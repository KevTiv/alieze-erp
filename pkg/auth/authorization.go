package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

// PolicyEngine defines the interface for policy enforcement
type PolicyEngine interface {
	// CheckPermission checks if a subject has permission to perform an action on an object
	CheckPermission(ctx context.Context, subject, object, action string) (bool, error)

	// GetRolesForUser gets all roles for a user
	GetRolesForUser(user string) ([]string, error)

	// GetPermissionsForUser gets all permissions for a user
	GetPermissionsForUser(user string) ([][]string, error)
}

// AuthorizationService handles authorization and permission checks
type AuthorizationService interface {
	// CheckPermission checks if a user has a specific permission
	CheckPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error

	// CheckResourceAccess checks if a user can access a specific resource
	CheckResourceAccess(ctx context.Context, userID, orgID uuid.UUID, resourceType string, resourceID uuid.UUID) error

	// HasOrganizationAccess checks if a user has access to an organization
	HasOrganizationAccess(ctx context.Context, userID, orgID uuid.UUID) bool

	// GetUserRoles returns all roles for a user in an organization
	GetUserRoles(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)

	// GetUserPermissions returns all permissions for a user in an organization
	GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
}

// BasicAuthorizationService provides a robust authorization implementation using PolicyEngine
type BasicAuthorizationService struct {
	policyEngine PolicyEngine
	logger       *slog.Logger
}

// NewBasicAuthorizationService creates a new basic authorization service
func NewBasicAuthorizationService(policyEngine PolicyEngine) *BasicAuthorizationService {
	return &BasicAuthorizationService{
		policyEngine: policyEngine,
		logger:       slog.Default().With("component", "basic-auth-service"),
	}
}

// CheckPermission checks if a user has a specific permission
func (a *BasicAuthorizationService) CheckPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error {
	if a.policyEngine == nil {
		a.logger.Error("Policy engine not configured")
		return fmt.Errorf("policy engine not configured")
	}

	// Parse permission in format "resource:action" (e.g., "contacts:create")
	resource := "contacts"
	action := permission

	// If permission contains ':', split it
	if parts := strings.SplitN(permission, ":", 2); len(parts) == 2 {
		resource = parts[0]
		action = parts[1]
	}

	// Create subject identifier: user:userID:orgID
	subject := fmt.Sprintf("user:%s:org:%s", userID.String(), orgID.String())

	// Check permission using policy engine
	allowed, err := a.policyEngine.CheckPermission(ctx, subject, resource, action)
	if err != nil {
		a.logger.Error("Permission check failed",
			"user_id", userID,
			"org_id", orgID,
			"resource", resource,
			"action", action,
			"error", err)
		return fmt.Errorf("permission check failed: %w", err)
	}

	if !allowed {
		a.logger.Warn("Permission denied",
			"user_id", userID,
			"org_id", orgID,
			"resource", resource,
			"action", action)
		return fmt.Errorf("permission denied: user '%s' cannot '%s' on '%s'", userID, action, resource)
	}

	a.logger.Debug("Permission granted",
		"user_id", userID,
		"org_id", orgID,
		"resource", resource,
		"action", action)

	return nil
}

// CheckResourceAccess checks if a user can access a specific resource
func (a *BasicAuthorizationService) CheckResourceAccess(ctx context.Context, userID, orgID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	if a.policyEngine == nil {
		a.logger.Error("Policy engine not configured")
		return fmt.Errorf("policy engine not configured")
	}

	// Create subject identifier
	subject := fmt.Sprintf("user:%s:org:%s", userID.String(), orgID.String())

	// Create object identifier
	object := fmt.Sprintf("%s:%s", resourceType, resourceID.String())

	// Check access using policy engine
	allowed, err := a.policyEngine.CheckPermission(ctx, subject, object, "access")
	if err != nil {
		a.logger.Error("Resource access check failed",
			"user_id", userID,
			"org_id", orgID,
			"resource_type", resourceType,
			"resource_id", resourceID,
			"error", err)
		return fmt.Errorf("resource access check failed: %w", err)
	}

	if !allowed {
		a.logger.Warn("Resource access denied",
			"user_id", userID,
			"org_id", orgID,
			"resource_type", resourceType,
			"resource_id", resourceID)
		return fmt.Errorf("access denied: user '%s' cannot access '%s' resource '%s'", userID, resourceType, resourceID)
	}

	a.logger.Debug("Resource access granted",
		"user_id", userID,
		"org_id", orgID,
		"resource_type", resourceType,
		"resource_id", resourceID)

	return nil
}

// HasOrganizationAccess checks if a user has access to an organization
func (a *BasicAuthorizationService) HasOrganizationAccess(ctx context.Context, userID, orgID uuid.UUID) bool {
	if a.policyEngine == nil {
		a.logger.Error("Policy engine not configured")
		return false
	}

	// Create subject identifier
	subject := fmt.Sprintf("user:%s", userID.String())

	// Create organization object identifier
	object := fmt.Sprintf("org:%s", orgID.String())

	// Check organization access
	allowed, err := a.policyEngine.CheckPermission(ctx, subject, object, "access")
	if err != nil {
		a.logger.Error("Organization access check failed",
			"user_id", userID,
			"org_id", orgID,
			"error", err)
		return false
	}

	if !allowed {
		a.logger.Warn("Organization access denied",
			"user_id", userID,
			"org_id", orgID)
		return false
	}

	a.logger.Debug("Organization access granted",
		"user_id", userID,
		"org_id", orgID)

	return true
}

// GetUserRoles returns all roles for a user in an organization
func (a *BasicAuthorizationService) GetUserRoles(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	if a.policyEngine == nil {
		a.logger.Error("Policy engine not configured")
		return nil, fmt.Errorf("policy engine not configured")
	}

	// Create subject identifier
	subject := fmt.Sprintf("user:%s:org:%s", userID.String(), orgID.String())

	// Get roles for user
	roles, err := a.policyEngine.GetRolesForUser(subject)
	if err != nil {
		a.logger.Error("Failed to get user roles",
			"user_id", userID,
			"org_id", orgID,
			"error", err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		a.logger.Warn("No roles found for user",
			"user_id", userID,
			"org_id", orgID)
		return []string{"user"}, nil // Default role
	}

	a.logger.Debug("Retrieved user roles",
		"user_id", userID,
		"org_id", orgID,
		"roles", roles)

	return roles, nil
}

// GetUserPermissions returns all permissions for a user in an organization
func (a *BasicAuthorizationService) GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	if a.policyEngine == nil {
		a.logger.Error("Policy engine not configured")
		return nil, fmt.Errorf("policy engine not configured")
	}

	// Create subject identifier
	subject := fmt.Sprintf("user:%s:org:%s", userID.String(), orgID.String())

	// Get permissions for user
	permissions, err := a.policyEngine.GetPermissionsForUser(subject)
	if err != nil {
		a.logger.Error("Failed to get user permissions",
			"user_id", userID,
			"org_id", orgID,
			"error", err)
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Convert permissions to string format
	var result []string
	for _, perm := range permissions {
		if len(perm) >= 3 {
			result = append(result, fmt.Sprintf("%s:%s", perm[1], perm[2]))
		}
	}

	if len(result) == 0 {
		a.logger.Warn("No permissions found for user",
			"user_id", userID,
			"org_id", orgID)
		return []string{"read", "write"}, nil // Default permissions
	}

	a.logger.Debug("Retrieved user permissions",
		"user_id", userID,
		"org_id", orgID,
		"permissions", result)

	return result, nil
}
