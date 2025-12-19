package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/internal/testutils"
	"github.com/KevTiv/alieze-erp/pkg/events"
)

type AssignmentRuleServiceTestSuite struct {
	suite.Suite
	service     *service.AssignmentRuleService
	repo        *testutils.MockAssignmentRuleRepository
	auth        *testutils.MockAuthService
	eventBus    *events.Bus
	ctx         context.Context
	orgID       uuid.UUID
	userID      uuid.UUID
	ruleID      uuid.UUID
	territoryID uuid.UUID
}

func (s *AssignmentRuleServiceTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.repo = testutils.NewMockAssignmentRuleRepository()
	s.auth = testutils.NewMockAuthService()
	s.eventBus = &events.Bus{}
	s.service = service.NewAssignmentRuleServiceWithEventBus(s.repo, s.auth, s.eventBus)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.userID = uuid.Must(uuid.NewV7())
	s.ruleID = uuid.Must(uuid.NewV7())
	s.territoryID = uuid.Must(uuid.NewV7())

	// Default mock behavior
	s.auth.WithOrganizationID(s.orgID)
	s.auth.WithUserID(s.userID)
}

func (s *AssignmentRuleServiceTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
}

func (s *AssignmentRuleServiceTestSuite) TestCreateAssignmentRuleSuccess() {
	s.T().Run("CreateAssignmentRule - Success", func(t *testing.T) {
		// Setup test data
		request := &types.CreateAssignmentRuleRequest{
			Name:             "Test Rule",
			Description:      "Test Description",
			RuleType:         "round_robin",
			TargetModel:      "leads",
			Priority:         1,
			IsActive:         true,
			Conditions:       types.AssignmentConditions{{Field: "country", Operator: "=", Value: "US"}},
			AssignmentConfig: types.AssignmentConfig{Users: []uuid.UUID{s.userID}},
			AssignToType:     "user",
		}

		// Mock repository behavior
		expectedRule := &types.AssignmentRule{
			ID:                    s.ruleID,
			OrganizationID:        s.orgID,
			Name:                  request.Name,
			Description:           request.Description,
			RuleType:              types.AssignmentRuleType(request.RuleType),
			TargetModel:           types.AssignmentTargetModel(request.TargetModel),
			Priority:              request.Priority,
			IsActive:              request.IsActive,
			Conditions:            request.Conditions,
			AssignmentConfig:      request.AssignmentConfig,
			AssignToType:          request.AssignToType,
			MaxAssignmentsPerUser: request.MaxAssignmentsPerUser,
			CreatedBy:             s.userID,
			UpdatedBy:             s.userID,
		}

		s.repo.WithCreateAssignmentRuleFunc(func(ctx context.Context, rule *types.AssignmentRule) error {
			require.Equal(t, s.orgID, rule.OrganizationID)
			require.Equal(t, request.Name, rule.Name)
			require.Equal(t, types.AssignmentRuleType(request.RuleType), rule.RuleType)
			require.Equal(t, types.AssignmentTargetModel(request.TargetModel), rule.TargetModel)
			require.Equal(t, s.userID, rule.CreatedBy)
			require.Equal(t, s.userID, rule.UpdatedBy)
			return nil
		})

		// Execute
		created, err := s.service.CreateAssignmentRule(s.ctx, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedRule.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedRule.Name, created.Name)
		require.Equal(t, expectedRule.Description, created.Description)
		require.Equal(t, expectedRule.RuleType, created.RuleType)
		require.Equal(t, expectedRule.TargetModel, created.TargetModel)
		require.Equal(t, expectedRule.Priority, created.Priority)
		require.Equal(t, expectedRule.IsActive, created.IsActive)
		require.Equal(t, expectedRule.Conditions, created.Conditions)
		require.Equal(t, expectedRule.AssignmentConfig, created.AssignmentConfig)
		require.Equal(t, expectedRule.AssignToType, created.AssignToType)
		require.Equal(t, expectedRule.CreatedBy, created.CreatedBy)
		require.Equal(t, expectedRule.UpdatedBy, created.UpdatedBy)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestCreateAssignmentRuleValidationError() {
	s.T().Run("CreateAssignmentRule - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			request     *types.CreateAssignmentRuleRequest
			expectedErr string
		}{
			{
				name:        "Empty Name",
				request:     &types.CreateAssignmentRuleRequest{RuleType: "round_robin", TargetModel: "leads"},
				expectedErr: "name is required",
			},
			{
				name:        "Empty Rule Type",
				request:     &types.CreateAssignmentRuleRequest{Name: "Test Rule", TargetModel: "leads"},
				expectedErr: "rule type is required",
			},
			{
				name:        "Empty Target Model",
				request:     &types.CreateAssignmentRuleRequest{Name: "Test Rule", RuleType: "round_robin"},
				expectedErr: "target model is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				created, err := s.service.CreateAssignmentRule(s.ctx, tc.request)

				// Assert
				require.Error(t, err)
				require.Nil(t, created)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *AssignmentRuleServiceTestSuite) TestCreateAssignmentRuleAuthError() {
	s.T().Run("CreateAssignmentRule - Auth Error", func(t *testing.T) {
		// Setup test data
		request := &types.CreateAssignmentRuleRequest{
			Name:        "Test Rule",
			RuleType:    "round_robin",
			TargetModel: "leads",
		}

		// Mock auth service to return error
		s.auth.WithError("GetOrganizationID", errors.New("failed to get organization"))

		// Execute
		created, err := s.service.CreateAssignmentRule(s.ctx, request)

		// Assert
		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "failed to get organization ID")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestGetAssignmentRuleSuccess() {
	s.T().Run("GetAssignmentRule - Success", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID
		expectedRule := &types.AssignmentRule{
			ID:             ruleID,
			OrganizationID: s.orgID,
			Name:           "Test Rule",
			Description:    "Test Description",
			RuleType:       types.AssignmentRuleTypeRoundRobin,
			TargetModel:    types.AssignmentTargetModelLeads,
			Priority:       1,
			IsActive:       true,
		}

		// Mock repository behavior
		s.repo.WithGetAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
			require.Equal(t, ruleID, id)
			return expectedRule, nil
		})

		// Execute
		rule, err := s.service.GetAssignmentRule(s.ctx, ruleID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, rule)
		require.Equal(t, expectedRule.ID, rule.ID)
		require.Equal(t, expectedRule.OrganizationID, rule.OrganizationID)
		require.Equal(t, expectedRule.Name, rule.Name)
		require.Equal(t, expectedRule.Description, rule.Description)
		require.Equal(t, expectedRule.RuleType, rule.RuleType)
		require.Equal(t, expectedRule.TargetModel, rule.TargetModel)
		require.Equal(t, expectedRule.Priority, rule.Priority)
		require.Equal(t, expectedRule.IsActive, rule.IsActive)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestGetAssignmentRuleNotFound() {
	s.T().Run("GetAssignmentRule - Not Found", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID

		// Mock repository behavior - return error
		s.repo.WithGetAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
			require.Equal(t, ruleID, id)
			return nil, errors.New("assignment rule not found")
		})

		// Execute
		rule, err := s.service.GetAssignmentRule(s.ctx, ruleID)

		// Assert
		require.Error(t, err)
		require.Nil(t, rule)
		require.Contains(t, err.Error(), "failed to get existing rule")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestUpdateAssignmentRuleSuccess() {
	s.T().Run("UpdateAssignmentRule - Success", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID
		newName := "Updated Rule"
		newPriority := 2

		request := &types.UpdateAssignmentRuleRequest{
			Name:     stringPtr(newName),
			Priority: intPtr(newPriority),
		}

		// Mock repository behavior
		existingRule := &types.AssignmentRule{
			ID:             ruleID,
			OrganizationID: s.orgID,
			Name:           "Original Rule",
			Priority:       1,
			IsActive:       true,
		}

		expectedRule := &types.AssignmentRule{
			ID:             ruleID,
			OrganizationID: s.orgID,
			Name:           newName,
			Priority:       newPriority,
			IsActive:       true,
			UpdatedBy:      s.userID,
			UpdatedAt:      time.Now(),
		}

		s.repo.WithGetAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
			require.Equal(t, ruleID, id)
			return existingRule, nil
		})

		s.repo.WithUpdateAssignmentRuleFunc(func(ctx context.Context, rule *types.AssignmentRule) error {
			require.Equal(t, ruleID, rule.ID)
			require.Equal(t, s.orgID, rule.OrganizationID)
			require.Equal(t, newName, rule.Name)
			require.Equal(t, newPriority, rule.Priority)
			require.Equal(t, s.userID, rule.UpdatedBy)
			return nil
		})

		// Execute
		updated, err := s.service.UpdateAssignmentRule(s.ctx, ruleID, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, expectedRule.ID, updated.ID)
		require.Equal(t, expectedRule.OrganizationID, updated.OrganizationID)
		require.Equal(t, expectedRule.Name, updated.Name)
		require.Equal(t, expectedRule.Priority, updated.Priority)
		require.Equal(t, expectedRule.IsActive, updated.IsActive)
		require.Equal(t, expectedRule.UpdatedBy, updated.UpdatedBy)
		require.NotZero(t, updated.UpdatedAt)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestUpdateAssignmentRuleNotFound() {
	s.T().Run("UpdateAssignmentRule - Not Found", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID
		request := &types.UpdateAssignmentRuleRequest{
			Name: stringPtr("Updated Rule"),
		}

		// Mock repository behavior - return error
		s.repo.WithGetAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
			require.Equal(t, ruleID, id)
			return nil, errors.New("assignment rule not found")
		})

		// Execute
		updated, err := s.service.UpdateAssignmentRule(s.ctx, ruleID, request)

		// Assert
		require.Error(t, err)
		require.Nil(t, updated)
		require.Contains(t, err.Error(), "failed to get existing rule")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestUpdateAssignmentRuleAuthError() {
	s.T().Run("UpdateAssignmentRule - Auth Error", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID
		request := &types.UpdateAssignmentRuleRequest{
			Name: stringPtr("Updated Rule"),
		}

		// Mock repository behavior
		existingRule := &types.AssignmentRule{
			ID:             ruleID,
			OrganizationID: s.orgID,
			Name:           "Original Rule",
		}

		s.repo.WithGetAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
			return existingRule, nil
		})

		// Mock auth service to return error
		s.auth.WithError("GetUserID", errors.New("failed to get user ID"))

		// Execute
		updated, err := s.service.UpdateAssignmentRule(s.ctx, ruleID, request)

		// Assert
		require.Error(t, err)
		require.Nil(t, updated)
		require.Contains(t, err.Error(), "failed to get user ID")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestDeleteAssignmentRuleSuccess() {
	s.T().Run("DeleteAssignmentRule - Success", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID

		// Mock repository behavior
		s.repo.WithDeleteAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, ruleID, id)
			return nil
		})

		// Execute
		err := s.service.DeleteAssignmentRule(s.ctx, ruleID)

		// Assert
		require.NoError(t, err)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestDeleteAssignmentRuleError() {
	s.T().Run("DeleteAssignmentRule - Error", func(t *testing.T) {
		// Setup test data
		ruleID := s.ruleID

		// Mock repository behavior - return error
		s.repo.WithDeleteAssignmentRuleFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, ruleID, id)
			return errors.New("failed to delete assignment rule")
		})

		// Execute
		err := s.service.DeleteAssignmentRule(s.ctx, ruleID)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete assignment rule")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestListAssignmentRulesSuccess() {
	s.T().Run("ListAssignmentRules - Success", func(t *testing.T) {
		// Setup test data
		targetModel := "leads"
		activeOnly := true

		// Mock repository behavior
		expectedRules := []*types.AssignmentRule{
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Rule 1",
				RuleType:       types.AssignmentRuleTypeRoundRobin,
				TargetModel:    types.AssignmentTargetModelLeads,
				Priority:       1,
				IsActive:       true,
			},
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Rule 2",
				RuleType:       types.AssignmentRuleTypeTerritory,
				TargetModel:    types.AssignmentTargetModelLeads,
				Priority:       2,
				IsActive:       true,
			},
		}

		s.repo.WithListAssignmentRulesFunc(func(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error) {
			require.Equal(t, s.orgID, orgID)
			require.Equal(t, targetModel, targetModel)
			require.Equal(t, activeOnly, activeOnly)
			return expectedRules, nil
		})

		// Execute
		rules, err := s.service.ListAssignmentRules(s.ctx, s.orgID, targetModel, activeOnly)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, rules)
		require.Len(t, rules, 2)
		require.Equal(t, expectedRules[0].Name, rules[0].Name)
		require.Equal(t, expectedRules[1].Name, rules[1].Name)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestCreateTerritorySuccess() {
	s.T().Run("CreateTerritory - Success", func(t *testing.T) {
		// Setup test data
		request := &types.CreateTerritoryRequest{
			Name:          "Test Territory",
			Description:   "Test Description",
			TerritoryType: "geographic",
			Conditions:    map[string]interface{}{"country": "US", "state": "CA"},
			Priority:      1,
			IsActive:      true,
		}

		// Mock repository behavior
		expectedTerritory := &types.Territory{
			ID:             s.territoryID,
			OrganizationID: s.orgID,
			Name:           request.Name,
			Description:    request.Description,
			TerritoryType:  request.TerritoryType,
			Conditions:     request.Conditions,
			Priority:       request.Priority,
			IsActive:       request.IsActive,
			CreatedBy:      s.userID,
			UpdatedBy:      s.userID,
		}

		s.repo.WithCreateTerritoryFunc(func(ctx context.Context, territory *types.Territory) error {
			require.Equal(t, s.orgID, territory.OrganizationID)
			require.Equal(t, request.Name, territory.Name)
			require.Equal(t, request.TerritoryType, territory.TerritoryType)
			require.Equal(t, s.userID, territory.CreatedBy)
			require.Equal(t, s.userID, territory.UpdatedBy)
			return nil
		})

		// Execute
		created, err := s.service.CreateTerritory(s.ctx, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedTerritory.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedTerritory.Name, created.Name)
		require.Equal(t, expectedTerritory.Description, created.Description)
		require.Equal(t, expectedTerritory.TerritoryType, created.TerritoryType)
		require.Equal(t, expectedTerritory.Conditions, created.Conditions)
		require.Equal(t, expectedTerritory.Priority, created.Priority)
		require.Equal(t, expectedTerritory.IsActive, created.IsActive)
		require.Equal(t, expectedTerritory.CreatedBy, created.CreatedBy)
		require.Equal(t, expectedTerritory.UpdatedBy, created.UpdatedBy)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestCreateTerritoryValidationError() {
	s.T().Run("CreateTerritory - Validation Error", func(t *testing.T) {
		// Setup test data - empty name should fail
		request := &types.CreateTerritoryRequest{
			TerritoryType: "geographic",
		}

		// Execute
		created, err := s.service.CreateTerritory(s.ctx, request)

		// Assert
		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "name is required")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestGetTerritorySuccess() {
	s.T().Run("GetTerritory - Success", func(t *testing.T) {
		// Setup test data
		territoryID := s.territoryID
		expectedTerritory := &types.Territory{
			ID:             territoryID,
			OrganizationID: s.orgID,
			Name:           "Test Territory",
			Description:    "Test Description",
			TerritoryType:  "geographic",
			Priority:       1,
			IsActive:       true,
		}

		// Mock repository behavior
		s.repo.WithGetTerritoryFunc(func(ctx context.Context, id uuid.UUID) (*types.Territory, error) {
			require.Equal(t, territoryID, id)
			return expectedTerritory, nil
		})

		// Execute
		territory, err := s.service.GetTerritory(s.ctx, territoryID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, territory)
		require.Equal(t, expectedTerritory.ID, territory.ID)
		require.Equal(t, expectedTerritory.OrganizationID, territory.OrganizationID)
		require.Equal(t, expectedTerritory.Name, territory.Name)
		require.Equal(t, expectedTerritory.Description, territory.Description)
		require.Equal(t, expectedTerritory.TerritoryType, territory.TerritoryType)
		require.Equal(t, expectedTerritory.Priority, territory.Priority)
		require.Equal(t, expectedTerritory.IsActive, territory.IsActive)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestAssignLeadSuccess() {
	s.T().Run("AssignLead - Success", func(t *testing.T) {
		// Setup test data
		leadID := uuid.Must(uuid.NewV7())
		assigneeID := uuid.Must(uuid.NewV7())
		assigneeName := "John Doe"
		conditions := map[string]interface{}{
			"country": "US",
			"source":  "web",
		}

		// Mock repository behavior
		s.repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
			require.Equal(t, leadID, id)
			return &types.Lead{
				ID:         leadID,
				AssignedTo: nil, // Not assigned yet
			}, nil
		})

		s.repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
			require.Equal(t, "leads", targetModel)
			require.Equal(t, conditions, conditions)
			return assigneeID, assigneeName, nil
		})

		s.repo.WithAssignLeadFunc(func(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error {
			require.Equal(t, leadID, leadID)
			require.Equal(t, assigneeID, userID)
			require.Equal(t, "auto_assignment", reason)
			return nil
		})

		// Execute
		result, err := s.service.AssignLead(s.ctx, leadID, conditions)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, leadID, result.LeadID)
		require.Equal(t, assigneeID, result.AssignedToID)
		require.Equal(t, assigneeName, result.AssignedToName)
		require.Equal(t, "auto_assignment", result.Reason)
		require.True(t, result.Changed)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestAssignLeadAlreadyAssigned() {
	s.T().Run("AssignLead - Already Assigned", func(t *testing.T) {
		// Setup test data
		leadID := uuid.Must(uuid.NewV7())
		assigneeID := uuid.Must(uuid.NewV7())
		assigneeName := "John Doe"
		conditions := map[string]interface{}{
			"country": "US",
		}

		// Mock repository behavior - lead already assigned to the same user
		s.repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
			require.Equal(t, leadID, id)
			return &types.Lead{
				ID:         leadID,
				AssignedTo: &assigneeID, // Already assigned
			}, nil
		})

		s.repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
			return assigneeID, assigneeName, nil // Same assignee
		})

		// Execute
		result, err := s.service.AssignLead(s.ctx, leadID, conditions)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, leadID, result.LeadID)
		require.Equal(t, assigneeID, result.AssignedToID)
		require.Equal(t, assigneeName, result.AssignedToName)
		require.Equal(t, "already_assigned", result.Reason)
		require.False(t, result.Changed)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestAssignLeadReassignment() {
	s.T().Run("AssignLead - Reassignment", func(t *testing.T) {
		// Setup test data
		leadID := uuid.Must(uuid.NewV7())
		oldAssigneeID := uuid.Must(uuid.NewV7())
		newAssigneeID := uuid.Must(uuid.NewV7())
		assigneeName := "Jane Smith"
		conditions := map[string]interface{}{
			"country": "US",
		}

		// Mock repository behavior - lead assigned to different user
		s.repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
			require.Equal(t, leadID, id)
			return &types.Lead{
				ID:         leadID,
				AssignedTo: &oldAssigneeID, // Assigned to different user
			}, nil
		})

		s.repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
			return newAssigneeID, assigneeName, nil // New assignee
		})

		s.repo.WithAssignLeadFunc(func(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error {
			require.Equal(t, leadID, leadID)
			require.Equal(t, newAssigneeID, userID)
			require.Equal(t, "reassignment", reason)
			return nil
		})

		// Execute
		result, err := s.service.AssignLead(s.ctx, leadID, conditions)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, leadID, result.LeadID)
		require.Equal(t, newAssigneeID, result.AssignedToID)
		require.Equal(t, assigneeName, result.AssignedToName)
		require.Equal(t, "reassignment", result.Reason)
		require.True(t, result.Changed)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestAssignLeadNoAssigneeFound() {
	s.T().Run("AssignLead - No Assignee Found", func(t *testing.T) {
		// Setup test data
		leadID := uuid.Must(uuid.NewV7())
		conditions := map[string]interface{}{
			"country": "XX", // No assignee for this country
		}

		// Mock repository behavior
		s.repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
			return &types.Lead{
				ID:         leadID,
				AssignedTo: nil,
			}, nil
		})

		s.repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
			return uuid.Nil, "", nil // No assignee found
		})

		// Execute
		result, err := s.service.AssignLead(s.ctx, leadID, conditions)

		// Assert
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "no suitable assignee found")
	})
}

func (s *AssignmentRuleServiceTestSuite) TestGetAssignmentStatsByUserSuccess() {
	s.T().Run("GetAssignmentStatsByUser - Success", func(t *testing.T) {
		// Setup test data
		targetModel := "leads"

		// Mock repository behavior
		expectedStats := []*types.AssignmentStatsByUser{
			{
				UserID:            s.userID,
				UserName:          "John Doe",
				ActiveAssignments: 10,
			},
			{
				UserID:            uuid.Must(uuid.NewV7()),
				UserName:          "Jane Smith",
				ActiveAssignments: 5,
			},
		}

		s.repo.WithGetAssignmentStatsByUserFunc(func(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error) {
			require.Equal(t, s.orgID, orgID)
			require.Equal(t, targetModel, targetModel)
			return expectedStats, nil
		})

		// Execute
		stats, err := s.service.GetAssignmentStatsByUser(s.ctx, s.orgID, targetModel)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, stats)
		require.Len(t, stats, 2)
		require.Equal(t, expectedStats[0].UserID, stats[0].UserID)
		require.Equal(t, expectedStats[0].UserName, stats[0].UserName)
		require.Equal(t, expectedStats[0].ActiveAssignments, stats[0].ActiveAssignments)
		require.Equal(t, expectedStats[1].UserID, stats[1].UserID)
		require.Equal(t, expectedStats[1].UserName, stats[1].UserName)
		require.Equal(t, expectedStats[1].ActiveAssignments, stats[1].ActiveAssignments)
	})
}

func (s *AssignmentRuleServiceTestSuite) TestGetAssignmentRuleEffectivenessSuccess() {
	s.T().Run("GetAssignmentRuleEffectiveness - Success", func(t *testing.T) {
		// Mock repository behavior
		expectedEffectiveness := []*types.AssignmentRuleEffectiveness{
			{
				RuleID:           s.ruleID,
				RuleName:         "Test Rule",
				TotalAssignments: 100,
				AssignmentsToday: 5,
			},
		}

		s.repo.WithGetAssignmentRuleEffectivenessFunc(func(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error) {
			require.Equal(t, s.orgID, orgID)
			return expectedEffectiveness, nil
		})

		// Execute
		effectiveness, err := s.service.GetAssignmentRuleEffectiveness(s.ctx, s.orgID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, effectiveness)
		require.Len(t, effectiveness, 1)
		require.Equal(t, expectedEffectiveness[0].RuleID, effectiveness[0].RuleID)
		require.Equal(t, expectedEffectiveness[0].RuleName, effectiveness[0].RuleName)
		require.Equal(t, expectedEffectiveness[0].TotalAssignments, effectiveness[0].TotalAssignments)
		require.Equal(t, expectedEffectiveness[0].AssignmentsToday, effectiveness[0].AssignmentsToday)
	})
}

// Run the test suite
func TestAssignmentRuleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AssignmentRuleServiceTestSuite))
}
