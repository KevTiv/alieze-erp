package crm

import (
	"context"
	"fmt"
	"log/slog"

	"alieze-erp/internal/modules/crm/handler"
	"alieze-erp/internal/modules/crm/repository"
	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/pkg/registry"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// CRMModule represents the CRM module
type CRMModule struct {
	contactHandler *handler.ContactHandler
	logger         *slog.Logger
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

	// Create services - using nil for auth service for now
	// TODO: Integrate with new auth/permission system
	// Create auth service adapter that uses the new policy engine
	authService := NewPolicyAuthServiceAdapter(deps.PolicyEngine)

	// Create services using the new auth service, rule engine, and event bus
	contactService := service.NewContactServiceWithDependencies(contactRepo, authService, deps.RuleEngine, deps.EventBus)

	// Create handlers
	m.contactHandler = handler.NewContactHandler(contactService)

	m.logger.Info("CRM module initialized successfully")
	return nil
}

// RegisterRoutes registers CRM module routes
func (m *CRMModule) RegisterRoutes(router interface{}) {
	if m.contactHandler != nil && router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			m.contactHandler.RegisterRoutes(r)
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
	// TODO: Update contact last activity, sales stats, etc.
	return nil
}

// handleOrderConfirmed handles order confirmation events
func (m *CRMModule) handleOrderConfirmed(ctx context.Context, event interface{}) error {
	m.logger.Info("Received order.confirmed event", "event", event)
	// TODO: Update contact as active customer
	return nil
}

// handleInvoiceCreated handles invoice creation events
func (m *CRMModule) handleInvoiceCreated(ctx context.Context, event interface{}) error {
	m.logger.Info("Received invoice.created event", "event", event)
	// TODO: Track customer invoicing activity
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
	// TODO: Implement organization ID retrieval from context
	// For now, return a default organization ID
	return uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000001")), nil
}

func (a *PolicyAuthServiceAdapter) GetUserID(ctx context.Context) (uuid.UUID, error) {
	// TODO: Implement user ID retrieval from context
	// For now, return a default user ID
	return uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000001")), nil
}

func (a *PolicyAuthServiceAdapter) CheckPermission(ctx context.Context, permission string) error {
	// Use the new policy engine
	if a.policyEngine != nil {
		if engine, ok := a.policyEngine.(interface{
			CheckPermission(ctx context.Context, subject, object, action string) (bool, error)
		}); ok {
			// Parse permission in format "resource:action"
			// For now, we'll use a simple approach
			allowed, err := engine.CheckPermission(ctx, "user", "contacts", permission)
			if err != nil {
				a.logger.Error("Permission check failed", "permission", permission, "error", err)
				return fmt.Errorf("permission check failed: %w", err)
			}
			if !allowed {
				return fmt.Errorf("permission denied: %s", permission)
			}
			return nil
		}
	}

	// Fallback: allow all permissions for now
	a.logger.Warn("No policy engine available, allowing all permissions")
	return nil
}
