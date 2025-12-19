package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/crm/base"
	"github.com/KevTiv/alieze-erp/pkg/crm/errors"
	"github.com/KevTiv/alieze-erp/pkg/crm/validation"

	"github.com/google/uuid"
)

// ContactRequest represents a request to create a contact
type ContactRequest struct {
	Name           string     `json:"name"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	IsCustomer     bool       `json:"is_customer"`
	IsVendor       bool       `json:"is_vendor"`
	Street         *string    `json:"street,omitempty"`
	City           *string    `json:"city,omitempty"`
	StateID        *uuid.UUID `json:"state_id,omitempty"`
	CountryID      *uuid.UUID `json:"country_id,omitempty"`
	OrganizationID uuid.UUID  `json:"organization_id"`
}

// ContactUpdateRequest represents a request to update a contact
type ContactUpdateRequest struct {
	Name       *string    `json:"name,omitempty"`
	Email      *string    `json:"email,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	IsCustomer *bool      `json:"is_customer,omitempty"`
	IsVendor   *bool      `json:"is_vendor,omitempty"`
	Street     *string    `json:"street,omitempty"`
	City       *string    `json:"city,omitempty"`
	StateID    *uuid.UUID `json:"state_id,omitempty"`
	CountryID  *uuid.UUID `json:"country_id,omitempty"`
}

// ContactServiceV2 implements standardized contact service
type ContactServiceV2 struct {
	*base.CRUDService[types.Contact, ContactRequest, ContactUpdateRequest, types.ContactFilter]
}

// NewContactServiceV2 creates a new standardized contact service
func NewContactServiceV2(
	repo base.Repository[types.Contact, types.ContactFilter],
	authService base.AuthService,
	opts base.ServiceOptions,
) *ContactServiceV2 {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &ContactServiceV2{
		CRUDService: base.NewCRUDService[types.Contact, ContactRequest, ContactUpdateRequest, types.ContactFilter](
			repo, authService, opts,
		),
	}
}

// CreateContact creates a new contact
func (s *ContactServiceV2) CreateContact(ctx context.Context, req ContactRequest) (*types.Contact, error) {
	// Validate input
	if err := s.validateContactRequest(req); err != nil {
		return nil, err
	}

	// Convert request to entity
	contact := s.requestToContact(req)

	// Check authorization
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, contact.OrganizationID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Create contact
	result, err := s.GetRepository().Create(ctx, contact)
	if err != nil {
		return nil, errors.Wrap(err, "CREATE_FAILED", "failed to create contact")
	}

	// Log operation
	s.LogOperation(ctx, "create_contact", result.ID, map[string]interface{}{
		"organization_id": contact.OrganizationID,
		"name":            contact.Name,
		"is_customer":     contact.IsCustomer,
		"is_vendor":       contact.IsVendor,
	})

	// Publish event
	s.PublishEvent(ctx, "contact.created", result)

	return result, nil
}

// GetContact retrieves a contact by ID
func (s *ContactServiceV2) GetContact(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
	result, err := s.GetRepository().FindByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "GET_FAILED", "failed to get contact")
	}

	if result == nil {
		return nil, errors.ErrNotFound
	}

	// Check authorization
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, result.OrganizationID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	return result, nil
}

// UpdateContact updates an existing contact
func (s *ContactServiceV2) UpdateContact(ctx context.Context, id uuid.UUID, req ContactUpdateRequest) (*types.Contact, error) {
	// Get existing contact
	existing, err := s.GetContact(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate update request
	if err := s.validateContactUpdateRequest(req); err != nil {
		return nil, err
	}

	// Apply updates
	updated := s.applyContactUpdate(existing, req)

	// Update contact
	result, err := s.GetRepository().Update(ctx, *updated)
	if err != nil {
		return nil, errors.Wrap(err, "UPDATE_FAILED", "failed to update contact")
	}

	// Log operation
	s.LogOperation(ctx, "update_contact", result.ID, map[string]interface{}{
		"organization_id": result.OrganizationID,
		"changes":         s.getContactChanges(existing, result),
	})

	// Publish event
	s.PublishEvent(ctx, "contact.updated", result)

	return result, nil
}

// DeleteContact deletes a contact
func (s *ContactServiceV2) DeleteContact(ctx context.Context, id uuid.UUID) error {
	// Get existing contact for authorization
	existing, err := s.GetContact(ctx, id)
	if err != nil {
		return err
	}

	// Delete contact
	err = s.GetRepository().Delete(ctx, id)
	if err != nil {
		return errors.Wrap(err, "DELETE_FAILED", "failed to delete contact")
	}

	// Log operation
	s.LogOperation(ctx, "delete_contact", id, map[string]interface{}{
		"organization_id": existing.OrganizationID,
	})

	// Publish event
	s.PublishEvent(ctx, "contact.deleted", map[string]interface{}{
		"id":              id,
		"organization_id": existing.OrganizationID,
	})

	return nil
}

// ListContacts lists contacts with filtering
func (s *ContactServiceV2) ListContacts(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, int, error) {
	// Validate filter
	if err := s.validateContactFilter(filter); err != nil {
		return nil, 0, err
	}

	// Check organization access
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, filter.OrganizationID); err != nil {
		return nil, 0, errors.ErrOrganizationAccess
	}

	// Get contacts
	contacts, err := s.GetRepository().FindAll(ctx, filter)
	if err != nil {
		return nil, 0, errors.Wrap(err, "LIST_FAILED", "failed to list contacts")
	}

	// Get count
	count, err := s.GetRepository().Count(ctx, filter)
	if err != nil {
		return nil, 0, errors.Wrap(err, "COUNT_FAILED", "failed to count contacts")
	}

	return contacts, count, nil
}

// Helper methods

func (s *ContactServiceV2) validateContactRequest(req ContactRequest) error {
	return validation.ValidateMultiple(
		func() error { return validation.ValidateRequired("name", req.Name) },
		func() error { return validation.ValidateLength("name", req.Name, 1, 255) },
		func() error { return validation.ValidateEmail(safeString(req.Email)) },
		func() error { return validation.ValidatePhone(safeString(req.Phone)) },
		func() error { return validation.ValidateUUID(req.OrganizationID.String()) },
	)
}

func (s *ContactServiceV2) validateContactUpdateRequest(req ContactUpdateRequest) error {
	if req.Name != nil {
		if err := validation.ValidateLength("name", *req.Name, 1, 255); err != nil {
			return err
		}
	}
	if req.Email != nil {
		if err := validation.ValidateEmail(*req.Email); err != nil {
			return err
		}
	}
	if req.Phone != nil {
		if err := validation.ValidatePhone(*req.Phone); err != nil {
			return err
		}
	}
	if req.StateID != nil {
		if err := validation.ValidateUUID(req.StateID.String()); err != nil {
			return err
		}
	}
	if req.CountryID != nil {
		if err := validation.ValidateUUID(req.CountryID.String()); err != nil {
			return err
		}
	}
	return nil
}

func (s *ContactServiceV2) validateContactFilter(filter types.ContactFilter) error {
	return validation.ValidateMultiple(
		func() error { return validation.ValidateUUID(filter.OrganizationID.String()) },
		func() error {
			if filter.Email != nil {
				return validation.ValidateEmail(*filter.Email)
			}
			return nil
		},
		func() error {
			if filter.Phone != nil {
				return validation.ValidatePhone(*filter.Phone)
			}
			return nil
		},
		func() error {
			if filter.Limit < 0 {
				return &validation.ValidationError{Field: "limit", Message: "must be non-negative"}
			}
			if filter.Offset < 0 {
				return &validation.ValidationError{Field: "offset", Message: "must be non-negative"}
			}
			return nil
		},
	)
}

func (s *ContactServiceV2) requestToContact(req ContactRequest) types.Contact {
	return types.Contact{
		ID:             uuid.New(),
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		IsCustomer:     req.IsCustomer,
		IsVendor:       req.IsVendor,
		Street:         req.Street,
		City:           req.City,
		StateID:        req.StateID,
		CountryID:      req.CountryID,
		OrganizationID: req.OrganizationID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (s *ContactServiceV2) applyContactUpdate(existing *types.Contact, req ContactUpdateRequest) *types.Contact {
	updated := *existing

	if req.Name != nil {
		updated.Name = *req.Name
	}
	if req.Email != nil {
		updated.Email = req.Email
	}
	if req.Phone != nil {
		updated.Phone = req.Phone
	}
	if req.IsCustomer != nil {
		updated.IsCustomer = *req.IsCustomer
	}
	if req.IsVendor != nil {
		updated.IsVendor = *req.IsVendor
	}
	if req.Street != nil {
		updated.Street = req.Street
	}
	if req.City != nil {
		updated.City = req.City
	}
	if req.StateID != nil {
		updated.StateID = req.StateID
	}
	if req.CountryID != nil {
		updated.CountryID = req.CountryID
	}

	updated.UpdatedAt = time.Now()

	return &updated
}

func (s *ContactServiceV2) getContactChanges(existing, updated *types.Contact) map[string]interface{} {
	changes := make(map[string]interface{})

	if existing.Name != updated.Name {
		changes["name"] = map[string]string{"old": existing.Name, "new": updated.Name}
	}
	if (existing.Email == nil) != (updated.Email == nil) || (existing.Email != nil && updated.Email != nil && *existing.Email != *updated.Email) {
		changes["email"] = map[string]interface{}{"old": existing.Email, "new": updated.Email}
	}
	if (existing.Phone == nil) != (updated.Phone == nil) || (existing.Phone != nil && updated.Phone != nil && *existing.Phone != *updated.Phone) {
		changes["phone"] = map[string]interface{}{"old": existing.Phone, "new": updated.Phone}
	}
	if existing.IsCustomer != updated.IsCustomer {
		changes["is_customer"] = map[string]bool{"old": existing.IsCustomer, "new": updated.IsCustomer}
	}
	if existing.IsVendor != updated.IsVendor {
		changes["is_vendor"] = map[string]bool{"old": existing.IsVendor, "new": updated.IsVendor}
	}

	return changes
}

// CreateRelationship creates a relationship between contacts
func (s *ContactServiceV2) CreateRelationship(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, req types.ContactRelationshipCreateRequest) (*types.ContactRelationship, error) {
	// Validate organization access
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Check if contact exists
	exists, err := s.GetRepository().(interface {
		ContactExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	}).ContactExists(ctx, orgID, contactID)
	if err != nil {
		return nil, errors.Wrap(err, "VALIDATION_FAILED", "failed to check contact existence")
	}
	if !exists {
		return nil, errors.New("contact_not_found", "contact does not exist")
	}

	// Create the relationship
	relationship := &types.ContactRelationship{
		ID:               uuid.New(),
		OrganizationID:   orgID,
		ContactID:        contactID,
		RelatedContactID: req.RelatedContactID,
		Type:             types.ContactRelationshipType(req.Type),
		Notes:            req.Notes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Validate relationship type
	if !types.IsValidRelationshipType(relationship.Type) {
		return nil, errors.New("invalid_relationship_type", "invalid relationship type")
	}

	// Create relationship in repository
	err = s.GetRepository().(interface {
		CreateRelationship(context.Context, *types.ContactRelationship) error
	}).CreateRelationship(ctx, relationship)
	if err != nil {
		return nil, errors.Wrap(err, "CREATE_FAILED", "failed to create relationship")
	}

	// Log operation
	s.LogOperation(ctx, "create_contact_relationship", relationship.ID, map[string]interface{}{
		"organization_id":    orgID,
		"contact_id":         contactID,
		"related_contact_id": req.RelatedContactID,
		"type":               req.Type,
	})

	// Publish event
	s.PublishEvent(ctx, "contact.relationship.created", relationship)

	return relationship, nil
}

// ListRelationships lists relationships for a contact
func (s *ContactServiceV2) ListRelationships(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, relationshipType string, limit int) ([]*types.ContactRelationship, error) {
	// Validate organization access
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Check if contact exists
	exists, err := s.GetRepository().(interface {
		ContactExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	}).ContactExists(ctx, orgID, contactID)
	if err != nil {
		return nil, errors.Wrap(err, "VALIDATION_FAILED", "failed to check contact existence")
	}
	if !exists {
		return nil, errors.New("contact_not_found", "contact does not exist")
	}

	// Get relationships from repository
	relationships, err := s.GetRepository().(interface {
		FindRelationships(context.Context, uuid.UUID, uuid.UUID, string, int) ([]*types.ContactRelationship, error)
	}).FindRelationships(ctx, orgID, contactID, relationshipType, limit)
	if err != nil {
		return nil, errors.Wrap(err, "QUERY_FAILED", "failed to find relationships")
	}

	return relationships, nil
}

// AddToSegments adds a contact to segments
func (s *ContactServiceV2) AddToSegments(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, req types.ContactSegmentationRequest) error {
	// Validate organization access
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, orgID); err != nil {
		return errors.ErrOrganizationAccess
	}

	// Check if contact exists
	exists, err := s.GetRepository().(interface {
		ContactExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	}).ContactExists(ctx, orgID, contactID)
	if err != nil {
		return errors.Wrap(err, "VALIDATION_FAILED", "failed to check contact existence")
	}
	if !exists {
		return errors.New("contact_not_found", "contact does not exist")
	}

	// Add to segments
	err = s.GetRepository().(interface {
		AddContactToSegments(context.Context, uuid.UUID, uuid.UUID, []string) error
	}).AddContactToSegments(ctx, orgID, contactID, req.SegmentIDs)
	if err != nil {
		return errors.Wrap(err, "UPDATE_FAILED", "failed to add contact to segments")
	}

	// Add tags if provided
	if len(req.CustomTags) > 0 {
		err = s.GetRepository().(interface {
			AddContactTags(context.Context, uuid.UUID, uuid.UUID, []string) error
		}).AddContactTags(ctx, orgID, contactID, req.CustomTags)
		if err != nil {
			return errors.Wrap(err, "UPDATE_FAILED", "failed to add contact tags")
		}
	}

	// Log operation
	s.LogOperation(ctx, "add_contact_to_segments", contactID, map[string]interface{}{
		"organization_id": orgID,
		"segment_ids":     req.SegmentIDs,
		"tags":            req.CustomTags,
	})

	// Publish event
	s.PublishEvent(ctx, "contact.segments.added", map[string]interface{}{
		"contact_id": contactID,
		"segments":   req.SegmentIDs,
		"tags":       req.CustomTags,
	})

	return nil
}

// CalculateContactScore calculates engagement and lead scores for a contact
func (s *ContactServiceV2) CalculateContactScore(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) (*types.ContactScore, error) {
	// Validate organization access
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Check if contact exists
	exists, err := s.GetRepository().(interface {
		ContactExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	}).ContactExists(ctx, orgID, contactID)
	if err != nil {
		return nil, errors.Wrap(err, "VALIDATION_FAILED", "failed to check contact existence")
	}
	if !exists {
		return nil, errors.New("contact_not_found", "contact does not exist")
	}

	// For now, return a basic score calculation
	// In a real implementation, this would analyze contact activity, engagement, etc.
	score := &types.ContactScore{
		EngagementScore: 50, // Placeholder
		LeadScore:       60, // Placeholder
		EngagementFactors: map[string]interface{}{
			"recent_activity": "medium",
			"response_rate":   "good",
		},
		LeadFactors: map[string]interface{}{
			"fit":      "good",
			"interest": "medium",
			"budget":   "unknown",
		},
		LastUpdated: time.Now(),
	}

	return score, nil
}

// Helper function to safely dereference string pointers
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
