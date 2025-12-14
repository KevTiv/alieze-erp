package testutils

import (
	"context"

	"alieze-erp/internal/modules/crm/domain"

	"github.com/google/uuid"
)

// MockContactRepository implements the repository.ContactRepository interface for testing
type MockContactRepository struct {
	createFunc   func(ctx context.Context, contact domain.Contact) (*domain.Contact, error)
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Contact, error)
	findAllFunc  func(ctx context.Context, filter domain.ContactFilter) ([]domain.Contact, error)
	updateFunc   func(ctx context.Context, contact domain.Contact) (*domain.Contact, error)
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
	countFunc    func(ctx context.Context, filter domain.ContactFilter) (int, error)
}

// NewMockContactRepository creates a new mock contact repository
func NewMockContactRepository() *MockContactRepository {
	return &MockContactRepository{}
}

// Create implements the repository interface
func (m *MockContactRepository) Create(ctx context.Context, contact domain.Contact) (*domain.Contact, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, contact)
	}
	return &contact, nil
}

// FindByID implements the repository interface
func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return &domain.Contact{ID: id, OrganizationID: uuid.Must(uuid.NewV7()), Name: "Test Contact"}, nil
}

// FindAll implements the repository interface
func (m *MockContactRepository) FindAll(ctx context.Context, filter domain.ContactFilter) ([]domain.Contact, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return []domain.Contact{
		{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Contact 1"},
		{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Contact 2"},
	}, nil
}

// Update implements the repository interface
func (m *MockContactRepository) Update(ctx context.Context, contact domain.Contact) (*domain.Contact, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, contact)
	}
	return &contact, nil
}

// Delete implements the repository interface
func (m *MockContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// Count implements the repository interface
func (m *MockContactRepository) Count(ctx context.Context, filter domain.ContactFilter) (int, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, filter)
	}
	return 2, nil
}

// Helper methods to set mock behaviors
func (m *MockContactRepository) WithCreateFunc(f func(ctx context.Context, contact domain.Contact) (*domain.Contact, error)) *MockContactRepository {
	m.createFunc = f
	return m
}

func (m *MockContactRepository) WithFindByIDFunc(f func(ctx context.Context, id uuid.UUID) (*domain.Contact, error)) *MockContactRepository {
	m.findByIDFunc = f
	return m
}

func (m *MockContactRepository) WithFindAllFunc(f func(ctx context.Context, filter domain.ContactFilter) ([]domain.Contact, error)) *MockContactRepository {
	m.findAllFunc = f
	return m
}

func (m *MockContactRepository) WithUpdateFunc(f func(ctx context.Context, contact domain.Contact) (*domain.Contact, error)) *MockContactRepository {
	m.updateFunc = f
	return m
}

func (m *MockContactRepository) WithDeleteFunc(f func(ctx context.Context, id uuid.UUID) error) *MockContactRepository {
	m.deleteFunc = f
	return m
}

func (m *MockContactRepository) WithCountFunc(f func(ctx context.Context, filter domain.ContactFilter) (int, error)) *MockContactRepository {
	m.countFunc = f
	return m
}
