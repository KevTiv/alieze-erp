package testutils

import (
	"context"

	"alieze-erp/internal/modules/products/types"
	"alieze-erp/internal/modules/products/repository"

	"github.com/google/uuid"
)

// MockProductRepository implements the repository.ProductRepo interface for testing
type MockProductRepository struct {
	createFunc   func(ctx context.Context, product types.Product) (*types.Product, error)
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*types.Product, error)
	findAllFunc  func(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	updateFunc   func(ctx context.Context, product types.Product) (*types.Product, error)
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
	countFunc    func(ctx context.Context, filter types.ProductFilter) (int, error)
}

// Ensure MockProductRepository implements ProductRepo interface
var _ repository.ProductRepo = &MockProductRepository{}

// NewMockProductRepository creates a new mock product repository
func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{}
}

// Create implements the repository interface
func (m *MockProductRepository) Create(ctx context.Context, product types.Product) (*types.Product, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, product)
	}
	return &product, nil
}

// FindByID implements the repository interface
func (m *MockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Product, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return &types.Product{ID: id, OrganizationID: uuid.Must(uuid.NewV7()), Name: "Test Product"}, nil
}

// FindAll implements the repository interface
func (m *MockProductRepository) FindAll(ctx context.Context, filter types.ProductFilter) ([]types.Product, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return []types.Product{
		{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Product 1"},
		{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Product 2"},
	}, nil
}

// Update implements the repository interface
func (m *MockProductRepository) Update(ctx context.Context, product types.Product) (*types.Product, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, product)
	}
	return &product, nil
}

// Delete implements the repository interface
func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// Count implements the repository interface
func (m *MockProductRepository) Count(ctx context.Context, filter types.ProductFilter) (int, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, filter)
	}
	return 2, nil
}

// Helper methods to set mock behaviors
func (m *MockProductRepository) WithCreateFunc(f func(ctx context.Context, product types.Product) (*types.Product, error)) *MockProductRepository {
	m.createFunc = f
	return m
}

func (m *MockProductRepository) WithFindByIDFunc(f func(ctx context.Context, id uuid.UUID) (*types.Product, error)) *MockProductRepository {
	m.findByIDFunc = f
	return m
}

func (m *MockProductRepository) WithFindAllFunc(f func(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)) *MockProductRepository {
	m.findAllFunc = f
	return m
}

func (m *MockProductRepository) WithUpdateFunc(f func(ctx context.Context, product types.Product) (*types.Product, error)) *MockProductRepository {
	m.updateFunc = f
	return m
}

func (m *MockProductRepository) WithDeleteFunc(f func(ctx context.Context, id uuid.UUID) error) *MockProductRepository {
	m.deleteFunc = f
	return m
}

func (m *MockProductRepository) WithCountFunc(f func(ctx context.Context, filter types.ProductFilter) (int, error)) *MockProductRepository {
	m.countFunc = f
	return m
}
