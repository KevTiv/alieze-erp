package service

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"
	productsRepo "alieze-erp/internal/modules/products/repository"

	"github.com/google/uuid"
)

// BatchOperationService interface
type BatchOperationService interface {
	// CRUD Operations
	CreateBatchOperation(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error)
	GetBatchOperation(ctx context.Context, id uuid.UUID) (*types.BatchOperation, error)
	ListBatchOperations(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.BatchOperation, error)
	ListBatchOperationsByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.BatchOperation, error)
	ListBatchOperationsByType(ctx context.Context, organizationID uuid.UUID, operationType types.BatchOperationType) ([]types.BatchOperation, error)
	ListBatchOperationsByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]types.BatchOperation, error)
	ListBatchOperationsByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.BatchOperation, error)
	UpdateBatchOperation(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error)
	DeleteBatchOperation(ctx context.Context, id uuid.UUID) error

	// Batch Operation Item Operations
	CreateBatchOperationItem(ctx context.Context, item types.BatchOperationItem) (*types.BatchOperationItem, error)
	GetBatchOperationItem(ctx context.Context, id uuid.UUID) (*types.BatchOperationItem, error)
	ListBatchOperationItems(ctx context.Context, batchOperationID uuid.UUID) ([]types.BatchOperationItem, error)
	ListBatchOperationItemsByStatus(ctx context.Context, batchOperationID uuid.UUID, status string) ([]types.BatchOperationItem, error)
	UpdateBatchOperationItem(ctx context.Context, item types.BatchOperationItem) (*types.BatchOperationItem, error)
	DeleteBatchOperationItem(ctx context.Context, id uuid.UUID) error

	// Business Logic Operations
	CreateStockAdjustmentBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, items []types.BatchOperationItem) (*types.BatchOperation, error)
	CreateStockTransferBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, sourceLocationID, destLocationID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error)
	CreateStockCountBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, locationID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error)
	CreatePriceUpdateBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, currencyID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error)
	CreateLocationUpdateBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, newLocationID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error)
	CreateStatusUpdateBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, newStatus string, items []types.BatchOperationItem) (*types.BatchOperation, error)

	ProcessBatchOperation(ctx context.Context, operationID uuid.UUID, processedBy uuid.UUID) (*types.BatchOperationResult, error)
	GetBatchOperationStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, operationType *types.BatchOperationType) (types.BatchOperationStatistics, error)

	// Validation and Utility Methods
	ValidateBatchOperation(ctx context.Context, operation types.BatchOperation) error
	ValidateBatchOperationItem(ctx context.Context, item types.BatchOperationItem) error
	CalculateBatchOperationTotals(ctx context.Context, batchOperationID uuid.UUID) (int, int, int, error)
}

type batchOperationService struct {
	batchOperationRepo repository.BatchOperationRepository
	batchOperationItemRepo repository.BatchOperationItemRepository
	inventoryService   *InventoryService
	productsRepo       productsRepo.ProductRepo
}

func NewBatchOperationService(
	batchOperationRepo repository.BatchOperationRepository,
	batchOperationItemRepo repository.BatchOperationItemRepository,
	inventoryService *InventoryService,
	productsRepo productsRepo.ProductRepo,
) BatchOperationService {
	return &batchOperationService{
		batchOperationRepo:     batchOperationRepo,
		batchOperationItemRepo: batchOperationItemRepo,
		inventoryService:       inventoryService,
		productsRepo:           productsRepo,
	}
}

// CRUD Operations

func (s *batchOperationService) CreateBatchOperation(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error) {
	if err := s.ValidateBatchOperation(ctx, operation); err != nil {
		return nil, fmt.Errorf("invalid batch operation: %w", err)
	}

	return s.batchOperationRepo.Create(ctx, operation)
}

func (s *batchOperationService) GetBatchOperation(ctx context.Context, id uuid.UUID) (*types.BatchOperation, error) {
	return s.batchOperationRepo.FindByID(ctx, id)
}

func (s *batchOperationService) ListBatchOperations(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.BatchOperation, error) {
	return s.batchOperationRepo.FindAll(ctx, organizationID, limit)
}

func (s *batchOperationService) ListBatchOperationsByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.BatchOperation, error) {
	return s.batchOperationRepo.FindByStatus(ctx, organizationID, status)
}

func (s *batchOperationService) ListBatchOperationsByType(ctx context.Context, organizationID uuid.UUID, operationType types.BatchOperationType) ([]types.BatchOperation, error) {
	return s.batchOperationRepo.FindByType(ctx, organizationID, operationType)
}

func (s *batchOperationService) ListBatchOperationsByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]types.BatchOperation, error) {
	return s.batchOperationRepo.FindByDateRange(ctx, organizationID, fromTime, toTime)
}

func (s *batchOperationService) ListBatchOperationsByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.BatchOperation, error) {
	return s.batchOperationRepo.FindByProduct(ctx, organizationID, productID)
}

func (s *batchOperationService) UpdateBatchOperation(ctx context.Context, operation types.BatchOperation) (*types.BatchOperation, error) {
	if err := s.ValidateBatchOperation(ctx, operation); err != nil {
		return nil, fmt.Errorf("invalid batch operation: %w", err)
	}

	return s.batchOperationRepo.Update(ctx, operation)
}

func (s *batchOperationService) DeleteBatchOperation(ctx context.Context, id uuid.UUID) error {
	// First delete all items associated with this batch operation
	err := s.batchOperationItemRepo.DeleteByBatchOperation(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete batch operation items: %w", err)
	}

	// Then delete the batch operation itself
	return s.batchOperationRepo.Delete(ctx, id)
}

// Batch Operation Item Operations

func (s *batchOperationService) CreateBatchOperationItem(ctx context.Context, item types.BatchOperationItem) (*types.BatchOperationItem, error) {
	if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
		return nil, fmt.Errorf("invalid batch operation item: %w", err)
	}

	return s.batchOperationItemRepo.Create(ctx, item)
}

func (s *batchOperationService) GetBatchOperationItem(ctx context.Context, id uuid.UUID) (*types.BatchOperationItem, error) {
	return s.batchOperationItemRepo.FindByID(ctx, id)
}

func (s *batchOperationService) ListBatchOperationItems(ctx context.Context, batchOperationID uuid.UUID) ([]types.BatchOperationItem, error) {
	return s.batchOperationItemRepo.FindByBatchOperation(ctx, batchOperationID)
}

func (s *batchOperationService) ListBatchOperationItemsByStatus(ctx context.Context, batchOperationID uuid.UUID, status string) ([]types.BatchOperationItem, error) {
	return s.batchOperationItemRepo.FindByStatus(ctx, batchOperationID, status)
}

func (s *batchOperationService) UpdateBatchOperationItem(ctx context.Context, item types.BatchOperationItem) (*types.BatchOperationItem, error) {
	if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
		return nil, fmt.Errorf("invalid batch operation item: %w", err)
	}

	return s.batchOperationItemRepo.Update(ctx, item)
}

func (s *batchOperationService) DeleteBatchOperationItem(ctx context.Context, id uuid.UUID) error {
	return s.batchOperationItemRepo.Delete(ctx, id)
}

// Business Logic Operations

func (s *batchOperationService) CreateStockAdjustmentBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, items []types.BatchOperationItem) (*types.BatchOperation, error) {
	// Validate items
	for _, item := range items {
		if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("invalid batch operation item: %w", err)
		}
	}

	// Create the batch operation
	operation := types.BatchOperation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		CompanyID:      &companyID,
		OperationType:  types.BatchOperationTypeStockAdjustment,
		Status:         types.BatchOperationStatusDraft,
		Reference:      reference,
		Description:    &description,
		Priority:       1,
		SourceType:     nil,
		SourceID:       nil,
		CreatedBy:      &createdBy,
		TotalItems:     len(items),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdOperation, err := s.batchOperationRepo.Create(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}

	// Create all items
	for i, item := range items {
		item.BatchOperationID = createdOperation.ID
		item.Sequence = i + 1
		item.Status = "pending"

		// Set operation-specific data
		if item.OperationData == nil {
			item.OperationData = make(map[string]interface{})
		}

		_, err := s.batchOperationItemRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch operation item: %w", err)
		}
	}

	return createdOperation, nil
}

func (s *batchOperationService) CreateStockTransferBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, sourceLocationID, destLocationID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error) {
	// Validate items
	for _, item := range items {
		if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("invalid batch operation item: %w", err)
		}
	}

	// Create the batch operation
	operation := types.BatchOperation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		CompanyID:      &companyID,
		OperationType:  types.BatchOperationTypeStockTransfer,
		Status:         types.BatchOperationStatusDraft,
		Reference:      reference,
		Description:    &description,
		Priority:       1,
		SourceType:     nil,
		SourceID:       nil,
		CreatedBy:      &createdBy,
		TotalItems:     len(items),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdOperation, err := s.batchOperationRepo.Create(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}

	// Create all items
	for i, item := range items {
		item.BatchOperationID = createdOperation.ID
		item.Sequence = i + 1
		item.Status = "pending"
		item.SourceLocationID = &sourceLocationID
		item.DestLocationID = &destLocationID

		// Set operation-specific data
		if item.OperationData == nil {
			item.OperationData = make(map[string]interface{})
		}

		_, err := s.batchOperationItemRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch operation item: %w", err)
		}
	}

	return createdOperation, nil
}

func (s *batchOperationService) CreateStockCountBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, locationID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error) {
	// Validate items
	for _, item := range items {
		if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("invalid batch operation item: %w", err)
		}
	}

	// Create the batch operation
	operation := types.BatchOperation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		CompanyID:      &companyID,
		OperationType:  types.BatchOperationTypeStockCount,
		Status:         types.BatchOperationStatusDraft,
		Reference:      reference,
		Description:    &description,
		Priority:       1,
		SourceType:     nil,
		SourceID:       nil,
		CreatedBy:      &createdBy,
		TotalItems:     len(items),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdOperation, err := s.batchOperationRepo.Create(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}

	// Create all items
	for i, item := range items {
		item.BatchOperationID = createdOperation.ID
		item.Sequence = i + 1
		item.Status = "pending"

		// Set operation-specific data
		if item.OperationData == nil {
			item.OperationData = make(map[string]interface{})
		}

		_, err := s.batchOperationItemRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch operation item: %w", err)
		}
	}

	return createdOperation, nil
}

func (s *batchOperationService) CreatePriceUpdateBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, currencyID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error) {
	// Validate items
	for _, item := range items {
		if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("invalid batch operation item: %w", err)
		}
	}

	// Create the batch operation
	operation := types.BatchOperation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		CompanyID:      &companyID,
		OperationType:  types.BatchOperationTypePriceUpdate,
		Status:         types.BatchOperationStatusDraft,
		Reference:      reference,
		Description:    &description,
		Priority:       1,
		SourceType:     nil,
		SourceID:       nil,
		CreatedBy:      &createdBy,
		TotalItems:     len(items),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdOperation, err := s.batchOperationRepo.Create(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}

	// Create all items
	for i, item := range items {
		item.BatchOperationID = createdOperation.ID
		item.Sequence = i + 1
		item.Status = "pending"

		// Set operation-specific data
		if item.OperationData == nil {
			item.OperationData = make(map[string]interface{})
		}

		_, err := s.batchOperationItemRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch operation item: %w", err)
		}
	}

	return createdOperation, nil
}

func (s *batchOperationService) CreateLocationUpdateBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, newLocationID uuid.UUID, items []types.BatchOperationItem) (*types.BatchOperation, error) {
	// Validate items
	for _, item := range items {
		if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("invalid batch operation item: %w", err)
		}
	}

	// Create the batch operation
	operation := types.BatchOperation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		CompanyID:      &companyID,
		OperationType:  types.BatchOperationTypeLocationUpdate,
		Status:         types.BatchOperationStatusDraft,
		Reference:      reference,
		Description:    &description,
		Priority:       1,
		SourceType:     nil,
		SourceID:       nil,
		CreatedBy:      &createdBy,
		TotalItems:     len(items),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdOperation, err := s.batchOperationRepo.Create(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}

	// Create all items
	for i, item := range items {
		item.BatchOperationID = createdOperation.ID
		item.Sequence = i + 1
		item.Status = "pending"
		item.DestLocationID = &newLocationID

		// Set operation-specific data
		if item.OperationData == nil {
			item.OperationData = make(map[string]interface{})
		}

		_, err := s.batchOperationItemRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch operation item: %w", err)
		}
	}

	return createdOperation, nil
}

func (s *batchOperationService) CreateStatusUpdateBatch(ctx context.Context, organizationID, companyID, createdBy uuid.UUID, reference, description string, newStatus string, items []types.BatchOperationItem) (*types.BatchOperation, error) {
	// Validate items
	for _, item := range items {
		if err := s.ValidateBatchOperationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("invalid batch operation item: %w", err)
		}
	}

	// Create the batch operation
	operation := types.BatchOperation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		CompanyID:      &companyID,
		OperationType:  types.BatchOperationTypeStatusUpdate,
		Status:         types.BatchOperationStatusDraft,
		Reference:      reference,
		Description:    &description,
		Priority:       1,
		SourceType:     nil,
		SourceID:       nil,
		CreatedBy:      &createdBy,
		TotalItems:     len(items),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdOperation, err := s.batchOperationRepo.Create(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch operation: %w", err)
	}

	// Create all items
	for i, item := range items {
		item.BatchOperationID = createdOperation.ID
		item.Sequence = i + 1
		item.Status = "pending"

		// Set operation-specific data
		if item.OperationData == nil {
			item.OperationData = make(map[string]interface{})
		}

		_, err := s.batchOperationItemRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to create batch operation item: %w", err)
		}
	}

	return createdOperation, nil
}

func (s *batchOperationService) ProcessBatchOperation(ctx context.Context, operationID uuid.UUID, processedBy uuid.UUID) (*types.BatchOperationResult, error) {
	// Get the operation to validate it exists and is in the right state
	operation, err := s.batchOperationRepo.FindByID(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch operation: %w", err)
	}
	if operation == nil {
		return nil, fmt.Errorf("batch operation not found")
	}

	// Validate operation can be processed
	if operation.Status != types.BatchOperationStatusDraft && operation.Status != types.BatchOperationStatusPending {
		return nil, fmt.Errorf("batch operation cannot be processed in status: %s", operation.Status)
	}

	// Process the operation using the database function
	result, err := s.batchOperationRepo.ProcessBatchOperation(ctx, operationID, processedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to process batch operation: %w", err)
	}

	return result, nil
}

func (s *batchOperationService) GetBatchOperationStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, operationType *types.BatchOperationType) (types.BatchOperationStatistics, error) {
	return s.batchOperationRepo.GetStatistics(ctx, organizationID, fromTime, toTime, operationType)
}

// Validation and Utility Methods

func (s *batchOperationService) ValidateBatchOperation(ctx context.Context, operation types.BatchOperation) error {
	if operation.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if operation.OperationType == "" {
		return fmt.Errorf("operation type is required")
	}

	// Validate operation type
	switch operation.OperationType {
	case types.BatchOperationTypeStockAdjustment,
	types.BatchOperationTypeStockTransfer,
	types.BatchOperationTypeStockCount,
	types.BatchOperationTypePriceUpdate,
	types.BatchOperationTypeLocationUpdate,
	types.BatchOperationTypeStatusUpdate:
		// Valid types
	default:
		return fmt.Errorf("invalid operation type: %s", operation.OperationType)
	}

	if operation.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Validate status
	switch operation.Status {
		case types.BatchOperationStatusDraft,
		types.BatchOperationStatusPending,
		types.BatchOperationStatusProcessing,
		types.BatchOperationStatusCompleted,
		types.BatchOperationStatusFailed,
		types.BatchOperationStatusCancelled:
		// Valid statuses
	default:
		return fmt.Errorf("invalid status: %s", operation.Status)
	}

	if operation.Reference == "" {
		return fmt.Errorf("reference is required")
	}

	// Validate company exists if provided
	if operation.CompanyID != nil && *operation.CompanyID != uuid.Nil {
		// TODO: Add company validation when company service is available
	}

	return nil
}

func (s *batchOperationService) ValidateBatchOperationItem(ctx context.Context, item types.BatchOperationItem) error {
	if item.BatchOperationID == uuid.Nil {
		return fmt.Errorf("batch operation ID is required")
	}

	if item.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	if item.ProductName == "" {
		return fmt.Errorf("product name is required")
	}

	if item.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Validate product exists
	product, err := s.productsRepo.FindByID(ctx, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to validate product: %w", err)
	}
	if product == nil {
		return fmt.Errorf("product not found: %s", item.ProductID)
	}

	// Validate locations if provided
	if item.SourceLocationID != nil && *item.SourceLocationID != uuid.Nil {
		location, err := s.inventoryService.GetStockLocation(ctx, *item.SourceLocationID)
		if err != nil {
			return fmt.Errorf("failed to validate source location: %w", err)
		}
		if location == nil {
			return fmt.Errorf("source location not found: %s", *item.SourceLocationID)
		}
	}

	if item.DestLocationID != nil && *item.DestLocationID != uuid.Nil {
		location, err := s.inventoryService.GetStockLocation(ctx, *item.DestLocationID)
		if err != nil {
			return fmt.Errorf("failed to validate destination location: %w", err)
		}
		if location == nil {
			return fmt.Errorf("destination location not found: %s", *item.DestLocationID)
		}
	}

	// Validate lot if provided
	if item.LotID != nil && *item.LotID != uuid.Nil {
		// TODO: Add lot validation when lot service is available
	}

	return nil
}

func (s *batchOperationService) CalculateBatchOperationTotals(ctx context.Context, batchOperationID uuid.UUID) (int, int, int, error) {
	items, err := s.batchOperationItemRepo.FindByBatchOperation(ctx, batchOperationID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get batch operation items: %w", err)
	}

	totalItems := len(items)
	successfulItems := 0
	failedItems := 0

	for _, item := range items {
		if item.Status == "processed" || item.Status == "success" {
			successfulItems++
		} else if item.Status == "failed" {
			failedItems++
		}
	}

	return totalItems, successfulItems, failedItems, nil
}
