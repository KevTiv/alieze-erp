package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
)

type TaxRepository interface {
	Create(ctx context.Context, tax domain.Tax) (*domain.Tax, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tax, error)
	FindAll(ctx context.Context, filters TaxFilter) ([]domain.Tax, error)
	Update(ctx context.Context, tax domain.Tax) (*domain.Tax, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByType(ctx context.Context, organizationID uuid.UUID, typeTaxUse string) ([]domain.Tax, error)
}

type TaxFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	TypeTaxUse     *string // sale, purchase, none
	AmountType     *string // percent, fixed, division, group
	Active         *bool
	Search         *string // Search in name or description
	Limit          int
	Offset         int
}

type taxRepository struct {
	db *sql.DB
}

func NewTaxRepository(db *sql.DB) TaxRepository {
	return &taxRepository{db: db}
}

func (r *taxRepository) Create(ctx context.Context, tax domain.Tax) (*domain.Tax, error) {
	query := `
		INSERT INTO account_taxes
		(id, organization_id, company_id, name, type_tax_use, amount_type,
		 amount, price_include, include_base_amount, is_base_affected,
		 description, sequence, active, tax_group_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, organization_id, company_id, name, type_tax_use, amount_type,
		 amount, price_include, include_base_amount, is_base_affected,
		 description, sequence, active, tax_group_id, created_at, updated_at
	`

	now := time.Now()
	if tax.ID == uuid.Nil {
		tax.ID = uuid.New()
	}
	if tax.CreatedAt.IsZero() {
		tax.CreatedAt = now
	}
	if tax.UpdatedAt.IsZero() {
		tax.UpdatedAt = now
	}

	var createdTax domain.Tax
	err := r.db.QueryRowContext(ctx, query,
		tax.ID, tax.OrganizationID, tax.CompanyID, tax.Name,
		tax.TypeTaxUse, tax.AmountType, tax.Amount, tax.PriceInclude,
		tax.IncludeBaseAmount, tax.IsBaseAffected, tax.Description,
		tax.Sequence, tax.Active, tax.TaxGroupID, tax.CreatedAt, tax.UpdatedAt,
	).Scan(
		&createdTax.ID, &createdTax.OrganizationID, &createdTax.CompanyID,
		&createdTax.Name, &createdTax.TypeTaxUse, &createdTax.AmountType,
		&createdTax.Amount, &createdTax.PriceInclude, &createdTax.IncludeBaseAmount,
		&createdTax.IsBaseAffected, &createdTax.Description, &createdTax.Sequence,
		&createdTax.Active, &createdTax.TaxGroupID, &createdTax.CreatedAt, &createdTax.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tax: %w", err)
	}

	return &createdTax, nil
}

func (r *taxRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tax, error) {
	query := `
		SELECT id, organization_id, company_id, name, type_tax_use, amount_type,
		 amount, price_include, include_base_amount, is_base_affected,
		 description, sequence, active, tax_group_id, created_at, updated_at
		FROM account_taxes
		WHERE id = $1
	`

	var tax domain.Tax
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tax.ID, &tax.OrganizationID, &tax.CompanyID,
		&tax.Name, &tax.TypeTaxUse, &tax.AmountType,
		&tax.Amount, &tax.PriceInclude, &tax.IncludeBaseAmount,
		&tax.IsBaseAffected, &tax.Description, &tax.Sequence,
		&tax.Active, &tax.TaxGroupID, &tax.CreatedAt, &tax.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find tax: %w", err)
	}

	return &tax, nil
}

func (r *taxRepository) FindAll(ctx context.Context, filters TaxFilter) ([]domain.Tax, error) {
	query := `
		SELECT id, organization_id, company_id, name, type_tax_use, amount_type,
		 amount, price_include, include_base_amount, is_base_affected,
		 description, sequence, active, tax_group_id, created_at, updated_at
		FROM account_taxes
		WHERE organization_id = $1
	`
	args := []interface{}{filters.OrganizationID}
	argCount := 1

	if filters.CompanyID != nil {
		argCount++
		query += fmt.Sprintf(" AND company_id = $%d", argCount)
		args = append(args, *filters.CompanyID)
	}

	if filters.TypeTaxUse != nil {
		argCount++
		query += fmt.Sprintf(" AND type_tax_use = $%d", argCount)
		args = append(args, *filters.TypeTaxUse)
	}

	if filters.AmountType != nil {
		argCount++
		query += fmt.Sprintf(" AND amount_type = $%d", argCount)
		args = append(args, *filters.AmountType)
	}

	if filters.Active != nil {
		argCount++
		query += fmt.Sprintf(" AND active = $%d", argCount)
		args = append(args, *filters.Active)
	}

	if filters.Search != nil && *filters.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern)
	}

	query += " ORDER BY sequence ASC, name ASC"

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
		return nil, fmt.Errorf("failed to find taxes: %w", err)
	}
	defer rows.Close()

	var taxes []domain.Tax
	for rows.Next() {
		var tax domain.Tax
		err := rows.Scan(
			&tax.ID, &tax.OrganizationID, &tax.CompanyID,
			&tax.Name, &tax.TypeTaxUse, &tax.AmountType,
			&tax.Amount, &tax.PriceInclude, &tax.IncludeBaseAmount,
			&tax.IsBaseAffected, &tax.Description, &tax.Sequence,
			&tax.Active, &tax.TaxGroupID, &tax.CreatedAt, &tax.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tax: %w", err)
		}
		taxes = append(taxes, tax)
	}

	return taxes, nil
}

func (r *taxRepository) Update(ctx context.Context, tax domain.Tax) (*domain.Tax, error) {
	query := `
		UPDATE account_taxes
		SET name = $2, type_tax_use = $3, amount_type = $4, amount = $5,
		    price_include = $6, include_base_amount = $7, is_base_affected = $8,
		    description = $9, sequence = $10, active = $11, tax_group_id = $12,
		    updated_at = $13
		WHERE id = $1
		RETURNING id, organization_id, company_id, name, type_tax_use, amount_type,
		 amount, price_include, include_base_amount, is_base_affected,
		 description, sequence, active, tax_group_id, created_at, updated_at
	`

	tax.UpdatedAt = time.Now()

	var updatedTax domain.Tax
	err := r.db.QueryRowContext(ctx, query,
		tax.ID, tax.Name, tax.TypeTaxUse, tax.AmountType, tax.Amount,
		tax.PriceInclude, tax.IncludeBaseAmount, tax.IsBaseAffected,
		tax.Description, tax.Sequence, tax.Active, tax.TaxGroupID, tax.UpdatedAt,
	).Scan(
		&updatedTax.ID, &updatedTax.OrganizationID, &updatedTax.CompanyID,
		&updatedTax.Name, &updatedTax.TypeTaxUse, &updatedTax.AmountType,
		&updatedTax.Amount, &updatedTax.PriceInclude, &updatedTax.IncludeBaseAmount,
		&updatedTax.IsBaseAffected, &updatedTax.Description, &updatedTax.Sequence,
		&updatedTax.Active, &updatedTax.TaxGroupID, &updatedTax.CreatedAt, &updatedTax.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tax not found")
		}
		return nil, fmt.Errorf("failed to update tax: %w", err)
	}

	return &updatedTax, nil
}

func (r *taxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete by setting active = false
	query := `
		UPDATE account_taxes
		SET active = false, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete tax: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tax not found")
	}

	return nil
}

func (r *taxRepository) FindByType(ctx context.Context, organizationID uuid.UUID, typeTaxUse string) ([]domain.Tax, error) {
	query := `
		SELECT id, organization_id, company_id, name, type_tax_use, amount_type,
		 amount, price_include, include_base_amount, is_base_affected,
		 description, sequence, active, tax_group_id, created_at, updated_at
		FROM account_taxes
		WHERE organization_id = $1 AND type_tax_use = $2 AND active = true
		ORDER BY sequence ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, typeTaxUse)
	if err != nil {
		return nil, fmt.Errorf("failed to find taxes by type: %w", err)
	}
	defer rows.Close()

	var taxes []domain.Tax
	for rows.Next() {
		var tax domain.Tax
		err := rows.Scan(
			&tax.ID, &tax.OrganizationID, &tax.CompanyID,
			&tax.Name, &tax.TypeTaxUse, &tax.AmountType,
			&tax.Amount, &tax.PriceInclude, &tax.IncludeBaseAmount,
			&tax.IsBaseAffected, &tax.Description, &tax.Sequence,
			&tax.Active, &tax.TaxGroupID, &tax.CreatedAt, &tax.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tax: %w", err)
		}
		taxes = append(taxes, tax)
	}

	return taxes, nil
}
