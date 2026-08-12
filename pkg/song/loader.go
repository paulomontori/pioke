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

// LoadMetadata lê apenas os metadados/cabeçalho de um arquivo .json, .yaml/.yml,
// .musicxml/.xml ou .mxl
func LoadMetadata(filePath string) (*SongMetadata, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if s, ok, err := loadScoreFile(filePath, ext); ok {
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

// loadScoreFile despacha para o parser de partitura (MusicXML texto ou .mxl comprimido)
// quando a extensão corresponde. O segundo valor de retorno indica se a extensão foi
// reconhecida — quando falso, o chamador deve seguir para a lógica de JSON/YAML.
func loadScoreFile(filePath, ext string) (*model.Song, bool, error) {
	switch ext {
	case ".musicxml", ".xml":
		s, err := parser.ParseMusicXML(filePath)
		return s, true, err
	case ".mxl":
		s, err := parser.ParseMXL(filePath)
		return s, true, err
	default:
		return nil, false, nil
	}
}

// LoadSong realiza a leitura e conversão do arquivo completo para model.Song
// (.json, .yaml/.yml, .musicxml/.xml ou .mxl)
func LoadSong(filePath string) (*model.Song, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if s, ok, err := loadScoreFile(filePath, ext); ok {
		return s, err
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
