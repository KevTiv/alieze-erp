package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// BasicLeadRepository handles basic lead data operations
type BasicLeadRepository struct {
	db *sql.DB
}

func NewBasicLeadRepository(db *sql.DB) *BasicLeadRepository {
	return &BasicLeadRepository{db: db}
}

func (r *BasicLeadRepository) Create(ctx context.Context, lead types.Lead) (*types.Lead, error) {
	if lead.ID == uuid.Nil {
		lead.ID = uuid.New()
	}

	if lead.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if lead.Name == "" {
		return nil, errors.New("name is required")
	}

	if lead.Status == "" {
		lead.Status = "new"
	}

	if lead.Active == false {
		lead.Active = true
	}

	query := `
		INSERT INTO leads (
			id, organization_id, name, email, phone, stage_id, status, active,
			created_at, updated_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, organization_id, name, email, phone, stage_id, status, active,
			created_at, updated_at, deleted_at
	`

	now := time.Now()

	result := r.db.QueryRowContext(ctx, query,
		lead.ID,
		lead.OrganizationID,
		lead.Name,
		lead.Email,
		lead.Phone,
		lead.StageID,
		lead.Status,
		lead.Active,
		now,
		now,
		nil,
	)

	var created types.Lead
	err := result.Scan(
		&created.ID,
		&created.OrganizationID,
		&created.Name,
		&created.Email,
		&created.Phone,
		&created.StageID,
		&created.Status,
		&created.Active,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}

	return &created, nil
}

func (r *BasicLeadRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid lead id")
	}

	query := `
		SELECT id, organization_id, name, email, phone, stage_id, status, active,
			created_at, updated_at, deleted_at
		FROM leads
		WHERE id = $1 AND deleted_at IS NULL
	`

	var lead types.Lead
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lead.ID,
		&lead.OrganizationID,
		&lead.Name,
		&lead.Email,
		&lead.Phone,
		&lead.StageID,
		&lead.Status,
		&lead.Active,
		&lead.CreatedAt,
		&lead.UpdatedAt,
		&lead.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lead not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	return &lead, nil
}

func (r *BasicLeadRepository) FindAll(ctx context.Context, filter types.LeadFilter) ([]types.Lead, error) {
	query := `SELECT id, organization_id, name, email, phone, stage_id, status, active,
		created_at, updated_at, deleted_at
		FROM leads WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.Email != nil && *filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	if filter.Phone != nil && *filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Phone+"%")
		argIndex++
	}

	if filter.StageID != nil && *filter.StageID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("stage_id = $%d", argIndex))
		args = append(args, *filter.StageID)
		argIndex++
	}

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filter.Active)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY name ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find leads: %w", err)
	}
	defer rows.Close()

	var leads []types.Lead
	for rows.Next() {
		var lead types.Lead
		err := rows.Scan(
			&lead.ID,
			&lead.OrganizationID,
			&lead.Name,
			&lead.Email,
			&lead.Phone,
			&lead.StageID,
			&lead.Status,
			&lead.Active,
			&lead.CreatedAt,
			&lead.UpdatedAt,
			&lead.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during lead iteration: %w", err)
	}

	return leads, nil
}

func (r *BasicLeadRepository) Update(ctx context.Context, lead types.Lead) (*types.Lead, error) {
	if lead.ID == uuid.Nil {
		return nil, errors.New("lead id is required")
	}

	if lead.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if lead.Name == "" {
		return nil, errors.New("name is required")
	}

	lead.UpdatedAt = time.Now()

	query := `
		UPDATE leads SET
			organization_id = $1,
			name = $2,
			email = $3,
			phone = $4,
			stage_id = $5,
			status = $6,
			active = $7,
			updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
		RETURNING id, organization_id, name, email, phone, stage_id, status, active,
			created_at, updated_at, deleted_at
	`

	result := r.db.QueryRowContext(ctx, query,
		lead.OrganizationID,
		lead.Name,
		lead.Email,
		lead.Phone,
		lead.StageID,
		lead.Status,
		lead.Active,
		lead.UpdatedAt,
		lead.ID,
	)

	var updated types.Lead
	err := result.Scan(
		&updated.ID,
		&updated.OrganizationID,
		&updated.Name,
		&updated.Email,
		&updated.Phone,
		&updated.StageID,
		&updated.Status,
		&updated.Active,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update lead: %w", err)
	}

	return &updated, nil
}

func (r *BasicLeadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid lead id")
	}

	query := `
		UPDATE leads SET
			deleted_at = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lead not found or already deleted")
	}

	return nil
}

func (r *BasicLeadRepository) Count(ctx context.Context, filter types.LeadFilter) (int, error) {
	query := `SELECT COUNT(*) FROM leads WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.Email != nil && *filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	if filter.Phone != nil && *filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Phone+"%")
		argIndex++
	}

	if filter.StageID != nil && *filter.StageID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("stage_id = $%d", argIndex))
		args = append(args, *filter.StageID)
		argIndex++
	}

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filter.Active)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count leads: %w", err)
	}

	return count, nil
}
