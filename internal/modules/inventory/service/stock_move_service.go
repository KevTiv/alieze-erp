package service

import (
	"context"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockMoveService handles business logic for stock moves
type StockMoveService struct {
	repo   *repository.StockMoveRepository
}

// NewStockMoveService creates a new StockMoveService
func NewStockMoveService(repo *repository.StockMoveRepository) *StockMoveService {
	return &StockMoveService{
		repo: repo,
	}
}

// Create creates a new stock move
func (s *StockMoveService) Create(ctx context.Context, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a stock move by ID
func (s *StockMoveService) GetByID(ctx context.Context, id uuid.UUID) (*types.StockMove, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all stock moves for an organization
func (s *StockMoveService) List(ctx context.Context, orgID uuid.UUID) ([]types.StockMove, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a stock move
func (s *StockMoveService) Update(ctx context.Context, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a stock move
func (s *StockMoveService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
