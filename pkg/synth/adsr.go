package synth

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
)

const (
	SampleRate     = 44100
	ChannelCount   = 2
	BytesPerSample = 2 // 16-bit PCM
)

// ADSR Envelope
type ADSR struct {
	AttackTime  float64 // Em segundos
	DecayTime   float64 // Em segundos
	SustainLev  float64 // Nível de 0.0 a 1.0
	ReleaseTime float64 // Em segundos
}

// NoteFrequency calcula a frequência em Hz de uma nota MIDI n
// Fórmula: f = 440 * 2^((n - 69) / 12)
func NoteFrequency(midiNote int) float64 {
	return 440.0 * math.Pow(2.0, float64(midiNote-69)/12.0)
}

// GeneratePCMWithADSR gera amostras estéreo PCM de 16-bit com envelope ADSR
func GeneratePCMWithADSR(freqs []float64, duration time.Duration) []byte {
	numSamples := int(float64(SampleRate) * duration.Seconds())
	buf := new(bytes.Buffer)

	steps := make([]float64, len(freqs))
	for i, f := range freqs {
		steps[i] = 2 * math.Pi * f / float64(SampleRate)
	}

	adsr := ADSR{
		AttackTime:  0.05,
		DecayTime:   0.1,
		SustainLev:  0.7,
		ReleaseTime: 0.2,
	}

	totalTime := duration.Seconds()
	attackSamples := int(adsr.AttackTime * SampleRate)
	decaySamples := int(adsr.DecayTime * SampleRate)
	releaseSamples := int(adsr.ReleaseTime * SampleRate)

	for i := 0; i < numSamples; i++ {
		var mixed float64
		for _, step := range steps {
			mixed += math.Sin(step * float64(i))
		}
		mixed = mixed / float64(len(freqs))

		// Cálculo do envelope ADSR
		t := float64(i) / float64(SampleRate)
		var envelope float64

		if i < attackSamples {
			envelope = float64(i) / float64(attackSamples)
		} else if i < attackSamples+decaySamples {
			decayProgress := float64(i-attackSamples) / float64(decaySamples)
			envelope = 1.0 - (1.0-adsr.SustainLev)*decayProgress
		} else if t < totalTime-adsr.ReleaseTime {
			envelope = adsr.SustainLev
		} else {
			relProgress := float64(numSamples-i) / float64(releaseSamples)
			if relProgress < 0 {
				relProgress = 0
			}
			envelope = adsr.SustainLev * relProgress
		}

		sampleValue := int16(mixed * envelope * 0.3 * 32767.0)

		_ = binary.Write(buf, binary.LittleEndian, sampleValue)
		_ = binary.Write(buf, binary.LittleEndian, sampleValue)
	}

	return buf.Bytes()
}
