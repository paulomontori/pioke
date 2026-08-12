// Package playback conduz a reprodução ao vivo de uma música (TUI + áudio) e, opcionalmente,
// grava o resultado sintetizado em um arquivo WAV. É a lógica compartilhada usada por todos os
// entrypoints de CLI do PioKe, para que não haja duas cópias divergentes do mesmo comportamento.
package playback

import (
	"fmt"

	"pioke/pkg/audio"
	"pioke/pkg/engine"
	"pioke/pkg/model"
	"pioke/pkg/synth"
	"pioke/pkg/ui"
	"pioke/pkg/ui/tui"
)

// Run inicializa a TUI, o motor de reprodução e o sintetizador de áudio para a música informada.
// Bloqueia até o encerramento da TUI. Se outputFile não for vazio, grava o mesmo áudio sintetizado
// em um arquivo WAV ao final.
func Run(s *model.Song, termUI *tui.TerminalUI, outputFile string) error {
	termUI.DisplayHeader(s)
	if err := termUI.Init(); err != nil {
		return fmt.Errorf("erro ao inicializar TUI: %w", err)
	}

	// A reprodução ao vivo toca o mesmo buffer PCM pré-renderizado que seria gravado no WAV, em
	// vez de sintetizar nota a nota em tempo real — assim ela soa exatamente igual ao arquivo
	// exportado, sem depender de um agendador em tempo real (sujeito a atraso de GC, troca de
	// goroutine, jitter do SO) para acertar cada troca de nota.
	pcm := synth.RenderSequence(buildSegments(s))

	audioSynth := audio.NewSynth()
	audioSynth.Play(pcm)

	eng := engine.NewEngine(s)
	eng.Play()

	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		for pbEvent := range eng.Events() {
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
	audioSynth.Stop()

	if outputFile == "" {
		return nil
	}

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
