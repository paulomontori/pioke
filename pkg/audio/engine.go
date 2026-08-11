package audio

import (
	"io"
	"sync"

	"pioke/pkg/synth"
)

const (
	DefaultSampleRate = 44100
	DefaultChannels   = 2
)

// PCMStream encapsula um leitor de amostras em formato de bytes para consumo de drivers de som
type PCMStream struct {
	synth    *synth.Synthesizer
	mu       sync.Mutex
	floatBuf []float32 // Buffer reutilizado para evitar alocações contínuas
}

// NewPCMStream recebe o ponteiro *synth.Synthesizer
func NewPCMStream(s *synth.Synthesizer) *PCMStream {
	return &PCMStream{
		synth:    s,
		floatBuf: make([]float32, 1024),
	}
}

// Reset limpa buffers pendentes e interrompe a síntese atual
func (p *PCMStream) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.synth != nil {
		p.synth.Stop()
	}
}

// Read lê amostras sintetizadas e converte float32 PCM em bytes (LittleEndian 16-bit PCM)
func (p *PCMStream) Read(b []byte) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSamples := len(b) / 2
	if numSamples == 0 {
		return 0, nil
	}

	if p.synth == nil {
		for i := range b {
			b[i] = 0
		}
		return len(b), nil
	}

	// Redimensiona o buffer temporário apenas se a requisição for maior
	if len(p.floatBuf) < numSamples {
		p.floatBuf = make([]float32, numSamples)
	}

	subBuf := p.floatBuf[:numSamples]
	p.synth.ReadPCM(subBuf)

	for i, sample := range subBuf {
		// Converte float32 [-1.0, 1.0] para int16
		val := int16(sample * 32767.0)
		b[i*2] = byte(val)
		b[i*2+1] = byte(val >> 8)
	}

	return len(b), nil
}

var _ io.Reader = (*PCMStream)(nil)