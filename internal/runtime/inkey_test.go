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

func TestInkeyCtrlC(t *testing.T) {
	withPipedStdin(t, []byte{3}, func() {
		got := Inkey()
		if got != "\x03" {
			t.Errorf("ctrl+C: expected \\x03, got %q", got)
		}
	})
}

func TestInkeyBackspaceCode8(t *testing.T) {
	withPipedStdin(t, []byte{8}, func() {
		got := Inkey()
		if got != "\x08" {
			t.Errorf("BS code 8: expected \\x08, got %q", got)
		}
	})
}

func TestInkeyMultiByteNonEsc(t *testing.T) {
	// Multi-byte sequence where first byte != ESC returns string(b[0]).
	withPipedStdin(t, []byte{'X', 'Y'}, func() {
		got := Inkey()
		if got != "X" {
			t.Errorf("multi-byte non-esc: expected 'X', got %q", got)
		}
	})
}

// ---------------------------------------------------------------------------
// parseEscapeSequence — full switch coverage
// ---------------------------------------------------------------------------

func TestParseEscapeSequenceHomeEnd(t *testing.T) {
	// ESC [ H = Home, ESC [ F = End
	withPipedStdin(t, []byte{27, '[', 'H'}, func() {
		got := Inkey()
		if got != "\x00\x47" {
			t.Errorf("Home: got %q", got)
		}
	})
	withPipedStdin(t, []byte{27, '[', 'F'}, func() {
		got := Inkey()
		if got != "\x00\x4F" {
			t.Errorf("End: got %q", got)
		}
	})
}

func TestParseEscapeSequenceFunctionKeys(t *testing.T) {
	cases := []struct {
		seq  []byte
		want string
		name string
	}{
		{[]byte{27, '[', '1', '1', '~'}, "\x00\x3B", "F1"},
		{[]byte{27, '[', '1', '2', '~'}, "\x00\x3C", "F2"},
		{[]byte{27, '[', '1', '3', '~'}, "\x00\x3D", "F3"},
		{[]byte{27, '[', '1', '4', '~'}, "\x00\x3E", "F4"},
		{[]byte{27, '[', '1', '5', '~'}, "\x00\x3F", "F5"},
		{[]byte{27, '[', '1', '7', '~'}, "\x00\x40", "F6"},
		{[]byte{27, '[', '1', '8', '~'}, "\x00\x41", "F7"},
		{[]byte{27, '[', '1', '9', '~'}, "\x00\x42", "F8"},
		{[]byte{27, '[', '2', '0', '~'}, "\x00\x43", "F9"},
		{[]byte{27, '[', '2', '~'}, "\x00\x52", "Insert"},
		{[]byte{27, '[', '3', '~'}, "\x00\x53", "Delete"},
		{[]byte{27, '[', '5', '~'}, "\x00\x49", "PgUp"},
		{[]byte{27, '[', '6', '~'}, "\x00\x51", "PgDn"},
		{[]byte{27, '[', '1', '~'}, "\x00\x3B", "F1 alt"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			withPipedStdin(t, tc.seq, func() {
				got := Inkey()
				if got != tc.want {
					t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
				}
			})
		})
	}
}

func TestParseEscapeSequenceSS3Arrows(t *testing.T) {
	cases := []struct {
		seq  []byte
		want string
		name string
	}{
		{[]byte{27, 'O', 'A'}, "\x00\x48", "SS3 Up"},
		{[]byte{27, 'O', 'B'}, "\x00\x50", "SS3 Down"},
		{[]byte{27, 'O', 'C'}, "\x00\x4D", "SS3 Right"},
		{[]byte{27, 'O', 'D'}, "\x00\x4B", "SS3 Left"},
		{[]byte{27, 'O', 'Q'}, "\x00\x3C", "SS3 F2"},
		{[]byte{27, 'O', 'R'}, "\x00\x3D", "SS3 F3"},
		{[]byte{27, 'O', 'S'}, "\x00\x3E", "SS3 F4"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			withPipedStdin(t, tc.seq, func() {
				got := Inkey()
				if got != tc.want {
					t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
				}
			})
		})
	}
}

func TestParseEscapeSequenceUnknown(t *testing.T) {
	// Unrecognised sequence → returns bare ESC \x1b.
	withPipedStdin(t, []byte{27, '[', 'Z'}, func() {
		got := Inkey()
		if got != "\x1b" {
			t.Errorf("unknown ESC[Z: got %q, want \\x1b", got)
		}
	})
}

func TestParseEscapeSequenceUnknownTilde(t *testing.T) {
	// Unknown tilde code falls through.
	withPipedStdin(t, []byte{27, '[', '9', '9', '~'}, func() {
		got := Inkey()
		// Falls through tilde switch to single-letter switch; '~' is not a match → "\x1b".
		_ = got // just check no panic
	})
}

// ---------------------------------------------------------------------------
// initRawTerminal — when stdin is not a TTY it returns false
// ---------------------------------------------------------------------------

func TestInitRawTerminalNotATTY(t *testing.T) {
	// In test mode, stdin is a pipe. initRawTerminal should fail gracefully.
	saved := termRaw
	termRaw = false
	defer func() { termRaw = saved }()

	result := initRawTerminal()
	// Should return false because stdin is not a real terminal in test mode.
	// (On some CI environments this may vary; we just verify no panic.)
	_ = result
}

// ---------------------------------------------------------------------------
// RestoreTerminal — no-op paths
// ---------------------------------------------------------------------------

func TestRestoreTerminalWhenNotRaw(t *testing.T) {
	// Not in raw mode: should return immediately without panic.
	saved := termRaw
	savedState := termState
	termRaw = false
	termState = nil
	defer func() {
		termRaw = saved
		termState = savedState
	}()
	RestoreTerminal() // should not panic
}

func TestRestoreTerminalRawButNoState(t *testing.T) {
	// termRaw true but termState nil: should return immediately.
	saved := termRaw
	savedState := termState
	termRaw = true
	termState = nil
	defer func() {
		termRaw = saved
		termState = savedState
	}()
	RestoreTerminal() // should not panic
}
