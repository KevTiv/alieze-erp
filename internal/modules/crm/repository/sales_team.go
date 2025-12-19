package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

type salesTeamRepository struct {
	db *sql.DB
}

func NewSalesTeamRepository(db *sql.DB) types.SalesTeamRepository {
	return &salesTeamRepository{db: db}
}

func (r *salesTeamRepository) Create(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error) {
	query := `INSERT INTO sales_teams (id, organization_id, company_id, name, code, team_leader_id, member_ids, is_active, created_at, updated_at, created_by, updated_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, organization_id, company_id, name, code, team_leader_id, member_ids, is_active, created_at, updated_at, created_by, updated_by`

	var created types.SalesTeam
	err := r.db.QueryRowContext(ctx, query,
		team.ID, team.OrganizationID, team.CompanyID, team.Name, team.Code,
		team.TeamLeaderID, team.MemberIDs, team.IsActive, team.CreatedAt, team.UpdatedAt,
		team.CreatedBy, team.UpdatedBy).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Name, &created.Code,
		&created.TeamLeaderID, &created.MemberIDs, &created.IsActive, &created.CreatedAt, &created.UpdatedAt,
		&created.CreatedBy, &created.UpdatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create sales team: %w", err)
	}

	return &created, nil
}

func (r *salesTeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
	query := `SELECT id, organization_id, company_id, name, code, team_leader_id, member_ids, is_active, created_at, updated_at, created_by, updated_by, deleted_at FROM sales_teams WHERE id = $1`

	var team types.SalesTeam
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.OrganizationID, &team.CompanyID, &team.Name, &team.Code,
		&team.TeamLeaderID, &team.MemberIDs, &team.IsActive, &team.CreatedAt, &team.UpdatedAt,
		&team.CreatedBy, &team.UpdatedBy, &team.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("sales team not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get sales team: %w", err)
	}

	return &team, nil
}

func (r *salesTeamRepository) FindAll(ctx context.Context, filter types.SalesTeamFilter) ([]*types.SalesTeam, error) {
	query := `SELECT id, organization_id, company_id, name, code, team_leader_id, member_ids, is_active, created_at, updated_at, created_by, updated_by, deleted_at FROM sales_teams WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.CompanyID != nil {
		query += " AND company_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.CompanyID)
	}

	if filter.Name != nil {
		query += " AND name LIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Code != nil {
		query += " AND code = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.Code)
	}

	if filter.TeamLeaderID != nil {
		query += " AND team_leader_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.TeamLeaderID)
	}

	if filter.IsActive != nil {
		query += " AND is_active = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *filter.IsActive)
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
		return nil, fmt.Errorf("failed to query sales teams: %w", err)
	}
	defer rows.Close()

	var teams []*types.SalesTeam
	for rows.Next() {
		var team types.SalesTeam
		if err := rows.Scan(&team.ID, &team.OrganizationID, &team.CompanyID, &team.Name, &team.Code,
			&team.TeamLeaderID, &team.MemberIDs, &team.IsActive, &team.CreatedAt, &team.UpdatedAt,
			&team.CreatedBy, &team.UpdatedBy, &team.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan sales team: %w", err)
		}
		teams = append(teams, &team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales teams: %w", err)
	}

	return teams, nil
}

// Count counts sales teams matching the filter criteria
func (r *salesTeamRepository) Count(ctx context.Context, filter types.SalesTeamFilter) (int, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return 0, errors.New("organization ID not found in context")
	}

	query := `SELECT COUNT(*) FROM sales_teams WHERE organization_id = $1`
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
		return 0, fmt.Errorf("failed to count sales teams: %w", err)
	}

	return count, nil
}

func (r *salesTeamRepository) Update(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error) {
	query := `UPDATE sales_teams SET company_id = $1, name = $2, code = $3, team_leader_id = $4, member_ids = $5, is_active = $6, updated_at = $7, updated_by = $8 WHERE id = $9 RETURNING id, organization_id, company_id, name, code, team_leader_id, member_ids, is_active, created_at, updated_at, created_by, updated_by`

	var updated types.SalesTeam
	err := r.db.QueryRowContext(ctx, query,
		team.CompanyID, team.Name, team.Code, team.TeamLeaderID, team.MemberIDs,
		team.IsActive, team.UpdatedAt, team.UpdatedBy, team.ID).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.Name, &updated.Code,
		&updated.TeamLeaderID, &updated.MemberIDs, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt,
		&updated.CreatedBy, &updated.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("sales team not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update sales team: %w", err)
	}

	return &updated, nil
}

func (r *salesTeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sales_teams WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sales team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("sales team not found: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *salesTeamRepository) FindByMember(ctx context.Context, memberID uuid.UUID) ([]types.SalesTeam, error) {
	query := `SELECT id, organization_id, company_id, name, code, team_leader_id, member_ids, is_active, created_at, updated_at, created_by, updated_by, deleted_at FROM sales_teams WHERE $1 = ANY(member_ids)`

	rows, err := r.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales teams by member: %w", err)
	}
	defer rows.Close()

	var teams []types.SalesTeam
	for rows.Next() {
		var team types.SalesTeam
		if err := rows.Scan(&team.ID, &team.OrganizationID, &team.CompanyID, &team.Name, &team.Code,
			&team.TeamLeaderID, &team.MemberIDs, &team.IsActive, &team.CreatedAt, &team.UpdatedAt,
			&team.CreatedBy, &team.UpdatedBy, &team.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan sales team: %w", err)
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sales teams: %w", err)
	}

	return teams, nil
}
