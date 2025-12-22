package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
)

// MockContactValidationRepository is a mock implementation
type MockContactValidationRepository struct {
	mock.Mock
}

func (m *MockContactValidationRepository) CreateValidationRule(ctx context.Context, rule *types.ContactValidationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockContactValidationRepository) GetValidationRule(ctx context.Context, id uuid.UUID) (*types.ContactValidationRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactValidationRule), args.Error(1)
}

func (m *MockContactValidationRepository) UpdateValidationRule(ctx context.Context, rule *types.ContactValidationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockContactValidationRepository) DeleteValidationRule(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactValidationRepository) ListValidationRules(ctx context.Context, filter types.ValidationRuleFilter) ([]*types.ContactValidationRule, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactValidationRule), args.Error(1)
}

// MockAuthorizationService is a mock for auth
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CanRead(ctx context.Context, userID, orgID uuid.UUID, resource string) bool {
	args := m.Called(ctx, userID, orgID, resource)
	return args.Bool(0)
}

func (m *MockAuthorizationService) CanWrite(ctx context.Context, userID, orgID uuid.UUID, resource string) bool {
	args := m.Called(ctx, userID, orgID, resource)
	return args.Bool(0)
}

func (m *MockAuthorizationService) CanDelete(ctx context.Context, userID, orgID uuid.UUID, resource string) bool {
	args := m.Called(ctx, userID, orgID, resource)
	return args.Bool(0)
}

// MockContactRepository for validation service
type MockContactRepositoryForValidation struct {
	mock.Mock
}

func (m *MockContactRepositoryForValidation) GetContact(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Contact), args.Error(1)
}

func (m *MockContactRepositoryForValidation) CreateContact(ctx context.Context, contact *types.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepositoryForValidation) UpdateContact(ctx context.Context, contact *types.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepositoryForValidation) ListContacts(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Contact), args.Error(1)
}

// MockEventPublisher for tests
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event interface{}) {
	m.Called(ctx, event)
}

func (m *MockEventPublisher) Subscribe(eventType string, handler interface{}) {
	m.Called(eventType, handler)
}

func TestContactValidationService_ValidateContact_RequiredField(t *testing.T) {
	// Setup
	mockRepo := new(MockContactValidationRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactValidationService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Create a required field rule
	rule := &types.ContactValidationRule{
		ID:               uuid.New(),
		OrganizationID:   orgID,
		Field:            "email",
		RuleType:         "required",
		Severity:         "error",
		IsActive:         true,
		ValidationConfig: types.JSONBMap{},
	}

	// Mock auth
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)

	// Mock validation rules retrieval
	mockRepo.On("ListValidationRules", ctx, mock.MatchedBy(func(f types.ValidationRuleFilter) bool {
		return f.OrganizationID == orgID
	})).Return([]*types.ContactValidationRule{rule}, nil)

	// Test with missing email (should fail validation)
	req := types.ContactValidateRequest{
		ContactData: types.Contact{
			OrganizationID: orgID,
			FirstName:      strPtr("John"),
			LastName:       strPtr("Doe"),
			// Email missing
		},
	}

	result, err := svc.ValidateContact(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "email", result.Errors[0].Field)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactValidationService_ValidateContact_FormatValidation(t *testing.T) {
	// Setup
	mockRepo := new(MockContactValidationRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactValidationService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Create email format rule
	rule := &types.ContactValidationRule{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Field:          "email",
		RuleType:       "format",
		Severity:       "error",
		IsActive:       true,
		ValidationConfig: types.JSONBMap{
			"pattern": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		},
	}

	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("ListValidationRules", ctx, mock.Anything).Return([]*types.ContactValidationRule{rule}, nil)

	// Test with invalid email format
	invalidEmail := "not-an-email"
	req := types.ContactValidateRequest{
		ContactData: types.Contact{
			OrganizationID: orgID,
			Email:          &invalidEmail,
		},
	}

	result, err := svc.ValidateContact(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 1)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactValidationService_ValidateContact_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockContactValidationRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactValidationService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// No rules
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("ListValidationRules", ctx, mock.Anything).Return([]*types.ContactValidationRule{}, nil)

	// Test with valid contact
	validEmail := "john@example.com"
	req := types.ContactValidateRequest{
		ContactData: types.Contact{
			OrganizationID: orgID,
			Email:          &validEmail,
			FirstName:      strPtr("John"),
			LastName:       strPtr("Doe"),
		},
	}

	result, err := svc.ValidateContact(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsValid)
	assert.Len(t, result.Errors, 0)
	assert.Greater(t, result.QualityScore, 0)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactValidationService_CalculateDataQualityScore(t *testing.T) {
	testCases := []struct {
		name          string
		contact       types.Contact
		expectedScore int
	}{
		{
			name: "Complete contact - high score",
			contact: types.Contact{
				Email:     strPtr("john@example.com"),
				Phone:     strPtr("+1234567890"),
				FirstName: strPtr("John"),
				LastName:  strPtr("Doe"),
				Company:   strPtr("Acme Inc"),
				City:      strPtr("New York"),
			},
			expectedScore: 80, // Should be high
		},
		{
			name: "Minimal contact - low score",
			contact: types.Contact{
				FirstName: strPtr("John"),
			},
			expectedScore: 20, // Should be low
		},
		{
			name:          "Empty contact - zero score",
			contact:       types.Contact{},
			expectedScore: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockContactValidationRepository)
			mockContactRepo := new(MockContactRepositoryForValidation)
			mockAuth := new(MockAuthorizationService)
			mockEvents := new(MockEventPublisher)

			svc := service.NewContactValidationService(
				mockRepo,
				mockContactRepo,
				mockAuth,
				mockEvents,
			)

			orgID := uuid.New()
			userID := uuid.New()
			ctx := context.WithValue(context.Background(), "user_id", userID)

			mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
			mockRepo.On("ListValidationRules", ctx, mock.Anything).Return([]*types.ContactValidationRule{}, nil)

			req := types.ContactValidateRequest{
				ContactData: tc.contact,
			}

			result, err := svc.ValidateContact(ctx, orgID, req)

			require.NoError(t, err)
			require.NotNil(t, result)
			if tc.expectedScore > 0 {
				assert.GreaterOrEqual(t, result.QualityScore, tc.expectedScore-20)
			} else {
				assert.Equal(t, tc.expectedScore, result.QualityScore)
			}
		})
	}
}

func TestContactValidationService_Unauthorized(t *testing.T) {
	// Setup
	mockRepo := new(MockContactValidationRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactValidationService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock unauthorized access
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(false)

	req := types.ContactValidateRequest{
		ContactData: types.Contact{
			OrganizationID: orgID,
		},
	}

	result, err := svc.ValidateContact(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "ListValidationRules")
}

// Helper function
func strPtr(s string) *string {
	return &s
}
