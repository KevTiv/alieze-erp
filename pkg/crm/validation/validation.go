package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validator defines the interface for validation
type Validator interface {
	Validate() error
}

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s %s", v.Field, v.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}

	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = err.Error()
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// ValidateRequired checks if a field is required and not empty
func ValidateRequired(field string, value interface{}) error {
	if value == nil {
		return &ValidationError{Field: field, Message: "is required", Value: value}
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
	case *string:
		if v == nil {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
		if strings.TrimSpace(*v) == "" {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
	case int, int32, int64:
		if v == 0 {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
	case *int:
		if v == nil || *v == 0 {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
	case *int32:
		if v == nil || *v == 0 {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
	case *int64:
		if v == nil || *v == 0 {
			return &ValidationError{Field: field, Message: "is required", Value: value}
		}
	}
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Email is optional
	}

	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	if !matched {
		return &ValidationError{Field: "email", Message: "invalid email format", Value: email}
	}
	return nil
}

// ValidateUUID validates UUID format
func ValidateUUID(id string) error {
	if id == "" {
		return &ValidationError{Field: "id", Message: "is required", Value: id}
	}

	_, err := uuid.Parse(id)
	if err != nil {
		return &ValidationError{Field: "id", Message: "invalid UUID format", Value: id}
	}
	return nil
}

// ValidatePhone validates phone number format (basic validation)
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}

	// Remove common phone number characters
	cleaned := regexp.MustCompile(`[\s\-\(\)\+]`).ReplaceAllString(phone, "")

	// Check if it contains only digits
	matched, _ := regexp.MatchString(`^\d+$`, cleaned)
	if !matched || len(cleaned) < 10 {
		return &ValidationError{Field: "phone", Message: "invalid phone number format", Value: phone}
	}
	return nil
}

// ValidateLength validates string length
func ValidateLength(field string, value string, minLength, maxLength int) error {
	if value == "" {
		return nil // Optional field
	}

	length := len(value)
	if length < minLength {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", minLength),
			Value:   value,
		}
	}

	if length > maxLength {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be no more than %d characters", maxLength),
			Value:   value,
		}
	}
	return nil
}

// ValidateEnum validates that a value is in the allowed enum values
func ValidateEnum(field string, value interface{}, allowedValues []string) error {
	if value == nil {
		return nil // Optional field
	}

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case *string:
		if v == nil {
			return nil
		}
		strValue = *v
	default:
		return &ValidationError{
			Field:   field,
			Message: "must be a string",
			Value:   value,
		}
	}

	for _, allowed := range allowedValues {
		if strValue == allowed {
			return nil
		}
	}

	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", ")),
		Value:   value,
	}
}

// ValidateDate validates date format and range
func ValidateDate(field string, value interface{}, minDate, maxDate *time.Time) error {
	if value == nil {
		return nil // Optional field
	}

	var date time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		date = v
	case *time.Time:
		if v == nil {
			return nil
		}
		date = *v
	case string:
		date, err = time.Parse(time.RFC3339, v)
		if err != nil {
			date, err = time.Parse("2006-01-02", v)
			if err != nil {
				return &ValidationError{
					Field:   field,
					Message: "invalid date format",
					Value:   value,
				}
			}
		}
	default:
		return &ValidationError{
			Field:   field,
			Message: "must be a date",
			Value:   value,
		}
	}

	if minDate != nil && date.Before(*minDate) {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be after %s", minDate.Format("2006-01-02")),
			Value:   value,
		}
	}

	if maxDate != nil && date.After(*maxDate) {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be before %s", maxDate.Format("2006-01-02")),
			Value:   value,
		}
	}

	return nil
}

// ValidateSlice validates slice properties
func ValidateSlice(field string, value interface{}, minLength, maxLength int) error {
	if value == nil {
		return nil // Optional field
	}

	var length int
	switch v := value.(type) {
	case []string:
		length = len(v)
	case []interface{}:
		length = len(v)
	case []*string:
		length = len(v)
	default:
		return &ValidationError{
			Field:   field,
			Message: "must be a slice or array",
			Value:   value,
		}
	}

	if length < minLength {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must have at least %d items", minLength),
			Value:   value,
		}
	}

	if maxLength > 0 && length > maxLength {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must have no more than %d items", maxLength),
			Value:   value,
		}
	}

	return nil
}

// ValidatePositive validates that a number is positive
func ValidatePositive(field string, value interface{}) error {
	if value == nil {
		return nil // Optional field
	}

	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	case *int:
		if v == nil {
			return nil
		}
		num = float64(*v)
	case *int32:
		if v == nil {
			return nil
		}
		num = float64(*v)
	case *int64:
		if v == nil {
			return nil
		}
		num = float64(*v)
	case *float32:
		if v == nil {
			return nil
		}
		num = float64(*v)
	case *float64:
		if v == nil {
			return nil
		}
		num = *v
	default:
		return &ValidationError{
			Field:   field,
			Message: "must be a number",
			Value:   value,
		}
	}

	if num <= 0 {
		return &ValidationError{
			Field:   field,
			Message: "must be positive",
			Value:   value,
		}
	}

	return nil
}

// ValidateStruct validates a struct that implements the Validator interface
func ValidateStruct(v Validator) error {
	return v.Validate()
}

// ValidateMultiple runs multiple validation functions and returns combined errors
func ValidateMultiple(validations ...func() error) error {
	var errors ValidationErrors

	for _, validation := range validations {
		if err := validation(); err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				errors = append(errors, *validationErr)
			} else {
				// Convert generic error to ValidationError
				errors = append(errors, ValidationError{
					Field:   "general",
					Message: err.Error(),
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// IsValidEmail checks if an email is valid (returns bool)
func IsValidEmail(email string) bool {
	return ValidateEmail(email) == nil
}

// IsValidUUID checks if a UUID is valid (returns bool)
func IsValidUUID(id string) bool {
	return ValidateUUID(id) == nil
}

// IsValidPhone checks if a phone number is valid (returns bool)
func IsValidPhone(phone string) bool {
	return ValidatePhone(phone) == nil
}
