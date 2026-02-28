package runtime

import (
	"math"
	"testing"
	"time"
)

// capturedTone records a single Sound call for testing.
type capturedTone struct {
	freq float64
	dur  float64
}

// installMockBackend replaces SoundBackend with a recorder and disables sleep
// by fast-forwarding time. Returns a pointer to the collected tones and a
// restore function.
func installMockBackend(t *testing.T) (*[]capturedTone, func()) {
	t.Helper()
	tones := &[]capturedTone{}
	orig := SoundBackend
	SoundBackend = func(freq, dur float64) {
		*tones = append(*tones, capturedTone{freq, dur})
	}
	return tones, func() { SoundBackend = orig }
}

func TestSoundNoOp(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	// Negative/zero args should be ignored.
	Sound(0, 10)
	Sound(440, 0)
	Sound(-100, 5)
	if len(*tones) != 0 {
		t.Errorf("expected no tones for invalid args, got %d", len(*tones))
	}
}

func TestSoundCallsBackend(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	// Temporarily override Sleep so the test doesn't block.
	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	Sound(440, 18.2) // 1 second at 440 Hz
	if len(*tones) != 1 {
		t.Fatalf("expected 1 tone, got %d", len(*tones))
	}
	if math.Abs((*tones)[0].freq-440) > 0.1 {
		t.Errorf("freq: got %.2f want 440", (*tones)[0].freq)
	}
	if math.Abs((*tones)[0].dur-1.0) > 0.01 {
		t.Errorf("dur: got %.4f want 1.0", (*tones)[0].dur)
	}
}

func TestPlaySimpleNotes(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	// C D E F G A B — 7 notes in default octave 4, default length 4 (quarter).
	Play("CDEFGAB")
	if len(*tones) != 7 {
		t.Fatalf("expected 7 tones for CDEFGAB, got %d", len(*tones))
	}
	// Verify ascending frequencies.
	for i := 1; i < len(*tones); i++ {
		if (*tones)[i].freq <= (*tones)[i-1].freq {
			t.Errorf("note %d: freq %.2f <= prev %.2f (expected ascending)", i, (*tones)[i].freq, (*tones)[i-1].freq)
		}
	}
}

func TestPlayOctaveChange(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	// O4 A then O5 A — octave 5 should be exactly double octave 4.
	Play("O4 A O5 A")
	if len(*tones) != 2 {
		t.Fatalf("expected 2 tones, got %d", len(*tones))
	}
	ratio := (*tones)[1].freq / (*tones)[0].freq
	if math.Abs(ratio-2.0) > 0.01 {
		t.Errorf("O5 A should be 2x O4 A, got ratio %.4f", ratio)
	}
}

func TestPlayOctaveUpDown(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	// > raises octave, < lowers it.
	Play("O4 A > A < A")
	if len(*tones) != 3 {
		t.Fatalf("expected 3 tones, got %d", len(*tones))
	}
	// [0]=O4 A, [1]=O5 A, [2]=O4 A
	if math.Abs((*tones)[0].freq-(*tones)[2].freq) > 0.01 {
		t.Errorf("> then < should restore octave")
	}
	if math.Abs((*tones)[1].freq/(*tones)[0].freq-2.0) > 0.01 {
		t.Errorf("> should double frequency")
	}
}

func TestPlaySharpFlat(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	// C# should be higher than C; C- (Cb) should be lower.
	Play("O4 C C# C-")
	if len(*tones) != 3 {
		t.Fatalf("expected 3 tones, got %d", len(*tones))
	}
	c := (*tones)[0].freq
	cs := (*tones)[1].freq
	cb := (*tones)[2].freq
	if cs <= c {
		t.Errorf("C# (%.2f) should be > C (%.2f)", cs, c)
	}
	if cb >= c {
		t.Errorf("Cb (%.2f) should be < C (%.2f)", cb, c)
	}
}

func TestPlayTempoAndLength(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	// T120 L4 A: quarter note at 120 BPM = 0.5 s × musicStyle (7/8) ≈ 0.4375 s
	Play("T120 L4 A")
	if len(*tones) != 1 {
		t.Fatalf("expected 1 tone, got %d", len(*tones))
	}
	want := (60.0 / 120.0) * (4.0 / 4.0) * (7.0 / 8.0)
	if math.Abs((*tones)[0].dur-want) > 0.001 {
		t.Errorf("duration: got %.4f want %.4f", (*tones)[0].dur, want)
	}
}

func TestPlayPause(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()

	var slept time.Duration
	sleepFn = func(d time.Duration) { slept += d }

	// P4 = quarter rest at T120 = 0.5 s raw duration.
	Play("T120 P4")
	if len(*tones) != 0 {
		t.Errorf("pause should produce no tones, got %d", len(*tones))
	}
	if slept == 0 {
		t.Error("pause should sleep")
	}
}

func TestPlayDottedNote(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	Play("T120 L4 A A.")
	if len(*tones) != 2 {
		t.Fatalf("expected 2 tones, got %d", len(*tones))
	}
	// Dotted note should be 1.5x longer.
	ratio := (*tones)[1].dur / (*tones)[0].dur
	if math.Abs(ratio-1.5) > 0.001 {
		t.Errorf("dotted note should be 1.5x, got ratio %.4f", ratio)
	}
}

func TestPlayMusicStyle(t *testing.T) {
	tones, restore := installMockBackend(t)
	defer restore()

	origSleep := sleepFn
	defer func() { sleepFn = origSleep }()
	sleepFn = func(time.Duration) {}

	Play("T120 L4 MN A ML A MS A")
	if len(*tones) != 3 {
		t.Fatalf("expected 3 tones, got %d", len(*tones))
	}
	beat := 60.0 / 120.0 // 0.5 s
	wantMN := beat * (7.0 / 8.0)
	wantML := beat * 1.0
	wantMS := beat * (3.0 / 4.0)
	if math.Abs((*tones)[0].dur-wantMN) > 0.001 {
		t.Errorf("MN dur: got %.4f want %.4f", (*tones)[0].dur, wantMN)
	}
	if math.Abs((*tones)[1].dur-wantML) > 0.001 {
		t.Errorf("ML dur: got %.4f want %.4f", (*tones)[1].dur, wantML)
	}
	if math.Abs((*tones)[2].dur-wantMS) > 0.001 {
		t.Errorf("MS dur: got %.4f want %.4f", (*tones)[2].dur, wantMS)
	}
}
