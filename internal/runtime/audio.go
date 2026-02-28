package runtime

import (
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// SoundBackend is the function called by Sound to produce a tone. It defaults
// to a no-op stub. Programs or tests can replace it with a real audio driver.
var SoundBackend func(freqHz float64, durationSec float64) = func(float64, float64) {}

// sleepFn is the sleep implementation. Overridden in tests to avoid blocking.
var sleepFn = time.Sleep

// Sound generates a tone at freqHz for a duration measured in 18.2-tick-per-
// second BASIC clock ticks. The default backend is a no-op; replace SoundBackend
// with a real audio driver if needed.
func Sound(freq, duration float64) {
	if freq <= 0 || duration <= 0 {
		return
	}
	durationSec := duration / 18.2
	SoundBackend(freq, durationSec)
	// Sleep to simulate blocking behavior expected by BASIC programs.
	sleepFn(time.Duration(durationSec * float64(time.Second)))
}

// noteFrequencies maps note names (C, D, E, F, G, A, B) to their base frequency
// in octave 4 (middle octave) in Hz.
var noteFrequencies = map[byte]float64{
	'C': 261.63,
	'D': 293.66,
	'E': 329.63,
	'F': 349.23,
	'G': 392.00,
	'A': 440.00,
	'B': 493.88,
}

// halfStepsAboveC maps each natural note to the number of half-steps above C.
var halfStepsAboveC = map[byte]int{
	'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11,
}

// mmlNoteFreq computes the frequency for note (A-G) with sharps/flats at octave.
func mmlNoteFreq(note byte, sharps int, octave int) float64 {
	// Half-step offset from middle C (octave 4, C=0).
	semitone := halfStepsAboveC[note] + sharps + (octave-4)*12
	// A4 = 440 Hz; semitones relative to A4.
	a4Semitone := halfStepsAboveC['A'] + (4-4)*12 // = 9 semitones above C4
	relToA4 := float64(semitone - a4Semitone)
	return 440.0 * math.Pow(2, relToA4/12.0)
}

// mmlState holds the current state of the MML interpreter.
type mmlState struct {
	octave      int     // current octave (0-6, default 4)
	length      int     // default note length (1-64, default 4 = quarter)
	tempo       int     // quarter notes per minute (32-255, default 120)
	musicStyle  float64 // note duration fraction (MN=7/8, ML=1, MS=3/4)
}

// newMMLState returns a default MML state.
func newMMLState() mmlState {
	return mmlState{octave: 4, length: 4, tempo: 120, musicStyle: 7.0 / 8.0}
}

// noteDurationSec computes note duration in seconds from length (1,2,4,8,16,32,64)
// and whether it is dotted.
func (s mmlState) noteDurationSec(length int, dotted bool) float64 {
	// One beat = 60/tempo seconds. length=4 means quarter note = 1 beat.
	beatSec := 60.0 / float64(s.tempo)
	dur := beatSec * 4.0 / float64(length)
	if dotted {
		dur *= 1.5
	}
	return dur * s.musicStyle
}

// Play interprets a BASIC MML (Music Macro Language) command string and
// generates tones via Sound(). Commands are case-insensitive.
//
// Supported commands:
//
//	A-G           Play note (optional: # or + for sharp, - for flat,
//	              digit(s) for note length, dot for dotted note)
//	O n           Set octave (0-6)
//	>             Increase octave by 1
//	<             Decrease octave by 1
//	L n           Set default note length (1-64)
//	T n           Set tempo in BPM (32-255)
//	P n           Pause/rest for length n
//	MN            Music normal (7/8 of beat)
//	ML            Music legato (full beat)
//	MS            Music staccato (3/4 of beat)
//	MF            Music foreground (ignored — always synchronous)
//	MB            Music background (ignored — always synchronous)
func Play(cmd string) {
	s := newMMLState()
	upper := strings.ToUpper(cmd)
	i := 0
	n := len(upper)

	readInt := func() (int, bool) {
		j := i
		for j < n && unicode.IsDigit(rune(upper[j])) {
			j++
		}
		if j == i {
			return 0, false
		}
		v, err := strconv.Atoi(upper[i:j])
		if err != nil {
			return 0, false
		}
		i = j
		return v, true
	}

	for i < n {
		ch := upper[i]
		switch {
		case ch == ' ' || ch == ',':
			i++

		case ch == 'O':
			i++
			if v, ok := readInt(); ok && v >= 0 && v <= 6 {
				s.octave = v
			}

		case ch == '>':
			i++
			if s.octave < 6 {
				s.octave++
			}

		case ch == '<':
			i++
			if s.octave > 0 {
				s.octave--
			}

		case ch == 'L':
			i++
			if v, ok := readInt(); ok && v >= 1 && v <= 64 {
				s.length = v
			}

		case ch == 'T':
			i++
			if v, ok := readInt(); ok && v >= 32 && v <= 255 {
				s.tempo = v
			}

		case ch == 'M':
			i++
			if i < n {
				switch upper[i] {
				case 'N':
					s.musicStyle = 7.0 / 8.0
					i++
				case 'L':
					s.musicStyle = 1.0
					i++
				case 'S':
					s.musicStyle = 3.0 / 4.0
					i++
				case 'F', 'B':
					i++ // ignore foreground/background — always synchronous
				}
			}

		case ch == 'P':
			// Pause/rest.
			i++
			length := s.length
			if v, ok := readInt(); ok && v >= 1 {
				length = v
			}
			dotted := i < n && upper[i] == '.'
			if dotted {
				i++
			}
			dur := s.noteDurationSec(length, dotted)
			sleepFn(time.Duration(dur * float64(time.Second)))

		case ch >= 'A' && ch <= 'G':
			note := ch
			i++
			// Accidentals.
			sharps := 0
			for i < n && (upper[i] == '#' || upper[i] == '+') {
				sharps++
				i++
			}
			for i < n && upper[i] == '-' {
				sharps--
				i++
			}
			// Note length.
			length := s.length
			if v, ok := readInt(); ok && v >= 1 {
				length = v
			}
			dotted := i < n && upper[i] == '.'
			if dotted {
				i++
			}
			freq := mmlNoteFreq(note, sharps, s.octave)
			dur := s.noteDurationSec(length, dotted)
			SoundBackend(freq, dur)
			sleepFn(time.Duration(dur * float64(time.Second)))

		default:
			i++ // skip unknown character
		}
	}
}
