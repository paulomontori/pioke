package audio

import (
	"io"

	"pioke/pkg/synth"
)

const (
	DefaultSampleRate = 44100
	DefaultChannels   = 2
)

// PCMStream encapsula um leitor de amostras em formato de bytes para consumo de drivers de som
type PCMStream struct {
	synth synth.Synthesizer
}

func NewPCMStream(s synth.Synthesizer) *PCMStream {
	return &PCMStream{synth: s}
}

// Read lê amostras sintetizadas e converte float32 PCM em bytes (LittleEndian 16-bit PCM)
func (p *PCMStream) Read(b []byte) (n int, err error) {
	numSamples := len(b) / 2
	if numSamples == 0 {
		return 0, nil
	}

	floatBuf := make([]float32, numSamples)
	p.synth.ReadPCM(floatBuf)

	for i, sample := range floatBuf {
		// Converte float32 [-1.0, 1.0] para int16
		val := int16(sample * 32767.0)
		b[i*2] = byte(val)
		b[i*2+1] = byte(val >> 8)
	}

	return len(b), nil
}

var _ io.Reader = (*PCMStream)(nil)
