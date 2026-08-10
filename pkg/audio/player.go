package audio

import (
	"context"
	"sync"
	"time"

	"pioke/pkg/model"
	"pioke/pkg/synth"
)

// AudioPlayer escuta eventos de reprodução e aciona o sintetizador
type AudioPlayer struct {
	synth           synth.Synthesizer
	lastChordPlayed string
	mu              sync.Mutex
}

func NewAudioPlayer(s synth.Synthesizer) *AudioPlayer {
	return &AudioPlayer{
		synth: s,
	}
}

// Reset para a execução atual e reinicia o estado de acordes
func (p *AudioPlayer) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastChordPlayed = ""
	if p.synth != nil {
		p.synth.Stop()
	}
}

// Listen consome eventos do canal de reprodução de forma não-bloqueante
func (p *AudioPlayer) Listen(ctx context.Context, events <-chan model.PlaybackEvent) {
	for {
		select {
		case <-ctx.Done():
			p.Reset()
			return
		case ev, ok := <-events:
			if !ok {
				p.Reset()
				return
			}

			p.mu.Lock()
			if ev.State != model.PLAYING {
				p.lastChordPlayed = ""
				p.mu.Unlock()
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
					if p.synth != nil {
						p.synth.PlayChord(chord, dur)
					}
				}
			}
			p.mu.Unlock()
		}
	}
}
