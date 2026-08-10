package parser

import (
	"os"
	"testing"
	"time"
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
	if song.Timeline[0].Duration != time.Second {
		t.Errorf("Duração do evento 0 esperada 1s, obtida %v", song.Timeline[0].Duration)
	}
	if song.Timeline[1].Duration != 2*time.Second {
		t.Errorf("Duração do evento 1 esperada 2s, obtida %v", song.Timeline[1].Duration)
	}
}
