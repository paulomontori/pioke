package gui

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"pioke/pkg/song"
	"pioke/pkg/synth"
)

func testItems() []song.SongMetadata {
	return []song.SongMetadata{
		{FilePath: "songs/a.yaml", Title: "A", Artist: "Artist A", BPM: 90, Key: "C"},
		{FilePath: "songs/b.yaml", Title: "B", Artist: "Artist B", BPM: 100, Key: "D"},
		{FilePath: "songs/c.yaml", Title: "C", Artist: "Artist C", BPM: 110, Key: "E"},
	}
}

func newTestSelectScreen(items []song.SongMetadata) *SelectScreen {
	return NewSelectScreen(items, synth.TimbreAdditive)
}

func TestSelectScreenMoveSelectionWrapsAround(t *testing.T) {
	s := newTestSelectScreen(testItems()) // 3 itens: índices 0,1,2

	s.MoveSelection(-1)
	if s.selected != 2 {
		t.Errorf("esperado dar a volta pro último item (2) ao subir do topo, obtido %d", s.selected)
	}

	s.MoveSelection(1)
	if s.selected != 0 {
		t.Fatalf("esperado voltar pro primeiro item (0), obtido %d", s.selected)
	}

	s.MoveSelection(-1)
	s.MoveSelection(-1)
	if s.selected != 1 {
		t.Errorf("esperado selected=1 após duas voltas pra trás a partir de 0, obtido %d", s.selected)
	}

	s.MoveSelection(2)
	if s.selected != 0 {
		t.Errorf("esperado dar a volta pro primeiro item (0) ao passar do fim, obtido %d", s.selected)
	}
}

func TestSelectScreenMoveSelectionEmptyList(t *testing.T) {
	s := newTestSelectScreen(nil)
	s.MoveSelection(1)
	if s.selected != 0 {
		t.Errorf("esperado selected=0 com lista vazia, obtido %d", s.selected)
	}
}

func TestSelectScreenConfirm(t *testing.T) {
	s := newTestSelectScreen(testItems())
	s.MoveSelection(1)

	chosen := s.Confirm()
	if chosen == nil || chosen.Title != "B" {
		t.Fatalf("esperado música 'B', obtido %v", chosen)
	}
}

func TestSelectScreenConfirmEmptyList(t *testing.T) {
	s := newTestSelectScreen(nil)
	if chosen := s.Confirm(); chosen != nil {
		t.Errorf("esperado nil com lista vazia, obtido %v", chosen)
	}
}

func TestSelectScreenHitTest(t *testing.T) {
	s := newTestSelectScreen(testItems())

	if row := s.HitTest(50, 0); row != -1 {
		t.Errorf("esperado -1 acima da lista, obtido %d", row)
	}

	if row := s.HitTest(50, selectListStartY); row != 0 {
		t.Errorf("esperado linha 0, obtido %d", row)
	}

	if row := s.HitTest(50, selectListStartY+rowHeight+1); row != 1 {
		t.Errorf("esperado linha 1, obtido %d", row)
	}

	if row := s.HitTest(50, selectListStartY+rowHeight*len(testItems())); row != -1 {
		t.Errorf("esperado -1 abaixo da última linha, obtido %d", row)
	}
}

func TestSelectScreenToggleTimbre(t *testing.T) {
	s := newTestSelectScreen(testItems())
	if s.Timbre() != synth.TimbreAdditive {
		t.Fatalf("esperado timbre inicial additive, obtido %v", s.Timbre())
	}

	s.ToggleTimbre()
	if s.Timbre() != synth.TimbreKarplus {
		t.Errorf("esperado karplus após alternar, obtido %v", s.Timbre())
	}

	s.ToggleTimbre()
	if s.Timbre() != synth.TimbreAdditive {
		t.Errorf("esperado voltar a additive após alternar de novo, obtido %v", s.Timbre())
	}
}

func TestSelectScreenScrollKeepsSelectionVisible(t *testing.T) {
	items := make([]song.SongMetadata, visibleRows+5)
	for i := range items {
		items[i] = song.SongMetadata{FilePath: "songs/x.yaml", Title: "Song"}
	}
	s := newTestSelectScreen(items)

	for range visibleRows + 3 {
		s.MoveSelection(1)
	}

	if s.selected < s.scrollOffset || s.selected >= s.scrollOffset+visibleRows {
		t.Fatalf("seleção (%d) saiu da janela visível [%d, %d)", s.selected, s.scrollOffset, s.scrollOffset+visibleRows)
	}

	// volta exatamente até o índice 0 (sem dar a volta pra ponta oposta, que é testado à parte em
	// TestSelectScreenMoveSelectionWrapsAround).
	for range visibleRows + 3 {
		s.MoveSelection(-1)
	}
	if s.selected != 0 {
		t.Fatalf("esperado selected=0, obtido %d", s.selected)
	}
	if s.scrollOffset != 0 {
		t.Errorf("esperado scrollOffset=0 ao voltar pro topo, obtido %d", s.scrollOffset)
	}
}

// Regressão: artista/título com espaço em branco estranho (quebra de linha embutida por algum
// exportador de partitura) não deve virar texto multi-linha que atropela a linha vizinha na lista.
func TestBuildSongLabelCollapsesWhitespace(t *testing.T) {
	item := song.SongMetadata{Title: "Antigo Lugar", Artist: "Música de\nCARLOS NIEHUES\nMARCOS VENICIUS"}
	label := buildSongLabel(item)
	if strings.ContainsAny(label, "\n\r\t") {
		t.Fatalf("label não deveria conter quebras de linha/tabs: %q", label)
	}
	if label != "Antigo Lugar · Música de CARLOS NIEHUES MARCOS VENICIUS" {
		t.Errorf("esperava espaços colapsados numa linha só, obtido %q", label)
	}
}

func TestTruncateToWidthKeepsShortStringsUnchanged(t *testing.T) {
	fonts := NewFonts()
	if got := truncateToWidth("curto", fonts.Preview, 10000); got != "curto" {
		t.Errorf("string curta não deveria ser truncada, obtido %q", got)
	}
}

func TestTruncateToWidthShrinksLongStrings(t *testing.T) {
	fonts := NewFonts()
	long := strings.Repeat("um texto bem longo ", 20)
	got := truncateToWidth(long, fonts.Preview, 200)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("esperava reticências no fim do texto truncado, obtido %q", got)
	}
	if text.Advance(got, fonts.Preview) > 200 {
		t.Errorf("texto truncado ainda excede maxWidth: %q", got)
	}
}
