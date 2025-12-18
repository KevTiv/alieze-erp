package types

import (
	"time"

	"github.com/google/uuid"
)

// AccountJournalType represents the type of journal
type AccountJournalType string

const (
	AccountJournalTypeSale     AccountJournalType = "sale"
	AccountJournalTypePurchase AccountJournalType = "purchase"
	AccountJournalTypeCash     AccountJournalType = "cash"
	AccountJournalTypeBank     AccountJournalType = "bank"
	AccountJournalTypeGeneral  AccountJournalType = "general"
)

// AccountJournalState represents the state of a journal
type AccountJournalState string

const (
	AccountJournalStateDraft      AccountJournalState = "draft"
	AccountJournalStatePosted     AccountJournalState = "posted"
	AccountJournalStateCancelled  AccountJournalState = "cancelled"
)

// AccountJournal represents an accounting journal
type AccountJournal struct {
	ID                   uuid.UUID           `json:"id" db:"id"`
	OrganizationID       uuid.UUID           `json:"organization_id" db:"organization_id"`
	CompanyID            *uuid.UUID          `json:"company_id,omitempty" db:"company_id"`
	Name                 string              `json:"name" db:"name"`
	Code                 string              `json:"code" db:"code"`
	Type                 AccountJournalType  `json:"type" db:"type"`
	DefaultAccountID     *uuid.UUID          `json:"default_account_id,omitempty" db:"default_account_id"`
	RefundSequence       bool                `json:"refund_sequence" db:"refund_sequence"`
	SequenceID           *uuid.UUID          `json:"sequence_id,omitempty" db:"sequence_id"`
	CurrencyID           *uuid.UUID          `json:"currency_id,omitempty" db:"currency_id"`
	BankAccountID        *uuid.UUID          `json:"bank_account_id,omitempty" db:"bank_account_id"`
	Color                *int                `json:"color,omitempty" db:"color"`
	Active               bool                `json:"active" db:"active"`
	CreatedAt            time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at" db:"updated_at"`
}

// AccountJournalFilter represents filtering criteria for journals
type AccountJournalFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	Name           *string
	Code           *string
	Type           *AccountJournalType
	Active         *bool
	Limit          int
	Offset         int
}

// AccountJournalCreateRequest represents a request to create a journal
type AccountJournalCreateRequest struct {
	CompanyID         *uuid.UUID         `json:"company_id,omitempty"`
	Name              string             `json:"name"`
	Code              string             `json:"code"`
	Type              AccountJournalType `json:"type"`
	DefaultAccountID  *uuid.UUID         `json:"default_account_id,omitempty"`
	RefundSequence    bool               `json:"refund_sequence"`
	SequenceID        *uuid.UUID         `json:"sequence_id,omitempty"`
	CurrencyID        *uuid.UUID         `json:"currency_id,omitempty"`
	BankAccountID     *uuid.UUID         `json:"bank_account_id,omitempty"`
	Color             *int               `json:"color,omitempty"`
	Active            bool               `json:"active"`
}

// AccountJournalUpdateRequest represents a request to update a journal
type AccountJournalUpdateRequest struct {
	CompanyID         *uuid.UUID         `json:"company_id,omitempty"`
	Name              *string            `json:"name,omitempty"`
	Code              *string            `json:"code,omitempty"`
	Type              *AccountJournalType `json:"type,omitempty"`
	DefaultAccountID  *uuid.UUID         `json:"default_account_id,omitempty"`
	RefundSequence    *bool              `json:"refund_sequence,omitempty"`
	SequenceID        *uuid.UUID         `json:"sequence_id,omitempty"`
	CurrencyID        *uuid.UUID         `json:"currency_id,omitempty"`
	BankAccountID     *uuid.UUID         `json:"bank_account_id,omitempty"`
	Color             *int               `json:"color,omitempty"`
	Active            *bool              `json:"active,omitempty"`
}
