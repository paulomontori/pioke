package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"pioke/pkg/playback"
	"pioke/pkg/song"
	"pioke/pkg/synth"
	"pioke/pkg/ui/tui"
)

func main() {
	// Subcomando "score" tem seu próprio parsing de flags e não passa pelo fluxo de reprodução
	// abaixo (ver runScore em score.go).
	if len(os.Args) > 1 && os.Args[1] == "score" {
		runScore(os.Args[2:])
		return
	}

	var outputFile string
	flag.StringVar(&outputFile, "out", "", "Caminho do arquivo WAV de saída para salvar o áudio gerado (ex: output.wav)")
	var timbreFlag string
	flag.StringVar(&timbreFlag, "timbre", "", "Timbre de síntese: additive (padrão, harmônicos aditivos) ou karplus (corda dedilhada)")
	var record bool
	flag.BoolVar(&record, "record", false, "Grava o microfone durante a reprodução (salvo em recordings/), para avaliar a qualidade do canto depois")
	flag.Parse()

	// Fallback para capturar as flags -out/-timbre/-record caso tenham sido informadas após
	// argumentos posicionais (flag.Parse() para no primeiro argumento não reconhecido como flag).
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

	sampleFile := "songs/parabens.yaml"
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") && arg != outputFile && arg != timbreFlag {
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
