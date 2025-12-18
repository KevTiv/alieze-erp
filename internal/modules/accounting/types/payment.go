package types

import (
	"time"

	"github.com/google/uuid"
)

// PaymentType represents the type of payment
type PaymentType string

const (
	PaymentTypeInbound  PaymentType = "inbound"
	PaymentTypeOutbound PaymentType = "outbound"
)

// PartnerType represents the type of partner
type PartnerType string

const (
	PartnerTypeCustomer PartnerType = "customer"
	PartnerTypeSupplier PartnerType = "supplier"
)

// PaymentState represents the state of a payment
type PaymentState string

const (
	PaymentStateDraft      PaymentState = "draft"
	PaymentStatePosted     PaymentState = "posted"
	PaymentStateSent       PaymentState = "sent"
	PaymentStateReconciled PaymentState = "reconciled"
	PaymentStateCancelled  PaymentState = "cancelled"
)

// Payment represents a payment
type Payment struct {
	ID                   uuid.UUID      `json:"id" db:"id"`
	OrganizationID       uuid.UUID      `json:"organization_id" db:"organization_id"`
	CompanyID            *uuid.UUID     `json:"company_id,omitempty" db:"company_id"`
	PaymentType          PaymentType    `json:"payment_type" db:"payment_type"`
	PartnerType          PartnerType    `json:"partner_type" db:"partner_type"`
	PartnerID            *uuid.UUID     `json:"partner_id,omitempty" db:"partner_id"`
	Amount               float64        `json:"amount" db:"amount"`
	CurrencyID           *uuid.UUID     `json:"currency_id,omitempty" db:"currency_id"`
	PaymentDate          time.Time      `json:"payment_date" db:"payment_date"`
	Communication        *string        `json:"communication,omitempty" db:"communication"`
	JournalID            *uuid.UUID     `json:"journal_id,omitempty" db:"journal_id"`
	PaymentMethodID      *uuid.UUID     `json:"payment_method_id,omitempty" db:"payment_method_id"`
	DestinationAccountID *uuid.UUID     `json:"destination_account_id,omitempty" db:"destination_account_id"`
	State                PaymentState   `json:"state" db:"state"`
	Name                 *string        `json:"name,omitempty" db:"name"`
	Ref                  *string        `json:"ref,omitempty" db:"ref"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy            *uuid.UUID     `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy            *uuid.UUID     `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt            *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
	CustomFields         interface{}    `json:"custom_fields,omitempty" db:"custom_fields"`
	Metadata             interface{}    `json:"metadata,omitempty" db:"metadata"`
}

// PaymentFilter represents filtering criteria for payments
type PaymentFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	PaymentType    *PaymentType
	PartnerType    *PartnerType
	PartnerID      *uuid.UUID
	State          *PaymentState
	Limit          int
	Offset         int
}

// PaymentCreateRequest represents a request to create a payment
type PaymentCreateRequest struct {
	CompanyID            *uuid.UUID    `json:"company_id,omitempty"`
	PaymentType          PaymentType   `json:"payment_type"`
	PartnerType          PartnerType   `json:"partner_type"`
	PartnerID            *uuid.UUID    `json:"partner_id,omitempty"`
	Amount               float64       `json:"amount"`
	CurrencyID           *uuid.UUID    `json:"currency_id,omitempty"`
	PaymentDate          time.Time     `json:"payment_date"`
	Communication        *string       `json:"communication,omitempty"`
	JournalID            *uuid.UUID    `json:"journal_id,omitempty"`
	PaymentMethodID      *uuid.UUID    `json:"payment_method_id,omitempty"`
	DestinationAccountID *uuid.UUID    `json:"destination_account_id,omitempty"`
	State                PaymentState  `json:"state"`
	Name                 *string       `json:"name,omitempty"`
	Ref                  *string       `json:"ref,omitempty"`
	CustomFields         interface{}   `json:"custom_fields,omitempty"`
	Metadata             interface{}   `json:"metadata,omitempty"`
}

// PaymentUpdateRequest represents a request to update a payment
type PaymentUpdateRequest struct {
	CompanyID            *uuid.UUID    `json:"company_id,omitempty"`
	PaymentType          *PaymentType  `json:"payment_type,omitempty"`
	PartnerType          *PartnerType  `json:"partner_type,omitempty"`
	PartnerID            *uuid.UUID    `json:"partner_id,omitempty"`
	Amount               *float64      `json:"amount,omitempty"`
	CurrencyID           *uuid.UUID    `json:"currency_id,omitempty"`
	PaymentDate          *time.Time    `json:"payment_date,omitempty"`
	Communication        *string       `json:"communication,omitempty"`
	JournalID            *uuid.UUID    `json:"journal_id,omitempty"`
	PaymentMethodID      *uuid.UUID    `json:"payment_method_id,omitempty"`
	DestinationAccountID *uuid.UUID    `json:"destination_account_id,omitempty"`
	State                *PaymentState `json:"state,omitempty"`
	Name                 *string       `json:"name,omitempty"`
	Ref                  *string       `json:"ref,omitempty"`
	CustomFields         interface{}   `json:"custom_fields,omitempty"`
	Metadata             interface{}   `json:"metadata,omitempty"`
}
