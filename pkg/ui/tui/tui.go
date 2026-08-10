package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"pioke/pkg/model"
	"pioke/pkg/song"
	"pioke/pkg/ui"
)

// Model representa o estado do Bubble Tea para a interface do terminal
type Model struct {
	Song  *model.Song
	Event model.PlaybackEvent
}

// NewModel inicializa o modelo padrão do Bubble Tea
func NewModel() Model {
	return Model{}
}

// EventMsg representa a mensagem enviada pelo evento de reprodução
type EventMsg model.PlaybackEvent

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EventMsg:
		m.Event = model.PlaybackEvent(msg)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.Song == nil {
		return "Aguardando seleção de música...\n"
	}

	res := fmt.Sprintf("=== %s - %s ===\n", m.Song.Title, m.Song.Artist)
	if m.Event.ActiveEvent != nil && m.Event.ActiveEvent.Lyric != "" {
		res += fmt.Sprintf("\nLetra: %s\n", m.Event.ActiveEvent.Lyric)
	}
	if m.Event.ActiveEvent != nil && m.Event.ActiveEvent.ChordStr != "" {
		res += fmt.Sprintf("Acorde: %s\n", m.Event.ActiveEvent.ChordStr)
	}
	return res
}

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

func (t *TerminalUI) Init() error {
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

func (t *TerminalUI) DisplayHeader(s *model.Song) {
	t.model.Song = s
	if t.program != nil {
		t.program.Send(EventMsg(model.PlaybackEvent{
			ActiveEvent: &model.TimelineEvent{},
		}))
	}
}

func (t *TerminalUI) RenderTick(event ui.PlaybackEvent) error {
	t.HandleEvent(model.PlaybackEvent{
		Song:        event.Song,
		ActiveEvent: event.Current,
	})
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
