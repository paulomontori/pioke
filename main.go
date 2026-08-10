package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"pioke/pkg/audio"
	"pioke/pkg/engine"
	"pioke/pkg/parser"
	"pioke/pkg/ui/tui"
)

func main() {
	sampleFile := "examples/sample.json"
	if len(os.Args) > 1 {
		sampleFile = os.Args[1]
	}

	s, err := parser.ParseSong(sampleFile)
	if err != nil {
		log.Fatalf("Erro ao carregar a música: %v\n", err)
	}

	termUI := tui.NewTUI()
	termUI.DisplayHeader(s)
	if err := termUI.Init(); err != nil {
		log.Fatalf("Erro ao inicializar TUI: %v\n", err)
	}
	defer termUI.Close()

	synth := audio.NewSynth()
	eng := engine.NewEngine(s)
	eng.Start()

	go func() {
		for pbEvent := range eng.Events() {
			if pbEvent.Current != nil {
				chord := pbEvent.Current.ChordStr
				if chord == "" && pbEvent.Current.Chord != nil {
					chord = pbEvent.Current.Chord.Name
				}
				synth.PlayChord(chord)
			}
			_ = termUI.RenderTick(pbEvent)
		}
	}()

	time.Sleep(16 * time.Second)
	eng.Stop()
	fmt.Println("\nReprodução concluída.")
}
