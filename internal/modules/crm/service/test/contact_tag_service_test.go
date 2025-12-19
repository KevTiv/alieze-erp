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
	"github.com/KevTiv/alieze-erp/pkg/events"
)

// MockContactTagRepository is a mock implementation of ContactTagRepository
type MockContactTagRepository struct {
	mock.Mock
}

func (m *MockContactTagRepository) Create(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error) {
	args := m.Called(ctx, tag)
	return args.Get(0).(*types.ContactTag), args.Error(1)
}

func (m *MockContactTagRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.ContactTag, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.ContactTag), args.Error(1)
}

func (m *MockContactTagRepository) FindAll(ctx context.Context, filter types.ContactTagFilter) ([]*types.ContactTag, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.ContactTag), args.Error(1)
}

func (m *MockContactTagRepository) Update(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error) {
	args := m.Called(ctx, tag)
	return args.Get(0).(*types.ContactTag), args.Error(1)
}

func (m *MockContactTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactTagRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]types.ContactTag, error) {
	args := m.Called(ctx, contactID)
	return args.Get(0).([]types.ContactTag), args.Error(1)
}

func (m *MockContactTagRepository) Count(ctx context.Context, filter types.ContactTagFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int), args.Error(1)
}

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GetOrganizationID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAuthService) GetUserID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAuthService) CheckPermission(ctx context.Context, permission string) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func TestContactTagService_CreateContactTag(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	orgID := uuid.New()
	tag := types.ContactTag{
		Name:  "VIP Customer",
		Color: 0xFF0000,
	}

	createdTag := &types.ContactTag{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "VIP Customer",
		Color:          0xFF0000,
	}

	// Mock expectations
	mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:create").Return(nil)
	mockAuth.On("GetOrganizationID", mock.Anything).Return(orgID, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(t types.ContactTag) bool {
		return t.Name == "VIP Customer" && t.Color == 0xFF0000 && t.OrganizationID == orgID
	})).Return(createdTag, nil)

	// Test
	result, err := service.CreateContactTag(context.Background(), tag)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, createdTag.ID, result.ID)
	assert.Equal(t, createdTag.Name, result.Name)
	assert.Equal(t, createdTag.Color, result.Color)
	assert.Equal(t, orgID, result.OrganizationID)

	// Verify mock expectations
	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactTagService_CreateContactTag_ValidationError(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	// Test with empty name
	tag := types.ContactTag{
		Name:  "",
		Color: 0xFF0000,
	}

	// Test
	result, err := service.CreateContactTag(context.Background(), tag)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid contact tag")

	// Verify no repository calls were made
	mockRepo.AssertNotCalled(t, "Create")
}

func TestContactTagService_GetContactTag(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	orgID := uuid.New()
	tagID := uuid.New()
	expectedTag := &types.ContactTag{
		ID:             tagID,
		OrganizationID: orgID,
		Name:           "VIP Customer",
		Color:          0xFF0000,
	}

	// Mock expectations
	mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:read").Return(nil)
	mockAuth.On("GetOrganizationID", mock.Anything).Return(orgID, nil)
	mockRepo.On("FindByID", mock.Anything, tagID).Return(expectedTag, nil)

	// Test
	result, err := service.GetContactTag(context.Background(), tagID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedTag.ID, result.ID)
	assert.Equal(t, expectedTag.Name, result.Name)

	// Verify mock expectations
	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactTagService_ListContactTags(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	orgID := uuid.New()
	tags := []types.ContactTag{
		{ID: uuid.New(), OrganizationID: orgID, Name: "VIP Customer", Color: 0xFF0000},
		{ID: uuid.New(), OrganizationID: orgID, Name: "New Lead", Color: 0x00FF00},
	}

	// Mock expectations
	mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:read").Return(nil)
	mockAuth.On("GetOrganizationID", mock.Anything).Return(orgID, nil)
	// Convert tags to pointers for the mock
	tagPtrs := make([]*types.ContactTag, len(tags))
	for i := range tags {
		tagPtrs[i] = &tags[i]
	}
	mockRepo.On("FindAll", mock.Anything, mock.MatchedBy(func(f types.ContactTagFilter) bool {
		return f.OrganizationID == orgID
	})).Return(tagPtrs, nil)

	// Test
	result, err := service.ListContactTags(context.Background(), types.ContactTagFilter{})

	// Assertions
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, tags[0].ID, result[0].ID)
	assert.Equal(t, tags[1].ID, result[1].ID)

	// Verify mock expectations
	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactTagService_UpdateContactTag(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	orgID := uuid.New()
	tagID := uuid.New()

	existingTag := &types.ContactTag{
		ID:             tagID,
		OrganizationID: orgID,
		Name:           "VIP Customer",
		Color:          0xFF0000,
	}

	updatedTag := &types.ContactTag{
		ID:    tagID,
		Name:  "Premium Customer",
		Color: 0x00FF00,
	}

	// Mock expectations
	mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:update").Return(nil)
	mockAuth.On("GetOrganizationID", mock.Anything).Return(orgID, nil)
	mockRepo.On("FindByID", mock.Anything, tagID).Return(existingTag, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(t types.ContactTag) bool {
		return t.ID == tagID && t.Name == "Premium Customer" && t.Color == 0x00FF00
	})).Return(updatedTag, nil)

	// Test
	result, err := service.UpdateContactTag(context.Background(), *updatedTag)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, updatedTag.Name, result.Name)
	assert.Equal(t, updatedTag.Color, result.Color)

	// Verify mock expectations
	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactTagService_DeleteContactTag(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	orgID := uuid.New()
	tagID := uuid.New()

	existingTag := &types.ContactTag{
		ID:             tagID,
		OrganizationID: orgID,
		Name:           "VIP Customer",
		Color:          0xFF0000,
	}

	// Mock expectations
	mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:delete").Return(nil)
	mockAuth.On("GetOrganizationID", mock.Anything).Return(orgID, nil)
	mockRepo.On("FindByID", mock.Anything, tagID).Return(existingTag, nil)
	mockRepo.On("Delete", mock.Anything, tagID).Return(nil)

	// Test
	err := service.DeleteContactTag(context.Background(), tagID)

	// Assertions
	require.NoError(t, err)

	// Verify mock expectations
	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactTagService_OrganizationAccessControl(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	orgID := uuid.New()
	otherOrgID := uuid.New()
	tagID := uuid.New()

	existingTag := &types.ContactTag{
		ID:             tagID,
		OrganizationID: otherOrgID, // Different organization
		Name:           "VIP Customer",
		Color:          0xFF0000,
	}

	// Mock expectations
	mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:update").Return(nil)
	mockAuth.On("GetOrganizationID", mock.Anything).Return(orgID, nil)
	mockRepo.On("FindByID", mock.Anything, tagID).Return(existingTag, nil)

	// Test
	_, err := service.UpdateContactTag(context.Background(), types.ContactTag{ID: tagID, Name: "Updated"})

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to organization")

	// Verify repository update was not called
	mockRepo.AssertNotCalled(t, "Update")
}

func TestContactTagService_Validation(t *testing.T) {
	// Setup
	mockRepo := new(MockContactTagRepository)
	mockAuth := new(MockAuthService)
	eventBus := events.NewBus(false)

	service := service.NewContactTagService(mockRepo, mockAuth, eventBus)

	testCases := []struct {
		name        string
		tag         types.ContactTag
		expectedErr string
	}{
		{
			name:        "Empty name",
			tag:         types.ContactTag{Name: "", Color: 0xFF0000},
			expectedErr: "name is required",
		},
		{
			name:        "Name too long",
			tag:         types.ContactTag{Name: "This is a very long name that exceeds the maximum allowed length of 100 characters and should be rejected by the validation logic in the service layer", Color: 0xFF0000},
			expectedErr: "name must be 100 characters or less",
		},
		{
			name:        "Invalid color (negative)",
			tag:         types.ContactTag{Name: "Test", Color: -1},
			expectedErr: "color must be a valid RGB value",
		},
		{
			name:        "Invalid color (too large)",
			tag:         types.ContactTag{Name: "Test", Color: 16777216},
			expectedErr: "color must be a valid RGB value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock expectations - should not call repository due to validation failure
			mockAuth.On("CheckPermission", mock.Anything, "crm:contact_tags:create").Return(nil)
			mockAuth.On("GetOrganizationID", mock.Anything).Return(uuid.New(), nil)

			// Test
			_, err := service.CreateContactTag(context.Background(), tc.tag)

			// Assertions
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)

			// Verify no repository calls were made
			mockRepo.AssertNotCalled(t, "Create")

			// Clean up mocks for next test case
			mockAuth.ExpectedCalls = nil
		})
	}
}
