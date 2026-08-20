package main

import (
	"fmt"
	"os"

	"pioke/pkg/audio"
	"pioke/pkg/song"
	"pioke/pkg/synth"
)

func main() {
	path := os.Args[1]
	out := os.Args[2]

	s, err := song.LoadSong(path)
	if err != nil {
		fmt.Printf("ERRO ao carregar: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Title=%q eventos=%d\n", s.Title, len(s.Timeline))

	segs := synth.BuildSegments(s)
	voiced := 0
	var totalMS int64
	for _, seg := range segs {
		totalMS += seg.DurationMS
		if len(seg.Freqs) > 0 {
			voiced++
		}
	}
	fmt.Printf("segments=%d voiced=%d totalMS=%d (%.1fs)\n", len(segs), voiced, totalMS, float64(totalMS)/1000)

	var cursor int64
	firstVoicedAt := int64(-1)
	for _, seg := range segs {
		if firstVoicedAt < 0 && len(seg.Freqs) > 0 {
			firstVoicedAt = cursor
		}
		cursor += seg.DurationMS
	}
	fmt.Printf("primeiro segmento com som: %dms (%.1fs)\n", firstVoicedAt, float64(firstVoicedAt)/1000)

	pcm := synth.RenderSong(s)
	fmt.Printf("pcm bytes=%d\n", len(pcm))

	if err := audio.WriteWAV(out, pcm, synth.SampleRate, synth.ChannelCount); err != nil {
		fmt.Printf("ERRO ao gravar WAV: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("WAV gravado em %s\n", out)
}
