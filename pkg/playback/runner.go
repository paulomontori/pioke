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

// Start inicia a reprodução de áudio + motor de sincronização (e, opcionalmente, a captura do
// microfone) sem bloquear: os eventos do motor são encaminhados para renderer.RenderTick em uma
// goroutine própria, então chamadas concorrentes ao renderer (ex: uma tela de GUI redesenhando a
// ~60fps) veem o estado mais recente sem precisar rodar sua própria máquina de estados. O canal
// done é fechado quando a reprodução termina por conta própria (fim da música ou stop() já
// chamado); stop() para tudo (motor, áudio, gravação) e deve ser chamada mesmo se done já tiver
// fechado sozinho, para garantir que a gravação (se houver) seja salva.
func Start(s *model.Song, renderer ui.Renderer, timbre synth.Timbre, record bool, songFile string) (pcm []byte, done <-chan struct{}, stop func()) {
	// A reprodução ao vivo toca o mesmo buffer PCM pré-renderizado que seria gravado no WAV, em
	// vez de sintetizar nota a nota em tempo real — assim ela soa exatamente igual ao arquivo
	// exportado, sem depender de um agendador em tempo real (sujeito a atraso de GC, troca de
	// goroutine, jitter do SO) para acertar cada troca de nota.
	pcm = synth.RenderSongWithTimbre(s, timbre)

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
			_ = renderer.RenderTick(ui.PlaybackEvent{
				Song:     s,
				Current:  pbEvent.ActiveEvent,
				Position: pbEvent.CurrentTimeMS,
			})
		}
	}()

	stop = func() {
		eng.Stop()
		audioSynth.Stop()
		finishRecording(micRecorder, recordStart, playbackStart, s, songFile, timbre)
	}
	return pcm, doneChan, stop
}

// Run inicializa a TUI e a reprodução (ver Start) para a música informada, sintetizando com o
// timbre escolhido (ver synth.Timbre). Bloqueia até o encerramento da TUI. Se outputFile não for
// vazio, grava o mesmo áudio sintetizado em um arquivo WAV ao final. Se record for true, também
// captura o microfone durante a reprodução e salva em recordings/ (ver RecordingMeta) — pensado
// pra permitir avaliar depois a qualidade do canto de quem usou o PioKe.
func Run(s *model.Song, termUI *tui.TerminalUI, outputFile string, timbre synth.Timbre, record bool, songFile string) error {
	termUI.DisplayHeader(s)
	if err := termUI.Init(); err != nil {
		return fmt.Errorf("erro ao inicializar TUI: %w", err)
	}

	pcm, done, stop := Start(s, termUI, timbre, record, songFile)

	go func() {
		<-done
		termUI.Close()
	}()

	_ = termUI.Run()

	stop()

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
