package types

import (
	"time"

	"github.com/google/uuid"
)

// Warehouse represents a physical warehouse location
type Warehouse struct {
	ID                  uuid.UUID  `json:"id" db:"id" validate:"required,uuid"`
	OrganizationID      uuid.UUID  `json:"organization_id" db:"organization_id" validate:"required,uuid"`
	CompanyID           *uuid.UUID `json:"company_id,omitempty" db:"company_id" validate:"omitempty,uuid"`
	Name                string     `json:"name" db:"name" validate:"required,min=1,max=255"`
	Code                string     `json:"code" db:"code" validate:"required,min=1,max=50"`
	PartnerID           *uuid.UUID `json:"partner_id,omitempty" db:"partner_id" validate:"omitempty,uuid"`
	LotStockID          *uuid.UUID `json:"lot_stock_id,omitempty" db:"lot_stock_id" validate:"omitempty,uuid"`
	WHInputStockLocID   *uuid.UUID `json:"wh_input_stock_loc_id,omitempty" db:"wh_input_stock_loc_id" validate:"omitempty,uuid"`
	WHQCStockLocID      *uuid.UUID `json:"wh_qc_stock_loc_id,omitempty" db:"wh_qc_stock_loc_id" validate:"omitempty,uuid"`
	WHOutputStockLocID  *uuid.UUID `json:"wh_output_stock_loc_id,omitempty" db:"wh_output_stock_loc_id" validate:"omitempty,uuid"`
	WHPackStockLocID    *uuid.UUID `json:"wh_pack_stock_loc_id,omitempty" db:"wh_pack_stock_loc_id" validate:"omitempty,uuid"`
	ReceptionSteps      string     `json:"reception_steps" db:"reception_steps" validate:"required,oneof=one_step two_steps three_steps"` // one_step, two_steps, three_steps
	DeliverySteps       string     `json:"delivery_steps" db:"delivery_steps" validate:"required,oneof=ship_only pick_ship pick_pack_ship"`   // ship_only, pick_ship, pick_pack_ship
	Active              bool       `json:"active" db:"active"`
	Sequence            int        `json:"sequence" db:"sequence" validate:"gte=0"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty" db:"created_by" validate:"omitempty,uuid"`
	UpdatedBy           *uuid.UUID `json:"updated_by,omitempty" db:"updated_by" validate:"omitempty,uuid"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// StockLocation represents a location where stock can be stored
type StockLocation struct {
	ID             uuid.UUID  `json:"id" db:"id" validate:"required,uuid"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id" validate:"required,uuid"`
	CompanyID      *uuid.UUID `json:"company_id,omitempty" db:"company_id" validate:"omitempty,uuid"`
	Name           string     `json:"name" db:"name" validate:"required,min=1,max=255"`
	CompleteName   *string    `json:"complete_name,omitempty" db:"complete_name" validate:"omitempty,max=500"`
	LocationID     *uuid.UUID `json:"location_id,omitempty" db:"location_id" validate:"omitempty,uuid"` // Parent location
	Usage          string     `json:"usage" db:"usage" validate:"required,oneof=supplier view internal customer inventory production transit"`                       // supplier, view, internal, customer, inventory, production, transit
	Barcode        *string    `json:"barcode,omitempty" db:"barcode" validate:"omitempty,min=1,max=100"`
	RemovalStrategy string    `json:"removal_strategy" db:"removal_strategy" validate:"required,oneof=fifo lifo nearest"` // fifo, lifo, nearest
	Comment        *string    `json:"comment,omitempty" db:"comment" validate:"omitempty,max=1000"`
	PosX           *int       `json:"posx,omitempty" db:"posx" validate:"omitempty,gte=0"`
	PosY           *int       `json:"posy,omitempty" db:"posy" validate:"omitempty,gte=0"`
	PosZ           *int       `json:"posz,omitempty" db:"posz" validate:"omitempty,gte=0"`
	Active         bool       `json:"active" db:"active"`
	ScrapLocation  bool       `json:"scrap_location" db:"scrap_location"`
	ReturnLocation bool       `json:"return_location" db:"return_location"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty" db:"created_by" validate:"omitempty,uuid"`
	UpdatedBy      *uuid.UUID `json:"updated_by,omitempty" db:"updated_by" validate:"omitempty,uuid"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// StockMoveCreateRequest represents a request to create a stock move
type StockMoveCreateRequest struct {
	CompanyID       *uuid.UUID `json:"company_id,omitempty" validate:"omitempty,uuid"`
	Name            string     `json:"name" validate:"required,min=1,max=255"`
	Sequence        int        `json:"sequence" validate:"gte=0"`
	Priority        string     `json:"priority" validate:"required,oneof=0 1 2 3"`
	Date            time.Time  `json:"date" validate:"required"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty" validate:"omitempty"`
	State           string     `json:"state" validate:"required,oneof=draft waiting confirmed assigned done cancel"`
	ProductID       uuid.UUID  `json:"product_id" validate:"required,uuid"`
	ProductUomID    *uuid.UUID `json:"product_uom_id,omitempty" validate:"omitempty,uuid"`
	LocationID      uuid.UUID  `json:"location_id" validate:"required,uuid"`
	LocationDestID  uuid.UUID  `json:"location_dest_id" validate:"required,uuid"`
	PickingID       *uuid.UUID `json:"picking_id,omitempty" validate:"omitempty,uuid"`
	Quantity        float64    `json:"quantity" validate:"required,gt=0"`
	ReservedQuantity float64   `json:"reserved_quantity" validate:"gte=0"`
	Note            *string    `json:"note,omitempty" validate:"omitempty,max=1000"`
}

// StockMoveUpdateRequest represents a request to update a stock move
type StockMoveUpdateRequest struct {
	Sequence        *int        `json:"sequence,omitempty" validate:"omitempty,gte=0"`
	Priority        *string     `json:"priority,omitempty" validate:"omitempty,oneof=0 1 2 3"`
	Date            *time.Time  `json:"date,omitempty" validate:"omitempty"`
	ScheduledDate   *time.Time  `json:"scheduled_date,omitempty" validate:"omitempty"`
	State           *string     `json:"state,omitempty" validate:"omitempty,oneof=draft waiting confirmed assigned done cancel"`
	ProductID       *uuid.UUID  `json:"product_id,omitempty" validate:"omitempty,uuid"`
	ProductUomID    *uuid.UUID  `json:"product_uom_id,omitempty" validate:"omitempty,uuid"`
	LocationID      *uuid.UUID  `json:"location_id,omitempty" validate:"omitempty,uuid"`
	LocationDestID  *uuid.UUID  `json:"location_dest_id,omitempty" validate:"omitempty,uuid"`
	PickingID       *uuid.UUID  `json:"picking_id,omitempty" validate:"omitempty,uuid"`
	Quantity        *float64    `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	ReservedQuantity *float64    `json:"reserved_quantity,omitempty" validate:"omitempty,gte=0"`
	Note            *string     `json:"note,omitempty" validate:"omitempty,max=1000"`
}

// StockPickingCreateRequest represents a request to create a stock picking
type StockPickingCreateRequest struct {
	CompanyID       *uuid.UUID `json:"company_id,omitempty"`
	Name            string     `json:"name"`
	SequenceCode    *string    `json:"sequence_code,omitempty"`
	PickingTypeID   *uuid.UUID `json:"picking_type_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	LocationDestID  *uuid.UUID `json:"location_dest_id,omitempty"`
	PartnerID       *uuid.UUID `json:"partner_id,omitempty"`
	Date            time.Time  `json:"date"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty"`
	State           string     `json:"state"`
	Priority        string     `json:"priority"`
	Origin          *string    `json:"origin,omitempty"`
	Note            *string    `json:"note,omitempty"`
}

// StockPickingUpdateRequest represents a request to update a stock picking
type StockPickingUpdateRequest struct {
	SequenceCode    *string    `json:"sequence_code,omitempty"`
	PickingTypeID   *uuid.UUID `json:"picking_type_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	LocationDestID  *uuid.UUID `json:"location_dest_id,omitempty"`
	PartnerID       *uuid.UUID `json:"partner_id,omitempty"`
	Date            *time.Time `json:"date,omitempty"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty"`
	State           *string    `json:"state,omitempty"`
	Priority        *string    `json:"priority,omitempty"`
	Origin          *string    `json:"origin,omitempty"`
	Note            *string    `json:"note,omitempty"`
}

// StockQuant represents the current stock level at a specific location
type StockQuant struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID        *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	ProductID        uuid.UUID  `json:"product_id" db:"product_id"`
	LocationID       uuid.UUID  `json:"location_id" db:"location_id"`
	LotID            *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	PackageID        *uuid.UUID `json:"package_id,omitempty" db:"package_id"`
	OwnerID          *uuid.UUID `json:"owner_id,omitempty" db:"owner_id"`
	Quantity         float64    `json:"quantity" db:"quantity"`
	ReservedQuantity float64    `json:"reserved_quantity" db:"reserved_quantity"`
	InDate           *time.Time `json:"in_date,omitempty" db:"in_date"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// StockMove represents a movement of stock from one location to another
type StockMove struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	OrganizationID  uuid.UUID   `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID  `json:"company_id,omitempty" db:"company_id"`
	Name            string      `json:"name" db:"name"`
	Sequence        int         `json:"sequence" db:"sequence"`
	Priority        string      `json:"priority" db:"priority"` // 0, 1, 2, 3
	Date            time.Time   `json:"date" db:"date"`
	DateDeadline    *time.Time  `json:"date_deadline,omitempty" db:"date_deadline"`
	ProductID       uuid.UUID   `json:"product_id" db:"product_id"`
	ProductUOMQty   float64     `json:"product_uom_qty" db:"product_uom_qty"`
	ProductUOM      *uuid.UUID  `json:"product_uom,omitempty" db:"product_uom"`
	LocationID      uuid.UUID   `json:"location_id" db:"location_id"`
	LocationDestID  uuid.UUID   `json:"location_dest_id" db:"location_dest_id"`
	PartnerID       *uuid.UUID  `json:"partner_id,omitempty" db:"partner_id"`
	PickingID       *uuid.UUID  `json:"picking_id,omitempty" db:"picking_id"`
	State           string      `json:"state" db:"state"` // draft, waiting, confirmed, assigned, done, cancel
	ProcureMethod   string      `json:"procure_method" db:"procure_method"` // make_to_stock, make_to_order
	Origin          *string     `json:"origin,omitempty" db:"origin"`
	GroupID         *uuid.UUID  `json:"group_id,omitempty" db:"group_id"`
	RuleID          *uuid.UUID  `json:"rule_id,omitempty" db:"rule_id"`
	LotIDs          []uuid.UUID `json:"lot_ids,omitempty" db:"lot_ids"`
	Note            *string     `json:"note,omitempty" db:"note"`
	Reference       *string     `json:"reference,omitempty" db:"reference"`
	ScheduledDate   *time.Time  `json:"scheduled_date,omitempty" db:"scheduled_date"`
	Quantity        float64     `json:"quantity" db:"quantity"`
	ReservedQuantity float64    `json:"reserved_quantity" db:"reserved_quantity"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy       *uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *uuid.UUID  `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt       *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

// StockPicking represents a picking operation (delivery, receipt, etc.)
type StockPicking struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID      *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name           string     `json:"name" db:"name"`
	SequenceCode   *string    `json:"sequence_code,omitempty" db:"sequence_code"`
	PickingTypeID  *uuid.UUID `json:"picking_type_id,omitempty" db:"picking_type_id"`
	LocationID     *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	LocationDestID *uuid.UUID `json:"location_dest_id,omitempty" db:"location_dest_id"`
	PartnerID      *uuid.UUID `json:"partner_id,omitempty" db:"partner_id"`
	Date           time.Time  `json:"date" db:"date"`
	ScheduledDate  *time.Time `json:"scheduled_date,omitempty" db:"scheduled_date"`
	DateDeadline   *time.Time `json:"date_deadline,omitempty" db:"date_deadline"`
	DateDone       *time.Time `json:"date_done,omitempty" db:"date_done"`
	Origin         *string    `json:"origin,omitempty" db:"origin"`
	State          string     `json:"state" db:"state"` // draft, waiting, confirmed, assigned, done, cancel
	Priority       string     `json:"priority" db:"priority"` // 0, 1, 2, 3
	UserID         *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	OwnerID        *uuid.UUID `json:"owner_id,omitempty" db:"owner_id"`
	Note           *string    `json:"note,omitempty" db:"note"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy      *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
