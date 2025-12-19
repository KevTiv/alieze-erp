package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/pkg/events"

	"github.com/google/uuid"
)

// LeadStageService handles lead stage business logic
type LeadStageService struct {
	repo        types.LeadStageRepository
	authService AuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

func NewLeadStageService(repo types.LeadStageRepository, authService AuthService, eventBus *events.Bus) *LeadStageService {
	return &LeadStageService{
		repo:        repo,
		authService: authService,
		eventBus:    eventBus,
		logger:      slog.Default().With("service", "lead-stage"),
	}
}

func (s *LeadStageService) CreateLeadStage(ctx context.Context, req types.LeadStageCreateRequest) (*types.LeadStage, error) {
	// Validation
	if err := s.validateLeadStage(req); err != nil {
		return nil, fmt.Errorf("invalid lead stage: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_stages:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Generate ID
	stageID := uuid.New()

	// Create stage
	stage := types.LeadStage{
		ID:             stageID,
		OrganizationID: orgID,
		Name:           req.Name,
		Sequence:       req.Sequence,
		Probability:    req.Probability,
		Fold:           req.Fold,
		IsWon:          req.IsWon,
		Requirements:   req.Requirements,
		TeamID:         req.TeamID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	created, err := s.repo.Create(ctx, stage)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead stage: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lead_stage.created", created)

	s.logger.Info("Created lead stage", "stage_id", created.ID, "name", created.Name)

	return created, nil
}

func (s *LeadStageService) GetLeadStage(ctx context.Context, id uuid.UUID) (*types.LeadStage, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_stages:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the stage
	stage, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead stage: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if stage.OrganizationID != orgID {
		return nil, fmt.Errorf("lead stage does not belong to organization: %w", errors.New("access denied"))
	}

	return stage, nil
}

func (s *LeadStageService) ListLeadStages(ctx context.Context, filter types.LeadStageFilter) ([]types.LeadStage, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_stages:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization filter
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// List stages
	stages, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list lead stages: %w", err)
	}

	return stages, nil
}

func (s *LeadStageService) UpdateLeadStage(ctx context.Context, id uuid.UUID, req types.LeadStageUpdateRequest) (*types.LeadStage, error) {
	// Validation
	if err := s.validateLeadStageUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid lead stage update: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_stages:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing stage to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing lead stage: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("lead stage does not belong to organization: %w", errors.New("access denied"))
	}

	// Build update
	stage := types.LeadStage{
		ID:             id,
		OrganizationID: orgID,
		Name:           *req.Name,
		Sequence:       *req.Sequence,
		Probability:    *req.Probability,
		Fold:           *req.Fold,
		IsWon:          *req.IsWon,
		Requirements:   req.Requirements,
		TeamID:         req.TeamID,
		UpdatedAt:      time.Now(),
	}

	// Update
	updated, err := s.repo.Update(ctx, stage)
	if err != nil {
		return nil, fmt.Errorf("failed to update lead stage: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lead_stage.updated", updated)

	s.logger.Info("Updated lead stage", "stage_id", updated.ID, "name", updated.Name)

	return updated, nil
}

func (s *LeadStageService) DeleteLeadStage(ctx context.Context, id uuid.UUID) error {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_stages:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing stage to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing lead stage: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("lead stage does not belong to organization: %w", errors.New("access denied"))
	}

	// Delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead stage: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lead_stage.deleted", existing)

	s.logger.Info("Deleted lead stage", "stage_id", id)

	return nil
}

func (s *LeadStageService) validateLeadStage(req types.LeadStageCreateRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}

	if req.Probability < 0 || req.Probability > 100 {
		return errors.New("probability must be between 0 and 100")
	}

	if req.Sequence < 0 {
		return errors.New("sequence must be a positive number")
	}

	if req.Requirements != nil && len(*req.Requirements) > 10000 {
		return errors.New("requirements must be 10000 characters or less")
	}

	return nil
}

func (s *LeadStageService) validateLeadStageUpdate(req types.LeadStageUpdateRequest) error {
	if req.Name != nil {
		if *req.Name == "" {
			return errors.New("name cannot be empty")
		}

		if len(*req.Name) > 100 {
			return errors.New("name must be 100 characters or less")
		}
	}

	if req.Probability != nil {
		if *req.Probability < 0 || *req.Probability > 100 {
			return errors.New("probability must be between 0 and 100")
		}
	}

	if req.Sequence != nil {
		if *req.Sequence < 0 {
			return errors.New("sequence must be a positive number")
		}
	}

	if req.Requirements != nil && len(*req.Requirements) > 10000 {
		return errors.New("requirements must be 10000 characters or less")
	}

	return nil
}
