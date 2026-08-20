package gui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"pioke/pkg/model"
)

const (
	screenWidth  = 1280
	screenHeight = 720

	lyricAreaMaxWidth = screenWidth - 2*sidePadding
	sidePadding       = 60

	currentLineY  = 320
	previewLineY  = 400
	chordBadgeX   = screenWidth - 260
	chordBadgeY   = 20
	chordBadgeW   = 200
	chordBadgeH   = 60
	progressBarX  = sidePadding
	progressBarY  = 660
	progressBarW  = screenWidth - 2*sidePadding
	progressBarH  = 10
)

var (
	ColorBackground     = color.RGBA{R: 20, G: 20, B: 30, A: 255}
	ColorHeader         = color.RGBA{R: 150, G: 150, B: 170, A: 255}
	ColorSung           = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	ColorCurrent        = color.RGBA{R: 255, G: 215, B: 0, A: 255}
	ColorUpcoming       = color.RGBA{R: 100, G: 100, B: 115, A: 255}
	ColorChordBadgeBG   = color.RGBA{R: 90, G: 60, B: 160, A: 255}
	ColorChordText      = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	ColorProgressTrack  = color.RGBA{R: 60, G: 60, B: 70, A: 255}
	ColorProgressFill   = color.RGBA{R: 0, G: 200, B: 200, A: 255}
	ColorProgressText   = color.RGBA{R: 180, G: 180, B: 190, A: 255}
)

// View desenha a tela de reprodução ao vivo: linha atual em destaque (com progresso por
// sílaba/palavra), prévia da próxima linha, badge de acorde e barra de progresso. Mantém as spans
// e a quebra em linhas em cache por música (só recalcula quando a música muda), já que elas não
// dependem do tempo de reprodução — só a posição atual dentro delas muda a cada frame.
type View struct {
	fonts *Fonts

	cachedSong  *model.Song
	spans       []lyricSpan
	lines       [][]lyricSpan
	lastChord   string
}

func NewView() *View {
	return &View{fonts: NewFonts()}
}

func (v *View) Render(screen *ebiten.Image, song *model.Song, activeEvent *model.TimelineEvent, posMS int64) {
	screen.Fill(ColorBackground)

	if song == nil {
		v.drawCentered(screen, "Aguardando seleção de música...", v.fonts.Line, ColorHeader, screenHeight/2)
		return
	}

	v.ensureCache(song)

	v.drawHeader(screen, song)
	v.drawChordBadge(screen, activeEvent)
	v.drawLyrics(screen, posMS)
	v.drawProgressBar(screen, song, posMS)
}

// ensureCache recalcula spans/lines quando a música mudou desde o último Render — elas não
// dependem de posMS, só do conteúdo da música e da largura disponível na tela (fixa aqui).
func (v *View) ensureCache(song *model.Song) {
	if song == v.cachedSong {
		return
	}
	v.cachedSong = song
	v.spans = buildLyricSpans(song)
	v.lines = wrapSpansIntoLines(v.spans, lyricAreaMaxWidth, func(s string) float64 {
		return text.Advance(s, v.fonts.Line)
	})
	v.lastChord = ""
}

func (v *View) drawHeader(screen *ebiten.Image, song *model.Song) {
	header := fmt.Sprintf("%s - %s", song.Title, song.Artist)
	if song.BPM > 0 || song.Key != "" {
		header += fmt.Sprintf("  |  BPM: %d  |  Tom: %s", song.BPM, song.Key)
	}
	v.drawText(screen, header, v.fonts.Header, ColorHeader, sidePadding, 20, text.AlignStart)
}

func (v *View) drawChordBadge(screen *ebiten.Image, activeEvent *model.TimelineEvent) {
	if activeEvent != nil && activeEvent.ChordStr != "" {
		v.lastChord = activeEvent.ChordStr
	}
	if v.lastChord == "" {
		return
	}

	vector.FillRect(screen, chordBadgeX, chordBadgeY, chordBadgeW, chordBadgeH, ColorChordBadgeBG, true)
	cx := float64(chordBadgeX) + float64(chordBadgeW)/2
	cy := float64(chordBadgeY) + float64(chordBadgeH)/2
	v.drawTextAligned(screen, v.lastChord, v.fonts.Chord, ColorChordText, cx, cy, text.AlignCenter, text.AlignCenter)
}

func (v *View) drawLyrics(screen *ebiten.Image, posMS int64) {
	if len(v.spans) == 0 {
		v.drawCentered(screen, "(sem letra)", v.fonts.Line, ColorUpcoming, currentLineY)
		return
	}

	activeIdx := activeSpanIndex(v.spans, posMS)
	lineIdx := lineIndexForSpan(v.lines, activeIdx)
	lineStart := v.lineStartSpanIndex(lineIdx)

	v.drawLine(screen, v.lines[lineIdx], lineStart, activeIdx, currentLineY)

	if lineIdx+1 < len(v.lines) {
		preview := spansText(v.lines[lineIdx+1])
		v.drawCentered(screen, preview, v.fonts.Preview, ColorUpcoming, previewLineY)
	}
}

// drawLine desenha os spans de uma linha lado a lado, centralizados horizontalmente, coloridos
// conforme já foram cantados (índice global < activeIdx), estão tocando agora (== activeIdx) ou
// ainda não chegaram (> activeIdx).
func (v *View) drawLine(screen *ebiten.Image, line []lyricSpan, lineStart, activeIdx int, y float64) {
	full := spansText(line)
	totalWidth := text.Advance(full, v.fonts.Line)
	x := (float64(screenWidth) - totalWidth) / 2

	for i, sp := range line {
		globalIdx := lineStart + i
		col := ColorUpcoming
		switch {
		case globalIdx == activeIdx:
			col = ColorCurrent
		case globalIdx < activeIdx:
			col = ColorSung
		}

		v.drawTextAligned(screen, sp.Text, v.fonts.Line, col, x, y, text.AlignStart, text.AlignCenter)
		x += text.Advance(sp.Text, v.fonts.Line)
	}
}

func (v *View) lineStartSpanIndex(lineIdx int) int {
	start := 0
	for i := range lineIdx {
		start += len(v.lines[i])
	}
	return start
}

func (v *View) drawProgressBar(screen *ebiten.Image, song *model.Song, posMS int64) {
	totalMS := songDurationMS(song)

	vector.FillRect(screen, progressBarX, progressBarY, progressBarW, progressBarH, ColorProgressTrack, false)
	if totalMS > 0 {
		frac := float64(posMS) / float64(totalMS)
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		vector.FillRect(screen, progressBarX, progressBarY, float32(float64(progressBarW)*frac), progressBarH, ColorProgressFill, false)
	}

	label := fmt.Sprintf("%s / %s", formatMS(posMS), formatMS(totalMS))
	v.drawText(screen, label, v.fonts.Small, ColorProgressText, sidePadding, progressBarY+18, text.AlignStart)
}

func (v *View) drawCentered(screen *ebiten.Image, s string, face *text.GoTextFace, col color.Color, y float64) {
	drawTextAt(screen, s, face, col, float64(screenWidth)/2, y, text.AlignCenter, text.AlignCenter)
}

func (v *View) drawText(screen *ebiten.Image, s string, face *text.GoTextFace, col color.Color, x, y float64, primaryAlign text.Align) {
	drawTextAt(screen, s, face, col, x, y, primaryAlign, text.AlignStart)
}

func (v *View) drawTextAligned(screen *ebiten.Image, s string, face *text.GoTextFace, col color.Color, x, y float64, primaryAlign, secondaryAlign text.Align) {
	drawTextAt(screen, s, face, col, x, y, primaryAlign, secondaryAlign)
}

// drawTextAt desenha s com face/cor/alinhamento numa posição (x, y) — função livre (em vez de
// método) pra ser reaproveitada por qualquer tela do pacote (View, SelectScreen, ...), não só pela
// que a criou.
func drawTextAt(screen *ebiten.Image, s string, face *text.GoTextFace, col color.Color, x, y float64, primaryAlign, secondaryAlign text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.LineSpacing = face.Size * 1.2
	op.PrimaryAlign = primaryAlign
	op.SecondaryAlign = secondaryAlign
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, s, face, op)
}

// spansText concatena os textos de uma linha de spans (cada texto já traz o espaçamento entre
// palavras que precisar, seguindo a convenção usada nos arquivos de música).
func spansText(line []lyricSpan) string {
	var b strings.Builder
	for _, sp := range line {
		b.WriteString(sp.Text)
	}
	return b.String()
}

// songDurationMS calcula a duração total da música, com o mesmo fallback usado em
// pkg/engine/engine.go quando o último evento não informa DurationMS.
func songDurationMS(song *model.Song) int64 {
	if song == nil || len(song.Timeline) == 0 {
		return 0
	}
	last := song.Timeline[len(song.Timeline)-1]
	dur := last.DurationMS
	if dur <= 0 {
		dur = defaultEventDurationMS
	}
	return last.TimeMS + dur
}

func formatMS(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	totalSeconds := ms / 1000
	return fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
}
