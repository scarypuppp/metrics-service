package audit

import (
	"fmt"
	"sync"
)

// Event is an audit record describing a metrics update: when, which metrics and from what address.
type Event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	Address   string   `json:"ip_address"`
}

// Publisher fans out audit events to registered subscribers over buffered channels.
type Publisher struct {
	mu          sync.RWMutex
	subscribers map[string]chan Event
}

// NewPublisher creates a Publisher with no subscribers.
func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make(map[string]chan Event),
	}
}

// Subscribe registers a subscriber under the given id and returns its event channel.
func (p *Publisher) Subscribe(id string) <-chan Event {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan Event, 10)
	p.subscribers[id] = ch
	return ch
}

// Unsubscribe removes the subscriber with the given id and closes its channel.
func (p *Publisher) Unsubscribe(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ch, ok := p.subscribers[id]; ok {
		close(ch)
		delete(p.subscribers, id)
	}
}

// Publish delivers the event to all subscribers, skipping those whose channels are full.
func (p *Publisher) Publish(event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for id, ch := range p.subscribers {
		select {
		case ch <- event:
		default:
			fmt.Printf("event skipped %s", id)
		}
	}
}

// Close removes all subscribers and closes their channels.
func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, ch := range p.subscribers {
		close(ch)
		delete(p.subscribers, id)
	}
}
