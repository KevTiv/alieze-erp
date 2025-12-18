package types

import (
	"time"

	"github.com/google/uuid"
)

// Contact represents a CRM contact
type Contact struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Email          *string    `json:"email,omitempty" db:"email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	IsCustomer     bool       `json:"is_customer" db:"is_customer"`
	IsVendor       bool       `json:"is_vendor" db:"is_vendor"`
	Street         *string    `json:"street,omitempty" db:"street"`
	City           *string    `json:"city,omitempty" db:"city"`
	StateID        *uuid.UUID `json:"state_id,omitempty" db:"state_id"`
	CountryID      *uuid.UUID `json:"country_id,omitempty" db:"country_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ContactFilter represents filtering criteria for contacts
type ContactFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Email          *string
	Phone          *string
	IsCustomer     *bool
	IsVendor       *bool
	Limit          int
	Offset         int
}

// ContactRelationshipType represents the type of relationship between contacts
type ContactRelationshipType string

const (
	ContactRelationshipTypeColleague ContactRelationshipType = "colleague"
	ContactRelationshipTypeManager   ContactRelationshipType = "manager"
	ContactRelationshipTypeFamily    ContactRelationshipType = "family"
	ContactRelationshipTypePartner   ContactRelationshipType = "partner"
	ContactRelationshipTypeReferral  ContactRelationshipType = "referral"
	ContactRelationshipTypeOther    ContactRelationshipType = "other"
)

func IsValidRelationshipType(relType ContactRelationshipType) bool {
	switch relType {
	case ContactRelationshipTypeColleague, ContactRelationshipTypeManager,
		ContactRelationshipTypeFamily, ContactRelationshipTypePartner,
		ContactRelationshipTypeReferral, ContactRelationshipTypeOther:
		return true
	default:
		return false
	}
}

// ContactRelationship represents a relationship between two contacts
type ContactRelationship struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	OrganizationID  uuid.UUID              `json:"organization_id" db:"organization_id"`
	ContactID       uuid.UUID              `json:"contact_id" db:"contact_id"`
	RelatedContactID uuid.UUID              `json:"related_contact_id" db:"related_contact_id"`
	Type            ContactRelationshipType `json:"type" db:"type"`
	Notes           *string                `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// ContactRelationshipCreateRequest represents a request to create a contact relationship
type ContactRelationshipCreateRequest struct {
	RelatedContactID uuid.UUID              `json:"related_contact_id"`
	Type            ContactRelationshipType `json:"type"`
	Notes           *string                `json:"notes,omitempty"`
}

// ContactSegmentationRequest represents a request to add a contact to segments/tags
type ContactSegmentationRequest struct {
	SegmentIDs  []string `json:"segment_ids,omitempty"`
	CustomTags  []string `json:"custom_tags,omitempty"`
}

// ContactScore represents engagement and lead scores for a contact
type ContactScore struct {
	EngagementScore   int                     `json:"engagement_score"`
	LeadScore         int                     `json:"lead_score"`
	EngagementFactors map[string]interface{}  `json:"engagement_factors"`
	LeadFactors       map[string]interface{}  `json:"lead_factors"`
	LastUpdated       time.Time               `json:"last_updated"`
}
