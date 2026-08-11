# Guia de Criação de Músicas para o Pioke

Este manual orienta como criar arquivos de música compatíveis com o **Pioke** nos formatos **JSON** ou **YAML**. Os arquivos devem ser salvos na pasta de músicas (ex: `./songs`).

---

## 📄 Formatos Suportados

O Pioke aceita arquivos com extensões `.json`, `.yaml` e `.yml`. Você pode organizar a estrutura de forma **plana** (campos na raiz) ou **encapsulada** (com a chave `metadata`).

---

## 🎼 Estrutura Básica

Um arquivo de música é composto por duas partes principais:
1. **Cabeçalho/Metadados**: Informações sobre título, artista, tom e BPM.
2. **Linha do Tempo (`timeline`)**: Lista de eventos sincronizados por tempo contendo letras, acordes e propriedades avançadas de canto.

---

## 🛠️ Campos de Metadados

| Campo | Tipo | Obrigatório | Descrição |
| :--- | :--- | :--- | :--- |
| `title` | `string` | **Sim** | Nome da música |
| `artist` | `string` | Não | Nome do artista/banda |
| `bpm` | `int` | Não | Batidas por minuto |
| `key` | `string` | Não | Tom principal da música (ex: `"C"`, `"Am"`, `"E"`) |
| `time_signature` | `string` | Não | Fórmula de compasso (ex: `"4/4"`, `"3/4"`) |

---

## ⏱️ Campos da Linha do Tempo (`timeline`)

Cada item da lista `timeline` representa um evento num determinado momento da música:

| Campo | Tipo | Descrição |
| :--- | :--- | :--- |
| `timestamp` | `string` | Marcador em texto (ex: `"00:15.00"`). |
| `time_ms` | `int64` | Tempo em milissegundos a partir do início (ex: `15000`). |
| `duration_ms` | `int64` | Duração do evento/acorde em milissegundos. |
| `lyric` | `string` | Texto da letra simples para exibição rápida. |
| `chord` | `string` | Nome do acorde para execução no sintetizador (ex: `"C"`, `"Am"`, `"G"`). |
| `velocity` | `int` | Intensidade do volume (0-127). |
| `articulation` | `string` | Tipo de articulação (ex: `"legato"`, `"staccato"`). |
| `syllables` | `array` | Lista de objetos de sílabas para sincronização detalhada e notas musicais. |

### Estrutura de Sílabas (`syllables`)

Para um controle mais preciso do karaokê e avaliação de afinação, você pode detalhar as sílabas dentro de um evento da timeline:

| Campo | Tipo | Descrição |
| :--- | :--- | :--- |
| `text` | `string` | O texto da sílaba (ex: `"Pa"`, `"ra"`, `"béns"`). |
| `offset_ms` | `int64` | Tempo de início da sílaba relativo ao `time_ms` do evento pai. |
| `duration_ms` | `int64` | Duração da sílaba em milissegundos. |
| `pitch` | `string` | Nota musical esperada para a melodia (ex: `"G4"`, `"C5"`). |

---

## 💡 Exemplos Completos

### 1. Exemplo em YAML (Plano)

```yaml
title: "Boate Azul"
artist: "Joaquim & Manuel"
bpm: 105
key: "Dm"
timeline:
  - time_ms: 0
    duration_ms: 3000
    chord: "Dm"
    lyric: "Doente de amor, procurei remédio na boate azul"
  - time_ms: 3000
    duration_ms: 3000
    chord: "A7"
    lyric: "Comi uma fruta e tomei um café"
```

### 2. Exemplo em JSON (Avançado com Sílabas)

```json
{
  "metadata": {
    "title": "Evidências",
    "artist": "Chitãozinho & Xororó",
    "bpm": 90,
    "key": "E"
  },
  "timeline": [
    {
      "time_ms": 1000,
      "duration_ms": 2000,
      "chord": "E",
      "lyric": "Quando",
      "velocity": 100,
      "articulation": "legato",
      "syllables": [
        {
          "text": "Quan",
          "offset_ms": 0,
          "duration_ms": 1000,
          "pitch": "E4"
        },
        {
          "text": "do",
          "offset_ms": 1000,
          "duration_ms": 1000,
          "pitch": "F#4"
        }
      ]
    },
    {
      "time_ms": 3000,
      "duration_ms": 2000,
      "chord": "B",
      "lyric": "É porque eu te amo"
    }
  ]
}
```

---

## 🔍 Regras Importantes

1. **Validação do Título**: O campo `title` é obrigatório. Se um arquivo for criado sem título, a biblioteca do Pioke irá ignorá-lo automaticamente ao escanear a pasta.
2. **Nomes de Acordes**: Utilize notação de cifra padrão (ex: `C`, `C#`, `Db`, `Dm`, `E`, `Em`, `G7`).
3. **Ordenação Temporal**: Certifique-se de que os valores de `time_ms` estão em ordem crescente na lista `timeline`.
4. **Fallback de Sílabas**: Se o campo `syllables` não for fornecido, o Pioke criará automaticamente uma sílaba única usando o texto do campo `lyric` e a duração total do evento.
