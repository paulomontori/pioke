package synth

import (
	"bytes"
	"encoding/binary"
	"math/rand"

	"pioke/pkg/model"
)

// KarplusStrongDecay controla o quão rápido a corda "morre" depois de dedilhada: mais perto de
// 1.0 = corda mais viva, tocando por mais tempo; mais perto de 0.9 = mais abafada/curta. Faixa
// útil: ~0.990-0.999.
var KarplusStrongDecay = 0.996

// karplusStrongPluck sintetiza numSamples de uma corda dedilhada na frequência freq usando
// Karplus-Strong: uma linha de atraso do tamanho de um período da onda, inicializada com ruído (a
// excitação do dedilhado) e realimentada por uma média móvel com decaimento. O filtro passa-baixa
// da média suaviza o ruído a cada volta pela linha — o resultado é um ataque brilhante/percussivo
// decaindo pra um tom mais redondo, igual a uma corda dedilhada de verdade, sem precisar de
// envoltória ADSR desenhada à mão.
func karplusStrongPluck(freq float64, numSamples int) []float64 {
	if freq <= 0 || numSamples <= 0 {
		return nil
	}

	period := int(float64(SampleRate) / freq)
	if period < 2 {
		period = 2
	}

	buf := make([]float64, period)
	for i := range buf {
		buf[i] = rand.Float64()*2 - 1
	}

	out := make([]float64, numSamples)
	pos := 0
	for i := 0; i < numSamples; i++ {
		out[i] = buf[pos]
		next := (pos + 1) % period
		buf[pos] = KarplusStrongDecay * 0.5 * (buf[pos] + buf[next])
		pos = next
	}
	return out
}

// RenderSequenceKS sintetiza uma sequência de segmentos usando Karplus-Strong em vez da síntese
// aditiva de RenderSequence. Fisicamente, uma corda dedilhada é sempre um novo ataque — mesmo uma
// nota "legato" na partitura é uma nova dedilhada na vida real — então, ao contrário de
// RenderSequence, aqui CADA segmento dedilha suas próprias cordas do zero, sem herdar fase/estado
// do segmento anterior.
func RenderSequenceKS(segments []Segment) []byte {
	buf := new(bytes.Buffer)

	const sustainLevel = 0.8

	for _, seg := range segments {
		numSamples := int(float64(SampleRate) * float64(seg.DurationMS) / 1000.0)
		if numSamples <= 0 {
			continue
		}

		if len(seg.Freqs) == 0 {
			zero := make([]byte, numSamples*BytesPerSample*ChannelCount)
			buf.Write(zero)
			continue
		}

		voices := make([][]float64, 0, len(seg.Freqs))
		for _, f := range seg.Freqs {
			if pluck := karplusStrongPluck(f, numSamples); pluck != nil {
				voices = append(voices, pluck)
			}
		}
		if len(voices) == 0 {
			zero := make([]byte, numSamples*BytesPerSample*ChannelCount)
			buf.Write(zero)
			continue
		}

		for i := 0; i < numSamples; i++ {
			var mixed float64
			for _, v := range voices {
				mixed += v[i]
			}
			mixed /= float64(len(voices))

			sample := mixed * sustainLevel * 0.3 * 32767.0
			switch {
			case sample > 32767:
				sample = 32767
			case sample < -32768:
				sample = -32768
			}

			sampleValue := int16(sample)
			_ = binary.Write(buf, binary.LittleEndian, sampleValue)
			_ = binary.Write(buf, binary.LittleEndian, sampleValue)
		}
	}

	return buf.Bytes()
}

// RenderSongKS espelha RenderSong, mas usando RenderSequenceKS (Karplus-Strong) para sintetizar
// tanto a melodia quanto o acompanhamento.
func RenderSongKS(s *model.Song) []byte {
	melody := RenderSequenceKS(BuildSegments(s))

	accompaniment := buildAccompanimentSegments(s)
	if len(accompaniment) == 0 {
		return melody
	}

	return MixPCM(melody, RenderSequenceKS(accompaniment))
}
