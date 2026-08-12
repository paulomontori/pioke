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

// Run inicializa a TUI, o motor de reprodução e o sintetizador de áudio para a música informada,
// sintetizando com o timbre escolhido (ver synth.Timbre). Bloqueia até o encerramento da TUI. Se
// outputFile não for vazio, grava o mesmo áudio sintetizado em um arquivo WAV ao final. Se record
// for true, também captura o microfone durante a reprodução e salva em recordings/ (ver
// RecordingMeta) — pensado pra permitir avaliar depois a qualidade do canto de quem usou o PioKe.
func Run(s *model.Song, termUI *tui.TerminalUI, outputFile string, timbre synth.Timbre, record bool, songFile string) error {
	termUI.DisplayHeader(s)
	if err := termUI.Init(); err != nil {
		return fmt.Errorf("erro ao inicializar TUI: %w", err)
	}

	// A reprodução ao vivo toca o mesmo buffer PCM pré-renderizado que seria gravado no WAV, em
	// vez de sintetizar nota a nota em tempo real — assim ela soa exatamente igual ao arquivo
	// exportado, sem depender de um agendador em tempo real (sujeito a atraso de GC, troca de
	// goroutine, jitter do SO) para acertar cada troca de nota.
	pcm := synth.RenderSongWithTimbre(s, timbre)

	// A captura do microfone começa antes da reprodução: o offset real entre os dois instantes
	// (que pode variar por causa da inicialização do dispositivo de captura) é medido e salvo,
	// em vez de assumir que os dois começam exatamente juntos.
	micRecorder, recordStart := startRecording(record)

	audioSynth := audio.NewSynth()
	playbackStart := time.Now()
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
	finishRecording(micRecorder, recordStart, playbackStart, s, songFile, timbre)

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
