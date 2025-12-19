package testutils

import (
	"context"
	"fmt"

	authTypes "github.com/KevTiv/alieze-erp/internal/modules/auth/types"
	"github.com/KevTiv/alieze-erp/pkg/crm/errors"
	"github.com/google/uuid"
)

// MockAuthService implements the AuthService interface for testing
type MockAuthService struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Permissions    map[string]bool
	Errors         map[string]error
}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		OrganizationID: uuid.Must(uuid.NewV7()),
		UserID:         uuid.Must(uuid.NewV7()),
		Permissions:    make(map[string]bool),
		Errors:         make(map[string]error),
	}
}

func (m *MockAuthService) WithOrganizationID(orgID uuid.UUID) *MockAuthService {
	m.OrganizationID = orgID
	return m
}

func (m *MockAuthService) WithUserID(userID uuid.UUID) *MockAuthService {
	m.UserID = userID
	return m
}

func (m *MockAuthService) AllowPermission(permission string) *MockAuthService {
	m.Permissions[permission] = true
	return m
}

func (m *MockAuthService) DenyPermission(permission string) *MockAuthService {
	m.Permissions[permission] = false
	return m
}

func (m *MockAuthService) WithError(method string, err error) *MockAuthService {
	m.Errors[method] = err
	return m
}

func (m *MockAuthService) GetOrganizationID(ctx context.Context) (uuid.UUID, error) {
	if err, ok := m.Errors["GetOrganizationID"]; ok {
		return uuid.Nil, err
	}
	return m.OrganizationID, nil
}

func (m *MockAuthService) GetUserID(ctx context.Context) (uuid.UUID, error) {
	if err, ok := m.Errors["GetUserID"]; ok {
		return uuid.Nil, err
	}
	return m.UserID, nil
}

func (m *MockAuthService) CheckPermission(ctx context.Context, permission string) error {
	if err, ok := m.Errors["CheckPermission"]; ok {
		return err
	}

	if allowed, ok := m.Permissions[permission]; ok {
		if !allowed {
			return fmt.Errorf("permission denied: %s", permission)
		}
		return nil
	}

	// Default: allow all permissions for testing
	return nil
}

func (m *MockAuthService) CheckOrganizationAccess(ctx context.Context, orgID uuid.UUID) error {
	if err, ok := m.Errors["CheckOrganizationAccess"]; ok {
		return err
	}

	if m.OrganizationID != orgID {
		return errors.ErrOrganizationAccess
	}

	return nil
}

func (m *MockAuthService) CheckUserPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error {
	if err, ok := m.Errors["CheckUserPermission"]; ok {
		return err
	}

	if m.UserID != userID {
		return errors.ErrUnauthorized
	}

	if m.OrganizationID != orgID {
		return errors.ErrOrganizationAccess
	}

	if allowed, ok := m.Permissions[permission]; ok && !allowed {
		return errors.ErrPermissionDenied
	}

	return nil
}

func (m *MockAuthService) GetCurrentUser(ctx context.Context) (*authTypes.User, error) {
	if err, ok := m.Errors["GetCurrentUser"]; ok {
		return nil, err
	}

	return &authTypes.User{
		ID: m.UserID,
	}, nil
}
