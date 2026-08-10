# Prompt de Implementação — Fase 6: Interface Gráfica 2D (`pkg/ui/gui` - Ebitengine)

**Projeto:** Pioke (`github.com/paulomontori/pioke`)  
**Linguagem:** Go (Golang)  
**Objetivo da Task:** Implementar a interface gráfica 2D (GUI) usando a biblioteca Ebitengine (`github.com/hajimehoshi/ebiten/v2`), oferecendo uma visualização fluida e responsiva para telas desktop e Raspberry Pi OS (HDMI), com suporte a modo tela cheia, renderização de fontes customizadas e transição suave no destaque de letras e acordes.

---

## 📐 Visão Geral & Objetivos

A Fase 6 cria a experiência visual rica do Pioke. O `pkg/ui/gui` deve implementar a interface `Renderer` definida na Fase 5 e gerenciar a janela do Ebitengine:
1. Renderizar as linhas de texto (com suporte a UTF-8/Acentuação) e posições de acordes.
2. Animar/destacar a palavra ou sílaba atual de forma fluida (interpolação visual/smooth transition).
3. Suportar alternância para Fullscreen, ajuste automático de resolução e escala para telas de TV/monitores via Raspberry Pi.
4. Manter baixíssimo consumo de CPU/GPU para garantir 60 FPS estáveis no Raspberry Pi 3/4/5.

---

## 📁 Estrutura de Arquivos a Criar / Atualizar

```text
pkg/
└── ui/
    └── gui/
        ├── ebiten.go         # Implementação de ebiten.Game e do Renderer
        ├── font.go           # Carregamento e cache de fontes TrueType/OpenType (golang.org/x/image/font)
        ├── view.go           # Lógica de layout (cálculo de bounding box, centralização, rolagem)
        └── gui_test.go       # Testes de inicialização e mapeamento de estado
```

---

## 🛠️ Especificação Detalhada

### 1. `pkg/ui/gui/font.go`
- Carregar fonte TrueType incorporada via `embed.FS` ou do sistema.
- Criar rotinas para medir o tamanho do texto (largura/altura em pixels) para renderização centralizada e alinhada dos acordes sobre as palavras.

### 2. `pkg/ui/gui/ebiten.go`
- Implementar as interfaces `ebiten.Game` e `ui.Renderer`:
  ```go
  type GUIRenderer struct {
      // Estado de renderização
      currentSong *song.Song
      activeEvent engine.PlaybackEvent
      // ...
  }

  func (g *GUIRenderer) Update() error {
      // Processar inputs (ex: F11 para Fullscreen, Esc para sair)
      return nil
  }

  func (g *GUIRenderer) Draw(screen *ebiten.Image) {
      // Desenhar fundo, cabeçalho da música, versos da letra e acordes
  }

  func (g *GUIRenderer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
      return 1280, 720 // Resolução base de referência
  }
  ```

### 3. `pkg/ui/gui/view.go`
- **Renderização dos Acordes:** Posicionar as cifras em uma linha dedicada imediatamente acima de cada palavra/sílaba correspondente.
- **Destaque do Karaokê:**
  - Cor normal de texto inativo (ex: cinza claro/branco).
  - Cor do texto ativo (ex: amarelo/dourado brilhante).
  - Animação simples de transição/preenchimento (sweep effect ou fade).
- **Auto-scroll:** Ajustar a posição vertical do texto conforme a música avança, mantendo a linha ativa centralizada na tela.

---

## 🧪 Testes e Validação Manual

- Testar inicialização de janela com resolução nativa e redimensionamento.
- Validar se o renderizador respeita a interface `ui.Renderer` sem vazar lógica específica do Ebitengine fora de `pkg/ui/gui`.
- Verificar estabilidade da taxa de quadros (60 FPS) e ausência de travamentos no loop principal (`ebiten.RunGame`).

---

## 🚀 Requisitos de Entrega (Checklist para o Agente)

- [ ] Código implementado sob `pkg/ui/gui`.
- [ ] Fontes renderizadas sem artefatos e com suporte a caracteres da língua portuguesa.
- [ ] Suporte a alternância de tela cheia (Fullscreen) via atalho (ex: `F11` ou `Alt+Enter`).
- [ ] Interface desacoplada usando a abstração `ui.Renderer`.
