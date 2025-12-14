package testutils

import (
	"context"
	"fmt"

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
