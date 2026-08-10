package song

import (
	"encoding/json"
	"fmt"
	"os"

	"pioke/pkg/model"
)

// LoadMetadata lê apenas o cabeçalho e informações básicas de um arquivo de música
func LoadMetadata(filePath string) (*SongMetadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	var meta SongMetadata
	if err := json.Unmarshal(data, &meta); err == nil && meta.Title != "" {
		meta.FilePath = filePath
		return &meta, nil
	}

	// Estrutura alternativa caso o JSON envolva os dados em "metadata"
	var wrapper struct {
		Metadata SongMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Metadata.Title != "" {
		wrapper.Metadata.FilePath = filePath
		return &wrapper.Metadata, nil
	}

	return nil, fmt.Errorf("formato de música ou metadados inválidos em: %s", filePath)
}

// LoadSong realiza a leitura e conversão do arquivo completo para model.Song
func LoadSong(filePath string) (*model.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de música: %w", err)
	}

	var song model.Song
	if err := json.Unmarshal(data, &song); err != nil {
		return nil, fmt.Errorf("falha ao decodificar arquivo de música: %w", err)
	}

	return &song, nil
}
