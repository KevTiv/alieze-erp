package repository

import (
	"context"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// BatchOperationRepository interface
type BatchOperationRepository interface {
	Create(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.BatchOperation, error)
	FindAll(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.BatchOperation, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.BatchOperation, error)
	FindByType(ctx context.Context, organizationID uuid.UUID, operationType types.BatchOperationType) ([]types.BatchOperation, error)
	FindByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]types.BatchOperation, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.BatchOperation, error)
	Update(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Business logic methods
	ProcessBatchOperation(ctx context.Context, operationID uuid.UUID, processedBy uuid.UUID) (*types.BatchOperationResult, error)
	GetStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, operationType *types.BatchOperationType) (types.BatchOperationStatistics, error)
}

// BatchOperationItemRepository interface
type BatchOperationItemRepository interface {
	Create(ctx context.Context, item types.BatchOperationItem) (*types.BatchOperationItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.BatchOperationItem, error)
	FindByBatchOperation(ctx context.Context, batchOperationID uuid.UUID) ([]types.BatchOperationItem, error)
	FindByStatus(ctx context.Context, batchOperationID uuid.UUID, status string) ([]types.BatchOperationItem, error)
	Update(ctx context.Context, item types.BatchOperationItem) (*types.BatchOperationItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByBatchOperation(ctx context.Context, batchOperationID uuid.UUID) error
}
