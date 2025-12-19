package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	deliveryhandler "alieze-erp/internal/modules/delivery/handler"
	deliveryrepository "alieze-erp/internal/modules/delivery/repository"
	deliveryservice "alieze-erp/internal/modules/delivery/service"
	deliverytypes "alieze-erp/internal/modules/delivery/types"
	inventorytypes "alieze-erp/internal/modules/inventory/types"
	salestypes "alieze-erp/internal/modules/sales/types"
	"alieze-erp/pkg/registry"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// DeliveryModule represents the Delivery Tracking module
type DeliveryModule struct {
	deliveryVehicleHandler  *deliveryhandler.DeliveryVehicleHandler
	deliveryRouteHandler    *deliveryhandler.DeliveryRouteHandler
	deliveryTrackingHandler *deliveryhandler.DeliveryTrackingHandler
	deliveryRouteService    *deliveryservice.DeliveryRouteService
	deliveryTrackingService *deliveryservice.DeliveryTrackingService
	inventoryService        InventoryServiceInterface
	logger                  *slog.Logger
}

// InventoryServiceInterface defines the interface for inventory service dependency
type InventoryServiceInterface interface {
	GetStockMovesByPickingID(ctx context.Context, pickingID uuid.UUID) ([]inventorytypes.StockMove, error)
	GetStockPicking(ctx context.Context, pickingID uuid.UUID) (*inventorytypes.StockPicking, error)
}

// NewDeliveryModule creates a new Delivery Tracking module
func NewDeliveryModule() *DeliveryModule {
	return &DeliveryModule{}
}

// Name returns the module name
func (m *DeliveryModule) Name() string {
	return "delivery"
}

// Init initializes the Delivery Tracking module
func (m *DeliveryModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "delivery")
	m.logger.Info("Initializing Delivery Tracking module")

	// Create repositories
	deliveryVehicleRepo := deliveryrepository.NewDeliveryVehicleRepository(deps.DB)
	deliveryRouteRepo := deliveryrepository.NewDeliveryRouteRepository(deps.DB)
	deliveryTrackingRepo := deliveryrepository.NewDeliveryTrackingRepository(deps.DB)

	// Create services with event bus support
	deliveryVehicleService := deliveryservice.NewDeliveryVehicleService(deliveryVehicleRepo)
	// We need to pass the event bus to services if they need to publish events
	// Casting deps.EventBus to interface{} as the service expects
	m.deliveryRouteService = deliveryservice.NewDeliveryRouteServiceWithEventBus(deliveryRouteRepo, deps.EventBus)
	m.deliveryTrackingService = deliveryservice.NewDeliveryTrackingServiceWithEventBus(deliveryTrackingRepo, deps.EventBus)

	// Get inventory service from dependencies if available
	if deps.InventoryService != nil {
		if invService, ok := deps.InventoryService.(InventoryServiceInterface); ok {
			m.inventoryService = invService
			m.logger.Info("Delivery tracking initialized with inventory service integration")
		} else {
			m.logger.Warn("Inventory service type mismatch, continuing without inventory integration")
		}
	} else {
		m.logger.Warn("Inventory service not available - some delivery features may be limited")
	}

	// Create handlers
	m.deliveryVehicleHandler = deliveryhandler.NewDeliveryVehicleHandler(deliveryVehicleService)
	m.deliveryRouteHandler = deliveryhandler.NewDeliveryRouteHandler(m.deliveryRouteService)
	m.deliveryTrackingHandler = deliveryhandler.NewDeliveryTrackingHandler(m.deliveryTrackingService)

	m.logger.Info("Delivery Tracking module initialized successfully")
	return nil
}

// RegisterRoutes registers Delivery Tracking module routes
func (m *DeliveryModule) RegisterRoutes(router interface{}) {
	if router != nil {
		if r, ok := router.(*httprouter.Router); ok {
			if m.deliveryVehicleHandler != nil {
				m.deliveryVehicleHandler.RegisterRoutes(r)
			}
			if m.deliveryRouteHandler != nil {
				m.deliveryRouteHandler.RegisterRoutes(r)
			}

			if m.deliveryTrackingHandler != nil {
				m.deliveryTrackingHandler.RegisterRoutes(r)
			}
		}
	}
}

// RegisterEventHandlers registers event handlers for the Delivery Tracking module
func (m *DeliveryModule) RegisterEventHandlers(bus interface{}) {
	if bus == nil {
		return
	}

	// Subscribe to relevant events from other modules
	if eventBus, ok := bus.(interface {
		Subscribe(eventType string, handler func(ctx context.Context, event interface{}) error)
	}); ok {
		// Listen to sales order events for delivery creation
		// Note: The sales module publishes "order.confirmed", not "sales_order.confirmed"
		eventBus.Subscribe("order.confirmed", m.handleSalesOrderConfirmed)
		// Listen to stock picking events for shipment creation
		eventBus.Subscribe("stock_picking.ready", m.handleStockPickingReady)
		// Listen to inventory events for delivery updates
		eventBus.Subscribe("inventory.stock_move.done", m.handleStockMoveDone)

		m.logger.Info("Delivery Tracking module event handlers registered")
	}
}

// handleSalesOrderConfirmed handles sales order confirmed events
func (m *DeliveryModule) handleSalesOrderConfirmed(ctx context.Context, event interface{}) error {
	m.logger.Info("Received order.confirmed event", "event", event)

	// Attempt to cast event to SalesOrder
	// The event payload might be the struct itself or a map, depending on how it was published/marshaled
	var order salestypes.SalesOrder

	// Helper to decode map or struct to SalesOrder
	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	if err := json.Unmarshal(bytes, &order); err != nil {
		return fmt.Errorf("failed to unmarshal event data to SalesOrder: %w", err)
	}

	// Create a delivery route (draft) for this order
	// In a real scenario, we might try to group orders or optimize, but for now 1 order = 1 route
	route := deliverytypes.DeliveryRoute{
		ID:                    uuid.New(),
		OrganizationID:        order.OrganizationID,
		CompanyID:             &order.CompanyID,
		Name:                  fmt.Sprintf("Delivery for Order %s", order.Reference),
		RouteCode:             fmt.Sprintf("RT-%s", order.Reference), // Simple code generation
		TransportMode:         deliverytypes.TransportModeRoad,       // Default
		Status:                deliverytypes.RouteStatusDraft,
		ScheduledStartAt:      nil, // To be scheduled
		DestinationLocationID: nil, // We would need to resolve customer location ID here
		Metadata: map[string]interface{}{
			"source_order_id": order.ID.String(),
			"customer_id":     order.CustomerID.String(),
		},
	}

	createdRoute, err := m.deliveryRouteService.CreateDeliveryRoute(ctx, route)
	if err != nil {
		m.logger.Error("Failed to create delivery route from sales order", "error", err, "order_id", order.ID)
		return err
	}

	m.logger.Info("Created delivery route from sales order", "route_id", createdRoute.ID, "order_id", order.ID)
	return nil
}

// handleStockPickingReady handles stock picking ready events
func (m *DeliveryModule) handleStockPickingReady(ctx context.Context, event interface{}) error {
	m.logger.Info("Received stock_picking.ready event", "event", event)

	// Assuming event structure contains PickingID and related info
	// Since we don't have the Inventory types imported, we'll work with a map representation
	var eventData map[string]interface{}
	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	if err := json.Unmarshal(bytes, &eventData); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Extract necessary fields safely
	pickingIDStr, _ := eventData["id"].(string)
	orgIDStr, _ := eventData["organization_id"].(string)

	if pickingIDStr == "" || orgIDStr == "" {
		return fmt.Errorf("missing required fields in stock_picking.ready event")
	}

	pickingID, err := uuid.Parse(pickingIDStr)
	if err != nil {
		return fmt.Errorf("invalid picking_id: %w", err)
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return fmt.Errorf("invalid organization_id: %w", err)
	}

	// Create a shipment linked to this picking
	shipment := deliverytypes.DeliveryShipment{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PickingID:      pickingID,
		Status:         deliverytypes.ShipmentStatusDraft,
		ShipmentType:   deliverytypes.ShipmentTypeOutbound,
		Metadata:       eventData, // Store original event data for context
	}

	// Try to find if there is already a Route for the related order
	// This would require more complex lookup (Picking -> Order -> Route) which we skip for now

	createdShipment, err := m.deliveryTrackingService.CreateShipment(ctx, shipment)
	if err != nil {
		m.logger.Error("Failed to create shipment from stock picking", "error", err, "picking_id", pickingID)
		return err
	}

	m.logger.Info("Created shipment from stock picking", "shipment_id", createdShipment.ID, "picking_id", pickingID)
	return nil
}

// handleStockMoveDone handles inventory stock move done events
func (m *DeliveryModule) handleStockMoveDone(ctx context.Context, event interface{}) error {
	m.logger.Info("Processing inventory.stock_move.done event", "event", event)

	// Parse the stock move event
	stockMove, err := m.parseStockMoveEvent(event)
	if err != nil {
		m.logger.Error("Failed to parse stock move event", "error", err)
		return fmt.Errorf("failed to parse stock move event: %w", err)
	}

	// Skip if no picking ID (internal moves)
	if stockMove.PickingID == nil {
		m.logger.Debug("Stock move has no picking ID, skipping delivery processing", "move_id", stockMove.ID)
		return nil
	}

	pickingID := *stockMove.PickingID

	// Find associated shipment
	shipment, err := m.deliveryTrackingService.GetShipmentByPickingID(ctx, pickingID)
	if err != nil {
		m.logger.Error("Failed to find shipment for picking", "error", err, "picking_id", pickingID)
		return fmt.Errorf("failed to find shipment: %w", err)
	}

	if shipment == nil {
		m.logger.Debug("No shipment found for picking, creating one", "picking_id", pickingID)
		return m.createShipmentFromPicking(ctx, pickingID, stockMove.OrganizationID)
	}

	// Check if all moves for this picking are now done
	allMovesDone, err := m.areAllMovesCompleted(ctx, pickingID)
	if err != nil {
		m.logger.Error("Failed to check picking completion status", "error", err, "picking_id", pickingID)
		return err
	}

	if !allMovesDone {
		m.logger.Debug("Not all moves completed yet", "picking_id", pickingID, "shipment_id", shipment.ID)
		return m.updatePartialProgress(ctx, shipment, stockMove)
	}

	// All moves are done - update shipment status
	return m.processPickingCompletion(ctx, shipment, pickingID)
}

func (m *DeliveryModule) parseStockMoveEvent(event interface{}) (*inventorytypes.StockMove, error) {
	var stockMove inventorytypes.StockMove

	bytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := json.Unmarshal(bytes, &stockMove); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to StockMove: %w", err)
	}

	// Validate required fields
	if stockMove.ID == uuid.Nil {
		return nil, fmt.Errorf("stock move ID is required")
	}

	if stockMove.State != "done" {
		return nil, fmt.Errorf("expected stock move state 'done', got '%s'", stockMove.State)
	}

	return &stockMove, nil
}

func (m *DeliveryModule) areAllMovesCompleted(ctx context.Context, pickingID uuid.UUID) (bool, error) {
	if m.inventoryService == nil {
		return false, fmt.Errorf("inventory service not available")
	}

	stockMoves, err := m.inventoryService.GetStockMovesByPickingID(ctx, pickingID)
	if err != nil {
		return false, fmt.Errorf("failed to get stock moves: %w", err)
	}

	for _, move := range stockMoves {
		if move.State != "done" && move.State != "cancel" {
			return false, nil
		}
	}

	return true, nil
}

func (m *DeliveryModule) processPickingCompletion(ctx context.Context, shipment *deliverytypes.DeliveryShipment, pickingID uuid.UUID) error {
	// Get picking details to determine next status
	picking, err := m.inventoryService.GetStockPicking(ctx, pickingID)
	if err != nil {
		return fmt.Errorf("failed to get picking details: %w", err)
	}

	// Determine new shipment status based on picking type and current status
	newStatus := m.determineShipmentStatus(shipment, picking)

	// Update shipment status
	updatedShipment, err := m.deliveryTrackingService.UpdateShipmentStatus(ctx, shipment.ID, newStatus)
	if err != nil {
		return fmt.Errorf("failed to update shipment status: %w", err)
	}

	// Create tracking event
	err = m.createTrackingEvent(ctx, updatedShipment, fmt.Sprintf("All stock moves completed for picking %s", pickingID))
	if err != nil {
		m.logger.Error("Failed to create tracking event", "error", err, "shipment_id", shipment.ID)
	}

	// Update mobile stock if enabled
	if m.shouldTrackMobileStock(updatedShipment) {
		err = m.updateMobileStock(ctx, updatedShipment, pickingID)
		if err != nil {
			m.logger.Error("Failed to update mobile stock", "error", err, "shipment_id", shipment.ID)
		}
	}

	m.logger.Info("Successfully processed picking completion",
		"shipment_id", updatedShipment.ID,
		"picking_id", pickingID,
		"new_status", newStatus)

	return nil
}

func (m *DeliveryModule) shouldTrackMobileStock(shipment *deliverytypes.DeliveryShipment) bool {
	// Check configuration or shipment metadata to determine if mobile stock tracking is enabled
	if metadata, ok := shipment.Metadata["track_mobile_stock"].(bool); ok {
		return metadata
	}
	return false // Default to disabled
}

func (m *DeliveryModule) updateMobileStock(ctx context.Context, shipment *deliverytypes.DeliveryShipment, pickingID uuid.UUID) error {
	// Get stock moves to understand what products/quantities were moved
	stockMoves, err := m.inventoryService.GetStockMovesByPickingID(ctx, pickingID)
	if err != nil {
		return fmt.Errorf("failed to get stock moves for mobile tracking: %w", err)
	}

	// Get vehicle assignment for this shipment
	vehicleID, err := m.getVehicleForShipment(ctx, shipment)
	if err != nil {
		return fmt.Errorf("failed to get vehicle assignment: %w", err)
	}

	// Update vehicle inventory based on move type
	for _, move := range stockMoves {
		if move.State != "done" {
			continue
		}

		// Determine if this is loading (pickup) or unloading (delivery)
		isLoading := m.isLoadingOperation(&move, shipment)

		err = m.updateVehicleInventory(ctx, vehicleID, move.ProductID, move.Quantity, isLoading)
		if err != nil {
			m.logger.Error("Failed to update vehicle inventory",
				"error", err,
				"vehicle_id", vehicleID,
				"product_id", move.ProductID)
		}
	}

	return nil
}

func (m *DeliveryModule) updateVehicleInventory(ctx context.Context, vehicleID, productID uuid.UUID, quantity float64, isLoading bool) error {
	// This would integrate with a vehicle inventory tracking system
	// For now, just log the operation
	operation := "unload"
	if isLoading {
		operation = "load"
	}

	m.logger.Info("Vehicle inventory update",
		"vehicle_id", vehicleID,
		"product_id", productID,
		"quantity", quantity,
		"operation", operation)

	// TODO: Implement actual vehicle inventory tracking
	// This could involve:
	// - Updating vehicle_inventory table
	// - Publishing inventory.vehicle.updated events
	// - Triggering route optimization if needed

	return nil
}

func (m *DeliveryModule) determineShipmentStatus(shipment *deliverytypes.DeliveryShipment, picking *inventorytypes.StockPicking) deliverytypes.ShipmentStatus {
	// Business logic for status transitions based on picking type and current status
	switch shipment.Status {
	case deliverytypes.ShipmentStatusDraft:
		// First completion moves to scheduled
		return deliverytypes.ShipmentStatusScheduled
	case deliverytypes.ShipmentStatusScheduled, deliverytypes.ShipmentStatusInTransit:
		// If this is a delivery picking, mark as delivered
		if m.isDeliveryPicking(picking) {
			return deliverytypes.ShipmentStatusDelivered
		}
		// If this is a pickup, mark as in transit
		return deliverytypes.ShipmentStatusInTransit
	default:
		return shipment.Status // No change
	}
}

func (m *DeliveryModule) isDeliveryPicking(picking *inventorytypes.StockPicking) bool {
	// Determine if picking is for delivery (vs pickup)
	// This logic depends on your picking type configuration
	return picking.State == "done" && picking.LocationDestID != nil
}

func (m *DeliveryModule) createTrackingEvent(ctx context.Context, shipment *deliverytypes.DeliveryShipment, message string) error {
	event := deliverytypes.DeliveryTrackingEvent{
		ID:             uuid.New(),
		OrganizationID: shipment.OrganizationID,
		ShipmentID:     shipment.ID,
		EventType:      "status_change",
		Status:         string(shipment.Status),
		EventTime:      time.Now(),
		Source:         "stock_move_completion",
		Message:        message,
		RawPayload: map[string]interface{}{
			"triggered_by": "stock_move_completion",
		},
	}

	_, err := m.deliveryTrackingService.CreateTrackingEvent(ctx, event)
	return err
}

func (m *DeliveryModule) updatePartialProgress(ctx context.Context, shipment *deliverytypes.DeliveryShipment, stockMove *inventorytypes.StockMove) error {
	// Track partial completion for progress monitoring
	m.logger.Info("Partial picking progress",
		"shipment_id", shipment.ID,
		"completed_move", stockMove.ID,
		"product_id", stockMove.ProductID,
		"quantity", stockMove.Quantity)

	// Could create tracking events for partial progress
	return nil
}

func (m *DeliveryModule) createShipmentFromPicking(ctx context.Context, pickingID, orgID uuid.UUID) error {
	// Create a new shipment for this picking
	shipment := deliverytypes.DeliveryShipment{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PickingID:      pickingID,
		Status:         deliverytypes.ShipmentStatusDraft,
		ShipmentType:   deliverytypes.ShipmentTypeOutbound,
		Metadata: map[string]interface{}{
			"created_from": "stock_move_completion",
		},
	}

	_, err := m.deliveryTrackingService.CreateShipment(ctx, shipment)
	if err != nil {
		m.logger.Error("Failed to create shipment from stock move completion", "error", err, "picking_id", pickingID)
		return err
	}

	m.logger.Info("Created shipment from stock move completion", "picking_id", pickingID)
	return nil
}

func (m *DeliveryModule) getVehicleForShipment(ctx context.Context, shipment *deliverytypes.DeliveryShipment) (uuid.UUID, error) {
	// This would look up the vehicle assignment for the shipment
	// For now, return a zero UUID as a placeholder
	// In a real implementation, this would query the delivery route or vehicle assignment
	return uuid.Nil, nil
}

func (m *DeliveryModule) isLoadingOperation(move *inventorytypes.StockMove, shipment *deliverytypes.DeliveryShipment) bool {
	// Determine if this move represents loading (pickup) or unloading (delivery)
	// This depends on the shipment type and move direction
	// For simplicity, assume outbound shipments are unloading, inbound are loading
	return shipment.ShipmentType == deliverytypes.ShipmentTypeInbound
}

// Health checks the health of the Delivery Tracking module
func (m *DeliveryModule) Health() error {
	return nil
}
