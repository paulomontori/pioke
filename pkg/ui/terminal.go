package ui

import (
	"fmt"
	"time"

	"opentune/pkg/model"
)

// TerminalUI gerencia a exibição do karaokê na linha de comando
type TerminalUI struct{}

// NewTerminalUI cria uma nova instância de TerminalUI
func NewTerminalUI() *TerminalUI {
	return &TerminalUI{}
}

// DisplayHeader exibe o cabeçalho da música
func (ui *TerminalUI) DisplayHeader(s *model.Song) {
	fmt.Printf("========================================\n")
	fmt.Printf("   OpenTune / MicroMelody Karaoke\n")
	fmt.Printf("   Música: %s - %s\n", s.Title, s.Artist)
	if s.BPM > 0 {
		fmt.Printf("   BPM: %d | Tom: %s\n", s.BPM, s.Key)
	}
	fmt.Printf("========================================\n\n")
}

// RenderEvent exibe a letra e o acorde sincronizados em tempo real
func (ui *TerminalUI) RenderEvent(event model.TimelineEvent) {
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
