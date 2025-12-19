package testutils

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCRUDRepository provides a generic mock implementation of the base.Repository interface
type MockCRUDRepository[Entity any, Filter any] struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockCRUDRepository[Entity, Filter]) Create(ctx context.Context, entity Entity) (*Entity, error) {
	args := m.Called(ctx, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Entity), args.Error(1)
}

// FindByID mocks the FindByID method
func (m *MockCRUDRepository[Entity, Filter]) FindByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Entity), args.Error(1)
}

// FindAll mocks the FindAll method
func (m *MockCRUDRepository[Entity, Filter]) FindAll(ctx context.Context, filter Filter) ([]*Entity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Entity), args.Error(1)
}

// Update mocks the Update method
func (m *MockCRUDRepository[Entity, Filter]) Update(ctx context.Context, entity Entity) (*Entity, error) {
	args := m.Called(ctx, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Entity), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockCRUDRepository[Entity, Filter]) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Count mocks the Count method
func (m *MockCRUDRepository[Entity, Filter]) Count(ctx context.Context, filter Filter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

// WithCreateFunc provides a fluent interface for setting up Create method
func (m *MockCRUDRepository[Entity, Filter]) WithCreateFunc(fn func(context.Context, Entity) (*Entity, error)) *MockCRUDRepository[Entity, Filter] {
	m.On("Create", mock.Anything, mock.Anything).Return(fn)
	return m
}

// WithFindByIDFunc provides a fluent interface for setting up FindByID method
func (m *MockCRUDRepository[Entity, Filter]) WithFindByIDFunc(fn func(context.Context, uuid.UUID) (*Entity, error)) *MockCRUDRepository[Entity, Filter] {
	m.On("FindByID", mock.Anything, mock.Anything).Return(fn)
	return m
}

// WithFindAllFunc provides a fluent interface for setting up FindAll method
func (m *MockCRUDRepository[Entity, Filter]) WithFindAllFunc(fn func(context.Context, Filter) ([]*Entity, error)) *MockCRUDRepository[Entity, Filter] {
	m.On("FindAll", mock.Anything, mock.Anything).Return(fn)
	return m
}

// WithUpdateFunc provides a fluent interface for setting up Update method
func (m *MockCRUDRepository[Entity, Filter]) WithUpdateFunc(fn func(context.Context, Entity) (*Entity, error)) *MockCRUDRepository[Entity, Filter] {
	m.On("Update", mock.Anything, mock.Anything).Return(fn)
	return m
}

// WithDeleteFunc provides a fluent interface for setting up Delete method
func (m *MockCRUDRepository[Entity, Filter]) WithDeleteFunc(fn func(context.Context, uuid.UUID) error) *MockCRUDRepository[Entity, Filter] {
	m.On("Delete", mock.Anything, mock.Anything).Return(fn)
	return m
}

// WithCountFunc provides a fluent interface for setting up Count method
func (m *MockCRUDRepository[Entity, Filter]) WithCountFunc(fn func(context.Context, Filter) (int, error)) *MockCRUDRepository[Entity, Filter] {
	m.On("Count", mock.Anything, mock.Anything).Return(fn)
	return m
}

// NewMockCRUDRepository creates a new mock repository with default no-op implementations
func NewMockCRUDRepository[Entity any, Filter any]() *MockCRUDRepository[Entity, Filter] {
	m := &MockCRUDRepository[Entity, Filter]{}

	// Setup default no-op implementations
	m.WithCreateFunc(func(ctx context.Context, entity Entity) (*Entity, error) {
		return nil, nil
	})
	m.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*Entity, error) {
		return nil, nil
	})
	m.WithFindAllFunc(func(ctx context.Context, filter Filter) ([]*Entity, error) {
		return []*Entity{}, nil
	})
	m.WithUpdateFunc(func(ctx context.Context, entity Entity) (*Entity, error) {
		return nil, nil
	})
	m.WithDeleteFunc(func(ctx context.Context, id uuid.UUID) error {
		return nil
	})
	m.WithCountFunc(func(ctx context.Context, filter Filter) (int, error) {
		return 0, nil
	})

	return m
}
