package repository

import (
	"context"
	"database/sql"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// AnalyticsRepository interface for inventory analytics operations
type AnalyticsRepository interface {
	// Valuation operations
	GetInventoryValuation(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryValuation, error)
	GetValuationByProduct(ctx context.Context, orgID, productID uuid.UUID) (*domain.InventoryValuation, error)
	GetValuationSummary(ctx context.Context, orgID uuid.UUID) (*domain.AnalyticsSummary, error)

	// Turnover operations
	GetInventoryTurnover(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryTurnover, error)
	GetTurnoverByProduct(ctx context.Context, orgID, productID uuid.UUID) (*domain.InventoryTurnover, error)

	// Aging operations
	GetInventoryAging(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryAging, error)
	GetAgingSummary(ctx context.Context, orgID uuid.UUID) (map[string]float64, error)

	// Dead stock operations
	GetDeadStock(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryDeadStock, error)
	GetDeadStockSummary(ctx context.Context, orgID uuid.UUID) (*domain.AnalyticsSummary, error)

	// Movement summary operations
	GetMovementSummary(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryMovementSummary, error)

	// Reorder analysis operations
	GetReorderAnalysis(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryReorderAnalysis, error)
	GetProductsNeedingReorder(ctx context.Context, orgID uuid.UUID) ([]domain.InventoryReorderAnalysis, error)

	// Snapshot operations
	GetInventorySnapshot(ctx context.Context, orgID uuid.UUID) ([]domain.InventorySnapshot, error)

	// Refresh operations
	RefreshOrganizationAnalytics(ctx context.Context, orgID uuid.UUID) error
}

type analyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// GetInventoryValuation retrieves inventory valuation data for an organization
func (r *analyticsRepository) GetInventoryValuation(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryValuation, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, default_code, category_id,
			valuation_method, total_quantity, current_value, retail_value,
			unrealized_gain_loss, currency_id, uom_id, active, created_at, updated_at
		FROM inventory_valuation
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	// Apply filters
	if len(request.ProductIDs) > 0 {
		query += ` AND product_id = ANY($2)`
		params = append(params, request.ProductIDs)
	}

	if len(request.CategoryIDs) > 0 {
		query += ` AND category_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.CategoryIDs)
	}

	// Apply sorting and pagination
	query += ` ORDER BY current_value DESC`
	if request.Limit != nil {
		query += ` LIMIT $` + string(len(params)+1)
		params = append(params, *request.Limit)
	}

	if request.Offset != nil {
		query += ` OFFSET $` + string(len(params)+1)
		params = append(params, *request.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryValuation
	for rows.Next() {
		var val domain.InventoryValuation
		err := rows.Scan(
			&val.OrganizationID, &val.ProductID, &val.ProductName,
			&val.DefaultCode, &val.CategoryID, &val.ValuationMethod,
			&val.TotalQuantity, &val.CurrentValue, &val.RetailValue,
			&val.UnrealizedGainLoss, &val.CurrencyID, &val.UOMID,
			&val.Active, &val.CreatedAt, &val.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}

	return results, nil
}

// GetValuationByProduct retrieves valuation for a specific product
func (r *analyticsRepository) GetValuationByProduct(ctx context.Context, orgID, productID uuid.UUID) (*domain.InventoryValuation, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, default_code, category_id,
			valuation_method, total_quantity, current_value, retail_value,
			unrealized_gain_loss, currency_id, uom_id, active, created_at, updated_at
		FROM inventory_valuation
		WHERE organization_id = $1 AND product_id = $2
	`

	var val domain.InventoryValuation
	err := r.db.QueryRowContext(ctx, query, orgID, productID).Scan(
		&val.OrganizationID, &val.ProductID, &val.ProductName,
		&val.DefaultCode, &val.CategoryID, &val.ValuationMethod,
		&val.TotalQuantity, &val.CurrentValue, &val.RetailValue,
		&val.UnrealizedGainLoss, &val.CurrencyID, &val.UOMID,
		&val.Active, &val.CreatedAt, &val.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &val, nil
}

// GetValuationSummary retrieves summary valuation metrics
func (r *analyticsRepository) GetValuationSummary(ctx context.Context, orgID uuid.UUID) (*domain.AnalyticsSummary, error) {
	query := `
		SELECT
			COUNT(DISTINCT product_id) as total_products,
			COALESCE(SUM(current_value), 0) as total_inventory_value,
			COALESCE(SUM(retail_value), 0) as total_retail_value
		FROM inventory_valuation
		WHERE organization_id = $1
	`

	var summary domain.AnalyticsSummary
	summary.OrganizationID = orgID
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&summary.TotalProducts,
		&summary.TotalInventoryValue,
		&summary.TotalRetailValue,
	)

	if err != nil {
		return nil, err
	}

	// Calculate additional metrics
	summary.AverageTurnoverRatio = r.getAverageTurnoverRatio(ctx, orgID)
	summary.AverageDaysOfSupply = r.getAverageDaysOfSupply(ctx, orgID)
	summary.DeadStockValue, summary.DeadStockPercentage = r.getDeadStockMetrics(ctx, orgID)
	summary.ProductsNeedingReorder = r.getProductsNeedingReorderCount(ctx, orgID)
	summary.ProductsBelowSafetyStock = r.getProductsBelowSafetyStockCount(ctx, orgID)

	return &summary, nil
}

// GetInventoryTurnover retrieves inventory turnover data
func (r *analyticsRepository) GetInventoryTurnover(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryTurnover, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, category_id,
			annual_cogs, average_inventory, turnover_ratio, days_of_supply
		FROM inventory_turnover
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	// Apply filters
	if len(request.ProductIDs) > 0 {
		query += ` AND product_id = ANY($2)`
		params = append(params, request.ProductIDs)
	}

	if len(request.CategoryIDs) > 0 {
		query += ` AND category_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.CategoryIDs)
	}

	// Apply sorting and pagination
	query += ` ORDER BY turnover_ratio DESC`
	if request.Limit != nil {
		query += ` LIMIT $` + string(len(params)+1)
		params = append(params, *request.Limit)
	}

	if request.Offset != nil {
		query += ` OFFSET $` + string(len(params)+1)
		params = append(params, *request.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryTurnover
	for rows.Next() {
		var turnover domain.InventoryTurnover
		err := rows.Scan(
			&turnover.OrganizationID, &turnover.ProductID, &turnover.ProductName,
			&turnover.CategoryID, &turnover.AnnualCOGS, &turnover.AverageInventory,
			&turnover.TurnoverRatio, &turnover.DaysOfSupply,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, turnover)
	}

	return results, nil
}

// GetTurnoverByProduct retrieves turnover for a specific product
func (r *analyticsRepository) GetTurnoverByProduct(ctx context.Context, orgID, productID uuid.UUID) (*domain.InventoryTurnover, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, category_id,
			annual_cogs, average_inventory, turnover_ratio, days_of_supply
		FROM inventory_turnover
		WHERE organization_id = $1 AND product_id = $2
	`

	var turnover domain.InventoryTurnover
	err := r.db.QueryRowContext(ctx, query, orgID, productID).Scan(
		&turnover.OrganizationID, &turnover.ProductID, &turnover.ProductName,
		&turnover.CategoryID, &turnover.AnnualCOGS, &turnover.AverageInventory,
		&turnover.TurnoverRatio, &turnover.DaysOfSupply,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &turnover, nil
}

// GetInventoryAging retrieves stock aging data
func (r *analyticsRepository) GetInventoryAging(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryAging, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, category_id, default_code,
			location_id, location_name, lot_id, lot_name, quantity, in_date,
			age_bracket, days_in_stock, value
		FROM inventory_aging
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	// Apply filters
	if len(request.ProductIDs) > 0 {
		query += ` AND product_id = ANY($2)`
		params = append(params, request.ProductIDs)
	}

	if len(request.CategoryIDs) > 0 {
		query += ` AND category_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.CategoryIDs)
	}

	if len(request.LocationIDs) > 0 {
		query += ` AND location_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.LocationIDs)
	}

	// Apply sorting and pagination
	query += ` ORDER BY value DESC`
	if request.Limit != nil {
		query += ` LIMIT $` + string(len(params)+1)
		params = append(params, *request.Limit)
	}

	if request.Offset != nil {
		query += ` OFFSET $` + string(len(params)+1)
		params = append(params, *request.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryAging
	for rows.Next() {
		var aging domain.InventoryAging
		err := rows.Scan(
			&aging.OrganizationID, &aging.ProductID, &aging.ProductName,
			&aging.CategoryID, &aging.DefaultCode, &aging.LocationID,
			&aging.LocationName, &aging.LotID, &aging.LotName, &aging.Quantity,
			&aging.InDate, &aging.AgeBracket, &aging.DaysInStock, &aging.Value,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, aging)
	}

	return results, nil
}

// GetAgingSummary retrieves aging summary by age brackets
func (r *analyticsRepository) GetAgingSummary(ctx context.Context, orgID uuid.UUID) (map[string]float64, error) {
	query := `
		SELECT
			age_bracket,
			SUM(value) as total_value
		FROM inventory_aging
		WHERE organization_id = $1
		GROUP BY age_bracket
		ORDER BY
			CASE age_bracket
				WHEN '0-30_days' THEN 1
				WHEN '31-90_days' THEN 2
				WHEN '91-180_days' THEN 3
				WHEN '181-365_days' THEN 4
				ELSE 5
			END
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make(map[string]float64)
	for rows.Next() {
		var bracket string
		var value float64
		if err := rows.Scan(&bracket, &value); err != nil {
			return nil, err
		}
		summary[bracket] = value
	}

	return summary, nil
}

// GetDeadStock retrieves dead stock analysis
func (r *analyticsRepository) GetDeadStock(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryDeadStock, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, default_code, category_id,
			last_movement_date, days_since_movement, total_quantity, total_value, dead_stock_category
		FROM inventory_dead_stock
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	// Apply filters
	if len(request.ProductIDs) > 0 {
		query += ` AND product_id = ANY($2)`
		params = append(params, request.ProductIDs)
	}

	if len(request.CategoryIDs) > 0 {
		query += ` AND category_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.CategoryIDs)
	}

	// Apply sorting and pagination
	query += ` ORDER BY total_value DESC`
	if request.Limit != nil {
		query += ` LIMIT $` + string(len(params)+1)
		params = append(params, *request.Limit)
	}

	if request.Offset != nil {
		query += ` OFFSET $` + string(len(params)+1)
		params = append(params, *request.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryDeadStock
	for rows.Next() {
		var deadStock domain.InventoryDeadStock
		err := rows.Scan(
			&deadStock.OrganizationID, &deadStock.ProductID, &deadStock.ProductName,
			&deadStock.DefaultCode, &deadStock.CategoryID, &deadStock.LastMovementDate,
			&deadStock.DaysSinceMovement, &deadStock.TotalQuantity, &deadStock.TotalValue,
			&deadStock.DeadStockCategory,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, deadStock)
	}

	return results, nil
}

// GetDeadStockSummary retrieves dead stock summary metrics
func (r *analyticsRepository) GetDeadStockSummary(ctx context.Context, orgID uuid.UUID) (*domain.AnalyticsSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_products,
			COALESCE(SUM(total_value), 0) as dead_stock_value,
			COALESCE(SUM(total_value), 0) / NULLIF(
				(SELECT COALESCE(SUM(current_value), 1) FROM inventory_valuation WHERE organization_id = $1), 0
			) * 100 as dead_stock_percentage
		FROM inventory_dead_stock
		WHERE organization_id = $1
	`

	var summary domain.AnalyticsSummary
	summary.OrganizationID = orgID
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&summary.TotalProducts,
		&summary.DeadStockValue,
		&summary.DeadStockPercentage,
	)

	if err != nil {
		return nil, err
	}

	// Get additional metrics
	summary.TotalInventoryValue = r.getTotalInventoryValue(ctx, orgID)
	summary.TotalRetailValue = r.getTotalRetailValue(ctx, orgID)
	summary.AverageTurnoverRatio = r.getAverageTurnoverRatio(ctx, orgID)
	summary.AverageDaysOfSupply = r.getAverageDaysOfSupply(ctx, orgID)
	summary.ProductsNeedingReorder = r.getProductsNeedingReorderCount(ctx, orgID)
	summary.ProductsBelowSafetyStock = r.getProductsBelowSafetyStockCount(ctx, orgID)

	return &summary, nil
}

// GetMovementSummary retrieves inventory movement summary
func (r *analyticsRepository) GetMovementSummary(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryMovementSummary, error) {
	query := `
		SELECT
			organization_id, month, product_id, product_name, category_id,
			location_name, move_count, total_quantity, total_value, avg_move_quantity
		FROM inventory_movement_summary
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	// Apply date filters
	if request.DateFrom != nil {
		query += ` AND month >= $` + string(len(params)+1)
		params = append(params, *request.DateFrom)
	}

	if request.DateTo != nil {
		query += ` AND month <= $` + string(len(params)+1)
		params = append(params, *request.DateTo)
	}

	// Apply product/category filters
	if len(request.ProductIDs) > 0 {
		query += ` AND product_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.ProductIDs)
	}

	if len(request.CategoryIDs) > 0 {
		query += ` AND category_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.CategoryIDs)
	}

	// Apply sorting and pagination
	query += ` ORDER BY month DESC, total_value DESC`
	if request.Limit != nil {
		query += ` LIMIT $` + string(len(params)+1)
		params = append(params, *request.Limit)
	}

	if request.Offset != nil {
		query += ` OFFSET $` + string(len(params)+1)
		params = append(params, *request.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryMovementSummary
	for rows.Next() {
		var summary domain.InventoryMovementSummary
		err := rows.Scan(
			&summary.OrganizationID, &summary.Month, &summary.ProductID,
			&summary.ProductName, &summary.CategoryID, &summary.LocationName,
			&summary.MoveCount, &summary.TotalQuantity, &summary.TotalValue,
			&summary.AvgMoveQuantity,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, summary)
	}

	return results, nil
}

// GetReorderAnalysis retrieves reorder analysis
func (r *analyticsRepository) GetReorderAnalysis(ctx context.Context, orgID uuid.UUID, request domain.AnalyticsRequest) ([]domain.InventoryReorderAnalysis, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, default_code, category_id,
			current_stock, reorder_point, safety_stock, lead_time_days,
			daily_consumption, days_until_reorder, reorder_status, recommended_order_quantity
		FROM inventory_reorder_analysis
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	// Apply filters
	if len(request.ProductIDs) > 0 {
		query += ` AND product_id = ANY($2)`
		params = append(params, request.ProductIDs)
	}

	if len(request.CategoryIDs) > 0 {
		query += ` AND category_id = ANY($` + string(len(params)+1) + `)`
		params = append(params, request.CategoryIDs)
	}

	// Apply sorting and pagination
	query += ` ORDER BY
		CASE reorder_status
			WHEN 'reorder_now' THEN 1
			WHEN 'reorder_soon' THEN 2
			ELSE 3
		END,
		days_until_reorder`
	if request.Limit != nil {
		query += ` LIMIT $` + string(len(params)+1)
		params = append(params, *request.Limit)
	}

	if request.Offset != nil {
		query += ` OFFSET $` + string(len(params)+1)
		params = append(params, *request.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryReorderAnalysis
	for rows.Next() {
		var analysis domain.InventoryReorderAnalysis
		err := rows.Scan(
			&analysis.OrganizationID, &analysis.ProductID, &analysis.ProductName,
			&analysis.DefaultCode, &analysis.CategoryID, &analysis.CurrentStock,
			&analysis.ReorderPoint, &analysis.SafetyStock, &analysis.LeadTimeDays,
			&analysis.DailyConsumption, &analysis.DaysUntilReorder, &analysis.ReorderStatus,
			&analysis.RecommendedOrderQuantity,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, analysis)
	}

	return results, nil
}

// GetProductsNeedingReorder retrieves products that need reordering
func (r *analyticsRepository) GetProductsNeedingReorder(ctx context.Context, orgID uuid.UUID) ([]domain.InventoryReorderAnalysis, error) {
	query := `
		SELECT
			organization_id, product_id, product_name, default_code, category_id,
			current_stock, reorder_point, safety_stock, lead_time_days,
			daily_consumption, days_until_reorder, reorder_status, recommended_order_quantity
		FROM inventory_reorder_analysis
		WHERE organization_id = $1 AND reorder_status IN ('reorder_now', 'reorder_soon')
		ORDER BY
			CASE reorder_status
				WHEN 'reorder_now' THEN 1
				ELSE 2
			END,
			days_until_reorder
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventoryReorderAnalysis
	for rows.Next() {
		var analysis domain.InventoryReorderAnalysis
		err := rows.Scan(
			&analysis.OrganizationID, &analysis.ProductID, &analysis.ProductName,
			&analysis.DefaultCode, &analysis.CategoryID, &analysis.CurrentStock,
			&analysis.ReorderPoint, &analysis.SafetyStock, &analysis.LeadTimeDays,
			&analysis.DailyConsumption, &analysis.DaysUntilReorder, &analysis.ReorderStatus,
			&analysis.RecommendedOrderQuantity,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, analysis)
	}

	return results, nil
}

// GetInventorySnapshot retrieves a comprehensive inventory snapshot
func (r *analyticsRepository) GetInventorySnapshot(ctx context.Context, orgID uuid.UUID) ([]domain.InventorySnapshot, error) {
	query := `
		SELECT
			product_id, product_name, category_id, current_stock,
			reorder_point, safety_stock, reorder_status, days_until_reorder,
			current_value, retail_value
		FROM get_inventory_snapshot($1)
		ORDER BY current_value DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.InventorySnapshot
	for rows.Next() {
		var snapshot domain.InventorySnapshot
		err := rows.Scan(
			&snapshot.ProductID, &snapshot.ProductName, &snapshot.CategoryID,
			&snapshot.CurrentStock, &snapshot.ReorderPoint, &snapshot.SafetyStock,
			&snapshot.ReorderStatus, &snapshot.DaysUntilReorder,
			&snapshot.CurrentValue, &snapshot.RetailValue,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, snapshot)
	}

	return results, nil
}

// RefreshOrganizationAnalytics refreshes analytics for a specific organization
func (r *analyticsRepository) RefreshOrganizationAnalytics(ctx context.Context, orgID uuid.UUID) error {
	query := `SELECT refresh_organization_analytics($1)`
	_, err := r.db.ExecContext(ctx, query, orgID)
	return err
}

// Helper methods for summary calculations
func (r *analyticsRepository) getTotalInventoryValue(ctx context.Context, orgID uuid.UUID) float64 {
	query := `SELECT COALESCE(SUM(current_value), 0) FROM inventory_valuation WHERE organization_id = $1`
	var value float64
	r.db.QueryRowContext(ctx, query, orgID).Scan(&value)
	return value
}

func (r *analyticsRepository) getTotalRetailValue(ctx context.Context, orgID uuid.UUID) float64 {
	query := `SELECT COALESCE(SUM(retail_value), 0) FROM inventory_valuation WHERE organization_id = $1`
	var value float64
	r.db.QueryRowContext(ctx, query, orgID).Scan(&value)
	return value
}

func (r *analyticsRepository) getAverageTurnoverRatio(ctx context.Context, orgID uuid.UUID) float64 {
	query := `SELECT COALESCE(AVG(turnover_ratio), 0) FROM inventory_turnover WHERE organization_id = $1`
	var ratio float64
	r.db.QueryRowContext(ctx, query, orgID).Scan(&ratio)
	return ratio
}

func (r *analyticsRepository) getAverageDaysOfSupply(ctx context.Context, orgID uuid.UUID) float64 {
	query := `SELECT COALESCE(AVG(days_of_supply), 0) FROM inventory_turnover WHERE organization_id = $1`
	var days float64
	r.db.QueryRowContext(ctx, query, orgID).Scan(&days)
	return days
}

func (r *analyticsRepository) getDeadStockMetrics(ctx context.Context, orgID uuid.UUID) (float64, float64) {
	query := `
		SELECT
			COALESCE(SUM(total_value), 0) as dead_stock_value,
			COALESCE(SUM(total_value), 0) / NULLIF(
				(SELECT COALESCE(SUM(current_value), 1) FROM inventory_valuation WHERE organization_id = $1), 0
			) * 100 as dead_stock_percentage
		FROM inventory_dead_stock
		WHERE organization_id = $1
	`
	var value, percentage float64
	r.db.QueryRowContext(ctx, query, orgID).Scan(&value, &percentage)
	return value, percentage
}

func (r *analyticsRepository) getProductsNeedingReorderCount(ctx context.Context, orgID uuid.UUID) int {
	query := `SELECT COUNT(*) FROM inventory_reorder_analysis WHERE organization_id = $1 AND reorder_status = 'reorder_now'`
	var count int
	r.db.QueryRowContext(ctx, query, orgID).Scan(&count)
	return count
}

func (r *analyticsRepository) getProductsBelowSafetyStockCount(ctx context.Context, orgID uuid.UUID) int {
	query := `
		SELECT COUNT(*)
		FROM inventory_reorder_analysis
		WHERE organization_id = $1
		AND current_stock <= safety_stock
	`
	var count int
	r.db.QueryRowContext(ctx, query, orgID).Scan(&count)
	return count
}
