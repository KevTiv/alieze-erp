package types

import (
	"time"

	"github.com/google/uuid"
)

// CurrencyPosition represents the position of currency symbol
type CurrencyPosition string

const (
	CurrencyPositionBefore CurrencyPosition = "before"
	CurrencyPositionAfter  CurrencyPosition = "after"
)

// Currency represents a currency (ISO 4217)
type Currency struct {
	ID            uuid.UUID         `json:"id" db:"id"`
	Name          string            `json:"name" db:"name"`
	Symbol        string            `json:"symbol" db:"symbol"`
	Code          string            `json:"code" db:"code"`
	Rounding      float64           `json:"rounding" db:"rounding"`
	DecimalPlaces int               `json:"decimal_places" db:"decimal_places"`
	Position      CurrencyPosition  `json:"position" db:"position"`
	Active        bool              `json:"active" db:"active"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
}

// CurrencyFilter for querying currencies
type CurrencyFilter struct {
	Active     *bool
	Code       *string
	Name       *string
	Limit      int
	Offset     int
}

// CurrencyCreateRequest represents a request to create a currency
type CurrencyCreateRequest struct {
	Name          string           `json:"name" validate:"required"`
	Symbol        string           `json:"symbol" validate:"required"`
	Code          string           `json:"code" validate:"required,len=3"`
	Rounding      float64          `json:"rounding" validate:"required"`
	DecimalPlaces int              `json:"decimal_places" validate:"required,min=0,max=6"`
	Position      CurrencyPosition `json:"position" validate:"required,oneof=before after"`
	Active        bool             `json:"active"`
}

// CurrencyUpdateRequest represents a request to update a currency
type CurrencyUpdateRequest struct {
	Name          *string           `json:"name,omitempty"`
	Symbol        *string           `json:"symbol,omitempty"`
	Code          *string           `json:"code,omitempty"`
	Rounding      *float64          `json:"rounding,omitempty"`
	DecimalPlaces *int              `json:"decimal_places,omitempty"`
	Position      *CurrencyPosition `json:"position,omitempty"`
	Active        *bool             `json:"active,omitempty"`
}
