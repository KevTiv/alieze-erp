package types

import (
	"time"

	"github.com/google/uuid"
)

// UTMCampaign represents a UTM campaign for marketing tracking
type UTMCampaign struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name          string     `json:"name" db:"name"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// UTMMedium represents a UTM medium for marketing tracking
type UTMMedium struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name          string     `json:"name" db:"name"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// UTMSource represents a UTM source for marketing tracking
type UTMSource struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name          string     `json:"name" db:"name"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// UTMCampaignFilter for querying UTM campaigns
type UTMCampaignFilter struct {
	OrganizationID *uuid.UUID
	Name          *string
	Limit         int
	Offset        int
}

// UTMMediumFilter for querying UTM mediums
type UTMMediumFilter struct {
	OrganizationID *uuid.UUID
	Name          *string
	Limit         int
	Offset        int
}

// UTMSourceFilter for querying UTM sources
type UTMSourceFilter struct {
	OrganizationID *uuid.UUID
	Name          *string
	Limit         int
	Offset        int
}

// UTMCampaignCreateRequest represents a request to create a UTM campaign
type UTMCampaignCreateRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	Name          string    `json:"name" validate:"required"`
}

// UTMMediumCreateRequest represents a request to create a UTM medium
type UTMMediumCreateRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	Name          string    `json:"name" validate:"required"`
}

// UTMSourceCreateRequest represents a request to create a UTM source
type UTMSourceCreateRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	Name          string    `json:"name" validate:"required"`
}

// UTMCampaignUpdateRequest represents a request to update a UTM campaign
type UTMCampaignUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}

// UTMMediumUpdateRequest represents a request to update a UTM medium
type UTMMediumUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}

// UTMSourceUpdateRequest represents a request to update a UTM source
type UTMSourceUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}
