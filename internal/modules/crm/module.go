package crm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"alieze-erp/internal/modules/crm/handler"
	"alieze-erp/internal/modules/crm/repository"
	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/pkg/events"
	"alieze-erp/pkg/registry"
	"alieze-erp/pkg/rules"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// CRMModule represents the CRM module
type CRMModule struct {
	contactHandler    *handler.ContactHandler
	contactTagHandler *handler.ContactTagHandler
	salesTeamHandler  *handler.SalesTeamHandler
	activityHandler   *handler.ActivityHandler
	leadStageHandler  *handler.LeadStageHandler
	leadSourceHandler *handler.LeadSourceHandler
	lostReasonHandler *handler.LostReasonHandler
	leadHandler       *handler.LeadHandler
	assignmentRuleHandler *handler.AssignmentRuleHandler
	logger            *slog.Logger
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

	// Create services - using nil for auth service for now
	// TODO: Integrate with new auth/permission system
	// Create auth service adapter that uses the new policy engine
	authService := NewPolicyAuthServiceAdapter(deps.PolicyEngine)

	// Create services using the new auth service, rule engine, and event bus
	contactService := service.NewContactServiceWithDependencies(contactRepo, authService, deps.RuleEngine, deps.EventBus)

	// Cast event bus to concrete type for ContactTagService
	eventBus, ok := deps.EventBus.(*events.Bus)
	if !ok {
		m.logger.Error("Failed to cast event bus to *events.Bus")
		return fmt.Errorf("invalid event bus type")
	}
	contactTagService := service.NewContactTagService(contactTagRepo, authService, eventBus)
	salesTeamService := service.NewSalesTeamService(salesTeamRepo, authService, eventBus)
	activityService := service.NewActivityService(activityRepo, authService, eventBus)
	leadStageService := service.NewLeadStageService(leadStageRepo, authService, eventBus)
	leadSourceService := service.NewLeadSourceService(leadSourceRepo, authService, eventBus)
	lostReasonService := service.NewLostReasonService(lostReasonRepo, authService, eventBus)
	leadService := service.NewLeadService(service.NewLeadServiceOptions{
		LeadRepository: leadRepo,
		RuleEngine:     deps.RuleEngine.(*rules.RuleEngine),
		Logger:         m.logger,
	})
	assignmentRuleService := service.NewAssignmentRuleService(assignmentRuleRepo, authService)

	// Create handlers
	m.contactHandler = handler.NewContactHandler(contactService)
	m.contactTagHandler = handler.NewContactTagHandler(contactTagService)
	m.salesTeamHandler = handler.NewSalesTeamHandler(salesTeamService)
	m.activityHandler = handler.NewActivityHandler(activityService)
	m.leadStageHandler = handler.NewLeadStageHandler(leadStageService)
	m.leadSourceHandler = handler.NewLeadSourceHandler(leadSourceService)
	m.lostReasonHandler = handler.NewLostReasonHandler(lostReasonService)
	m.leadHandler = handler.NewLeadHandler(leadService)
	m.assignmentRuleHandler = handler.NewAssignmentRuleHandler(assignmentRuleService, authService)

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

// PolicyAuthServiceAdapter adapts the new policy engine to the existing auth service interface
type PolicyAuthServiceAdapter struct {
	policyEngine interface{}
	logger       *slog.Logger
}

func NewPolicyAuthServiceAdapter(policyEngine interface{}) *PolicyAuthServiceAdapter {
	return &PolicyAuthServiceAdapter{
		policyEngine: policyEngine,
		logger:       slog.Default().With("component", "policy-auth-adapter"),
	}
}

func (a *PolicyAuthServiceAdapter) GetOrganizationID(ctx context.Context) (uuid.UUID, error) {
	// Extract organization ID from context (set by auth middleware)
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		a.logger.Error("Organization ID not found in context")
		return uuid.Nil, fmt.Errorf("organization ID not found in context")
	}
	return orgID, nil
}

func (a *PolicyAuthServiceAdapter) GetUserID(ctx context.Context) (uuid.UUID, error) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		a.logger.Error("User ID not found in context")
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

func (a *PolicyAuthServiceAdapter) CheckPermission(ctx context.Context, permission string) error {
	// Get role from context (set by auth middleware)
	role, ok := ctx.Value("role").(string)
	if !ok {
		a.logger.Error("Role not found in context")
		return fmt.Errorf("role not found in context")
	}

	// Use the policy engine for RBAC check
	if a.policyEngine != nil {
		if engine, ok := a.policyEngine.(interface{
			CheckPermission(ctx context.Context, subject, object, action string) (bool, error)
		}); ok {
			// Parse permission in format "action" (e.g., "contacts:create" or just "create")
			// Extract resource and action
			resource := "contacts"
			action := permission

			// If permission contains ':', split it
			if strings.Contains(permission, ":") {
				parts := strings.SplitN(permission, ":", 2)
				resource = parts[0]
				action = parts[1]
			}

			// Check permission using role:roleName format for Casbin
			subject := fmt.Sprintf("role:%s", role)
			allowed, err := engine.CheckPermission(ctx, subject, resource, action)
			if err != nil {
				a.logger.Error("Permission check failed",
					"role", role,
					"resource", resource,
					"action", action,
					"error", err)
				return fmt.Errorf("permission check failed: %w", err)
			}
			if !allowed {
				a.logger.Warn("Permission denied",
					"role", role,
					"resource", resource,
					"action", action)
				return fmt.Errorf("permission denied: user with role '%s' cannot '%s' on '%s'", role, action, resource)
			}
			return nil
		}
	}

	// No policy engine available - this should not happen in production
	a.logger.Error("No policy engine available - denying all permissions")
	return fmt.Errorf("policy engine not configured")
}
