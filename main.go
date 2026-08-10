package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"pioke/pkg/audio"
	"pioke/pkg/engine"
	"pioke/pkg/song"
	"pioke/pkg/ui"
	"pioke/pkg/ui/tui"
)

func main() {
	sampleFile := "songs/parabens.yaml"
	if len(os.Args) > 1 {
		sampleFile = os.Args[1]
	}

	s, err := song.LoadSong(sampleFile)
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
	eng.Play()

	go func() {
		for pbEvent := range eng.Events() {
			if pbEvent.ActiveEvent != nil {
				chord := pbEvent.ActiveEvent.ChordStr
				if chord == "" && pbEvent.ActiveEvent.Chord != nil {
					chord = pbEvent.ActiveEvent.Chord.Name
				}
				synth.PlayChord(chord)
			}
			_ = termUI.RenderTick(ui.PlaybackEvent{
				Song:     s,
				Current:  pbEvent.ActiveEvent,
				Position: pbEvent.CurrentTimeMS,
			})
		}
	}()

	time.Sleep(16 * time.Second)
	eng.Stop()
	fmt.Println("\nReprodução concluída.")
}
