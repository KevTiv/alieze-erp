package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/internal/testutils"
)

func TestBulkCreateContacts(t *testing.T) {
	// Setup
	mockRepo := testutils.NewMockContactRepository()
	mockAuth := testutils.NewMockAuthService()
	svc := service.NewContactServiceV2(mockRepo, mockAuth, service.ServiceOptions{})

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name     string
		requests []service.ContactRequest
		setup    func()
		expect   func(t *testing.T, results []*types.Contact, errors []error)
	}{
		{
			name: "Success - single contact",
			requests: []service.ContactRequest{
				{
					Name:           "John Doe",
					Email:          stringPtr("john@example.com"),
					OrganizationID: orgID,
					IsCustomer:     true,
				},
			},
			setup: func() {
				mockRepo.WithCreateFunc(func(ctx context.Context, contact types.Contact) (*types.Contact, error) {
					return &contact, nil
				})
			},
			expect: func(t *testing.T, results []*types.Contact, errors []error) {
				assert.Len(t, results, 1)
				assert.Len(t, errors, 0)
				assert.Equal(t, "John Doe", results[0].Name)
			},
		},
		{
			name: "Success - multiple contacts",
			requests: []service.ContactRequest{
				{
					Name:           "John Doe",
					OrganizationID: orgID,
					IsCustomer:     true,
				},
				{
					Name:           "Jane Smith",
					OrganizationID: orgID,
					IsVendor:       true,
				},
			},
			setup: func() {
				mockRepo.WithCreateFunc(func(ctx context.Context, contact types.Contact) (*types.Contact, error) {
					return &contact, nil
				})
			},
			expect: func(t *testing.T, results []*types.Contact, errors []error) {
				assert.Len(t, results, 2)
				assert.Len(t, errors, 0)
				assert.Equal(t, "John Doe", results[0].Name)
				assert.Equal(t, "Jane Smith", results[1].Name)
			},
		},
		{
			name: "Partial success - some failures",
			requests: []service.ContactRequest{
				{
					Name:           "Valid Contact",
					OrganizationID: orgID,
					IsCustomer:     true,
				},
				{
					Name:           "", // Invalid - missing name
					OrganizationID: orgID,
				},
				{
					Name:           "Another Valid",
					OrganizationID: orgID,
					IsVendor:       true,
				},
			},
			setup: func() {
				mockRepo.WithCreateFunc(func(ctx context.Context, contact types.Contact) (*types.Contact, error) {
					if contact.Name == "" {
						return nil, assert.AnError
					}
					return &contact, nil
				})
			},
			expect: func(t *testing.T, results []*types.Contact, errors []error) {
				assert.Len(t, results, 2)
				assert.Len(t, errors, 1)
				assert.Equal(t, "Valid Contact", results[0].Name)
				assert.Equal(t, "Another Valid", results[1].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			results, errors := svc.BulkCreateContacts(context.Background(), tc.requests)

			tc.expect(t, results, errors)

			// Verify mock expectations
			mockRepo.Mock.AssertExpectations(t)
		})
	}
}

func TestAdvancedSearchContacts(t *testing.T) {
	// Setup
	mockRepo := testutils.NewMockContactRepository()
	mockAuth := testutils.NewMockAuthService()
	svc := service.NewContactServiceV2(mockRepo, mockAuth, service.ServiceOptions{})

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name   string
		filter types.AdvancedContactFilter
		setup  func()
		expect func(t *testing.T, contacts []*types.Contact, total int, err error)
	}{
		{
			name: "Success - basic search",
			filter: types.AdvancedContactFilter{
				OrganizationID: orgID,
				SearchQuery:    "John",
				Page:           1,
				PageSize:       10,
			},
			setup: func() {
				mockRepo.WithFindAllFunc(func(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
					contact := &types.Contact{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "John Doe",
						Email:          stringPtr("john@example.com"),
						IsCustomer:     true,
					}
					return []*types.Contact{contact}, nil
				})
				mockRepo.WithCountFunc(func(ctx context.Context, filter types.ContactFilter) (int, error) {
					return 1, nil
				})
			},
			expect: func(t *testing.T, contacts []*types.Contact, total int, err error) {
				require.NoError(t, err)
				assert.Len(t, contacts, 1)
				assert.Equal(t, 1, total)
				assert.Equal(t, "John Doe", contacts[0].Name)
			},
		},
		{
			name: "Success - with score range",
			filter: types.AdvancedContactFilter{
				OrganizationID: orgID,
				ScoreRange: struct {
					Min int `json:"min,omitempty"`
					Max int `json:"max,omitempty"`
				}{
					Min: 50,
					Max: 100,
				},
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				mockRepo.WithFindAllFunc(func(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
					contact1 := &types.Contact{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "High Score Contact",
						IsCustomer:     true,
					}
					contact2 := &types.Contact{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "Another Contact",
						IsVendor:       true,
					}
					return []*types.Contact{contact1, contact2}, nil
				})
				mockRepo.WithCountFunc(func(ctx context.Context, filter types.ContactFilter) (int, error) {
					return 2, nil
				})
			},
			expect: func(t *testing.T, contacts []*types.Contact, total int, err error) {
				require.NoError(t, err)
				assert.Len(t, contacts, 2)
				assert.Equal(t, 2, total)
			},
		},
		{
			name: "Error - invalid organization",
			filter: types.AdvancedContactFilter{
				OrganizationID: uuid.Nil, // Invalid
				Page:           1,
				PageSize:       10,
			},
			setup: func() {
				// No setup needed - should fail validation
			},
			expect: func(t *testing.T, contacts []*types.Contact, total int, err error) {
				require.Error(t, err)
				assert.Nil(t, contacts)
				assert.Equal(t, 0, total)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			contacts, total, err := svc.AdvancedSearchContacts(context.Background(), tc.filter)

			tc.expect(t, contacts, total, err)

			// Verify mock expectations
			mockRepo.Mock.AssertExpectations(t)
		})
	}
}

func TestGetCRMDashboard(t *testing.T) {
	// Setup
	mockRepo := testutils.NewMockContactRepository()
	mockAuth := testutils.NewMockAuthService()
	svc := service.NewContactServiceV2(mockRepo, mockAuth, service.ServiceOptions{})

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name      string
		timeRange string
		setup     func()
		expect    func(t *testing.T, dashboard *types.CRMDashboard, err error)
	}{
		{
			name:      "Success - 30 day dashboard",
			timeRange: "30d",
			setup: func() {
				mockRepo.WithFindAllFunc(func(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
					contact := &types.Contact{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "Top Contact",
						Email:          stringPtr("top@example.com"),
						IsCustomer:     true,
					}
					return []*types.Contact{contact}, nil
				})

				// Mock other required methods
				mockRepo.WithCountFunc(func(ctx context.Context, filter types.ContactFilter) (int, error) {
					return 10, nil
				})
			},
			expect: func(t *testing.T, dashboard *types.CRMDashboard, err error) {
				require.NoError(t, err)
				assert.NotNil(t, dashboard)
				assert.Equal(t, "30d", dashboard.TimeRange)
				assert.NotEmpty(t, dashboard.Summary)
				assert.NotEmpty(t, dashboard.Trends)
			},
		},
		{
			name:      "Success - 7 day dashboard",
			timeRange: "7d",
			setup: func() {
				mockRepo.WithFindAllFunc(func(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
					return []*types.Contact{}, nil
				})
				mockRepo.WithCountFunc(func(ctx context.Context, filter types.ContactFilter) (int, error) {
					return 5, nil
				})
			},
			expect: func(t *testing.T, dashboard *types.CRMDashboard, err error) {
				require.NoError(t, err)
				assert.NotNil(t, dashboard)
				assert.Equal(t, "7d", dashboard.TimeRange)
			},
		},
		{
			name:      "Error - invalid time range",
			timeRange: "invalid",
			setup: func() {
				// No setup needed - should fail validation
			},
			expect: func(t *testing.T, dashboard *types.CRMDashboard, err error) {
				require.Error(t, err)
				assert.Nil(t, dashboard)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			dashboard, err := svc.GetCRMDashboard(context.Background(), orgID, tc.timeRange)

			tc.expect(t, dashboard, err)

			// Verify mock expectations
			mockRepo.Mock.AssertExpectations(t)
		})
	}
}

func TestGetActivityDashboard(t *testing.T) {
	// Setup
	mockRepo := testutils.NewMockContactRepository()
	mockAuth := testutils.NewMockAuthService()
	svc := service.NewContactServiceV2(mockRepo, mockAuth, service.ServiceOptions{})

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name        string
		contactType string
		timeRange   string
		setup       func()
		expect      func(t *testing.T, dashboard *types.ActivityDashboard, err error)
	}{
		{
			name:        "Success - all contacts, 30 days",
			contactType: "all",
			timeRange:   "30d",
			setup: func() {
				mockRepo.WithFindAllFunc(func(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
					contact := &types.Contact{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "Active Contact",
						IsCustomer:     true,
					}
					return []*types.Contact{contact}, nil
				})
			},
			expect: func(t *testing.T, dashboard *types.ActivityDashboard, err error) {
				require.NoError(t, err)
				assert.NotNil(t, dashboard)
				assert.Equal(t, "all", dashboard.ContactType)
				assert.Equal(t, "30d", dashboard.TimeRange)
			},
		},
		{
			name:        "Success - customers only, 7 days",
			contactType: "customers",
			timeRange:   "7d",
			setup: func() {
				mockRepo.WithFindAllFunc(func(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
					return []*types.Contact{}, nil
				})
			},
			expect: func(t *testing.T, dashboard *types.ActivityDashboard, err error) {
				require.NoError(t, err)
				assert.NotNil(t, dashboard)
				assert.Equal(t, "customers", dashboard.ContactType)
				assert.Equal(t, "7d", dashboard.TimeRange)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			dashboard, err := svc.GetActivityDashboard(context.Background(), orgID, tc.contactType, tc.timeRange)

			tc.expect(t, dashboard, err)

			// Verify mock expectations
			mockRepo.Mock.AssertExpectations(t)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
