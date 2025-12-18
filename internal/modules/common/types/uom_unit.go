package types

import (
	"time"

	"github.com/google/uuid"
)

// UOMType represents the type of unit of measure
type UOMType string

const (
	UOMTypeReference UOMType = "reference"
	UOMTypeBigger    UOMType = "bigger"
	UOMTypeSmaller   UOMType = "smaller"
)

// UOMUnit represents a unit of measure
type UOMUnit struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	CategoryID uuid.UUID  `json:"category_id" db:"category_id"`
	Name       string     `json:"name" db:"name"`
	UOMType    UOMType    `json:"uom_type" db:"uom_type"`
	Factor     float64    `json:"factor" db:"factor"`
	FactorInv  float64    `json:"factor_inv" db:"factor_inv"`
	Rounding   float64    `json:"rounding" db:"rounding"`
	Active     bool       `json:"active" db:"active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// UOMUnitFilter for querying UOM units
type UOMUnitFilter struct {
	CategoryID *uuid.UUID
	Active     *bool
	Name       *string
	Limit      int
	Offset     int
}

// UOMUnitCreateRequest represents a request to create a UOM unit
type UOMUnitCreateRequest struct {
	CategoryID uuid.UUID  `json:"category_id" validate:"required"`
	Name       string     `json:"name" validate:"required"`
	UOMType    UOMType    `json:"uom_type" validate:"required,oneof=reference bigger smaller"`
	Factor     float64    `json:"factor" validate:"required"`
	FactorInv  float64    `json:"factor_inv" validate:"required"`
	Rounding   float64    `json:"rounding" validate:"required"`
	Active     bool       `json:"active"`
}

// UOMUnitUpdateRequest represents a request to update a UOM unit
type UOMUnitUpdateRequest struct {
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	Name       *string    `json:"name,omitempty"`
	UOMType    *UOMType   `json:"uom_type,omitempty"`
	Factor     *float64   `json:"factor,omitempty"`
	FactorInv  *float64   `json:"factor_inv,omitempty"`
	Rounding   *float64   `json:"rounding,omitempty"`
	Active     *bool      `json:"active,omitempty"`
}
