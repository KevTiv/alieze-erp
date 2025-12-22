package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/events"

	"github.com/google/uuid"
)

// ActivityService handles activity business logic
type ActivityService struct {
	repo        types.ActivityRepository
	authService auth.LegacyAuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

func NewActivityService(repo types.ActivityRepository, authService auth.LegacyAuthService, eventBus *events.Bus) *ActivityService {
	return &ActivityService{
		repo:        repo,
		authService: authService,
		eventBus:    eventBus,
		logger:      slog.Default().With("service", "activity"),
	}
}

func (s *ActivityService) CreateActivity(ctx context.Context, req types.ActivityCreateRequest) (*types.Activity, error) {
	// Validation
	if err := s.validateActivity(req); err != nil {
		return nil, fmt.Errorf("invalid activity: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:activities:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Set user
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate ID
	activityID := uuid.New()

	// Create activity
	activity := types.Activity{
		ID:             activityID,
		OrganizationID: orgID,
		ActivityType:   req.ActivityType,
		Summary:        req.Summary,
		Note:           req.Note,
		DateDeadline:   req.DateDeadline,
		UserID:         req.UserID,
		AssignedTo:     req.AssignedTo,
		ResModel:       req.ResModel,
		ResID:          req.ResID,
		State:          req.State,
		DoneDate:       req.DoneDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
	}

	created, err := s.repo.Create(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.activity.created", created)

	s.logger.Info("Created activity", "activity_id", created.ID, "type", created.ActivityType, "summary", created.Summary)

	return created, nil
}

func (s *ActivityService) GetActivity(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:activities:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the activity
	activity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if activity.OrganizationID != orgID {
		return nil, fmt.Errorf("activity does not belong to organization: %w", errors.New("access denied"))
	}

	return activity, nil
}

func (s *ActivityService) ListActivities(ctx context.Context, filter types.ActivityFilter) ([]*types.Activity, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:activities:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization filter
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// List activities
	activities, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	return activities, nil
}

func (s *ActivityService) UpdateActivity(ctx context.Context, id uuid.UUID, req types.ActivityUpdateRequest) (*types.Activity, error) {
	// Validation
	if err := s.validateActivityUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid activity update: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:activities:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing activity to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing activity: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("activity does not belong to organization: %w", errors.New("access denied"))
	}

	// Set user
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Build update
	activity := types.Activity{
		ID:             id,
		OrganizationID: orgID,
		ActivityType:   *req.ActivityType,
		Summary:        *req.Summary,
		Note:           req.Note,
		DateDeadline:   req.DateDeadline,
		UserID:         req.UserID,
		AssignedTo:     req.AssignedTo,
		ResModel:       req.ResModel,
		ResID:          req.ResID,
		State:          *req.State,
		DoneDate:       req.DoneDate,
		UpdatedAt:      time.Now(),
		UpdatedBy:      &userID,
	}

	// Update
	updated, err := s.repo.Update(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.activity.updated", updated)

	s.logger.Info("Updated activity", "activity_id", updated.ID, "type", updated.ActivityType)

	return updated, nil
}

func (s *ActivityService) CompleteActivity(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:activities:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing activity to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing activity: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("activity does not belong to organization: %w", errors.New("access denied"))
	}

	// Set user
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Complete the activity
	now := time.Now()
	activity := types.Activity{
		ID:             id,
		OrganizationID: orgID,
		ActivityType:   existing.ActivityType,
		Summary:        existing.Summary,
		Note:           existing.Note,
		DateDeadline:   existing.DateDeadline,
		UserID:         existing.UserID,
		AssignedTo:     existing.AssignedTo,
		ResModel:       existing.ResModel,
		ResID:          existing.ResID,
		State:          types.ActivityStateDone,
		DoneDate:       &now,
		UpdatedAt:      now,
		UpdatedBy:      &userID,
	}

	// Update
	updated, err := s.repo.Update(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to complete activity: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.activity.completed", updated)

	s.logger.Info("Completed activity", "activity_id", updated.ID, "type", updated.ActivityType)

	return updated, nil
}

func (s *ActivityService) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:activities:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing activity to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing activity: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("activity does not belong to organization: %w", errors.New("access denied"))
	}

	// Delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.activity.deleted", existing)

	s.logger.Info("Deleted activity", "activity_id", id)

	return nil
}

func (s *ActivityService) GetActivitiesByContact(ctx context.Context, contactID uuid.UUID) ([]*types.Activity, error) {
	// Permission check - verify contact access
	if err := s.authService.CheckPermission(ctx, "crm:contacts:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access for contact
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get activities by contact
	activities, err := s.repo.FindByContact(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by contact: %w", err)
	}

	// Filter by organization
	var filteredActivities []*types.Activity
	for _, activity := range activities {
		if activity.OrganizationID == orgID {
			filteredActivities = append(filteredActivities, activity)
		}
	}

	return filteredActivities, nil
}

func (s *ActivityService) GetActivitiesByLead(ctx context.Context, leadID uuid.UUID) ([]*types.Activity, error) {
	// Permission check - verify lead access
	if err := s.authService.CheckPermission(ctx, "crm:leads:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access for lead
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get activities by lead
	activities, err := s.repo.FindByLead(ctx, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by lead: %w", err)
	}

	// Filter by organization
	var filteredActivities []*types.Activity
	for _, activity := range activities {
		if activity.OrganizationID == orgID {
			filteredActivities = append(filteredActivities, activity)
		}
	}

	return filteredActivities, nil
}

func (s *ActivityService) validateActivity(req types.ActivityCreateRequest) error {
	if req.Summary == "" {
		return errors.New("summary is required")
	}

	if len(req.Summary) > 255 {
		return errors.New("summary must be 255 characters or less")
	}

	if req.Note != nil && len(*req.Note) > 10000 {
		return errors.New("note must be 10000 characters or less")
	}

	// Validate activity type
	switch req.ActivityType {
	case types.ActivityTypeCall, types.ActivityTypeMeeting, types.ActivityTypeEmail, types.ActivityTypeTodo, types.ActivityTypeNote:
		// Valid
	default:
		return errors.New("invalid activity type")
	}

	// Validate state
	switch req.State {
	case types.ActivityStatePlanned, types.ActivityStateDone, types.ActivityStateCancelled:
		// Valid
	default:
		return errors.New("invalid activity state")
	}

	return nil
}

func (s *ActivityService) validateActivityUpdate(req types.ActivityUpdateRequest) error {
	if req.Summary != nil {
		if *req.Summary == "" {
			return errors.New("summary cannot be empty")
		}

		if len(*req.Summary) > 255 {
			return errors.New("summary must be 255 characters or less")
		}
	}

	if req.Note != nil && len(*req.Note) > 10000 {
		return errors.New("note must be 10000 characters or less")
	}

	// Validate activity type
	if req.ActivityType != nil {
		switch *req.ActivityType {
		case types.ActivityTypeCall, types.ActivityTypeMeeting, types.ActivityTypeEmail, types.ActivityTypeTodo, types.ActivityTypeNote:
			// Valid
		default:
			return errors.New("invalid activity type")
		}
	}

	// Validate state
	if req.State != nil {
		switch *req.State {
		case types.ActivityStatePlanned, types.ActivityStateDone, types.ActivityStateCancelled:
			// Valid
		default:
			return errors.New("invalid activity state")
		}
	}

	return nil
}
