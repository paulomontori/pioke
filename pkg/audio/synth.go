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

// PlayChord sintetiza o acorde com envelope ADSR (duração padrão de 800ms) e executa na placa de som
func (s *Synth) PlayChord(chord string) {
	s.PlayChordFor(chord, 800*time.Millisecond)
}

// PlayChordFor sintetiza o acorde com envelope ADSR pela duração informada e executa na placa de som
func (s *Synth) PlayChordFor(chord string, duration time.Duration) {
	if !s.enabled || chord == "" || duration <= 0 {
		return
	}

	frequencies := synth.GetChordFrequencies(chord)
	if len(frequencies) == 0 {
		return
	}

	s.play(frequencies, duration)
}

// PlayNote sintetiza uma única nota de melodia (ex: "G4", "C#5") pela duração informada — usado para
// reproduzir as notas de syllables[].pitch em sequência
func (s *Synth) PlayNote(noteName string, duration time.Duration) {
	if !s.enabled || noteName == "" || duration <= 0 {
		return
	}

	freq, ok := synth.NoteNameToFrequency(noteName)
	if !ok {
		return
	}

	s.play([]float64{freq}, duration)
}

func (s *Synth) play(frequencies []float64, duration time.Duration) {
	pcmData := synth.GeneratePCMWithADSR(frequencies, duration)
	player := s.context.NewPlayer(bytes.NewReader(pcmData))
	player.Play()
}
