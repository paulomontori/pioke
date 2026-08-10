package ui

import (
	"pioke/pkg/model"
)

// PlaybackEvent representa o estado transmitido para a UI a cada tique ou evento
type PlaybackEvent struct {
	Song     *model.Song
	Current  *model.TimelineEvent
	Position int64 // Tempo decorrido em milissegundos
}

// Renderer define a interface genérica para exibições de UI (TUI, GUI, Web)
type Renderer interface {
	Init() error
	DisplayHeader(s *model.Song)
	RenderTick(event PlaybackEvent) error
	Close() error
}
