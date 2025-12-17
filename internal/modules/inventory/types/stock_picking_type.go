package types

import (
	"time"
	"github.com/google/uuid"
)

// StockPickingType represents a type of inventory transfer operation
type StockPickingType struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	OrganizationID         uuid.UUID `json:"organization_id" db:"organization_id"`
	Name                   string    `json:"name" db:"name"`
	Code                   *string   `json:"code" db:"code"`
	Sequence               int       `json:"sequence" db:"sequence"`
	SequenceID             *uuid.UUID `json:"sequence_id" db:"sequence_id"`
	DefaultLocationSrcID   *uuid.UUID `json:"default_location_src_id" db:"default_location_src_id"`
	DefaultLocationDestID  *uuid.UUID `json:"default_location_dest_id" db:"default_location_dest_id"`
	WarehouseID            *uuid.UUID `json:"warehouse_id" db:"warehouse_id"`
	Color                  *int       `json:"color" db:"color"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// StockPickingTypeCreateRequest represents the request to create a stock picking type
type StockPickingTypeCreateRequest struct {
	Name                   string    `json:"name" validate:"required,min=1,max=255"`
	Code                   *string   `json:"code"`
	Sequence               int       `json:"sequence"`
	SequenceID             *uuid.UUID `json:"sequence_id"`
	DefaultLocationSrcID   *uuid.UUID `json:"default_location_src_id"`
	DefaultLocationDestID  *uuid.UUID `json:"default_location_dest_id"`
	WarehouseID            *uuid.UUID `json:"warehouse_id"`
	Color                  *int       `json:"color"`
}

// StockPickingTypeUpdateRequest represents the request to update a stock picking type
type StockPickingTypeUpdateRequest struct {
	Name                   *string    `json:"name"`
	Code                   *string    `json:"code"`
	Sequence               *int       `json:"sequence"`
	SequenceID             *uuid.UUID `json:"sequence_id"`
	DefaultLocationSrcID   *uuid.UUID `json:"default_location_src_id"`
	DefaultLocationDestID  *uuid.UUID `json:"default_location_dest_id"`
	WarehouseID            *uuid.UUID `json:"warehouse_id"`
	Color                  *int       `json:"color"`
}
