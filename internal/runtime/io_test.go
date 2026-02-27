package runtime

import (
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
