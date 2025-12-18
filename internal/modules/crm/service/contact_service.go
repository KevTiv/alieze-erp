package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"alieze-erp/internal/modules/crm/types"
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
	ruleEngine  interface{} // Will be used for rule-based validation
	eventBus    interface{} // Event bus for publishing domain events
	logger      *log.Logger
}

func NewContactService(repo repository.ContactRepo, authService AuthService) *ContactService {
	return &ContactService{
		repo:        repo,
		authService: authService,
		logger:      log.New(log.Writer(), "contact-service: ", log.LstdFlags),
	}
}

// NewContactServiceWithRules creates a contact service with rule engine support
func NewContactServiceWithRules(repo repository.ContactRepo, authService AuthService, ruleEngine interface{}) *ContactService {
	service := NewContactService(repo, authService)
	service.ruleEngine = ruleEngine
	return service
}

// NewContactServiceWithDependencies creates a contact service with all dependencies
func NewContactServiceWithDependencies(repo repository.ContactRepo, authService AuthService, ruleEngine interface{}, eventBus interface{}) *ContactService {
	service := NewContactService(repo, authService)
	service.ruleEngine = ruleEngine
	service.eventBus = eventBus
	return service
}

func (s *ContactService) CreateContact(ctx context.Context, contact types.Contact) (*types.Contact, error) {
	// Use rule engine for validation if available
	if s.ruleEngine != nil {
		if ruleEngine, ok := s.ruleEngine.(interface{
			Validate(ctx context.Context, ruleName string, entity interface{}) error
		}); ok {
			if err := ruleEngine.Validate(ctx, "contact_create", contact); err != nil {
				return nil, fmt.Errorf("validation failed: %w", err)
			}
		}
	} else {
		// Fallback to hardcoded validation
		if contact.Name == "" {
			return nil, errors.New("contact name is required")
		}

		// Validate email format if provided
		if contact.Email != nil && *contact.Email != "" {
			if !isValidEmail(*contact.Email) {
				return nil, errors.New("invalid email format")
			}
		}
	}

	// Set organization from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	contact.OrganizationID = orgID

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

	// Publish contact.created event
	s.publishEvent(ctx, "contact.created", created)

	return created, nil
}

func (s *ContactService) GetContact(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
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

func (s *ContactService) ListContacts(ctx context.Context, filter types.ContactFilter) ([]types.Contact, int, error) {
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

func (s *ContactService) UpdateContact(ctx context.Context, contact types.Contact) (*types.Contact, error) {
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

	// Publish contact.updated event
	s.publishEvent(ctx, "contact.updated", updated)

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

	// Publish contact.deleted event
	s.publishEvent(ctx, "contact.deleted", map[string]interface{}{
		"id":              id,
		"organization_id": orgID,
	})

	return nil
}

// Helper functions
func isValidEmail(email string) bool {
	// Simple email validation
	return len(email) >= 5 && strings.Contains(email, "@") && strings.Contains(email, ".")
}

// publishEvent publishes an event to the event bus if available
func (s *ContactService) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if s.eventBus != nil {
		if bus, ok := s.eventBus.(interface {
			Publish(ctx context.Context, eventType string, payload interface{}) error
		}); ok {
			if err := bus.Publish(ctx, eventType, payload); err != nil {
				s.logger.Printf("Failed to publish event %s: %v", eventType, err)
			}
		}
	}
}

// ContactRelationship methods

func (s *ContactService) CreateRelationship(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
	req types.ContactRelationshipCreateRequest,
) (*types.ContactRelationship, error) {
	// Validate the contact exists and belongs to the organization
	exists, err := s.repo.ContactExists(ctx, orgID, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to check contact existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("contact not found")
	}

	// Validate the related contact exists and belongs to the same organization
	relatedExists, err := s.repo.ContactExists(ctx, orgID, req.RelatedContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to check related contact existence: %w", err)
	}
	if !relatedExists {
		return nil, fmt.Errorf("related contact not found")
	}

	// Create the relationship
	relationship := &types.ContactRelationship{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		ContactID:       contactID,
		RelatedContactID: req.RelatedContactID,
		Type:            req.Type,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Use repository to create the relationship
	err = s.repo.CreateRelationship(ctx, relationship)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	return relationship, nil
}

func (s *ContactService) ListRelationships(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
	relationshipType string,
	limit int,
) ([]*types.ContactRelationship, error) {
	// Validate the contact exists and belongs to the organization
	exists, err := s.repo.ContactExists(ctx, orgID, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to check contact existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("contact not found")
	}

	// Get relationships from repository
	relationships, err := s.repo.FindRelationships(ctx, orgID, contactID, relationshipType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships: %w", err)
	}

	return relationships, nil
}

func (s *ContactService) AddToSegments(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
	req types.ContactSegmentationRequest,
) error {
	// Validate the contact exists and belongs to the organization
	exists, err := s.repo.ContactExists(ctx, orgID, contactID)
	if err != nil {
		return fmt.Errorf("failed to check contact existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("contact not found")
	}

	// Add to predefined segments
	if len(req.SegmentIDs) > 0 {
		err = s.repo.AddContactToSegments(ctx, orgID, contactID, req.SegmentIDs)
		if err != nil {
			return fmt.Errorf("failed to add to segments: %w", err)
		}
	}

	// Add custom tags
	if len(req.CustomTags) > 0 {
		err = s.repo.AddContactTags(ctx, orgID, contactID, req.CustomTags)
		if err != nil {
			return fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return nil
}

func (s *ContactService) CalculateContactScore(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
) (*types.ContactScore, error) {
	// Validate the contact exists and belongs to the organization
	exists, err := s.repo.ContactExists(ctx, orgID, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to check contact existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("contact not found")
	}

	// For now, return a basic score structure
	// In a full implementation, this would calculate actual engagement and lead scores
	return &types.ContactScore{
		EngagementScore: 75,
		LeadScore:       65,
		EngagementFactors: map[string]interface{}{
			"activityFrequency": map[string]interface{}{
				"count": 8,
				"score": 30,
			},
			"recency": map[string]interface{}{
				"daysSinceLastActivity": 5,
				"score": 25,
			},
			"responseRate": map[string]interface{}{
				"rate": 0.8,
				"score": 20,
			},
		},
		LeadFactors: map[string]interface{}{
			"fit": map[string]interface{}{
				"score": 30,
			},
			"interest": map[string]interface{}{
				"score": 25,
			},
			"budget": map[string]interface{}{
				"score": 10,
			},
		},
		LastUpdated: time.Now(),
	}, nil
}
