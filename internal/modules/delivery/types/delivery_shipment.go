package types

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentStatus string

const (
	ShipmentStatusDraft     ShipmentStatus = "draft"
	ShipmentStatusScheduled ShipmentStatus = "scheduled"
	ShipmentStatusInTransit ShipmentStatus = "in_transit"
	ShipmentStatusDelivered ShipmentStatus = "delivered"
	ShipmentStatusFailed    ShipmentStatus = "failed"
	ShipmentStatusCancelled ShipmentStatus = "cancelled"
)

type ShipmentType string

const (
	ShipmentTypeOutbound ShipmentType = "outbound"
	ShipmentTypeInbound  ShipmentType = "inbound"
	ShipmentTypeInternal ShipmentType = "internal"
)

type DeliveryShipment struct {
	ID                  uuid.UUID      `json:"id" db:"id"`
	OrganizationID      uuid.UUID      `json:"organization_id" db:"organization_id"`
	CompanyID           *uuid.UUID     `json:"company_id" db:"company_id"`
	PickingID           uuid.UUID      `json:"picking_id" db:"picking_id"`
	RouteID             *uuid.UUID     `json:"route_id" db:"route_id"`
	AssignmentID        *uuid.UUID     `json:"assignment_id" db:"assignment_id"`
	TrackingNumber      string         `json:"tracking_number" db:"tracking_number"`
	CarrierName         string         `json:"carrier_name" db:"carrier_name"`
	CarrierCode         string         `json:"carrier_code" db:"carrier_code"`
	CarrierServiceLevel string        `json:"carrier_service_level" db:"carrier_service_level"`
	ShipmentType        ShipmentType   `json:"shipment_type" db:"shipment_type"`
	Status              ShipmentStatus `json:"status" db:"status"`
	RequiresSignature   bool           `json:"requires_signature" db:"requires_signature"`
	EstimatedDepartureAt *time.Time    `json:"estimated_departure_at" db:"estimated_departure_at"`
	EstimatedArrivalAt   *time.Time    `json:"estimated_arrival_at" db:"estimated_arrival_at"`
	DepartedAt          *time.Time    `json:"departed_at" db:"departed_at"`
	ArrivedAt           *time.Time    `json:"arrived_at" db:"arrived_at"`
	LastEventAt         *time.Time    `json:"last_event_at" db:"last_event_at"`
	LastLatitude        *float64       `json:"last_latitude" db:"last_latitude"`
	LastLongitude       *float64       `json:"last_longitude" db:"last_longitude"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy           *uuid.UUID     `json:"created_by" db:"created_by"`
	UpdatedBy           *uuid.UUID     `json:"updated_by" db:"updated_by"`
	DeletedAt           *time.Time     `json:"deleted_at" db:"deleted_at"`
}
