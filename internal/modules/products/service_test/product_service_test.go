package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"alieze-erp/internal/modules/products/domain"
	"alieze-erp/internal/modules/products/service"
	"alieze-erp/internal/testutils"
)

type ProductServiceTestSuite struct {
	suite.Suite
	service    *service.ProductService
	repo       *testutils.MockProductRepository
	auth       *testutils.MockAuthService
	ctx        context.Context
	orgID      uuid.UUID
	userID     uuid.UUID
	productID  uuid.UUID
	categoryID uuid.UUID
}

func (s *ProductServiceTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.repo = testutils.NewMockProductRepository()
	s.auth = testutils.NewMockAuthService()
	s.service = service.NewProductService(s.repo, s.auth)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.userID = uuid.Must(uuid.NewV7())
	s.productID = uuid.Must(uuid.NewV7())
	s.categoryID = uuid.Must(uuid.NewV7())

	// Default mock behavior
	s.auth.WithOrganizationID(s.orgID)
	s.auth.WithUserID(s.userID)
	s.auth.AllowPermission("products:create")
	s.auth.AllowPermission("products:read")
	s.auth.AllowPermission("products:update")
	s.auth.AllowPermission("products:delete")
}

func (s *ProductServiceTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
	// Clean up if needed
}

func (s *ProductServiceTestSuite) TestCreateProductSuccess() {
	s.T().Run("CreateProduct - Success", func(t *testing.T) {
		// Setup test data
		product := types.Product{
			Name:        "Test Product",
			DefaultCode: stringPtr("TEST-001"),
			Barcode:     stringPtr("1234567890"),
			ProductType: "storable",
			CategoryID:  &s.categoryID,
			ListPrice:   float64Ptr(99.99),
			Active:      true,
		}

		expectedProduct := product
		expectedProduct.ID = s.productID
		expectedProduct.OrganizationID = s.orgID

		// Setup mock expectations
		s.repo.WithCreateFunc(func(ctx context.Context, p types.Product) (*types.Product, error) {
			require.Equal(t, s.orgID, p.OrganizationID)
			require.Equal(t, "Test Product", p.Name)
			require.Equal(t, "storable", p.ProductType)
			return &expectedProduct, nil
		})

		// Execute
		created, err := s.service.CreateProduct(s.ctx, product)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedProduct.ID, created.ID)
		require.Equal(t, expectedProduct.Name, created.Name)
		require.Equal(t, expectedProduct.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedProduct.ProductType, created.ProductType)
	})
}

func (s *ProductServiceTestSuite) TestCreateProductMissingName() {
	s.T().Run("CreateProduct - Missing Name", func(t *testing.T) {
		product := types.Product{
			ProductType: "storable",
		}

		created, err := s.service.CreateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "product name is required")
	})
}

func (s *ProductServiceTestSuite) TestCreateProductInvalidProductType() {
	s.T().Run("CreateProduct - Invalid Product Type", func(t *testing.T) {
		product := types.Product{
			Name:        "Test Product",
			ProductType: "invalid",
		}

		created, err := s.service.CreateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "invalid product type")
	})
}

func (s *ProductServiceTestSuite) TestCreateProductPermissionDenied() {
	s.T().Run("CreateProduct - Permission Denied", func(t *testing.T) {
		// Setup permission denial
		s.auth.DenyPermission("products:create")

		product := types.Product{
			Name:        "Test Product",
			ProductType: "storable",
		}

		created, err := s.service.CreateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ProductServiceTestSuite) TestGetProductSuccess() {
	s.T().Run("GetProduct - Success", func(t *testing.T) {
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
		}

		// Setup mock expectations
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			require.Equal(t, s.productID, id)
			return &expectedProduct, nil
		})

		// Execute
		product, err := s.service.GetProduct(s.ctx, s.productID)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, product)
		require.Equal(t, expectedProduct.ID, product.ID)
		require.Equal(t, expectedProduct.Name, product.Name)
		require.Equal(t, expectedProduct.OrganizationID, product.OrganizationID)
	})
}

func (s *ProductServiceTestSuite) TestGetProductInvalidID() {
	s.T().Run("GetProduct - Invalid ID", func(t *testing.T) {
		product, err := s.service.GetProduct(s.ctx, uuid.Nil)

		require.Error(t, err)
		require.Nil(t, product)
		require.Contains(t, err.Error(), "invalid product id")
	})
}

func (s *ProductServiceTestSuite) TestGetProductPermissionDenied() {
	s.T().Run("GetProduct - Permission Denied", func(t *testing.T) {
		// Setup permission denial
		s.auth.DenyPermission("products:read")

		product, err := s.service.GetProduct(s.ctx, s.productID)

		require.Error(t, err)
		require.Nil(t, product)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ProductServiceTestSuite) TestGetProductOrganizationMismatch() {
	s.T().Run("GetProduct - Organization Mismatch", func(t *testing.T) {
		// Setup test data with different organization
		otherOrgID := uuid.Must(uuid.NewV7())
		product := types.Product{
			ID:             s.productID,
			OrganizationID: otherOrgID,
			Name:           "Test Product",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			return &product, nil
		})

		// Execute
		result, err := s.service.GetProduct(s.ctx, s.productID)

		// Verify
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *ProductServiceTestSuite) TestListProductsSuccess() {
	s.T().Run("ListProducts - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ProductFilter{
			Name:   stringPtr("Test"),
			Active: boolPtr(true),
		}

		expectedProducts := []types.Product{
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Test Product 1",
				DefaultCode:    stringPtr("TEST-001"),
				ProductType:    "storable",
				Active:         true,
			},
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "Test Product 2",
				DefaultCode:    stringPtr("TEST-002"),
				ProductType:    "service",
				Active:         true,
			},
		}

		// Setup mock expectations
		s.repo.WithFindAllFunc(func(ctx context.Context, f types.ProductFilter) ([]types.Product, error) {
			require.Equal(t, s.orgID, f.OrganizationID)
			require.Equal(t, "Test", *f.Name)
			require.Equal(t, true, *f.Active)
			return expectedProducts, nil
		})

		s.repo.WithCountFunc(func(ctx context.Context, f types.ProductFilter) (int, error) {
			return len(expectedProducts), nil
		})

		// Execute
		products, count, err := s.service.ListProducts(s.ctx, filter)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, products)
		require.Len(t, products, 2)
		require.Equal(t, 2, count)
		require.Equal(t, "Test Product 1", products[0].Name)
		require.Equal(t, "Test Product 2", products[1].Name)
	})
}

func (s *ProductServiceTestSuite) TestListProductsPermissionDenied() {
	s.T().Run("ListProducts - Permission Denied", func(t *testing.T) {
		// Setup permission denial
		s.auth.DenyPermission("products:read")

		filter := types.ProductFilter{
			Name: stringPtr("Test"),
		}

		products, count, err := s.service.ListProducts(s.ctx, filter)

		require.Error(t, err)
		require.Nil(t, products)
		require.Equal(t, 0, count)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ProductServiceTestSuite) TestListProductsDefaultPagination() {
	s.T().Run("ListProducts - Default Pagination", func(t *testing.T) {
		filter := types.ProductFilter{
			// No limit specified
		}

		expectedProducts := []types.Product{
			{ID: uuid.Must(uuid.NewV7()), OrganizationID: s.orgID, Name: "Product 1"},
		}

		s.repo.WithFindAllFunc(func(ctx context.Context, f types.ProductFilter) ([]types.Product, error) {
			require.Equal(t, 50, f.Limit) // Default limit should be 50
			return expectedProducts, nil
		})

		s.repo.WithCountFunc(func(ctx context.Context, f types.ProductFilter) (int, error) {
			return 1, nil
		})

		products, count, err := s.service.ListProducts(s.ctx, filter)

		require.NoError(t, err)
		require.NotNil(t, products)
		require.Len(t, products, 1)
		require.Equal(t, 1, count)
	})
}

func (s *ProductServiceTestSuite) TestUpdateProductSuccess() {
	s.T().Run("UpdateProduct - Success", func(t *testing.T) {
		// Setup test data
		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Old Product",
			ProductType:    "storable",
			Active:         true,
		}

		updatedProduct := types.Product{
			ID:          s.productID,
			Name:        "Updated Product",
			DefaultCode: stringPtr("UPDATED-001"),
			Barcode:     stringPtr("9876543210"),
			ProductType: "service",
			CategoryID:  &s.categoryID,
			ListPrice:   float64Ptr(149.99),
			Active:      false,
		}

		expectedProduct := updatedProduct
		expectedProduct.OrganizationID = s.orgID

		// Setup mock expectations
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			require.Equal(t, s.productID, id)
			return &existingProduct, nil
		})

		s.repo.WithUpdateFunc(func(ctx context.Context, p types.Product) (*types.Product, error) {
			require.Equal(t, s.productID, p.ID)
			require.Equal(t, s.orgID, p.OrganizationID)
			require.Equal(t, "Updated Product", p.Name)
			return &expectedProduct, nil
		})

		// Execute
		result, err := s.service.UpdateProduct(s.ctx, updatedProduct)

		// Verify
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, expectedProduct.ID, result.ID)
		require.Equal(t, expectedProduct.Name, result.Name)
		require.Equal(t, expectedProduct.ProductType, result.ProductType)
		require.Equal(t, expectedProduct.Active, result.Active)
	})
}

func (s *ProductServiceTestSuite) TestUpdateProductMissingID() {
	s.T().Run("UpdateProduct - Missing ID", func(t *testing.T) {
		product := types.Product{
			Name:        "Test Product",
			ProductType: "storable",
		}

		result, err := s.service.UpdateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "product id is required")
	})
}

func (s *ProductServiceTestSuite) TestUpdateProductMissingName() {
	s.T().Run("UpdateProduct - Missing Name", func(t *testing.T) {
		product := types.Product{
			ID:          s.productID,
			ProductType: "storable",
		}

		result, err := s.service.UpdateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "product name is required")
	})
}

func (s *ProductServiceTestSuite) TestUpdateProductInvalidProductType() {
	s.T().Run("UpdateProduct - Invalid Product Type", func(t *testing.T) {
		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Test Product",
			ProductType:    "storable",
		}

		product := types.Product{
			ID:          s.productID,
			Name:        "Test Product",
			ProductType: "invalid",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			return &existingProduct, nil
		})

		result, err := s.service.UpdateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "invalid product type")
	})
}

func (s *ProductServiceTestSuite) TestUpdateProductOrganizationMismatch() {
	s.T().Run("UpdateProduct - Organization Mismatch", func(t *testing.T) {
		// Setup test data with different organization
		otherOrgID := uuid.Must(uuid.NewV7())
		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: otherOrgID,
			Name:           "Test Product",
			ProductType:    "storable",
		}

		product := types.Product{
			ID:          s.productID,
			Name:        "Updated Product",
			ProductType: "storable",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			return &existingProduct, nil
		})

		result, err := s.service.UpdateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *ProductServiceTestSuite) TestUpdateProductPermissionDenied() {
	s.T().Run("UpdateProduct - Permission Denied", func(t *testing.T) {
		// Setup permission denial
		s.auth.DenyPermission("products:update")

		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Test Product",
			ProductType:    "storable",
		}

		product := types.Product{
			ID:          s.productID,
			Name:        "Updated Product",
			ProductType: "storable",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			return &existingProduct, nil
		})

		result, err := s.service.UpdateProduct(s.ctx, product)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ProductServiceTestSuite) TestDeleteProductSuccess() {
	s.T().Run("DeleteProduct - Success", func(t *testing.T) {
		// Setup test data
		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Test Product",
			ProductType:    "storable",
			Active:         true,
		}

		// Setup mock expectations
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			require.Equal(t, s.productID, id)
			return &existingProduct, nil
		})

		s.repo.WithDeleteFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, s.productID, id)
			return nil
		})

		// Execute
		err := s.service.DeleteProduct(s.ctx, s.productID)

		// Verify
		require.NoError(t, err)
	})
}

func (s *ProductServiceTestSuite) TestDeleteProductInvalidID() {
	s.T().Run("DeleteProduct - Invalid ID", func(t *testing.T) {
		err := s.service.DeleteProduct(s.ctx, uuid.Nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid product id")
	})
}

func (s *ProductServiceTestSuite) TestDeleteProductOrganizationMismatch() {
	s.T().Run("DeleteProduct - Organization Mismatch", func(t *testing.T) {
		// Setup test data with different organization
		otherOrgID := uuid.Must(uuid.NewV7())
		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: otherOrgID,
			Name:           "Test Product",
			ProductType:    "storable",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			return &existingProduct, nil
		})

		err := s.service.DeleteProduct(s.ctx, s.productID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *ProductServiceTestSuite) TestDeleteProductPermissionDenied() {
	s.T().Run("DeleteProduct - Permission Denied", func(t *testing.T) {
		// Setup permission denial
		s.auth.DenyPermission("products:delete")

		existingProduct := types.Product{
			ID:             s.productID,
			OrganizationID: s.orgID,
			Name:           "Test Product",
			ProductType:    "storable",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Product, error) {
			return &existingProduct, nil
		})

		err := s.service.DeleteProduct(s.ctx, s.productID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "permission denied")
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

func TestProductService(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}
