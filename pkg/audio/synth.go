package audio

import (
	"fmt"
	"math"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// Synth representa o sintetizador de áudio/acordes usando a biblioteca Beep
type Synth struct {
	enabled    bool
	sampleRate beep.SampleRate
}

// NewSynth cria uma nova instância de Synth e inicializa o dispositivo de áudio
func NewSynth() *Synth {
	sr := beep.SampleRate(44100)
	// Inicializa o alto-falante com taxa de amostragem de 44.1kHz e buffer ajustado
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

// PlayChord sintetiza/toca o acorde informado gerando uma onda senoidal simples
func (s *Synth) PlayChord(chord string) {
	if !s.enabled || chord == "" {
		return
	}

	freq := getChordFrequency(chord)
	if freq == 0 {
		fmt.Printf("[AUDIO SYNTH] Executando acorde: %s (frequência não mapeada)\n", chord)
		return
	}

	fmt.Printf("[AUDIO SYNTH] Executando acorde: %s (%.2f Hz)\n", chord, freq)

	// Gera tom de onda senoidal com duração de 500ms
	sine := sineWave(s.sampleRate, freq, time.Millisecond*500)
	speaker.Play(sine)
}

// getChordFrequency mapeia nomes de acordes básicos para frequências fundamentais em Hz
func getChordFrequency(chord string) float64 {
	switch chord {
	case "C":
		return 261.63 // C4
	case "D":
		return 293.66 // D4
	case "E":
		return 329.63 // E4
	case "F":
		return 349.23 // F4
	case "G":
		return 392.00 // G4
	case "A", "Am":
		return 440.00 // A4
	case "B", "Bm":
		return 493.88 // B4
	case "G#m":
		return 415.30 // G#4 / Ab4
	case "F#7":
		return 369.99 // F#4 / Gb4
	default:
		return 440.00 // Frequência padrão A4
	}
}

// sineWave produz um Streamer de onda senoidal com envelope básico
func sineWave(sr beep.SampleRate, freq float64, duration time.Duration) beep.Streamer {
	length := sr.N(duration)
	step := 2 * math.Pi * freq / float64(sr)
	i := 0

	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for j := range samples {
			if i >= length {
				return j, false
			}
			val := math.Sin(step * float64(i))
			// Envelope suave para evitar estalos
			fade := 1.0
			if i < 100 {
				fade = float64(i) / 100.0
			} else if i > length-100 {
				fade = float64(length-i) / 100.0
			}

			samples[j][0] = val * 0.3 * fade
			samples[j][1] = val * 0.3 * fade
			i++
		}
		return len(samples), true
	})
}
