# Prompt de Implementação — Fase 7: Navegação e Seletor de Músicas (`pkg/song/library`)

**Projeto:** Pioke (`github.com/paulomontori/pioke`)  
**Linguagem:** Go (Golang)  
**Objetivo da Task:** Criar o gerenciador de biblioteca de músicas para escanear diretórios locais (ex: `./songs`), listar, filtrar e carregar dinamicamente arquivos de música em formato `.json` ou `.yaml`, permitindo a navegação e troca de faixas sem reiniciar a aplicação.

---

## 📐 Visão Geral & Objetivos

A Fase 7 adiciona a funcionalidade de menu/seletor de faixas ao Pioke. O módulo `pkg/song/library` deve:
1. Ler o diretório configurado (`./songs`) e fazer o parse dos metadados de todas as músicas encontradas sem carregar o áudio/eventos completos na memória de uma só vez (lazy loading ou leitura rápida de cabeçalho).
2. Fornecer uma estrutura para listar músicas com filtros simples (título, artista, tom/key, bpm).
3. Permitir a alternância dinâmica de faixas: interromper o playback atual, resetar o motor de áudio/síntese e carregar a nova música selecionada.

---

## 📁 Estrutura de Arquivos a Criar / Atualizar

```text
pkg/
└── song/
    ├── library.go        # Gerenciador de diretório, varredura e lista de faixas
    ├── loader.go         # Parsers para arquivos .json e .yaml
    └── library_test.go    # Testes unitários para varredura e ordenação de biblioteca
```

---

## 🛠️ Especificação Detalhada

### 1. `pkg/song/library.go`
- **Estruturas de Dados:**
  ```go
  type SongMetadata struct {
      FilePath string `json:"file_path" yaml:"file_path"`
      Title    string `json:"title" yaml:"title"`
      Artist   string `json:"artist" yaml:"artist"`
      BPM      int    `json:"bpm" yaml:"bpm"`
      Key      string `json:"key" yaml:"key"`
  }

  type Library struct {
      songsDir string
      items    []SongMetadata
  }
  ```
- **Métodos Principais:**
  - `NewLibrary(dir string) *Library`: Instancia a biblioteca vinculada ao diretório especificado.
  - `Scan() ([]SongMetadata, error)`: Varre o diretório `./songs` utilizando `os.ReadDir` ou `filepath.WalkDir`, identificando arquivos `.json` e `.yaml` / `.yml`.
  - `GetSong(filePath string) (*song.Song, error)`: Carrega o objeto `Song` completo a ser entregue ao `engine`.

### 2. `pkg/song/loader.go`
- Suporte a múltiplos formatos de arquivo:
  - JSON (`encoding/json`)
  - YAML (`gopkg.in/yaml.v3`)
- Validação rápida de esquema/estrutura ao escanear para ignorar arquivos inválidos ou que não sejam faixas do Pioke.

---

## 🧪 Testes e Validação

1. **`pkg/song/library_test.go`**:
   - Criar diretórios temporários (`t.TempDir()`) com arquivos `.json` e `.yaml` válidos e inválidos.
   - Validar se o método `Scan()` encontra todas as faixas e ignora extensões/formatos desconhecidos.
   - Garantir ordenação alfabética por Título ou Artista na listagem.

---

## 🚀 Requisitos de Entrega (Checklist para o Agente)

- [x] Pacote `pkg/song/library` e `pkg/song/loader` implementados.
- [x] Suporte nativo a parsing de arquivos `.json` e `.yaml`.
- [x] Testes unitários cobrindo varredura de arquivos e tratamento de erros para arquivos corrompidos.
- [x] Integração com as interfaces de UI (TUI/GUI) para exibição do menu de seleção.
