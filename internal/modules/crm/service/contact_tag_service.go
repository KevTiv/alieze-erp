package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/crm/errors"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/google/uuid"
)

// ContactTagService handles contact tag operations
type ContactTagService struct {
	contactRepo types.ContactRepository
	authService auth.BaseAuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

// NewContactTagService creates a new tag service
func NewContactTagService(
	contactRepo types.ContactRepository,
	authService auth.BaseAuthService,
	eventBus *events.Bus,
	logger *slog.Logger,
) *ContactTagService {
	return &ContactTagService{
		contactRepo: contactRepo,
		authService: authService,
		eventBus:    eventBus,
		logger:      logger,
	}
}

// AddTags adds tags to a contact
func (s *ContactTagService) AddTags(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, req types.ContactTagRequest) (*types.ContactTagResponse, error) {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Get existing contact
	contact, err := s.contactRepo.FindByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	if contact == nil {
		return nil, errors.ErrNotFound
	}

	// Verify organization
	if contact.OrganizationID != orgID {
		return nil, errors.ErrOrganizationAccess
	}

	// Get current tags (from metadata or custom implementation)
	currentTags := s.getCurrentTags(contact)

	// Add new tags (avoid duplicates)
	addedTags := []string{}
	for _, tag := range req.Tags {
		if !contains(currentTags, tag) {
			currentTags = append(currentTags, tag)
			addedTags = append(addedTags, tag)
		}
	}

	// Update contact with new tags
	if len(addedTags) > 0 {
		err = s.updateContactTags(ctx, orgID, contactID, currentTags)
		if err != nil {
			return nil, fmt.Errorf("failed to update contact tags: %w", err)
		}

		// Publish event
		if s.eventBus != nil {
			s.eventBus.Publish(ctx, "contact.tags.added", map[string]interface{}{
				"contact_id":      contactID,
				"organization_id": orgID,
				"tags":            addedTags,
			})
		}

		s.logger.Info("tags added to contact",
			"contact_id", contactID,
			"org_id", orgID,
			"added_count", len(addedTags),
		)
	}

	return &types.ContactTagResponse{
		ContactID: contactID,
		Tags:      currentTags,
		Added:     addedTags,
	}, nil
}

// RemoveTag removes a specific tag from a contact
func (s *ContactTagService) RemoveTag(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, tag string) error {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return errors.ErrOrganizationAccess
	}

	// Get existing contact
	contact, err := s.contactRepo.FindByID(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}
	if contact == nil {
		return errors.ErrNotFound
	}

	// Verify organization
	if contact.OrganizationID != orgID {
		return errors.ErrOrganizationAccess
	}

	// Get current tags
	currentTags := s.getCurrentTags(contact)

	// Remove tag
	newTags := []string{}
	removed := false
	for _, t := range currentTags {
		if t != tag {
			newTags = append(newTags, t)
		} else {
			removed = true
		}
	}

	if !removed {
		return fmt.Errorf("tag not found on contact")
	}

	// Update contact
	err = s.updateContactTags(ctx, orgID, contactID, newTags)
	if err != nil {
		return fmt.Errorf("failed to update contact tags: %w", err)
	}

	// Publish event
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "contact.tag.removed", map[string]interface{}{
			"contact_id":      contactID,
			"organization_id": orgID,
			"tag":             tag,
		})
	}

	s.logger.Info("tag removed from contact",
		"contact_id", contactID,
		"org_id", orgID,
		"tag", tag,
	)

	return nil
}

// ListTags returns all tags used in the organization
func (s *ContactTagService) ListTags(ctx context.Context, orgID uuid.UUID) ([]string, error) {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Get all contacts for organization
	filter := types.ContactFilter{
		OrganizationID: orgID,
		Limit:          10000, // Large limit to get all contacts
	}

	contacts, err := s.contactRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	// Collect unique tags
	tagSet := make(map[string]bool)
	for _, contact := range contacts {
		tags := s.getCurrentTags(contact)
		for _, tag := range tags {
			tagSet[tag] = true
		}
	}

	// Convert to slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}

// getCurrentTags extracts tags from contact metadata
func (s *ContactTagService) getCurrentTags(contact *types.Contact) []string {
	// Tags can be stored in metadata JSONB field
	// For now, return empty slice - will be enhanced when contact metadata is implemented
	// TODO: Extract from contact.Metadata["tags"] once metadata field is available
	return []string{}
}

// updateContactTags updates contact tags in database
func (s *ContactTagService) updateContactTags(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, tags []string) error {
	// TODO: Update contact metadata with new tags
	// Use the repository method to add tags
	return s.contactRepo.AddContactTags(ctx, orgID, contactID, tags)
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}
