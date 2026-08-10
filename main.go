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

	var renderer ui.Renderer = ui.NewTerminalUI()
	if err := renderer.Init(); err != nil {
		log.Fatalf("Erro ao inicializar UI: %v\n", err)
	}
	defer renderer.Close()

	renderer.DisplayHeader(s)

	synth := audio.NewSynth()
	eng := engine.NewEngine(s)
	eng.Start()

	done := make(chan bool)
	go func() {
		for pbEvent := range eng.Events() {
			if pbEvent.Current != nil {
				chord := pbEvent.Current.ChordStr
				if chord == "" && pbEvent.Current.Chord != nil {
					chord = pbEvent.Current.Chord.Name
				}

				// Toca o acorde via sintetizador de áudio
				synth.PlayChord(chord)
			}

			// Renderiza na UI desacoplada
			_ = renderer.RenderTick(pbEvent)
		}
		done <- true
	}()

	// Aguarda a execução do protótipo
	time.Sleep(16 * time.Second)
	eng.Stop()
	fmt.Println("\nReprodução concluída.")
}
