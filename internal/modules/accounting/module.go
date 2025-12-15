package accounting

import (
	"context"
	"log/slog"

	"alieze-erp/internal/modules/accounting/handler"
	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/modules/accounting/service"
	"alieze-erp/pkg/registry"
	"github.com/julienschmidt/httprouter"
)

// AccountingModule represents the Accounting module
type AccountingModule struct {
	invoiceHandler *handler.InvoiceHandler
	paymentHandler *handler.PaymentHandler
	logger         *slog.Logger
}

// NewAccountingModule creates a new Accounting module
func NewAccountingModule() *AccountingModule {
	return &AccountingModule{}
}

// Name returns the module name
func (m *AccountingModule) Name() string {
	return "accounting"
}

// Init initializes the Accounting module
func (m *AccountingModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "accounting")
	m.logger.Info("Initializing Accounting module")

	// Create repositories
	invoiceRepo := repository.NewInvoiceRepository(deps.DB)
	paymentRepo := repository.NewPaymentRepository(deps.DB)

	// Create services with state machine and event bus support
	invoiceStateMachine, _ := m.getStateMachine(deps, "accounting.invoice")
	invoiceService := service.NewInvoiceServiceWithDependencies(invoiceRepo, paymentRepo, invoiceStateMachine, deps.EventBus)
	paymentService := service.NewPaymentService(paymentRepo)

	// Create handlers
	m.invoiceHandler = handler.NewInvoiceHandler(invoiceService)
	m.paymentHandler = handler.NewPaymentHandler(paymentService)

	m.logger.Info("Accounting module initialized successfully")
	return nil
}

// RegisterRoutes registers Accounting module routes
func (m *AccountingModule) RegisterRoutes(router interface{}) {
	if router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			if m.invoiceHandler != nil {
				m.invoiceHandler.RegisterRoutes(r)
			}
			if m.paymentHandler != nil {
				m.paymentHandler.RegisterRoutes(r)
			}
		}
	}
}

// RegisterEventHandlers registers event handlers for the Accounting module
func (m *AccountingModule) RegisterEventHandlers(bus interface{}) {
	if bus == nil {
		return
	}

	// Subscribe to relevant events from other modules
	if eventBus, ok := bus.(interface {
		Subscribe(eventType string, handler func(ctx context.Context, event interface{}) error)
	}); ok {
		// Listen to order confirmation events to potentially create invoices
		eventBus.Subscribe("order.confirmed", m.handleOrderConfirmed)

		// Listen to contact updates to sync partner information
		eventBus.Subscribe("contact.updated", m.handleContactUpdated)

		m.logger.Info("Accounting module event handlers registered")
	}
}

// handleOrderConfirmed handles order confirmation events
func (m *AccountingModule) handleOrderConfirmed(ctx context.Context, event interface{}) error {
	m.logger.Info("Received order.confirmed event", "event", event)
	// TODO: Implement invoice generation logic when order is confirmed
	return nil
}

// handleContactUpdated handles contact update events
func (m *AccountingModule) handleContactUpdated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received contact.updated event", "event", event)
	// TODO: Implement partner sync logic
	return nil
}

// Health checks the health of the Accounting module
func (m *AccountingModule) Health() error {
	return nil
}

// getStateMachine helper function to retrieve state machines from dependencies
func (m *AccountingModule) getStateMachine(deps registry.Dependencies, workflowID string) (interface{}, bool) {
	if deps.StateMachineFactory != nil {
		if factory, ok := deps.StateMachineFactory.(interface{
			GetStateMachine(workflowID string) (interface{}, bool)
		}); ok {
			return factory.GetStateMachine(workflowID)
		}
	}
	return nil, false
}
