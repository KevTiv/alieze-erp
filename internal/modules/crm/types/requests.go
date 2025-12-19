package types

import (
	"time"

	"github.com/google/uuid"
)

// Request types for Contact operations
type ContactCreateRequest struct {
	Name       string     `json:"name"`
	Email      *string    `json:"email,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	IsCustomer bool       `json:"is_customer"`
	IsVendor   bool       `json:"is_vendor"`
	Street     *string    `json:"street,omitempty"`
	City       *string    `json:"city,omitempty"`
	StateID    *uuid.UUID `json:"state_id,omitempty"`
	CountryID  *uuid.UUID `json:"country_id,omitempty"`
}

type ContactUpdateRequest struct {
	Name       *string    `json:"name,omitempty"`
	Email      *string    `json:"email,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	IsCustomer *bool      `json:"is_customer,omitempty"`
	IsVendor   *bool      `json:"is_vendor,omitempty"`
	Street     *string    `json:"street,omitempty"`
	City       *string    `json:"city,omitempty"`
	StateID    *uuid.UUID `json:"state_id,omitempty"`
	CountryID  *uuid.UUID `json:"country_id,omitempty"`
}

type ContactRelationshipCreateRequest struct {
	RelatedContactID uuid.UUID               `json:"related_contact_id"`
	Type             ContactRelationshipType `json:"type"`
	Notes            *string                 `json:"notes,omitempty"`
}

type ContactSegmentationRequest struct {
	SegmentIDs []string `json:"segment_ids,omitempty"`
	CustomTags []string `json:"custom_tags,omitempty"`
}

// LeadCreateRequest represents a request to create a lead
// This is the consolidated request type that replaces LeadEnhancedCreateRequest
type LeadCreateRequest struct {
	CompanyID        *uuid.UUID     `json:"company_id,omitempty"`
	Name             string         `json:"name"`
	ContactName      *string        `json:"contact_name,omitempty"`
	Email            *string        `json:"email,omitempty"`
	Phone            *string        `json:"phone,omitempty"`
	Mobile           *string        `json:"mobile,omitempty"`
	ContactID        *uuid.UUID     `json:"contact_id,omitempty"`
	UserID           *uuid.UUID     `json:"user_id,omitempty"`
	TeamID           *uuid.UUID     `json:"team_id,omitempty"`
	LeadType         LeadType       `json:"lead_type"`
	StageID          *uuid.UUID     `json:"stage_id,omitempty"`
	Priority         LeadPriority   `json:"priority"`
	SourceID         *uuid.UUID     `json:"source_id,omitempty"`
	MediumID         *uuid.UUID     `json:"medium_id,omitempty"`
	CampaignID       *uuid.UUID     `json:"campaign_id,omitempty"`
	ExpectedRevenue  *float64       `json:"expected_revenue,omitempty"`
	Probability      int            `json:"probability"`
	RecurringRevenue *float64       `json:"recurring_revenue,omitempty"`
	RecurringPlan    *string        `json:"recurring_plan,omitempty"`
	DateOpen         *time.Time     `json:"date_open,omitempty"`
	DateClosed       *time.Time     `json:"date_closed,omitempty"`
	DateDeadline     *time.Time     `json:"date_deadline,omitempty"`
	Active           bool           `json:"active"`
	Status           *string        `json:"status,omitempty"`
	AssignedTo       *uuid.UUID     `json:"assigned_to,omitempty"`
	WonStatus        *LeadWonStatus `json:"won_status,omitempty"`
	LostReasonID     *uuid.UUID     `json:"lost_reason_id,omitempty"`
	Street           *string        `json:"street,omitempty"`
	Street2          *string        `json:"street2,omitempty"`
	City             *string        `json:"city,omitempty"`
	StateID          *uuid.UUID     `json:"state_id,omitempty"`
	Zip              *string        `json:"zip,omitempty"`
	CountryID        *uuid.UUID     `json:"country_id,omitempty"`
	Website          *string        `json:"website,omitempty"`
	Description      *string        `json:"description,omitempty"`
	TagIDs           []uuid.UUID    `json:"tag_ids"`
	Color            *int           `json:"color,omitempty"`
	CustomFields     interface{}    `json:"custom_fields,omitempty"`
	Metadata         interface{}    `json:"metadata,omitempty"`
}

// LeadUpdateRequest represents a request to update a lead
// This is the consolidated request type that replaces LeadEnhancedUpdateRequest
type LeadUpdateRequest struct {
	CompanyID        *uuid.UUID     `json:"company_id,omitempty"`
	Name             *string        `json:"name,omitempty"`
	ContactName      *string        `json:"contact_name,omitempty"`
	Email            *string        `json:"email,omitempty"`
	Phone            *string        `json:"phone,omitempty"`
	Mobile           *string        `json:"mobile,omitempty"`
	ContactID        *uuid.UUID     `json:"contact_id,omitempty"`
	UserID           *uuid.UUID     `json:"user_id,omitempty"`
	TeamID           *uuid.UUID     `json:"team_id,omitempty"`
	LeadType         *LeadType      `json:"lead_type,omitempty"`
	StageID          *uuid.UUID     `json:"stage_id,omitempty"`
	Priority         *LeadPriority  `json:"priority,omitempty"`
	SourceID         *uuid.UUID     `json:"source_id,omitempty"`
	MediumID         *uuid.UUID     `json:"medium_id,omitempty"`
	CampaignID       *uuid.UUID     `json:"campaign_id,omitempty"`
	ExpectedRevenue  *float64       `json:"expected_revenue,omitempty"`
	Probability      *int           `json:"probability,omitempty"`
	RecurringRevenue *float64       `json:"recurring_revenue,omitempty"`
	RecurringPlan    *string        `json:"recurring_plan,omitempty"`
	DateOpen         *time.Time     `json:"date_open,omitempty"`
	DateClosed       *time.Time     `json:"date_closed,omitempty"`
	DateDeadline     *time.Time     `json:"date_deadline,omitempty"`
	Active           *bool          `json:"active,omitempty"`
	Status           *string        `json:"status,omitempty"`
	AssignedTo       *uuid.UUID     `json:"assigned_to,omitempty"`
	WonStatus        *LeadWonStatus `json:"won_status,omitempty"`
	LostReasonID     *uuid.UUID     `json:"lost_reason_id,omitempty"`
	Street           *string        `json:"street,omitempty"`
	Street2          *string        `json:"street2,omitempty"`
	City             *string        `json:"city,omitempty"`
	StateID          *uuid.UUID     `json:"state_id,omitempty"`
	Zip              *string        `json:"zip,omitempty"`
	CountryID        *uuid.UUID     `json:"country_id,omitempty"`
	Website          *string        `json:"website,omitempty"`
	Description      *string        `json:"description,omitempty"`
	TagIDs           *[]uuid.UUID   `json:"tag_ids,omitempty"`
	Color            *int           `json:"color,omitempty"`
	CustomFields     interface{}    `json:"custom_fields,omitempty"`
	Metadata         interface{}    `json:"metadata,omitempty"`
}
