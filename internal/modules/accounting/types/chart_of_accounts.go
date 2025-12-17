package types

import (
	"time"

	"github.com/google/uuid"
)

type AccountType struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AccountGroup struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID      uuid.UUID  `json:"company_id" db:"company_id"`
	Name           string     `json:"name" db:"name"`
	Code           string     `json:"code" db:"code"`
	TypeID         uuid.UUID  `json:"type_id" db:"type_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type Account struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	OrganizationID uuid.UUID   `json:"organization_id" db:"organization_id"`
	CompanyID      *uuid.UUID  `json:"company_id,omitempty" db:"company_id"`
	Name           string      `json:"name" db:"name"`
	Code           string      `json:"code" db:"code"`
	Deprecated     bool        `json:"deprecated" db:"deprecated"`
	AccountType    string      `json:"account_type" db:"account_type"`
	InternalType   *string     `json:"internal_type,omitempty" db:"internal_type"`
	InternalGroup  *string     `json:"internal_group,omitempty" db:"internal_group"`
	UserTypeID     *uuid.UUID  `json:"user_type_id,omitempty" db:"user_type_id"`
	Reconcile      bool        `json:"reconcile" db:"reconcile"`
	CurrencyID     *uuid.UUID  `json:"currency_id,omitempty" db:"currency_id"`
	GroupID        *uuid.UUID  `json:"group_id,omitempty" db:"group_id"`
	TaxIDs         []uuid.UUID `json:"tax_ids,omitempty" db:"tax_ids"`
	Note           *string     `json:"note,omitempty" db:"note"`
	TagIDs         []uuid.UUID `json:"tag_ids,omitempty" db:"tag_ids"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy      *uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy      *uuid.UUID  `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt      *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Journal struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID        *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name             string     `json:"name" db:"name"`
	Code             string     `json:"code" db:"code"`
	Type             string     `json:"type" db:"type"` // sale, purchase, cash, bank, general
	DefaultAccountID *uuid.UUID `json:"default_account_id,omitempty" db:"default_account_id"`
	RefundSequence   bool       `json:"refund_sequence" db:"refund_sequence"`
	SequenceID       *uuid.UUID `json:"sequence_id,omitempty" db:"sequence_id"`
	CurrencyID       *uuid.UUID `json:"currency_id,omitempty" db:"currency_id"`
	BankAccountID    *uuid.UUID `json:"bank_account_id,omitempty" db:"bank_account_id"`
	Color            *int       `json:"color,omitempty" db:"color"`
	Active           bool       `json:"active" db:"active"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type Tax struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	OrganizationID    uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID         *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name              string     `json:"name" db:"name"`
	TypeTaxUse        *string    `json:"type_tax_use,omitempty" db:"type_tax_use"` // sale, purchase, none
	AmountType        string     `json:"amount_type" db:"amount_type"`             // percent, fixed, division, group
	Amount            float64    `json:"amount" db:"amount"`
	PriceInclude      bool       `json:"price_include" db:"price_include"`
	IncludeBaseAmount bool       `json:"include_base_amount" db:"include_base_amount"`
	IsBaseAffected    bool       `json:"is_base_affected" db:"is_base_affected"`
	Description       *string    `json:"description,omitempty" db:"description"`
	Sequence          int        `json:"sequence" db:"sequence"`
	Active            bool       `json:"active" db:"active"`
	TaxGroupID        *uuid.UUID `json:"tax_group_id,omitempty" db:"tax_group_id"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}
