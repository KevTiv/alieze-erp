package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"alieze-erp/internal/modules/products/repository"
	"alieze-erp/internal/modules/products/types"
	"alieze-erp/internal/testutils"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	repo       *repository.ProductRepository
	mockDB     *testutils.MockDB
	ctx        context.Context
	orgID      uuid.UUID
	productID  uuid.UUID
	categoryID uuid.UUID
}

func (s *ProductRepositoryTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.mockDB = testutils.SetupMockDB(s.T())
	s.repo = repository.NewProductRepository(s.mockDB.DB)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.productID = uuid.Must(uuid.NewV7())
	s.categoryID = uuid.Must(uuid.NewV7())
}

func (s *ProductRepositoryTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
	// Don't close the mock DB as it causes issues with sqlmock
	// The mock DB will be automatically cleaned up when the test ends
}

func (s *ProductRepositoryTestSuite) TestCreateProductSuccess() {
	s.T().Run("CreateProduct - Success", func(t *testing.T) {
		// Setup test data
		product := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Test Product",
			DefaultCode:    stringPtr("TEST-001"),
			Barcode:        stringPtr("1234567890"),
			ProductType:    "storable",
			CategoryID:     &s.categoryID,
			ListPrice:      float64Ptr(99.99),
			Active:         true,
		}

		// Setup mock expectations
		s.mockDB.Mock.ExpectQuery("INSERT INTO products").
			WithArgs(
				product.ID,
				product.OrganizationID,
				product.Name,
				product.DefaultCode,
				product.Barcode,
				product.ProductType,
				product.CategoryID,
				product.ListPrice,
				product.Active,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				nil,              // deleted_at
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "default_code", "barcode", "product_type",
				"category_id", "list_price", "active", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				product.ID,
				product.OrganizationID,
				product.Name,
				product.DefaultCode,
				product.Barcode,
				product.ProductType,
				product.CategoryID,
				product.ListPrice,
				product.Active,
				time.Now(),
				time.Now(),
				nil,
			))

		// Execute
		created, err := s.repo.Create(s.ctx, product)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, product.ID, created.ID)
		require.Equal(t, product.Name, created.Name)
		require.Equal(t, product.OrganizationID, created.OrganizationID)

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestCreateProductMissingOrganization() {
	s.T().Run("CreateProduct - Missing Organization", func(t *testing.T) {
		product := types.Product{
			Name:        "Test Product",
			ProductType: "storable",
		}

		created, err := s.repo.Create(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "organization_id is required")
	})
}

func (s *ProductRepositoryTestSuite) TestCreateProductMissingName() {
	s.T().Run("CreateProduct - Missing Name", func(t *testing.T) {
		product := types.Product{
			OrganizationID: s.orgID,
			ProductType:    "storable",
		}

		created, err := s.repo.Create(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "name is required")
	})
}

func (s *ProductRepositoryTestSuite) TestFindByIDSuccess() {
	s.T().Run("FindByID - Success", func(t *testing.T) {
		// Setup test data
		expectedProduct := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Test Product",
			DefaultCode:    stringPtr("TEST-001"),
			Barcode:        stringPtr("1234567890"),
			ProductType:    "storable",
			CategoryID:     &s.categoryID,
			ListPrice:      float64Ptr(99.99),
			Active:         true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Setup mock expectations
		s.mockDB.Mock.ExpectQuery("SELECT.*FROM products").
			WithArgs(s.productID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "default_code", "barcode", "product_type",
				"category_id", "list_price", "active", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				expectedProduct.ID,
				expectedProduct.OrganizationID,
				expectedProduct.Name,
				expectedProduct.DefaultCode,
				expectedProduct.Barcode,
				expectedProduct.ProductType,
				expectedProduct.CategoryID,
				expectedProduct.ListPrice,
				expectedProduct.Active,
				expectedProduct.CreatedAt,
				expectedProduct.UpdatedAt,
				nil,
			))

		// Execute
		product, err := s.repo.FindByID(s.ctx, s.productID)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, product)
		require.Equal(t, expectedProduct.ID, product.ID)
		require.Equal(t, expectedProduct.Name, product.Name)

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestFindByIDNotFound() {
	s.T().Run("FindByID - Not Found", func(t *testing.T) {
		// Setup mock expectations
		s.mockDB.Mock.ExpectQuery("SELECT.*FROM products").
			WithArgs(s.productID).
			WillReturnError(sql.ErrNoRows)

		// Execute
		product, err := s.repo.FindByID(s.ctx, s.productID)

		// Verify
		require.Error(t, err)
		require.Nil(t, product)
		require.Contains(t, err.Error(), "product not found")

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestFindByIDInvalidID() {
	s.T().Run("FindByID - Invalid ID", func(t *testing.T) {
		product, err := s.repo.FindByID(s.ctx, uuid.Nil)

		require.Error(t, err)
		require.Nil(t, product)
		require.Contains(t, err.Error(), "invalid product id")
	})
}

func (s *ProductRepositoryTestSuite) TestFindAllSuccess() {
	s.T().Run("FindAll - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ProductFilter{
			OrganizationID: s.orgID,
			Name:           stringPtr("Test"),
			Active:         boolPtr(true),
			Limit:          10,
			Offset:         0,
		}

		// Setup mock expectations - the query will have organization_id, name pattern, and active
		s.mockDB.Mock.ExpectQuery("SELECT.*FROM products").
			WithArgs(
				filter.OrganizationID,
				"%"+*filter.Name+"%", // name pattern with wildcards
				*filter.Active,
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "default_code", "barcode", "product_type",
				"category_id", "list_price", "active", "created_at", "updated_at", "deleted_at",
			}).
				AddRow(uuid.Must(uuid.NewV7()), s.orgID, "Test Product 1", "TEST-001", "1234567890", "storable", s.categoryID, 99.99, true, time.Now(), time.Now(), nil).
				AddRow(uuid.Must(uuid.NewV7()), s.orgID, "Test Product 2", "TEST-002", "0987654321", "service", nil, 49.99, true, time.Now(), time.Now(), nil))

		// Execute
		products, err := s.repo.FindAll(s.ctx, filter)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, products)
		require.Len(t, products, 2)
		require.Equal(t, "Test Product 1", products[0].Name)
		require.Equal(t, "Test Product 2", products[1].Name)

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestFindAllEmptyResult() {
	s.T().Run("FindAll - Empty Result", func(t *testing.T) {
		filter := types.ProductFilter{
			OrganizationID: s.orgID,
			Name:           stringPtr("NonExistent"),
		}

		s.mockDB.Mock.ExpectQuery("SELECT.*FROM products").
			WithArgs(
				filter.OrganizationID,
				"%"+*filter.Name+"%", // name pattern with wildcards
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "default_code", "barcode", "product_type",
				"category_id", "list_price", "active", "created_at", "updated_at", "deleted_at",
			}))

		products, err := s.repo.FindAll(s.ctx, filter)

		require.NoError(t, err)
		require.Len(t, products, 0)

		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestUpdateProductSuccess() {
	s.T().Run("UpdateProduct - Success", func(t *testing.T) {
		// Setup test data
		product := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Updated Product",
			DefaultCode:    stringPtr("UPDATED-001"),
			Barcode:        stringPtr("9876543210"),
			ProductType:    "service",
			CategoryID:     &s.categoryID,
			ListPrice:      float64Ptr(149.99),
			Active:         false,
		}

		// Setup mock expectations
		s.mockDB.Mock.ExpectQuery("UPDATE products").
			WithArgs(
				product.OrganizationID,
				product.Name,
				product.DefaultCode,
				product.Barcode,
				product.ProductType,
				product.CategoryID,
				product.ListPrice,
				product.Active,
				sqlmock.AnyArg(), // updated_at
				product.ID,
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "default_code", "barcode", "product_type",
				"category_id", "list_price", "active", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				product.ID,
				product.OrganizationID,
				product.Name,
				product.DefaultCode,
				product.Barcode,
				product.ProductType,
				product.CategoryID,
				product.ListPrice,
				product.Active,
				time.Now(),
				time.Now(),
				nil,
			))

		// Execute
		updated, err := s.repo.Update(s.ctx, product)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, product.ID, updated.ID)
		require.Equal(t, product.Name, updated.Name)
		require.Equal(t, product.Active, updated.Active)

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestUpdateProductMissingID() {
	s.T().Run("UpdateProduct - Missing ID", func(t *testing.T) {
		product := types.Product{
			OrganizationID: s.orgID,
			Name:           "Test Product",
			ProductType:    "storable",
		}

		updated, err := s.repo.Update(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, updated)
		require.Contains(t, err.Error(), "product id is required")
	})
}

func (s *ProductRepositoryTestSuite) TestDeleteProductSuccess() {
	s.T().Run("DeleteProduct - Success", func(t *testing.T) {
		// Setup mock expectations
		s.mockDB.Mock.ExpectExec("UPDATE products").
			WithArgs(
				sqlmock.AnyArg(), // deleted_at
				sqlmock.AnyArg(), // updated_at
				s.productID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Execute
		err := s.repo.Delete(s.ctx, s.productID)

		// Verify
		require.NoError(t, err)

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestDeleteProductNotFound() {
	s.T().Run("DeleteProduct - Not Found", func(t *testing.T) {
		// Setup mock expectations
		s.mockDB.Mock.ExpectExec("UPDATE products").
			WithArgs(
				sqlmock.AnyArg(), // deleted_at
				sqlmock.AnyArg(), // updated_at
				s.productID,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Execute
		err := s.repo.Delete(s.ctx, s.productID)

		// Verify
		require.Error(t, err)
		require.Contains(t, err.Error(), "product not found or already deleted")

		// Ensure all expectations were met
		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestDeleteProductInvalidID() {
	s.T().Run("DeleteProduct - Invalid ID", func(t *testing.T) {
		err := s.repo.Delete(s.ctx, uuid.Nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid product id")
	})
}

func (s *ProductRepositoryTestSuite) TestCountProductsSuccess() {
	s.T().Run("CountProducts - Success", func(t *testing.T) {
		filter := types.ProductFilter{
			OrganizationID: s.orgID,
			Active:         boolPtr(true),
		}

		s.mockDB.Mock.ExpectQuery("SELECT COUNT").
			WithArgs(
				filter.OrganizationID,
				filter.Active,
			).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		count, err := s.repo.Count(s.ctx, filter)

		require.NoError(t, err)
		require.Equal(t, 5, count)

		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

func (s *ProductRepositoryTestSuite) TestCountProductsZero() {
	s.T().Run("CountProducts - Zero", func(t *testing.T) {
		filter := types.ProductFilter{
			OrganizationID: s.orgID,
			Name:           stringPtr("NonExistent"),
		}

		s.mockDB.Mock.ExpectQuery("SELECT COUNT").
			WithArgs(
				filter.OrganizationID,
				sqlmock.AnyArg(), // name pattern
			).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		count, err := s.repo.Count(s.ctx, filter)

		require.NoError(t, err)
		require.Equal(t, 0, count)

		if err := s.mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expectations: %s", err)
		}
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func TestProductRepository(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}
