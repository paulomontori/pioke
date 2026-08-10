package tui

import (
	"fmt"

	"pioke/pkg/model"
	"pioke/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EventMsg repassa um PlaybackEvent para a barra do Bubbletea
type EventMsg ui.PlaybackEvent

// BubbleUI implementa a interface ui.Renderer usando Bubbletea
type BubbleUI struct {
	program *tea.Program
	model   Model
}

// Model representa o estado interno da TUI
type Model struct {
	song     *model.Song
	current  *model.TimelineEvent
	position int64
	playing  bool
}

// NewTUI cria uma nova instância de UI baseada em Bubbletea
func NewTUI() *BubbleUI {
	m := Model{
		playing: true,
	}
	return &BubbleUI{
		model: m,
	}
}

func (b *BubbleUI) Init() error {
	b.program = tea.NewProgram(b.model)
	go func() {
		if _, err := b.program.Run(); err != nil {
			fmt.Printf("Erro ao executar TUI: %v\n", err)
		}
	}()
	return nil
}

func (b *BubbleUI) DisplayHeader(s *model.Song) {
	b.model.song = s
}

func (b *BubbleUI) RenderTick(e ui.PlaybackEvent) error {
	if b.program != nil {
		b.program.Send(EventMsg(e))
	}
	return nil
}

func (b *BubbleUI) Close() error {
	if b.program != nil {
		b.program.Quit()
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.playing = !m.playing
		}
	case EventMsg:
		m.song = msg.Song
		m.current = msg.Current
		m.position = msg.Position
	}
	return m, nil
}

func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	chordStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFC6")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(0, 1)

	lyricStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFF00"))

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(1, 0, 0, 0)

	title := "PioKe Karaoke Engine"
	if m.song != nil && m.song.Title != "" {
		title = fmt.Sprintf("PioKe - %s (%s)", m.song.Title, m.song.Artist)
	}

	header := headerStyle.Render(titleStyle.Render(title))

	chordStr := "-"
	lyricStr := "..."

	if m.current != nil {
		if m.current.ChordStr != "" {
			chordStr = m.current.ChordStr
		} else if m.current.Chord != nil {
			chordStr = m.current.Chord.Name
		}

		if m.current.Lyric != "" {
			lyricStr = m.current.Lyric
		} else if m.current.Lyrics != nil {
			lyricStr = m.current.Lyrics.Text
		}
	}

	content := fmt.Sprintf("\n Acorde:  %s\n Letra:   %s\n Tempo:   %05d ms\n",
		chordStyle.Render(chordStr),
		lyricStyle.Render(lyricStr),
		m.position,
	)

	status := statusStyle.Render("[Espaço] Play/Pause | [Q] Sair")

	return fmt.Sprintf("%s\n%s\n%s\n", header, content, status)
}
