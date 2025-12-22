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

// MockContactMergeRepository is a mock implementation
type MockContactMergeRepository struct {
	mock.Mock
}

func (m *MockContactMergeRepository) FindPotentialDuplicates(ctx context.Context, orgID uuid.UUID, threshold int, limit int) ([]*types.ContactDuplicate, error) {
	args := m.Called(ctx, orgID, threshold, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactDuplicate), args.Error(1)
}

func (m *MockContactMergeRepository) CreateDuplicate(ctx context.Context, duplicate *types.ContactDuplicate) error {
	args := m.Called(ctx, duplicate)
	return args.Error(0)
}

func (m *MockContactMergeRepository) GetDuplicate(ctx context.Context, id uuid.UUID) (*types.ContactDuplicate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactDuplicate), args.Error(1)
}

func (m *MockContactMergeRepository) UpdateDuplicateStatus(ctx context.Context, id uuid.UUID, status, resolutionType string, resolvedBy uuid.UUID) error {
	args := m.Called(ctx, id, status, resolutionType, resolvedBy)
	return args.Error(0)
}

func (m *MockContactMergeRepository) ListDuplicates(ctx context.Context, filter types.DuplicateFilter) ([]*types.ContactDuplicate, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactDuplicate), args.Error(1)
}

func (m *MockContactMergeRepository) CountDuplicates(ctx context.Context, filter types.DuplicateFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockContactMergeRepository) MergeContacts(ctx context.Context, masterID, duplicateID uuid.UUID, fieldSelections map[string]string, mergedBy uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, masterID, duplicateID, fieldSelections, mergedBy, orgID)
	return args.Error(0)
}

func (m *MockContactMergeRepository) GetMergeHistory(ctx context.Context, contactID uuid.UUID) ([]*types.ContactMergeHistory, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactMergeHistory), args.Error(1)
}

func (m *MockContactMergeRepository) CalculateSimilarity(contact1, contact2 *types.Contact) int {
	args := m.Called(contact1, contact2)
	return args.Int(0)
}

func TestContactMergeService_DetectDuplicates(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	duplicates := []*types.ContactDuplicate{
		{
			ID:              uuid.New(),
			OrganizationID:  orgID,
			Contact1ID:      uuid.New(),
			Contact2ID:      uuid.New(),
			SimilarityScore: 85,
			MatchingFields:  []string{"email", "name"},
			Status:          "pending",
		},
	}

	threshold := 80
	limit := 100

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockMergeRepo.On("FindPotentialDuplicates", ctx, orgID, threshold, limit).Return(duplicates, nil)
	mockMergeRepo.On("CreateDuplicate", ctx, mock.AnythingOfType("*types.ContactDuplicate")).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.DetectDuplicatesRequest{
		Threshold: &threshold,
		Limit:     &limit,
	}

	result, err := svc.DetectDuplicates(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.TotalFound)
	assert.Equal(t, threshold, result.Threshold)

	mockAuth.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
}

func TestContactMergeService_CalculateSimilarity(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	contact1ID := uuid.New()
	contact2ID := uuid.New()

	email := "john@example.com"
	contact1 := &types.Contact{
		ID:             contact1ID,
		OrganizationID: orgID,
		Email:          &email,
		FirstName:      strPtr("John"),
		LastName:       strPtr("Doe"),
	}

	contact2 := &types.Contact{
		ID:             contact2ID,
		OrganizationID: orgID,
		Email:          &email,
		FirstName:      strPtr("John"),
		LastName:       strPtr("Doe"),
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, contact1ID).Return(contact1, nil)
	mockContactRepo.On("GetContact", ctx, contact2ID).Return(contact2, nil)
	mockMergeRepo.On("CalculateSimilarity", contact1, contact2).Return(90)

	// Test
	req := types.CalculateSimilarityRequest{
		Contact1ID: contact1ID,
		Contact2ID: contact2ID,
	}

	result, err := svc.CalculateSimilarity(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, contact1ID, result.Contact1ID)
	assert.Equal(t, contact2ID, result.Contact2ID)
	assert.Equal(t, 90, result.SimilarityScore)
	assert.True(t, result.IsDuplicate) // Score >= 80

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
}

func TestContactMergeService_MergeContacts(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	masterID := uuid.New()
	duplicateID := uuid.New()

	masterContact := &types.Contact{
		ID:             masterID,
		OrganizationID: orgID,
		Email:          strPtr("john@example.com"),
	}

	duplicateContact := &types.Contact{
		ID:             duplicateID,
		OrganizationID: orgID,
		Email:          strPtr("john.doe@example.com"),
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, masterID).Return(masterContact, nil)
	mockContactRepo.On("GetContact", ctx, duplicateID).Return(duplicateContact, nil)
	mockMergeRepo.On("MergeContacts", ctx, masterID, duplicateID, mock.AnythingOfType("map[string]string"), userID, orgID).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.MergeContactsRequest{
		MasterContactID:    masterID,
		DuplicateContactID: duplicateID,
		MergeStrategy:      "keep_master",
		FieldSelections:    map[string]string{},
	}

	err := svc.MergeContacts(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactMergeService_MergeContacts_SameContact(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	contactID := uuid.New()

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)

	// Test
	req := types.MergeContactsRequest{
		MasterContactID:    contactID,
		DuplicateContactID: contactID, // Same ID
		MergeStrategy:      "keep_master",
	}

	err := svc.MergeContacts(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be different")

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertNotCalled(t, "GetContact")
	mockMergeRepo.AssertNotCalled(t, "MergeContacts")
}

func TestContactMergeService_ResolveDuplicate(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	duplicateID := uuid.New()
	duplicate := &types.ContactDuplicate{
		ID:             duplicateID,
		OrganizationID: orgID,
		Contact1ID:     uuid.New(),
		Contact2ID:     uuid.New(),
		Status:         "pending",
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockMergeRepo.On("GetDuplicate", ctx, duplicateID).Return(duplicate, nil)
	mockMergeRepo.On("UpdateDuplicateStatus", ctx, duplicateID, "resolved", "false_positive", userID).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.ResolveDuplicateRequest{
		ResolutionType: "false_positive",
	}

	err := svc.ResolveDuplicate(ctx, orgID, duplicateID, req)

	// Assertions
	require.NoError(t, err)

	mockAuth.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactMergeService_ResolveDuplicate_InvalidType(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	duplicateID := uuid.New()
	duplicate := &types.ContactDuplicate{
		ID:             duplicateID,
		OrganizationID: orgID,
		Status:         "pending",
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockMergeRepo.On("GetDuplicate", ctx, duplicateID).Return(duplicate, nil)

	// Test with invalid resolution type
	req := types.ResolveDuplicateRequest{
		ResolutionType: "invalid_type",
	}

	err := svc.ResolveDuplicate(ctx, orgID, duplicateID, req)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid resolution type")

	mockAuth.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
	mockMergeRepo.AssertNotCalled(t, "UpdateDuplicateStatus")
}

func TestContactMergeService_ListDuplicates(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	duplicates := []*types.ContactDuplicate{
		{
			ID:              uuid.New(),
			OrganizationID:  orgID,
			SimilarityScore: 90,
		},
		{
			ID:              uuid.New(),
			OrganizationID:  orgID,
			SimilarityScore: 85,
		},
	}

	filter := types.DuplicateFilter{
		OrganizationID: orgID,
		Limit:          50,
		Offset:         0,
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockMergeRepo.On("ListDuplicates", ctx, filter).Return(duplicates, nil)
	mockMergeRepo.On("CountDuplicates", ctx, filter).Return(2, nil)

	// Test
	result, err := svc.ListDuplicates(ctx, orgID, filter)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Duplicates, 2)
	assert.Equal(t, 2, result.Total)

	mockAuth.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
}

func TestContactMergeService_GetMergeHistory(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	contactID := uuid.New()
	contact := &types.Contact{
		ID:             contactID,
		OrganizationID: orgID,
	}

	history := []*types.ContactMergeHistory{
		{
			ID:                 uuid.New(),
			OrganizationID:     orgID,
			MasterContactID:    contactID,
			DuplicateContactID: uuid.New(),
			MergedBy:           userID,
		},
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, contactID).Return(contact, nil)
	mockMergeRepo.On("GetMergeHistory", ctx, contactID).Return(history, nil)

	// Test
	result, err := svc.GetMergeHistory(ctx, orgID, contactID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, contactID, result[0].MasterContactID)

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockMergeRepo.AssertExpectations(t)
}

func TestContactMergeService_Unauthorized(t *testing.T) {
	// Setup
	mockMergeRepo := new(MockContactMergeRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactMergeService(
		mockMergeRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock unauthorized
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(false)

	// Test
	req := types.DetectDuplicatesRequest{}
	result, err := svc.DetectDuplicates(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")

	mockAuth.AssertExpectations(t)
	mockMergeRepo.AssertNotCalled(t, "FindPotentialDuplicates")
}
