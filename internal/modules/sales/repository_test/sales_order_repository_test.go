package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"alieze-erp/internal/modules/sales/domain"
	"alieze-erp/internal/modules/sales/repository"
	"alieze-erp/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSalesOrderRepository_Create(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Test data
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	order := domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-001",
		Status:         domain.SalesOrderStatusDraft,
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
		Lines: []domain.SalesOrderLine{
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

	// Execute
	createdOrder, err := repo.Create(context.Background(), order)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdOrder)
	assert.Equal(t, order.ID, createdOrder.ID)
	assert.Equal(t, order.Reference, createdOrder.Reference)
	assert.Equal(t, order.Status, createdOrder.Status)
	assert.Len(t, createdOrder.Lines, 1)
	assert.Equal(t, order.Lines[0].ProductName, createdOrder.Lines[0].ProductName)
}

func TestSalesOrderRepository_FindByID(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Create test data first
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	order := domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-002",
		Status:         domain.SalesOrderStatusDraft,
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
		Lines: []domain.SalesOrderLine{
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

	// Create the order first
	createdOrder, err := repo.Create(context.Background(), order)
	require.NoError(t, err)

	// Execute
	foundOrder, err := repo.FindByID(context.Background(), createdOrder.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, foundOrder)
	assert.Equal(t, createdOrder.ID, foundOrder.ID)
	assert.Equal(t, createdOrder.Reference, foundOrder.Reference)
	assert.Equal(t, createdOrder.Status, foundOrder.Status)
	assert.Len(t, foundOrder.Lines, 1)
	assert.Equal(t, createdOrder.Lines[0].ProductName, foundOrder.Lines[0].ProductName)
}

func TestSalesOrderRepository_FindAll(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	// Create multiple orders
	for i := 0; i < 3; i++ {
		order := domain.SalesOrder{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			CustomerID:     customerID,
			Reference:      fmt.Sprintf("TEST-%03d", i),
			Status:         domain.SalesOrderStatusDraft,
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
			Lines: []domain.SalesOrderLine{
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
		_, err := repo.Create(context.Background(), order)
		require.NoError(t, err)
	}

	// Execute
	filters := repository.SalesOrderFilter{
		Limit:  10,
		Offset: 0,
	}
	orders, err := repo.FindAll(context.Background(), filters)

	// Assert
	require.NoError(t, err)
	assert.Len(t, orders, 3)
	for _, order := range orders {
		assert.Equal(t, orgID, order.OrganizationID)
		assert.Equal(t, companyID, order.CompanyID)
		assert.Equal(t, customerID, order.CustomerID)
		assert.Len(t, order.Lines, 1)
	}
}

func TestSalesOrderRepository_Update(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Create test data first
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	order := domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-003",
		Status:         domain.SalesOrderStatusDraft,
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
		Lines: []domain.SalesOrderLine{
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

	// Create the order first
	createdOrder, err := repo.Create(context.Background(), order)
	require.NoError(t, err)

	// Modify the order
	createdOrder.Reference = "UPDATED-003"
	createdOrder.Status = domain.SalesOrderStatusConfirmed
	createdOrder.Note = "Updated test order"
	createdOrder.UpdatedAt = time.Now()

	// Execute
	updatedOrder, err := repo.Update(context.Background(), *createdOrder)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, updatedOrder)
	assert.Equal(t, "UPDATED-003", updatedOrder.Reference)
	assert.Equal(t, domain.SalesOrderStatusConfirmed, updatedOrder.Status)
	assert.Equal(t, "Updated test order", updatedOrder.Note)
	assert.Len(t, updatedOrder.Lines, 1)
}

func TestSalesOrderRepository_Delete(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Create test data first
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	order := domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "TEST-004",
		Status:         domain.SalesOrderStatusDraft,
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
		Lines: []domain.SalesOrderLine{
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

	// Create the order first
	createdOrder, err := repo.Create(context.Background(), order)
	require.NoError(t, err)

	// Execute
	err = repo.Delete(context.Background(), createdOrder.ID)

	// Assert
	require.NoError(t, err)

	// Verify deletion
	deletedOrder, err := repo.FindByID(context.Background(), createdOrder.ID)
	require.NoError(t, err)
	assert.Nil(t, deletedOrder)
}

func TestSalesOrderRepository_FindByCustomerID(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	customerID1 := uuid.New()
	customerID2 := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	// Create orders for customer 1
	for i := 0; i < 2; i++ {
		order := domain.SalesOrder{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			CustomerID:     customerID1,
			Reference:      fmt.Sprintf("CUST1-%03d", i),
			Status:         domain.SalesOrderStatusDraft,
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
			Lines: []domain.SalesOrderLine{
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
		_, err := repo.Create(context.Background(), order)
		require.NoError(t, err)
	}

	// Create order for customer 2
	order := domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID2,
		Reference:      "CUST2-001",
		Status:         domain.SalesOrderStatusDraft,
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
		Lines: []domain.SalesOrderLine{
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
	_, err := repo.Create(context.Background(), order)
	require.NoError(t, err)

	// Execute
	orders, err := repo.FindByCustomerID(context.Background(), customerID1)

	// Assert
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	for _, order := range orders {
		assert.Equal(t, customerID1, order.CustomerID)
		assert.Equal(t, orgID, order.OrganizationID)
		assert.Equal(t, companyID, order.CompanyID)
		assert.Len(t, order.Lines, 1)
	}
}

func TestSalesOrderRepository_FindByStatus(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewSalesOrderRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	customerID := uuid.New()
	pricelistID := uuid.New()
	currencyID := uuid.New()
	productID := uuid.New()
	uomID := uuid.New()

	// Create draft orders
	for i := 0; i < 2; i++ {
		order := domain.SalesOrder{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			CustomerID:     customerID,
			Reference:      fmt.Sprintf("DRAFT-%03d", i),
			Status:         domain.SalesOrderStatusDraft,
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
			Lines: []domain.SalesOrderLine{
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
		_, err := repo.Create(context.Background(), order)
		require.NoError(t, err)
	}

	// Create confirmed order
	order := domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		CustomerID:     customerID,
		Reference:      "CONF-001",
		Status:         domain.SalesOrderStatusConfirmed,
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
		Lines: []domain.SalesOrderLine{
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
	_, err := repo.Create(context.Background(), order)
	require.NoError(t, err)

	// Execute
	orders, err := repo.FindByStatus(context.Background(), domain.SalesOrderStatusDraft)

	// Assert
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	for _, order := range orders {
		assert.Equal(t, domain.SalesOrderStatusDraft, order.Status)
		assert.Equal(t, orgID, order.OrganizationID)
		assert.Equal(t, companyID, order.CompanyID)
		assert.Len(t, order.Lines, 1)
	}
}
