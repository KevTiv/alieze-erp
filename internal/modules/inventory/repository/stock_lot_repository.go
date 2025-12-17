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

// StockLotRepository handles database operations for stock lots
type StockLotRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewStockLotRepository creates a new StockLotRepository
func NewStockLotRepository(db *sqlx.DB, logger *slog.Logger) *StockLotRepository {
	return &StockLotRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new stock lot
func (r *StockLotRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockLotCreateRequest) (*types.StockLot, error) {
	query := `
		INSERT INTO stock_lots (organization_id, name, ref, product_id, expiration_date, use_date, removal_date, alert_date, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, organization_id, company_id, name, ref, product_id, expiration_date, use_date, removal_date, alert_date, note, created_at, updated_at
	`

	var lot types.StockLot
	err := r.db.QueryRowxContext(ctx, query, orgID, req.Name, req.Ref, req.ProductID, req.ExpirationDate, req.UseDate, req.RemovalDate, req.AlertDate, req.Note).StructScan(&lot)
	if err != nil {
		r.logger.Error("Failed to create stock lot", "error", err)
		return nil, err
	}

	return &lot, nil
}

// GetByID retrieves a stock lot by ID
func (r *StockLotRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockLot, error) {
	query := `
		SELECT id, organization_id, company_id, name, ref, product_id, expiration_date, use_date, removal_date, alert_date, note, created_at, updated_at
		FROM stock_lots
		WHERE id = $1
	`

	var lot types.StockLot
	err := r.db.GetContext(ctx, &lot, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get stock lot", "error", err, "id", id)
		return nil, err
	}

	return &lot, nil
}

// List retrieves all stock lots for an organization
func (r *StockLotRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockLot, error) {
	query := `
		SELECT id, organization_id, company_id, name, ref, product_id, expiration_date, use_date, removal_date, alert_date, note, created_at, updated_at
		FROM stock_lots
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`

	var lots []types.StockLot
	err := r.db.SelectContext(ctx, &lots, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list stock lots", "error", err)
		return nil, err
	}

	return lots, nil
}

// Update updates a stock lot
func (r *StockLotRepository) Update(ctx context.Context, id uuid.UUID, req types.StockLotUpdateRequest) (*types.StockLot, error) {
	query := `UPDATE stock_lots SET `
	args := []interface{}{}
	argCount := 1

	if req.Ref != nil {
		query += `ref = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Ref)
		argCount++
	}

	if req.ExpirationDate != nil {
		query += `expiration_date = $` + string(rune(argCount)) + `, `
		args = append(args, *req.ExpirationDate)
		argCount++
	}

	if req.UseDate != nil {
		query += `use_date = $` + string(rune(argCount)) + `, `
		args = append(args, *req.UseDate)
		argCount++
	}

	if req.RemovalDate != nil {
		query += `removal_date = $` + string(rune(argCount)) + `, `
		args = append(args, *req.RemovalDate)
		argCount++
	}

	if req.AlertDate != nil {
		query += `alert_date = $` + string(rune(argCount)) + `, `
		args = append(args, *req.AlertDate)
		argCount++
	}

	if req.Note != nil {
		query += `note = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Note)
		argCount++
	}

	if len(args) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += `, updated_at = NOW() WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, company_id, name, ref, product_id, expiration_date, use_date, removal_date, alert_date, note, created_at, updated_at`
	args = append(args, id)

	var lot types.StockLot
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&lot)
	if err != nil {
		r.logger.Error("Failed to update stock lot", "error", err)
		return nil, err
	}

	return &lot, nil
}

// Delete deletes a stock lot
func (r *StockLotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM stock_lots
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete stock lot", "error", err)
		return err
	}

	return nil
}
