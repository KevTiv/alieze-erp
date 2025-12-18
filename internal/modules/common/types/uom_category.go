package types

import (
	"time"

	"github.com/google/uuid"
)

// UOMCategory represents a unit of measure category
type UOMCategory struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// UOMCategoryFilter for querying UOM categories
type UOMCategoryFilter struct {
	Name   *string
	Limit  int
	Offset int
}

// UOMCategoryCreateRequest represents a request to create a UOM category
type UOMCategoryCreateRequest struct {
	Name string `json:"name" validate:"required"`
}

// UOMCategoryUpdateRequest represents a request to update a UOM category
type UOMCategoryUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}
