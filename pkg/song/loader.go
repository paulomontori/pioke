package song

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"pioke/pkg/model"
)

// LoadMetadata lê apenas os metadados/cabeçalho de um arquivo .json ou .yaml/.yml
func LoadMetadata(filePath string) (*SongMetadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var meta SongMetadata

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &meta); err == nil && meta.Title != "" {
			meta.FilePath = filePath
			return &meta, nil
		}
		var wrapper struct {
			Metadata SongMetadata `json:"metadata"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Metadata.Title != "" {
			wrapper.Metadata.FilePath = filePath
			return &wrapper.Metadata, nil
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &meta); err == nil && meta.Title != "" {
			meta.FilePath = filePath
			return &meta, nil
		}
		var wrapper struct {
			Metadata SongMetadata `yaml:"metadata"`
		}
		if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Metadata.Title != "" {
			wrapper.Metadata.FilePath = filePath
			return &wrapper.Metadata, nil
		}
	default:
		return nil, fmt.Errorf("formato não suportado: %s", ext)
	}

	return nil, fmt.Errorf("formato de música ou metadados inválidos em: %s", filePath)
}

// LoadSong realiza a leitura e conversão do arquivo completo para model.Song (.json ou .yaml/.yml)
func LoadSong(filePath string) (*model.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de música: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var song model.Song

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &song); err != nil {
			return nil, fmt.Errorf("falha ao decodificar arquivo JSON de música: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &song); err != nil {
			return nil, fmt.Errorf("falha ao decodificar arquivo YAML de música: %w", err)
		}
	default:
		return nil, fmt.Errorf("formato não suportado: %s", ext)
	}

	return &song, nil
}
