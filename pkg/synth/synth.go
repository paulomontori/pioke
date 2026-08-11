package synth

import (
	"math"
	"sync"
)

// Mapeamento das frequências dos acordes em Hz
var chordFrequencies = map[string][]float64{
	"C":  {261.63, 329.63, 392.00},         // C4, E4, G4
	"G":  {196.00, 246.94, 293.66},         // G3, B3, D4
	"G7": {196.00, 246.94, 293.66, 349.23}, // G3, B3, D4, F4
	"F":  {174.61, 220.00, 261.63},         // F3, A3, C4
	"C7": {261.63, 329.63, 392.00, 466.16}, // C4, E4, G4, Bb4
	"Am": {220.00, 261.63, 329.63},         // A3, C4, E4
	"Dm": {146.83, 174.61, 220.00},         // D3, F3, A3
	"Em": {164.81, 196.00, 246.94},         // E3, G3, B3
}

type Synthesizer struct {
	mu         sync.Mutex
	sampleRate int
	phases     []float64
	activeFreq []float64
	volume     float64
}

func NewSynthesizer(sampleRate int) *Synthesizer {
	return &Synthesizer{
		sampleRate: sampleRate,
		volume:     0.3,
	}
}

// SetChord altera o acorde ativo mantendo a fase contínua
func (s *Synthesizer) SetChord(chordName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	freqs, exists := chordFrequencies[chordName]
	if !exists {
		s.activeFreq = nil
		s.phases = nil
		return
	}

	s.activeFreq = freqs
	if len(s.phases) != len(freqs) {
		s.phases = make([]float64, len(freqs))
	}
}

// Stop interrompe a emissão de som
func (s *Synthesizer) Stop() {
	s.SetChord("")
}

// ReadPCM preenche um buffer de amostras float32 (usado pelo engine/gravador WAV)
func (s *Synthesizer) ReadPCM(samples []float32) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.activeFreq) == 0 {
		for i := range samples {
			samples[i] = 0
		}
		return len(samples)
	}

	numVoices := float64(len(s.activeFreq))

	for i := 0; i < len(samples); i++ {
		var sample float64
		for voiceIdx, freq := range s.activeFreq {
			phaseStep := (2.0 * math.Pi * freq) / float64(s.sampleRate)
			sample += math.Sin(s.phases[voiceIdx])
			s.phases[voiceIdx] += phaseStep
			if s.phases[voiceIdx] >= 2.0*math.Pi {
				s.phases[voiceIdx] -= 2.0 * math.Pi
			}
		}
		samples[i] = float32((sample / numVoices) * s.volume)
	}

	return len(samples)
}

// Read implementa io.Reader preenchendo bytes PCM 16-bit (usado pelo Oto/Player)
func (s *Synthesizer) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	numSamples := len(p) / 2
	if len(s.activeFreq) == 0 {
		for i := range p {
			p[i] = 0
		}
		return len(p), nil
	}

	numVoices := float64(len(s.activeFreq))

	for i := 0; i < numSamples; i++ {
		var sample float64
		for voiceIdx, freq := range s.activeFreq {
			phaseStep := (2.0 * math.Pi * freq) / float64(s.sampleRate)
			sample += math.Sin(s.phases[voiceIdx])
			s.phases[voiceIdx] += phaseStep
			if s.phases[voiceIdx] >= 2.0*math.Pi {
				s.phases[voiceIdx] -= 2.0 * math.Pi
			}
		}

		sample = (sample / numVoices) * s.volume
		intSample := int16(sample * 32767.0)

		p[i*2] = byte(intSample)
		p[i*2+1] = byte(intSample >> 8)
	}

	return len(p), nil
}