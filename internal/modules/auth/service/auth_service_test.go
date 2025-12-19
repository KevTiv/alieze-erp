package service

import (
	"context"
	"testing"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/auth/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_RegisterUser(t *testing.T) {
	mockRepo := repository.NewMockAuthRepository()
	svc := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("Successful registration", func(t *testing.T) {
		req := types.RegisterRequest{
			Email:            "newuser@example.com",
			Password:         "password123",
			OrganizationName: "Test Org",
			FirstName:        "John",
			LastName:         "Doe",
		}

		profile, err := svc.RegisterUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, req.Email, profile.Email)
		assert.False(t, profile.IsSuperAdmin)

		// Verify user was created
		user, err := mockRepo.FindUserByEmail(ctx, req.Email)
		require.NoError(t, err)
		assert.NotNil(t, user)

		// Verify password was hashed
		err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(req.Password))
		require.NoError(t, err)

		// Verify organization was created
		orgUsers, err := mockRepo.FindOrganizationUsersByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, orgUsers, 1)
		assert.Equal(t, "owner", orgUsers[0].Role)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		req := types.RegisterRequest{
			Email:            "invalid-email",
			Password:         "password123",
			OrganizationName: "Test Org",
			FirstName:        "John",
			LastName:         "Doe",
		}

		_, err := svc.RegisterUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("User already exists", func(t *testing.T) {
		existingUser := repository.CreateTestUser()
		existingUser.Email = "existing@example.com"
		mockRepo.AddUser(existingUser)

		req := types.RegisterRequest{
			Email:            "existing@example.com",
			Password:         "password123",
			OrganizationName: "Test Org",
			FirstName:        "John",
			LastName:         "Doe",
		}

		_, err := svc.RegisterUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user with this email already exists")
	})

	t.Run("Weak password", func(t *testing.T) {
		req := types.RegisterRequest{
			Email:            "test@example.com",
			Password:         "short",
			OrganizationName: "Test Org",
			FirstName:        "John",
			LastName:         "Doe",
		}

		_, err := svc.RegisterUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password")
	})
}

func TestAuthService_LoginUser(t *testing.T) {
	mockRepo := repository.NewMockAuthRepository()
	svc := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("Successful login", func(t *testing.T) {
		// Create test user
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := repository.CreateTestUser()
		user.Email = "login-test@example.com"
		user.EncryptedPassword = string(hashedPassword)
		mockRepo.AddUser(user)

		// Create organization and user membership
		orgID := uuid.New()
		mockRepo.AddOrganization(orgID, "Test Org")
		orgUser := repository.CreateTestOrganizationUser(user.ID, orgID)
		mockRepo.AddOrganizationUser(orgUser)

		req := types.LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		response, err := svc.LoginUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user.Email, response.User.Email)
		assert.Equal(t, user.ID, response.User.ID)
	})

	t.Run("Invalid credentials - wrong password", func(t *testing.T) {
		user := repository.CreateTestUser()
		user.Email = "wrong-pass@example.com"
		mockRepo.AddUser(user)

		req := types.LoginRequest{
			Email:    user.Email,
			Password: "wrongpassword",
		}

		_, err := svc.LoginUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("Invalid credentials - user not found", func(t *testing.T) {
		req := types.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		_, err := svc.LoginUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("User with no organization access", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := repository.CreateTestUser()
		user.Email = "no-org@example.com"
		user.EncryptedPassword = string(hashedPassword)
		mockRepo.AddUser(user)

		req := types.LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		_, err := svc.LoginUser(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no organization access")
	})
}

func TestAuthService_GetUserProfile(t *testing.T) {
	mockRepo := repository.NewMockAuthRepository()
	svc := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("Get existing user profile", func(t *testing.T) {
		user := repository.CreateTestUser()
		mockRepo.AddUser(user)

		profile, err := svc.GetUserProfile(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, user.ID, profile.ID)
		assert.Equal(t, user.Email, profile.Email)
		assert.Equal(t, user.IsSuperAdmin, profile.IsSuperAdmin)
	})

	t.Run("User not found", func(t *testing.T) {
		_, err := svc.GetUserProfile(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestAuthService_GetOrganizationID(t *testing.T) {
	mockRepo := repository.NewMockAuthRepository()
	svc := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("Get organization ID for user", func(t *testing.T) {
		userID := uuid.New()
		orgID := uuid.New()
		orgUser := repository.CreateTestOrganizationUser(userID, orgID)
		mockRepo.AddOrganizationUser(orgUser)

		foundOrgID, err := svc.GetOrganizationID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, orgID, foundOrgID)
	})

	t.Run("User with no organization", func(t *testing.T) {
		_, err := svc.GetOrganizationID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no organization access")
	})
}

func TestAuthService_GetUserRole(t *testing.T) {
	mockRepo := repository.NewMockAuthRepository()
	svc := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("Get user role", func(t *testing.T) {
		userID := uuid.New()
		orgID := uuid.New()
		orgUser := repository.CreateTestOrganizationUser(userID, orgID)
		orgUser.Role = "admin"
		mockRepo.AddOrganizationUser(orgUser)

		role, err := svc.GetUserRole(ctx, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, "admin", role)
	})

	t.Run("User not in organization", func(t *testing.T) {
		_, err := svc.GetUserRole(ctx, uuid.New(), uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found in organization")
	})
}

func TestAuthService_CheckPermission(t *testing.T) {
	mockRepo := repository.NewMockAuthRepository()
	svc := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("Super admin has all permissions", func(t *testing.T) {
		user := repository.CreateTestUser()
		user.IsSuperAdmin = true
		mockRepo.AddUser(user)

		userID := user.ID
		orgID := uuid.New()
		orgUser := repository.CreateTestOrganizationUser(userID, orgID)
		mockRepo.AddOrganizationUser(orgUser)

		err := svc.CheckPermission(ctx, userID, orgID, "any:permission")
		assert.NoError(t, err)
	})

	t.Run("Owner has all permissions", func(t *testing.T) {
		userID := uuid.New()
		orgID := uuid.New()
		orgUser := repository.CreateTestOrganizationUser(userID, orgID)
		orgUser.Role = "owner"
		mockRepo.AddOrganizationUser(orgUser)

		err := svc.CheckPermission(ctx, userID, orgID, "any:permission")
		assert.NoError(t, err)
	})

	t.Run("Admin has all permissions", func(t *testing.T) {
		userID := uuid.New()
		orgID := uuid.New()
		orgUser := repository.CreateTestOrganizationUser(userID, orgID)
		orgUser.Role = "admin"
		mockRepo.AddOrganizationUser(orgUser)

		err := svc.CheckPermission(ctx, userID, orgID, "any:permission")
		assert.NoError(t, err)
	})

	t.Run("Regular user gets permission denied", func(t *testing.T) {
		userID := uuid.New()
		orgID := uuid.New()
		orgUser := repository.CreateTestOrganizationUser(userID, orgID)
		orgUser.Role = "user"
		mockRepo.AddOrganizationUser(orgUser)

		err := svc.CheckPermission(ctx, userID, orgID, "some:permission")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("User not in organization", func(t *testing.T) {
		err := svc.CheckPermission(ctx, uuid.New(), uuid.New(), "some:permission")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found in organization")
	})
}

func TestPasswordHashing(t *testing.T) {
	t.Run("Hash and verify password", func(t *testing.T) {
		password := "test-password-123"
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		err = bcrypt.CompareHashAndPassword(hashed, []byte(password))
		require.NoError(t, err)

		// Wrong password should fail
		err = bcrypt.CompareHashAndPassword(hashed, []byte("wrong-password"))
		assert.Error(t, err)
	})
}
