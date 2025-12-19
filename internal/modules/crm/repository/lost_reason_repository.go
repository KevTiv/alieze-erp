package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

type lostReasonRepository struct {
	db *sql.DB
}

func NewLostReasonRepository(db *sql.DB) types.LostReasonRepository {
	return &lostReasonRepository{db: db}
}

func (r *lostReasonRepository) Create(ctx context.Context, reason types.LostReason) (*types.LostReason, error) {
	query := `INSERT INTO lost_reasons (id, organization_id, name, active, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, organization_id, name, active, created_at`

	var created types.LostReason
	err := r.db.QueryRowContext(ctx, query, reason.ID, reason.OrganizationID, reason.Name, reason.Active, reason.CreatedAt).Scan(
		&created.ID, &created.OrganizationID, &created.Name, &created.Active, &created.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create lost reason: %w", err)
	}

	return &created, nil
}

func (r *lostReasonRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.LostReason, error) {
	query := `SELECT id, organization_id, name, active, created_at FROM lost_reasons WHERE id = $1`

	var reason types.LostReason
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reason.ID, &reason.OrganizationID, &reason.Name, &reason.Active, &reason.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lost reason not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get lost reason: %w", err)
	}

	return &reason, nil
}

func (r *lostReasonRepository) FindAll(ctx context.Context, filter types.LostReasonFilter) ([]types.LostReason, error) {
	query := `SELECT id, organization_id, name, active, created_at FROM lost_reasons WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.Name != nil {
		query += " AND name LIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Active != nil {
		query += " AND active = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.Active)
	}

	query += " ORDER BY name"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query lost reasons: %w", err)
	}
	defer rows.Close()

	var reasons []types.LostReason
	for rows.Next() {
		var reason types.LostReason
		if err := rows.Scan(&reason.ID, &reason.OrganizationID, &reason.Name, &reason.Active, &reason.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lost reason: %w", err)
		}
		reasons = append(reasons, reason)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lost reasons: %w", err)
	}

	return reasons, nil
}

func (r *lostReasonRepository) Update(ctx context.Context, reason types.LostReason) (*types.LostReason, error) {
	query := `UPDATE lost_reasons SET name = $1, active = $2 WHERE id = $3 RETURNING id, organization_id, name, active, created_at`

	var updated types.LostReason
	err := r.db.QueryRowContext(ctx, query, reason.Name, reason.Active, reason.ID).Scan(
		&updated.ID, &updated.OrganizationID, &updated.Name, &updated.Active, &updated.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lost reason not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update lost reason: %w", err)
	}

	return &updated, nil
}

func (r *lostReasonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM lost_reasons WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lost reason: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lost reason not found: %w", sql.ErrNoRows)
	}

	return nil
}

// Count counts lost reasons matching the filter criteria
func (r *lostReasonRepository) Count(ctx context.Context, filter types.LostReasonFilter) (int, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return 0, errors.New("organization ID not found in context")
	}

	query := `SELECT COUNT(*) FROM lost_reasons WHERE organization_id = $1`
	args := []interface{}{orgID}
	argIndex := 2

	if filter.Name != nil && *filter.Name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count lost reasons: %w", err)
	}

	return count, nil
}
