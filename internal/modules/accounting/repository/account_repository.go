package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (*domain.Account, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	FindAll(ctx context.Context, filters AccountFilter) ([]domain.Account, error)
	Update(ctx context.Context, account domain.Account) (*domain.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByCode(ctx context.Context, organizationID uuid.UUID, code string) (*domain.Account, error)
	FindByType(ctx context.Context, organizationID uuid.UUID, accountType string) ([]domain.Account, error)
}

type AccountFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	AccountType    *string
	Deprecated     *bool
	Reconcile      *bool
	Search         *string // Search in name or code
	Limit          int
	Offset         int
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account domain.Account) (*domain.Account, error) {
	query := `
		INSERT INTO account_accounts
		(id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by, deleted_at
	`

	now := time.Now()
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	if account.UpdatedAt.IsZero() {
		account.UpdatedAt = now
	}

	var createdAccount domain.Account
	err := r.db.QueryRowContext(ctx, query,
		account.ID, account.OrganizationID, account.CompanyID, account.Name,
		account.Code, account.Deprecated, account.AccountType, account.InternalType,
		account.InternalGroup, account.UserTypeID, account.Reconcile, account.CurrencyID,
		account.GroupID, pq.Array(account.TaxIDs), account.Note, pq.Array(account.TagIDs),
		account.CreatedAt, account.UpdatedAt, account.CreatedBy, account.UpdatedBy,
	).Scan(
		&createdAccount.ID, &createdAccount.OrganizationID, &createdAccount.CompanyID,
		&createdAccount.Name, &createdAccount.Code, &createdAccount.Deprecated,
		&createdAccount.AccountType, &createdAccount.InternalType, &createdAccount.InternalGroup,
		&createdAccount.UserTypeID, &createdAccount.Reconcile, &createdAccount.CurrencyID,
		&createdAccount.GroupID, pq.Array(&createdAccount.TaxIDs), &createdAccount.Note,
		pq.Array(&createdAccount.TagIDs), &createdAccount.CreatedAt, &createdAccount.UpdatedAt,
		&createdAccount.CreatedBy, &createdAccount.UpdatedBy, &createdAccount.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &createdAccount, nil
}

func (r *accountRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by, deleted_at
		FROM account_accounts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var account domain.Account
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.OrganizationID, &account.CompanyID,
		&account.Name, &account.Code, &account.Deprecated,
		&account.AccountType, &account.InternalType, &account.InternalGroup,
		&account.UserTypeID, &account.Reconcile, &account.CurrencyID,
		&account.GroupID, pq.Array(&account.TaxIDs), &account.Note,
		pq.Array(&account.TagIDs), &account.CreatedAt, &account.UpdatedAt,
		&account.CreatedBy, &account.UpdatedBy, &account.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	return &account, nil
}

func (r *accountRepository) FindAll(ctx context.Context, filters AccountFilter) ([]domain.Account, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by, deleted_at
		FROM account_accounts
		WHERE organization_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{filters.OrganizationID}
	argCount := 1

	if filters.CompanyID != nil {
		argCount++
		query += fmt.Sprintf(" AND company_id = $%d", argCount)
		args = append(args, *filters.CompanyID)
	}

	if filters.AccountType != nil {
		argCount++
		query += fmt.Sprintf(" AND account_type = $%d", argCount)
		args = append(args, *filters.AccountType)
	}

	if filters.Deprecated != nil {
		argCount++
		query += fmt.Sprintf(" AND deprecated = $%d", argCount)
		args = append(args, *filters.Deprecated)
	}

	if filters.Reconcile != nil {
		argCount++
		query += fmt.Sprintf(" AND reconcile = $%d", argCount)
		args = append(args, *filters.Reconcile)
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
		return nil, fmt.Errorf("failed to find accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var account domain.Account
		err := rows.Scan(
			&account.ID, &account.OrganizationID, &account.CompanyID,
			&account.Name, &account.Code, &account.Deprecated,
			&account.AccountType, &account.InternalType, &account.InternalGroup,
			&account.UserTypeID, &account.Reconcile, &account.CurrencyID,
			&account.GroupID, pq.Array(&account.TaxIDs), &account.Note,
			pq.Array(&account.TagIDs), &account.CreatedAt, &account.UpdatedAt,
			&account.CreatedBy, &account.UpdatedBy, &account.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *accountRepository) Update(ctx context.Context, account domain.Account) (*domain.Account, error) {
	query := `
		UPDATE account_accounts
		SET name = $2, code = $3, deprecated = $4, account_type = $5,
		    internal_type = $6, internal_group = $7, user_type_id = $8,
		    reconcile = $9, currency_id = $10, group_id = $11,
		    tax_ids = $12, note = $13, tag_ids = $14,
		    updated_at = $15, updated_by = $16
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by, deleted_at
	`

	account.UpdatedAt = time.Now()

	var updatedAccount domain.Account
	err := r.db.QueryRowContext(ctx, query,
		account.ID, account.Name, account.Code, account.Deprecated, account.AccountType,
		account.InternalType, account.InternalGroup, account.UserTypeID, account.Reconcile,
		account.CurrencyID, account.GroupID, pq.Array(account.TaxIDs), account.Note,
		pq.Array(account.TagIDs), account.UpdatedAt, account.UpdatedBy,
	).Scan(
		&updatedAccount.ID, &updatedAccount.OrganizationID, &updatedAccount.CompanyID,
		&updatedAccount.Name, &updatedAccount.Code, &updatedAccount.Deprecated,
		&updatedAccount.AccountType, &updatedAccount.InternalType, &updatedAccount.InternalGroup,
		&updatedAccount.UserTypeID, &updatedAccount.Reconcile, &updatedAccount.CurrencyID,
		&updatedAccount.GroupID, pq.Array(&updatedAccount.TaxIDs), &updatedAccount.Note,
		pq.Array(&updatedAccount.TagIDs), &updatedAccount.CreatedAt, &updatedAccount.UpdatedAt,
		&updatedAccount.CreatedBy, &updatedAccount.UpdatedBy, &updatedAccount.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return &updatedAccount, nil
}

func (r *accountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	query := `
		UPDATE account_accounts
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

func (r *accountRepository) FindByCode(ctx context.Context, organizationID uuid.UUID, code string) (*domain.Account, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by, deleted_at
		FROM account_accounts
		WHERE organization_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	var account domain.Account
	err := r.db.QueryRowContext(ctx, query, organizationID, code).Scan(
		&account.ID, &account.OrganizationID, &account.CompanyID,
		&account.Name, &account.Code, &account.Deprecated,
		&account.AccountType, &account.InternalType, &account.InternalGroup,
		&account.UserTypeID, &account.Reconcile, &account.CurrencyID,
		&account.GroupID, pq.Array(&account.TaxIDs), &account.Note,
		pq.Array(&account.TagIDs), &account.CreatedAt, &account.UpdatedAt,
		&account.CreatedBy, &account.UpdatedBy, &account.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find account by code: %w", err)
	}

	return &account, nil
}

func (r *accountRepository) FindByType(ctx context.Context, organizationID uuid.UUID, accountType string) ([]domain.Account, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, deprecated, account_type,
		 internal_type, internal_group, user_type_id, reconcile, currency_id,
		 group_id, tax_ids, note, tag_ids, created_at, updated_at, created_by, updated_by, deleted_at
		FROM account_accounts
		WHERE organization_id = $1 AND account_type = $2 AND deleted_at IS NULL
		ORDER BY code ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, accountType)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts by type: %w", err)
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var account domain.Account
		err := rows.Scan(
			&account.ID, &account.OrganizationID, &account.CompanyID,
			&account.Name, &account.Code, &account.Deprecated,
			&account.AccountType, &account.InternalType, &account.InternalGroup,
			&account.UserTypeID, &account.Reconcile, &account.CurrencyID,
			&account.GroupID, pq.Array(&account.TaxIDs), &account.Note,
			pq.Array(&account.TagIDs), &account.CreatedAt, &account.UpdatedAt,
			&account.CreatedBy, &account.UpdatedBy, &account.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
