package types

import (
	"time"

	"github.com/google/uuid"
)

type VehicleType string

const (
	VehicleTypeTruck  VehicleType = "truck"
	VehicleTypeVan    VehicleType = "van"
	VehicleTypeBike   VehicleType = "bike"
	VehicleTypeCar    VehicleType = "car"
	VehicleTypeDrone  VehicleType = "drone"
	VehicleTypeOther  VehicleType = "other"
)

type DeliveryVehicle struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	OrganizationID    uuid.UUID      `json:"organization_id" db:"organization_id"`
	Name              string         `json:"name" db:"name"`
	RegistrationNumber string        `json:"registration_number" db:"registration_number"`
	VehicleIdentifier string        `json:"vehicle_identifier" db:"vehicle_identifier"`
	VehicleType       VehicleType    `json:"vehicle_type" db:"vehicle_type"`
	Capacity          float64        `json:"capacity" db:"capacity"`
	CapacityUOMID     *uuid.UUID     `json:"capacity_uom_id" db:"capacity_uom_id"`
	Active            bool           `json:"active" db:"active"`
	LastServiceAt     *time.Time     `json:"last_service_at" db:"last_service_at"`
	ServiceIntervalDays *int         `json:"service_interval_days" db:"service_interval_days"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy         *uuid.UUID     `json:"created_by" db:"created_by"`
	UpdatedBy         *uuid.UUID     `json:"updated_by" db:"updated_by"`
	DeletedAt         *time.Time     `json:"deleted_at" db:"deleted_at"`
}
