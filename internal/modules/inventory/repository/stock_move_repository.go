package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockMoveRepository interface defines the contract for stock move operations
type StockMoveRepository interface {
	Create(ctx context.Context, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error)
	CreateWithTx(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error)
	BulkCreate(ctx context.Context, orgID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error)
	BulkCreateWithTx(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error)
	GetByID(ctx context.Context, id uuid.UUID) (*types.StockMove, error)
	GetByPickingID(ctx context.Context, pickingID uuid.UUID) ([]types.StockMove, error)
	List(ctx context.Context, orgID uuid.UUID) ([]types.StockMove, error)
	Update(ctx context.Context, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error)
	UpdateWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error
	UpdateState(ctx context.Context, id uuid.UUID, state string) error
}

// stockMoveRepository implements StockMoveRepository
type stockMoveRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewStockMoveRepository creates a new StockMoveRepository
func NewStockMoveRepository(db *sql.DB, logger *slog.Logger) StockMoveRepository {
	return &stockMoveRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new stock move
func (r *stockMoveRepository) Create(ctx context.Context, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	return r.CreateWithTx(ctx, nil, orgID, req)
}

// CreateWithTx creates a new stock move within a transaction
func (r *stockMoveRepository) CreateWithTx(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, req types.StockMoveCreateRequest) (*types.StockMove, error) {
	query := `
		INSERT INTO stock_moves (organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		RETURNING id, organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at
	`

	var move types.StockMove
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, orgID, req.CompanyID, req.Name, req.Sequence, req.Priority, req.Date, req.ScheduledDate, req.State, req.ProductID, req.ProductUomID, req.LocationID, req.LocationDestID, req.PickingID, req.Quantity, req.ReservedQuantity, req.Note).Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, orgID, req.CompanyID, req.Name, req.Sequence, req.Priority, req.Date, req.ScheduledDate, req.State, req.ProductID, req.ProductUomID, req.LocationID, req.LocationDestID, req.PickingID, req.Quantity, req.ReservedQuantity, req.Note).Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
	}

	if err != nil {
		r.logger.Error("Failed to create stock move", "error", err)
		return nil, err
	}

	return &move, nil
}

// GetByID retrieves a stock move by ID
func (r *stockMoveRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.StockMove, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at
		FROM stock_moves
		WHERE id = $1
	`

	var move types.StockMove
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("Failed to get stock move", "error", err, "id", id)
		return nil, err
	}

	return &move, nil
}

// GetByPickingID retrieves all stock moves for a given picking ID
func (r *stockMoveRepository) GetByPickingID(ctx context.Context, pickingID uuid.UUID) ([]types.StockMove, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at
		FROM stock_moves
		WHERE picking_id = $1
		ORDER BY sequence ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pickingID)
	if err != nil {
		r.logger.Error("Failed to get stock moves by picking ID", "error", err, "picking_id", pickingID)
		return nil, err
	}
	defer rows.Close()

	var moves []types.StockMove
	for rows.Next() {
		var move types.StockMove
		err := rows.Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan stock move", "error", err)
			return nil, err
		}
		moves = append(moves, move)
	}

	return moves, nil
}

// List retrieves all stock moves for an organization
func (r *stockMoveRepository) List(ctx context.Context, orgID uuid.UUID) ([]types.StockMove, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at
		FROM stock_moves
		WHERE organization_id = $1
		ORDER BY date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		r.logger.Error("Failed to list stock moves", "error", err)
		return nil, err
	}
	defer rows.Close()

	var moves []types.StockMove
	for rows.Next() {
		var move types.StockMove
		err := rows.Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan stock move", "error", err)
			return nil, err
		}
		moves = append(moves, move)
	}

	return moves, nil
}

// Update updates a stock move
func (r *stockMoveRepository) Update(ctx context.Context, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error) {
	return r.UpdateWithTx(ctx, nil, id, req)
}

// UpdateWithTx updates a stock move within a transaction
func (r *stockMoveRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, req types.StockMoveUpdateRequest) (*types.StockMove, error) {
	query := `UPDATE stock_moves SET `
	args := []interface{}{}
	argCount := 1

	if req.Sequence != nil {
		query += `sequence = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Sequence)
		argCount++
	}

	if req.Priority != nil {
		query += `priority = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Priority)
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

	if req.ProductID != nil {
		query += `product_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.ProductID)
		argCount++
	}

	if req.ProductUomID != nil {
		query += `product_uom = $` + string(rune(argCount)) + `, `
		args = append(args, *req.ProductUomID)
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

	if req.PickingID != nil {
		query += `picking_id = $` + string(rune(argCount)) + `, `
		args = append(args, *req.PickingID)
		argCount++
	}

	if req.Quantity != nil {
		query += `quantity = $` + string(rune(argCount)) + `, `
		args = append(args, *req.Quantity)
		argCount++
	}

	if req.ReservedQuantity != nil {
		query += `reserved_quantity = $` + string(rune(argCount)) + `, `
		args = append(args, *req.ReservedQuantity)
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
	query += `, updated_at = NOW() WHERE id = $` + string(rune(argCount)) + ` RETURNING id, organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at`
	args = append(args, id)

	var move types.StockMove
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, args...).Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, args...).Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
	}

	if err != nil {
		r.logger.Error("Failed to update stock move", "error", err)
		return nil, err
	}

	return &move, nil
}

// Delete deletes a stock move
func (r *stockMoveRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DeleteWithTx(ctx, nil, id)
}

// DeleteWithTx deletes a stock move within a transaction
func (r *stockMoveRepository) DeleteWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM stock_moves
		WHERE id = $1
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, id)
	}

	if err != nil {
		r.logger.Error("Failed to delete stock move", "error", err)
		return err
	}

	return nil
}

// UpdateState updates the state of a stock move
func (r *stockMoveRepository) UpdateState(ctx context.Context, id uuid.UUID, state string) error {
	query := `
		UPDATE stock_moves
		SET state = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, state, id)
	if err != nil {
		r.logger.Error("Failed to update stock move state", "error", err, "id", id, "state", state)
		return err
	}

	return nil
}

// BulkCreate creates multiple stock moves in a single operation
func (r *stockMoveRepository) BulkCreate(ctx context.Context, orgID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error) {
	return r.BulkCreateWithTx(ctx, nil, orgID, reqs)
}

// BulkCreateWithTx creates multiple stock moves within a transaction
func (r *stockMoveRepository) BulkCreateWithTx(ctx context.Context, tx *sql.Tx, orgID uuid.UUID, reqs []types.StockMoveCreateRequest) ([]types.StockMove, error) {
	if len(reqs) == 0 {
		return []types.StockMove{}, nil
	}

	// Use a transaction if none provided
	useExternalTx := tx != nil
	var internalTx *sql.Tx
	var err error

	if !useExternalTx {
		internalTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			r.logger.Error("Failed to begin transaction for bulk create", "error", err)
			return nil, err
		}
		tx = internalTx
		defer tx.Rollback()
	}

	// Prepare bulk insert query
	query := `
		INSERT INTO stock_moves (organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at)
		VALUES %s
		RETURNING id, organization_id, company_id, name, sequence, priority, date, scheduled_date, state, product_id, product_uom_id, location_id, location_dest_id, picking_id, quantity, reserved_quantity, note, created_at, updated_at
	`

	// Build values and args for bulk insert
	var valueClauses []string
	var args []interface{}
	argCount := 1

	for i, req := range reqs {
		valueClauses = append(valueClauses, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, NOW(), NOW())",
			argCount, argCount+1, argCount+2, argCount+3, argCount+4, argCount+5, argCount+6, argCount+7,
			argCount+8, argCount+9, argCount+10, argCount+11, argCount+12, argCount+13, argCount+14, argCount+15,
		))

		args = append(args,
			orgID, req.CompanyID, req.Name, req.Sequence, req.Priority, req.Date, req.ScheduledDate, req.State,
			req.ProductID, req.ProductUomID, req.LocationID, req.LocationDestID, req.PickingID, req.Quantity, req.ReservedQuantity, req.Note,
		)

		argCount += 16

		// Add comma for all except last
		if i < len(reqs)-1 {
			valueClauses[i] += ","
		}
	}

	// Execute bulk insert
	query = fmt.Sprintf(query, strings.Join(valueClauses, " "))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to bulk create stock moves", "error", err)
		return nil, err
	}
	defer rows.Close()

	// Collect results
	var moves []types.StockMove
	for rows.Next() {
		var move types.StockMove
		err := rows.Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority, &move.Date, &move.ScheduledDate, &move.State, &move.ProductID, &move.ProductUOM, &move.LocationID, &move.LocationDestID, &move.PickingID, &move.Quantity, &move.ReservedQuantity, &move.Note, &move.CreatedAt, &move.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan stock move in bulk create", "error", err)
			return nil, err
		}
		moves = append(moves, move)
	}

	// Commit if we created our own transaction
	if !useExternalTx {
		if err := tx.Commit(); err != nil {
			r.logger.Error("Failed to commit bulk create transaction", "error", err)
			return nil, err
		}
	}

	return moves, nil
}
