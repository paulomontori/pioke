package engine

import (
	"testing"
	"time"

	"pioke/pkg/model"
)

func TestEnginePlayback(t *testing.T) {
	song := &model.Song{
		Title:  "Test Engine Song",
		Artist: "Test Engine Artist",
		Timeline: []model.TimelineEvent{
			{
				TimeMS: 20,
				Lyric:  "First Event",
			},
			{
				TimeMS: 50,
				Lyric:  "Second Event",
			},
		},
	}

	eng := NewEngine(song)

	if eng.State() != STOPPED {
		t.Errorf("Estado inicial esperado STOPPED, obtido %v", eng.State())
	}

	eng.Play()
	if eng.State() != PLAYING {
		t.Errorf("Estado após Play esperado PLAYING, obtido %v", eng.State())
	}

	var eventsReceived int
	var receivedFirst, receivedSecond bool
	timeout := time.After(300 * time.Millisecond)

Loop:
	for {
		select {
		case ev := <-eng.Events():
			eventsReceived++
			if ev.ActiveEvent != nil {
				if ev.ActiveEvent.Lyric == "First Event" {
					receivedFirst = true
				}
				if ev.ActiveEvent.Lyric == "Second Event" {
					receivedSecond = true
				}
			}
			if receivedFirst && receivedSecond {
				break Loop
			}
		case <-timeout:
			t.Fatalf("Timeout aguardando eventos do engine. Recebidos: %d, First: %v, Second: %v", eventsReceived, receivedFirst, receivedSecond)
		}
	}

	eng.Pause()
	if eng.State() != PAUSED {
		t.Errorf("Estado após Pause esperado PAUSED, obtido %v", eng.State())
	}

	eng.Stop()
	if eng.State() != STOPPED {
		t.Errorf("Estado após Stop esperado STOPPED, obtido %v", eng.State())
	}
	if eng.Position() != 0 {
		t.Errorf("Posição após Stop esperada 0, obtida %v", eng.Position())
	}
}

func TestEngineSeek(t *testing.T) {
	song := &model.Song{Title: "Seek Test"}
	eng := NewEngine(song)

	eng.Seek(1500)
	if eng.Position() != 1500 {
		t.Errorf("Posição esperada 1500ms após Seek, obtida %v", eng.Position())
	}
}
