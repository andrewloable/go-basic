// strings.go — String manipulation built-ins for the BASIC runtime.
//
// # Compiler Design Note: Bridging 1-Based and 0-Based Indexing
//
// One of the most pervasive translation challenges between BASIC and Go is
// string indexing. In Turbo BASIC:
//
//   - Strings are 1-based: the first character is at position 1.
//   - MID$("Hello", 1, 3) = "Hel"  (start=1, length=3)
//   - LEFT$ and RIGHT$ count from the respective ends.
//
// In Go, strings are 0-based slices: s[0] is the first byte. Every function
// in this file that accepts a BASIC position must convert it with:
//
//	goIndex = basicPosition - 1
//
// Forgetting this off-by-one is a classic source of bugs when implementing
// string functions for a transpiler. The conversion is done once here, in the
// runtime, so the code generator never has to worry about it — it simply emits
// rt.Mid(s, start, length) and the adjustment happens automatically.
//
// # Binary Encoding Functions (MKI$, MKD$, CVI, CVD, …)
//
// BASIC used strings as raw byte arrays for binary file I/O. MKI$ converts an
// integer to a 2-byte little-endian string, CVI reverses it. These are legacy
// functions used with FIELD/GET/PUT random-access file operations.

package runtime

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// Left returns the leftmost n characters of s.
// BASIC: LEFT$(s, n) — no index adjustment needed; counts from the left edge.
// If n > len(s), returns s unchanged. If n < 0, returns "".
func Left(s string, n int) string {
	if n < 0 {
		return ""
	}
	if n >= len(s) {
		return s
	}
	return s[:n]
}

// Right returns the rightmost n characters of s.
// BASIC: RIGHT$(s, n) — counts from the right edge; no 1-based adjustment needed.
// If n > len(s), returns s unchanged. If n < 0, returns "".
func Right(s string, n int) string {
	if n < 0 {
		return ""
	}
	if n >= len(s) {
		return s
	}
	return s[len(s)-n:]
}

// Mid returns a substring starting at 1-based position start with the given length.
// BASIC: MID$(s, start[, length]) — the central string function in BASIC programs.
//
// Key index translation: BASIC start=1 maps to Go index 0 (idx = start - 1).
// When length is -1, the caller wants "from start to end of string."
// If start < 1, it is treated as 1. If start > len(s), returns "".
func Mid(s string, start, length int) string {
	if start < 1 {
		start = 1
	}
	if start > len(s) {
		return ""
	}
	// Convert to 0-based index.
	idx := start - 1
	if length < 0 {
		// Return from start to end.
		return s[idx:]
	}
	end := idx + length
	if end > len(s) {
		end = len(s)
	}
	return s[idx:end]
}

// Instr finds the first occurrence of find in s beginning at 1-based position start.
// BASIC: INSTR([start,] s, find$) — returns a 1-based position, or 0 if not found.
//
// This function must both accept a 1-based start and return a 1-based result,
// so it performs two index translations:
//  1. Convert input start (1-based) to Go slice index (0-based): idx = start - 1.
//  2. Convert the result of strings.Index (0-based) back to 1-based: return idx+pos+1.
//
// If start is 0, it defaults to 1 (permissive BASIC compatibility).
func Instr(start int, s, find string) int {
	if start <= 0 {
		start = 1
	}
	if start > len(s) {
		return 0
	}
	// Convert to 0-based index for the search region.
	idx := start - 1
	pos := strings.Index(s[idx:], find)
	if pos < 0 {
		return 0
	}
	// Convert back to 1-based position.
	return idx + pos + 1
}

// Len returns the length of a string.
func Len(s string) int {
	return len(s)
}

// Asc returns the ASCII code of the first character of s.
// BASIC: ASC(s$) — inverse of CHR$. Returns error for empty string because
// BASIC raises "Illegal function call" in that case rather than returning 0.
func Asc(s string) (int, error) {
	if len(s) == 0 {
		return 0, errors.New("illegal function call: empty string")
	}
	return int(s[0]), nil
}

// Chr returns the single-character string whose ASCII code is n (0-255).
// BASIC: CHR$(n) — used for printing control characters, box-drawing, etc.
// Returns an error if n is outside [0, 255].
func Chr(n int) (string, error) {
	if n < 0 || n > 255 {
		return "", errors.New("illegal function call: argument out of range (0-255)")
	}
	return string(byte(n)), nil
}

// Str converts a number to its string representation.
// BASIC: STR$(n) — positive numbers include a leading space where a minus sign
// would appear; this is a quirk of BASIC's formatting convention preserved here
// for fidelity. Inverse of VAL.
func Str(n float64) string {
	// BASIC behavior: positive numbers have a leading space where the sign would be.
	if n == 0 {
		// Avoid "-0".
		return " 0"
	}
	if n < 0 {
		return fmt.Sprintf("%g", n)
	}
	return fmt.Sprintf(" %g", n)
}

// Val converts a string to a float64, consuming as many numeric characters as
// possible from the left and ignoring the rest.
// BASIC: VAL(s$) — stops at the first character that cannot be part of a number.
// Handles leading whitespace, optional sign, decimal point, and E notation.
// Returns 0 for strings with no numeric prefix.
//
// Implementing VAL is a mini-lexer exercise: the function hand-rolls a numeric
// scanner rather than using strconv.ParseFloat so it can apply BASIC's exact
// stop-on-first-non-digit rule (which differs from Go's ParseFloat error model).
func Val(s string) float64 {
	// Skip leading whitespace.
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	if i >= len(s) {
		return 0
	}

	start := i

	// Optional sign.
	if s[i] == '+' || s[i] == '-' {
		i++
	}

	hasDigits := false

	// Integer part.
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		hasDigits = true
		i++
	}

	// Decimal point and fractional part.
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			hasDigits = true
			i++
		}
	}

	if !hasDigits {
		return 0
	}

	// Exponent part (E or e notation).
	if i < len(s) && (s[i] == 'E' || s[i] == 'e') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		expHasDigits := false
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			expHasDigits = true
			j++
		}
		if expHasDigits {
			i = j
		}
		// If no digits after E, ignore the E part and stop before it.
	}

	// Parse the valid numeric portion.
	numStr := s[start:i]
	var result float64
	_, err := fmt.Sscanf(numStr, "%g", &result)
	if err != nil {
		return 0
	}
	return result
}

// Hex returns the uppercase hexadecimal string for n, with no "0x" prefix.
// BASIC: HEX$(n) — negative values are treated as unsigned 16-bit, matching
// Turbo BASIC's behaviour on the 16-bit DOS platform.
func Hex(n int) string {
	if n < 0 {
		// Turbo BASIC treats negative numbers as unsigned 16-bit.
		return fmt.Sprintf("%X", uint16(n))
	}
	return fmt.Sprintf("%X", n)
}

// Oct returns the octal string for n, with no "0o" prefix.
// BASIC: OCT$(n) — like HEX$, negatives are treated as unsigned 16-bit.
func Oct(n int) string {
	if n < 0 {
		return fmt.Sprintf("%o", uint16(n))
	}
	return fmt.Sprintf("%o", n)
}

// Bin returns the binary string representation of an integer.
// Output has no prefix. Turbo BASIC extension.
func Bin(n int) string {
	if n < 0 {
		return fmt.Sprintf("%b", uint16(n))
	}
	return fmt.Sprintf("%b", n)
}

// UCase converts a string to uppercase.
func UCase(s string) string {
	return strings.ToUpper(s)
}

// LCase converts a string to lowercase.
func LCase(s string) string {
	return strings.ToLower(s)
}

// LTrim removes leading spaces from a string.
func LTrim(s string) string {
	return strings.TrimLeftFunc(s, func(r rune) bool {
		return r == ' '
	})
}

// RTrim removes trailing spaces from a string.
func RTrim(s string) string {
	return strings.TrimRightFunc(s, func(r rune) bool {
		return r == ' '
	})
}

// Trim removes leading and trailing spaces from a string. Turbo BASIC extension.
func Trim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r == ' '
	})
}

// Space returns a string of n spaces. If n < 0, returns "".
func Space(n int) string {
	if n < 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

// StringRepeat returns a string of n repetitions of the byte char.
// Corresponds to BASIC's STRING$(n, char) function.
// If n < 0, returns "".
func StringRepeat(n int, char byte) string {
	if n < 0 {
		return ""
	}
	return strings.Repeat(string(char), n)
}

// Mki converts an int16 to a 2-byte little-endian binary string.
// BASIC: MKI$(n) — used to store integers in fixed-length random-access file
// records. BASIC's strings were just byte arrays, so binary data was packed
// into them. The file is then read back with CVI to recover the int16.
func Mki(n int16) string {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(n))
	return string(buf)
}

// Mkl converts an int32 to a 4-byte little-endian binary string.
// BASIC: MKL$(n) — like MKI$ but for long integer (&) values (4 bytes).
func Mkl(n int32) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(n))
	return string(buf)
}

// Mks converts a float32 to a 4-byte IEEE 754 little-endian binary string.
// BASIC: MKS$(n) — stores a single-precision float (!) in a file record.
func Mks(n float32) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(n))
	return string(buf)
}

// Mkd converts a float64 to an 8-byte IEEE 754 little-endian binary string.
// BASIC: MKD$(n) — stores a double-precision float (#) in a file record.
func Mkd(n float64) string {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(n))
	return string(buf)
}

// Cvi converts a 2-byte string to int16 (little-endian).
// Corresponds to BASIC's CVI function.
func Cvi(s string) (int16, error) {
	if len(s) < 2 {
		return 0, errors.New("illegal function call: string too short for CVI (need 2 bytes)")
	}
	v := binary.LittleEndian.Uint16([]byte(s[:2]))
	return int16(v), nil
}

// Cvl converts a 4-byte string to int32 (little-endian).
// Corresponds to BASIC's CVL function.
func Cvl(s string) (int32, error) {
	if len(s) < 4 {
		return 0, errors.New("illegal function call: string too short for CVL (need 4 bytes)")
	}
	v := binary.LittleEndian.Uint32([]byte(s[:4]))
	return int32(v), nil
}

// Cvs converts a 4-byte string to float32 (IEEE 754 little-endian).
// Corresponds to BASIC's CVS function.
func Cvs(s string) (float32, error) {
	if len(s) < 4 {
		return 0, errors.New("illegal function call: string too short for CVS (need 4 bytes)")
	}
	bits := binary.LittleEndian.Uint32([]byte(s[:4]))
	return math.Float32frombits(bits), nil
}

// Cvd converts an 8-byte string to float64 (IEEE 754 little-endian).
// Corresponds to BASIC's CVD function.
func Cvd(s string) (float64, error) {
	if len(s) < 8 {
		return 0, errors.New("illegal function call: string too short for CVD (need 8 bytes)")
	}
	bits := binary.LittleEndian.Uint64([]byte(s[:8]))
	return math.Float64frombits(bits), nil
}

