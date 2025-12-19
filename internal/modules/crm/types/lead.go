package types

import (
	"time"

	"github.com/google/uuid"
)

// LeadType represents the type of lead
type LeadType string

const (
	LeadTypeLead        LeadType = "lead"
	LeadTypeOpportunity LeadType = "opportunity"
)

// LeadPriority represents the priority of a lead
type LeadPriority string

const (
	LeadPriorityLow    LeadPriority = "low"
	LeadPriorityMedium LeadPriority = "medium"
	LeadPriorityHigh   LeadPriority = "high"
	LeadPriorityUrgent LeadPriority = "urgent"
)

// LeadWonStatus represents the won status of a lead
type LeadWonStatus string

const (
	LeadWonStatusWon     LeadWonStatus = "won"
	LeadWonStatusLost    LeadWonStatus = "lost"
	LeadWonStatusOngoing LeadWonStatus = "ongoing"
)

// Lead represents a comprehensive sales lead with all database fields
type Lead struct {
	ID                  uuid.UUID      `json:"id" db:"id"`
	OrganizationID      uuid.UUID      `json:"organization_id" db:"organization_id"`
	CompanyID           *uuid.UUID     `json:"company_id,omitempty" db:"company_id"`
	Name                string         `json:"name" db:"name"`
	ContactName         *string        `json:"contact_name,omitempty" db:"contact_name"`
	Email               *string        `json:"email,omitempty" db:"email"`
	Phone               *string        `json:"phone,omitempty" db:"phone"`
	Mobile              *string        `json:"mobile,omitempty" db:"mobile"`
	ContactID           *uuid.UUID     `json:"contact_id,omitempty" db:"contact_id"`
	UserID              *uuid.UUID     `json:"user_id,omitempty" db:"user_id"`
	TeamID              *uuid.UUID     `json:"team_id,omitempty" db:"team_id"`
	LeadType            LeadType       `json:"lead_type" db:"lead_type"`
	StageID             *uuid.UUID     `json:"stage_id,omitempty" db:"stage_id"`
	Priority            LeadPriority   `json:"priority" db:"priority"`
	SourceID            *uuid.UUID     `json:"source_id,omitempty" db:"source_id"`
	MediumID            *uuid.UUID     `json:"medium_id,omitempty" db:"medium_id"`
	CampaignID          *uuid.UUID     `json:"campaign_id,omitempty" db:"campaign_id"`
	ExpectedRevenue     *float64       `json:"expected_revenue,omitempty" db:"expected_revenue"`
	Probability         int            `json:"probability" db:"probability"`
	RecurringRevenue    *float64       `json:"recurring_revenue,omitempty" db:"recurring_revenue"`
	RecurringPlan       *string        `json:"recurring_plan,omitempty" db:"recurring_plan"`
	DateOpen            *time.Time     `json:"date_open,omitempty" db:"date_open"`
	DateClosed          *time.Time     `json:"date_closed,omitempty" db:"date_closed"`
	DateDeadline        *time.Time     `json:"date_deadline,omitempty" db:"date_deadline"`
	DateLastStageUpdate *time.Time     `json:"date_last_stage_update,omitempty" db:"date_last_stage_update"`
	Active              bool           `json:"active" db:"active"`
	Status              *string        `json:"status,omitempty" db:"status"`
	AssignedTo          *uuid.UUID     `json:"assigned_to,omitempty" db:"assigned_to"`
	WonStatus           *LeadWonStatus `json:"won_status,omitempty" db:"won_status"`
	LostReasonID        *uuid.UUID     `json:"lost_reason_id,omitempty" db:"lost_reason_id"`
	Street              *string        `json:"street,omitempty" db:"street"`
	Street2             *string        `json:"street2,omitempty" db:"street2"`
	City                *string        `json:"city,omitempty" db:"city"`
	StateID             *uuid.UUID     `json:"state_id,omitempty" db:"state_id"`
	Zip                 *string        `json:"zip,omitempty" db:"zip"`
	CountryID           *uuid.UUID     `json:"country_id,omitempty" db:"country_id"`
	Website             *string        `json:"website,omitempty" db:"website"`
	Description         *string        `json:"description,omitempty" db:"description"`
	TagIDs              []uuid.UUID    `json:"tag_ids" db:"tag_ids"`
	Color               *int           `json:"color,omitempty" db:"color"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy           *uuid.UUID     `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy           *uuid.UUID     `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt           *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
	CustomFields        interface{}    `json:"custom_fields,omitempty" db:"custom_fields"`
	Metadata            interface{}    `json:"metadata,omitempty" db:"metadata"`
}

// LeadFilter represents filtering criteria for enhanced leads
type LeadFilter struct {
	OrganizationID     uuid.UUID
	CompanyID          *uuid.UUID
	Name               *string
	ContactName        *string
	Email              *string
	Phone              *string
	Mobile             *string
	ContactID          *uuid.UUID
	UserID             *uuid.UUID
	TeamID             *uuid.UUID
	LeadType           *LeadType
	StageID            *uuid.UUID
	Priority           *LeadPriority
	SourceID           *uuid.UUID
	MediumID           *uuid.UUID
	CampaignID         *uuid.UUID
	ExpectedRevenueMin *float64
	ExpectedRevenueMax *float64
	ProbabilityMin     *int
	ProbabilityMax     *int
	WonStatus          *LeadWonStatus
	LostReasonID       *uuid.UUID
	Active             *bool
	Status             *string
	AssignedTo         *uuid.UUID
	DateOpenFrom       *time.Time
	DateOpenTo         *time.Time
	DateDeadlineFrom   *time.Time
	DateDeadlineTo     *time.Time
	CountryID          *uuid.UUID
	StateID            *uuid.UUID
	City               *string
	CreatedBy          *uuid.UUID
	UpdatedBy          *uuid.UUID
	Color              *string
	Limit              int
	Offset             int
}
