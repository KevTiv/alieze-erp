package repository

import (
	"context"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// SalesTeamRepository defines the interface for sales team data access
type SalesTeamRepository interface {
	Create(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error)
	FindAll(ctx context.Context, filter types.SalesTeamFilter) ([]types.SalesTeam, error)
	Update(ctx context.Context, team types.SalesTeam) (*types.SalesTeam, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByMember(ctx context.Context, memberID uuid.UUID) ([]types.SalesTeam, error)
}
