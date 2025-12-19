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

// LeadSourceService handles lead source business logic
type LeadSourceService struct {
	repo        types.LeadSourceRepository
	authService AuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

func NewLeadSourceService(repo types.LeadSourceRepository, authService AuthService, eventBus *events.Bus) *LeadSourceService {
	return &LeadSourceService{
		repo:        repo,
		authService: authService,
		eventBus:    eventBus,
		logger:      slog.Default().With("service", "lead-source"),
	}
}

func (s *LeadSourceService) CreateLeadSource(ctx context.Context, req types.LeadSourceCreateRequest) (*types.LeadSource, error) {
	// Validation
	if err := s.validateLeadSource(req); err != nil {
		return nil, fmt.Errorf("invalid lead source: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_sources:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Generate ID
	sourceID := uuid.New()

	// Create source
	source := types.LeadSource{
		ID:             sourceID,
		OrganizationID: orgID,
		Name:           req.Name,
		CreatedAt:      time.Now(),
	}

	created, err := s.repo.Create(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead source: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lead_source.created", created)

	s.logger.Info("Created lead source", "source_id", created.ID, "name", created.Name)

	return created, nil
}

func (s *LeadSourceService) GetLeadSource(ctx context.Context, id uuid.UUID) (*types.LeadSource, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_sources:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the source
	source, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead source: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if source.OrganizationID != orgID {
		return nil, fmt.Errorf("lead source does not belong to organization: %w", errors.New("access denied"))
	}

	return source, nil
}

func (s *LeadSourceService) ListLeadSources(ctx context.Context, filter types.LeadSourceFilter) ([]*types.LeadSource, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_sources:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization filter
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// List sources
	sources, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list lead sources: %w", err)
	}

	return sources, nil
}

func (s *LeadSourceService) UpdateLeadSource(ctx context.Context, id uuid.UUID, req types.LeadSourceUpdateRequest) (*types.LeadSource, error) {
	// Validation
	if err := s.validateLeadSourceUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid lead source update: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_sources:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing source to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing lead source: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("lead source does not belong to organization: %w", errors.New("access denied"))
	}

	// Build update
	source := types.LeadSource{
		ID:             id,
		OrganizationID: orgID,
		Name:           *req.Name,
		CreatedAt:      existing.CreatedAt, // Keep original created date
	}

	// Update
	updated, err := s.repo.Update(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to update lead source: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lead_source.updated", updated)

	s.logger.Info("Updated lead source", "source_id", updated.ID, "name", updated.Name)

	return updated, nil
}

func (s *LeadSourceService) DeleteLeadSource(ctx context.Context, id uuid.UUID) error {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:lead_sources:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing source to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing lead source: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("lead source does not belong to organization: %w", errors.New("access denied"))
	}

	// Delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead source: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.lead_source.deleted", existing)

	s.logger.Info("Deleted lead source", "source_id", id)

	return nil
}

func (s *LeadSourceService) validateLeadSource(req types.LeadSourceCreateRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}

	return nil
}

func (s *LeadSourceService) validateLeadSourceUpdate(req types.LeadSourceUpdateRequest) error {
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
