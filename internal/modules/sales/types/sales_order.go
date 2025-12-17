package types

import (
	"time"

	"github.com/google/uuid"
)

type SalesOrderStatus string

const (
	SalesOrderStatusDraft     SalesOrderStatus = "draft"
	SalesOrderStatusQuotation SalesOrderStatus = "quotation"
	SalesOrderStatusConfirmed SalesOrderStatus = "confirmed"
	SalesOrderStatusCancelled SalesOrderStatus = "cancelled"
	SalesOrderStatusDone      SalesOrderStatus = "done"
)

type SalesOrder struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	OrganizationID   uuid.UUID        `json:"organization_id" db:"organization_id"`
	CompanyID        uuid.UUID        `json:"company_id" db:"company_id"`
	CustomerID       uuid.UUID        `json:"customer_id" db:"customer_id"`
	SalesTeamID      *uuid.UUID       `json:"sales_team_id,omitempty" db:"sales_team_id"`
	Reference        string           `json:"reference" db:"reference"`
	Status           SalesOrderStatus `json:"status" db:"status"`
	OrderDate        time.Time        `json:"order_date" db:"order_date"`
	ConfirmationDate *time.Time       `json:"confirmation_date,omitempty" db:"confirmation_date"`
	ValidityDate     *time.Time       `json:"validity_date,omitempty" db:"validity_date"`
	PaymentTermID    *uuid.UUID       `json:"payment_term_id,omitempty" db:"payment_term_id"`
	FiscalPositionID *uuid.UUID       `json:"fiscal_position_id,omitempty" db:"fiscal_position_id"`
	PricelistID      uuid.UUID        `json:"pricelist_id" db:"pricelist_id"`
	CurrencyID       uuid.UUID        `json:"currency_id" db:"currency_id"`
	AmountUntaxed    float64          `json:"amount_untaxed" db:"amount_untaxed"`
	AmountTax        float64          `json:"amount_tax" db:"amount_tax"`
	AmountTotal      float64          `json:"amount_total" db:"amount_total"`
	Note             string           `json:"note" db:"note"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
	CreatedBy        uuid.UUID        `json:"created_by" db:"created_by"`
	UpdatedBy        uuid.UUID        `json:"updated_by" db:"updated_by"`
	Lines            []SalesOrderLine `json:"lines" db:"-"`
}

type SalesOrderLine struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	SalesOrderID  uuid.UUID  `json:"sales_order_id" db:"sales_order_id"`
	ProductID     uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName   string     `json:"product_name" db:"product_name"`
	Description   string     `json:"description" db:"description"`
	Quantity      float64    `json:"quantity" db:"quantity"`
	UomID         uuid.UUID  `json:"uom_id" db:"uom_id"`
	UnitPrice     float64    `json:"unit_price" db:"unit_price"`
	Discount      float64    `json:"discount" db:"discount"`
	TaxID         *uuid.UUID `json:"tax_id,omitempty" db:"tax_id"`
	PriceSubtotal float64    `json:"price_subtotal" db:"price_subtotal"`
	PriceTax      float64    `json:"price_tax" db:"price_tax"`
	PriceTotal    float64    `json:"price_total" db:"price_total"`
	Sequence      int        `json:"sequence" db:"sequence"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Pricelist struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	CompanyID      uuid.UUID       `json:"company_id" db:"company_id"`
	Name           string          `json:"name" db:"name"`
	CurrencyID     uuid.UUID       `json:"currency_id" db:"currency_id"`
	IsActive       bool            `json:"is_active" db:"is_active"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy      uuid.UUID       `json:"updated_by" db:"updated_by"`
	Items          []PricelistItem `json:"items" db:"-"`
}

type PricelistItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PricelistID uuid.UUID `json:"pricelist_id" db:"pricelist_id"`
	ProductID   uuid.UUID `json:"product_id" db:"product_id"`
	MinQuantity float64   `json:"min_quantity" db:"min_quantity"`
	FixedPrice  *float64  `json:"fixed_price,omitempty" db:"fixed_price"`
	Discount    *float64  `json:"discount,omitempty" db:"discount"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
