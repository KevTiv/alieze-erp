package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"alieze-erp/internal/modules/crm/repository"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/internal/testutils"
)

func TestContactTagRepository(t *testing.T) {
	// Setup
	db := testutils.NewMockDB()
	repo := repository.NewContactTagRepository(db)

	orgID := uuid.New()

	// Test Create
	tag := types.ContactTag{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "VIP Customer",
		Color:          0xFF0000, // Red
	}

	created, err := repo.Create(context.Background(), tag)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, tag.Name, created.Name)
	assert.Equal(t, tag.Color, created.Color)
	assert.Equal(t, tag.OrganizationID, created.OrganizationID)
	assert.NotZero(t, created.CreatedAt)

	// Test FindByID
	found, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Name, found.Name)

	// Test FindAll
	filter := types.ContactTagFilter{
		OrganizationID: orgID,
	}

	allTags, err := repo.FindAll(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, allTags, 1)
	assert.Equal(t, created.ID, allTags[0].ID)

	// Test Update
	updatedTag := types.ContactTag{
		ID:    created.ID,
		Name:  "Premium Customer",
		Color: 0x00FF00, // Green
	}

	updated, err := repo.Update(context.Background(), updatedTag)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, updatedTag.Name, updated.Name)
	assert.Equal(t, updatedTag.Color, updated.Color)

	// Test Delete
	err = repo.Delete(context.Background(), created.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(context.Background(), created.ID)
	assert.Error(t, err)
}
