package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"pioke/pkg/audio"
	"pioke/pkg/engine"
	"pioke/pkg/parser"
	"pioke/pkg/synth"
)

func main() {
	filePathFlag := flag.String("file", "", "Caminho do arquivo de música em formato JSON")
	noAudioFlag := flag.Bool("no-audio", false "Desabilita a inicialização e saída do motor de áudio")
	flag.Parse()

	filePath := *filePathFlag
	if filePath == "" {
		if flag.NArg() > 0 {
			filePath = flag.Arg(0)
		} else {
			filePath = filepath.Join("songs", "evidencias.json")
		}
	}

	fmt.Printf("Carregando música: %s\n", filePath)
	song, err := parser.LoadJSON(filePath)
	if err != nil {
		log.Fatalf("Erro ao carregar música: %v\n", err)
	}

	fmt.Println("========================================")
	fmt.Printf("   Música: %s\n", song.Title)
	fmt.Printf("   Artista: %s\n", song.Artist)
	if song.BPM > 0 {
		fmt.Printf("   BPM: %d | Tom: %s\n", song.BPM, song.Key)
	}
	fmt.Println("========================================\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trata sinais do sistema (Ctrl+C) para desligamento gracioso
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nEncerrando reprodução...")
		cancel()
	}()

	polySynth := synth.NewPolySynth(audio.DefaultSampleRate)

	if !*noAudioFlag {
		player := audio.NewAudioPlayer(polySynth)
		// Em um ambiente com driver de som real, PCMStream enviaria para as caixas de áudio
		engAudio := audio.NewPCMStream(polySynth)
		_ = engAudio

		eng := engine.NewEngine(song)
		go player.Listen(ctx, eng.Events())

		runCLIPlayback(ctx, eng, song)
	} else {
		eng := engine.NewEngine(song)
		runCLIPlayback(ctx, eng, song)
	}

	fmt.Println("\nExecução concluída.")
}

func runCLIPlayback(ctx context.Context, eng *engine.Engine, song interface{}) {
	eng.Play()
	events := eng.Events()

	lastLyric := ""
	lastChord := ""

	for {
		select {
		case <-ctx.Done():
			eng.Stop()
			return
		case ev, ok := <-events:
			if !ok {
				return
			}

			if ev.ActiveEvent != nil {
				active := ev.ActiveEvent

				if active.ChordStr != "" && active.ChordStr != lastChord {
					lastChord = active.ChordStr
					fmt.Printf("\n[Acorde: %s]\n", lastChord)
				}

				if active.Lyric != "" && active.Lyric != lastLyric {
					lastLyric = active.Lyric
					secs := float64(ev.CurrentTimeMS) / 1000.0
					fmt.Printf("[%02d:%05.2f] Letra: %s\n", int(secs)/60, float64(int(secs)%60)+secs-float64(int(secs)), lastLyric)
				}
			}

			// Aguarda uma pequena pausa se a música tiver chegado ao fim da timeline
			if eng.State() == engine.STOPPED {
				return
			}
		case <-time.After(500 * time.Millisecond):
			if eng.State() != engine.PLAYING {
				return
			}
		}
	}
}
