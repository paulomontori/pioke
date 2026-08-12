package audio

import (
	"fmt"
	"time"

	"pioke/pkg/synth"

	"github.com/ebitengine/oto/v3"
)

// Synth representa a interface de saída de áudio ao vivo: um único oto.Player de longa duração
// alimentado por liveVoice, em vez de um novo player para cada nota — é isso que evita o
// "picotado" ao trocar de nota rapidamente (ex: uma linha de melodia importada de MusicXML).
type Synth struct {
	voice   *liveVoice
	enabled bool
}

// NewSynth inicializa o contexto de áudio do Oto v3 e já inicia o player contínuo (em silêncio
// até a primeira chamada a PlayChord/PlayNote).
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

	voice := newLiveVoice(synth.SampleRate)
	player := otoCtx.NewPlayer(voice)
	player.Play()

	return &Synth{voice: voice, enabled: true}
}

// PlayChord toca o acorde (ou silencia, se chord == "") continuamente até a próxima chamada.
func (s *Synth) PlayChord(chord string) {
	if !s.enabled {
		return
	}
	if chord == "" {
		s.voice.SetFreqs(nil)
		return
	}
	s.voice.SetFreqs(synth.GetChordFrequencies(chord))
}

// PlayChordFor toca o acorde continuamente; duration é ignorado — quem decide até quando o
// acorde soa é a próxima chamada a PlayChord/PlayChordFor/PlayNote/PlayChord(""), disparada
// pelo laço de reprodução (pkg/playback) no instante certo da timeline. O parâmetro existe só
// para manter a assinatura estável nos chamadores existentes.
func (s *Synth) PlayChordFor(chord string, _ time.Duration) {
	s.PlayChord(chord)
}

// PlayNote toca uma única nota de melodia (ex: "G4", "C#5") continuamente — usado para
// reproduzir as notas de syllables[].pitch em sequência. duration é ignorado pelo mesmo motivo
// descrito em PlayChordFor.
func (s *Synth) PlayNote(noteName string, _ time.Duration) {
	if !s.enabled || noteName == "" {
		return
	}
	freq, ok := synth.NoteNameToFrequency(noteName)
	if !ok {
		return
	}
	s.voice.SetFreqs([]float64{freq})
}
