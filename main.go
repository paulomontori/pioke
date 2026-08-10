package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"opentune/pkg/engine"
	"opentune/pkg/parser"
)

func main() {
	sampleFile := "examples/sample.json"
	if len(os.Args) > 1 {
		sampleFile = os.Args[1]
	}

	fmt.Printf("Carregando musica de %s...\n", sampleFile)
	s, err := parser.ParseSong(sampleFile)
	if err != nil {
		log.Fatalf("Erro ao carregar a musica: %v\n", err)
	}

	fmt.Printf("\n--- Tocando: %s - %s ---\n\n", s.Title, s.Artist)

	eng := engine.NewEngine(s)
	eng.Start()

	// Consome os eventos emitidos pelo engine em tempo real
	done := make(chan bool)
	go func() {
		for event := range eng.Events() {
			currentTime := event.Duration / time.Millisecond

			chord := event.ChordStr
			if chord == "" && event.Chord != nil {
				chord = event.Chord.Name
			}

			lyric := event.Lyric
			if lyric == "" && event.Lyrics != nil {
				lyric = event.Lyrics.Text
			}

			fmt.Printf("[%05d ms] Acorde: %-6s | Letra: %s\n", currentTime, chord, lyric)
		}
		done <- true
	}()

	// Aguarda o termino da execucao da musica ou um tempo limite para demonstracao
	time.Sleep(16 * time.Second)
	eng.Stop()
	fmt.Println("\nReproducao concluida.")
}
