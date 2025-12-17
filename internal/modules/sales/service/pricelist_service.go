package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/sales/types"
	"alieze-erp/internal/modules/sales/repository"

	"github.com/google/uuid"
)

type PricelistService struct {
	repo repository.PricelistRepository
}

func NewPricelistService(repo repository.PricelistRepository) *PricelistService {
	return &PricelistService{
		repo: repo,
	}
}

func (s *PricelistService) CreatePricelist(ctx context.Context, pricelist types.Pricelist) (*types.Pricelist, error) {
	// Validate the pricelist
	if err := s.validatePricelist(pricelist); err != nil {
		return nil, fmt.Errorf("invalid pricelist: %w", err)
	}

	// Set default values
	if pricelist.ID == uuid.Nil {
		pricelist.ID = uuid.New()
	}
	if pricelist.IsActive {
		// Check if there's already an active pricelist with the same name
		existing, err := s.repo.FindActiveByCompanyID(ctx, pricelist.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing pricelists: %w", err)
		}
		for _, p := range existing {
			if p.Name == pricelist.Name && p.ID != pricelist.ID {
				return nil, fmt.Errorf("active pricelist with name '%s' already exists", pricelist.Name)
			}
		}
	}

	// Create the pricelist
	createdPricelist, err := s.repo.Create(ctx, pricelist)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricelist: %w", err)
	}

	return createdPricelist, nil
}

func (s *PricelistService) GetPricelist(ctx context.Context, id uuid.UUID) (*types.Pricelist, error) {
	pricelist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricelist: %w", err)
	}
	if pricelist == nil {
		return nil, nil
	}

	return pricelist, nil
}

func (s *PricelistService) ListPricelists(ctx context.Context, organizationID uuid.UUID) ([]types.Pricelist, error) {
	pricelists, err := s.repo.FindAll(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricelists: %w", err)
	}

	return pricelists, nil
}

func (s *PricelistService) ListPricelistsByCompany(ctx context.Context, companyID uuid.UUID) ([]types.Pricelist, error) {
	pricelists, err := s.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricelists by company: %w", err)
	}

	return pricelists, nil
}

func (s *PricelistService) ListActivePricelistsByCompany(ctx context.Context, companyID uuid.UUID) ([]types.Pricelist, error) {
	pricelists, err := s.repo.FindActiveByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active pricelists by company: %w", err)
	}

	return pricelists, nil
}

func (s *PricelistService) UpdatePricelist(ctx context.Context, pricelist types.Pricelist) (*types.Pricelist, error) {
	// Validate the pricelist
	if err := s.validatePricelist(pricelist); err != nil {
		return nil, fmt.Errorf("invalid pricelist: %w", err)
	}

	// Check if pricelist exists
	existing, err := s.repo.FindByID(ctx, pricelist.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pricelist existence: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("pricelist not found")
	}

	// If activating, check for name conflicts
	if pricelist.IsActive && !existing.IsActive {
		activePricelists, err := s.repo.FindActiveByCompanyID(ctx, pricelist.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("failed to check active pricelists: %w", err)
		}
		for _, p := range activePricelists {
			if p.Name == pricelist.Name && p.ID != pricelist.ID {
				return nil, fmt.Errorf("active pricelist with name '%s' already exists", pricelist.Name)
			}
		}
	}

	// Update the pricelist
	updatedPricelist, err := s.repo.Update(ctx, pricelist)
	if err != nil {
		return nil, fmt.Errorf("failed to update pricelist: %w", err)
	}

	return updatedPricelist, nil
}

func (s *PricelistService) DeletePricelist(ctx context.Context, id uuid.UUID) error {
	// Check if pricelist exists
	pricelist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check pricelist existence: %w", err)
	}
	if pricelist == nil {
		return nil
	}

	// Prevent deletion of active pricelists
	if pricelist.IsActive {
		return fmt.Errorf("cannot delete active pricelists")
	}

	// Delete the pricelist
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete pricelist: %w", err)
	}

	return nil
}

func (s *PricelistService) validatePricelist(pricelist types.Pricelist) error {
	if pricelist.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if pricelist.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	if pricelist.Name == "" {
		return fmt.Errorf("name is required")
	}

	if pricelist.CurrencyID == uuid.Nil {
		return fmt.Errorf("currency ID is required")
	}

	// Validate items
	for _, item := range pricelist.Items {
		if item.ProductID == uuid.Nil {
			return fmt.Errorf("product ID is required for all items")
		}
		if item.MinQuantity < 0 {
			return fmt.Errorf("minimum quantity cannot be negative")
		}
		if item.FixedPrice != nil && *item.FixedPrice < 0 {
			return fmt.Errorf("fixed price cannot be negative")
		}
		if item.Discount != nil && (*item.Discount < 0 || *item.Discount > 100) {
			return fmt.Errorf("discount must be between 0 and 100")
		}
	}

	return nil
}
