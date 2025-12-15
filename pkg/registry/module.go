package registry

import "context"

// Module represents a modular component of the ERP system
type Module interface {
	Name() string
	Init(ctx context.Context, deps Dependencies) error
	RegisterRoutes(router interface{})
	RegisterEventHandlers(bus interface{})
	Health() error
}
