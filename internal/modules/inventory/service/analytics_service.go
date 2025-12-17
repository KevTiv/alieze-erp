package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type AnalyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// Valuation Services

func (s *AnalyticsService) GetInventoryValuation(ctx context.Context, orgID uuid.UUID, request types.AnalyticsRequest) ([]types.InventoryValuation, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetInventoryValuation(ctx, orgID, request)
}

func (s *AnalyticsService) GetValuationByProduct(ctx context.Context, orgID, productID uuid.UUID) (*types.InventoryValuation, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product_id is required")
	}

	return s.analyticsRepo.GetValuationByProduct(ctx, orgID, productID)
}

func (s *AnalyticsService) GetValuationSummary(ctx context.Context, orgID uuid.UUID) (*types.AnalyticsSummary, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetValuationSummary(ctx, orgID)
}

// Turnover Services

func (s *AnalyticsService) GetInventoryTurnover(ctx context.Context, orgID uuid.UUID, request types.AnalyticsRequest) ([]types.InventoryTurnover, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetInventoryTurnover(ctx, orgID, request)
}

func (s *AnalyticsService) GetTurnoverByProduct(ctx context.Context, orgID, productID uuid.UUID) (*types.InventoryTurnover, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product_id is required")
	}

	return s.analyticsRepo.GetTurnoverByProduct(ctx, orgID, productID)
}

// Aging Services

func (s *AnalyticsService) GetInventoryAging(ctx context.Context, orgID uuid.UUID, request types.AnalyticsRequest) ([]types.InventoryAging, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetInventoryAging(ctx, orgID, request)
}

func (s *AnalyticsService) GetAgingSummary(ctx context.Context, orgID uuid.UUID) (map[string]float64, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetAgingSummary(ctx, orgID)
}

// Dead Stock Services

func (s *AnalyticsService) GetDeadStock(ctx context.Context, orgID uuid.UUID, request types.AnalyticsRequest) ([]types.InventoryDeadStock, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetDeadStock(ctx, orgID, request)
}

func (s *AnalyticsService) GetDeadStockSummary(ctx context.Context, orgID uuid.UUID) (*types.AnalyticsSummary, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetDeadStockSummary(ctx, orgID)
}

// Movement Summary Services

func (s *AnalyticsService) GetMovementSummary(ctx context.Context, orgID uuid.UUID, request types.AnalyticsRequest) ([]types.InventoryMovementSummary, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetMovementSummary(ctx, orgID, request)
}

// Reorder Analysis Services

func (s *AnalyticsService) GetReorderAnalysis(ctx context.Context, orgID uuid.UUID, request types.AnalyticsRequest) ([]types.InventoryReorderAnalysis, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetReorderAnalysis(ctx, orgID, request)
}

func (s *AnalyticsService) GetProductsNeedingReorder(ctx context.Context, orgID uuid.UUID) ([]types.InventoryReorderAnalysis, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetProductsNeedingReorder(ctx, orgID)
}

// Snapshot Services

func (s *AnalyticsService) GetInventorySnapshot(ctx context.Context, orgID uuid.UUID) ([]types.InventorySnapshot, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.GetInventorySnapshot(ctx, orgID)
}

// Refresh Services

func (s *AnalyticsService) RefreshOrganizationAnalytics(ctx context.Context, orgID uuid.UUID) error {
	if orgID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}

	return s.analyticsRepo.RefreshOrganizationAnalytics(ctx, orgID)
}

// Dashboard Services

func (s *AnalyticsService) GetInventoryDashboard(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	dashboard := make(map[string]interface{})

	// Get summary metrics
	summary, err := s.GetValuationSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}
	dashboard["summary"] = summary

	// Get top products by value
	valuationRequest := types.AnalyticsRequest{
		Limit: func() *int { i := 10; return &i }(),
	}
	valuations, err := s.GetInventoryValuation(ctx, orgID, valuationRequest)
	if err != nil {
		return nil, err
	}
	dashboard["top_products_by_value"] = valuations

	// Get aging summary
	agingSummary, err := s.GetAgingSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}
	dashboard["aging_summary"] = agingSummary

	// Get products needing reorder
	reorderProducts, err := s.GetProductsNeedingReorder(ctx, orgID)
	if err != nil {
		return nil, err
	}
	dashboard["products_needing_reorder"] = reorderProducts

	// Get dead stock summary
	deadStockSummary, err := s.GetDeadStockSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}
	dashboard["dead_stock_summary"] = deadStockSummary

	return dashboard, nil
}
