package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/sales/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/sales/service"
	"github.com/KevTiv/alieze-erp/internal/modules/sales/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSalesOrderRepository is a mock implementation for testing
type MockSalesOrderRepository struct {
	mock.Mock
}

func (m *MockSalesOrderRepository) Create(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(*types.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.SalesOrder, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) FindAll(ctx context.Context, filters repository.SalesOrderFilter) ([]types.SalesOrder, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]types.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) Update(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(*types.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSalesOrderRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]types.SalesOrder, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]types.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) FindByStatus(ctx context.Context, status types.SalesOrderStatus) ([]types.SalesOrder, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]types.SalesOrder), args.Error(1)
}

// MockPricelistRepository is a mock implementation for testing
type MockPricelistRepository struct {
	mock.Mock
}

func (m *MockPricelistRepository) Create(ctx context.Context, pricelist types.Pricelist) (*types.Pricelist, error) {
	args := m.Called(ctx, pricelist)
	return args.Get(0).(*types.Pricelist), args.Error(1)
}

func (m *MockPricelistRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Pricelist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Pricelist), args.Error(1)
}

func (m *MockPricelistRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.Pricelist, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]types.Pricelist), args.Error(1)
}

func (m *MockPricelistRepository) Update(ctx context.Context, pricelist types.Pricelist) (*types.Pricelist, error) {
	args := m.Called(ctx, pricelist)
	return args.Get(0).(*types.Pricelist), args.Error(1)
}

func (m *MockPricelistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPricelistRepository) FindByCompanyID(ctx context.Context, companyID uuid.UUID) ([]types.Pricelist, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]types.Pricelist), args.Error(1)
}

func (m *MockPricelistRepository) FindActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]types.Pricelist, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]types.Pricelist), args.Error(1)
}

func TestSalesOrderService_CreateSalesOrder_Success(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	order := types.SalesOrder{
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-001",
		PricelistID:    pricelistID,
		CurrencyID:     currencyID,
		Lines: []types.SalesOrderLine{
			{
				ProductID:   productID,
				ProductName: "Test Product",
				Quantity:    2.0,
				UomID:       uomID,
				UnitPrice:   50.0,
			},
		},
	}

	expectedOrder := order
	expectedOrder.ID = uuid.New()
	expectedOrder.Status = types.SalesOrderStatusDraft
	expectedOrder.OrderDate = time.Now()
	expectedOrder.AmountUntaxed = 100.0
	expectedOrder.AmountTax = 0.0
	expectedOrder.AmountTotal = 100.0
	expectedOrder.CreatedAt = time.Now()
	expectedOrder.UpdatedAt = time.Now()
	expectedOrder.Lines[0].ID = uuid.New()
	expectedOrder.Lines[0].PriceSubtotal = 100.0
	expectedOrder.Lines[0].PriceTax = 0.0
	expectedOrder.Lines[0].PriceTotal = 100.0
	expectedOrder.Lines[0].Sequence = 1
	expectedOrder.Lines[0].CreatedAt = time.Now()
	expectedOrder.Lines[0].UpdatedAt = time.Now()

	// Mock expectations
	mockOrderRepo.On("Create", context.Background(), mock.AnythingOfType("types.SalesOrder")).
		Return(&expectedOrder, nil)

	// Execute
	createdOrder, err := service.CreateSalesOrder(context.Background(), order)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, expectedOrder.ID, createdOrder.ID)
	assert.Equal(t, types.SalesOrderStatusDraft, createdOrder.Status)
	assert.Equal(t, 100.0, createdOrder.AmountUntaxed)
	assert.Equal(t, 100.0, createdOrder.AmountTotal)
	assert.Len(t, createdOrder.Lines, 1)
	assert.Equal(t, 100.0, createdOrder.Lines[0].PriceSubtotal)

	mockOrderRepo.AssertExpectations(t)
}

func TestSalesOrderService_CreateSalesOrder_ValidationError(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data - missing required fields
	order := types.SalesOrder{
		// Missing OrganizationID, CompanyID, CustomerID, PricelistID, CurrencyID
		Lines: []types.SalesOrderLine{
			{
				ProductID: uuid.New(),
				Quantity:  2.0,
				UomID:     uuid.New(),
				UnitPrice: 50.0,
			},
		},
	}

	// Execute
	createdOrder, err := service.CreateSalesOrder(context.Background(), order)

	// Assert
	require.Error(t, err)
	assert.Nil(t, createdOrder)
	assert.Contains(t, err.Error(), "organization ID is required")

	mockOrderRepo.AssertNotCalled(t, "Create")
}

func TestSalesOrderService_ConfirmSalesOrder_Success(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data
	orderID := uuid.New()
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	existingOrder := types.SalesOrder{
		ID:             orderID,
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-002",
		Status:         types.SalesOrderStatusDraft,
		OrderDate:      time.Now(),
		PricelistID:    pricelistID,
		CurrencyID:     currencyID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		Note:           "Test order",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.SalesOrderLine{
			{
				ID:            uuid.New(),
				ProductID:     productID,
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UomID:         uomID,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	expectedUpdatedOrder := existingOrder
	expectedUpdatedOrder.Status = types.SalesOrderStatusConfirmed
	now := time.Now()
	expectedUpdatedOrder.ConfirmationDate = &now
	expectedUpdatedOrder.UpdatedAt = now

	// Mock expectations
	mockOrderRepo.On("FindByID", context.Background(), orderID).Return(&existingOrder, nil)
	mockOrderRepo.On("Update", context.Background(), mock.AnythingOfType("types.SalesOrder")).
		Return(&expectedUpdatedOrder, nil)

	// Execute
	confirmedOrder, err := service.ConfirmSalesOrder(context.Background(), orderID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, confirmedOrder)
	assert.Equal(t, types.SalesOrderStatusConfirmed, confirmedOrder.Status)
	require.NotNil(t, confirmedOrder.ConfirmationDate)
	assert.WithinDuration(t, time.Now(), *confirmedOrder.ConfirmationDate, time.Second)

	mockOrderRepo.AssertExpectations(t)
}

func TestSalesOrderService_ConfirmSalesOrder_InvalidStatus(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data - order already confirmed
	orderID := uuid.New()
	existingOrder := types.SalesOrder{
		ID:     orderID,
		Status: types.SalesOrderStatusConfirmed,
	}

	// Mock expectations
	mockOrderRepo.On("FindByID", context.Background(), orderID).Return(&existingOrder, nil)

	// Execute
	confirmedOrder, err := service.ConfirmSalesOrder(context.Background(), orderID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, confirmedOrder)
	assert.Contains(t, err.Error(), "only draft or quotation orders can be confirmed")

	mockOrderRepo.AssertExpectations(t)
}

func TestSalesOrderService_CancelSalesOrder_Success(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data
	orderID := uuid.New()
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()

	existingOrder := types.SalesOrder{
		ID:             orderID,
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-003",
		Status:         types.SalesOrderStatusDraft,
		OrderDate:      time.Now(),
		PricelistID:    pricelistID,
		CurrencyID:     currencyID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		Note:           "Test order",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
	}

	expectedUpdatedOrder := existingOrder
	expectedUpdatedOrder.Status = types.SalesOrderStatusCancelled
	expectedUpdatedOrder.UpdatedAt = time.Now()

	// Mock expectations
	mockOrderRepo.On("FindByID", context.Background(), orderID).Return(&existingOrder, nil)
	mockOrderRepo.On("Update", context.Background(), mock.AnythingOfType("types.SalesOrder")).
		Return(&expectedUpdatedOrder, nil)

	// Execute
	cancelledOrder, err := service.CancelSalesOrder(context.Background(), orderID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cancelledOrder)
	assert.Equal(t, types.SalesOrderStatusCancelled, cancelledOrder.Status)

	mockOrderRepo.AssertExpectations(t)
}

func TestSalesOrderService_CalculateOrderAmounts(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data
	productID := uuid.New()
	uomID := uuid.New()

	order := types.SalesOrder{
		Lines: []types.SalesOrderLine{
			{
				ProductID:   productID,
				ProductName: "Product 1",
				Quantity:    2.0,
				UomID:       uomID,
				UnitPrice:   50.0,
				Discount:    10.0, // 10% discount
			},
			{
				ProductID:   uuid.New(),
				ProductName: "Product 2",
				Quantity:    3.0,
				UomID:       uomID,
				UnitPrice:   30.0,
				Discount:    0.0,
			},
		},
	}

	// Execute
	err := service.CalculateOrderAmounts(&order)

	// Assert
	require.NoError(t, err)

	// Product 1: 2 * 50 = 100, 10% discount = 90
	assert.Equal(t, 90.0, order.Lines[0].PriceSubtotal)
	assert.Equal(t, 0.0, order.Lines[0].PriceTax)
	assert.Equal(t, 90.0, order.Lines[0].PriceTotal)

	// Product 2: 3 * 30 = 90, no discount = 90
	assert.Equal(t, 90.0, order.Lines[1].PriceSubtotal)
	assert.Equal(t, 0.0, order.Lines[1].PriceTax)
	assert.Equal(t, 90.0, order.Lines[1].PriceTotal)

	// Order totals
	assert.Equal(t, 180.0, order.AmountUntaxed) // 90 + 90
	assert.Equal(t, 0.0, order.AmountTax)       // 0 + 0
	assert.Equal(t, 180.0, order.AmountTotal)   // 90 + 90
}

func TestSalesOrderService_DeleteSalesOrder_ConfirmedOrder(t *testing.T) {
	// Setup
	mockOrderRepo := new(MockSalesOrderRepository)
	mockPricelistRepo := new(MockPricelistRepository)
	service := service.NewSalesOrderService(mockOrderRepo, mockPricelistRepo)

	// Test data - confirmed order
	orderID := uuid.New()
	existingOrder := types.SalesOrder{
		ID:     orderID,
		Status: types.SalesOrderStatusConfirmed,
	}

	// Mock expectations
	mockOrderRepo.On("FindByID", context.Background(), orderID).Return(&existingOrder, nil)

	// Execute
	err := service.DeleteSalesOrder(context.Background(), orderID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete confirmed or done sales orders")

	mockOrderRepo.AssertExpectations(t)
}
