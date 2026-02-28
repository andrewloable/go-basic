//go:build windows

package runtime

import (
	"syscall"
	"unsafe"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procPeekConsoleInputW  = kernel32.NewProc("PeekConsoleInputW")
	procReadConsoleInputW  = kernel32.NewProc("ReadConsoleInputW")

	stdinHandle   syscall.Handle
	origConsMode  uint32
	consoleSetup  bool
)

const (
	stdInputHandle = ^uintptr(0) - 10 + 1 // STD_INPUT_HANDLE = -10

	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableWindowInput    = 0x0008

	keyEventType = 0x0001
)

// inputRecord matches the Windows INPUT_RECORD structure (KEY_EVENT_RECORD variant).
type inputRecord struct {
	EventType uint16
	_         [2]byte // padding
	// KEY_EVENT_RECORD fields:
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

// initConsole switches stdin to raw mode for non-blocking key reads.
func initConsole() bool {
	if consoleSetup {
		return true
	}
	h, _, _ := procGetStdHandle.Call(stdInputHandle)
	if h == 0 || h == ^uintptr(0) {
		return false
	}
	stdinHandle = syscall.Handle(h)

	r, _, _ := procGetConsoleMode.Call(uintptr(stdinHandle), uintptr(unsafe.Pointer(&origConsMode)))
	if r == 0 {
		return false
	}

	// Disable line input and echo; keep processed input for Ctrl+C.
	newMode := origConsMode &^ (enableLineInput | enableEchoInput)
	r, _, _ = procSetConsoleMode.Call(uintptr(stdinHandle), uintptr(newMode))
	if r == 0 {
		return false
	}
	consoleSetup = true
	return true
}

// RestoreTerminal restores the original console mode.
func RestoreTerminal() {
	if !consoleSetup {
		return
	}
	procSetConsoleMode.Call(uintptr(stdinHandle), uintptr(origConsMode))
	consoleSetup = false
}

// Inkey returns the next character from the keyboard buffer without waiting.
// Returns "" if no key is available.
func Inkey() string {
	// When the Ebitengine graphics window is active, read keys from it.
	if IsGraphicsMode() {
		return InkeyFromGraphics()
	}

	if !consoleSetup {
		if !initConsole() {
			return ""
		}
	}

	// Check if there are pending input events.
	var numEvents uint32
	r, _, _ := procPeekConsoleInputW.Call(
		uintptr(stdinHandle),
		0, 0,
		uintptr(unsafe.Pointer(&numEvents)),
	)
	if r == 0 || numEvents == 0 {
		return ""
	}

	// Read the next input event.
	var rec inputRecord
	var numRead uint32
	r, _, _ = procReadConsoleInputW.Call(
		uintptr(stdinHandle),
		uintptr(unsafe.Pointer(&rec)),
		1,
		uintptr(unsafe.Pointer(&numRead)),
	)
	if r == 0 || numRead == 0 {
		return ""
	}

	// Only process key-down events.
	if rec.EventType != keyEventType || rec.KeyDown == 0 {
		return ""
	}

	ch := rec.UnicodeChar
	scan := rec.VirtualScanCode

	// If a printable character was generated, return it directly.
	if ch != 0 {
		switch ch {
		case 13: // Enter
			return "\r"
		case 8: // Backspace
			return "\x08"
		case 27: // Escape
			return "\x1b"
		case 3: // Ctrl+C
			return "\x03"
		default:
			return string(rune(ch))
		}
	}

	// No printable character — map the scan code to QBasic convention.
	scanStr := mapWindowsScanCode(scan)
	if scanStr != "" {
		return scanStr
	}

	return ""
}

// mapWindowsScanCode maps Windows virtual scan codes to QBasic two-byte strings.
func mapWindowsScanCode(scan uint16) string {
	switch scan {
	case 0x48: return "\x00\x48" // Up arrow
	case 0x50: return "\x00\x50" // Down arrow
	case 0x4D: return "\x00\x4D" // Right arrow
	case 0x4B: return "\x00\x4B" // Left arrow
	case 0x47: return "\x00\x47" // Home
	case 0x4F: return "\x00\x4F" // End
	case 0x49: return "\x00\x49" // Page Up
	case 0x51: return "\x00\x51" // Page Down
	case 0x52: return "\x00\x52" // Insert
	case 0x53: return "\x00\x53" // Delete
	case 0x3B: return "\x00\x3B" // F1
	case 0x3C: return "\x00\x3C" // F2
	case 0x3D: return "\x00\x3D" // F3
	case 0x3E: return "\x00\x3E" // F4
	case 0x3F: return "\x00\x3F" // F5
	case 0x40: return "\x00\x40" // F6
	case 0x41: return "\x00\x41" // F7
	case 0x42: return "\x00\x42" // F8
	case 0x43: return "\x00\x43" // F9
	case 0x44: return "\x00\x44" // F10
	}
	return ""
}
