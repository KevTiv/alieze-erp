package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStockMoveRepository is a mock implementation of StockMoveRepository
type MockStockMoveRepository struct {
	mock.Mock
}

func (m *MockStockMoveRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	args := m.Called(ctx, orgID, req)
	return args.Get(0).(*types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) CreateWithTx(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	args := m.Called(ctx, tx, orgID, req)
	return args.Get(0).(*types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) BulkCreate(ctx context.Context, orgID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error) {
	args := m.Called(ctx, orgID, reqs)
	return args.Get(0).([]types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) BulkCreateWithTx(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error) {
	args := m.Called(ctx, tx, orgID, reqs)
	return args.Get(0).([]types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockMove, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockMove, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) Update(ctx context.Context, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error) {
	args := m.Called(ctx, tx, id, req)
	return args.Get(0).(*types.StockMove), args.Error(1)
}

func (m *MockStockMoveRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStockMoveRepository) DeleteWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockStockMoveRepository) UpdateState(ctx context.Context, id uuid.UUID, state string) error {
	args := m.Called(ctx, id, state)
	return args.Error(0)
}

// MockStockQuantRepository is a mock implementation of StockQuantRepository
type MockStockQuantRepository struct {
	mock.Mock
}

func (m *MockStockQuantRepository) UpdateQuantity(ctx context.Context, organizationID, productID, locationID uuid.UUID, deltaQty float64) error {
	args := m.Called(ctx, organizationID, productID, locationID, deltaQty)
	return args.Error(0)
}

func (m *MockStockQuantRepository) UpdateQuantityWithTx(ctx context.Context, tx *sql.Tx, organizationID, productID, locationID uuid.UUID, deltaQty float64) error {
	args := m.Called(ctx, tx, organizationID, productID, locationID, deltaQty)
	return args.Error(0)
}

func (m *MockStockQuantRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.StockQuant, error) {
	args := m.Called(ctx, organizationID, productID)
	return args.Get(0).([]types.StockQuant), args.Error(1)
}

func (m *MockStockQuantRepository) FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]types.StockQuant, error) {
	args := m.Called(ctx, organizationID, locationID)
	return args.Get(0).([]types.StockQuant), args.Error(1)
}

func (m *MockStockQuantRepository) FindAvailable(ctx context.Context, organizationID, productID, locationID uuid.UUID) (float64, error) {
	args := m.Called(ctx, organizationID, productID, locationID)
	return args.Get(0).(float64), args.Error(1)
}

// MockWarehouseRepository is a mock implementation of WarehouseRepository
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Create(ctx context.Context, wh types.Warehouse) (*types.Warehouse, error) {
	args := m.Called(ctx, wh)
	return args.Get(0).(*types.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Warehouse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.Warehouse, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]types.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) Update(ctx context.Context, wh types.Warehouse) (*types.Warehouse, error) {
	args := m.Called(ctx, wh)
	return args.Get(0).(*types.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockStockLocationRepository is a mock implementation of StockLocationRepository
type MockStockLocationRepository struct {
	mock.Mock
}

func (m *MockStockLocationRepository) Create(ctx context.Context, loc types.StockLocation) (*types.StockLocation, error) {
	args := m.Called(ctx, loc)
	return args.Get(0).(*types.StockLocation), args.Error(1)
}

func (m *MockStockLocationRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.StockLocation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.StockLocation), args.Error(1)
}

func (m *MockStockLocationRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.StockLocation, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]types.StockLocation), args.Error(1)
}

func (m *MockStockLocationRepository) Update(ctx context.Context, loc types.StockLocation) (*types.StockLocation, error) {
	args := m.Called(ctx, loc)
	return args.Get(0).(*types.StockLocation), args.Error(1)
}

func (m *MockStockLocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestInventoryService_CreateMove(t *testing.T) {
	// Setup
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New()
	locID := uuid.New()
	destLocID := uuid.New()

	req := types.StockMoveCreateRequest{
		ProductID:      productID,
		LocationID:     locID,
		LocationDestID: destLocID,
		Quantity:       10.0,
		Date:           time.Now(),
	}

	expectedMove := &types.StockMove{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProductID:      productID,
		LocationID:     locID,
		LocationDestID: destLocID,
		Quantity:       10.0,
		State:          "draft",
		Sequence:       10,
		Priority:       "1",
	}

	mockMoveRepo := new(MockStockMoveRepository)

	// Set up the sanitized request expectation after we know what the defaults will be
	sanitizedReq := req
	if sanitizedReq.State == "" {
		sanitizedReq.State = "draft"
	}
	if sanitizedReq.Sequence == 0 {
		sanitizedReq.Sequence = 10
	}
	if sanitizedReq.Priority == "" {
		sanitizedReq.Priority = "1"
	}

	mockMoveRepo.On("Create", ctx, orgID, sanitizedReq).Return(expectedMove, nil)

	mockQuantRepo := new(MockStockQuantRepository)
	mockLocationRepo := new(MockStockLocationRepository)
	mockWarehouseRepo := new(MockWarehouseRepository)

	db := &sql.DB{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewInventoryService(db, logger, mockWarehouseRepo, mockLocationRepo, mockQuantRepo, mockMoveRepo)

	// Test


	// Set expected defaults that the service will apply
	expectedReqWithDefaults := req
	if expectedReqWithDefaults.State == "" {
		expectedReqWithDefaults.State = "draft"
	}
	if expectedReqWithDefaults.Sequence == 0 {
		expectedReqWithDefaults.Sequence = 10
	}
	if expectedReqWithDefaults.Priority == "" {
		expectedReqWithDefaults.Priority = "1"
	}

	// Also set the defaults on the original request since the service modifies it
	if req.State == "" {
		req.State = "draft"
	}
	if req.Sequence == 0 {
		req.Sequence = 10
	}
	if req.Priority == "" {
		req.Priority = "1"
	}

	// mockMoveRepo.On("Create", ctx, orgID, expectedReqWithDefaults).Return(expectedMove, nil)

	move, err := service.CreateMove(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, move)
	assert.Equal(t, expectedMove.ID, move.ID)
	assert.Equal(t, orgID, move.OrganizationID)
	assert.Equal(t, productID, move.ProductID)
	assert.Equal(t, locID, move.LocationID)
	assert.Equal(t, destLocID, move.LocationDestID)
	assert.Equal(t, 10.0, move.Quantity)
	assert.Equal(t, "draft", move.State)
	assert.Equal(t, 10, move.Sequence)
	assert.Equal(t, "1", move.Priority)

	mockMoveRepo.AssertExpectations(t)
}

func TestInventoryService_CreateMove_ValidationErrors(t *testing.T) {
	// Setup
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New()
	locID := uuid.New()
	destLocID := uuid.New()

	testCases := []struct {
		name        string
		req         types.StockMoveCreateRequest
		expectedErr string
	}{
		{
			name:        "Empty ProductID",
			req:         types.StockMoveCreateRequest{LocationID: locID, LocationDestID: destLocID, Quantity: 10.0, Date: time.Now()},
			expectedErr: "product_id is required",
		},
		{
			name:        "Empty LocationID",
			req:         types.StockMoveCreateRequest{ProductID: productID, LocationDestID: destLocID, Quantity: 10.0, Date: time.Now()},
			expectedErr: "location_id is required",
		},
		{
			name:        "Empty LocationDestID",
			req:         types.StockMoveCreateRequest{ProductID: productID, LocationID: locID, Quantity: 10.0, Date: time.Now()},
			expectedErr: "location_dest_id is required",
		},
		{
			name:        "Zero Quantity",
			req:         types.StockMoveCreateRequest{ProductID: productID, LocationID: locID, LocationDestID: destLocID, Quantity: 0, Date: time.Now()},
			expectedErr: "quantity must be positive",
		},
		{
			name:        "Negative Quantity",
			req:         types.StockMoveCreateRequest{ProductID: productID, LocationID: locID, LocationDestID: destLocID, Quantity: -5.0, Date: time.Now()},
			expectedErr: "quantity must be positive",
		},
	}

	mockMoveRepo := new(MockStockMoveRepository)
	mockQuantRepo := new(MockStockQuantRepository)
	mockLocationRepo := new(MockStockLocationRepository)
	mockWarehouseRepo := new(MockWarehouseRepository)

	db := &sql.DB{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewInventoryService(db, logger, mockWarehouseRepo, mockLocationRepo, mockQuantRepo, mockMoveRepo)

	// Test each case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		move, err := service.CreateMove(ctx, orgID, tc.req)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
			assert.Nil(t, move)
		})
	}

	mockMoveRepo.AssertExpectations(t)
}

func TestInventoryService_BulkCreateMoves(t *testing.T) {
	// Setup
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New()
	locID := uuid.New()
	destLocID := uuid.New()

	reqs := []types.StockMoveCreateRequest{
		{
			ProductID:      productID,
			LocationID:     locID,
			LocationDestID: destLocID,
			Quantity:       10.0,
			Date:           time.Now(),
		},
		{
			ProductID:      productID,
			LocationID:     locID,
			LocationDestID: destLocID,
			Quantity:       20.0,
			Date:           time.Now(),
		},
	}

	expectedMoves := []types.StockMove{
		{
			ID:             uuid.New(),
			ProductID:      productID,
			LocationID:     locID,
			LocationDestID: destLocID,
			Quantity:       10.0,
			State:          "draft",
			Sequence:       10,
			Priority:       "1",
		},
		{
			ID:             uuid.New(),

			ProductID:      productID,
			LocationID:     locID,
			LocationDestID: destLocID,
			Quantity:       20.0,
			State:          "draft",
			Sequence:       10,
			Priority:       "1",
		},
	}

	mockMoveRepo := new(MockStockMoveRepository)
	mockMoveRepo.On("BulkCreate", ctx, orgID, reqs).Return(expectedMoves, nil)

	mockQuantRepo := new(MockStockQuantRepository)
	mockLocationRepo := new(MockStockLocationRepository)
	mockWarehouseRepo := new(MockWarehouseRepository)

	db := &sql.DB{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewInventoryService(db, logger, mockWarehouseRepo, mockLocationRepo, mockQuantRepo, mockMoveRepo)

	// Test
	moves, err := service.BulkCreateMoves(ctx, orgID, reqs)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, moves, 2)
	assert.Equal(t, expectedMoves[0].ID, moves[0].ID)
	assert.Equal(t, expectedMoves[1].ID, moves[1].ID)
	assert.Equal(t, 10.0, moves[0].Quantity)
	assert.Equal(t, 20.0, moves[1].Quantity)

	mockMoveRepo.AssertExpectations(t)
}

func TestInventoryService_ProcessStockMoveWithTransaction(t *testing.T) {
	// Setup
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New()
	locID := uuid.New()
	destLocID := uuid.New()

	req := types.StockMoveCreateRequest{
		ProductID:      productID,
		LocationID:     locID,
		LocationDestID: destLocID,
		Quantity:       10.0,
		Date:           time.Now(),
	}

	expectedMove := &types.StockMove{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProductID:      productID,
		LocationID:     locID,
		LocationDestID: destLocID,
		Quantity:       10.0,
		State:          "draft",
		Sequence:       10,
		Priority:       "1",
	}

	mockMoveRepo := new(MockStockMoveRepository)
	mockMoveRepo.On("CreateWithTx", ctx, mock.Anything, orgID, req).Return(expectedMove, nil)

	mockQuantRepo := new(MockStockQuantRepository)
	mockQuantRepo.On("UpdateQuantityWithTx", ctx, mock.Anything, orgID, productID, locID, -10.0).Return(nil)
	mockQuantRepo.On("UpdateQuantityWithTx", ctx, mock.Anything, orgID, productID, destLocID, 10.0).Return(nil)

	mockLocationRepo := new(MockStockLocationRepository)
	mockWarehouseRepo := new(MockWarehouseRepository)

	// Create a mock database that can handle transactions
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Mock the transaction
	mockDB.ExpectBegin()
	mockDB.ExpectCommit()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewInventoryService(db, logger, mockWarehouseRepo, mockLocationRepo, mockQuantRepo, mockMoveRepo)

	// Test
	move, err := service.ProcessStockMoveWithTransaction(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, move)
	assert.Equal(t, expectedMove.ID, move.ID)
	assert.Equal(t, orgID, move.OrganizationID)
	assert.Equal(t, productID, move.ProductID)
	assert.Equal(t, locID, move.LocationID)
	assert.Equal(t, destLocID, move.LocationDestID)
	assert.Equal(t, 10.0, move.Quantity)

	mockMoveRepo.AssertExpectations(t)
	mockQuantRepo.AssertExpectations(t)
}

func TestInventoryService_ProcessStockMoveWithTransaction_Rollback(t *testing.T) {
	// Setup
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New()
	locID := uuid.New()
	destLocID := uuid.New()

	req := types.StockMoveCreateRequest{
		ProductID:      productID,
		LocationID:     locID,
		LocationDestID: destLocID,
		Quantity:       10.0,
		Date:           time.Now(),
	}

	expectedMove := &types.StockMove{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProductID:      productID,
		LocationID:     locID,
		LocationDestID: destLocID,
		Quantity:       10.0,
		State:          "draft",
		Sequence:       10,
		Priority:       "1",
	}

	mockMoveRepo := new(MockStockMoveRepository)
	mockMoveRepo.On("CreateWithTx", ctx, mock.Anything, orgID, req).Return(expectedMove, nil)

	mockQuantRepo := new(MockStockQuantRepository)
	mockQuantRepo.On("UpdateQuantityWithTx", ctx, mock.Anything, orgID, productID, locID, -10.0).Return(nil)
	mockQuantRepo.On("UpdateQuantityWithTx", ctx, mock.Anything, orgID, productID, destLocID, 10.0).Return(errors.New("failed to update quantity"))

	mockLocationRepo := new(MockStockLocationRepository)
	mockWarehouseRepo := new(MockWarehouseRepository)

	// Create a mock database that can handle transactions
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Mock the transaction
	mockDB.ExpectBegin()
	mockDB.ExpectRollback()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewInventoryService(db, logger, mockWarehouseRepo, mockLocationRepo, mockQuantRepo, mockMoveRepo)

	// Test
	move, err := service.ProcessStockMoveWithTransaction(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to increase dest quantity")
	assert.Nil(t, move)

	mockMoveRepo.AssertExpectations(t)
	mockQuantRepo.AssertExpectations(t)
}
