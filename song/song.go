package song

import "time"

// Song representa a estrutura completa de uma música no OpenTune
type Song struct {
	Metadata Metadata        `json:"metadata" yaml:"metadata"`
	Timeline []TimelineEvent `json:"timeline" yaml:"timeline"`
}

// Metadata armazena as informações gerais sobre a música
type Metadata struct {
	Title   string `json:"title" yaml:"title"`
	Artist  string `json:"artist" yaml:"artist"`
	BPM     int    `json:"bpm" yaml:"bpm"`
	Key     string `json:"key" yaml:"key"`
	TimeSig string `json:"time_signature" yaml:"time_signature"`
}

// TimelineEvent representa um evento sincronizado na linha do tempo
type TimelineEvent struct {
	Timestamp string        `json:"timestamp" yaml:"timestamp"` // ex: "01:23.45"
	Duration  time.Duration `json:"-" yaml:"-"`
	Lyrics    *LyricLine    `json:"lyrics,omitempty" yaml:"lyrics,omitempty"`
	Chord     *ChordEvent   `json:"chord,omitempty" yaml:"chord,omitempty"`
}

// LyricLine representa uma linha ou palavra da letra e seu tempo
type LyricLine struct {
	Text      string `json:"text" yaml:"text"`
	Syllables []Word `json:"syllables,omitempty" yaml:"syllables,omitempty"`
}

// Word/Syllable para sincronização detalhada de karaokê
type Word struct {
	Text      string `json:"text" yaml:"text"`
	Timestamp string `json:"timestamp" yaml:"timestamp"`
}

// ChordEvent representa a execução de um acorde
type ChordEvent struct {
	Name     string `json:"name" yaml:"name"`         // ex: "Cmaj7", "Am"
	Duration string `json:"duration" yaml:"duration"` // ex: "1b" (1 beat), "2s" (2 seconds)
}
