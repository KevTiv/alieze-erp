package testutils

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// MockActivityRepository implements the repository.ActivityRepository interface for testing
type MockActivityRepository struct {
	createFunc   func(ctx context.Context, activity types.Activity) (*types.Activity, error)
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*types.Activity, error)
	findAllFunc  func(ctx context.Context, filter types.ActivityFilter) ([]*types.Activity, error)
	updateFunc   func(ctx context.Context, activity types.Activity) (*types.Activity, error)
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
	countFunc    func(ctx context.Context, filter types.ActivityFilter) (int, error)
}

// NewMockActivityRepository creates a new mock activity repository
func NewMockActivityRepository() *MockActivityRepository {
	return &MockActivityRepository{}
}

// Create implements the repository interface
func (m *MockActivityRepository) Create(ctx context.Context, activity types.Activity) (*types.Activity, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, activity)
	}
	return &activity, nil
}

// FindByID implements the repository interface
func (m *MockActivityRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return &types.Activity{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		ActivityType:   types.ActivityTypeMeeting,
		Summary:        "Test Activity",
		State:          types.ActivityStatePlanned,
	}, nil
}

// FindAll implements the repository interface
func (m *MockActivityRepository) FindAll(ctx context.Context, filter types.ActivityFilter) ([]*types.Activity, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return []*types.Activity{
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: filter.OrganizationID,
			ActivityType:   types.ActivityTypeCall,
			Summary:        "Activity 1",
			State:          types.ActivityStatePlanned,
		},
		{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: filter.OrganizationID,
			ActivityType:   types.ActivityTypeEmail,
			Summary:        "Activity 2",
			State:          types.ActivityStateDone,
		},
	}, nil
}

// Update implements the repository interface
func (m *MockActivityRepository) Update(ctx context.Context, activity types.Activity) (*types.Activity, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, activity)
	}
	return &activity, nil
}

// Delete implements the repository interface
func (m *MockActivityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// Count implements the repository interface
func (m *MockActivityRepository) Count(ctx context.Context, filter types.ActivityFilter) (int, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, filter)
	}
	return 2, nil
}

// FindByContact implements the repository interface
func (m *MockActivityRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]*types.Activity, error) {
	// Default implementation returns empty slice
	return []*types.Activity{}, nil
}

// FindByLead implements the repository interface
func (m *MockActivityRepository) FindByLead(ctx context.Context, leadID uuid.UUID) ([]*types.Activity, error) {
	// Default implementation returns empty slice
	return []*types.Activity{}, nil
}

// Helper methods to set mock behaviors
func (m *MockActivityRepository) WithCreateFunc(f func(ctx context.Context, activity types.Activity) (*types.Activity, error)) *MockActivityRepository {
	m.createFunc = f
	return m
}

func (m *MockActivityRepository) WithFindByIDFunc(f func(ctx context.Context, id uuid.UUID) (*types.Activity, error)) *MockActivityRepository {
	m.findByIDFunc = f
	return m
}

func (m *MockActivityRepository) WithFindAllFunc(f func(ctx context.Context, filter types.ActivityFilter) ([]*types.Activity, error)) *MockActivityRepository {
	m.findAllFunc = f
	return m
}

func (m *MockActivityRepository) WithUpdateFunc(f func(ctx context.Context, activity types.Activity) (*types.Activity, error)) *MockActivityRepository {
	m.updateFunc = f
	return m
}

func (m *MockActivityRepository) WithDeleteFunc(f func(ctx context.Context, id uuid.UUID) error) *MockActivityRepository {
	m.deleteFunc = f
	return m
}

func (m *MockActivityRepository) WithCountFunc(f func(ctx context.Context, filter types.ActivityFilter) (int, error)) *MockActivityRepository {
	m.countFunc = f
	return m
}
