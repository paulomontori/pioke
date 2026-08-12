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
