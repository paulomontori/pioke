package synth

import (
	"sync"
	"time"
)

// Synthesizer define a interface para geração polifônica de áudio sintetizado
type Synthesizer interface {
	PlayChord(name string, duration time.Duration)
	PlayNote(freq float64, duration time.Duration)
	ReadPCM(samples []float32) int
	Stop()
}

type ActiveVoice struct {
	Frequencies []float64
	StartTime   time.Time
	Duration    time.Duration
	Envelope    Envelope
	WaveType    WaveType
}

// PolySynth gerencia vozes ativas e gera buffer PCM polifônico sem clipping
type PolySynth struct {
	mu         sync.Mutex
	voices     []ActiveVoice
	sampleRate int
	sampleTime float64
	waveType   WaveType
	envelope   Envelope
}

// NewPolySynth cria uma nova instância do sintetizador polifônico
func NewPolySynth(sampleRate int) *PolySynth {
	return &PolySynth{
		sampleRate: sampleRate,
		waveType:   WaveSine,
		envelope:   DefaultEnvelope(),
		voices:     make([]ActiveVoice, 0),
	}
}

func (s *PolySynth) PlayChord(name string, duration time.Duration) {
	freqs := GetChordFrequencies(name)
	if len(freqs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.voices = append(s.voices, ActiveVoice{
		Frequencies: freqs,
		StartTime:   time.Now(),
		Duration:    duration,
		Envelope:    s.envelope,
		WaveType:    s.waveType,
	})
}

func (s *PolySynth) PlayNote(freq float64, duration time.Duration) {
	if freq <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.voices = append(s.voices, ActiveVoice{
		Frequencies: []float64{freq},
		StartTime:   time.Now(),
		Duration:    duration,
		Envelope:    s.envelope,
		WaveType:    s.waveType,
	})
}

// ReadPCM preenche o buffer de amostragem em float32 (-1.0 a 1.0)
func (s *PolySynth) ReadPCM(samples []float32) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	dt := 1.0 / float64(s.sampleRate)
	now := time.Now()

	activeVoices := make([]ActiveVoice, 0, len(s.voices))

	for _, v := range s.voices {
		if now.Sub(v.StartTime) <= v.Duration+v.Envelope.Release {
			activeVoices = append(activeVoices, v)
		}
	}
	s.voices = activeVoices

	for i := range samples {
		var sampleAcc float64

		for _, voice := range s.voices {
			elapsed := now.Sub(voice.StartTime) + time.Duration(float64(i)*dt*float64(time.Second))
			envAmp := voice.Envelope.AmplitudeEnvelope(elapsed, voice.Duration)

			if envAmp <= 0 {
				continue
			}

			var chordAcc float64
			for _, freq := range voice.Frequencies {
				t := s.sampleTime + float64(i)*dt
				chordAcc += GenerateWaveSample(voice.WaveType, freq, t)
			}
			if len(voice.Frequencies) > 0 {
				chordAcc /= float64(len(voice.Frequencies))
			}

			sampleAcc += chordAcc * envAmp
		}

		// Garante atenuação para evitar ruídos e clipping quando há múltiplas vozes
		if len(s.voices) > 1 {
			sampleAcc /= float64(len(s.voices))
		}

		// Clamp no intervalo [-1.0, 1.0]
		if sampleAcc > 1.0 {
			sampleAcc = 1.0
		} else if sampleAcc < -1.0 {
			sampleAcc = -1.0
		}

		samples[i] = float32(sampleAcc)
	}

	s.sampleTime += float64(len(samples)) * dt
	return len(samples)
}

func (s *PolySynth) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voices = nil
}
