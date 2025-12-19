package service

import (
	"context"

	"alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// InventoryIntegrationService provides a facade for other modules to interact with inventory
// This service implements the interfaces required by the delivery module
type InventoryIntegrationService struct {
	stockMoveService    *StockMoveService
	stockPickingService *StockPickingService
}

// NewInventoryIntegrationService creates a new InventoryIntegrationService
func NewInventoryIntegrationService(
	stockMoveService *StockMoveService,
	stockPickingService *StockPickingService,
) *InventoryIntegrationService {
	return &InventoryIntegrationService{
		stockMoveService:    stockMoveService,
		stockPickingService: stockPickingService,
	}
}

// GetStockMovesByPickingID retrieves all stock moves for a given picking ID
func (s *InventoryIntegrationService) GetStockMovesByPickingID(ctx context.Context, pickingID uuid.UUID) ([]types.StockMove, error) {
	return s.stockMoveService.GetStockMovesByPickingID(ctx, pickingID)
}

// GetStockPicking retrieves a stock picking by ID
func (s *InventoryIntegrationService) GetStockPicking(ctx context.Context, pickingID uuid.UUID) (*types.StockPicking, error) {
	return s.stockPickingService.GetStockPicking(ctx, pickingID)
}
