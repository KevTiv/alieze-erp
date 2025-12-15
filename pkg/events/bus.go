package events

import (
	"context"
	"sync"
	"time"
)

// HandlerFunc is a function that handles events
type HandlerFunc func(ctx context.Context, event Event) error

// Bus is an event bus that manages event publishing and subscription
type Bus struct {
	handlers map[string][]HandlerFunc
	mu       sync.RWMutex
	async    bool
}

// NewBus creates a new event bus
func NewBus(async bool) *Bus {
	return &Bus{
		handlers: make(map[string][]HandlerFunc),
		async:    async,
	}
}

// Publish sends an event to all subscribers
func (b *Bus) Publish(ctx context.Context, eventType string, payload interface{}) error {
	event := Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
		Context:   ctx,
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, exists := b.handlers[eventType]; exists {
		for _, handler := range handlers {
			if b.async {
				go func(h HandlerFunc) {
					_ = h(ctx, event)
				}(handler)
			} else {
				if err := handler(ctx, event); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Subscribe adds a handler for a specific event type
func (b *Bus) Subscribe(eventType string, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}
