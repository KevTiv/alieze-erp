package domain

import (
	"time"

	"github.com/google/uuid"
)

// BatchOperationType represents the type of batch operation
type BatchOperationType string

const (
	BatchOperationTypeStockAdjustment BatchOperationType = "stock_adjustment"
	BatchOperationTypeStockTransfer    BatchOperationType = "stock_transfer"
	BatchOperationTypeStockCount       BatchOperationType = "stock_count"
	BatchOperationTypePriceUpdate      BatchOperationType = "price_update"
	BatchOperationTypeLocationUpdate   BatchOperationType = "location_update"
	BatchOperationTypeStatusUpdate     BatchOperationType = "status_update"
)

// BatchOperationStatus represents the status of a batch operation
type BatchOperationStatus string

const (
	BatchOperationStatusDraft     BatchOperationStatus = "draft"
	BatchOperationStatusPending   BatchOperationStatus = "pending"
	BatchOperationStatusProcessing BatchOperationStatus = "processing"
	BatchOperationStatusCompleted  BatchOperationStatus = "completed"
	BatchOperationStatusFailed    BatchOperationStatus = "failed"
	BatchOperationStatusCancelled  BatchOperationStatus = "cancelled"
)

// BatchOperation represents a batch operation for bulk inventory updates
type BatchOperation struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID `json:"company_id,omitempty" db:"company_id"`

	// Operation details
	OperationType   BatchOperationType `json:"operation_type" db:"operation_type"`
	Status          BatchOperationStatus `json:"status" db:"status"`
	Reference       string `json:"reference" db:"reference"`
	Description     *string `json:"description,omitempty" db:"description"`
	Priority        int `json:"priority" db:"priority"`

	// Source information
	SourceType      *string `json:"source_type,omitempty" db:"source_type"`
	SourceID        *uuid.UUID `json:"source_id,omitempty" db:"source_id"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`

	// Processing information
	ProcessedBy     *uuid.UUID `json:"processed_by,omitempty" db:"processed_by"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	TotalItems      int `json:"total_items" db:"total_items"`
	SuccessfulItems int `json:"successful_items" db:"successful_items"`
	FailedItems     int `json:"failed_items" db:"failed_items"`

	// Error handling
	ErrorMessage    *string `json:"error_message,omitempty" db:"error_message"`
	ErrorDetails    map[string]interface{} `json:"error_details,omitempty" db:"error_details"`

	// Standard fields
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Metadata
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// BatchOperationItem represents an individual item within a batch operation
type BatchOperationItem struct {
	ID              uuid.UUID `json:"id" db:"id"`
	BatchOperationID uuid.UUID `json:"batch_operation_id" db:"batch_operation_id"`
	Sequence        int `json:"sequence" db:"sequence"`

	// Item identification
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	ProductName     string `json:"product_name" db:"product_name"`
	LotID           *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	SerialNumber    *string `json:"serial_number,omitempty" db:"serial_number"`

	// Location information
	SourceLocationID *uuid.UUID `json:"source_location_id,omitempty" db:"source_location_id"`
	DestLocationID   *uuid.UUID `json:"dest_location_id,omitempty" db:"dest_location_id"`

	// Quantity information
	CurrentQuantity  float64 `json:"current_quantity" db:"current_quantity"`
	AdjustmentQuantity float64 `json:"adjustment_quantity" db:"adjustment_quantity"`
	NewQuantity      float64 `json:"new_quantity" db:"new_quantity"`

	// Operation-specific data
	OperationData    map[string]interface{} `json:"operation_data" db:"operation_data"`

	// Processing status
	Status          string `json:"status" db:"status"` // "pending", "processed", "failed"
	ErrorMessage    *string `json:"error_message,omitempty" db:"error_message"`

	// Standard fields
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// BatchOperationResult represents the result of processing a batch operation
type BatchOperationResult struct {
	BatchOperationID uuid.UUID `json:"batch_operation_id"`
	Status           BatchOperationStatus `json:"status"`
	TotalItems       int `json:"total_items"`
	SuccessfulItems  int `json:"successful_items"`
	FailedItems      int `json:"failed_items"`
	ErrorMessage     *string `json:"error_message,omitempty"`

	// Detailed results for each item
	ItemResults []BatchOperationItemResult `json:"item_results"`

	// Summary statistics
	ProcessingTime time.Duration `json:"processing_time"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
}

// BatchOperationItemResult represents the result of processing an individual batch item
type BatchOperationItemResult struct {
	ItemID      uuid.UUID `json:"item_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Status      string `json:"status"` // "success", "failed", "skipped"
	ErrorMessage *string `json:"error_message,omitempty"`

	// Before/after values
	BeforeValue interface{} `json:"before_value,omitempty"`
	AfterValue  interface{} `json:"after_value,omitempty"`

	// Any resulting documents/records
	ResultID     *uuid.UUID `json:"result_id,omitempty"`
	ResultType   *string `json:"result_type,omitempty"`
}

// BatchOperationStatistics represents statistics about batch operations
type BatchOperationStatistics struct {
	TotalOperations      int     `json:"total_operations"`
	CompletedOperations  int     `json:"completed_operations"`
	FailedOperations     int     `json:"failed_operations"`
	PendingOperations    int     `json:"pending_operations"`
	ProcessingOperations int     `json:"processing_operations"`

	SuccessRate      float64 `json:"success_rate"`
	FailureRate      float64 `json:"failure_rate"`
	AverageItems      float64 `json:"average_items"`
	AverageProcessing time.Duration `json:"average_processing"`

	// By operation type
	ByType map[BatchOperationType]int `json:"by_type"`

	// Time-based statistics
	Last30Days int `json:"last_30_days"`
	Last7Days  int `json:"last_7_days"`
	Today      int `json:"today"`
}

// StockAdjustmentData represents data specific to stock adjustment operations
type StockAdjustmentData struct {
	AdjustmentType string `json:"adjustment_type"` // "increase", "decrease", "set"
	Reason         string `json:"reason"`
	Reference      string `json:"reference"`
	CostPrice      *float64 `json:"cost_price,omitempty"`
}

// StockTransferData represents data specific to stock transfer operations
type StockTransferData struct {
	TransferType string `json:"transfer_type"` // "internal", "incoming", "outgoing"
	Reference    string `json:"reference"`
	Priority     string `json:"priority"`
}

// PriceUpdateData represents data specific to price update operations
type PriceUpdateData struct {
	UpdateType string `json:"update_type"` // "standard", "sale", "cost"
	NewPrice   float64 `json:"new_price"`
	CurrencyID uuid.UUID `json:"currency_id"`
}

// LocationUpdateData represents data specific to location update operations
type LocationUpdateData struct {
	UpdateType string `json:"update_type"` // "move", "assign", "remove"
	NewLocationID uuid.UUID `json:"new_location_id"`
}

// StatusUpdateData represents data specific to status update operations
type StatusUpdateData struct {
	NewStatus string `json:"new_status"`
	Reason    string `json:"reason"`
}
