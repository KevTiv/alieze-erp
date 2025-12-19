package types

import (
	"time"

	"github.com/google/uuid"
)

type StopStatus string

const (
	StopStatusPlanned   StopStatus = "planned"
	StopStatusEnRoute   StopStatus = "en_route"
	StopStatusArrived   StopStatus = "arrived"
	StopStatusCompleted StopStatus = "completed"
	StopStatusSkipped   StopStatus = "skipped"
	StopStatusFailed    StopStatus = "failed"
)

type DeliveryRouteStop struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	OrganizationID    uuid.UUID         `json:"organization_id" db:"organization_id"`
	RouteID           uuid.UUID         `json:"route_id" db:"route_id"`
	AssignmentID      *uuid.UUID        `json:"assignment_id" db:"assignment_id"`
	ShipmentID        *uuid.UUID        `json:"shipment_id" db:"shipment_id"`
	StopSequence      int               `json:"stop_sequence" db:"stop_sequence"`
	ContactID         *uuid.UUID        `json:"contact_id" db:"contact_id"`
	LocationID        *uuid.UUID        `json:"location_id" db:"location_id"`
	Address           map[string]interface{} `json:"address" db:"address"`
	PlannedArrivalAt  *time.Time        `json:"planned_arrival_at" db:"planned_arrival_at"`
	PlannedDepartureAt *time.Time       `json:"planned_departure_at" db:"planned_departure_at"`
	ActualArrivalAt   *time.Time        `json:"actual_arrival_at" db:"actual_arrival_at"`
	ActualDepartureAt *time.Time        `json:"actual_departure_at" db:"actual_departure_at"`
	Status            StopStatus        `json:"status" db:"status"`
	Notes             string            `json:"notes" db:"notes"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
	CreatedBy         *uuid.UUID        `json:"created_by" db:"created_by"`
	UpdatedBy         *uuid.UUID        `json:"updated_by" db:"updated_by"`
}
