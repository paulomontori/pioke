package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"opentune/pkg/model"

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

	for i := range s.Timeline {
		event := &s.Timeline[i]
		if event.TimeMS > 0 {
			event.Duration = time.Duration(event.TimeMS) * time.Millisecond
		} else if event.Timestamp != "" {
			d, err := parseTimestamp(event.Timestamp)
			if err != nil {
				return fmt.Errorf("timestamp inválido na linha do tempo [%s]: %w", event.Timestamp, err)
			}
			event.Duration = d
		}
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
