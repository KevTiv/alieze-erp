package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/internal/testutils"
	"alieze-erp/pkg/rules"
)

type LeadServiceTestSuite struct {
	suite.Suite
	service    *service.LeadService
	repo       *testutils.MockLeadRepository
	ruleEngine *rules.RuleEngine
	ctx        context.Context
	orgID      uuid.UUID
	userID     uuid.UUID
	leadID     uuid.UUID
	stageID    uuid.UUID
	sourceID   uuid.UUID
}

func (s *LeadServiceTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.repo = testutils.NewMockLeadRepository()
	s.ruleEngine = &rules.RuleEngine{}
	s.service = service.NewLeadService(service.NewLeadServiceOptions{
		LeadRepository: s.repo,
		RuleEngine:     s.ruleEngine,
	})
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.userID = uuid.Must(uuid.NewV7())
	s.leadID = uuid.Must(uuid.NewV7())
	s.stageID = uuid.Must(uuid.NewV7())
	s.sourceID = uuid.Must(uuid.NewV7())
}

func (s *LeadServiceTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
}

func (s *LeadServiceTestSuite) TestCreateLeadSuccess() {
	s.T().Run("CreateLead - Success", func(t *testing.T) {
		// Setup test data
		leadRequest := types.LeadCreateRequest{
			Name:            "Test Lead",
			ContactName:     stringPtr("John Doe"),
			Email:           stringPtr("john@example.com"),
			Phone:           stringPtr("1234567890"),
			StageID:         &s.stageID,
			SourceID:        &s.sourceID,
			ExpectedRevenue: floatPtr(1000.0),
			Probability:     50,
			Active:          true,
		}

		// Mock repository behavior
		expectedLead := &types.LeadEnhanced{
			ID:              uuid.Must(uuid.NewV7()),
			OrganizationID:  s.orgID,
			Name:            leadRequest.Name,
			ContactName:     leadRequest.ContactName,
			Email:           leadRequest.Email,
			Phone:           leadRequest.Phone,
			StageID:         leadRequest.StageID,
			SourceID:        leadRequest.SourceID,
			ExpectedRevenue: leadRequest.ExpectedRevenue,
			Probability:     leadRequest.Probability,
			Active:          leadRequest.Active,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		s.repo.WithCreateFunc(func(ctx context.Context, lead *types.LeadEnhanced) error {
			require.Equal(t, s.orgID, lead.OrganizationID)
			require.Equal(t, leadRequest.Name, lead.Name)
			require.Equal(t, leadRequest.ContactName, lead.ContactName)
			require.Equal(t, leadRequest.Email, lead.Email)
			require.Equal(t, leadRequest.Phone, lead.Phone)
			require.Equal(t, types.LeadTypeLead, lead.LeadType)
			require.Equal(t, types.LeadPriorityMedium, lead.Priority)
			require.Equal(t, 10, lead.Probability) // Default probability
			return nil
		})

		// Execute
		created, err := s.service.CreateLead(s.ctx, s.orgID, leadRequest)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedLead.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedLead.Name, created.Name)
		require.Equal(t, expectedLead.ContactName, created.ContactName)
		require.Equal(t, expectedLead.Email, created.Email)
		require.Equal(t, expectedLead.Phone, created.Phone)
		require.Equal(t, expectedLead.StageID, created.StageID)
		require.Equal(t, expectedLead.SourceID, created.SourceID)
		require.Equal(t, expectedLead.ExpectedRevenue, created.ExpectedRevenue)
		require.Equal(t, expectedLead.Probability, created.Probability)
		require.Equal(t, expectedLead.Active, created.Active)
		require.NotZero(t, created.CreatedAt)
		require.NotZero(t, created.UpdatedAt)
	})
}

func (s *LeadServiceTestSuite) TestCreateLeadValidationError() {
	s.T().Run("CreateLead - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			lead        types.LeadCreateRequest
			expectedErr string
		}{
			{
				name:        "Empty Name",
				lead:        types.LeadCreateRequest{},
				expectedErr: "lead name is required",
			},
			{
				name: "Invalid Email",
				lead: types.LeadCreateRequest{
					Name:  "Test Lead",
					Email: stringPtr("invalid-email"),
				},
				expectedErr: "invalid email format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				created, err := s.service.CreateLead(s.ctx, s.orgID, tc.lead)

				// Assert
				require.Error(t, err)
				require.Nil(t, created)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *LeadServiceTestSuite) TestCreateLeadRuleEngineError() {
	s.T().Run("CreateLead - Rule Engine Error", func(t *testing.T) {
		// Setup test data
		leadRequest := types.LeadCreateRequest{
			Name: "Test Lead",
		}

		// Mock rule engine to return error
		mockRuleEngine := &MockRuleEngine{
			validateError: errors.New("rule validation failed"),
		}

		testService := service.NewLeadService(service.NewLeadServiceOptions{
			LeadRepository: s.repo,
			RuleEngine:     mockRuleEngine,
		})

		// Execute
		created, err := testService.CreateLead(s.ctx, s.orgID, leadRequest)

		// Assert
		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "failed to apply business rules")
	})
}

func (s *LeadServiceTestSuite) TestGetLeadSuccess() {
	s.T().Run("GetLead - Success", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID
		expectedLead := &types.LeadEnhanced{
			ID:             leadID,
			OrganizationID: s.orgID,
			Name:           "Test Lead",
			ContactName:    stringPtr("John Doe"),
			Email:          stringPtr("john@example.com"),
			Phone:          stringPtr("1234567890"),
			StageID:        &s.stageID,
			SourceID:       &s.sourceID,
			Active:         true,
		}

		// Mock repository behavior
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			require.Equal(t, leadID, id)
			return expectedLead, nil
		})

		// Execute
		lead, err := s.service.GetLead(s.ctx, s.orgID, leadID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, lead)
		require.Equal(t, expectedLead.ID, lead.ID)
		require.Equal(t, expectedLead.OrganizationID, lead.OrganizationID)
		require.Equal(t, expectedLead.Name, lead.Name)
		require.Equal(t, expectedLead.ContactName, lead.ContactName)
		require.Equal(t, expectedLead.Email, lead.Email)
		require.Equal(t, expectedLead.Phone, lead.Phone)
		require.Equal(t, expectedLead.StageID, lead.StageID)
		require.Equal(t, expectedLead.SourceID, lead.SourceID)
		require.Equal(t, expectedLead.Active, lead.Active)
	})
}

func (s *LeadServiceTestSuite) TestGetLeadNotFound() {
	s.T().Run("GetLead - Not Found", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID

		// Mock repository behavior - return nil (not found)
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			require.Equal(t, leadID, id)
			return nil, nil
		})

		// Execute
		lead, err := s.service.GetLead(s.ctx, s.orgID, leadID)

		// Assert
		require.Error(t, err)
		require.Nil(t, lead)
		require.Contains(t, err.Error(), "lead not found")
	})
}

func (s *LeadServiceTestSuite) TestGetLeadOrganizationMismatch() {
	s.T().Run("GetLead - Organization Mismatch", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID
		otherOrgID := uuid.Must(uuid.NewV7())

		// Mock repository behavior - return lead from different organization
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			return &types.LeadEnhanced{
				ID:             leadID,
				OrganizationID: otherOrgID, // Different organization
				Name:           "Test Lead",
			}, nil
		})

		// Execute
		lead, err := s.service.GetLead(s.ctx, s.orgID, leadID)

		// Assert
		require.Error(t, err)
		require.Nil(t, lead)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *LeadServiceTestSuite) TestUpdateLeadSuccess() {
	s.T().Run("UpdateLead - Success", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID
		newName := "Updated Lead"
		newEmail := "updated@example.com"
		newProbability := 75

		updateRequest := types.LeadEnhancedUpdateRequest{
			Name:        stringPtr(newName),
			Email:       stringPtr(newEmail),
			Probability: intPtr(newProbability),
		}

		// Mock repository behavior
		existingLead := &types.LeadEnhanced{
			ID:             leadID,
			OrganizationID: s.orgID,
			Name:           "Original Lead",
			Email:          stringPtr("original@example.com"),
			Probability:    50,
			Active:         true,
		}

		expectedLead := &types.LeadEnhanced{
			ID:             leadID,
			OrganizationID: s.orgID,
			Name:           newName,
			Email:          stringPtr(newEmail),
			Probability:    newProbability,
			Active:         true,
			UpdatedAt:      time.Now(),
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			require.Equal(t, leadID, id)
			return existingLead, nil
		})

		s.repo.WithUpdateFunc(func(ctx context.Context, lead *types.LeadEnhanced) error {
			require.Equal(t, leadID, lead.ID)
			require.Equal(t, s.orgID, lead.OrganizationID)
			require.Equal(t, newName, lead.Name)
			require.Equal(t, stringPtr(newEmail), lead.Email)
			require.Equal(t, newProbability, lead.Probability)
			return nil
		})

		// Execute
		updated, err := s.service.UpdateLead(s.ctx, s.orgID, leadID, updateRequest)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, expectedLead.ID, updated.ID)
		require.Equal(t, expectedLead.OrganizationID, updated.OrganizationID)
		require.Equal(t, expectedLead.Name, updated.Name)
		require.Equal(t, expectedLead.Email, updated.Email)
		require.Equal(t, expectedLead.Probability, updated.Probability)
		require.Equal(t, expectedLead.Active, updated.Active)
		require.NotZero(t, updated.UpdatedAt)
	})
}

func (s *LeadServiceTestSuite) TestUpdateLeadValidationError() {
	s.T().Run("UpdateLead - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			leadID      uuid.UUID
			update      types.LeadEnhancedUpdateRequest
			expectedErr string
		}{
			{
				name:        "Invalid Lead ID",
				leadID:      uuid.Nil,
				update:      types.LeadEnhancedUpdateRequest{},
				expectedErr: "invalid lead ID",
			},
			{
				name:   "Invalid Email",
				leadID: s.leadID,
				update: types.LeadEnhancedUpdateRequest{
					Email: stringPtr("invalid-email"),
				},
				expectedErr: "invalid email format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// For valid lead ID cases, mock the repository
				if tc.leadID != uuid.Nil {
					existingLead := &types.LeadEnhanced{
						ID:             tc.leadID,
						OrganizationID: s.orgID,
						Name:           "Original Lead",
					}

					s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
						return existingLead, nil
					})
				}

				// Execute
				updated, err := s.service.UpdateLead(s.ctx, s.orgID, tc.leadID, tc.update)

				// Assert
				require.Error(t, err)
				require.Nil(t, updated)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *LeadServiceTestSuite) TestUpdateLeadOrganizationMismatch() {
	s.T().Run("UpdateLead - Organization Mismatch", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID
		otherOrgID := uuid.Must(uuid.NewV7())

		updateRequest := types.LeadEnhancedUpdateRequest{
			Name: stringPtr("Updated Lead"),
		}

		// Mock repository behavior - return lead from different organization
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			return &types.LeadEnhanced{
				ID:             leadID,
				OrganizationID: otherOrgID, // Different organization
				Name:           "Original Lead",
			}, nil
		})

		// Execute
		updated, err := s.service.UpdateLead(s.ctx, s.orgID, leadID, updateRequest)

		// Assert
		require.Error(t, err)
		require.Nil(t, updated)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *LeadServiceTestSuite) TestDeleteLeadSuccess() {
	s.T().Run("DeleteLead - Success", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID

		// Mock repository behavior
		existingLead := &types.LeadEnhanced{
			ID:             leadID,
			OrganizationID: s.orgID,
			Name:           "Lead to Delete",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			require.Equal(t, leadID, id)
			return existingLead, nil
		})

		s.repo.WithDeleteFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, leadID, id)
			return nil
		})

		// Execute
		err := s.service.DeleteLead(s.ctx, s.orgID, leadID)

		// Assert
		require.NoError(t, err)
	})
}

func (s *LeadServiceTestSuite) TestDeleteLeadInvalidID() {
	s.T().Run("DeleteLead - Invalid ID", func(t *testing.T) {
		// Execute
		err := s.service.DeleteLead(s.ctx, s.orgID, uuid.Nil)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid lead ID")
	})
}

func (s *LeadServiceTestSuite) TestDeleteLeadOrganizationMismatch() {
	s.T().Run("DeleteLead - Organization Mismatch", func(t *testing.T) {
		// Setup test data
		leadID := s.leadID
		otherOrgID := uuid.Must(uuid.NewV7())

		// Mock repository behavior - return lead from different organization
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
			return &types.LeadEnhanced{
				ID:             leadID,
				OrganizationID: otherOrgID, // Different organization
				Name:           "Lead to Delete",
			}, nil
		})

		// Execute
		err := s.service.DeleteLead(s.ctx, s.orgID, leadID)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *LeadServiceTestSuite) TestListLeadsSuccess() {
	s.T().Run("ListLeads - Success", func(t *testing.T) {
		// Setup test data
		filter := types.LeadEnhancedFilter{
			Name:  stringPtr("Test"),
			Limit: 10,
		}

		// Mock repository behavior
		expectedLeads := []*types.LeadEnhanced{
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Test Lead 1",
				ContactName:    stringPtr("John Doe"),
				Email:          stringPtr("john@example.com"),
				Active:         true,
			},
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Test Lead 2",
				ContactName:    stringPtr("Jane Smith"),
				Email:          stringPtr("jane@example.com"),
				Active:         false,
			},
		}

		s.repo.WithFindAllFunc(func(ctx context.Context, f types.LeadEnhancedFilter) ([]*types.LeadEnhanced, error) {
			require.Equal(t, s.orgID, f.OrganizationID)
			require.Equal(t, filter.Name, f.Name)
			require.Equal(t, filter.Limit, f.Limit)
			return expectedLeads, nil
		})

		// Execute
		leads, err := s.service.ListLeads(s.ctx, s.orgID, filter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, leads)
		require.Len(t, leads, 2)
		require.Equal(t, expectedLeads[0].Name, leads[0].Name)
		require.Equal(t, expectedLeads[1].Name, leads[1].Name)
	})
}

func (s *LeadServiceTestSuite) TestCountLeadsSuccess() {
	s.T().Run("CountLeads - Success", func(t *testing.T) {
		// Setup test data
		filter := types.LeadEnhancedFilter{
			Active: boolPtr(true),
		}

		expectedCount := 42

		// Mock repository behavior
		s.repo.WithCountFunc(func(ctx context.Context, f types.LeadEnhancedFilter) (int, error) {
			require.Equal(t, s.orgID, f.OrganizationID)
			require.Equal(t, filter.Active, f.Active)
			return expectedCount, nil
		})

		// Execute
		count, err := s.service.CountLeads(s.ctx, s.orgID, filter)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedCount, count)
	})
}

// Helper functions
type MockRuleEngine struct {
	validateError error
}

func (m *MockRuleEngine) Validate(ctx context.Context, entityType string, entity interface{}) error {
	return m.validateError
}

// Run the test suite
func TestLeadServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LeadServiceTestSuite))
}
