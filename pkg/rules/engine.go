package rules

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sync"

	"gopkg.in/yaml.v3"
)

// ValidatorFunc validates an entity according to business rules
type ValidatorFunc func(ctx context.Context, entity interface{}) error

// RuleEngine manages validation rules and business logic
type RuleEngine struct {
	validators map[string]ValidatorFunc
	config     *RuleConfig
	mu         sync.RWMutex
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(config *RuleConfig) *RuleEngine {
	re := &RuleEngine{
		validators: make(map[string]ValidatorFunc),
		config:     config,
	}

	// Register built-in validators
	re.RegisterBuiltInValidators()

	return re
}

// RegisterBuiltInValidators registers all built-in validator functions
func (re *RuleEngine) RegisterBuiltInValidators() {
	// Register common validators from validators package
	// These will be imported and registered when the rule engine is initialized
}

// LoadConfigFromFile loads rule configuration from a YAML file
func (re *RuleEngine) LoadConfigFromFile(filePath string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read rule config file: %w", err)
	}

	// Parse YAML
	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse rule config: %w", err)
	}

	re.config = &config
	return nil
}

// LoadConfigsFromDirectory loads all rule configuration files from a directory
func (re *RuleEngine) LoadConfigsFromDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read rule config directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml" {
			filePath := filepath.Join(dirPath, file.Name())
			if err := re.LoadConfigFromFile(filePath); err != nil {
				return fmt.Errorf("failed to load config %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

// RegisterValidator adds a validator function
func (re *RuleEngine) RegisterValidator(name string, validator ValidatorFunc) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.validators[name] = validator
}

// Validate runs validation for a specific rule
func (re *RuleEngine) Validate(ctx context.Context, ruleName string, entity interface{}) error {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// First try direct validator lookup
	if validator, exists := re.validators[ruleName]; exists {
		return validator(ctx, entity)
	}

	// If no direct validator, try to find validation rules in config
	if re.config != nil && re.config.Modules != nil {
		// Parse ruleName in format "module.rule_name"
		// For now, we'll look for the rule in any module
		for _, moduleRules := range re.config.Modules {
			if moduleRules.Validation != nil {
				if validationRules, exists := moduleRules.Validation[ruleName]; exists {
					return re.validateWithRules(ctx, entity, validationRules)
				}
			}
		}
	}

	return nil
}

// validateWithRules validates an entity using a list of validation rules
func (re *RuleEngine) validateWithRules(ctx context.Context, entity interface{}, rules []ValidationRule) error {
	for _, rule := range rules {
		if validator, exists := re.validators[rule.Validator]; exists {
			// Get the field value from the entity
			fieldValue, err := re.getFieldValue(entity, rule.Field)
			if err != nil {
				return fmt.Errorf("failed to get field %s: %w", rule.Field, err)
			}

			// Apply the validator to the field value
			if err := validator(ctx, fieldValue); err != nil {
				return fmt.Errorf("validation failed for field %s: %w", rule.Field, err)
			}
		}
	}

	return nil
}

// getFieldValue gets a field value from an entity using reflection
func (re *RuleEngine) getFieldValue(entity interface{}, fieldName string) (interface{}, error) {
	reflectValue := reflect.ValueOf(entity)

	// Handle pointers
	if reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() {
			return nil, errors.New("entity is nil")
		}
		reflectValue = reflectValue.Elem()
	}

	// Get the field by name
	if reflectValue.Kind() == reflect.Struct {
		fieldValue := reflectValue.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			return nil, fmt.Errorf("field %s not found", fieldName)
		}
		return fieldValue.Interface(), nil
	}

	return nil, fmt.Errorf("entity is not a struct: %T", entity)
}

// Evaluate evaluates a rule and returns a boolean result
func (re *RuleEngine) Evaluate(ctx context.Context, ruleName string, entity interface{}) (bool, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// Check if there's a validator for this rule
	if validator, exists := re.validators[ruleName]; exists {
		err := validator(ctx, entity)
		// If validation passes, return true; if it fails, return false
		return err == nil, nil
	}

	// If no validator found, check config for business rules
	if re.config != nil && re.config.Modules != nil {
		for _, moduleRules := range re.config.Modules {
			if moduleRules.Validation != nil {
				if validationRules, exists := moduleRules.Validation[ruleName]; exists {
					err := re.validateWithRules(ctx, entity, validationRules)
					return err == nil, nil
				}
			}
		}
	}

	// If no rule found, return false (rule doesn't exist or doesn't apply)
	return false, fmt.Errorf("rule %s not found", ruleName)
}

// GetConfig returns the rule configuration
func (re *RuleEngine) GetConfig() *RuleConfig {
	return re.config
}
