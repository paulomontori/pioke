package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSON(t *testing.T) {
	// Localiza o arquivo evidencias.json a partir da raiz do projeto ou diretório de testes
	path := filepath.Join("..", "..", "songs", "evidencias.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Tenta caminho relativo do diretório atual se executado na raiz
		path = filepath.Join("songs", "evidencias.json")
	}

	song, err := LoadJSON(path)
	if err != nil {
		t.Fatalf("Erro ao carregar LoadJSON: %v", err)
	}

	if song.Title != "Evidências" {
		t.Errorf("Título esperado 'Evidências', obtido '%s'", song.Title)
	}

	if song.Artist != "Chitãozinho & Xororó" {
		t.Errorf("Artista esperado 'Chitãozinho & Xororó', obtido '%s'", song.Artist)
	}

	if len(song.Timeline) < 4 {
		t.Errorf("Esperado pelo menos 4 eventos na timeline, obtido %d", len(song.Timeline))
	}

	firstEvent := song.Timeline[0]
	if firstEvent.ChordStr != "E" {
		t.Errorf("Acorde do primeiro evento esperado 'E', obtido '%s'", firstEvent.ChordStr)
	}

	if firstEvent.Lyric == "" {
		t.Error("Letra do primeiro evento não deveria ser vazia")
	}
}

func TestLoadJSONInvalidPath(t *testing.T) {
	_, err := LoadJSON("non_existent_file.json")
	if err == nil {
		t.Error("Esperava erro ao tentar carregar arquivo inexistente, obtido nil")
	}
}
