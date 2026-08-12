package synth

import (
	"slices"

	"pioke/pkg/model"
)

// BuildSegments converte a timeline de uma música em uma sequência plana de segmentos de áudio
// (notas/acordes/silêncios), na ordem cronológica, preenchendo silêncio nos intervalos entre
// eventos. É determinístico — não depende de temporização de reprodução em tempo real — por isso
// tanto a gravação WAV quanto a reprodução ao vivo (que toca o PCM renderizado a partir daqui)
// ficam sempre corretas e idênticas entre si.
//
// Dentro de cada evento, a melodia (Syllables) e o acompanhamento de outras vozes
// (Accompaniment, ex: baixo de violão) são combinados por interseção de intervalos: em qualquer
// instante, o segmento soa a união das frequências de tudo que está ativo naquele ponto — uma
// nota da melodia sustentando enquanto o baixo troca de nota, por exemplo, vira um único segmento
// com as duas frequências, sem recortar a nota da melodia ao meio.
func BuildSegments(s *model.Song) []Segment {
	var segs []Segment
	var cursorMS int64

	appendSilence := func(durMS int64) {
		if durMS > 0 {
			segs = append(segs, Segment{DurationMS: durMS})
		}
	}

	for i := range s.Timeline {
		event := &s.Timeline[i]

		if event.TimeMS > cursorMS {
			appendSilence(event.TimeMS - cursorMS)
			cursorMS = event.TimeMS
		}

		if len(event.Syllables) > 0 || len(event.Accompaniment) > 0 {
			notes := eventNoteIntervals(event)
			for _, seg := range flattenPolyphonic(notes) {
				if len(seg.Freqs) == 0 {
					appendSilence(seg.DurationMS)
				} else {
					segs = append(segs, seg)
				}
				cursorMS += seg.DurationMS
			}
			continue
		}

		dur := event.DurationMS
		if dur <= 0 {
			dur = 800
		}
		segs = append(segs, Segment{Freqs: GetChordFrequencies(chordOf(event)), DurationMS: dur})
		cursorMS += dur
	}

	return segs
}

func chordOf(ev *model.TimelineEvent) string {
	if ev == nil {
		return ""
	}
	if ev.ChordStr != "" {
		return ev.ChordStr
	}
	if ev.Chord != nil {
		return ev.Chord.Name
	}
	return ""
}

// noteInterval é uma nota soando entre [startMS, endMS), relativo ao início do evento, já
// resolvida em frequências (uma só para uma nota de melodia/acompanhamento; várias quando um
// "syllable" sem pitch cai de volta para a cifra do evento).
type noteInterval struct {
	startMS int64
	endMS   int64
	freqs   []float64
}

// eventNoteIntervals resolve Syllables (melodia, com fallback para a cifra do evento quando a
// sílaba não tem pitch) e Accompaniment (outras vozes) em intervalos de frequência prontos para
// flattenPolyphonic.
func eventNoteIntervals(event *model.TimelineEvent) []noteInterval {
	var notes []noteInterval

	for _, syl := range event.Syllables {
		dur := syl.DurationMS
		if dur <= 0 {
			continue
		}
		var freqs []float64
		if syl.Pitch != "" {
			if freq, ok := NoteNameToFrequency(syl.Pitch); ok {
				freqs = []float64{freq}
			}
		}
		if len(freqs) == 0 {
			freqs = GetChordFrequencies(chordOf(event))
		}
		if len(freqs) == 0 {
			continue // sem pitch e sem cifra: intervalo silencioso, não precisa de nota explícita
		}
		notes = append(notes, noteInterval{startMS: syl.OffsetMS, endMS: syl.OffsetMS + dur, freqs: freqs})
	}

	for _, acc := range event.Accompaniment {
		dur := acc.DurationMS
		if dur <= 0 || acc.Pitch == "" {
			continue
		}
		freq, ok := NoteNameToFrequency(acc.Pitch)
		if !ok {
			continue
		}
		notes = append(notes, noteInterval{startMS: acc.OffsetMS, endMS: acc.OffsetMS + dur, freqs: []float64{freq}})
	}

	return notes
}

// flattenPolyphonic converte uma lista de notas possivelmente sobrepostas em uma sequência plana
// e mínima de Segments: em cada sub-intervalo entre dois pontos de início/fim de nota, o segmento
// soa a união das frequências de toda nota ativa ali. Sub-intervalos adjacentes com exatamente o
// mesmo conjunto de frequências são mesclados em um só segmento, para não recortar (e reatacar) a
// envoltória sem necessidade quando nada realmente mudou.
func flattenPolyphonic(notes []noteInterval) []Segment {
	if len(notes) == 0 {
		return nil
	}

	var totalMS int64
	boundarySet := map[int64]struct{}{0: {}}
	for _, n := range notes {
		if n.endMS > totalMS {
			totalMS = n.endMS
		}
		boundarySet[n.startMS] = struct{}{}
		boundarySet[n.endMS] = struct{}{}
	}
	boundarySet[totalMS] = struct{}{}

	bounds := make([]int64, 0, len(boundarySet))
	for b := range boundarySet {
		bounds = append(bounds, b)
	}
	slices.Sort(bounds)

	var segs []Segment
	for i := 0; i+1 < len(bounds); i++ {
		start, end := bounds[i], bounds[i+1]
		if end <= start {
			continue
		}

		var freqs []float64
		for _, n := range notes {
			if n.startMS <= start && n.endMS >= end {
				freqs = append(freqs, n.freqs...)
			}
		}

		if len(segs) > 0 && sameFreqs(segs[len(segs)-1].Freqs, freqs) {
			segs[len(segs)-1].DurationMS += end - start
			continue
		}
		segs = append(segs, Segment{Freqs: freqs, DurationMS: end - start})
	}

	return segs
}

func sameFreqs(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
