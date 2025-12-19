package service

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
)

// StockRuleService handles business logic for stock rules
type StockRuleService struct {
	repo   *repository.StockRuleRepository
}

// NewStockRuleService creates a new StockRuleService
func NewStockRuleService(repo *repository.StockRuleRepository) *StockRuleService {
	return &StockRuleService{
		repo: repo,
	}
}

// Create creates a new stock rule
func (s *StockRuleService) Create(ctx context.Context, orgID uuid.UUID, req types.StockRuleCreateRequest) (*types.StockRule, error) {
	return s.repo.Create(ctx, orgID, req)
}

// GetByID retrieves a stock rule by ID
func (s *StockRuleService) GetByID(ctx context.Context, id uuid.UUID) (*types.StockRule, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all stock rules for an organization
func (s *StockRuleService) List(ctx context.Context, orgID uuid.UUID) ([]types.StockRule, error) {
	return s.repo.List(ctx, orgID)
}

// Update updates a stock rule
func (s *StockRuleService) Update(ctx context.Context, id uuid.UUID, req types.StockRuleUpdateRequest) (*types.StockRule, error) {
	return s.repo.Update(ctx, id, req)
}

// Delete deletes a stock rule
func (s *StockRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
