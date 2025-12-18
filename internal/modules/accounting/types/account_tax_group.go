package types

import (
	"time"

	"github.com/google/uuid"
)

// AccountTaxGroup represents a tax group for organizing related taxes
type AccountTaxGroup struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name          string     `json:"name" db:"name"`
	Sequence      int        `json:"sequence" db:"sequence"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// AccountTaxGroupFilter represents filtering criteria for tax groups
type AccountTaxGroupFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Limit          int
	Offset         int
}

// AccountTaxGroupCreateRequest represents a request to create a tax group
type AccountTaxGroupCreateRequest struct {
	Name     string `json:"name"`
	Sequence int    `json:"sequence"`
}

// AccountTaxGroupUpdateRequest represents a request to update a tax group
type AccountTaxGroupUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Sequence *int    `json:"sequence,omitempty"`
}
