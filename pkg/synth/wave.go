package synth

import (
	"math"
	"time"
)

// WaveType define o tipo de forma de onda
type WaveType int

const (
	WaveSine WaveType = iota
	WaveTriangle
	WaveSquare
	WaveSawtooth
)

// Envelope do ADSR
type Envelope struct {
	Attack  time.Duration
	Decay   time.Duration
	Sustain float64 // Nível de 0.0 a 1.0
	Release time.Duration
}

// DefaultEnvelope retorna um envelope padrão suave para notas sintetizadas
func DefaultEnvelope() Envelope {
	return Envelope{
		Attack:  20 * time.Millisecond,
		Decay:   50 * time.Millisecond,
		Sustain: 0.7,
		Release: 100 * time.Millisecond,
	}
}

// AmplitudeEnvelope calcula a amplitude multiplicadora (0.0 a 1.0) baseada no tempo decorrido e duração total
func (env Envelope) AmplitudeEnvelope(t, totalDuration time.Duration) float64 {
	if t < 0 || t > totalDuration {
		return 0.0
	}

	attackSec := env.Attack.Seconds()
	decaySec := env.Decay.Seconds()
	releaseSec := env.Release.Seconds()
	totalSec := totalDuration.Seconds()
	tSec := t.Seconds()

	// Se a nota é mais curta do que o attack + release, ajusta proporcionalmente
	if totalSec < attackSec+releaseSec {
		sustainSec := math.Max(0, totalSec-attackSec-releaseSec)
		if tSec < attackSec && attackSec > 0 {
			return tSec / attackSec
		} else if tSec < attackSec+sustainSec {
			return 1.0
		} else if releaseSec > 0 {
			relPos := (tSec - attackSec - sustainSec) / releaseSec
			return math.Max(0.0, 1.0-relPos)
		}
		return 0.0
	}

	// Fase de Release (no final da nota)
	sustainEndSec := totalSec - releaseSec
	if tSec >= sustainEndSec {
		if releaseSec <= 0 {
			return 0.0
		}
		relPos := (tSec - sustainEndSec) / releaseSec
		return math.Max(0.0, env.Sustain*(1.0-relPos))
	}

	// Fase de Attack
	if tSec < attackSec {
		if attackSec <= 0 {
			return 1.0
		}
		return tSec / attackSec
	}

	// Fase de Decay
	if tSec < attackSec+decaySec {
		if decaySec <= 0 {
			return env.Sustain
		}
		decayPos := (tSec - attackSec) / decaySec
		return 1.0 - (1.0-env.Sustain)*decayPos
	}

	// Fase de Sustain
	return env.Sustain
}

// GenerateWaveSample gera o valor da onda normalizado (-1.0 a 1.0) para uma dada frequência e tempo t
func GenerateWaveSample(waveType WaveType, freq float64, t float64) float64 {
	if freq <= 0 {
		return 0.0
	}
	phase := math.Mod(freq*t, 1.0)

	switch waveType {
	case WaveSine:
		return math.Sin(2 * math.Pi * freq * t)
	case WaveTriangle:
		if phase < 0.5 {
			return 4.0*phase - 1.0
		}
		return 3.0 - 4.0*phase
	case WaveSquare:
		if phase < 0.5 {
			return 1.0
		}
		return -1.0
	case WaveSawtooth:
		return 2.0*phase - 1.0
	default:
		return math.Sin(2 * math.Pi * freq * t)
	}
}

// GeneratePolyphonicSample calcula a amostra instantânea somando múltiplas frequências ativas e normalizando pelo número de vozes
func GeneratePolyphonicSample(freqs []float64, t float64) float64 {
	if len(freqs) == 0 {
		return 0.0
	}

	var sum float64
	for _, freq := range freqs {
		if freq > 0 {
			sum += math.Sin(2 * math.Pi * freq * t)
		}
	}

	return sum / float64(len(freqs))
}
