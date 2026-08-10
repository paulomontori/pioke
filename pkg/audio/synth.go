package audio

import (
	"fmt"
)

// Synth representa o sintetizador de áudio/acordes
type Synth struct {
	enabled bool
}

// NewSynth cria uma nova instância de Synth
func NewSynth() *Synth {
	return &Synth{
		enabled: true,
	}
}

// PlayChord sintetiza/toca o acorde informado
func (s *Synth) PlayChord(chord string) {
	if !s.enabled || chord == "" {
		return
	}
	fmt.Printf("[AUDIO SYNTH] Executando acorde: %s\n", chord)
}
