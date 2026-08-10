package gui

import (
	"testing"

	"pioke/pkg/model"
	"pioke/pkg/ui"
)

func TestGUIRendererState(t *testing.T) {
	gui := NewGUIRenderer()
	song := &model.Song{
		Title:  "GUI Test Song",
		Artist: "GUI Test Artist",
	}

	gui.DisplayHeader(song)
	if gui.song == nil || gui.song.Title != "GUI Test Song" {
		t.Errorf("Esperado título 'GUI Test Song', obtido '%v'", gui.song)
	}

	event := ui.PlaybackEvent{
		Song: song,
		Current: &model.TimelineEvent{
			ChordStr: "C",
			Lyric:    "Test Lyric",
		},
		Position: 500,
	}

	err := gui.RenderTick(event)
	if err != nil {
		t.Fatalf("Erro em RenderTick: %v", err)
	}

	if gui.positionMS != 500 {
		t.Errorf("Posição esperada 500ms, obtida %d", gui.positionMS)
	}

	if gui.activeEvent == nil || gui.activeEvent.ChordStr != "C" {
		t.Errorf("Acorde esperado 'C', obtido '%v'", gui.activeEvent)
	}
}
