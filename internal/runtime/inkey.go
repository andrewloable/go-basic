//go:build linux || darwin

package runtime

import (
	"os"
	"syscall"
	"unsafe"
)

// termState stores the original terminal attributes so they can be restored.
var termState *syscall.Termios
var termRaw bool

// initRawTerminal switches stdin to raw mode (no echo, no line buffering).
// Returns false if stdin is not a terminal or if raw mode setup fails.
func initRawTerminal() bool {
	if termRaw {
		return true
	}
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		ioctlGetTermios,
		uintptr(unsafe.Pointer(&t))); errno != 0 {
		return false // not a terminal
	}
	saved := t
	termState = &saved

	// Raw mode: disable ICANON, ECHO; set VMIN=0, VTIME=0 for non-blocking read.
	t.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	t.Cflag |= syscall.CS8
	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	t.Cc[syscall.VMIN] = 0
	t.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		ioctlSetTermios,
		uintptr(unsafe.Pointer(&t))); errno != 0 {
		termState = nil
		return false
	}
	termRaw = true
	return true
}

// RestoreTerminal restores the original terminal settings. Call this on program
// exit or when done using INKEY$. It is safe to call if raw mode was never set.
func RestoreTerminal() {
	if !termRaw || termState == nil {
		return
	}
	syscall.Syscall(syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		ioctlSetTermios,
		uintptr(unsafe.Pointer(termState)))
	termRaw = false
	termState = nil
}

// Inkey returns the next character from stdin without waiting. Returns "" if no
// key is available. Special keys (arrows, function keys, etc.) return a two-byte
// string matching QBasic scan-code convention: chr(0) + chr(scanCode).
//
// On first call it switches stdin to raw mode. If stdin is not a tty (piped),
// it always returns "".
func Inkey() string {
	// When the Ebitengine graphics window is active, read keys from it
	// instead of the terminal.
	if IsGraphicsMode() {
		return InkeyFromGraphics()
	}

	if !termRaw {
		if !initRawTerminal() {
			return "" // not a terminal — degrade gracefully
		}
	}

	buf := make([]byte, 8)
	n, err := os.Stdin.Read(buf)
	if n == 0 || err != nil {
		return ""
	}

	b := buf[:n]

	// Single printable byte or control character.
	if n == 1 {
		switch b[0] {
		case 127, 8: // Backspace / DEL
			return "\x08"
		case 13: // Enter
			return "\r"
		case 3: // Ctrl+C — propagate as ctrl character
			return "\x03"
		default:
			return string(b[0])
		}
	}

	// Escape sequences for special keys.
	if b[0] == 27 {
		return parseEscapeSequence(b[1:])
	}

	return string(b[0])
}

// parseEscapeSequence maps an ANSI escape sequence to a QBasic-compatible
// two-byte string (chr(0) + chr(scanCode)).
func parseEscapeSequence(seq []byte) string {
	if len(seq) == 0 {
		return "\x1b" // bare ESC
	}

	// ESC [ ...
	if seq[0] == '[' && len(seq) >= 2 {
		// Multi-byte sequences ending in '~' take priority (function/editing keys).
		if seq[len(seq)-1] == '~' {
			code := string(seq[1 : len(seq)-1])
			switch code {
			case "11", "1": return "\x00\x3B" // F1  = scan 59
			case "12":      return "\x00\x3C" // F2  = scan 60
			case "13":      return "\x00\x3D" // F3  = scan 61
			case "14":      return "\x00\x3E" // F4  = scan 62
			case "15":      return "\x00\x3F" // F5  = scan 63
			case "17":      return "\x00\x40" // F6  = scan 64
			case "18":      return "\x00\x41" // F7  = scan 65
			case "19":      return "\x00\x42" // F8  = scan 66
			case "20":      return "\x00\x43" // F9  = scan 67
			case "21":      return "\x00\x44" // F10 = scan 68
			case "2":       return "\x00\x52" // Insert  = scan 82
			case "3":       return "\x00\x53" // Delete  = scan 83
			case "5":       return "\x00\x49" // PgUp    = scan 73
			case "6":       return "\x00\x51" // PgDn    = scan 81
			}
		}
		// Single-letter sequences (cursor keys, Home, End).
		switch seq[1] {
		case 'A': return "\x00\x48" // Up arrow    = scan 72
		case 'B': return "\x00\x50" // Down arrow  = scan 80
		case 'C': return "\x00\x4D" // Right arrow = scan 77
		case 'D': return "\x00\x4B" // Left arrow  = scan 75
		case 'H': return "\x00\x47" // Home        = scan 71
		case 'F': return "\x00\x4F" // End         = scan 79
		}
	}

	// ESC O ...  (SS3 sequences used by some terminals for arrow keys)
	if seq[0] == 'O' && len(seq) >= 2 {
		switch seq[1] {
		case 'A': return "\x00\x48"
		case 'B': return "\x00\x50"
		case 'C': return "\x00\x4D"
		case 'D': return "\x00\x4B"
		case 'P': return "\x00\x3B" // F1
		case 'Q': return "\x00\x3C" // F2
		case 'R': return "\x00\x3D" // F3
		case 'S': return "\x00\x3E" // F4
		}
	}

	return "\x1b" // unrecognised escape sequence — return bare ESC
}
