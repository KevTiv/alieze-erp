package repository

import (
	"context"
	"database/sql"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// CycleCountRepository interface for cycle counting operations
type CycleCountRepository interface {
	// Cycle count plan operations
	CreateCycleCountPlan(ctx context.Context, request types.CreateCycleCountPlanRequest) (*types.CycleCountPlan, error)
	GetCycleCountPlan(ctx context.Context, orgID uuid.UUID, planID uuid.UUID) (*types.CycleCountPlan, error)
	ListCycleCountPlans(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.CycleCountPlan, error)
	UpdateCycleCountPlanStatus(ctx context.Context, orgID uuid.UUID, planID uuid.UUID, status string) (bool, error)

	// Cycle count session operations
	CreateCycleCountSession(ctx context.Context, request types.CreateCycleCountSessionRequest) (*types.CycleCountSession, error)
	GetCycleCountSession(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (*types.CycleCountSession, error)
	ListCycleCountSessions(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.CycleCountSession, error)
	CompleteCycleCountSession(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (bool, error)

	// Cycle count line operations
	AddCycleCountLine(ctx context.Context, request types.AddCycleCountLineRequest) (*types.CycleCountLine, error)
	GetCycleCountLine(ctx context.Context, orgID uuid.UUID, lineID uuid.UUID) (*types.CycleCountLine, error)
	ListCycleCountLines(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) ([]types.CycleCountLine, error)
	VerifyCycleCountLine(ctx context.Context, request types.VerifyCycleCountLineRequest) (bool, error)

	// Cycle count adjustment operations
	CreateAdjustmentFromVariance(ctx context.Context, request types.CreateAdjustmentRequest) (*types.CycleCountAdjustment, error)
	GetCycleCountAdjustment(ctx context.Context, orgID uuid.UUID, adjustmentID uuid.UUID) (*types.CycleCountAdjustment, error)
	ListCycleCountAdjustments(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.CycleCountAdjustment, error)
	ApproveCycleCountAdjustment(ctx context.Context, request types.ApproveAdjustmentRequest) (bool, error)

	// Cycle count accuracy operations
	GetCycleCountAccuracyMetrics(ctx context.Context, request types.GetAccuracyMetricsRequest) (*types.CycleCountMetrics, error)
	GetProductsNeedingCycleCount(ctx context.Context, request types.GetProductsNeedingCountRequest) ([]types.ProductNeedingCycleCount, error)
	GetCycleCountAccuracyHistory(ctx context.Context, orgID uuid.UUID, productID *uuid.UUID, limit, offset int) ([]types.CycleCountAccuracy, error)
}

type cycleCountRepository struct {
	db *sql.DB
}

func NewCycleCountRepository(db *sql.DB) CycleCountRepository {
	return &cycleCountRepository{db: db}
}

// CreateCycleCountPlan creates a new cycle count plan
func (r *cycleCountRepository) CreateCycleCountPlan(ctx context.Context, request types.CreateCycleCountPlanRequest) (*types.CycleCountPlan, error) {
	query := `
		INSERT INTO inventory_cycle_count_plans (
			id, organization_id, name, description, frequency, abc_class,
			start_date, end_date, status, priority, created_by, assigned_to,
			created_at, updated_at, metadata
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW(), $12
		) RETURNING
			id, organization_id, name, description, frequency, abc_class,
			start_date, end_date, status, priority, created_by, assigned_to,
			created_at, updated_at, deleted_at, metadata
	`

	var plan types.CycleCountPlan
	err := r.db.QueryRowContext(ctx, query,
		request.OrganizationID, request.Name, request.Description, request.Frequency,
		request.ABCClass, request.StartDate, request.EndDate, "draft", request.Priority,
		request.CreatedBy, request.AssignedTo, request.Metadata,
	).Scan(
		&plan.ID, &plan.OrganizationID, &plan.Name, &plan.Description, &plan.Frequency,
		&plan.ABCClass, &plan.StartDate, &plan.EndDate, &plan.Status, &plan.Priority,
		&plan.CreatedBy, &plan.AssignedTo, &plan.CreatedAt, &plan.UpdatedAt,
		&plan.DeletedAt, &plan.Metadata,
	)

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// GetCycleCountPlan retrieves a cycle count plan
func (r *cycleCountRepository) GetCycleCountPlan(ctx context.Context, orgID uuid.UUID, planID uuid.UUID) (*types.CycleCountPlan, error) {
	query := `
		SELECT
			id, organization_id, name, description, frequency, abc_class,
			start_date, end_date, status, priority, created_by, assigned_to,
			created_at, updated_at, deleted_at, metadata
		FROM inventory_cycle_count_plans
		WHERE organization_id = $1 AND id = $2
	`

	var plan types.CycleCountPlan
	err := r.db.QueryRowContext(ctx, query, orgID, planID).Scan(
		&plan.ID, &plan.OrganizationID, &plan.Name, &plan.Description, &plan.Frequency,
		&plan.ABCClass, &plan.StartDate, &plan.EndDate, &plan.Status, &plan.Priority,
		&plan.CreatedBy, &plan.AssignedTo, &plan.CreatedAt, &plan.UpdatedAt,
		&plan.DeletedAt, &plan.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// ListCycleCountPlans retrieves cycle count plans for an organization
func (r *cycleCountRepository) ListCycleCountPlans(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.CycleCountPlan, error) {
	query := `
		SELECT
			id, organization_id, name, description, frequency, abc_class,
			start_date, end_date, status, priority, created_by, assigned_to,
			created_at, updated_at, deleted_at, metadata
		FROM inventory_cycle_count_plans
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	if status != nil {
		query += ` AND status = $` + string(len(params)+1)
		params = append(params, *status)
	}

	query += ` ORDER BY priority, created_at DESC LIMIT $` + string(len(params)+1) + ` OFFSET $` + string(len(params)+2)
	params = append(params, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []types.CycleCountPlan
	for rows.Next() {
		var plan types.CycleCountPlan
		err := rows.Scan(
			&plan.ID, &plan.OrganizationID, &plan.Name, &plan.Description, &plan.Frequency,
			&plan.ABCClass, &plan.StartDate, &plan.EndDate, &plan.Status, &plan.Priority,
			&plan.CreatedBy, &plan.AssignedTo, &plan.CreatedAt, &plan.UpdatedAt,
			&plan.DeletedAt, &plan.Metadata,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

// UpdateCycleCountPlanStatus updates the status of a cycle count plan
func (r *cycleCountRepository) UpdateCycleCountPlanStatus(ctx context.Context, orgID uuid.UUID, planID uuid.UUID, status string) (bool, error) {
	query := `
		UPDATE inventory_cycle_count_plans
		SET status = $3, updated_at = NOW()
		WHERE organization_id = $1 AND id = $2
	`

	result, err := r.db.ExecContext(ctx, query, orgID, planID, status)
	if err != nil {
		return false, err
	}

	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// CreateCycleCountSession creates a new cycle count session
func (r *cycleCountRepository) CreateCycleCountSession(ctx context.Context, request types.CreateCycleCountSessionRequest) (*types.CycleCountSession, error) {
	query := `
		INSERT INTO inventory_cycle_count_sessions (
			id, organization_id, plan_id, name, location_id, user_id,
			status, count_method, device_id, notes, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		) RETURNING
			id, organization_id, plan_id, name, location_id, user_id,
			start_time, end_time, status, count_method, device_id, notes,
			created_at, updated_at
	`

	var session types.CycleCountSession
	err := r.db.QueryRowContext(ctx, query,
		request.OrganizationID, request.PlanID, request.Name, request.LocationID,
		request.UserID, "in_progress", request.CountMethod, request.DeviceID, request.Notes,
	).Scan(
		&session.ID, &session.OrganizationID, &session.PlanID, &session.Name,
		&session.LocationID, &session.UserID, &session.StartTime, &session.EndTime,
		&session.Status, &session.CountMethod, &session.DeviceID, &session.Notes,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetCycleCountSession retrieves a cycle count session
func (r *cycleCountRepository) GetCycleCountSession(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (*types.CycleCountSession, error) {
	query := `
		SELECT
			id, organization_id, plan_id, name, location_id, user_id,
			start_time, end_time, status, count_method, device_id, notes,
			created_at, updated_at
		FROM inventory_cycle_count_sessions
		WHERE organization_id = $1 AND id = $2
	`

	var session types.CycleCountSession
	err := r.db.QueryRowContext(ctx, query, orgID, sessionID).Scan(
		&session.ID, &session.OrganizationID, &session.PlanID, &session.Name,
		&session.LocationID, &session.UserID, &session.StartTime, &session.EndTime,
		&session.Status, &session.CountMethod, &session.DeviceID, &session.Notes,
		&session.CreatedAt, &session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// ListCycleCountSessions retrieves cycle count sessions for an organization
func (r *cycleCountRepository) ListCycleCountSessions(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.CycleCountSession, error) {
	query := `
		SELECT
			id, organization_id, plan_id, name, location_id, user_id,
			start_time, end_time, status, count_method, device_id, notes,
			created_at, updated_at
		FROM inventory_cycle_count_sessions
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	if status != nil {
		query += ` AND status = $` + string(len(params)+1)
		params = append(params, *status)
	}

	query += ` ORDER BY start_time DESC LIMIT $` + string(len(params)+1) + ` OFFSET $` + string(len(params)+2)
	params = append(params, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []types.CycleCountSession
	for rows.Next() {
		var session types.CycleCountSession
		err := rows.Scan(
			&session.ID, &session.OrganizationID, &session.PlanID, &session.Name,
			&session.LocationID, &session.UserID, &session.StartTime, &session.EndTime,
			&session.Status, &session.CountMethod, &session.DeviceID, &session.Notes,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CompleteCycleCountSession marks a session as completed
func (r *cycleCountRepository) CompleteCycleCountSession(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (bool, error) {
	query := `
		UPDATE inventory_cycle_count_sessions
		SET status = 'completed', end_time = NOW(), updated_at = NOW()
		WHERE organization_id = $1 AND id = $2 AND status = 'in_progress'
	`

	result, err := r.db.ExecContext(ctx, query, orgID, sessionID)
	if err != nil {
		return false, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return false, nil
	}

	// Record accuracy history for each line in the session
	accuracyQuery := `
		INSERT INTO inventory_cycle_count_accuracy (
			id, organization_id, product_id, location_id, count_date,
			system_quantity, counted_quantity, variance, variance_percentage,
			count_method, counted_by, created_at
		)
		SELECT
			gen_random_uuid(), $1, product_id, location_id, DATE(NOW()),
			system_quantity, counted_quantity, variance, variance_percentage,
			'manual', counted_by, NOW()
		FROM inventory_cycle_count_lines
		WHERE session_id = $2 AND organization_id = $1
	`

	_, err = r.db.ExecContext(ctx, accuracyQuery, orgID, sessionID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// AddCycleCountLine adds a count line to a session
func (r *cycleCountRepository) AddCycleCountLine(ctx context.Context, request types.AddCycleCountLineRequest) (*types.CycleCountLine, error) {
	// Get current system quantity
	var systemQty float64
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM stock_quants
		WHERE organization_id = $1 AND product_id = $2 AND location_id = $3
		AND (lot_id = $4 OR ($4 IS NULL AND lot_id IS NULL))
		AND (package_id = $5 OR ($5 IS NULL AND package_id IS NULL))
	`

	err := r.db.QueryRowContext(ctx, query,
		request.OrganizationID, request.ProductID, request.LocationID,
		request.LotID, request.PackageID,
	).Scan(&systemQty)
	if err != nil {
		return nil, err
	}

	// Create count line
	lineQuery := `
		INSERT INTO inventory_cycle_count_lines (
			id, session_id, organization_id, product_id, location_id,
			counted_quantity, system_quantity, counted_by, lot_id, package_id,
			count_time, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), NOW()
		) RETURNING
			id, session_id, organization_id, product_id, location_id,
			counted_quantity, system_quantity, variance, variance_percentage,
			count_time, counted_by, created_at, updated_at
	`

	var line types.CycleCountLine
	err = r.db.QueryRowContext(ctx, lineQuery,
		request.SessionID, request.OrganizationID, request.ProductID, request.LocationID,
		request.CountedQuantity, systemQty, request.CountedBy, request.LotID, request.PackageID,
	).Scan(
		&line.ID, &line.SessionID, &line.OrganizationID, &line.ProductID, &line.LocationID,
		&line.CountedQuantity, &line.SystemQuantity, &line.Variance, &line.VariancePercentage,
		&line.CountTime, &line.CountedBy, &line.CreatedAt, &line.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &line, nil
}

// GetCycleCountLine retrieves a count line
func (r *cycleCountRepository) GetCycleCountLine(ctx context.Context, orgID uuid.UUID, lineID uuid.UUID) (*types.CycleCountLine, error) {
	query := `
		SELECT
			id, session_id, organization_id, product_id, location_id,
			counted_quantity, system_quantity, variance, variance_percentage,
			count_time, counted_by, verified_by, verification_time,
			status, notes, resolution_notes, created_at, updated_at
		FROM inventory_cycle_count_lines
		WHERE organization_id = $1 AND id = $2
	`

	var line types.CycleCountLine
	err := r.db.QueryRowContext(ctx, query, orgID, lineID).Scan(
		&line.ID, &line.SessionID, &line.OrganizationID, &line.ProductID, &line.LocationID,
		&line.CountedQuantity, &line.SystemQuantity, &line.Variance, &line.VariancePercentage,
		&line.CountTime, &line.CountedBy, &line.VerifiedBy, &line.VerificationTime,
		&line.Status, &line.Notes, &line.ResolutionNotes, &line.CreatedAt, &line.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &line, nil
}

// ListCycleCountLines retrieves count lines for a session
func (r *cycleCountRepository) ListCycleCountLines(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) ([]types.CycleCountLine, error) {
	query := `
		SELECT
			id, session_id, organization_id, product_id, location_id,
			counted_quantity, system_quantity, variance, variance_percentage,
			count_time, counted_by, verified_by, verification_time,
			status, notes, resolution_notes, created_at, updated_at
		FROM inventory_cycle_count_lines
		WHERE organization_id = $1 AND session_id = $2
		ORDER BY count_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []types.CycleCountLine
	for rows.Next() {
		var line types.CycleCountLine
		err := rows.Scan(
			&line.ID, &line.SessionID, &line.OrganizationID, &line.ProductID, &line.LocationID,
			&line.CountedQuantity, &line.SystemQuantity, &line.Variance, &line.VariancePercentage,
			&line.CountTime, &line.CountedBy, &line.VerifiedBy, &line.VerificationTime,
			&line.Status, &line.Notes, &line.ResolutionNotes, &line.CreatedAt, &line.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// VerifyCycleCountLine verifies a count line
func (r *cycleCountRepository) VerifyCycleCountLine(ctx context.Context, request types.VerifyCycleCountLineRequest) (bool, error) {
	query := `
		UPDATE inventory_cycle_count_lines
		SET
			status = $3,
			verified_by = $4,
			verification_time = NOW(),
			updated_at = NOW()
		WHERE organization_id = $1 AND id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		request.OrganizationID, request.LineID, request.Status, request.VerifiedBy,
	)
	if err != nil {
		return false, err
	}

	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// CreateAdjustmentFromVariance creates an adjustment from a count variance
func (r *cycleCountRepository) CreateAdjustmentFromVariance(ctx context.Context, request types.CreateAdjustmentRequest) (*types.CycleCountAdjustment, error) {
	// Get count line
	var line types.CycleCountLine
	query := `
		SELECT
			product_id, location_id, system_quantity, counted_quantity
		FROM inventory_cycle_count_lines
		WHERE organization_id = $1 AND id = $2
	`

	err := r.db.QueryRowContext(ctx, query, request.OrganizationID, request.LineID).Scan(
		&line.ProductID, &line.LocationID, &line.SystemQuantity, &line.CountedQuantity,
	)
	if err != nil {
		return nil, err
	}

	// Create adjustment
	adjustmentQuery := `
		INSERT INTO inventory_cycle_count_adjustments (
			id, organization_id, count_line_id, product_id, location_id,
			old_quantity, new_quantity, adjustment_type, reason,
			adjusted_by, status, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending', NOW(), NOW()
		) RETURNING
			id, organization_id, count_line_id, product_id, location_id,
			old_quantity, new_quantity, adjustment_quantity, adjustment_type,
			reason, adjusted_by, status, created_at, updated_at
	`

	var adjustment types.CycleCountAdjustment
	err = r.db.QueryRowContext(ctx, adjustmentQuery,
		request.OrganizationID, request.LineID, line.ProductID, line.LocationID,
		line.SystemQuantity, line.CountedQuantity, request.AdjustmentType, request.Reason,
		request.AdjustedBy,
	).Scan(
		&adjustment.ID, &adjustment.OrganizationID, &adjustment.CountLineID,
		&adjustment.ProductID, &adjustment.LocationID, &adjustment.OldQuantity,
		&adjustment.NewQuantity, &adjustment.AdjustmentQuantity, &adjustment.AdjustmentType,
		&adjustment.Reason, &adjustment.AdjustedBy, &adjustment.Status,
		&adjustment.CreatedAt, &adjustment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &adjustment, nil
}

// GetCycleCountAdjustment retrieves an adjustment
func (r *cycleCountRepository) GetCycleCountAdjustment(ctx context.Context, orgID uuid.UUID, adjustmentID uuid.UUID) (*types.CycleCountAdjustment, error) {
	query := `
		SELECT
			id, organization_id, count_line_id, product_id, location_id,
			old_quantity, new_quantity, adjustment_quantity, adjustment_type,
			reason, adjustment_time, adjusted_by, approved_by, approval_time,
			status, notes, created_at, updated_at
		FROM inventory_cycle_count_adjustments
		WHERE organization_id = $1 AND id = $2
	`

	var adjustment types.CycleCountAdjustment
	err := r.db.QueryRowContext(ctx, query, orgID, adjustmentID).Scan(
		&adjustment.ID, &adjustment.OrganizationID, &adjustment.CountLineID,
		&adjustment.ProductID, &adjustment.LocationID, &adjustment.OldQuantity,
		&adjustment.NewQuantity, &adjustment.AdjustmentQuantity, &adjustment.AdjustmentType,
		&adjustment.Reason, &adjustment.AdjustmentTime, &adjustment.AdjustedBy,
		&adjustment.ApprovedBy, &adjustment.ApprovalTime, &adjustment.Status,
		&adjustment.Notes, &adjustment.CreatedAt, &adjustment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &adjustment, nil
}

// ListCycleCountAdjustments retrieves adjustments for an organization
func (r *cycleCountRepository) ListCycleCountAdjustments(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.CycleCountAdjustment, error) {
	query := `
		SELECT
			id, organization_id, count_line_id, product_id, location_id,
			old_quantity, new_quantity, adjustment_quantity, adjustment_type,
			reason, adjustment_time, adjusted_by, approved_by, approval_time,
			status, notes, created_at, updated_at
		FROM inventory_cycle_count_adjustments
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	if status != nil {
		query += ` AND status = $` + string(len(params)+1)
		params = append(params, *status)
	}

	query += ` ORDER BY adjustment_time DESC LIMIT $` + string(len(params)+1) + ` OFFSET $` + string(len(params)+2)
	params = append(params, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []types.CycleCountAdjustment
	for rows.Next() {
		var adjustment types.CycleCountAdjustment
		err := rows.Scan(
			&adjustment.ID, &adjustment.OrganizationID, &adjustment.CountLineID,
			&adjustment.ProductID, &adjustment.LocationID, &adjustment.OldQuantity,
			&adjustment.NewQuantity, &adjustment.AdjustmentQuantity, &adjustment.AdjustmentType,
			&adjustment.Reason, &adjustment.AdjustmentTime, &adjustment.AdjustedBy,
			&adjustment.ApprovedBy, &adjustment.ApprovalTime, &adjustment.Status,
			&adjustment.Notes, &adjustment.CreatedAt, &adjustment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		adjustments = append(adjustments, adjustment)
	}

	return adjustments, nil
}

// ApproveCycleCountAdjustment approves an adjustment and updates stock
func (r *cycleCountRepository) ApproveCycleCountAdjustment(ctx context.Context, request types.ApproveAdjustmentRequest) (bool, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Get adjustment
	var adjustment types.CycleCountAdjustment
	query := `
		SELECT
			product_id, location_id, new_quantity
		FROM inventory_cycle_count_adjustments
		WHERE organization_id = $1 AND id = $2 AND status = 'pending'
		FOR UPDATE
	`

	err = tx.QueryRowContext(ctx, query, request.OrganizationID, request.AdjustmentID).Scan(
		&adjustment.ProductID, &adjustment.LocationID, &adjustment.NewQuantity,
	)
	if err != nil {
		return false, err
	}

	// Update adjustment status
	updateQuery := `
		UPDATE inventory_cycle_count_adjustments
		SET
			status = 'approved',
			approved_by = $3,
			approval_time = NOW(),
			updated_at = NOW()
		WHERE organization_id = $1 AND id = $2
	`

	_, err = tx.ExecContext(ctx, updateQuery,
		request.OrganizationID, request.AdjustmentID, request.ApprovedBy,
	)
	if err != nil {
		return false, err
	}

	// Update stock quantity
	stockQuery := `
		INSERT INTO stock_quants (
			id, organization_id, product_id, location_id, quantity,
			reserved_quantity, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, 0, NOW(), NOW()
		)
		ON CONFLICT (product_id, location_id, organization_id)
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			updated_at = NOW()
	`

	_, err = tx.ExecContext(ctx, stockQuery,
		request.OrganizationID, adjustment.ProductID, adjustment.LocationID, adjustment.NewQuantity,
	)
	if err != nil {
		return false, err
	}

	// Update count line status
	lineQuery := `
		UPDATE inventory_cycle_count_lines
		SET
			status = 'adjusted',
			updated_at = NOW()
		WHERE organization_id = $1 AND id = (
			SELECT count_line_id FROM inventory_cycle_count_adjustments
			WHERE organization_id = $1 AND id = $2
		)
	`

	_, err = tx.ExecContext(ctx, lineQuery, request.OrganizationID, request.AdjustmentID)
	if err != nil {
		return false, err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetCycleCountAccuracyMetrics retrieves accuracy metrics
func (r *cycleCountRepository) GetCycleCountAccuracyMetrics(ctx context.Context, request types.GetAccuracyMetricsRequest) (*types.CycleCountMetrics, error) {
	query := `
		SELECT
			COUNT(*) as total_counts,
			SUM(CASE WHEN variance_percentage = 0 THEN 1 ELSE 0 END) as accurate_counts,
			SUM(CASE WHEN variance_percentage != 0 THEN 1 ELSE 0 END) as variance_counts,
			COALESCE(SUM(CASE WHEN variance_percentage = 0 THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as accuracy_percentage,
			COALESCE(AVG(ABS(variance_percentage)), 0) as average_variance_percentage,
			COALESCE(SUM(variance), 0) as total_variance_quantity
		FROM inventory_cycle_count_accuracy
		WHERE organization_id = $1
		AND ($2 IS NULL OR count_date >= $2)
		AND ($3 IS NULL OR count_date <= $3)
	`

	var metrics types.CycleCountMetrics
	err := r.db.QueryRowContext(ctx, query,
		request.OrganizationID, request.DateFrom, request.DateTo,
	).Scan(
		&metrics.TotalCounts, &metrics.AccurateCounts, &metrics.VarianceCounts,
		&metrics.AccuracyPercentage, &metrics.AverageVariancePercentage,
		&metrics.TotalVarianceQuantity,
	)

	if err != nil {
		return nil, err
	}

	return &metrics, nil
}

// GetProductsNeedingCycleCount retrieves products that need cycle counting
func (r *cycleCountRepository) GetProductsNeedingCycleCount(ctx context.Context, request types.GetProductsNeedingCountRequest) ([]types.ProductNeedingCycleCount, error) {
	daysSince := 30
	minVariance := 5.0

	if request.DaysSinceLastCount != nil {
		daysSince = *request.DaysSinceLastCount
	}
	if request.MinVariancePercentage != nil {
		minVariance = *request.MinVariancePercentage
	}

	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			p.default_code,
			p.category_id,
			MAX(cca.count_date) as last_count_date,
			COALESCE(DATE_PART('day', NOW() - MAX(cca.count_date)), 999) as days_since_count,
			MAX(cca.variance_percentage) as last_variance_percentage,
			COALESCE(AVG(cca.variance_percentage), 0) as average_variance_percentage,
			CASE
				WHEN MAX(cca.count_date) IS NULL THEN 1
				WHEN DATE_PART('day', NOW() - MAX(cca.count_date)) > $2 THEN 2
				WHEN MAX(cca.variance_percentage) > $3 THEN 3
				ELSE 4
			END as count_priority
		FROM products p
		LEFT JOIN inventory_cycle_count_accuracy cca ON
			p.id = cca.product_id AND p.organization_id = cca.organization_id
		WHERE p.organization_id = $1
		GROUP BY p.id
		HAVING
			MAX(cca.count_date) IS NULL OR
			DATE_PART('day', NOW() - MAX(cca.count_date)) > $2 OR
			MAX(cca.variance_percentage) > $3
		ORDER BY count_priority, days_since_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query, request.OrganizationID, daysSince, minVariance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []types.ProductNeedingCycleCount
	for rows.Next() {
		var product types.ProductNeedingCycleCount
		err := rows.Scan(
			&product.ProductID, &product.ProductName, &product.DefaultCode,
			&product.CategoryID, &product.LastCountDate, &product.DaysSinceCount,
			&product.LastVariancePercentage, &product.AverageVariancePercentage,
			&product.CountPriority,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// GetCycleCountAccuracyHistory retrieves accuracy history
func (r *cycleCountRepository) GetCycleCountAccuracyHistory(ctx context.Context, orgID uuid.UUID, productID *uuid.UUID, limit, offset int) ([]types.CycleCountAccuracy, error) {
	query := `
		SELECT
			id, organization_id, product_id, location_id, count_date,
			system_quantity, counted_quantity, variance, variance_percentage,
			accuracy_score, count_method, counted_by, verified_by, created_at
		FROM inventory_cycle_count_accuracy
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	if productID != nil {
		query += ` AND product_id = $` + string(len(params)+1)
		params = append(params, *productID)
	}

	query += ` ORDER BY count_date DESC LIMIT $` + string(len(params)+1) + ` OFFSET $` + string(len(params)+2)
	params = append(params, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []types.CycleCountAccuracy
	for rows.Next() {
		var record types.CycleCountAccuracy
		err := rows.Scan(
			&record.ID, &record.OrganizationID, &record.ProductID, &record.LocationID,
			&record.CountDate, &record.SystemQuantity, &record.CountedQuantity,
			&record.Variance, &record.VariancePercentage, &record.AccuracyScore,
			&record.CountMethod, &record.CountedBy, &record.VerifiedBy, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, record)
	}

	return history, nil
}
