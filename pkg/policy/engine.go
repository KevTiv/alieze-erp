package policy

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// PolicyConfig defines the structure for policy configuration
type PolicyConfig struct {
	Permissions map[string]map[string]string `yaml:"permissions"`
	Roles      map[string]struct {
		Permissions []string `yaml:"permissions"`
	} `yaml:"roles"`
}

// Engine manages authorization policies
type Engine struct {
	enforcer   interface{}
	validators map[string]func(ctx context.Context, subject, object, action string) bool
	config     *PolicyConfig
	mu         sync.RWMutex
}

// NewEngine creates a new policy engine
func NewEngine(enforcer interface{}) *Engine {
	return &Engine{
		enforcer:   enforcer,
		validators: make(map[string]func(ctx context.Context, subject, object, action string) bool),
		config:     nil,
	}
}

// LoadConfigFromFile loads policy configuration from a YAML file
func (e *Engine) LoadConfigFromFile(filePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read policy config file: %w", err)
	}

	// Parse YAML
	var config PolicyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse policy config: %w", err)
	}

	e.config = &config
	return nil
}

// LoadConfigsFromDirectory loads all policy configuration files from a directory
func (e *Engine) LoadConfigsFromDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read policy config directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml" {
			filePath := filepath.Join(dirPath, file.Name())
			if err := e.LoadConfigFromFile(filePath); err != nil {
				return fmt.Errorf("failed to load config %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

// CheckPermission checks if a subject has permission to perform an action on an object
func (e *Engine) CheckPermission(ctx context.Context, subject, object, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check with Casbin enforcer first
	if enforcer, ok := e.enforcer.(interface {
		CheckPermission(ctx context.Context, subject, object, action string) (bool, error)
	}); ok {
		return enforcer.CheckPermission(ctx, subject, object, action)
	}

	// Fallback to configuration-based validation
	if e.config != nil {
		// For now, we'll implement a simple permission system
		// In production, this would be more sophisticated with role inheritance, etc.

		// For demo purposes, we'll allow all actions if no specific config
		// In a real system, you'd check the subject's roles and permissions
		return true, nil
	}

	// Fallback to simple validation
	if validator, exists := e.validators[action]; exists {
		return validator(ctx, subject, object, action), nil
	}

	return false, fmt.Errorf("no policy validator found for action: %s", action)
}

// RegisterValidator adds a custom permission validator
func (e *Engine) RegisterValidator(action string, validator func(ctx context.Context, subject, object, action string) bool) {
	e.validators[action] = validator
}
