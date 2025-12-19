package repository

import (
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/pkg/crm/base"
	"context"
	"github.com/google/uuid"
)

// ContactRepositoryAdapter implements the base.Repository interface for Contact entities
type ContactRepositoryAdapter struct {
	repo types.ContactRepository
}

// NewContactRepositoryAdapter creates a new adapter for ContactRepository
func NewContactRepositoryAdapter(repo types.ContactRepository) *ContactRepositoryAdapter {
	return &ContactRepositoryAdapter{repo: repo}
}

// Create implements base.Repository.Create
func (a *ContactRepositoryAdapter) Create(ctx context.Context, contact types.Contact) (*types.Contact, error) {
	return a.repo.Create(ctx, contact)
}

// FindByID implements base.Repository.FindByID
func (a *ContactRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
	return a.repo.FindByID(ctx, id)
}

// FindAll implements base.Repository.FindAll
func (a *ContactRepositoryAdapter) FindAll(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
	contacts, err := a.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Convert []Contact to []*Contact
	result := make([]*types.Contact, len(contacts))
	for i := range contacts {
		result[i] = &contacts[i]
	}

	return result, nil
}

// Update implements base.Repository.Update
func (a *ContactRepositoryAdapter) Update(ctx context.Context, contact types.Contact) (*types.Contact, error) {
	return a.repo.Update(ctx, contact)
}

// Delete implements base.Repository.Delete
func (a *ContactRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

// Count implements base.Repository.Count
func (a *ContactRepositoryAdapter) Count(ctx context.Context, filter types.ContactFilter) (int, error) {
	return a.repo.Count(ctx, filter)
}

// Compile-time interface compliance check
var _ base.Repository[types.Contact, types.ContactFilter] = (*ContactRepositoryAdapter)(nil)
