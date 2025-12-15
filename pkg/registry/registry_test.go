package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockModule is a test implementation of Module interface
type MockModule struct {
	name string
}

func (m *MockModule) Name() string {
	return m.name
}

func (m *MockModule) Init(ctx context.Context, deps Dependencies) error {
	return nil
}

func (m *MockModule) RegisterRoutes(router interface{}) {
	// No-op for test
}

func (m *MockModule) RegisterEventHandlers(bus interface{}) {
	// No-op for test
}

func (m *MockModule) Health() error {
	return nil
}

func TestRegistry(t *testing.T) {
	ctx := context.Background()

	// Create dependencies
	deps := Dependencies{
		Logger: nil, // We'll add proper logger in real implementation
	}

	// Create registry
	registry := NewRegistry(deps)

	// Test registering modules
	crmModule := &MockModule{name: "crm"}
	salesModule := &MockModule{name: "sales"}

	registry.Register(crmModule)
	registry.Register(salesModule)

	// Test initialization
	err := registry.InitAll(ctx)
	assert.NoError(t, err)

	// Test getting modules
	module, exists := registry.GetModule("crm")
	assert.True(t, exists)
	assert.Equal(t, "crm", module.Name())

	// Test getting all modules
	allModules := registry.GetAllModules()
	assert.Len(t, allModules, 2)
}

// MockFailingModule is a test module that fails initialization
type MockFailingModule struct {
	name string
}

func (m *MockFailingModule) Name() string {
	return m.name
}

func (m *MockFailingModule) Init(ctx context.Context, deps Dependencies) error {
	return assert.AnError
}

func (m *MockFailingModule) RegisterRoutes(router interface{}) {
	// No-op for test
}

func (m *MockFailingModule) RegisterEventHandlers(bus interface{}) {
	// No-op for test
}

func (m *MockFailingModule) Health() error {
	return nil
}

func TestRegistryInitError(t *testing.T) {
	ctx := context.Background()

	deps := Dependencies{}
	registry := NewRegistry(deps)

	// Create a module that fails initialization
	failingModule := &MockFailingModule{name: "failing"}

	registry.Register(failingModule)

	err := registry.InitAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to init failing")
}
