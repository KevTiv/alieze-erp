package server

import (
	"context"
	"fmt"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/service"

	"github.com/google/uuid"
)

// SimpleAuthServiceAdapter adapts the auth service to the simpler interface expected by CRM and Products services
type SimpleAuthServiceAdapter struct {
	authService *service.AuthService
}

func NewSimpleAuthServiceAdapter(authService *service.AuthService) *SimpleAuthServiceAdapter {
	return &SimpleAuthServiceAdapter{authService: authService}
}

func (a *SimpleAuthServiceAdapter) GetOrganizationID(ctx context.Context) (uuid.UUID, error) {
	// This would normally come from the context set by auth middleware
	// For now, return a dummy UUID
	return uuid.New(), nil
}

func (a *SimpleAuthServiceAdapter) GetUserID(ctx context.Context) (uuid.UUID, error) {
	// This would normally come from the context set by auth middleware
	// For now, return a dummy UUID
	return uuid.New(), nil
}

func (a *SimpleAuthServiceAdapter) CheckPermission(ctx context.Context, permission string) error {
	// Get user and org IDs (would normally come from context)
	userID, err := a.GetUserID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	orgID, err := a.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Call the actual auth service with the full signature
	return a.authService.CheckPermission(ctx, userID, orgID, permission)
}
