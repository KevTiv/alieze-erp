package service

import (
	"context"
	"errors"

	"alieze-erp/internal/modules/common/repository"
	"alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// StateService handles business logic for states
type StateService struct {
	repository *repository.StateRepository
}

func NewStateService(repository *repository.StateRepository) *StateService {
	return &StateService{repository: repository}
}

func (s *StateService) Create(ctx context.Context, req types.StateCreateRequest) (*types.State, error) {
	// Validate required fields
	if req.CountryID == uuid.Nil {
		return nil, errors.New("country_id is required")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Code == "" {
		return nil, errors.New("code is required")
	}

	// Check if country exists
	countryRepo := repository.NewCountryRepository(s.repository.db)
	country, err := countryRepo.GetByID(ctx, req.CountryID)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return nil, errors.New("country not found")
	}

	// Check if state with same country and code already exists
	existingStates, err := s.repository.List(ctx, types.StateFilter{
		CountryID: &req.CountryID,
		Code:      &req.Code,
	})
	if err != nil {
		return nil, err
	}
	if len(existingStates) > 0 {
		return nil, errors.New("state with this code already exists for the country")
	}

	// Create state entity
	state := types.State{
		CountryID: req.CountryID,
		Name:      req.Name,
		Code:      req.Code,
	}

	return s.repository.Create(ctx, state)
}

func (s *StateService) GetByID(ctx context.Context, id uuid.UUID) (*types.State, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *StateService) List(ctx context.Context, filter types.StateFilter) ([]types.State, error) {
	return s.repository.List(ctx, filter)
}

func (s *StateService) ListByCountry(ctx context.Context, countryID uuid.UUID) ([]types.State, error) {
	return s.repository.ListByCountry(ctx, countryID)
}

func (s *StateService) Update(ctx context.Context, id uuid.UUID, req types.StateUpdateRequest) (*types.State, error) {
	// Check if state exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("state not found")
	}

	// If country is being updated, check if new country exists
	if req.CountryID != nil && *req.CountryID != existing.CountryID {
		countryRepo := repository.NewCountryRepository(s.repository.db)
		country, err := countryRepo.GetByID(ctx, *req.CountryID)
		if err != nil {
			return nil, err
		}
		if country == nil {
			return nil, errors.New("new country not found")
		}
	}

	// If code is being updated, check if new code already exists for the country
	if req.Code != nil && *req.Code != existing.Code {
		countryID := existing.CountryID
		if req.CountryID != nil {
			countryID = *req.CountryID
		}
		existingStates, err := s.repository.List(ctx, types.StateFilter{
			CountryID: &countryID,
			Code:      req.Code,
		})
		if err != nil {
			return nil, err
		}
		if len(existingStates) > 0 {
			return nil, errors.New("state with this code already exists for the country")
		}
	}

	return s.repository.Update(ctx, id, req)
}

func (s *StateService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if state exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("state not found")
	}

	return s.repository.Delete(ctx, id)
}
