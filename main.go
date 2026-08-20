package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"pioke/pkg/playback"
	"pioke/pkg/song"
	"pioke/pkg/synth"
	"pioke/pkg/ui/gui"
	"pioke/pkg/ui/tui"
)

func main() {
	var outputFile string
	flag.StringVar(&outputFile, "out", "", "Caminho do arquivo WAV de saída para salvar o áudio gerado (ex: output.wav)")
	var timbreFlag string
	flag.StringVar(&timbreFlag, "timbre", "", "Timbre de síntese: additive (padrão, harmônicos aditivos) ou karplus (corda dedilhada)")
	var record bool
	flag.BoolVar(&record, "record", false, "Grava o microfone durante a reprodução (salvo em recordings/), para avaliar a qualidade do canto depois")
	var uiFlag string
	flag.StringVar(&uiFlag, "ui", "tui", "Interface: tui (padrão, terminal) ou gui (janela gráfica com seleção de músicas)")
	flag.Parse()

	// Fallback para capturar as flags -out/-timbre/-record/-ui caso tenham sido informadas após
	// argumentos posicionais (flag.Parse() para no primeiro argumento não reconhecido como flag).
	if uiFlag == "tui" {
		for i, arg := range os.Args {
			if (arg == "-ui" || arg == "--ui") && i+1 < len(os.Args) {
				uiFlag = os.Args[i+1]
				break
			} else if strings.HasPrefix(arg, "-ui=") {
				uiFlag = strings.TrimPrefix(arg, "-ui=")
				break
			} else if strings.HasPrefix(arg, "--ui=") {
				uiFlag = strings.TrimPrefix(arg, "--ui=")
				break
			}
		}
	}
	if !record {
		for _, arg := range os.Args {
			if arg == "-record" || arg == "--record" {
				record = true
				break
			}
		}
	}
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
	if timbreFlag == "" {
		for i, arg := range os.Args {
			if (arg == "-timbre" || arg == "--timbre") && i+1 < len(os.Args) {
				timbreFlag = os.Args[i+1]
				break
			} else if strings.HasPrefix(arg, "-timbre=") {
				timbreFlag = strings.TrimPrefix(arg, "-timbre=")
				break
			} else if strings.HasPrefix(arg, "--timbre=") {
				timbreFlag = strings.TrimPrefix(arg, "--timbre=")
				break
			}
		}
	}

	timbre, err := synth.ParseTimbre(timbreFlag)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if uiFlag == "gui" {
		if err := gui.NewApp("songs", timbre, record).Run(); err != nil {
			log.Fatalf("%v\n", err)
		}
		return
	}

	sampleFile := "songs/parabens.yaml"
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") && arg != outputFile && arg != timbreFlag && arg != uiFlag {
			sampleFile = arg
			break
		}
	}

	s, err := song.LoadSong(sampleFile)
	if err != nil {
		log.Fatalf("Erro ao carregar a música %s: %v\n", sampleFile, err)
	}

	if err := playback.Run(s, tui.NewTUI(), outputFile, timbre, record, sampleFile); err != nil {
		log.Fatalf("%v\n", err)
	}
}
