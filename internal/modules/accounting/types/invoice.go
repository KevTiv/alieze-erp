package types

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusOpen      InvoiceStatus = "open"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

type InvoiceType string

const (
	InvoiceTypeCustomer InvoiceType = "customer"
	InvoiceTypeSupplier InvoiceType = "supplier"
)

type Invoice struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	OrganizationID   uuid.UUID     `json:"organization_id" db:"organization_id"`
	CompanyID        uuid.UUID     `json:"company_id" db:"company_id"`
	PartnerID        uuid.UUID     `json:"partner_id" db:"partner_id"`
	Reference        string        `json:"reference" db:"reference"`
	Status           InvoiceStatus `json:"status" db:"status"`
	Type             InvoiceType   `json:"type" db:"type"`
	InvoiceDate      time.Time     `json:"invoice_date" db:"invoice_date"`
	DueDate          time.Time     `json:"due_date" db:"due_date"`
	PaymentTermID    *uuid.UUID    `json:"payment_term_id,omitempty" db:"payment_term_id"`
	FiscalPositionID *uuid.UUID    `json:"fiscal_position_id,omitempty" db:"fiscal_position_id"`
	CurrencyID       uuid.UUID     `json:"currency_id" db:"currency_id"`
	JournalID        uuid.UUID     `json:"journal_id" db:"journal_id"`
	AmountUntaxed    float64       `json:"amount_untaxed" db:"amount_untaxed"`
	AmountTax        float64       `json:"amount_tax" db:"amount_tax"`
	AmountTotal      float64       `json:"amount_total" db:"amount_total"`
	AmountResidual   float64       `json:"amount_residual" db:"amount_residual"`
	Note             string        `json:"note" db:"note"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
	CreatedBy        uuid.UUID     `json:"created_by" db:"created_by"`
	UpdatedBy        uuid.UUID     `json:"updated_by" db:"updated_by"`
	Lines            []InvoiceLine `json:"lines" db:"-"`
	Payments         []Payment     `json:"payments" db:"-"`
}

type InvoiceLine struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	InvoiceID     uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	ProductID     *uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	ProductName   string     `json:"product_name" db:"product_name"`
	Description   string     `json:"description" db:"description"`
	Quantity      float64    `json:"quantity" db:"quantity"`
	UomID         *uuid.UUID `json:"uom_id,omitempty" db:"uom_id"`
	UnitPrice     float64    `json:"unit_price" db:"unit_price"`
	Discount      float64    `json:"discount" db:"discount"`
	TaxID         *uuid.UUID `json:"tax_id,omitempty" db:"tax_id"`
	PriceSubtotal float64    `json:"price_subtotal" db:"price_subtotal"`
	PriceTax      float64    `json:"price_tax" db:"price_tax"`
	PriceTotal    float64    `json:"price_total" db:"price_total"`
	Sequence      int        `json:"sequence" db:"sequence"`
	AccountID     uuid.UUID  `json:"account_id" db:"account_id"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Payment struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID      uuid.UUID `json:"company_id" db:"company_id"`
	InvoiceID      uuid.UUID `json:"invoice_id" db:"invoice_id"`
	PartnerID      uuid.UUID `json:"partner_id" db:"partner_id"`
	PaymentDate    time.Time `json:"payment_date" db:"payment_date"`
	Amount         float64   `json:"amount" db:"amount"`
	CurrencyID     uuid.UUID `json:"currency_id" db:"currency_id"`
	JournalID      uuid.UUID `json:"journal_id" db:"journal_id"`
	PaymentMethod  string    `json:"payment_method" db:"payment_method"`
	Reference      string    `json:"reference" db:"reference"`
	Note           string    `json:"note" db:"note"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy      uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy      uuid.UUID `json:"updated_by" db:"updated_by"`
}
