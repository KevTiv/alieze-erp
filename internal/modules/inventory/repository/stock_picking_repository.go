package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockPickingRepository handles database operations for stock pickings
type StockPickingRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewStockPickingRepository creates a new StockPickingRepository
func NewStockPickingRepository(db *sql.DB, logger *slog.Logger) *StockPickingRepository {
	return &StockPickingRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new stock picking
func (r *StockPickingRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockPickingCreateRequest) (*types.StockPicking, error) {
	query := `
		INSERT INTO stock_pickings (organization_id, company_id, name, sequence_code, picking_type_id, location_id, location_dest_id, partner_id, date, scheduled_date, state, priority, origin, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING id, organization_id, company_id, name, sequence_code, picking_type_id, location_id, location_dest_id, partner_id, date, scheduled_date, state, priority, origin, note, created_at, updated_at
	`

	var picking types.StockPicking
	err := r.db.QueryRowContext(ctx, query, orgID, req.CompanyID, req.Name, req.SequenceCode, req.PickingTypeID, req.LocationID, req.LocationDestID, req.PartnerID, req.Date, req.ScheduledDate, req.State, req.Priority, req.Origin, req.Note).Scan(
		&picking.ID, &picking.OrganizationID, &picking.CompanyID, &picking.Name, &picking.SequenceCode, &picking.PickingTypeID, &picking.LocationID, &picking.LocationDestID, &picking.PartnerID, &picking.Date, &picking.ScheduledDate, &picking.State, &picking.Priority, &picking.Origin, &picking.Note, &picking.CreatedAt, &picking.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to create stock picking", "error", err)
		return nil, err
	}

	return &picking, nil
}

// GetByID retrieves a stock picking by ID
func (r *StockPickingRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockPicking, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence_code, picking_type_id, location_id, location_dest_id, partner_id, date, scheduled_date, state, priority, origin, note, created_at, updated_at
		FROM stock_pickings
		WHERE id = $1
	`

	var picking types.StockPicking
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&picking.ID, &picking.OrganizationID, &picking.CompanyID, &picking.Name, &picking.SequenceCode, &picking.PickingTypeID, &picking.LocationID, &picking.LocationDestID, &picking.PartnerID, &picking.Date, &picking.ScheduledDate, &picking.State, &picking.Priority, &picking.Origin, &picking.Note, &picking.CreatedAt, &picking.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get stock picking", "error", err, "id", id)
		return nil, err
	}

	return &picking, nil
}

// List retrieves all stock pickings for an organization
func (r *StockPickingRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockPicking, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence_code, picking_type_id, location_id, location_dest_id, partner_id, date, scheduled_date, state, priority, origin, note, created_at, updated_at
		FROM stock_pickings
		WHERE organization_id = $1
		ORDER BY date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list stock pickings", "error", err)
		return nil, err
	}
	defer rows.Close()

	var pickings []types.StockPicking
	for rows.Next() {
		var picking types.StockPicking
		err := rows.Scan(
			&picking.ID, &picking.OrganizationID, &picking.CompanyID, &picking.Name, &picking.SequenceCode, &picking.PickingTypeID, &picking.LocationID, &picking.LocationDestID, &picking.PartnerID, &picking.Date, &picking.ScheduledDate, &picking.State, &picking.Priority, &picking.Origin, &picking.Note, &picking.CreatedAt, &picking.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan stock picking", "error", err)
			return nil, err
		}
		pickings = append(pickings, picking)
	}

	return pickings, nil
}

// Update updates a stock picking
func (r *StockPickingRepository) Update(ctx context.Context, id uuid.UUID, req types.StockPickingUpdateRequest) (*types.StockPicking, error) {
	query := `UPDATE stock_pickings SET `
	args := []interface{}{}
	argCount := 1

	if req.SequenceCode != nil {
		query += `sequence_code = $` + string(rune(argCount)) + `, `
		args = append(args, *req.SequenceCode)
		argCount++
	}

	if req.PickingTypeID != nil {
		query += `picking_type_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.PickingTypeID)
		argCount++
	}

	if req.LocationID != nil {
		query += `location_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.LocationID)
		argCount++
	}

	if req.LocationDestID != nil {
		query += `location_dest_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.LocationDestID)
		argCount++
	}

	if req.PartnerID != nil {
		query += `partner_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.PartnerID)
		argCount++
	}

	if req.Date != nil {
		query += `date = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Date)
		argCount++
	}

	if req.ScheduledDate != nil {
		query += `scheduled_date = $` + string(rune(argCount)) + `, `
		args = append(args, *req.ScheduledDate)
		argCount++
	}

	if req.State != nil {
		query += `state = $` + string(rune(argCount)) + `, `
		args = append(args, *req.State)
		argCount++
	}

	if req.Priority != nil {
		query += `priority = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Priority)
		argCount++
	}

	if req.Origin != nil {
		query += `origin = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Origin)
		argCount++
	}

	if req.Note != nil {
		query += `note = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Note)
		argCount++
	}

	if len(args) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += `, updated_at = NOW() WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, company_id, name, sequence_code, picking_type_id, location_id, location_dest_id, partner_id, date, scheduled_date, state, priority, origin, note, created_at, updated_at`
	args = append(args, id)

	var picking types.StockPicking
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&picking.ID, &picking.OrganizationID, &picking.CompanyID, &picking.Name, &picking.SequenceCode, &picking.PickingTypeID, &picking.LocationID, &picking.LocationDestID, &picking.PartnerID, &picking.Date, &picking.ScheduledDate, &picking.State, &picking.Priority, &picking.Origin, &picking.Note, &picking.CreatedAt, &picking.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to update stock picking", "error", err)
		return nil, err
	}

	return &picking, nil
}

// Delete deletes a stock picking
func (r *StockPickingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM stock_pickings
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete stock picking", "error", err)
		return err
	}

	return nil
}
