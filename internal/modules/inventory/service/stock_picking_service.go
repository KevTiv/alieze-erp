package service

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockPickingService handles business logic for stock pickings
type StockPickingService struct {
	repo   *repository.StockPickingRepository
}

// NewStockPickingService creates a new StockPickingService
func NewStockPickingService(repo *repository.StockPickingRepository) *StockPickingService {
	return &StockPickingService{
		repo: repo,
	}
}

// Create creates a new stock picking
func (s *StockPickingService) Create(ctx context.Context, orgID uuid.UUID, req types.StockPickingCreateRequest) (*types.StockPicking, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a stock picking by ID
func (s *StockPickingService) GetByID(ctx context.Context, id uuid.UUID) (*types.StockPicking, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all stock pickings for an organization
func (s *StockPickingService) List(ctx context.Context, orgID uuid.UUID) ([]types.StockPicking, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a stock picking
func (s *StockPickingService) Update(ctx context.Context, id uuid.UUID, req types.StockPickingUpdateRequest) (*types.StockPicking, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a stock picking
func (s *StockPickingService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// GetStockPicking retrieves a stock picking by ID (alias for GetByID for interface compatibility)
func (s *StockPickingService) GetStockPicking(ctx context.Context, id uuid.UUID) (*types.StockPicking, error) {
	return s.repo.GetByID(ctx, id)
}
