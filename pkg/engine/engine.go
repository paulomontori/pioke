package engine

import (
	"sync"
	"time"

	"opentune/song"
)

// PlaybackState representa o estado atual da reprodução
type PlaybackState int

const (
	StateStopped PlaybackState = iota
	StatePlaying
	StatePaused
)

// Engine gerencia o ciclo de vida e tempo real de reprodução da música
type Engine struct {
	song       *song.Song
	state      PlaybackState
	position   time.Duration
	ticker     *time.Ticker
	stopChan   chan struct{}
	eventsChan chan song.TimelineEvent
	mu         sync.Mutex
}

// NewEngine cria uma nova instância de Engine para a música informada
func NewEngine(s *song.Song) *Engine {
	return &Engine{
		song:       s,
		state:      StateStopped,
		position:   0,
		stopChan:   make(chan struct{}),
		eventsChan: make(chan song.TimelineEvent, 100),
	}
}

// Events retorna o canal de leitura para eventos da linha do tempo
func (e *Engine) Events() <-chan song.TimelineEvent {
	return e.eventsChan
}

// State retorna o estado atual do motor de reprodução
func (e *Engine) State() PlaybackState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

// Position retorna a posição atual do tempo de reprodução
func (e *Engine) Position() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.position
}

// Start inicia ou retoma a reprodução usando time.Ticker
func (e *Engine) Start() {
	e.mu.Lock()
	if e.state == StatePlaying {
		e.mu.Unlock()
		return
	}
	e.state = StatePlaying
	e.stopChan = make(chan struct{})
	e.mu.Unlock()

	go e.run()
}

// Pause pausa a reprodução mantendo a posição atual
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state != StatePlaying {
		return
	}
	e.state = StatePaused
	close(e.stopChan)
}

// Stop para a reprodução e reinicia a posição
func (e *Engine) Stop() {
	e.mu.Lock()
	if e.state == StateStopped {
		e.mu.Unlock()
		return
	}
	isPlaying := (e.state == StatePlaying)
	e.state = StateStopped
	e.position = 0
	if isPlaying {
		close(e.stopChan)
	}
	e.mu.Unlock()
}

// Seek altera a posição atual da reprodução
func (e *Engine) Seek(d time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.position = d
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
			e.position += interval
			currentPos := e.position
			e.mu.Unlock()

			// Verifica se há eventos a emitir no tempo atual
			for i, event := range e.song.Timeline {
				if i > lastEmittedIndex && event.Duration <= currentPos {
					e.eventsChan <- event
					lastEmittedIndex = i
				}
			}
		}
	}
}
