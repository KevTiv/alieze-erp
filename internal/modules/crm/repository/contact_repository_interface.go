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

	// Relationship methods
	CreateRelationship(ctx context.Context, relationship *types.ContactRelationship) error
	FindRelationships(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, relationshipType string, limit int) ([]*types.ContactRelationship, error)
	ContactExists(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) (bool, error)

	// Segment and tag methods
	AddContactToSegments(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, segmentIDs []string) error
	AddContactTags(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, tags []string) error
}
