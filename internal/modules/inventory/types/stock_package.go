package types

import (
	"time"
	"github.com/google/uuid"
)

// StockPackage represents a stock package for grouping products
type StockPackage struct {
	ID            uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name          string    `json:"name" db:"name"`
	LocationID    *uuid.UUID `json:"location_id" db:"location_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// StockPackageCreateRequest represents the request to create a stock package
type StockPackageCreateRequest struct {
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	LocationID *uuid.UUID `json:"location_id"`
}

// StockPackageUpdateRequest represents the request to update a stock package
type StockPackageUpdateRequest struct {
	Name       *string    `json:"name"`
	LocationID *uuid.UUID `json:"location_id"`
}
