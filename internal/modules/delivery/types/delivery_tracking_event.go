package types

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryTrackingEvent struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	OrganizationID uuid.UUID      `json:"organization_id" db:"organization_id"`
	ShipmentID   uuid.UUID        `json:"shipment_id" db:"shipment_id"`
	StopID       *uuid.UUID       `json:"stop_id" db:"stop_id"`
	EventType    string           `json:"event_type" db:"event_type"`
	Status       string           `json:"status" db:"status"`
	EventTime    time.Time        `json:"event_time" db:"event_time"`
	Source       string           `json:"source" db:"source"`
	Message      string           `json:"message" db:"message"`
	RawPayload   map[string]interface{} `json:"raw_payload" db:"raw_payload"`
	Latitude     *float64        `json:"latitude" db:"latitude"`
	Longitude    *float64        `json:"longitude" db:"longitude"`
	Altitude     *float64        `json:"altitude" db:"altitude"`
	SpeedKPH     *float64        `json:"speed_kph" db:"speed_kph"`
	Heading      *float64        `json:"heading" db:"heading"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
	CreatedBy    *uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy    *uuid.UUID       `json:"updated_by" db:"updated_by"`
}
