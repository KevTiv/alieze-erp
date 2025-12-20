package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/types"
	"github.com/google/uuid"
)

// LegacyAuthService defines the legacy auth service interface used by older services
type LegacyAuthService interface {
	CheckPermission(ctx context.Context, permission string) error
	GetOrganizationID(ctx context.Context) (uuid.UUID, error)
	GetUserID(ctx context.Context) (uuid.UUID, error)
}

// BaseAuthService defines the base auth service interface for standardized services
type BaseAuthService interface {
	CheckOrganizationAccess(ctx context.Context, orgID uuid.UUID) error
	CheckUserPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error
	GetCurrentUser(ctx context.Context) (*types.User, error)
}

// PolicyAuthAdapter adapts the policy engine to both legacy and base auth service interfaces
type PolicyAuthAdapter struct {
	policyEngine interface{}
	ruleEngine   interface{}
	logger       *slog.Logger
}

// NewPolicyAuthAdapter creates a new auth service adapter
func NewPolicyAuthAdapter(policyEngine interface{}) *PolicyAuthAdapter {
	return &PolicyAuthAdapter{
		policyEngine: policyEngine,
		logger:       slog.Default().With("component", "policy-auth-adapter"),
	}
}

// NewPolicyAuthAdapterWithRules creates a new auth service adapter with rule engine integration
func NewPolicyAuthAdapterWithRules(policyEngine, ruleEngine interface{}) *PolicyAuthAdapter {
	return &PolicyAuthAdapter{
		policyEngine: policyEngine,
		ruleEngine:   ruleEngine,
		logger:       slog.Default().With("component", "policy-auth-adapter"),
	}
}

// LegacyAuthService implementation

func (a *PolicyAuthAdapter) GetOrganizationID(ctx context.Context) (uuid.UUID, error) {
	// Extract organization ID from context (set by auth middleware)
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		a.logger.Error("Organization ID not found in context")
		return uuid.Nil, fmt.Errorf("organization ID not found in context")
	}
	return orgID, nil
}

func (a *PolicyAuthAdapter) GetUserID(ctx context.Context) (uuid.UUID, error) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		a.logger.Error("User ID not found in context")
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

func (a *PolicyAuthAdapter) CheckPermission(ctx context.Context, permission string) error {
	// Get role from context (set by auth middleware)
	role, ok := ctx.Value("role").(string)
	if !ok {
		a.logger.Error("Role not found in context")
		return fmt.Errorf("role not found in context")
	}

	return a.checkPermissionWithRole(ctx, role, permission)
}

// BaseAuthService implementation

func (a *PolicyAuthAdapter) CheckOrganizationAccess(ctx context.Context, orgID uuid.UUID) error {
	// Extract organization ID from context (set by auth middleware)
	contextOrgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		a.logger.Error("Organization ID not found in context")
		return fmt.Errorf("organization ID not found in context")
	}

	// Verify that the requested organization matches the context organization
	if contextOrgID != orgID {
		a.logger.Warn("Organization access denied",
			"context_org_id", contextOrgID,
			"requested_org_id", orgID)
		return fmt.Errorf("access denied: organization mismatch")
	}

	return nil
}

func (a *PolicyAuthAdapter) CheckUserPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error {
	// First check organization access
	if err := a.CheckOrganizationAccess(ctx, orgID); err != nil {
		return err
	}

	// Get role from context
	role, ok := ctx.Value("role").(string)
	if !ok {
		a.logger.Error("Role not found in context")
		return fmt.Errorf("role not found in context")
	}

	return a.checkPermissionWithRole(ctx, role, permission)
}

func (a *PolicyAuthAdapter) GetCurrentUser(ctx context.Context) (*types.User, error) {
	// Extract user ID from context
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		a.logger.Error("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Extract organization ID from context
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		a.logger.Error("Organization ID not found in context")
		return nil, fmt.Errorf("organization ID not found in context")
	}

	// Extract role from context
	role, ok := ctx.Value("role").(string)
	if !ok {
		role = "user" // Default role
	}

	// Extract email if available
	email, _ := ctx.Value("email").(string)

	// Extract super admin flag
	isSuperAdmin, _ := ctx.Value("is_super_admin").(bool)

	// Return a user object from context
	// In a real implementation, this might query a user repository for full details
	return &types.User{
		ID:             userID,
		Email:          email,
		OrganizationID: orgID,
		Role:           role,
		IsSuperAdmin:   isSuperAdmin,
	}, nil
}

// Helper methods

func (a *PolicyAuthAdapter) checkPermissionWithRole(ctx context.Context, role, permission string) error {
	// Check if user is super admin first using rule engine
	if a.ruleEngine != nil {
		if ruleEngine, ok := a.ruleEngine.(interface {
			Evaluate(ctx context.Context, rule string, data interface{}) (bool, error)
		}); ok {
			// Check for super admin privilege
			isSuperAdmin, err := ruleEngine.Evaluate(ctx, "is_super_admin", map[string]interface{}{
				"role": role,
				"ctx":  ctx,
			})
			if err == nil && isSuperAdmin {
				a.logger.Debug("Permission granted to super admin",
					"role", role,
					"permission", permission)
				return nil
			}
		}
	}

	// Use the policy engine for RBAC check
	if a.policyEngine != nil {
		if engine, ok := a.policyEngine.(interface {
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
