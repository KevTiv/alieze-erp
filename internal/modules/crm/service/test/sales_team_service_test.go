package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/internal/testutils"
	"alieze-erp/pkg/events"
)

type SalesTeamServiceTestSuite struct {
	suite.Suite
	service   *service.SalesTeamService
	repo      *testutils.MockSalesTeamRepository
	auth      *testutils.MockAuthService
	eventBus  *events.Bus
	ctx       context.Context
	orgID     uuid.UUID
	userID    uuid.UUID
	teamID    uuid.UUID
	leaderID  uuid.UUID
	memberIDs []uuid.UUID
}

func (s *SalesTeamServiceTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.repo = testutils.NewMockSalesTeamRepository()
	s.auth = testutils.NewMockAuthService()
	s.eventBus = &events.Bus{}
	s.service = service.NewSalesTeamService(s.repo, s.auth, s.eventBus)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.userID = uuid.Must(uuid.NewV7())
	s.teamID = uuid.Must(uuid.NewV7())
	s.leaderID = uuid.Must(uuid.NewV7())
	s.memberIDs = []uuid.UUID{
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
	}

	// Default mock behavior
	s.auth.WithOrganizationID(s.orgID)
	s.auth.WithUserID(s.userID)
}

func (s *SalesTeamServiceTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
}

func (s *SalesTeamServiceTestSuite) TestCreateSalesTeamSuccess() {
	s.T().Run("CreateSalesTeam - Success", func(t *testing.T) {
		// Setup test data
		request := types.SalesTeamCreateRequest{
			Name:         "Test Sales Team",
			Code:         stringPtr("TEST"),
			TeamLeaderID: &s.leaderID,
			MemberIDs:    s.memberIDs,
			IsActive:     true,
		}

		// Mock repository behavior
		expectedTeam := types.SalesTeam{
			ID:             s.teamID,
			OrganizationID: s.orgID,
			Name:           request.Name,
			Code:           request.Code,
			TeamLeaderID:   request.TeamLeaderID,
			MemberIDs:      request.MemberIDs,
			IsActive:       request.IsActive,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		s.repo.WithCreateFunc(func(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error) {
			require.Equal(t, s.orgID, team.OrganizationID)
			require.Equal(t, request.Name, team.Name)
			require.Equal(t, request.Code, team.Code)
			require.Equal(t, request.TeamLeaderID, team.TeamLeaderID)
			require.Equal(t, request.MemberIDs, team.MemberIDs)
			require.Equal(t, request.IsActive, team.IsActive)
			return &expectedTeam, nil
		})

		// Execute
		created, err := s.service.CreateSalesTeam(s.ctx, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedTeam.ID, created.ID)
		require.Equal(t, expectedTeam.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedTeam.Name, created.Name)
		require.Equal(t, expectedTeam.Code, created.Code)
		require.Equal(t, expectedTeam.TeamLeaderID, created.TeamLeaderID)
		require.Equal(t, expectedTeam.MemberIDs, created.MemberIDs)
		require.Equal(t, expectedTeam.IsActive, created.IsActive)
		require.NotZero(t, created.CreatedAt)
		require.NotZero(t, created.UpdatedAt)
	})
}

func (s *SalesTeamServiceTestSuite) TestCreateSalesTeamValidationError() {
	s.T().Run("CreateSalesTeam - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			request     types.SalesTeamCreateRequest
			expectedErr string
		}{
			{
				name:        "Empty Name",
				request:     types.SalesTeamCreateRequest{},
				expectedErr: "sales team name is required",
			},
			{
				name: "Empty Member List",
				request: types.SalesTeamCreateRequest{
					Name:      "Test Team",
					MemberIDs: []uuid.UUID{}, // Empty member list
				},
				expectedErr: "sales team must have at least one member",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				created, err := s.service.CreateSalesTeam(s.ctx, tc.request)

				// Assert
				require.Error(t, err)
				require.Nil(t, created)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *SalesTeamServiceTestSuite) TestGetSalesTeamSuccess() {
	s.T().Run("GetSalesTeam - Success", func(t *testing.T) {
		// Setup test data
		teamID := s.teamID
		expectedTeam := types.SalesTeam{
			ID:             teamID,
			OrganizationID: s.orgID,
			Name:           "Test Sales Team",
			Code:           stringPtr("TEST"),
			TeamLeaderID:   &s.leaderID,
			MemberIDs:      s.memberIDs,
			IsActive:       true,
		}

		// Mock repository behavior
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
			require.Equal(t, teamID, id)
			return &expectedTeam, nil
		})

		// Execute
		team, err := s.service.GetSalesTeam(s.ctx, teamID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, team)
		require.Equal(t, expectedTeam.ID, team.ID)
		require.Equal(t, expectedTeam.OrganizationID, team.OrganizationID)
		require.Equal(t, expectedTeam.Name, team.Name)
		require.Equal(t, expectedTeam.Code, team.Code)
		require.Equal(t, expectedTeam.TeamLeaderID, team.TeamLeaderID)
		require.Equal(t, expectedTeam.MemberIDs, team.MemberIDs)
		require.Equal(t, expectedTeam.IsActive, team.IsActive)
	})
}

func (s *SalesTeamServiceTestSuite) TestGetSalesTeamNotFound() {
	s.T().Run("GetSalesTeam - Not Found", func(t *testing.T) {
		// Setup test data
		teamID := s.teamID

		// Mock repository behavior - return error
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
			require.Equal(t, teamID, id)
			return nil, nil // Not found
		})

		// Execute
		team, err := s.service.GetSalesTeam(s.ctx, teamID)

		// Assert
		require.Error(t, err)
		require.Nil(t, team)
		require.Contains(t, err.Error(), "sales team not found")
	})
}

func (s *SalesTeamServiceTestSuite) TestUpdateSalesTeamSuccess() {
	s.T().Run("UpdateSalesTeam - Success", func(t *testing.T) {
		// Setup test data
		teamID := s.teamID
		newName := "Updated Sales Team"
		newCode := "UPDATED"
		newLeaderID := uuid.Must(uuid.NewV7())
		newMemberIDs := []uuid.UUID{uuid.Must(uuid.NewV7())}

		request := types.SalesTeamUpdateRequest{
			Name:         stringPtr(newName),
			Code:         stringPtr(newCode),
			TeamLeaderID: &newLeaderID,
			MemberIDs:    &newMemberIDs,
			IsActive:     boolPtr(false),
		}

		// Mock repository behavior
		existingTeam := types.SalesTeam{
			ID:             teamID,
			OrganizationID: s.orgID,
			Name:           "Original Team",
			Code:           stringPtr("ORIG"),
			TeamLeaderID:   &s.leaderID,
			MemberIDs:      s.memberIDs,
			IsActive:       true,
		}

		expectedTeam := types.SalesTeam{
			ID:             teamID,
			OrganizationID: s.orgID,
			Name:           newName,
			Code:           stringPtr(newCode),
			TeamLeaderID:   &newLeaderID,
			MemberIDs:      newMemberIDs,
			IsActive:       false,
			UpdatedAt:      time.Now(),
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
			require.Equal(t, teamID, id)
			return &existingTeam, nil
		})

		s.repo.WithUpdateFunc(func(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error) {
			require.Equal(t, teamID, team.ID)
			require.Equal(t, s.orgID, team.OrganizationID)
			require.Equal(t, newName, team.Name)
			require.Equal(t, stringPtr(newCode), team.Code)
			require.Equal(t, &newLeaderID, team.TeamLeaderID)
			require.Equal(t, newMemberIDs, team.MemberIDs)
			require.Equal(t, false, team.IsActive)
			return &expectedTeam, nil
		})

		// Execute
		updated, err := s.service.UpdateSalesTeam(s.ctx, teamID, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, expectedTeam.ID, updated.ID)
		require.Equal(t, expectedTeam.OrganizationID, updated.OrganizationID)
		require.Equal(t, expectedTeam.Name, updated.Name)
		require.Equal(t, expectedTeam.Code, updated.Code)
		require.Equal(t, expectedTeam.TeamLeaderID, updated.TeamLeaderID)
		require.Equal(t, expectedTeam.MemberIDs, updated.MemberIDs)
		require.Equal(t, expectedTeam.IsActive, updated.IsActive)
		require.NotZero(t, updated.UpdatedAt)
	})
}

func (s *SalesTeamServiceTestSuite) TestDeleteSalesTeamSuccess() {
	s.T().Run("DeleteSalesTeam - Success", func(t *testing.T) {
		// Setup test data
		teamID := s.teamID

		// Mock repository behavior
		existingTeam := types.SalesTeam{
			ID:             teamID,
			OrganizationID: s.orgID,
			Name:           "Team to Delete",
			IsActive:       true,
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
			require.Equal(t, teamID, id)
			return &existingTeam, nil
		})

		s.repo.WithDeleteFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, teamID, id)
			return nil
		})

		// Execute
		err := s.service.DeleteSalesTeam(s.ctx, teamID)

		// Assert
		require.NoError(t, err)
	})
}

func (s *SalesTeamServiceTestSuite) TestListSalesTeamsSuccess() {
	s.T().Run("ListSalesTeams - Success", func(t *testing.T) {
		// Setup test data
		filter := types.SalesTeamFilter{
			Name:     stringPtr("Test"),
			IsActive: boolPtr(true),
			Limit:    10,
		}

		// Mock repository behavior
		expectedTeams := []types.SalesTeam{
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Test Team 1",
				Code:           stringPtr("TT1"),
				IsActive:       true,
			},
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Test Team 2",
				Code:           stringPtr("TT2"),
				IsActive:       true,
			},
		}

		s.repo.WithFindAllFunc(func(ctx context.Context, f types.SalesTeamFilter) ([]types.SalesTeam, error) {
			require.Equal(t, s.orgID, f.OrganizationID)
			require.Equal(t, filter.Name, f.Name)
			require.Equal(t, filter.IsActive, f.IsActive)
			require.Equal(t, filter.Limit, f.Limit)
			return expectedTeams, nil
		})

		// Execute
		teams, err := s.service.ListSalesTeams(s.ctx, filter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, teams)
		require.Len(t, teams, 2)
		require.Equal(t, expectedTeams[0].Name, teams[0].Name)
		require.Equal(t, expectedTeams[1].Name, teams[1].Name)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// Run the test suite
func TestSalesTeamServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SalesTeamServiceTestSuite))
}
