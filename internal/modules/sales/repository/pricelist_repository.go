package repository

import (
	"context"
	"database/sql"
	"fmt"

	"alieze-erp/internal/modules/sales/types"

	"github.com/google/uuid"
)

type PricelistRepository interface {
	Create(ctx context.Context, pricelist domain.Pricelist) (*domain.Pricelist, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Pricelist, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Pricelist, error)
	Update(ctx context.Context, pricelist domain.Pricelist) (*domain.Pricelist, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Pricelist, error)
	FindActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Pricelist, error)
}

type pricelistRepository struct {
	db *sql.DB
}

func NewPricelistRepository(db *sql.DB) PricelistRepository {
	return &pricelistRepository{db: db}
}

func (r *pricelistRepository) Create(ctx context.Context, pricelist domain.Pricelist) (*domain.Pricelist, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create pricelist
	query := `
		INSERT INTO pricelists
		(id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by
	`

	var createdPricelist domain.Pricelist
	err = tx.QueryRowContext(ctx, query,
		pricelist.ID, pricelist.OrganizationID, pricelist.CompanyID, pricelist.Name,
		pricelist.CurrencyID, pricelist.IsActive, pricelist.CreatedAt, pricelist.UpdatedAt,
		pricelist.CreatedBy, pricelist.UpdatedBy,
	).Scan(
		&createdPricelist.ID, &createdPricelist.OrganizationID, &createdPricelist.CompanyID,
		&createdPricelist.Name, &createdPricelist.CurrencyID, &createdPricelist.IsActive,
		&createdPricelist.CreatedAt, &createdPricelist.UpdatedAt,
		&createdPricelist.CreatedBy, &createdPricelist.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricelist: %w", err)
	}

	// Create pricelist items
	for _, item := range pricelist.Items {
		itemQuery := `
			INSERT INTO pricelist_items
			(id, pricelist_id, product_id, min_quantity, fixed_price, discount,
			 created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, pricelist_id, product_id, min_quantity, fixed_price, discount,
			 created_at, updated_at
		`

		var createdItem domain.PricelistItem
		err = tx.QueryRowContext(ctx, itemQuery,
			item.ID, createdPricelist.ID, item.ProductID, item.MinQuantity,
			item.FixedPrice, item.Discount, item.CreatedAt, item.UpdatedAt,
		).Scan(
			&createdItem.ID, &createdItem.PricelistID, &createdItem.ProductID,
			&createdItem.MinQuantity, &createdItem.FixedPrice, &createdItem.Discount,
			&createdItem.CreatedAt, &createdItem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create pricelist item: %w", err)
		}
		createdPricelist.Items = append(createdPricelist.Items, createdItem)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &createdPricelist, nil
}

func (r *pricelistRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Pricelist, error) {
	query := `
		SELECT id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by
		FROM pricelists
		WHERE id = $1
	`

	var pricelist domain.Pricelist
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pricelist.ID, &pricelist.OrganizationID, &pricelist.CompanyID,
		&pricelist.Name, &pricelist.CurrencyID, &pricelist.IsActive,
		&pricelist.CreatedAt, &pricelist.UpdatedAt,
		&pricelist.CreatedBy, &pricelist.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find pricelist: %w", err)
	}

	// Load items
	items, err := r.findItemsByPricelistID(ctx, pricelist.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load pricelist items: %w", err)
	}
	pricelist.Items = items

	return &pricelist, nil
}

func (r *pricelistRepository) findItemsByPricelistID(ctx context.Context, pricelistID uuid.UUID) ([]domain.PricelistItem, error) {
	query := `
		SELECT id, pricelist_id, product_id, min_quantity, fixed_price, discount,
		 created_at, updated_at
		FROM pricelist_items
		WHERE pricelist_id = $1
		ORDER BY min_quantity
	`

	rows, err := r.db.QueryContext(ctx, query, pricelistID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pricelist items: %w", err)
	}
	defer rows.Close()

	var items []domain.PricelistItem
	for rows.Next() {
		var item domain.PricelistItem
		err = rows.Scan(
			&item.ID, &item.PricelistID, &item.ProductID,
			&item.MinQuantity, &item.FixedPrice, &item.Discount,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricelist item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *pricelistRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Pricelist, error) {
	query := `
		SELECT id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by
		FROM pricelists
		WHERE organization_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pricelists: %w", err)
	}
	defer rows.Close()

	var pricelists []domain.Pricelist
	for rows.Next() {
		var pricelist domain.Pricelist
		err = rows.Scan(
			&pricelist.ID, &pricelist.OrganizationID, &pricelist.CompanyID,
			&pricelist.Name, &pricelist.CurrencyID, &pricelist.IsActive,
			&pricelist.CreatedAt, &pricelist.UpdatedAt,
			&pricelist.CreatedBy, &pricelist.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricelist: %w", err)
		}
		pricelists = append(pricelists, pricelist)
	}

	// Load items for each pricelist
	for i := range pricelists {
		items, err := r.findItemsByPricelistID(ctx, pricelists[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load pricelist items: %w", err)
		}
		pricelists[i].Items = items
	}

	return pricelists, nil
}

func (r *pricelistRepository) Update(ctx context.Context, pricelist domain.Pricelist) (*domain.Pricelist, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update pricelist
	query := `
		UPDATE pricelists
		SET name = $1, currency_id = $2, is_active = $3,
		 updated_at = $4, updated_by = $5
		WHERE id = $6
		RETURNING id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by
	`

	var updatedPricelist domain.Pricelist
	err = tx.QueryRowContext(ctx, query,
		pricelist.Name, pricelist.CurrencyID, pricelist.IsActive,
		pricelist.UpdatedAt, pricelist.UpdatedBy, pricelist.ID,
	).Scan(
		&updatedPricelist.ID, &updatedPricelist.OrganizationID, &updatedPricelist.CompanyID,
		&updatedPricelist.Name, &updatedPricelist.CurrencyID, &updatedPricelist.IsActive,
		&updatedPricelist.CreatedAt, &updatedPricelist.UpdatedAt,
		&updatedPricelist.CreatedBy, &updatedPricelist.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update pricelist: %w", err)
	}

	// Delete existing items and create new ones
	_, err = tx.ExecContext(ctx, "DELETE FROM pricelist_items WHERE pricelist_id = $1", pricelist.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing pricelist items: %w", err)
	}

	// Create new items
	for _, item := range pricelist.Items {
		itemQuery := `
			INSERT INTO pricelist_items
			(id, pricelist_id, product_id, min_quantity, fixed_price, discount,
			 created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, pricelist_id, product_id, min_quantity, fixed_price, discount,
			 created_at, updated_at
		`

		var createdItem domain.PricelistItem
		err = tx.QueryRowContext(ctx, itemQuery,
			item.ID, updatedPricelist.ID, item.ProductID, item.MinQuantity,
			item.FixedPrice, item.Discount, item.CreatedAt, item.UpdatedAt,
		).Scan(
			&createdItem.ID, &createdItem.PricelistID, &createdItem.ProductID,
			&createdItem.MinQuantity, &createdItem.FixedPrice, &createdItem.Discount,
			&createdItem.CreatedAt, &createdItem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create pricelist item: %w", err)
		}
		updatedPricelist.Items = append(updatedPricelist.Items, createdItem)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedPricelist, nil
}

func (r *pricelistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete items first
	_, err = tx.ExecContext(ctx, "DELETE FROM pricelist_items WHERE pricelist_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete pricelist items: %w", err)
	}

	// Delete pricelist
	_, err = tx.ExecContext(ctx, "DELETE FROM pricelists WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete pricelist: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *pricelistRepository) FindByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Pricelist, error) {
	query := `
		SELECT id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by
		FROM pricelists
		WHERE company_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pricelists by company: %w", err)
	}
	defer rows.Close()

	var pricelists []domain.Pricelist
	for rows.Next() {
		var pricelist domain.Pricelist
		err = rows.Scan(
			&pricelist.ID, &pricelist.OrganizationID, &pricelist.CompanyID,
			&pricelist.Name, &pricelist.CurrencyID, &pricelist.IsActive,
			&pricelist.CreatedAt, &pricelist.UpdatedAt,
			&pricelist.CreatedBy, &pricelist.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricelist: %w", err)
		}
		pricelists = append(pricelists, pricelist)
	}

	// Load items for each pricelist
	for i := range pricelists {
		items, err := r.findItemsByPricelistID(ctx, pricelists[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load pricelist items: %w", err)
		}
		pricelists[i].Items = items
	}

	return pricelists, nil
}

func (r *pricelistRepository) FindActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Pricelist, error) {
	query := `
		SELECT id, organization_id, company_id, name, currency_id, is_active,
		 created_at, updated_at, created_by, updated_by
		FROM pricelists
		WHERE company_id = $1 AND is_active = TRUE
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active pricelists by company: %w", err)
	}
	defer rows.Close()

	var pricelists []domain.Pricelist
	for rows.Next() {
		var pricelist domain.Pricelist
		err = rows.Scan(
			&pricelist.ID, &pricelist.OrganizationID, &pricelist.CompanyID,
			&pricelist.Name, &pricelist.CurrencyID, &pricelist.IsActive,
			&pricelist.CreatedAt, &pricelist.UpdatedAt,
			&pricelist.CreatedBy, &pricelist.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricelist: %w", err)
		}
		pricelists = append(pricelists, pricelist)
	}

	// Load items for each pricelist
	for i := range pricelists {
		items, err := r.findItemsByPricelistID(ctx, pricelists[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load pricelist items: %w", err)
		}
		pricelists[i].Items = items
	}

	return pricelists, nil
}
