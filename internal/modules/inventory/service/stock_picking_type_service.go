package service

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockPickingTypeService handles business logic for stock picking types
type StockPickingTypeService struct {
	repo   *repository.StockPickingTypeRepository
}

// NewStockPickingTypeService creates a new StockPickingTypeService
func NewStockPickingTypeService(repo *repository.StockPickingTypeRepository) *StockPickingTypeService {
	return &StockPickingTypeService{
		repo: repo,
	}
}

// Create creates a new stock picking type
func (s *StockPickingTypeService) Create(ctx context.Context, orgID uuid.UUID, req types.StockPickingTypeCreateRequest) (*types.StockPickingType, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a stock picking type by ID
func (s *StockPickingTypeService) GetByID(ctx context.Context, id uuid.UUID) (*types.StockPickingType, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all stock picking types for an organization
func (s *StockPickingTypeService) List(ctx context.Context, orgID uuid.UUID) ([]types.StockPickingType, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a stock picking type
func (s *StockPickingTypeService) Update(ctx context.Context, id uuid.UUID, req types.StockPickingTypeUpdateRequest) (*types.StockPickingType, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a stock picking type
func (s *StockPickingTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
