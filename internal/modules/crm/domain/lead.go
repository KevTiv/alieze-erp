package domain

import (
	"time"

	"github.com/google/uuid"
)

// Lead represents a sales lead
type Lead struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Email          *string    `json:"email,omitempty" db:"email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	StageID        *uuid.UUID `json:"stage_id,omitempty" db:"stage_id"`
	Status         string     `json:"status" db:"status"`
	Active         bool       `json:"active" db:"active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// LeadFilter represents filtering criteria for leads
type LeadFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Email          *string
	Phone          *string
	StageID        *uuid.UUID
	Status         *string
	Active         *bool
	Limit          int
	Offset         int
}
