package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"pioke/pkg/audio"
	"pioke/pkg/engine"
	"pioke/pkg/song"
	"pioke/pkg/synth"
	"pioke/pkg/ui"
	"pioke/pkg/ui/tui"
)

func main() {
	var outputFile string
	flag.StringVar(&outputFile, "out", "", "Caminho do arquivo WAV de saída para salvar o áudio gerado (ex: output.wav)")
	flag.Parse()

	// Fallback para capturar a flag -out caso tenha sido informada após argumentos posicionais
	if outputFile == "" {
		for i, arg := range os.Args {
			if (arg == "-out" || arg == "--out") && i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				break
			} else if strings.HasPrefix(arg, "-out=") {
				outputFile = strings.TrimPrefix(arg, "-out=")
				break
			} else if strings.HasPrefix(arg, "--out=") {
				outputFile = strings.TrimPrefix(arg, "--out=")
				break
			}
		}
	}

	sampleFile := "songs/parabens.yaml"
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") && arg != outputFile {
			sampleFile = arg
			break
		}
	}

	s, err := song.LoadSong(sampleFile)
	if err != nil {
		log.Fatalf("Erro ao carregar a música %s: %v\n", sampleFile, err)
	}

	termUI := tui.NewTUI()
	termUI.DisplayHeader(s)
	if err := termUI.Init(); err != nil {
		log.Fatalf("Erro ao inicializar TUI: %v\n", err)
	}

	audioSynth := audio.NewSynth()
	eng := engine.NewEngine(s)
	eng.Play()

	var pcmBuffer []byte
	var pcmMu sync.Mutex

	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		
		var currentChord string
		var chordStartMS int64
		var lastTimeMS int64

		for pbEvent := range eng.Events() {
			lastTimeMS = pbEvent.CurrentTimeMS
			chord := ""
			
			if pbEvent.ActiveEvent != nil {
				chord = pbEvent.ActiveEvent.ChordStr
				if chord == "" && pbEvent.ActiveEvent.Chord != nil {
					chord = pbEvent.ActiveEvent.Chord.Name
				}
			}

			// Se o acorde mudou, renderiza o acorde anterior com a duração exata
			if chord != currentChord {
				if outputFile != "" && currentChord != "" {
					durMS := pbEvent.CurrentTimeMS - chordStartMS
					if durMS > 0 {
						freqs := synth.GetChordFrequencies(currentChord)
						if len(freqs) > 0 {
							pcm := synth.GeneratePCMWithADSR(freqs, time.Duration(durMS)*time.Millisecond)
							pcmMu.Lock()
							pcmBuffer = append(pcmBuffer, pcm...)
							pcmMu.Unlock()
						}
					}
				}

				currentChord = chord
				chordStartMS = pbEvent.CurrentTimeMS
				
				if chord != "" {
					audioSynth.PlayChord(chord)
				} else {
					audioSynth.PlayChord("") // Silêncio
				}
			}

			// Envia atualização contínua para a interface visual
			_ = termUI.RenderTick(ui.PlaybackEvent{
				Song:     s,
				Current:  pbEvent.ActiveEvent,
				Position: pbEvent.CurrentTimeMS,
			})
		}

		// Flush do último acorde quando a música terminar
		if outputFile != "" && currentChord != "" {
			durMS := lastTimeMS - chordStartMS
			if durMS <= 0 {
				durMS = 1636 // fallback para o último acorde
			}
			freqs := synth.GetChordFrequencies(currentChord)
			if len(freqs) > 0 {
				pcm := synth.GeneratePCMWithADSR(freqs, time.Duration(durMS)*time.Millisecond)
				pcmMu.Lock()
				pcmBuffer = append(pcmBuffer, pcm...)
				pcmMu.Unlock()
			}
		}
	}()

	// Aguarda o término dos eventos para fechar a TUI automaticamente
	go func() {
		<-doneChan
		termUI.Close()
	}()

	// Inicia a TUI interativa (bloqueante)
	_ = termUI.Run()

	eng.Stop()

	// Gera o arquivo de áudio WAV ao término se a opção -out for definida
	if outputFile != "" {
		pcmMu.Lock()
		defer pcmMu.Unlock()
		if len(pcmBuffer) > 0 {
			err := audio.WriteWAV(outputFile, pcmBuffer, synth.SampleRate, 1) // 1 canal (Mono)
			if err != nil {
				fmt.Printf("\nErro ao salvar arquivo de áudio: %v\n", err)
			} else {
				fmt.Printf("\nÁudio gravado com sucesso em: %s\n", outputFile)
			}
		} else {
			fmt.Println("\nNenhum áudio gerado para gravar.")
		}
	}
}
