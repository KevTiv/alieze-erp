package repository

import (
	"context"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// ContactRepo defines the interface for contact data access
type ContactRepo interface {
	Create(ctx context.Context, contact domain.Contact) (*domain.Contact, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error)
	FindAll(ctx context.Context, filter domain.ContactFilter) ([]domain.Contact, error)
	Update(ctx context.Context, contact domain.Contact) (*domain.Contact, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter domain.ContactFilter) (int, error)
}
