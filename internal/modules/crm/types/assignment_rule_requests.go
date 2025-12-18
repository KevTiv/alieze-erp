package types

import (
	"time"

	"github.com/google/uuid"
)

// CreateAssignmentRuleRequest represents a request to create an assignment rule
type CreateAssignmentRuleRequest struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	RuleType              string                 `json:"rule_type"`
	TargetModel           string                 `json:"target_model"`
	Priority              int                    `json:"priority"`
	IsActive              bool                   `json:"is_active"`
	Conditions            AssignmentConditions   `json:"conditions"`
	AssignmentConfig      AssignmentConfig       `json:"assignment_config"`
	AssignToType          string                 `json:"assign_to_type"`
	MaxAssignmentsPerUser int                    `json:"max_assignments_per_user"`
	AssignmentWindowStart *time.Time             `json:"assignment_window_start"`
	AssignmentWindowEnd   *time.Time             `json:"assignment_window_end"`
	ActiveDays            []int                  `json:"active_days"`
}

// UpdateAssignmentRuleRequest represents a request to update an assignment rule
type UpdateAssignmentRuleRequest struct {
	Name                  *string                `json:"name,omitempty"`
	Description           *string                `json:"description,omitempty"`
	RuleType              *string                `json:"rule_type,omitempty"`
	TargetModel           *string                `json:"target_model,omitempty"`
	Priority              *int                   `json:"priority,omitempty"`
	IsActive              *bool                  `json:"is_active,omitempty"`
	Conditions            *AssignmentConditions  `json:"conditions,omitempty"`
	AssignmentConfig      *AssignmentConfig      `json:"assignment_config,omitempty"`
	AssignToType          *string                `json:"assign_to_type,omitempty"`
	MaxAssignmentsPerUser *int                   `json:"max_assignments_per_user,omitempty"`
	AssignmentWindowStart *time.Time             `json:"assignment_window_start,omitempty"`
	AssignmentWindowEnd   *time.Time             `json:"assignment_window_end,omitempty"`
	ActiveDays            *[]int                 `json:"active_days,omitempty"`
}

// CreateTerritoryRequest represents a request to create a territory
type CreateTerritoryRequest struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	TerritoryType string        `json:"territory_type"`
	Conditions    interface{}   `json:"conditions"`
	AssignedUsers []uuid.UUID   `json:"assigned_users"`
	AssignedTeams []uuid.UUID   `json:"assigned_teams"`
	Priority      int           `json:"priority"`
	IsActive      bool          `json:"is_active"`
}

// UpdateTerritoryRequest represents a request to update a territory
type UpdateTerritoryRequest struct {
	Name          *string       `json:"name,omitempty"`
	Description   *string       `json:"description,omitempty"`
	TerritoryType *string       `json:"territory_type,omitempty"`
	Conditions    *interface{}  `json:"conditions,omitempty"`
	AssignedUsers *[]uuid.UUID  `json:"assigned_users,omitempty"`
	AssignedTeams *[]uuid.UUID  `json:"assigned_teams,omitempty"`
	Priority      *int          `json:"priority,omitempty"`
	IsActive      *bool         `json:"is_active,omitempty"`
}

// AssignmentResult represents the result of an assignment operation
type AssignmentResult struct {
	LeadID        uuid.UUID `json:"lead_id"`
	AssignedToID  uuid.UUID `json:"assigned_to_id"`
	AssignedToName string    `json:"assigned_to_name"`
	Reason        string    `json:"reason"`
	Changed       bool      `json:"changed"`
}
