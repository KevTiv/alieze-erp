package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"alieze-erp/internal/modules/products/domain"

	"github.com/google/uuid"
)

// ProductRepository handles product data operations
type ProductRepository struct {
	db *sql.DB
}

// Ensure ProductRepository implements ProductRepo interface
var _ ProductRepo = &ProductRepository{}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product domain.Product) (*domain.Product, error) {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	if product.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if product.Name == "" {
		return nil, errors.New("name is required")
	}

	if product.ProductType == "" {
		product.ProductType = "storable"
	}

	if product.Active == false {
		product.Active = true
	}

	query := `
		INSERT INTO products (
			id, organization_id, name, default_code, barcode, product_type,
			category_id, list_price, active, created_at, updated_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, organization_id, name, default_code, barcode, product_type,
			category_id, list_price, active, created_at, updated_at, deleted_at
	`

	now := time.Now()

	result := r.db.QueryRowContext(ctx, query,
		product.ID,
		product.OrganizationID,
		product.Name,
		product.DefaultCode,
		product.Barcode,
		product.ProductType,
		product.CategoryID,
		product.ListPrice,
		product.Active,
		now,
		now,
		nil,
	)

	var created domain.Product
	err := result.Scan(
		&created.ID,
		&created.OrganizationID,
		&created.Name,
		&created.DefaultCode,
		&created.Barcode,
		&created.ProductType,
		&created.CategoryID,
		&created.ListPrice,
		&created.Active,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &created, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product id")
	}

	query := `
		SELECT id, organization_id, name, default_code, barcode, product_type,
			category_id, list_price, active, created_at, updated_at, deleted_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.OrganizationID,
		&product.Name,
		&product.DefaultCode,
		&product.Barcode,
		&product.ProductType,
		&product.CategoryID,
		&product.ListPrice,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) FindAll(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	query := `SELECT id, organization_id, name, default_code, barcode, product_type,
		category_id, list_price, active, created_at, updated_at, deleted_at
		FROM products WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.DefaultCode != nil && *filter.DefaultCode != "" {
		conditions = append(conditions, fmt.Sprintf("default_code ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.DefaultCode+"%")
		argIndex++
	}

	if filter.Barcode != nil && *filter.Barcode != "" {
		conditions = append(conditions, fmt.Sprintf("barcode ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Barcode+"%")
		argIndex++
	}

	if filter.ProductType != nil && *filter.ProductType != "" {
		conditions = append(conditions, fmt.Sprintf("product_type = $%d", argIndex))
		args = append(args, *filter.ProductType)
		argIndex++
	}

	if filter.CategoryID != nil && *filter.CategoryID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filter.Active)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY name ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.OrganizationID,
			&product.Name,
			&product.DefaultCode,
			&product.Barcode,
			&product.ProductType,
			&product.CategoryID,
			&product.ListPrice,
			&product.Active,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during product iteration: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product domain.Product) (*domain.Product, error) {
	if product.ID == uuid.Nil {
		return nil, errors.New("product id is required")
	}

	if product.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if product.Name == "" {
		return nil, errors.New("name is required")
	}

	product.UpdatedAt = time.Now()

	query := `
		UPDATE products SET
			organization_id = $1,
			name = $2,
			default_code = $3,
			barcode = $4,
			product_type = $5,
			category_id = $6,
			list_price = $7,
			active = $8,
			updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
		RETURNING id, organization_id, name, default_code, barcode, product_type,
			category_id, list_price, active, created_at, updated_at, deleted_at
	`

	result := r.db.QueryRowContext(ctx, query,
		product.OrganizationID,
		product.Name,
		product.DefaultCode,
		product.Barcode,
		product.ProductType,
		product.CategoryID,
		product.ListPrice,
		product.Active,
		product.UpdatedAt,
		product.ID,
	)

	var updated domain.Product
	err := result.Scan(
		&updated.ID,
		&updated.OrganizationID,
		&updated.Name,
		&updated.DefaultCode,
		&updated.Barcode,
		&updated.ProductType,
		&updated.CategoryID,
		&updated.ListPrice,
		&updated.Active,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &updated, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid product id")
	}

	query := `
		UPDATE products SET
			deleted_at = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found or already deleted")
	}

	return nil
}

func (r *ProductRepository) Count(ctx context.Context, filter domain.ProductFilter) (int, error) {
	query := `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.DefaultCode != nil && *filter.DefaultCode != "" {
		conditions = append(conditions, fmt.Sprintf("default_code ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.DefaultCode+"%")
		argIndex++
	}

	if filter.Barcode != nil && *filter.Barcode != "" {
		conditions = append(conditions, fmt.Sprintf("barcode ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Barcode+"%")
		argIndex++
	}

	if filter.ProductType != nil && *filter.ProductType != "" {
		conditions = append(conditions, fmt.Sprintf("product_type = $%d", argIndex))
		args = append(args, *filter.ProductType)
		argIndex++
	}

	if filter.CategoryID != nil && *filter.CategoryID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filter.Active)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}
