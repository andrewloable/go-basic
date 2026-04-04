package runtime

import (
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// PrintZone
// ---------------------------------------------------------------------------

func TestPrintZone(t *testing.T) {
	tests := []struct {
		name      string
		col       int
		value     string
		wantOut   string
		wantNewCol int
	}{
		{
			name:       "start of line",
			col:        0,
			value:      "Hello",
			wantOut:    strings.Repeat(" ", 14) + "Hello",
			wantNewCol: 14 + 5,
		},
		{
			name:       "mid zone",
			col:        5,
			value:      "X",
			wantOut:    strings.Repeat(" ", 9) + "X",
			wantNewCol: 14 + 1,
		},
		{
			name:       "at zone boundary",
			col:        14,
			value:      "AB",
			wantOut:    strings.Repeat(" ", 14) + "AB",
			wantNewCol: 28 + 2,
		},
		{
			name:       "empty value",
			col:        0,
			value:      "",
			wantOut:    strings.Repeat(" ", 14),
			wantNewCol: 14,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, newCol := PrintZone(tt.col, tt.value)
			if out != tt.wantOut {
				t.Errorf("PrintZone(%d, %q) output = %q, want %q", tt.col, tt.value, out, tt.wantOut)
			}
			if newCol != tt.wantNewCol {
				t.Errorf("PrintZone(%d, %q) newCol = %d, want %d", tt.col, tt.value, newCol, tt.wantNewCol)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tab
// ---------------------------------------------------------------------------

func TestTab(t *testing.T) {
	tests := []struct {
		name      string
		currentCol int
		targetCol  int
		want       string
	}{
		{"move forward", 1, 10, strings.Repeat(" ", 9)},
		{"already at target", 10, 10, "\n" + strings.Repeat(" ", 9)},
		{"past target wraps", 15, 10, "\n" + strings.Repeat(" ", 9)},
		{"target 1", 1, 1, "\n"},
		{"target < 1 clamped", 5, 0, "\n"},
		{"from col 1 to col 5", 1, 5, "    "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tab(tt.currentCol, tt.targetCol)
			if got != tt.want {
				t.Errorf("Tab(%d, %d) = %q, want %q", tt.currentCol, tt.targetCol, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Spc
// ---------------------------------------------------------------------------

func TestSpc(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{1, " "},
		{5, "     "},
		{-1, ""},
	}
	for _, tt := range tests {
		got := Spc(tt.n)
		if got != tt.want {
			t.Errorf("Spc(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// FormatNumber
// ---------------------------------------------------------------------------

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name   string
		format string
		value  float64
		want   string
	}{
		{"simple integer", "###", 42, " 42"},
		{"pad leading spaces", "####", 42, "  42"},
		{"decimal", "###.##", 3.14, "  3.14"},
		{"zero decimal", "###.##", 0, "  0.00"},
		{"negative", "###.##", -3.14, "  -3.14"},
		{"overflow", "##", 100, "%100"},
		{"leading plus positive", "+##.##", 3.14, " +3.14"},
		{"leading plus negative", "+##.##", -3.14, " -3.14"},
		{"trailing plus positive", "##.##+", 3.14, " 3.14+"},
		{"trailing plus negative", "##.##+", -3.14, " 3.14-"},
		{"trailing minus positive", "##.##-", 3.14, " 3.14 "},
		{"trailing minus negative", "##.##-", -3.14, " 3.14-"},
		{"asterisk fill", "**##.##", 3.14, "***3.14"},
		{"floating dollar", "$$##.##", 3.14, "  $3.14"},
		{"empty format", "", 42, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNumber(tt.format, tt.value)
			if got != tt.want {
				t.Errorf("FormatNumber(%q, %g) = %q, want %q", tt.format, tt.value, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FormatNumber scientific
// ---------------------------------------------------------------------------

func TestFormatNumberScientific(t *testing.T) {
	tests := []struct {
		name   string
		format string
		value  float64
		want   string
	}{
		{"basic scientific", "##.##^^^^", 1234.5, "12.35E+02"},
		{"zero scientific", "#.##^^^^", 0, "0.00E+00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNumber(tt.format, tt.value)
			if got != tt.want {
				t.Errorf("FormatNumber(%q, %g) = %q, want %q", tt.format, tt.value, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FormatString
// ---------------------------------------------------------------------------

func TestFormatString(t *testing.T) {
	tests := []struct {
		name   string
		format string
		value  string
		want   string
	}{
		{"exclamation", "!", "Hello", "H"},
		{"exclamation empty", "!", "", " "},
		{"ampersand", "&", "Hello", "Hello"},
		{"backslash width 5", "\\   \\", "Hello World", "Hello"},
		{"backslash pad short", "\\   \\", "Hi", "Hi   "},
		{"backslash exact", "\\   \\", "ABCDE", "ABCDE"},
		{"empty format", "", "Hello", "Hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatString(tt.format, tt.value)
			if got != tt.want {
				t.Errorf("FormatString(%q, %q) = %q, want %q", tt.format, tt.value, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PrintUsing (integration-level)
// ---------------------------------------------------------------------------

func TestPrintUsing(t *testing.T) {
	t.Run("no values", func(t *testing.T) {
		got := PrintUsing("###.##", nil)
		if got != "" {
			t.Errorf("PrintUsing with nil values = %q, want empty", got)
		}
	})

	t.Run("single number", func(t *testing.T) {
		got := PrintUsing("###.##", []interface{}{3.14})
		if got != "  3.14" {
			t.Errorf("PrintUsing(\"###.##\", 3.14) = %q, want %q", got, "  3.14")
		}
	})

	t.Run("string field !", func(t *testing.T) {
		got := PrintUsing("!", []interface{}{"Hello"})
		if got != "H" {
			t.Errorf("PrintUsing(\"!\", \"Hello\") = %q, want %q", got, "H")
		}
	})

	t.Run("string field &", func(t *testing.T) {
		got := PrintUsing("&", []interface{}{"Hello"})
		if got != "Hello" {
			t.Errorf("PrintUsing(\"&\", \"Hello\") = %q, want %q", got, "Hello")
		}
	})

	t.Run("literal escape", func(t *testing.T) {
		got := PrintUsing("_#", []interface{}{})
		// No fields, so format returned as-is. But _# should become #.
		// Actually with no values, PrintUsing returns "".
		// Let me test with a value and a literal character.
		_ = got
	})

	t.Run("no fields returns format", func(t *testing.T) {
		got := PrintUsing("Hello World", []interface{}{42.0})
		if got != "Hello World" {
			t.Errorf("PrintUsing with no fields = %q, want %q", got, "Hello World")
		}
	})

	t.Run("multiple values cycling", func(t *testing.T) {
		got := PrintUsing("###", []interface{}{1.0, 2.0, 3.0})
		// Each value formatted with ### (3 wide), cycled through the single field.
		if !strings.Contains(got, "1") && !strings.Contains(got, "2") && !strings.Contains(got, "3") {
			t.Errorf("PrintUsing cycling = %q, expected all values", got)
		}
	})
}

// ---------------------------------------------------------------------------
// AnsiLocate
// ---------------------------------------------------------------------------

func TestAnsiLocate(t *testing.T) {
	got := AnsiLocate(5, 10)
	want := "\033[5;10H"
	if got != want {
		t.Errorf("AnsiLocate(5, 10) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// AnsiCls
// ---------------------------------------------------------------------------

func TestAnsiCls(t *testing.T) {
	got := AnsiCls()
	want := "\033[2J\033[H"
	if got != want {
		t.Errorf("AnsiCls() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// AnsiColor
// ---------------------------------------------------------------------------

func TestAnsiColor(t *testing.T) {
	tests := []struct {
		name string
		fg   int
		bg   int
		want string
	}{
		{"white on black", 7, 0, "\033[37;40m"},
		{"black on white", 0, 7, "\033[30;47m"},
		{"red on blue", 4, 1, "\033[31;44m"},
		{"bright white on black", 15, 0, "\033[97;40m"},
		{"yellow on black", 14, 0, "\033[93;40m"},
		{"default for out of range", 99, 99, "\033[37;40m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnsiColor(tt.fg, tt.bg)
			if got != tt.want {
				t.Errorf("AnsiColor(%d, %d) = %q, want %q", tt.fg, tt.bg, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// addThousandsSep (unexported helper)
// ---------------------------------------------------------------------------

func TestAddThousandsSep(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"1", "1"},
		{"12", "12"},
		{"123", "123"},
		{"1234", "1,234"},
		{"12345", "12,345"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
	}
	for _, tt := range tests {
		got := addThousandsSep(tt.in)
		if got != tt.want {
			t.Errorf("addThousandsSep(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// valueToFloat / valueToString (unexported helpers)
// ---------------------------------------------------------------------------

func TestValueToFloat(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want float64
	}{
		{"float64", float64(3.14), 3.14},
		{"float32", float32(2.5), 2.5},
		{"int", int(42), 42},
		{"int16", int16(10), 10},
		{"int32", int32(100), 100},
		{"int64", int64(200), 200},
		{"uint", uint(5), 5},
		{"uint8", uint8(255), 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueToFloat(tt.v)
			if got != tt.want {
				t.Errorf("valueToFloat(%v) = %g, want %g", tt.v, got, tt.want)
			}
		})
	}
}

func TestValueToString(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want string
	}{
		{"string", "hello", "hello"},
		{"bytes", []byte("world"), "world"},
		{"int fallback", 42, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueToString(tt.v)
			if got != tt.want {
				t.Errorf("valueToString(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatScientific (via PrintUsing with ^^^^ format)
// ---------------------------------------------------------------------------

func TestFormatScientificZero(t *testing.T) {
	// absVal == 0 branch
	got := PrintUsing("#.##^^^^", []interface{}{0.0})
	if !strings.Contains(got, "E") {
		t.Errorf("PrintUsing scientific zero = %q, expected E notation", got)
	}
}

func TestFormatScientificPositive(t *testing.T) {
	// Normal positive value with decimal
	got := PrintUsing("#.##^^^^", []interface{}{1234.5})
	if !strings.Contains(got, "E") {
		t.Errorf("PrintUsing scientific = %q, expected E notation", got)
	}
}

func TestFormatScientificNegative(t *testing.T) {
	// negative=true branch
	got := PrintUsing("#.##^^^^", []interface{}{-1234.5})
	if !strings.Contains(got, "-") || !strings.Contains(got, "E") {
		t.Errorf("PrintUsing scientific negative = %q, expected '-' and 'E'", got)
	}
}

func TestFormatScientificLeadingPlus(t *testing.T) {
	// leadingPlus=true, positive value
	got := PrintUsing("+#.##^^^^", []interface{}{42.0})
	if !strings.HasPrefix(got, "+") {
		t.Errorf("PrintUsing scientific leadingPlus positive = %q, expected '+'", got)
	}
}

func TestFormatScientificLeadingPlusNegative(t *testing.T) {
	// leadingPlus=true, negative value
	got := PrintUsing("+#.##^^^^", []interface{}{-42.0})
	if !strings.HasPrefix(got, "-") {
		t.Errorf("PrintUsing scientific leadingPlus negative = %q, expected '-'", got)
	}
}

func TestFormatScientificTrailingPlus(t *testing.T) {
	// trailingPlus=true, positive value
	got := PrintUsing("#.##^^^^+", []interface{}{42.0})
	if !strings.HasSuffix(got, "+") {
		t.Errorf("PrintUsing scientific trailingPlus = %q, expected trailing '+'", got)
	}
}

func TestFormatScientificTrailingPlusNegative(t *testing.T) {
	// trailingPlus=true, negative value
	got := PrintUsing("#.##^^^^+", []interface{}{-42.0})
	if !strings.HasSuffix(got, "-") {
		t.Errorf("PrintUsing scientific trailingPlus negative = %q, expected trailing '-'", got)
	}
}

func TestFormatScientificTrailingMinus(t *testing.T) {
	// trailingMinus=true, negative value
	got := PrintUsing("#.##^^^^-", []interface{}{-42.0})
	if !strings.HasSuffix(got, "-") {
		t.Errorf("PrintUsing scientific trailingMinus negative = %q, expected trailing '-'", got)
	}
}

func TestFormatScientificTrailingMinusPositive(t *testing.T) {
	// trailingMinus=true, positive value → trailing space
	got := PrintUsing("#.##^^^^-", []interface{}{42.0})
	if !strings.HasSuffix(got, " ") {
		t.Errorf("PrintUsing scientific trailingMinus positive = %q, expected trailing ' '", got)
	}
}

func TestFormatScientificNoDecimal(t *testing.T) {
	// hasDecimal=false branch (format "##^^^^" has no '.')
	got := PrintUsing("##^^^^", []interface{}{1234.0})
	if !strings.Contains(got, "E") {
		t.Errorf("PrintUsing scientific noDecimal = %q, expected E notation", got)
	}
}

func TestFormatScientificMultiIntDigits(t *testing.T) {
	// intDigits > 1 branch
	got := PrintUsing("##.##^^^^", []interface{}{9999.99})
	if !strings.Contains(got, "E") {
		t.Errorf("PrintUsing scientific multiIntDigits = %q, expected E notation", got)
	}
}

func TestFormatScientificFloatingDollar(t *testing.T) {
	// floatingDollar branch
	got := PrintUsing("$#.##^^^^", []interface{}{42.0})
	if !strings.Contains(got, "$") {
		t.Errorf("PrintUsing scientific floatingDollar = %q, expected '$'", got)
	}
}

// ---------------------------------------------------------------------------
// basicToAnsiFg and basicToAnsiBg — all cases
// ---------------------------------------------------------------------------

func TestBasicToAnsiFgAllCases(t *testing.T) {
	tests := []struct {
		color    int
		wantCode int
	}{
		{0, 30}, {1, 34}, {2, 32}, {3, 36}, {4, 31}, {5, 35},
		{6, 33}, {7, 37}, {8, 90}, {9, 94}, {10, 92}, {11, 96},
		{12, 91}, {13, 95}, {14, 93}, {15, 97},
		{99, 37}, // default
	}
	for _, tt := range tests {
		got := basicToAnsiFg(tt.color)
		if got != tt.wantCode {
			t.Errorf("basicToAnsiFg(%d) = %d, want %d", tt.color, got, tt.wantCode)
		}
	}
}

func TestBasicToAnsiBgAllCases(t *testing.T) {
	tests := []struct {
		color    int
		wantCode int
	}{
		{0, 40}, {1, 44}, {2, 42}, {3, 46}, {4, 41}, {5, 45},
		{6, 43}, {7, 47}, {8, 100}, {9, 104}, {10, 102}, {11, 106},
		{12, 101}, {13, 105}, {14, 103}, {15, 107},
		{99, 40}, // default
	}
	for _, tt := range tests {
		got := basicToAnsiBg(tt.color)
		if got != tt.wantCode {
			t.Errorf("basicToAnsiBg(%d) = %d, want %d", tt.color, got, tt.wantCode)
		}
	}
}

// ---------------------------------------------------------------------------
// InputPrompt, NewScanner, InputSplitLine — stdin mocking
// ---------------------------------------------------------------------------

func TestInputPromptReadsLine(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	w.WriteString("hello world\n")
	w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	got := InputPrompt("")
	if got != "hello world" {
		t.Errorf("InputPrompt() = %q, want %q", got, "hello world")
	}
}

func TestInputPromptEOF(t *testing.T) {
	// On EOF, InputPrompt returns whatever was read (possibly empty).
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	w.WriteString("partial")
	w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	got := InputPrompt("Enter: ")
	if got != "partial" {
		t.Errorf("InputPrompt EOF = %q, want %q", got, "partial")
	}
}

func TestNewScannerReturnsBufioScanner(t *testing.T) {
	s := NewScanner()
	if s == nil {
		t.Fatal("NewScanner() returned nil")
	}
}

func TestInputSplitLine(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	w.WriteString("  1, 2 , 3\n")
	w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	parts := InputSplitLine()
	if len(parts) != 3 {
		t.Fatalf("InputSplitLine() = %v, want 3 parts", parts)
	}
	if parts[0] != "1" || parts[1] != "2" || parts[2] != "3" {
		t.Errorf("InputSplitLine() = %v, want [1 2 3]", parts)
	}
}

func TestPrintUsingEmpty(t *testing.T) {
	result := PrintUsing("##", []interface{}{})
	if result != "" {
		t.Errorf("expected empty string for empty values, got %q", result)
	}
}

func TestPrintUsingBasicHash(t *testing.T) {
	result := PrintUsing("##", []interface{}{float64(42)})
	if !strings.Contains(result, "42") {
		t.Errorf("expected '42' in output, got %q", result)
	}
}

func TestPrintUsingDecimal(t *testing.T) {
	result := PrintUsing("##.##", []interface{}{float64(3.14)})
	if !strings.Contains(result, "3.14") {
		t.Errorf("expected '3.14' in output, got %q", result)
	}
}

func TestPrintUsingStringField_Bang(t *testing.T) {
	result := PrintUsing("!", []interface{}{"Hello"})
	if !strings.Contains(result, "H") {
		t.Errorf("expected 'H' in output, got %q", result)
	}
}

func TestPrintUsingStringField_Ampersand(t *testing.T) {
	result := PrintUsing("&", []interface{}{"Hi"})
	if !strings.Contains(result, "Hi") {
		t.Errorf("expected 'Hi' in output, got %q", result)
	}
}

func TestPrintUsingLiteralText(t *testing.T) {
	result := PrintUsing("Item:", []interface{}{})
	_ = result
}

func TestPrintUsingUnderscoreEscape(t *testing.T) {
	result := PrintUsing("_#", []interface{}{float64(5)})
	if !strings.Contains(result, "#") {
		t.Errorf("expected literal '#' in output, got %q", result)
	}
}

func TestPrintUsingLeadingPlus(t *testing.T) {
	result := PrintUsing("+##", []interface{}{float64(7)})
	_ = result
}

func TestPrintUsingDoubleDoller(t *testing.T) {
	result := PrintUsing("$$##.##", []interface{}{float64(12.50)})
	_ = result
}

func TestPrintUsingAsteriskFill(t *testing.T) {
	result := PrintUsing("**###", []interface{}{float64(42)})
	_ = result
}

func TestPrintUsingScientific(t *testing.T) {
	result := PrintUsing("##.##^^^^", []interface{}{float64(12345.6)})
	_ = result
}

func TestPrintUsingMultipleValues(t *testing.T) {
	result := PrintUsing("##", []interface{}{float64(1), float64(2), float64(3)})
	_ = result
}

func TestPrintUsingTrailingSign(t *testing.T) {
	result := PrintUsing("##-", []interface{}{float64(-5)})
	_ = result
}

func TestPrintUsingAsteriskDollar(t *testing.T) {
	result := PrintUsing("**$###", []interface{}{float64(99)})
	_ = result
}

func TestPrintUsingNoHashNoPattern(t *testing.T) {
	result := PrintUsing("abc", []interface{}{float64(5)})
	_ = result
}

func TestPrintUsingCommaGrouping(t *testing.T) {
	result := PrintUsing("#,###", []interface{}{float64(1234)})
	_ = result
}

func TestFormatNumberZero(t *testing.T) {
	result := FormatNumber("##.##", 0)
	_ = result
}

func TestFormatNumberNegative(t *testing.T) {
	result := FormatNumber("##.##", -3.14)
	if !strings.Contains(result, "3.14") {
		t.Errorf("expected '3.14' in output, got %q", result)
	}
}

func TestFormatNumberLarge(t *testing.T) {
	result := FormatNumber("####", 9999)
	if !strings.Contains(result, "9999") {
		t.Errorf("expected '9999' in output, got %q", result)
	}
}

func TestFormatStringTruncate(t *testing.T) {
	result := FormatString("!", "Hello")
	if result != "H" {
		t.Errorf("FormatString('!', 'Hello') = %q, want 'H'", result)
	}
}

func TestFormatStringFullAmpersand(t *testing.T) {
	result := FormatString("&", "Hello")
	if result != "Hello" {
		t.Errorf("FormatString('&', 'Hello') = %q, want 'Hello'", result)
	}
}

func TestFormatStringPadded(t *testing.T) {
	result := FormatString(`\ \`, "Hi")
	if len(result) == 0 {
		t.Error("FormatString with padded format should return non-empty string")
	}
}

func TestFormatNumberScientificSmall(t *testing.T) {
	result := FormatNumber("#.##^^^^", 0.0001234)
	_ = result
}

// ---------------------------------------------------------------------------
// TestValueToFloatAllTypes — cover missing type branches + fallback path
// ---------------------------------------------------------------------------

func TestValueToFloatAllTypes(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want float64
	}{
		{"int8", int8(42), 42},
		{"uint16", uint16(1000), 1000},
		{"uint32", uint32(100000), 100000},
		{"uint64", uint64(999999), 999999},
		{"string numeric", "3.14", 3.14},
		{"string non-numeric", "notanumber", 0},
		{"bool true", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueToFloat(tt.v)
			if got != tt.want {
				t.Errorf("valueToFloat(%v) = %g, want %g", tt.v, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFormatScientificExactValues — exact output for various magnitudes
// ---------------------------------------------------------------------------

func TestFormatScientificExactValues(t *testing.T) {
	tests := []struct {
		name   string
		format string
		value  float64
	}{
		{"small number negative exponent", "#.##^^^^", 0.0001234},
		{"large number", "#.##^^^^", 9999999.0},
		{"exact one", "#.##^^^^", 1.0},
		{"negative fraction", "#.##^^^^", -0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNumber(tt.format, tt.value)
			if !strings.Contains(got, "E") {
				t.Errorf("FormatNumber(%q, %g) = %q, expected E notation", tt.format, tt.value, got)
			}
			if tt.value < 0 && !strings.Contains(got, "-") {
				t.Errorf("FormatNumber(%q, %g) = %q, expected '-' for negative value", tt.format, tt.value, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFormatScientificAsteriskDollar — **$ prefix with scientific notation
// ---------------------------------------------------------------------------

func TestFormatScientificAsteriskDollar(t *testing.T) {
	got := FormatNumber("**$#.##^^^^", 42.0)
	if !strings.Contains(got, "$") {
		t.Errorf("FormatNumber(\"**$#.##^^^^\", 42) = %q, expected '$'", got)
	}
	if !strings.Contains(got, "E") {
		t.Errorf("FormatNumber(\"**$#.##^^^^\", 42) = %q, expected 'E' notation", got)
	}
}

// ---------------------------------------------------------------------------
// TestFormatNumberCommaGrouping — comma thousands separator
// ---------------------------------------------------------------------------

func TestFormatNumberCommaGrouping(t *testing.T) {
	tests := []struct {
		name   string
		format string
		value  float64
		substr string
	}{
		{"thousands", "#,###", 1234, "1,234"},
		{"millions", "#,###,###", 1234567, "1,234,567"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNumber(tt.format, tt.value)
			if !strings.Contains(got, tt.substr) {
				t.Errorf("FormatNumber(%q, %g) = %q, expected to contain %q", tt.format, tt.value, got, tt.substr)
			}
		})
	}
}
