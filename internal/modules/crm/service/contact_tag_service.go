package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/repository"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/pkg/events"
)

// ContactTagService handles contact tag business logic
type ContactTagService struct {
	repo        repository.ContactTagRepository
	authService AuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

func NewContactTagService(repo repository.ContactTagRepository, authService AuthService, eventBus *events.Bus) *ContactTagService {
	return &ContactTagService{
		repo:        repo,
		authService: authService,
		eventBus:    eventBus,
		logger:      slog.Default().With("service", "contact-tag"),
	}
}

func (s *ContactTagService) CreateContactTag(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error) {
	// Validation
	if err := s.validateContactTag(tag); err != nil {
		return nil, fmt.Errorf("invalid contact tag: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:contact_tags:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	tag.OrganizationID = orgID

	// Generate ID if not provided
	if tag.ID == uuid.Nil {
		tag.ID = uuid.New()
	}

	// Create
	created, err := s.repo.Create(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact tag: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.contact_tag.created", created)

	s.logger.Info("Created contact tag", "tag_id", created.ID, "name", created.Name)

	return created, nil
}

func (s *ContactTagService) GetContactTag(ctx context.Context, id uuid.UUID) (*types.ContactTag, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:contact_tags:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the tag
	tag, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact tag: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if tag.OrganizationID != orgID {
		return nil, fmt.Errorf("contact tag does not belong to organization: %w", errors.New("access denied"))
	}

	return tag, nil
}

func (s *ContactTagService) ListContactTags(ctx context.Context, filter types.ContactTagFilter) ([]types.ContactTag, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:contact_tags:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization filter
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// List tags
	tags, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact tags: %w", err)
	}

	return tags, nil
}

func (s *ContactTagService) UpdateContactTag(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error) {
	// Validation
	if err := s.validateContactTag(tag); err != nil {
		return nil, fmt.Errorf("invalid contact tag: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:contact_tags:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing tag to verify organization
	existing, err := s.repo.FindByID(ctx, tag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing contact tag: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("contact tag does not belong to organization: %w", errors.New("access denied"))
	}

	// Update
	updated, err := s.repo.Update(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact tag: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.contact_tag.updated", updated)

	s.logger.Info("Updated contact tag", "tag_id", updated.ID, "name", updated.Name)

	return updated, nil
}

func (s *ContactTagService) DeleteContactTag(ctx context.Context, id uuid.UUID) error {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:contact_tags:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing tag to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing contact tag: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("contact tag does not belong to organization: %w", errors.New("access denied"))
	}

	// Delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact tag: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.contact_tag.deleted", existing)

	s.logger.Info("Deleted contact tag", "tag_id", id)

	return nil
}



func (s *ContactTagService) validateContactTag(tag types.ContactTag) error {
	if tag.Name == "" {
		return errors.New("name is required")
	}

	if len(tag.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}

	if tag.Color < 0 || tag.Color > 16777215 {
		return errors.New("color must be a valid RGB value")
	}

	return nil
}
