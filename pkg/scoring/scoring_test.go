package scoring

import (
	"math"
	"testing"

	"pioke/pkg/model"
)

const testSampleRate = 44100

// genSine gera durationMS de senoide pura em freqHz, amostras float64 normalizadas em [-1,1].
func genSine(freqHz float64, durationMS int, sampleRate int) []float64 {
	n := durationMS * sampleRate / 1000
	out := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(sampleRate)
		out[i] = 0.5 * math.Sin(2*math.Pi*freqHz*t)
	}
	return out
}

func genSilence(durationMS int, sampleRate int) []float64 {
	return make([]float64, durationMS*sampleRate/1000)
}

func encodePCM16Mono(samples []float64) []byte {
	out := make([]byte, len(samples)*2)
	for i, s := range samples {
		v := int16(s * 32767)
		out[2*i] = byte(uint16(v))
		out[2*i+1] = byte(uint16(v) >> 8)
	}
	return out
}

func TestDetectPitchYIN_KnownFrequency(t *testing.T) {
	const freq = 220.0 // A3
	samples := genSine(freq, 200, testSampleRate)

	cfg := DefaultAnalysisConfig(testSampleRate)
	maxTau := cfg.maxTau()
	result := detectPitchYIN(samples[:2*maxTau], testSampleRate, cfg.MinFreqHz, cfg.MaxFreqHz, cfg.YINThreshold)

	if !result.OK {
		t.Fatalf("esperava detecção bem-sucedida para senoide de %.0fHz", freq)
	}
	if math.Abs(result.FreqHz-freq) > 1.0 {
		t.Errorf("frequência detectada = %.2fHz, esperado ~%.2fHz", result.FreqHz, freq)
	}
	if result.Confidence < 0.9 {
		t.Errorf("confiança = %.2f, esperado alta confiança para senoide pura", result.Confidence)
	}
}

func TestAnalyzePCM_SilenceIsUnvoiced(t *testing.T) {
	cfg := DefaultAnalysisConfig(testSampleRate)
	frames := AnalyzePCM(genSilence(500, testSampleRate), cfg)

	if len(frames) == 0 {
		t.Fatal("esperava frames retornados para 500ms de silêncio")
	}
	for _, f := range frames {
		if f.Voiced {
			t.Errorf("frame em t=%dms marcado como vozeado durante silêncio", f.TimeMS)
		}
	}
}

// buildTestSong monta uma música de duas notas (A3=220Hz, D4=293.66Hz) de 400ms cada, começando
// em t=0 e t=400ms — o suficiente pra exercitar a extração de referência e o scoring nota-a-nota.
func buildTestSong() *model.Song {
	return &model.Song{
		Title: "Teste",
		Timeline: []model.TimelineEvent{
			{
				TimeMS: 0,
				Syllables: []model.Syllable{
					{Text: "la", OffsetMS: 0, DurationMS: 400, Pitch: "A3"},
					{Text: "re", OffsetMS: 400, DurationMS: 400, Pitch: "D4"},
				},
			},
		},
	}
}

func TestExtractReferenceNotes(t *testing.T) {
	notes := ExtractReferenceNotes(buildTestSong())
	if len(notes) != 2 {
		t.Fatalf("esperava 2 notas de referência, obteve %d", len(notes))
	}
	if notes[0].StartMS != 0 || notes[0].DurationMS != 400 || notes[0].Pitch != "A3" {
		t.Errorf("nota 0 = %+v, inesperada", notes[0])
	}
	if notes[1].StartMS != 400 || notes[1].DurationMS != 400 || notes[1].Pitch != "D4" {
		t.Errorf("nota 1 = %+v, inesperada", notes[1])
	}
	if math.Abs(notes[0].FreqHz-220.0) > 0.5 {
		t.Errorf("FreqHz da nota 0 = %.2f, esperado ~220Hz (A3)", notes[0].FreqHz)
	}
}

// buildRecording sintetiza uma "gravação" cantando cada nota de referência na frequência dada por
// sungFreqOf (permite simular afinação perfeita, desafinada, ou silêncio por nota), com o onset
// exatamente no tempo esperado.
func buildRecording(notes []ReferenceNote, sungFreqOf func(i int) float64) []byte {
	var samples []float64
	var cursorMS int64
	for i, n := range notes {
		if n.StartMS > cursorMS {
			samples = append(samples, genSilence(int(n.StartMS-cursorMS), testSampleRate)...)
			cursorMS = n.StartMS
		}
		freq := sungFreqOf(i)
		if freq <= 0 {
			samples = append(samples, genSilence(int(n.DurationMS), testSampleRate)...)
		} else {
			samples = append(samples, genSine(freq, int(n.DurationMS), testSampleRate)...)
		}
		cursorMS += n.DurationMS
	}
	return encodePCM16Mono(samples)
}

func TestScore_PerfectPitch(t *testing.T) {
	song := buildTestSong()
	notes := ExtractReferenceNotes(song)
	pcm := buildRecording(notes, func(i int) float64 { return notes[i].FreqHz })

	result := Score(song, pcm, testSampleRate, 0, PresetMedium)

	if result.FinalScore < 90 {
		t.Errorf("FinalScore = %.1f, esperado >= 90 para canto afinado e no tempo", result.FinalScore)
	}
	if result.CoverageScore != 100 {
		t.Errorf("CoverageScore = %.1f, esperado 100", result.CoverageScore)
	}
	for _, nb := range result.Notes {
		if !nb.Covered {
			t.Errorf("nota %q não coberta, esperado cobertura total", nb.Pitch)
		}
		if nb.PitchScore < 90 {
			t.Errorf("nota %q PitchScore = %.1f, esperado alto", nb.Pitch, nb.PitchScore)
		}
	}
}

func TestScore_OffKeyLowersScore(t *testing.T) {
	song := buildTestSong()
	notes := ExtractReferenceNotes(song)
	// canta um tom inteiro (200 cents) acima do alvo em toda nota — bem além do limiar even do easy.
	pcm := buildRecording(notes, func(i int) float64 { return notes[i].FreqHz * math.Pow(2, 200.0/1200.0) })

	result := Score(song, pcm, testSampleRate, 0, PresetHard)

	if result.PitchScore > 20 {
		t.Errorf("PitchScore = %.1f, esperado baixo para desvio de 200 cents no nível hard", result.PitchScore)
	}
}

func TestScore_SilenceGivesZeroCoverage(t *testing.T) {
	song := buildTestSong()
	notes := ExtractReferenceNotes(song)
	pcm := buildRecording(notes, func(i int) float64 { return -1 }) // silêncio nas duas notas

	result := Score(song, pcm, testSampleRate, 0, PresetEasy)

	if result.CoverageScore != 0 {
		t.Errorf("CoverageScore = %.1f, esperado 0 para gravação silenciosa", result.CoverageScore)
	}
	if result.FinalScore != 0 {
		t.Errorf("FinalScore = %.1f, esperado 0 quando nenhuma nota foi coberta", result.FinalScore)
	}
}

func TestScore_OctaveErrorModes(t *testing.T) {
	song := buildTestSong()
	notes := ExtractReferenceNotes(song)
	// canta uma oitava acima em toda nota.
	pcm := buildRecording(notes, func(i int) float64 { return notes[i].FreqHz * 2 })

	easy := Score(song, pcm, testSampleRate, 0, PresetEasy)
	if easy.PitchScore < 90 {
		t.Errorf("easy (oitava ignorada): PitchScore = %.1f, esperado alto", easy.PitchScore)
	}

	hard := Score(song, pcm, testSampleRate, 0, PresetHard)
	if hard.PitchScore > 10 {
		t.Errorf("hard (oitava não corrigida): PitchScore = %.1f, esperado baixo", hard.PitchScore)
	}

	medium := Score(song, pcm, testSampleRate, 0, PresetMedium)
	if medium.PitchScore <= hard.PitchScore || medium.PitchScore >= easy.PitchScore {
		t.Errorf("medium (penalidade parcial): PitchScore = %.1f, esperado entre hard (%.1f) e easy (%.1f)", medium.PitchScore, hard.PitchScore, easy.PitchScore)
	}
}

func TestScore_PlaybackOffsetAlignsTimeline(t *testing.T) {
	song := buildTestSong()
	notes := ExtractReferenceNotes(song)
	const offsetMS = 1200

	// simula gravação que começou 1200ms antes da reprodução: silêncio de offsetMS, depois a
	// música cantada afinada.
	var samples []float64
	samples = append(samples, genSilence(offsetMS, testSampleRate)...)
	rec := buildRecording(notes, func(i int) float64 { return notes[i].FreqHz })
	samples = append(samples, DecodePCM16Mono(rec)...)
	pcm := encodePCM16Mono(samples)

	result := Score(song, pcm, testSampleRate, offsetMS, PresetMedium)
	if result.CoverageScore != 100 {
		t.Errorf("CoverageScore = %.1f, esperado 100 quando o offset é aplicado corretamente", result.CoverageScore)
	}
	if result.FinalScore < 85 {
		t.Errorf("FinalScore = %.1f, esperado alto quando a timeline é alinhada pelo playback_offset_ms", result.FinalScore)
	}
}

func TestPresetByName(t *testing.T) {
	for _, name := range []string{"easy", "medium", "hard"} {
		if _, err := PresetByName(name); err != nil {
			t.Errorf("PresetByName(%q) retornou erro inesperado: %v", name, err)
		}
	}
	if _, err := PresetByName("impossivel"); err == nil {
		t.Error("PresetByName com nome inválido deveria retornar erro")
	}
}
