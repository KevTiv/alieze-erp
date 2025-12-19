package types

import (
	"time"

	"github.com/google/uuid"
)

type TransportMode string

const (
	TransportModeRoad TransportMode = "road"
	TransportModeAir  TransportMode = "air"
	TransportModeSea  TransportMode = "sea"
	TransportModeRail TransportMode = "rail"
	TransportModeOther TransportMode = "other"
)

type RouteStatus string

const (
	RouteStatusDraft      RouteStatus = "draft"
	RouteStatusScheduled  RouteStatus = "scheduled"
	RouteStatusInProgress RouteStatus = "in_progress"
	RouteStatusCompleted  RouteStatus = "completed"
	RouteStatusCancelled  RouteStatus = "cancelled"
)

type DeliveryRoute struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	OrganizationID    uuid.UUID      `json:"organization_id" db:"organization_id"`
	CompanyID         *uuid.UUID     `json:"company_id" db:"company_id"`
	WarehouseID       *uuid.UUID     `json:"warehouse_id" db:"warehouse_id"`
	Name              string         `json:"name" db:"name"`
	RouteCode         string         `json:"route_code" db:"route_code"`
	TransportMode     TransportMode  `json:"transport_mode" db:"transport_mode"`
	Status            RouteStatus    `json:"status" db:"status"`
	ScheduledStartAt  *time.Time     `json:"scheduled_start_at" db:"scheduled_start_at"`
	ScheduledEndAt    *time.Time     `json:"scheduled_end_at" db:"scheduled_end_at"`
	ActualStartAt     *time.Time     `json:"actual_start_at" db:"actual_start_at"`
	ActualEndAt       *time.Time     `json:"actual_end_at" db:"actual_end_at"`
	OriginLocationID  *uuid.UUID     `json:"origin_location_id" db:"origin_location_id"`
	DestinationLocationID *uuid.UUID `json:"destination_location_id" db:"destination_location_id"`
	Notes             string         `json:"notes" db:"notes"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy         *uuid.UUID     `json:"created_by" db:"created_by"`
	UpdatedBy         *uuid.UUID     `json:"updated_by" db:"updated_by"`
	DeletedAt         *time.Time     `json:"deleted_at" db:"deleted_at"`
}
