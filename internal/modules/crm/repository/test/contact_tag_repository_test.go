package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/internal/testutils"
)

func TestContactTagRepository(t *testing.T) {
	// Setup
	mockDB := testutils.NewMockDB()
	repo := repository.NewContactTagRepository(mockDB.DB)

	orgID := uuid.Must(uuid.NewV7())
	tagID := uuid.Must(uuid.NewV7())

	t.Run("Create - Success", func(t *testing.T) {
		// Setup test data
		tag := types.ContactTag{
			ID:             tagID,
			OrganizationID: orgID,
			Name:           "VIP Customer",
			Color:          0xFF0000, // Red
		}

		// Expected query and result
		now := time.Now()
		mockDB.Mock.ExpectQuery("INSERT INTO contact_tags").
			WithArgs(
				tag.ID,
				tag.OrganizationID,
				tag.Name,
				tag.Color,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "color", "created_at"}).
				AddRow(tag.ID, tag.OrganizationID, tag.Name, tag.Color, now))

		// Execute
		created, err := repo.Create(context.Background(), tag)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, tag.ID, created.ID)
		assert.Equal(t, tag.Name, created.Name)
		assert.Equal(t, tag.Color, created.Color)
		assert.Equal(t, tag.OrganizationID, created.OrganizationID)
		assert.NotZero(t, created.CreatedAt)

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("FindByID - Success", func(t *testing.T) {
		// Setup test data
		tagID := uuid.Must(uuid.NewV7())
		now := time.Now()

		// Expected query and result
		mockDB.Mock.ExpectQuery("SELECT.*FROM contact_tags").
			WithArgs(tagID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "color", "created_at"}).
				AddRow(tagID, orgID, "VIP Customer", 0xFF0000, now))

		// Execute
		found, err := repo.FindByID(context.Background(), tagID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, tagID, found.ID)
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "VIP Customer", found.Name)
		assert.Equal(t, 0xFF0000, found.Color)
		assert.NotZero(t, found.CreatedAt)

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("FindByID - Not Found", func(t *testing.T) {
		// Setup test data
		tagID := uuid.Must(uuid.NewV7())

		// Expected query and result
		mockDB.Mock.ExpectQuery("SELECT.*FROM contact_tags").
			WithArgs(tagID).
			WillReturnError(sql.ErrNoRows)

		// Execute
		found, err := repo.FindByID(context.Background(), tagID)

		// Assert
		require.Error(t, err)
		require.Nil(t, found)
		assert.Contains(t, err.Error(), "contact tag not found")

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("FindAll - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ContactTagFilter{
			OrganizationID: orgID,
			Name:           stringPtr("VIP"),
			Limit:          10,
			Offset:         0,
		}

		// Expected query and result
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "organization_id", "name", "color", "created_at"}).
			AddRow(uuid.Must(uuid.NewV7()), orgID, "VIP Customer", 0xFF0000, now).
			AddRow(uuid.Must(uuid.NewV7()), orgID, "VIP Partner", 0x00FF00, now)

		mockDB.Mock.ExpectQuery("SELECT.*FROM contact_tags").
			WithArgs(orgID, "%VIP%").
			WillReturnRows(rows)

		// Execute
		allTags, err := repo.FindAll(context.Background(), filter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, allTags)
		assert.Len(t, allTags, 2)
		assert.Equal(t, "VIP Customer", allTags[0].Name)
		assert.Equal(t, "VIP Partner", allTags[1].Name)

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("Update - Success", func(t *testing.T) {
		// Setup test data
		tag := types.ContactTag{
			ID:    tagID,
			Name:  "Premium Customer",
			Color: 0x00FF00, // Green
		}

		// Expected query and result
		now := time.Now()
		mockDB.Mock.ExpectQuery("UPDATE contact_tags").
			WithArgs(
				tag.Name,
				tag.Color,
				tag.ID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "color", "created_at"}).
				AddRow(tag.ID, orgID, tag.Name, tag.Color, now))

		// Execute
		updated, err := repo.Update(context.Background(), tag)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, tag.ID, updated.ID)
		assert.Equal(t, tag.Name, updated.Name)
		assert.Equal(t, tag.Color, updated.Color)
		assert.NotZero(t, updated.CreatedAt)

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("Delete - Success", func(t *testing.T) {
		// Setup test data
		tagID := uuid.Must(uuid.NewV7())

		// Expected query and result
		result := sqlmock.NewResult(0, 1) // 1 row affected
		mockDB.Mock.ExpectExec("DELETE FROM contact_tags").
			WithArgs(tagID).
			WillReturnResult(result)

		// Execute
		err := repo.Delete(context.Background(), tagID)

		// Assert
		require.NoError(t, err)

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("Delete - Not Found", func(t *testing.T) {
		// Setup test data
		tagID := uuid.Must(uuid.NewV7())

		// Expected query and result
		result := sqlmock.NewResult(0, 0) // 0 rows affected
		mockDB.Mock.ExpectExec("DELETE FROM contact_tags").
			WithArgs(tagID).
			WillReturnResult(result)

		// Execute
		err := repo.Delete(context.Background(), tagID)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contact tag not found")

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})

	t.Run("Count - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ContactTagFilter{
			OrganizationID: orgID,
			Name:           stringPtr("VIP"),
		}

		// Expected query and result
		count := 3
		rows := sqlmock.NewRows([]string{"count"}).AddRow(count)

		mockDB.Mock.ExpectQuery("SELECT COUNT").
			WithArgs(orgID, "%VIP%").
			WillReturnRows(rows)

		// Execute
		total, err := repo.Count(context.Background(), filter)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, count, total)

		// Ensure all expectations were met
		err = mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}
