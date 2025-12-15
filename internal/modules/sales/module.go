package sales

import (
	"context"
	"log/slog"

	"alieze-erp/internal/modules/sales/handler"
	"alieze-erp/internal/modules/sales/repository"
	"alieze-erp/internal/modules/sales/service"
	"alieze-erp/pkg/registry"
	"github.com/julienschmidt/httprouter"
)

// SalesModule represents the Sales module
type SalesModule struct {
	salesOrderHandler *handler.SalesOrderHandler
	pricelistHandler  *handler.PricelistHandler
	logger            *slog.Logger
}

// NewSalesModule creates a new Sales module
func NewSalesModule() *SalesModule {
	return &SalesModule{}
}

// Name returns the module name
func (m *SalesModule) Name() string {
	return "sales"
}

// Init initializes the Sales module
func (m *SalesModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "sales")
	m.logger.Info("Initializing Sales module")

	// Create repositories
	salesOrderRepo := repository.NewSalesOrderRepository(deps.DB)
	pricelistRepo := repository.NewPricelistRepository(deps.DB)

	// Create services with event bus support
	salesOrderService := service.NewSalesOrderServiceWithEventBus(salesOrderRepo, pricelistRepo, deps.EventBus)
	pricelistService := service.NewPricelistService(pricelistRepo)

	// Create handlers
	m.salesOrderHandler = handler.NewSalesOrderHandler(salesOrderService)
	m.pricelistHandler = handler.NewPricelistHandler(pricelistService)

	m.logger.Info("Sales module initialized successfully")
	return nil
}

// RegisterRoutes registers Sales module routes
func (m *SalesModule) RegisterRoutes(router interface{}) {
	if router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			if m.salesOrderHandler != nil {
				m.salesOrderHandler.RegisterRoutes(r)
			}
			if m.pricelistHandler != nil {
				m.pricelistHandler.RegisterRoutes(r)
			}
		}
	}
}

// RegisterEventHandlers registers event handlers for the Sales module
func (m *SalesModule) RegisterEventHandlers(bus interface{}) {
	if bus == nil {
		return
	}

	// Subscribe to relevant events from other modules
	if eventBus, ok := bus.(interface {
		Subscribe(eventType string, handler func(ctx context.Context, event interface{}) error)
	}); ok {
		// Listen to contact creation/updates to sync customer data
		eventBus.Subscribe("contact.created", m.handleContactCreated)
		eventBus.Subscribe("contact.updated", m.handleContactUpdated)

		// Listen to invoice events for order fulfillment tracking
		eventBus.Subscribe("invoice.paid", m.handleInvoicePaid)

		m.logger.Info("Sales module event handlers registered")
	}
}

// handleContactCreated handles contact creation events
func (m *SalesModule) handleContactCreated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received contact.created event", "event", event)
	// TODO: Implement customer sync logic
	return nil
}

// handleContactUpdated handles contact update events
func (m *SalesModule) handleContactUpdated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received contact.updated event", "event", event)
	// TODO: Implement customer update sync logic
	return nil
}

// handleInvoicePaid handles invoice paid events
func (m *SalesModule) handleInvoicePaid(ctx context.Context, event interface{}) error {
	m.logger.Info("Received invoice.paid event", "event", event)
	// TODO: Mark related orders as fully invoiced/paid
	return nil
}

// Health checks the health of the Sales module
func (m *SalesModule) Health() error {
	return nil
}
