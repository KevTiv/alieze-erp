package types

import (
	"time"

	"github.com/google/uuid"
)

// LostReason represents a reason why a lead was lost
type LostReason struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name          string     `json:"name" db:"name"`
	Active        bool       `json:"active" db:"active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// LostReasonFilter represents filtering criteria for lost reasons
type LostReasonFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Active         *bool
	Limit          int
	Offset         int
}

// LostReasonCreateRequest represents a request to create a lost reason
type LostReasonCreateRequest struct {
	Name    string `json:"name"`
	Active  bool   `json:"active"`
}

// LostReasonUpdateRequest represents a request to update a lost reason
type LostReasonUpdateRequest struct {
	Name    *string `json:"name,omitempty"`
	Active  *bool   `json:"active,omitempty"`
}
