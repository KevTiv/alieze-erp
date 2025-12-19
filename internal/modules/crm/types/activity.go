package types

import (
	"time"

	"github.com/google/uuid"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeCall    ActivityType = "call"
	ActivityTypeMeeting ActivityType = "meeting"
	ActivityTypeEmail   ActivityType = "email"
	ActivityTypeTodo    ActivityType = "todo"
	ActivityTypeNote    ActivityType = "note"
)

// ActivityState represents the state of an activity
type ActivityState string

const (
	ActivityStatePlanned   ActivityState = "planned"
	ActivityStateDone      ActivityState = "done"
	ActivityStateCancelled ActivityState = "cancelled"
)

// Activity represents a CRM activity
type Activity struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	OrganizationID uuid.UUID     `json:"organization_id" db:"organization_id"`
	ActivityType   ActivityType  `json:"activity_type" db:"activity_type"`
	Summary        string        `json:"summary" db:"summary"`
	Note           *string       `json:"note,omitempty" db:"note"`
	DateDeadline   *time.Time    `json:"date_deadline,omitempty" db:"date_deadline"`
	UserID         *uuid.UUID    `json:"user_id,omitempty" db:"user_id"`
	AssignedTo     *uuid.UUID    `json:"assigned_to,omitempty" db:"assigned_to"`
	ResModel       *string       `json:"res_model,omitempty" db:"res_model"`
	ResID          *uuid.UUID    `json:"res_id,omitempty" db:"res_id"`
	State          ActivityState `json:"state" db:"state"`
	DoneDate       *time.Time    `json:"done_date,omitempty" db:"done_date"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	CreatedBy      *uuid.UUID    `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy      *uuid.UUID    `json:"updated_by,omitempty" db:"updated_by"`
}

// ActivityFilter represents filtering criteria for activities
type ActivityFilter struct {
	OrganizationID uuid.UUID
	ActivityType   *ActivityType
	State          *ActivityState
	UserID         *uuid.UUID
	AssignedTo     *uuid.UUID
	ResModel       *string
	ResID          *uuid.UUID
	Limit          int
	Offset         int
}

// ActivityCreateRequest represents a request to create an activity
type ActivityCreateRequest struct {
	ActivityType ActivityType  `json:"activity_type"`
	Summary      string        `json:"summary"`
	Note         *string       `json:"note,omitempty"`
	DateDeadline *time.Time    `json:"date_deadline,omitempty"`
	UserID       *uuid.UUID    `json:"user_id,omitempty"`
	AssignedTo   *uuid.UUID    `json:"assigned_to,omitempty"`
	ResModel     *string       `json:"res_model,omitempty"`
	ResID        *uuid.UUID    `json:"res_id,omitempty"`
	State        ActivityState `json:"state"`
	DoneDate     *time.Time    `json:"done_date,omitempty"`
}

// ActivityUpdateRequest represents a request to update an activity
type ActivityUpdateRequest struct {
	ActivityType *ActivityType  `json:"activity_type,omitempty"`
	Summary      *string        `json:"summary,omitempty"`
	Note         *string        `json:"note,omitempty"`
	DateDeadline *time.Time     `json:"date_deadline,omitempty"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	AssignedTo   *uuid.UUID     `json:"assigned_to,omitempty"`
	ResModel     *string        `json:"res_model,omitempty"`
	ResID        *uuid.UUID     `json:"res_id,omitempty"`
	State        *ActivityState `json:"state,omitempty"`
	DoneDate     *time.Time     `json:"done_date,omitempty"`
}
