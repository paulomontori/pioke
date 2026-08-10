package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"opentune/pkg/audio"
	"opentune/pkg/engine"
	"opentune/pkg/parser"
	"opentune/pkg/ui"
)

func main() {
	sampleFile := "examples/sample.json"
	if len(os.Args) > 1 {
		sampleFile = os.Args[1]
	}

	fmt.Printf("Carregando música de %s...\n", sampleFile)
	s, err := parser.ParseSong(sampleFile)
	if err != nil {
		log.Fatalf("Erro ao carregar a música: %v\n", err)
	}

	termUI := ui.NewTerminalUI()
	termUI.DisplayHeader(s)

	synth := audio.NewSynth()
	eng := engine.NewEngine(s)
	eng.Start()

	done := make(chan bool)
	go func() {
		for event := range eng.Events() {
			chord := event.ChordStr
			if chord == "" && event.Chord != nil {
				chord = event.Chord.Name
			}

			// Toca o acorde via sintetizador de áudio
			synth.PlayChord(chord)

			// Exibe na UI
			termUI.RenderEvent(event)
		}
		done <- true
	}()

	// Aguarda a execução da demonstração
	time.Sleep(16 * time.Second)
	eng.Stop()
	fmt.Println("\nReprodução concluída.")
}
