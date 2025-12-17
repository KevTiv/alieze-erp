package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
)

type TaxService struct {
	repo repository.TaxRepository
}

func NewTaxService(repo repository.TaxRepository) *TaxService {
	return &TaxService{
		repo: repo,
	}
}

func (s *TaxService) CreateTax(ctx context.Context, tax types.Tax) (*types.Tax, error) {
	// Validate the tax
	if err := s.validateTax(tax); err != nil {
		return nil, fmt.Errorf("invalid tax: %w", err)
	}

	// Set defaults
	if tax.Sequence == 0 {
		tax.Sequence = 10
	}
	if tax.Active == false && tax.ID == uuid.Nil {
		// Default to active for new taxes
		tax.Active = true
	}

	// Create the tax
	createdTax, err := s.repo.Create(ctx, tax)
	if err != nil {
		return nil, fmt.Errorf("failed to create tax: %w", err)
	}

	return createdTax, nil
}

func (s *TaxService) GetTax(ctx context.Context, id uuid.UUID) (*types.Tax, error) {
	tax, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax: %w", err)
	}
	if tax == nil {
		return nil, fmt.Errorf("tax not found")
	}

	return tax, nil
}

func (s *TaxService) ListTaxes(ctx context.Context, filters repository.TaxFilter) ([]types.Tax, error) {
	taxes, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list taxes: %w", err)
	}

	return taxes, nil
}

func (s *TaxService) UpdateTax(ctx context.Context, tax types.Tax) (*types.Tax, error) {
	// Validate the tax
	if err := s.validateTax(tax); err != nil {
		return nil, fmt.Errorf("invalid tax: %w", err)
	}

	// Check if tax exists
	existing, err := s.repo.FindByID(ctx, tax.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing tax: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("tax not found")
	}

	// Update the tax
	updatedTax, err := s.repo.Update(ctx, tax)
	if err != nil {
		return nil, fmt.Errorf("failed to update tax: %w", err)
	}

	return updatedTax, nil
}

func (s *TaxService) DeleteTax(ctx context.Context, id uuid.UUID) error {
	// Check if tax exists
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check existing tax: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("tax not found")
	}

	// Delete the tax (soft delete - sets active = false)
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete tax: %w", err)
	}

	return nil
}

func (s *TaxService) GetTaxesByType(ctx context.Context, organizationID uuid.UUID, typeTaxUse string) ([]types.Tax, error) {
	taxes, err := s.repo.FindByType(ctx, organizationID, typeTaxUse)
	if err != nil {
		return nil, fmt.Errorf("failed to get taxes by type: %w", err)
	}

	return taxes, nil
}

func (s *TaxService) validateTax(tax types.Tax) error {
	if tax.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if tax.Name == "" {
		return fmt.Errorf("name is required")
	}

	if tax.AmountType == "" {
		return fmt.Errorf("amount type is required")
	}

	// Validate amount type
	validAmountTypes := map[string]bool{
		"percent":  true,
		"fixed":    true,
		"division": true,
		"group":    true,
	}

	if !validAmountTypes[tax.AmountType] {
		return fmt.Errorf("invalid amount type: %s (must be one of: percent, fixed, division, group)", tax.AmountType)
	}

	// Validate type_tax_use if provided
	if tax.TypeTaxUse != nil {
		validTypeTaxUse := map[string]bool{
			"sale":     true,
			"purchase": true,
			"none":     true,
		}

		if !validTypeTaxUse[*tax.TypeTaxUse] {
			return fmt.Errorf("invalid type_tax_use: %s (must be one of: sale, purchase, none)", *tax.TypeTaxUse)
		}
	}

	// Validate amount for percent type
	if tax.AmountType == "percent" && (tax.Amount < 0 || tax.Amount > 100) {
		return fmt.Errorf("percent tax amount must be between 0 and 100")
	}

	return nil
}
