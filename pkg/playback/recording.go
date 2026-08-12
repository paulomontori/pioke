package playback

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"pioke/pkg/audio"
	"pioke/pkg/model"
	"pioke/pkg/synth"
)

// RecordingMeta descreve uma gravação de microfone salva junto de uma sessão de reprodução —
// suficiente pra, depois, alinhar o áudio cantado com a timeline esperada da música (offset entre
// o início da gravação e o início real da reprodução) e reproduzi-lo com o contexto certo (qual
// música, qual timbre estava tocando).
type RecordingMeta struct {
	SongFile   string `json:"song_file"`
	SongTitle  string `json:"song_title,omitempty"`
	SongArtist string `json:"song_artist,omitempty"`
	Timbre     string `json:"timbre"`
	SampleRate int    `json:"sample_rate"`
	Channels   int    `json:"channels"`

	// RecordingStartedAt é o instante (RFC3339) em que a captura do microfone começou de fato.
	RecordingStartedAt string `json:"recording_started_at"`

	// PlaybackOffsetMS é quantos milissegundos depois do início da gravação a reprodução da
	// música realmente começou a tocar — o instante 0 da timeline da música corresponde a este
	// offset dentro do WAV gravado, não ao início do arquivo.
	PlaybackOffsetMS int64 `json:"playback_offset_ms"`
}

// startRecording inicia a captura do microfone, se solicitada; retorna nil sem erro quando record
// é false. Falhas ao iniciar a captura (ex: sem microfone disponível) são reportadas mas não
// interrompem a reprodução — cantar sem gravar ainda é melhor que não tocar a música.
func startRecording(record bool) (*audio.MicRecorder, time.Time) {
	if !record {
		return nil, time.Time{}
	}
	rec, err := audio.NewMicRecorder(synth.SampleRate)
	if err != nil {
		fmt.Printf("\n[gravação] aviso: não foi possível captar o microfone (%v) — a música vai tocar normalmente, sem gravar.\n", err)
		return nil, time.Time{}
	}
	return rec, time.Now()
}

// finishRecording para a captura e salva o WAV + metadados em recordings/. songFile é o caminho
// original da música (guardado no metadata pra permitir recarregar a timeline esperada depois).
func finishRecording(rec *audio.MicRecorder, recordStart, playbackStart time.Time, s *model.Song, songFile string, timbre synth.Timbre) {
	if rec == nil {
		return
	}
	pcm := rec.Stop()
	if len(pcm) == 0 {
		fmt.Println("\n[gravação] nenhum áudio foi captado do microfone.")
		return
	}

	if err := os.MkdirAll("recordings", 0o755); err != nil {
		fmt.Printf("\n[gravação] erro ao criar diretório recordings/: %v\n", err)
		return
	}

	slug := slugify(s.Title)
	if slug == "" {
		slug = "musica"
	}
	base := fmt.Sprintf("recordings/%s_%s", slug, recordStart.Format("20060102-150405"))

	wavPath := base + ".wav"
	if err := audio.WriteWAV(wavPath, pcm, synth.SampleRate, 1); err != nil {
		fmt.Printf("\n[gravação] erro ao salvar WAV: %v\n", err)
		return
	}

	meta := RecordingMeta{
		SongFile:            songFile,
		SongTitle:           s.Title,
		SongArtist:          s.Artist,
		Timbre:              string(timbre),
		SampleRate:          synth.SampleRate,
		Channels:            1,
		RecordingStartedAt:  recordStart.Format(time.RFC3339),
		PlaybackOffsetMS:    playbackStart.Sub(recordStart).Milliseconds(),
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		fmt.Printf("\n[gravação] erro ao serializar metadados: %v\n", err)
		return
	}
	metaPath := base + ".json"
	if err := os.WriteFile(metaPath, metaBytes, 0o644); err != nil {
		fmt.Printf("\n[gravação] erro ao salvar metadados: %v\n", err)
		return
	}

	fmt.Printf("\nGravação salva em %s (metadados em %s)\n", wavPath, metaPath)
}

// slugify converte um título de música em um nome de arquivo seguro e legível: acentos viram a
// letra base (ex: "Parabéns" -> "parabens", não "parab-ns"), minúsculas, apenas
// letras/números/hífen.
func slugify(title string) string {
	ascii, _, err := transform.String(
		transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		title,
	)
	if err != nil {
		ascii = title
	}

	var b strings.Builder
	lastDash := true // evita hífen duplicado/inicial
	for _, r := range strings.ToLower(ascii) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
