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
				Duration: 20 * time.Millisecond,
				Lyric:    "First Event",
			},
			{
				Duration: 50 * time.Millisecond,
				Lyric:    "Second Event",
			},
		},
	}

	eng := NewEngine(song)

	if eng.State() != StateStopped {
		t.Errorf("Estado inicial esperado StateStopped, obtido %v", eng.State())
	}

	eng.Start()
	if eng.State() != StatePlaying {
		t.Errorf("Estado após Start esperado StatePlaying, obtido %v", eng.State())
	}

	var eventsReceived int
	timeout := time.After(200 * time.Millisecond)

Loop:
	for {
		select {
		case ev := <-eng.Events():
			eventsReceived++
			if eventsReceived == 1 && ev.Current.Lyric != "First Event" {
				t.Errorf("Primeiro evento esperado 'First Event', obtido %q", ev.Current.Lyric)
			}
			if eventsReceived == 2 {
				if ev.Current.Lyric != "Second Event" {
					t.Errorf("Segundo evento esperado 'Second Event', obtido %q", ev.Current.Lyric)
				}
				break Loop
			}
		case <-timeout:
			t.Fatalf("Timeout aguardando eventos do engine. Recebidos: %d", eventsReceived)
		}
	}

	eng.Pause()
	if eng.State() != StatePaused {
		t.Errorf("Estado após Pause esperado StatePaused, obtido %v", eng.State())
	}

	eng.Stop()
	if eng.State() != StateStopped {
		t.Errorf("Estado após Stop esperado StateStopped, obtido %v", eng.State())
	}
	if eng.Position() != 0 {
		t.Errorf("Posição após Stop esperada 0, obtida %v", eng.Position())
	}
}

func TestEngineSeek(t *testing.T) {
	song := &model.Song{Title: "Seek Test"}
	eng := NewEngine(song)

	eng.Seek(1500 * time.Millisecond)
	if eng.Position() != 1500*time.Millisecond {
		t.Errorf("Posição esperada 1.5s após Seek, obtida %v", eng.Position())
	}
}
