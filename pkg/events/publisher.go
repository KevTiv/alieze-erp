package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventPublisher defines interface for publishing events
type EventPublisher interface {
	// Publish publishes an event with the given type and payload
	Publish(ctx context.Context, eventType string, payload interface{}) error

	// PublishAsync publishes an event asynchronously
	PublishAsync(ctx context.Context, eventType string, payload interface{}) error

	// PublishWithMetadata publishes an event with additional metadata
	PublishWithMetadata(ctx context.Context, eventType string, payload interface{}, metadata EventMetadata) error
}

// EventMetadata contains additional metadata for events
type EventMetadata struct {
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	CorrelationID  *string    `json:"correlation_id,omitempty"`
	Source         string     `json:"source,omitempty"`
	Version        string     `json:"version,omitempty"`
	Timestamp      time.Time  `json:"timestamp"`
}

// EnhancedEvent represents an enhanced event with metadata
type EnhancedEvent struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Payload   interface{}     `json:"payload"`
	Metadata  EventMetadata   `json:"metadata"`
	Timestamp time.Time       `json:"timestamp"`
	Context   context.Context `json:"-"`
}

// BusEventPublisher wraps the existing Bus to implement EventPublisher
type BusEventPublisher struct {
	bus *Bus
}

// NewBusEventPublisher creates a new EventPublisher using the Bus
func NewBusEventPublisher(bus *Bus) EventPublisher {
	return &BusEventPublisher{bus: bus}
}

// Publish publishes an event with the given type and payload
func (p *BusEventPublisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	metadata := EventMetadata{
		Timestamp: time.Now(),
		Source:    "crm-service",
		Version:   "1.0",
	}

	return p.PublishWithMetadata(ctx, eventType, payload, metadata)
}

// PublishAsync publishes an event asynchronously
func (p *BusEventPublisher) PublishAsync(ctx context.Context, eventType string, payload interface{}) error {
	// For now, implement async publishing in a goroutine
	// TODO: Add proper async error handling and retry logic
	go func() {
		_ = p.Publish(ctx, eventType, payload)
	}()
	return nil
}

// PublishWithMetadata publishes an event with additional metadata
func (p *BusEventPublisher) PublishWithMetadata(ctx context.Context, eventType string, payload interface{}, metadata EventMetadata) error {
	// Enhance the payload with metadata if needed
	enhancedPayload := map[string]interface{}{
		"data":     payload,
		"metadata": metadata,
	}

	return p.bus.Publish(ctx, eventType, enhancedPayload)
}
