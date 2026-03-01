// textbuffer.go — Text cell buffer for rendering text in the Ebitengine window.
//
// When the Ebitengine window is open, this buffer captures text output (via the
// stdout interceptor) and provides a grid of character cells. In SCREEN 0 (text
// mode), the Ebitengine Draw() method renders the cells using the CP437 bitmap
// font. In graphics modes (SCREEN 1-13), characters are rendered directly into
// the framebuffer as they are written, matching QBasic's behavior.

package runtime

import "sync"

// TextCell represents a single character cell in the text buffer.
type TextCell struct {
	Char byte // CP437 character code
	Fg   byte // foreground color index (0-15)
	Bg   byte // background color index (0-15)
}

// TextBuffer holds a grid of text cells with cursor state.
type TextBuffer struct {
	mu           sync.Mutex
	Cols, Rows   int          // dimensions (default 80x25)
	Cells        [][]TextCell // [row][col]
	CursorR      int          // 0-based cursor row
	CursorC      int          // 0-based cursor column
	CurFg, CurBg byte         // current color state
}

// TextBuf is the package-level text buffer instance, created when the
// text writer is installed (any SCREEN mode with the Ebitengine window open).
var TextBuf *TextBuffer

// NewTextBuffer allocates a text buffer of the given dimensions,
// filled with spaces using fg=7 (light gray) and bg=0 (black).
func NewTextBuffer(cols, rows int) *TextBuffer {
	tb := &TextBuffer{
		Cols:  cols,
		Rows:  rows,
		CurFg: 7,
		CurBg: 0,
	}
	tb.Cells = make([][]TextCell, rows)
	for r := range tb.Cells {
		tb.Cells[r] = make([]TextCell, cols)
		for c := range tb.Cells[r] {
			tb.Cells[r][c] = TextCell{Char: ' ', Fg: 7, Bg: 0}
		}
	}
	return tb
}

// Clear resets all cells to spaces and moves the cursor to the top-left.
func (tb *TextBuffer) Clear() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for r := range tb.Cells {
		for c := range tb.Cells[r] {
			tb.Cells[r][c] = TextCell{Char: ' ', Fg: tb.CurFg, Bg: tb.CurBg}
		}
	}
	tb.CursorR = 0
	tb.CursorC = 0
}

// SetCursor moves the cursor to the given 0-based row and column.
func (tb *TextBuffer) SetCursor(row, col int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if row < 0 {
		row = 0
	}
	if row >= tb.Rows {
		row = tb.Rows - 1
	}
	if col < 0 {
		col = 0
	}
	if col >= tb.Cols {
		col = tb.Cols - 1
	}
	tb.CursorR = row
	tb.CursorC = col
}

// SetColor sets the current foreground and background color indices.
func (tb *TextBuffer) SetColor(fg, bg byte) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.CurFg = fg
	tb.CurBg = bg
}

// PutChar writes a character at the current cursor position and advances the cursor.
// Handles control characters: \n (newline), \r (carriage return), \t (tab).
func (tb *TextBuffer) PutChar(ch byte) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.putCharLocked(ch)
}

// putCharLocked is the internal implementation (caller must hold tb.mu).
func (tb *TextBuffer) putCharLocked(ch byte) {
	switch ch {
	case '\n':
		tb.CursorC = 0
		tb.CursorR++
		if tb.CursorR >= tb.Rows {
			tb.scrollUpLocked()
			tb.CursorR = tb.Rows - 1
		}
	case '\r':
		tb.CursorC = 0
	case '\t':
		// Advance to next 8-column tab stop.
		target := (tb.CursorC/8 + 1) * 8
		for tb.CursorC < target && tb.CursorC < tb.Cols {
			tb.Cells[tb.CursorR][tb.CursorC] = TextCell{Char: ' ', Fg: tb.CurFg, Bg: tb.CurBg}
			tb.renderCellToFB(tb.CursorR, tb.CursorC, ' ', tb.CurFg, tb.CurBg)
			tb.CursorC++
		}
		if tb.CursorC >= tb.Cols {
			tb.CursorC = 0
			tb.CursorR++
			if tb.CursorR >= tb.Rows {
				tb.scrollUpLocked()
				tb.CursorR = tb.Rows - 1
			}
		}
	case '\b':
		// Backspace: move cursor back one column.
		if tb.CursorC > 0 {
			tb.CursorC--
		}
	default:
		// Printable character.
		if tb.CursorR < tb.Rows && tb.CursorC < tb.Cols {
			tb.Cells[tb.CursorR][tb.CursorC] = TextCell{Char: ch, Fg: tb.CurFg, Bg: tb.CurBg}
			tb.renderCellToFB(tb.CursorR, tb.CursorC, ch, tb.CurFg, tb.CurBg)
		}
		tb.CursorC++
		if tb.CursorC >= tb.Cols {
			tb.CursorC = 0
			tb.CursorR++
			if tb.CursorR >= tb.Rows {
				tb.scrollUpLocked()
				tb.CursorR = tb.Rows - 1
			}
		}
	}
}

// renderCellToFB renders a character glyph into the graphics framebuffer.
// Called when in a graphics SCREEN mode (not SCREEN 0 text mode) so that
// PRINT/INPUT text appears in the Ebitengine window, matching QBasic behavior.
// No-op in text mode (SCREEN 0) or when no framebuffer is active.
func (tb *TextBuffer) renderCellToFB(row, col int, ch, fg, bg byte) {
	if gfxTextMode || !gfxActive {
		return
	}
	s := CurrentScreen
	if s == nil || s.Framebuffer == nil {
		return
	}
	// Calculate character cell height: screenHeight / textRows.
	// QBasic uses different font heights per mode (e.g. 14px for SCREEN 9,
	// 8px for SCREEN 13). We scale the 8×16 font to fit.
	charH := s.Height / tb.Rows
	if charH < 1 {
		charH = 1
	}
	charW := 8
	px := col * charW
	py := row * charH

	gfxMu.Lock()
	DrawCharToFramebuffer(s.Framebuffer, s.Width, s.Height, px, py, charH, ch, fg, bg)
	gfxMu.Unlock()
}

// PutString writes a string of characters to the buffer.
func (tb *TextBuffer) PutString(s string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for i := 0; i < len(s); i++ {
		tb.putCharLocked(s[i])
	}
}

// scrollUpLocked shifts all rows up by one and clears the bottom row.
func (tb *TextBuffer) scrollUpLocked() {
	for r := 1; r < tb.Rows; r++ {
		copy(tb.Cells[r-1], tb.Cells[r])
	}
	// Clear the bottom row.
	bottom := tb.Cells[tb.Rows-1]
	for c := range bottom {
		bottom[c] = TextCell{Char: ' ', Fg: tb.CurFg, Bg: tb.CurBg}
	}
}
