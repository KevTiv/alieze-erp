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

// LostReasonService handles lost reason business logic
type LostReasonService struct {
	repo        types.LostReasonRepository
	authService AuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

func NewLostReasonService(repo types.LostReasonRepository, authService AuthService, eventBus *events.Bus) *LostReasonService {
	return &LostReasonService{
		repo:        repo,
		authService: authService,
		eventBus:    eventBus,
		logger:      slog.Default().With("service", "lost-reason"),
	}
}

func (s *LostReasonService) CreateLostReason(ctx context.Context, req types.LostReasonCreateRequest) (*types.LostReason, error) {
	// Validation
	if err := s.validateLostReason(req); err != nil {
		return nil, fmt.Errorf("invalid lost reason: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lost_reasons:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Generate ID
	reasonID := uuid.New()

	// Create reason
	reason := types.LostReason{
		ID:             reasonID,
		OrganizationID: orgID,
		Name:           req.Name,
		Active:         req.Active,
		CreatedAt:      time.Now(),
	}

	created, err := s.repo.Create(ctx, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to create lost reason: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lost_reason.created", created)

	s.logger.Info("Created lost reason", "reason_id", created.ID, "name", created.Name)

	return created, nil
}

func (s *LostReasonService) GetLostReason(ctx context.Context, id uuid.UUID) (*types.LostReason, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lost_reasons:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the reason
	reason, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lost reason: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if reason.OrganizationID != orgID {
		return nil, fmt.Errorf("lost reason does not belong to organization: %w", errors.New("access denied"))
	}

	return reason, nil
}

func (s *LostReasonService) ListLostReasons(ctx context.Context, filter types.LostReasonFilter) ([]types.LostReason, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lost_reasons:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization filter
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// List reasons
	reasons, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list lost reasons: %w", err)
	}

	return reasons, nil
}

func (s *LostReasonService) UpdateLostReason(ctx context.Context, id uuid.UUID, req types.LostReasonUpdateRequest) (*types.LostReason, error) {
	// Validation
	if err := s.validateLostReasonUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid lost reason update: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lost_reasons:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing reason to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing lost reason: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("lost reason does not belong to organization: %w", errors.New("access denied"))
	}

	// Build update
	reason := types.LostReason{
		ID:             id,
		OrganizationID: orgID,
		Name:           *req.Name,
		Active:         *req.Active,
		CreatedAt:      existing.CreatedAt, // Keep original created date
	}

	// Update
	updated, err := s.repo.Update(ctx, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to update lost reason: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lost_reason.updated", updated)

	s.logger.Info("Updated lost reason", "reason_id", updated.ID, "name", updated.Name)

	return updated, nil
}

func (s *LostReasonService) DeleteLostReason(ctx context.Context, id uuid.UUID) error {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lost_reasons:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing reason to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing lost reason: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("lost reason does not belong to organization: %w", errors.New("access denied"))
	}

	// Delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete lost reason: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lost_reason.deleted", existing)

	s.logger.Info("Deleted lost reason", "reason_id", id)

	return nil
}

func (s *LostReasonService) validateLostReason(req types.LostReasonCreateRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}

	return nil
}

func (s *LostReasonService) validateLostReasonUpdate(req types.LostReasonUpdateRequest) error {
	if req.Name != nil {
		if *req.Name == "" {
			return errors.New("name cannot be empty")
		}

		if len(*req.Name) > 100 {
			return errors.New("name must be 100 characters or less")
		}
	}

	return nil
}
