package repository

import (
	"context"

	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// ContactTagRepository defines the interface for contact tag data access
type ContactTagRepository interface {
	Create(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.ContactTag, error)
	FindAll(ctx context.Context, filter types.ContactTagFilter) ([]types.ContactTag, error)
	Update(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByContact(ctx context.Context, contactID uuid.UUID) ([]types.ContactTag, error)
}
