package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactSegmentRepository_CreateSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success - static segment", func(t *testing.T) {
		segment := &types.ContactSegment{
			OrganizationID: uuid.New(),
			Name:           "Test Segment",
			SegmentType:    "static",
			MemberCount:    0,
		}

		now := time.Now()
		mock.ExpectQuery(`INSERT INTO contact_segments`).
			WithArgs(
				sqlmock.AnyArg(), // id
				segment.OrganizationID,
				segment.Name,
				segment.Description,
				segment.SegmentType,
				segment.Criteria,
				segment.Color,
				segment.Icon,
				segment.MemberCount,
				segment.CreatedBy,
			).
			WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
				AddRow(now, now))

		err := repo.CreateSegment(ctx, segment)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, segment.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - dynamic segment with criteria", func(t *testing.T) {
		criteria := types.JSONBMap{
			"is_customer": true,
			"city":        "New York",
		}

		segment := &types.ContactSegment{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "NYC Customers",
			SegmentType:    "dynamic",
			Criteria:       criteria,
			MemberCount:    0,
		}

		now := time.Now()
		mock.ExpectQuery(`INSERT INTO contact_segments`).
			WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
				AddRow(now, now))

		err := repo.CreateSegment(ctx, segment)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - database failure", func(t *testing.T) {
		segment := &types.ContactSegment{
			OrganizationID: uuid.New(),
			Name:           "Test Segment",
			SegmentType:    "static",
		}

		mock.ExpectQuery(`INSERT INTO contact_segments`).
			WillReturnError(sqlmock.ErrCancelled)

		err := repo.CreateSegment(ctx, segment)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_GetSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success - segment found", func(t *testing.T) {
		segmentID := uuid.New()
		orgID := uuid.New()
		now := time.Now()

		mock.ExpectQuery(`SELECT (.+) FROM contact_segments WHERE id`).
			WithArgs(segmentID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "description", "segment_type",
				"criteria", "color", "icon", "member_count", "last_calculated_at",
				"created_at", "updated_at", "created_by",
			}).AddRow(
				segmentID, orgID, "Test Segment", nil, "static",
				types.JSONBMap{}, nil, nil, 5, nil,
				now, now, nil,
			))

		segment, err := repo.GetSegment(ctx, segmentID)

		assert.NoError(t, err)
		assert.NotNil(t, segment)
		assert.Equal(t, segmentID, segment.ID)
		assert.Equal(t, "Test Segment", segment.Name)
		assert.Equal(t, 5, segment.MemberCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		segmentID := uuid.New()

		mock.ExpectQuery(`SELECT (.+) FROM contact_segments WHERE id`).
			WithArgs(segmentID).
			WillReturnRows(sqlmock.NewRows([]string{}))

		segment, err := repo.GetSegment(ctx, segmentID)

		assert.NoError(t, err)
		assert.Nil(t, segment)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_UpdateSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		segment := &types.ContactSegment{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "Updated Segment",
			SegmentType:    "static",
			Criteria:       types.JSONBMap{"is_customer": true},
		}

		now := time.Now()
		mock.ExpectQuery(`UPDATE contact_segments SET`).
			WithArgs(
				segment.ID,
				segment.Name,
				segment.Description,
				segment.Criteria,
				segment.Color,
				segment.Icon,
				segment.OrganizationID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

		err := repo.UpdateSegment(ctx, segment)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		segment := &types.ContactSegment{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "Test",
			SegmentType:    "static",
		}

		mock.ExpectQuery(`UPDATE contact_segments SET`).
			WillReturnRows(sqlmock.NewRows([]string{}))

		err := repo.UpdateSegment(ctx, segment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_DeleteSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		segmentID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM contact_segment_members`).
			WithArgs(segmentID).
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectExec(`DELETE FROM contact_segments`).
			WithArgs(segmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.DeleteSegment(ctx, segmentID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		segmentID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM contact_segment_members`).
			WithArgs(segmentID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`DELETE FROM contact_segments`).
			WithArgs(segmentID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()

		err := repo.DeleteSegment(ctx, segmentID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_ListSegments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success - no filters", func(t *testing.T) {
		orgID := uuid.New()
		now := time.Now()

		filter := types.SegmentFilter{
			OrganizationID: orgID,
			Limit:          10,
			Offset:         0,
		}

		mock.ExpectQuery(`SELECT (.+) FROM contact_segments WHERE organization_id`).
			WithArgs(orgID, 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "description", "segment_type",
				"criteria", "color", "icon", "member_count", "last_calculated_at",
				"created_at", "updated_at", "created_by",
			}).
				AddRow(uuid.New(), orgID, "Segment 1", nil, "static", types.JSONBMap{}, nil, nil, 10, nil, now, now, nil).
				AddRow(uuid.New(), orgID, "Segment 2", nil, "dynamic", types.JSONBMap{}, nil, nil, 5, &now, now, now, nil))

		segments, err := repo.ListSegments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, segments, 2)
		assert.Equal(t, "Segment 1", segments[0].Name)
		assert.Equal(t, "Segment 2", segments[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - with type filter", func(t *testing.T) {
		orgID := uuid.New()
		segmentType := "dynamic"

		filter := types.SegmentFilter{
			OrganizationID: orgID,
			SegmentType:    &segmentType,
			Limit:          10,
			Offset:         0,
		}

		mock.ExpectQuery(`SELECT (.+) FROM contact_segments WHERE organization_id (.+) AND segment_type`).
			WithArgs(orgID, segmentType, 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "description", "segment_type",
				"criteria", "color", "icon", "member_count", "last_calculated_at",
				"created_at", "updated_at", "created_by",
			}).AddRow(uuid.New(), orgID, "Dynamic Segment", nil, "dynamic", types.JSONBMap{}, nil, nil, 5, nil, time.Now(), time.Now(), nil))

		segments, err := repo.ListSegments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, segments, 1)
		assert.Equal(t, "dynamic", segments[0].SegmentType)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_AddContactsToSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success - add multiple contacts", func(t *testing.T) {
		segmentID := uuid.New()
		orgID := uuid.New()
		addedBy := uuid.New()
		contactIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO contact_segment_members`)
		for range contactIDs {
			mock.ExpectExec(`INSERT INTO contact_segment_members`).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectExec(`UPDATE contact_segments SET member_count`).
			WithArgs(segmentID, segmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.AddContactsToSegment(ctx, segmentID, contactIDs, addedBy, orgID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty contact list", func(t *testing.T) {
		segmentID := uuid.New()
		orgID := uuid.New()
		addedBy := uuid.New()

		err := repo.AddContactsToSegment(ctx, segmentID, []uuid.UUID{}, addedBy, orgID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_RemoveContactsFromSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		segmentID := uuid.New()
		contactIDs := []uuid.UUID{uuid.New(), uuid.New()}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM contact_segment_members`).
			WithArgs(segmentID, contactIDs).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(`UPDATE contact_segments SET member_count`).
			WithArgs(segmentID, segmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.RemoveContactsFromSegment(ctx, segmentID, contactIDs)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty contact list", func(t *testing.T) {
		segmentID := uuid.New()

		err := repo.RemoveContactsFromSegment(ctx, segmentID, []uuid.UUID{})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_GetSegmentMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		segmentID := uuid.New()
		orgID := uuid.New()
		now := time.Now()

		email := "test@example.com"
		phone := "+1234567890"

		mock.ExpectQuery(`SELECT c.id, c.organization_id, c.name`).
			WithArgs(segmentID, 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "email", "phone",
				"is_customer", "is_vendor", "street", "city", "state_id",
				"country_id", "created_at", "updated_at",
			}).
				AddRow(uuid.New(), orgID, "Contact 1", &email, &phone, true, false, nil, nil, nil, nil, now, now).
				AddRow(uuid.New(), orgID, "Contact 2", &email, nil, false, true, nil, nil, nil, nil, now, now))

		contacts, err := repo.GetSegmentMembers(ctx, segmentID, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, contacts, 2)
		assert.Equal(t, "Contact 1", contacts[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_ClearSegmentMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		segmentID := uuid.New()

		mock.ExpectExec(`DELETE FROM contact_segment_members WHERE segment_id`).
			WithArgs(segmentID).
			WillReturnResult(sqlmock.NewResult(0, 5))

		err := repo.ClearSegmentMembers(ctx, segmentID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContactSegmentRepository_UpdateSegmentMemberCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewContactSegmentRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		segmentID := uuid.New()
		count := 42

		mock.ExpectExec(`UPDATE contact_segments SET member_count`).
			WithArgs(segmentID, count).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateSegmentMemberCount(ctx, segmentID, count)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
