package audio

import (
	"github.com/ebitengine/oto/v3"
	"pioke/pkg/synth"
)

type AudioPlayer struct {
	otoCtx *oto.Context
	player *oto.Player
	synth  *synth.Synthesizer
}

func NewAudioPlayer(sampleRate int) (*AudioPlayer, error) {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, err
	}
	<-readyChan

	syn := synth.NewSynthesizer(sampleRate)
	player := otoCtx.NewPlayer(syn)
	player.Play()

	return &AudioPlayer{
		otoCtx: otoCtx,
		player: player,
		synth:  syn,
	}, nil
}

func (p *AudioPlayer) PlayChord(chordName string) {
	p.synth.SetChord(chordName)
}

func (p *AudioPlayer) Stop() {
	p.synth.Stop()
}