// Package audit реализует паттерн Observer для логирования изменений метрик
package audit

import (
	"log/slog"
	"sync"
	"time"
)

// Event описывает событие изменения метрик
type Event struct {
	// TS — Unix-время события в секундах
	TS int64 `json:"ts"`
	// Metrics — имена изменённых метрик
	Metrics []string `json:"metrics"`
	// IPAddress — IP-адрес клиента, инициировавшего изменение
	IPAddress string `json:"ip_address"`
}

// Observer получает уведомления об изменениях метрик
type Observer interface {
	// Update обрабатывает событие изменения метрик
	Update(event Event)
	// ID возвращает уникальный идентификатор наблюдателя
	ID() string
}

const observerQueueSize = 64

type observerRunner struct {
	observer Observer
	events   chan Event
	done     chan struct{}
}

func newObserverRunner(o Observer) *observerRunner {
	r := &observerRunner{
		observer: o,
		events:   make(chan Event, observerQueueSize),
		done:     make(chan struct{}),
	}
	go func() {
		defer close(r.done)
		for event := range r.events {
			r.observer.Update(event)
		}
	}()
	return r
}

func (r *observerRunner) notify(event Event) {
	select {
	case r.events <- event:
	default:
	}
}

func (r *observerRunner) stop() {
	close(r.events)
	<-r.done
}

func closeObserver(o Observer) {
	if c, ok := o.(interface{ Close() error }); ok {
		_ = c.Close()
	}
}

// Publisher рассылает события аудита зарегистрированным наблюдателям
type Publisher struct {
	mu        sync.RWMutex
	observers map[string]*observerRunner
}

// NewPublisher создаёт новый Publisher и регистрирует наблюдателей для файла и URL
func NewPublisher(logger *slog.Logger, auditFile, auditURL string) *Publisher {
	p := &Publisher{observers: make(map[string]*observerRunner)}
	if auditFile != "" {
		p.Register(NewFileObserver(logger, auditFile))
	}
	if auditURL != "" {
		p.Register(NewURLObserver(logger, auditURL))
	}
	return p
}

// Register регистрирует новый наблюдатель
func (p *Publisher) Register(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.observers[o.ID()]; ok {
		r.stop()
		closeObserver(r.observer)
	}
	p.observers[o.ID()] = newObserverRunner(o)
}

// Deregister удаляет наблюдателя из Publisher
func (p *Publisher) Deregister(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.observers[o.ID()]; ok {
		r.stop()
		closeObserver(r.observer)
		delete(p.observers, o.ID())
	}
}

// Close останавливает всех наблюдателей и освобождает их ресурсы
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, r := range p.observers {
		r.stop()
		closeObserver(r.observer)
		delete(p.observers, id)
	}
	return nil
}

// Notify отправляет событие всем наблюдателям
func (p *Publisher) Notify(metrics []string, ipAddress string) {
	if len(metrics) == 0 {
		return
	}

	event := Event{
		TS:        time.Now().Unix(),
		Metrics:   append([]string(nil), metrics...),
		IPAddress: ipAddress,
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, r := range p.observers {
		r.notify(event)
	}
}

// Enabled проверяет, включен ли Publisher
func (p *Publisher) Enabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.observers) > 0
}
