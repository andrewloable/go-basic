package runtime

import (
	"io"
	"math"
	"sync"

	"github.com/ebitengine/oto/v3"
)

const (
	audioSampleRate = 44100
	audioBitDepth   = 2 // 16-bit signed PCM
	audioChannels   = 1 // mono
)

var (
	audioCtx     *oto.Context
	audioOnce    sync.Once
	audioInitErr error
)

// initAudio lazily initializes the oto audio context on first use.
func initAudio() error {
	audioOnce.Do(func() {
		op := &oto.NewContextOptions{
			SampleRate:   audioSampleRate,
			ChannelCount: audioChannels,
			Format:       oto.FormatSignedInt16LE,
		}
		var readyCh chan struct{}
		audioCtx, readyCh, audioInitErr = oto.NewContext(op)
		if audioInitErr == nil {
			<-readyCh // wait for the context to be ready
		}
	})
	return audioInitErr
}

// sineWavePCM generates 16-bit signed PCM samples for a sine wave tone.
func sineWavePCM(freqHz, durationSec float64) []byte {
	nSamples := int(float64(audioSampleRate) * durationSec)
	if nSamples < 1 {
		nSamples = 1
	}

	buf := make([]byte, nSamples*audioBitDepth)
	for i := 0; i < nSamples; i++ {
		t := float64(i) / float64(audioSampleRate)
		sample := math.Sin(2 * math.Pi * freqHz * t)

		// Apply a short fade-in/fade-out envelope to avoid clicks/pops.
		fadeSamples := audioSampleRate / 100 // ~10ms fade
		if fadeSamples > nSamples/2 {
			fadeSamples = nSamples / 2
		}
		if fadeSamples < 1 {
			fadeSamples = 1
		}
		if i < fadeSamples {
			sample *= float64(i) / float64(fadeSamples)
		} else if i > nSamples-fadeSamples {
			sample *= float64(nSamples-i) / float64(fadeSamples)
		}

		// Scale to 16-bit signed int range (-32768 to 32767).
		// Use 80% volume to avoid clipping.
		val := int16(sample * 0.8 * 32767)
		buf[i*2] = byte(val)
		buf[i*2+1] = byte(val >> 8)
	}
	return buf
}

// playTone plays a sine wave tone at the given frequency and duration.
func playTone(freqHz, durationSec float64) {
	if freqHz <= 0 || durationSec <= 0 {
		return
	}
	// Clamp frequency to Nyquist limit.
	if freqHz > float64(audioSampleRate)/2 {
		freqHz = float64(audioSampleRate) / 2
	}

	if err := initAudio(); err != nil {
		return // silently fail if audio not available
	}

	pcm := sineWavePCM(freqHz, durationSec)
	player := audioCtx.NewPlayer(&pcmReader{data: pcm})
	player.Play()

	// Wait for playback to finish.
	for player.IsPlaying() {
		// Busy-wait is fine for short tones; the BASIC program expects
		// Sound() to block for the duration anyway.
	}
}

// pcmReader wraps a byte slice as an io.Reader for oto.Player.
type pcmReader struct {
	data []byte
	pos  int
}

func (r *pcmReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		return n, io.EOF
	}
	return n, nil
}

func init() {
	// Wire the real audio backend into SoundBackend.
	SoundBackend = playTone
}
