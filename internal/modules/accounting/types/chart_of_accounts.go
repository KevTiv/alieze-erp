package domain

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
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID      uuid.UUID `json:"company_id" db:"company_id"`
	Code           string    `json:"code" db:"code"`
	Name           string    `json:"name" db:"name"`
	TypeID         uuid.UUID `json:"type_id" db:"type_id"`
	GroupID        uuid.UUID `json:"group_id" db:"group_id"`
	CurrencyID     uuid.UUID `json:"currency_id" db:"currency_id"`
	Reconcile      bool      `json:"reconcile" db:"reconcile"`
	Active         bool      `json:"active" db:"active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Journal struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	OrganizationID         uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID              uuid.UUID  `json:"company_id" db:"company_id"`
	Name                   string     `json:"name" db:"name"`
	Code                   string     `json:"code" db:"code"`
	Type                   string     `json:"type" db:"type"`
	CurrencyID             uuid.UUID  `json:"currency_id" db:"currency_id"`
	DefaultCreditAccountID *uuid.UUID `json:"default_credit_account_id,omitempty" db:"default_credit_account_id"`
	DefaultDebitAccountID  *uuid.UUID `json:"default_debit_account_id,omitempty" db:"default_debit_account_id"`
	Active                 bool       `json:"active" db:"active"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}
