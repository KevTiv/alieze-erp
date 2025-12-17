package service

import (
	"context"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockLotService handles business logic for stock lots
type StockLotService struct {
	repo   *repository.StockLotRepository
}

// NewStockLotService creates a new StockLotService
func NewStockLotService(repo *repository.StockLotRepository) *StockLotService {
	return &StockLotService{
		repo: repo,
	}
}

// Create creates a new stock lot
func (s *StockLotService) Create(ctx context.Context, orgID uuid.UUID, req types.StockLotCreateRequest) (*types.StockLot, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a stock lot by ID
func (s *StockLotService) GetByID(ctx context.Context, id uuid.UUID) (*types.StockLot, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all stock lots for an organization
func (s *StockLotService) List(ctx context.Context, orgID uuid.UUID) ([]types.StockLot, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a stock lot
func (s *StockLotService) Update(ctx context.Context, id uuid.UUID, req types.StockLotUpdateRequest) (*types.StockLot, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a stock lot
func (s *StockLotService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
