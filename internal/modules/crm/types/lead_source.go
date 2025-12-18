package types

import (
	"time"

	"github.com/google/uuid"
)

// LeadSource represents a source of leads
type LeadSource struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name          string     `json:"name" db:"name"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// LeadSourceFilter represents filtering criteria for lead sources
type LeadSourceFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Limit          int
	Offset         int
}

// LeadSourceCreateRequest represents a request to create a lead source
type LeadSourceCreateRequest struct {
	Name string `json:"name"`
}

// LeadSourceUpdateRequest represents a request to update a lead source
type LeadSourceUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}
