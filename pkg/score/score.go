package score

import (
	"math"
)

// PitchScorer calcula pontuações baseadas na precisão da afinação cantada.
type PitchScorer struct {
	totalScore float64
	samples    int
}

// NewPitchScorer inicializa um novo avaliador de pontuação.
func NewPitchScorer() *PitchScorer {
	return &PitchScorer{}
}

// FreqToMIDINote converte uma frequência em Hz para a nota MIDI mais próxima.
func FreqToMIDINote(freq float64) int {
	if freq <= 0 {
		return 0
	}
	return int(math.Round(69 + 12*math.Log2(freq/440.0)))
}

// CalculateAccuracy calcula a precisão (0 a 100) comparando a frequência cantada com a esperada.
func (ps *PitchScorer) CalculateAccuracy(targetFreq, sungFreq float64) float64 {
	if targetFreq <= 0 || sungFreq <= 0 {
		return 0.0
	}

	// Calcula a diferença em semitons
	diffSemitones := math.Abs(12 * math.Log2(sungFreq/targetFreq))

	// Se a diferença for maior que 2 semitons, pontuação é 0
	if diffSemitones >= 2.0 {
		return 0.0
	}

	// Pontuação linear de 100 (afinação perfeita) a 0 (2 semitons de diferença)
	accuracy := (1.0 - (diffSemitones / 2.0)) * 100.0
	return accuracy
}

// AddSample registra uma nova amostra e acumula na pontuação média.
func (ps *PitchScorer) AddSample(targetFreq, sungFreq float64) float64 {
	acc := ps.CalculateAccuracy(targetFreq, sungFreq)
	ps.totalScore += acc
	ps.samples++
	return acc
}

// GetAverageScore retorna a pontuação média final acumulada (0 a 100).
func (ps *PitchScorer) GetAverageScore() float64 {
	if ps.samples == 0 {
		return 0.0
	}
	return ps.totalScore / float64(ps.samples)
}

// Reset reinicia os contadores de pontuação.
func (ps *PitchScorer) Reset() {
	ps.totalScore = 0.0
	ps.samples = 0
}
