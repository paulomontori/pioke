# Refatoração Crítica: Polifonia Real e Stream Não-Bloqueante (pkg/synth & pkg/audio)

**Problemas Identificados:**
1. Os acordes não estão soando polifônicos (apenas uma nota estática ou timbre incorreto é ouvido).
2. O áudio sofre engasgos constantes/tremolo devido a bloqueios na implementação do `io.Reader` do stream de áudio.

---

## 🛠️ Especificação da Correção

### 1. `pkg/synth/chord.go` — Definição Precisa dos Acordes
- [x] Cada acorde retorna um slice com as frequências reais em Hz de **todas** as suas notas:
  - `C` (Dó Maior): `[261.63, 329.63, 392.00]` (C4, E4, G4)
  - `G7` (Sol com Sétima): `[196.00, 246.94, 293.66, 349.23]` (G3, B3, D4, F4)
  - `F` (Fá Maior): `[174.61, 220.00, 261.63]` (F3, A3, C4)
  - `C7` (Dó com Sétima): `[261.63, 329.63, 392.00, 466.16]` (C4, E4, G4, Bb4)

### 2. `pkg/synth/wave.go` — Síntese Polifônica com Soma de Senóides
- [x] Ao gerar amostras para um acorde, calcula o valor instantâneo de cada frequência ativa e soma as ondas dividindo pela quantidade de vozes para evitar saturação.

### 3. `pkg/audio/engine.go` — Stream Contínuo Não-Bloqueante
- [x] O driver de áudio (oto ou beep) lê de uma struct que implementa `io.Reader` sem nunca travar ou bloquear.
- [x] Se não houver notas ativas, preenche o buffer com silêncio e retorna `len(buf), nil`.
- [x] Nunca utiliza `time.Sleep` ou canais bloqueantes dentro do método `Read`.
