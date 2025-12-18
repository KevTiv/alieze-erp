package service

import (
	"context"
	"errors"

	"alieze-erp/internal/modules/common/repository"
	"alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// CountryService handles business logic for countries
type CountryService struct {
	repository *repository.CountryRepository
}

func NewCountryService(repository *repository.CountryRepository) *CountryService {
	return &CountryService{repository: repository}
}

func (s *CountryService) Create(ctx context.Context, req types.CountryCreateRequest) (*types.Country, error) {
	// Validate required fields
	if req.Code == "" {
		return nil, errors.New("code is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// Check if country with same code already exists
	existing, err := s.repository.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("country with this code already exists")
	}

	// Create country entity
	country := types.Country{
		Code:        req.Code,
		Name:        req.Name,
		PhoneCode:   req.PhoneCode,
		CurrencyID:  req.CurrencyID,
		AddressFormat: req.AddressFormat,
	}

	return s.repository.Create(ctx, country)
}

func (s *CountryService) GetByID(ctx context.Context, id uuid.UUID) (*types.Country, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *CountryService) GetByCode(ctx context.Context, code string) (*types.Country, error) {
	return s.repository.GetByCode(ctx, code)
}

func (s *CountryService) List(ctx context.Context, filter types.CountryFilter) ([]types.Country, error) {
	return s.repository.List(ctx, filter)
}

func (s *CountryService) Update(ctx context.Context, id uuid.UUID, req types.CountryUpdateRequest) (*types.Country, error) {
	// Check if country exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("country not found")
	}

	// If code is being updated, check if new code already exists
	if req.Code != nil && *req.Code != existing.Code {
		existingWithCode, err := s.repository.GetByCode(ctx, *req.Code)
		if err != nil {
			return nil, err
		}
		if existingWithCode != nil {
			return nil, errors.New("country with this code already exists")
		}
	}

	return s.repository.Update(ctx, id, req)
}

func (s *CountryService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if country exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("country not found")
	}

	// Check if country has states - if so, prevent deletion
	// This would require a query to check for dependent states
	// For now, we'll allow deletion and let the database handle referential integrity

	return s.repository.Delete(ctx, id)
}

// GetCountryWithStates returns a country with its states
type CountryWithStates struct {
	Country *types.Country
	States  []types.State
}

func (s *CountryService) GetCountryWithStates(ctx context.Context, countryID uuid.UUID) (*CountryWithStates, error) {
	country, err := s.repository.GetByID(ctx, countryID)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return nil, errors.New("country not found")
	}

	// Get states for this country
	stateRepo := repository.NewStateRepository(s.repository.db)
	states, err := stateRepo.ListByCountry(ctx, countryID)
	if err != nil {
		return nil, err
	}

	return &CountryWithStates{
		Country: country,
		States:  states,
	}, nil
}
