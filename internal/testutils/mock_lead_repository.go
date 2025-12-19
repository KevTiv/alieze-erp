package testutils

import (
	"context"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// MockLeadRepository implements the repository.LeadRepository interface for testing
type MockLeadRepository struct {
	createFunc   func(ctx context.Context, lead types.Lead) error
	findByIDFunc func(ctx context.Context, id uuid.UUID) (types.Lead, error)
	findAllFunc  func(ctx context.Context, filter types.LeadFilter) ([]types.Lead, error)
	updateFunc   func(ctx context.Context, lead types.Lead) error
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
	countFunc    func(ctx context.Context, filter types.LeadFilter) (int, error)
}

// NewMockLeadRepository creates a new mock lead repository
func NewMockLeadRepository() *MockLeadRepository {
	return &MockLeadRepository{}
}

// Create implements the repository interface
func (m *MockLeadRepository) Create(ctx context.Context, lead types.Lead) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, lead)
	}
	return nil
}

// FindByID implements the repository interface
func (m *MockLeadRepository) FindByID(ctx context.Context, id uuid.UUID) (types.Lead, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return types.Lead{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		Name:           "Test Lead",
	}, nil
}

// FindAll implements the repository interface
func (m *MockLeadRepository) FindAll(ctx context.Context, filter types.LeadFilter) ([]types.Lead, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return []types.Lead{
		{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Lead 1"},
		{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Lead 2"},
	}, nil
}

// Update implements the repository interface
func (m *MockLeadRepository) Update(ctx context.Context, lead types.Lead) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, lead)
	}
	return nil
}

// Delete implements the repository interface
func (m *MockLeadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// Count implements the repository interface
func (m *MockLeadRepository) Count(ctx context.Context, filter types.LeadFilter) (int, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, filter)
	}
	return 2, nil
}

// Helper methods to set mock behaviors
func (m *MockLeadRepository) WithCreateFunc(f func(ctx context.Context, lead types.Lead) error) *MockLeadRepository {
	m.createFunc = f
	return m
}

func (m *MockLeadRepository) WithFindByIDFunc(f func(ctx context.Context, id uuid.UUID) (types.Lead, error)) *MockLeadRepository {
	m.findByIDFunc = f
	return m
}

func (m *MockLeadRepository) WithFindAllFunc(f func(ctx context.Context, filter types.LeadFilter) ([]types.Lead, error)) *MockLeadRepository {
	m.findAllFunc = f
	return m
}

func (m *MockLeadRepository) WithUpdateFunc(f func(ctx context.Context, lead types.Lead) error) *MockLeadRepository {
	m.updateFunc = f
	return m
}

func (m *MockLeadRepository) WithDeleteFunc(f func(ctx context.Context, id uuid.UUID) error) *MockLeadRepository {
	m.deleteFunc = f
	return m
}

func (m *MockLeadRepository) WithCountFunc(f func(ctx context.Context, filter types.LeadFilter) (int, error)) *MockLeadRepository {
	m.countFunc = f
	return m
}
