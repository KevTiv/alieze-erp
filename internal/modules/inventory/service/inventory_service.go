package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type InventoryService struct {
	db            *sql.DB
	logger        *slog.Logger
	warehouseRepo repository.WarehouseRepository
	locationRepo  repository.StockLocationRepository
	quantRepo     repository.StockQuantRepository
	moveRepo      repository.StockMoveRepository
}

func NewInventoryService(
	db *sql.DB,
	logger *slog.Logger,
	warehouseRepo repository.WarehouseRepository,
	locationRepo repository.StockLocationRepository,
	quantRepo repository.StockQuantRepository,
	moveRepo repository.StockMoveRepository,
) *InventoryService {
	return &InventoryService{
		db:            db,
		logger:        logger,
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		quantRepo:     quantRepo,
		moveRepo:      moveRepo,
	}
}

// Warehouse operations
func (s *InventoryService) CreateWarehouse(ctx context.Context, wh types.Warehouse) (*types.Warehouse, error) {
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

func (s *InventoryService) GetWarehouse(ctx context.Context, id uuid.UUID) (*types.Warehouse, error) {
	wh, err := s.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wh == nil {
		return nil, fmt.Errorf("warehouse not found")
	}
	return wh, nil
}

func (s *InventoryService) ListWarehouses(ctx context.Context, organizationID uuid.UUID) ([]types.Warehouse, error) {
	return s.warehouseRepo.FindAll(ctx, organizationID)
}

func (s *InventoryService) UpdateWarehouse(ctx context.Context, wh types.Warehouse) (*types.Warehouse, error) {
	return s.warehouseRepo.Update(ctx, wh)
}

func (s *InventoryService) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	return s.warehouseRepo.Delete(ctx, id)
}

// Location operations
func (s *InventoryService) CreateLocation(ctx context.Context, loc types.StockLocation) (*types.StockLocation, error) {
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

func (s *InventoryService) GetLocation(ctx context.Context, id uuid.UUID) (*types.StockLocation, error) {
	loc, err := s.locationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, fmt.Errorf("location not found")
	}
	return loc, nil
}

func (s *InventoryService) ListLocations(ctx context.Context, organizationID uuid.UUID) ([]types.StockLocation, error) {
	return s.locationRepo.FindAll(ctx, organizationID)
}

func (s *InventoryService) UpdateLocation(ctx context.Context, loc types.StockLocation) (*types.StockLocation, error) {
	return s.locationRepo.Update(ctx, loc)
}

func (s *InventoryService) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	return s.locationRepo.Delete(ctx, id)
}

// Stock operations
func (s *InventoryService) GetProductStock(ctx context.Context, organizationID, productID uuid.UUID) ([]types.StockQuant, error) {
	return s.quantRepo.FindByProduct(ctx, organizationID, productID)
}

func (s *InventoryService) GetLocationStock(ctx context.Context, organizationID, locationID uuid.UUID) ([]types.StockQuant, error) {
	return s.quantRepo.FindByLocation(ctx, organizationID, locationID)
}

func (s *InventoryService) GetAvailableQuantity(ctx context.Context, organizationID, productID, locationID uuid.UUID) (float64, error) {
	return s.quantRepo.FindAvailable(ctx, organizationID, productID, locationID)
}

// Move operations
func (s *InventoryService) CreateMove(ctx context.Context, organizationID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	// Input sanitization
	sanitizedReq := req
	if req.Name != "" {
		sanitizedReq.Name = sanitizeInput(req.Name)
	}
	if req.Note != nil && *req.Note != "" {
		sanitizedNote := sanitizeInput(*req.Note)
		sanitizedReq.Note = &sanitizedNote
	}

	s.logger.Info("Creating stock move",
		"organization_id", organizationID,
		"product_id", req.ProductID,
		"quantity", req.Quantity,
		"from_location", req.LocationID,
		"to_location", req.LocationDestID,
		"state", req.State,
	)

	if organizationID == uuid.Nil {
		s.logger.Error("Validation failed: organization_id is required")
		return nil, fmt.Errorf("organization_id is required")
	}
	if req.ProductID == uuid.Nil {
		s.logger.Error("Validation failed: product_id is required")
		return nil, fmt.Errorf("product_id is required")
	}
	if req.LocationID == uuid.Nil {
		s.logger.Error("Validation failed: location_id is required")
		return nil, fmt.Errorf("location_id is required")
	}
	if req.LocationDestID == uuid.Nil {
		s.logger.Error("Validation failed: location_dest_id is required")
		return nil, fmt.Errorf("location_dest_id is required")
	}
	if req.Quantity <= 0 {
		s.logger.Error("Validation failed: quantity must be positive")
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Set defaults
	if req.State == "" {
		sanitizedReq.State = "draft"
	}
	if req.Sequence == 0 {
		sanitizedReq.Sequence = 10
	}
	if req.Priority == "" {
		sanitizedReq.Priority = "1"
	}

	move, err := s.moveRepo.Create(ctx, organizationID, sanitizedReq)
	if err != nil {
		s.logger.Error("Failed to create stock move", "error", err)
		return nil, fmt.Errorf("failed to create stock move: %w", err)
	}

	// Audit log
	logAuditEvent(ctx, s.logger, "inventory.stock_move.create", move.ID, "Stock move created",
		map[string]interface{}{
			"organization_id": organizationID,
			"product_id":      move.ProductID,
			"quantity":        move.Quantity,
			"from_location":   move.LocationID,
			"to_location":     move.LocationDestID,
		})

	s.logger.Info("Stock move created successfully", "move_id", move.ID)
	return move, nil
}

// sanitizeInput sanitizes string input to prevent XSS and SQL injection
func sanitizeInput(input string) string {
	// Basic sanitization - remove potentially harmful characters
	// In a production environment, you would use a proper HTML sanitizer
	sanitized := input
	sanitized = strings.ReplaceAll(sanitized, "<", "&")
	sanitized = strings.ReplaceAll(sanitized, ">", ">")
	sanitized = strings.ReplaceAll(sanitized, "'", "'")
	sanitized = strings.ReplaceAll(sanitized, "\"", "\"")
	sanitized = strings.ReplaceAll(sanitized, ";", ";")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// extractMoveIDs extracts IDs from a slice of stock moves for logging
func extractMoveIDs(moves []types.StockMove) []uuid.UUID {
	ids := make([]uuid.UUID, len(moves))
	for i, move := range moves {
		ids[i] = move.ID
	}
	return ids
}

// logAuditEvent logs an audit event for inventory operations
func logAuditEvent(ctx context.Context, logger *slog.Logger, eventType string, entityID uuid.UUID, message string, details map[string]interface{}) {
	// Extract user info from context if available
	userID := "system"
	if ctx != nil {
		if user, ok := ctx.Value("user_id").(uuid.UUID); ok && user != uuid.Nil {
			userID = user.String()
		}
	}

	// Create audit log entry
	auditDetails := map[string]interface{}{
		"event_type":  eventType,
		"entity_id":   entityID,
		"user_id":     userID,
		"message":     message,
		"timestamp":   time.Now().UTC(),
	}

	// Add additional details
	for k, v := range details {
		auditDetails[k] = v
	}

	logger.Info("AUDIT", "event_type", eventType, "entity_id", entityID, "user_id", userID, "message", message, "timestamp", time.Now().UTC())
}

func (s *InventoryService) GetMove(ctx context.Context, id uuid.UUID) (*types.StockMove, error) {
	move, err := s.moveRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if move == nil {
		return nil, fmt.Errorf("stock move not found")
	}
	return move, nil
}

func (s *InventoryService) ListMoves(ctx context.Context, organizationID uuid.UUID) ([]types.StockMove, error) {
	return s.moveRepo.List(ctx, organizationID)
}

func (s *InventoryService) ConfirmMove(ctx context.Context, id uuid.UUID) error {
	// Update move to confirmed state and update stock quantities
	move, err := s.moveRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if move == nil {
		return fmt.Errorf("stock move not found")
	}

	// Update state to done
	if err := s.moveRepo.UpdateState(ctx, id, "done"); err != nil {
		s.logger.Error("Failed to update stock move state", "error", err, "move_id", id, "state", "done")
		return fmt.Errorf("failed to update stock move state: %w", err)
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

// BulkCreateMoves creates multiple stock moves in a single operation
func (s *InventoryService) BulkCreateMoves(ctx context.Context, organizationID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error) {
	if len(reqs) == 0 {
		return []types.StockMove{}, nil
	}

	if organizationID == uuid.Nil {
		s.logger.Error("Validation failed in bulk create: organization_id is required")
		return nil, fmt.Errorf("organization_id is required")
	}

	s.logger.Info("Bulk creating stock moves", "count", len(reqs), "organization_id", organizationID)

	// Validate all requests
	for i, req := range reqs {
		if req.ProductID == uuid.Nil {
			s.logger.Error("Validation failed in bulk create", "index", i, "error", "product_id is required")
			return nil, fmt.Errorf("request %d: product_id is required", i)
		}
		if req.LocationID == uuid.Nil {
			s.logger.Error("Validation failed in bulk create", "index", i, "error", "location_id is required")
			return nil, fmt.Errorf("request %d: location_id is required", i)
		}
		if req.LocationDestID == uuid.Nil {
			s.logger.Error("Validation failed in bulk create", "index", i, "error", "location_dest_id is required")
			return nil, fmt.Errorf("request %d: location_dest_id is required", i)
		}
		if req.Quantity <= 0 {
			s.logger.Error("Validation failed in bulk create", "index", i, "error", "quantity must be positive")
			return nil, fmt.Errorf("request %d: quantity must be positive", i)
		}

		// Set defaults
		if req.State == "" {
			reqs[i].State = "draft"
		}
		if req.Sequence == 0 {
			reqs[i].Sequence = 10
		}
		if req.Priority == "" {
			reqs[i].Priority = "1"
		}
	}

	moves, err := s.moveRepo.BulkCreate(ctx, organizationID, reqs)
	if err != nil {
		s.logger.Error("Failed to bulk create stock moves", "error", err, "count", len(reqs))
		return nil, fmt.Errorf("failed to bulk create stock moves: %w", err)
	}

	// Audit log for bulk operation
	logAuditEvent(ctx, s.logger, "inventory.stock_move.bulk_create", uuid.Nil, "Bulk stock moves created",
		map[string]interface{}{
			"organization_id": organizationID,
			"count":           len(moves),
			"move_ids":        extractMoveIDs(moves),
		})

	s.logger.Info("Bulk stock moves created successfully", "count", len(moves))
	return moves, nil
}

// ProcessStockMoveWithTransaction processes a stock move with transaction support
func (s *InventoryService) ProcessStockMoveWithTransaction(ctx context.Context, organizationID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	if organizationID == uuid.Nil {
		s.logger.Error("Validation failed: organization_id is required")
		return nil, fmt.Errorf("organization_id is required")
	}

	s.logger.Info("Processing stock move with transaction",
		"organization_id", organizationID,
		"product_id", req.ProductID,
		"quantity", req.Quantity,
		"from_location", req.LocationID,
		"to_location", req.LocationDestID,
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create stock move
	move, err := s.moveRepo.CreateWithTx(ctx, tx, organizationID, req)
	if err != nil {
		s.logger.Error("Failed to create stock move in transaction", "error", err)
		return nil, fmt.Errorf("failed to create stock move: %w", err)
	}

	// Update stock quantities
	if err := s.quantRepo.UpdateQuantityWithTx(ctx, tx, organizationID, req.ProductID, req.LocationID, -req.Quantity); err != nil {
		s.logger.Error("Failed to decrease source quantity", "error", err)
		return nil, fmt.Errorf("failed to decrease source quantity: %w", err)
	}
	if err := s.quantRepo.UpdateQuantityWithTx(ctx, tx, organizationID, req.ProductID, req.LocationDestID, req.Quantity); err != nil {
		s.logger.Error("Failed to increase dest quantity", "error", err)
		return nil, fmt.Errorf("failed to increase dest quantity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Stock move processed successfully with transaction", "move_id", move.ID)
	return move, nil
}
