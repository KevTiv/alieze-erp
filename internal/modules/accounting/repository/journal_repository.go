package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
)

type JournalRepository interface {
	Create(ctx context.Context, journal domain.Journal) (*domain.Journal, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Journal, error)
	FindAll(ctx context.Context, filters JournalFilter) ([]domain.Journal, error)
	Update(ctx context.Context, journal domain.Journal) (*domain.Journal, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByCode(ctx context.Context, organizationID uuid.UUID, code string) (*domain.Journal, error)
	FindByType(ctx context.Context, organizationID uuid.UUID, journalType string) ([]domain.Journal, error)
}

type JournalFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	Type           *string // sale, purchase, cash, bank, general
	Active         *bool
	Search         *string // Search in name or code
	Limit          int
	Offset         int
}

type journalRepository struct {
	db *sql.DB
}

func NewJournalRepository(db *sql.DB) JournalRepository {
	return &journalRepository{db: db}
}

func (r *journalRepository) Create(ctx context.Context, journal domain.Journal) (*domain.Journal, error) {
	query := `
		INSERT INTO account_journals
		(id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at
	`

	now := time.Now()
	if journal.ID == uuid.Nil {
		journal.ID = uuid.New()
	}
	if journal.CreatedAt.IsZero() {
		journal.CreatedAt = now
	}
	if journal.UpdatedAt.IsZero() {
		journal.UpdatedAt = now
	}

	var createdJournal domain.Journal
	err := r.db.QueryRowContext(ctx, query,
		journal.ID, journal.OrganizationID, journal.CompanyID, journal.Name,
		journal.Code, journal.Type, journal.DefaultAccountID, journal.RefundSequence,
		journal.SequenceID, journal.CurrencyID, journal.BankAccountID, journal.Color,
		journal.Active, journal.CreatedAt, journal.UpdatedAt,
	).Scan(
		&createdJournal.ID, &createdJournal.OrganizationID, &createdJournal.CompanyID,
		&createdJournal.Name, &createdJournal.Code, &createdJournal.Type,
		&createdJournal.DefaultAccountID, &createdJournal.RefundSequence,
		&createdJournal.SequenceID, &createdJournal.CurrencyID,
		&createdJournal.BankAccountID, &createdJournal.Color, &createdJournal.Active,
		&createdJournal.CreatedAt, &createdJournal.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal: %w", err)
	}

	return &createdJournal, nil
}

func (r *journalRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Journal, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at
		FROM account_journals
		WHERE id = $1
	`

	var journal domain.Journal
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&journal.ID, &journal.OrganizationID, &journal.CompanyID,
		&journal.Name, &journal.Code, &journal.Type,
		&journal.DefaultAccountID, &journal.RefundSequence,
		&journal.SequenceID, &journal.CurrencyID,
		&journal.BankAccountID, &journal.Color, &journal.Active,
		&journal.CreatedAt, &journal.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find journal: %w", err)
	}

	return &journal, nil
}

func (r *journalRepository) FindAll(ctx context.Context, filters JournalFilter) ([]domain.Journal, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at
		FROM account_journals
		WHERE organization_id = $1
	`
	args := []interface{}{filters.OrganizationID}
	argCount := 1

	if filters.CompanyID != nil {
		argCount++
		query += fmt.Sprintf(" AND company_id = $%d", argCount)
		args = append(args, *filters.CompanyID)
	}

	if filters.Type != nil {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filters.Type)
	}

	if filters.Active != nil {
		argCount++
		query += fmt.Sprintf(" AND active = $%d", argCount)
		args = append(args, *filters.Active)
	}

	if filters.Search != nil && *filters.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern)
	}

	query += " ORDER BY code ASC"

	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find journals: %w", err)
	}
	defer rows.Close()

	var journals []domain.Journal
	for rows.Next() {
		var journal domain.Journal
		err := rows.Scan(
			&journal.ID, &journal.OrganizationID, &journal.CompanyID,
			&journal.Name, &journal.Code, &journal.Type,
			&journal.DefaultAccountID, &journal.RefundSequence,
			&journal.SequenceID, &journal.CurrencyID,
			&journal.BankAccountID, &journal.Color, &journal.Active,
			&journal.CreatedAt, &journal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journal: %w", err)
		}
		journals = append(journals, journal)
	}

	return journals, nil
}

func (r *journalRepository) Update(ctx context.Context, journal domain.Journal) (*domain.Journal, error) {
	query := `
		UPDATE account_journals
		SET name = $2, code = $3, type = $4, default_account_id = $5,
		    refund_sequence = $6, sequence_id = $7, currency_id = $8,
		    bank_account_id = $9, color = $10, active = $11, updated_at = $12
		WHERE id = $1
		RETURNING id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at
	`

	journal.UpdatedAt = time.Now()

	var updatedJournal domain.Journal
	err := r.db.QueryRowContext(ctx, query,
		journal.ID, journal.Name, journal.Code, journal.Type,
		journal.DefaultAccountID, journal.RefundSequence, journal.SequenceID,
		journal.CurrencyID, journal.BankAccountID, journal.Color,
		journal.Active, journal.UpdatedAt,
	).Scan(
		&updatedJournal.ID, &updatedJournal.OrganizationID, &updatedJournal.CompanyID,
		&updatedJournal.Name, &updatedJournal.Code, &updatedJournal.Type,
		&updatedJournal.DefaultAccountID, &updatedJournal.RefundSequence,
		&updatedJournal.SequenceID, &updatedJournal.CurrencyID,
		&updatedJournal.BankAccountID, &updatedJournal.Color, &updatedJournal.Active,
		&updatedJournal.CreatedAt, &updatedJournal.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("journal not found")
		}
		return nil, fmt.Errorf("failed to update journal: %w", err)
	}

	return &updatedJournal, nil
}

func (r *journalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Hard delete for journals (could be changed to soft delete if needed)
	query := `DELETE FROM account_journals WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete journal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("journal not found")
	}

	return nil
}

func (r *journalRepository) FindByCode(ctx context.Context, organizationID uuid.UUID, code string) (*domain.Journal, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at
		FROM account_journals
		WHERE organization_id = $1 AND code = $2
	`

	var journal domain.Journal
	err := r.db.QueryRowContext(ctx, query, organizationID, code).Scan(
		&journal.ID, &journal.OrganizationID, &journal.CompanyID,
		&journal.Name, &journal.Code, &journal.Type,
		&journal.DefaultAccountID, &journal.RefundSequence,
		&journal.SequenceID, &journal.CurrencyID,
		&journal.BankAccountID, &journal.Color, &journal.Active,
		&journal.CreatedAt, &journal.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find journal by code: %w", err)
	}

	return &journal, nil
}

func (r *journalRepository) FindByType(ctx context.Context, organizationID uuid.UUID, journalType string) ([]domain.Journal, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, type, default_account_id,
		 refund_sequence, sequence_id, currency_id, bank_account_id, color,
		 active, created_at, updated_at
		FROM account_journals
		WHERE organization_id = $1 AND type = $2
		ORDER BY code ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, journalType)
	if err != nil {
		return nil, fmt.Errorf("failed to find journals by type: %w", err)
	}
	defer rows.Close()

	var journals []domain.Journal
	for rows.Next() {
		var journal domain.Journal
		err := rows.Scan(
			&journal.ID, &journal.OrganizationID, &journal.CompanyID,
			&journal.Name, &journal.Code, &journal.Type,
			&journal.DefaultAccountID, &journal.RefundSequence,
			&journal.SequenceID, &journal.CurrencyID,
			&journal.BankAccountID, &journal.Color, &journal.Active,
			&journal.CreatedAt, &journal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journal: %w", err)
		}
		journals = append(journals, journal)
	}

	return journals, nil
}
