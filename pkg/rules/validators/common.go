package rules

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Common validators

// RequireUUID validates that a field is a valid UUID
func RequireUUID(ctx context.Context, entity interface{}) error {
	// This validator should be used with a field name parameter
	// For now, we'll implement a basic version that checks if the entity itself is a valid UUID

	// Check if entity is a UUID
	if u, ok := entity.(uuid.UUID); ok {
		// Check if UUID is nil (all zeros)
		var nilUUID uuid.UUID
		if u == nilUUID {
			return errors.New("uuid cannot be nil")
		}
		return nil
	}

	// Check if entity is a pointer to UUID
	if uuidPtr, ok := entity.(*uuid.UUID); ok {
		if uuidPtr == nil {
			return errors.New("uuid cannot be nil")
		}
		var nilUUID uuid.UUID
		if *uuidPtr == nilUUID {
			return errors.New("uuid cannot be nil")
		}
		return nil
	}

	return errors.New("entity is not a UUID")
}

// RequireString validates that a string field is not empty
func RequireString(ctx context.Context, entity interface{}) error {
	// Check if entity is a string
	if str, ok := entity.(string); ok {
		if str == "" {
			return errors.New("string cannot be empty")
		}
		return nil
	}

	// Check if entity is a pointer to string
	if strPtr, ok := entity.(*string); ok {
		if strPtr == nil || *strPtr == "" {
			return errors.New("string cannot be empty")
		}
		return nil
	}

	return errors.New("entity is not a string")
}

// EmailFormat validates that a field contains a valid email
func EmailFormat(ctx context.Context, entity interface{}) error {
	// Get the string value
	var emailStr string

	if str, ok := entity.(string); ok {
		emailStr = str
	} else if strPtr, ok := entity.(*string); ok {
		if strPtr == nil {
			return errors.New("email cannot be nil")
		}
		emailStr = *strPtr
	} else {
		return errors.New("entity is not a string")
	}

	// Simple email validation
	if emailStr == "" {
		return errors.New("email cannot be empty")
	}

	// Check for @ symbol
	if !strings.Contains(emailStr, "@") {
		return errors.New("email must contain @ symbol")
	}

	// Check for . in domain part
	parts := strings.Split(emailStr, "@")
	if len(parts) != 2 {
		return errors.New("email must have exactly one @ symbol")
	}

	if !strings.Contains(parts[1], ".") {
		return errors.New("email domain must contain a dot")
	}

	return nil
}

// MinLength validates minimum string length
func MinLength(min int) func(ctx context.Context, entity interface{}) error {
	return func(ctx context.Context, entity interface{}) error {
		var str string

		// Get string value
		if s, ok := entity.(string); ok {
			str = s
		} else if sPtr, ok := entity.(*string); ok {
			if sPtr == nil {
				return errors.New("string cannot be nil")
			}
			str = *sPtr
		} else {
			return errors.New("entity is not a string")
		}

		if len(str) < min {
			return fmt.Errorf("string length %d is less than minimum %d", len(str), min)
		}

		return nil
	}
}

// MaxLength validates maximum string length
func MaxLength(max int) func(ctx context.Context, entity interface{}) error {
	return func(ctx context.Context, entity interface{}) error {
		var str string

		// Get string value
		if s, ok := entity.(string); ok {
			str = s
		} else if sPtr, ok := entity.(*string); ok {
			if sPtr == nil {
				return errors.New("string cannot be nil")
			}
			str = *sPtr
		} else {
			return errors.New("entity is not a string")
		}

		if len(str) > max {
			return fmt.Errorf("string length %d exceeds maximum %d", len(str), max)
		}

		return nil
	}
}

// MinArrayLength validates minimum array length
func MinArrayLength(min int) func(ctx context.Context, entity interface{}) error {
	return func(ctx context.Context, entity interface{}) error {
		var length int

		// Handle different slice/array types
		switch v := entity.(type) {
		case []interface{}:
			length = len(v)
		case []string:
			length = len(v)
		case []int:
			length = len(v)
		case []float64:
			length = len(v)
		case *[]interface{}:
			if v == nil {
				return errors.New("array cannot be nil")
			}
			length = len(*v)
		case *[]string:
			if v == nil {
				return errors.New("array cannot be nil")
			}
			length = len(*v)
		default:
			// Try reflection for other slice types
			reflectValue := reflect.ValueOf(entity)
			if reflectValue.Kind() == reflect.Slice || reflectValue.Kind() == reflect.Array {
				length = reflectValue.Len()
			} else if reflectValue.Kind() == reflect.Ptr {
				if reflectValue.IsNil() {
					return errors.New("array cannot be nil")
				}
				reflectValue = reflectValue.Elem()
				if reflectValue.Kind() == reflect.Slice || reflectValue.Kind() == reflect.Array {
					length = reflectValue.Len()
				}
			}
		}

		if length < min {
			return fmt.Errorf("array length %d is less than minimum %d", length, min)
		}

		return nil
	}
}

// InRange validates that a numeric value is within range
func InRange(min, max float64) func(ctx context.Context, entity interface{}) error {
	return func(ctx context.Context, entity interface{}) error {
		var value float64

		// Handle different numeric types
		switch v := entity.(type) {
		case float64:
			value = v
		case float32:
			value = float64(v)
		case int:
			value = float64(v)
		case int32:
			value = float64(v)
		case int64:
			value = float64(v)
		case *float64:
			if v == nil {
				return errors.New("value cannot be nil")
			}
			value = *v
		case *float32:
			if v == nil {
				return errors.New("value cannot be nil")
			}
			value = float64(*v)
		case *int:
			if v == nil {
				return errors.New("value cannot be nil")
			}
			value = float64(*v)
		default:
			return errors.New("entity is not a numeric type")
		}

		if value < min {
			return fmt.Errorf("value %f is less than minimum %f", value, min)
		}
		if value > max {
			return fmt.Errorf("value %f is greater than maximum %f", value, max)
		}

		return nil
	}
}

// MatchPattern validates that a string matches a regex pattern
func MatchPattern(pattern string) func(ctx context.Context, entity interface{}) error {
	return func(ctx context.Context, entity interface{}) error {
		// Compile regex
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}

		// Implementation would match string against pattern
		return fmt.Errorf("match_pattern validator not implemented for pattern %s", pattern)
	}
}
