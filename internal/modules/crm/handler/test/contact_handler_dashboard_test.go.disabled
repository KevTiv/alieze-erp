package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/handler"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/internal/testutils"
)

func TestBulkCreateContactsHandler(t *testing.T) {
	// Setup
	mockService := testutils.NewMockContactService()
	handler := handler.NewContactHandler(mockService)
	router := httprouter.New()
	handler.RegisterRoutes(router)

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name           string
		requestBody    interface{}
		setup          func()
		expectedStatus int
		expect         func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Success - create multiple contacts",
			requestBody: []service.ContactRequest{
				{
					Name:           "John Doe",
					Email:          stringPtr("john@example.com"),
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
				contacts := []*types.Contact{
					{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "John Doe",
						Email:          stringPtr("john@example.com"),
						IsCustomer:     true,
					},
					{
						ID:             uuid.Must(uuid.NewV7()),
						OrganizationID: orgID,
						Name:           "Jane Smith",
						IsVendor:       true,
					},
				}
				mockService.WithBulkCreateContactsFunc(func(ctx context.Context, requests []service.ContactRequest) ([]*types.Contact, []error) {
					return contacts, nil
				})
			},
			expectedStatus: http.StatusCreated,
			expect: func(t *testing.T, response map[string]interface{}) {
				success := response["success"].([]interface{})
				errors := response["errors"].([]interface{})
				assert.Len(t, success, 2)
				assert.Len(t, errors, 0)
			},
		},
		{
			name: "Partial success - some errors",
			requestBody: []service.ContactRequest{
				{
					Name:           "Valid Contact",
					OrganizationID: orgID,
					IsCustomer:     true,
				},
				{
					Name:           "", // Invalid
					OrganizationID: orgID,
				},
			},
			setup: func() {
				contact := &types.Contact{
					ID:             uuid.Must(uuid.NewV7()),
					OrganizationID: orgID,
					Name:           "Valid Contact",
					IsCustomer:     true,
				}
				error := assert.AnError
				mockService.WithBulkCreateContactsFunc(func(ctx context.Context, requests []service.ContactRequest) ([]*types.Contact, []error) {
					return []*types.Contact{contact}, []error{error}
				})
			},
			expectedStatus: http.StatusCreated,
			expect: func(t *testing.T, response map[string]interface{}) {
				success := response["success"].([]interface{})
				errors := response["errors"].([]interface{})
				assert.Len(t, success, 1)
				assert.Len(t, errors, 1)
			},
		},
		{
			name:           "Error - invalid request body",
			requestBody:    "invalid json",
			setup:          func() {},
			expectedStatus: http.StatusBadRequest,
			expect: func(t *testing.T, response map[string]interface{}) {
				// Expect error response
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			// Create request
			body, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/crm/contacts/bulk", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Add organization ID to context
			ctx := context.WithValue(req.Context(), "organizationID", orgID)
			req = req.WithContext(ctx)

			// Record response
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check status
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Parse response
			if tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				tc.expect(t, response)
			}

			// Verify mock expectations
			mockService.Mock.AssertExpectations(t)
		})
	}
}

func TestAdvancedSearchContactsHandler(t *testing.T) {
	// Setup
	mockService := testutils.NewMockContactService()
	handler := handler.NewContactHandler(mockService)
	router := httprouter.New()
	handler.RegisterRoutes(router)

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name           string
		requestBody    types.AdvancedContactFilter
		setup          func()
		expectedStatus int
		expect         func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Success - basic search",
			requestBody: types.AdvancedContactFilter{
				OrganizationID: orgID,
				SearchQuery:    "John",
				Page:           1,
				PageSize:       10,
			},
			setup: func() {
				contact := &types.Contact{
					ID:             uuid.Must(uuid.NewV7()),
					OrganizationID: orgID,
					Name:           "John Doe",
					Email:          stringPtr("john@example.com"),
					IsCustomer:     true,
				}
				mockService.WithAdvancedSearchContactsFunc(func(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error) {
					return []*types.Contact{contact}, 1, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, response map[string]interface{}) {
				data := response["data"].([]interface{})
				total := response["total"].(float64)
				page := response["page"].(float64)
				pageSize := response["pageSize"].(float64)
				assert.Len(t, data, 1)
				assert.Equal(t, 1.0, total)
				assert.Equal(t, 1.0, page)
				assert.Equal(t, 10.0, pageSize)
			},
		},
		{
			name: "Success - with score range",
			requestBody: types.AdvancedContactFilter{
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
				mockService.WithAdvancedSearchContactsFunc(func(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error) {
					return []*types.Contact{contact1, contact2}, 2, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, response map[string]interface{}) {
				data := response["data"].([]interface{})
				total := response["total"].(float64)
				assert.Len(t, data, 2)
				assert.Equal(t, 2.0, total)
			},
		},
		{
			name: "Error - invalid request body",
			requestBody: types.AdvancedContactFilter{
				OrganizationID: uuid.Nil, // Invalid
			},
			setup: func() {
				mockService.WithAdvancedSearchContactsFunc(func(ctx context.Context, filter types.AdvancedContactFilter) ([]*types.Contact, int, error) {
					return nil, 0, assert.AnError
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expect: func(t *testing.T, response map[string]interface{}) {
				// Expect error response
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			// Create request
			body, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/crm/contacts/search/advanced", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Add organization ID to context
			ctx := context.WithValue(req.Context(), "organizationID", orgID)
			req = req.WithContext(ctx)

			// Record response
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check status
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Parse response
			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				tc.expect(t, response)
			}

			// Verify mock expectations
			mockService.Mock.AssertExpectations(t)
		})
	}
}

func TestGetCRMDashboardHandler(t *testing.T) {
	// Setup
	mockService := testutils.NewMockContactService()
	handler := handler.NewContactHandler(mockService)
	router := httprouter.New()
	handler.RegisterRoutes(router)

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name           string
		queryParams    string
		setup          func()
		expectedStatus int
		expect         func(t *testing.T, dashboard *types.CRMDashboard)
	}{
		{
			name:        "Success - 30 day dashboard",
			queryParams: "?time_range=30d",
			setup: func() {
				dashboard := &types.CRMDashboard{
					TimeRange: "30d",
					Summary: types.DashboardSummary{
						TotalContacts:     10,
						NewContacts:       2,
						ActiveContacts:    8,
						AtRiskContacts:    1,
						HighValueContacts: 3,
					},
					Trends: types.DashboardTrends{
						ContactGrowth: make([]types.TrendDataPoint, 0),
						Engagement:    make([]types.TrendDataPoint, 0),
						ResponseRate:  make([]types.TrendDataPoint, 0),
					},
					TopContacts:      make([]types.TopContact, 0),
					RecentActivities: make([]types.RecentActivity, 0),
				}
				mockService.WithGetCRMDashboardFunc(func(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error) {
					return dashboard, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, dashboard *types.CRMDashboard) {
				assert.Equal(t, "30d", dashboard.TimeRange)
				assert.Equal(t, 10, dashboard.Summary.TotalContacts)
			},
		},
		{
			name:        "Success - 7 day dashboard (default)",
			queryParams: "", // No time_range specified, should default to 30d
			setup: func() {
				dashboard := &types.CRMDashboard{
					TimeRange: "30d", // Default
					Summary: types.DashboardSummary{
						TotalContacts: 5,
					},
				}
				mockService.WithGetCRMDashboardFunc(func(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error) {
					return dashboard, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, dashboard *types.CRMDashboard) {
				assert.Equal(t, "30d", dashboard.TimeRange) // Should default to 30d
			},
		},
		{
			name:        "Error - invalid time range",
			queryParams: "?time_range=invalid",
			setup: func() {
				mockService.WithGetCRMDashboardFunc(func(ctx context.Context, orgID uuid.UUID, timeRange string) (*types.CRMDashboard, error) {
					return nil, assert.AnError
				})
			},
			expectedStatus: http.StatusInternalServerError,
			expect: func(t *testing.T, dashboard *types.CRMDashboard) {
				// Should not reach here due to error
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			// Create request
			req, _ := http.NewRequest("GET", "/api/crm/dashboard"+tc.queryParams, nil)

			// Add organization ID to context
			ctx := context.WithValue(req.Context(), "organizationID", orgID)
			req = req.WithContext(ctx)

			// Record response
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check status
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Parse response
			if tc.expectedStatus == http.StatusOK {
				var dashboard types.CRMDashboard
				err := json.NewDecoder(rr.Body).Decode(&dashboard)
				require.NoError(t, err)
				tc.expect(t, &dashboard)
			}

			// Verify mock expectations
			mockService.Mock.AssertExpectations(t)
		})
	}
}

func TestGetActivityDashboardHandler(t *testing.T) {
	// Setup
	mockService := testutils.NewMockContactService()
	handler := handler.NewContactHandler(mockService)
	router := httprouter.New()
	handler.RegisterRoutes(router)

	orgID := uuid.Must(uuid.NewV7())

	testCases := []struct {
		name           string
		queryParams    string
		setup          func()
		expectedStatus int
		expect         func(t *testing.T, dashboard *types.ActivityDashboard)
	}{
		{
			name:        "Success - all contacts, 30 days",
			queryParams: "?contact_type=all&time_range=30d",
			setup: func() {
				dashboard := &types.ActivityDashboard{
					TimeRange:   "30d",
					ContactType: "all",
					ActivitySummary: types.ActivitySummary{
						TotalActivities: 10,
						ActivityTypes:   map[string]int{"call": 3, "email": 5, "meeting": 2},
					},
				}
				mockService.WithGetActivityDashboardFunc(func(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error) {
					return dashboard, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, dashboard *types.ActivityDashboard) {
				assert.Equal(t, "all", dashboard.ContactType)
				assert.Equal(t, "30d", dashboard.TimeRange)
				assert.Equal(t, 10, dashboard.ActivitySummary.TotalActivities)
			},
		},
		{
			name:        "Success - customers only, 7 days",
			queryParams: "?contact_type=customers&time_range=7d",
			setup: func() {
				dashboard := &types.ActivityDashboard{
					TimeRange:   "7d",
					ContactType: "customers",
					ActivitySummary: types.ActivitySummary{
						TotalActivities: 5,
					},
				}
				mockService.WithGetActivityDashboardFunc(func(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error) {
					return dashboard, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, dashboard *types.ActivityDashboard) {
				assert.Equal(t, "customers", dashboard.ContactType)
				assert.Equal(t, "7d", dashboard.TimeRange)
			},
		},
		{
			name:        "Success - defaults",
			queryParams: "", // No parameters, should use defaults
			setup: func() {
				dashboard := &types.ActivityDashboard{
					TimeRange:   "30d", // Default time range
					ContactType: "all", // Default contact type
				}
				mockService.WithGetActivityDashboardFunc(func(ctx context.Context, orgID uuid.UUID, contactType string, timeRange string) (*types.ActivityDashboard, error) {
					return dashboard, nil
				})
			},
			expectedStatus: http.StatusOK,
			expect: func(t *testing.T, dashboard *types.ActivityDashboard) {
				assert.Equal(t, "all", dashboard.ContactType)
				assert.Equal(t, "30d", dashboard.TimeRange)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			// Create request
			req, _ := http.NewRequest("GET", "/api/crm/dashboard/activity"+tc.queryParams, nil)

			// Add organization ID to context
			ctx := context.WithValue(req.Context(), "organizationID", orgID)
			req = req.WithContext(ctx)

			// Record response
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check status
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Parse response
			if tc.expectedStatus == http.StatusOK {
				var dashboard types.ActivityDashboard
				err := json.NewDecoder(rr.Body).Decode(&dashboard)
				require.NoError(t, err)
				tc.expect(t, &dashboard)
			}

			// Verify mock expectations
			mockService.Mock.AssertExpectations(t)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
