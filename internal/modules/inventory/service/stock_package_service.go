package service

import (
	"context"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockPackageService handles business logic for stock packages
type StockPackageService struct {
	repo   *repository.StockPackageRepository
}

// NewStockPackageService creates a new StockPackageService
func NewStockPackageService(repo *repository.StockPackageRepository) *StockPackageService {
	return &StockPackageService{
		repo: repo,
	}
}

// Create creates a new stock package
func (s *StockPackageService) Create(ctx context.Context, orgID uuid.UUID, req types.StockPackageCreateRequest) (*types.StockPackage, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a stock package by ID
func (s *StockPackageService) GetByID(ctx context.Context, id uuid.UUID) (*types.StockPackage, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all stock packages for an organization
func (s *StockPackageService) List(ctx context.Context, orgID uuid.UUID) ([]types.StockPackage, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a stock package
func (s *StockPackageService) Update(ctx context.Context, id uuid.UUID, req types.StockPackageUpdateRequest) (*types.StockPackage, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a stock package
func (s *StockPackageService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
