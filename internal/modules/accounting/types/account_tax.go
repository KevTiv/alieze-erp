package types

import (
	"time"

	"github.com/google/uuid"
)

// AccountTaxTypeUse represents the type of tax use
type AccountTaxTypeUse string

const (
	AccountTaxTypeUseSale     AccountTaxTypeUse = "sale"
	AccountTaxTypeUsePurchase AccountTaxTypeUse = "purchase"
	AccountTaxTypeUseNone     AccountTaxTypeUse = "none"
)

// AccountTaxAmountType represents how the tax amount is calculated
type AccountTaxAmountType string

const (
	AccountTaxAmountTypeGroup    AccountTaxAmountType = "group"
	AccountTaxAmountTypeFixed    AccountTaxAmountType = "fixed"
	AccountTaxAmountTypePercent  AccountTaxAmountType = "percent"
	AccountTaxAmountTypeDivision AccountTaxAmountType = "division"
)

// AccountTax represents a tax definition
type AccountTax struct {
	ID              uuid.UUID          `json:"id" db:"id"`
	OrganizationID  uuid.UUID          `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID         `json:"company_id,omitempty" db:"company_id"`
	Name            string             `json:"name" db:"name"`
	TypeTaxUse      AccountTaxTypeUse  `json:"type_tax_use" db:"type_tax_use"`
	AmountType      AccountTaxAmountType `json:"amount_type" db:"amount_type"`
	Amount          float64            `json:"amount" db:"amount"`
	PriceInclude    bool               `json:"price_include" db:"price_include"`
	IncludeBaseAmount bool             `json:"include_base_amount" db:"include_base_amount"`
	IsBaseAffected  bool               `json:"is_base_affected" db:"is_base_affected"`
	Description     *string            `json:"description,omitempty" db:"description"`
	Sequence        int                `json:"sequence" db:"sequence"`
	Active          bool               `json:"active" db:"active"`
	TaxGroupID      *uuid.UUID         `json:"tax_group_id,omitempty" db:"tax_group_id"`
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
}

// AccountTaxFilter represents filtering criteria for taxes
type AccountTaxFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	Name           *string
	TypeTaxUse     *AccountTaxTypeUse
	AmountType     *AccountTaxAmountType
	Active         *bool
	TaxGroupID     *uuid.UUID
	Limit          int
	Offset         int
}

// AccountTaxCreateRequest represents a request to create a tax
type AccountTaxCreateRequest struct {
	CompanyID      *uuid.UUID         `json:"company_id,omitempty"`
	Name           string             `json:"name"`
	TypeTaxUse     AccountTaxTypeUse  `json:"type_tax_use"`
	AmountType     AccountTaxAmountType `json:"amount_type"`
	Amount         float64            `json:"amount"`
	PriceInclude   bool               `json:"price_include"`
	IncludeBaseAmount bool            `json:"include_base_amount"`
	IsBaseAffected bool               `json:"is_base_affected"`
	Description    *string            `json:"description,omitempty"`
	Sequence       int                `json:"sequence"`
	Active         bool               `json:"active"`
	TaxGroupID     *uuid.UUID         `json:"tax_group_id,omitempty"`
}

// AccountTaxUpdateRequest represents a request to update a tax
type AccountTaxUpdateRequest struct {
	CompanyID      *uuid.UUID         `json:"company_id,omitempty"`
	Name           *string            `json:"name,omitempty"`
	TypeTaxUse     *AccountTaxTypeUse  `json:"type_tax_use,omitempty"`
	AmountType     *AccountTaxAmountType `json:"amount_type,omitempty"`
	Amount         *float64           `json:"amount,omitempty"`
	PriceInclude   *bool              `json:"price_include,omitempty"`
	IncludeBaseAmount *bool           `json:"include_base_amount,omitempty"`
	IsBaseAffected *bool              `json:"is_base_affected,omitempty"`
	Description    *string            `json:"description,omitempty"`
	Sequence       *int               `json:"sequence,omitempty"`
	Active         *bool              `json:"active,omitempty"`
	TaxGroupID     *uuid.UUID         `json:"tax_group_id,omitempty"`
}
