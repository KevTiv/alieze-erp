package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type batchOperationRepository struct {
	db *sql.DB
}

func NewBatchOperationRepository(db *sql.DB) BatchOperationRepository {
	return &batchOperationRepository{db: db}
}

func (r *batchOperationRepository) Create(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error) {
	query := `
		INSERT INTO batch_operations
		(id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
	`

	if operation.ID == uuid.Nil {
		operation.ID = uuid.New()
	}
	now := time.Now()
	if operation.CreatedAt.IsZero() {
		operation.CreatedAt = now
	}
	if operation.UpdatedAt.IsZero() {
		operation.UpdatedAt = now
	}

	var created types.BatchOperation
	err := r.db.QueryRowContext(ctx, query,
		operation.ID, operation.OrganizationID, operation.CompanyID, operation.OperationType, operation.Status,
		operation.Reference, operation.Description, operation.Priority, operation.SourceType, operation.SourceID,
		operation.CreatedBy, operation.TotalItems, operation.SuccessfulItems, operation.FailedItems,
		operation.ErrorMessage, operation.ErrorDetails, operation.Metadata, operation.CreatedAt, operation.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.OperationType, &created.Status,
		&created.Reference, &created.Description, &created.Priority, &created.SourceType, &created.SourceID,
		&created.CreatedBy, &created.ProcessedBy, &created.ProcessedAt, &created.TotalItems, &created.SuccessfulItems,
		&created.FailedItems, &created.ErrorMessage, &created.ErrorDetails, &created.Metadata, &created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}
	return &created, nil
}

func (r *batchOperationRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.BatchOperation, error) {
	query := `
		SELECT id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
		FROM batch_operations WHERE id = $1 AND deleted_at IS NULL
	`

	var operation types.BatchOperation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&operation.ID, &operation.OrganizationID, &operation.CompanyID, &operation.OperationType, &operation.Status,
		&operation.Reference, &operation.Description, &operation.Priority, &operation.SourceType, &operation.SourceID,
		&operation.CreatedBy, &operation.ProcessedBy, &operation.ProcessedAt, &operation.TotalItems, &operation.SuccessfulItems,
		&operation.FailedItems, &operation.ErrorMessage, &operation.ErrorDetails, &operation.Metadata, &operation.CreatedAt,
		&operation.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operation: %w", err)
	}
	return &operation, nil
}

func (r *batchOperationRepository) FindAll(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.BatchOperation, error) {
	query := `
		SELECT id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
		FROM batch_operations WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operations: %w", err)
	}
	defer rows.Close()

	var operations []types.BatchOperation
	for rows.Next() {
		var operation types.BatchOperation
		err := rows.Scan(
			&operation.ID, &operation.OrganizationID, &operation.CompanyID, &operation.OperationType, &operation.Status,
			&operation.Reference, &operation.Description, &operation.Priority, &operation.SourceType, &operation.SourceID,
			&operation.CreatedBy, &operation.ProcessedBy, &operation.ProcessedAt, &operation.TotalItems, &operation.SuccessfulItems,
			&operation.FailedItems, &operation.ErrorMessage, &operation.ErrorDetails, &operation.Metadata, &operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation: %w", err)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (r *batchOperationRepository) FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.BatchOperation, error) {
	query := `
		SELECT id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
		FROM batch_operations WHERE organization_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operations by status: %w", err)
	}
	defer rows.Close()

	var operations []types.BatchOperation
	for rows.Next() {
		var operation types.BatchOperation
		err := rows.Scan(
			&operation.ID, &operation.OrganizationID, &operation.CompanyID, &operation.OperationType, &operation.Status,
			&operation.Reference, &operation.Description, &operation.Priority, &operation.SourceType, &operation.SourceID,
			&operation.CreatedBy, &operation.ProcessedBy, &operation.ProcessedAt, &operation.TotalItems, &operation.SuccessfulItems,
			&operation.FailedItems, &operation.ErrorMessage, &operation.ErrorDetails, &operation.Metadata, &operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation: %w", err)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (r *batchOperationRepository) FindByType(ctx context.Context, organizationID uuid.UUID, operationType types.BatchOperationType) ([]types.BatchOperation, error) {
	query := `
		SELECT id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
		FROM batch_operations WHERE organization_id = $1 AND operation_type = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, operationType)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operations by type: %w", err)
	}
	defer rows.Close()

	var operations []types.BatchOperation
	for rows.Next() {
		var operation types.BatchOperation
		err := rows.Scan(
			&operation.ID, &operation.OrganizationID, &operation.CompanyID, &operation.OperationType, &operation.Status,
			&operation.Reference, &operation.Description, &operation.Priority, &operation.SourceType, &operation.SourceID,
			&operation.CreatedBy, &operation.ProcessedBy, &operation.ProcessedAt, &operation.TotalItems, &operation.SuccessfulItems,
			&operation.FailedItems, &operation.ErrorMessage, &operation.ErrorDetails, &operation.Metadata, &operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation: %w", err)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (r *batchOperationRepository) FindByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]types.BatchOperation, error) {
	query := `
		SELECT id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
		FROM batch_operations WHERE organization_id = $1 AND created_at BETWEEN $2 AND $3 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operations by date range: %w", err)
	}
	defer rows.Close()

	var operations []types.BatchOperation
	for rows.Next() {
		var operation types.BatchOperation
		err := rows.Scan(
			&operation.ID, &operation.OrganizationID, &operation.CompanyID, &operation.OperationType, &operation.Status,
			&operation.Reference, &operation.Description, &operation.Priority, &operation.SourceType, &operation.SourceID,
			&operation.CreatedBy, &operation.ProcessedBy, &operation.ProcessedAt, &operation.TotalItems, &operation.SuccessfulItems,
			&operation.FailedItems, &operation.ErrorMessage, &operation.ErrorDetails, &operation.Metadata, &operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation: %w", err)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (r *batchOperationRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.BatchOperation, error) {
	query := `
		SELECT DISTINCT bo.id, bo.organization_id, bo.company_id, bo.operation_type, bo.status, bo.reference, bo.description, bo.priority,
		 bo.source_type, bo.source_id, bo.created_by, bo.processed_by, bo.processed_at, bo.total_items, bo.successful_items, bo.failed_items,
		 bo.error_message, bo.error_details, bo.metadata, bo.created_at, bo.updated_at
		FROM batch_operations bo
		JOIN batch_operation_items boi ON bo.id = boi.batch_operation_id
		WHERE bo.organization_id = $1 AND boi.product_id = $2 AND bo.deleted_at IS NULL
		ORDER BY bo.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch operations by product: %w", err)
	}
	defer rows.Close()

	var operations []types.BatchOperation
	for rows.Next() {
		var operation types.BatchOperation
		err := rows.Scan(
			&operation.ID, &operation.OrganizationID, &operation.CompanyID, &operation.OperationType, &operation.Status,
			&operation.Reference, &operation.Description, &operation.Priority, &operation.SourceType, &operation.SourceID,
			&operation.CreatedBy, &operation.ProcessedBy, &operation.ProcessedAt, &operation.TotalItems, &operation.SuccessfulItems,
			&operation.FailedItems, &operation.ErrorMessage, &operation.ErrorDetails, &operation.Metadata, &operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation: %w", err)
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

func (r *batchOperationRepository) Update(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error) {
	query := `
		UPDATE batch_operations
		SET operation_type = $2, status = $3, reference = $4, description = $5, priority = $6,
		 source_type = $7, source_id = $8, created_by = $9, processed_by = $10, processed_at = $11,
		 total_items = $12, successful_items = $13, failed_items = $14, error_message = $15,
		 error_details = $16, metadata = $17, updated_at = $18
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, operation_type, status, reference, description, priority,
		 source_type, source_id, created_by, processed_by, processed_at, total_items, successful_items, failed_items,
		 error_message, error_details, metadata, created_at, updated_at
	`

	operation.UpdatedAt = time.Now()
	var updated types.BatchOperation
	err := r.db.QueryRowContext(ctx, query,
		operation.ID, operation.OperationType, operation.Status, operation.Reference, operation.Description, operation.Priority,
		operation.SourceType, operation.SourceID, operation.CreatedBy, operation.ProcessedBy, operation.ProcessedAt,
		operation.TotalItems, operation.SuccessfulItems, operation.FailedItems, operation.ErrorMessage,
		operation.ErrorDetails, operation.Metadata, operation.UpdatedAt,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.OperationType, &updated.Status,
		&updated.Reference, &updated.Description, &updated.Priority, &updated.SourceType, &updated.SourceID,
		&updated.CreatedBy, &updated.ProcessedBy, &updated.ProcessedAt, &updated.TotalItems, &updated.SuccessfulItems,
		&updated.FailedItems, &updated.ErrorMessage, &updated.ErrorDetails, &updated.Metadata, &updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("batch operation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update batch operation: %w", err)
	}
	return &updated, nil
}

func (r *batchOperationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE batch_operations SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete batch operation: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("batch operation not found")
	}
	return nil
}

func (r *batchOperationRepository) ProcessBatchOperation(ctx context.Context, operationID uuid.UUID, processedBy uuid.UUID) (*types.BatchOperationResult, error) {
	query := `
		SELECT
			batch_operation_id, status, total_items, successful_items, failed_items, error_message
		FROM process_batch_operation($1, $2)
	`

	rows, err := r.db.QueryContext(ctx, query, operationID, processedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to process batch operation: %w", err)
	}
	defer rows.Close()

	var result types.BatchOperationResult
	var itemResults []types.BatchOperationItemResult

	for rows.Next() {
		var batchOpID uuid.UUID
		var statusStr, errorMsg sql.NullString
		var totalItems, successfulItems, failedItems sql.NullInt32

		err := rows.Scan(
			&batchOpID, &statusStr, &totalItems, &successfulItems, &failedItems, &errorMsg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation result: %w", err)
		}

		if statusStr.Valid {
			result.Status = types.BatchOperationStatus(statusStr.String)
		}
		if totalItems.Valid {
			result.TotalItems = int(totalItems.Int32)
		}
		if successfulItems.Valid {
			result.SuccessfulItems = int(successfulItems.Int32)
		}
		if failedItems.Valid {
			result.FailedItems = int(failedItems.Int32)
		}
		if errorMsg.Valid {
			result.ErrorMessage = &errorMsg.String
		}
	}

	// Get item-level results
	itemQuery := `
		SELECT
			item_id, product_id, status, error_message, before_value, after_value, result_id, result_type
		FROM batch_operation_item_results
		WHERE batch_operation_id = $1
		ORDER BY created_at
	`

	itemRows, err := r.db.QueryContext(ctx, itemQuery, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch operation item results: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var itemResult types.BatchOperationItemResult
		var itemID, productID uuid.UUID
		var statusStr, errorMsg, beforeValue, afterValue sql.NullString
		var resultID sql.NullString
		var resultType sql.NullString

		err := itemRows.Scan(
			&itemID, &productID, &statusStr, &errorMsg, &beforeValue, &afterValue, &resultID, &resultType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch operation item result: %w", err)
		}

		itemResult.ItemID = itemID
		itemResult.ProductID = productID

		if statusStr.Valid {
			itemResult.Status = statusStr.String
		}
		if errorMsg.Valid {
			itemResult.ErrorMessage = &errorMsg.String
		}
		if beforeValue.Valid {
			itemResult.BeforeValue = beforeValue.String
		}
		if afterValue.Valid {
			itemResult.AfterValue = afterValue.String
		}
		if resultID.Valid {
			resultIDUUID := uuid.Must(uuid.Parse(resultID.String))
			itemResult.ResultID = &resultIDUUID
		}
		if resultType.Valid {
			itemResult.ResultType = &resultType.String
		}

		itemResults = append(itemResults, itemResult)
	}

	result.ItemResults = itemResults

	// Set timestamps
	result.StartedAt = time.Now().Add(-15 * time.Minute) // Approximate
	result.CompletedAt = time.Now()
	result.ProcessingTime = result.CompletedAt.Sub(result.StartedAt)

	return &result, nil
}

func (r *batchOperationRepository) GetStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, operationType *types.BatchOperationType) (types.BatchOperationStatistics, error) {
	query := `
		SELECT
			get_batch_operation_statistics($1, $2, $3, $4)
	`

	var stats types.BatchOperationStatistics

	var fromTimeParam, toTimeParam, opTypeParam interface{}
	if fromTime != nil {
		fromTimeParam = *fromTime
	}
	if toTime != nil {
		toTimeParam = *toTime
	}
	if operationType != nil {
		opTypeParam = *operationType
	}

	err := r.db.QueryRowContext(ctx, query, organizationID, fromTimeParam, toTimeParam, opTypeParam).Scan(
		&stats.TotalOperations, &stats.CompletedOperations, &stats.FailedOperations,
		&stats.PendingOperations, &stats.ProcessingOperations, &stats.SuccessRate,
		&stats.FailureRate, &stats.AverageItems, &stats.AverageProcessing,
		&stats.Last30Days, &stats.Last7Days, &stats.Today,
	)

	if err != nil {
		return types.BatchOperationStatistics{}, fmt.Errorf("failed to get batch operation statistics: %w", err)
	}

	return stats, nil
}
