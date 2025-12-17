package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ProcurementGroupRepository handles database operations for procurement groups
type ProcurementGroupRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewProcurementGroupRepository creates a new ProcurementGroupRepository
func NewProcurementGroupRepository(db *sqlx.DB, logger *slog.Logger) *ProcurementGroupRepository {
	return &ProcurementGroupRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new procurement group
func (r *ProcurementGroupRepository) Create(ctx context.Context, orgID uuid.UUID, req types.ProcurementGroupCreateRequest) (*types.ProcurementGroup, error) {
	query := `
		INSERT INTO procurement_groups (organization_id, name, partner_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, organization_id, name, partner_id, created_at
	`

	var group types.ProcurementGroup
	err := r.db.QueryRowxContext(ctx, query, orgID, req.Name, req.PartnerID).StructScan(&group)
	if err != nil {
		r.logger.Error("Failed to create procurement group", "error", err)
		return nil, err
	}

	return &group, nil
}

// GetByID retrieves a procurement group by ID
func (r *ProcurementGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.ProcurementGroup, error) {
	query := `
		SELECT id, organization_id, name, partner_id, created_at
		FROM procurement_groups
		WHERE id = $1
	`

	var group types.ProcurementGroup
	err := r.db.GetContext(ctx, &group, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get procurement group", "error", err, "id", id)
		return nil, err
	}

	return &group, nil
}

// List retrieves all procurement groups for an organization
func (r *ProcurementGroupRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.ProcurementGroup, error) {
	query := `
		SELECT id, organization_id, name, partner_id, created_at
		FROM procurement_groups
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`

	var groups []types.ProcurementGroup
	err := r.db.SelectContext(ctx, &groups, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list procurement groups", "error", err)
		return nil, err
	}

	return groups, nil
}

// Update updates a procurement group
func (r *ProcurementGroupRepository) Update(ctx context.Context, id uuid.UUID, req types.ProcurementGroupUpdateRequest) (*types.ProcurementGroup, error) {
	query := `UPDATE procurement_groups SET `
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		query += `name = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Name)
		argCount++
	}

	if req.PartnerID != nil {
		query += `partner_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.PartnerID)
		argCount++
	}

	if len(args) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += ` WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, name, partner_id, created_at`
	args = append(args, id)

	var group types.ProcurementGroup
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&group)
	if err != nil {
		r.logger.Error("Failed to update procurement group", "error", err)
		return nil, err
	}

	return &group, nil
}

// Delete deletes a procurement group
func (r *ProcurementGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM procurement_groups
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete procurement group", "error", err)
		return err
	}

	return nil
}
