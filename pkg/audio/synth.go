package audio

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// Synth representa o sintetizador de áudio/acordes com suporte a mixagem de amostras PCM
type Synth struct {
	enabled    bool
	sampleRate beep.SampleRate
}

// NewSynth cria uma nova instância de Synth e inicializa o dispositivo de áudio
func NewSynth() *Synth {
	sr := beep.SampleRate(44100)
	err := speaker.Init(sr, sr.N(time.Second/10))
	if err != nil {
		fmt.Printf("[AUDIO SYNTH] Erro ao inicializar o alto-falante: %v\n", err)
		return &Synth{enabled: false, sampleRate: sr}
	}

	return &Synth{
		enabled:    true,
		sampleRate: sr,
	}
}

// PlayChord sintetiza e executa o acorde em tempo real usando mixagem de frequências
func (s *Synth) PlayChord(chord string) {
	if !s.enabled || chord == "" {
		return
	}

	frequencies := getChordFrequencies(chord)
	if len(frequencies) == 0 {
		fmt.Printf("[AUDIO SYNTH] Executando acorde: %s (frequência não mapeada)\n", chord)
		return
	}

	fmt.Printf("[AUDIO SYNTH] Executando acorde polifônico: %s (%v Hz)\n", chord, frequencies)

	// Gera a mixagem de ondas senoidais para todas as notas do acorde
	streamer := multiToneWave(s.sampleRate, frequencies, time.Millisecond*800)
	speaker.Play(streamer)
}

// getChordFrequencies converte o nome do acorde nas frequências das suas notas componentes (Hz)
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
		return []float64{440.00} // Tom padrão (A4)
	}
}

// multiToneWave mixa múltiplas frequências no buffer PCM
func multiToneWave(sr beep.SampleRate, freqs []float64, duration time.Duration) beep.Streamer {
	length := sr.N(duration)
	steps := make([]float64, len(freqs))
	for i, f := range freqs {
		steps[i] = 2 * math.Pi * f / float64(sr)
	}
	i := 0

	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for j := range samples {
			if i >= length {
				return j, false
			}

			var mixedSample float64
			for _, step := range steps {
				mixedSample += math.Sin(step * float64(i))
			}
			mixedSample = mixedSample / float64(len(freqs)) // Normalização do volume

			// Envelope suave de ataque e decaimento (Fade In / Fade Out)
			fade := 1.0
			if i < 200 {
				fade = float64(i) / 200.0
			} else if i > length-200 {
				fade = float64(length-i) / 200.0
			}

			finalVal := mixedSample * 0.3 * fade
			samples[j][0] = finalVal
			samples[j][1] = finalVal
			i++
		}
		return len(samples), true
	})
}
