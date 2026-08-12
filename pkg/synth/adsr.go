package synth

import (
	"math"
)

const (
	SampleRate     = 44100
	ChannelCount   = 2
	BytesPerSample = 2 // 16-bit PCM
)

// NoteFrequency calcula a frequência em Hz de uma nota MIDI n
// Fórmula: f = 440 * 2^((n - 69) / 12)
func NoteFrequency(midiNote int) float64 {
	return 440.0 * math.Pow(2.0, float64(midiNote-69)/12.0)
}
