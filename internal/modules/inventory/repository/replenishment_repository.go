package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// ReplenishmentRuleRepository interface
type ReplenishmentRuleRepository interface {
	Create(ctx context.Context, rule types.ReplenishmentRule) (*types.ReplenishmentRule, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.ReplenishmentRule, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.ReplenishmentRule, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.ReplenishmentRule, error)
	FindByCategory(ctx context.Context, organizationID, categoryID uuid.UUID) ([]types.ReplenishmentRule, error)
	FindByWarehouse(ctx context.Context, organizationID, warehouseID uuid.UUID) ([]types.ReplenishmentRule, error)
	FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]types.ReplenishmentRule, error)
	Update(ctx context.Context, rule types.ReplenishmentRule) (*types.ReplenishmentRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CheckAndCreateReplenishmentOrders(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.ReplenishmentCheckResult, error)
	UpdateRuleCheckTimes(ctx context.Context, organizationID uuid.UUID) error
}

type replenishmentRuleRepository struct {
	db *sql.DB
}

func NewReplenishmentRuleRepository(db *sql.DB) ReplenishmentRuleRepository {
	return &replenishmentRuleRepository{db: db}
}

func (r *replenishmentRuleRepository) Create(ctx context.Context, rule types.ReplenishmentRule) (*types.ReplenishmentRule, error) {
	query := `
		INSERT INTO replenishment_rules
		(id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
		 $19, $20, $21, $22, $23, $24, $25, $26, $27)
		RETURNING id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
	`

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = now
	}

	var created types.ReplenishmentRule
	err := r.db.QueryRowContext(ctx, query,
		rule.ID, rule.OrganizationID, rule.CompanyID, rule.Name, rule.Description,
		rule.ProductID, rule.ProductCategoryID, rule.WarehouseID, rule.LocationID,
		rule.TriggerType, rule.MinQuantity, rule.MaxQuantity, rule.ReorderPoint,
		rule.SafetyStock, rule.ProcureMethod, rule.OrderQuantity, rule.MultipleOf,
		rule.LeadTimeDays, rule.CheckFrequency, rule.LastCheckedAt, rule.NextCheckAt,
		rule.SourceLocationID, rule.DestLocationID, rule.Active, rule.Priority,
		rule.CreatedAt, rule.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Name, &created.Description,
		&created.ProductID, &created.ProductCategoryID, &created.WarehouseID, &created.LocationID,
		&created.TriggerType, &created.MinQuantity, &created.MaxQuantity, &created.ReorderPoint,
		&created.SafetyStock, &created.ProcureMethod, &created.OrderQuantity, &created.MultipleOf,
		&created.LeadTimeDays, &created.CheckFrequency, &created.LastCheckedAt, &created.NextCheckAt,
		&created.SourceLocationID, &created.DestLocationID, &created.Active, &created.Priority,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create replenishment rule: %w", err)
	}
	return &created, nil
}

func (r *replenishmentRuleRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.ReplenishmentRule, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
		FROM replenishment_rules WHERE id = $1 AND deleted_at IS NULL
	`

	var rule types.ReplenishmentRule
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.OrganizationID, &rule.CompanyID, &rule.Name, &rule.Description,
		&rule.ProductID, &rule.ProductCategoryID, &rule.WarehouseID, &rule.LocationID,
		&rule.TriggerType, &rule.MinQuantity, &rule.MaxQuantity, &rule.ReorderPoint,
		&rule.SafetyStock, &rule.ProcureMethod, &rule.OrderQuantity, &rule.MultipleOf,
		&rule.LeadTimeDays, &rule.CheckFrequency, &rule.LastCheckedAt, &rule.NextCheckAt,
		&rule.SourceLocationID, &rule.DestLocationID, &rule.Active, &rule.Priority,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment rule: %w", err)
	}
	return &rule, nil
}

func (r *replenishmentRuleRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.ReplenishmentRule, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
		FROM replenishment_rules WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment rules: %w", err)
	}
	defer rows.Close()

	var rules []types.ReplenishmentRule
	for rows.Next() {
		var rule types.ReplenishmentRule
		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.CompanyID, &rule.Name, &rule.Description,
			&rule.ProductID, &rule.ProductCategoryID, &rule.WarehouseID, &rule.LocationID,
			&rule.TriggerType, &rule.MinQuantity, &rule.MaxQuantity, &rule.ReorderPoint,
			&rule.SafetyStock, &rule.ProcureMethod, &rule.OrderQuantity, &rule.MultipleOf,
			&rule.LeadTimeDays, &rule.CheckFrequency, &rule.LastCheckedAt, &rule.NextCheckAt,
			&rule.SourceLocationID, &rule.DestLocationID, &rule.Active, &rule.Priority,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *replenishmentRuleRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.ReplenishmentRule, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
		FROM replenishment_rules WHERE organization_id = $1 AND product_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment rules by product: %w", err)
	}
	defer rows.Close()

	var rules []types.ReplenishmentRule
	for rows.Next() {
		var rule types.ReplenishmentRule
		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.CompanyID, &rule.Name, &rule.Description,
			&rule.ProductID, &rule.ProductCategoryID, &rule.WarehouseID, &rule.LocationID,
			&rule.TriggerType, &rule.MinQuantity, &rule.MaxQuantity, &rule.ReorderPoint,
			&rule.SafetyStock, &rule.ProcureMethod, &rule.OrderQuantity, &rule.MultipleOf,
			&rule.LeadTimeDays, &rule.CheckFrequency, &rule.LastCheckedAt, &rule.NextCheckAt,
			&rule.SourceLocationID, &rule.DestLocationID, &rule.Active, &rule.Priority,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *replenishmentRuleRepository) FindByCategory(ctx context.Context, organizationID, categoryID uuid.UUID) ([]types.ReplenishmentRule, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
		FROM replenishment_rules WHERE organization_id = $1 AND product_category_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment rules by category: %w", err)
	}
	defer rows.Close()

	var rules []types.ReplenishmentRule
	for rows.Next() {
		var rule types.ReplenishmentRule
		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.CompanyID, &rule.Name, &rule.Description,
			&rule.ProductID, &rule.ProductCategoryID, &rule.WarehouseID, &rule.LocationID,
			&rule.TriggerType, &rule.MinQuantity, &rule.MaxQuantity, &rule.ReorderPoint,
			&rule.SafetyStock, &rule.ProcureMethod, &rule.OrderQuantity, &rule.MultipleOf,
			&rule.LeadTimeDays, &rule.CheckFrequency, &rule.LastCheckedAt, &rule.NextCheckAt,
			&rule.SourceLocationID, &rule.DestLocationID, &rule.Active, &rule.Priority,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *replenishmentRuleRepository) FindByWarehouse(ctx context.Context, organizationID, warehouseID uuid.UUID) ([]types.ReplenishmentRule, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
		FROM replenishment_rules WHERE organization_id = $1 AND warehouse_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment rules by warehouse: %w", err)
	}
	defer rows.Close()

	var rules []types.ReplenishmentRule
	for rows.Next() {
		var rule types.ReplenishmentRule
		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.CompanyID, &rule.Name, &rule.Description,
			&rule.ProductID, &rule.ProductCategoryID, &rule.WarehouseID, &rule.LocationID,
			&rule.TriggerType, &rule.MinQuantity, &rule.MaxQuantity, &rule.ReorderPoint,
			&rule.SafetyStock, &rule.ProcureMethod, &rule.OrderQuantity, &rule.MultipleOf,
			&rule.LeadTimeDays, &rule.CheckFrequency, &rule.LastCheckedAt, &rule.NextCheckAt,
			&rule.SourceLocationID, &rule.DestLocationID, &rule.Active, &rule.Priority,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *replenishmentRuleRepository) FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]types.ReplenishmentRule, error) {
	query := `
		SELECT id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
		FROM replenishment_rules WHERE organization_id = $1 AND location_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find replenishment rules by location: %w", err)
	}
	defer rows.Close()

	var rules []types.ReplenishmentRule
	for rows.Next() {
		var rule types.ReplenishmentRule
		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.CompanyID, &rule.Name, &rule.Description,
			&rule.ProductID, &rule.ProductCategoryID, &rule.WarehouseID, &rule.LocationID,
			&rule.TriggerType, &rule.MinQuantity, &rule.MaxQuantity, &rule.ReorderPoint,
			&rule.SafetyStock, &rule.ProcureMethod, &rule.OrderQuantity, &rule.MultipleOf,
			&rule.LeadTimeDays, &rule.CheckFrequency, &rule.LastCheckedAt, &rule.NextCheckAt,
			&rule.SourceLocationID, &rule.DestLocationID, &rule.Active, &rule.Priority,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *replenishmentRuleRepository) Update(ctx context.Context, rule types.ReplenishmentRule) (*types.ReplenishmentRule, error) {
	query := `
		UPDATE replenishment_rules
		SET name = $2, description = $3, product_id = $4, product_category_id = $5,
		 warehouse_id = $6, location_id = $7, trigger_type = $8, min_quantity = $9,
		 max_quantity = $10, reorder_point = $11, safety_stock = $12, procure_method = $13,
		 order_quantity = $14, multiple_of = $15, lead_time_days = $16, check_frequency = $17,
		 last_checked_at = $18, next_check_at = $19, source_location_id = $20, dest_location_id = $21,
		 active = $22, priority = $23, updated_at = $24
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, name, description, product_id, product_category_id,
		 warehouse_id, location_id, trigger_type, min_quantity, max_quantity, reorder_point,
		 safety_stock, procure_method, order_quantity, multiple_of, lead_time_days,
		 check_frequency, last_checked_at, next_check_at, source_location_id, dest_location_id,
		 active, priority, created_at, updated_at
	`

	rule.UpdatedAt = time.Now()
	var updated types.ReplenishmentRule
	err := r.db.QueryRowContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.ProductID, rule.ProductCategoryID,
		rule.WarehouseID, rule.LocationID, rule.TriggerType, rule.MinQuantity,
		rule.MaxQuantity, rule.ReorderPoint, rule.SafetyStock, rule.ProcureMethod,
		rule.OrderQuantity, rule.MultipleOf, rule.LeadTimeDays, rule.CheckFrequency,
		rule.LastCheckedAt, rule.NextCheckAt, rule.SourceLocationID, rule.DestLocationID,
		rule.Active, rule.Priority, rule.UpdatedAt,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.Name, &updated.Description,
		&updated.ProductID, &updated.ProductCategoryID, &updated.WarehouseID, &updated.LocationID,
		&updated.TriggerType, &updated.MinQuantity, &updated.MaxQuantity, &updated.ReorderPoint,
		&updated.SafetyStock, &updated.ProcureMethod, &updated.OrderQuantity, &updated.MultipleOf,
		&updated.LeadTimeDays, &updated.CheckFrequency, &updated.LastCheckedAt, &updated.NextCheckAt,
		&updated.SourceLocationID, &updated.DestLocationID, &updated.Active, &updated.Priority,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("replenishment rule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update replenishment rule: %w", err)
	}
	return &updated, nil
}

func (r *replenishmentRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE replenishment_rules SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete replenishment rule: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("replenishment rule not found")
	}
	return nil
}

func (r *replenishmentRuleRepository) CheckAndCreateReplenishmentOrders(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.ReplenishmentCheckResult, error) {
	query := `
		SELECT
			order_id, product_id, product_name, quantity, status, rule_name
		FROM check_and_create_replenishment_orders($1, $2)
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to check and create replenishment orders: %w", err)
	}
	defer rows.Close()

	var results []types.ReplenishmentCheckResult
	for rows.Next() {
		var result types.ReplenishmentCheckResult
		var orderID, ruleName sql.NullString
		var status sql.NullString

		err := rows.Scan(
			&orderID, &result.ProductID, &result.ProductName, &result.RecommendedQuantity, &status, &ruleName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replenishment check result: %w", err)
		}

		// Set default values if null
		if ruleName.Valid {
			result.RuleName = ruleName.String
		}
		if status.Valid {
			result.Status = status.String
		}

		results = append(results, result)
	}

	return results, nil
}

func (r *replenishmentRuleRepository) UpdateRuleCheckTimes(ctx context.Context, organizationID uuid.UUID) error {
	query := `SELECT update_replenishment_rule_check_times($1)`

	_, err := r.db.ExecContext(ctx, query, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update replenishment rule check times: %w", err)
	}
	return nil
}
