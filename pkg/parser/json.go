package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/username/pioke/pkg/model"
)

// LoadJSON lê e valida um arquivo de música em formato JSON.
func LoadJSON(filePath string) (*model.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo json: %w", err)
	}

	var song model.Song
	if err := json.Unmarshal(data, &song); err != nil {
		return nil, fmt.Errorf("erro ao decodificar json: %w", err)
	}

	if err := validateAndProcessSong(&song); err != nil {
		return nil, fmt.Errorf("falha na validação da música: %w", err)
	}

	return &song, nil
}
