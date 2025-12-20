package testutils

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/crm/base"

	"github.com/google/uuid"
)

// MockContactService implements a mock contact service for testing
type MockContactService struct {
	bulkCreateContactsFunc   func(ctx context.Context, requests []service.ContactRequest) ([]*types.Contact, []error)
	advancedSearchFunc       func(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error)
	getCRMDashboardFunc      func(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error)
	getActivityDashboardFunc func(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error)
}

// NewMockContactService creates a new mock contact service
func NewMockContactService() *MockContactService {
	return &MockContactService{}
}

// BulkCreateContacts implements the service method
func (m *MockContactService) BulkCreateContacts(ctx context.Context, requests []service.ContactRequest) ([]*types.Contact, []error) {
	if m.bulkCreateContactsFunc != nil {
		return m.bulkCreateContactsFunc(ctx, requests)
	}
	// Default behavior: return success for all requests
	var results []*types.Contact
	for _, req := range requests {
		contact := types.Contact{
			ID:             uuid.New(),
			OrganizationID: req.OrganizationID,
			Name:           req.Name,
			Email:          req.Email,
			Phone:          req.Phone,
			IsCustomer:     req.IsCustomer,
			IsVendor:       req.IsVendor,
			Street:         req.Street,
			City:           req.City,
			StateID:        req.StateID,
			CountryID:      req.CountryID,
		}
		results = append(results, &contact)
	}
	return results, nil
}

// AdvancedSearchContacts implements the service method
func (m *MockContactService) AdvancedSearchContacts(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error) {
	if m.advancedSearchFunc != nil {
		return m.advancedSearchFunc(ctx, filter)
	}
	// Default behavior: return empty results
	return []*types.Contact{}, 0, nil
}

// GetCRMDashboard implements the service method
func (m *MockContactService) GetCRMDashboard(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error) {
	if m.getCRMDashboardFunc != nil {
		return m.getCRMDashboardFunc(ctx, orgID, timeRange)
	}
	// Default behavior: return empty dashboard
	return &types.CRMDashboard{
		TimeRange:        timeRange,
		Summary:          types.DashboardSummary{},
		Trends:           types.DashboardTrends{},
		TopContacts:      []types.TopContact{},
		RecentActivities: []types.RecentActivity{},
	}, nil
}

// GetActivityDashboard implements the service method
func (m *MockContactService) GetActivityDashboard(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error) {
	if m.getActivityDashboardFunc != nil {
		return m.getActivityDashboardFunc(ctx, orgID, contactType, timeRange)
	}
	// Default behavior: return empty dashboard
	return &types.ActivityDashboard{
		TimeRange:         timeRange,
		ContactType:       contactType,
		ActivitySummary:   types.ActivitySummary{},
		RecentActivities:  []types.RecentActivity{},
		ContactEngagement: []types.ContactEngagement{},
	}, nil
}

// WithBulkCreateContactsFunc sets the mock behavior for BulkCreateContacts
func (m *MockContactService) WithBulkCreateContactsFunc(f func(ctx context.Context, requests []service.ContactRequest) ([]*types.Contact, []error)) *MockContactService {
	m.bulkCreateContactsFunc = f
	return m
}

// WithAdvancedSearchContactsFunc sets the mock behavior for AdvancedSearchContacts
func (m *MockContactService) WithAdvancedSearchContactsFunc(f func(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error)) *MockContactService {
	m.advancedSearchFunc = f
	return m
}

// WithGetCRMDashboardFunc sets the mock behavior for GetCRMDashboard
func (m *MockContactService) WithGetCRMDashboardFunc(f func(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error)) *MockContactService {
	m.getCRMDashboardFunc = f
	return m
}

// WithGetActivityDashboardFunc sets the mock behavior for GetActivityDashboard
func (m *MockContactService) WithGetActivityDashboardFunc(f func(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error)) *MockContactService {
	m.getActivityDashboardFunc = f
	return m
}

// Mock methods for interface compliance
func (m *MockContactService) CreateContact(ctx context.Context, req service.ContactRequest) (*types.Contact, error) {
	return &types.Contact{}, nil
}

func (m *MockContactService) GetContact(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
	return &types.Contact{}, nil
}

func (m *MockContactService) UpdateContact(ctx context.Context, id uuid.UUID, req service.ContactUpdateRequest) (*types.Contact, error) {
	return &types.Contact{}, nil
}

func (m *MockContactService) DeleteContact(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockContactService) ListContacts(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, int, error) {
	return []*types.Contact{}, 0, nil
}

func (m *MockContactService) CreateRelationship(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, req types.ContactRelationshipCreateRequest) (*types.ContactRelationship, error) {
	return &types.ContactRelationship{}, nil
}

func (m *MockContactService) ListRelationships(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, relationshipType string, limit int) ([]*types.ContactRelationship, error) {
	return []*types.ContactRelationship{}, nil
}

func (m *MockContactService) AddToSegments(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, req types.ContactSegmentationRequest) error {
	return nil
}

func (m *MockContactService) CalculateContactScore(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) (*types.ContactScore, error) {
	return &types.ContactScore{}, nil
}

func (m *MockContactService) GetAuthService() base.AuthService {
	return nil
}

func (m *MockContactService) GetRepository() interface{} {
	return nil
}

func (m *MockContactService) LogOperation(ctx context.Context, operation string, entityID uuid.UUID, details map[string]interface{}) {
	// No-op
}

func (m *MockContactService) PublishEvent(ctx context.Context, eventType string, data interface{}) {
	// No-op
}
