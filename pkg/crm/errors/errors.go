package errors

import (
	"fmt"
	"net/http"
)

// CRMError represents a standardized error in the CRM module
type CRMError struct {
	Code    string
	Message string
	HTTP    int
	Err     error
}

// Error implements the error interface
func (e *CRMError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *CRMError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code
func (e *CRMError) HTTPStatus() int {
	return e.HTTP
}

// IsCRMError checks if an error is a CRMError
func IsCRMError(err error) (*CRMError, bool) {
	if crmErr, ok := err.(*CRMError); ok {
		return crmErr, true
	}
	return nil, false
}

// Predefined error types
var (
	// Validation errors
	ErrNotFound        = &CRMError{Code: "NOT_FOUND", Message: "resource not found", HTTP: http.StatusNotFound}
	ErrInvalidInput    = &CRMError{Code: "INVALID_INPUT", Message: "invalid input", HTTP: http.StatusBadRequest}
	ErrValidationError = &CRMError{Code: "VALIDATION_ERROR", Message: "validation failed", HTTP: http.StatusBadRequest}

	// Authorization errors
	ErrPermissionDenied   = &CRMError{Code: "PERMISSION_DENIED", Message: "permission denied", HTTP: http.StatusForbidden}
	ErrOrganizationAccess = &CRMError{Code: "ORG_ACCESS", Message: "organization access denied", HTTP: http.StatusForbidden}
	ErrUnauthorized       = &CRMError{Code: "UNAUTHORIZED", Message: "unauthorized access", HTTP: http.StatusUnauthorized}

	// Business logic errors
	ErrConflict     = &CRMError{Code: "CONFLICT", Message: "conflict", HTTP: http.StatusConflict}
	ErrDuplicate    = &CRMError{Code: "DUPLICATE", Message: "duplicate resource", HTTP: http.StatusConflict}
	ErrInvalidState = &CRMError{Code: "INVALID_STATE", Message: "invalid state transition", HTTP: http.StatusBadRequest}

	// System errors
	ErrInternal           = &CRMError{Code: "INTERNAL", Message: "internal server error", HTTP: http.StatusInternalServerError}
	ErrDatabase           = &CRMError{Code: "DATABASE", Message: "database error", HTTP: http.StatusInternalServerError}
	ErrServiceUnavailable = &CRMError{Code: "SERVICE_UNAVAILABLE", Message: "service unavailable", HTTP: http.StatusServiceUnavailable}
)

// Wrap creates a new CRMError by wrapping an existing error
func Wrap(err error, code string, message string) *CRMError {
	httpStatus := http.StatusInternalServerError
	switch code {
	case "NOT_FOUND":
		httpStatus = http.StatusNotFound
	case "INVALID_INPUT", "VALIDATION_ERROR", "INVALID_STATE":
		httpStatus = http.StatusBadRequest
	case "PERMISSION_DENIED", "ORG_ACCESS":
		httpStatus = http.StatusForbidden
	case "UNAUTHORIZED":
		httpStatus = http.StatusUnauthorized
	case "CONFLICT", "DUPLICATE":
		httpStatus = http.StatusConflict
	}

	return &CRMError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
		Err:     err,
	}
}

// WrapWithStatus creates a new CRMError with a specific HTTP status
func WrapWithStatus(err error, code string, message string, httpStatus int) *CRMError {
	return &CRMError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
		Err:     err,
	}
}

// New creates a new CRMError without wrapping an existing error
func New(code string, message string) *CRMError {
	httpStatus := http.StatusInternalServerError
	switch code {
	case "NOT_FOUND":
		httpStatus = http.StatusNotFound
	case "INVALID_INPUT", "VALIDATION_ERROR", "INVALID_STATE":
		httpStatus = http.StatusBadRequest
	case "PERMISSION_DENIED", "ORG_ACCESS":
		httpStatus = http.StatusForbidden
	case "UNAUTHORIZED":
		httpStatus = http.StatusUnauthorized
	case "CONFLICT", "DUPLICATE":
		httpStatus = http.StatusConflict
	}

	return &CRMError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
	}
}

// NewWithStatus creates a new CRMError with a specific HTTP status
func NewWithStatus(code string, message string, httpStatus int) *CRMError {
	return &CRMError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
	}
}

// ValidationError creates a validation error with field details
type ValidationError struct {
	Code    string
	Field   string
	Message string
	HTTP    int
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s %s", v.Field, v.Message)
}

// HTTPStatus returns the HTTP status code
func (v *ValidationError) HTTPStatus() int {
	return v.HTTP
}

// NewValidationError creates a new validation error
func NewValidationError(field string, message string) *ValidationError {
	return &ValidationError{
		Code:    "VALIDATION_ERROR",
		Field:   field,
		Message: message,
		HTTP:    http.StatusBadRequest,
	}
}

// BusinessError represents a business logic error
type BusinessError struct {
	Code    string
	Message string
	HTTP    int
	Details map[string]interface{}
}

// Error implements the error interface
func (b *BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", b.Code, b.Message)
}

// HTTPStatus returns the HTTP status code
func (b *BusinessError) HTTPStatus() int {
	return b.HTTP
}

// NewBusinessError creates a new business error
func NewBusinessError(code string, message string, details map[string]interface{}) *BusinessError {
	httpStatus := http.StatusBadRequest
	switch code {
	case "CONFLICT", "DUPLICATE":
		httpStatus = http.StatusConflict
	case "INVALID_STATE":
		httpStatus = http.StatusBadRequest
	}

	return &BusinessError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
		Details: details,
	}
}
