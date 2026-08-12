// Package scoring avalia a qualidade de uma gravação de voz (microfone) contra a melodia de
// referência de uma música do PioKe — pontuação de afinação, ritmo e cobertura. Não depende do
// pacote playback nem altera o pipeline de reprodução do guia vocal; consome apenas model.Song
// (já carregado por pkg/parser/pkg/song) e o PCM gravado.
package scoring

import (
	"pioke/pkg/model"
	"pioke/pkg/synth"
)

// ReferenceNote é uma nota da melodia vocal de referência, com tempo absoluto (relativo ao início
// da música, não ao evento da timeline que a contém).
type ReferenceNote struct {
	StartMS    int64
	DurationMS int64
	Pitch      string  // nome da nota, ex: "C4"
	FreqHz     float64
	Lyric      string  // sílaba cantada nesta nota, útil pra UI mostrar qual trecho é este
}

func (n ReferenceNote) EndMS() int64 { return n.StartMS + n.DurationMS }

// ExtractReferenceNotes achata a melodia vocal da timeline (TimelineEvent.Syllables, já com notas
// ligadas por tie mescladas pelo parser — ver pkg/parser/musicxml.go) em uma sequência plana de
// notas com tempo absoluto. É a mesma fonte de dados (pitch + offset + duração por sílaba) que
// synth.BuildSegments usa pra sintetizar o guia vocal — aqui reaproveitada para comparação com o
// canto gravado, em vez de duplicar a extração num parser próprio do pacote scoring.
func ExtractReferenceNotes(s *model.Song) []ReferenceNote {
	var notes []ReferenceNote
	for i := range s.Timeline {
		event := &s.Timeline[i]
		for _, syl := range event.Syllables {
			if syl.Pitch == "" || syl.DurationMS <= 0 {
				continue
			}
			freq, ok := synth.NoteNameToFrequency(syl.Pitch)
			if !ok {
				continue
			}
			notes = append(notes, ReferenceNote{
				StartMS:    event.TimeMS + syl.OffsetMS,
				DurationMS: syl.DurationMS,
				Pitch:      syl.Pitch,
				FreqHz:     freq,
				Lyric:      syl.Text,
			})
		}
	}
	return notes
}
