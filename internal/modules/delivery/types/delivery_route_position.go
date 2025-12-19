package types

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryRoutePosition struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	OrganizationID uuid.UUID      `json:"organization_id" db:"organization_id"`
	RouteID      uuid.UUID        `json:"route_id" db:"route_id"`
	AssignmentID *uuid.UUID       `json:"assignment_id" db:"assignment_id"`
	VehicleID    *uuid.UUID       `json:"vehicle_id" db:"vehicle_id"`
	RecordedAt   time.Time        `json:"recorded_at" db:"recorded_at"`
	Latitude     float64          `json:"latitude" db:"latitude"`
	Longitude    float64          `json:"longitude" db:"longitude"`
	Altitude     *float64        `json:"altitude" db:"altitude"`
	SpeedKPH     *float64        `json:"speed_kph" db:"speed_kph"`
	Heading      *float64        `json:"heading" db:"heading"`
	Source       string           `json:"source" db:"source"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
