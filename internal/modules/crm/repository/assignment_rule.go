package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/alieze-erp/internal/modules/crm/types"
)

// AssignmentRuleRepository defines the interface for assignment rule operations
type AssignmentRuleRepository interface {
	// Assignment Rules
	CreateAssignmentRule(ctx context.Context, rule *types.AssignmentRule) error
	GetAssignmentRule(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error)
	UpdateAssignmentRule(ctx context.Context, rule *types.AssignmentRule) error
	DeleteAssignmentRule(ctx context.Context, id uuid.UUID) error
	ListAssignmentRules(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error)

	// Assignment History
	CreateAssignmentHistory(ctx context.Context, history *types.AssignmentHistory) error
	GetAssignmentHistory(ctx context.Context, id uuid.UUID) (*types.AssignmentHistory, error)
	ListAssignmentHistory(ctx context.Context, orgID uuid.UUID, targetModel string, limit int) ([]*types.AssignmentHistory, error)

	// User Assignment Load
	GetUserAssignmentLoad(ctx context.Context, userID uuid.UUID, targetModel string) (*types.UserAssignmentLoad, error)
	UpdateUserAssignmentLoad(ctx context.Context, load *types.UserAssignmentLoad) error
	ListUserAssignmentLoads(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.UserAssignmentLoad, error)

	// Territories
	CreateTerritory(ctx context.Context, territory *types.Territory) error
	GetTerritory(ctx context.Context, id uuid.UUID) (*types.Territory, error)
	UpdateTerritory(ctx context.Context, territory *types.Territory) error
	DeleteTerritory(ctx context.Context, id uuid.UUID) error
	ListTerritories(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*types.Territory, error)

	// Assignment Functions
	AssignLead(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error
	GetNextAssignee(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error)

	// Analytics
	GetAssignmentStatsByUser(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error)
	GetAssignmentRuleEffectiveness(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error)
}
