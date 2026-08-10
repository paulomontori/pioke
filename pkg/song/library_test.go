package song

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryScan(t *testing.T) {
	tempDir := t.TempDir()

	// Arquivo 1 Válido (JSON com wrapper metadata)
	song1 := []byte(`{
		"metadata": {
			"title": "Evidências",
			"artist": "Chitãozinho & Xororó",
			"bpm": 90,
			"key": "E"
		}
	}`)
	_ = os.WriteFile(filepath.Join(tempDir, "evidencias.json"), song1, 0644)

	// Arquivo 2 Válido (JSON plano)
	song2 := []byte(`{
		"title": "Ainda Ontem Chorei de Saudade",
		"artist": "João Mineiro & Marciano",
		"bpm": 110,
		"key": "A"
	}`)
	_ = os.WriteFile(filepath.Join(tempDir, "ainda_ontem.json"), song2, 0644)

	// Arquivo 3 Válido (YAML plano)
	song3 := []byte(`
title: "Boate Azul"
artist: "Joaquim & Manuel"
bpm: 105
key: "Dm"
`)
	_ = os.WriteFile(filepath.Join(tempDir, "boate_azul.yaml"), song3, 0644)

	// Arquivo Inválido (Deve ser ignorado pela extensão)
	_ = os.WriteFile(filepath.Join(tempDir, "invalido.txt"), []byte("conteudo qualquer"), 0644)

	// Arquivo JSON Inválido sem título (Deve ser ignorado)
	_ = os.WriteFile(filepath.Join(tempDir, "sem_titulo.json"), []byte(`{"foo": "bar"}`), 0644)

	lib := NewLibrary(tempDir)
	items, err := lib.Scan()
	if err != nil {
		t.Fatalf("Erro inesperado no Scan: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Esperado 3 faixas na biblioteca, encontrado %d", len(items))
	}

	// Verifica ordenação alfabética por Título ("Ainda Ontem...", "Boate Azul", "Evidências")
	if items[0].Title != "Ainda Ontem Chorei de Saudade" {
		t.Errorf("Esperado primeira música 'Ainda Ontem Chorei de Saudade', obtido '%s'", items[0].Title)
	}

	if items[1].Title != "Boate Azul" {
		t.Errorf("Esperado segunda música 'Boate Azul', obtido '%s'", items[1].Title)
	}

	if items[2].Title != "Evidências" {
		t.Errorf("Esperado terceira música 'Evidências', obtido '%s'", items[2].Title)
	}
}

func TestLibraryGetSong(t *testing.T) {
	tempDir := t.TempDir()

	// Teste JSON
	jsonPath := filepath.Join(tempDir, "song.json")
	jsonContent := []byte(`{
		"title": "Test JSON Song",
		"artist": "Test Artist",
		"timeline": []
	}`)
	_ = os.WriteFile(jsonPath, jsonContent, 0644)

	// Teste YAML
	yamlPath := filepath.Join(tempDir, "song.yaml")
	yamlContent := []byte(`
title: "Test YAML Song"
artist: "Test Artist 2"
timeline: []
`)
	_ = os.WriteFile(yamlPath, yamlContent, 0644)

	lib := NewLibrary(tempDir)

	songJSON, err := lib.GetSong(jsonPath)
	if err != nil {
		t.Fatalf("Erro ao buscar música JSON com GetSong: %v", err)
	}
	if songJSON.Title != "Test JSON Song" {
		t.Errorf("Título esperado 'Test JSON Song', obtido '%s'", songJSON.Title)
	}

	songYAML, err := lib.GetSong(yamlPath)
	if err != nil {
		t.Fatalf("Erro ao buscar música YAML com GetSong: %v", err)
	}
	if songYAML.Title != "Test YAML Song" {
		t.Errorf("Título esperado 'Test YAML Song', obtido '%s'", songYAML.Title)
	}
}
