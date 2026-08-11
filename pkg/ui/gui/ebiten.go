package gui

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"pioke/pkg/model"
	"pioke/pkg/ui"
)

type GUIRenderer struct {
	mu          sync.Mutex
	song        *model.Song
	activeEvent *model.TimelineEvent
	positionMS  int64
	view        *View
	fullscreen  bool
}

func NewGUIRenderer() *GUIRenderer {
	return &GUIRenderer{
		view: NewView(),
	}
}

func (g *GUIRenderer) Init() error {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("PioKe - Karaoke Player")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return nil
}

func (g *GUIRenderer) DisplayHeader(s *model.Song) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.song = s
}

func (g *GUIRenderer) RenderTick(event ui.PlaybackEvent) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if event.Song != nil {
		g.song = event.Song
	}
	g.activeEvent = event.Current
	g.positionMS = event.Position
	return nil
}

func (g *GUIRenderer) Update() error {
	// Alterna tela cheia com F11
	if ebiten.IsKeyPressed(ebiten.KeyF11) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}
	return nil
}

func (g *GUIRenderer) Draw(screen *ebiten.Image) {
	g.mu.Lock()
	song := g.song
	active := g.activeEvent
	pos := g.positionMS
	g.mu.Unlock()

	g.view.Render(screen, song, active, pos)
}

func (g *GUIRenderer) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}

func (g *GUIRenderer) Run() error {
	return ebiten.RunGame(g)
}

func (g *GUIRenderer) Close() error {
	return nil
}

var _ ui.Renderer = (*GUIRenderer)(nil)
var _ ebiten.Game = (*GUIRenderer)(nil)
