package testutils

import (
	"context"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// MockSalesTeamRepository implements the repository.SalesTeamRepository interface for testing
type MockSalesTeamRepository struct {
	createFunc       func(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error)
	findByIDFunc     func(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error)
	findAllFunc      func(ctx context.Context, filter types.SalesTeamFilter) ([]types.SalesTeam, error)
	updateFunc       func(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error)
	deleteFunc       func(ctx context.Context, id uuid.UUID) error
	countFunc        func(ctx context.Context, filter types.SalesTeamFilter) (int, error)
	findByMemberFunc func(ctx context.Context, memberID uuid.UUID) ([]types.SalesTeam, error)
}

// NewMockSalesTeamRepository creates a new mock sales team repository
func NewMockSalesTeamRepository() *MockSalesTeamRepository {
	return &MockSalesTeamRepository{}
}

// Create implements the repository interface
func (m *MockSalesTeamRepository) Create(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, team)
	}
	return &team, nil
}

// FindByID implements the repository interface
func (m *MockSalesTeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return &types.SalesTeam{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		Name:           "Test Sales Team",
		IsActive:       true,
	}, nil
}

// FindAll implements the repository interface
func (m *MockSalesTeamRepository) FindAll(ctx context.Context, filter types.SalesTeamFilter) ([]types.SalesTeam, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return []types.SalesTeam{
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: filter.OrganizationID,
			Name:           "Sales Team 1",
			IsActive:       true,
		},
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: filter.OrganizationID,
			Name:           "Sales Team 2",
			IsActive:       true,
		},
	}, nil
}

// Update implements the repository interface
func (m *MockSalesTeamRepository) Update(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, team)
	}
	return &team, nil
}

// Delete implements the repository interface
func (m *MockSalesTeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// Count implements the repository interface
func (m *MockSalesTeamRepository) Count(ctx context.Context, filter types.SalesTeamFilter) (int, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, filter)
	}
	return 2, nil
}

// FindByMember implements the repository interface
func (m *MockSalesTeamRepository) FindByMember(ctx context.Context, memberID uuid.UUID) ([]types.SalesTeam, error) {
	if m.findByMemberFunc != nil {
		return m.findByMemberFunc(ctx, memberID)
	}
	return []types.SalesTeam{
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: uuid.Must(uuid.NewV7()),
			Name:           "Team with Member",
			MemberIDs:      []uuid.UUID{memberID},
			IsActive:       true,
		},
	}, nil
}

// Helper methods to set mock behaviors
func (m *MockSalesTeamRepository) WithCreateFunc(f func(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error)) *MockSalesTeamRepository {
	m.createFunc = f
	return m
}

func (m *MockSalesTeamRepository) WithFindByIDFunc(f func(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error)) *MockSalesTeamRepository {
	m.findByIDFunc = f
	return m
}

func (m *MockSalesTeamRepository) WithFindAllFunc(f func(ctx context.Context, filter types.SalesTeamFilter) ([]types.SalesTeam, error)) *MockSalesTeamRepository {
	m.findAllFunc = f
	return m
}

func (m *MockSalesTeamRepository) WithUpdateFunc(f func(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error)) *MockSalesTeamRepository {
	m.updateFunc = f
	return m
}

func (m *MockSalesTeamRepository) WithDeleteFunc(f func(ctx context.Context, id uuid.UUID) error) *MockSalesTeamRepository {
	m.deleteFunc = f
	return m
}

func (m *MockSalesTeamRepository) WithCountFunc(f func(ctx context.Context, filter types.SalesTeamFilter) (int, error)) *MockSalesTeamRepository {
	m.countFunc = f
	return m
}

func (m *MockSalesTeamRepository) WithFindByMemberFunc(f func(ctx context.Context, memberID uuid.UUID) ([]types.SalesTeam, error)) *MockSalesTeamRepository {
	m.findByMemberFunc = f
	return m
}
