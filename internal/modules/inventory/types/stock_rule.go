package types

import (
	"time"
	"github.com/google/uuid"
)

// StockRule represents an inventory routing rule
type StockRule struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	Name            string    `json:"name" db:"name"`
	Action          string    `json:"action" db:"action"`
	LocationSrcID   *uuid.UUID `json:"location_src_id" db:"location_src_id"`
	LocationDestID  *uuid.UUID `json:"location_dest_id" db:"location_dest_id"`
	Sequence        int       `json:"sequence" db:"sequence"`
	Active          bool      `json:"active" db:"active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// StockRuleCreateRequest represents the request to create a stock rule
type StockRuleCreateRequest struct {
	Name           string    `json:"name" validate:"required,min=1,max=255"`
	Action         string    `json:"action" validate:"required,min=1,max=50"`
	LocationSrcID  *uuid.UUID `json:"location_src_id"`
	LocationDestID *uuid.UUID `json:"location_dest_id"`
	Sequence       int       `json:"sequence"`
	Active         bool      `json:"active"`
}

// StockRuleUpdateRequest represents the request to update a stock rule
type StockRuleUpdateRequest struct {
	Name           *string    `json:"name"`
	Action         *string    `json:"action"`
	LocationSrcID  *uuid.UUID `json:"location_src_id"`
	LocationDestID *uuid.UUID `json:"location_dest_id"`
	Sequence       *int       `json:"sequence"`
	Active         *bool      `json:"active"`
}
