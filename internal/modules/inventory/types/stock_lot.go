package types

import (
	"time"
	"github.com/google/uuid"
)

// StockLot represents a stock lot (batch/serial number)
type StockLot struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID `json:"company_id" db:"company_id"`
	Name            string    `json:"name" db:"name"`
	Ref             *string   `json:"ref" db:"ref"`
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	ExpirationDate  *time.Time `json:"expiration_date" db:"expiration_date"`
	UseDate         *time.Time `json:"use_date" db:"use_date"`
	RemovalDate     *time.Time `json:"removal_date" db:"removal_date"`
	AlertDate       *time.Time `json:"alert_date" db:"alert_date"`
	Note            *string   `json:"note" db:"note"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// StockLotCreateRequest represents the request to create a stock lot
type StockLotCreateRequest struct {
	Name            string     `json:"name" validate:"required,min=1,max=255"`
	Ref             *string    `json:"ref"`
	ProductID       uuid.UUID  `json:"product_id" validate:"required"`
	ExpirationDate  *time.Time `json:"expiration_date"`
	UseDate         *time.Time `json:"use_date"`
	RemovalDate     *time.Time `json:"removal_date"`
	AlertDate       *time.Time `json:"alert_date"`
	Note            *string    `json:"note"`
}

// StockLotUpdateRequest represents the request to update a stock lot
type StockLotUpdateRequest struct {
	Ref             *string    `json:"ref"`
	ExpirationDate  *time.Time `json:"expiration_date"`
	UseDate         *time.Time `json:"use_date"`
	RemovalDate     *time.Time `json:"removal_date"`
	AlertDate       *time.Time `json:"alert_date"`
	Note            *string    `json:"note"`
}
