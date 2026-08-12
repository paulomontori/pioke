package song

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pioke/pkg/model"
)

// SongMetadata armazena os cabeçalhos e informações principais de uma música
type SongMetadata struct {
	FilePath string `json:"file_path" yaml:"file_path"`
	Title    string `json:"title" yaml:"title"`
	Artist   string `json:"artist" yaml:"artist"`
	BPM      int    `json:"bpm" yaml:"bpm"`
	Key      string `json:"key" yaml:"key"`
}

// Library gerencia o diretório de músicas e sua listagem
type Library struct {
	songsDir string
	items    []SongMetadata
}

// NewLibrary instancia a biblioteca vinculada ao diretório especificado
func NewLibrary(dir string) *Library {
	return &Library{
		songsDir: dir,
		items:    make([]SongMetadata, 0),
	}
}

// Scan varre o diretório e mapeia os arquivos .json, .yaml e .yml
func (l *Library) Scan() ([]SongMetadata, error) {
	l.items = make([]SongMetadata, 0)

	entries, err := os.ReadDir(l.songsDir)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler diretório de músicas: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" && ext != ".musicxml" && ext != ".xml" && ext != ".mxl" {
			continue
		}

		fullPath := filepath.Join(l.songsDir, entry.Name())
		meta, err := LoadMetadata(fullPath)
		if err != nil {
			// Ignora arquivos corrompidos ou em formatos inválidos
			continue
		}

		l.items = append(l.items, *meta)
	}

	// Ordena as músicas alfabeticamente por Título
	sort.Slice(l.items, func(i, j int) bool {
		return strings.ToLower(l.items[i].Title) < strings.ToLower(l.items[j].Title)
	})

	return l.items, nil
}

// GetSong carrega o objeto Song completo para reprodução no engine
func (l *Library) GetSong(filePath string) (*model.Song, error) {
	return LoadSong(filePath)
}

// Items retorna as músicas encontradas na última varredura
func (l *Library) Items() []SongMetadata {
	return l.items
}
