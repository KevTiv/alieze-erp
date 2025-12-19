package base

import (
	"context"
	"log/slog"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/types"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/google/uuid"
)

// Repository defines the standard interface for CRUD operations
type Repository[Entity any, Filter any] interface {
	Create(ctx context.Context, entity Entity) (*Entity, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Entity, error)
	FindAll(ctx context.Context, filter Filter) ([]*Entity, error)
	Update(ctx context.Context, entity Entity) (*Entity, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter Filter) (int, error)
}

// ExtendedRepository provides additional methods for organization-scoped operations
type ExtendedRepository[Entity any, Filter any] interface {
	Repository[Entity, Filter]
	FindByOrganization(ctx context.Context, orgID uuid.UUID, filter Filter) ([]*Entity, error)
	Exists(ctx context.Context, orgID, id uuid.UUID) (bool, error)
}

// AuthService defines the interface for authorization operations
type AuthService interface {
	CheckOrganizationAccess(ctx context.Context, orgID uuid.UUID) error
	CheckUserPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error
	GetCurrentUser(ctx context.Context) (*types.User, error)
}

// RuleEngine defines the interface for business rule validation
type RuleEngine interface {
	Validate(ctx context.Context, rule string, data interface{}) error
	Evaluate(ctx context.Context, rule string, data interface{}) (bool, error)
}

// ServiceOptions provides optional dependencies for services
type ServiceOptions struct {
	RuleEngine RuleEngine
	EventBus   *events.Bus
	Logger     *slog.Logger
}

// CRUDService provides a generic implementation of CRUD operations
type CRUDService[Entity any, Request any, UpdateRequest any, Filter any] struct {
	repo        Repository[Entity, Filter]
	authService AuthService
	ruleEngine  RuleEngine
	eventBus    *events.Bus
	logger      *slog.Logger
}

// NewCRUDService creates a new generic CRUD service
func NewCRUDService[Entity any, Request any, UpdateRequest any, Filter any](
	repo Repository[Entity, Filter],
	authService AuthService,
	opts ServiceOptions,
) *CRUDService[Entity, Request, UpdateRequest, Filter] {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &CRUDService[Entity, Request, UpdateRequest, Filter]{
		repo:        repo,
		authService: authService,
		ruleEngine:  opts.RuleEngine,
		eventBus:    opts.EventBus,
		logger:      logger,
	}
}

// GetRepository returns the underlying repository
func (s *CRUDService[Entity, Request, UpdateRequest, Filter]) GetRepository() Repository[Entity, Filter] {
	return s.repo
}

// GetAuthService returns the auth service
func (s *CRUDService[Entity, Request, UpdateRequest, Filter]) GetAuthService() AuthService {
	return s.authService
}

// GetLogger returns the logger
func (s *CRUDService[Entity, Request, UpdateRequest, Filter]) GetLogger() *slog.Logger {
	return s.logger
}

// PublishEvent publishes an event if the event bus is available
func (s *CRUDService[Entity, Request, UpdateRequest, Filter]) PublishEvent(ctx context.Context, eventType string, data interface{}) {
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, eventType, data)
	}
}

// ValidateWithRuleEngine validates data using the rule engine if available
func (s *CRUDService[Entity, Request, UpdateRequest, Filter]) ValidateWithRuleEngine(ctx context.Context, rule string, data interface{}) error {
	if s.ruleEngine != nil {
		return s.ruleEngine.Validate(ctx, rule, data)
	}
	return nil
}

// LogOperation logs a service operation with context
func (s *CRUDService[Entity, Request, UpdateRequest, Filter]) LogOperation(ctx context.Context, operation string, entityID uuid.UUID, details map[string]interface{}) {
	logData := map[string]any{
		"operation": operation,
		"entity_id": entityID,
	}

	// Merge details
	for k, v := range details {
		logData[k] = v
	}

	s.logger.InfoContext(ctx, "CRM operation", logData)
}
