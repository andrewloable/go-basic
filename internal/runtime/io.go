// io.go — Console input/output functions for the BASIC runtime.
//
// # Compiler Design Note: BASIC's I/O Model
//
// Turbo BASIC provides a rich set of console I/O statements that are more
// powerful than Go's fmt.Println:
//
//   - PRINT — outputs values separated by commas (zone tabs) or semicolons.
//   - PRINT USING — formats output using a mini format language.
//   - INPUT — reads comma-separated values from a single line.
//   - LINE INPUT — reads a full line (no comma splitting).
//   - LOCATE r,c — moves the text cursor using ANSI escape sequences.
//   - CLS — clears the screen using ANSI escape sequences.
//   - COLOR f,b — sets foreground/background using ANSI SGR codes.
//
// # ANSI Escape Sequences
//
// ANSI escape sequences are special byte sequences that most terminals
// interpret as commands rather than visible characters. They begin with the
// ESC character (byte 0x1B, written \033 in Go) followed by '[' and a
// parameter string ending with a letter.
//
// Examples used here:
//   - "\033[row;colH"   — move cursor to row, col (LOCATE)
//   - "\033[2J\033[H"   — clear screen and home cursor (CLS)
//   - "\033[fg;bgm"     — set SGR colour attributes (COLOR)
//
// The code generator emits fmt.Print(rt.AnsiLocate(r, c)) etc., so the
// runtime returns strings and the generated code decides when to write them.
//
// # PRINT USING Format Language
//
// PRINT USING is a mini format language embedded inside a BASIC string:
//
//   "#"     — a digit position (leading spaces for unused positions)
//   "."     — decimal point
//   ","     — thousands separator (only between # characters)
//   "+"     — force sign at leading or trailing position
//   "-"     — trailing minus for negative numbers (space for positive)
//   "$$"    — floating dollar sign (replaces leading space)
//   "**"    — asterisk fill for leading spaces
//   "**$"   — asterisk fill with floating dollar sign
//   "^^^^"  — scientific notation (appended after digit spec)
//   "!"     — first character of string only
//   "&"     — full string
//   "\ \"   — fixed-width string field (width = chars between backslashes + 2)
//   "_x"    — output literal character x (escape)
//
// Implementing PRINT USING requires a two-pass parser: first scan the format
// string into a list of literal and field segments, then apply values to fields
// cyclically. See PrintUsing, parseNumericField, FormatNumber, FormatString.

package runtime

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
)

// zoneWidth is the number of characters per PRINT zone (comma-separated values).
const zoneWidth = 14

// PrintZone formats a value into the next 14-character print zone.
//
// BASIC's PRINT statement separates values with commas or semicolons:
//   - Semicolon: no gap — values appear immediately adjacent.
//   - Comma: jump to the next 14-character tab stop (zone).
//
// This function handles the comma case. col is the current 0-based column;
// the function advances to the next multiple of zoneWidth and returns the
// padded string plus the new column position. The generated code tracks the
// column itself and passes it to each PrintZone call.
func PrintZone(col int, value string) (output string, newCol int) {
	// Advance to the next zone boundary.
	zone := col / zoneWidth
	targetCol := zone * zoneWidth
	if targetCol <= col {
		targetCol += zoneWidth
	}
	padding := targetCol - col
	out := strings.Repeat(" ", padding) + value
	newCol = targetCol + len(value)
	return out, newCol
}

// PrintUsing formats a value using a BASIC PRINT USING format string.
// Format characters for numbers:
//
//	# = digit position
//	. = decimal point
//	, = thousands separator (only between #'s before decimal)
//	+ = force sign (at start or end)
//	- = trailing minus for negatives (at end)
//	$$ = floating dollar sign
//	** = asterisk fill for leading spaces
//	**$ = asterisk fill with floating dollar
//	^^^^ = scientific notation
//
// Format characters for strings:
//
//	! = first character only
//	\ \ = fixed-width field (width = number of spaces between backslashes + 2)
//	& = full string
//	_ = literal next char
//
// The format string may contain multiple fields that are reused cyclically if there are
// more values than fields. Returns the formatted string.
func PrintUsing(format string, values []interface{}) string {
	if len(values) == 0 {
		return ""
	}

	// Parse the format string into a sequence of segments: either literal text
	// or format fields.
	type segment struct {
		isField bool
		text    string // literal text for non-field segments, format spec for fields
		isStr   bool   // true if the field is a string format field
	}

	var segments []segment
	var fields []int // indices into segments that are format fields

	i := 0
	fmtRunes := format

	for i < len(fmtRunes) {
		ch := fmtRunes[i]

		// Check for underscore (literal escape).
		if ch == '_' {
			if i+1 < len(fmtRunes) {
				segments = append(segments, segment{isField: false, text: string(fmtRunes[i+1])})
				i += 2
			} else {
				segments = append(segments, segment{isField: false, text: "_"})
				i++
			}
			continue
		}

		// Check for string format fields.
		if ch == '!' {
			idx := len(segments)
			segments = append(segments, segment{isField: true, text: "!", isStr: true})
			fields = append(fields, idx)
			i++
			continue
		}

		if ch == '&' {
			idx := len(segments)
			segments = append(segments, segment{isField: true, text: "&", isStr: true})
			fields = append(fields, idx)
			i++
			continue
		}

		if ch == '\\' {
			// Count characters between the two backslashes.
			j := i + 1
			for j < len(fmtRunes) && fmtRunes[j] != '\\' {
				j++
			}
			if j < len(fmtRunes) {
				// Width = j - i + 1 (includes both backslashes).
				fieldStr := fmtRunes[i : j+1]
				idx := len(segments)
				segments = append(segments, segment{isField: true, text: fieldStr, isStr: true})
				fields = append(fields, idx)
				i = j + 1
			} else {
				// No closing backslash; treat as literal.
				segments = append(segments, segment{isField: false, text: "\\"})
				i++
			}
			continue
		}

		// Check for numeric format fields.
		if ch == '#' || ch == '+' || ch == '$' || ch == '*' || (ch == '.' && i+1 < len(fmtRunes) && fmtRunes[i+1] == '#') {
			fieldStart := i
			field := parseNumericField(fmtRunes, &i)
			if len(field) > 0 && isNumericField(field) {
				idx := len(segments)
				segments = append(segments, segment{isField: true, text: field, isStr: false})
				fields = append(fields, idx)
			} else {
				// Not a valid numeric field; treat the first character as literal.
				segments = append(segments, segment{isField: false, text: string(fmtRunes[fieldStart])})
				i = fieldStart + 1
			}
			continue
		}

		// Check for '-' at the start potentially being a numeric field isn't standard;
		// '-' is a trailing sign indicator only. Treat as literal.

		// Everything else is literal text.
		segments = append(segments, segment{isField: false, text: string(ch)})
		i++
	}

	if len(fields) == 0 {
		// No format fields found; just return the format as-is.
		return format
	}

	// Now apply values to fields cyclically.
	var result strings.Builder
	fieldIdx := 0
	valIdx := 0

	for valIdx < len(values) {
		// Output segments from the current position in the segment list.
		// On the first pass, start from segment 0.
		// On subsequent passes (cycling), start from the first field.
		var startSeg, endSeg int
		if fieldIdx < len(fields) {
			if valIdx == 0 && fieldIdx == 0 {
				startSeg = 0
			} else {
				startSeg = fields[fieldIdx]
			}
		} else {
			// Cycle: reset field index, output from first field.
			fieldIdx = 0
			startSeg = fields[0]
		}

		// Find the end of this field's segment range (up to but not including next field or end).
		fldSegIdx := fields[fieldIdx]
		if fieldIdx+1 < len(fields) {
			endSeg = fields[fieldIdx+1]
		} else {
			endSeg = len(segments)
		}

		// Output literal segments before the field.
		for s := startSeg; s < fldSegIdx; s++ {
			result.WriteString(segments[s].text)
		}

		// Format the value using this field.
		seg := segments[fldSegIdx]
		v := values[valIdx]

		if seg.isStr {
			sv := valueToString(v)
			result.WriteString(FormatString(seg.text, sv))
		} else {
			nv := valueToFloat(v)
			result.WriteString(FormatNumber(seg.text, nv))
		}

		// Output literal segments after the field (up to endSeg).
		for s := fldSegIdx + 1; s < endSeg; s++ {
			result.WriteString(segments[s].text)
		}

		valIdx++
		fieldIdx++

		// If we've used all fields, cycle.
		if fieldIdx >= len(fields) && valIdx < len(values) {
			fieldIdx = 0
		}
	}

	return result.String()
}

// parseNumericField extracts a numeric format field starting at position *pos in the
// format string, advancing *pos past the field. Returns the field string.
func parseNumericField(format string, pos *int) string {
	i := *pos
	var field strings.Builder

	// Handle leading sign: +
	if i < len(format) && format[i] == '+' {
		field.WriteByte('+')
		i++
	}

	// Handle $$ (floating dollar sign).
	if i+1 < len(format) && format[i] == '$' && format[i+1] == '$' {
		field.WriteString("$$")
		i += 2
	}

	// Handle ** (asterisk fill) and **$ (asterisk fill with floating dollar).
	if i+1 < len(format) && format[i] == '*' && format[i+1] == '*' {
		field.WriteString("**")
		i += 2
		if i < len(format) && format[i] == '$' {
			field.WriteByte('$')
			i++
		}
	}

	// Collect # and , before decimal point.
	hasHash := false
	for i < len(format) && (format[i] == '#' || format[i] == ',') {
		field.WriteByte(format[i])
		if format[i] == '#' {
			hasHash = true
		}
		i++
	}

	// Decimal point and fractional digits.
	if i < len(format) && format[i] == '.' {
		field.WriteByte('.')
		i++
		for i < len(format) && format[i] == '#' {
			field.WriteByte('#')
			hasHash = true
			i++
		}
	}

	// Scientific notation: ^^^^
	if i+3 < len(format) && format[i] == '^' && format[i+1] == '^' && format[i+2] == '^' && format[i+3] == '^' {
		field.WriteString("^^^^")
		i += 4
		hasHash = true // allow scientific notation fields
	}

	// Trailing sign: + or -
	if i < len(format) && (format[i] == '+' || format[i] == '-') {
		// Only append if the field has content (to avoid consuming a standalone + or -).
		if field.Len() > 0 {
			field.WriteByte(format[i])
			i++
		}
	}

	if !hasHash && !strings.Contains(field.String(), "**") && !strings.Contains(field.String(), "$$") {
		// Not a valid numeric field. Restore position.
		return ""
	}

	*pos = i
	return field.String()
}

// isNumericField returns true if the field string contains numeric format characters.
func isNumericField(field string) bool {
	for _, ch := range field {
		switch ch {
		case '#', '.', ',', '+', '-', '$', '*', '^':
			return true
		}
	}
	return false
}

// FormatNumber formats a single number according to a PRINT USING numeric format field.
// If the number overflows the format field, the result is prefixed with %.
func FormatNumber(format string, value float64) string {
	if len(format) == 0 {
		return fmt.Sprintf("%g", value)
	}

	// Parse the format field to extract components.
	i := 0

	// Leading sign.
	leadingPlus := false
	if i < len(format) && format[i] == '+' {
		leadingPlus = true
		i++
	}

	// Floating dollar.
	floatingDollar := false
	if i+1 < len(format) && format[i] == '$' && format[i+1] == '$' {
		floatingDollar = true
		i += 2
	}

	// Asterisk fill.
	asteriskFill := false
	asteriskDollar := false
	if i+1 < len(format) && format[i] == '*' && format[i+1] == '*' {
		asteriskFill = true
		i += 2
		if i < len(format) && format[i] == '$' {
			asteriskDollar = true
			i++
		}
	}

	// Count integer # and commas.
	intDigits := 0
	hasComma := false
	for i < len(format) && (format[i] == '#' || format[i] == ',') {
		if format[i] == '#' {
			intDigits++
		} else {
			hasComma = true
		}
		i++
	}

	// Decimal point and fractional digits.
	decDigits := 0
	hasDecimal := false
	if i < len(format) && format[i] == '.' {
		hasDecimal = true
		i++
		for i < len(format) && format[i] == '#' {
			decDigits++
			i++
		}
	}

	// Scientific notation.
	scientific := false
	if i+3 < len(format) && format[i] == '^' && format[i+1] == '^' && format[i+2] == '^' && format[i+3] == '^' {
		scientific = true
		i += 4
	}

	// Trailing sign.
	trailingPlus := false
	trailingMinus := false
	if i < len(format) {
		if format[i] == '+' {
			trailingPlus = true
			i++
		} else if format[i] == '-' {
			trailingMinus = true
			i++
		}
	}

	negative := value < 0
	absVal := math.Abs(value)

	// The floating dollar and ** each contribute digit positions.
	// $$ provides one digit position (the second $ is replaced by the value).
	// ** provides two digit positions.
	// **$ provides two digit positions (from **) plus the dollar sign.
	totalIntDigits := intDigits
	if floatingDollar {
		totalIntDigits++ // One extra digit position from $$.
	}
	if asteriskFill {
		totalIntDigits += 2 // Two extra digit positions from **.
	}

	if scientific {
		return formatScientific(absVal, negative, totalIntDigits, decDigits, hasDecimal,
			leadingPlus, trailingPlus, trailingMinus, floatingDollar, asteriskDollar)
	}

	// Round the value to the specified decimal places.
	rounded := absVal
	if hasDecimal {
		factor := math.Pow(10, float64(decDigits))
		rounded = math.Round(absVal*factor) / factor
	} else {
		rounded = math.Round(absVal)
	}

	// Format the number.
	var numStr string
	if hasDecimal {
		numStr = fmt.Sprintf("%.*f", decDigits, rounded)
	} else {
		numStr = fmt.Sprintf("%.0f", rounded)
	}

	// Split into integer and decimal parts.
	intPart := numStr
	decPart := ""
	if dotIdx := strings.Index(numStr, "."); dotIdx >= 0 {
		intPart = numStr[:dotIdx]
		decPart = numStr[dotIdx+1:]
	}

	// Apply thousands separator if requested.
	if hasComma && len(intPart) > 3 {
		intPart = addThousandsSep(intPart)
	}

	// Build the digit portion.
	var digitStr string
	if hasDecimal {
		digitStr = intPart + "." + decPart
	} else {
		digitStr = intPart
	}

	// Calculate the total field width for the numeric portion.
	// The field width is determined by the total # positions, decimal point, and commas.
	fieldWidth := totalIntDigits
	if hasDecimal {
		fieldWidth += 1 + decDigits // dot + fractional digits
	}

	// Determine the sign character.
	signChar := ""
	if leadingPlus {
		if negative {
			signChar = "-"
		} else {
			signChar = "+"
		}
	} else if trailingPlus || trailingMinus {
		// Sign goes at end; we still need space in the integer part.
	} else {
		// Default: negative numbers get a "-" that takes space from the field.
		if negative {
			signChar = "-"
		}
	}

	// Check for overflow: the digit string (without sign) must fit in fieldWidth.
	overflow := false
	if len(digitStr) > fieldWidth {
		overflow = true
	}

	// Pad or fill the result.
	var result strings.Builder

	if overflow {
		result.WriteByte('%')
		// Show the number as best we can.
		if signChar != "" && !trailingPlus && !trailingMinus {
			result.WriteString(signChar)
		}
		if floatingDollar || asteriskDollar {
			result.WriteByte('$')
		}
		result.WriteString(digitStr)
		if trailingPlus {
			if negative {
				result.WriteByte('-')
			} else {
				result.WriteByte('+')
			}
		} else if trailingMinus {
			if negative {
				result.WriteByte('-')
			} else {
				result.WriteByte(' ')
			}
		}
		return result.String()
	}

	// Calculate padding.
	padLen := fieldWidth - len(digitStr)

	// Build the leading portion.
	if leadingPlus || (!trailingPlus && !trailingMinus && negative) {
		// Sign is part of the leading area.
		if asteriskFill {
			// Fill with asterisks, then sign, then dollar if applicable.
			if padLen > 0 {
				result.WriteString(strings.Repeat("*", padLen))
			}
			result.WriteString(signChar)
			if asteriskDollar {
				result.WriteByte('$')
			}
		} else if floatingDollar {
			if padLen > 0 {
				result.WriteString(strings.Repeat(" ", padLen))
			}
			result.WriteString(signChar)
			result.WriteByte('$')
		} else {
			if padLen > 0 {
				result.WriteString(strings.Repeat(" ", padLen))
			}
			result.WriteString(signChar)
		}
	} else {
		// No leading sign (or trailing sign handles it).
		if asteriskFill {
			if asteriskDollar {
				if padLen > 0 {
					result.WriteString(strings.Repeat("*", padLen))
				}
				result.WriteByte('$')
			} else {
				if padLen > 0 {
					result.WriteString(strings.Repeat("*", padLen))
				}
			}
		} else if floatingDollar {
			if padLen > 0 {
				result.WriteString(strings.Repeat(" ", padLen))
			}
			result.WriteByte('$')
		} else {
			if padLen > 0 {
				result.WriteString(strings.Repeat(" ", padLen))
			}
		}
	}

	result.WriteString(digitStr)

	// Trailing sign.
	if trailingPlus {
		if negative {
			result.WriteByte('-')
		} else {
			result.WriteByte('+')
		}
	} else if trailingMinus {
		if negative {
			result.WriteByte('-')
		} else {
			result.WriteByte(' ')
		}
	}

	return result.String()
}

// formatScientific formats a number in scientific notation for PRINT USING with ^^^^.
func formatScientific(absVal float64, negative bool, intDigits, decDigits int,
	hasDecimal, leadingPlus, trailingPlus, trailingMinus, floatingDollar, asteriskDollar bool) string {

	// Total digit positions before the exponent.
	totalDigits := intDigits
	if hasDecimal {
		totalDigits += decDigits
	}

	// Format in scientific notation with the right number of significant digits.
	// BASIC scientific notation: d.ddd^^^^  where ^^^^ becomes E+nn
	sigDigits := totalDigits
	if sigDigits < 1 {
		sigDigits = 1
	}

	var expStr string
	if absVal == 0 {
		expStr = fmt.Sprintf("%.*fE+00", sigDigits-1, 0.0)
	} else {
		exp := int(math.Floor(math.Log10(absVal)))
		mantissa := absVal / math.Pow(10, float64(exp))
		// Adjust so we have intDigits digits before the decimal.
		if intDigits > 1 {
			mantissa = mantissa * math.Pow(10, float64(intDigits-1))
			exp = exp - (intDigits - 1)
		}
		// Round mantissa.
		if hasDecimal {
			factor := math.Pow(10, float64(decDigits))
			mantissa = math.Round(mantissa*factor) / factor
		} else {
			mantissa = math.Round(mantissa)
		}

		sign := "+"
		absExp := exp
		if exp < 0 {
			sign = "-"
			absExp = -exp
		}
		if hasDecimal {
			expStr = fmt.Sprintf("%.*f", decDigits, mantissa)
		} else {
			expStr = fmt.Sprintf("%.0f", mantissa)
		}
		expStr = fmt.Sprintf("%sE%s%02d", expStr, sign, absExp)
	}

	var result strings.Builder

	// Leading sign.
	if leadingPlus {
		if negative {
			result.WriteByte('-')
		} else {
			result.WriteByte('+')
		}
	} else if negative {
		result.WriteByte('-')
	}

	if floatingDollar || asteriskDollar {
		result.WriteByte('$')
	}

	result.WriteString(expStr)

	// Trailing sign.
	if trailingPlus {
		if negative {
			result.WriteByte('-')
		} else {
			result.WriteByte('+')
		}
	} else if trailingMinus {
		if negative {
			result.WriteByte('-')
		} else {
			result.WriteByte(' ')
		}
	}

	return result.String()
}

// addThousandsSep inserts commas as thousands separators into an integer string.
func addThousandsSep(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var result strings.Builder
	firstGroup := n % 3
	if firstGroup == 0 {
		firstGroup = 3
	}
	result.WriteString(s[:firstGroup])
	for i := firstGroup; i < n; i += 3 {
		result.WriteByte(',')
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// FormatString formats a single string according to a PRINT USING string format field.
//
//	! = first character only
//	\ \ = fixed-width field (width = number of chars between backslashes + 2)
//	& = full string
func FormatString(format string, value string) string {
	if len(format) == 0 {
		return value
	}

	switch {
	case format == "!":
		if len(value) > 0 {
			return string(value[0])
		}
		return " "

	case format == "&":
		return value

	case format[0] == '\\' && format[len(format)-1] == '\\':
		// Width = len(format): the two backslashes plus the spaces between them.
		width := len(format)
		if len(value) >= width {
			return value[:width]
		}
		// Pad with trailing spaces.
		return value + strings.Repeat(" ", width-len(value))

	default:
		return value
	}
}

// Tab returns spaces needed to move to column n (1-based).
// currentCol is the current 1-based column. targetCol is the desired 1-based column.
// If already past targetCol, moves to targetCol on the next line (returns newline + spaces).
func Tab(currentCol, targetCol int) string {
	if targetCol < 1 {
		targetCol = 1
	}
	if currentCol < targetCol {
		return strings.Repeat(" ", targetCol-currentCol)
	}
	// Already at or past the target column; go to next line.
	return "\n" + strings.Repeat(" ", targetCol-1)
}

// Spc returns a string of n spaces.
func Spc(n int) string {
	if n < 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

// InputPrompt displays a prompt to stdout and reads one line from stdin.
// BASIC: INPUT "prompt"; var — the runtime prints the prompt then waits.
// The returned string is the raw line with the trailing newline stripped; the
// generated code then converts it to the variable's target type.
func InputPrompt(prompt string) string {
	fmt.Print(prompt)
	if IsGraphicsWindowOpen() {
		return GraphicsReadLine()
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return strings.TrimRight(line, "\r\n")
	}
	return strings.TrimRight(line, "\r\n")
}

// InputScanner is an interface that provides Scan/Text methods for INPUT.
// When the graphics window is open, a GraphicsLineScanner is used instead
// of bufio.Scanner so that keyboard input comes from the Ebitengine key buffer.
type InputScanner interface {
	Scan() bool
	Text() string
}

// GraphicsLineScanner reads one line from the Ebitengine key buffer.
type GraphicsLineScanner struct {
	line string
}

func (s *GraphicsLineScanner) Scan() bool {
	s.line = GraphicsReadLine()
	return true
}

func (s *GraphicsLineScanner) Text() string {
	return s.line
}

// NewScanner creates a scanner for INPUT/LINE INPUT statements.
// When the graphics window is open, returns a GraphicsLineScanner that reads
// from the Ebitengine key buffer. Otherwise returns a bufio.Scanner for stdin.
func NewScanner() InputScanner {
	if IsGraphicsWindowOpen() {
		return &GraphicsLineScanner{}
	}
	return bufio.NewScanner(os.Stdin)
}

// InputSplitLine reads one line from stdin and splits it on commas.
//
// BASIC's INPUT statement accepts multiple variables on a single line:
//
//	INPUT a, b, c
//
// The user types "1, 2, 3" and BASIC assigns each comma-delimited token to the
// corresponding variable. The generated code calls InputSplitLine() once and
// then indexes the returned slice — one element per variable in the INPUT list.
func InputSplitLine() []string {
	var line string
	if IsGraphicsWindowOpen() {
		line = GraphicsReadLine()
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		line = scanner.Text()
	}
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// AnsiLocate returns the ANSI CSI escape sequence that moves the terminal
// cursor to the given row and column (both 1-based, matching BASIC's LOCATE).
// The generated code calls fmt.Print(rt.AnsiLocate(r, c)) for LOCATE r,c.
func AnsiLocate(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

// AnsiCls returns the ANSI escape sequence that clears the terminal screen
// and moves the cursor to the top-left corner (home position).
func AnsiCls() string {
	return "\033[2J\033[H"
}

// Cls implements the BASIC CLS statement.
//
//	CLS 0 (or no arg) — clear everything (graphics framebuffer + text)
//	CLS 1             — clear graphics framebuffer only
//	CLS 2             — clear text viewport area of the graphics framebuffer
//
// In text mode (SCREEN 0), CLS sends ANSI clear. In graphics modes,
// CLS clears the appropriate portion of the framebuffer and/or text buffer
// directly (without going through the pipe, which would trigger a redundant
// ClearScreen via the text writer's \033[2J handler).
func Cls(mode int) {
	s := CurrentScreen
	if s.Framebuffer != nil {
		switch mode {
		case 0:
			// Clear entire framebuffer.
			for y := 0; y < s.Height; y++ {
				for x := 0; x < s.Width; x++ {
					s.Framebuffer[y][x] = 0
				}
			}
			// Also clear the text buffer directly (no ANSI pipe).
			if TextBuf != nil {
				TextBuf.Clear()
			}
		case 1:
			// Clear graphics only.
			for y := 0; y < s.Height; y++ {
				for x := 0; x < s.Width; x++ {
					s.Framebuffer[y][x] = 0
				}
			}
		case 2:
			// Clear the text viewport region of the framebuffer.
			// VIEW PRINT top TO bottom defines text rows (1-based).
			top := s.TextTop
			bottom := s.TextBottom
			if top == 0 && bottom == 0 {
				top = 1
				bottom = 25
			}
			charH := s.Height / 25
			if charH < 1 {
				charH = 1
			}
			py1 := (top - 1) * charH
			py2 := bottom * charH
			if py2 > s.Height {
				py2 = s.Height
			}
			for y := py1; y < py2; y++ {
				for x := 0; x < s.Width; x++ {
					s.Framebuffer[y][x] = 0
				}
			}
		}
		return
	}
	// Text mode (no framebuffer): send ANSI clear.
	fmt.Print(AnsiCls())
}

// AnsiColor returns the ANSI SGR escape sequence for the given BASIC color codes.
//
// BASIC COLOR fg, bg uses a 16-color palette derived from CGA hardware:
//
//	0=black, 1=blue, 2=green, 3=cyan, 4=red, 5=magenta,
//	6=brown/dark yellow, 7=white/light gray, 8=dark gray, 9=light blue,
//	10=light green, 11=light cyan, 12=light red, 13=light magenta,
//	14=yellow, 15=bright white
//
// ANSI SGR foreground codes are 30-37 (normal) and 90-97 (bright). Background
// codes are foreground + 10. The mapping is not a simple offset because BASIC
// and ANSI order their colours differently — the switch tables in basicToAnsiFg
// and basicToAnsiBg encode the translation explicitly.
func AnsiColor(fg, bg int) string {
	fgCode := basicToAnsiFg(fg)
	bgCode := basicToAnsiBg(bg)
	return fmt.Sprintf("\033[%d;%dm", fgCode, bgCode)
}

// basicToAnsiFg converts a BASIC color code (0-15) to an ANSI foreground SGR code.
func basicToAnsiFg(c int) int {
	// BASIC color -> ANSI foreground code mapping.
	// BASIC colors 0-7 map to ANSI 30-37 (but reordered to match BASIC's scheme).
	// BASIC colors 8-15 map to ANSI 90-97 (bright variants).
	switch c {
	case 0:
		return 30 // black
	case 1:
		return 34 // blue
	case 2:
		return 32 // green
	case 3:
		return 36 // cyan
	case 4:
		return 31 // red
	case 5:
		return 35 // magenta
	case 6:
		return 33 // brown/dark yellow
	case 7:
		return 37 // white/light gray
	case 8:
		return 90 // dark gray (bright black)
	case 9:
		return 94 // light blue
	case 10:
		return 92 // light green
	case 11:
		return 96 // light cyan
	case 12:
		return 91 // light red
	case 13:
		return 95 // light magenta
	case 14:
		return 93 // yellow (bright dark yellow)
	case 15:
		return 97 // bright white
	default:
		return 37 // default to white
	}
}

// basicToAnsiBg converts a BASIC color code (0-15) to an ANSI background SGR code.
func basicToAnsiBg(c int) int {
	// Background codes are foreground codes + 10.
	switch c {
	case 0:
		return 40 // black
	case 1:
		return 44 // blue
	case 2:
		return 42 // green
	case 3:
		return 46 // cyan
	case 4:
		return 41 // red
	case 5:
		return 45 // magenta
	case 6:
		return 43 // brown/dark yellow
	case 7:
		return 47 // white/light gray
	case 8:
		return 100 // dark gray
	case 9:
		return 104 // light blue
	case 10:
		return 102 // light green
	case 11:
		return 106 // light cyan
	case 12:
		return 101 // light red
	case 13:
		return 105 // light magenta
	case 14:
		return 103 // yellow
	case 15:
		return 107 // bright white
	default:
		return 40 // default to black
	}
}

// valueToString converts an interface{} value to a string for PRINT USING.
func valueToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// valueToFloat converts an interface{} value to a float64 for PRINT USING.
func valueToFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	default:
		// Try fmt.Sscanf as fallback.
		var f float64
		fmt.Sscanf(fmt.Sprintf("%v", v), "%g", &f)
		return f
	}
}
