package scoring

import (
	"math"
	"sort"

	"pioke/pkg/model"
)

// NoteBreakdown é o resultado da comparação de uma ReferenceNote com a gravação — o suficiente
// pra UI destacar quais trechos foram bem ou mal cantados.
type NoteBreakdown struct {
	ReferenceNote

	Covered         bool    // houve voz detectada na janela de tempo desta nota
	CentsDeviation  float64 // desvio médio em cents (já com correção de oitava aplicada, se o preset corrigir); só significativo quando Covered
	OctaveCorrected bool    // a correção de oitava mudou o resultado (só relevante fora de OctaveStrict)
	PitchScore      float64 // 0-100

	OnsetDeviationMS int64 // onset detectado - start_ms de referência (ms; negativo = adiantado); só significativo quando Covered
	RhythmScore      float64 // 0-100

	Score float64 // combinação de PitchScore e RhythmScore para esta nota (pesos do preset, renormalizados)
}

// Result é o resultado completo da avaliação de uma gravação contra a melodia de referência de
// uma música, sob um ScoringPreset.
type Result struct {
	Preset ScoringPreset

	FinalScore    float64 // 0-100
	PitchScore    float64 // 0-100, média entre todas as notas
	RhythmScore   float64 // 0-100, média entre todas as notas
	CoverageScore float64 // 0-100, % de notas com voz detectada

	Notes []NoteBreakdown
}

// Score compara a gravação (recordingPCM: PCM mono 16-bit little-endian, o mesmo formato salvo por
// audio.MicRecorder/audio.WriteWAV) contra a melodia vocal de referência extraída de song, sob o
// preset de dificuldade informado.
//
// playbackOffsetMS é o RecordingMeta.PlaybackOffsetMS salvo junto da gravação (ver
// pkg/playback.RecordingMeta): quantos ms depois do início da gravação a reprodução da música
// realmente começou — usado para converter o tempo da timeline (relativo ao início da música) em
// tempo da gravação (relativo ao início do WAV captado).
func Score(song *model.Song, recordingPCM []byte, sampleRate int, playbackOffsetMS int64, preset ScoringPreset) Result {
	notes := ExtractReferenceNotes(song)
	cfg := DefaultAnalysisConfig(sampleRate)
	frames := AnalyzePCM(DecodePCM16Mono(recordingPCM), cfg)

	breakdown := make([]NoteBreakdown, len(notes))
	var sumPitch, sumRhythm float64
	var covered int

	for i, note := range notes {
		startRec := note.StartMS + playbackOffsetMS
		endRec := startRec + note.DurationMS

		prevEndRec := int64(math.MinInt64 / 2)
		if i > 0 {
			prevEndRec = notes[i-1].StartMS + playbackOffsetMS + notes[i-1].DurationMS
		}
		nextStartRec := int64(math.MaxInt64 / 2)
		if i+1 < len(notes) {
			nextStartRec = notes[i+1].StartMS + playbackOffsetMS
		}

		nb := scoreNote(note, frames, startRec, endRec, prevEndRec, nextStartRec, preset)
		breakdown[i] = nb

		sumPitch += nb.PitchScore
		sumRhythm += nb.RhythmScore
		if nb.Covered {
			covered++
		}
	}

	result := Result{Preset: preset, Notes: breakdown}
	if len(notes) == 0 {
		return result
	}

	result.PitchScore = sumPitch / float64(len(notes))
	result.RhythmScore = sumRhythm / float64(len(notes))
	result.CoverageScore = 100 * float64(covered) / float64(len(notes))
	result.FinalScore = (preset.WeightPitch*result.PitchScore +
		preset.WeightRhythm*result.RhythmScore +
		preset.WeightCoverage*result.CoverageScore) / 100

	return result
}

// scoreNote avalia uma única nota de referência contra os frames vozados da gravação.
//
// startRec/endRec delimitam a janela exata da nota (tempo da gravação) usada para a comparação de
// afinação e para decidir se a nota está "coberta" (teve voz detectada), conforme pedido: "pega os
// frames vozados da gravação na janela de tempo daquela nota". prevEndRec/nextStartRec (fim da
// nota anterior / início da próxima, também em tempo da gravação) limitam uma busca mais ampla de
// onset para a pontuação de ritmo — deliberadamente mais larga que a janela de afinação, já que um
// desvio de tempo grande é exatamente o que essa pontuação existe para capturar; se a busca de
// onset usasse a mesma janela estrita, cantar adiantado ou atrasado o bastante para escapar da
// janela zeraria o ritmo por "cobertura" em vez de por desvio, escondendo o motivo real.
func scoreNote(note ReferenceNote, frames []Frame, startRec, endRec, prevEndRec, nextStartRec int64, preset ScoringPreset) NoteBreakdown {
	nb := NoteBreakdown{ReferenceNote: note}

	pitchFrames := framesInWindow(frames, startRec, endRec)
	nb.Covered = len(pitchFrames) > 0
	if nb.Covered {
		nb.CentsDeviation, nb.OctaveCorrected, nb.PitchScore = scorePitch(pitchFrames, note.FreqHz, preset)
	}

	searchStart := max(startRec-2*preset.TimeToleranceMS, prevEndRec)
	searchEnd := min(endRec, nextStartRec)
	onsetFrames := framesInWindow(frames, searchStart, searchEnd)
	if onsetMS, hasOnset := firstVoicedOnset(onsetFrames); hasOnset {
		nb.OnsetDeviationMS = onsetMS - startRec
		nb.RhythmScore = scoreFromDeviation(float64(abs64(nb.OnsetDeviationMS)), float64(preset.TimeToleranceMS))
	}

	if nb.Covered {
		weightSum := preset.WeightPitch + preset.WeightRhythm
		if weightSum > 0 {
			nb.Score = (preset.WeightPitch*nb.PitchScore + preset.WeightRhythm*nb.RhythmScore) / weightSum
		}
	}

	return nb
}

// framesInWindow retorna os frames vozados cujo TimeMS cai em [startMS, endMS). frames é assumido
// ordenado por TimeMS (garantido por AnalyzePCM).
func framesInWindow(frames []Frame, startMS, endMS int64) []Frame {
	lo := sort.Search(len(frames), func(i int) bool { return frames[i].TimeMS >= startMS })
	hi := sort.Search(len(frames), func(i int) bool { return frames[i].TimeMS >= endMS })
	if lo >= hi {
		return nil
	}

	var out []Frame
	for _, f := range frames[lo:hi] {
		if f.Voiced {
			out = append(out, f)
		}
	}
	return out
}

// firstVoicedOnset retorna o TimeMS do primeiro frame vozado (a lista já vem filtrada por
// framesInWindow, então basta pegar o primeiro elemento).
func firstVoicedOnset(voiced []Frame) (int64, bool) {
	if len(voiced) == 0 {
		return 0, false
	}
	return voiced[0].TimeMS, true
}

// scorePitch calcula o desvio em cents (com correção de oitava conforme preset.OctaveErrorMode) e
// a pontuação de afinação para uma nota, a partir dos frames vozados na janela dessa nota.
func scorePitch(voiced []Frame, targetFreq float64, preset ScoringPreset) (avgCents float64, octaveCorrected bool, pitchScore float64) {
	if len(voiced) == 0 || targetFreq <= 0 {
		return 0, false, 0
	}

	var sumCents, sumScore float64
	var anyShift bool
	for _, f := range voiced {
		if f.FreqHz <= 0 {
			continue
		}
		rawCents := 1200 * math.Log2(f.FreqHz/targetFreq)

		cents := rawCents
		shifted := false
		if preset.OctaveErrorMode != OctaveStrict {
			k := math.Round(rawCents / 1200)
			cents = rawCents - k*1200
			shifted = k != 0
		}

		frameScore := scoreFromCents(math.Abs(cents), preset.CentsFullScoreAbs, preset.CentsZeroScoreAbs)
		if shifted && preset.OctaveErrorMode == OctavePenalize {
			frameScore *= preset.OctavePenaltyFactor
		}

		sumCents += cents
		sumScore += frameScore
		anyShift = anyShift || shifted
	}

	n := float64(len(voiced))
	return sumCents / n, anyShift, sumScore / n
}

// scoreFromCents mapeia um desvio de afinação (cents, valor absoluto) para uma pontuação 0-100:
// máxima até fullScoreAbs, zero a partir de zeroScoreAbs, linear entre os dois.
func scoreFromCents(absCents, fullScoreAbs, zeroScoreAbs float64) float64 {
	switch {
	case absCents <= fullScoreAbs:
		return 100
	case absCents >= zeroScoreAbs:
		return 0
	default:
		return 100 * (zeroScoreAbs - absCents) / (zeroScoreAbs - fullScoreAbs)
	}
}

// scoreFromDeviation mapeia um desvio absoluto (ex: ms de onset) para uma pontuação 0-100: 100 em
// desvio 0, caindo linearmente até 0 em desvio == tolerance.
func scoreFromDeviation(absDeviation, tolerance float64) float64 {
	if tolerance <= 0 {
		if absDeviation <= 0 {
			return 100
		}
		return 0
	}
	score := 100 * (1 - absDeviation/tolerance)
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
