package synth

import (
	"bytes"
	"encoding/binary"
	"math"
)

// Segment representa um trecho de áudio contínuo: uma nota/acorde sustentado (Freqs não vazio)
// ou um silêncio (Freqs vazio), com duração em milissegundos.
type Segment struct {
	Freqs      []float64
	DurationMS int64
}

// RenderSequence sintetiza uma sequência de segmentos (notas, acordes e silêncios) como um único
// buffer PCM estéreo de 16 bits contínuo.
//
// Sintetizar cada nota isoladamente com sua própria envolvente ADSR completa soa "picotado": mesmo
// quando duas notas são adjacentes (sem silêncio real entre elas), o volume cai quase a zero no fim
// de uma e sobe do zero no início da outra. Aqui o volume só sobe (attack) ao sair do silêncio e só
// desce (release) ao entrar em silêncio; entre notas adjacentes não-silenciosas a fase da onda e o
// nível de volume são preservados, e apenas a frequência muda — produzindo uma transição suave
// (legato) em vez de um novo ataque a cada nota.
func RenderSequence(segments []Segment) []byte {
	buf := new(bytes.Buffer)

	const attackSec = 0.015
	const releaseSec = 0.030
	const sustainLevel = 0.8

	sampleRateF := float64(SampleRate)
	attackSamples := int(attackSec * sampleRateF)
	releaseSamples := int(releaseSec * sampleRateF)

	var phases []float64

	for si, seg := range segments {
		numSamples := int(float64(SampleRate) * float64(seg.DurationMS) / 1000.0)
		if numSamples <= 0 {
			continue
		}

		if len(seg.Freqs) == 0 {
			phases = nil
			zero := make([]byte, numSamples*BytesPerSample*ChannelCount)
			buf.Write(zero)
			continue
		}

		steps := make([]float64, len(seg.Freqs))
		for i, f := range seg.Freqs {
			steps[i] = 2 * math.Pi * f / SampleRate
		}
		if len(phases) != len(steps) {
			phases = make([]float64, len(steps))
		}

		prevVoiced := si > 0 && len(segments[si-1].Freqs) > 0
		attackWindow := attackSamples
		if prevVoiced {
			attackWindow = 0 // já está soando: só troca a frequência, sem reatacar
		}

		nextVoiced := si+1 < len(segments) && len(segments[si+1].Freqs) > 0
		releaseWindow := releaseSamples
		if nextVoiced {
			releaseWindow = 0 // a próxima nota continua soando: não libera agora
		}

		for i := 0; i < numSamples; i++ {
			var mixed float64
			for vi := range steps {
				mixed += math.Sin(phases[vi])
				phases[vi] += steps[vi]
				if phases[vi] >= 2*math.Pi {
					phases[vi] -= 2 * math.Pi
				}
			}
			mixed /= float64(len(steps))

			envelope := sustainLevel
			switch {
			case attackWindow > 0 && i < attackWindow:
				envelope = sustainLevel * float64(i) / float64(attackWindow)
			case releaseWindow > 0 && i >= numSamples-releaseWindow:
				remaining := numSamples - i
				envelope = sustainLevel * float64(remaining) / float64(releaseWindow)
			}

			sampleValue := int16(mixed * envelope * 0.3 * 32767.0)
			_ = binary.Write(buf, binary.LittleEndian, sampleValue)
			_ = binary.Write(buf, binary.LittleEndian, sampleValue)
		}
	}

	return buf.Bytes()
}
