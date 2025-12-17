package service

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"
	productsRepo "alieze-erp/internal/modules/products/repository"

	"github.com/google/uuid"
)

type ReplenishmentService struct {
	replenishmentRuleRepo repository.ReplenishmentRuleRepository
	replenishmentOrderRepo repository.ReplenishmentOrderRepository
	inventoryRepo         repository.InventoryRepository
	productsRepo          productsRepo.ProductRepo
}

func NewReplenishmentService(
	replenishmentRuleRepo repository.ReplenishmentRuleRepository,
	replenishmentOrderRepo repository.ReplenishmentOrderRepository,
	inventoryRepo repository.InventoryRepository,
	productsRepo productsRepo.ProductRepo,
) *ReplenishmentService {
	return &ReplenishmentService{
		replenishmentRuleRepo: replenishmentRuleRepo,
		replenishmentOrderRepo: replenishmentOrderRepo,
		inventoryRepo:         inventoryRepo,
		productsRepo:          productsRepo,
	}
}

// Replenishment Rule Operations

func (s *ReplenishmentService) CreateReplenishmentRule(ctx context.Context, rule domain.ReplenishmentRule) (*domain.ReplenishmentRule, error) {
	if rule.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if rule.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if rule.TriggerType == "" {
		rule.TriggerType = "reorder_point"
	}
	if rule.ProcureMethod == "" {
		rule.ProcureMethod = "make_to_stock"
	}
	if rule.CheckFrequency == "" {
		rule.CheckFrequency = "daily"
	}
	if rule.Priority == 0 {
		rule.Priority = 10
	}
	if rule.Active == false {
		rule.Active = true
	}

	return s.replenishmentRuleRepo.Create(ctx, rule)
}

func (s *ReplenishmentService) GetReplenishmentRule(ctx context.Context, id uuid.UUID) (*domain.ReplenishmentRule, error) {
	rule, err := s.replenishmentRuleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, fmt.Errorf("replenishment rule not found")
	}
	return rule, nil
}

func (s *ReplenishmentService) ListReplenishmentRules(ctx context.Context, organizationID uuid.UUID) ([]domain.ReplenishmentRule, error) {
	return s.replenishmentRuleRepo.FindAll(ctx, organizationID)
}

func (s *ReplenishmentService) UpdateReplenishmentRule(ctx context.Context, rule domain.ReplenishmentRule) (*domain.ReplenishmentRule, error) {
	return s.replenishmentRuleRepo.Update(ctx, rule)
}

func (s *ReplenishmentService) DeleteReplenishmentRule(ctx context.Context, id uuid.UUID) error {
	return s.replenishmentRuleRepo.Delete(ctx, id)
}

// Replenishment Order Operations

func (s *ReplenishmentService) CreateReplenishmentOrder(ctx context.Context, order domain.ReplenishmentOrder) (*domain.ReplenishmentOrder, error) {
	if order.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if order.RuleID == uuid.Nil {
		return nil, fmt.Errorf("rule_id is required")
	}
	if order.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product_id is required")
	}
	if order.ProductName == "" {
		return nil, fmt.Errorf("product_name is required")
	}
	if order.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if order.Status == "" {
		order.Status = "draft"
	}
	if order.Priority == 0 {
		order.Priority = 10
	}
	if order.ProcureMethod == "" {
		order.ProcureMethod = "make_to_stock"
	}

	return s.replenishmentOrderRepo.Create(ctx, order)
}

func (s *ReplenishmentService) GetReplenishmentOrder(ctx context.Context, id uuid.UUID) (*domain.ReplenishmentOrder, error) {
	order, err := s.replenishmentOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("replenishment order not found")
	}
	return order, nil
}

func (s *ReplenishmentService) ListReplenishmentOrders(ctx context.Context, organizationID uuid.UUID) ([]domain.ReplenishmentOrder, error) {
	return s.replenishmentOrderRepo.FindAll(ctx, organizationID)
}

func (s *ReplenishmentService) ListReplenishmentOrdersByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]domain.ReplenishmentOrder, error) {
	return s.replenishmentOrderRepo.FindByStatus(ctx, organizationID, status)
}

func (s *ReplenishmentService) UpdateReplenishmentOrder(ctx context.Context, order domain.ReplenishmentOrder) (*domain.ReplenishmentOrder, error) {
	return s.replenishmentOrderRepo.Update(ctx, order)
}

func (s *ReplenishmentService) DeleteReplenishmentOrder(ctx context.Context, id uuid.UUID) error {
	return s.replenishmentOrderRepo.Delete(ctx, id)
}

// Replenishment Processing Operations

func (s *ReplenishmentService) CheckAndCreateReplenishmentOrders(ctx context.Context, organizationID uuid.UUID, limit int) ([]domain.ReplenishmentCheckResult, error) {
	if limit <= 0 {
		limit = 100
	}

	return s.replenishmentRuleRepo.CheckAndCreateReplenishmentOrders(ctx, organizationID, limit)
}

func (s *ReplenishmentService) ProcessReplenishmentOrders(ctx context.Context, organizationID uuid.UUID, limit int) ([]domain.ReplenishmentOrder, error) {
	if limit <= 0 {
		limit = 20
	}

	return s.replenishmentOrderRepo.ProcessReplenishmentOrders(ctx, organizationID, limit)
}

func (s *ReplenishmentService) RunReplenishmentCycle(ctx context.Context, organizationID uuid.UUID) (map[string]interface{}, error) {
	return s.replenishmentOrderRepo.RunReplenishmentCycle(ctx, organizationID)
}

// Manual Replenishment Check

func (s *ReplenishmentService) CheckReplenishmentNeeds(ctx context.Context, organizationID uuid.UUID) ([]domain.ReplenishmentCheckResult, error) {
	// This is a simplified version that checks basic reorder points
	// For a more comprehensive check, use CheckAndCreateReplenishmentOrders

	// Get all active replenishment rules
	rules, err := s.replenishmentRuleRepo.FindAll(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replenishment rules: %w", err)
	}

	var results []domain.ReplenishmentCheckResult

	for _, rule := range rules {
		if !rule.Active {
			continue
		}

		// For now, we'll just check products with specific rules
		// In a real implementation, we'd check all products against their reorder points
		if rule.ProductID != nil {
			// Get current stock for this product
			stock, err := s.inventoryRepo.GetProductStock(ctx, organizationID, *rule.ProductID)
			if err != nil {
				continue // Skip if we can't get stock
			}

			currentQty := 0.0
			for _, quant := range stock {
				currentQty += quant.Quantity
			}

			// Determine reorder point and safety stock
			reorderPoint := 0.0
			safetyStock := 0.0

			if rule.ReorderPoint != nil {
				reorderPoint = *rule.ReorderPoint
			}
			if rule.SafetyStock != nil {
				safetyStock = *rule.SafetyStock
			}

			// Check if replenishment is needed
			if currentQty <= reorderPoint {
				recommendedQty := reorderPoint - currentQty + safetyStock
				if recommendedQty <= 0 {
					recommendedQty = safetyStock
				}

				// Get actual product name from database
				productName := "Product " + rule.ProductID.String() // Default fallback
				if rule.ProductID != nil {
					// Query the products table to get the actual product name
					product, err := s.productsRepo.FindByID(ctx, *rule.ProductID)
					if err == nil && product != nil {
						productName = product.Name
					}
				}

				// Get location name if available
				locationName := ""
				if rule.LocationID != nil {
					location, err := s.inventoryRepo.GetLocation(ctx, *rule.LocationID)
					if err == nil && location != nil {
						locationName = location.Name
					}
				}

				results = append(results, domain.ReplenishmentCheckResult{
					ProductID:           *rule.ProductID,
					ProductName:         productName,
					CurrentQuantity:      currentQty,
					ReorderPoint:        reorderPoint,
					SafetyStock:         safetyStock,
					RecommendedQuantity: recommendedQty,
					LocationID:          uuid.Nil, // Would be set to actual location ID in real implementation
					LocationName:        locationName,
					RuleID:             rule.ID,
					RuleName:           rule.Name,
					Priority:           rule.Priority,
					ProcureMethod:      rule.ProcureMethod,
				})
			}
		}
	}

	return results, nil
}

// Scheduled Replenishment Check

func (s *ReplenishmentService) ScheduledReplenishmentCheck(ctx context.Context, organizationID uuid.UUID) error {
	// Run the replenishment cycle
	_, err := s.RunReplenishmentCycle(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to run replenishment cycle: %w", err)
	}

	// Update rule check times
	err = s.replenishmentRuleRepo.UpdateRuleCheckTimes(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update rule check times: %w", err)
	}

	return nil
}

// Get Replenishment Statistics

func (s *ReplenishmentService) GetReplenishmentStatistics(ctx context.Context, organizationID uuid.UUID) (map[string]interface{}, error) {
	// Get counts of replenishment rules and orders
	rules, err := s.replenishmentRuleRepo.FindAll(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replenishment rules: %w", err)
	}

	orders, err := s.replenishmentOrderRepo.FindAll(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replenishment orders: %w", err)
	}

	draftOrders, err := s.replenishmentOrderRepo.FindByStatus(ctx, organizationID, "draft")
	if err != nil {
		return nil, fmt.Errorf("failed to get draft replenishment orders: %w", err)
	}

	activeRules := 0
	for _, rule := range rules {
		if rule.Active {
			activeRules++
		}
	}

	return map[string]interface{}{
		"total_rules":        len(rules),
		"active_rules":       activeRules,
		"total_orders":       len(orders),
		"draft_orders":      len(draftOrders),
		"processed_orders":  len(orders) - len(draftOrders),
		"last_updated":      time.Now(),
	}, nil
}
