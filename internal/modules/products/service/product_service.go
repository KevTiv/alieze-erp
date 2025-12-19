package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/KevTiv/alieze-erp/internal/modules/products/types"
	"github.com/KevTiv/alieze-erp/internal/modules/products/repository"

	"github.com/google/uuid"
)

// AuthService defines the interface for authentication/authorization
type AuthService interface {
	GetOrganizationID(ctx context.Context) (uuid.UUID, error)
	GetUserID(ctx context.Context) (uuid.UUID, error)
	CheckPermission(ctx context.Context, permission string) error
}

// ProductService handles product business logic
type ProductService struct {
	repo        repository.ProductRepo
	authService AuthService
	logger      *log.Logger
}

func NewProductService(repo repository.ProductRepo, authService AuthService) *ProductService {
	return &ProductService{
		repo:        repo,
		authService: authService,
		logger:      log.New(log.Writer(), "product-service: ", log.LstdFlags),
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product types.Product) (*types.Product, error) {
	// Validate required fields
	if product.Name == "" {
		return nil, errors.New("product name is required")
	}

	// Set organization from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	product.OrganizationID = orgID

	// Set defaults
	if product.ProductType == "" {
		product.ProductType = "storable"
	}

	if product.Active == false {
		product.Active = true
	}

	// Validate product type
	if !isValidProductType(product.ProductType) {
		return nil, errors.New("invalid product type")
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "products:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Create the product
	created, err := s.repo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	s.logger.Printf("Created product %s for organization %s", created.ID, created.OrganizationID)

	return created, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*types.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product id")
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "products:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if product.OrganizationID != orgID {
		return nil, fmt.Errorf("product does not belong to organization %s", orgID)
	}

	return product, nil
}

func (s *ProductService) ListProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, int, error) {
	// Set organization from context
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "products:read"); err != nil {
		return nil, 0, fmt.Errorf("permission denied: %w", err)
	}

	// Set default pagination
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	products, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	count, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	return products, count, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, product types.Product) (*types.Product, error) {
	if product.ID == uuid.Nil {
		return nil, errors.New("product id is required")
	}

	if product.Name == "" {
		return nil, errors.New("product name is required")
	}

	// Get existing product to verify organization
	existing, err := s.repo.FindByID(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing product: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("product does not belong to organization %s", orgID)
	}

	// Set organization
	product.OrganizationID = orgID

	// Validate product type
	if !isValidProductType(product.ProductType) {
		return nil, errors.New("invalid product type")
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "products:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	updated, err := s.repo.Update(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	s.logger.Printf("Updated product %s for organization %s", updated.ID, updated.OrganizationID)

	return updated, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid product id")
	}

	// Get existing product to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find existing product: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("product does not belong to organization %s", orgID)
	}

	// Check permissions
	if err := s.authService.CheckPermission(ctx, "products:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	s.logger.Printf("Deleted product %s for organization %s", id, orgID)

	return nil
}

// Helper validation functions
func isValidProductType(productType string) bool {
	switch strings.ToLower(productType) {
	case "consumable", "service", "storable":
		return true
	default:
		return false
	}
}
