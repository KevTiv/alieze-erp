package service

import (
	"context"
	"fmt"
	"time"

	deliveryrepository "alieze-erp/internal/modules/delivery/repository"
	deliverytypes "alieze-erp/internal/modules/delivery/types"
	"alieze-erp/pkg/events"

	"github.com/google/uuid"
)

type DeliveryTrackingService struct {
	repo     deliveryrepository.DeliveryTrackingRepository
	eventBus *events.Bus
}

func NewDeliveryTrackingService(repo deliveryrepository.DeliveryTrackingRepository) *DeliveryTrackingService {
	return &DeliveryTrackingService{
		repo: repo,
	}
}

func NewDeliveryTrackingServiceWithEventBus(repo deliveryrepository.DeliveryTrackingRepository, eventBus *events.Bus) *DeliveryTrackingService {
	service := NewDeliveryTrackingService(repo)
	service.eventBus = eventBus
	return service
}

func (s *DeliveryTrackingService) CreateShipment(ctx context.Context, shipment deliverytypes.DeliveryShipment) (*deliverytypes.DeliveryShipment, error) {
	// Validate the shipment
	if err := s.validateShipment(shipment); err != nil {
		return nil, fmt.Errorf("invalid shipment: %w", err)
	}

	// Set default values
	if shipment.ID == uuid.Nil {
		shipment.ID = uuid.New()
	}
	if shipment.ShipmentType == "" {
		shipment.ShipmentType = deliverytypes.ShipmentTypeOutbound
	}
	if shipment.Status == "" {
		shipment.Status = deliverytypes.ShipmentStatusDraft
	}
	if shipment.Metadata == nil {
		shipment.Metadata = make(map[string]interface{})
	}

	// Create the shipment
	createdShipment, err := s.repo.CreateShipment(ctx, shipment)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Publish event
	s.publishShipmentEvent(ctx, "delivery_shipment.created", *createdShipment)

	return createdShipment, nil
}

func (s *DeliveryTrackingService) GetShipment(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryShipment, error) {
	shipment, err := s.repo.FindShipmentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}
	if shipment == nil {
		return nil, nil
	}
	return shipment, nil
}

func (s *DeliveryTrackingService) GetShipmentByPickingID(ctx context.Context, pickingID uuid.UUID) (*deliverytypes.DeliveryShipment, error) {
	shipment, err := s.repo.FindShipmentsByPickingID(ctx, pickingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment by picking: %w", err)
	}
	if shipment == nil {
		return nil, nil
	}
	return shipment, nil
}

func (s *DeliveryTrackingService) ListShipmentsByRoute(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryShipment, error) {
	return s.repo.FindShipmentsByRouteID(ctx, routeID)
}

func (s *DeliveryTrackingService) UpdateShipmentStatus(ctx context.Context, shipmentID uuid.UUID, status deliverytypes.ShipmentStatus) (*deliverytypes.DeliveryShipment, error) {
	// Get the shipment
	shipment, err := s.repo.FindShipmentByID(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}
	if shipment == nil {
		return nil, fmt.Errorf("shipment not found")
	}

	// Update status
	shipment.Status = status
	now := time.Now()

	// Update appropriate timestamps based on status
	switch status {
	case deliverytypes.ShipmentStatusInTransit:
		if shipment.DepartedAt == nil {
			shipment.DepartedAt = &now
		}
	case deliverytypes.ShipmentStatusDelivered:
		if shipment.ArrivedAt == nil {
			shipment.ArrivedAt = &now
		}
	case deliverytypes.ShipmentStatusFailed, deliverytypes.ShipmentStatusCancelled:
		// No specific timestamp updates for these statuses
	}

	// Update the shipment
	updatedShipment, err := s.repo.UpdateShipment(ctx, *shipment)
	if err != nil {
		return nil, fmt.Errorf("failed to update shipment status: %w", err)
	}

	// Publish event
	s.publishShipmentEvent(ctx, "delivery_shipment.status_updated", *updatedShipment)

	return updatedShipment, nil
}

func (s *DeliveryTrackingService) CreateTrackingEvent(ctx context.Context, event deliverytypes.DeliveryTrackingEvent) (*deliverytypes.DeliveryTrackingEvent, error) {
	// Validate the event
	if err := s.validateTrackingEvent(event); err != nil {
		return nil, fmt.Errorf("invalid tracking event: %w", err)
	}

	// Set default values
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}
	if event.RawPayload == nil {
		event.RawPayload = make(map[string]interface{})
	}

	// Create the event
	createdEvent, err := s.repo.CreateTrackingEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracking event: %w", err)
	}

	// Update shipment status if event contains status
	if event.Status != "" {
		_, err = s.UpdateShipmentStatus(ctx, event.ShipmentID, deliverytypes.ShipmentStatus(event.Status))
		if err != nil {
			// Log error but don't fail the event creation
			fmt.Printf("Warning: failed to update shipment status from tracking event: %v\n", err)
		}
	}

	// Publish event
	s.publishTrackingEvent(ctx, "delivery_tracking.event_created", *createdEvent)

	return createdEvent, nil
}

func (s *DeliveryTrackingService) GetTrackingEvents(ctx context.Context, shipmentID uuid.UUID) ([]deliverytypes.DeliveryTrackingEvent, error) {
	return s.repo.FindTrackingEventsByShipmentID(ctx, shipmentID)
}

func (s *DeliveryTrackingService) GetLatestTrackingEvent(ctx context.Context, shipmentID uuid.UUID) (*deliverytypes.DeliveryTrackingEvent, error) {
	return s.repo.FindLatestTrackingEventByShipmentID(ctx, shipmentID)
}

func (s *DeliveryTrackingService) CreateRoutePosition(ctx context.Context, position deliverytypes.DeliveryRoutePosition) (*deliverytypes.DeliveryRoutePosition, error) {
	// Validate the position
	if err := s.validateRoutePosition(position); err != nil {
		return nil, fmt.Errorf("invalid route position: %w", err)
	}

	// Set default values
	if position.ID == uuid.Nil {
		position.ID = uuid.New()
	}
	if position.RecordedAt.IsZero() {
		position.RecordedAt = time.Now()
	}
	if position.Source == "" {
		position.Source = "gps"
	}
	if position.Metadata == nil {
		position.Metadata = make(map[string]interface{})
	}

	// Create the position
	createdPosition, err := s.repo.CreateRoutePosition(ctx, position)
	if err != nil {
		return nil, fmt.Errorf("failed to create route position: %w", err)
	}

	// Publish event
	s.publishRoutePositionEvent(ctx, "delivery_route.position_created", *createdPosition)

	return createdPosition, nil
}

func (s *DeliveryTrackingService) GetRoutePositions(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRoutePosition, error) {
	return s.repo.FindRoutePositionsByRouteID(ctx, routeID)
}

func (s *DeliveryTrackingService) GetLatestRoutePosition(ctx context.Context, routeID uuid.UUID) (*deliverytypes.DeliveryRoutePosition, error) {
	return s.repo.FindLatestRoutePositionByRouteID(ctx, routeID)
}

func (s *DeliveryTrackingService) CreateRouteAssignment(ctx context.Context, assignment deliverytypes.DeliveryRouteAssignment) (*deliverytypes.DeliveryRouteAssignment, error) {
	// Validate the assignment
	if err := s.validateRouteAssignment(assignment); err != nil {
		return nil, fmt.Errorf("invalid route assignment: %w", err)
	}

	// Set default values
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	if assignment.AssignmentStatus == "" {
		assignment.AssignmentStatus = deliverytypes.AssignmentStatusAssigned
	}
	if assignment.Metadata == nil {
		assignment.Metadata = make(map[string]interface{})
	}

	// Create the assignment
	createdAssignment, err := s.repo.CreateRouteAssignment(ctx, assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to create route assignment: %w", err)
	}

	// Publish event
	s.publishRouteAssignmentEvent(ctx, "delivery_route.assignment_created", *createdAssignment)

	return createdAssignment, nil
}

func (s *DeliveryTrackingService) GetRouteAssignments(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRouteAssignment, error) {
	return s.repo.FindRouteAssignmentsByRouteID(ctx, routeID)
}

func (s *DeliveryTrackingService) CreateRouteStop(ctx context.Context, stop deliverytypes.DeliveryRouteStop) (*deliverytypes.DeliveryRouteStop, error) {
	// Validate the stop
	if err := s.validateRouteStop(stop); err != nil {
		return nil, fmt.Errorf("invalid route stop: %w", err)
	}

	// Set default values
	if stop.ID == uuid.Nil {
		stop.ID = uuid.New()
	}
	if stop.Status == "" {
		stop.Status = deliverytypes.StopStatusPlanned
	}
	if stop.Metadata == nil {
		stop.Metadata = make(map[string]interface{})
	}

	// Create the stop
	createdStop, err := s.repo.CreateRouteStop(ctx, stop)
	if err != nil {
		return nil, fmt.Errorf("failed to create route stop: %w", err)
	}

	// Publish event
	s.publishRouteStopEvent(ctx, "delivery_route.stop_created", *createdStop)

	return createdStop, nil
}

func (s *DeliveryTrackingService) GetRouteStops(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRouteStop, error) {
	return s.repo.FindRouteStopsByRouteID(ctx, routeID)
}

func (s *DeliveryTrackingService) GetRouteStopByShipment(ctx context.Context, shipmentID uuid.UUID) (*deliverytypes.DeliveryRouteStop, error) {
	return s.repo.FindRouteStopByShipmentID(ctx, shipmentID)
}

func (s *DeliveryTrackingService) UpdateRouteStopStatus(ctx context.Context, stopID uuid.UUID, status deliverytypes.StopStatus) (*deliverytypes.DeliveryRouteStop, error) {
	// Get the stop
	stop, err := s.repo.FindRouteStopByShipmentID(ctx, stopID)
	if err != nil {
		return nil, fmt.Errorf("failed to find route stop: %w", err)
	}
	if stop == nil {
		return nil, fmt.Errorf("route stop not found")
	}

	// Update status
	stop.Status = status
	now := time.Now()

	// Update appropriate timestamps based on status
	switch status {
	case deliverytypes.StopStatusArrived:
		if stop.ActualArrivalAt == nil {
			stop.ActualArrivalAt = &now
		}
	case deliverytypes.StopStatusCompleted:
		if stop.ActualDepartureAt == nil {
			stop.ActualDepartureAt = &now
		}
	}

	// Update the stop
	updatedStop, err := s.repo.UpdateRouteStop(ctx, *stop)
	if err != nil {
		return nil, fmt.Errorf("failed to update route stop status: %w", err)
	}

	// Publish event
	s.publishRouteStopEvent(ctx, "delivery_route.stop_updated", *updatedStop)

	return updatedStop, nil
}

func (s *DeliveryTrackingService) validateShipment(shipment deliverytypes.DeliveryShipment) error {
	if shipment.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if shipment.PickingID == uuid.Nil {
		return fmt.Errorf("picking_id is required")
	}
	if shipment.TrackingNumber != "" && len(shipment.TrackingNumber) > 100 {
		return fmt.Errorf("tracking_number must be 100 characters or less")
	}
	if shipment.CarrierName != "" && len(shipment.CarrierName) > 120 {
		return fmt.Errorf("carrier_name must be 120 characters or less")
	}
	return nil
}

func (s *DeliveryTrackingService) validateTrackingEvent(event deliverytypes.DeliveryTrackingEvent) error {
	if event.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if event.ShipmentID == uuid.Nil {
		return fmt.Errorf("shipment_id is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.Latitude != nil && (*event.Latitude < -90 || *event.Latitude > 90) {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if event.Longitude != nil && (*event.Longitude < -180 || *event.Longitude > 180) {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

func (s *DeliveryTrackingService) validateRoutePosition(position deliverytypes.DeliveryRoutePosition) error {
	if position.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if position.RouteID == uuid.Nil {
		return fmt.Errorf("route_id is required")
	}
	if position.Latitude < -90 || position.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if position.Longitude < -180 || position.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

func (s *DeliveryTrackingService) validateRouteAssignment(assignment deliverytypes.DeliveryRouteAssignment) error {
	if assignment.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if assignment.RouteID == uuid.Nil {
		return fmt.Errorf("route_id is required")
	}
	return nil
}

func (s *DeliveryTrackingService) validateRouteStop(stop deliverytypes.DeliveryRouteStop) error {
	if stop.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if stop.RouteID == uuid.Nil {
		return fmt.Errorf("route_id is required")
	}
	if stop.StopSequence <= 0 {
		return fmt.Errorf("stop_sequence must be positive")
	}
	return nil
}

func (s *DeliveryTrackingService) publishShipmentEvent(ctx context.Context, eventType string, shipment deliverytypes.DeliveryShipment) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":                   shipment.ID,
			"organization_id":      shipment.OrganizationID,
			"picking_id":           shipment.PickingID,
			"route_id":             shipment.RouteID,
			"tracking_number":      shipment.TrackingNumber,
			"carrier_name":         shipment.CarrierName,
			"shipment_type":        shipment.ShipmentType,
			"status":               shipment.Status,
			"estimated_arrival_at": shipment.EstimatedArrivalAt,
			"arrived_at":           shipment.ArrivedAt,
			"metadata":             shipment.Metadata,
			"created_at":           shipment.CreatedAt,
			"updated_at":           shipment.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}

func (s *DeliveryTrackingService) publishTrackingEvent(ctx context.Context, eventType string, trackingEvent deliverytypes.DeliveryTrackingEvent) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":              trackingEvent.ID,
			"organization_id": trackingEvent.OrganizationID,
			"shipment_id":     trackingEvent.ShipmentID,
			"event_type":      trackingEvent.EventType,
			"status":          trackingEvent.Status,
			"event_time":      trackingEvent.EventTime,
			"source":          trackingEvent.Source,
			"message":         trackingEvent.Message,
			"latitude":        trackingEvent.Latitude,
			"longitude":       trackingEvent.Longitude,
			"raw_payload":     trackingEvent.RawPayload,
			"created_at":      trackingEvent.CreatedAt,
			"updated_at":      trackingEvent.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}

func (s *DeliveryTrackingService) publishRoutePositionEvent(ctx context.Context, eventType string, position deliverytypes.DeliveryRoutePosition) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":              position.ID,
			"organization_id": position.OrganizationID,
			"route_id":        position.RouteID,
			"vehicle_id":      position.VehicleID,
			"recorded_at":     position.RecordedAt,
			"latitude":        position.Latitude,
			"longitude":       position.Longitude,
			"speed_kph":       position.SpeedKPH,
			"heading":         position.Heading,
			"source":          position.Source,
			"metadata":        position.Metadata,
			"created_at":      position.CreatedAt,
			"updated_at":      position.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}

func (s *DeliveryTrackingService) publishRouteAssignmentEvent(ctx context.Context, eventType string, assignment deliverytypes.DeliveryRouteAssignment) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":                 assignment.ID,
			"organization_id":    assignment.OrganizationID,
			"route_id":           assignment.RouteID,
			"vehicle_id":         assignment.VehicleID,
			"driver_employee_id": assignment.DriverEmployeeID,
			"assignment_status":  assignment.AssignmentStatus,
			"assigned_at":        assignment.AssignedAt,
			"acknowledged_at":    assignment.AcknowledgedAt,
			"metadata":           assignment.Metadata,
			"created_at":         assignment.CreatedAt,
			"updated_at":         assignment.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}

func (s *DeliveryTrackingService) publishRouteStopEvent(ctx context.Context, eventType string, stop deliverytypes.DeliveryRouteStop) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":                 stop.ID,
			"organization_id":    stop.OrganizationID,
			"route_id":           stop.RouteID,
			"shipment_id":        stop.ShipmentID,
			"stop_sequence":      stop.StopSequence,
			"contact_id":         stop.ContactID,
			"location_id":        stop.LocationID,
			"status":             stop.Status,
			"planned_arrival_at": stop.PlannedArrivalAt,
			"actual_arrival_at":  stop.ActualArrivalAt,
			"metadata":           stop.Metadata,
			"created_at":         stop.CreatedAt,
			"updated_at":         stop.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}
