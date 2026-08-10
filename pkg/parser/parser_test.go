package parser

import (
	"os"
	"testing"
	"time"

	"pioke/pkg/model"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"01:23.45", 83450 * time.Millisecond, false},
		{"12.34", 12340 * time.Millisecond, false},
		{"00:05", 5 * time.Second, false},
		{"invalid", 0, true},
		{"12:34:56", 0, true},
	}

	for _, tt := range tests {
		got, err := parseTimestamp(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseTimestamp(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("parseTimestamp(%q) = %v, esperado %v", tt.input, got, tt.expected)
		}
	}
}

func TestParseSongJSON(t *testing.T) {
	jsonContent := `{
		"metadata": {
			"title": "Test Song",
			"artist": "Test Artist",
			"bpm": 120
		},
		"timeline": [
			{
				"timestamp": "00:01.00",
				"lyric": "Hello"
			},
			{
				"time_ms": 2000,
				"lyric": "World"
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "song_test_*.json")
	if err != nil {
		t.Fatalf("Erro ao criar arquivo temporário: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("Erro ao escrever no arquivo temporário: %v", err)
	}
	tmpFile.Close()

	song, err := ParseSong(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseSong falhou: %v", err)
	}

	if song.Title != "Test Song" {
		t.Errorf("Título esperado 'Test Song', obtido %q", song.Title)
	}
	if song.Artist != "Test Artist" {
		t.Errorf("Artista esperado 'Test Artist', obtido %q", song.Artist)
	}
	if len(song.Timeline) != 2 {
		t.Fatalf("Esperado 2 eventos na timeline, obtido %d", len(song.Timeline))
	}
	
	// Evento 0: Timestamp 1s. Próximo evento é 2s. Duração inferida = 1000ms
	if song.Timeline[0].DurationMS != 1000 {
		t.Errorf("Duração do evento 0 esperada 1000ms, obtida %d", song.Timeline[0].DurationMS)
	}
	if song.Timeline[0].Duration != 1000*time.Millisecond {
		t.Errorf("Duration (start time) do evento 0 esperada 1000ms, obtida %v", song.Timeline[0].Duration)
	}
	
	// Evento 1: Último evento, usa default (BPM 120, 4/4 -> 2000ms)
	if song.Timeline[1].DurationMS != 2000 {
		t.Errorf("Duração do evento 1 esperada 2000ms, obtida %d", song.Timeline[1].DurationMS)
	}
	if song.Timeline[1].Duration != 2000*time.Millisecond {
		t.Errorf("Duration (start time) do evento 1 esperada 2000ms, obtida %v", song.Timeline[1].Duration)
	}
}

func TestValidateAndProcessSong_AutoCalculateTimes(t *testing.T) {
	// Cenário 1: Música sem time_ms, duration_ms e sem timestamp
	// BPM = 99, TimeSig = 4/4
	// Duração de 1 batida = 60000 / 99 = 606.06 ms
	// Duração de 1 compasso (4 batidas) = 2424 ms
	song := &model.Song{
		Metadata: model.Metadata{
			Title:   "Evidências",
			BPM:     99,
			TimeSig: "4/4",
		},
		Timeline: []model.TimelineEvent{
			{ChordStr: "E"},
			{ChordStr: "E5+"},
			{ChordStr: "A"},
		},
	}

	err := validateAndProcessSong(song)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	expectedDurationMS := int64(2424)

	// Evento 0
	if song.Timeline[0].TimeMS != 0 {
		t.Errorf("Evento 0: TimeMS esperado 0, obtido %d", song.Timeline[0].TimeMS)
	}
	if song.Timeline[0].DurationMS != expectedDurationMS {
		t.Errorf("Evento 0: DurationMS esperado %d, obtido %d", expectedDurationMS, song.Timeline[0].DurationMS)
	}
	if song.Timeline[0].Duration != 0 {
		t.Errorf("Evento 0: Duration (start time) esperado 0, obtido %v", song.Timeline[0].Duration)
	}
	if song.Timeline[0].Chord == nil || song.Timeline[0].Chord.Name != "E" {
		t.Errorf("Evento 0: Objeto Chord não foi preenchido corretamente")
	}

	// Evento 1
	if song.Timeline[1].TimeMS != expectedDurationMS {
		t.Errorf("Evento 1: TimeMS esperado %d, obtido %d", expectedDurationMS, song.Timeline[1].TimeMS)
	}
	if song.Timeline[1].Duration != time.Duration(expectedDurationMS)*time.Millisecond {
		t.Errorf("Evento 1: Duration (start time) esperado %v, obtido %v", time.Duration(expectedDurationMS)*time.Millisecond, song.Timeline[1].Duration)
	}

	// Evento 2
	if song.Timeline[2].TimeMS != expectedDurationMS*2 {
		t.Errorf("Evento 2: TimeMS esperado %d, obtido %d", expectedDurationMS*2, song.Timeline[2].TimeMS)
	}
}

func TestValidateAndProcessSong_ExplicitTimes(t *testing.T) {
	// Cenário 2: Música com tempos explícitos (como parabens.yaml)
	song := &model.Song{
		Metadata: model.Metadata{
			Title: "Parabéns",
		},
		Timeline: []model.TimelineEvent{
			{TimeMS: 0, DurationMS: 1636},
			{TimeMS: 1636, DurationMS: 1636},
			{TimeMS: 3272, DurationMS: 2500},
		},
	}

	err := validateAndProcessSong(song)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	// Verifica se os tempos foram mantidos e Duration preenchido com o start time
	if song.Timeline[1].TimeMS != 1636 {
		t.Errorf("Evento 1: TimeMS foi alterado, esperado 1636, obtido %d", song.Timeline[1].TimeMS)
	}
	if song.Timeline[2].DurationMS != 2500 {
		t.Errorf("Evento 2: DurationMS esperado 2500, obtido %d", song.Timeline[2].DurationMS)
	}
	if song.Timeline[2].Duration != 3272*time.Millisecond {
		t.Errorf("Evento 2: Duration (start time) esperado 3272ms, obtido %v", song.Timeline[2].Duration)
	}
}

func TestValidateAndProcessSong_InferDurationFromNext(t *testing.T) {
	// Cenário 3: Música com timestamps, mas sem duration_ms
	song := &model.Song{
		Metadata: model.Metadata{
			Title: "Teste Inferência",
		},
		Timeline: []model.TimelineEvent{
			{Timestamp: "00:00.000"},
			{Timestamp: "00:00.500"},
			{Timestamp: "00:01.500"},
		},
	}

	err := validateAndProcessSong(song)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	if song.Timeline[0].DurationMS != 500 {
		t.Errorf("Evento 0: DurationMS esperado 500, obtido %d", song.Timeline[0].DurationMS)
	}
	if song.Timeline[1].DurationMS != 1000 {
		t.Errorf("Evento 1: DurationMS esperado 1000, obtido %d", song.Timeline[1].DurationMS)
	}
	// O último evento usa o default (2000ms pois não tem BPM definido)
	if song.Timeline[2].DurationMS != 2000 {
		t.Errorf("Evento 2: DurationMS esperado 2000, obtido %d", song.Timeline[2].DurationMS)
	}
}
