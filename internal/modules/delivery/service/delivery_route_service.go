package service

import (
	"context"
	"fmt"
	"time"

	deliveryrepository "github.com/KevTiv/alieze-erp/internal/modules/delivery/repository"
	deliverytypes "github.com/KevTiv/alieze-erp/internal/modules/delivery/types"
	"github.com/KevTiv/alieze-erp/pkg/events"

	"github.com/google/uuid"
)

type DeliveryRouteService struct {
	repo     deliveryrepository.DeliveryRouteRepository
	eventBus *events.Bus
}

func NewDeliveryRouteService(repo deliveryrepository.DeliveryRouteRepository) *DeliveryRouteService {
	return &DeliveryRouteService{
		repo: repo,
	}
}

func NewDeliveryRouteServiceWithEventBus(repo deliveryrepository.DeliveryRouteRepository, eventBus *events.Bus) *DeliveryRouteService {
	service := NewDeliveryRouteService(repo)
	service.eventBus = eventBus
	return service
}

func (s *DeliveryRouteService) CreateDeliveryRoute(ctx context.Context, route deliverytypes.DeliveryRoute) (*deliverytypes.DeliveryRoute, error) {
	// Validate the route
	if err := s.validateDeliveryRoute(route); err != nil {
		return nil, fmt.Errorf("invalid delivery route: %w", err)
	}

	// Set default values
	if route.ID == uuid.Nil {
		route.ID = uuid.New()
	}
	if route.TransportMode == "" {
		route.TransportMode = deliverytypes.TransportModeRoad
	}
	if route.Status == "" {
		route.Status = deliverytypes.RouteStatusDraft
	}
	if route.Metadata == nil {
		route.Metadata = make(map[string]interface{})
	}

	// Create the route
	createdRoute, err := s.repo.Create(ctx, route)
	if err != nil {
		return nil, fmt.Errorf("failed to create delivery route: %w", err)
	}

	// Publish event if event bus is available
	s.publishRouteEvent(ctx, "delivery_route.created", *createdRoute)

	return createdRoute, nil
}

func (s *DeliveryRouteService) GetDeliveryRoute(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryRoute, error) {
	route, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery route: %w", err)
	}
	if route == nil {
		return nil, nil
	}
	return route, nil
}

func (s *DeliveryRouteService) ListDeliveryRoutes(ctx context.Context, orgID uuid.UUID, statusFilter *deliverytypes.RouteStatus) ([]deliverytypes.DeliveryRoute, error) {
	if statusFilter != nil {
		return s.repo.FindByStatus(ctx, orgID, *statusFilter)
	}
	return s.repo.FindByOrganizationID(ctx, orgID)
}

func (s *DeliveryRouteService) UpdateDeliveryRoute(ctx context.Context, route deliverytypes.DeliveryRoute) (*deliverytypes.DeliveryRoute, error) {
	// Validate the route
	if err := s.validateDeliveryRoute(route); err != nil {
		return nil, fmt.Errorf("invalid delivery route: %w", err)
	}

	// Get existing route to check if it exists
	existing, err := s.repo.FindByID(ctx, route.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing route: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("delivery route not found")
	}

	// Update the route
	updatedRoute, err := s.repo.Update(ctx, route)
	if err != nil {
		return nil, fmt.Errorf("failed to update delivery route: %w", err)
	}

	// Publish event if event bus is available
	s.publishRouteEvent(ctx, "delivery_route.updated", *updatedRoute)

	return updatedRoute, nil
}

func (s *DeliveryRouteService) DeleteDeliveryRoute(ctx context.Context, id uuid.UUID) error {
	// Get existing route to check if it exists
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find existing route: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("delivery route not found")
	}

	// Delete the route
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery route: %w", err)
	}

	// Publish event if event bus is available
	s.publishRouteEvent(ctx, "delivery_route.deleted", *existing)

	return nil
}

func (s *DeliveryRouteService) StartRoute(ctx context.Context, routeID uuid.UUID) (*deliverytypes.DeliveryRoute, error) {
	// Get the route
	route, err := s.repo.FindByID(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find route: %w", err)
	}
	if route == nil {
		return nil, fmt.Errorf("route not found")
	}

	// Check if route can be started
	if route.Status != deliverytypes.RouteStatusScheduled {
		return nil, fmt.Errorf("route must be in 'scheduled' status to start")
	}

	// Update route status
	now := time.Now()
	route.Status = deliverytypes.RouteStatusInProgress
	route.ActualStartAt = &now

	updatedRoute, err := s.repo.Update(ctx, *route)
	if err != nil {
		return nil, fmt.Errorf("failed to update route status: %w", err)
	}

	// Publish event
	s.publishRouteEvent(ctx, "delivery_route.started", *updatedRoute)

	return updatedRoute, nil
}

func (s *DeliveryRouteService) CompleteRoute(ctx context.Context, routeID uuid.UUID) (*deliverytypes.DeliveryRoute, error) {
	// Get the route
	route, err := s.repo.FindByID(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find route: %w", err)
	}
	if route == nil {
		return nil, fmt.Errorf("route not found")
	}

	// Check if route can be completed
	if route.Status != deliverytypes.RouteStatusInProgress {
		return nil, fmt.Errorf("route must be in 'in_progress' status to complete")
	}

	// Update route status
	now := time.Now()
	route.Status = deliverytypes.RouteStatusCompleted
	route.ActualEndAt = &now

	updatedRoute, err := s.repo.Update(ctx, *route)
	if err != nil {
		return nil, fmt.Errorf("failed to update route status: %w", err)
	}

	// Publish event
	s.publishRouteEvent(ctx, "delivery_route.completed", *updatedRoute)

	return updatedRoute, nil
}

func (s *DeliveryRouteService) validateDeliveryRoute(route deliverytypes.DeliveryRoute) error {
	if route.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if route.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(route.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}
	if route.RouteCode != "" && len(route.RouteCode) > 50 {
		return fmt.Errorf("route_code must be 50 characters or less")
	}
	if route.Notes != "" && len(route.Notes) > 10000 {
		return fmt.Errorf("notes must be 10000 characters or less")
	}
	return nil
}

func (s *DeliveryRouteService) publishRouteEvent(ctx context.Context, eventType string, route deliverytypes.DeliveryRoute) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":                 route.ID,
			"organization_id":    route.OrganizationID,
			"name":               route.Name,
			"route_code":         route.RouteCode,
			"transport_mode":     route.TransportMode,
			"status":             route.Status,
			"scheduled_start_at": route.ScheduledStartAt,
			"scheduled_end_at":   route.ScheduledEndAt,
			"actual_start_at":    route.ActualStartAt,
			"actual_end_at":      route.ActualEndAt,
			"metadata":           route.Metadata,
			"created_at":         route.CreatedAt,
			"updated_at":         route.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}
