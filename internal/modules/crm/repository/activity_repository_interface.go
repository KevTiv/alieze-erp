package repository

import (
	"context"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// ActivityRepository defines the interface for activity data access
type ActivityRepository interface {
	Create(ctx context.Context, activity types.Activity) (*types.Activity, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.Activity, error)
	FindAll(ctx context.Context, filter types.ActivityFilter) ([]types.Activity, error)
	Update(ctx context.Context, activity types.Activity) (*types.Activity, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByContact(ctx context.Context, contactID uuid.UUID) ([]types.Activity, error)
	FindByLead(ctx context.Context, leadID uuid.UUID) ([]types.Activity, error)
}
