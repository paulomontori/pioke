package playback

import (
	"sync"
	"testing"
	"time"

	"pioke/pkg/model"
	"pioke/pkg/synth"
	"pioke/pkg/ui"
)

// fakeRenderer é um ui.Renderer mínimo que só conta quantas vezes RenderTick foi chamado, usado
// pra verificar que Start encaminha os eventos do engine sem precisar de uma TUI/GUI real.
type fakeRenderer struct {
	mu    sync.Mutex
	ticks int
}

func (f *fakeRenderer) Init() error               { return nil }
func (f *fakeRenderer) DisplayHeader(*model.Song)  {}
func (f *fakeRenderer) Close() error               { return nil }
func (f *fakeRenderer) RenderTick(ui.PlaybackEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ticks++
	return nil
}

func (f *fakeRenderer) tickCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ticks
}

func TestStartForwardsEventsAndStopStopsCleanly(t *testing.T) {
	s := &model.Song{
		Title:  "Teste",
		Artist: "Testador",
		Timeline: []model.TimelineEvent{
			{TimeMS: 0, DurationMS: 50, Lyric: "la"},
		},
	}

	renderer := &fakeRenderer{}
	_, done, stop := Start(s, renderer, synth.TimbreAdditive, false, "songs/teste.yaml")

	// Dá tempo pro engine (ticker de 10ms) emitir ao menos um evento antes de parar.
	time.Sleep(30 * time.Millisecond)
	stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("done não fechou após stop()")
	}

	if renderer.tickCount() == 0 {
		t.Error("esperava ao menos uma chamada a RenderTick antes de stop()")
	}
}

func TestStartWithoutRecordingDoesNotBlock(t *testing.T) {
	s := &model.Song{Title: "Vazia"}

	renderer := &fakeRenderer{}
	_, done, stop := Start(s, renderer, synth.TimbreAdditive, false, "songs/vazia.yaml")
	stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("done não fechou após stop() em música sem timeline")
	}
}
