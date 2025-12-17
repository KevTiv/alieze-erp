package repository

import (
	"context"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// BatchOperationRepository interface
type BatchOperationRepository interface {
	Create(ctx context.Context, operation domain.BatchOperation) (*domain.BatchOperation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.BatchOperation, error)
	FindAll(ctx context.Context, organizationID uuid.UUID, limit int) ([]domain.BatchOperation, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]domain.BatchOperation, error)
	FindByType(ctx context.Context, organizationID uuid.UUID, operationType domain.BatchOperationType) ([]domain.BatchOperation, error)
	FindByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]domain.BatchOperation, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.BatchOperation, error)
	Update(ctx context.Context, operation domain.BatchOperation) (*domain.BatchOperation, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Business logic methods
	ProcessBatchOperation(ctx context.Context, operationID uuid.UUID, processedBy uuid.UUID) (*domain.BatchOperationResult, error)
	GetStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, operationType *domain.BatchOperationType) (domain.BatchOperationStatistics, error)
}

// BatchOperationItemRepository interface
type BatchOperationItemRepository interface {
	Create(ctx context.Context, item domain.BatchOperationItem) (*domain.BatchOperationItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.BatchOperationItem, error)
	FindByBatchOperation(ctx context.Context, batchOperationID uuid.UUID) ([]domain.BatchOperationItem, error)
	FindByStatus(ctx context.Context, batchOperationID uuid.UUID, status string) ([]domain.BatchOperationItem, error)
	Update(ctx context.Context, item domain.BatchOperationItem) (*domain.BatchOperationItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByBatchOperation(ctx context.Context, batchOperationID uuid.UUID) error
}
