package types

import (
	"time"
	"github.com/google/uuid"
)

// ProcurementGroup represents a procurement group for organizing purchases
type ProcurementGroup struct {
	ID            uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name          string    `json:"name" db:"name"`
	PartnerID     *uuid.UUID `json:"partner_id" db:"partner_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// ProcurementGroupCreateRequest represents the request to create a procurement group
type ProcurementGroupCreateRequest struct {
	Name      string    `json:"name" validate:"required,min=1,max=255"`
	PartnerID *uuid.UUID `json:"partner_id"`
}

// ProcurementGroupUpdateRequest represents the request to update a procurement group
type ProcurementGroupUpdateRequest struct {
	Name      *string    `json:"name"`
	PartnerID *uuid.UUID `json:"partner_id"`
}
