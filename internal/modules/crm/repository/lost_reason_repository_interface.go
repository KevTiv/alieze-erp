package repository

import (
	"context"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// LostReasonRepository defines the interface for lost reason data access
type LostReasonRepository interface {
	Create(ctx context.Context, reason types.LostReason) (*types.LostReason, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.LostReason, error)
	FindAll(ctx context.Context, filter types.LostReasonFilter) ([]types.LostReason, error)
	Update(ctx context.Context, reason types.LostReason) (*types.LostReason, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
