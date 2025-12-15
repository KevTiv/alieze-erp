package auth

import (
	"context"
	"log/slog"

	"alieze-erp/internal/modules/auth/handler"
	"alieze-erp/internal/modules/auth/middleware"
	"alieze-erp/internal/modules/auth/repository"
	"alieze-erp/internal/modules/auth/service"
	"alieze-erp/pkg/registry"
	"github.com/julienschmidt/httprouter"
)

// AuthModule represents the Auth module
type AuthModule struct {
	authHandler    *handler.AuthHandler
	authMiddleware *middleware.AuthMiddleware
	logger         *slog.Logger
}

// NewAuthModule creates a new Auth module
func NewAuthModule() *AuthModule {
	return &AuthModule{}
}

// Name returns the module name
func (m *AuthModule) Name() string {
	return "auth"
}

// Init initializes the Auth module
func (m *AuthModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "auth")
	m.logger.Info("Initializing Auth module")

	// Create repositories
	authRepo := repository.NewAuthRepository(deps.DB)

	// Create services
	authService := service.NewAuthService(authRepo)

	// Create handlers
	m.authHandler = handler.NewAuthHandler(authService)
	m.authMiddleware = middleware.NewAuthMiddleware()

	m.logger.Info("Auth module initialized successfully")
	return nil
}

// RegisterRoutes registers Auth module routes
func (m *AuthModule) RegisterRoutes(router interface{}) {
	if m.authHandler != nil && router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			m.authHandler.RegisterRoutes(r)
		}
	}
}

// RegisterEventHandlers registers event handlers for the Auth module
func (m *AuthModule) RegisterEventHandlers(bus interface{}) {
	// TODO: Implement event handlers when event system is integrated
}

// Health checks the health of the Auth module
func (m *AuthModule) Health() error {
	return nil
}

// GetMiddleware returns the auth middleware for use in the server
func (m *AuthModule) GetMiddleware() *middleware.AuthMiddleware {
	return m.authMiddleware
}
