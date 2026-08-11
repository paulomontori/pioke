package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"pioke/pkg/model"
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

	// Lógica de fallback: se não houver sílabas, mas houver letra (lyric), cria uma sílaba padrão
	for i := range song.Timeline {
		if len(song.Timeline[i].Syllables) == 0 && song.Timeline[i].Lyric != "" {
			song.Timeline[i].Syllables = []model.Syllable{
				{
					Text:       song.Timeline[i].Lyric,
					OffsetMS:   0,
					DurationMS: song.Timeline[i].DurationMS,
				},
			}
		}
	}

	if err := validateAndProcessSong(&song); err != nil {
		return nil, fmt.Errorf("falha na validação da música: %w", err)
	}

	return &song, nil
}
