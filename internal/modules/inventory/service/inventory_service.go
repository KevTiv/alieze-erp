package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type InventoryService struct {
	warehouseRepo repository.WarehouseRepository
	locationRepo  repository.StockLocationRepository
	quantRepo     repository.StockQuantRepository
	moveRepo      repository.StockMoveRepository
}

func NewInventoryService(
	warehouseRepo repository.WarehouseRepository,
	locationRepo repository.StockLocationRepository,
	quantRepo repository.StockQuantRepository,
	moveRepo repository.StockMoveRepository,
) *InventoryService {
	return &InventoryService{
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		quantRepo:     quantRepo,
		moveRepo:      moveRepo,
	}
}

// Warehouse operations
func (s *InventoryService) CreateWarehouse(ctx context.Context, wh domain.Warehouse) (*domain.Warehouse, error) {
	if wh.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if wh.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if wh.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	// Set defaults
	if wh.ReceptionSteps == "" {
		wh.ReceptionSteps = "one_step"
	}
	if wh.DeliverySteps == "" {
		wh.DeliverySteps = "ship_only"
	}
	if wh.Sequence == 0 {
		wh.Sequence = 10
	}
	wh.Active = true

	return s.warehouseRepo.Create(ctx, wh)
}

func (s *InventoryService) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	wh, err := s.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wh == nil {
		return nil, fmt.Errorf("warehouse not found")
	}
	return wh, nil
}

func (s *InventoryService) ListWarehouses(ctx context.Context, organizationID uuid.UUID) ([]domain.Warehouse, error) {
	return s.warehouseRepo.FindAll(ctx, organizationID)
}

func (s *InventoryService) UpdateWarehouse(ctx context.Context, wh domain.Warehouse) (*domain.Warehouse, error) {
	return s.warehouseRepo.Update(ctx, wh)
}

func (s *InventoryService) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	return s.warehouseRepo.Delete(ctx, id)
}

// Location operations
func (s *InventoryService) CreateLocation(ctx context.Context, loc domain.StockLocation) (*domain.StockLocation, error) {
	if loc.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if loc.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if loc.Usage == "" {
		loc.Usage = "internal"
	}
	if loc.RemovalStrategy == "" {
		loc.RemovalStrategy = "fifo"
	}
	loc.Active = true

	return s.locationRepo.Create(ctx, loc)
}

func (s *InventoryService) GetLocation(ctx context.Context, id uuid.UUID) (*domain.StockLocation, error) {
	loc, err := s.locationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, fmt.Errorf("location not found")
	}
	return loc, nil
}

func (s *InventoryService) ListLocations(ctx context.Context, organizationID uuid.UUID) ([]domain.StockLocation, error) {
	return s.locationRepo.FindAll(ctx, organizationID)
}

func (s *InventoryService) UpdateLocation(ctx context.Context, loc domain.StockLocation) (*domain.StockLocation, error) {
	return s.locationRepo.Update(ctx, loc)
}

func (s *InventoryService) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	return s.locationRepo.Delete(ctx, id)
}

// Stock operations
func (s *InventoryService) GetProductStock(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.StockQuant, error) {
	return s.quantRepo.FindByProduct(ctx, organizationID, productID)
}

func (s *InventoryService) GetLocationStock(ctx context.Context, organizationID, locationID uuid.UUID) ([]domain.StockQuant, error) {
	return s.quantRepo.FindByLocation(ctx, organizationID, locationID)
}

func (s *InventoryService) GetAvailableQuantity(ctx context.Context, organizationID, productID, locationID uuid.UUID) (float64, error) {
	return s.quantRepo.FindAvailable(ctx, organizationID, productID, locationID)
}

// Move operations
func (s *InventoryService) CreateMove(ctx context.Context, move domain.StockMove) (*domain.StockMove, error) {
	if move.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if move.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product_id is required")
	}
	if move.LocationID == uuid.Nil {
		return nil, fmt.Errorf("location_id is required")
	}
	if move.LocationDestID == uuid.Nil {
		return nil, fmt.Errorf("location_dest_id is required")
	}
	if move.ProductUOMQty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	if move.State == "" {
		move.State = "draft"
	}
	if move.ProcureMethod == "" {
		move.ProcureMethod = "make_to_stock"
	}
	if move.Priority == "" {
		move.Priority = "1"
	}
	if move.Sequence == 0 {
		move.Sequence = 10
	}

	return s.moveRepo.Create(ctx, move)
}

func (s *InventoryService) GetMove(ctx context.Context, id uuid.UUID) (*domain.StockMove, error) {
	move, err := s.moveRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if move == nil {
		return nil, fmt.Errorf("stock move not found")
	}
	return move, nil
}

func (s *InventoryService) ListMoves(ctx context.Context, organizationID uuid.UUID) ([]domain.StockMove, error) {
	return s.moveRepo.FindAll(ctx, organizationID)
}

func (s *InventoryService) ConfirmMove(ctx context.Context, id uuid.UUID) error {
	// Update move to confirmed state and update stock quantities
	move, err := s.moveRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if move == nil {
		return fmt.Errorf("stock move not found")
	}

	// Update state to done
	if err := s.moveRepo.UpdateState(ctx, id, "done"); err != nil {
		return err
	}

	// Update quantities
	if err := s.quantRepo.UpdateQuantity(ctx, move.OrganizationID, move.ProductID, move.LocationID, -move.ProductUOMQty); err != nil {
		return fmt.Errorf("failed to decrease source quantity: %w", err)
	}
	if err := s.quantRepo.UpdateQuantity(ctx, move.OrganizationID, move.ProductID, move.LocationDestID, move.ProductUOMQty); err != nil {
		return fmt.Errorf("failed to increase dest quantity: %w", err)
	}

	return nil
}
