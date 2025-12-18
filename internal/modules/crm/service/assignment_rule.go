package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/alieze-erp/internal/modules/crm/repository"
	"github.com/alieze-erp/internal/modules/crm/types"
	"github.com/alieze-erp/pkg/auth"
)

// AssignmentRuleService handles business logic for assignment rules
type AssignmentRuleService struct {
	repo repository.AssignmentRuleRepository
	authService auth.Service
}

// NewAssignmentRuleService creates a new assignment rule service
func NewAssignmentRuleService(repo repository.AssignmentRuleRepository, authService auth.Service) *AssignmentRuleService {
	return &AssignmentRuleService{
		repo:       repo,
		authService: authService,
	}
}

// CreateAssignmentRule creates a new assignment rule
func (s *AssignmentRuleService) CreateAssignmentRule(ctx context.Context, req *types.CreateAssignmentRuleRequest) (*types.AssignmentRule, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if req.RuleType == "" {
		return nil, fmt.Errorf("rule type is required")
	}

	if req.TargetModel == "" {
		return nil, fmt.Errorf("target model is required")
	}

	// Get organization ID and user ID from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization ID: %w", err)
	}

	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Create assignment rule
	rule := &types.AssignmentRule{
		ID:            uuid.New(),
		OrganizationID: orgID,
		Name:          req.Name,
		Description:   req.Description,
		RuleType:      types.AssignmentRuleType(req.RuleType),
		TargetModel:   types.AssignmentTargetModel(req.TargetModel),
		Priority:      req.Priority,
		IsActive:      req.IsActive,
		Conditions:    req.Conditions,
		AssignmentConfig: req.AssignmentConfig,
		AssignToType:  req.AssignToType,
		MaxAssignmentsPerUser: req.MaxAssignmentsPerUser,
		AssignmentWindowStart: req.AssignmentWindowStart,
		AssignmentWindowEnd:   req.AssignmentWindowEnd,
		ActiveDays:    req.ActiveDays,
		CreatedBy:     userID,
		UpdatedBy:     userID,
	}

	err = s.repo.CreateAssignmentRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment rule: %w", err)
	}

	return rule, nil
}

// GetAssignmentRule retrieves an assignment rule by ID
func (s *AssignmentRuleService) GetAssignmentRule(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
	return s.repo.GetAssignmentRule(ctx, id)
}

// UpdateAssignmentRule updates an existing assignment rule
func (s *AssignmentRuleService) UpdateAssignmentRule(ctx context.Context, id uuid.UUID, req *types.UpdateAssignmentRuleRequest) (*types.AssignmentRule, error) {
	// Get existing rule
	existingRule, err := s.repo.GetAssignmentRule(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing rule: %w", err)
	}

	// Get user ID from context
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Update fields
	if req.Name != "" {
		existingRule.Name = req.Name
	}
	if req.Description != "" {
		existingRule.Description = req.Description
	}
	if req.RuleType != "" {
		existingRule.RuleType = types.AssignmentRuleType(req.RuleType)
	}
	if req.TargetModel != "" {
		existingRule.TargetModel = types.AssignmentTargetModel(req.TargetModel)
	}
	if req.Priority != 0 {
		existingRule.Priority = req.Priority
	}
	if req.IsActive != nil {
		existingRule.IsActive = *req.IsActive
	}
	if req.Conditions != nil {
		existingRule.Conditions = req.Conditions
	}
	if req.AssignmentConfig != nil {
		existingRule.AssignmentConfig = req.AssignmentConfig
	}
	if req.AssignToType != "" {
		existingRule.AssignToType = req.AssignToType
	}
	if req.MaxAssignmentsPerUser != 0 {
		existingRule.MaxAssignmentsPerUser = req.MaxAssignmentsPerUser
	}
	if req.AssignmentWindowStart != nil {
		existingRule.AssignmentWindowStart = req.AssignmentWindowStart
	}
	if req.AssignmentWindowEnd != nil {
		existingRule.AssignmentWindowEnd = req.AssignmentWindowEnd
	}
	if req.ActiveDays != nil {
		existingRule.ActiveDays = req.ActiveDays
	}
	existingRule.UpdatedBy = userID
	existingRule.UpdatedAt = time.Now()

	err = s.repo.UpdateAssignmentRule(ctx, existingRule)
	if err != nil {
		return nil, fmt.Errorf("failed to update assignment rule: %w", err)
	}

	return existingRule, nil
}

// DeleteAssignmentRule deletes an assignment rule
func (s *AssignmentRuleService) DeleteAssignmentRule(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAssignmentRule(ctx, id)
}

// ListAssignmentRules lists assignment rules with filters
func (s *AssignmentRuleService) ListAssignmentRules(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error) {
	return s.repo.ListAssignmentRules(ctx, orgID, targetModel, activeOnly)
}

// CreateTerritory creates a new territory
func (s *AssignmentRuleService) CreateTerritory(ctx context.Context, req *types.CreateTerritoryRequest) (*types.Territory, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Get organization ID and user ID from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization ID: %w", err)
	}

	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Create territory
	territory := &types.Territory{
		ID:            uuid.New(),
		OrganizationID: orgID,
		Name:          req.Name,
		Description:   req.Description,
		TerritoryType: req.TerritoryType,
		Conditions:    req.Conditions,
		AssignedUsers: req.AssignedUsers,
		AssignedTeams: req.AssignedTeams,
		Priority:      req.Priority,
		IsActive:      req.IsActive,
		CreatedBy:     userID,
		UpdatedBy:     userID,
	}

	err = s.repo.CreateTerritory(ctx, territory)
	if err != nil {
		return nil, fmt.Errorf("failed to create territory: %w", err)
	}

	return territory, nil
}

// GetTerritory retrieves a territory by ID
func (s *AssignmentRuleService) GetTerritory(ctx context.Context, id uuid.UUID) (*types.Territory, error) {
	return s.repo.GetTerritory(ctx, id)
}

// UpdateTerritory updates an existing territory
func (s *AssignmentRuleService) UpdateTerritory(ctx context.Context, id uuid.UUID, req *types.UpdateTerritoryRequest) (*types.Territory, error) {
	// Get existing territory
	existingTerritory, err := s.repo.GetTerritory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing territory: %w", err)
	}

	// Get user ID from context
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Update fields
	if req.Name != "" {
		existingTerritory.Name = req.Name
	}
	if req.Description != "" {
		existingTerritory.Description = req.Description
	}
	if req.TerritoryType != "" {
		existingTerritory.TerritoryType = req.TerritoryType
	}
	if req.Conditions != nil {
		existingTerritory.Conditions = req.Conditions
	}
	if req.AssignedUsers != nil {
		existingTerritory.AssignedUsers = req.AssignedUsers
	}
	if req.AssignedTeams != nil {
		existingTerritory.AssignedTeams = req.AssignedTeams
	}
	if req.Priority != 0 {
		existingTerritory.Priority = req.Priority
	}
	if req.IsActive != nil {
		existingTerritory.IsActive = *req.IsActive
	}
	existingTerritory.UpdatedBy = userID
	existingTerritory.UpdatedAt = time.Now()

	err = s.repo.UpdateTerritory(ctx, existingTerritory)
	if err != nil {
		return nil, fmt.Errorf("failed to update territory: %w", err)
	}

	return existingTerritory, nil
}

// DeleteTerritory deletes a territory
func (s *AssignmentRuleService) DeleteTerritory(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteTerritory(ctx, id)
}

// ListTerritories lists territories with filters
func (s *AssignmentRuleService) ListTerritories(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*types.Territory, error) {
	return s.repo.ListTerritories(ctx, orgID, activeOnly)
}

// AssignLead assigns a lead to a user based on assignment rules
func (s *AssignmentRuleService) AssignLead(ctx context.Context, leadID uuid.UUID, conditions map[string]interface{}) (*types.AssignmentResult, error) {
	// Get lead to check current assignment
	lead, err := s.getLead(ctx, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	// Determine next assignee
	assigneeID, assigneeName, err := s.repo.GetNextAssignee(ctx, "leads", conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get next assignee: %w", err)
	}

	if assigneeID == uuid.Nil {
		return nil, fmt.Errorf("no suitable assignee found")
	}

	// Check if assignment is needed
	if lead.AssignedTo != uuid.Nil && lead.AssignedTo == assigneeID {
		return &types.AssignmentResult{
			LeadID:        leadID,
			AssignedToID:  assigneeID,
			AssignedToName: assigneeName,
			Reason:        "already_assigned",
			Changed:       false,
		}, nil
	}

	// Assign lead
	reason := "auto_assignment"
	if lead.AssignedTo != uuid.Nil {
		reason = "reassignment"
	}

	err = s.repo.AssignLead(ctx, leadID, assigneeID, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to assign lead: %w", err)
	}

	return &types.AssignmentResult{
		LeadID:        leadID,
		AssignedToID:  assigneeID,
		AssignedToName: assigneeName,
		Reason:        reason,
		Changed:       true,
	}, nil
}

// getLead is a helper function to get lead details
func (s *AssignmentRuleService) getLead(ctx context.Context, leadID uuid.UUID) (*types.Lead, error) {
	// This would be implemented with the actual lead repository
	// For now, we'll return a mock structure
	return &types.Lead{
		ID:        leadID,
		AssignedTo: uuid.Nil, // This would be the actual assigned user
	}, nil
}

// GetAssignmentStatsByUser retrieves assignment statistics by user
func (s *AssignmentRuleService) GetAssignmentStatsByUser(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error) {
	return s.repo.GetAssignmentStatsByUser(ctx, orgID, targetModel)
}

// GetAssignmentRuleEffectiveness retrieves effectiveness metrics for assignment rules
func (s *AssignmentRuleService) GetAssignmentRuleEffectiveness(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error) {
	return s.repo.GetAssignmentRuleEffectiveness(ctx, orgID)
}
