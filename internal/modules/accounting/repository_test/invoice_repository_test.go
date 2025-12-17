package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"alieze-erp/internal/modules/accounting/domain"
	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceRepository_Create(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Test data
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID,
		Reference:      "INV-001",
		Status:         types.InvoiceStatusDraft,
		Type:           types.InvoiceTypeCustomer,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	// Execute
	createdInvoice, err := repo.Create(context.Background(), invoice)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdInvoice)
	assert.Equal(t, invoice.ID, createdInvoice.ID)
	assert.Equal(t, invoice.Reference, createdInvoice.Reference)
	assert.Equal(t, invoice.Status, createdInvoice.Status)
	assert.Equal(t, invoice.Type, createdInvoice.Type)
	assert.Len(t, createdInvoice.Lines, 1)
	assert.Equal(t, invoice.Lines[0].ProductName, createdInvoice.Lines[0].ProductName)
	assert.Equal(t, invoice.AmountTotal, createdInvoice.AmountResidual)
}

func TestInvoiceRepository_FindByID(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data first
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID,
		Reference:      "INV-002",
		Status:         types.InvoiceStatusDraft,
		Type:           types.InvoiceTypeCustomer,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	// Create the invoice first
	createdInvoice, err := repo.Create(context.Background(), invoice)
	require.NoError(t, err)

	// Execute
	foundInvoice, err := repo.FindByID(context.Background(), createdInvoice.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, foundInvoice)
	assert.Equal(t, createdInvoice.ID, foundInvoice.ID)
	assert.Equal(t, createdInvoice.Reference, foundInvoice.Reference)
	assert.Equal(t, createdInvoice.Status, foundInvoice.Status)
	assert.Equal(t, createdInvoice.Type, foundInvoice.Type)
	assert.Len(t, foundInvoice.Lines, 1)
	assert.Equal(t, createdInvoice.Lines[0].ProductName, foundInvoice.Lines[0].ProductName)
	assert.Equal(t, createdInvoice.AmountTotal, foundInvoice.AmountResidual)
	assert.Len(t, foundInvoice.Payments, 0)
}

func TestInvoiceRepository_FindAll(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	// Create multiple invoices
	for i := 0; i < 3; i++ {
		invoice := types.Invoice{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			PartnerID:      partnerID,
			Reference:      fmt.Sprintf("INV-%03d", i),
			Status:         types.InvoiceStatusDraft,
			Type:           types.InvoiceTypeCustomer,
			InvoiceDate:    time.Now(),
			DueDate:        time.Now().AddDate(0, 0, 30),
			CurrencyID:     currencyID,
			JournalID:      journalID,
			AmountUntaxed:  100.0,
			AmountTax:      20.0,
			AmountTotal:    120.0,
			AmountResidual: 120.0,
			Note:           "Test invoice",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      uuid.New(),
			UpdatedBy:      uuid.New(),
			Lines: []types.InvoiceLine{
				{
					ID:            uuid.New(),
					ProductName:   "Test Product",
					Description:   "Test Description",
					Quantity:      2.0,
					UnitPrice:     50.0,
					Discount:      0.0,
					PriceSubtotal: 100.0,
					PriceTax:      20.0,
					PriceTotal:    120.0,
					Sequence:      1,
					AccountID:     accountID,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				},
			},
		}
		_, err := repo.Create(context.Background(), invoice)
		require.NoError(t, err)
	}

	// Execute
	filters := repository.InvoiceFilter{
		Limit:  10,
		Offset: 0,
	}
	invoices, err := repo.FindAll(context.Background(), filters)

	// Assert
	require.NoError(t, err)
	assert.Len(t, invoices, 3)
	for _, invoice := range invoices {
		assert.Equal(t, orgID, invoice.OrganizationID)
		assert.Equal(t, companyID, invoice.CompanyID)
		assert.Equal(t, partnerID, invoice.PartnerID)
		assert.Equal(t, types.InvoiceTypeCustomer, invoice.Type)
		assert.Len(t, invoice.Lines, 1)
		assert.Len(t, invoice.Payments, 0)
	}
}

func TestInvoiceRepository_Update(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data first
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID,
		Reference:      "INV-003",
		Status:         types.InvoiceStatusDraft,
		Type:           types.InvoiceTypeCustomer,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	// Create the invoice first
	createdInvoice, err := repo.Create(context.Background(), invoice)
	require.NoError(t, err)

	// Modify the invoice
	createdInvoice.Reference = "UPDATED-003"
	createdInvoice.Status = types.InvoiceStatusOpen
	createdInvoice.Note = "Updated test invoice"
	createdInvoice.AmountResidual = 100.0
	createdInvoice.UpdatedAt = time.Now()

	// Execute
	updatedInvoice, err := repo.Update(context.Background(), *createdInvoice)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, updatedInvoice)
	assert.Equal(t, "UPDATED-003", updatedInvoice.Reference)
	assert.Equal(t, types.InvoiceStatusOpen, updatedInvoice.Status)
	assert.Equal(t, "Updated test invoice", updatedInvoice.Note)
	assert.Equal(t, 100.0, updatedInvoice.AmountResidual)
	assert.Len(t, updatedInvoice.Lines, 1)
}

func TestInvoiceRepository_Delete(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data first
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID,
		Reference:      "INV-004",
		Status:         types.InvoiceStatusDraft,
		Type:           types.InvoiceTypeCustomer,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	// Create the invoice first
	createdInvoice, err := repo.Create(context.Background(), invoice)
	require.NoError(t, err)

	// Execute
	err = repo.Delete(context.Background(), createdInvoice.ID)

	// Assert
	require.NoError(t, err)

	// Verify deletion
	deletedInvoice, err := repo.FindByID(context.Background(), createdInvoice.ID)
	require.NoError(t, err)
	assert.Nil(t, deletedInvoice)
}

func TestInvoiceRepository_FindByPartnerID(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID1 := uuid.New()
	partnerID2 := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	// Create invoices for partner 1
	for i := 0; i < 2; i++ {
		invoice := types.Invoice{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			PartnerID:      partnerID1,
			Reference:      fmt.Sprintf("PART1-%03d", i),
			Status:         types.InvoiceStatusDraft,
			Type:           types.InvoiceTypeCustomer,
			InvoiceDate:    time.Now(),
			DueDate:        time.Now().AddDate(0, 0, 30),
			CurrencyID:     currencyID,
			JournalID:      journalID,
			AmountUntaxed:  100.0,
			AmountTax:      20.0,
			AmountTotal:    120.0,
			AmountResidual: 120.0,
			Note:           "Test invoice",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      uuid.New(),
			UpdatedBy:      uuid.New(),
			Lines: []types.InvoiceLine{
				{
					ID:            uuid.New(),
					ProductName:   "Test Product",
					Description:   "Test Description",
					Quantity:      2.0,
					UnitPrice:     50.0,
					Discount:      0.0,
					PriceSubtotal: 100.0,
					PriceTax:      20.0,
					PriceTotal:    120.0,
					Sequence:      1,
					AccountID:     accountID,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				},
			},
		}
		_, err := repo.Create(context.Background(), invoice)
		require.NoError(t, err)
	}

	// Create invoice for partner 2
	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID2,
		Reference:      "PART2-001",
		Status:         types.InvoiceStatusDraft,
		Type:           types.InvoiceTypeCustomer,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}
	_, err := repo.Create(context.Background(), invoice)
	require.NoError(t, err)

	// Execute
	invoices, err := repo.FindByPartnerID(context.Background(), partnerID1)

	// Assert
	require.NoError(t, err)
	assert.Len(t, invoices, 2)
	for _, invoice := range invoices {
		assert.Equal(t, partnerID1, invoice.PartnerID)
		assert.Equal(t, orgID, invoice.OrganizationID)
		assert.Equal(t, companyID, invoice.CompanyID)
		assert.Equal(t, types.InvoiceTypeCustomer, invoice.Type)
		assert.Len(t, invoice.Lines, 1)
		assert.Len(t, invoice.Payments, 0)
	}
}

func TestInvoiceRepository_FindByStatus(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	// Create draft invoices
	for i := 0; i < 2; i++ {
		invoice := types.Invoice{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			PartnerID:      partnerID,
			Reference:      fmt.Sprintf("DRAFT-%03d", i),
			Status:         types.InvoiceStatusDraft,
			Type:           types.InvoiceTypeCustomer,
			InvoiceDate:    time.Now(),
			DueDate:        time.Now().AddDate(0, 0, 30),
			CurrencyID:     currencyID,
			JournalID:      journalID,
			AmountUntaxed:  100.0,
			AmountTax:      20.0,
			AmountTotal:    120.0,
			AmountResidual: 120.0,
			Note:           "Test invoice",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      uuid.New(),
			UpdatedBy:      uuid.New(),
			Lines: []types.InvoiceLine{
				{
					ID:            uuid.New(),
					ProductName:   "Test Product",
					Description:   "Test Description",
					Quantity:      2.0,
					UnitPrice:     50.0,
					Discount:      0.0,
					PriceSubtotal: 100.0,
					PriceTax:      20.0,
					PriceTotal:    120.0,
					Sequence:      1,
					AccountID:     accountID,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				},
			},
		}
		_, err := repo.Create(context.Background(), invoice)
		require.NoError(t, err)
	}

	// Create open invoice
	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID,
		Reference:      "OPEN-001",
		Status:         types.InvoiceStatusOpen,
		Type:           types.InvoiceTypeCustomer,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}
	_, err := repo.Create(context.Background(), invoice)
	require.NoError(t, err)

	// Execute
	invoices, err := repo.FindByStatus(context.Background(), types.InvoiceStatusDraft)

	// Assert
	require.NoError(t, err)
	assert.Len(t, invoices, 2)
	for _, invoice := range invoices {
		assert.Equal(t, types.InvoiceStatusDraft, invoice.Status)
		assert.Equal(t, orgID, invoice.OrganizationID)
		assert.Equal(t, companyID, invoice.CompanyID)
		assert.Equal(t, partnerID, invoice.PartnerID)
		assert.Equal(t, types.InvoiceTypeCustomer, invoice.Type)
		assert.Len(t, invoice.Lines, 1)
		assert.Len(t, invoice.Payments, 0)
	}
}

func TestInvoiceRepository_FindByType(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB(t)
	defer testutils.TeardownTestDB(t, db)

	repo := repository.NewInvoiceRepository(db)

	// Create test data
	orgID := uuid.New()
	companyID := uuid.New()
	partnerID := uuid.New()
	currencyID := uuid.New()
	journalID := uuid.New()
	accountID := uuid.New()

	// Create customer invoices
	for i := 0; i < 2; i++ {
		invoice := types.Invoice{
			ID:             uuid.New(),
			OrganizationID: orgID,
			CompanyID:      companyID,
			PartnerID:      partnerID,
			Reference:      fmt.Sprintf("CUST-%03d", i),
			Status:         types.InvoiceStatusDraft,
			Type:           types.InvoiceTypeCustomer,
			InvoiceDate:    time.Now(),
			DueDate:        time.Now().AddDate(0, 0, 30),
			CurrencyID:     currencyID,
			JournalID:      journalID,
			AmountUntaxed:  100.0,
			AmountTax:      20.0,
			AmountTotal:    120.0,
			AmountResidual: 120.0,
			Note:           "Test invoice",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      uuid.New(),
			UpdatedBy:      uuid.New(),
			Lines: []types.InvoiceLine{
				{
					ID:            uuid.New(),
					ProductName:   "Test Product",
					Description:   "Test Description",
					Quantity:      2.0,
					UnitPrice:     50.0,
					Discount:      0.0,
					PriceSubtotal: 100.0,
					PriceTax:      20.0,
					PriceTotal:    120.0,
					Sequence:      1,
					AccountID:     accountID,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				},
			},
		}
		_, err := repo.Create(context.Background(), invoice)
		require.NoError(t, err)
	}

	// Create supplier invoice
	invoice := types.Invoice{
		ID:             uuid.New(),
		OrganizationID: orgID,
		CompanyID:      companyID,
		PartnerID:      partnerID,
		Reference:      "SUPP-001",
		Status:         types.InvoiceStatusDraft,
		Type:           types.InvoiceTypeSupplier,
		InvoiceDate:    time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyID:     currencyID,
		JournalID:      journalID,
		AmountUntaxed:  100.0,
		AmountTax:      20.0,
		AmountTotal:    120.0,
		AmountResidual: 120.0,
		Note:           "Test invoice",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		Lines: []types.InvoiceLine{
			{
				ID:            uuid.New(),
				ProductName:   "Test Product",
				Description:   "Test Description",
				Quantity:      2.0,
				UnitPrice:     50.0,
				Discount:      0.0,
				PriceSubtotal: 100.0,
				PriceTax:      20.0,
				PriceTotal:    120.0,
				Sequence:      1,
				AccountID:     accountID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}
	_, err := repo.Create(context.Background(), invoice)
	require.NoError(t, err)

	// Execute
	invoices, err := repo.FindByType(context.Background(), types.InvoiceTypeCustomer)

	// Assert
	require.NoError(t, err)
	assert.Len(t, invoices, 2)
	for _, invoice := range invoices {
		assert.Equal(t, types.InvoiceTypeCustomer, invoice.Type)
		assert.Equal(t, orgID, invoice.OrganizationID)
		assert.Equal(t, companyID, invoice.CompanyID)
		assert.Equal(t, partnerID, invoice.PartnerID)
		assert.Equal(t, types.InvoiceStatusDraft, invoice.Status)
		assert.Len(t, invoice.Lines, 1)
		assert.Len(t, invoice.Payments, 0)
	}
}
