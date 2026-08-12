package scoring

import "math"

// Frame é o resultado da análise de pitch de uma pequena janela de tempo da gravação.
type Frame struct {
	TimeMS     int64   // tempo do início do frame, relativo ao início da gravação (WAV completo)
	FreqHz     float64 // frequência estimada; só significativa quando Voiced é true
	Confidence float64 // confiança do YIN (0-1)
	RMS        float64 // energia RMS do frame (amplitude normalizada -1..1)
	Voiced     bool    // passou no gate de vozeamento (energia + confiança do YIN)
}

// AnalysisConfig controla a janela de análise de pitch e o gate de vozeamento. DefaultAnalysisConfig
// cobre a faixa de uma voz cantada; os campos existem para permitir ajuste fino (testes, outras
// fontes de áudio) sem tocar no algoritmo.
type AnalysisConfig struct {
	SampleRate int

	// HopSize: passo entre frames sucessivos, em amostras — controla a resolução temporal da
	// análise (ex: detecção de onset).
	HopSize int
	// MinFreqHz/MaxFreqHz: faixa de frequência fundamental aceita pelo YIN. Delimita também o
	// tamanho da janela de análise necessária (frames mais longos pra frequências mais graves).
	MinFreqHz float64
	MaxFreqHz float64
	// YINThreshold: limiar absoluto do algoritmo YIN (ver detectPitchYIN) — quanto menor, mais
	// exigente com a periodicidade do sinal.
	YINThreshold float64

	// MinRMS: energia mínima do frame para ser considerado candidato a vozeado — descarta
	// silêncio puro antes mesmo de rodar o YIN.
	MinRMS float64
	// MinConfidence: confiança mínima do YIN para o frame ser considerado vozeado — descarta
	// ruído não-periódico (respiração, consoantes) que passou do gate de energia.
	MinConfidence float64
}

// DefaultAnalysisConfig retorna parâmetros calibrados para voz cantada (fundamental entre 70Hz —
// grave de voz masculina, com margem — e 1000Hz — agudo/falsete, com margem).
func DefaultAnalysisConfig(sampleRate int) AnalysisConfig {
	return AnalysisConfig{
		SampleRate:    sampleRate,
		HopSize:       sampleRate / 100, // ~10ms: resolução suficiente até para a tolerância de ritmo do nível hard (±80ms)
		MinFreqHz:     70,
		MaxFreqHz:     1000,
		YINThreshold:  0.15,
		MinRMS:        0.01,
		MinConfidence: 0.6,
	}
}

// maxTau é o maior período (em amostras) que o detector precisa cobrir, dado MinFreqHz.
func (c AnalysisConfig) maxTau() int {
	return int(float64(c.SampleRate) / c.MinFreqHz)
}

// AnalyzePCM roda a análise de pitch + gate de vozeamento sobre uma gravação PCM mono, amostra por
// amostra normalizada em [-1, 1], produzindo um Frame a cada HopSize amostras.
func AnalyzePCM(samples []float64, cfg AnalysisConfig) []Frame {
	maxTau := cfg.maxTau()
	windowSize := maxTau // tamanho da janela "atual"; o YIN também olha maxTau amostras à frente
	needed := 2 * maxTau
	if cfg.HopSize <= 0 || len(samples) < needed {
		return nil
	}

	var frames []Frame
	for start := 0; start+needed <= len(samples); start += cfg.HopSize {
		window := samples[start : start+windowSize]
		rms := rmsOf(window)

		frame := Frame{
			TimeMS: int64(start) * 1000 / int64(cfg.SampleRate),
			RMS:    rms,
		}

		if rms >= cfg.MinRMS {
			yr := detectPitchYIN(samples[start:start+needed], cfg.SampleRate, cfg.MinFreqHz, cfg.MaxFreqHz, cfg.YINThreshold)
			if yr.OK {
				frame.FreqHz = yr.FreqHz
				frame.Confidence = yr.Confidence
				frame.Voiced = yr.Confidence >= cfg.MinConfidence
			}
		}

		frames = append(frames, frame)
	}
	return frames
}

func rmsOf(window []float64) float64 {
	if len(window) == 0 {
		return 0
	}
	var sum float64
	for _, s := range window {
		sum += s * s
	}
	return math.Sqrt(sum / float64(len(window)))
}

// DecodePCM16Mono converte PCM mono 16-bit little-endian (o formato gravado por audio.MicRecorder
// e lido de volta por audio.ReadWAV) em amostras float64 normalizadas em [-1, 1].
func DecodePCM16Mono(pcm []byte) []float64 {
	n := len(pcm) / 2
	out := make([]float64, n)
	for i := range n {
		v := int16(uint16(pcm[2*i]) | uint16(pcm[2*i+1])<<8)
		out[i] = float64(v) / 32768.0
	}
	return out
}
