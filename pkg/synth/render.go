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
// buffer PCM estéreo de 16 bits contínuo — uma única linha/voz por vez (ex: só a melodia, ou só o
// acompanhamento); para tocar vozes simultâneas juntas, renderize cada uma separadamente e some
// com MixPCM.
//
// Sintetizar cada nota isoladamente com sua própria envolvente ADSR completa soa "picotado": mesmo
// quando duas notas são adjacentes (sem silêncio real entre elas), o volume cai quase a zero no fim
// de uma e sobe do zero no início da outra. Aqui o volume só sobe (attack) ao sair do silêncio e só
// desce (release) ao entrar em silêncio; entre notas adjacentes não-silenciosas a fase da onda e o
// nível de volume são preservados, e apenas a frequência muda — produzindo uma transição suave
// (legato) em vez de um novo ataque a cada nota.
//
// Cada nota é síntese aditiva (fundamental + harmônicos com peso decrescente, ver
// HarmonicAmplitudes) em vez de uma senoide pura — timbre mais próximo de uma corda dedilhada do
// que de um bipe eletrônico. A fase do harmônico n é derivada da fase da fundamental (n × phase)
// em vez de um acumulador próprio: matematicamente idêntico a integrar n×f diretamente (sin é
// 2π-periódico, então a redução módulo 2π da fundamental comuta com a multiplicação por n), sem
// custo extra de estado.
func RenderSequence(segments []Segment) []byte {
	buf := new(bytes.Buffer)

	const attackSec = 0.015
	const releaseSec = 0.030
	const sustainLevel = 0.8

	sampleRateF := float64(SampleRate)
	attackSamples := int(attackSec * sampleRateF)
	releaseSamples := int(releaseSec * sampleRateF)
	nyquist := sampleRateF / 2

	harmonics := HarmonicAmplitudes
	if len(harmonics) == 0 {
		harmonics = []float64{1.0}
	}

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

		// Harmônicos acima do Nyquist (SampleRate/2) dobram (aliasing) em vez de simplesmente
		// não soar — melhor deixar de fora do que distorcer. maxHarmonics[vi] é quantos
		// harmônicos de HarmonicAmplitudes cabem abaixo do Nyquist pra frequência da voz vi;
		// varia por nota (grave cabe todos, aguda cabe menos), sempre ao menos a fundamental.
		maxHarmonics := make([]int, len(seg.Freqs))
		harmonicWeightSum := make([]float64, len(seg.Freqs))
		for i, f := range seg.Freqs {
			limit := len(harmonics)
			if f > 0 {
				if byNyquist := int(nyquist / f); byNyquist < limit {
					limit = byNyquist
				}
			}
			if limit < 1 {
				limit = 1
			}
			maxHarmonics[i] = limit
			var sum float64
			for h := 0; h < limit; h++ {
				sum += harmonics[h]
			}
			if sum <= 0 {
				sum = 1
			}
			harmonicWeightSum[i] = sum
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
				var voice float64
				limit := maxHarmonics[vi]
				for h := 0; h < limit; h++ {
					voice += harmonics[h] * math.Sin(float64(h+1)*phases[vi])
				}
				mixed += voice / harmonicWeightSum[vi]

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

// MixPCM soma duas trilhas PCM estéreo 16-bit amostra a amostra (a trilha mais curta é
// completada com silêncio), arredondando (clipping) para o range de int16 quando a soma estoura.
func MixPCM(a, b []byte) []byte {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	// mantém alinhado a um número inteiro de amostras estéreo 16-bit (4 bytes por amostra)
	n -= n % 4

	out := make([]byte, n)
	for i := 0; i+1 < n; i += 2 {
		var av, bv int32
		if i+1 < len(a) {
			av = int32(int16(uint16(a[i]) | uint16(a[i+1])<<8))
		}
		if i+1 < len(b) {
			bv = int32(int16(uint16(b[i]) | uint16(b[i+1])<<8))
		}
		sum := av + bv
		switch {
		case sum > 32767:
			sum = 32767
		case sum < -32768:
			sum = -32768
		}
		out[i] = byte(uint16(sum))
		out[i+1] = byte(uint16(sum) >> 8)
	}
	return out
}
