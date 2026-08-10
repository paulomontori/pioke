package audio

import (
	"bytes"
	"fmt"
	"time"

	"pioke/pkg/synth"

	"github.com/ebitengine/oto/v3"
)

// Synth representa a interface de saída de áudio alimentada pelo gerador de síntese
type Synth struct {
	context *oto.Context
	enabled bool
}

// NewSynth inicializa o contexto de áudio do Oto v3
func NewSynth() *Synth {
	op := &oto.NewContextOptions{
		SampleRate:   synth.SampleRate,
		ChannelCount: synth.ChannelCount,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, ready, err := oto.NewContext(op)
	if err != nil {
		fmt.Printf("[AUDIO ENGINE] Aviso: Dispositivo de áudio indisponível: %v\n", err)
		return &Synth{enabled: false}
	}

	<-ready

	return &Synth{
		context: otoCtx,
		enabled: true,
	}
}

// PlayChord sintetiza o acorde com envelope ADSR e executa na placa de som
func (s *Synth) PlayChord(chord string) {
	if !s.enabled || chord == "" {
		return
	}

	frequencies := synth.GetChordFrequencies(chord)
	if len(frequencies) == 0 {
		return
	}

	pcmData := synth.GeneratePCMWithADSR(frequencies, time.Millisecond*800)
	player := s.context.NewPlayer(bytes.NewReader(pcmData))
	player.Play()
}
