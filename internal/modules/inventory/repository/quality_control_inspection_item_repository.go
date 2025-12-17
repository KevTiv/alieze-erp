package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type qualityControlInspectionItemRepository struct {
	db *sql.DB
}

func NewQualityControlInspectionItemRepository(db *sql.DB) QualityControlInspectionItemRepository {
	return &qualityControlInspectionItemRepository{db: db}
}

func (r *qualityControlInspectionItemRepository) Create(ctx context.Context, item domain.QualityControlInspectionItem) (*domain.QualityControlInspectionItem, error) {
	query := `
		INSERT INTO quality_control_inspection_items
		(id, inspection_id, checklist_item_id, description, result, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, inspection_id, checklist_item_id, description, result, notes, created_at
	`

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.Result == "" {
		item.Result = "pending"
	}

	var created domain.QualityControlInspectionItem
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.InspectionID, item.ChecklistItemID, item.Description, item.Result, item.Notes, item.CreatedAt,
	).Scan(
		&created.ID, &created.InspectionID, &created.ChecklistItemID, &created.Description, &created.Result, &created.Notes, &created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality control inspection item: %w", err)
	}

	return &created, nil
}

func (r *qualityControlInspectionItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityControlInspectionItem, error) {
	query := `
		SELECT id, inspection_id, checklist_item_id, description, result, notes, created_at
		FROM quality_control_inspection_items WHERE id = $1
	`

	var item domain.QualityControlInspectionItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.InspectionID, &item.ChecklistItemID, &item.Description, &item.Result, &item.Notes, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspection item: %w", err)
	}

	return &item, nil
}

func (r *qualityControlInspectionItemRepository) FindByInspection(ctx context.Context, inspectionID uuid.UUID) ([]domain.QualityControlInspectionItem, error) {
	query := `
		SELECT id, inspection_id, checklist_item_id, description, result, notes, created_at
		FROM quality_control_inspection_items WHERE inspection_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspection items: %w", err)
	}
	defer rows.Close()

	var items []domain.QualityControlInspectionItem
	for rows.Next() {
		var item domain.QualityControlInspectionItem
		err := rows.Scan(
			&item.ID, &item.InspectionID, &item.ChecklistItemID, &item.Description, &item.Result, &item.Notes, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection item: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *qualityControlInspectionItemRepository) Update(ctx context.Context, item domain.QualityControlInspectionItem) (*domain.QualityControlInspectionItem, error) {
	query := `
		UPDATE quality_control_inspection_items
		SET description = $2, result = $3, notes = $4
		WHERE id = $1
		RETURNING id, inspection_id, checklist_item_id, description, result, notes, created_at
	`

	var updated domain.QualityControlInspectionItem
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.Description, item.Result, item.Notes,
	).Scan(
		&updated.ID, &updated.InspectionID, &updated.ChecklistItemID, &updated.Description, &updated.Result, &updated.Notes, &updated.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quality control inspection item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update quality control inspection item: %w", err)
	}

	return &updated, nil
}

func (r *qualityControlInspectionItemRepository) UpdateResult(ctx context.Context, itemID uuid.UUID, result, notes string) error {
	query := `UPDATE quality_control_inspection_items SET result = $2, notes = $3 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, itemID, result, notes)
	if err != nil {
		return fmt.Errorf("failed to update quality control inspection item result: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality control inspection item not found")
	}
	return nil
}

func (r *qualityControlInspectionItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM quality_control_inspection_items WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quality control inspection item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality control inspection item not found")
	}
	return nil
}

func (r *qualityControlInspectionItemRepository) DeleteByInspection(ctx context.Context, inspectionID uuid.UUID) error {
	query := `DELETE FROM quality_control_inspection_items WHERE inspection_id = $1`
	_, err := r.db.ExecContext(ctx, query, inspectionID)
	if err != nil {
		return fmt.Errorf("failed to delete quality control inspection items by inspection: %w", err)
	}
	return nil
}
