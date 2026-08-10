package song

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryScan(t *testing.T) {
	tempDir := t.TempDir()

	// Arquivo 1 Válido
	song1 := []byte(`{
		"metadata": {
			"title": "Evidências",
			"artist": "Chitãozinho & Xororó",
			"bpm": 90,
			"key": "E"
		}
	}`)
	_ = os.WriteFile(filepath.Join(tempDir, "evidencias.json"), song1, 0644)

	// Arquivo 2 Válido
	song2 := []byte(`{
		"title": "Ainda Ontem Chorei de Saudade",
		"artist": "João Mineiro & Marciano",
		"bpm": 110,
		"key": "A"
	}`)
	_ = os.WriteFile(filepath.Join(tempDir, "ainda_ontem.json"), song2, 0644)

	// Arquivo Inválido (Deve ser ignorado)
	_ = os.WriteFile(filepath.Join(tempDir, "invalido.txt"), []byte("conteudo qualquer"), 0644)

	lib := NewLibrary(tempDir)
	items, err := lib.Scan()
	if err != nil {
		t.Fatalf("Erro inesperado no Scan: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("Esperado 2 faixas na biblioteca, encontrado %d", len(items))
	}

	// Verifica ordenação alfabética por Título ("Ainda Ontem..." vem antes de "Evidências")
	if items[0].Title != "Ainda Ontem Chorei de Saudade" {
		t.Errorf("Esperado primeira música 'Ainda Ontem Chorei de Saudade', obtido '%s'", items[0].Title)
	}

	if items[1].Title != "Evidências" {
		t.Errorf("Esperado segunda música 'Evidências', obtido '%s'", items[1].Title)
	}
}

func TestLibraryGetSong(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "song.json")

	songContent := []byte(`{
		"title": "Test Song",
		"artist": "Test Artist",
		"timeline": []
	}`)
	_ = os.WriteFile(filePath, songContent, 0644)

	lib := NewLibrary(tempDir)
	song, err := lib.GetSong(filePath)
	if err != nil {
		t.Fatalf("Erro ao buscar música com GetSong: %v", err)
	}

	if song.Title != "Test Song" {
		t.Errorf("Título esperado 'Test Song', obtido '%s'", song.Title)
	}
}
