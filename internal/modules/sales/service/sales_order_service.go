package service

import (
	"context"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/sales/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/sales/types"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/KevTiv/alieze-erp/pkg/tax"

	"github.com/google/uuid"
)

type SalesOrderService struct {
	repo          repository.SalesOrderRepository
	pricelistRepo repository.PricelistRepository
	eventBus      *events.Bus
	taxCalc       *tax.Calculator
}

func NewSalesOrderService(repo repository.SalesOrderRepository, pricelistRepo repository.PricelistRepository, taxCalc *tax.Calculator) *SalesOrderService {
	return &SalesOrderService{
		repo:          repo,
		pricelistRepo: pricelistRepo,
		taxCalc:       taxCalc,
	}
}

// NewSalesOrderServiceWithEventBus creates a sales order service with event bus support
func NewSalesOrderServiceWithEventBus(repo repository.SalesOrderRepository, pricelistRepo repository.PricelistRepository, taxCalc *tax.Calculator, eventBus *events.Bus) *SalesOrderService {
	service := NewSalesOrderService(repo, pricelistRepo, taxCalc)
	service.eventBus = eventBus
	return service
}

func (s *SalesOrderService) CreateSalesOrder(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	// Validate the order
	if err := s.validateSalesOrder(order); err != nil {
		return nil, fmt.Errorf("invalid sales order: %w", err)
	}

	// Set default values
	if order.Status == "" {
		order.Status = types.SalesOrderStatusDraft
	}
	if order.OrderDate.IsZero() {
		order.OrderDate = time.Now()
	}
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	// Calculate amounts
	if err := s.calculateOrderAmounts(ctx, &order); err != nil {
		return nil, fmt.Errorf("failed to calculate order amounts: %w", err)
	}

	// Create the order
	createdOrder, err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales order: %w", err)
	}

	// Publish order.created event
	s.publishEvent(ctx, "order.created", createdOrder)

	return createdOrder, nil
}

func (s *SalesOrderService) GetSalesOrder(ctx context.Context, id uuid.UUID) (*types.SalesOrder, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}
	if order == nil {
		return nil, nil
	}

	return order, nil
}

func (s *SalesOrderService) ListSalesOrders(ctx context.Context, filters repository.SalesOrderFilter) ([]types.SalesOrder, error) {
	orders, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales orders: %w", err)
	}

	return orders, nil
}

func (s *SalesOrderService) UpdateSalesOrder(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	// Validate the order
	if err := s.validateSalesOrder(order); err != nil {
		return nil, fmt.Errorf("invalid sales order: %w", err)
	}

	// Calculate amounts
	if err := s.calculateOrderAmounts(ctx, &order); err != nil {
		return nil, fmt.Errorf("failed to calculate order amounts: %w", err)
	}

	// Update the order
	updatedOrder, err := s.repo.Update(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales order: %w", err)
	}

	// Publish order.updated event
	s.publishEvent(ctx, "order.updated", updatedOrder)

	return updatedOrder, nil
}

func (s *SalesOrderService) DeleteSalesOrder(ctx context.Context, id uuid.UUID) error {
	// Check if order exists
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check sales order existence: %w", err)
	}
	if order == nil {
		return nil
	}

	// Prevent deletion of confirmed orders
	if order.Status == types.SalesOrderStatusConfirmed || order.Status == types.SalesOrderStatusDone {
		return fmt.Errorf("cannot delete confirmed or done sales orders")
	}

	// Delete the order
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete sales order: %w", err)
	}

	// Publish order.deleted event
	s.publishEvent(ctx, "order.deleted", map[string]interface{}{
		"id":              id,
		"organization_id": order.OrganizationID,
	})

	return nil
}

func (s *SalesOrderService) ConfirmSalesOrder(ctx context.Context, id uuid.UUID) (*types.SalesOrder, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("sales order not found")
	}

	// Validate order for confirmation
	if order.Status != types.SalesOrderStatusDraft && order.Status != types.SalesOrderStatusQuotation {
		return nil, fmt.Errorf("only draft or quotation orders can be confirmed")
	}

	if len(order.Lines) == 0 {
		return nil, fmt.Errorf("sales order must have at least one line to be confirmed")
	}

	// Update status and confirmation date
	order.Status = types.SalesOrderStatusConfirmed
	now := time.Now()
	order.ConfirmationDate = &now
	order.UpdatedAt = now

	updatedOrder, err := s.repo.Update(ctx, *order)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales order: %w", err)
	}

	// Publish order.confirmed event
	s.publishEvent(ctx, "order.confirmed", updatedOrder)

	return updatedOrder, nil
}

func (s *SalesOrderService) CancelSalesOrder(ctx context.Context, id uuid.UUID) (*types.SalesOrder, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("sales order not found")
	}

	// Validate order for cancellation
	if order.Status == types.SalesOrderStatusCancelled || order.Status == types.SalesOrderStatusDone {
		return nil, fmt.Errorf("order cannot be cancelled in its current state")
	}

	// Update status
	order.Status = types.SalesOrderStatusCancelled
	order.UpdatedAt = time.Now()

	updatedOrder, err := s.repo.Update(ctx, *order)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales order: %w", err)
	}

	// Publish order.cancelled event
	s.publishEvent(ctx, "order.cancelled", updatedOrder)

	return updatedOrder, nil
}

func (s *SalesOrderService) GetSalesOrdersByCustomer(ctx context.Context, customerID uuid.UUID) ([]types.SalesOrder, error) {
	orders, err := s.repo.FindByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales orders by customer: %w", err)
	}

	return orders, nil
}

func (s *SalesOrderService) GetSalesOrdersByStatus(ctx context.Context, status types.SalesOrderStatus) ([]types.SalesOrder, error) {
	orders, err := s.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales orders by status: %w", err)
	}

	return orders, nil
}

func (s *SalesOrderService) validateSalesOrder(order types.SalesOrder) error {
	if order.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if order.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	if order.CustomerID == uuid.Nil {
		return fmt.Errorf("customer ID is required")
	}

	if order.PricelistID == uuid.Nil {
		return fmt.Errorf("pricelist ID is required")
	}

	if order.CurrencyID == uuid.Nil {
		return fmt.Errorf("currency ID is required")
	}

	// Validate lines
	if len(order.Lines) == 0 {
		return fmt.Errorf("sales order must have at least one line")
	}

	for _, line := range order.Lines {
		if line.ProductID == uuid.Nil {
			return fmt.Errorf("product ID is required for all lines")
		}
		if line.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive for all lines")
		}
		if line.UnitPrice < 0 {
			return fmt.Errorf("unit price cannot be negative")
		}
		if line.UomID == uuid.Nil {
			return fmt.Errorf("unit of measure ID is required for all lines")
		}
	}

	return nil
}

func (s *SalesOrderService) calculateOrderAmounts(ctx context.Context, order *types.SalesOrder) error {
	var amountUntaxed, amountTax, amountTotal float64

	for i, line := range order.Lines {
		// Calculate line subtotal
		line.PriceSubtotal = line.Quantity * line.UnitPrice
		if line.Discount > 0 {
			line.PriceSubtotal = line.PriceSubtotal * (1 - line.Discount/100)
		}

		// Calculate line tax using tax calculator
		line.PriceTax = 0
		if line.TaxID != nil && s.taxCalc != nil {
			taxAmount, err := s.taxCalc.CalculateLineTax(ctx, *line.TaxID, line.PriceSubtotal)
			if err != nil {
				// Log error but don't fail the calculation
				fmt.Printf("Failed to calculate tax for line: %v\n", err)
				line.PriceTax = 0
			} else {
				line.PriceTax = taxAmount
			}
		}

		// Calculate line total
		line.PriceTotal = line.PriceSubtotal + line.PriceTax

		// Update the line in the slice
		order.Lines[i] = line

		// Accumulate order totals
		amountUntaxed += line.PriceSubtotal
		amountTax += line.PriceTax
		amountTotal += line.PriceTotal
	}

	order.AmountUntaxed = amountUntaxed
	order.AmountTax = amountTax
	order.AmountTotal = amountTotal

	return nil
}

// publishEvent publishes an event to the event bus if available
func (s *SalesOrderService) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if s.eventBus != nil {
		if err := s.eventBus.Publish(ctx, eventType, payload); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish event %s: %v\n", eventType, err)
		}
	}
}
