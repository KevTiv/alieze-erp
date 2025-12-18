package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// UOMCategoryRepository handles UOM category data operations
type UOMCategoryRepository struct {
	db *sql.DB
}

func NewUOMCategoryRepository(db *sql.DB) *UOMCategoryRepository {
	return &UOMCategoryRepository{db: db}
}

func (r *UOMCategoryRepository) Create(ctx context.Context, category types.UOMCategory) (*types.UOMCategory, error) {
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	if category.Name == "" {
		return nil, errors.New("name is required")
	}

	query := `
		INSERT INTO uom_categories (
			id, name, created_at
		) VALUES (
			$1, $2, $3
		) RETURNING id, name, created_at
	`

	now := time.Now()

	var created types.UOMCategory
	err := r.db.QueryRowContext(ctx, query,
		category.ID, category.Name, now,
	).Scan(
		&created.ID, &created.Name, &created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *UOMCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.UOMCategory, error) {
	query := `
		SELECT id, name, created_at
		FROM uom_categories
		WHERE id = $1
	`

	var category types.UOMCategory
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

func (r *UOMCategoryRepository) List(ctx context.Context, filter types.UOMCategoryFilter) ([]types.UOMCategory, error) {
	query := `
		SELECT id, name, created_at
		FROM uom_categories
		WHERE (1=1)
	`

	params := []interface{}{}
	paramIndex := 1

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

	var categories []types.UOMCategory
	for rows.Next() {
		var category types.UOMCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *UOMCategoryRepository) Update(ctx context.Context, id uuid.UUID, update types.UOMCategoryUpdateRequest) (*types.UOMCategory, error) {
	if update.Name == nil {
		return nil, errors.New("no fields to update")
	}

	query := "UPDATE uom_categories SET name = $1 WHERE id = $2 RETURNING id, name, created_at"

	var updated types.UOMCategory
	err := r.db.QueryRowContext(ctx, query, *update.Name, id).Scan(
		&updated.ID, &updated.Name, &updated.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *UOMCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM uom_categories WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
