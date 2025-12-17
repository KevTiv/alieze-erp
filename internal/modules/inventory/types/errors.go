package types

import (
	"fmt"
)

// Inventory module error types
var (
	ErrStockMoveNotFound      = fmt.Errorf("stock move not found")
	ErrInsufficientStock      = fmt.Errorf("insufficient stock available")
	ErrInvalidStockMove       = fmt.Errorf("invalid stock move operation")
	ErrLocationNotFound       = fmt.Errorf("stock location not found")
	ErrWarehouseNotFound      = fmt.Errorf("warehouse not found")
	ErrProductNotFound        = fmt.Errorf("product not found")
	ErrStockQuantNotFound     = fmt.Errorf("stock quant not found")
	ErrInvalidLocationType    = fmt.Errorf("invalid location type for operation")
	ErrStockMoveAlreadyDone   = fmt.Errorf("stock move already completed")
	ErrStockMoveCannotCancel  = fmt.Errorf("stock move cannot be canceled in current state")
	ErrNegativeStockQuantity  = fmt.Errorf("negative stock quantity not allowed")
	ErrReservedQuantityExceedsAvailable = fmt.Errorf("reserved quantity exceeds available quantity")
)

// BusinessLogicError represents a business logic validation error
type BusinessLogicError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e BusinessLogicError) Error() string {
	return e.Message
}

// NewBusinessLogicError creates a new business logic error
func NewBusinessLogicError(code, message string) BusinessLogicError {
	return BusinessLogicError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the business logic error
func (e BusinessLogicError) WithDetail(key string, value interface{}) BusinessLogicError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (es ValidationErrors) Error() string {
	if len(es) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("%d validation errors occurred", len(es))
}

// Add adds a validation error to the collection
func (es *ValidationErrors) Add(field, message string) {
	*es = append(*es, ValidationError{
		Field:   field,
		Message: message,
	})
}
