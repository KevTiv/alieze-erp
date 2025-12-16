package domain

import (
	"time"

	"github.com/google/uuid"
)

// BarcodeScan represents a barcode scanning event
type BarcodeScan struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID        *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ScanType      string     `json:"scan_type" db:"scan_type"` // product, location, lot, package, move
	ScannedBarcode string     `json:"scanned_barcode" db:"scanned_barcode"`
	EntityID      *uuid.UUID `json:"entity_id,omitempty" db:"entity_id"`
	EntityType    *string    `json:"entity_type,omitempty" db:"entity_type"`
	LocationID    *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	Quantity      *float64   `json:"quantity,omitempty" db:"quantity"`
	ScanTime      time.Time  `json:"scan_time" db:"scan_time"`
	DeviceInfo    *string    `json:"device_info,omitempty" db:"device_info"`
	IPAddress     *string    `json:"ip_address,omitempty" db:"ip_address"`
	Success       bool       `json:"success" db:"success"`
	ErrorMessage  *string    `json:"error_message,omitempty" db:"error_message"`
	Metadata      *string    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// MobileScanningSession represents a mobile scanning session
type MobileScanningSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID       *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	DeviceID     *string    `json:"device_id,omitempty" db:"device_id"`
	SessionType  string     `json:"session_type" db:"session_type"` // inventory, picking, receiving, counting
	Status       string     `json:"status" db:"status"`             // active, completed, cancelled
	StartTime    time.Time  `json:"start_time" db:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty" db:"end_time"`
	LocationID   *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	Reference    *string    `json:"reference,omitempty" db:"reference"`
	Metadata     *string    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// MobileScanningSessionLine represents an item in a scanning session
type MobileScanningSessionLine struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	SessionID       uuid.UUID  `json:"session_id" db:"session_id"`
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	ScanID          *uuid.UUID `json:"scan_id,omitempty" db:"scan_id"`
	ProductID       *uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	ProductVariantID *uuid.UUID `json:"product_variant_id,omitempty" db:"product_variant_id"`
	LocationID      *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	LotID           *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	PackageID       *uuid.UUID `json:"package_id,omitempty" db:"package_id"`
	Quantity        *float64   `json:"quantity,omitempty" db:"quantity"`
	ScannedQuantity *float64   `json:"scanned_quantity,omitempty" db:"scanned_quantity"`
	UOMID          *uuid.UUID `json:"uom_id,omitempty" db:"uom_id"`
	Status         string     `json:"status" db:"status"` // scanned, verified, processed, error
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	Sequence       int        `json:"sequence" db:"sequence"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// BarcodeEntity represents an entity found by barcode
type BarcodeEntity struct {
	EntityType    string     `json:"entity_type"`
	EntityID      uuid.UUID  `json:"entity_id"`
	EntityName    string     `json:"entity_name"`
	AdditionalInfo *string    `json:"additional_info,omitempty"`
}

// BarcodeScanRequest represents a barcode scan request
type BarcodeScanRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID        uuid.UUID `json:"user_id"`
	Barcode       string    `json:"barcode"`
	Quantity      *float64  `json:"quantity,omitempty"`
	LocationID    *uuid.UUID `json:"location_id,omitempty"`
	DeviceInfo    *string   `json:"device_info,omitempty"`
	IPAddress     *string   `json:"ip_address,omitempty"`
}

// BarcodeScanResponse represents a barcode scan response
type BarcodeScanResponse struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
	ScanID       uuid.UUID   `json:"scan_id"`
	Entity       *BarcodeEntity `json:"entity,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
}

// CreateScanningSessionRequest represents a request to create a scanning session
type CreateScanningSessionRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID        uuid.UUID `json:"user_id"`
	SessionType   string    `json:"session_type"` // inventory, picking, receiving, counting
	LocationID    *uuid.UUID `json:"location_id,omitempty"`
	Reference     *string   `json:"reference,omitempty"`
	DeviceID      *string   `json:"device_id,omitempty"`
	Metadata      *string   `json:"metadata,omitempty"`
}

// AddScanToSessionRequest represents a request to add a scan to a session
type AddScanToSessionRequest struct {
	SessionID    uuid.UUID  `json:"session_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID       uuid.UUID  `json:"user_id"`
	Barcode      string     `json:"barcode"`
	Quantity     *float64   `json:"quantity,omitempty"`
	LocationID   *uuid.UUID `json:"location_id,omitempty"`
	DeviceInfo   *string    `json:"device_info,omitempty"`
}

// CompleteSessionRequest represents a request to complete a scanning session
type CompleteSessionRequest struct {
	SessionID    uuid.UUID `json:"session_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
}

// BarcodeGenerationRequest represents a request to generate barcodes
type BarcodeGenerationRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	EntityType    string    `json:"entity_type"` // product, location, lot, package
	EntityID      uuid.UUID `json:"entity_id"`
	Prefix        *string   `json:"prefix,omitempty"`
}

// BarcodeGenerationResponse represents a barcode generation response
type BarcodeGenerationResponse struct {
	Success     bool   `json:"success"`
	Barcode     string `json:"barcode"`
	EntityType  string `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
}
