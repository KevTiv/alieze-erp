package registry

import (
	"context"
	"fmt"

	"github.com/julienschmidt/httprouter"
)

// Registry manages all modules in the ERP system
type Registry struct {
	modules map[string]Module
	deps    Dependencies
}

// NewRegistry creates a new module registry
func NewRegistry(deps Dependencies) *Registry {
	return &Registry{
		modules: make(map[string]Module),
		deps:    deps,
	}
}

// Register adds a module to the registry
func (r *Registry) Register(module Module) {
	r.modules[module.Name()] = module
}

// InitAll initializes all registered modules
func (r *Registry) InitAll(ctx context.Context) error {
	for name, module := range r.modules {
		if err := module.Init(ctx, r.deps); err != nil {
			return fmt.Errorf("failed to init %s: %w", name, err)
		}
	}
	return nil
}

// GetModule returns a module by name
func (r *Registry) GetModule(name string) (Module, bool) {
	module, exists := r.modules[name]
	return module, exists
}

// GetAllModules returns all registered modules
func (r *Registry) GetAllModules() map[string]Module {
	return r.modules
}

// RegisterAllRoutes registers routes for all modules
func (r *Registry) RegisterAllRoutes(router *httprouter.Router) {
	for _, module := range r.modules {
		module.RegisterRoutes(router)
	}
}

// RegisterAllEventHandlers registers event handlers for all modules
func (r *Registry) RegisterAllEventHandlers(bus interface{}) {
	for _, module := range r.modules {
		module.RegisterEventHandlers(bus)
	}
}

// UpdateDependencies updates the dependencies for all modules
func (r *Registry) UpdateDependencies(deps Dependencies) {
	r.deps = deps
}
