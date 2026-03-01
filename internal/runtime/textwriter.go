// textwriter.go — Stdout interceptor for text-mode rendering in Ebitengine.
//
// When the Ebitengine window is open and the program switches to SCREEN 0,
// this module replaces os.Stdout with a pipe. A background goroutine reads
// from the pipe, parses ANSI escape sequences (emitted by LOCATE/CLS/COLOR),
// and writes characters into the TextBuffer. The Ebitengine Draw() method
// then renders the TextBuffer contents using the CP437 bitmap font.

package runtime

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// textWriter implements io.Writer and parses ANSI CSI sequences from the
// output stream, translating them into TextBuffer operations.
type textWriter struct {
	buf *TextBuffer
}

// Write parses bytes for ANSI escape sequences and writes characters
// to the text buffer.
func (tw *textWriter) Write(p []byte) (n int, err error) {
	i := 0
	for i < len(p) {
		if p[i] == 0x1B && i+1 < len(p) && p[i+1] == '[' {
			// Start of CSI sequence: \033[ ... <letter>
			j := i + 2
			// Collect parameter bytes (digits, semicolons, question marks).
			for j < len(p) && ((p[j] >= '0' && p[j] <= '9') || p[j] == ';' || p[j] == '?') {
				j++
			}
			if j >= len(p) {
				// Incomplete sequence at end of buffer — write remaining as literal.
				for k := i; k < len(p); k++ {
					tw.buf.PutChar(p[k])
				}
				break
			}
			// p[j] is the command letter.
			cmd := p[j]
			params := string(p[i+2 : j])
			j++ // advance past command letter
			tw.handleCSI(cmd, params)
			i = j
		} else {
			tw.buf.PutChar(p[i])
			i++
		}
	}
	return len(p), nil
}

// handleCSI processes a parsed CSI sequence.
func (tw *textWriter) handleCSI(cmd byte, params string) {
	switch cmd {
	case 'H', 'f':
		// Cursor position: \033[row;colH
		row, col := 1, 1
		parts := strings.Split(params, ";")
		if len(parts) >= 1 && parts[0] != "" {
			if v, err := strconv.Atoi(parts[0]); err == nil {
				row = v
			}
		}
		if len(parts) >= 2 && parts[1] != "" {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				col = v
			}
		}
		tw.buf.SetCursor(row-1, col-1) // Convert 1-based to 0-based.

	case 'J':
		// Erase display: \033[2J = clear screen.
		// Only clears the text buffer. The graphics framebuffer is managed
		// by rt.Cls() which is called directly from generated code.
		if params == "2" || params == "" {
			tw.buf.Clear()
		}

	case 'm':
		// SGR (Select Graphic Rendition): \033[fg;bgm
		if params == "" || params == "0" {
			// Reset to defaults.
			tw.buf.SetColor(7, 0)
			return
		}
		parts := strings.Split(params, ";")
		fg := tw.buf.CurFg
		bg := tw.buf.CurBg
		for _, p := range parts {
			code, err := strconv.Atoi(p)
			if err != nil {
				continue
			}
			if basicColor, ok := ansiFgToBasic(code); ok {
				fg = basicColor
			} else if basicColor, ok := ansiBgToBasic(code); ok {
				bg = basicColor
			} else if code == 0 {
				fg = 7
				bg = 0
			}
		}
		tw.buf.SetColor(fg, bg)

	case 'A':
		// Cursor up.
		n := 1
		if params != "" {
			if v, err := strconv.Atoi(params); err == nil {
				n = v
			}
		}
		tw.buf.mu.Lock()
		tw.buf.CursorR -= n
		if tw.buf.CursorR < 0 {
			tw.buf.CursorR = 0
		}
		tw.buf.mu.Unlock()

	case 'B':
		// Cursor down.
		n := 1
		if params != "" {
			if v, err := strconv.Atoi(params); err == nil {
				n = v
			}
		}
		tw.buf.mu.Lock()
		tw.buf.CursorR += n
		if tw.buf.CursorR >= tw.buf.Rows {
			tw.buf.CursorR = tw.buf.Rows - 1
		}
		tw.buf.mu.Unlock()

	case 'C':
		// Cursor forward.
		n := 1
		if params != "" {
			if v, err := strconv.Atoi(params); err == nil {
				n = v
			}
		}
		tw.buf.mu.Lock()
		tw.buf.CursorC += n
		if tw.buf.CursorC >= tw.buf.Cols {
			tw.buf.CursorC = tw.buf.Cols - 1
		}
		tw.buf.mu.Unlock()

	case 'D':
		// Cursor backward.
		n := 1
		if params != "" {
			if v, err := strconv.Atoi(params); err == nil {
				n = v
			}
		}
		tw.buf.mu.Lock()
		tw.buf.CursorC -= n
		if tw.buf.CursorC < 0 {
			tw.buf.CursorC = 0
		}
		tw.buf.mu.Unlock()
	}
}

// ansiFgToBasic maps ANSI foreground SGR codes to BASIC color indices (0-15).
func ansiFgToBasic(code int) (byte, bool) {
	switch code {
	case 30:
		return 0, true // black
	case 34:
		return 1, true // blue
	case 32:
		return 2, true // green
	case 36:
		return 3, true // cyan
	case 31:
		return 4, true // red
	case 35:
		return 5, true // magenta
	case 33:
		return 6, true // brown
	case 37:
		return 7, true // light gray
	case 90:
		return 8, true // dark gray
	case 94:
		return 9, true // light blue
	case 92:
		return 10, true // light green
	case 96:
		return 11, true // light cyan
	case 91:
		return 12, true // light red
	case 95:
		return 13, true // light magenta
	case 93:
		return 14, true // yellow
	case 97:
		return 15, true // bright white
	}
	return 0, false
}

// ansiBgToBasic maps ANSI background SGR codes to BASIC color indices (0-15).
func ansiBgToBasic(code int) (byte, bool) {
	switch code {
	case 40:
		return 0, true
	case 44:
		return 1, true
	case 42:
		return 2, true
	case 46:
		return 3, true
	case 41:
		return 4, true
	case 45:
		return 5, true
	case 43:
		return 6, true
	case 47:
		return 7, true
	case 100:
		return 8, true
	case 104:
		return 9, true
	case 102:
		return 10, true
	case 106:
		return 11, true
	case 101:
		return 12, true
	case 105:
		return 13, true
	case 103:
		return 14, true
	case 107:
		return 15, true
	}
	return 0, false
}

var (
	// textWriterMu protects textwriter install/uninstall.
	textWriterMu sync.Mutex

	// savedStdout holds the original os.Stdout while the text writer is active.
	savedStdout *os.File

	// pipeWriter is the write end of the pipe that replaces os.Stdout.
	pipeWriter *os.File

	// pipeReader is the read end of the pipe.
	pipeReader *os.File

	// textWriterActive tracks whether the interceptor is installed.
	textWriterActive bool
)

// InstallTextWriter replaces os.Stdout with a pipe that feeds characters
// into the TextBuffer (80x25). Must be called when the Ebitengine window is open.
func InstallTextWriter() {
	InstallTextWriterWithSize(80, 25)
}

// InstallTextWriterWithSize replaces os.Stdout with a pipe that feeds characters
// into a TextBuffer of the given dimensions. If the text writer is already
// active, it replaces the TextBuffer without reinstalling the pipe.
func InstallTextWriterWithSize(cols, rows int) {
	textWriterMu.Lock()
	defer textWriterMu.Unlock()

	if textWriterActive {
		// Already active — just replace the text buffer with new dimensions.
		TextBuf = NewTextBuffer(cols, rows)
		return
	}

	// Create the text buffer.
	TextBuf = NewTextBuffer(cols, rows)

	// Create a pipe to intercept stdout.
	pr, pw, err := os.Pipe()
	if err != nil {
		return // silently fail — text will go to terminal
	}

	savedStdout = os.Stdout
	pipeReader = pr
	pipeWriter = pw
	os.Stdout = pw
	textWriterActive = true

	// Spawn a goroutine to read from the pipe and feed the text writer.
	go func() {
		tw := &textWriter{}
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				// Always use the current TextBuf (it may be replaced on mode switch).
				tw.buf = TextBuf
				tw.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()
}

// UninstallTextWriter restores the original os.Stdout and closes the pipe.
func UninstallTextWriter() {
	textWriterMu.Lock()
	defer textWriterMu.Unlock()

	if !textWriterActive {
		return
	}

	// Restore stdout first so subsequent writes go to the real stdout.
	os.Stdout = savedStdout

	// Close the pipe to terminate the reader goroutine.
	pipeWriter.Close()
	// Drain and close the read end.
	io.Copy(io.Discard, pipeReader)
	pipeReader.Close()

	savedStdout = nil
	pipeWriter = nil
	pipeReader = nil
	textWriterActive = false
}

// IsTextWriterActive returns true if the stdout interceptor is installed.
func IsTextWriterActive() bool {
	textWriterMu.Lock()
	defer textWriterMu.Unlock()
	return textWriterActive
}
