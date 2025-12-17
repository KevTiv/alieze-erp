package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type qualityControlChecklistRepository struct {
	db *sql.DB
}

func NewQualityControlChecklistRepository(db *sql.DB) QualityControlChecklistRepository {
	return &qualityControlChecklistRepository{db: db}
}

func (r *qualityControlChecklistRepository) Create(ctx context.Context, checklist types.QualityControlChecklist) (*types.QualityControlChecklist, error) {
	query := `
		INSERT INTO quality_control_checklists
		(id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
	`

	if checklist.ID == uuid.Nil {
		checklist.ID = uuid.New()
	}
	if checklist.CreatedAt.IsZero() {
		checklist.CreatedAt = time.Now()
	}
	if checklist.UpdatedAt.IsZero() {
		checklist.UpdatedAt = time.Now()
	}
	if checklist.Priority == 0 {
		checklist.Priority = 10
	}
	if checklist.Active == false {
		checklist.Active = true
	}

	var created types.QualityControlChecklist
	err := r.db.QueryRowContext(ctx, query,
		checklist.ID, checklist.OrganizationID, checklist.CompanyID, checklist.Name, checklist.Description,
		checklist.ProductID, checklist.ProductCategoryID, checklist.InspectionType, checklist.Active,
		checklist.Priority, checklist.CreatedAt, checklist.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Name, &created.Description,
		&created.ProductID, &created.ProductCategoryID, &created.InspectionType, &created.Active,
		&created.Priority, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality control checklist: %w", err)
	}

	return &created, nil
}

func (r *qualityControlChecklistRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.QualityControlChecklist, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
		FROM quality_control_checklists WHERE id = $1 AND deleted_at IS NULL
	`

	var checklist types.QualityControlChecklist
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&checklist.ID, &checklist.OrganizationID, &checklist.CompanyID, &checklist.Name, &checklist.Description,
		&checklist.ProductID, &checklist.ProductCategoryID, &checklist.InspectionType, &checklist.Active,
		&checklist.Priority, &checklist.CreatedAt, &checklist.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control checklist: %w", err)
	}

	return &checklist, nil
}

func (r *qualityControlChecklistRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlChecklist, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
		FROM quality_control_checklists WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control checklists: %w", err)
	}
	defer rows.Close()

	var checklists []types.QualityControlChecklist
	for rows.Next() {
		var checklist types.QualityControlChecklist
		err := rows.Scan(
			&checklist.ID, &checklist.OrganizationID, &checklist.CompanyID, &checklist.Name, &checklist.Description,
			&checklist.ProductID, &checklist.ProductCategoryID, &checklist.InspectionType, &checklist.Active,
			&checklist.Priority, &checklist.CreatedAt, &checklist.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control checklist: %w", err)
		}

		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

func (r *qualityControlChecklistRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.QualityControlChecklist, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
		FROM quality_control_checklists WHERE organization_id = $1 AND product_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control checklists by product: %w", err)
	}
	defer rows.Close()

	var checklists []types.QualityControlChecklist
	for rows.Next() {
		var checklist types.QualityControlChecklist
		err := rows.Scan(
			&checklist.ID, &checklist.OrganizationID, &checklist.CompanyID, &checklist.Name, &checklist.Description,
			&checklist.ProductID, &checklist.ProductCategoryID, &checklist.InspectionType, &checklist.Active,
			&checklist.Priority, &checklist.CreatedAt, &checklist.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control checklist: %w", err)
		}

		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

func (r *qualityControlChecklistRepository) FindByCategory(ctx context.Context, organizationID, categoryID uuid.UUID) ([]types.QualityControlChecklist, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
		FROM quality_control_checklists WHERE organization_id = $1 AND product_category_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control checklists by category: %w", err)
	}
	defer rows.Close()

	var checklists []types.QualityControlChecklist
	for rows.Next() {
		var checklist types.QualityControlChecklist
		err := rows.Scan(
			&checklist.ID, &checklist.OrganizationID, &checklist.CompanyID, &checklist.Name, &checklist.Description,
			&checklist.ProductID, &checklist.ProductCategoryID, &checklist.InspectionType, &checklist.Active,
			&checklist.Priority, &checklist.CreatedAt, &checklist.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control checklist: %w", err)
		}

		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

func (r *qualityControlChecklistRepository) FindByInspectionType(ctx context.Context, organizationID uuid.UUID, inspectionType string) ([]types.QualityControlChecklist, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
		FROM quality_control_checklists WHERE organization_id = $1 AND inspection_type = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, inspectionType)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control checklists by inspection type: %w", err)
	}
	defer rows.Close()

	var checklists []types.QualityControlChecklist
	for rows.Next() {
		var checklist types.QualityControlChecklist
		err := rows.Scan(
			&checklist.ID, &checklist.OrganizationID, &checklist.CompanyID, &checklist.Name, &checklist.Description,
			&checklist.ProductID, &checklist.ProductCategoryID, &checklist.InspectionType, &checklist.Active,
			&checklist.Priority, &checklist.CreatedAt, &checklist.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control checklist: %w", err)
		}

		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

func (r *qualityControlChecklistRepository) FindActive(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlChecklist, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
		FROM quality_control_checklists WHERE organization_id = $1 AND active = true AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active quality control checklists: %w", err)
	}
	defer rows.Close()

	var checklists []types.QualityControlChecklist
	for rows.Next() {
		var checklist types.QualityControlChecklist
		err := rows.Scan(
			&checklist.ID, &checklist.OrganizationID, &checklist.CompanyID, &checklist.Name, &checklist.Description,
			&checklist.ProductID, &checklist.ProductCategoryID, &checklist.InspectionType, &checklist.Active,
			&checklist.Priority, &checklist.CreatedAt, &checklist.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control checklist: %w", err)
		}

		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

func (r *qualityControlChecklistRepository) Update(ctx context.Context, checklist types.QualityControlChecklist) (*types.QualityControlChecklist, error) {
	query := `
		UPDATE quality_control_checklists
		SET name = $2, description = $3, product_id = $4, product_category_id = $5,
		 inspection_type = $6, active = $7, priority = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, name, description, product_id, product_category_id,
		 inspection_type, active, priority, created_at, updated_at
	`

	checklist.UpdatedAt = time.Now()
	var updated types.QualityControlChecklist
	err := r.db.QueryRowContext(ctx, query,
		checklist.ID, checklist.Name, checklist.Description, checklist.ProductID, checklist.ProductCategoryID,
		checklist.InspectionType, checklist.Active, checklist.Priority, checklist.UpdatedAt,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.Name, &updated.Description,
		&updated.ProductID, &updated.ProductCategoryID, &updated.InspectionType, &updated.Active,
		&updated.Priority, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quality control checklist not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update quality control checklist: %w", err)
	}

	return &updated, nil
}

func (r *qualityControlChecklistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE quality_control_checklists SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete quality control checklist: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality control checklist not found")
	}
	return nil
}
