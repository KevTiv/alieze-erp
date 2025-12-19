package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

type leadSourceRepository struct {
	db *sql.DB
}

func NewLeadSourceRepository(db *sql.DB) types.LeadSourceRepository {
	return &leadSourceRepository{db: db}
}

func (r *leadSourceRepository) Create(ctx context.Context, source types.LeadSource) (*types.LeadSource, error) {
	query := `INSERT INTO lead_sources (id, organization_id, name, created_at) VALUES ($1, $2, $3, $4) RETURNING id, organization_id, name, created_at`

	var created types.LeadSource
	err := r.db.QueryRowContext(ctx, query, source.ID, source.OrganizationID, source.Name, source.CreatedAt).Scan(
		&created.ID, &created.OrganizationID, &created.Name, &created.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create lead source: %w", err)
	}

	return &created, nil
}

func (r *leadSourceRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.LeadSource, error) {
	query := `SELECT id, organization_id, name, created_at FROM lead_sources WHERE id = $1`

	var source types.LeadSource
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&source.ID, &source.OrganizationID, &source.Name, &source.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lead source not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get lead source: %w", err)
	}

	return &source, nil
}

func (r *leadSourceRepository) FindAll(ctx context.Context, filter types.LeadSourceFilter) ([]*types.LeadSource, error) {
	query := `SELECT id, organization_id, name, created_at FROM lead_sources WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.Name != nil {
		query += " AND name LIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+*filter.Name+"%")
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
		return nil, fmt.Errorf("failed to query lead sources: %w", err)
	}
	defer rows.Close()

	var sources []*types.LeadSource
	for rows.Next() {
		var source types.LeadSource
		if err := rows.Scan(&source.ID, &source.OrganizationID, &source.Name, &source.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lead source: %w", err)
		}
		sources = append(sources, &source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lead sources: %w", err)
	}

	return sources, nil
}

func (r *leadSourceRepository) Update(ctx context.Context, source types.LeadSource) (*types.LeadSource, error) {
	query := `UPDATE lead_sources SET name = $1 WHERE id = $2 RETURNING id, organization_id, name, created_at`

	var updated types.LeadSource
	err := r.db.QueryRowContext(ctx, query, source.Name, source.ID).Scan(
		&updated.ID, &updated.OrganizationID, &updated.Name, &updated.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lead source not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update lead source: %w", err)
	}

	return &updated, nil
}

func (r *leadSourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM lead_sources WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead source: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lead source not found: %w", sql.ErrNoRows)
	}

	return nil
}

// Count counts lead sources matching the filter criteria
func (r *leadSourceRepository) Count(ctx context.Context, filter types.LeadSourceFilter) (int, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return 0, errors.New("organization ID not found in context")
	}

	query := `SELECT COUNT(*) FROM lead_sources WHERE organization_id = $1`
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
		return 0, fmt.Errorf("failed to count lead sources: %w", err)
	}

	return count, nil
}
