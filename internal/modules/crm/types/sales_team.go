package types

import (
	"time"

	"github.com/google/uuid"
)

// SalesTeam represents a sales team
type SalesTeam struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CompanyID      *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name          string     `json:"name" db:"name"`
	Code          *string    `json:"code,omitempty" db:"code"`
	TeamLeaderID   *uuid.UUID `json:"team_leader_id,omitempty" db:"team_leader_id"`
	MemberIDs     []uuid.UUID `json:"member_ids" db:"member_ids"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SalesTeamFilter represents filtering criteria for sales teams
type SalesTeamFilter struct {
	OrganizationID uuid.UUID
	CompanyID      *uuid.UUID
	Name           *string
	Code           *string
	TeamLeaderID   *uuid.UUID
	IsActive       *bool
	Limit          int
	Offset         int
}

// SalesTeamCreateRequest represents a request to create a sales team
type SalesTeamCreateRequest struct {
	CompanyID    *uuid.UUID `json:"company_id,omitempty"`
	Name         string     `json:"name"`
	Code         *string    `json:"code,omitempty"`
	TeamLeaderID *uuid.UUID `json:"team_leader_id,omitempty"`
	MemberIDs    []uuid.UUID `json:"member_ids"`
	IsActive     bool       `json:"is_active"`
}

// SalesTeamUpdateRequest represents a request to update a sales team
type SalesTeamUpdateRequest struct {
	CompanyID    *uuid.UUID `json:"company_id,omitempty"`
	Name         *string    `json:"name,omitempty"`
	Code         *string    `json:"code,omitempty"`
	TeamLeaderID *uuid.UUID `json:"team_leader_id,omitempty"`
	MemberIDs    *[]uuid.UUID `json:"member_ids,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}
