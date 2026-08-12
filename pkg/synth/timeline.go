package synth

import "pioke/pkg/model"

// RenderSong sintetiza a timeline completa de uma música em um único buffer PCM estéreo 16-bit,
// combinando a melodia (Syllables) e o acompanhamento de outras vozes (Accompaniment, ex: baixo
// de violão) quando presente.
//
// As duas linhas são renderizadas SEPARADAMENTE (cada uma com RenderSequence, que preserva fase
// contínua entre notas adjacentes da mesma linha) e depois somadas amostra a amostra com MixPCM —
// em vez de achatar as duas em uma lista única de segmentos com contagem de vozes variável, o que
// forçava reiniciar a fase de tudo (inclusive notas já soando) toda vez que o acompanhamento
// entrava ou saía, produzindo estalos audíveis a cada uma dessas trocas.
func RenderSong(s *model.Song) []byte {
	melody := RenderSequence(BuildSegments(s))

	accompaniment := buildAccompanimentSegments(s)
	if len(accompaniment) == 0 {
		return melody
	}

	return MixPCM(melody, RenderSequence(accompaniment))
}

// BuildSegments converte a melodia da timeline (Syllables, com fallback para a cifra do evento
// quando a sílaba não tem pitch) em uma sequência plana de segmentos de áudio, preenchendo
// silêncio nos intervalos entre eventos e entre sílabas. É determinística — não depende de
// temporização de reprodução em tempo real.
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

		if len(event.Syllables) > 0 {
			for _, syl := range event.Syllables {
				sylStart := event.TimeMS + syl.OffsetMS
				if sylStart > cursorMS {
					appendSilence(sylStart - cursorMS)
					cursorMS = sylStart
				}
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

				segs = append(segs, Segment{Freqs: freqs, DurationMS: dur})
				cursorMS += dur
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

// buildAccompanimentSegments espelha BuildSegments, mas para TimelineEvent.Accompaniment (outras
// vozes simultâneas à melodia, ex: baixo de violão) — sem fallback de cifra e sem letra, já que
// essa linha não carrega texto cantado.
func buildAccompanimentSegments(s *model.Song) []Segment {
	var segs []Segment
	var cursorMS int64

	appendSilence := func(durMS int64) {
		if durMS > 0 {
			segs = append(segs, Segment{DurationMS: durMS})
		}
	}

	for i := range s.Timeline {
		event := &s.Timeline[i]
		if len(event.Accompaniment) == 0 {
			continue
		}

		if event.TimeMS > cursorMS {
			appendSilence(event.TimeMS - cursorMS)
			cursorMS = event.TimeMS
		}

		for _, acc := range event.Accompaniment {
			accStart := event.TimeMS + acc.OffsetMS
			if accStart > cursorMS {
				appendSilence(accStart - cursorMS)
				cursorMS = accStart
			}
			dur := acc.DurationMS
			if dur <= 0 || acc.Pitch == "" {
				continue
			}
			freq, ok := NoteNameToFrequency(acc.Pitch)
			if !ok {
				continue
			}
			segs = append(segs, Segment{Freqs: []float64{freq}, DurationMS: dur})
			cursorMS += dur
		}
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
