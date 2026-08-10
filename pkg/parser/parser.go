package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pioke/pkg/model"

	"gopkg.in/yaml.v3"
)

// ParseSong lê um arquivo (.json ou .yaml), valida sua estrutura, converte marcas de tempo e carrega para a memória.
func ParseSong(filePath string) (*model.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de música: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var s model.Song

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("erro ao decodificar YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("formato de arquivo não suportado: %s", ext)
	}

	if err := validateAndProcessSong(&s); err != nil {
		return nil, fmt.Errorf("falha na validação e processamento da música: %w", err)
	}

	return &s, nil
}

// validateAndProcessSong valida campos obrigatórios e converte timestamps para time.Duration
func validateAndProcessSong(s *model.Song) error {
	if s.Title == "" && s.Metadata.Title != "" {
		s.Title = s.Metadata.Title
	}
	if s.Artist == "" && s.Metadata.Artist != "" {
		s.Artist = s.Metadata.Artist
	}

	if s.Title == "" && s.Metadata.Title == "" {
		return fmt.Errorf("metadado obrigatório 'title' está ausente")
	}

	// Obtém BPM
	bpm := s.BPM
	if bpm == 0 {
		bpm = s.Metadata.BPM
	}

	// Obtém Time Signature (Fórmula de Compasso)
	timeSig := s.TimeSig
	if timeSig == "" {
		timeSig = s.Metadata.TimeSig
	}

	// Define batidas por compasso padrão (4/4)
	beatsPerMeasure := 4
	if timeSig != "" {
		parts := strings.Split(timeSig, "/")
		if len(parts) >= 1 {
			if b, err := strconv.Atoi(parts[0]); err == nil && b > 0 {
				beatsPerMeasure = b
			}
		}
	}

	// Calcula a duração padrão de 1 compasso baseado no BPM e Time Signature
	var defaultDurationMS int64 = 2000
	if bpm > 0 {
		beatDurationMS := 60000.0 / float64(bpm)
		defaultDurationMS = int64(math.Round(beatDurationMS * float64(beatsPerMeasure)))
	}

	var accumulatedTimeMS int64 = 0

	for i := range s.Timeline {
		event := &s.Timeline[i]
		
		// 1. Se TimeMS está zerado mas temos um Timestamp em string, usamos ele
		if event.TimeMS == 0 && event.Timestamp != "" && event.Timestamp != "00:00.00" {
			d, err := parseTimestamp(event.Timestamp)
			if err != nil {
				return fmt.Errorf("timestamp inválido na linha do tempo [%s]: %w", event.Timestamp, err)
			}
			event.TimeMS = d.Milliseconds()
		}

		// 2. Se não temos TimeMS nem Timestamp, calculamos sequencialmente
		// O evento 0 começa em 0. Os demais começam no accumulatedTimeMS.
		if event.TimeMS == 0 && event.Timestamp == "" {
			if i > 0 {
				event.TimeMS = accumulatedTimeMS
			} else {
				event.TimeMS = 0
			}
		}

		// 3. Se a duração não foi definida, usamos a duração calculada do compasso
		if event.DurationMS <= 0 {
			event.DurationMS = defaultDurationMS
		}

		// 4. Preenche Duration (time.Duration) com o TimeMS (usado como start time no sistema)
		event.Duration = time.Duration(event.TimeMS) * time.Millisecond

		// 5. Preenche o Timestamp em string caso esteja vazio (útil para UI)
		if event.Timestamp == "" {
			totalSeconds := event.TimeMS / 1000
			mins := totalSeconds / 60
			secs := float64(event.TimeMS%60000) / 1000.0
			event.Timestamp = fmt.Sprintf("%02d:%05.2f", mins, secs)
		}

		// 6. Atualiza o tempo acumulado para o próximo evento
		accumulatedTimeMS = event.TimeMS + event.DurationMS
	}

	return nil
}

// parseTimestamp converte formatos de tempo como "01:23.45" ou "12.34" em time.Duration
func parseTimestamp(ts string) (time.Duration, error) {
	parts := strings.Split(ts, ":")
	var minutes float64
	var seconds float64
	var err error

	if len(parts) == 2 {
		minutes, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, fmt.Errorf("minutos inválidos: %w", err)
		}
		seconds, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("segundos inválidos: %w", err)
		}
	} else if len(parts) == 1 {
		seconds, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, fmt.Errorf("segundos inválidos: %w", err)
		}
	} else {
		return 0, fmt.Errorf("formato de timestamp desconhecido")
	}

	totalSeconds := (minutes * 60) + seconds
	return time.Duration(totalSeconds * float64(time.Second)), nil
}
