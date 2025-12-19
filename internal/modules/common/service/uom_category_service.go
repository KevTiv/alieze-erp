package service

import (
	"context"
	"errors"

	"github.com/KevTiv/alieze-erp/internal/modules/common/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// UOMCategoryService handles business logic for UOM categories
type UOMCategoryService struct {
	repository *repository.UOMCategoryRepository
}

func NewUOMCategoryService(repository *repository.UOMCategoryRepository) *UOMCategoryService {
	return &UOMCategoryService{repository: repository}
}

func (s *UOMCategoryService) Create(ctx context.Context, req types.UOMCategoryCreateRequest) (*types.UOMCategory, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// Create UOM category entity
	category := types.UOMCategory{
		Name: req.Name,
	}

	return s.repository.Create(ctx, category)
}

func (s *UOMCategoryService) GetByID(ctx context.Context, id uuid.UUID) (*types.UOMCategory, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *UOMCategoryService) List(ctx context.Context, filter types.UOMCategoryFilter) ([]types.UOMCategory, error) {
	return s.repository.List(ctx, filter)
}

func (s *UOMCategoryService) Update(ctx context.Context, id uuid.UUID, req types.UOMCategoryUpdateRequest) (*types.UOMCategory, error) {
	// Check if UOM category exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("UOM category not found")
	}

	return s.repository.Update(ctx, id, req)
}

func (s *UOMCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if UOM category exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("UOM category not found")
	}

	// Check if category has units - if so, prevent deletion
	// This would require a query to check for dependent units
	// For now, we'll allow deletion and let the database handle referential integrity

	return s.repository.Delete(ctx, id)
}

// GetUOMCategoryWithUnits returns a UOM category with its units
type UOMCategoryWithUnits struct {
	Category *types.UOMCategory
	Units    []types.UOMUnit
}

func (s *UOMCategoryService) GetUOMCategoryWithUnits(ctx context.Context, categoryID uuid.UUID) (*UOMCategoryWithUnits, error) {
	category, err := s.repository.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("UOM category not found")
	}

	// Get units for this category
	unitRepo := repository.NewUOMUnitRepository(s.repository.db)
	units, err := unitRepo.ListByCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	return &UOMCategoryWithUnits{
		Category: category,
		Units:    units,
	}, nil
}
