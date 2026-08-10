package synth

import (
	"testing"
	"time"
)

func TestGetChordFrequencies(t *testing.T) {
	freqsC := GetChordFrequencies("C")
	if len(freqsC) != 3 {
		t.Fatalf("Esperado 3 frequências para o acorde C, obtido %d", len(freqsC))
	}

	freqsAm7 := GetChordFrequencies("Am7")
	if len(freqsAm7) != 4 {
		t.Fatalf("Esperado 4 frequências para Am7, obtido %d", len(freqsAm7))
	}
}

func TestPolySynthReadPCM(t *testing.T) {
	s := NewPolySynth(44100)
	s.PlayChord("C", 200*time.Millisecond)

	buf := make([]float32, 512)
	n := s.ReadPCM(buf)
	if n != 512 {
		t.Errorf("Esperado ler 512 amostras, lido %d", n)
	}

	for i, sample := range buf {
		if sample < -1.0 || sample > 1.0 {
			t.Errorf("Amostra no índice %d fora do intervalo [-1.0, 1.0]: %f", i, sample)
		}
	}
}

func TestEnvelopeADSR(t *testing.T) {
	env := DefaultEnvelope()
	total := 500 * time.Millisecond

	// Início no attack (amplitude crescendo)
	amp0 := env.AmplitudeEnvelope(0, total)
	ampAttack := env.AmplitudeEnvelope(10*time.Millisecond, total)
	if ampAttack <= amp0 {
		t.Errorf("Esperado crescimento no attack, amp0=%f, ampAttack=%f", amp0, ampAttack)
	}

	// Fora da duração total + release
	ampEnd := env.AmplitudeEnvelope(700*time.Millisecond, total)
	if ampEnd != 0.0 {
		t.Errorf("Esperado amplitude 0.0 após release completo, obtido %f", ampEnd)
	}
}
