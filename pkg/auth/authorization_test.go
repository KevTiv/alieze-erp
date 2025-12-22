package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPolicyEngine is a mock implementation of PolicyEngine for testing
type MockPolicyEngine struct {
	mock.Mock
}

func (m *MockPolicyEngine) CheckPermission(ctx context.Context, subject, object, action string) (bool, error) {
	args := m.Called(ctx, subject, object, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockPolicyEngine) GetRolesForUser(user string) ([]string, error) {
	args := m.Called(user)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPolicyEngine) GetPermissionsForUser(user string) ([][]string, error) {
	args := m.Called(user)
	return args.Get(0).([][]string), args.Error(1)
}

func TestNewBasicAuthorizationService(t *testing.T) {
	mockEngine := new(MockPolicyEngine)
	service := NewBasicAuthorizationService(mockEngine)

	assert.NotNil(t, service)
	assert.Equal(t, mockEngine, service.policyEngine)
	assert.NotNil(t, service.logger)
}

func TestCheckPermission(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	permission := "contacts:create"

	t.Run("successful permission check", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		mockEngine.On("CheckPermission", ctx, "user:"+userID.String()+":org:"+orgID.String(), "contacts", "create").Return(true, nil)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckPermission(ctx, userID, orgID, permission)

		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("failed permission check", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		mockEngine.On("CheckPermission", ctx, "user:"+userID.String()+":org:"+orgID.String(), "contacts", "create").Return(false, nil)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckPermission(ctx, userID, orgID, permission)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
		mockEngine.AssertExpectations(t)
	})

	t.Run("permission check error", func(t *testing.T) {
		expectedErr := errors.New("policy engine error")
		mockEngine := new(MockPolicyEngine)
		mockEngine.On("CheckPermission", ctx, "user:"+userID.String()+":org:"+orgID.String(), "contacts", "create").Return(false, expectedErr)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckPermission(ctx, userID, orgID, permission)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission check failed")
		assert.True(t, errors.Is(err, expectedErr))
		mockEngine.AssertExpectations(t)
	})

	t.Run("permission without resource prefix", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		mockEngine.On("CheckPermission", ctx, "user:"+userID.String()+":org:"+orgID.String(), "contacts", "read").Return(true, nil)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckPermission(ctx, userID, orgID, "read")

		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("nil policy engine", func(t *testing.T) {
		service := NewBasicAuthorizationService(nil)
		err := service.CheckPermission(ctx, userID, orgID, permission)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "policy engine not configured")
	})
}

func TestCheckResourceAccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	resourceID := uuid.New()
	resourceType := "contacts"

	t.Run("successful resource access check", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		object := resourceType + ":" + resourceID.String()
		mockEngine.On("CheckPermission", ctx, subject, object, "access").Return(true, nil)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckResourceAccess(ctx, userID, orgID, resourceType, resourceID)

		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("failed resource access check", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		object := resourceType + ":" + resourceID.String()
		mockEngine.On("CheckPermission", ctx, subject, object, "access").Return(false, nil)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckResourceAccess(ctx, userID, orgID, resourceType, resourceID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
		mockEngine.AssertExpectations(t)
	})

	t.Run("resource access check error", func(t *testing.T) {
		expectedErr := errors.New("policy engine error")
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		object := resourceType + ":" + resourceID.String()
		mockEngine.On("CheckPermission", ctx, subject, object, "access").Return(false, expectedErr)

		service := NewBasicAuthorizationService(mockEngine)
		err := service.CheckResourceAccess(ctx, userID, orgID, resourceType, resourceID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource access check failed")
		assert.True(t, errors.Is(err, expectedErr))
		mockEngine.AssertExpectations(t)
	})
}

func TestHasOrganizationAccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()

	t.Run("has organization access", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String()
		object := "org:" + orgID.String()
		mockEngine.On("CheckPermission", ctx, subject, object, "access").Return(true, nil)

		service := NewBasicAuthorizationService(mockEngine)
		hasAccess := service.HasOrganizationAccess(ctx, userID, orgID)

		assert.True(t, hasAccess)
		mockEngine.AssertExpectations(t)
	})

	t.Run("no organization access", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String()
		object := "org:" + orgID.String()
		mockEngine.On("CheckPermission", ctx, subject, object, "access").Return(false, nil)

		service := NewBasicAuthorizationService(mockEngine)
		hasAccess := service.HasOrganizationAccess(ctx, userID, orgID)

		assert.False(t, hasAccess)
		mockEngine.AssertExpectations(t)
	})

	t.Run("organization access check error", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String()
		object := "org:" + orgID.String()
		mockEngine.On("CheckPermission", ctx, subject, object, "access").Return(false, errors.New("policy engine error"))

		service := NewBasicAuthorizationService(mockEngine)
		hasAccess := service.HasOrganizationAccess(ctx, userID, orgID)

		assert.False(t, hasAccess)
		mockEngine.AssertExpectations(t)
	})

	t.Run("nil policy engine", func(t *testing.T) {
		service := NewBasicAuthorizationService(nil)
		hasAccess := service.HasOrganizationAccess(ctx, userID, orgID)

		assert.False(t, hasAccess)
	})
}

func TestGetUserRoles(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()

	t.Run("get user roles successfully", func(t *testing.T) {
		expectedRoles := []string{"admin", "manager"}
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		mockEngine.On("GetRolesForUser", subject).Return(expectedRoles, nil)

		service := NewBasicAuthorizationService(mockEngine)
		roles, err := service.GetUserRoles(ctx, userID, orgID)

		assert.NoError(t, err)
		assert.Equal(t, expectedRoles, roles)
		mockEngine.AssertExpectations(t)
	})

	t.Run("no roles found - return default", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		mockEngine.On("GetRolesForUser", subject).Return([]string{}, nil)

		service := NewBasicAuthorizationService(mockEngine)
		roles, err := service.GetUserRoles(ctx, userID, orgID)

		assert.NoError(t, err)
		assert.Equal(t, []string{"user"}, roles)
		mockEngine.AssertExpectations(t)
	})

	t.Run("get user roles error", func(t *testing.T) {
		expectedErr := errors.New("policy engine error")
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		mockEngine.On("GetRolesForUser", subject).Return([]string{}, expectedErr)

		service := NewBasicAuthorizationService(mockEngine)
		roles, err := service.GetUserRoles(ctx, userID, orgID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user roles")
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, roles)
		mockEngine.AssertExpectations(t)
	})

	t.Run("nil policy engine", func(t *testing.T) {
		service := NewBasicAuthorizationService(nil)
		roles, err := service.GetUserRoles(ctx, userID, orgID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "policy engine not configured")
		assert.Nil(t, roles)
	})
}

func TestGetUserPermissions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()

	t.Run("get user permissions successfully", func(t *testing.T) {
		expectedPermissions := [][]string{
			{"user:" + userID.String() + ":org:" + orgID.String(), "contacts", "read"},
			{"user:" + userID.String() + ":org:" + orgID.String(), "contacts", "create"},
		}
		expectedResult := []string{"contacts:read", "contacts:create"}

		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		mockEngine.On("GetPermissionsForUser", subject).Return(expectedPermissions, nil)

		service := NewBasicAuthorizationService(mockEngine)
		permissions, err := service.GetUserPermissions(ctx, userID, orgID)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, permissions)
		mockEngine.AssertExpectations(t)
	})

	t.Run("no permissions found - return default", func(t *testing.T) {
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		mockEngine.On("GetPermissionsForUser", subject).Return([][]string{}, nil)

		service := NewBasicAuthorizationService(mockEngine)
		permissions, err := service.GetUserPermissions(ctx, userID, orgID)

		assert.NoError(t, err)
		assert.Equal(t, []string{"read", "write"}, permissions)
		mockEngine.AssertExpectations(t)
	})

	t.Run("get user permissions error", func(t *testing.T) {
		expectedErr := errors.New("policy engine error")
		mockEngine := new(MockPolicyEngine)
		subject := "user:" + userID.String() + ":org:" + orgID.String()
		mockEngine.On("GetPermissionsForUser", subject).Return([][]string{}, expectedErr)

		service := NewBasicAuthorizationService(mockEngine)
		permissions, err := service.GetUserPermissions(ctx, userID, orgID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user permissions")
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, permissions)
		mockEngine.AssertExpectations(t)
	})

	t.Run("nil policy engine", func(t *testing.T) {
		service := NewBasicAuthorizationService(nil)
		permissions, err := service.GetUserPermissions(ctx, userID, orgID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "policy engine not configured")
		assert.Nil(t, permissions)
	})
}
