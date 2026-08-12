# PioKe 🎤🎶

> **Interactive Real-Time Karaoke Engine & Chord Synthesizer built in Go.**

PioKe é uma engine de karaokê e sintetizador de acordes em tempo real, leve e multiplataforma, desenvolvida em Go. O projeto foi projetado para rodar com baixíssimo consumo de CPU e memória em computadores de baixo custo, Raspberry Pi, Smart TVs e sistemas embarcados.

---

## 🚀 Funcionalidades Implementadas

* **Parsing de Arquivos de Música:** Suporte para carregamento e validação de arquivos de música em formato **JSON** e **YAML**.
* **Motor de Sincronização de Alta Precisão:** Loop de reprodução baseado em `time.Ticker` com precisão de milissegundos.
* **Sintetizador de Áudio Polifônico com Timbre Selecionável:** cálculo de frequências musicais baseado em afinação A4 (440Hz), com envelope de attack/release para evitar picotado entre notas adjacentes. Dois motores de síntese, escolhíveis em tempo de execução (flag `-timbre`): **aditivo** (fundamental + harmônicos com peso decrescente — padrão) e **Karplus-Strong** (corda dedilhada, cada nota dedilha sua própria corda do zero).
* **Pipeline de Áudio Multiplataforma:** Reprodução de amostras PCM em tempo real via **Oto v3** (`github.com/ebitengine/oto/v3`).
* **Camada de UI Desacoplada:** Interface `Renderer` no pacote `pkg/ui` pronta para suportar interfaces de terminal (TUI) e interfaces gráficas (GUI).
* **Arquitetura Modular:** Separação limpa entre os pacotes `pkg/model`, `pkg/parser`, `pkg/engine`, `pkg/synth`, `pkg/audio` e `pkg/ui`.

---

## 🛠️ Arquitetura do Sistema

```
                  ┌──────────────────────────────┐
                  │   Song File (.json / .yaml)   │
                  └──────────────┬───────────────┘
                                 │
                                 ▼
                  ┌──────────────────────────────┐
                  │          pkg/parser          │
                  └──────────────┬───────────────┘
                                 │ (pkg/model.Song)
                                 ▼
                  ┌──────────────────────────────┐
                  │          pkg/engine          │◄── Clock / Timer (10ms)
                  └──────┬───────────────┬───────┘
                         │               │
        PlaybackEvents   │               │ PlaybackEvents
                         ▼               ▼
      ┌────────────────────┐          ┌────────────────────┐
      │     pkg/audio      │          │       pkg/ui       │
      │   (Oto v3 Synth)   │          │ (Terminal Renderer)│
      └────────────────────┘          └────────────────────┘
```

---

## 📁 Estrutura do Projeto

```
.
├── cmd/
│   └── pioke-cli/         # Ponto de entrada oficial da aplicação CLI
├── examples/
│   └── sample.json        # Arquivo de exemplo de música (Evidências)
├── pkg/
│   ├── audio/             # Gerenciamento do dispositivo de áudio (Oto v3)
│   ├── engine/            # Motor de reprodução e emissão de eventos
│   ├── model/             # Estruturas de dados (Song, TimelineEvent, Metadata)
│   ├── parser/            # Leitor e validador de arquivos JSON/YAML
│   ├── synth/             # Sintetizador ADSR e cálculo de frequências
│   └── ui/                # Abstração da UI e renderizador de terminal
├── doc.md                 # Documentação e especificação do projeto
├── tasks.md               # Roadmap de tarefas e implementação
├── go.mod                 # Dependências do Go
└── README.md              # Documentação principal do repositório
```

---

## 📄 Formato de Arquivo de Música (`.json`)

```json
{
  "title": "Evidências",
  "artist": "Chitãozinho & Xororó",
  "bpm": 100,
  "key": "E",
  "time_signature": "4/4",
  "timeline": [
    {
      "time_ms": 12500,
      "duration_ms": 1200,
      "chord": "E",
      "lyric": "Quando eu digo "
    },
    {
      "time_ms": 13700,
      "duration_ms": 2000,
      "chord": "G#m",
      "lyric": "que não quero mais você..."
    }
  ]
}
```

---

## 💻 Como Executar

### Pré-requisitos
* **Go 1.21** ou superior instalado.
* Dispositivo de saída de áudio configurado.

### Executando a Aplicação
Para rodar o projeto com a música de exemplo padrão:

```bash
go run main.go
```

Ou passar um arquivo de música customizado:

```bash
go run main.go caminho/para/sua_musica.json
```

Você também pode compilar e rodar a CLI diretamente:

```bash
go run cmd/pioke-cli/main.go
```

### Flags

* `-timbre additive|karplus` — escolhe o motor de síntese: `additive` (padrão, harmônicos aditivos) ou `karplus` (Karplus-Strong, corda dedilhada). Ex: `go run main.go -timbre karplus musica.mxl`.
* `-out caminho.wav` — além de tocar ao vivo, grava o áudio sintetizado (com o timbre escolhido) em um arquivo WAV.

As duas flags podem vir em qualquer posição em relação ao caminho da música.

---

## 📜 Licença

Este projeto está sob a licença **GNU General Public License v3.0 (GPLv3)**. Veja o arquivo `LICENSE` para mais detalhes.
