package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// ReplenishmentOrderRepository interface
type ReplenishmentOrderRepository interface {
	Create(ctx context.Context, order types.ReplenishmentOrder) (*types.ReplenishmentOrder, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.ReplenishmentOrder, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.ReplenishmentOrder, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.ReplenishmentOrder, error)
	FindByRule(ctx context.Context, organizationID, ruleID uuid.UUID) ([]types.ReplenishmentOrder, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.ReplenishmentOrder, error)
	Update(ctx context.Context, order types.ReplenishmentOrder) (*types.ReplenishmentOrder, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ProcessReplenishmentOrders(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.ReplenishmentOrder, error)
	RunReplenishmentCycle(ctx context.Context, organizationID uuid.UUID) (map[string]interface{}, error)
}

type replenishmentOrderRepository struct {
	db *sql.DB
}

func NewReplenishmentOrderRepository(db *sql.DB) ReplenishmentOrderRepository {
	return &replenishmentOrderRepository{db: db}
}

func (r *replenishmentOrderRepository) Create(ctx context.Context, order types.ReplenishmentOrder) (*types.ReplenishmentOrder, error) {
	query := `
		INSERT INTO replenishment_orders
		(id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
	`

	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = now
	}

	var created types.ReplenishmentOrder
	err := r.db.QueryRowContext(ctx, query,
		order.ID, order.OrganizationID, order.CompanyID, order.RuleID, order.ProductID,
		order.ProductName, order.Quantity, order.UOMID, order.SourceLocationID, order.DestLocationID,
		order.Status, order.Priority, order.ScheduledDate, order.ProcureMethod,
		order.Reference, order.Notes, order.CreatedAt, order.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.RuleID, &created.ProductID,
		&created.ProductName, &created.Quantity, &created.UOMID, &created.SourceLocationID, &created.DestLocationID,
		&created.Status, &created.Priority, &created.ScheduledDate, &created.ProcureMethod,
		&created.Reference, &created.Notes, &created.PurchaseOrderID, &created.ManufacturingOrderID,
		&created.TransferID, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create replenishment order: %w", err)
	}
	return &created, nil
}

func (r *replenishmentOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.ReplenishmentOrder, error) {
	query := `
		SELECT id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
		FROM replenishment_orders WHERE id = $1 AND deleted_at IS NULL
	`

	var order types.ReplenishmentOrder
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.OrganizationID, &order.CompanyID, &order.RuleID, &order.ProductID,
		&order.ProductName, &order.Quantity, &order.UOMID, &order.SourceLocationID, &order.DestLocationID,
		&order.Status, &order.Priority, &order.ScheduledDate, &order.ProcureMethod,
		&order.Reference, &order.Notes, &order.PurchaseOrderID, &order.ManufacturingOrderID,
		&order.TransferID, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment order: %w", err)
	}
	return &order, nil
}

func (r *replenishmentOrderRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.ReplenishmentOrder, error) {
	query := `
		SELECT id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
		FROM replenishment_orders WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC, priority ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment orders: %w", err)
	}
	defer rows.Close()

	var orders []types.ReplenishmentOrder
	for rows.Next() {
		var order types.ReplenishmentOrder
		err := rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.RuleID, &order.ProductID,
			&order.ProductName, &order.Quantity, &order.UOMID, &order.SourceLocationID, &order.DestLocationID,
			&order.Status, &order.Priority, &order.ScheduledDate, &order.ProcureMethod,
			&order.Reference, &order.Notes, &order.PurchaseOrderID, &order.ManufacturingOrderID,
			&order.TransferID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment order: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *replenishmentOrderRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.ReplenishmentOrder, error) {
	query := `
		SELECT id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
		FROM replenishment_orders WHERE organization_id = $1 AND product_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC, priority ASC
		LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment orders by product: %w", err)
	}
	defer rows.Close()

	var orders []types.ReplenishmentOrder
	for rows.Next() {
		var order types.ReplenishmentOrder
		err := rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.RuleID, &order.ProductID,
			&order.ProductName, &order.Quantity, &order.UOMID, &order.SourceLocationID, &order.DestLocationID,
			&order.Status, &order.Priority, &order.ScheduledDate, &order.ProcureMethod,
			&order.Reference, &order.Notes, &order.PurchaseOrderID, &order.ManufacturingOrderID,
			&order.TransferID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment order: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *replenishmentOrderRepository) FindByRule(ctx context.Context, organizationID, ruleID uuid.UUID) ([]types.ReplenishmentOrder, error) {
	query := `
		SELECT id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
		FROM replenishment_orders WHERE organization_id = $1 AND rule_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC, priority ASC
		LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment orders by rule: %w", err)
	}
	defer rows.Close()

	var orders []types.ReplenishmentOrder
	for rows.Next() {
		var order types.ReplenishmentOrder
		err := rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.RuleID, &order.ProductID,
			&order.ProductName, &order.Quantity, &order.UOMID, &order.SourceLocationID, &order.DestLocationID,
			&order.Status, &order.Priority, &order.ScheduledDate, &order.ProcureMethod,
			&order.Reference, &order.Notes, &order.PurchaseOrderID, &order.ManufacturingOrderID,
			&order.TransferID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment order: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *replenishmentOrderRepository) FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.ReplenishmentOrder, error) {
	query := `
		SELECT id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
		FROM replenishment_orders WHERE organization_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, created_at ASC
		LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment orders by status: %w", err)
	}
	defer rows.Close()

	var orders []types.ReplenishmentOrder
	for rows.Next() {
		var order types.ReplenishmentOrder
		err := rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.RuleID, &order.ProductID,
			&order.ProductName, &order.Quantity, &order.UOMID, &order.SourceLocationID, &order.DestLocationID,
			&order.Status, &order.Priority, &order.ScheduledDate, &order.ProcureMethod,
			&order.Reference, &order.Notes, &order.PurchaseOrderID, &order.ManufacturingOrderID,
			&order.TransferID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment order: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *replenishmentOrderRepository) Update(ctx context.Context, order types.ReplenishmentOrder) (*types.ReplenishmentOrder, error) {
	query := `
		UPDATE replenishment_orders
		SET company_id = $2, rule_id = $3, product_name = $4, quantity = $5, uom_id = $6,
		 source_location_id = $7, dest_location_id = $8, status = $9, priority = $10,
		 scheduled_date = $11, procure_method = $12, reference = $13, notes = $14,
		 purchase_order_id = $15, manufacturing_order_id = $16, transfer_id = $17, updated_at = $18
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, rule_id, product_id, product_name, quantity, uom_id,
		 source_location_id, dest_location_id, status, priority, scheduled_date, procure_method,
		 reference, notes, purchase_order_id, manufacturing_order_id, transfer_id, created_at, updated_at
	`

	order.UpdatedAt = time.Now()
	var updated types.ReplenishmentOrder
	err := r.db.QueryRowContext(ctx, query,
		order.ID, order.CompanyID, order.RuleID, order.ProductName, order.Quantity, order.UOMID,
		order.SourceLocationID, order.DestLocationID, order.Status, order.Priority,
		order.ScheduledDate, order.ProcureMethod, order.Reference, order.Notes,
		order.PurchaseOrderID, order.ManufacturingOrderID, order.TransferID, order.UpdatedAt,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.RuleID, &updated.ProductID,
		&updated.ProductName, &updated.Quantity, &updated.UOMID, &updated.SourceLocationID, &updated.DestLocationID,
		&updated.Status, &updated.Priority, &updated.ScheduledDate, &updated.ProcureMethod,
		&updated.Reference, &updated.Notes, &updated.PurchaseOrderID, &updated.ManufacturingOrderID,
		&updated.TransferID, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("replenishment order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update replenishment order: %w", err)
	}
	return &updated, nil
}

func (r *replenishmentOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE replenishment_orders SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete replenishment order: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("replenishment order not found")
	}
	return nil
}

func (r *replenishmentOrderRepository) ProcessReplenishmentOrders(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.ReplenishmentOrder, error) {
	query := `
		SELECT order_id, product_id, product_name, quantity, status, procure_method
		FROM process_replenishment_orders($1, $2)
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to process replenishment orders: %w", err)
	}
	defer rows.Close()

	var orders []types.ReplenishmentOrder
	for rows.Next() {
		var order types.ReplenishmentOrder
		var orderID, procureMethod sql.NullString
		var status sql.NullString

		err := rows.Scan(
			&orderID, &order.ProductID, &order.ProductName, &order.Quantity, &status, &procureMethod,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan processed replenishment order: %w", err)
		}

		if orderID.Valid {
			order.ID = uuid.MustParse(orderID.String)
		}
		if status.Valid {
			order.Status = status.String
		}
		if procureMethod.Valid {
			order.ProcureMethod = procureMethod.String
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *replenishmentOrderRepository) RunReplenishmentCycle(ctx context.Context, organizationID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT action, count, message
		FROM run_replenishment_cycle($1)
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to run replenishment cycle: %w", err)
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var action, message sql.NullString
		var count sql.NullInt64

		err := rows.Scan(&action, &count, &message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment cycle result: %w", err)
		}

		if action.Valid {
			if count.Valid {
				result[action.String] = count.Int64
			}
			if message.Valid {
				result[action.String+"_message"] = message.String
			}
		}
	}

	return result, nil
}
