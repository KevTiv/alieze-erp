package sales

import (
	"context"
	"fmt"
	"log/slog"

	"alieze-erp/internal/modules/sales/handler"
	"alieze-erp/internal/modules/sales/repository"
	"alieze-erp/internal/modules/sales/service"
	"alieze-erp/pkg/registry"
	"github.com/google/uuid"
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

	// Extract contact data from event
	contactData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for contact.created")
		return fmt.Errorf("invalid event data format")
	}

	contactID, ok := contactData["id"].(uuid.UUID)
	if !ok {
		if contactIDStr, ok := contactData["id"].(string); ok {
			parsedID, err := uuid.Parse(contactIDStr)
			if err != nil {
				m.logger.Warn("Invalid contact id in contact.created event", "id", contactIDStr)
				return err
			}
			contactID = parsedID
		}
	}

	contactName, _ := contactData["name"].(string)
	isCustomer, _ := contactData["is_customer"].(bool)

	m.logger.Info("Contact created notification received",
		"contact_id", contactID,
		"name", contactName,
		"is_customer", isCustomer)

	// Note: Sales orders reference contacts via customer_id foreign key,
	// so contact data is automatically accessible without sync.
	// This handler exists for future enhancements:
	// - Cache invalidation (Redis/in-memory caches)
	// - Search index updates (Elasticsearch)
	// - Notification triggers (alert sales team of new customers)
	// - Audit logging (track which contacts affect which orders)

	return nil
}

// handleContactUpdated handles contact update events
func (m *SalesModule) handleContactUpdated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received contact.updated event", "event", event)

	// Extract contact data from event
	contactData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for contact.updated")
		return fmt.Errorf("invalid event data format")
	}

	contactID, ok := contactData["id"].(uuid.UUID)
	if !ok {
		if contactIDStr, ok := contactData["id"].(string); ok {
			parsedID, err := uuid.Parse(contactIDStr)
			if err != nil {
				m.logger.Warn("Invalid contact id in contact.updated event", "id", contactIDStr)
				return err
			}
			contactID = parsedID
		}
	}

	contactName, _ := contactData["name"].(string)

	m.logger.Info("Contact updated notification received",
		"contact_id", contactID,
		"name", contactName)

	// Note: Sales orders reference contacts via customer_id foreign key,
	// so contact updates are automatically visible to sales orders.
	// This handler exists for future enhancements:
	// - Cache invalidation when customer data changes
	// - Search index updates
	// - Notifications to sales reps about customer info changes
	// - Triggering re-validation of orders if critical customer data changes

	return nil
}

// handleInvoicePaid handles invoice paid events
func (m *SalesModule) handleInvoicePaid(ctx context.Context, event interface{}) error {
	m.logger.Info("Received invoice.paid event", "event", event)

	// Extract invoice data from event
	invoiceData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for invoice.paid")
		return fmt.Errorf("invalid event data format")
	}

	invoiceID, ok := invoiceData["id"].(uuid.UUID)
	if !ok {
		if invoiceIDStr, ok := invoiceData["id"].(string); ok {
			parsedID, err := uuid.Parse(invoiceIDStr)
			if err != nil {
				m.logger.Warn("Invalid invoice id in invoice.paid event", "id", invoiceIDStr)
				return err
			}
			invoiceID = parsedID
		}
	}

	// Check if invoice has an origin (sales order reference)
	invoiceOrigin, _ := invoiceData["invoice_origin"].(string)

	m.logger.Info("Processing paid invoice",
		"invoice_id", invoiceID,
		"invoice_origin", invoiceOrigin)

	// FUTURE ENHANCEMENT: Update related sales order payment status
	// This would require:
	// 1. Parse invoice_origin to extract order reference
	// 2. Find the sales order by reference
	// 3. Update the order's payment status or invoice_status field
	// Implementation: salesOrderService.UpdateInvoiceStatus(ctx, orderReference, "invoiced")
	if invoiceOrigin != "" {
		m.logger.Info("Invoice paid for order", "order_reference", invoiceOrigin)
	}

	return nil
}

// Health checks the health of the Sales module
func (m *SalesModule) Health() error {
	return nil
}
