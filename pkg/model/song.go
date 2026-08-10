package model

import "time"

// Song representa a estrutura completa de uma música no PioKe
type Song struct {
	Metadata Metadata        `json:"metadata" yaml:"metadata"`
	Title    string          `json:"title,omitempty" yaml:"title,omitempty"`
	Artist   string          `json:"artist,omitempty" yaml:"artist,omitempty"`
	BPM      int             `json:"bpm,omitempty" yaml:"bpm,omitempty"`
	Key      string          `json:"key,omitempty" yaml:"key,omitempty"`
	TimeSig  string          `json:"time_signature,omitempty" yaml:"time_signature,omitempty"`
	Timeline []TimelineEvent `json:"timeline" yaml:"timeline"`
}

// Metadata armazena as informações gerais sobre a música
type Metadata struct {
	Title   string `json:"title,omitempty" yaml:"title,omitempty"`
	Artist  string `json:"artist,omitempty" yaml:"artist,omitempty"`
	BPM     int    `json:"bpm,omitempty" yaml:"bpm,omitempty"`
	Key     string `json:"key,omitempty" yaml:"key,omitempty"`
	TimeSig string `json:"time_signature,omitempty" yaml:"time_signature,omitempty"`
}

// TimelineEvent representa um evento sincronizado na linha do tempo
type TimelineEvent struct {
	Timestamp  string        `json:"timestamp,omitempty" yaml:"timestamp,omitempty"` // ex: "01:23.45"
	TimeMS     int64         `json:"time_ms,omitempty" yaml:"time_ms,omitempty"`     // tempo em ms
	DurationMS int64         `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
	Duration   time.Duration `json:"-" yaml:"-"`
	Lyric      string        `json:"lyric,omitempty" yaml:"lyric,omitempty"`
	ChordStr   string        `json:"chord,omitempty" yaml:"chord,omitempty"`
	Lyrics     *LyricLine    `json:"lyrics,omitempty" yaml:"lyrics,omitempty"`
	Chord      *ChordEvent   `json:"chord_obj,omitempty" yaml:"chord_obj,omitempty"`
}

// LyricLine representa uma linha ou palavra da letra e seu tempo
type LyricLine struct {
	Text      string `json:"text" yaml:"text"`
	Syllables []Word `json:"syllables,omitempty" yaml:"syllables,omitempty"`
}

// Word para sincronização detalhada de karaokê
type Word struct {
	Text      string `json:"text" yaml:"text"`
	Timestamp string `json:"timestamp" yaml:"timestamp"`
}

// ChordEvent representa a execução de um acorde
type ChordEvent struct {
	Name     string `json:"name" yaml:"name"`         // ex: "Cmaj7", "Am"
	Duration string `json:"duration" yaml:"duration"` // ex: "1b" (1 beat), "2s" (2 seconds)
}
