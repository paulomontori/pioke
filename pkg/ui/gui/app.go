package gui

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"pioke/pkg/playback"
	"pioke/pkg/song"
	"pioke/pkg/synth"
)

type screenState int

const (
	screenSelect screenState = iota
	screenPlayback
)

// App é o ponto de entrada da GUI: mantém a biblioteca de músicas, a tela de seleção e a tela de
// reprodução (reaproveitando GUIRenderer), alternando entre elas conforme o usuário escolhe uma
// música ou volta pra lista.
type App struct {
	library *song.Library
	sel     *SelectScreen
	player  *GUIRenderer
	state   screenState

	timbre synth.Timbre
	record bool

	doneCh <-chan struct{}
	stopFn func()
}

// NewApp cria a aplicação GUI, que varre songsDir em busca de músicas quando Run é chamado.
func NewApp(songsDir string, timbre synth.Timbre, record bool) *App {
	return &App{
		library: song.NewLibrary(songsDir),
		player:  NewGUIRenderer(),
		timbre:  timbre,
		record:  record,
	}
}

// Run configura a janela, varre a biblioteca de músicas e bloqueia rodando o loop do Ebiten até a
// janela ser fechada.
func (a *App) Run() error {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("PioKe - Karaoke Player")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	items, err := a.library.Scan()
	if err != nil {
		log.Printf("aviso: erro ao varrer biblioteca de músicas: %v", err)
	}
	a.sel = NewSelectScreen(items, a.timbre)

	return ebiten.RunGame(a)
}

func (a *App) Update() error {
	switch a.state {
	case screenSelect:
		a.updateSelect()
	case screenPlayback:
		a.updatePlayback()
	}
	return nil
}

func (a *App) updateSelect() {
	chosen := a.sel.Update()
	if chosen == nil {
		return
	}

	s, err := a.library.GetSong(chosen.FilePath)
	if err != nil {
		log.Printf("erro ao carregar música %s: %v", chosen.FilePath, err)
		return
	}

	a.player.DisplayHeader(s)
	_, done, stop := playback.Start(s, a.player, a.sel.Timbre(), a.record, chosen.FilePath)
	a.doneCh = done
	a.stopFn = stop
	a.state = screenPlayback
}

func (a *App) updatePlayback() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.backToSelect()
		return
	}

	select {
	case <-a.doneCh:
		a.backToSelect()
		return
	default:
	}

	_ = a.player.Update()
}

// backToSelect para a reprodução em andamento (se houver) e volta pra tela de seleção.
func (a *App) backToSelect() {
	if a.stopFn != nil {
		a.stopFn()
		a.stopFn = nil
	}
	a.doneCh = nil
	a.state = screenSelect
}

func (a *App) Draw(screen *ebiten.Image) {
	switch a.state {
	case screenSelect:
		a.sel.Draw(screen)
	case screenPlayback:
		a.player.Draw(screen)
		ebitenutil.DebugPrintAt(screen, "Esc: voltar à seleção", 20, 690)
	}
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}

var _ ebiten.Game = (*App)(nil)
