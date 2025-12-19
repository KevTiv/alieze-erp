package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

type leadStageRepository struct {
	db *sql.DB
}

func NewLeadStageRepository(db *sql.DB) types.LeadStageRepository {
	return &leadStageRepository{db: db}
}

func (r *leadStageRepository) Create(ctx context.Context, stage types.LeadStage) (*types.LeadStage, error) {
	query := `INSERT INTO lead_stages (id, organization_id, name, sequence, probability, fold, is_won, requirements, team_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, organization_id, name, sequence, probability, fold, is_won, requirements, team_id, created_at, updated_at`

	var created types.LeadStage
	err := r.db.QueryRowContext(ctx, query,
		stage.ID, stage.OrganizationID, stage.Name, stage.Sequence, stage.Probability,
		stage.Fold, stage.IsWon, stage.Requirements, stage.TeamID, stage.CreatedAt, stage.UpdatedAt).Scan(
		&created.ID, &created.OrganizationID, &created.Name, &created.Sequence, &created.Probability,
		&created.Fold, &created.IsWon, &created.Requirements, &created.TeamID, &created.CreatedAt, &created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create lead stage: %w", err)
	}

	return &created, nil
}

func (r *leadStageRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.LeadStage, error) {
	query := `SELECT id, organization_id, name, sequence, probability, fold, is_won, requirements, team_id, created_at, updated_at FROM lead_stages WHERE id = $1`

	var stage types.LeadStage
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&stage.ID, &stage.OrganizationID, &stage.Name, &stage.Sequence, &stage.Probability,
		&stage.Fold, &stage.IsWon, &stage.Requirements, &stage.TeamID, &stage.CreatedAt, &stage.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lead stage not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get lead stage: %w", err)
	}

	return &stage, nil
}

func (r *leadStageRepository) FindAll(ctx context.Context, filter types.LeadStageFilter) ([]*types.LeadStage, error) {
	query := `SELECT id, organization_id, name, sequence, probability, fold, is_won, requirements, team_id, created_at, updated_at FROM lead_stages WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.Name != nil {
		query += " AND name LIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.IsWon != nil {
		query += " AND is_won = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.IsWon)
	}

	if filter.TeamID != nil {
		query += " AND team_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.TeamID)
	}

	query += " ORDER BY sequence"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query lead stages: %w", err)
	}
	defer rows.Close()

	var stages []*types.LeadStage
	for rows.Next() {
		var stage types.LeadStage
		if err := rows.Scan(&stage.ID, &stage.OrganizationID, &stage.Name, &stage.Sequence, &stage.Probability,
			&stage.Fold, &stage.IsWon, &stage.Requirements, &stage.TeamID, &stage.CreatedAt, &stage.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lead stage: %w", err)
		}
		stages = append(stages, &stage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lead stages: %w", err)
	}

	return stages, nil
}

func (r *leadStageRepository) Update(ctx context.Context, stage types.LeadStage) (*types.LeadStage, error) {
	query := `UPDATE lead_stages SET name = $1, sequence = $2, probability = $3, fold = $4, is_won = $5, requirements = $6, team_id = $7, updated_at = $8 WHERE id = $9 RETURNING id, organization_id, name, sequence, probability, fold, is_won, requirements, team_id, created_at, updated_at`

	var updated types.LeadStage
	err := r.db.QueryRowContext(ctx, query,
		stage.Name, stage.Sequence, stage.Probability, stage.Fold, stage.IsWon,
		stage.Requirements, stage.TeamID, stage.UpdatedAt, stage.ID).Scan(
		&updated.ID, &updated.OrganizationID, &updated.Name, &updated.Sequence, &updated.Probability,
		&updated.Fold, &updated.IsWon, &updated.Requirements, &updated.TeamID, &updated.CreatedAt, &updated.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lead stage not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update lead stage: %w", err)
	}

	return &updated, nil
}

func (r *leadStageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM lead_stages WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead stage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lead stage not found: %w", sql.ErrNoRows)
	}

	return nil
}

// Count counts lead stages matching the filter criteria
func (r *leadStageRepository) Count(ctx context.Context, filter types.LeadStageFilter) (int, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return 0, errors.New("organization ID not found in context")
	}

	query := `SELECT COUNT(*) FROM lead_stages WHERE organization_id = $1`
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
		return 0, fmt.Errorf("failed to count lead stages: %w", err)
	}

	return count, nil
}
