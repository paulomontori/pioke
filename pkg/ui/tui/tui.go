package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"pioke/pkg/model"
	"pioke/pkg/ui"
)

// TerminalUI encapsula o programa Bubble Tea e implementa a interface ui.Renderer
type TerminalUI struct {
	program *tea.Program
	model   Model
	ctx     context.Context
}

func NewTerminalUI() *TerminalUI {
	return &TerminalUI{
		model: NewModel(),
	}
}

func (t *TerminalUI) Init(ctx context.Context) error {
	t.ctx = ctx
	t.program = tea.NewProgram(t.model, tea.WithAltScreen())
	return nil
}

func (t *TerminalUI) SetSong(s *model.Song) {
	t.model.Song = s
}

func (t *TerminalUI) HandleEvent(event model.PlaybackEvent) {
	if t.program != nil {
		t.program.Send(EventMsg(event))
	}
}

func (t *TerminalUI) Run() error {
	if t.program == nil {
		return fmt.Errorf("programa TUI não foi inicializado")
	}

	_, err := t.program.Run()
	return err
}

func (t *TerminalUI) Close() error {
	if t.program != nil {
		t.program.Quit()
	}
	return nil
}

var _ ui.Renderer = (*TerminalUI)(nil)
