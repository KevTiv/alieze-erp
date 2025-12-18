package repository

import (
	"context"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// LeadStageRepository defines the interface for lead stage data access
type LeadStageRepository interface {
	Create(ctx context.Context, stage types.LeadStage) (*types.LeadStage, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.LeadStage, error)
	FindAll(ctx context.Context, filter types.LeadStageFilter) ([]types.LeadStage, error)
	Update(ctx context.Context, stage types.LeadStage) (*types.LeadStage, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
