package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"database/sql"

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

// BulkCreateContacts creates multiple contacts in a single operation
func (s *ContactServiceV2) BulkCreateContacts(ctx context.Context, requests []ContactRequest) ([]*types.Contact, []error) {
	var results []*types.Contact
	var errors []error

	// Validate organization access for all requests
	if len(requests) > 0 {
		orgID := requests[0].OrganizationID
		if err := s.GetAuthService().CheckOrganizationAccess(ctx, orgID); err != nil {
			// All requests will fail with the same error
			for range requests {
				errors = append(errors, errors.ErrOrganizationAccess)
			}
			return nil, errors
		}
	}

	for _, req := range requests {
		// Validate individual request
		if err := s.validateContactRequest(req); err != nil {
			errors = append(errors, err)
			continue
		}

		// Convert request to entity
		contact := s.requestToContact(req)

		// Create contact
		result, err := s.GetRepository().Create(ctx, contact)
		if err != nil {
			errors = append(errors, errors.Wrap(err, "CREATE_FAILED", "failed to create contact"))
			continue
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

		results = append(results, result)
	}

	return results, errors
}

// AdvancedSearchContacts performs advanced search with multiple criteria
func (s *ContactServiceV2) AdvancedSearchContacts(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error) {
	// Validate filter
	if err := s.validateAdvancedContactFilter(filter); err != nil {
		return nil, 0, err
	}

	// Check organization access
	if err := s.GetAuthService().CheckOrganizationAccess(ctx, filter.OrganizationID); err != nil {
		return nil, 0, errors.ErrOrganizationAccess
	}

	// Build query based on filter criteria
	query, args := s.buildAdvancedSearchQuery(filter)

	// Execute query
	rows, err := s.GetRepository().(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "SEARCH_FAILED", "failed to execute search query")
	}
	defer rows.Close()

	var contacts []*types.Contact
	for rows.Next() {
		var contact types.Contact
		if err := rows.Scan(
			&contact.ID,
			&contact.OrganizationID,
			&contact.Name,
			&contact.Email,
			&contact.Phone,
			&contact.IsCustomer,
			&contact.IsVendor,
			&contact.Street,
			&contact.City,
			&contact.StateID,
			&contact.CountryID,
			&contact.CreatedAt,
			&contact.UpdatedAt,
			&contact.DeletedAt,
		); err != nil {
			return nil, 0, errors.Wrap(err, "SCAN_FAILED", "failed to scan contact")
		}
		contacts = append(contacts, &contact)
	}

	// Get total count
	count, err := s.getAdvancedSearchCount(ctx, filter)
	if err != nil {
		return nil, 0, errors.Wrap(err, "COUNT_FAILED", "failed to get search count")
	}

	return contacts, count, nil
}

// validateAdvancedContactFilter validates the advanced contact filter
func (s *ContactServiceV2) validateAdvancedContactFilter(filter types.AdvancedContactFilter) error {
	if filter.OrganizationID == uuid.Nil {
		return errors.New("organization_id_required", "organization_id is required")
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	return nil
}

// buildAdvancedSearchQuery builds the SQL query for advanced search
func (s *ContactServiceV2) buildAdvancedSearchQuery(filter types.AdvancedContactFilter) (string, []interface{}) {
	query := `SELECT id, organization_id, name, email, phone, is_customer, is_vendor, street, city, state_id, country_id, created_at, updated_at, deleted_at
		FROM contacts
		WHERE organization_id = $1 AND deleted_at IS NULL`

	args := []interface{}{filter.OrganizationID}
	argIndex := 2

	// Add search query condition
	if filter.SearchQuery != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", argIndex, argIndex+1, argIndex+2)
		searchParam := "%" + filter.SearchQuery + "%"
		args = append(args, searchParam, searchParam, searchParam)
		argIndex += 3
	}

	// Add tag conditions (would require tags table)
	if len(filter.Tags) > 0 {
		// Implementation would join with contact_tags table
	}

	// Add score range condition
	if filter.ScoreRange.Min > 0 || filter.ScoreRange.Max > 0 {
		if filter.ScoreRange.Min > 0 && filter.ScoreRange.Max > 0 {
			query += fmt.Sprintf(" AND engagement_score BETWEEN $%d AND $%d", argIndex, argIndex+1)
			args = append(args, filter.ScoreRange.Min, filter.ScoreRange.Max)
			argIndex += 2
		} else if filter.ScoreRange.Min > 0 {
			query += fmt.Sprintf(" AND engagement_score >= $%d", argIndex)
			args = append(args, filter.ScoreRange.Min)
			argIndex++
		} else if filter.ScoreRange.Max > 0 {
			query += fmt.Sprintf(" AND engagement_score <= $%d", argIndex)
			args = append(args, filter.ScoreRange.Max)
			argIndex++
		}
	}

	// Add last contacted condition
	if !filter.LastContacted.From.IsZero() || !filter.LastContacted.To.IsZero() {
		if !filter.LastContacted.From.IsZero() && !filter.LastContacted.To.IsZero() {
			query += fmt.Sprintf(" AND last_contacted_at BETWEEN $%d AND $%d", argIndex, argIndex+1)
			args = append(args, filter.LastContacted.From, filter.LastContacted.To)
			argIndex += 2
		} else if !filter.LastContacted.From.IsZero() {
			query += fmt.Sprintf(" AND last_contacted_at >= $%d", argIndex)
			args = append(args, filter.LastContacted.From)
			argIndex++
		} else if !filter.LastContacted.To.IsZero() {
			query += fmt.Sprintf(" AND last_contacted_at <= $%d", argIndex)
			args = append(args, filter.LastContacted.To)
			argIndex++
		}
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	return query, args
}

// getAdvancedSearchCount gets the total count for advanced search
func (s *ContactServiceV2) getAdvancedSearchCount(ctx context.Context, filter types.AdvancedContactFilter) (int, error) {
	query, args := s.buildAdvancedSearchQuery(filter)
	// Replace SELECT with COUNT
	countQuery := strings.Replace(query,
		"SELECT id, organization_id, name, email, phone, is_customer, is_vendor, street, city, state_id, country_id, created_at, updated_at, deleted_at",
		"SELECT COUNT(*)", 1)
	// Remove ORDER BY, LIMIT, OFFSET
	countQuery = regexCount.ReplaceAllString(countQuery, "")

	var count int
	if err := s.GetRepository().(interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}).QueryRowContext(ctx, countQuery, args...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

var regexCount = regexp.MustCompile(`ORDER BY.*$`)

// GetCRMDashboard retrieves comprehensive CRM dashboard data
func (s *ContactServiceV2) GetCRMDashboard(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error) {
	// Parse time range
	fromDate, toDate, err := parseTimeRange(timeRange)
	if err != nil {
		return nil, errors.Wrap(err, "INVALID_TIME_RANGE", "invalid time range")
	}

	// Check cache first
	if cachedData, found := s.getCachedDashboard(ctx, orgID, timeRange, ""); found {
		if dashboard, ok := cachedData.(*types.CRMDashboard); ok {
			return dashboard, nil
		}
	}

	dashboard := &types.CRMDashboard{
		TimeRange: timeRange,
		Summary:   types.DashboardSummary{},
		Trends: types.DashboardTrends{
			ContactGrowth: make([]types.TrendDataPoint, 0),
			Engagement:    make([]types.TrendDataPoint, 0),
			ResponseRate:  make([]types.TrendDataPoint, 0),
		},
		TopContacts:      make([]types.TopContact, 0),
		RecentActivities: make([]types.RecentActivity, 0),
	}

	// Get summary data
	summary, err := s.getDashboardSummary(ctx, orgID, fromDate, toDate)
	if err != nil {
		return nil, errors.Wrap(err, "DASHBOARD_ERROR", "failed to get dashboard summary")
	}
	dashboard.Summary = *summary

	// Get trend data
	trends, err := s.getDashboardTrends(ctx, orgID, fromDate, toDate)
	if err != nil {
		return nil, errors.Wrap(err, "DASHBOARD_ERROR", "failed to get dashboard trends")
	}
	dashboard.Trends = *trends

	// Get top contacts
	topContacts, err := s.getTopContacts(ctx, orgID, fromDate, toDate)
	if err != nil {
		return nil, errors.Wrap(err, "DASHBOARD_ERROR", "failed to get top contacts")
	}
	dashboard.TopContacts = topContacts

	// Get recent activities
	activities, err := s.getRecentActivities(ctx, orgID, fromDate, toDate)
	if err != nil {
		return nil, errors.Wrap(err, "DASHBOARD_ERROR", "failed to get recent activities")
	}
	dashboard.RecentActivities = activities

	// Cache the dashboard data
	s.setDashboardCache(ctx, orgID, timeRange, "", dashboard)

	return dashboard, nil
}

// GetActivityDashboard retrieves activity-focused dashboard data
func (s *ContactServiceV2) GetActivityDashboard(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error) {
	// Parse time range
	fromDate, toDate, err := parseTimeRange(timeRange)
	if err != nil {
		return nil, errors.Wrap(err, "INVALID_TIME_RANGE", "invalid time range")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s", timeRange, contactType)
	if cachedData, found := s.getCachedDashboard(ctx, orgID, timeRange, cacheKey); found {
		if dashboard, ok := cachedData.(*types.ActivityDashboard); ok {
			return dashboard, nil
		}
	}

	dashboard := &types.ActivityDashboard{
		TimeRange:   timeRange,
		ContactType: contactType,
		ActivitySummary: types.ActivitySummary{
			ActivityTypes:  make(map[string]int),
			ActivityTrends: make([]types.ActivityTrend, 0),
		},
		RecentActivities:  make([]types.RecentActivity, 0),
		ContactEngagement: make([]types.ContactEngagement, 0),
	}

	// Get activity data based on contact type
	activities, err := s.getActivityData(ctx, orgID, contactType, fromDate, toDate)
	if err != nil {
		return nil, errors.Wrap(err, "DASHBOARD_ERROR", "failed to get activity data")
	}
	dashboard.RecentActivities = activities

	// Calculate activity statistics
	dashboard.ActivitySummary = calculateActivityStatistics(activities, fromDate, toDate)

	// Get contact engagement data
	engagement, err := s.getContactEngagement(ctx, orgID, contactType, fromDate, toDate)
	if err != nil {
		return nil, errors.Wrap(err, "DASHBOARD_ERROR", "failed to get contact engagement")
	}
	dashboard.ContactEngagement = engagement

	// Cache the dashboard data
	s.setDashboardCache(ctx, orgID, timeRange, cacheKey, dashboard)

	return dashboard, nil
}

// parseTimeRange parses time range strings into date ranges
func parseTimeRange(timeRange string) (time.Time, time.Time, error) {
	now := time.Now()
	toDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	var fromDate time.Time

	switch timeRange {
	case "7d":
		fromDate = toDate.AddDate(0, 0, -7)
	case "30d":
		fromDate = toDate.AddDate(0, 0, -30)
	case "90d":
		fromDate = toDate.AddDate(0, 0, -90)
	case "1y":
		fromDate = toDate.AddDate(-1, 0, 0)
	case "custom":
		// Would need additional parameters for custom range
		fallthrough
	default:
		return time.Time{}, time.Time{}, errors.New("invalid_time_range", "invalid time range specified")
	}

	return fromDate, toDate, nil
}

// getDashboardSummary gets summary statistics for the dashboard
func (s *ContactServiceV2) getDashboardSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*types.DashboardSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_contacts,
			COUNT(CASE WHEN created_at >= $1 AND created_at <= $2 THEN 1 END) as new_contacts,
			COUNT(CASE WHEN updated_at >= $1 THEN 1 END) as active_contacts,
			COUNT(CASE WHEN last_contacted_at < $1 AND last_contacted_at IS NOT NULL THEN 1 END) as at_risk_contacts,
			COUNT(CASE WHEN engagement_score >= 80 THEN 1 END) as high_value_contacts
		FROM contacts
		WHERE organization_id = $3 AND deleted_at IS NULL
	`

	var summary types.DashboardSummary
	err := s.GetRepository().(interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}).QueryRowContext(ctx, query, fromDate, toDate, orgID).Scan(
		&summary.TotalContacts,
		&summary.NewContacts,
		&summary.ActiveContacts,
		&summary.AtRiskContacts,
		&summary.HighValueContacts,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}

	return &summary, nil
}

// getDashboardTrends gets trend data for the dashboard
func (s *ContactServiceV2) getDashboardTrends(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*types.DashboardTrends, error) {
	// This would implement trend calculations
	// For now, return empty trends
	return &types.DashboardTrends{
		ContactGrowth: make([]types.TrendDataPoint, 0),
		Engagement:    make([]types.TrendDataPoint, 0),
		ResponseRate:  make([]types.TrendDataPoint, 0),
	}, nil
}

// getTopContacts gets top contacts for the dashboard
func (s *ContactServiceV2) getTopContacts(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) ([]types.TopContact, error) {
	query := `
		SELECT
			id, name, company, engagement_score,
			last_activity_type, next_recommended_action
		FROM contacts
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY engagement_score DESC, updated_at DESC
		LIMIT 10
	`

	rows, err := s.GetRepository().(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	}).QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get top contacts: %w", err)
	}
	defer rows.Close()

	var contacts []types.TopContact
	for rows.Next() {
		var contact types.TopContact
		if err := rows.Scan(
			&contact.ID,
			&contact.Name,
			&contact.Company,
			&contact.Score,
			&contact.LastActivity,
			&contact.NextAction,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// getRecentActivities gets recent activities for the dashboard
func (s *ContactServiceV2) getRecentActivities(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) ([]types.RecentActivity, error) {
	query := `
		SELECT
			a.id, a.contact_id, c.name as contact_name,
			a.type, a.subject, a.created_at as date, a.status
		FROM activities a
		JOIN contacts c ON a.contact_id = c.id
		WHERE c.organization_id = $1 AND a.created_at BETWEEN $2 AND $3
		ORDER BY a.created_at DESC
		LIMIT 20
	`

	rows, err := s.GetRepository().(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	}).QueryContext(ctx, query, orgID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}
	defer rows.Close()

	var activities []types.RecentActivity
	for rows.Next() {
		var activity types.RecentActivity
		if err := rows.Scan(
			&activity.ID,
			&activity.ContactID,
			&activity.ContactName,
			&activity.Type,
			&activity.Subject,
			&activity.Date,
			&activity.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// getActivityData gets activity data for the activity dashboard
func (s *ContactServiceV2) getActivityData(ctx context.Context, orgID uuid.UUID, contactType string, fromDate, toDate time.Time) ([]types.RecentActivity, error) {
	query := `
		SELECT
			a.id, a.contact_id, c.name as contact_name,
			a.type, a.subject, a.created_at as date, a.status
		FROM activities a
		JOIN contacts c ON a.contact_id = c.id
		WHERE c.organization_id = $1 AND a.created_at BETWEEN $2 AND $3
	`

	args := []interface{}{orgID, fromDate, toDate}

	// Add contact type filter
	if contactType != "all" {
		switch contactType {
		case "customers":
			query += " AND c.is_customer = true"
		case "vendors":
			query += " AND c.is_vendor = true"
		case "leads":
			query += " AND c.is_lead = true"
		}
	}

	query += " ORDER BY a.created_at DESC LIMIT 50"

	rows, err := s.GetRepository().(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity data: %w", err)
	}
	defer rows.Close()

	var activities []types.RecentActivity
	for rows.Next() {
		var activity types.RecentActivity
		if err := rows.Scan(
			&activity.ID,
			&activity.ContactID,
			&activity.ContactName,
			&activity.Type,
			&activity.Subject,
			&activity.Date,
			&activity.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// calculateActivityStatistics calculates activity statistics
func calculateActivityStatistics(activities []types.RecentActivity, fromDate, toDate time.Time) types.ActivitySummary {
	summary := types.ActivitySummary{
		TotalActivities: len(activities),
		ActivityTypes:   make(map[string]int),
		ActivityTrends:  make([]types.ActivityTrend, 0),
	}

	// Count activity types
	for _, activity := range activities {
		summary.ActivityTypes[activity.Type]++
	}

	// Calculate trends by date
	dateMap := make(map[string]types.ActivityTrend)
	for _, activity := range activities {
		dateKey := activity.Date.Format("2006-01-02")
		trend, exists := dateMap[dateKey]
		if !exists {
			trend = types.ActivityTrend{Date: dateKey}
		}

		switch activity.Type {
		case "call":
			trend.Calls++
		case "email":
			trend.Emails++
		case "meeting":
			trend.Meetings++
		default:
			trend.Other++
		}

		dateMap[dateKey] = trend
	}

	// Convert map to slice and sort by date
	for _, trend := range dateMap {
		summary.ActivityTrends = append(summary.ActivityTrends, trend)
	}

	sort.Slice(summary.ActivityTrends, func(i, j int) bool {
		return summary.ActivityTrends[i].Date < summary.ActivityTrends[j].Date
	})

	return summary
}

// getContactEngagement gets contact engagement data
func (s *ContactServiceV2) getContactEngagement(ctx context.Context, orgID uuid.UUID, contactType string, fromDate, toDate time.Time) ([]types.ContactEngagement, error) {
	query := `
		SELECT
			c.id, c.name,
			c.engagement_score,
			MAX(a.type) as last_activity_type,
			CURRENT_DATE - MAX(a.created_at::date) as days_since_last_contact
		FROM contacts c
		LEFT JOIN activities a ON c.id = a.contact_id
		WHERE c.organization_id = $1 AND c.deleted_at IS NULL
	`

	args := []interface{}{orgID}

	// Add contact type filter
	if contactType != "all" {
		switch contactType {
		case "customers":
			query += " AND c.is_customer = true"
		case "vendors":
			query += " AND c.is_vendor = true"
		case "leads":
			query += " AND c.is_lead = true"
		}
	}

	query += ` GROUP BY c.id, c.name, c.engagement_score ORDER BY c.engagement_score DESC LIMIT 15`

	rows, err := s.GetRepository().(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact engagement: %w", err)
	}
	defer rows.Close()

	var engagement []types.ContactEngagement
	for rows.Next() {
		var item types.ContactEngagement
		if err := rows.Scan(
			&item.ContactID,
			&item.ContactName,
			&item.EngagementScore,
			&item.LastActivity,
			&item.DaysSinceLastContact,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contact engagement: %w", err)
		}
		// Set recommended action based on engagement
		item.RecommendedAction = getRecommendedAction(item)
		engagement = append(engagement, item)
	}

	return engagement, nil
}

// getRecommendedAction suggests actions based on contact engagement
func getRecommendedAction(engagement types.ContactEngagement) string {
	if engagement.DaysSinceLastContact > 30 {
		return "Re-engage - schedule follow-up"
	}
	if engagement.EngagementScore < 50 {
		return "Improve engagement - send personalized content"
	}
	if engagement.LastActivity == "email" {
		return "Follow up with call"
	}
	return "Maintain engagement - regular check-ins"
}

// getCachedDashboard retrieves cached dashboard data
func (s *ContactServiceV2) getCachedDashboard(ctx context.Context, orgID uuid.UUID, timeRange string, cacheKey string) (interface{}, bool) {
	// In a real implementation, this would check a cache (Redis, in-memory, etc.)
	// For now, return false to indicate no cache
	return nil, false
}

// setDashboardCache stores dashboard data in cache
func (s *ContactServiceV2) setDashboardCache(ctx context.Context, orgID uuid.UUID, timeRange string, cacheKey string, data interface{}) {
	// In a real implementation, this would store in cache with TTL
	// For now, do nothing
}

// Helper function to safely dereference string pointers
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
