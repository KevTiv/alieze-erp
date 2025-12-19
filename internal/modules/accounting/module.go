package accounting

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/KevTiv/alieze-erp/internal/modules/accounting/handler"
	"github.com/KevTiv/alieze-erp/internal/modules/accounting/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/accounting/service"
	"github.com/KevTiv/alieze-erp/pkg/registry"
	"github.com/KevTiv/alieze-erp/pkg/tax"
	"github.com/KevTiv/alieze-erp/pkg/workflow"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// AccountingModule represents the Accounting module
type AccountingModule struct {
	invoiceHandler *handler.InvoiceHandler
	paymentHandler *handler.PaymentHandler
	accountHandler *handler.AccountHandler
	journalHandler *handler.JournalHandler
	taxHandler     *handler.TaxHandler
	balanceHandler *handler.BalanceHandler
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
	accountRepo := repository.NewAccountRepository(deps.DB)
	journalRepo := repository.NewJournalRepository(deps.DB)
	taxRepo := repository.NewTaxRepository(deps.DB)

	// Create tax calculator
	taxCalc := tax.NewCalculator(deps.DB)

	// Create services with state machine and event bus support
	var invoiceStateMachine *workflow.StateMachine
	if deps.StateMachineFactory != nil {
		if sm, exists := deps.StateMachineFactory.GetStateMachine("accounting.invoice"); exists {
			invoiceStateMachine = sm
		}
	}
	invoiceService := service.NewInvoiceServiceWithDependencies(invoiceRepo, paymentRepo, taxCalc, invoiceStateMachine, deps.EventBus)
	paymentService := service.NewPaymentService(paymentRepo)
	accountService := service.NewAccountService(accountRepo)
	journalService := service.NewJournalService(journalRepo)
	taxService := service.NewTaxService(taxRepo)

	// Create handlers
	m.invoiceHandler = handler.NewInvoiceHandler(invoiceService)
	m.paymentHandler = handler.NewPaymentHandler(paymentService)
	m.accountHandler = handler.NewAccountHandler(accountService)
	m.journalHandler = handler.NewJournalHandler(journalService)
	m.taxHandler = handler.NewTaxHandler(taxService)
	m.balanceHandler = handler.NewBalanceHandler()

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
			if m.accountHandler != nil {
				m.accountHandler.RegisterRoutes(r)
			}
			if m.journalHandler != nil {
				m.journalHandler.RegisterRoutes(r)
			}
			if m.taxHandler != nil {
				m.taxHandler.RegisterRoutes(r)
			}
			if m.balanceHandler != nil {
				m.balanceHandler.RegisterRoutes(r)
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

	// Extract order data from event
	orderData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for order.confirmed")
		return fmt.Errorf("invalid event data format")
	}

	orderID, ok := orderData["id"].(uuid.UUID)
	if !ok {
		if orderIDStr, ok := orderData["id"].(string); ok {
			parsedID, err := uuid.Parse(orderIDStr)
			if err != nil {
				m.logger.Warn("Invalid order id in order.confirmed event", "id", orderIDStr)
				return err
			}
			orderID = parsedID
		}
	}

	customerID, ok := orderData["customer_id"].(uuid.UUID)
	if !ok {
		if customerIDStr, ok := orderData["customer_id"].(string); ok {
			parsedID, err := uuid.Parse(customerIDStr)
			if err != nil {
				m.logger.Warn("Invalid customer_id in order.confirmed event", "customer_id", customerIDStr)
				return err
			}
			customerID = parsedID
		}
	}

	orderReference, _ := orderData["reference"].(string)
	amountTotal, _ := orderData["amount_total"].(float64)

	m.logger.Info("Auto-generating invoice for confirmed order",
		"order_id", orderID,
		"customer_id", customerID,
		"reference", orderReference,
		"amount", amountTotal)

	// FUTURE ENHANCEMENT: Implement automatic invoice creation from confirmed orders
	// This would require implementing invoiceService.CreateFromOrder(ctx, orderID) which:
	// 1. Creates invoice with invoice_origin = orderReference
	// 2. Copies line items from sales_order_lines to invoice_lines
	// 3. Sets partner_id = customerID
	// 4. Sets amounts from order (amount_untaxed, amount_tax, amount_total)
	// 5. Publishes invoice.created event
	m.logger.Info("Invoice should be auto-generated",
		"order_reference", orderReference,
		"customer_id", customerID)

	return nil
}

// handleContactUpdated handles contact update events
func (m *AccountingModule) handleContactUpdated(ctx context.Context, event interface{}) error {
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

	// Note: Invoices reference contacts via partner_id foreign key,
	// so contact updates are automatically visible to invoices.
	// This handler exists for future enhancements:
	// - Cache invalidation when partner data changes
	// - Search index updates for invoice partner information
	// - Notifications about partner changes affecting unpaid invoices
	// - Audit logging for compliance (partner info changes on financial documents)

	return nil
}

// Health checks the health of the Accounting module
func (m *AccountingModule) Health() error {
	return nil
}

// getStateMachine helper function to retrieve state machines from dependencies
func (m *AccountingModule) getStateMachine(deps registry.Dependencies, workflowID string) (interface{}, bool) {
	if deps.StateMachineFactory != nil {
		if factory, ok := deps.StateMachineFactory.(interface {
			GetStateMachine(workflowID string) (interface{}, bool)
		}); ok {
			return factory.GetStateMachine(workflowID)
		}
	}
	return nil, false
}
