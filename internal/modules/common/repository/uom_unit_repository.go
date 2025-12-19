package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// UOMUnitRepository handles UOM unit data operations
type UOMUnitRepository struct {
	db *sql.DB
}

func NewUOMUnitRepository(db *sql.DB) *UOMUnitRepository {
	return &UOMUnitRepository{db: db}
}

func (r *UOMUnitRepository) Create(ctx context.Context, unit types.UOMUnit) (*types.UOMUnit, error) {
	if unit.ID == uuid.Nil {
		unit.ID = uuid.New()
	}

	if unit.CategoryID == uuid.Nil {
		return nil, errors.New("category_id is required")
	}

	if unit.Name == "" {
		return nil, errors.New("name is required")
	}

	if unit.UOMType == "" {
		unit.UOMType = types.UOMTypeReference
	}

	if unit.Factor == 0 {
		unit.Factor = 1.0
	}

	if unit.FactorInv == 0 {
		unit.FactorInv = 1.0
	}

	if unit.Rounding == 0 {
		unit.Rounding = 0.01
	}

	if !unit.Active {
		unit.Active = true
	}

	query := `
		INSERT INTO uom_units (
			id, category_id, name, uom_type, factor, factor_inv, rounding, active, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, category_id, name, uom_type, factor, factor_inv, rounding, active, created_at
	`

	now := time.Now()

	var created types.UOMUnit
	err := r.db.QueryRowContext(ctx, query,
		unit.ID, unit.CategoryID, unit.Name, unit.UOMType, unit.Factor, unit.FactorInv, unit.Rounding, unit.Active, now,
	).Scan(
		&created.ID, &created.CategoryID, &created.Name, &created.UOMType, &created.Factor, &created.FactorInv, &created.Rounding, &created.Active, &created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *UOMUnitRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.UOMUnit, error) {
	query := `
		SELECT id, category_id, name, uom_type, factor, factor_inv, rounding, active, created_at
		FROM uom_units
		WHERE id = $1
	`

	var unit types.UOMUnit
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&unit.ID, &unit.CategoryID, &unit.Name, &unit.UOMType, &unit.Factor, &unit.FactorInv, &unit.Rounding, &unit.Active, &unit.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &unit, nil
}

func (r *UOMUnitRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]types.UOMUnit, error) {
	query := `
		SELECT id, category_id, name, uom_type, factor, factor_inv, rounding, active, created_at
		FROM uom_units
		WHERE category_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []types.UOMUnit
	for rows.Next() {
		var unit types.UOMUnit
		err := rows.Scan(
			&unit.ID, &unit.CategoryID, &unit.Name, &unit.UOMType, &unit.Factor, &unit.FactorInv, &unit.Rounding, &unit.Active, &unit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}

	return units, nil
}

func (r *UOMUnitRepository) List(ctx context.Context, filter types.UOMUnitFilter) ([]types.UOMUnit, error) {
	query := `
		SELECT id, category_id, name, uom_type, factor, factor_inv, rounding, active, created_at
		FROM uom_units
		WHERE (1=1)
	`

	params := []interface{}{}
	paramIndex := 1

	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", paramIndex)
		params = append(params, *filter.CategoryID)
		paramIndex++
	}

	if filter.Active != nil {
		query += fmt.Sprintf(" AND active = $%d", paramIndex)
		params = append(params, *filter.Active)
		paramIndex++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name ILIKE $%d", paramIndex)
		params = append(params, "%"+*filter.Name+"%")
		paramIndex++
	}

	query += " ORDER BY name"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		params = append(params, filter.Limit)
		paramIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramIndex)
		params = append(params, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []types.UOMUnit
	for rows.Next() {
		var unit types.UOMUnit
		err := rows.Scan(
			&unit.ID, &unit.CategoryID, &unit.Name, &unit.UOMType, &unit.Factor, &unit.FactorInv, &unit.Rounding, &unit.Active, &unit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}

	return units, nil
}

func (r *UOMUnitRepository) Update(ctx context.Context, id uuid.UUID, update types.UOMUnitUpdateRequest) (*types.UOMUnit, error) {
	if update.CategoryID == nil && update.Name == nil && update.UOMType == nil &&
	   update.Factor == nil && update.FactorInv == nil && update.Rounding == nil && update.Active == nil {
		return nil, errors.New("no fields to update")
	}

	query := "UPDATE uom_units SET "
	params := []interface{}{}
	paramIndex := 1

	if update.CategoryID != nil {
		query += fmt.Sprintf("category_id = $%d, ", paramIndex)
		params = append(params, *update.CategoryID)
		paramIndex++
	}

	if update.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, *update.Name)
		paramIndex++
	}

	if update.UOMType != nil {
		query += fmt.Sprintf("uom_type = $%d, ", paramIndex)
		params = append(params, *update.UOMType)
		paramIndex++
	}

	if update.Factor != nil {
		query += fmt.Sprintf("factor = $%d, ", paramIndex)
		params = append(params, *update.Factor)
		paramIndex++
	}

	if update.FactorInv != nil {
		query += fmt.Sprintf("factor_inv = $%d, ", paramIndex)
		params = append(params, *update.FactorInv)
		paramIndex++
	}

	if update.Rounding != nil {
		query += fmt.Sprintf("rounding = $%d, ", paramIndex)
		params = append(params, *update.Rounding)
		paramIndex++
	}

	if update.Active != nil {
		query += fmt.Sprintf("active = $%d, ", paramIndex)
		params = append(params, *update.Active)
		paramIndex++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, category_id, name, uom_type, factor, factor_inv, rounding, active, created_at", paramIndex)
	params = append(params, id)

	var updated types.UOMUnit
	err := r.db.QueryRowContext(ctx, query, params...).Scan(
		&updated.ID, &updated.CategoryID, &updated.Name, &updated.UOMType, &updated.Factor, &updated.FactorInv, &updated.Rounding, &updated.Active, &updated.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *UOMUnitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM uom_units WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
