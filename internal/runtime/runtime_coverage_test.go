package runtime

import (
	"strings"
	"testing"
)

// ===========================================================================
// Graphics stubs (all at 0% coverage)
// ===========================================================================

func TestScreenMode(t *testing.T) {
	cases := []int{0, 1, 2, 7, 8, 9, 10, 11, 12, 13, 99}
	for _, mode := range cases {
		ScreenMode(mode)
		// Just verify it doesn't panic
	}
	// Verify mode 0 resets to defaults
	ScreenMode(0)
	if CurrentScreen.Width != 80 {
		t.Errorf("ScreenMode(0) width = %d, want 80", CurrentScreen.Width)
	}
}

func TestCircleStub(t *testing.T) {
	Circle(100, 100, 50, 15, 0, 0, 1) // should not panic
}

func TestDrawLineStub(t *testing.T) {
	DrawLine(0, 0, 100, 100, 15, "B") // should not panic
}

func TestPsetStub(t *testing.T) {
	Pset(50, 50, 7) // should not panic
}

func TestPaintStub(t *testing.T) {
	Paint(50, 50, 4, 15) // should not panic
}

func TestDrawStub(t *testing.T) {
	Draw("U10 R10 D10 L10") // should not panic
}

func TestViewPortStub(t *testing.T) {
	ViewPort(0, 0, 640, 480, 0, 0) // should not panic
}

func TestViewPrintStub(t *testing.T) {
	ViewPrint(1, 24) // should not panic
}

// ===========================================================================
// Audio stubs (all at 0% coverage)
// ===========================================================================

func TestPlayStub(t *testing.T) {
	Play("T120 O4 L4 CDEFGAB") // should not panic
}

func TestSoundStub(t *testing.T) {
	Sound(440.0, 18.2) // should not panic
}

// ===========================================================================
// PrintUsing coverage (at 70.9%)
// ===========================================================================

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
	// "!" truncates to first character
	result := PrintUsing("!", []interface{}{"Hello"})
	if !strings.Contains(result, "H") {
		t.Errorf("expected 'H' in output, got %q", result)
	}
}

func TestPrintUsingStringField_Ampersand(t *testing.T) {
	// "&" outputs full string
	result := PrintUsing("&", []interface{}{"Hi"})
	if !strings.Contains(result, "Hi") {
		t.Errorf("expected 'Hi' in output, got %q", result)
	}
}

func TestPrintUsingLiteralText(t *testing.T) {
	// Literal text without field passes through
	result := PrintUsing("Item:", []interface{}{})
	_ = result // no crash
}

func TestPrintUsingUnderscoreEscape(t *testing.T) {
	// _ followed by char escapes that char as literal
	result := PrintUsing("_#", []interface{}{float64(5)})
	if !strings.Contains(result, "#") {
		t.Errorf("expected literal '#' in output, got %q", result)
	}
}

func TestPrintUsingLeadingPlus(t *testing.T) {
	// + before field shows sign
	result := PrintUsing("+##", []interface{}{float64(7)})
	_ = result // verify no crash
}

func TestPrintUsingDoubleDoller(t *testing.T) {
	// $$ floating dollar sign
	result := PrintUsing("$$##.##", []interface{}{float64(12.50)})
	_ = result // verify no crash
}

func TestPrintUsingAsteriskFill(t *testing.T) {
	// ** asterisk fill
	result := PrintUsing("**###", []interface{}{float64(42)})
	_ = result // verify no crash
}

func TestPrintUsingScientific(t *testing.T) {
	// ^^^^ scientific notation
	result := PrintUsing("##.##^^^^", []interface{}{float64(12345.6)})
	_ = result // verify no crash
}

func TestPrintUsingMultipleValues(t *testing.T) {
	// Multiple values cycle through format
	result := PrintUsing("##", []interface{}{float64(1), float64(2), float64(3)})
	_ = result // verify no crash
}

func TestPrintUsingTrailingSign(t *testing.T) {
	// Trailing - sign for negative
	result := PrintUsing("##-", []interface{}{float64(-5)})
	_ = result // verify no crash
}

// ===========================================================================
// parseNumericField paths (internal, tested via PrintUsing)
// ===========================================================================

func TestPrintUsingAsteriskDollar(t *testing.T) {
	// **$ asterisk fill with floating dollar
	result := PrintUsing("**$###", []interface{}{float64(99)})
	_ = result
}

func TestPrintUsingNoHashNoPattern(t *testing.T) {
	// Format string with no valid numeric pattern - literal text only
	result := PrintUsing("abc", []interface{}{float64(5)})
	_ = result
}

// ===========================================================================
// io.go: isNumericField (tested indirectly via PrintUsing edge cases)
// ===========================================================================

func TestPrintUsingCommaGrouping(t *testing.T) {
	// Comma in format for grouping thousands
	result := PrintUsing("#,###", []interface{}{float64(1234)})
	_ = result
}

// ===========================================================================
// io.go: FormatNumber edge cases
// ===========================================================================

func TestFormatNumberZero(t *testing.T) {
	// FormatNumber is exported and we can call it directly
	result := FormatNumber("##.##", 0)
	_ = result // just verify no panic
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

// ===========================================================================
// io.go: FormatString
// ===========================================================================

func TestFormatStringTruncate(t *testing.T) {
	// "!" format - first char only
	result := FormatString("!", "Hello")
	if result != "H" {
		t.Errorf("FormatString('!', 'Hello') = %q, want 'H'", result)
	}
}

func TestFormatStringFullAmpersand(t *testing.T) {
	// "&" format - full string
	result := FormatString("&", "Hello")
	if result != "Hello" {
		t.Errorf("FormatString('&', 'Hello') = %q, want 'Hello'", result)
	}
}

func TestFormatStringPadded(t *testing.T) {
	// "\\...\\" format - padded to fixed width
	result := FormatString(`\ \`, "Hi")
	if len(result) == 0 {
		t.Error("FormatString with padded format should return non-empty string")
	}
}

// ===========================================================================
// io.go: formatScientific (tested via FormatNumber with scientific notation)
// ===========================================================================

func TestFormatNumberScientificSmall(t *testing.T) {
	result := FormatNumber("#.##^^^^", 0.0001234)
	_ = result
}

// ===========================================================================
// fileio: Lset / Rset with wrong mode
// ===========================================================================

func TestLsetNoFile(t *testing.T) {
	fm := NewFileManager()
	// Lset on file not open
	err := fm.Lset(1, "field", "value")
	if err == nil {
		t.Error("expected error for Lset on closed file")
	}
}

func TestRsetNoFile(t *testing.T) {
	fm := NewFileManager()
	err := fm.Rset(1, "field", "value")
	if err == nil {
		t.Error("expected error for Rset on closed file")
	}
}

// ===========================================================================
// fileio: getWriter error path (output file not opened)
// ===========================================================================

func TestFileWriteNotOpen(t *testing.T) {
	fm := NewFileManager()
	err := fm.FileWrite(99, []interface{}{"test"})
	if err == nil {
		t.Error("expected error writing to unopened file")
	}
}

func TestFilePrintNotOpen(t *testing.T) {
	fm := NewFileManager()
	err := fm.FilePrint(99, " hello")
	if err == nil {
		t.Error("expected error printing to unopened file")
	}
}

// ===========================================================================
// fileio: Eof, Loc, Lof error paths
// ===========================================================================

func TestEofNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Eof(99)
	if err == nil {
		t.Error("expected error for Eof on closed file")
	}
}

func TestLocNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Loc(99)
	if err == nil {
		t.Error("expected error for Loc on closed file")
	}
}

func TestLofNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Lof(99)
	if err == nil {
		t.Error("expected error for Lof on closed file")
	}
}
