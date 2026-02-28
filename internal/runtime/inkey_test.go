package runtime

import (
	"os"
	"testing"
)

// withPipedStdin replaces os.Stdin with a pipe, writes data into it, and runs f.
// This lets tests exercise Inkey() without a real terminal.
func withPipedStdin(t *testing.T, data []byte, f func()) {
	t.Helper()
	// Temporarily disable raw-mode so Inkey falls through to the pipe read.
	wasRaw := termRaw
	savedState := termState
	termRaw = true // pretend already in raw mode so initRawTerminal is skipped
	termState = nil

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = r

	_, _ = w.Write(data)
	w.Close()

	defer func() {
		os.Stdin = origStdin
		r.Close()
		termRaw = wasRaw
		termState = savedState
	}()

	f()
}

func TestInkeyEmptyWhenNoByte(t *testing.T) {
	withPipedStdin(t, []byte{}, func() {
		got := Inkey()
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

func TestInkeyPrintable(t *testing.T) {
	withPipedStdin(t, []byte("A"), func() {
		got := Inkey()
		if got != "A" {
			t.Errorf("expected 'A', got %q", got)
		}
	})
}

func TestInkeyEnter(t *testing.T) {
	withPipedStdin(t, []byte{13}, func() {
		got := Inkey()
		if got != "\r" {
			t.Errorf("expected CR, got %q", got)
		}
	})
}

func TestInkeyBackspace(t *testing.T) {
	withPipedStdin(t, []byte{127}, func() {
		got := Inkey()
		if got != "\x08" {
			t.Errorf("expected BS, got %q", got)
		}
	})
}

func TestInkeyArrowUp(t *testing.T) {
	// ESC [ A = cursor up = scan 72
	withPipedStdin(t, []byte{27, '[', 'A'}, func() {
		got := Inkey()
		if got != "\x00\x48" {
			t.Errorf("up arrow: expected \\x00\\x48, got %q", got)
		}
	})
}

func TestInkeyArrowDown(t *testing.T) {
	withPipedStdin(t, []byte{27, '[', 'B'}, func() {
		got := Inkey()
		if got != "\x00\x50" {
			t.Errorf("down arrow: expected \\x00\\x50, got %q", got)
		}
	})
}

func TestInkeyArrowRight(t *testing.T) {
	withPipedStdin(t, []byte{27, '[', 'C'}, func() {
		got := Inkey()
		if got != "\x00\x4D" {
			t.Errorf("right arrow: expected \\x00\\x4D, got %q", got)
		}
	})
}

func TestInkeyArrowLeft(t *testing.T) {
	withPipedStdin(t, []byte{27, '[', 'D'}, func() {
		got := Inkey()
		if got != "\x00\x4B" {
			t.Errorf("left arrow: expected \\x00\\x4B, got %q", got)
		}
	})
}

func TestInkeyF1(t *testing.T) {
	// ESC O P or ESC [ 1 1 ~
	withPipedStdin(t, []byte{27, 'O', 'P'}, func() {
		got := Inkey()
		if got != "\x00\x3B" {
			t.Errorf("F1: expected \\x00\\x3B, got %q", got)
		}
	})
}

func TestInkeyF10(t *testing.T) {
	withPipedStdin(t, []byte{27, '[', '2', '1', '~'}, func() {
		got := Inkey()
		if got != "\x00\x44" {
			t.Errorf("F10: expected \\x00\\x44, got %q", got)
		}
	})
}

func TestInkeyBareEsc(t *testing.T) {
	withPipedStdin(t, []byte{27}, func() {
		got := Inkey()
		if got != "\x1b" {
			t.Errorf("bare ESC: expected \\x1b, got %q", got)
		}
	})
}
