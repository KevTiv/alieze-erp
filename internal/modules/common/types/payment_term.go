package types

import (
	"time"

	"github.com/google/uuid"
)

// PaymentTerm represents payment terms for an organization
type PaymentTerm struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID     *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name          string     `json:"name" db:"name"`
	Note          string     `json:"note" db:"note"`
	Active        bool       `json:"active" db:"active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PaymentTermFilter for querying payment terms
type PaymentTermFilter struct {
	OrganizationID *uuid.UUID
	CompanyID     *uuid.UUID
	Active        *bool
	Name          *string
	Limit         int
	Offset        int
}

// PaymentTermCreateRequest represents a request to create payment terms
type PaymentTermCreateRequest struct {
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	CompanyID     *uuid.UUID `json:"company_id,omitempty"`
	Name          string     `json:"name" validate:"required"`
	Note          string     `json:"note"`
	Active        bool       `json:"active"`
}

// PaymentTermUpdateRequest represents a request to update payment terms
type PaymentTermUpdateRequest struct {
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Note      *string    `json:"note,omitempty"`
	Active    *bool      `json:"active,omitempty"`
}
