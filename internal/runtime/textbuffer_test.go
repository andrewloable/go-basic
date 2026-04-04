package runtime

import "testing"

func TestNewTextBuffer(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	if tb.Cols != 80 {
		t.Errorf("Cols = %d, want 80", tb.Cols)
	}
	if tb.Rows != 25 {
		t.Errorf("Rows = %d, want 25", tb.Rows)
	}
	if tb.CurFg != 7 {
		t.Errorf("CurFg = %d, want 7", tb.CurFg)
	}
	if tb.CurBg != 0 {
		t.Errorf("CurBg = %d, want 0", tb.CurBg)
	}
	if tb.CursorR != 0 {
		t.Errorf("CursorR = %d, want 0", tb.CursorR)
	}
	if tb.CursorC != 0 {
		t.Errorf("CursorC = %d, want 0", tb.CursorC)
	}
	for r := 0; r < tb.Rows; r++ {
		for c := 0; c < tb.Cols; c++ {
			cell := tb.Cells[r][c]
			if cell.Char != ' ' || cell.Fg != 7 || cell.Bg != 0 {
				t.Fatalf("Cells[%d][%d] = {%q, %d, %d}, want {' ', 7, 0}", r, c, cell.Char, cell.Fg, cell.Bg)
			}
		}
	}
}

func TestTextBufferClear(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tb.PutString("Hello")
	tb.SetColor(4, 1)
	tb.Clear()

	if tb.CursorR != 0 || tb.CursorC != 0 {
		t.Errorf("cursor after Clear = (%d, %d), want (0, 0)", tb.CursorR, tb.CursorC)
	}
	for r := 0; r < tb.Rows; r++ {
		for c := 0; c < tb.Cols; c++ {
			cell := tb.Cells[r][c]
			if cell.Char != ' ' || cell.Fg != 4 || cell.Bg != 1 {
				t.Fatalf("Cells[%d][%d] = {%q, %d, %d}, want {' ', 4, 1}", r, c, cell.Char, cell.Fg, cell.Bg)
			}
		}
	}
}

func TestTextBufferSetCursor(t *testing.T) {
	tb := NewTextBuffer(80, 25)

	tb.SetCursor(-1, -1)
	if tb.CursorR != 0 || tb.CursorC != 0 {
		t.Errorf("SetCursor(-1,-1) = (%d, %d), want (0, 0)", tb.CursorR, tb.CursorC)
	}

	tb.SetCursor(100, 100)
	if tb.CursorR != 24 || tb.CursorC != 79 {
		t.Errorf("SetCursor(100,100) = (%d, %d), want (24, 79)", tb.CursorR, tb.CursorC)
	}

	tb.SetCursor(5, 10)
	if tb.CursorR != 5 || tb.CursorC != 10 {
		t.Errorf("SetCursor(5,10) = (%d, %d), want (5, 10)", tb.CursorR, tb.CursorC)
	}
}

func TestTextBufferSetColor(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tb.SetColor(12, 3)
	if tb.CurFg != 12 {
		t.Errorf("CurFg = %d, want 12", tb.CurFg)
	}
	if tb.CurBg != 3 {
		t.Errorf("CurBg = %d, want 3", tb.CurBg)
	}
}

func TestTextBufferPutCharPrintable(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tb.PutChar('A')
	if tb.Cells[0][0].Char != 'A' {
		t.Errorf("Cells[0][0].Char = %q, want 'A'", tb.Cells[0][0].Char)
	}
	if tb.CursorR != 0 || tb.CursorC != 1 {
		t.Errorf("cursor = (%d, %d), want (0, 1)", tb.CursorR, tb.CursorC)
	}
}

func TestTextBufferPutCharNewline(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tb.PutChar('\n')
	if tb.CursorR != 1 || tb.CursorC != 0 {
		t.Errorf("cursor after newline = (%d, %d), want (1, 0)", tb.CursorR, tb.CursorC)
	}
}

func TestTextBufferPutCharCarriageReturn(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tb.SetCursor(3, 10)
	tb.PutChar('\r')
	if tb.CursorC != 0 {
		t.Errorf("CursorC after CR = %d, want 0", tb.CursorC)
	}
	if tb.CursorR != 3 {
		t.Errorf("CursorR after CR = %d, want 3 (unchanged)", tb.CursorR)
	}
}

func TestTextBufferPutCharTab(t *testing.T) {
	tb := NewTextBuffer(80, 25)

	// At col 0, tab should go to col 8.
	tb.PutChar('\t')
	if tb.CursorC != 8 {
		t.Errorf("CursorC after tab at col 0 = %d, want 8", tb.CursorC)
	}

	// At col 8, set cursor to col 5 and tab again.
	tb.SetCursor(0, 5)
	tb.PutChar('\t')
	if tb.CursorC != 8 {
		t.Errorf("CursorC after tab at col 5 = %d, want 8", tb.CursorC)
	}
}

func TestTextBufferPutCharBackspace(t *testing.T) {
	tb := NewTextBuffer(80, 25)

	// At col 5, backspace should go to col 4.
	tb.SetCursor(0, 5)
	tb.PutChar('\b')
	if tb.CursorC != 4 {
		t.Errorf("CursorC after backspace at col 5 = %d, want 4", tb.CursorC)
	}

	// At col 0, backspace should stay at 0.
	tb.SetCursor(0, 0)
	tb.PutChar('\b')
	if tb.CursorC != 0 {
		t.Errorf("CursorC after backspace at col 0 = %d, want 0", tb.CursorC)
	}
}

func TestTextBufferLineWrap(t *testing.T) {
	tb := NewTextBuffer(10, 3)
	for i := 0; i < 10; i++ {
		tb.PutChar('X')
	}
	if tb.CursorR != 1 || tb.CursorC != 0 {
		t.Errorf("cursor after 10 chars on 10-col buffer = (%d, %d), want (1, 0)", tb.CursorR, tb.CursorC)
	}
}

func TestTextBufferScroll(t *testing.T) {
	tb := NewTextBuffer(10, 3)
	// Write 'A' on row 0, 'B' on row 1.
	tb.PutChar('A')
	tb.PutChar('\n')
	tb.PutChar('B')
	tb.PutChar('\n')
	tb.PutChar('C')
	tb.PutChar('\n') // This should trigger a scroll.

	// After scroll, row 0 should have what was row 1 ('B'), row 1 should have 'C'.
	if tb.Cells[0][0].Char != 'B' {
		t.Errorf("Cells[0][0].Char after scroll = %q, want 'B'", tb.Cells[0][0].Char)
	}
	if tb.Cells[1][0].Char != 'C' {
		t.Errorf("Cells[1][0].Char after scroll = %q, want 'C'", tb.Cells[1][0].Char)
	}
	// Bottom row should be cleared.
	if tb.Cells[2][0].Char != ' ' {
		t.Errorf("Cells[2][0].Char after scroll = %q, want ' '", tb.Cells[2][0].Char)
	}
}

func TestTextBufferPutString(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tb.PutString("Hi")
	if tb.Cells[0][0].Char != 'H' {
		t.Errorf("Cells[0][0].Char = %q, want 'H'", tb.Cells[0][0].Char)
	}
	if tb.Cells[0][1].Char != 'i' {
		t.Errorf("Cells[0][1].Char = %q, want 'i'", tb.Cells[0][1].Char)
	}
	if tb.CursorR != 0 || tb.CursorC != 2 {
		t.Errorf("cursor = (%d, %d), want (0, 2)", tb.CursorR, tb.CursorC)
	}
}

func TestTextWriterPlainText(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tw := &textWriter{buf: tb}
	tw.Write([]byte("Hello"))
	expected := "Hello"
	for i, ch := range []byte(expected) {
		if tb.Cells[0][i].Char != ch {
			t.Errorf("Cells[0][%d].Char = %q, want %q", i, tb.Cells[0][i].Char, ch)
		}
	}
}

func TestTextWriterCSICursorPosition(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tw := &textWriter{buf: tb}
	tw.Write([]byte("\033[5;10H"))
	if tb.CursorR != 4 || tb.CursorC != 9 {
		t.Errorf("cursor after ESC[5;10H = (%d, %d), want (4, 9)", tb.CursorR, tb.CursorC)
	}
}

func TestTextWriterCSIClearScreen(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tw := &textWriter{buf: tb}
	tw.Write([]byte("ABCDE"))
	tw.Write([]byte("\033[2J"))
	// After clear, cells should be spaces.
	for c := 0; c < 5; c++ {
		if tb.Cells[0][c].Char != ' ' {
			t.Errorf("Cells[0][%d].Char after clear = %q, want ' '", c, tb.Cells[0][c].Char)
		}
	}
}

func TestTextWriterCSIColor(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tw := &textWriter{buf: tb}
	// ANSI 31 = red (BASIC 4), ANSI 44 = blue bg (BASIC 1).
	tw.Write([]byte("\033[31;44m"))
	if tb.CurFg != 4 {
		t.Errorf("CurFg after ESC[31;44m = %d, want 4 (red)", tb.CurFg)
	}
	if tb.CurBg != 1 {
		t.Errorf("CurBg after ESC[31;44m = %d, want 1 (blue)", tb.CurBg)
	}
}

func TestTextWriterCSICursorMovement(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tw := &textWriter{buf: tb}
	tb.SetCursor(5, 5)

	// Move cursor up 2.
	tw.Write([]byte("\033[2A"))
	if tb.CursorR != 3 || tb.CursorC != 5 {
		t.Errorf("cursor after ESC[2A = (%d, %d), want (3, 5)", tb.CursorR, tb.CursorC)
	}

	// Move cursor right 3.
	tw.Write([]byte("\033[3C"))
	if tb.CursorR != 3 || tb.CursorC != 8 {
		t.Errorf("cursor after ESC[3C = (%d, %d), want (3, 8)", tb.CursorR, tb.CursorC)
	}
}

func TestTextWriterIncompleteCSI(t *testing.T) {
	tb := NewTextBuffer(80, 25)
	tw := &textWriter{buf: tb}
	// Incomplete CSI at end of buffer - should be written as literal chars.
	tw.Write([]byte("\033["))
	// ESC (0x1B) and '[' should be written as literal characters.
	if tb.Cells[0][0].Char != 0x1B {
		t.Errorf("Cells[0][0].Char = 0x%02X, want 0x1B (ESC)", tb.Cells[0][0].Char)
	}
	if tb.Cells[0][1].Char != '[' {
		t.Errorf("Cells[0][1].Char = %q, want '['", tb.Cells[0][1].Char)
	}
}

func TestAnsiFgToBasic(t *testing.T) {
	tests := []struct {
		code    int
		want    byte
		wantOK  bool
	}{
		{30, 0, true},
		{97, 15, true},
		{999, 0, false},
	}
	for _, tc := range tests {
		got, ok := ansiFgToBasic(tc.code)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("ansiFgToBasic(%d) = (%d, %v), want (%d, %v)", tc.code, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestAnsiBgToBasic(t *testing.T) {
	tests := []struct {
		code    int
		want    byte
		wantOK  bool
	}{
		{40, 0, true},
		{107, 15, true},
		{999, 0, false},
	}
	for _, tc := range tests {
		got, ok := ansiBgToBasic(tc.code)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("ansiBgToBasic(%d) = (%d, %v), want (%d, %v)", tc.code, got, ok, tc.want, tc.wantOK)
		}
	}
}
