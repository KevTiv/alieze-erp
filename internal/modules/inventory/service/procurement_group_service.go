package service

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// ProcurementGroupService handles business logic for procurement groups
type ProcurementGroupService struct {
	repo   *repository.ProcurementGroupRepository
}

// NewProcurementGroupService creates a new ProcurementGroupService
func NewProcurementGroupService(repo *repository.ProcurementGroupRepository) *ProcurementGroupService {
	return &ProcurementGroupService{
		repo: repo,
	}
}

// Create creates a new procurement group
func (s *ProcurementGroupService) Create(ctx context.Context, orgID uuid.UUID, req types.ProcurementGroupCreateRequest) (*types.ProcurementGroup, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a procurement group by ID
func (s *ProcurementGroupService) GetByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*types.ProcurementGroup, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all procurement groups for an organization
func (s *ProcurementGroupService) List(ctx context.Context, orgID uuid.UUID) ([]types.ProcurementGroup, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a procurement group
func (s *ProcurementGroupService) Update(ctx context.Context, id uuid.UUID, req types.ProcurementGroupUpdateRequest) (*types.ProcurementGroup, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a procurement group
func (s *ProcurementGroupService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
