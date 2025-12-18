package types

import (
	"time"

	"github.com/google/uuid"
)

// ContactTag represents a tag for categorizing contacts
type ContactTag struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Color          int        `json:"color" db:"color"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ContactTagFilter represents filtering criteria for contact tags
type ContactTagFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Limit          int
	Offset         int
}
