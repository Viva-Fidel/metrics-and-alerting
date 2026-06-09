package audit

import (
	"sync"
	"time"
)

type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Observer interface {
	Update(event Event)
	ID() string
}

type Publisher struct {
	mu        sync.RWMutex
	observers map[string]Observer
}

// NewPublisher создает новый Publisher и регистрирует наблюдателей для файла и URL
func NewPublisher(auditFile, auditURL string) *Publisher {
	p := &Publisher{observers: make(map[string]Observer)}
	if auditFile != "" {
		p.Register(NewFileObserver(auditFile))
	}
	if auditURL != "" {
		p.Register(NewURLObserver(auditURL))
	}
	return p
}

// Register регистрирует новый наблюдатель
func (p *Publisher) Register(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers[o.ID()] = o
}

// Deregister удаляет наблюдателя из Publisher
func (p *Publisher) Deregister(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.observers, o.ID())
}

// Notify отправляет событие всем наблюдателям
func (p *Publisher) Notify(metrics []string, ipAddress string) {
	if p == nil || len(metrics) == 0 {
		return
	}

	event := Event{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddress,
	}

	p.mu.RLock()
	observers := make([]Observer, 0, len(p.observers))
	for _, o := range p.observers {
		observers = append(observers, o)
	}
	p.mu.RUnlock()

	for _, o := range observers {
		o.Update(event)
	}
}

// Enabled проверяет, включен ли Publisher
func (p *Publisher) Enabled() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.observers) > 0
}
