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

// MockContactRelationshipRepository is a mock implementation
type MockContactRelationshipRepository struct {
	mock.Mock
}

func (m *MockContactRelationshipRepository) CreateRelationshipType(ctx context.Context, relType *types.RelationshipType) error {
	args := m.Called(ctx, relType)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) GetRelationshipType(ctx context.Context, id uuid.UUID) (*types.RelationshipType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.RelationshipType), args.Error(1)
}

func (m *MockContactRelationshipRepository) UpdateRelationshipType(ctx context.Context, relType *types.RelationshipType) error {
	args := m.Called(ctx, relType)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) DeleteRelationshipType(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) ListRelationshipTypes(ctx context.Context, filter types.RelationshipTypeFilter) ([]*types.RelationshipType, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.RelationshipType), args.Error(1)
}

func (m *MockContactRelationshipRepository) CreateRelationship(ctx context.Context, rel *types.ContactRelationship) error {
	args := m.Called(ctx, rel)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) GetRelationship(ctx context.Context, id uuid.UUID) (*types.ContactRelationship, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactRelationship), args.Error(1)
}

func (m *MockContactRelationshipRepository) UpdateRelationship(ctx context.Context, rel *types.ContactRelationship) error {
	args := m.Called(ctx, rel)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) ListRelationships(ctx context.Context, contactID uuid.UUID) ([]*types.ContactRelationship, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactRelationship), args.Error(1)
}

func (m *MockContactRelationshipRepository) UpdateRelationshipStrength(ctx context.Context, id uuid.UUID, strength int) error {
	args := m.Called(ctx, id, strength)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) RecordInteraction(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactRelationshipRepository) GetRelationshipNetwork(ctx context.Context, contactID uuid.UUID, depth int) (*types.RelationshipNetwork, error) {
	args := m.Called(ctx, contactID, depth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.RelationshipNetwork), args.Error(1)
}

func TestContactRelationshipService_CreateRelationshipType(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("CreateRelationshipType", ctx, mock.MatchedBy(func(rt *types.RelationshipType) bool {
		return rt.Name == "Partner" && rt.OrganizationID == orgID && !rt.IsSystem
	})).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.CreateRelationshipTypeRequest{
		Name:            "Partner",
		Description:     strPtr("Business partner"),
		IsBidirectional: true,
	}

	result, err := svc.CreateRelationshipType(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Partner", result.Name)
	assert.False(t, result.IsSystem) // Custom types are never system

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func (m *MockContactRelationshipRepository) UpdateRelationshipType(ctx context.Context, relType *types.RelationshipType) error {
	args := m.Called(ctx, relType)
	return args.Error(0)
}

func TestContactRelationshipService_UpdateRelationshipType_CannotModifySystem(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	typeID := uuid.New()
	systemType := &types.RelationshipType{
		ID:             typeID,
		OrganizationID: orgID,
		Name:           "Colleague",
		IsSystem:       true, // System type
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetRelationshipType", ctx, typeID).Return(systemType, nil)

	// Test
	req := types.UpdateRelationshipTypeRequest{
		Name: strPtr("Updated Name"),
	}

	result, err := svc.UpdateRelationshipType(ctx, orgID, typeID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot modify system")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateRelationshipType")
}

func TestContactRelationshipService_CreateRelationship(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	fromContactID := uuid.New()
	toContactID := uuid.New()
	relTypeID := uuid.New()

	fromContact := &types.Contact{
		ID:             fromContactID,
		OrganizationID: orgID,
	}

	toContact := &types.Contact{
		ID:             toContactID,
		OrganizationID: orgID,
	}

	relType := &types.RelationshipType{
		ID:             relTypeID,
		OrganizationID: orgID,
		Name:           "Colleague",
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, fromContactID).Return(fromContact, nil)
	mockContactRepo.On("GetContact", ctx, toContactID).Return(toContact, nil)
	mockRepo.On("GetRelationshipType", ctx, relTypeID).Return(relType, nil)
	mockRepo.On("CreateRelationship", ctx, mock.MatchedBy(func(r *types.ContactRelationship) bool {
		return r.FromContactID == fromContactID && r.ToContactID == toContactID && r.StrengthScore == 50
	})).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.CreateRelationshipRequest{
		FromContactID:      fromContactID,
		ToContactID:        toContactID,
		RelationshipTypeID: relTypeID,
	}

	result, err := svc.CreateRelationship(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, fromContactID, result.FromContactID)
	assert.Equal(t, toContactID, result.ToContactID)
	assert.Equal(t, 50, result.StrengthScore) // Default strength

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactRelationshipService_UpdateRelationshipStrength(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	relationshipID := uuid.New()
	relationship := &types.ContactRelationship{
		ID:             relationshipID,
		OrganizationID: orgID,
		StrengthScore:  50,
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetRelationship", ctx, relationshipID).Return(relationship, nil)
	mockRepo.On("UpdateRelationshipStrength", ctx, relationshipID, 80).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.UpdateRelationshipStrengthRequest{
		StrengthScore: 80,
	}

	err := svc.UpdateRelationshipStrength(ctx, orgID, relationshipID, req)

	// Assertions
	require.NoError(t, err)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactRelationshipService_UpdateRelationshipStrength_InvalidScore(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	relationshipID := uuid.New()
	relationship := &types.ContactRelationship{
		ID:             relationshipID,
		OrganizationID: orgID,
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetRelationship", ctx, relationshipID).Return(relationship, nil)

	// Test with invalid score (>100)
	req := types.UpdateRelationshipStrengthRequest{
		StrengthScore: 150,
	}

	err := svc.UpdateRelationshipStrength(ctx, orgID, relationshipID, req)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 0 and 100")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateRelationshipStrength")
}

func TestContactRelationshipService_RecordInteraction(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	relationshipID := uuid.New()
	relationship := &types.ContactRelationship{
		ID:             relationshipID,
		OrganizationID: orgID,
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetRelationship", ctx, relationshipID).Return(relationship, nil)
	mockRepo.On("RecordInteraction", ctx, relationshipID).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	err := svc.RecordInteraction(ctx, orgID, relationshipID)

	// Assertions
	require.NoError(t, err)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactRelationshipService_GetRelationshipNetwork(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
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

	network := &types.RelationshipNetwork{
		CenterContactID: contactID,
		Depth:           2,
		TotalNodes:      5,
		TotalEdges:      4,
		Nodes: []types.NetworkNode{
			{ContactID: contactID, Depth: 0},
			{ContactID: uuid.New(), Depth: 1},
		},
		Edges: []types.NetworkEdge{
			{FromContactID: contactID, ToContactID: uuid.New(), StrengthScore: 80},
		},
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, contactID).Return(contact, nil)
	mockRepo.On("GetRelationshipNetwork", ctx, contactID, 2).Return(network, nil)

	// Test
	result, err := svc.GetRelationshipNetwork(ctx, orgID, contactID, 2)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, contactID, result.CenterContactID)
	assert.Equal(t, 2, result.Depth)
	assert.Equal(t, 5, result.TotalNodes)

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactRelationshipService_GetRelationshipNetwork_InvalidDepth(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
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

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, contactID).Return(contact, nil)

	// Test with invalid depth
	result, err := svc.GetRelationshipNetwork(ctx, orgID, contactID, 5)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "between 1 and 3")

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetRelationshipNetwork")
}

func TestContactRelationshipService_ListRelationships(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
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

	relationships := []*types.ContactRelationship{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			FromContactID:  contactID,
			ToContactID:    uuid.New(),
			StrengthScore:  90,
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			FromContactID:  uuid.New(),
			ToContactID:    contactID,
			StrengthScore:  75,
		},
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockContactRepo.On("GetContact", ctx, contactID).Return(contact, nil)
	mockRepo.On("ListRelationships", ctx, contactID).Return(relationships, nil)

	// Test
	result, err := svc.ListRelationships(ctx, orgID, contactID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 2)

	mockAuth.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactRelationshipService_DeleteRelationship(t *testing.T) {
	// Setup
	mockRepo := new(MockContactRelationshipRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)

	svc := service.NewContactRelationshipService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	relationshipID := uuid.New()
	relationship := &types.ContactRelationship{
		ID:             relationshipID,
		OrganizationID: orgID,
	}

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetRelationship", ctx, relationshipID).Return(relationship, nil)
	mockRepo.On("DeleteRelationship", ctx, relationshipID).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	err := svc.DeleteRelationship(ctx, orgID, relationshipID)

	// Assertions
	require.NoError(t, err)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}
