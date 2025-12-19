package products

import (
	"context"
	"log/slog"

	"github.com/KevTiv/alieze-erp/internal/modules/products/handler"
	"github.com/KevTiv/alieze-erp/internal/modules/products/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/products/service"
	"github.com/KevTiv/alieze-erp/pkg/registry"
	"github.com/julienschmidt/httprouter"
)

// ProductsModule represents the Products module
type ProductsModule struct {
	productHandler *handler.ProductHandler
	logger         *slog.Logger
}

// NewProductsModule creates a new Products module
func NewProductsModule() *ProductsModule {
	return &ProductsModule{}
}

// Name returns the module name
func (m *ProductsModule) Name() string {
	return "products"
}

// Init initializes the Products module
func (m *ProductsModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "products")
	m.logger.Info("Initializing Products module")

	// Create repositories
	productRepo := repository.NewProductRepository(deps.DB)

	// Create services
	productService := service.NewProductService(productRepo, nil)

	// Create handlers
	m.productHandler = handler.NewProductHandler(productService)

	m.logger.Info("Products module initialized successfully")
	return nil
}

// RegisterRoutes registers Products module routes
func (m *ProductsModule) RegisterRoutes(router interface{}) {
	if m.productHandler != nil && router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			m.productHandler.RegisterRoutes(r)
		}
	}
}

// RegisterEventHandlers registers event handlers for the Products module
func (m *ProductsModule) RegisterEventHandlers(bus interface{}) {
	// TODO: Implement event handlers when event system is integrated
}

// Health checks the health of the Products module
func (m *ProductsModule) Health() error {
	return nil
}
