package service_test

import (
	"context"
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

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func activityTypePtr(at types.ActivityType) *types.ActivityType {
	return &at
}

func activityStatePtr(as types.ActivityState) *types.ActivityState {
	return &as
}

type ActivityServiceTestSuite struct {
	suite.Suite
	service    *service.ActivityService
	repo       *testutils.MockActivityRepository
	auth       *testutils.MockAuthService
	eventBus   *events.Bus
	ctx        context.Context
	orgID      uuid.UUID
	userID     uuid.UUID
	activityID uuid.UUID
	contactID  uuid.UUID
}

func (s *ActivityServiceTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.repo = testutils.NewMockActivityRepository()
	s.auth = testutils.NewMockAuthService()
	s.eventBus = &events.Bus{}
	s.service = service.NewActivityService(s.repo, s.auth, s.eventBus)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.userID = uuid.Must(uuid.NewV7())
	s.activityID = uuid.Must(uuid.NewV7())
	s.contactID = uuid.Must(uuid.NewV7())

	// Default mock behavior
	s.auth.WithOrganizationID(s.orgID)
	s.auth.WithUserID(s.userID)
}

func (s *ActivityServiceTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
}

func (s *ActivityServiceTestSuite) TestCreateActivitySuccess() {
	s.T().Run("CreateActivity - Success", func(t *testing.T) {
		// Setup test data
		now := time.Now()
		deadline := now.Add(24 * time.Hour)
		request := types.ActivityCreateRequest{
			ActivityType: types.ActivityTypeMeeting,
			Summary:      "Client Meeting",
			Note:         stringPtr("Discuss project requirements"),
			DateDeadline: &deadline,
			UserID:       &s.userID,
			AssignedTo:   &s.userID,
			ResModel:     stringPtr("contacts"),
			ResID:        &s.contactID,
			State:        types.ActivityStatePlanned,
		}

		// Mock repository behavior
		expectedActivity := types.Activity{
			ID:             s.activityID,
			OrganizationID: s.orgID,
			ActivityType:   request.ActivityType,
			Summary:        request.Summary,
			Note:           request.Note,
			DateDeadline:   request.DateDeadline,
			UserID:         request.UserID,
			AssignedTo:     request.AssignedTo,
			ResModel:       request.ResModel,
			ResID:          request.ResID,
			State:          request.State,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		s.repo.WithCreateFunc(func(ctx context.Context, activity types.Activity) (*types.Activity, error) {
			require.Equal(t, s.orgID, activity.OrganizationID)
			require.Equal(t, request.ActivityType, activity.ActivityType)
			require.Equal(t, request.Summary, activity.Summary)
			require.Equal(t, request.Note, activity.Note)
			require.Equal(t, request.DateDeadline, activity.DateDeadline)
			require.Equal(t, request.UserID, activity.UserID)
			require.Equal(t, request.AssignedTo, activity.AssignedTo)
			require.Equal(t, request.ResModel, activity.ResModel)
			require.Equal(t, request.ResID, activity.ResID)
			require.Equal(t, request.State, activity.State)
			return &expectedActivity, nil
		})

		// Execute
		created, err := s.service.CreateActivity(s.ctx, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedActivity.ID, created.ID)
		require.Equal(t, expectedActivity.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedActivity.ActivityType, created.ActivityType)
		require.Equal(t, expectedActivity.Summary, created.Summary)
		require.Equal(t, expectedActivity.Note, created.Note)
		require.Equal(t, expectedActivity.DateDeadline, created.DateDeadline)
		require.Equal(t, expectedActivity.UserID, created.UserID)
		require.Equal(t, expectedActivity.AssignedTo, created.AssignedTo)
		require.Equal(t, expectedActivity.ResModel, created.ResModel)
		require.Equal(t, expectedActivity.ResID, created.ResID)
		require.Equal(t, expectedActivity.State, created.State)
		require.NotZero(t, created.CreatedAt)
		require.NotZero(t, created.UpdatedAt)
	})
}

func (s *ActivityServiceTestSuite) TestCreateActivityValidationError() {
	s.T().Run("CreateActivity - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			request     types.ActivityCreateRequest
			expectedErr string
		}{
			{
				name:        "Empty Summary",
				request:     types.ActivityCreateRequest{},
				expectedErr: "activity summary is required",
			},
			{
				name: "Invalid Activity Type",
				request: types.ActivityCreateRequest{
					ActivityType: "invalid_type",
					Summary:      "Test Activity",
				},
				expectedErr: "invalid activity type",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				created, err := s.service.CreateActivity(s.ctx, tc.request)

				// Assert
				require.Error(t, err)
				require.Nil(t, created)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *ActivityServiceTestSuite) TestGetActivitySuccess() {
	s.T().Run("GetActivity - Success", func(t *testing.T) {
		// Setup test data
		activityID := s.activityID
		now := time.Now()
		deadline := now.Add(24 * time.Hour)
		expectedActivity := types.Activity{
			ID:             activityID,
			OrganizationID: s.orgID,
			ActivityType:   types.ActivityTypeCall,
			Summary:        "Client Call",
			Note:           stringPtr("Follow up on proposal"),
			DateDeadline:   &deadline,
			UserID:         &s.userID,
			AssignedTo:     &s.userID,
			ResModel:       stringPtr("contacts"),
			ResID:          &s.contactID,
			State:          types.ActivityStatePlanned,
		}

		// Mock repository behavior
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
			require.Equal(t, activityID, id)
			return &expectedActivity, nil
		})

		// Execute
		activity, err := s.service.GetActivity(s.ctx, activityID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, activity)
		require.Equal(t, expectedActivity.ID, activity.ID)
		require.Equal(t, expectedActivity.OrganizationID, activity.OrganizationID)
		require.Equal(t, expectedActivity.ActivityType, activity.ActivityType)
		require.Equal(t, expectedActivity.Summary, activity.Summary)
		require.Equal(t, expectedActivity.Note, activity.Note)
		require.Equal(t, expectedActivity.DateDeadline, activity.DateDeadline)
		require.Equal(t, expectedActivity.UserID, activity.UserID)
		require.Equal(t, expectedActivity.AssignedTo, activity.AssignedTo)
		require.Equal(t, expectedActivity.ResModel, activity.ResModel)
		require.Equal(t, expectedActivity.ResID, activity.ResID)
		require.Equal(t, expectedActivity.State, activity.State)
	})
}

func (s *ActivityServiceTestSuite) TestGetActivityNotFound() {
	s.T().Run("GetActivity - Not Found", func(t *testing.T) {
		// Setup test data
		activityID := s.activityID

		// Mock repository behavior - return error
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
			require.Equal(t, activityID, id)
			return nil, nil // Not found
		})

		// Execute
		activity, err := s.service.GetActivity(s.ctx, activityID)

		// Assert
		require.Error(t, err)
		require.Nil(t, activity)
		require.Contains(t, err.Error(), "activity not found")
	})
}

func (s *ActivityServiceTestSuite) TestUpdateActivitySuccess() {
	s.T().Run("UpdateActivity - Success", func(t *testing.T) {
		// Setup test data
		activityID := s.activityID
		newSummary := "Updated Meeting"
		newState := types.ActivityStateDone
		doneDate := time.Now()

		request := types.ActivityUpdateRequest{
			Summary:  stringPtr(newSummary),
			State:    &newState,
			DoneDate: &doneDate,
		}

		// Mock repository behavior
		existingActivity := types.Activity{
			ID:             activityID,
			OrganizationID: s.orgID,
			ActivityType:   types.ActivityTypeMeeting,
			Summary:        "Original Meeting",
			State:          types.ActivityStatePlanned,
		}

		expectedActivity := types.Activity{
			ID:             activityID,
			OrganizationID: s.orgID,
			ActivityType:   types.ActivityTypeMeeting,
			Summary:        newSummary,
			State:          newState,
			DoneDate:       &doneDate,
			UpdatedAt:      time.Now(),
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
			require.Equal(t, activityID, id)
			return &existingActivity, nil
		})

		s.repo.WithUpdateFunc(func(ctx context.Context, activity types.Activity) (*types.Activity, error) {
			require.Equal(t, activityID, activity.ID)
			require.Equal(t, s.orgID, activity.OrganizationID)
			require.Equal(t, newSummary, activity.Summary)
			require.Equal(t, newState, activity.State)
			require.Equal(t, &doneDate, activity.DoneDate)
			return &expectedActivity, nil
		})

		// Execute
		updated, err := s.service.UpdateActivity(s.ctx, activityID, request)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, expectedActivity.ID, updated.ID)
		require.Equal(t, expectedActivity.OrganizationID, updated.OrganizationID)
		require.Equal(t, expectedActivity.Summary, updated.Summary)
		require.Equal(t, expectedActivity.State, updated.State)
		require.Equal(t, expectedActivity.DoneDate, updated.DoneDate)
		require.NotZero(t, updated.UpdatedAt)
	})
}

func (s *ActivityServiceTestSuite) TestDeleteActivitySuccess() {
	s.T().Run("DeleteActivity - Success", func(t *testing.T) {
		// Setup test data
		activityID := s.activityID

		// Mock repository behavior
		existingActivity := types.Activity{
			ID:             activityID,
			OrganizationID: s.orgID,
			Summary:        "Activity to Delete",
			State:          types.ActivityStatePlanned,
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
			require.Equal(t, activityID, id)
			return &existingActivity, nil
		})

		s.repo.WithDeleteFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, activityID, id)
			return nil
		})

		// Execute
		err := s.service.DeleteActivity(s.ctx, activityID)

		// Assert
		require.NoError(t, err)
	})
}

func (s *ActivityServiceTestSuite) TestListActivitiesSuccess() {
	s.T().Run("ListActivities - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ActivityFilter{
			ActivityType: activityTypePtr(types.ActivityTypeMeeting),
			State:        activityStatePtr(types.ActivityStatePlanned),
			Limit:        10,
		}

		// Mock repository behavior
		now := time.Now()
		deadline := now.Add(24 * time.Hour)
		expectedActivities := []*types.Activity{
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				ActivityType:   types.ActivityTypeMeeting,
				Summary:        "Meeting 1",
				DateDeadline:   &deadline,
				State:          types.ActivityStatePlanned,
			},
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				ActivityType:   types.ActivityTypeMeeting,
				Summary:        "Meeting 2",
				DateDeadline:   &deadline,
				State:          types.ActivityStatePlanned,
			},
		}

		s.repo.WithFindAllFunc(func(ctx context.Context, f types.ActivityFilter) ([]*types.Activity, error) {
			require.Equal(t, s.orgID, f.OrganizationID)
			require.Equal(t, filter.ActivityType, f.ActivityType)
			require.Equal(t, filter.State, f.State)
			require.Equal(t, filter.Limit, f.Limit)
			return expectedActivities, nil
		})

		// Execute
		activities, err := s.service.ListActivities(s.ctx, filter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, activities)
		require.Len(t, activities, 2)
		require.Equal(t, expectedActivities[0].Summary, activities[0].Summary)
		require.Equal(t, expectedActivities[1].Summary, activities[1].Summary)
	})
}

func (s *ActivityServiceTestSuite) TestCompleteActivitySuccess() {
	s.T().Run("CompleteActivity - Success", func(t *testing.T) {
		// Setup test data
		activityID := s.activityID
		doneDate := time.Now()

		// Mock repository behavior
		existingActivity := types.Activity{
			ID:             activityID,
			OrganizationID: s.orgID,
			ActivityType:   types.ActivityTypeCall,
			Summary:        "Follow up call",
			State:          types.ActivityStatePlanned,
		}

		expectedActivity := types.Activity{
			ID:             activityID,
			OrganizationID: s.orgID,
			ActivityType:   types.ActivityTypeCall,
			Summary:        "Follow up call",
			State:          types.ActivityStateDone,
			DoneDate:       &doneDate,
			UpdatedAt:      time.Now(),
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
			require.Equal(t, activityID, id)
			return &existingActivity, nil
		})

		s.repo.WithUpdateFunc(func(ctx context.Context, activity types.Activity) (*types.Activity, error) {
			require.Equal(t, activityID, activity.ID)
			require.Equal(t, types.ActivityStateDone, activity.State)
			require.NotNil(t, activity.DoneDate)
			return &expectedActivity, nil
		})

		// Execute
		completed, err := s.service.CompleteActivity(s.ctx, activityID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, completed)
		require.Equal(t, expectedActivity.ID, completed.ID)
		require.Equal(t, types.ActivityStateDone, completed.State)
		require.Equal(t, &doneDate, completed.DoneDate)
		require.NotZero(t, completed.UpdatedAt)
	})
}

// Run the test suite
func TestActivityServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ActivityServiceTestSuite))
}
