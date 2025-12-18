package types

import (
	"time"

	"github.com/google/uuid"
)

// State represents a state/province (ISO 3166-2)
type State struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CountryID uuid.UUID  `json:"country_id" db:"country_id"`
	Name      string     `json:"name" db:"name"`
	Code      string     `json:"code" db:"code"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// StateFilter for querying states
type StateFilter struct {
	CountryID *uuid.UUID
	Code      *string
	Name      *string
	Limit     int
	Offset    int
}

// StateCreateRequest represents a request to create a state
type StateCreateRequest struct {
	CountryID uuid.UUID `json:"country_id" validate:"required"`
	Name      string    `json:"name" validate:"required"`
	Code      string    `json:"code" validate:"required"`
}

// StateUpdateRequest represents a request to update a state
type StateUpdateRequest struct {
	CountryID *uuid.UUID `json:"country_id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Code      *string    `json:"code,omitempty"`
}
