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
	ContactRelationshipTypeOther     ContactRelationshipType = "other"
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
	ID               uuid.UUID               `json:"id" db:"id"`
	OrganizationID   uuid.UUID               `json:"organization_id" db:"organization_id"`
	ContactID        uuid.UUID               `json:"contact_id" db:"contact_id"`
	RelatedContactID uuid.UUID               `json:"related_contact_id" db:"related_contact_id"`
	Type             ContactRelationshipType `json:"type" db:"type"`
	Notes            *string                 `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at" db:"updated_at"`
}

// ContactScore represents engagement and lead scores for a contact
type ContactScore struct {
	EngagementScore   int                    `json:"engagement_score"`
	LeadScore         int                    `json:"lead_score"`
	EngagementFactors map[string]interface{} `json:"engagement_factors"`
	LeadFactors       map[string]interface{} `json:"lead_factors"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// ContactTag represents a tag that can be applied to contacts
type ContactTag struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Color          int       `json:"color" db:"color"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ContactTagFilter represents filtering criteria for contact tags
type ContactTagFilter struct {
	OrganizationID uuid.UUID
	Name           *string
	Limit          int
	Offset         int
}

// AdvancedContactFilter represents advanced filtering criteria for contacts
type AdvancedContactFilter struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	SearchQuery    string    `json:"search_query,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	Segments       []string  `json:"segments,omitempty"`
	ScoreRange     struct {
		Min int `json:"min,omitempty"`
		Max int `json:"max,omitempty"`
	} `json:"score_range,omitempty"`
	LastContacted struct {
		From time.Time `json:"from,omitempty"`
		To   time.Time `json:"to,omitempty"`
	} `json:"last_contacted,omitempty"`
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// CRMDashboard represents a comprehensive CRM dashboard
type CRMDashboard struct {
	TimeRange        string           `json:"time_range"`
	Summary          DashboardSummary `json:"summary"`
	Trends           DashboardTrends  `json:"trends"`
	TopContacts      []TopContact     `json:"top_contacts"`
	RecentActivities []RecentActivity `json:"recent_activities"`
}

// DashboardSummary represents summary statistics for the dashboard
type DashboardSummary struct {
	TotalContacts     int `json:"total_contacts"`
	NewContacts       int `json:"new_contacts"`
	ActiveContacts    int `json:"active_contacts"`
	AtRiskContacts    int `json:"at_risk_contacts"`
	HighValueContacts int `json:"high_value_contacts"`
}

// DashboardTrends represents trend data for the dashboard
type DashboardTrends struct {
	ContactGrowth []TrendDataPoint `json:"contact_growth"`
	Engagement    []TrendDataPoint `json:"engagement"`
	ResponseRate  []TrendDataPoint `json:"response_rate"`
}

// TrendDataPoint represents a single data point in a trend
type TrendDataPoint struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

// TopContact represents a top contact for the dashboard
type TopContact struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Company      string    `json:"company,omitempty"`
	Score        int       `json:"score"`
	LastActivity string    `json:"last_activity"`
	NextAction   string    `json:"next_action"`
}

// RecentActivity represents a recent activity for the dashboard
type RecentActivity struct {
	ID          uuid.UUID `json:"id"`
	ContactID   uuid.UUID `json:"contact_id"`
	ContactName string    `json:"contact_name"`
	Type        string    `json:"type"` // call, email, meeting, etc.
	Subject     string    `json:"subject"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"`
}

// ActivityDashboard represents an activity-focused dashboard
type ActivityDashboard struct {
	TimeRange         string              `json:"time_range"`
	ContactType       string              `json:"contact_type"`
	ActivitySummary   ActivitySummary     `json:"activity_summary"`
	RecentActivities  []RecentActivity    `json:"recent_activities"`
	ContactEngagement []ContactEngagement `json:"contact_engagement"`
}

// ActivitySummary represents activity summary statistics
type ActivitySummary struct {
	TotalActivities int             `json:"total_activities"`
	ActivityTypes   map[string]int  `json:"activity_types"`
	ActivityTrends  []ActivityTrend `json:"activity_trends"`
}

// ActivityTrend represents activity trends by date
type ActivityTrend struct {
	Date     string `json:"date"`
	Calls    int    `json:"calls"`
	Emails   int    `json:"emails"`
	Meetings int    `json:"meetings"`
	Other    int    `json:"other"`
}

// ContactEngagement represents contact engagement metrics
type ContactEngagement struct {
	ContactID            uuid.UUID `json:"contact_id"`
	ContactName          string    `json:"contact_name"`
	EngagementScore      int       `json:"engagement_score"`
	LastActivity         string    `json:"last_activity"`
	DaysSinceLastContact int       `json:"days_since_last_contact"`
	RecommendedAction    string    `json:"recommended_action"`
}

// DashboardCache represents cached dashboard data
type DashboardCache struct {
	OrganizationID uuid.UUID   `json:"organization_id"`
	TimeRange      string      `json:"time_range"`
	ContactType    string      `json:"contact_type,omitempty"`
	Data           interface{} `json:"data"` // Can be CRMDashboard or ActivityDashboard
	CachedAt       time.Time   `json:"cached_at"`
	ExpiresAt      time.Time   `json:"expires_at"`
}
