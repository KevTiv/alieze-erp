package policy

import (
	"context"
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	pgxadapter "github.com/pckhoi/casbin-pgx-adapter/v2"
)

// CasbinEnforcer is a wrapper around Casbin's enforcer
type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
	adapter  *pgxadapter.Adapter
	mockMode bool
}

// NewCasbinEnforcer creates a new Casbin enforcer with PostgreSQL adapter
func NewCasbinEnforcer(connString string, modelPath string) (*CasbinEnforcer, error) {
	// If no connection string provided, use mock mode
	if connString == "" {
		log.Println("[Policy] No connection string provided, using mock mode")
		return &CasbinEnforcer{
			mockMode: true,
		}, nil
	}

	// Create adapter from connection string
	adapter, err := pgxadapter.NewAdapter(connString,
		pgxadapter.WithTableName("casbin_rules"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// Load model
	var m model.Model
	if modelPath != "" {
		m, err = model.NewModelFromFile(modelPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load model from file: %w", err)
		}
	} else {
		// Use default RBAC model
		m, err = model.NewModelFromString(defaultRBACModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create default model: %w", err)
		}
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Load policies from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	return &CasbinEnforcer{
		enforcer: enforcer,
		adapter:  adapter,
		mockMode: false,
	}, nil
}

// CheckPermission checks if a subject has permission using Casbin
func (ce *CasbinEnforcer) CheckPermission(ctx context.Context, subject, object, action string) (bool, error) {
	if ce.mockMode {
		// Mock implementation - allow all for now
		log.Printf("[Policy Mock] Allowing permission: subject=%s, object=%s, action=%s", subject, object, action)
		return true, nil
	}

	// Use Casbin's Enforce method
	ok, err := ce.enforcer.Enforce(subject, object, action)
	if err != nil {
		return false, fmt.Errorf("casbin enforce error: %w", err)
	}

	return ok, nil
}

// AddPolicy adds a policy rule (p policy)
func (ce *CasbinEnforcer) AddPolicy(subject, object, action string) error {
	if ce.mockMode {
		log.Printf("[Policy Mock] Would add policy: %s, %s, %s", subject, object, action)
		return nil
	}

	_, err := ce.enforcer.AddPolicy(subject, object, action)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	return ce.enforcer.SavePolicy()
}

// AddGroupingPolicy adds a role inheritance rule (g policy)
func (ce *CasbinEnforcer) AddGroupingPolicy(user, role string) error {
	if ce.mockMode {
		log.Printf("[Policy Mock] Would add grouping policy: %s -> %s", user, role)
		return nil
	}

	_, err := ce.enforcer.AddGroupingPolicy(user, role)
	if err != nil {
		return fmt.Errorf("failed to add grouping policy: %w", err)
	}

	return ce.enforcer.SavePolicy()
}

// RemovePolicy removes a policy rule
func (ce *CasbinEnforcer) RemovePolicy(subject, object, action string) error {
	if ce.mockMode {
		log.Printf("[Policy Mock] Would remove policy: %s, %s, %s", subject, object, action)
		return nil
	}

	_, err := ce.enforcer.RemovePolicy(subject, object, action)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	return ce.enforcer.SavePolicy()
}

// GetRolesForUser gets all roles for a user
func (ce *CasbinEnforcer) GetRolesForUser(user string) ([]string, error) {
	if ce.mockMode {
		return []string{"admin"}, nil // Mock: everyone is admin
	}

	return ce.enforcer.GetRolesForUser(user)
}

// GetPermissionsForUser gets all permissions for a user
func (ce *CasbinEnforcer) GetPermissionsForUser(user string) ([][]string, error) {
	if ce.mockMode {
		return [][]string{}, nil
	}

	return ce.enforcer.GetPermissionsForUser(user)
}

// Default RBAC model following Casbin conventions
const defaultRBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
