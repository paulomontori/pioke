package song

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"pioke/pkg/model"
	"pioke/pkg/parser"
)

// LoadMetadata lê apenas os metadados/cabeçalho de um arquivo .json, .yaml/.yml ou .musicxml/.xml
func LoadMetadata(filePath string) (*SongMetadata, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".musicxml" || ext == ".xml" {
		s, err := parser.ParseMusicXML(filePath)
		if err != nil || s.Title == "" {
			return nil, fmt.Errorf("formato de música ou metadados inválidos em: %s", filePath)
		}
		return &SongMetadata{FilePath: filePath, Title: s.Title, Artist: s.Artist, BPM: s.BPM, Key: s.Key}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

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

// LoadSong realiza a leitura e conversão do arquivo completo para model.Song
// (.json, .yaml/.yml ou .musicxml/.xml)
func LoadSong(filePath string) (*model.Song, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".musicxml" || ext == ".xml" {
		return parser.ParseMusicXML(filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de música: %w", err)
	}

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
