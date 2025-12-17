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

// StockPackageRepository handles database operations for stock packages
type StockPackageRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewStockPackageRepository creates a new StockPackageRepository
func NewStockPackageRepository(db *sqlx.DB, logger *slog.Logger) *StockPackageRepository {
	return &StockPackageRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new stock package
func (r *StockPackageRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockPackageCreateRequest) (*types.StockPackage, error) {
	query := `
		INSERT INTO stock_packages (organization_id, name, location_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, organization_id, name, location_id, created_at
	`

	var packageModel types.StockPackage
	err := r.db.QueryRowxContext(ctx, query, orgID, req.Name, req.LocationID).StructScan(&packageModel)
	if err != nil {
		r.logger.Error("Failed to create stock package", "error", err)
		return nil, err
	}

	return &packageModel, nil
}

// GetByID retrieves a stock package by ID
func (r *StockPackageRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockPackage, error) {
	query := `
		SELECT id, organization_id, name, location_id, created_at
		FROM stock_packages
		WHERE id = $1 AND deleted_at IS NULL
	`

	var packageModel types.StockPackage
	err := r.db.GetContext(ctx, &packageModel, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get stock package", "error", err, "id", id)
		return nil, err
	}

	return &packageModel, nil
}

// List retrieves all stock packages for an organization
func (r *StockPackageRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockPackage, error) {
	query := `
		SELECT id, organization_id, name, location_id, created_at
		FROM stock_packages
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var packages []types.StockPackage
	err := r.db.SelectContext(ctx, &packages, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list stock packages", "error", err)
		return nil, err
	}

	return packages, nil
}

// Update updates a stock package
func (r *StockPackageRepository) Update(ctx context.Context, id uuid.UUID, req types.StockPackageUpdateRequest) (*types.StockPackage, error) {
	query := `UPDATE stock_packages SET `
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		query += `name = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Name)
		argCount++
	}

	if req.LocationID != nil {
		query += `location_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.LocationID)
		argCount++
	}

	if len(args) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += ` WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, name, location_id, created_at`
	args = append(args, id)

	var packageModel types.StockPackage
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&packageModel)
	if err != nil {
		r.logger.Error("Failed to update stock package", "error", err)
		return nil, err
	}

	return &packageModel, nil
}

// Delete deletes a stock package (soft delete)
func (r *StockPackageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE stock_packages
		SET deleted_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete stock package", "error", err)
		return err
	}

	return nil
}
