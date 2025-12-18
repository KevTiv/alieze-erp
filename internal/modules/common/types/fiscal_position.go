package types

import (
	"time"

	"github.com/google/uuid"
)

// FiscalPosition represents fiscal position (tax mapping rules)
type FiscalPosition struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID     *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name          string     `json:"name" db:"name"`
	AutoApply     bool       `json:"auto_apply" db:"auto_apply"`
	VATRequired   bool       `json:"vat_required" db:"vat_required"`
	CountryID     *uuid.UUID `json:"country_id,omitempty" db:"country_id"`
	StateIDs      []uuid.UUID `json:"state_ids" db:"state_ids"`
	ZipFrom       *int       `json:"zip_from,omitempty" db:"zip_from"`
	ZipTo         *int       `json:"zip_to,omitempty" db:"zip_to"`
	Note          string     `json:"note" db:"note"`
	Active        bool       `json:"active" db:"active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// FiscalPositionFilter for querying fiscal positions
type FiscalPositionFilter struct {
	OrganizationID *uuid.UUID
	CompanyID     *uuid.UUID
	Active        *bool
	Name          *string
	CountryID     *uuid.UUID
	Limit         int
	Offset        int
}

// FiscalPositionCreateRequest represents a request to create a fiscal position
type FiscalPositionCreateRequest struct {
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	CompanyID     *uuid.UUID `json:"company_id,omitempty"`
	Name          string     `json:"name" validate:"required"`
	AutoApply     bool       `json:"auto_apply"`
	VATRequired   bool       `json:"vat_required"`
	CountryID     *uuid.UUID `json:"country_id,omitempty"`
	StateIDs      []uuid.UUID `json:"state_ids"`
	ZipFrom       *int       `json:"zip_from,omitempty"`
	ZipTo         *int       `json:"zip_to,omitempty"`
	Note          string     `json:"note"`
	Active        bool       `json:"active"`
}

// FiscalPositionUpdateRequest represents a request to update a fiscal position
type FiscalPositionUpdateRequest struct {
	CompanyID   *uuid.UUID  `json:"company_id,omitempty"`
	Name        *string     `json:"name,omitempty"`
	AutoApply   *bool       `json:"auto_apply,omitempty"`
	VATRequired *bool       `json:"vat_required,omitempty"`
	CountryID   *uuid.UUID  `json:"country_id,omitempty"`
	StateIDs    *[]uuid.UUID `json:"state_ids,omitempty"`
	ZipFrom     *int        `json:"zip_from,omitempty"`
	ZipTo       *int        `json:"zip_to,omitempty"`
	Note        *string     `json:"note,omitempty"`
	Active      *bool       `json:"active,omitempty"`
}
