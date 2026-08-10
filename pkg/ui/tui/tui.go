package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"pioke/pkg/model"
	"pioke/pkg/song"
	"pioke/pkg/ui"
)

// TerminalUI encapsula o programa Bubble Tea e implementa a interface ui.Renderer
type TerminalUI struct {
	program *tea.Program
	model   Model
	ctx     context.Context
	library *song.Library
}

func NewTerminalUI() *TerminalUI {
	return &TerminalUI{
		model: NewModel(),
	}
}

func (t *TerminalUI) SetLibrary(lib *song.Library) {
	t.library = lib
}

func (t *TerminalUI) Init(ctx context.Context) error {
	t.ctx = ctx
	if t.library != nil {
		if items, err := t.library.Scan(); err == nil && len(items) > 0 {
			// Carrega a primeira música encontrada por padrão se nenhuma tiver sido definida
			if t.model.Song == nil {
				if s, err := t.library.GetSong(items[0].FilePath); err == nil {
					t.model.Song = s
				}
			}
		}
	}

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
