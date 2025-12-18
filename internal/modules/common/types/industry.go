package types

import (
	"time"

	"github.com/google/uuid"
)

// Industry represents an industry classification
type Industry struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Code      string     `json:"code" db:"code"`
	FullName  string     `json:"full_name" db:"full_name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// IndustryFilter for querying industries
type IndustryFilter struct {
	Name   *string
	Code   *string
	Limit  int
	Offset int
}

// IndustryCreateRequest represents a request to create an industry
type IndustryCreateRequest struct {
	Name     string `json:"name" validate:"required"`
	Code     string `json:"code"`
	FullName string `json:"full_name"`
}

// IndustryUpdateRequest represents a request to update an industry
type IndustryUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Code     *string `json:"code,omitempty"`
	FullName *string `json:"full_name,omitempty"`
}
