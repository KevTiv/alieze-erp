package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReplenishmentRule represents automatic stock replenishment rules
// These rules define when and how to automatically replenish stock
type ReplenishmentRule struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name            string `json:"name" db:"name"`
	Description     *string `json:"description,omitempty" db:"description"`
	ProductID       *uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	ProductCategoryID *uuid.UUID `json:"product_category_id,omitempty" db:"product_category_id"`
	WarehouseID     *uuid.UUID `json:"warehouse_id,omitempty" db:"warehouse_id"`
	LocationID      *uuid.UUID `json:"location_id,omitempty" db:"location_id"`

	// Replenishment triggers
	TriggerType     string `json:"trigger_type" db:"trigger_type"` // "reorder_point", "safety_stock", "manual"
	MinQuantity     *float64 `json:"min_quantity,omitempty" db:"min_quantity"`
	MaxQuantity     *float64 `json:"max_quantity,omitempty" db:"max_quantity"`
	ReorderPoint    *float64 `json:"reorder_point,omitempty" db:"reorder_point"`
	SafetyStock     *float64 `json:"safety_stock,omitempty" db:"safety_stock"`

	// Procurement settings
	ProcureMethod   string `json:"procure_method" db:"procure_method"` // "make_to_stock", "make_to_order"
	OrderQuantity   *float64 `json:"order_quantity,omitempty" db:"order_quantity"`
	MultipleOf      *float64 `json:"multiple_of,omitempty" db:"multiple_of"`
	LeadTimeDays    *int `json:"lead_time_days,omitempty" db:"lead_time_days"`

	// Scheduling
	CheckFrequency  string `json:"check_frequency" db:"check_frequency"` // "daily", "weekly", "monthly", "real_time"
	LastCheckedAt   *time.Time `json:"last_checked_at,omitempty" db:"last_checked_at"`
	NextCheckAt     *time.Time `json:"next_check_at,omitempty" db:"next_check_at"`

	// Source and destination
	SourceLocationID *uuid.UUID `json:"source_location_id,omitempty" db:"source_location_id"`
	DestLocationID   *uuid.UUID `json:"dest_location_id,omitempty" db:"dest_location_id"`

	// Status
	Active          bool `json:"active" db:"active"`
	Priority        int `json:"priority" db:"priority"`

	// Standard fields
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ReplenishmentOrder represents an automatically generated replenishment order
// This can be used to create purchase orders, manufacturing orders, or transfer orders
type ReplenishmentOrder struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	RuleID          uuid.UUID `json:"rule_id" db:"rule_id"`
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	ProductName     string `json:"product_name" db:"product_name"`
	Quantity        float64 `json:"quantity" db:"quantity"`
	UOMID           *uuid.UUID `json:"uom_id,omitempty" db:"uom_id"`

	// Source and destination
	SourceLocationID *uuid.UUID `json:"source_location_id,omitempty" db:"source_location_id"`
	DestLocationID   *uuid.UUID `json:"dest_location_id,omitempty" db:"dest_location_id"`

	// Status
	Status          string `json:"status" db:"status"` // "draft", "confirmed", "processed", "cancelled"
	Priority        int `json:"priority" db:"priority"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty" db:"scheduled_date"`

	// Procurement details
	ProcureMethod   string `json:"procure_method" db:"procure_method"`
	Reference       *string `json:"reference,omitempty" db:"reference"`
	Notes           *string `json:"notes,omitempty" db:"notes"`

	// Resulting documents
	PurchaseOrderID *uuid.UUID `json:"purchase_order_id,omitempty" db:"purchase_order_id"`
	ManufacturingOrderID *uuid.UUID `json:"manufacturing_order_id,omitempty" db:"manufacturing_order_id"`
	TransferID      *uuid.UUID `json:"transfer_id,omitempty" db:"transfer_id"`

	// Standard fields
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ReplenishmentCheckResult represents the result of checking replenishment rules
// This is used to determine what needs to be replenished
type ReplenishmentCheckResult struct {
	ProductID       uuid.UUID `json:"product_id"`
	ProductName     string `json:"product_name"`
	CurrentQuantity float64 `json:"current_quantity"`
	ReorderPoint    float64 `json:"reorder_point"`
	SafetyStock     float64 `json:"safety_stock"`
	RecommendedQuantity float64 `json:"recommended_quantity"`
	LocationID      uuid.UUID `json:"location_id"`
	LocationName    string `json:"location_name"`
	RuleID          uuid.UUID `json:"rule_id"`
	RuleName        string `json:"rule_name"`
	Priority        int `json:"priority"`
	ProcureMethod   string `json:"procure_method"`
}
