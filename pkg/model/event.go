package model

// PlaybackState representa o estado atual da reprodução
type PlaybackState int

const (
	STOPPED PlaybackState = iota
	PLAYING
	PAUSED
)

// PlaybackEvent representa o estado do evento de reprodução emitido pelo engine
type PlaybackEvent struct {
	CurrentTimeMS int64          `json:"current_time_ms"`
	ActiveEvent   *TimelineEvent `json:"active_event,omitempty"`
	State         PlaybackState  `json:"state"`
}
