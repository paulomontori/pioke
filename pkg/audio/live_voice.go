package audio

import (
	"math"
	"sync"
)

// liveVoice é um io.Reader que sintetiza continuamente as frequências ativas no momento (uma
// nota, um acorde, ou nenhuma). Ao contrário de criar um novo oto.Player — com sua própria
// envoltória ADSR completa — para cada nota, aqui um único player de longa duração lê deste
// stream: SetFreqs troca a frequência tocada preservando a fase e o nível de volume atual.
// Só há rampa de subida (attack) ao sair do silêncio e de descida (release) ao entrar em
// silêncio; trocar de uma nota para a próxima (sem silêncio real entre elas) não reataca a
// envoltória — mesma ideia de synth.RenderSequence, aplicada em tempo real.
type liveVoice struct {
	mu         sync.Mutex
	sampleRate int
	freqs      []float64
	phases     []float64
	envelope   float64 // amplitude atual, 0.0 a liveSustainLevel
}

const (
	liveSustainLevel = 0.8
	liveAttackSec    = 0.015
	liveReleaseSec   = 0.030
)

func newLiveVoice(sampleRate int) *liveVoice {
	return &liveVoice{sampleRate: sampleRate}
}

// SetFreqs troca as frequências ativas (nil para silêncio). Reaproveita as fases existentes
// quando o número de vozes não muda, para uma transição sem estalos entre notas adjacentes.
func (v *liveVoice) SetFreqs(freqs []float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.freqs = freqs
	if len(v.phases) != len(freqs) {
		v.phases = make([]float64, len(freqs))
	}
}

// Read implementa io.Reader preenchendo PCM estéreo de 16 bits (usado pelo oto.Player).
func (v *liveVoice) Read(p []byte) (int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	numSamples := len(p) / 4 // 16-bit estéreo = 4 bytes por amostra
	sampleRateF := float64(v.sampleRate)
	attackStep := liveSustainLevel / (liveAttackSec * sampleRateF)
	releaseStep := liveSustainLevel / (liveReleaseSec * sampleRateF)

	voiced := len(v.freqs) > 0
	target := 0.0
	if voiced {
		target = liveSustainLevel
	}

	for i := 0; i < numSamples; i++ {
		switch {
		case v.envelope < target:
			v.envelope += attackStep
			if v.envelope > target {
				v.envelope = target
			}
		case v.envelope > target:
			v.envelope -= releaseStep
			if v.envelope < target {
				v.envelope = target
			}
		}

		var mixed float64
		if voiced {
			for vi, freq := range v.freqs {
				step := 2 * math.Pi * freq / sampleRateF
				mixed += math.Sin(v.phases[vi])
				v.phases[vi] += step
				if v.phases[vi] >= 2*math.Pi {
					v.phases[vi] -= 2 * math.Pi
				}
			}
			mixed /= float64(len(v.freqs))
		}

		sampleValue := int16(mixed * v.envelope * 0.3 * 32767.0)
		p[i*4] = byte(sampleValue)
		p[i*4+1] = byte(sampleValue >> 8)
		p[i*4+2] = byte(sampleValue)
		p[i*4+3] = byte(sampleValue >> 8)
	}

	return numSamples * 4, nil
}
