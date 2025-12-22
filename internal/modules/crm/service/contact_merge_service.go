package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/events"
)

// ContactMergeService handles duplicate detection and contact merging
type ContactMergeService struct {
	repo           repository.ContactMergeRepository
	contactRepo    types.ContactRepository
	authService    auth.AuthorizationService
	eventPublisher events.EventPublisher
}

func NewContactMergeService(
	repo repository.ContactMergeRepository,
	contactRepo types.ContactRepository,
	authService auth.AuthorizationService,
	eventPublisher events.EventPublisher,
) *ContactMergeService {
	return &ContactMergeService{
		repo:           repo,
		contactRepo:    contactRepo,
		authService:    authService,
		eventPublisher: eventPublisher,
	}
}

// DetectDuplicates initiates duplicate detection for an organization
func (s *ContactMergeService) DetectDuplicates(ctx context.Context, orgID uuid.UUID, threshold *int, limit *int) (*types.DetectDuplicatesResponse, error) {
	// Set default threshold if not provided
	actualThreshold := 80
	if threshold != nil {
		actualThreshold = *threshold
	}

	// Set default limit
	actualLimit := 100
	if limit != nil {
		actualLimit = *limit
	}

	// Find potential duplicates
	duplicates, err := s.repo.FindPotentialDuplicates(ctx, orgID, actualThreshold, actualLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}

	// Store detected duplicates in database
	created := 0
	for _, dup := range duplicates {
		// Check if this duplicate pair already exists
		err := s.repo.CreateDuplicate(ctx, dup)
		if err != nil {
			// Skip if already exists, otherwise return error
			continue
		}
		created++

		// Publish event
		s.eventPublisher.Publish(ctx, "contact.duplicate.detected", map[string]interface{}{
			"organization_id":  orgID.String(),
			"duplicate_id":     dup.ID.String(),
			"contact1_id":      dup.ContactID1.String(),
			"contact2_id":      dup.ContactID2.String(),
			"similarity_score": dup.SimilarityScore,
			"matching_fields":  dup.MatchingFields,
		})
	}

	return &types.DetectDuplicatesResponse{
		TotalFound:   len(duplicates),
		TotalCreated: created,
		Threshold:    actualThreshold,
		Duplicates:   duplicates,
	}, nil
}

// CalculateSimilarity calculates similarity score between two contacts
func (s *ContactMergeService) CalculateSimilarity(ctx context.Context, orgID uuid.UUID, contact1ID uuid.UUID, contact2ID uuid.UUID) (*types.CalculateSimilarityResponse, error) {
	// Get both contacts
	contact1, err := s.contactRepo.FindByID(ctx, contact1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact1: %w", err)
	}

	contact2, err := s.contactRepo.FindByID(ctx, contact2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact2: %w", err)
	}

	// Verify both contacts belong to the organization
	if contact1.OrganizationID != orgID || contact2.OrganizationID != orgID {
		return nil, fmt.Errorf("contacts do not belong to the specified organization")
	}

	// Calculate similarity
	score := s.repo.CalculateSimilarity(contact1, contact2)

	return &types.CalculateSimilarityResponse{
		Contact1ID:      contact1ID,
		Contact2ID:      contact2ID,
		SimilarityScore: float64(score),
		IsDuplicate:     score >= 80,
	}, nil
}

// MergeContacts merges two contacts
func (s *ContactMergeService) MergeContacts(ctx context.Context, orgID uuid.UUID, masterContactID uuid.UUID, mergeContactID uuid.UUID, strategy string, fieldSelections map[string]string) error {
	// Validate that master and duplicate are different contacts
	if masterContactID == mergeContactID {
		return fmt.Errorf("master and merge contact IDs must be different")
	}

	// Get both contacts to verify they exist and belong to the organization
	masterContact, err := s.contactRepo.FindByID(ctx, masterContactID)
	if err != nil {
		return fmt.Errorf("failed to get master contact: %w", err)
	}

	mergeContact, err := s.contactRepo.FindByID(ctx, mergeContactID)
	if err != nil {
		return fmt.Errorf("failed to get merge contact: %w", err)
	}

	// Verify both contacts belong to the organization
	if masterContact.OrganizationID != orgID || mergeContact.OrganizationID != orgID {
		return fmt.Errorf("contacts do not belong to the specified organization")
	}

	// Build field selections based on strategy
	actualFieldSelections := s.buildFieldSelections(strategy, fieldSelections)

	// Perform merge
	userID := uuid.Nil // In production, this would come from context
	err = s.repo.MergeContacts(ctx, masterContactID, mergeContactID, actualFieldSelections, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to merge contacts: %w", err)
	}

	// Publish event
	s.eventPublisher.Publish(ctx, "contact.merged", map[string]interface{}{
		"organization_id":   orgID.String(),
		"master_contact_id": masterContactID.String(),
		"merge_contact_id":  mergeContactID.String(),
		"merge_strategy":    strategy,
	})

	return nil
}

// ResolveDuplicate marks a duplicate as resolved without merging
func (s *ContactMergeService) ResolveDuplicate(ctx context.Context, orgID uuid.UUID, duplicateID uuid.UUID, resolutionType string) error {
	// Get duplicate to verify it exists and belongs to organization
	duplicate, err := s.repo.GetDuplicate(ctx, duplicateID)
	if err != nil {
		return fmt.Errorf("failed to get duplicate: %w", err)
	}

	if duplicate.OrganizationID != orgID {
		return fmt.Errorf("duplicate does not belong to the specified organization")
	}

	// Validate resolution type
	validTypes := map[string]bool{
		"false_positive": true,
		"ignore":         true,
		"merged":         true,
	}
	if !validTypes[resolutionType] {
		return fmt.Errorf("invalid resolution type: %s", resolutionType)
	}

	// Update status
	userID := uuid.Nil // In production, this would come from context
	err = s.repo.UpdateDuplicateStatus(ctx, duplicateID, "resolved", userID)
	if err != nil {
		return fmt.Errorf("failed to update duplicate status: %w", err)
	}

	// Publish event
	s.eventPublisher.Publish(ctx, "contact.duplicate.resolved", map[string]interface{}{
		"organization_id": orgID.String(),
		"duplicate_id":    duplicateID.String(),
		"resolution_type": resolutionType,
	})

	return nil
}

// ListDuplicates retrieves a list of potential duplicates
func (s *ContactMergeService) ListDuplicates(ctx context.Context, orgID uuid.UUID, status *string, limit *int, offset *int) (*types.ListDuplicatesResponse, error) {
	// Create filter
	filter := types.DuplicateFilter{
		OrganizationID: orgID,
		Status:         status,
	}

	// Get duplicates
	duplicates, err := s.repo.ListDuplicates(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list duplicates: %w", err)
	}

	// Get total count
	total, err := s.repo.CountDuplicates(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count duplicates: %w", err)
	}

	return &types.ListDuplicatesResponse{
		Duplicates: duplicates,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// GetMergeHistory retrieves merge history for a contact
func (s *ContactMergeService) GetMergeHistory(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) ([]*types.ContactMergeHistory, error) {
	// Verify contact belongs to organization
	contact, err := s.contactRepo.FindByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	if contact.OrganizationID != orgID {
		return nil, fmt.Errorf("contact does not belong to the specified organization")
	}

	// Get merge history
	history, err := s.repo.GetMergeHistory(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merge history: %w", err)
	}

	return history, nil
}

// buildFieldSelections creates field selection map based on merge strategy
func (s *ContactMergeService) buildFieldSelections(strategy string, customSelections map[string]string) map[string]string {
	selections := make(map[string]string)

	switch strategy {
	case "keep_master":
		// All fields from master (default behavior)
		return selections

	case "keep_latest":
		// This would require comparing updated_at timestamps
		// For now, default to master
		return selections

	case "custom":
		// Use provided custom selections
		return customSelections

	default:
		// Default to keep_master
		return selections
	}
}
