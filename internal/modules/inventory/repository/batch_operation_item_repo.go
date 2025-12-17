package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type batchOperationItemRepository struct {
	db *sql.DB
}

func NewBatchOperationItemRepository(db *sql.DB) BatchOperationItemRepository {
	return &batchOperationItemRepository{db: db}
}

func (r *batchOperationItemRepository) Create(ctx context.Context, item domain.BatchOperationItem) (*domain.BatchOperationItem, error) {
	query := `
		INSERT INTO batch_operation_items
		(id, batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
		 source_location_id, dest_location_id, current_quantity, adjustment_quantity, new_quantity,
		 operation_data, status, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
		 source_location_id, dest_location_id, current_quantity, adjustment_quantity, new_quantity,
		 operation_data, status, error_message, created_at, updated_at
	`

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = now
	}

	var created domain.BatchOperationItem
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.BatchOperationID, item.Sequence, item.ProductID, item.ProductName, item.LotID, item.SerialNumber,
		item.SourceLocationID, item.DestLocationID, item.CurrentQuantity, item.AdjustmentQuantity, item.NewQuantity,
		item.OperationData, item.Status, item.ErrorMessage, item.CreatedAt, item.UpdatedAt,
	).Scan(
		&created.ID, &created.BatchOperationID, &created.Sequence, &created.ProductID, &created.ProductName, &created.LotID, &created.SerialNumber,
		&created.SourceLocationID, &created.DestLocationID, &created.CurrentQuantity, &created.AdjustmentQuantity, &created.NewQuantity,
		&created.OperationData, &created.Status, &created.ErrorMessage, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation item: %w", err)
	}
	return &created, nil
}

func (r *batchOperationItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.BatchOperationItem, error) {
	query := `
		SELECT id, batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
		 source_location_id, dest_location_id, current_quantity, adjustment_quantity, new_quantity,
		 operation_data, status, error_message, created_at, updated_at
		FROM batch_operation_items WHERE id = $1
	`

	var item domain.BatchOperationItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.BatchOperationID, &item.Sequence, &item.ProductID, &item.ProductName, &item.LotID, &item.SerialNumber,
		&item.SourceLocationID, &item.DestLocationID, &item.CurrentQuantity, &item.AdjustmentQuantity, &item.NewQuantity,
		&item.OperationData, &item.Status, &item.ErrorMessage, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operation item: %w", err)
	}
	return &item, nil
}

func (r *batchOperationItemRepository) FindByBatchOperation(ctx context.Context, batchOperationID uuid.UUID) ([]domain.BatchOperationItem, error) {
	query := `
		SELECT id, batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
		 source_location_id, dest_location_id, current_quantity, adjustment_quantity, new_quantity,
		 operation_data, status, error_message, created_at, updated_at
		FROM batch_operation_items WHERE batch_operation_id = $1
		ORDER BY sequence ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, batchOperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operation items: %w", err)
	}
	defer rows.Close()

	var items []domain.BatchOperationItem
	for rows.Next() {
		var item domain.BatchOperationItem
		err := rows.Scan(
			&item.ID, &item.BatchOperationID, &item.Sequence, &item.ProductID, &item.ProductName, &item.LotID, &item.SerialNumber,
			&item.SourceLocationID, &item.DestLocationID, &item.CurrentQuantity, &item.AdjustmentQuantity, &item.NewQuantity,
			&item.OperationData, &item.Status, &item.ErrorMessage, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *batchOperationItemRepository) FindByStatus(ctx context.Context, batchOperationID uuid.UUID, status string) ([]domain.BatchOperationItem, error) {
	query := `
		SELECT id, batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
		 source_location_id, dest_location_id, current_quantity, adjustment_quantity, new_quantity,
		 operation_data, status, error_message, created_at, updated_at
		FROM batch_operation_items WHERE batch_operation_id = $1 AND status = $2
		ORDER BY sequence ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, batchOperationID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operation items by status: %w", err)
	}
	defer rows.Close()

	var items []domain.BatchOperationItem
	for rows.Next() {
		var item domain.BatchOperationItem
		err := rows.Scan(
			&item.ID, &item.BatchOperationID, &item.Sequence, &item.ProductID, &item.ProductName, &item.LotID, &item.SerialNumber,
			&item.SourceLocationID, &item.DestLocationID, &item.CurrentQuantity, &item.AdjustmentQuantity, &item.NewQuantity,
			&item.OperationData, &item.Status, &item.ErrorMessage, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *batchOperationItemRepository) Update(ctx context.Context, item domain.BatchOperationItem) (*domain.BatchOperationItem, error) {
	query := `
		UPDATE batch_operation_items
		SET sequence = $2, product_id = $3, product_name = $4, lot_id = $5, serial_number = $6,
		 source_location_id = $7, dest_location_id = $8, current_quantity = $9, adjustment_quantity = $10,
		 new_quantity = $11, operation_data = $12, status = $13, error_message = $14, updated_at = $15
		WHERE id = $1
		RETURNING id, batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
		 source_location_id, dest_location_id, current_quantity, adjustment_quantity, new_quantity,
		 operation_data, status, error_message, created_at, updated_at
	`

	item.UpdatedAt = time.Now()
	var updated domain.BatchOperationItem
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.Sequence, item.ProductID, item.ProductName, item.LotID, item.SerialNumber,
		item.SourceLocationID, item.DestLocationID, item.CurrentQuantity, item.AdjustmentQuantity, item.NewQuantity,
		item.OperationData, item.Status, item.ErrorMessage, item.UpdatedAt,
	).Scan(
		&updated.ID, &updated.BatchOperationID, &updated.Sequence, &updated.ProductID, &updated.ProductName, &updated.LotID, &updated.SerialNumber,
		&updated.SourceLocationID, &updated.DestLocationID, &updated.CurrentQuantity, &updated.AdjustmentQuantity, &updated.NewQuantity,
		&updated.OperationData, &updated.Status, &updated.ErrorMessage, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("batch operation item not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update batch operation item: %w", err)
	}
	return &updated, nil
}

func (r *batchOperationItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM batch_operation_items WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete batch operation item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("batch operation item not found")
	}
	return nil
}

func (r *batchOperationItemRepository) DeleteByBatchOperation(ctx context.Context, batchOperationID uuid.UUID) error {
	query := `DELETE FROM batch_operation_items WHERE batch_operation_id = $1`
	_, err := r.db.ExecContext(ctx, query, batchOperationID)
	if err != nil {
		return fmt.Errorf("failed to delete batch operation items: %w", err)
	}
	return nil
}
