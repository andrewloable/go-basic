package runtime

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Left
// ---------------------------------------------------------------------------

func TestLeft(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"normal", "HELLO", 3, "HEL"},
		{"zero", "HELLO", 0, ""},
		{"negative", "HELLO", -1, ""},
		{"exact length", "HELLO", 5, "HELLO"},
		{"exceeds length", "HELLO", 10, "HELLO"},
		{"empty string", "", 5, ""},
		{"one char", "A", 1, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Left(tt.s, tt.n)
			if got != tt.want {
				t.Errorf("Left(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Right
// ---------------------------------------------------------------------------

func TestRight(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"normal", "HELLO", 3, "LLO"},
		{"zero", "HELLO", 0, ""},
		{"negative", "HELLO", -1, ""},
		{"exact length", "HELLO", 5, "HELLO"},
		{"exceeds length", "HELLO", 10, "HELLO"},
		{"empty string", "", 5, ""},
		{"one char", "A", 1, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Right(tt.s, tt.n)
			if got != tt.want {
				t.Errorf("Right(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Mid
// ---------------------------------------------------------------------------

func TestMid(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		start  int
		length int
		want   string
	}{
		{"normal", "HELLO WORLD", 7, 5, "WORLD"},
		{"start 1", "HELLO", 1, 3, "HEL"},
		{"start 0 clamped to 1", "HELLO", 0, 3, "HEL"},
		{"negative start clamped", "HELLO", -5, 3, "HEL"},
		{"start past end", "HELLO", 10, 3, ""},
		{"length -1 means rest", "HELLO", 3, -1, "LLO"},
		{"length exceeds", "HELLO", 3, 100, "LLO"},
		{"exact", "HELLO", 1, 5, "HELLO"},
		{"empty string", "", 1, 5, ""},
		{"single char", "A", 1, 1, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Mid(tt.s, tt.start, tt.length)
			if got != tt.want {
				t.Errorf("Mid(%q, %d, %d) = %q, want %q", tt.s, tt.start, tt.length, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Instr
// ---------------------------------------------------------------------------

func TestInstr(t *testing.T) {
	tests := []struct {
		name  string
		start int
		s     string
		find  string
		want  int
	}{
		{"found at start", 1, "HELLO", "HE", 1},
		{"found in middle", 1, "HELLO", "LL", 3},
		{"not found", 1, "HELLO", "XY", 0},
		{"start 0 defaults to 1", 0, "HELLO", "HE", 1},
		{"start past end", 10, "HELLO", "HE", 0},
		{"start in middle", 3, "ABCABC", "ABC", 4},
		{"find empty string", 1, "HELLO", "", 1},
		{"empty source", 1, "", "A", 0},
		{"negative start defaults", -1, "HELLO", "HE", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Instr(tt.start, tt.s, tt.find)
			if got != tt.want {
				t.Errorf("Instr(%d, %q, %q) = %d, want %d", tt.start, tt.s, tt.find, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Len
// ---------------------------------------------------------------------------

func TestLen(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"HELLO", 5},
		{"", 0},
		{"A", 1},
		{"Hello World", 11},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := Len(tt.s)
			if got != tt.want {
				t.Errorf("Len(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Asc
// ---------------------------------------------------------------------------

func TestAsc(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got, err := Asc("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 65 {
			t.Errorf("Asc(\"A\") = %d, want 65", got)
		}
	})

	t.Run("space", func(t *testing.T) {
		got, err := Asc(" ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 32 {
			t.Errorf("Asc(\" \") = %d, want 32", got)
		}
	})

	t.Run("multi-char uses first", func(t *testing.T) {
		got, err := Asc("Hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 72 {
			t.Errorf("Asc(\"Hello\") = %d, want 72", got)
		}
	})

	t.Run("empty string error", func(t *testing.T) {
		_, err := Asc("")
		if err == nil {
			t.Fatalf("expected error for empty string, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// Chr
// ---------------------------------------------------------------------------

func TestChr(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got, err := Chr(65)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "A" {
			t.Errorf("Chr(65) = %q, want \"A\"", got)
		}
	})

	t.Run("zero", func(t *testing.T) {
		got, err := Chr(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "\x00" {
			t.Errorf("Chr(0) = %q, want \"\\x00\"", got)
		}
	})

	t.Run("255", func(t *testing.T) {
		got, err := Chr(255)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := string(byte(255))
		if got != want {
			t.Errorf("Chr(255) = %q, want %q", got, want)
		}
	})

	t.Run("negative error", func(t *testing.T) {
		_, err := Chr(-1)
		if err == nil {
			t.Fatalf("expected error for -1, got nil")
		}
	})

	t.Run("over 255 error", func(t *testing.T) {
		_, err := Chr(256)
		if err == nil {
			t.Fatalf("expected error for 256, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// Str
// ---------------------------------------------------------------------------

func TestStr(t *testing.T) {
	tests := []struct {
		name string
		n    float64
		want string
	}{
		{"positive int", 42, " 42"},
		{"negative int", -42, "-42"},
		{"zero", 0, " 0"},
		{"positive float", 3.14, " 3.14"},
		{"negative float", -3.14, "-3.14"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Str(tt.n)
			if got != tt.want {
				t.Errorf("Str(%g) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Val
// ---------------------------------------------------------------------------

func TestVal(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want float64
	}{
		{"integer", "42", 42},
		{"negative", "-42", -42},
		{"float", "3.14", 3.14},
		{"leading whitespace", "  42", 42},
		{"trailing chars", "42abc", 42},
		{"scientific", "1.5E2", 150},
		{"scientific neg exp", "1.5E-2", 0.015},
		{"empty", "", 0},
		{"only spaces", "   ", 0},
		{"no digits", "abc", 0},
		{"plus sign", "+42", 42},
		{"decimal only", ".5", 0.5},
		{"E no digits after", "42E", 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Val(tt.s)
			if math.Abs(got-tt.want) > 1e-12 {
				t.Errorf("Val(%q) = %g, want %g", tt.s, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Hex, Oct, Bin
// ---------------------------------------------------------------------------

func TestHex(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "0"},
		{"positive", 255, "FF"},
		{"large", 65535, "FFFF"},
		{"negative treats as uint16", -1, "FFFF"},
		{"negative small", -256, "FF00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Hex(tt.n)
			if got != tt.want {
				t.Errorf("Hex(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestOct(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "0"},
		{"positive", 8, "10"},
		{"255", 255, "377"},
		{"negative treats as uint16", -1, "177777"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Oct(tt.n)
			if got != tt.want {
				t.Errorf("Oct(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestBin(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"five", 5, "101"},
		{"255", 255, "11111111"},
		{"negative treats as uint16", -1, "1111111111111111"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Bin(tt.n)
			if got != tt.want {
				t.Errorf("Bin(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UCase, LCase
// ---------------------------------------------------------------------------

func TestUCase(t *testing.T) {
	tests := []struct {
		s, want string
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"ALREADY", "ALREADY"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := UCase(tt.s); got != tt.want {
				t.Errorf("UCase(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestLCase(t *testing.T) {
	tests := []struct {
		s, want string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"already", "already"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := LCase(tt.s); got != tt.want {
				t.Errorf("LCase(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// LTrim, RTrim, Trim
// ---------------------------------------------------------------------------

func TestLTrim(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"leading spaces", "  hello", "hello"},
		{"no spaces", "hello", "hello"},
		{"trailing spaces kept", "hello  ", "hello  "},
		{"both sides", "  hello  ", "hello  "},
		{"empty", "", ""},
		{"only spaces", "   ", ""},
		{"tabs not trimmed", "\thello", "\thello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LTrim(tt.s); got != tt.want {
				t.Errorf("LTrim(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestRTrim(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"trailing spaces", "hello  ", "hello"},
		{"no spaces", "hello", "hello"},
		{"leading spaces kept", "  hello", "  hello"},
		{"both sides", "  hello  ", "  hello"},
		{"empty", "", ""},
		{"only spaces", "   ", ""},
		{"tabs not trimmed", "hello\t", "hello\t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RTrim(tt.s); got != tt.want {
				t.Errorf("RTrim(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"both sides", "  hello  ", "hello"},
		{"no spaces", "hello", "hello"},
		{"empty", "", ""},
		{"only spaces", "   ", ""},
		{"tabs not trimmed", "\thello\t", "\thello\t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Trim(tt.s); got != tt.want {
				t.Errorf("Trim(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Space
// ---------------------------------------------------------------------------

func TestSpace(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, ""},
		{"negative", -1, ""},
		{"one", 1, " "},
		{"five", 5, "     "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Space(tt.n); got != tt.want {
				t.Errorf("Space(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StringRepeat
// ---------------------------------------------------------------------------

func TestStringRepeat(t *testing.T) {
	tests := []struct {
		name string
		n    int
		char byte
		want string
	}{
		{"normal", 3, 'A', "AAA"},
		{"zero", 0, 'A', ""},
		{"negative", -1, 'A', ""},
		{"one", 1, '*', "*"},
		{"five stars", 5, '*', "*****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringRepeat(tt.n, tt.char); got != tt.want {
				t.Errorf("StringRepeat(%d, %c) = %q, want %q", tt.n, tt.char, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Mki / Cvi round-trip
// ---------------------------------------------------------------------------

func TestMkiCvi(t *testing.T) {
	tests := []int16{0, 1, -1, 32767, -32768, 256, -256, 12345}
	for _, v := range tests {
		s := Mki(v)
		if len(s) != 2 {
			t.Errorf("Mki(%d) produced %d bytes, want 2", v, len(s))
		}
		got, err := Cvi(s)
		if err != nil {
			t.Fatalf("Cvi(Mki(%d)) error: %v", v, err)
		}
		if got != v {
			t.Errorf("Cvi(Mki(%d)) = %d, want %d", v, got, v)
		}
	}
}

func TestCviTooShort(t *testing.T) {
	_, err := Cvi("x")
	if err == nil {
		t.Fatal("expected error for short string, got nil")
	}
}

// ---------------------------------------------------------------------------
// Mkl / Cvl round-trip
// ---------------------------------------------------------------------------

func TestMklCvl(t *testing.T) {
	tests := []int32{0, 1, -1, 2147483647, -2147483648, 100000}
	for _, v := range tests {
		s := Mkl(v)
		if len(s) != 4 {
			t.Errorf("Mkl(%d) produced %d bytes, want 4", v, len(s))
		}
		got, err := Cvl(s)
		if err != nil {
			t.Fatalf("Cvl(Mkl(%d)) error: %v", v, err)
		}
		if got != v {
			t.Errorf("Cvl(Mkl(%d)) = %d, want %d", v, got, v)
		}
	}
}

func TestCvlTooShort(t *testing.T) {
	_, err := Cvl("ab")
	if err == nil {
		t.Fatal("expected error for short string, got nil")
	}
}

// ---------------------------------------------------------------------------
// Mks / Cvs round-trip
// ---------------------------------------------------------------------------

func TestMksCvs(t *testing.T) {
	tests := []float32{0, 1, -1, 3.14, -273.15, 1e10}
	for _, v := range tests {
		s := Mks(v)
		if len(s) != 4 {
			t.Errorf("Mks(%g) produced %d bytes, want 4", v, len(s))
		}
		got, err := Cvs(s)
		if err != nil {
			t.Fatalf("Cvs(Mks(%g)) error: %v", v, err)
		}
		if got != v {
			t.Errorf("Cvs(Mks(%g)) = %g, want %g", v, got, v)
		}
	}
}

func TestCvsTooShort(t *testing.T) {
	_, err := Cvs("ab")
	if err == nil {
		t.Fatal("expected error for short string, got nil")
	}
}

// ---------------------------------------------------------------------------
// Mkd / Cvd round-trip
// ---------------------------------------------------------------------------

func TestMkdCvd(t *testing.T) {
	tests := []float64{0, 1, -1, 3.141592653589793, -273.15, 1e100, math.SmallestNonzeroFloat64}
	for _, v := range tests {
		s := Mkd(v)
		if len(s) != 8 {
			t.Errorf("Mkd(%g) produced %d bytes, want 8", v, len(s))
		}
		got, err := Cvd(s)
		if err != nil {
			t.Fatalf("Cvd(Mkd(%g)) error: %v", v, err)
		}
		if got != v {
			t.Errorf("Cvd(Mkd(%g)) = %g, want %g", v, got, v)
		}
	}
}

func TestCvdTooShort(t *testing.T) {
	_, err := Cvd("abcd")
	if err == nil {
		t.Fatal("expected error for short string, got nil")
	}
}
