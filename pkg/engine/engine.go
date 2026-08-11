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
	defer close(e.eventsChan)

	activeIndex := -1

	// Calcula a duração total da música
	var totalDurationMS int64
	if e.song != nil && len(e.song.Timeline) > 0 {
		lastEvent := e.song.Timeline[len(e.song.Timeline)-1]
		dur := lastEvent.DurationMS
		if dur <= 0 {
			dur = 2000
		}
		totalDurationMS = lastEvent.TimeMS + dur
	}

	// Captura o tempo real de início para evitar drift (atraso acumulado)
	e.mu.Lock()
	startPos := e.positionMS
	e.mu.Unlock()
	playStart := time.Now()

	for {
		select {
		case <-e.stopChan:
			// Salva a posição exata ao pausar
			e.mu.Lock()
			if e.state == model.PAUSED {
				e.positionMS = startPos + time.Since(playStart).Milliseconds()
			}
			e.mu.Unlock()
			return
		case <-e.ticker.C:
			e.mu.Lock()
			// Calcula a posição baseada no relógio real, não em contagem de ticks
			currentPos := startPos + time.Since(playStart).Milliseconds()
			e.positionMS = currentPos
			currentState := e.state
			e.mu.Unlock()

			// Avança o índice do evento ativo conforme a linha do tempo é alcançada
			var activeEvent *model.TimelineEvent
			if e.song != nil {
				for activeIndex+1 < len(e.song.Timeline) && e.song.Timeline[activeIndex+1].TimeMS <= currentPos {
					activeIndex++
				}

				// O evento permanece ativo durante toda a sua duração — não apenas no tick em que
				// começou — para que consumidores (áudio, gravação WAV) não confundam "nenhum evento
				// novo neste tick" com "nada tocando".
				if activeIndex >= 0 {
					event := &e.song.Timeline[activeIndex]
					withinWindow := true
					switch {
					case event.DurationMS > 0:
						withinWindow = currentPos < event.TimeMS+event.DurationMS
					case activeIndex+1 < len(e.song.Timeline):
						withinWindow = currentPos < e.song.Timeline[activeIndex+1].TimeMS
					case totalDurationMS > 0:
						withinWindow = currentPos < totalDurationMS
					}
					if withinWindow {
						activeEvent = event
					}
				}
			}

			e.eventsChan <- model.PlaybackEvent{
				CurrentTimeMS: currentPos,
				ActiveEvent:   activeEvent,
				State:         currentState,
			}

			// Se ultrapassou o fim da música, encerra a reprodução
			if totalDurationMS > 0 && currentPos >= totalDurationMS+500 {
				e.mu.Lock()
				e.state = model.STOPPED
				e.mu.Unlock()
				return
			}
		}
	}
}
