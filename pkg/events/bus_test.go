package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventBus(t *testing.T) {
	ctx := context.Background()

	// Create event bus
	bus := NewBus(false) // synchronous

	// Track events received
	eventsReceived := []Event{}
	handler := func(ctx context.Context, event Event) error {
		eventsReceived = append(eventsReceived, event)
		return nil
	}

	// Subscribe to test event
	bus.Subscribe("test.event", handler)

	// Publish event
	err := bus.Publish(ctx, "test.event", map[string]string{"message": "hello"})
	assert.NoError(t, err)

	// Verify event was received
	assert.Len(t, eventsReceived, 1)
	assert.Equal(t, "test.event", eventsReceived[0].Type)
	assert.Equal(t, map[string]string{"message": "hello"}, eventsReceived[0].Payload)
}

func TestEventBusAsync(t *testing.T) {
	ctx := context.Background()

	// Create async event bus
	bus := NewBus(true)

	eventsReceived := []Event{}
	handler := func(ctx context.Context, event Event) error {
		eventsReceived = append(eventsReceived, event)
		return nil
	}

	bus.Subscribe("async.event", handler)

	// Publish event
	err := bus.Publish(ctx, "async.event", "async payload")
	assert.NoError(t, err)

	// Give async handlers time to process
	time.Sleep(100 * time.Millisecond)

	// Verify event was received
	assert.Len(t, eventsReceived, 1)
	assert.Equal(t, "async.event", eventsReceived[0].Type)
	assert.Equal(t, "async payload", eventsReceived[0].Payload)
}

func TestEventBusMultipleHandlers(t *testing.T) {
	ctx := context.Background()

	bus := NewBus(false)

	handler1Called := false
	handler2Called := false

	handler1 := func(ctx context.Context, event Event) error {
		handler1Called = true
		return nil
	}

	handler2 := func(ctx context.Context, event Event) error {
		handler2Called = true
		return nil
	}

	bus.Subscribe("multi.event", handler1)
	bus.Subscribe("multi.event", handler2)

	err := bus.Publish(ctx, "multi.event", nil)
	assert.NoError(t, err)

	assert.True(t, handler1Called)
	assert.True(t, handler2Called)
}
