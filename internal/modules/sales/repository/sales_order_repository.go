package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/sales/types"

	"github.com/google/uuid"
)

type SalesOrderRepository interface {
	Create(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.SalesOrder, error)
	FindAll(ctx context.Context, filters SalesOrderFilter) ([]types.SalesOrder, error)
	Update(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]types.SalesOrder, error)
	FindByStatus(ctx context.Context, status types.SalesOrderStatus) ([]types.SalesOrder, error)
}

type SalesOrderFilter struct {
	CustomerID *uuid.UUID
	Status     *types.SalesOrderStatus
	DateFrom   *time.Time
	DateTo     *time.Time
	Limit      int
	Offset     int
}

type salesOrderRepository struct {
	db *sql.DB
}

func NewSalesOrderRepository(db *sql.DB) SalesOrderRepository {
	return &salesOrderRepository{db: db}
}

func (r *salesOrderRepository) Create(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create sales order
	query := `
		INSERT INTO sales_orders
		(id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by
	`

	var createdOrder types.SalesOrder
	err = tx.QueryRowContext(ctx, query,
		order.ID, order.OrganizationID, order.CompanyID, order.CustomerID, order.SalesTeamID,
		order.Reference, order.Status, order.OrderDate, order.ConfirmationDate, order.ValidityDate,
		order.PaymentTermID, order.FiscalPositionID, order.PricelistID, order.CurrencyID,
		order.AmountUntaxed, order.AmountTax, order.AmountTotal, order.Note,
		order.CreatedAt, order.UpdatedAt, order.CreatedBy, order.UpdatedBy,
	).Scan(
		&createdOrder.ID, &createdOrder.OrganizationID, &createdOrder.CompanyID, &createdOrder.CustomerID,
		&createdOrder.SalesTeamID, &createdOrder.Reference, &createdOrder.Status,
		&createdOrder.OrderDate, &createdOrder.ConfirmationDate, &createdOrder.ValidityDate,
		&createdOrder.PaymentTermID, &createdOrder.FiscalPositionID, &createdOrder.PricelistID,
		&createdOrder.CurrencyID, &createdOrder.AmountUntaxed, &createdOrder.AmountTax,
		&createdOrder.AmountTotal, &createdOrder.Note, &createdOrder.CreatedAt,
		&createdOrder.UpdatedAt, &createdOrder.CreatedBy, &createdOrder.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales order: %w", err)
	}

	// Create sales order lines
	for _, line := range order.Lines {
		lineQuery := `
			INSERT INTO sales_order_lines
			(id, sales_order_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING id, sales_order_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 created_at, updated_at
		`

		var createdLine types.SalesOrderLine
		err = tx.QueryRowContext(ctx, lineQuery,
			line.ID, createdOrder.ID, line.ProductID, line.ProductName, line.Description,
			line.Quantity, line.UomID, line.UnitPrice, line.Discount, line.TaxID,
			line.PriceSubtotal, line.PriceTax, line.PriceTotal, line.Sequence,
			line.CreatedAt, line.UpdatedAt,
		).Scan(
			&createdLine.ID, &createdLine.SalesOrderID, &createdLine.ProductID, &createdLine.ProductName,
			&createdLine.Description, &createdLine.Quantity, &createdLine.UomID, &createdLine.UnitPrice,
			&createdLine.Discount, &createdLine.TaxID, &createdLine.PriceSubtotal, &createdLine.PriceTax,
			&createdLine.PriceTotal, &createdLine.Sequence, &createdLine.CreatedAt, &createdLine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create sales order line: %w", err)
		}
		createdOrder.Lines = append(createdOrder.Lines, createdLine)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &createdOrder, nil
}

func (r *salesOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.SalesOrder, error) {
	query := `
		SELECT id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by
		FROM sales_orders
		WHERE id = $1
	`

	var order types.SalesOrder
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.OrganizationID, &order.CompanyID, &order.CustomerID,
		&order.SalesTeamID, &order.Reference, &order.Status,
		&order.OrderDate, &order.ConfirmationDate, &order.ValidityDate,
		&order.PaymentTermID, &order.FiscalPositionID, &order.PricelistID,
		&order.CurrencyID, &order.AmountUntaxed, &order.AmountTax,
		&order.AmountTotal, &order.Note, &order.CreatedAt,
		&order.UpdatedAt, &order.CreatedBy, &order.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find sales order: %w", err)
	}

	// Load lines
	lines, err := r.findLinesByOrderID(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load sales order lines: %w", err)
	}
	order.Lines = lines

	return &order, nil
}

func (r *salesOrderRepository) findLinesByOrderID(ctx context.Context, orderID uuid.UUID) ([]types.SalesOrderLine, error) {
	query := `
		SELECT id, sales_order_id, product_id, product_name, description, quantity, uom_id,
		 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
		 created_at, updated_at
		FROM sales_order_lines
		WHERE sales_order_id = $1
		ORDER BY sequence
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales order lines: %w", err)
	}
	defer rows.Close()

	var lines []types.SalesOrderLine
	for rows.Next() {
		var line types.SalesOrderLine
		err = rows.Scan(
			&line.ID, &line.SalesOrderID, &line.ProductID, &line.ProductName,
			&line.Description, &line.Quantity, &line.UomID, &line.UnitPrice,
			&line.Discount, &line.TaxID, &line.PriceSubtotal, &line.PriceTax,
			&line.PriceTotal, &line.Sequence, &line.CreatedAt, &line.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales order line: %w", err)
		}
		lines = append(lines, line)
	}

	return lines, nil
}

func (r *salesOrderRepository) FindAll(ctx context.Context, filters SalesOrderFilter) ([]types.SalesOrder, error) {
	query := `
		SELECT id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by
		FROM sales_orders
		WHERE organization_id = $1
	`

	params := []interface{}{}
	paramIndex := 1

	// Apply filters
	if filters.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", paramIndex+1)
		params = append(params, *filters.CustomerID)
		paramIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", paramIndex+1)
		params = append(params, *filters.Status)
		paramIndex++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND order_date >= $%d", paramIndex+1)
		params = append(params, *filters.DateFrom)
		paramIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND order_date <= $%d", paramIndex+1)
		params = append(params, *filters.DateTo)
		paramIndex++
	}

	query += fmt.Sprintf(" ORDER BY order_date DESC LIMIT $%d OFFSET $%d", paramIndex+1, paramIndex+2)
	params = append(params, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales orders: %w", err)
	}
	defer rows.Close()

	var orders []types.SalesOrder
	for rows.Next() {
		var order types.SalesOrder
		err = rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.CustomerID,
			&order.SalesTeamID, &order.Reference, &order.Status,
			&order.OrderDate, &order.ConfirmationDate, &order.ValidityDate,
			&order.PaymentTermID, &order.FiscalPositionID, &order.PricelistID,
			&order.CurrencyID, &order.AmountUntaxed, &order.AmountTax,
			&order.AmountTotal, &order.Note, &order.CreatedAt,
			&order.UpdatedAt, &order.CreatedBy, &order.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales order: %w", err)
		}
		orders = append(orders, order)
	}

	// Load lines for each order
	for i := range orders {
		lines, err := r.findLinesByOrderID(ctx, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load sales order lines: %w", err)
		}
		orders[i].Lines = lines
	}

	return orders, nil
}

func (r *salesOrderRepository) Update(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update sales order
	query := `
		UPDATE sales_orders
		SET customer_id = $1, sales_team_id = $2, reference = $3, status = $4,
		 order_date = $5, confirmation_date = $6, validity_date = $7,
		 payment_term_id = $8, fiscal_position_id = $9, pricelist_id = $10,
		 currency_id = $11, amount_untaxed = $12, amount_tax = $13, amount_total = $14,
		 note = $15, updated_at = $16, updated_by = $17
		WHERE id = $18
		RETURNING id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by
	`

	var updatedOrder types.SalesOrder
	err = tx.QueryRowContext(ctx, query,
		order.CustomerID, order.SalesTeamID, order.Reference, order.Status,
		order.OrderDate, order.ConfirmationDate, order.ValidityDate,
		order.PaymentTermID, order.FiscalPositionID, order.PricelistID,
		order.CurrencyID, order.AmountUntaxed, order.AmountTax, order.AmountTotal,
		order.Note, order.UpdatedAt, order.UpdatedBy, order.ID,
	).Scan(
		&updatedOrder.ID, &updatedOrder.OrganizationID, &updatedOrder.CompanyID, &updatedOrder.CustomerID,
		&updatedOrder.SalesTeamID, &updatedOrder.Reference, &updatedOrder.Status,
		&updatedOrder.OrderDate, &updatedOrder.ConfirmationDate, &updatedOrder.ValidityDate,
		&updatedOrder.PaymentTermID, &updatedOrder.FiscalPositionID, &updatedOrder.PricelistID,
		&updatedOrder.CurrencyID, &updatedOrder.AmountUntaxed, &updatedOrder.AmountTax,
		&updatedOrder.AmountTotal, &updatedOrder.Note, &updatedOrder.CreatedAt,
		&updatedOrder.UpdatedAt, &updatedOrder.CreatedBy, &updatedOrder.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales order: %w", err)
	}

	// Delete existing lines and create new ones
	_, err = tx.ExecContext(ctx, "DELETE FROM sales_order_lines WHERE sales_order_id = $1", order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing sales order lines: %w", err)
	}

	// Create new lines
	for _, line := range order.Lines {
		lineQuery := `
			INSERT INTO sales_order_lines
			(id, sales_order_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING id, sales_order_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 created_at, updated_at
		`

		var createdLine types.SalesOrderLine
		err = tx.QueryRowContext(ctx, lineQuery,
			line.ID, updatedOrder.ID, line.ProductID, line.ProductName, line.Description,
			line.Quantity, line.UomID, line.UnitPrice, line.Discount, line.TaxID,
			line.PriceSubtotal, line.PriceTax, line.PriceTotal, line.Sequence,
			line.CreatedAt, line.UpdatedAt,
		).Scan(
			&createdLine.ID, &createdLine.SalesOrderID, &createdLine.ProductID, &createdLine.ProductName,
			&createdLine.Description, &createdLine.Quantity, &createdLine.UomID, &createdLine.UnitPrice,
			&createdLine.Discount, &createdLine.TaxID, &createdLine.PriceSubtotal, &createdLine.PriceTax,
			&createdLine.PriceTotal, &createdLine.Sequence, &createdLine.CreatedAt, &createdLine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create sales order line: %w", err)
		}
		updatedOrder.Lines = append(updatedOrder.Lines, createdLine)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedOrder, nil
}

func (r *salesOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete lines first
	_, err = tx.ExecContext(ctx, "DELETE FROM sales_order_lines WHERE sales_order_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete sales order lines: %w", err)
	}

	// Delete order
	_, err = tx.ExecContext(ctx, "DELETE FROM sales_orders WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete sales order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *salesOrderRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]types.SalesOrder, error) {
	query := `
		SELECT id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by
		FROM sales_orders
		WHERE customer_id = $1
		ORDER BY order_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales orders by customer: %w", err)
	}
	defer rows.Close()

	var orders []types.SalesOrder
	for rows.Next() {
		var order types.SalesOrder
		err = rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.CustomerID,
			&order.SalesTeamID, &order.Reference, &order.Status,
			&order.OrderDate, &order.ConfirmationDate, &order.ValidityDate,
			&order.PaymentTermID, &order.FiscalPositionID, &order.PricelistID,
			&order.CurrencyID, &order.AmountUntaxed, &order.AmountTax,
			&order.AmountTotal, &order.Note, &order.CreatedAt,
			&order.UpdatedAt, &order.CreatedBy, &order.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales order: %w", err)
		}
		orders = append(orders, order)
	}

	// Load lines for each order
	for i := range orders {
		lines, err := r.findLinesByOrderID(ctx, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load sales order lines: %w", err)
		}
		orders[i].Lines = lines
	}

	return orders, nil
}

func (r *salesOrderRepository) FindByStatus(ctx context.Context, status types.SalesOrderStatus) ([]types.SalesOrder, error) {
	query := `
		SELECT id, organization_id, company_id, customer_id, sales_team_id, reference, status,
		 order_date, confirmation_date, validity_date, payment_term_id, fiscal_position_id,
		 pricelist_id, currency_id, amount_untaxed, amount_tax, amount_total, note,
		 created_at, updated_at, created_by, updated_by
		FROM sales_orders
		WHERE status = $1
		ORDER BY order_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales orders by status: %w", err)
	}
	defer rows.Close()

	var orders []types.SalesOrder
	for rows.Next() {
		var order types.SalesOrder
		err = rows.Scan(
			&order.ID, &order.OrganizationID, &order.CompanyID, &order.CustomerID,
			&order.SalesTeamID, &order.Reference, &order.Status,
			&order.OrderDate, &order.ConfirmationDate, &order.ValidityDate,
			&order.PaymentTermID, &order.FiscalPositionID, &order.PricelistID,
			&order.CurrencyID, &order.AmountUntaxed, &order.AmountTax,
			&order.AmountTotal, &order.Note, &order.CreatedAt,
			&order.UpdatedAt, &order.CreatedBy, &order.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sales order: %w", err)
		}
		orders = append(orders, order)
	}

	// Load lines for each order
	for i := range orders {
		lines, err := r.findLinesByOrderID(ctx, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load sales order lines: %w", err)
		}
		orders[i].Lines = lines
	}

	return orders, nil
}
