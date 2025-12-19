package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

type activityRepository struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) types.ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(ctx context.Context, activity types.Activity) (*types.Activity, error) {
	query := `INSERT INTO activities (id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by`

	var created types.Activity
	err := r.db.QueryRowContext(ctx, query,
		activity.ID, activity.OrganizationID, activity.ActivityType, activity.Summary, activity.Note,
		activity.DateDeadline, activity.UserID, activity.AssignedTo, activity.ResModel, activity.ResID,
		activity.State, activity.DoneDate, activity.CreatedAt, activity.UpdatedAt, activity.CreatedBy, activity.UpdatedBy).Scan(
		&created.ID, &created.OrganizationID, &created.ActivityType, &created.Summary, &created.Note,
		&created.DateDeadline, &created.UserID, &created.AssignedTo, &created.ResModel, &created.ResID,
		&created.State, &created.DoneDate, &created.CreatedAt, &created.UpdatedAt, &created.CreatedBy, &created.UpdatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	return &created, nil
}

func (r *activityRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Activity, error) {
	query := `SELECT id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by FROM activities WHERE id = $1`

	var activity types.Activity
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&activity.ID, &activity.OrganizationID, &activity.ActivityType, &activity.Summary, &activity.Note,
		&activity.DateDeadline, &activity.UserID, &activity.AssignedTo, &activity.ResModel, &activity.ResID,
		&activity.State, &activity.DoneDate, &activity.CreatedAt, &activity.UpdatedAt, &activity.CreatedBy, &activity.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("activity not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

func (r *activityRepository) FindAll(ctx context.Context, filter types.ActivityFilter) ([]*types.Activity, error) {
	query := `SELECT id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by FROM activities WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.ActivityType != nil {
		query += " AND activity_type = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.ActivityType)
	}

	if filter.State != nil {
		query += " AND state = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.State)
	}

	if filter.UserID != nil {
		query += " AND user_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.UserID)
	}

	if filter.AssignedTo != nil {
		query += " AND assigned_to = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.AssignedTo)
	}

	if filter.ResModel != nil {
		query += " AND res_model = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.ResModel)
	}

	if filter.ResID != nil {
		query += " AND res_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.ResID)
	}

	query += " ORDER BY date_deadline, created_at"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var activities []*types.Activity
	for rows.Next() {
		var activity types.Activity
		if err := rows.Scan(&activity.ID, &activity.OrganizationID, &activity.ActivityType, &activity.Summary, &activity.Note,
			&activity.DateDeadline, &activity.UserID, &activity.AssignedTo, &activity.ResModel, &activity.ResID,
			&activity.State, &activity.DoneDate, &activity.CreatedAt, &activity.UpdatedAt, &activity.CreatedBy, &activity.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, &activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}

// Count counts activities matching the filter criteria
func (r *activityRepository) Count(ctx context.Context, filter types.ActivityFilter) (int, error) {
	query := `SELECT COUNT(*) FROM activities WHERE organization_id = $1`
	args := []interface{}{filter.OrganizationID}
	argIndex := 2

	if filter.ActivityType != nil && *filter.ActivityType != "" {
		query += fmt.Sprintf(" AND activity_type = $%d", argIndex)
		args = append(args, *filter.ActivityType)
		argIndex++
	}

	if filter.State != nil && *filter.State != "" {
		query += fmt.Sprintf(" AND state = $%d", argIndex)
		args = append(args, *filter.State)
		argIndex++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.AssignedTo != nil {
		query += fmt.Sprintf(" AND assigned_to = $%d", argIndex)
		args = append(args, *filter.AssignedTo)
		argIndex++
	}

	if filter.ResModel != nil && *filter.ResModel != "" {
		query += fmt.Sprintf(" AND res_model = $%d", argIndex)
		args = append(args, *filter.ResModel)
		argIndex++
	}

	if filter.ResID != nil {
		query += fmt.Sprintf(" AND res_id = $%d", argIndex)
		args = append(args, *filter.ResID)
		argIndex++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count activities: %w", err)
	}

	return count, nil
}

func (r *activityRepository) Update(ctx context.Context, activity types.Activity) (*types.Activity, error) {
	query := `UPDATE activities SET activity_type = $1, summary = $2, note = $3, date_deadline = $4, user_id = $5, assigned_to = $6, res_model = $7, res_id = $8, state = $9, done_date = $10, updated_at = $11, updated_by = $12 WHERE id = $13 RETURNING id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by`

	var updated types.Activity
	err := r.db.QueryRowContext(ctx, query,
		activity.ActivityType, activity.Summary, activity.Note, activity.DateDeadline, activity.UserID,
		activity.AssignedTo, activity.ResModel, activity.ResID, activity.State, activity.DoneDate,
		activity.UpdatedAt, activity.UpdatedBy, activity.ID).Scan(
		&updated.ID, &updated.OrganizationID, &updated.ActivityType, &updated.Summary, &updated.Note,
		&updated.DateDeadline, &updated.UserID, &updated.AssignedTo, &updated.ResModel, &updated.ResID,
		&updated.State, &updated.DoneDate, &updated.CreatedAt, &updated.UpdatedAt, &updated.CreatedBy, &updated.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("activity not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	return &updated, nil
}

func (r *activityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM activities WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("activity not found: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *activityRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]*types.Activity, error) {
	query := `SELECT id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by FROM activities WHERE res_model = 'contacts' AND res_id = $1 ORDER BY date_deadline, created_at`

	rows, err := r.db.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities by contact: %w", err)
	}
	defer rows.Close()

	var activities []*types.Activity
	for rows.Next() {
		var activity types.Activity
		if err := rows.Scan(&activity.ID, &activity.OrganizationID, &activity.ActivityType, &activity.Summary, &activity.Note,
			&activity.DateDeadline, &activity.UserID, &activity.AssignedTo, &activity.ResModel, &activity.ResID,
			&activity.State, &activity.DoneDate, &activity.CreatedAt, &activity.UpdatedAt, &activity.CreatedBy, &activity.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, &activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}

func (r *activityRepository) FindByLead(ctx context.Context, leadID uuid.UUID) ([]*types.Activity, error) {
	query := `SELECT id, organization_id, activity_type, summary, note, date_deadline, user_id, assigned_to, res_model, res_id, state, done_date, created_at, updated_at, created_by, updated_by FROM activities WHERE res_model = 'leads' AND res_id = $1 ORDER BY date_deadline, created_at`

	rows, err := r.db.QueryContext(ctx, query, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities by lead: %w", err)
	}
	defer rows.Close()

	var activities []*types.Activity
	for rows.Next() {
		var activity types.Activity
		if err := rows.Scan(&activity.ID, &activity.OrganizationID, &activity.ActivityType, &activity.Summary, &activity.Note,
			&activity.DateDeadline, &activity.UserID, &activity.AssignedTo, &activity.ResModel, &activity.ResID,
			&activity.State, &activity.DoneDate, &activity.CreatedAt, &activity.UpdatedAt, &activity.CreatedBy, &activity.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, &activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}
