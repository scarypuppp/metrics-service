package audit

import (
	"fmt"
	"sync"
	"time"

	models "github.com/scarypuppp/metrics-service/internal/model"
)

type Event struct {
	Timestamp time.Time
	Metrics   []models.Metrics
	Address   string
}

type Publisher struct {
	mu          sync.RWMutex
	subscribers map[string]chan Event
}

func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make(map[string]chan Event),
	}
}

func (p *Publisher) Subscribe(id string) <-chan Event {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan Event, 10)
	p.subscribers[id] = ch
	return ch
}

func (p *Publisher) Unsubscribe(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ch, ok := p.subscribers[id]; ok {
		close(ch)
		delete(p.subscribers, id)
	}
}

func (p *Publisher) Publish(event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for id, ch := range p.subscribers {
		select {
		case ch <- event:
		default:
			fmt.Printf("подписчик %s не успевает, событие пропущено\n", id)
		}
	}
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, ch := range p.subscribers {
		close(ch)
		delete(p.subscribers, id)
	}
}
