# Refatoração Crítica: Polifonia Real e Stream Não-Bloqueante (pkg/synth & pkg/audio)

**Problemas Identificados:**
1. Os acordes não estão soando polifônicos (apenas uma nota estática ou timbre incorreto é ouvido).
2. O áudio sofre engasgos constantes/tremolo devido a bloqueios na implementação do `io.Reader` do stream de áudio.

---

## 🛠️ Especificação da Correção

### 1. `pkg/synth/chord.go` — Definição Precisa dos Acordes
Garanta que cada acorde retorne um slice com as frequências reais em Hz de **todas** as suas notas:
- `C` (Dó Maior): `[261.63, 329.63, 392.00]` (C4, E4, G4)
- `G7` (Sol com Sétima): `[196.00, 246.94, 293.66, 349.23]` (G3, B3, D4, F4)
- `F` (Fá Maior): `[174.61, 220.00, 261.63]` (F3, A3, C4)
- `C7` (Dó com Sétima): `[261.63, 329.63, 392.00, 466.16]` (C4, E4, G4, Bb4)

### 2. `pkg/synth/wave.go` — Síntese Polifônica com Soma de Senóides
Ao gerar amostras para um acorde:
- Para cada amostra $t$, calcule o valor instantâneo de cada frequência ativa.
- **Some as ondas e normalize pela quantidade de vozes** para evitar saturação:
  ```go
  var sample float64
  for _, freq := range activeFrequencies {
      sample += math.Sin(2 * math.Pi * freq * phase)
  }
  sample /= float64(len(activeFrequencies))


### 3. pkg/audio/stream.go — Stream Contínuo Não-Bloqueante
O driver de áudio (oto ou beep) deve ler de uma struct que implementa io.Reader sem nunca travar ou bloquear:

Go
func (s *SynthStream) Read(buf []byte) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Se não houver notas ativas, preencha 'buf' com zeros (silêncio) e retorne len(buf), nil.
    // Se houver notas ativas, calcule as amostras PCM e converta para int16 (little-endian).
    return len(buf), nil
}
Nunca use time.Sleep ou canais bloqueantes dentro do método Read.

Quando um novo acorde for acionado via PlaybackEvent, apenas atualize o slice activeFrequencies protegido por mutex.