package types

import (
	"time"

	"github.com/google/uuid"
)

// Country represents a country (ISO 3166-1)
type Country struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Code        string     `json:"code" db:"code"`
	Name        string     `json:"name" db:"name"`
	PhoneCode   string     `json:"phone_code" db:"phone_code"`
	CurrencyID  *uuid.UUID `json:"currency_id,omitempty" db:"currency_id"`
	AddressFormat string    `json:"address_format" db:"address_format"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// CountryFilter for querying countries
type CountryFilter struct {
	Code       *string
	Name       *string
	Limit      int
	Offset     int
}

// CountryCreateRequest represents a request to create a country
type CountryCreateRequest struct {
	Code        string     `json:"code" validate:"required,len=2"`
	Name        string     `json:"name" validate:"required"`
	PhoneCode   string     `json:"phone_code"`
	CurrencyID  *uuid.UUID `json:"currency_id,omitempty"`
	AddressFormat string    `json:"address_format"`
}

// CountryUpdateRequest represents a request to update a country
type CountryUpdateRequest struct {
	Code        *string     `json:"code,omitempty"`
	Name        *string     `json:"name,omitempty"`
	PhoneCode   *string     `json:"phone_code,omitempty"`
	CurrencyID  *uuid.UUID  `json:"currency_id,omitempty"`
	AddressFormat *string    `json:"address_format,omitempty"`
}
