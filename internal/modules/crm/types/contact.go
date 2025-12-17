package types

import (
	"time"

	"github.com/google/uuid"
)

// Contact represents a CRM contact
type Contact struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Email          *string    `json:"email,omitempty" db:"email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	IsCustomer     bool       `json:"is_customer" db:"is_customer"`
	IsVendor       bool       `json:"is_vendor" db:"is_vendor"`
	Street         *string    `json:"street,omitempty" db:"street"`
	City           *string    `json:"city,omitempty" db:"city"`
	StateID        *uuid.UUID `json:"state_id,omitempty" db:"state_id"`
	CountryID      *uuid.UUID `json:"country_id,omitempty" db:"country_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ContactFilter represents filtering criteria for contacts
type ContactFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Email          *string
	Phone          *string
	IsCustomer     *bool
	IsVendor       *bool
	Limit          int
	Offset         int
}
