package types

import (
	"time"

	"github.com/google/uuid"
)

// LeadStage represents a stage in the lead/opportunity pipeline
type LeadStage struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name         string     `json:"name" db:"name"`
	Sequence     int        `json:"sequence" db:"sequence"`
	Probability  int        `json:"probability" db:"probability"`
	Fold         bool       `json:"fold" db:"fold"`
	IsWon        bool       `json:"is_won" db:"is_won"`
	Requirements *string    `json:"requirements,omitempty" db:"requirements"`
	TeamID       *uuid.UUID `json:"team_id,omitempty" db:"team_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// LeadStageFilter represents filtering criteria for lead stages
type LeadStageFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	IsWon          *bool
	TeamID         *uuid.UUID
	Limit          int
	Offset         int
}

// LeadStageCreateRequest represents a request to create a lead stage
type LeadStageCreateRequest struct {
	Name        string     `json:"name"`
	Sequence    int        `json:"sequence"`
	Probability int        `json:"probability"`
	Fold        bool       `json:"fold"`
	IsWon       bool       `json:"is_won"`
	Requirements *string    `json:"requirements,omitempty"`
	TeamID      *uuid.UUID `json:"team_id,omitempty"`
}

// LeadStageUpdateRequest represents a request to update a lead stage
type LeadStageUpdateRequest struct {
	Name        *string    `json:"name,omitempty"`
	Sequence    *int       `json:"sequence,omitempty"`
	Probability *int       `json:"probability,omitempty"`
	Fold        *bool      `json:"fold,omitempty"`
	IsWon       *bool      `json:"is_won,omitempty"`
	Requirements *string    `json:"requirements,omitempty"`
	TeamID      *uuid.UUID `json:"team_id,omitempty"`
}
