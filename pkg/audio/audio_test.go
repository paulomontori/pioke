package audio

import (
	"context"
	"testing"
	"time"

	"pioke/pkg/model"
	"pioke/pkg/synth"
)

func TestAudioPlayerListen(t *testing.T) {
	polySynth := synth.NewPolySynth(44100)
	player := NewAudioPlayer(polySynth)

	events := make(chan model.PlaybackEvent, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go player.Listen(ctx, events)

	events <- model.PlaybackEvent{
		State: model.PLAYING,
		ActiveEvent: &model.TimelineEvent{
			ChordStr:   "C",
			DurationMS: 100,
		},
	}

	events <- model.PlaybackEvent{
		State: model.PLAYING,
		ActiveEvent: &model.TimelineEvent{
			ChordStr:   "G",
			DurationMS: 100,
		},
	}

	time.Sleep(50 * time.Millisecond)
	close(events)
}

func TestPCMStreamRead(t *testing.T) {
	polySynth := synth.NewPolySynth(44100)
	polySynth.PlayChord("Am", 100*time.Millisecond)

	stream := NewPCMStream(polySynth)
	buf := make([]byte, 1024)

	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Erro ao ler PCMStream: %v", err)
	}

	if n != 1024 {
		t.Errorf("Esperava ler 1024 bytes, leu %d", n)
	}
}
