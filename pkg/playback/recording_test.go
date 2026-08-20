package playback

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// makeFakeRecording cria um par <base>.wav + <base>.json em dir, com a data de modificação
// ajustada explicitamente (em vez de depender da resolução do relógio do sistema entre criações
// sucessivas, que pode não ser fina o bastante para garantir ordem).
func makeFakeRecording(t *testing.T, dir, base string, modTime time.Time) {
	t.Helper()
	for _, ext := range []string{".wav", ".json"} {
		path := filepath.Join(dir, base+ext)
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("erro ao criar %s: %v", path, err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("erro ao ajustar mtime de %s: %v", path, err)
		}
	}
}

func TestEnforceRecordingLimit_KeepsNewestUpToMax(t *testing.T) {
	dir := t.TempDir()

	base := time.Now()
	const total = maxRecordings + 3
	for i := range total {
		makeFakeRecording(t, dir, fmt.Sprintf("musica_%02d", i), base.Add(time.Duration(i)*time.Minute))
	}

	enforceRecordingLimit(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("erro ao ler dir: %v", err)
	}

	var wavCount int
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".wav" {
			wavCount++
		}
	}
	if wavCount != maxRecordings {
		t.Fatalf("esperava %d gravações restantes, obteve %d", maxRecordings, wavCount)
	}

	// as 3 mais antigas (i=0,1,2) devem ter sido removidas, e cada .json deve ter sido removido
	// junto do seu .wav correspondente (nenhum arquivo órfão).
	for i := range 3 {
		name := fmt.Sprintf("musica_%02d", i)
		if _, err := os.Stat(filepath.Join(dir, name+".wav")); !os.IsNotExist(err) {
			t.Errorf("gravação antiga %s.wav deveria ter sido removida", name)
		}
		if _, err := os.Stat(filepath.Join(dir, name+".json")); !os.IsNotExist(err) {
			t.Errorf("metadados órfãos: %s.json deveria ter sido removido junto do .wav", name)
		}
	}
}

func TestEnforceRecordingLimit_NoOpBelowLimit(t *testing.T) {
	dir := t.TempDir()
	base := time.Now()
	for i := range 3 {
		makeFakeRecording(t, dir, fmt.Sprintf("musica_%02d", i), base.Add(time.Duration(i)*time.Minute))
	}

	enforceRecordingLimit(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("erro ao ler dir: %v", err)
	}
	if len(entries) != 6 { // 3 pares .wav+.json
		t.Fatalf("esperava 6 arquivos (nenhum removido abaixo do limite), obteve %d", len(entries))
	}
}
