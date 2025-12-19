package types

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentStatus string

const (
	AssignmentStatusAssigned  AssignmentStatus = "assigned"
	AssignmentStatusAccepted  AssignmentStatus = "accepted"
	AssignmentStatusDeclined  AssignmentStatus = "declined"
	AssignmentStatusReleased  AssignmentStatus = "released"
	AssignmentStatusCompleted AssignmentStatus = "completed"
)

type DeliveryRouteAssignment struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	OrganizationID    uuid.UUID         `json:"organization_id" db:"organization_id"`
	RouteID           uuid.UUID         `json:"route_id" db:"route_id"`
	VehicleID         *uuid.UUID        `json:"vehicle_id" db:"vehicle_id"`
	DriverEmployeeID  *uuid.UUID        `json:"driver_employee_id" db:"driver_employee_id"`
	DriverContactID   *uuid.UUID        `json:"driver_contact_id" db:"driver_contact_id"`
	AssignmentStatus  AssignmentStatus  `json:"assignment_status" db:"assignment_status"`
	AssignedAt        time.Time         `json:"assigned_at" db:"assigned_at"`
	AcknowledgedAt    *time.Time        `json:"acknowledged_at" db:"acknowledged_at"`
	ReleasedAt        *time.Time        `json:"released_at" db:"released_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
	CreatedBy         *uuid.UUID        `json:"created_by" db:"created_by"`
	UpdatedBy         *uuid.UUID        `json:"updated_by" db:"updated_by"`
}
