package gui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"pioke/pkg/model"
)

var (
	ColorBackground = color.RGBA{R: 20, G: 20, B: 30, A: 255}
	ColorHeader     = color.RGBA{R: 125, G: 86, B: 244, A: 255}
	ColorActive     = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	ColorInactive   = color.RGBA{R: 150, G: 150, B: 150, A: 255}
	ColorChord      = color.RGBA{R: 255, G: 215, B: 0, A: 255}
)

type View struct {
	fontManager *FontManager
}

func NewView() *View {
	return &View{
		fontManager: NewFontManager(),
	}
}

func (v *View) Render(screen *ebiten.Image, song *model.Song, activeEvent *model.TimelineEvent, posMS int64) {
	screen.Fill(ColorBackground)

	// 1. Renderiza Cabeçalho
	if song != nil {
		header := fmt.Sprintf("%s - %s", song.Title, song.Artist)
		if song.BPM > 0 || song.Key != "" {
			header += fmt.Sprintf(" | BPM: %d | Tom: %s", song.BPM, song.Key)
		}
		ebitenutil.DebugPrintAt(screen, header, 20, 20)
	}

	// 2. Renderiza Acorde Ativo
	if activeEvent != nil && activeEvent.ChordStr != "" {
		chordText := fmt.Sprintf("Acorde: %s", activeEvent.ChordStr)
		ebitenutil.DebugPrintAt(screen, chordText, 20, 50)
	}

	// 3. Renderiza Timeline / Letra
	if song != nil && len(song.Timeline) > 0 {
		startY := 100
		for i, ev := range song.Timeline {
			if ev.Lyric == "" {
				continue
			}

			y := startY + (i * 25)
			if activeEvent != nil && ev.Lyric == activeEvent.Lyric {
				ebitenutil.DebugPrintAt(screen, "> "+ev.Lyric, 40, y)
			} else {
				ebitenutil.DebugPrintAt(screen, "  "+ev.Lyric, 40, y)
			}
		}
	}
}
