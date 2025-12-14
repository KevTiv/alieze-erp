package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockAuthRepository(t *testing.T) {
	repo := NewMockAuthRepository()
	ctx := context.Background()

	t.Run("CreateUser and FindUserByID", func(t *testing.T) {
		user := CreateTestUser()
		created, err := repo.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, user.ID, created.ID)
		assert.Equal(t, user.Email, created.Email)

		found, err := repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("FindUserByEmail", func(t *testing.T) {
		user := CreateTestUser()
		user.Email = "email-test@example.com"
		repo.AddUser(user)

		found, err := repo.FindUserByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user := CreateTestUser()
		repo.AddUser(user)

		user.Email = "updated@example.com"
		updated, err := repo.UpdateUser(ctx, user)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "updated@example.com", updated.Email)
	})

	t.Run("CreateOrganization", func(t *testing.T) {
		userID := uuid.New()
		orgID, err := repo.CreateOrganization(ctx, "Test Org", userID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, *orgID)

		name, err := repo.FindOrganizationByID(ctx, *orgID)
		require.NoError(t, err)
		assert.NotNil(t, name)
		assert.Equal(t, "Test Org", *name)
	})

	t.Run("OrganizationUser operations", func(t *testing.T) {
		userID := uuid.New()
		orgID := uuid.New()
		orgUser := CreateTestOrganizationUser(userID, orgID)

		created, err := repo.CreateOrganizationUser(ctx, orgUser)
		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, orgUser.UserID, created.UserID)

		found, err := repo.FindOrganizationUser(ctx, orgID, userID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, orgUser.Role, found.Role)

		orgUsers, err := repo.FindOrganizationUsersByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, orgUsers, 1)
		assert.Equal(t, orgUser.ID, orgUsers[0].ID)
	})

	t.Run("UpdateUserPassword", func(t *testing.T) {
		user := CreateTestUser()
		repo.AddUser(user)

		newPassword := "new-hashed-password"
		err := repo.UpdateUserPassword(ctx, user.ID, newPassword)
		require.NoError(t, err)

		updatedUser, err := repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, newPassword, updatedUser.EncryptedPassword)
	})

	t.Run("Error handling", func(t *testing.T) {
		repo.SetError("FindUserByEmail", assert.AnError)
		defer repo.ClearErrors()

		_, err := repo.FindUserByEmail(ctx, "test@example.com")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})
}
