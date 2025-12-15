package events

import (
	"context"
	"time"
)

// Event represents a domain event in the system
type Event struct {
	Type      string
	Payload   interface{}
	Source    string
	Timestamp time.Time
	Context   context.Context
}
