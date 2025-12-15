package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice domain.Invoice) (*domain.Invoice, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error)
	FindAll(ctx context.Context, filters InvoiceFilter) ([]domain.Invoice, error)
	Update(ctx context.Context, invoice domain.Invoice) (*domain.Invoice, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]domain.Invoice, error)
	FindByStatus(ctx context.Context, status domain.InvoiceStatus) ([]domain.Invoice, error)
	FindByType(ctx context.Context, invoiceType domain.InvoiceType) ([]domain.Invoice, error)
}

type InvoiceFilter struct {
	PartnerID *uuid.UUID
	Status    *domain.InvoiceStatus
	Type      *domain.InvoiceType
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}

type invoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(ctx context.Context, invoice domain.Invoice) (*domain.Invoice, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create invoice
	query := `
		INSERT INTO invoices
		(id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
	`

	var createdInvoice domain.Invoice
	err = tx.QueryRowContext(ctx, query,
		invoice.ID, invoice.OrganizationID, invoice.CompanyID, invoice.PartnerID,
		invoice.Reference, invoice.Status, invoice.Type, invoice.InvoiceDate, invoice.DueDate,
		invoice.PaymentTermID, invoice.FiscalPositionID, invoice.CurrencyID, invoice.JournalID,
		invoice.AmountUntaxed, invoice.AmountTax, invoice.AmountTotal, invoice.AmountResidual,
		invoice.Note, invoice.CreatedAt, invoice.UpdatedAt, invoice.CreatedBy, invoice.UpdatedBy,
	).Scan(
		&createdInvoice.ID, &createdInvoice.OrganizationID, &createdInvoice.CompanyID,
		&createdInvoice.PartnerID, &createdInvoice.Reference, &createdInvoice.Status,
		&createdInvoice.Type, &createdInvoice.InvoiceDate, &createdInvoice.DueDate,
		&createdInvoice.PaymentTermID, &createdInvoice.FiscalPositionID, &createdInvoice.CurrencyID,
		&createdInvoice.JournalID, &createdInvoice.AmountUntaxed, &createdInvoice.AmountTax,
		&createdInvoice.AmountTotal, &createdInvoice.AmountResidual, &createdInvoice.Note,
		&createdInvoice.CreatedAt, &createdInvoice.UpdatedAt, &createdInvoice.CreatedBy,
		&createdInvoice.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create invoice lines
	for _, line := range invoice.Lines {
		lineQuery := `
			INSERT INTO invoice_lines
			(id, invoice_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 account_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			RETURNING id, invoice_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 account_id, created_at, updated_at
		`

		var createdLine domain.InvoiceLine
		err = tx.QueryRowContext(ctx, lineQuery,
			line.ID, createdInvoice.ID, line.ProductID, line.ProductName, line.Description,
			line.Quantity, line.UomID, line.UnitPrice, line.Discount, line.TaxID,
			line.PriceSubtotal, line.PriceTax, line.PriceTotal, line.Sequence,
			line.AccountID, line.CreatedAt, line.UpdatedAt,
		).Scan(
			&createdLine.ID, &createdLine.InvoiceID, &createdLine.ProductID, &createdLine.ProductName,
			&createdLine.Description, &createdLine.Quantity, &createdLine.UomID, &createdLine.UnitPrice,
			&createdLine.Discount, &createdLine.TaxID, &createdLine.PriceSubtotal, &createdLine.PriceTax,
			&createdLine.PriceTotal, &createdLine.Sequence, &createdLine.AccountID,
			&createdLine.CreatedAt, &createdLine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice line: %w", err)
		}
		createdInvoice.Lines = append(createdInvoice.Lines, createdLine)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &createdInvoice, nil
}

func (r *invoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	query := `
		SELECT id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
		FROM invoices
		WHERE id = $1
	`

	var invoice domain.Invoice
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invoice.ID, &invoice.OrganizationID, &invoice.CompanyID, &invoice.PartnerID,
		&invoice.Reference, &invoice.Status, &invoice.Type, &invoice.InvoiceDate,
		&invoice.DueDate, &invoice.PaymentTermID, &invoice.FiscalPositionID, &invoice.CurrencyID,
		&invoice.JournalID, &invoice.AmountUntaxed, &invoice.AmountTax, &invoice.AmountTotal,
		&invoice.AmountResidual, &invoice.Note, &invoice.CreatedAt, &invoice.UpdatedAt,
		&invoice.CreatedBy, &invoice.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	// Load lines
	lines, err := r.findLinesByInvoiceID(ctx, invoice.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice lines: %w", err)
	}
	invoice.Lines = lines

	// Load payments
	payments, err := r.findPaymentsByInvoiceID(ctx, invoice.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice payments: %w", err)
	}
	invoice.Payments = payments

	return &invoice, nil
}

func (r *invoiceRepository) findLinesByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]domain.InvoiceLine, error) {
	query := `
		SELECT id, invoice_id, product_id, product_name, description, quantity, uom_id,
		 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
		 account_id, created_at, updated_at
		FROM invoice_lines
		WHERE invoice_id = $1
		ORDER BY sequence
	`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice lines: %w", err)
	}
	defer rows.Close()

	var lines []domain.InvoiceLine
	for rows.Next() {
		var line domain.InvoiceLine
		err = rows.Scan(
			&line.ID, &line.InvoiceID, &line.ProductID, &line.ProductName,
			&line.Description, &line.Quantity, &line.UomID, &line.UnitPrice,
			&line.Discount, &line.TaxID, &line.PriceSubtotal, &line.PriceTax,
			&line.PriceTotal, &line.Sequence, &line.AccountID,
			&line.CreatedAt, &line.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice line: %w", err)
		}
		lines = append(lines, line)
	}

	return lines, nil
}

func (r *invoiceRepository) findPaymentsByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]domain.Payment, error) {
	query := `
		SELECT id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by
		FROM payments
		WHERE invoice_id = $1
		ORDER BY payment_date
	`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice payments: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err = rows.Scan(
			&payment.ID, &payment.OrganizationID, &payment.CompanyID, &payment.InvoiceID,
			&payment.PartnerID, &payment.PaymentDate, &payment.Amount, &payment.CurrencyID,
			&payment.JournalID, &payment.PaymentMethod, &payment.Reference, &payment.Note,
			&payment.CreatedAt, &payment.UpdatedAt, &payment.CreatedBy, &payment.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *invoiceRepository) FindAll(ctx context.Context, filters InvoiceFilter) ([]domain.Invoice, error) {
	query := `
		SELECT id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
		FROM invoices
		WHERE organization_id = $1
	`

	params := []interface{}{}
	paramIndex := 1

	// Apply filters
	if filters.PartnerID != nil {
		query += fmt.Sprintf(" AND partner_id = $%d", paramIndex+1)
		params = append(params, *filters.PartnerID)
		paramIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", paramIndex+1)
		params = append(params, *filters.Status)
		paramIndex++
	}

	if filters.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", paramIndex+1)
		params = append(params, *filters.Type)
		paramIndex++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND invoice_date >= $%d", paramIndex+1)
		params = append(params, *filters.DateFrom)
		paramIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND invoice_date <= $%d", paramIndex+1)
		params = append(params, *filters.DateTo)
		paramIndex++
	}

	query += fmt.Sprintf(" ORDER BY invoice_date DESC LIMIT $%d OFFSET $%d", paramIndex+1, paramIndex+2)
	params = append(params, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err = rows.Scan(
			&invoice.ID, &invoice.OrganizationID, &invoice.CompanyID, &invoice.PartnerID,
			&invoice.Reference, &invoice.Status, &invoice.Type, &invoice.InvoiceDate,
			&invoice.DueDate, &invoice.PaymentTermID, &invoice.FiscalPositionID, &invoice.CurrencyID,
			&invoice.JournalID, &invoice.AmountUntaxed, &invoice.AmountTax, &invoice.AmountTotal,
			&invoice.AmountResidual, &invoice.Note, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.CreatedBy, &invoice.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	// Load lines and payments for each invoice
	for i := range invoices {
		lines, err := r.findLinesByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		invoices[i].Lines = lines

		payments, err := r.findPaymentsByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice payments: %w", err)
		}
		invoices[i].Payments = payments
	}

	return invoices, nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoice domain.Invoice) (*domain.Invoice, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update invoice
	query := `
		UPDATE invoices
		SET partner_id = $1, reference = $2, status = $3, type = $4,
		 invoice_date = $5, due_date = $6, payment_term_id = $7, fiscal_position_id = $8,
		 currency_id = $9, journal_id = $10, amount_untaxed = $11, amount_tax = $12,
		 amount_total = $13, amount_residual = $14, note = $15,
		 updated_at = $16, updated_by = $17
		WHERE id = $18
		RETURNING id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
	`

	var updatedInvoice domain.Invoice
	err = tx.QueryRowContext(ctx, query,
		invoice.PartnerID, invoice.Reference, invoice.Status, invoice.Type,
		invoice.InvoiceDate, invoice.DueDate, invoice.PaymentTermID, invoice.FiscalPositionID,
		invoice.CurrencyID, invoice.JournalID, invoice.AmountUntaxed, invoice.AmountTax,
		invoice.AmountTotal, invoice.AmountResidual, invoice.Note,
		invoice.UpdatedAt, invoice.UpdatedBy, invoice.ID,
	).Scan(
		&updatedInvoice.ID, &updatedInvoice.OrganizationID, &updatedInvoice.CompanyID,
		&updatedInvoice.PartnerID, &updatedInvoice.Reference, &updatedInvoice.Status,
		&updatedInvoice.Type, &updatedInvoice.InvoiceDate, &updatedInvoice.DueDate,
		&updatedInvoice.PaymentTermID, &updatedInvoice.FiscalPositionID, &updatedInvoice.CurrencyID,
		&updatedInvoice.JournalID, &updatedInvoice.AmountUntaxed, &updatedInvoice.AmountTax,
		&updatedInvoice.AmountTotal, &updatedInvoice.AmountResidual, &updatedInvoice.Note,
		&updatedInvoice.CreatedAt, &updatedInvoice.UpdatedAt, &updatedInvoice.CreatedBy,
		&updatedInvoice.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Delete existing lines and create new ones
	_, err = tx.ExecContext(ctx, "DELETE FROM invoice_lines WHERE invoice_id = $1", invoice.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing invoice lines: %w", err)
	}

	// Create new lines
	for _, line := range invoice.Lines {
		lineQuery := `
			INSERT INTO invoice_lines
			(id, invoice_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 account_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			RETURNING id, invoice_id, product_id, product_name, description, quantity, uom_id,
			 unit_price, discount, tax_id, price_subtotal, price_tax, price_total, sequence,
			 account_id, created_at, updated_at
		`

		var createdLine domain.InvoiceLine
		err = tx.QueryRowContext(ctx, lineQuery,
			line.ID, updatedInvoice.ID, line.ProductID, line.ProductName, line.Description,
			line.Quantity, line.UomID, line.UnitPrice, line.Discount, line.TaxID,
			line.PriceSubtotal, line.PriceTax, line.PriceTotal, line.Sequence,
			line.AccountID, line.CreatedAt, line.UpdatedAt,
		).Scan(
			&createdLine.ID, &createdLine.InvoiceID, &createdLine.ProductID, &createdLine.ProductName,
			&createdLine.Description, &createdLine.Quantity, &createdLine.UomID, &createdLine.UnitPrice,
			&createdLine.Discount, &createdLine.TaxID, &createdLine.PriceSubtotal, &createdLine.PriceTax,
			&createdLine.PriceTotal, &createdLine.Sequence, &createdLine.AccountID,
			&createdLine.CreatedAt, &createdLine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice line: %w", err)
		}
		updatedInvoice.Lines = append(updatedInvoice.Lines, createdLine)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedInvoice, nil
}

func (r *invoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete lines first
	_, err = tx.ExecContext(ctx, "DELETE FROM invoice_lines WHERE invoice_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice lines: %w", err)
	}

	// Delete payments first
	_, err = tx.ExecContext(ctx, "DELETE FROM payments WHERE invoice_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice payments: %w", err)
	}

	// Delete invoice
	_, err = tx.ExecContext(ctx, "DELETE FROM invoices WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *invoiceRepository) FindByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]domain.Invoice, error) {
	query := `
		SELECT id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
		FROM invoices
		WHERE partner_id = $1
		ORDER BY invoice_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices by partner: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err = rows.Scan(
			&invoice.ID, &invoice.OrganizationID, &invoice.CompanyID, &invoice.PartnerID,
			&invoice.Reference, &invoice.Status, &invoice.Type, &invoice.InvoiceDate,
			&invoice.DueDate, &invoice.PaymentTermID, &invoice.FiscalPositionID, &invoice.CurrencyID,
			&invoice.JournalID, &invoice.AmountUntaxed, &invoice.AmountTax, &invoice.AmountTotal,
			&invoice.AmountResidual, &invoice.Note, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.CreatedBy, &invoice.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	// Load lines and payments for each invoice
	for i := range invoices {
		lines, err := r.findLinesByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		invoices[i].Lines = lines

		payments, err := r.findPaymentsByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice payments: %w", err)
		}
		invoices[i].Payments = payments
	}

	return invoices, nil
}

func (r *invoiceRepository) FindByStatus(ctx context.Context, status domain.InvoiceStatus) ([]domain.Invoice, error) {
	query := `
		SELECT id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
		FROM invoices
		WHERE status = $1
		ORDER BY invoice_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices by status: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err = rows.Scan(
			&invoice.ID, &invoice.OrganizationID, &invoice.CompanyID, &invoice.PartnerID,
			&invoice.Reference, &invoice.Status, &invoice.Type, &invoice.InvoiceDate,
			&invoice.DueDate, &invoice.PaymentTermID, &invoice.FiscalPositionID, &invoice.CurrencyID,
			&invoice.JournalID, &invoice.AmountUntaxed, &invoice.AmountTax, &invoice.AmountTotal,
			&invoice.AmountResidual, &invoice.Note, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.CreatedBy, &invoice.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	// Load lines and payments for each invoice
	for i := range invoices {
		lines, err := r.findLinesByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		invoices[i].Lines = lines

		payments, err := r.findPaymentsByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice payments: %w", err)
		}
		invoices[i].Payments = payments
	}

	return invoices, nil
}

func (r *invoiceRepository) FindByType(ctx context.Context, invoiceType domain.InvoiceType) ([]domain.Invoice, error) {
	query := `
		SELECT id, organization_id, company_id, partner_id, reference, status, type,
		 invoice_date, due_date, payment_term_id, fiscal_position_id, currency_id,
		 journal_id, amount_untaxed, amount_tax, amount_total, amount_residual, note,
		 created_at, updated_at, created_by, updated_by
		FROM invoices
		WHERE type = $1
		ORDER BY invoice_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, invoiceType)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices by type: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err = rows.Scan(
			&invoice.ID, &invoice.OrganizationID, &invoice.CompanyID, &invoice.PartnerID,
			&invoice.Reference, &invoice.Status, &invoice.Type, &invoice.InvoiceDate,
			&invoice.DueDate, &invoice.PaymentTermID, &invoice.FiscalPositionID, &invoice.CurrencyID,
			&invoice.JournalID, &invoice.AmountUntaxed, &invoice.AmountTax, &invoice.AmountTotal,
			&invoice.AmountResidual, &invoice.Note, &invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.CreatedBy, &invoice.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	// Load lines and payments for each invoice
	for i := range invoices {
		lines, err := r.findLinesByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		invoices[i].Lines = lines

		payments, err := r.findPaymentsByInvoiceID(ctx, invoices[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice payments: %w", err)
		}
		invoices[i].Payments = payments
	}

	return invoices, nil
}
