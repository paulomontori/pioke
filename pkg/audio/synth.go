package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/oto/v3"
)

const (
	sampleRate      = 44100
	channelCount    = 2
	bytesPerSample  = 2 // 16-bit PCM
)

// Synth representa o sintetizador de áudio utilizando a biblioteca Oto v3
type Synth struct {
	context *oto.Context
	enabled bool
}

// NewSynth inicializa o contexto de áudio do Oto v3
func NewSynth() *Synth {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, ready, err := oto.NewContext(op)
	if err != nil {
		fmt.Printf("[AUDIO SYNTH] Erro ao criar contexto Oto: %v\n", err)
		return &Synth{enabled: false}
	}

	<-ready

	return &Synth{
		context: otoCtx,
		enabled: true,
	}
}

// PlayChord sintetiza o acorde em PCM e executa no contexto do Oto v3
func (s *Synth) PlayChord(chord string) {
	if !s.enabled || chord == "" {
		return
	}

	frequencies := getChordFrequencies(chord)
	if len(frequencies) == 0 {
		fmt.Printf("[AUDIO SYNTH] Executando acorde: %s (frequência não mapeada)\n", chord)
		return
	}

	fmt.Printf("[AUDIO SYNTH] Executando acorde polifônico via Oto v3: %s (%v Hz)\n", chord, frequencies)

	pcmData := generatePCMChord(frequencies, time.Millisecond*800)
	player := s.context.NewPlayer(bytes.NewReader(pcmData))
	player.Play()
}

// getChordFrequencies retorna as frequências fundamentais em Hz das notas de um acorde
func getChordFrequencies(chord string) []float64 {
	chord = strings.TrimSpace(chord)
	switch chord {
	case "C":
		return []float64{261.63, 329.63, 392.00} // C4, E4, G4
	case "D":
		return []float64{293.66, 369.99, 440.00} // D4, F#4, A4
	case "E":
		return []float64{329.63, 415.30, 493.88} // E4, G#4, B4
	case "Em":
		return []float64{329.63, 392.00, 493.88} // E4, G4, B4
	case "F":
		return []float64{349.23, 440.00, 523.25} // F4, A4, C5
	case "G":
		return []float64{392.00, 493.88, 587.33} // G4, B4, D5
	case "G#m":
		return []float64{415.30, 493.88, 622.25} // G#4, B4, D#5
	case "A":
		return []float64{440.00, 554.37, 659.25} // A4, C#5, E5
	case "Am":
		return []float64{440.00, 523.25, 659.25} // A4, C5, E5
	case "B":
		return []float64{493.88, 622.25, 739.99} // B4, D#5, F#5
	case "Bm":
		return []float64{493.88, 587.33, 739.99} // B4, D5, F#5
	case "F#7":
		return []float64{369.99, 466.16, 554.37, 659.25} // F#4, A#4, C#5, E5
	default:
		return []float64{440.00}
	}
}

// generatePCMChord gera amostras PCM de 16-bit Little Endian para o grupo de frequências
func generatePCMChord(freqs []float64, duration time.Duration) []byte {
	numSamples := int(float64(sampleRate) * duration.Seconds())
	buf := new(bytes.Buffer)

	steps := make([]float64, len(freqs))
	for i, f := range freqs {
		steps[i] = 2 * math.Pi * f / float64(sampleRate)
	}

	for i := 0; i < numSamples; i++ {
		var mixed float64
		for _, step := range steps {
			mixed += math.Sin(step * float64(i))
		}
		mixed = mixed / float64(len(freqs))

		// Application de fade-in e fade-out (envelope simples)
		fade := 1.0
		if i < 500 {
			fade = float64(i) / 500.0
		} else if i > numSamples-500 {
			fade = float64(numSamples-i) / 500.0
		}

		// Escala para PCM de 16-bits (Int16)
		sampleValue := int16(mixed * 0.3 * fade * 32767.0)

		// Canal esquerdo e direito (estéreo)
		_ = binary.Write(buf, binary.LittleEndian, sampleValue)
		_ = binary.Write(buf, binary.LittleEndian, sampleValue)
	}

	return buf.Bytes()
}
