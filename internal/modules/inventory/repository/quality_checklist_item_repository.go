package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type qualityChecklistItemRepository struct {
	db *sql.DB
}

func NewQualityChecklistItemRepository(db *sql.DB) QualityChecklistItemRepository {
	return &qualityChecklistItemRepository{db: db}
}

func (r *qualityChecklistItemRepository) Create(ctx context.Context, item types.QualityChecklistItem) (*types.QualityChecklistItem, error) {
	query := `
		INSERT INTO quality_checklist_items
		(id, checklist_id, description, criteria, sequence, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, checklist_id, description, criteria, sequence, active, created_at
	`

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.Sequence == 0 {
		item.Sequence = 10
	}
	if item.Active == false {
		item.Active = true
	}

	var created types.QualityChecklistItem
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.ChecklistID, item.Description, item.Criteria, item.Sequence, item.Active, item.CreatedAt,
	).Scan(
		&created.ID, &created.ChecklistID, &created.Description, &created.Criteria, &created.Sequence, &created.Active, &created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality checklist item: %w", err)
	}

	return &created, nil
}

func (r *qualityChecklistItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.QualityChecklistItem, error) {
	query := `
		SELECT id, checklist_id, description, criteria, sequence, active, created_at
		FROM quality_checklist_items WHERE id = $1
	`

	var item types.QualityChecklistItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.ChecklistID, &item.Description, &item.Criteria, &item.Sequence, &item.Active, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find quality checklist item: %w", err)
	}

	return &item, nil
}

func (r *qualityChecklistItemRepository) FindByChecklist(ctx context.Context, checklistID uuid.UUID) ([]types.QualityChecklistItem, error) {
	query := `
		SELECT id, checklist_id, description, criteria, sequence, active, created_at
		FROM quality_checklist_items WHERE checklist_id = $1
		ORDER BY sequence ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, checklistID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality checklist items: %w", err)
	}
	defer rows.Close()

	var items []types.QualityChecklistItem
	for rows.Next() {
		var item types.QualityChecklistItem
		err := rows.Scan(
			&item.ID, &item.ChecklistID, &item.Description, &item.Criteria, &item.Sequence, &item.Active, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality checklist item: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *qualityChecklistItemRepository) FindActiveByChecklist(ctx context.Context, checklistID uuid.UUID) ([]types.QualityChecklistItem, error) {
	query := `
		SELECT id, checklist_id, description, criteria, sequence, active, created_at
		FROM quality_checklist_items WHERE checklist_id = $1 AND active = true
		ORDER BY sequence ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, checklistID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active quality checklist items: %w", err)
	}
	defer rows.Close()

	var items []types.QualityChecklistItem
	for rows.Next() {
		var item types.QualityChecklistItem
		err := rows.Scan(
			&item.ID, &item.ChecklistID, &item.Description, &item.Criteria, &item.Sequence, &item.Active, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality checklist item: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *qualityChecklistItemRepository) Update(ctx context.Context, item types.QualityChecklistItem) (*types.QualityChecklistItem, error) {
	query := `
		UPDATE quality_checklist_items
		SET description = $2, criteria = $3, sequence = $4, active = $5
		WHERE id = $1
		RETURNING id, checklist_id, description, criteria, sequence, active, created_at
	`

	var updated types.QualityChecklistItem
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.Description, item.Criteria, item.Sequence, item.Active,
	).Scan(
		&updated.ID, &updated.ChecklistID, &updated.Description, &updated.Criteria, &updated.Sequence, &updated.Active, &updated.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quality checklist item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update quality checklist item: %w", err)
	}

	return &updated, nil
}

func (r *qualityChecklistItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM quality_checklist_items WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quality checklist item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality checklist item not found")
	}
	return nil
}

func (r *qualityChecklistItemRepository) DeleteByChecklist(ctx context.Context, checklistID uuid.UUID) error {
	query := `DELETE FROM quality_checklist_items WHERE checklist_id = $1`
	_, err := r.db.ExecContext(ctx, query, checklistID)
	if err != nil {
		return fmt.Errorf("failed to delete quality checklist items by checklist: %w", err)
	}
	return nil
}
