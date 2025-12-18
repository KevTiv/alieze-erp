package types

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticAccount represents an analytic account (cost center)
type AnalyticAccount struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID     *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name          string     `json:"name" db:"name"`
	Code          string     `json:"code" db:"code"`
	PartnerID     *uuid.UUID `json:"partner_id,omitempty" db:"partner_id"`
	Active        bool       `json:"active" db:"active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// AnalyticAccountFilter for querying analytic accounts
type AnalyticAccountFilter struct {
	OrganizationID *uuid.UUID
	CompanyID     *uuid.UUID
	Active        *bool
	Name          *string
	Code          *string
	Limit         int
	Offset        int
}

// AnalyticAccountCreateRequest represents a request to create an analytic account
type AnalyticAccountCreateRequest struct {
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	CompanyID     *uuid.UUID `json:"company_id,omitempty"`
	Name          string     `json:"name" validate:"required"`
	Code          string     `json:"code"`
	PartnerID     *uuid.UUID `json:"partner_id,omitempty"`
	Active        bool       `json:"active"`
}

// AnalyticAccountUpdateRequest represents a request to update an analytic account
type AnalyticAccountUpdateRequest struct {
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Code      *string    `json:"code,omitempty"`
	PartnerID *uuid.UUID `json:"partner_id,omitempty"`
	Active    *bool      `json:"active,omitempty"`
}
