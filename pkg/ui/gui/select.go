package gui

import (
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"pioke/pkg/song"
	"pioke/pkg/synth"
)

const (
	logoDisplayHeight = 56
	logoDisplayWidth  = 52 // ~168/178 de logoDisplayHeight, arredondado

	selectHeaderY    = 20
	selectListStartY = 130
	selectListBottom = 660

	// rowHeight é o espaçamento vertical, em pixels, entre cada linha visível da lista de músicas
	// — usado tanto pra desenhar quanto pra converter a posição do mouse em índice de linha
	// (HitTest).
	rowHeight = 34

	timbreLabelY = 34
)

// visibleRows é quantas linhas da lista cabem na área reservada pra ela — o resto rola por cima
// (ver scrollOffset), em vez de estourar a tela quando a biblioteca tem muitas músicas.
var visibleRows = (selectListBottom - selectListStartY) / rowHeight

// SelectScreen exibe a lista de músicas encontradas na biblioteca, um seletor de timbre de síntese
// (violão dedilhado ou sintético aditivo — ver pkg/synth.Timbre) e permite escolher uma música por
// teclado (setas + Enter) ou mouse (hover + clique). A navegação/confirmação em si (MoveSelection,
// Confirm, HitTest, ToggleTimbre) é mantida separada da leitura de input do Ebiten (Update) para
// poder ser testada sem um game loop rodando.
type SelectScreen struct {
	items  []song.SongMetadata
	labels []string // "Título · Artista", já truncado pra largura da tela

	selected     int
	scrollOffset int

	timbre synth.Timbre
	fonts  *Fonts
}

// NewSelectScreen cria uma tela de seleção para os itens informados (tipicamente o resultado de
// song.Library.Scan()), com defaultTimbre pré-selecionado (normalmente o valor vindo da flag
// -timbre da CLI).
func NewSelectScreen(items []song.SongMetadata, defaultTimbre synth.Timbre) *SelectScreen {
	fonts := NewFonts()
	maxLabelWidth := float64(screenWidth - 2*sidePadding)

	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = truncateToWidth(buildSongLabel(item), fonts.Preview, maxLabelWidth)
	}

	return &SelectScreen{
		items:  items,
		labels: labels,
		timbre: defaultTimbre,
		fonts:  fonts,
	}
}

// buildSongLabel monta o texto de exibição de uma música (só título e artista — BPM/tom não
// ajudam a escolher uma música pra cantar, então ficam de fora pra manter a lista limpa),
// colapsando qualquer espaço em branco (inclusive quebras de linha, que alguns exportadores de
// partitura deixam em créditos de compositor) numa linha só — sem isso, um artista com texto
// multi-linha atropelava as linhas vizinhas na lista.
func buildSongLabel(item song.SongMetadata) string {
	label := item.Title
	if item.Artist != "" {
		label += " · " + item.Artist
	}
	return strings.Join(strings.Fields(label), " ")
}

// truncateToWidth encurta s (preservando o início, que é a parte mais identificável de um título)
// até caber em maxWidth, acrescentando "…" quando precisa cortar.
func truncateToWidth(s string, face *text.GoTextFace, maxWidth float64) string {
	if text.Advance(s, face) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "…"
		if text.Advance(candidate, face) <= maxWidth {
			return candidate
		}
	}
	return "…"
}

// MoveSelection desloca o item selecionado por delta (positivo desce, negativo sobe), dando a
// volta nas pontas da lista (da última música pra primeira, e vice-versa) e rolando a janela
// visível (scrollOffset) pra manter a seleção sempre à vista. Não faz nada se a lista estiver
// vazia.
func (s *SelectScreen) MoveSelection(delta int) {
	n := len(s.items)
	if n == 0 {
		return
	}
	s.selected = ((s.selected+delta)%n + n) % n
	s.ensureSelectedVisible()
}

func (s *SelectScreen) ensureSelectedVisible() {
	if s.selected < s.scrollOffset {
		s.scrollOffset = s.selected
	}
	if s.selected >= s.scrollOffset+visibleRows {
		s.scrollOffset = s.selected - visibleRows + 1
	}
}

// Confirm retorna a música atualmente selecionada, ou nil se a lista estiver vazia.
func (s *SelectScreen) Confirm() *song.SongMetadata {
	if len(s.items) == 0 {
		return nil
	}
	item := s.items[s.selected]
	return &item
}

// Timbre retorna o timbre de síntese escolhido no momento (ver ToggleTimbre).
func (s *SelectScreen) Timbre() synth.Timbre {
	return s.timbre
}

// ToggleTimbre alterna entre os timbres disponíveis (ver synth.Timbres).
func (s *SelectScreen) ToggleTimbre() {
	i := slices.Index(synth.Timbres, s.timbre)
	s.timbre = synth.Timbres[(i+1)%len(synth.Timbres)]
}

// HitTest converte uma posição de mouse (x, y) no índice (absoluto, em s.items) da linha da lista
// sob o cursor, ou -1 se não houver nenhuma linha visível naquela posição.
func (s *SelectScreen) HitTest(x, y int) int {
	if y < selectListStartY {
		return -1
	}
	row := (y-selectListStartY)/rowHeight + s.scrollOffset
	if row < s.scrollOffset || row >= s.scrollOffset+visibleRows || row >= len(s.items) {
		return -1
	}
	return row
}

// Update lê teclado e mouse; retorna a música escolhida quando o usuário confirma a seleção (Enter
// ou clique em uma linha), e nil em qualquer outro caso (inclusive quando só o timbre foi trocado).
func (s *SelectScreen) Update() *song.SongMetadata {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		s.MoveSelection(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		s.MoveSelection(-1)
	}
	// ◀/▶ é o gesto universal de "ajustar um valor" em controles remotos de TV e gamepads, que só
	// costumam ter um D-pad + OK/voltar — nada garante uma tecla "T" disponível nesses aparelhos.
	// T continua funcionando como atalho extra pra quem estiver no teclado.
	if inpututil.IsKeyJustPressed(ebiten.KeyT) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		s.ToggleTimbre()
	}

	mx, my := ebiten.CursorPosition()
	if row := s.HitTest(mx, my); row >= 0 {
		s.selected = row
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			return s.Confirm()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return s.Confirm()
	}
	return nil
}

func (s *SelectScreen) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBackground)

	s.drawHeader(screen)
	s.drawTimbreBadge(screen)
	s.drawList(screen)
	s.drawFooter(screen)
}

func (s *SelectScreen) drawHeader(screen *ebiten.Image) {
	logoOpts := &ebiten.DrawImageOptions{}
	logoBounds := logoImage.Bounds()
	scale := float64(logoDisplayHeight) / float64(logoBounds.Dy())
	logoOpts.GeoM.Scale(scale, scale)
	logoOpts.GeoM.Translate(sidePadding, selectHeaderY)
	screen.DrawImage(logoImage, logoOpts)

	textX := float64(sidePadding + logoDisplayWidth + 20)
	drawTextAt(screen, "PioKe", s.fonts.Line, ColorSung, textX, selectHeaderY, text.AlignStart, text.AlignStart)
	drawTextAt(screen, "Escolha uma música", s.fonts.Header, ColorHeader, textX, selectHeaderY+52, text.AlignStart, text.AlignStart)
}

// drawTimbreBadge mostra o timbre escolhido como texto simples (sem caixa/preenchimento) no canto
// superior direito — só o valor já em destaque (cor de acento) já chama atenção o suficiente, sem
// pesar visualmente a tela.
func (s *SelectScreen) drawTimbreBadge(screen *ebiten.Image) {
	x := float64(screenWidth - sidePadding)
	drawTextAt(screen, "TIMBRE", s.fonts.Small, ColorProgressText, x, timbreLabelY, text.AlignEnd, text.AlignStart)
	drawTextAt(screen, "< "+timbreLabel(s.timbre)+" >", s.fonts.Header, ColorCurrent, x, timbreLabelY+18, text.AlignEnd, text.AlignStart)
}

func (s *SelectScreen) drawList(screen *ebiten.Image) {
	if len(s.items) == 0 {
		drawTextAt(screen, "Nenhuma música encontrada.", s.fonts.Preview, ColorUpcoming, sidePadding, selectListStartY, text.AlignStart, text.AlignStart)
		return
	}

	end := min(s.scrollOffset+visibleRows, len(s.items))
	for i := s.scrollOffset; i < end; i++ {
		y := selectListStartY + (i-s.scrollOffset)*rowHeight

		// A cor sozinha já marca a linha selecionada (sem precisar de um prefixo "> " de texto,
		// que lembra terminal e pesa a lista).
		col := ColorHeader
		if i == s.selected {
			col = ColorCurrent
		}
		drawTextAt(screen, s.labels[i], s.fonts.Preview, col, sidePadding, float64(y), text.AlignStart, text.AlignStart)
	}
}

func (s *SelectScreen) drawFooter(screen *ebiten.Image) {
	hint := "↑ ↓ navegar   Enter tocar   ← → timbre"
	drawTextAt(screen, hint, s.fonts.Small, ColorProgressText, sidePadding, selectListBottom+16, text.AlignStart, text.AlignStart)
}

// timbreLabel traduz um synth.Timbre pro nome exibido na tela de seleção.
func timbreLabel(t synth.Timbre) string {
	if t == synth.TimbreKarplus {
		return "Violão"
	}
	return "Aditivo"
}
