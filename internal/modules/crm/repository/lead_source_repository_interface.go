package repository

import (
	"context"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// LeadSourceRepository defines the interface for lead source data access
type LeadSourceRepository interface {
	Create(ctx context.Context, source types.LeadSource) (*types.LeadSource, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.LeadSource, error)
	FindAll(ctx context.Context, filter types.LeadSourceFilter) ([]types.LeadSource, error)
	Update(ctx context.Context, source types.LeadSource) (*types.LeadSource, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
