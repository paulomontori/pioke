package engine

import (
	"sync"
	"time"

	"pioke/pkg/model"
)

// Re-exporta ou mapeia os estados definidos em model para conveniência
const (
	STOPPED = model.STOPPED
	PLAYING = model.PLAYING
	PAUSED  = model.PAUSED
)

// Engine gerencia o ciclo de vida e tempo real de reprodução da música
type Engine struct {
	song       *model.Song
	state      model.PlaybackState
	positionMS int64
	ticker     *time.Ticker
	stopChan   chan struct{}
	eventsChan chan model.PlaybackEvent
	mu         sync.Mutex
}

// NewEngine cria uma nova instância de Engine para a música informada
func NewEngine(s *model.Song) *Engine {
	return &Engine{
		song:       s,
		state:      model.STOPPED,
		positionMS: 0,
		stopChan:   make(chan struct{}),
		eventsChan: make(chan model.PlaybackEvent, 100),
	}
}

// Events retorna o canal de leitura para eventos de reprodução
func (e *Engine) Events() <-chan model.PlaybackEvent {
	return e.eventsChan
}

// State retorna o estado atual do motor de reprodução
func (e *Engine) State() model.PlaybackState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

// Position retorna a posição atual do tempo de reprodução em ms
func (e *Engine) Position() int64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.positionMS
}

// Play inicia ou retoma a reprodução usando time.Ticker
func (e *Engine) Play() {
	e.mu.Lock()
	if e.state == model.PLAYING {
		e.mu.Unlock()
		return
	}
	e.state = model.PLAYING
	e.stopChan = make(chan struct{})
	e.mu.Unlock()

	go e.run()
}

// Start é um alias para Play para garantir compatibilidade
func (e *Engine) Start() {
	e.Play()
}

// Pause pausa a reprodução mantendo a posição atual
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state != model.PLAYING {
		return
	}
	e.state = model.PAUSED
	close(e.stopChan)
}

// Stop para a reprodução e reinicia a posição
func (e *Engine) Stop() {
	e.mu.Lock()
	if e.state == model.STOPPED {
		e.mu.Unlock()
		return
	}
	isPlaying := (e.state == model.PLAYING)
	e.state = model.STOPPED
	e.positionMS = 0
	if isPlaying {
		close(e.stopChan)
	}
	e.mu.Unlock()
}

// Seek altera a posição atual da reprodução em milissegundos
func (e *Engine) Seek(ms int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.positionMS = ms
}

// run é executado em uma goroutine e emite eventos conforme a linha do tempo avança
func (e *Engine) run() {
	interval := 10 * time.Millisecond // Resolução de ticker de 10ms
	e.ticker = time.NewTicker(interval)
	defer e.ticker.Stop()

	lastEmittedIndex := -1

	for {
		select {
		case <-e.stopChan:
			return
		case <-e.ticker.C:
			e.mu.Lock()
			e.positionMS += 10
			currentPos := e.positionMS
			currentState := e.state
			e.mu.Unlock()

			var activeEvent *model.TimelineEvent

			// Verifica se há eventos a emitir no tempo atual
			if e.song != nil {
				for i := range e.song.Timeline {
					event := &e.song.Timeline[i]
					eventTime := event.TimeMS
					if eventTime == 0 && event.Duration > 0 {
						eventTime = event.Duration.Milliseconds()
					}

					if i > lastEmittedIndex && eventTime <= currentPos {
						activeEvent = event
						lastEmittedIndex = i
					}
				}
			}

			e.eventsChan <- model.PlaybackEvent{
				CurrentTimeMS: currentPos,
				ActiveEvent:   activeEvent,
				State:         currentState,
			}
		}
	}
}
