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

// StockRuleRepository handles database operations for stock rules
type StockRuleRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewStockRuleRepository creates a new StockRuleRepository
func NewStockRuleRepository(db *sqlx.DB, logger *slog.Logger) *StockRuleRepository {
	return &StockRuleRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new stock rule
func (r *StockRuleRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockRuleCreateRequest) (*types.StockRule, error) {
	query := `
		INSERT INTO stock_rules (organization_id, name, action, location_src_id, location_dest_id, sequence, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, organization_id, name, action, location_src_id, location_dest_id, sequence, active, created_at, updated_at
	`

	var rule types.StockRule
	err := r.db.QueryRowxContext(ctx, query, orgID, req.Name, req.Action, req.LocationSrcID, req.LocationDestID, req.Sequence, req.Active).StructScan(&rule)
	if err != nil {
		r.logger.Error("Failed to create stock rule", "error", err)
		return nil, err
	}

	return &rule, nil
}

// GetByID retrieves a stock rule by ID
func (r *StockRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockRule, error) {
	query := `
		SELECT id, organization_id, name, action, location_src_id, location_dest_id, sequence, active, created_at, updated_at
		FROM stock_rules
		WHERE id = $1
	`

	var rule types.StockRule
	err := r.db.GetContext(ctx, &rule, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get stock rule", "error", err, "id", id)
		return nil, err
	}

	return &rule, nil
}

// List retrieves all stock rules for an organization
func (r *StockRuleRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockRule, error) {
	query := `
		SELECT id, organization_id, name, action, location_src_id, location_dest_id, sequence, active, created_at, updated_at
		FROM stock_rules
		WHERE organization_id = $1
		ORDER BY sequence, created_at DESC
	`

	var rules []types.StockRule
	err := r.db.SelectContext(ctx, &rules, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list stock rules", "error", err)
		return nil, err
	}

	return rules, nil
}

// Update updates a stock rule
func (r *StockRuleRepository) Update(ctx context.Context, id uuid.UUID, req types.StockRuleUpdateRequest) (*types.StockRule, error) {
	query := `UPDATE stock_rules SET `
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		query += `name = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Name)
		argCount++
	}

	if req.Action != nil {
		query += `action = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Action)
		argCount++
	}

	if req.LocationSrcID != nil {
		query += `location_src_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.LocationSrcID)
		argCount++
	}

	if req.LocationDestID != nil {
		query += `location_dest_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.LocationDestID)
		argCount++
	}

	if req.Sequence != nil {
		query += `sequence = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Sequence)
		argCount++
	}

	if req.Active != nil {
		query += `active = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Active)
		argCount++
	}

	if len(args) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += `, updated_at = NOW() WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, name, action, location_src_id, location_dest_id, sequence, active, created_at, updated_at`
	args = append(args, id)

	var rule types.StockRule
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&rule)
	if err != nil {
		r.logger.Error("Failed to update stock rule", "error", err)
		return nil, err
	}

	return &rule, nil
}

// Delete deletes a stock rule
func (r *StockRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM stock_rules
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete stock rule", "error", err)
		return err
	}

	return nil
}
