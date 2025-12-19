package service

import (
	"context"
	"errors"

	"github.com/KevTiv/alieze-erp/internal/modules/common/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// UOMUnitService handles business logic for UOM units
type UOMUnitService struct {
	repository *repository.UOMUnitRepository
}

func NewUOMUnitService(repository *repository.UOMUnitRepository) *UOMUnitService {
	return &UOMUnitService{repository: repository}
}

func (s *UOMUnitService) Create(ctx context.Context, req types.UOMUnitCreateRequest) (*types.UOMUnit, error) {
	// Validate required fields
	if req.CategoryID == uuid.Nil {
		return nil, errors.New("category_id is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// Check if category exists
	categoryRepo := repository.NewUOMCategoryRepository(s.repository.db)
	category, err := categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("UOM category not found")
	}

	// Create UOM unit entity
	unit := types.UOMUnit{
		CategoryID: req.CategoryID,
		Name:       req.Name,
		UOMType:    req.UOMType,
		Factor:     req.Factor,
		FactorInv:  req.FactorInv,
		Rounding:   req.Rounding,
		Active:     req.Active,
	}

	return s.repository.Create(ctx, unit)
}

func (s *UOMUnitService) GetByID(ctx context.Context, id uuid.UUID) (*types.UOMUnit, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *UOMUnitService) List(ctx context.Context, filter types.UOMUnitFilter) ([]types.UOMUnit, error) {
	return s.repository.List(ctx, filter)
}

func (s *UOMUnitService) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]types.UOMUnit, error) {
	return s.repository.ListByCategory(ctx, categoryID)
}

func (s *UOMUnitService) Update(ctx context.Context, id uuid.UUID, req types.UOMUnitUpdateRequest) (*types.UOMUnit, error) {
	// Check if UOM unit exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("UOM unit not found")
	}

	// If category is being updated, check if new category exists
	if req.CategoryID != nil && *req.CategoryID != existing.CategoryID {
		categoryRepo := repository.NewUOMCategoryRepository(s.repository.db)
		category, err := categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, errors.New("new UOM category not found")
		}
	}

	return s.repository.Update(ctx, id, req)
}

func (s *UOMUnitService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if UOM unit exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil, errors.New("UOM unit not found")
	}

	return s.repository.Delete(ctx, id)
}

// ConvertUnit converts a quantity from one unit to another within the same category
func (s *UOMUnitService) ConvertUnit(ctx context.Context, fromUnitID, toUnitID uuid.UUID, quantity float64) (float64, error) {
	fromUnit, err := s.repository.GetByID(ctx, fromUnitID)
	if err != nil {
		return 0, err
	}
	if fromUnit == nil {
		return 0, errors.New("from unit not found")
	}

	toUnit, err := s.repository.GetByID(ctx, toUnitID)
	if err != nil {
		return 0, err
	}
	if toUnit == nil {
		return 0, errors.New("to unit not found")
	}

	// Check if units are in the same category
	if fromUnit.CategoryID != toUnit.CategoryID {
		return 0, errors.New("units must be in the same category for conversion")
	}

	// Convert using the factor
	// If converting from reference unit to another unit: quantity * factor
	// If converting from another unit to reference unit: quantity * factor_inv
	// For general conversion: quantity * (fromUnit.FactorInv * toUnit.Factor)
	convertedQuantity := quantity * (fromUnit.FactorInv * toUnit.Factor)

	// Apply rounding
	convertedQuantity = float64(int(convertedQuantity/toUnit.Rounding+0.5)) * toUnit.Rounding

	return convertedQuantity, nil
}
