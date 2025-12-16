package inventory

import (
	"context"
	"log/slog"

	"alieze-erp/internal/modules/inventory/handler"
	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/service"
	"alieze-erp/pkg/registry"

	"github.com/julienschmidt/httprouter"
)

// InventoryModule represents the Inventory module
type InventoryModule struct {
	inventoryHandler  *handler.InventoryHandler
	analyticsHandler *handler.AnalyticsHandler
	barcodeHandler   *handler.BarcodeHandler
	cycleCountHandler *handler.CycleCountHandler
	logger            *slog.Logger
}

// NewInventoryModule creates a new Inventory module
func NewInventoryModule() *InventoryModule {
	return &InventoryModule{}
}

// Name returns the module name
func (m *InventoryModule) Name() string {
	return "inventory"
}

// Init initializes the Inventory module
func (m *InventoryModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "inventory")
	m.logger.Info("Initializing Inventory module")

	// Create repositories
	warehouseRepo := repository.NewWarehouseRepository(deps.DB)
	locationRepo := repository.NewStockLocationRepository(deps.DB)
	quantRepo := repository.NewStockQuantRepository(deps.DB)
	moveRepo := repository.NewStockMoveRepository(deps.DB)
	analyticsRepo := repository.NewAnalyticsRepository(deps.DB)
	barcodeRepo := repository.NewBarcodeRepository(deps.DB)
	cycleCountRepo := repository.NewCycleCountRepository(deps.DB)

	// Create services
	inventoryService := service.NewInventoryService(warehouseRepo, locationRepo, quantRepo, moveRepo)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	barcodeService := service.NewBarcodeService(barcodeRepo)
	cycleCountService := service.NewCycleCountService(cycleCountRepo)

	// Create handlers
	m.inventoryHandler = handler.NewInventoryHandler(inventoryService)
	m.analyticsHandler = handler.NewAnalyticsHandler(analyticsService)
	m.barcodeHandler = handler.NewBarcodeHandler(barcodeService)
	m.cycleCountHandler = handler.NewCycleCountHandler(cycleCountService)

	m.logger.Info("Inventory module initialized successfully")
	return nil
}

// RegisterRoutes registers Inventory module routes
func (m *InventoryModule) RegisterRoutes(router interface{}) {
	if router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			if m.inventoryHandler != nil {
				m.inventoryHandler.RegisterRoutes(r)
			}
			if m.analyticsHandler != nil {
				m.analyticsHandler.RegisterRoutes(r)
			}
			if m.barcodeHandler != nil {
				m.barcodeHandler.RegisterRoutes(r)
			}
			if m.cycleCountHandler != nil {
				m.cycleCountHandler.RegisterRoutes(r)
			}
		}
	}
}

// RegisterEventHandlers registers event handlers for the Inventory module
func (m *InventoryModule) RegisterEventHandlers(bus interface{}) {
	// Inventory module doesn't currently need event handlers
	// but this method is required by the module interface
}

// Health checks the health of the Inventory module
func (m *InventoryModule) Health() error {
	return nil
}
