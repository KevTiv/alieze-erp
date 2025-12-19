package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// AssignmentRuleType represents the type of assignment rule
type AssignmentRuleType string

const (
	AssignmentRuleTypeRoundRobin AssignmentRuleType = "round_robin"
	AssignmentRuleTypeWeighted   AssignmentRuleType = "weighted"
	AssignmentRuleTypeTerritory  AssignmentRuleType = "territory"
	AssignmentRuleTypeCustom     AssignmentRuleType = "custom"
)

// AssignmentTargetModel represents the target entity type
type AssignmentTargetModel string

const (
	AssignmentTargetModelLeads         AssignmentTargetModel = "leads"
	AssignmentTargetModelContacts      AssignmentTargetModel = "contacts"
	AssignmentTargetModelOpportunities AssignmentTargetModel = "opportunities"
)

// AssignmentRule represents an assignment rule configuration
type AssignmentRule struct {
	ID                    uuid.UUID             `json:"id" db:"id"`
	OrganizationID        uuid.UUID             `json:"organization_id" db:"organization_id"`
	Name                  string                `json:"name" db:"name"`
	Description           string                `json:"description" db:"description"`
	RuleType              AssignmentRuleType    `json:"rule_type" db:"rule_type"`
	TargetModel           AssignmentTargetModel `json:"target_model" db:"target_model"`
	Priority              int                   `json:"priority" db:"priority"`
	IsActive              bool                  `json:"is_active" db:"is_active"`
	Conditions            AssignmentConditions  `json:"conditions" db:"conditions"`
	AssignmentConfig      AssignmentConfig      `json:"assignment_config" db:"assignment_config"`
	AssignToType          string                `json:"assign_to_type" db:"assign_to_type"`
	MaxAssignmentsPerUser int                   `json:"max_assignments_per_user" db:"max_assignments_per_user"`
	AssignmentWindowStart *time.Time            `json:"assignment_window_start" db:"assignment_window_start"`
	AssignmentWindowEnd   *time.Time            `json:"assignment_window_end" db:"assignment_window_end"`
	ActiveDays            []int                 `json:"active_days" db:"active_days"`
	CreatedAt             time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at" db:"updated_at"`
	AssignedTo            uuid.UUID             `json:"assigned_to" db:"assigned_to"`
	CreatedBy             uuid.UUID             `json:"created_by" db:"created_by"`
	UpdatedBy             uuid.UUID             `json:"updated_by" db:"updated_by"`
}

// AssignmentCondition represents a single condition for rule matching
type AssignmentCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// AssignmentConditions is a slice of conditions
type AssignmentConditions []AssignmentCondition

// Value implements the driver.Valuer interface
func (ac AssignmentConditions) Value() (driver.Value, error) {
	return json.Marshal(ac)
}

// Scan implements the sql.Scanner interface
func (ac *AssignmentConditions) Scan(value interface{}) error {
	if value == nil {
		*ac = make(AssignmentConditions, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ac)
}

// AssignmentConfig represents the configuration for different rule types
type AssignmentConfig struct {
	// Round Robin
	Users        []uuid.UUID `json:"users,omitempty"`
	CurrentIndex int         `json:"current_index,omitempty"`

	// Weighted
	Assignments []WeightedAssignment `json:"assignments,omitempty"`

	// Territory
	Territories []TerritoryAssignment `json:"territories,omitempty"`

	// Custom
	Logic  string                 `json:"logic,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Value implements the driver.Valuer interface
func (ac AssignmentConfig) Value() (driver.Value, error) {
	return json.Marshal(ac)
}

// Scan implements the sql.Scanner interface
func (ac *AssignmentConfig) Scan(value interface{}) error {
	if value == nil {
		*ac = AssignmentConfig{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ac)
}

// WeightedAssignment represents a user with weight for weighted assignment
type WeightedAssignment struct {
	UserID uuid.UUID `json:"user_id"`
	Weight int       `json:"weight"`
}

// TerritoryAssignment represents a territory with users and conditions
type TerritoryAssignment struct {
	Name       string                 `json:"name"`
	Users      []uuid.UUID            `json:"users"`
	Conditions map[string]interface{} `json:"conditions"`
}

// AssignmentHistory represents an assignment history record
type AssignmentHistory struct {
	ID                     uuid.UUID   `json:"id" db:"id"`
	OrganizationID         uuid.UUID   `json:"organization_id" db:"organization_id"`
	RuleID                 uuid.UUID   `json:"rule_id" db:"rule_id"`
	RuleName               string      `json:"rule_name" db:"rule_name"`
	TargetModel            string      `json:"target_model" db:"target_model"`
	TargetID               uuid.UUID   `json:"target_id" db:"target_id"`
	TargetName             string      `json:"target_name" db:"target_name"`
	AssignedToType         string      `json:"assigned_to_type" db:"assigned_to_type"`
	AssignedToID           uuid.UUID   `json:"assigned_to_id" db:"assigned_to_id"`
	AssignedToName         string      `json:"assigned_to_name" db:"assigned_to_name"`
	PreviousAssignedToID   uuid.UUID   `json:"previous_assigned_to_id" db:"previous_assigned_to_id"`
	PreviousAssignedToName string      `json:"previous_assigned_to_name" db:"previous_assigned_to_name"`
	AssignmentReason       string      `json:"assignment_reason" db:"assignment_reason"`
	Metadata               interface{} `json:"metadata" db:"metadata"`
	AssignedAt             time.Time   `json:"assigned_at" db:"assigned_at"`
	AssignedBy             uuid.UUID   `json:"assigned_by" db:"assigned_by"`
}

// UserAssignmentLoad represents current assignment load per user
type UserAssignmentLoad struct {
	ID                uuid.UUID `json:"id" db:"id"`
	OrganizationID    uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	TargetModel       string    `json:"target_model" db:"target_model"`
	ActiveAssignments int       `json:"active_assignments" db:"active_assignments"`
	TotalAssignments  int       `json:"total_assignments" db:"total_assignments"`
	LastAssignedAt    time.Time `json:"last_assigned_at" db:"last_assigned_at"`
	MaxCapacity       int       `json:"max_capacity" db:"max_capacity"`
	Weight            int       `json:"weight" db:"weight"`
	IsAvailable       bool      `json:"is_available" db:"is_available"`
	UnavailableUntil  time.Time `json:"unavailable_until" db:"unavailable_until"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Territory represents a territory definition
type Territory struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	OrganizationID uuid.UUID   `json:"organization_id" db:"organization_id"`
	Name           string      `json:"name" db:"name"`
	Description    string      `json:"description" db:"description"`
	TerritoryType  string      `json:"territory_type" db:"territory_type"`
	Conditions     interface{} `json:"conditions" db:"conditions"`
	AssignedUsers  []uuid.UUID `json:"assigned_users" db:"assigned_users"`
	AssignedTeams  []uuid.UUID `json:"assigned_teams" db:"assigned_teams"`
	Priority       int         `json:"priority" db:"priority"`
	IsActive       bool        `json:"is_active" db:"is_active"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy      uuid.UUID   `json:"created_by" db:"created_by"`
	UpdatedBy      uuid.UUID   `json:"updated_by" db:"updated_by"`
}

// AssignmentStatsByUser represents assignment statistics by user
type AssignmentStatsByUser struct {
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	UserName          string    `json:"user_name" db:"user_name"`
	UserEmail         string    `json:"user_email" db:"user_email"`
	TargetModel       string    `json:"target_model" db:"target_model"`
	ActiveAssignments int       `json:"active_assignments" db:"active_assignments"`
	TotalAssignments  int       `json:"total_assignments" db:"total_assignments"`
	LastAssignedAt    time.Time `json:"last_assigned_at" db:"last_assigned_at"`
	Weight            int       `json:"weight" db:"weight"`
	IsAvailable       bool      `json:"is_available" db:"is_available"`
	AssignmentsToday  int       `json:"assignments_today" db:"assignments_today"`
}

// AssignmentRuleEffectiveness represents effectiveness metrics for assignment rules
type AssignmentRuleEffectiveness struct {
	RuleID              uuid.UUID `json:"rule_id" db:"rule_id"`
	RuleName            string    `json:"rule_name" db:"rule_name"`
	RuleType            string    `json:"rule_type" db:"rule_type"`
	TargetModel         string    `json:"target_model" db:"target_model"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	TotalAssignments    int       `json:"total_assignments" db:"total_assignments"`
	AssignmentsToday    int       `json:"assignments_today" db:"assignments_today"`
	AssignmentsThisWeek int       `json:"assignments_this_week" db:"assignments_this_week"`
	LastUsedAt          time.Time `json:"last_used_at" db:"last_used_at"`
	UniqueAssignees     int       `json:"unique_assignees" db:"unique_assignees"`
}

// AssignmentRuleFilter represents filter criteria for assignment rules
type AssignmentRuleFilter struct {
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty" db:"organization_id"`
	TargetModel    *AssignmentTargetModel `json:"target_model,omitempty" db:"target_model"`
	IsActive       *bool                  `json:"is_active,omitempty" db:"is_active"`
	RuleType       *AssignmentRuleType    `json:"rule_type,omitempty" db:"rule_type"`
}
