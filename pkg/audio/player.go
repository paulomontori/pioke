package audio

import (
	"context"
	"time"

	"pioke/pkg/model"
	"pioke/pkg/synth"
)

// AudioPlayer escuta eventos de reprodução e aciona o sintetizador
type AudioPlayer struct {
	synth           synth.Synthesizer
	lastChordPlayed string
}

func NewAudioPlayer(s synth.Synthesizer) *AudioPlayer {
	return &AudioPlayer{
		synth: s,
	}
}

// Listen consome eventos do canal de reprodução de forma não-bloqueante
func (p *AudioPlayer) Listen(ctx context.Context, events <-chan model.PlaybackEvent) {
	for {
		select {
		case <-ctx.Done():
			p.synth.Stop()
			return
		case ev, ok := <-events:
			if !ok {
				p.synth.Stop()
				return
			}

			if ev.State != model.PLAYING {
				p.lastChordPlayed = ""
				continue
			}

			if ev.ActiveEvent != nil && ev.ActiveEvent.ChordStr != "" {
				chord := ev.ActiveEvent.ChordStr
				if chord != p.lastChordPlayed {
					p.lastChordPlayed = chord
					dur := time.Duration(ev.ActiveEvent.DurationMS) * time.Millisecond
					if dur <= 0 {
						dur = 2 * time.Second
					}
					p.synth.PlayChord(chord, dur)
				}
			}
		}
	}
}
