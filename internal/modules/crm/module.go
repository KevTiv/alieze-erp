package crm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/handler"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/crm/base"
	"github.com/KevTiv/alieze-erp/pkg/registry"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// CRMModule represents the CRM module
type CRMModule struct {
	contactHandler        *handler.ContactHandler
	contactTagHandler     *handler.ContactTagHandler
	salesTeamHandler      *handler.SalesTeamHandler
	activityHandler       *handler.ActivityHandler
	leadStageHandler      *handler.LeadStageHandler
	leadSourceHandler     *handler.LeadSourceHandler
	lostReasonHandler     *handler.LostReasonHandler
	leadHandler           *handler.LeadHandler
	assignmentRuleHandler *handler.AssignmentRuleHandler
	logger                *slog.Logger
}

// NewCRMModule creates a new CRM module
func NewCRMModule() *CRMModule {
	return &CRMModule{}
}

// Name returns the module name
func (m *CRMModule) Name() string {
	return "crm"
}

// Init initializes the CRM module
func (m *CRMModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "crm")
	m.logger.Info("Initializing CRM module")

	// Create repositories
	contactRepo := repository.NewContactRepository(deps.DB)
	contactTagRepo := repository.NewContactTagRepository(deps.DB)
	salesTeamRepo := repository.NewSalesTeamRepository(deps.DB)
	activityRepo := repository.NewActivityRepository(deps.DB)
	leadStageRepo := repository.NewLeadStageRepository(deps.DB)
	leadSourceRepo := repository.NewLeadSourceRepository(deps.DB)
	lostReasonRepo := repository.NewLostReasonRepository(deps.DB)
	leadRepo := repository.NewLeadRepository(deps.DB)
	assignmentRuleRepo := repository.NewAssignmentRuleRepository(deps.DB)

	// Create services - using shared auth adapter with rule engine integration
	// The adapter implements both legacy and base auth service interfaces
	authAdapter := auth.NewPolicyAuthAdapterWithRules(deps.PolicyEngine, deps.RuleEngine)

	// Create services using the auth adapter, rule engine, and event bus
	contactService := service.NewContactServiceV2(contactRepo, authAdapter, base.ServiceOptions{
		Logger:     m.logger,
		RuleEngine: deps.RuleEngine,
		EventBus:   deps.EventBus,
	})
	contactTagService := service.NewContactTagService(contactTagRepo, authAdapter, deps.EventBus)
	salesTeamService := service.NewSalesTeamService(salesTeamRepo, authAdapter, deps.EventBus)
	activityService := service.NewActivityService(activityRepo, authAdapter, deps.EventBus)
	leadStageService := service.NewLeadStageService(leadStageRepo, authAdapter, deps.EventBus)
	leadSourceService := service.NewLeadSourceService(leadSourceRepo, authAdapter, deps.EventBus)
	lostReasonService := service.NewLostReasonService(lostReasonRepo, authAdapter, deps.EventBus)
	assignmentRuleService := service.NewAssignmentRuleService(assignmentRuleRepo, authAdapter, deps.EventBus)
	leadService := service.NewLeadService(leadRepo, authAdapter, deps.EventBus, assignmentRuleService)

	// Create handlers
	m.contactHandler = handler.NewContactHandler(contactService)
	m.contactTagHandler = handler.NewContactTagHandler(contactTagService)
	m.salesTeamHandler = handler.NewSalesTeamHandler(salesTeamService)
	m.activityHandler = handler.NewActivityHandler(activityService)
	m.leadStageHandler = handler.NewLeadStageHandler(leadStageService)
	m.leadSourceHandler = handler.NewLeadSourceHandler(leadSourceService)
	m.lostReasonHandler = handler.NewLostReasonHandler(lostReasonService)
	m.leadHandler = handler.NewLeadHandler(leadService)
	m.assignmentRuleHandler = handler.NewAssignmentRuleHandler(assignmentRuleService, authAdapter)

	m.logger.Info("CRM module initialized successfully")
	return nil
}

// RegisterRoutes registers CRM module routes
func (m *CRMModule) RegisterRoutes(router interface{}) {
	if router == nil {
		return
	}

	if r, ok := router.(*httprouter.Router); ok {
		if m.contactHandler != nil {
			m.contactHandler.RegisterRoutes(r)
		}
		if m.contactTagHandler != nil {
			m.contactTagHandler.RegisterRoutes(r)
		}
		if m.salesTeamHandler != nil {
			m.salesTeamHandler.RegisterRoutes(r)
		}
		if m.activityHandler != nil {
			m.activityHandler.RegisterRoutes(r)
		}
		if m.leadStageHandler != nil {
			m.leadStageHandler.RegisterRoutes(r)
		}
		if m.leadSourceHandler != nil {
			m.leadSourceHandler.RegisterRoutes(r)
		}
		if m.lostReasonHandler != nil {
			m.lostReasonHandler.RegisterRoutes(r)
		}
		if m.leadHandler != nil {
			m.leadHandler.RegisterRoutes(r)
		}
		if m.assignmentRuleHandler != nil {
			m.assignmentRuleHandler.RegisterRoutes(r)
		}
	}
}

// RegisterEventHandlers registers event handlers for the CRM module
func (m *CRMModule) RegisterEventHandlers(bus interface{}) {
	if bus == nil {
		return
	}

	// Subscribe to relevant events from other modules
	if eventBus, ok := bus.(interface {
		Subscribe(eventType string, handler func(ctx context.Context, event interface{}) error)
	}); ok {
		// Listen to order events to track customer activity
		eventBus.Subscribe("order.created", m.handleOrderCreated)
		eventBus.Subscribe("order.confirmed", m.handleOrderConfirmed)

		// Listen to invoice events to update contact payment status
		eventBus.Subscribe("invoice.created", m.handleInvoiceCreated)

		m.logger.Info("CRM module event handlers registered")
	}
}

// handleOrderCreated handles order creation events
func (m *CRMModule) handleOrderCreated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received order.created event", "event", event)

	// Extract order data from event
	// The event payload should contain the sales order
	orderData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for order.created")
		return fmt.Errorf("invalid event data format")
	}

	// Extract customer ID and order ID
	customerID, ok := orderData["customer_id"].(uuid.UUID)
	if !ok {
		// Try string conversion
		if customerIDStr, ok := orderData["customer_id"].(string); ok {
			parsedID, err := uuid.Parse(customerIDStr)
			if err != nil {
				m.logger.Warn("Invalid customer_id in order.created event", "customer_id", customerIDStr)
				return err
			}
			customerID = parsedID
		} else {
			m.logger.Warn("Missing customer_id in order.created event")
			return fmt.Errorf("missing customer_id in event")
		}
	}

	orderID, ok := orderData["id"].(uuid.UUID)
	if !ok {
		if orderIDStr, ok := orderData["id"].(string); ok {
			parsedID, err := uuid.Parse(orderIDStr)
			if err != nil {
				m.logger.Warn("Invalid order id in order.created event", "id", orderIDStr)
				return err
			}
			orderID = parsedID
		}
	}

	orgID, ok := orderData["organization_id"].(uuid.UUID)
	if !ok {
		if orgIDStr, ok := orderData["organization_id"].(string); ok {
			parsedID, err := uuid.Parse(orgIDStr)
			if err != nil {
				m.logger.Warn("Invalid organization_id in order.created event", "organization_id", orgIDStr)
				return err
			}
			orgID = parsedID
		}
	}

	// Create activity record for this order
	// Note: This requires activity repository which we'll add
	m.logger.Info("Creating activity for order",
		"customer_id", customerID,
		"order_id", orderID,
		"organization_id", orgID)

	// FUTURE ENHANCEMENT: Create activity record once ActivityRepository is implemented
	// activityRepo.Create(ctx, Activity{
	//     OrganizationID: orgID,
	//     ActivityType: "note",
	//     Summary: fmt.Sprintf("Sales order %s created", orderData["reference"]),
	//     ResModel: "sales.order",
	//     ResID: orderID,
	//     CreatedAt: time.Now(),
	// })
	// This will enable tracking customer engagement history in the CRM

	return nil
}

// handleOrderConfirmed handles order confirmation events
func (m *CRMModule) handleOrderConfirmed(ctx context.Context, event interface{}) error {
	m.logger.Info("Received order.confirmed event", "event", event)

	// Extract order data from event
	orderData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for order.confirmed")
		return fmt.Errorf("invalid event data format")
	}

	// Extract customer ID
	customerID, ok := orderData["customer_id"].(uuid.UUID)
	if !ok {
		if customerIDStr, ok := orderData["customer_id"].(string); ok {
			parsedID, err := uuid.Parse(customerIDStr)
			if err != nil {
				m.logger.Warn("Invalid customer_id in order.confirmed event", "customer_id", customerIDStr)
				return err
			}
			customerID = parsedID
		} else {
			m.logger.Warn("Missing customer_id in order.confirmed event")
			return fmt.Errorf("missing customer_id in event")
		}
	}

	// Update contact to mark as active customer
	// We need to access the contact repository through the handler
	if m.contactHandler != nil {
		m.logger.Info("Marking contact as active customer", "customer_id", customerID)

		// FUTURE ENHANCEMENT: Add MarkAsCustomer method to contact service
		// This would execute: UPDATE contacts SET is_customer = true WHERE id = $1
		// Implementation: contactService.MarkAsCustomer(ctx, customerID)
		m.logger.Info("Contact should be marked as customer", "contact_id", customerID)
	}

	return nil
}

// handleInvoiceCreated handles invoice creation events
func (m *CRMModule) handleInvoiceCreated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received invoice.created event", "event", event)

	// Extract invoice data from event
	invoiceData, ok := event.(map[string]interface{})
	if !ok {
		m.logger.Warn("Invalid event data format for invoice.created")
		return fmt.Errorf("invalid event data format")
	}

	// Extract partner ID (customer)
	partnerID, ok := invoiceData["partner_id"].(uuid.UUID)
	if !ok {
		if partnerIDStr, ok := invoiceData["partner_id"].(string); ok {
			parsedID, err := uuid.Parse(partnerIDStr)
			if err != nil {
				m.logger.Warn("Invalid partner_id in invoice.created event", "partner_id", partnerIDStr)
				return err
			}
			partnerID = parsedID
		} else {
			m.logger.Warn("Missing partner_id in invoice.created event")
			return fmt.Errorf("missing partner_id in event")
		}
	}

	invoiceID, ok := invoiceData["id"].(uuid.UUID)
	if !ok {
		if invoiceIDStr, ok := invoiceData["id"].(string); ok {
			parsedID, err := uuid.Parse(invoiceIDStr)
			if err != nil {
				m.logger.Warn("Invalid invoice id in invoice.created event", "id", invoiceIDStr)
				return err
			}
			invoiceID = parsedID
		}
	}

	orgID, ok := invoiceData["organization_id"].(uuid.UUID)
	if !ok {
		if orgIDStr, ok := invoiceData["organization_id"].(string); ok {
			parsedID, err := uuid.Parse(orgIDStr)
			if err != nil {
				m.logger.Warn("Invalid organization_id in invoice.created event", "organization_id", orgIDStr)
				return err
			}
			orgID = parsedID
		}
	}

	m.logger.Info("Tracking customer invoicing activity",
		"partner_id", partnerID,
		"invoice_id", invoiceID,
		"organization_id", orgID)

	// FUTURE ENHANCEMENT: Create activity record once ActivityRepository is implemented
	// activityRepo.Create(ctx, Activity{
	//     OrganizationID: orgID,
	//     ActivityType: "note",
	//     Summary: fmt.Sprintf("Invoice %s created", invoiceData["reference"]),
	//     ResModel: "account.invoice",
	//     ResID: invoiceID,
	//     CreatedAt: time.Now(),
	// })
	// This will enable tracking customer invoicing activity in the CRM

	return nil
}

// Health checks the health of the CRM module
func (m *CRMModule) Health() error {
	return nil
}
