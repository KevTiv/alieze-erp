package testutils

import (
	"context"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// MockLeadRepository implements the repository.LeadRepository interface for testing
type MockLeadRepository struct {
	createFunc              func(ctx context.Context, lead types.Lead) (*types.Lead, error)
	findByIDFunc            func(ctx context.Context, id uuid.UUID) (*types.Lead, error)
	findAllFunc             func(ctx context.Context, filter types.LeadFilter) ([]*types.Lead, error)
	updateFunc              func(ctx context.Context, lead types.Lead) (*types.Lead, error)
	deleteFunc              func(ctx context.Context, id uuid.UUID) error
	countFunc               func(ctx context.Context, filter types.LeadFilter) (int, error)
	countByStageFunc        func(ctx context.Context) (map[uuid.UUID]int, error)
	findByDateRangeFunc     func(ctx context.Context, startDate, endDate time.Time) ([]types.Lead, error)
	findByDeadlineRangeFunc func(ctx context.Context, startDate, endDate time.Time) ([]types.Lead, error)
	findOverdueFunc         func(ctx context.Context) ([]types.Lead, error)
	findHighValueFunc       func(ctx context.Context, minValue float64) ([]types.Lead, error)
	findBySearchTermFunc    func(ctx context.Context, searchTerm string) ([]types.Lead, error)
}

// NewMockLeadRepository creates a new mock lead repository
func NewMockLeadRepository() *MockLeadRepository {
	return &MockLeadRepository{}
}

// Create implements the repository interface
func (m *MockLeadRepository) Create(ctx context.Context, lead types.Lead) (*types.Lead, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, lead)
	}
	return &lead, nil
}

// FindByID implements the repository interface
func (m *MockLeadRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	lead := &types.Lead{
		ID:             id,
		OrganizationID: uuid.Must(uuid.NewV7()),
		Name:           "Test Lead",
	}
	return lead, nil
}

// FindAll implements the repository interface
func (m *MockLeadRepository) FindAll(ctx context.Context, filter types.LeadFilter) ([]*types.Lead, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	lead1 := &types.Lead{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Lead 1"}
	lead2 := &types.Lead{ID: uuid.Must(uuid.NewV7()), OrganizationID: filter.OrganizationID, Name: "Lead 2"}
	return []*types.Lead{lead1, lead2}, nil
}

// Update implements the repository interface
func (m *MockLeadRepository) Update(ctx context.Context, lead types.Lead) (*types.Lead, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, lead)
	}
	return &lead, nil
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

// CountByStage implements the repository interface
func (m *MockLeadRepository) CountByStage(ctx context.Context) (map[uuid.UUID]int, error) {
	if m.countByStageFunc != nil {
		return m.countByStageFunc(ctx)
	}
	// Return default mock data
	stageID1 := uuid.Must(uuid.NewV7())
	stageID2 := uuid.Must(uuid.NewV7())
	return map[uuid.UUID]int{
		stageID1: 5,
		stageID2: 3,
	}, nil
}

// FindByDateRange implements the repository interface
func (m *MockLeadRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]types.Lead, error) {
	if m.findByDateRangeFunc != nil {
		return m.findByDateRangeFunc(ctx, startDate, endDate)
	}
	return []types.Lead{
		{ID: uuid.Must(uuid.NewV7()), Name: "Lead in date range"},
	}, nil
}

// FindByDeadlineRange implements the repository interface
func (m *MockLeadRepository) FindByDeadlineRange(ctx context.Context, startDate, endDate time.Time) ([]types.Lead, error) {
	if m.findByDeadlineRangeFunc != nil {
		return m.findByDeadlineRangeFunc(ctx, startDate, endDate)
	}
	return []types.Lead{
		{ID: uuid.Must(uuid.NewV7()), Name: "Lead with deadline in range"},
	}, nil
}

// FindOverdue implements the repository interface
func (m *MockLeadRepository) FindOverdue(ctx context.Context) ([]types.Lead, error) {
	if m.findOverdueFunc != nil {
		return m.findOverdueFunc(ctx)
	}
	return []types.Lead{
		{ID: uuid.Must(uuid.NewV7()), Name: "Overdue Lead"},
	}, nil
}

// FindHighValue implements the repository interface
func (m *MockLeadRepository) FindHighValue(ctx context.Context, minValue float64) ([]types.Lead, error) {
	if m.findHighValueFunc != nil {
		return m.findHighValueFunc(ctx, minValue)
	}
	expectedRevenue := minValue + 1000
	return []types.Lead{
		{ID: uuid.Must(uuid.NewV7()), Name: "High Value Lead", ExpectedRevenue: &expectedRevenue},
	}, nil
}

// FindBySearchTerm implements the repository interface
func (m *MockLeadRepository) FindBySearchTerm(ctx context.Context, searchTerm string) ([]types.Lead, error) {
	if m.findBySearchTermFunc != nil {
		return m.findBySearchTermFunc(ctx, searchTerm)
	}
	return []types.Lead{
		{ID: uuid.Must(uuid.NewV7()), Name: "Lead matching " + searchTerm},
	}, nil
}

// Helper methods to set mock behaviors
func (m *MockLeadRepository) WithCreateFunc(f func(ctx context.Context, lead types.Lead) (*types.Lead, error)) *MockLeadRepository {
	m.createFunc = f
	return m
}

func (m *MockLeadRepository) WithFindByIDFunc(f func(ctx context.Context, id uuid.UUID) (*types.Lead, error)) *MockLeadRepository {
	m.findByIDFunc = f
	return m
}

func (m *MockLeadRepository) WithFindAllFunc(f func(ctx context.Context, filter types.LeadFilter) ([]*types.Lead, error)) *MockLeadRepository {
	m.findAllFunc = f
	return m
}

func (m *MockLeadRepository) WithUpdateFunc(f func(ctx context.Context, lead types.Lead) (*types.Lead, error)) *MockLeadRepository {
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

func (m *MockLeadRepository) WithCountByStageFunc(f func(ctx context.Context) (map[uuid.UUID]int, error)) *MockLeadRepository {
	m.countByStageFunc = f
	return m
}

func (m *MockLeadRepository) WithFindByDateRangeFunc(f func(ctx context.Context, startDate, endDate time.Time) ([]types.Lead, error)) *MockLeadRepository {
	m.findByDateRangeFunc = f
	return m
}

func (m *MockLeadRepository) WithFindByDeadlineRangeFunc(f func(ctx context.Context, startDate, endDate time.Time) ([]types.Lead, error)) *MockLeadRepository {
	m.findByDeadlineRangeFunc = f
	return m
}

func (m *MockLeadRepository) WithFindOverdueFunc(f func(ctx context.Context) ([]types.Lead, error)) *MockLeadRepository {
	m.findOverdueFunc = f
	return m
}

func (m *MockLeadRepository) WithFindHighValueFunc(f func(ctx context.Context, minValue float64) ([]types.Lead, error)) *MockLeadRepository {
	m.findHighValueFunc = f
	return m
}

func (m *MockLeadRepository) WithFindBySearchTermFunc(f func(ctx context.Context, searchTerm string) ([]types.Lead, error)) *MockLeadRepository {
	m.findBySearchTermFunc = f
	return m
}
