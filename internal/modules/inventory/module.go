package inventory

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/handler"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/service"
	productsRepo "github.com/KevTiv/alieze-erp/internal/modules/products/repository"
	"github.com/KevTiv/alieze-erp/pkg/registry"

	"github.com/julienschmidt/httprouter"
)

// InventoryModule represents the Inventory module
type InventoryModule struct {
	inventoryHandler          *handler.InventoryHandler
	analyticsHandler         *handler.AnalyticsHandler
	barcodeHandler           *handler.BarcodeHandler
	cycleCountHandler        *handler.CycleCountHandler
	replenishmentHandler     *handler.ReplenishmentHandler
	batchOperationHandler   *handler.BatchOperationHandler
	qualityControlHandler   *handler.QualityControlHandler
	stockPackageHandler     *handler.StockPackageHandler
	stockLotHandler         *handler.StockLotHandler
	procurementGroupHandler *handler.ProcurementGroupHandler
	stockRuleHandler        *handler.StockRuleHandler
	stockPickingTypeHandler *handler.StockPickingTypeHandler
	stockPickingHandler     *handler.StockPickingHandler
	stockMoveHandler        *handler.StockMoveHandler
	integrationService      *service.InventoryIntegrationService
	logger                 *slog.Logger
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
	replenishmentRuleRepo := repository.NewReplenishmentRuleRepository(deps.DB)
	replenishmentOrderRepo := repository.NewReplenishmentOrderRepository(deps.DB)

	// Batch Operation Repositories
	batchOperationRepo := repository.NewBatchOperationRepository(deps.DB)
	batchOperationItemRepo := repository.NewBatchOperationItemRepository(deps.DB)

	// Quality Control Repositories
	qcInspectionRepo := repository.NewQualityControlInspectionRepository(deps.DB)
	qcChecklistRepo := repository.NewQualityControlChecklistRepository(deps.DB)
	qcChecklistItemRepo := repository.NewQualityChecklistItemRepository(deps.DB)
	qcInspectionItemRepo := repository.NewQualityControlInspectionItemRepository(deps.DB)
	qcAlertRepo := repository.NewQualityControlAlertRepository(deps.DB)

	// New Inventory Repositories
	stockPackageRepo := repository.NewStockPackageRepository(deps.DB, m.logger)
	stockLotRepo := repository.NewStockLotRepository(deps.DB, m.logger)
	procurementGroupRepo := repository.NewProcurementGroupRepository(deps.DB, m.logger)
	stockRuleRepo := repository.NewStockRuleRepository(deps.DB, m.logger)
	stockPickingTypeRepo := repository.NewStockPickingTypeRepository(deps.DB, m.logger)
	stockPickingRepo := repository.NewStockPickingRepository(deps.DB, m.logger)
	stockMoveRepo := repository.NewStockMoveRepository(deps.DB, m.logger)

	// Get products repository from dependencies
	productsRepo, ok := deps.ProductRepo.(productsRepo.ProductRepo)
	if !ok {
		m.logger.Error("Failed to get products repository")
		return fmt.Errorf("products repository not available")
	}

	// Create services
	inventoryService := service.NewInventoryService(warehouseRepo, locationRepo, quantRepo, moveRepo)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	barcodeService := service.NewBarcodeService(barcodeRepo)
	cycleCountService := service.NewCycleCountService(cycleCountRepo)
	replenishmentService := service.NewReplenishmentService(replenishmentRuleRepo, replenishmentOrderRepo, inventoryService, productsRepo)
	batchOperationService := service.NewBatchOperationService(batchOperationRepo, batchOperationItemRepo, inventoryService, productsRepo)
	qualityControlService := service.NewQualityControlService(
		qcInspectionRepo, qcChecklistRepo, qcChecklistItemRepo, qcInspectionItemRepo, qcAlertRepo, inventoryService,
	)

	// New Inventory Services
	stockPackageService := service.NewStockPackageService(stockPackageRepo)
	stockLotService := service.NewStockLotService(stockLotRepo)
	procurementGroupService := service.NewProcurementGroupService(procurementGroupRepo)
	stockRuleService := service.NewStockRuleService(stockRuleRepo)
	stockPickingTypeService := service.NewStockPickingTypeService(stockPickingTypeRepo)
	stockPickingService := service.NewStockPickingService(stockPickingRepo)
	stockMoveService := service.NewStockMoveService(stockMoveRepo)

	// Create integration service for other modules
	m.integrationService = service.NewInventoryIntegrationService(stockMoveService, stockPickingService)

	// Create handlers
	m.inventoryHandler = handler.NewInventoryHandler(inventoryService)
	m.analyticsHandler = handler.NewAnalyticsHandler(analyticsService)
	m.barcodeHandler = handler.NewBarcodeHandler(barcodeService)
	m.cycleCountHandler = handler.NewCycleCountHandler(cycleCountService)
	m.replenishmentHandler = handler.NewReplenishmentHandler(replenishmentService)
	m.batchOperationHandler = handler.NewBatchOperationHandler(batchOperationService)
	m.qualityControlHandler = handler.NewQualityControlHandler(qualityControlService, deps.AuthService)

	// New Inventory Handlers
	m.stockPackageHandler = handler.NewStockPackageHandler(stockPackageService)
	m.stockLotHandler = handler.NewStockLotHandler(stockLotService)
	m.procurementGroupHandler = handler.NewProcurementGroupHandler(procurementGroupService)
	m.stockRuleHandler = handler.NewStockRuleHandler(stockRuleService)
	m.stockPickingTypeHandler = handler.NewStockPickingTypeHandler(stockPickingTypeService)
	m.stockPickingHandler = handler.NewStockPickingHandler(stockPickingService)
	m.stockMoveHandler = handler.NewStockMoveHandler(stockMoveService)

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
			if m.replenishmentHandler != nil {
				m.replenishmentHandler.RegisterRoutes(r)
			}
			if m.batchOperationHandler != nil {
				m.batchOperationHandler.RegisterRoutes(r)
			}
			if m.qualityControlHandler != nil {
				m.qualityControlHandler.RegisterRoutes(r)
			}
			if m.stockPackageHandler != nil {
				m.stockPackageHandler.RegisterRoutes(r)
			}
			if m.stockLotHandler != nil {
				m.stockLotHandler.RegisterRoutes(r)
			}
			if m.procurementGroupHandler != nil {
				m.procurementGroupHandler.RegisterRoutes(r)
			}
			if m.stockRuleHandler != nil {
				m.stockRuleHandler.RegisterRoutes(r)
			}
			if m.stockPickingTypeHandler != nil {
				m.stockPickingTypeHandler.RegisterRoutes(r)
			}
			if m.stockPickingHandler != nil {
				m.stockPickingHandler.RegisterRoutes(r)
			}
			if m.stockMoveHandler != nil {
				m.stockMoveHandler.RegisterRoutes(r)
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

// GetIntegrationService returns the integration service for use by other modules
func (m *InventoryModule) GetIntegrationService() *service.InventoryIntegrationService {
	return m.integrationService
}
