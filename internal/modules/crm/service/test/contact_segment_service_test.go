package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContactSegmentRepository is a mock implementation of ContactSegmentRepository
type MockContactSegmentRepository struct {
	mock.Mock
}

func (m *MockContactSegmentRepository) CreateSegment(ctx context.Context, segment *types.ContactSegment) error {
	args := m.Called(ctx, segment)
	return args.Error(0)
}

func (m *MockContactSegmentRepository) GetSegment(ctx context.Context, id uuid.UUID) (*types.ContactSegment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactSegment), args.Error(1)
}

func (m *MockContactSegmentRepository) UpdateSegment(ctx context.Context, segment *types.ContactSegment) error {
	args := m.Called(ctx, segment)
	return args.Error(0)
}

func (m *MockContactSegmentRepository) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactSegmentRepository) ListSegments(ctx context.Context, filter types.SegmentFilter) ([]*types.ContactSegment, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.ContactSegment), args.Error(1)
}

func (m *MockContactSegmentRepository) CountSegments(ctx context.Context, filter types.SegmentFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockContactSegmentRepository) AddContactsToSegment(ctx context.Context, segmentID uuid.UUID, contactIDs []uuid.UUID, addedBy uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, segmentID, contactIDs, addedBy, orgID)
	return args.Error(0)
}

func (m *MockContactSegmentRepository) RemoveContactsFromSegment(ctx context.Context, segmentID uuid.UUID, contactIDs []uuid.UUID) error {
	args := m.Called(ctx, segmentID, contactIDs)
	return args.Error(0)
}

func (m *MockContactSegmentRepository) GetSegmentMembers(ctx context.Context, segmentID uuid.UUID, limit, offset int) ([]*types.Contact, error) {
	args := m.Called(ctx, segmentID, limit, offset)
	return args.Get(0).([]*types.Contact), args.Error(1)
}

func (m *MockContactSegmentRepository) GetContactSegments(ctx context.Context, contactID uuid.UUID) ([]*types.ContactSegment, error) {
	args := m.Called(ctx, contactID)
	return args.Get(0).([]*types.ContactSegment), args.Error(1)
}

func (m *MockContactSegmentRepository) IsContactInSegment(ctx context.Context, segmentID uuid.UUID, contactID uuid.UUID) (bool, error) {
	args := m.Called(ctx, segmentID, contactID)
	return args.Bool(0), args.Error(1)
}

func (m *MockContactSegmentRepository) ClearSegmentMembers(ctx context.Context, segmentID uuid.UUID) error {
	args := m.Called(ctx, segmentID)
	return args.Error(0)
}

func (m *MockContactSegmentRepository) UpdateSegmentMemberCount(ctx context.Context, segmentID uuid.UUID, count int) error {
	args := m.Called(ctx, segmentID, count)
	return args.Error(0)
}

// MockContactRepository for segment service tests
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) FindAll(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Contact), args.Error(1)
}

// MockAuthService for segment service tests
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) CheckOrganizationAccess(ctx context.Context, orgID uuid.UUID) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockAuthService) GetCurrentUser(ctx context.Context) (*types.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func TestContactSegmentService_CreateSegment(t *testing.T) {
	t.Run("success - static segment", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()
		userID := uuid.New()

		req := types.SegmentCreateRequest{
			Name:        "Test Segment",
			SegmentType: "static",
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		authService.On("GetCurrentUser", ctx).Return(&types.User{ID: userID}, nil)
		segmentRepo.On("CreateSegment", ctx, mock.MatchedBy(func(s *types.ContactSegment) bool {
			return s.Name == "Test Segment" && s.SegmentType == "static" && s.OrganizationID == orgID
		})).Return(nil)

		// Execute
		segment, err := service.CreateSegment(ctx, orgID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, segment)
		assert.Equal(t, "Test Segment", segment.Name)
		assert.Equal(t, "static", segment.SegmentType)
		authService.AssertExpectations(t)
		segmentRepo.AssertExpectations(t)
	})

	t.Run("success - dynamic segment with criteria", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()

		criteria := map[string]interface{}{
			"is_customer": true,
			"city":        "New York",
		}

		req := types.SegmentCreateRequest{
			Name:        "NYC Customers",
			SegmentType: "dynamic",
			Criteria:    criteria,
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		authService.On("GetCurrentUser", ctx).Return(nil, nil)
		segmentRepo.On("CreateSegment", ctx, mock.MatchedBy(func(s *types.ContactSegment) bool {
			return s.Name == "NYC Customers" && s.SegmentType == "dynamic"
		})).Return(nil)

		// Execute
		segment, err := service.CreateSegment(ctx, orgID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, segment)
		assert.Equal(t, "dynamic", segment.SegmentType)
		segmentRepo.AssertExpectations(t)
	})

	t.Run("error - invalid segment type", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()

		req := types.SegmentCreateRequest{
			Name:        "Test",
			SegmentType: "invalid",
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)

		// Execute
		segment, err := service.CreateSegment(ctx, orgID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, segment)
		assert.Contains(t, err.Error(), "invalid segment type")
	})
}

func TestContactSegmentService_RecalculateSegment(t *testing.T) {
	t.Run("success - recalculate dynamic segment", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()
		segmentID := uuid.New()

		criteria := map[string]interface{}{
			"is_customer": true,
			"city":        "New York",
		}

		segment := &types.ContactSegment{
			ID:             segmentID,
			OrganizationID: orgID,
			Name:           "NYC Customers",
			SegmentType:    "dynamic",
			Criteria:       criteria,
		}

		// Create test contacts
		city := "New York"
		contacts := []*types.Contact{
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Name:           "Contact 1",
				IsCustomer:     true,
				City:           &city,
			},
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Name:           "Contact 2",
				IsCustomer:     false, // Won't match
				City:           &city,
			},
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Name:           "Contact 3",
				IsCustomer:     true,
				City:           &city,
			},
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		segmentRepo.On("GetSegment", ctx, segmentID).Return(segment, nil)
		segmentRepo.On("ClearSegmentMembers", ctx, segmentID).Return(nil)
		contactRepo.On("FindAll", ctx, mock.MatchedBy(func(f types.ContactFilter) bool {
			return f.OrganizationID == orgID
		})).Return(contacts, nil)
		segmentRepo.On("AddContactsToSegment", ctx, segmentID, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 2 // Only 2 contacts match the criteria
		}), uuid.Nil, orgID).Return(nil)
		segmentRepo.On("UpdateSegmentMemberCount", ctx, segmentID, 2).Return(nil)

		// Execute
		count, err := service.RecalculateSegment(ctx, orgID, segmentID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		segmentRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("error - static segment cannot be recalculated", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()
		segmentID := uuid.New()

		segment := &types.ContactSegment{
			ID:             segmentID,
			OrganizationID: orgID,
			SegmentType:    "static",
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		segmentRepo.On("GetSegment", ctx, segmentID).Return(segment, nil)

		// Execute
		count, err := service.RecalculateSegment(ctx, orgID, segmentID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "only dynamic segments")
	})
}

func TestContactSegmentService_evaluateCriteria(t *testing.T) {
	service := &ContactSegmentService{}

	t.Run("is_customer criteria", func(t *testing.T) {
		criteria := map[string]interface{}{
			"is_customer": true,
		}

		contact := &types.Contact{
			IsCustomer: true,
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.True(t, result)

		contact.IsCustomer = false
		result = service.evaluateCriteria(contact, criteria)
		assert.False(t, result)
	})

	t.Run("city criteria", func(t *testing.T) {
		criteria := map[string]interface{}{
			"city": "New York",
		}

		city := "New York"
		contact := &types.Contact{
			City: &city,
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.True(t, result)

		otherCity := "Boston"
		contact.City = &otherCity
		result = service.evaluateCriteria(contact, criteria)
		assert.False(t, result)
	})

	t.Run("email_contains criteria", func(t *testing.T) {
		criteria := map[string]interface{}{
			"email_contains": "@example.com",
		}

		email := "test@example.com"
		contact := &types.Contact{
			Email: &email,
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.True(t, result)

		otherEmail := "test@other.com"
		contact.Email = &otherEmail
		result = service.evaluateCriteria(contact, criteria)
		assert.False(t, result)
	})

	t.Run("name_contains criteria", func(t *testing.T) {
		criteria := map[string]interface{}{
			"name_contains": "John",
		}

		contact := &types.Contact{
			Name: "John Doe",
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.True(t, result)

		contact.Name = "Jane Smith"
		result = service.evaluateCriteria(contact, criteria)
		assert.False(t, result)
	})

	t.Run("created_after criteria", func(t *testing.T) {
		criteria := map[string]interface{}{
			"created_after": "2024-01-01",
		}

		contact := &types.Contact{
			CreatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.True(t, result)

		contact.CreatedAt = time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
		result = service.evaluateCriteria(contact, criteria)
		assert.False(t, result)
	})

	t.Run("multiple criteria - all must match", func(t *testing.T) {
		criteria := map[string]interface{}{
			"is_customer": true,
			"city":        "New York",
		}

		city := "New York"
		contact := &types.Contact{
			IsCustomer: true,
			City:       &city,
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.True(t, result)

		// One criterion doesn't match
		contact.IsCustomer = false
		result = service.evaluateCriteria(contact, criteria)
		assert.False(t, result)
	})

	t.Run("empty criteria", func(t *testing.T) {
		criteria := map[string]interface{}{}

		contact := &types.Contact{
			Name: "Test",
		}

		result := service.evaluateCriteria(contact, criteria)
		assert.False(t, result) // Empty criteria should return false
	})
}

func TestContactSegmentService_AddContactsToSegment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()
		segmentID := uuid.New()
		userID := uuid.New()

		segment := &types.ContactSegment{
			ID:             segmentID,
			OrganizationID: orgID,
			SegmentType:    "static",
		}

		req := types.SegmentMemberRequest{
			ContactIDs: []uuid.UUID{uuid.New(), uuid.New()},
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		segmentRepo.On("GetSegment", ctx, segmentID).Return(segment, nil)
		authService.On("GetCurrentUser", ctx).Return(&types.User{ID: userID}, nil)
		segmentRepo.On("AddContactsToSegment", ctx, segmentID, req.ContactIDs, userID, orgID).Return(nil)

		// Execute
		err := service.AddContactsToSegment(ctx, orgID, segmentID, req)

		// Assert
		assert.NoError(t, err)
		segmentRepo.AssertExpectations(t)
	})

	t.Run("error - cannot add to dynamic segment", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()
		segmentID := uuid.New()

		segment := &types.ContactSegment{
			ID:             segmentID,
			OrganizationID: orgID,
			SegmentType:    "dynamic",
		}

		req := types.SegmentMemberRequest{
			ContactIDs: []uuid.UUID{uuid.New()},
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		segmentRepo.On("GetSegment", ctx, segmentID).Return(segment, nil)

		// Execute
		err := service.AddContactsToSegment(ctx, orgID, segmentID, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot manually add contacts to dynamic segments")
	})
}

func TestContactSegmentService_GetSegmentMembers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		segmentRepo := new(MockContactSegmentRepository)
		contactRepo := new(MockContactRepository)
		authService := new(MockAuthService)
		service := NewContactSegmentService(segmentRepo, contactRepo, authService, nil, nil)

		ctx := context.Background()
		orgID := uuid.New()
		segmentID := uuid.New()

		segment := &types.ContactSegment{
			ID:             segmentID,
			OrganizationID: orgID,
			MemberCount:    2,
		}

		contacts := []*types.Contact{
			{ID: uuid.New(), Name: "Contact 1"},
			{ID: uuid.New(), Name: "Contact 2"},
		}

		// Mock expectations
		authService.On("CheckOrganizationAccess", ctx, orgID).Return(nil)
		segmentRepo.On("GetSegment", ctx, segmentID).Return(segment, nil)
		segmentRepo.On("GetSegmentMembers", ctx, segmentID, 10, 0).Return(contacts, nil)

		// Execute
		result, count, err := service.GetSegmentMembers(ctx, orgID, segmentID, 10, 0)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, count)
		segmentRepo.AssertExpectations(t)
	})
}
