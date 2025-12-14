package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/domain"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment domain.Payment) (*domain.Payment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	FindAll(ctx context.Context, filters PaymentFilter) ([]domain.Payment, error)
	Update(ctx context.Context, payment domain.Payment) (*domain.Payment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]domain.Payment, error)
	FindByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]domain.Payment, error)
}

type PaymentFilter struct {
	InvoiceID *uuid.UUID
	PartnerID *uuid.UUID
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment domain.Payment) (*domain.Payment, error) {
	query := `
		INSERT INTO payments
		(id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by
	`

	var createdPayment domain.Payment
	err := r.db.QueryRowContext(ctx, query,
		payment.ID, payment.OrganizationID, payment.CompanyID, payment.InvoiceID,
		payment.PartnerID, payment.PaymentDate, payment.Amount, payment.CurrencyID,
		payment.JournalID, payment.PaymentMethod, payment.Reference, payment.Note,
		payment.CreatedAt, payment.UpdatedAt, payment.CreatedBy, payment.UpdatedBy,
	).Scan(
		&createdPayment.ID, &createdPayment.OrganizationID, &createdPayment.CompanyID,
		&createdPayment.InvoiceID, &createdPayment.PartnerID, &createdPayment.PaymentDate,
		&createdPayment.Amount, &createdPayment.CurrencyID, &createdPayment.JournalID,
		&createdPayment.PaymentMethod, &createdPayment.Reference, &createdPayment.Note,
		&createdPayment.CreatedAt, &createdPayment.UpdatedAt, &createdPayment.CreatedBy,
		&createdPayment.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return &createdPayment, nil
}

func (r *paymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by
		FROM payments
		WHERE id = $1
	`

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID, &payment.OrganizationID, &payment.CompanyID,
		&payment.InvoiceID, &payment.PartnerID, &payment.PaymentDate,
		&payment.Amount, &payment.CurrencyID, &payment.JournalID,
		&payment.PaymentMethod, &payment.Reference, &payment.Note,
		&payment.CreatedAt, &payment.UpdatedAt, &payment.CreatedBy,
		&payment.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

func (r *paymentRepository) FindAll(ctx context.Context, filters PaymentFilter) ([]domain.Payment, error) {
	query := `
		SELECT id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by
		FROM payments
		WHERE organization_id = $1
	`

	params := []interface{}{}
	paramIndex := 1

	// Apply filters
	if filters.InvoiceID != nil {
		query += fmt.Sprintf(" AND invoice_id = $%d", paramIndex+1)
		params = append(params, *filters.InvoiceID)
		paramIndex++
	}

	if filters.PartnerID != nil {
		query += fmt.Sprintf(" AND partner_id = $%d", paramIndex+1)
		params = append(params, *filters.PartnerID)
		paramIndex++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND payment_date >= $%d", paramIndex+1)
		params = append(params, *filters.DateFrom)
		paramIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND payment_date <= $%d", paramIndex+1)
		params = append(params, *filters.DateTo)
		paramIndex++
	}

	query += fmt.Sprintf(" ORDER BY payment_date DESC LIMIT $%d OFFSET $%d", paramIndex+1, paramIndex+2)
	params = append(params, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err = rows.Scan(
			&payment.ID, &payment.OrganizationID, &payment.CompanyID,
			&payment.InvoiceID, &payment.PartnerID, &payment.PaymentDate,
			&payment.Amount, &payment.CurrencyID, &payment.JournalID,
			&payment.PaymentMethod, &payment.Reference, &payment.Note,
			&payment.CreatedAt, &payment.UpdatedAt, &payment.CreatedBy,
			&payment.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment domain.Payment) (*domain.Payment, error) {
	query := `
		UPDATE payments
		SET invoice_id = $1, partner_id = $2, payment_date = $3, amount = $4,
		 currency_id = $5, journal_id = $6, payment_method = $7, reference = $8,
		 note = $9, updated_at = $10, updated_by = $11
		WHERE id = $12
		RETURNING id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by
	`

	var updatedPayment domain.Payment
	err := r.db.QueryRowContext(ctx, query,
		payment.InvoiceID, payment.PartnerID, payment.PaymentDate, payment.Amount,
		payment.CurrencyID, payment.JournalID, payment.PaymentMethod, payment.Reference,
		payment.Note, payment.UpdatedAt, payment.UpdatedBy, payment.ID,
	).Scan(
		&updatedPayment.ID, &updatedPayment.OrganizationID, &updatedPayment.CompanyID,
		&updatedPayment.InvoiceID, &updatedPayment.PartnerID, &updatedPayment.PaymentDate,
		&updatedPayment.Amount, &updatedPayment.CurrencyID, &updatedPayment.JournalID,
		&updatedPayment.PaymentMethod, &updatedPayment.Reference, &updatedPayment.Note,
		&updatedPayment.CreatedAt, &updatedPayment.UpdatedAt, &updatedPayment.CreatedBy,
		&updatedPayment.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return &updatedPayment, nil
}

func (r *paymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM payments WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	return nil
}

func (r *paymentRepository) FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]domain.Payment, error) {
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
		return nil, fmt.Errorf("failed to query payments by invoice: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err = rows.Scan(
			&payment.ID, &payment.OrganizationID, &payment.CompanyID,
			&payment.InvoiceID, &payment.PartnerID, &payment.PaymentDate,
			&payment.Amount, &payment.CurrencyID, &payment.JournalID,
			&payment.PaymentMethod, &payment.Reference, &payment.Note,
			&payment.CreatedAt, &payment.UpdatedAt, &payment.CreatedBy,
			&payment.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *paymentRepository) FindByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]domain.Payment, error) {
	query := `
		SELECT id, organization_id, company_id, invoice_id, partner_id, payment_date,
		 amount, currency_id, journal_id, payment_method, reference, note,
		 created_at, updated_at, created_by, updated_by
		FROM payments
		WHERE partner_id = $1
		ORDER BY payment_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments by partner: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err = rows.Scan(
			&payment.ID, &payment.OrganizationID, &payment.CompanyID,
			&payment.InvoiceID, &payment.PartnerID, &payment.PaymentDate,
			&payment.Amount, &payment.CurrencyID, &payment.JournalID,
			&payment.PaymentMethod, &payment.Reference, &payment.Note,
			&payment.CreatedAt, &payment.UpdatedAt, &payment.CreatedBy,
			&payment.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}
