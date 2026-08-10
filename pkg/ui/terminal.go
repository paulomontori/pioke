package ui

import (
	"fmt"
	"time"

	"pioke/pkg/model"
)

// TerminalUI gerencia a exibição do karaokê na linha de comando
type TerminalUI struct{}

// NewTerminalUI cria uma nova instância de TerminalUI
func NewTerminalUI() *TerminalUI {
	return &TerminalUI{}
}

// Init inicializa a UI do terminal
func (ui *TerminalUI) Init() error {
	return nil
}

// Close finaliza a UI do terminal
func (ui *TerminalUI) Close() error {
	return nil
}

// DisplayHeader exibe o cabeçalho da música
func (ui *TerminalUI) DisplayHeader(s *model.Song) {
	fmt.Printf("========================================\n")
	fmt.Printf("   PioKe Karaoke Engine\n")
	fmt.Printf("   Música: %s - %s\n", s.Title, s.Artist)
	if s.BPM > 0 {
		fmt.Printf("   BPM: %d | Tom: %s\n", s.BPM, s.Key)
	}
	fmt.Printf("========================================\n\n")
}

// RenderTick implementa a interface Renderer para a UI de terminal
func (ui *TerminalUI) RenderTick(e PlaybackEvent) error {
	if e.Current == nil {
		return nil
	}

	currentTime := e.Current.Duration / time.Millisecond

	chord := e.Current.ChordStr
	if chord == "" && e.Current.Chord != nil {
		chord = e.Current.Chord.Name
	}

	lyric := e.Current.Lyric
	if lyric == "" && e.Current.Lyrics != nil {
		lyric = e.Current.Lyrics.Text
	}

	fmt.Printf("[%05d ms] Acorde: %-6s | Letra: %s\n", currentTime, chord, lyric)
	return nil
}

// RenderEvent exibe a letra e o acorde sincronizados em tempo real (compatibilidade)
func (ui *TerminalUI) RenderEvent(event model.TimelineEvent) {
	ui.RenderTick(PlaybackEvent{
		Current: &event,
	})
}
