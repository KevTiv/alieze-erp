package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"alieze-erp/internal/modules/common/repository"
	"alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// CurrencyService handles business logic for currencies
type CurrencyService struct {
	repository *repository.CurrencyRepository
}

func NewCurrencyService(repository *repository.CurrencyRepository) *CurrencyService {
	return &CurrencyService{repository: repository}
}

func (s *CurrencyService) Create(ctx context.Context, req types.CurrencyCreateRequest) (*types.Currency, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Code == "" {
		return nil, errors.New("code is required")
	}
	if req.Symbol == "" {
		return nil, errors.New("symbol is required")
	}

	// Check if currency with same code already exists
	existing, err := s.repository.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("currency with this code already exists")
	}

	// Create currency entity
	currency := types.Currency{
		Name:          req.Name,
		Symbol:        req.Symbol,
		Code:          req.Code,
		Rounding:      req.Rounding,
		DecimalPlaces: req.DecimalPlaces,
		Position:      req.Position,
		Active:        req.Active,
	}

	return s.repository.Create(ctx, currency)
}

func (s *CurrencyService) GetByID(ctx context.Context, id uuid.UUID) (*types.Currency, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *CurrencyService) GetByCode(ctx context.Context, code string) (*types.Currency, error) {
	return s.repository.GetByCode(ctx, code)
}

func (s *CurrencyService) List(ctx context.Context, filter types.CurrencyFilter) ([]types.Currency, error) {
	return s.repository.List(ctx, filter)
}

func (s *CurrencyService) Update(ctx context.Context, id uuid.UUID, req types.CurrencyUpdateRequest) (*types.Currency, error) {
	// Check if currency exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("currency not found")
	}

	// If code is being updated, check if new code already exists
	if req.Code != nil && *req.Code != existing.Code {
		existingWithCode, err := s.repository.GetByCode(ctx, *req.Code)
		if err != nil {
			return nil, err
		}
		if existingWithCode != nil {
			return nil, errors.New("currency with this code already exists")
		}
	}

	return s.repository.Update(ctx, id, req)
}

func (s *CurrencyService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if currency exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("currency not found")
	}

	// Check if currency is referenced by other entities (e.g., countries, companies)
	// This would require additional database queries to check foreign key constraints
	// For now, we'll allow deletion and let the database handle referential integrity

	return s.repository.Delete(ctx, id)
}

// GetDefaultCurrency returns the default currency (USD)
func (s *CurrencyService) GetDefaultCurrency(ctx context.Context) (*types.Currency, error) {
	return s.repository.GetByCode(ctx, "USD")
}

// FormatAmount formats an amount according to currency rules
func (s *CurrencyService) FormatAmount(ctx context.Context, currencyCode string, amount float64) (string, error) {
	currency, err := s.repository.GetByCode(ctx, currencyCode)
	if err != nil {
		return "", err
	}
	if currency == nil {
		return "", errors.New("currency not found")
	}

	// Simple formatting - in a real implementation, this would be more sophisticated
	var formatted string
	if currency.Position == types.CurrencyPositionBefore {
		formatted = currency.Symbol + " " + fmt.Sprintf("%."+strconv.Itoa(currency.DecimalPlaces)+"f", amount)
	} else {
		formatted = fmt.Sprintf("%."+strconv.Itoa(currency.DecimalPlaces)+"f", amount) + " " + currency.Symbol
	}

	return formatted, nil
}
