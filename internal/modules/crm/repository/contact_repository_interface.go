package repository

import (
	"context"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// ContactRepo defines the interface for contact data access
type ContactRepo interface {
	Create(ctx context.Context, contact types.Contact) (*types.Contact, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.Contact, error)
	FindAll(ctx context.Context, filter types.ContactFilter) ([]types.Contact, error)
	Update(ctx context.Context, contact types.Contact) (*types.Contact, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter types.ContactFilter) (int, error)
}
