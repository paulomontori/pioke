package audio

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"pioke/pkg/synth"

	"github.com/ebitengine/oto/v3"
)

// Synth representa a interface de saída de áudio: um contexto Oto de longa duração que toca um
// buffer PCM pré-renderizado (o mesmo produzido por synth.RenderSequence para o WAV), em vez de
// sintetizar nota a nota em tempo real a partir de um agendador. Isso garante que a reprodução ao
// vivo soa exatamente como o arquivo exportado: sem risco de uma troca de nota deixar de ser
// amostrada por atraso do agendador em tempo real (GC, troca de goroutine, jitter do SO, etc.) —
// quem faz a temporização real é o próprio driver de áudio, lendo de um buffer já correto.
type Synth struct {
	ctx     *oto.Context
	player  *oto.Player
	enabled bool
}

var audioDebug = os.Getenv("PIOKE_AUDIO_DEBUG") != ""

// NewSynth inicializa o contexto de áudio do Oto v3. Chame Play para iniciar a reprodução de um
// buffer PCM.
func NewSynth() *Synth {
	op := &oto.NewContextOptions{
		SampleRate:   synth.SampleRate,
		ChannelCount: synth.ChannelCount,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, ready, err := oto.NewContext(op)
	if err != nil {
		fmt.Printf("[AUDIO ENGINE] Aviso: Dispositivo de áudio indisponível: %v\n", err)
		return &Synth{enabled: false}
	}

	<-ready

	return &Synth{ctx: otoCtx, enabled: true}
}

// Play inicia a reprodução do buffer PCM estéreo 16-bit informado — tipicamente
// synth.RenderSequence(segments), a mesma sequência determinística usada na gravação WAV.
func (s *Synth) Play(pcm []byte) {
	if !s.enabled || len(pcm) == 0 {
		return
	}

	player := s.ctx.NewPlayer(bytes.NewReader(pcm))
	player.Play()
	s.player = player

	if audioDebug {
		go func() {
			for i := 0; i < 30; i++ {
				time.Sleep(time.Second)
				fmt.Printf("[AUDIO DEBUG t=%ds] IsPlaying=%v BufferedSize=%d Err=%v Volume=%v\n",
					i+1, player.IsPlaying(), player.BufferedSize(), player.Err(), player.Volume())
			}
		}()
	}
}

// Stop pausa a reprodução em andamento, se houver.
func (s *Synth) Stop() {
	if s.player != nil {
		s.player.Pause()
	}
}
