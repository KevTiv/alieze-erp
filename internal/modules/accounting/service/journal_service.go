package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
)

type JournalService struct {
	repo repository.JournalRepository
}

func NewJournalService(repo repository.JournalRepository) *JournalService {
	return &JournalService{
		repo: repo,
	}
}

func (s *JournalService) CreateJournal(ctx context.Context, journal types.Journal) (*types.Journal, error) {
	// Validate the journal
	if err := s.validateJournal(journal); err != nil {
		return nil, fmt.Errorf("invalid journal: %w", err)
	}

	// Check for duplicate code within organization
	existing, err := s.repo.FindByCode(ctx, journal.OrganizationID, journal.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing journal: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("journal with code '%s' already exists", journal.Code)
	}

	// Create the journal
	createdJournal, err := s.repo.Create(ctx, journal)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal: %w", err)
	}

	return createdJournal, nil
}

func (s *JournalService) GetJournal(ctx context.Context, id uuid.UUID) (*types.Journal, error) {
	journal, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal: %w", err)
	}
	if journal == nil {
		return nil, fmt.Errorf("journal not found")
	}

	return journal, nil
}

func (s *JournalService) ListJournals(ctx context.Context, filters repository.JournalFilter) ([]types.Journal, error) {
	journals, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list journals: %w", err)
	}

	return journals, nil
}

func (s *JournalService) UpdateJournal(ctx context.Context, journal types.Journal) (*types.Journal, error) {
	// Validate the journal
	if err := s.validateJournal(journal); err != nil {
		return nil, fmt.Errorf("invalid journal: %w", err)
	}

	// Check if journal exists
	existing, err := s.repo.FindByID(ctx, journal.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing journal: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("journal not found")
	}

	// Check for duplicate code if code is being changed
	if existing.Code != journal.Code {
		duplicate, err := s.repo.FindByCode(ctx, journal.OrganizationID, journal.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to check duplicate code: %w", err)
		}
		if duplicate != nil {
			return nil, fmt.Errorf("journal with code '%s' already exists", journal.Code)
		}
	}

	// Update the journal
	updatedJournal, err := s.repo.Update(ctx, journal)
	if err != nil {
		return nil, fmt.Errorf("failed to update journal: %w", err)
	}

	return updatedJournal, nil
}

func (s *JournalService) DeleteJournal(ctx context.Context, id uuid.UUID) error {
	// Check if journal exists
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check existing journal: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("journal not found")
	}

	// Delete the journal
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete journal: %w", err)
	}

	return nil
}

func (s *JournalService) GetJournalsByType(ctx context.Context, organizationID uuid.UUID, journalType string) ([]types.Journal, error) {
	journals, err := s.repo.FindByType(ctx, organizationID, journalType)
	if err != nil {
		return nil, fmt.Errorf("failed to get journals by type: %w", err)
	}

	return journals, nil
}

func (s *JournalService) validateJournal(journal types.Journal) error {
	if journal.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if journal.Name == "" {
		return fmt.Errorf("name is required")
	}

	if journal.Code == "" {
		return fmt.Errorf("code is required")
	}

	if journal.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate journal type is one of the valid types
	validTypes := map[string]bool{
		"sale":     true,
		"purchase": true,
		"cash":     true,
		"bank":     true,
		"general":  true,
	}

	if !validTypes[journal.Type] {
		return fmt.Errorf("invalid journal type: %s (must be one of: sale, purchase, cash, bank, general)", journal.Type)
	}

	return nil
}
