package synth

import (
	"math"
	"strings"
)

// NoteToFrequency converte nota MIDI (ex: 60 = C4, 69 = A4) em frequência Hz
func NoteToFrequency(midiNote int) float64 {
	return 440.0 * math.Pow(2.0, float64(midiNote-69)/12.0)
}

// GetChordFrequencies recebe uma cifra (ex: "C", "Am", "G7", "Em") e retorna as frequências Hz das notas
func GetChordFrequencies(chord string) []float64 {
	chord = strings.TrimSpace(chord)
	if chord == "" {
		return nil
	}

	// Mapeamento de notas base para MIDI offset (C4 = 60)
	baseNotes := map[string]int{
		"C": 60, "C#": 61, "Db": 61,
		"D": 62, "D#": 63, "Eb": 63,
		"E": 64,
		"F": 65, "F#": 66, "Gb": 66,
		"G": 67, "G#": 68, "Ab": 68,
		"A": 69, "A#": 70, "Bb": 70,
		"B": 71,
	}

	root := ""
	quality := ""

	// Extrai nota fundamental
	if len(chord) >= 2 && (chord[1] == '#' || chord[1] == 'b') {
		root = chord[:2]
		quality = chord[2:]
	} else {
		root = chord[:1]
		quality = chord[1:]
	}

	baseMidi, exists := baseNotes[strings.Title(root)]
	if !exists {
		return nil
	}

	// Intervalos em semitons em relação à tônica
	var intervals []int

	switch quality {
	case "", "maj", "M": // Maior (1, 3, 5)
		intervals = []int{0, 4, 7}
	case "m", "min": // Menor (1, b3, 5)
		intervals = []int{0, 3, 7}
	case "7", "dom7": // Sétima dominante (1, 3, 5, b7)
		intervals = []int{0, 4, 7, 10}
	case "m7", "min7": // Menor com sétima (1, b3, 5, b7)
		intervals = []int{0, 3, 7, 10}
	case "maj7", "M7": // Maior com sétima (1, 3, 5, 7)
		intervals = []int{0, 4, 7, 11}
	case "sus4": // Suspensa quarta (1, 4, 5)
		intervals = []int{0, 5, 7}
	case "sus2": // Suspensa segunda (1, 2, 5)
		intervals = []int{0, 2, 7}
	case "dim": // Diminuto (1, b3, b5)
		intervals = []int{0, 3, 6}
	default:
		// Fallback para tríade maior se o sufixo não for totalmente reconhecido
		if strings.HasPrefix(quality, "m") {
			intervals = []int{0, 3, 7}
		} else {
			intervals = []int{0, 4, 7}
		}
	}

	freqs := make([]float64, len(intervals))
	for i, interval := range intervals {
		freqs[i] = NoteToFrequency(baseMidi + interval)
	}

	return freqs
}
