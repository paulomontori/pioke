package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"opentune/song"
)

// ParseSong lê um arquivo (.json) e converte para a estrutura song.Song
func ParseSong(filePath string) (*song.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de musica: %w", err)
	}

	ext := filepath.Ext(filePath)
	var s song.Song

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("formato de arquivo nao suportado: %s", ext)
	}

	return &s, nil
}
