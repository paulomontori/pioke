package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pioke/pkg/audio"
	"pioke/pkg/playback"
	"pioke/pkg/scoring"
	"pioke/pkg/song"
)

// runScore implementa o subcomando "score": avalia uma gravação salva por -record (ver
// playback.RecordingMeta) contra a melodia de referência da música original, sob o nível de
// dificuldade escolhido, e imprime o resultado no terminal.
func runScore(args []string) {
	fs := flag.NewFlagSet("score", flag.ExitOnError)
	level := fs.String("level", "medium", "Nível de dificuldade: easy, medium ou hard")
	songFile := fs.String("song", "", "Caminho da música de referência (padrão: a música gravada nos metadados da gravação)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: pioke-cli score <gravacao.wav> [-level easy|medium|hard] [-song musica.yaml]")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}
	wavPath := fs.Arg(0)

	meta, err := loadRecordingMeta(wavPath)
	if err != nil {
		log.Fatalf("erro ao carregar metadados da gravação: %v", err)
	}

	sf := *songFile
	if sf == "" {
		sf = meta.SongFile
	}
	if sf == "" {
		log.Fatalf("não foi possível determinar a música de referência (metadados sem song_file) — informe com -song")
	}

	s, err := song.LoadSong(sf)
	if err != nil {
		log.Fatalf("erro ao carregar a música %s: %v\n", sf, err)
	}

	preset, err := scoring.PresetByName(*level)
	if err != nil {
		log.Fatalf("%v", err)
	}

	pcm, sampleRate, channels, err := audio.ReadWAV(wavPath)
	if err != nil {
		log.Fatalf("erro ao ler gravação %s: %v", wavPath, err)
	}
	if channels != 1 {
		log.Fatalf("gravação %s tem %d canais; esperado mono (1 canal)", wavPath, channels)
	}

	result := scoring.Score(s, pcm, sampleRate, meta.PlaybackOffsetMS, preset)
	printScoreReport(wavPath, meta, result)
}

// loadRecordingMeta lê o .json de metadados salvo ao lado do .wav pelo modo -record (mesmo nome
// base, ver playback.finishRecording).
func loadRecordingMeta(wavPath string) (playback.RecordingMeta, error) {
	metaPath := strings.TrimSuffix(wavPath, filepath.Ext(wavPath)) + ".json"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return playback.RecordingMeta{}, fmt.Errorf("não foi possível ler %s (esperado ao lado do WAV, gerado por -record): %w", metaPath, err)
	}
	var meta playback.RecordingMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return playback.RecordingMeta{}, fmt.Errorf("metadados inválidos em %s: %w", metaPath, err)
	}
	return meta, nil
}

// printScoreReport imprime o resultado da avaliação no terminal: pontuação final, os três
// componentes que a compõem, e o detalhamento por nota (sílaba, se coberta, desvio de afinação em
// cents e desvio de onset em ms).
func printScoreReport(wavPath string, meta playback.RecordingMeta, result scoring.Result) {
	title := meta.SongTitle
	if title == "" {
		title = wavPath
	}

	fmt.Printf("\n%s — nível %s\n", title, result.Preset.Name)
	fmt.Printf("Pontuação final: %.0f/100\n", result.FinalScore)
	fmt.Printf("  Afinação:  %5.1f  (peso %.0f%%)\n", result.PitchScore, result.Preset.WeightPitch)
	fmt.Printf("  Ritmo:     %5.1f  (peso %.0f%%)\n", result.RhythmScore, result.Preset.WeightRhythm)
	fmt.Printf("  Cobertura: %5.1f  (peso %.0f%%)\n", result.CoverageScore, result.Preset.WeightCoverage)

	if len(result.Notes) == 0 {
		return
	}

	fmt.Println("\nDetalhamento por nota:")
	for _, nb := range result.Notes {
		lyric := nb.Lyric
		if lyric == "" {
			lyric = "(sem sílaba)"
		}
		if !nb.Covered {
			if nb.RhythmScore > 0 || nb.OnsetDeviationMS != 0 {
				fmt.Printf("  [%6dms] %-4s %-12s -- sem afinação capturada na janela da nota (onset %+dms)\n", nb.StartMS, nb.Pitch, lyric, nb.OnsetDeviationMS)
			} else {
				fmt.Printf("  [%6dms] %-4s %-12s -- sem voz detectada\n", nb.StartMS, nb.Pitch, lyric)
			}
			continue
		}
		octaveMark := ""
		if nb.OctaveCorrected {
			octaveMark = " (oitava)"
		}
		fmt.Printf("  [%6dms] %-4s %-12s afinação %5.1f (%+.0fct%s)  ritmo %5.1f (%+dms)  nota %5.1f\n",
			nb.StartMS, nb.Pitch, lyric, nb.PitchScore, nb.CentsDeviation, octaveMark, nb.RhythmScore, nb.OnsetDeviationMS, nb.Score)
	}

	worst := worstNotes(result.Notes, 3)
	if len(worst) > 0 {
		fmt.Println("\nPontos de atenção:")
		for _, nb := range worst {
			lyric := nb.Lyric
			if lyric == "" {
				lyric = "(sem sílaba)"
			}
			fmt.Printf("  [%6dms] %-4s %-12s nota %.1f\n", nb.StartMS, nb.Pitch, lyric, nb.Score)
		}
	}
}

// worstNotes retorna até n notas cobertas com a menor pontuação combinada — as candidatas mais
// úteis pra mostrar como feedback de "essa parte você mandou mal".
func worstNotes(notes []scoring.NoteBreakdown, n int) []scoring.NoteBreakdown {
	var covered []scoring.NoteBreakdown
	for _, nb := range notes {
		if nb.Covered {
			covered = append(covered, nb)
		}
	}
	sort.Slice(covered, func(i, j int) bool { return covered[i].Score < covered[j].Score })
	if len(covered) > n {
		covered = covered[:n]
	}
	return covered
}
