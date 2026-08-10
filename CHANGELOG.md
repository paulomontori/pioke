# Registro de Commits e Alterações — Pioke

Histórico sequencial das implementações e commits realizados no projeto:

## Commits Realizados

1. **`8557c9d`** - `feat: adiciona parser de JSON para músicas e testes`
   - Implementação de `pkg/parser/json.go` para carregar e validar músicas em formato JSON.
   - Criação da música de exemplo `songs/evidencias.json`.
   - Adição dos testes unitários em `pkg/parser/json_test.go`.

2. **`5ccdb48`** - `feat: atualizar motor de reproducao e modelo de eventos`
   - Criação da estrutura `PlaybackEvent` e enumeração de estados `PlaybackState` em `pkg/model/event.go`.
   - Atualização do motor de reprodução `pkg/engine/engine.go` com métodos `Play`, `Pause`, `Stop`, `Seek` e loop via `time.Ticker`.
   - Atualização dos testes unitários do engine em `pkg/engine/engine_test.go`.

3. **`8b1e485`** - `feat: implementa sintetizador de áudio e motor PCM`
   - Implementação de geradores de onda, envelope ADSR e sintetizador polifônico em `pkg/synth/wave.go`, `pkg/synth/chord.go` e `pkg/synth/synth.go`.
   - Criação da conversão PCM e reprodutor de eventos de áudio em `pkg/audio/engine.go` e `pkg/audio/player.go`.
   - Adição dos testes para sintetizador e motor de áudio em `pkg/synth/synth_test.go` e `pkg/audio/audio_test.go`.

4. **`ff9c8d1`** - `feat: adiciona executável CLI para teste de reprodução`
   - Implementação da aplicação de linha de comando em `cmd/pioke-cli/main.go` para teste de reprodução integrada com encerramento gracioso via sinais do sistema.

5. **`5cbc6d0`** - `fix: corrige sintaxe das flags no CLI e caminho de importação`
   - Correção de sintaxe na inicialização da flag `-no-audio` em `cmd/pioke-cli/main.go`.
   - Ajuste nos pacotes de importação em `pkg/parser/json.go`.

6. **`4a5722c`** - `feat: implementar abstração e estrutura do Terminal UI com Bubble Tea`
   - Abstração da interface visual `Renderer` em `pkg/ui/renderer.go`.
   - Implementação da interface TUI usando Lipgloss e Bubble Tea em `pkg/ui/tui/styles.go`, `pkg/ui/tui/model.go` e `pkg/ui/tui/tui.go`.
   - Adição dos testes unitários em `pkg/ui/tui/tui_test.go`.

7. **`130b550`** - `feat: adiciona interface gráfica 2D usando Ebitengine`
   - Implementação da Fase 6 do projeto em `pkg/ui/gui/font.go`, `pkg/ui/gui/view.go` e `pkg/ui/gui/ebiten.go`.
   - Suporte a renderização 2D, redimensionamento de janela, alternância em tela cheia (F11) e testes em `pkg/ui/gui/gui_test.go`.

8. **`Fase 7`** - `feat: adiciona biblioteca e gerenciador de músicas`
   - Criação do pacote `pkg/song` para varredura de diretórios e carregamento dinâmico de faixas em `pkg/song/library.go` e `pkg/song/loader.go`.
   - Adição dos testes unitários de biblioteca em `pkg/song/library_test.go`.
