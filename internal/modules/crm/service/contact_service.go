package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"alieze-erp/internal/modules/crm/domain"
	"alieze-erp/internal/modules/crm/repository"

	"github.com/google/uuid"
)

// AuthService defines the interface for authentication/authorization
type AuthService interface {
	GetOrganizationID(ctx context.Context) (uuid.UUID, error)
	GetUserID(ctx context.Context) (uuid.UUID, error)
	CheckPermission(ctx context.Context, permission string) error
}

// ContactService handles contact business logic
type ContactService struct {
	repo        repository.ContactRepo
	authService AuthService
	logger      *log.Logger
}

func NewContactService(repo repository.ContactRepo, authService AuthService) *ContactService {
	return &ContactService{
		repo:        repo,
		authService: authService,
		logger:      log.New(log.Writer(), "contact-service: ", log.LstdFlags),
	}
}

func (s *ContactService) CreateContact(ctx context.Context, contact domain.Contact) (*domain.Contact, error) {
	// Validate required fields
	if contact.Name == "" {
		return nil, errors.New("contact name is required")
	}

	// Set organization from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	contact.OrganizationID = orgID

	// Validate email format if provided
	if contact.Email != nil && *contact.Email != "" {
		if !isValidEmail(*contact.Email) {
			return nil, errors.New("invalid email format")
		}
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "contacts:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Create the contact
	created, err := s.repo.Create(ctx, contact)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	s.logger.Printf("Created contact %s for organization %s", created.ID, created.OrganizationID)

	return created, nil
}

func (s *ContactService) GetContact(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid contact id")
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "contacts:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	contact, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if contact.OrganizationID != orgID {
		return nil, fmt.Errorf("contact does not belong to organization %s", orgID)
	}

	return contact, nil
}

func (s *ContactService) ListContacts(ctx context.Context, filter domain.ContactFilter) ([]domain.Contact, int, error) {
	// Set organization from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "contacts:read"); err != nil {
		return nil, 0, fmt.Errorf("permission denied: %w", err)
	}

	// Set default pagination
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	contacts, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list contacts: %w", err)
	}

	count, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count contacts: %w", err)
	}

	return contacts, count, nil
}

func (s *ContactService) UpdateContact(ctx context.Context, contact domain.Contact) (*domain.Contact, error) {
	if contact.ID == uuid.Nil {
		return nil, errors.New("contact id is required")
	}

	if contact.Name == "" {
		return nil, errors.New("contact name is required")
	}

	// Get existing contact to verify organization
	existing, err := s.repo.FindByID(ctx, contact.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing contact: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("contact does not belong to organization %s", orgID)
	}

	// Set organization
	contact.OrganizationID = orgID

	// Validate email format if provided
	if contact.Email != nil && *contact.Email != "" {
		if !isValidEmail(*contact.Email) {
			return nil, errors.New("invalid email format")
		}
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "contacts:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	updated, err := s.repo.Update(ctx, contact)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	s.logger.Printf("Updated contact %s for organization %s", updated.ID, updated.OrganizationID)

	return updated, nil
}

func (s *ContactService) DeleteContact(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid contact id")
	}

	// Get existing contact to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find existing contact: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("contact does not belong to organization %s", orgID)
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "contacts:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	s.logger.Printf("Deleted contact %s for organization %s", id, orgID)

	return nil
}

// Helper functions
func isValidEmail(email string) bool {
	// Simple email validation
	return len(email) >= 5 && strings.Contains(email, "@") && strings.Contains(email, ".")
}
