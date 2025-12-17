package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// StockPickingTypeRepository handles database operations for stock picking types
type StockPickingTypeRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewStockPickingTypeRepository creates a new StockPickingTypeRepository
func NewStockPickingTypeRepository(db *sqlx.DB, logger *slog.Logger) *StockPickingTypeRepository {
	return &StockPickingTypeRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new stock picking type
func (r *StockPickingTypeRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockPickingTypeCreateRequest) (*types.StockPickingType, error) {
	query := `
		INSERT INTO stock_picking_types (organization_id, name, code, sequence, sequence_id, default_location_src_id, default_location_dest_id, warehouse_id, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, organization_id, name, code, sequence, sequence_id, default_location_src_id, default_location_dest_id, warehouse_id, color, created_at, updated_at
	`

	var pickingType types.StockPickingType
	err := r.db.QueryRowxContext(ctx, query, orgID, req.Name, req.Code, req.Sequence, req.SequenceID, req.DefaultLocationSrcID, req.DefaultLocationDestID, req.WarehouseID, req.Color).StructScan(&pickingType)
	if err != nil {
		r.logger.Error("Failed to create stock picking type", "error", err)
		return nil, err
	}

	return &pickingType, nil
}

// GetByID retrieves a stock picking type by ID
func (r *StockPickingTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockPickingType, error) {
	query := `
		SELECT id, organization_id, name, code, sequence, sequence_id, default_location_src_id, default_location_dest_id, warehouse_id, color, created_at, updated_at
		FROM stock_picking_types
		WHERE id = $1
	`

	var pickingType types.StockPickingType
	err := r.db.GetContext(ctx, &pickingType, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get stock picking type", "error", err, "id", id)
		return nil, err
	}

	return &pickingType, nil
}

// List retrieves all stock picking types for an organization
func (r *StockPickingTypeRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockPickingType, error) {
	query := `
		SELECT id, organization_id, name, code, sequence, sequence_id, default_location_src_id, default_location_dest_id, warehouse_id, color, created_at, updated_at
		FROM stock_picking_types
		WHERE organization_id = $1
		ORDER BY sequence, created_at DESC
	`

	var pickingTypes []types.StockPickingType
	err := r.db.SelectContext(ctx, &pickingTypes, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list stock picking types", "error", err)
		return nil, err
	}

	return pickingTypes, nil
}

// Update updates a stock picking type
func (r *StockPickingTypeRepository) Update(ctx context.Context, id uuid.UUID, req types.StockPickingTypeUpdateRequest) (*types.StockPickingType, error) {
	query := `UPDATE stock_picking_types SET `
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		query += `name = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Name)
		argCount++
	}

	if req.Code != nil {
		query += `code = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Code)
		argCount++
	}

	if req.Sequence != nil {
		query += `sequence = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Sequence)
		argCount++
	}

	if req.SequenceID != nil {
		query += `sequence_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.SequenceID)
		argCount++
	}

	if req.DefaultLocationSrcID != nil {
		query += `default_location_src_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.DefaultLocationSrcID)
		argCount++
	}

	if req.DefaultLocationDestID != nil {
		query += `default_location_dest_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.DefaultLocationDestID)
		argCount++
	}

	if req.WarehouseID != nil {
		query += `warehouse_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.WarehouseID)
		argCount++
	}

	if req.Color != nil {
		query += `color = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Color)
		argCount++
	}

	if len(args) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += `, updated_at = NOW() WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, name, code, sequence, sequence_id, default_location_src_id, default_location_dest_id, warehouse_id, color, created_at, updated_at`
	args = append(args, id)

	var pickingType types.StockPickingType
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&pickingType)
	if err != nil {
		r.logger.Error("Failed to update stock picking type", "error", err)
		return nil, err
	}

	return &pickingType, nil
}

// Delete deletes a stock picking type
func (r *StockPickingTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM stock_picking_types
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete stock picking type", "error", err)
		return err
	}

	return nil
}
