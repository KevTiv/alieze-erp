package domain

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the catalog
type Product struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	DefaultCode    *string    `json:"default_code,omitempty" db:"default_code"`
	Barcode        *string    `json:"barcode,omitempty" db:"barcode"`
	ProductType    string     `json:"product_type" db:"product_type"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	ListPrice      *float64   `json:"list_price,omitempty" db:"list_price"`
	Active         bool       `json:"active" db:"active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ProductFilter represents filtering criteria for products
type ProductFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	DefaultCode    *string
	Barcode        *string
	ProductType    *string
	CategoryID     *uuid.UUID
	Active         *bool
	Limit          int
	Offset         int
}
