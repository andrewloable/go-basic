//go:build !linux && !darwin && !windows

package runtime

// Inkey returns the next character from the keyboard buffer without waiting.
// Returns empty string if no key is available. This stub is used on platforms
// where raw terminal I/O is not implemented.
func Inkey() string {
	// When the Ebitengine graphics window is active, read keys from it.
	if IsGraphicsMode() {
		return InkeyFromGraphics()
	}
	return ""
}

// RestoreTerminal is a no-op on unsupported platforms.
func RestoreTerminal() {}
