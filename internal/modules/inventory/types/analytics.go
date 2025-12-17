package types

import (
	"time"

	"github.com/google/uuid"
)

// InventoryValuation represents the valuation of inventory products
type InventoryValuation struct {
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID       uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName     string     `json:"product_name" db:"product_name"`
	DefaultCode     *string    `json:"default_code,omitempty" db:"default_code"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	ValuationMethod string     `json:"valuation_method" db:"valuation_method"`
	TotalQuantity   float64    `json:"total_quantity" db:"total_quantity"`
	CurrentValue    float64    `json:"current_value" db:"current_value"`
	RetailValue     float64    `json:"retail_value" db:"retail_value"`
	UnrealizedGainLoss float64 `json:"unrealized_gain_loss" db:"unrealized_gain_loss"`
	CurrencyID      *uuid.UUID `json:"currency_id,omitempty" db:"currency_id"`
	UOMID           *uuid.UUID `json:"uom_id,omitempty" db:"uom_id"`
	Active          bool       `json:"active" db:"active"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// InventoryTurnover represents inventory turnover metrics
type InventoryTurnover struct {
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID       uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName     string     `json:"product_name" db:"product_name"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	AnnualCOGS      float64    `json:"annual_cogs" db:"annual_cogs"`
	AverageInventory float64   `json:"average_inventory" db:"average_inventory"`
	TurnoverRatio   float64    `json:"turnover_ratio" db:"turnover_ratio"`
	DaysOfSupply    float64    `json:"days_of_supply" db:"days_of_supply"`
}

// InventoryAging represents stock aging analysis
type InventoryAging struct {
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName    string     `json:"product_name" db:"product_name"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	DefaultCode    *string    `json:"default_code,omitempty" db:"default_code"`
	LocationID     uuid.UUID  `json:"location_id" db:"location_id"`
	LocationName   string     `json:"location_name" db:"location_name"`
	LotID          *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	LotName        *string    `json:"lot_name,omitempty" db:"lot_name"`
	Quantity       float64    `json:"quantity" db:"quantity"`
	InDate         *time.Time `json:"in_date,omitempty" db:"in_date"`
	AgeBracket     string     `json:"age_bracket" db:"age_bracket"`
	DaysInStock    int        `json:"days_in_stock" db:"days_in_stock"`
	Value          float64    `json:"value" db:"value"`
}

// InventoryDeadStock represents dead stock analysis
type InventoryDeadStock struct {
	OrganizationID      uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID           uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName         string     `json:"product_name" db:"product_name"`
	DefaultCode         *string    `json:"default_code,omitempty" db:"default_code"`
	CategoryID          *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	LastMovementDate    *time.Time `json:"last_movement_date,omitempty" db:"last_movement_date"`
	DaysSinceMovement   *int       `json:"days_since_movement,omitempty" db:"days_since_movement"`
	TotalQuantity       float64    `json:"total_quantity" db:"total_quantity"`
	TotalValue          float64    `json:"total_value" db:"total_value"`
	DeadStockCategory   string     `json:"dead_stock_category" db:"dead_stock_category"`
}

// InventoryMovementSummary represents inventory movement metrics
type InventoryMovementSummary struct {
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	Month           time.Time  `json:"month" db:"month"`
	ProductID       uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName     string     `json:"product_name" db:"product_name"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	LocationName    string     `json:"location_name" db:"location_name"`
	MoveCount       int        `json:"move_count" db:"move_count"`
	TotalQuantity   float64    `json:"total_quantity" db:"total_quantity"`
	TotalValue      float64    `json:"total_value" db:"total_value"`
	AvgMoveQuantity float64    `json:"avg_move_quantity" db:"avg_move_quantity"`
}

// InventoryReorderAnalysis represents reorder recommendations
type InventoryReorderAnalysis struct {
	OrganizationID          uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID               uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName             string     `json:"product_name" db:"product_name"`
	DefaultCode             *string    `json:"default_code,omitempty" db:"default_code"`
	CategoryID              *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	CurrentStock            float64    `json:"current_stock" db:"current_stock"`
	ReorderPoint            *float64   `json:"reorder_point,omitempty" db:"reorder_point"`
	SafetyStock             *float64   `json:"safety_stock,omitempty" db:"safety_stock"`
	LeadTimeDays            *int       `json:"lead_time_days,omitempty" db:"lead_time_days"`
	DailyConsumption        float64    `json:"daily_consumption" db:"daily_consumption"`
	DaysUntilReorder        *float64   `json:"days_until_reorder,omitempty" db:"days_until_reorder"`
	ReorderStatus           string     `json:"reorder_status" db:"reorder_status"`
	RecommendedOrderQuantity float64   `json:"recommended_order_quantity" db:"recommended_order_quantity"`
}

// InventorySnapshot represents a comprehensive inventory snapshot
type InventorySnapshot struct {
	ProductID               uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName             string     `json:"product_name" db:"product_name"`
	CategoryID              *uuid.UUID `json:"category_id,omitempty" db:"category_id"`
	CurrentStock            float64    `json:"current_stock" db:"current_stock"`
	ReorderPoint            *float64   `json:"reorder_point,omitempty" db:"reorder_point"`
	SafetyStock             *float64   `json:"safety_stock,omitempty" db:"safety_stock"`
	ReorderStatus           string     `json:"reorder_status" db:"reorder_status"`
	DaysUntilReorder        *float64   `json:"days_until_reorder,omitempty" db:"days_until_reorder"`
	CurrentValue            float64    `json:"current_value" db:"current_value"`
	RetailValue             float64    `json:"retail_value" db:"retail_value"`
}

// AnalyticsRequest represents request parameters for analytics endpoints
type AnalyticsRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	ProductIDs     []uuid.UUID `json:"product_ids,omitempty"`
	CategoryIDs    []uuid.UUID `json:"category_ids,omitempty"`
	LocationIDs    []uuid.UUID `json:"location_ids,omitempty"`
	DateFrom       *time.Time `json:"date_from,omitempty"`
	DateTo         *time.Time `json:"date_to,omitempty"`
	GroupBy        string     `json:"group_by,omitempty"` // product, category, location
	Limit          *int       `json:"limit,omitempty"`
	Offset         *int       `json:"offset,omitempty"`
}

// AnalyticsSummary represents summary metrics for dashboard
type AnalyticsSummary struct {
	OrganizationID          uuid.UUID `json:"organization_id" db:"organization_id"`
	TotalProducts           int       `json:"total_products" db:"total_products"`
	TotalInventoryValue     float64   `json:"total_inventory_value" db:"total_inventory_value"`
	TotalRetailValue        float64   `json:"total_retail_value" db:"total_retail_value"`
	AverageTurnoverRatio    float64   `json:"average_turnover_ratio" db:"average_turnover_ratio"`
	AverageDaysOfSupply     float64   `json:"average_days_of_supply" db:"average_days_of_supply"`
	DeadStockValue          float64   `json:"dead_stock_value" db:"dead_stock_value"`
	DeadStockPercentage     float64   `json:"dead_stock_percentage" db:"dead_stock_percentage"`
	ProductsNeedingReorder  int       `json:"products_needing_reorder" db:"products_needing_reorder"`
	ProductsBelowSafetyStock int      `json:"products_below_safety_stock" db:"products_below_safety_stock"`
}
