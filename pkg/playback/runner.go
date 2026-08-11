// Package playback conduz a reprodução ao vivo de uma música (TUI + áudio) e, opcionalmente,
// grava o resultado sintetizado em um arquivo WAV. É a lógica compartilhada usada por todos os
// entrypoints de CLI do PioKe, para que não haja duas cópias divergentes do mesmo comportamento.
package playback

import (
	"fmt"
	"time"

	"pioke/pkg/audio"
	"pioke/pkg/engine"
	"pioke/pkg/model"
	"pioke/pkg/synth"
	"pioke/pkg/ui"
	"pioke/pkg/ui/tui"
)

// Run inicializa a TUI, o motor de reprodução e o sintetizador de áudio para a música informada.
// Bloqueia até o encerramento da TUI. Se outputFile não for vazio, sintetiza a música completa
// (incluindo as notas de melodia por sílaba, quando presentes) e grava em um arquivo WAV ao final.
func Run(s *model.Song, termUI *tui.TerminalUI, outputFile string) error {
	termUI.DisplayHeader(s)
	if err := termUI.Init(); err != nil {
		return fmt.Errorf("erro ao inicializar TUI: %w", err)
	}

	audioSynth := audio.NewSynth()
	eng := engine.NewEngine(s)
	eng.Play()

	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)

		var currentEvent *model.TimelineEvent
		currentSyllableIdx := -1

		for pbEvent := range eng.Events() {
			// Evento de nível superior mudou (novo acorde/linha, ou silêncio entre eventos)
			if pbEvent.ActiveEvent != currentEvent {
				currentEvent = pbEvent.ActiveEvent
				currentSyllableIdx = -1

				switch {
				case currentEvent == nil:
					audioSynth.PlayChord("")
				case len(currentEvent.Syllables) == 0:
					if chord := chordOf(currentEvent); chord != "" {
						dur := currentEvent.DurationMS
						if dur <= 0 {
							dur = 800
						}
						audioSynth.PlayChordFor(chord, time.Duration(dur)*time.Millisecond)
					}
				}
				// Eventos com sílabas: a primeira nota é disparada abaixo, no bloco de sílabas.
			}

			// Dentro do evento ativo, avança pelas sílabas (notas de melodia) conforme o tempo passa
			if currentEvent != nil && len(currentEvent.Syllables) > 0 {
				elapsedMS := pbEvent.CurrentTimeMS - currentEvent.TimeMS
				idx := activeSyllableIndex(currentEvent.Syllables, elapsedMS)
				if idx != currentSyllableIdx {
					currentSyllableIdx = idx
					if idx >= 0 {
						syl := currentEvent.Syllables[idx]
						dur := syl.DurationMS
						if dur <= 0 {
							dur = 200
						}

						if syl.Pitch != "" {
							audioSynth.PlayNote(syl.Pitch, time.Duration(dur)*time.Millisecond)
						} else if chord := chordOf(currentEvent); chord != "" {
							audioSynth.PlayChordFor(chord, time.Duration(dur)*time.Millisecond)
						}
					}
				}
			}

			_ = termUI.RenderTick(ui.PlaybackEvent{
				Song:     s,
				Current:  pbEvent.ActiveEvent,
				Position: pbEvent.CurrentTimeMS,
			})
		}
	}()

	go func() {
		<-doneChan
		termUI.Close()
	}()

	_ = termUI.Run()

	eng.Stop()

	if outputFile == "" {
		return nil
	}

	pcm := synth.RenderSequence(buildSegments(s))
	if len(pcm) == 0 {
		fmt.Println("\nNenhum áudio gerado para gravar.")
		return nil
	}

	if err := audio.WriteWAV(outputFile, pcm, synth.SampleRate, synth.ChannelCount); err != nil {
		return fmt.Errorf("erro ao salvar arquivo de áudio: %w", err)
	}
	fmt.Printf("\nÁudio gravado com sucesso em: %s\n", outputFile)
	return nil
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

// buildSegments converte a timeline da música em uma sequência plana de segmentos de áudio
// (notas/acordes/silêncios), na ordem cronológica, preenchendo silêncio nos intervalos entre eventos.
// É determinístico — não depende de temporização de reprodução em tempo real — por isso a gravação
// WAV fica sempre correta e sem cortes, independente de jitter do ticker ou do dispositivo de áudio.
func buildSegments(s *model.Song) []synth.Segment {
	var segs []synth.Segment
	var cursorMS int64

	appendSilence := func(durMS int64) {
		if durMS > 0 {
			segs = append(segs, synth.Segment{DurationMS: durMS})
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
					if freq, ok := synth.NoteNameToFrequency(syl.Pitch); ok {
						freqs = []float64{freq}
					}
				}
				if len(freqs) == 0 {
					freqs = synth.GetChordFrequencies(chordOf(event))
				}

				segs = append(segs, synth.Segment{Freqs: freqs, DurationMS: dur})
				cursorMS += dur
			}
			continue
		}

		dur := event.DurationMS
		if dur <= 0 {
			dur = 800
		}
		segs = append(segs, synth.Segment{Freqs: synth.GetChordFrequencies(chordOf(event)), DurationMS: dur})
		cursorMS += dur
	}

	return segs
}

// activeSyllableIndex retorna o índice da última sílaba cujo offset_ms já foi alcançado, ou -1
// se o tempo decorrido ainda não atingiu a primeira sílaba. Assume syllables ordenadas por offset_ms.
func activeSyllableIndex(syllables []model.Syllable, elapsedMS int64) int {
	idx := -1
	for i := range syllables {
		if syllables[i].OffsetMS <= elapsedMS {
			idx = i
		} else {
			break
		}
	}
	return idx
}
