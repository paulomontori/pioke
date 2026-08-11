package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"pioke/pkg/playback"
	"pioke/pkg/song"
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

	if err := playback.Run(s, tui.NewTUI(), outputFile); err != nil {
		log.Fatalf("%v\n", err)
	}
}
