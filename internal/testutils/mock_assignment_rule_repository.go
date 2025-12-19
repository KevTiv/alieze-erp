package testutils

import (
	"context"
	"time"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// MockAssignmentRuleRepository implements the repository.AssignmentRuleRepository interface for testing
type MockAssignmentRuleRepository struct {
	createAssignmentRuleFunc           func(ctx context.Context, rule *types.AssignmentRule) error
	getAssignmentRuleFunc              func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error)
	updateAssignmentRuleFunc           func(ctx context.Context, rule *types.AssignmentRule) error
	deleteAssignmentRuleFunc           func(ctx context.Context, id uuid.UUID) error
	listAssignmentRulesFunc            func(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error)
	createTerritoryFunc                func(ctx context.Context, territory *types.Territory) error
	getTerritoryFunc                   func(ctx context.Context, id uuid.UUID) (*types.Territory, error)
	updateTerritoryFunc                func(ctx context.Context, territory *types.Territory) error
	deleteTerritoryFunc                func(ctx context.Context, id uuid.UUID) error
	listTerritoriesFunc                func(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*types.Territory, error)
	assignLeadFunc                     func(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error
	getNextAssigneeFunc                func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error)
	getLeadFunc                        func(ctx context.Context, id uuid.UUID) (*types.Lead, error)
	getAssignmentStatsByUserFunc       func(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error)
	getAssignmentRuleEffectivenessFunc func(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error)
	createAssignmentHistoryFunc        func(ctx context.Context, history *types.AssignmentHistory) error
	getAssignmentHistoryFunc           func(ctx context.Context, id uuid.UUID) (*types.AssignmentHistory, error)
	getUserAssignmentLoadFunc          func(ctx context.Context, userID uuid.UUID, targetModel string) (*types.UserAssignmentLoad, error)
	updateUserAssignmentLoadFunc       func(ctx context.Context, load *types.UserAssignmentLoad) error
	listUserAssignmentLoadsFunc        func(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.UserAssignmentLoad, error)
	listAssignmentHistoryFunc          func(ctx context.Context, orgID uuid.UUID, targetModel string, limit int) ([]*types.AssignmentHistory, error)
}

// NewMockAssignmentRuleRepository creates a new mock assignment rule repository
func NewMockAssignmentRuleRepository() *MockAssignmentRuleRepository {
	return &MockAssignmentRuleRepository{}
}

// CreateAssignmentRule implements the repository interface
func (m *MockAssignmentRuleRepository) CreateAssignmentRule(ctx context.Context, rule *types.AssignmentRule) error {
	if m.createAssignmentRuleFunc != nil {
		return m.createAssignmentRuleFunc(ctx, rule)
	}
	return nil
}

// GetAssignmentRule implements the repository interface
func (m *MockAssignmentRuleRepository) GetAssignmentRule(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
	if m.getAssignmentRuleFunc != nil {
		return m.getAssignmentRuleFunc(ctx, id)
	}
	return &types.AssignmentRule{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		Name:           "Test Rule",
		RuleType:       types.AssignmentRuleTypeRoundRobin,
		TargetModel:    types.AssignmentTargetModelLeads,
	}, nil
}

// UpdateAssignmentRule implements the repository interface
func (m *MockAssignmentRuleRepository) UpdateAssignmentRule(ctx context.Context, rule *types.AssignmentRule) error {
	if m.updateAssignmentRuleFunc != nil {
		return m.updateAssignmentRuleFunc(ctx, rule)
	}
	return nil
}

// DeleteAssignmentRule implements the repository interface
func (m *MockAssignmentRuleRepository) DeleteAssignmentRule(ctx context.Context, id uuid.UUID) error {
	if m.deleteAssignmentRuleFunc != nil {
		return m.deleteAssignmentRuleFunc(ctx, id)
	}
	return nil
}

// ListAssignmentRules implements the repository interface
func (m *MockAssignmentRuleRepository) ListAssignmentRules(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error) {
	if m.listAssignmentRulesFunc != nil {
		return m.listAssignmentRulesFunc(ctx, orgID, targetModel, activeOnly)
	}
	return []*types.AssignmentRule{
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: orgID,
			Name:           "Rule 1",
			RuleType:       types.AssignmentRuleTypeRoundRobin,
			TargetModel:    types.AssignmentTargetModelLeads,
			IsActive:       true,
		},
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: orgID,
			Name:           "Rule 2",
			RuleType:       types.AssignmentRuleTypeTerritory,
			TargetModel:    types.AssignmentTargetModelLeads,
			IsActive:       true,
		},
	}, nil
}

// CreateTerritory implements the repository interface
func (m *MockAssignmentRuleRepository) CreateTerritory(ctx context.Context, territory *types.Territory) error {
	if m.createTerritoryFunc != nil {
		return m.createTerritoryFunc(ctx, territory)
	}
	return nil
}

// GetTerritory implements the repository interface
func (m *MockAssignmentRuleRepository) GetTerritory(ctx context.Context, id uuid.UUID) (*types.Territory, error) {
	if m.getTerritoryFunc != nil {
		return m.getTerritoryFunc(ctx, id)
	}
	return &types.Territory{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		Name:           "Test Territory",
		TerritoryType:  "geographic",
	}, nil
}

// UpdateTerritory implements the repository interface
func (m *MockAssignmentRuleRepository) UpdateTerritory(ctx context.Context, territory *types.Territory) error {
	if m.updateTerritoryFunc != nil {
		return m.updateTerritoryFunc(ctx, territory)
	}
	return nil
}

// DeleteTerritory implements the repository interface
func (m *MockAssignmentRuleRepository) DeleteTerritory(ctx context.Context, id uuid.UUID) error {
	if m.deleteTerritoryFunc != nil {
		return m.deleteTerritoryFunc(ctx, id)
	}
	return nil
}

// ListTerritories implements the repository interface
func (m *MockAssignmentRuleRepository) ListTerritories(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*types.Territory, error) {
	if m.listTerritoriesFunc != nil {
		return m.listTerritoriesFunc(ctx, orgID, activeOnly)
	}
	return []*types.Territory{
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: orgID,
			Name:           "Territory 1",
			TerritoryType:  "geographic",
			IsActive:       true,
		},
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: orgID,
			Name:           "Territory 2",
			TerritoryType:  "industry",
			IsActive:       true,
		},
	}, nil
}

// AssignLead implements the repository interface
func (m *MockAssignmentRuleRepository) AssignLead(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error {
	if m.assignLeadFunc != nil {
		return m.assignLeadFunc(ctx, leadID, userID, reason)
	}
	return nil
}

// GetNextAssignee implements the repository interface
func (m *MockAssignmentRuleRepository) GetNextAssignee(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
	if m.getNextAssigneeFunc != nil {
		return m.getNextAssigneeFunc(ctx, targetModel, conditions)
	}
	return uuid.Must(uuid.NewV7()), "Test User", nil
}

// getLead is a helper function for testing
func (m *MockAssignmentRuleRepository) GetLead(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
	if m.getLeadFunc != nil {
		return m.getLeadFunc(ctx, id)
	}
	return &types.Lead{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		Name:           "Test Lead",
		LeadType:       types.LeadTypeLead,
		Priority:       types.LeadPriorityMedium,
		Probability:    10,
		Active:         true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// GetAssignmentStatsByUser implements the repository interface
func (m *MockAssignmentRuleRepository) GetAssignmentStatsByUser(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error) {
	if m.getAssignmentStatsByUserFunc != nil {
		return m.getAssignmentStatsByUserFunc(ctx, orgID, targetModel)
	}
	return []*types.AssignmentStatsByUser{
		{
			UserID:            uuid.Must(uuid.NewV7()),
			UserName:          "User 1",
			ActiveAssignments: 10,
		},
		{
			UserID:            uuid.Must(uuid.NewV7()),
			UserName:          "User 2",
			ActiveAssignments: 5,
		},
	}, nil
}

// GetAssignmentRuleEffectiveness implements the repository interface
func (m *MockAssignmentRuleRepository) GetAssignmentRuleEffectiveness(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error) {
	if m.getAssignmentRuleEffectivenessFunc != nil {
		return m.getAssignmentRuleEffectivenessFunc(ctx, orgID)
	}
	return []*types.AssignmentRuleEffectiveness{
		{
			RuleID:           uuid.Must(uuid.NewV7()),
			RuleName:         "Rule 1",
			TotalAssignments: 100,
			AssignmentsToday: 5,
		},
	}, nil
}

// CreateAssignmentHistory implements the repository interface
func (m *MockAssignmentRuleRepository) CreateAssignmentHistory(ctx context.Context, history *types.AssignmentHistory) error {
	if m.createAssignmentHistoryFunc != nil {
		return m.createAssignmentHistoryFunc(ctx, history)
	}
	return nil
}

// GetAssignmentHistory implements the repository interface
func (m *MockAssignmentRuleRepository) GetAssignmentHistory(ctx context.Context, id uuid.UUID) (*types.AssignmentHistory, error) {
	if m.getAssignmentHistoryFunc != nil {
		return m.getAssignmentHistoryFunc(ctx, id)
	}
	return &types.AssignmentHistory{
		ID:           id,
		RuleName:     "Test Rule",
		TargetModel:  "leads",
		AssignedToID: uuid.Must(uuid.NewV7()),
	}, nil
}

// GetUserAssignmentLoad implements the repository interface
func (m *MockAssignmentRuleRepository) GetUserAssignmentLoad(ctx context.Context, userID uuid.UUID, targetModel string) (*types.UserAssignmentLoad, error) {
	if m.getUserAssignmentLoadFunc != nil {
		return m.getUserAssignmentLoadFunc(ctx, userID, targetModel)
	}
	return &types.UserAssignmentLoad{
		UserID:            userID,
		TargetModel:       targetModel,
		ActiveAssignments: 0,
		TotalAssignments:  0,
		MaxCapacity:       100,
		IsAvailable:       true,
	}, nil
}

// UpdateUserAssignmentLoad implements the repository interface
func (m *MockAssignmentRuleRepository) UpdateUserAssignmentLoad(ctx context.Context, load *types.UserAssignmentLoad) error {
	if m.updateUserAssignmentLoadFunc != nil {
		return m.updateUserAssignmentLoadFunc(ctx, load)
	}
	return nil
}

// ListUserAssignmentLoads implements the repository interface
func (m *MockAssignmentRuleRepository) ListUserAssignmentLoads(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.UserAssignmentLoad, error) {
	if m.listUserAssignmentLoadsFunc != nil {
		return m.listUserAssignmentLoadsFunc(ctx, orgID, targetModel)
	}
	return []*types.UserAssignmentLoad{
		{
			UserID:            uuid.Must(uuid.NewV7()),
			TargetModel:       targetModel,
			ActiveAssignments: 5,
			TotalAssignments:  10,
			MaxCapacity:       100,
			IsAvailable:       true,
		},
	}, nil
}

// ListAssignmentHistory implements the repository interface
func (m *MockAssignmentRuleRepository) ListAssignmentHistory(ctx context.Context, orgID uuid.UUID, targetModel string, limit int) ([]*types.AssignmentHistory, error) {
	if m.listAssignmentHistoryFunc != nil {
		return m.listAssignmentHistoryFunc(ctx, orgID, targetModel, limit)
	}
	return []*types.AssignmentHistory{
		{
			ID:           uuid.Must(uuid.NewV7()),
			RuleName:     "Test Rule",
			TargetModel:  targetModel,
			AssignedToID: uuid.Must(uuid.NewV7()),
		},
	}, nil
}

// Helper methods to set mock behaviors
func (m *MockAssignmentRuleRepository) WithCreateAssignmentRuleFunc(f func(ctx context.Context, rule *types.AssignmentRule) error) *MockAssignmentRuleRepository {
	m.createAssignmentRuleFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetAssignmentRuleFunc(f func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error)) *MockAssignmentRuleRepository {
	m.getAssignmentRuleFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithUpdateAssignmentRuleFunc(f func(ctx context.Context, rule *types.AssignmentRule) error) *MockAssignmentRuleRepository {
	m.updateAssignmentRuleFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithDeleteAssignmentRuleFunc(f func(ctx context.Context, id uuid.UUID) error) *MockAssignmentRuleRepository {
	m.deleteAssignmentRuleFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithListAssignmentRulesFunc(f func(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error)) *MockAssignmentRuleRepository {
	m.listAssignmentRulesFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithCreateTerritoryFunc(f func(ctx context.Context, territory *types.Territory) error) *MockAssignmentRuleRepository {
	m.createTerritoryFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetTerritoryFunc(f func(ctx context.Context, id uuid.UUID) (*types.Territory, error)) *MockAssignmentRuleRepository {
	m.getTerritoryFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithUpdateTerritoryFunc(f func(ctx context.Context, territory *types.Territory) error) *MockAssignmentRuleRepository {
	m.updateTerritoryFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithDeleteTerritoryFunc(f func(ctx context.Context, id uuid.UUID) error) *MockAssignmentRuleRepository {
	m.deleteTerritoryFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithListTerritoriesFunc(f func(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*types.Territory, error)) *MockAssignmentRuleRepository {
	m.listTerritoriesFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithAssignLeadFunc(f func(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error) *MockAssignmentRuleRepository {
	m.assignLeadFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetNextAssigneeFunc(f func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error)) *MockAssignmentRuleRepository {
	m.getNextAssigneeFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetLeadFunc(f func(ctx context.Context, id uuid.UUID) (*types.Lead, error)) *MockAssignmentRuleRepository {
	m.getLeadFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetAssignmentStatsByUserFunc(f func(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error)) *MockAssignmentRuleRepository {
	m.getAssignmentStatsByUserFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetAssignmentRuleEffectivenessFunc(f func(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error)) *MockAssignmentRuleRepository {
	m.getAssignmentRuleEffectivenessFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithCreateAssignmentHistoryFunc(f func(ctx context.Context, history *types.AssignmentHistory) error) *MockAssignmentRuleRepository {
	m.createAssignmentHistoryFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetAssignmentHistoryFunc(f func(ctx context.Context, id uuid.UUID) (*types.AssignmentHistory, error)) *MockAssignmentRuleRepository {
	m.getAssignmentHistoryFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithGetUserAssignmentLoadFunc(f func(ctx context.Context, userID uuid.UUID, targetModel string) (*types.UserAssignmentLoad, error)) *MockAssignmentRuleRepository {
	m.getUserAssignmentLoadFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithUpdateUserAssignmentLoadFunc(f func(ctx context.Context, load *types.UserAssignmentLoad) error) *MockAssignmentRuleRepository {
	m.updateUserAssignmentLoadFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithListUserAssignmentLoadsFunc(f func(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.UserAssignmentLoad, error)) *MockAssignmentRuleRepository {
	m.listUserAssignmentLoadsFunc = f
	return m
}

func (m *MockAssignmentRuleRepository) WithListAssignmentHistoryFunc(f func(ctx context.Context, orgID uuid.UUID, targetModel string, limit int) ([]*types.AssignmentHistory, error)) *MockAssignmentRuleRepository {
	m.listAssignmentHistoryFunc = f
	return m
}
