package service

import (
	"context"
	"fmt"

	deliveryrepository "alieze-erp/internal/modules/delivery/repository"
	deliverytypes "alieze-erp/internal/modules/delivery/types"
	"alieze-erp/pkg/events"

	"github.com/google/uuid"
)

type DeliveryVehicleService struct {
	repo     deliveryrepository.DeliveryVehicleRepository
	eventBus *events.Bus
}

func NewDeliveryVehicleService(repo deliveryrepository.DeliveryVehicleRepository) *DeliveryVehicleService {
	return &DeliveryVehicleService{
		repo: repo,
	}
}

func NewDeliveryVehicleServiceWithEventBus(repo deliveryrepository.DeliveryVehicleRepository, eventBus *events.Bus) *DeliveryVehicleService {
	service := NewDeliveryVehicleService(repo)
	service.eventBus = eventBus
	return service
}

func (s *DeliveryVehicleService) CreateDeliveryVehicle(ctx context.Context, vehicle deliverytypes.DeliveryVehicle) (*deliverytypes.DeliveryVehicle, error) {
	// Validate the vehicle
	if err := s.validateDeliveryVehicle(vehicle); err != nil {
		return nil, fmt.Errorf("invalid delivery vehicle: %w", err)
	}

	// Set default values
	if vehicle.ID == uuid.Nil {
		vehicle.ID = uuid.New()
	}
	if vehicle.VehicleType == "" {
		vehicle.VehicleType = deliverytypes.VehicleTypeTruck
	}
	if vehicle.Active == false {
		vehicle.Active = true // Default to active
	}
	if vehicle.Metadata == nil {
		vehicle.Metadata = make(map[string]interface{})
	}

	// Create the vehicle
	createdVehicle, err := s.repo.Create(ctx, vehicle)
	if err != nil {
		return nil, fmt.Errorf("failed to create delivery vehicle: %w", err)
	}

	// Publish event if event bus is available
	s.publishVehicleEvent(ctx, "delivery_vehicle.created", *createdVehicle)

	return createdVehicle, nil
}

func (s *DeliveryVehicleService) GetDeliveryVehicle(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryVehicle, error) {
	vehicle, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery vehicle: %w", err)
	}
	if vehicle == nil {
		return nil, nil
	}
	return vehicle, nil
}

func (s *DeliveryVehicleService) ListDeliveryVehicles(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]deliverytypes.DeliveryVehicle, error) {
	var vehicles []deliverytypes.DeliveryVehicle
	var err error

	if activeOnly {
		vehicles, err = s.repo.FindActiveByOrganizationID(ctx, orgID)
	} else {
		vehicles, err = s.repo.FindByOrganizationID(ctx, orgID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list delivery vehicles: %w", err)
	}

	return vehicles, nil
}

func (s *DeliveryVehicleService) UpdateDeliveryVehicle(ctx context.Context, vehicle deliverytypes.DeliveryVehicle) (*deliverytypes.DeliveryVehicle, error) {
	// Validate the vehicle
	if err := s.validateDeliveryVehicle(vehicle); err != nil {
		return nil, fmt.Errorf("invalid delivery vehicle: %w", err)
	}

	// Get existing vehicle to check if it exists
	existing, err := s.repo.FindByID(ctx, vehicle.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing vehicle: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("delivery vehicle not found")
	}

	// Update the vehicle
	updatedVehicle, err := s.repo.Update(ctx, vehicle)
	if err != nil {
		return nil, fmt.Errorf("failed to update delivery vehicle: %w", err)
	}

	// Publish event if event bus is available
	s.publishVehicleEvent(ctx, "delivery_vehicle.updated", *updatedVehicle)

	return updatedVehicle, nil
}

func (s *DeliveryVehicleService) DeleteDeliveryVehicle(ctx context.Context, id uuid.UUID) error {
	// Get existing vehicle to check if it exists
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find existing vehicle: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("delivery vehicle not found")
	}

	// Delete the vehicle
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery vehicle: %w", err)
	}

	// Publish event if event bus is available
	s.publishVehicleEvent(ctx, "delivery_vehicle.deleted", *existing)

	return nil
}

func (s *DeliveryVehicleService) validateDeliveryVehicle(vehicle deliverytypes.DeliveryVehicle) error {
	if vehicle.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if vehicle.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(vehicle.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}
	if vehicle.RegistrationNumber != "" && len(vehicle.RegistrationNumber) > 100 {
		return fmt.Errorf("registration_number must be 100 characters or less")
	}
	if vehicle.VehicleIdentifier != "" && len(vehicle.VehicleIdentifier) > 100 {
		return fmt.Errorf("vehicle_identifier must be 100 characters or less")
	}
	if vehicle.Capacity < 0 {
		return fmt.Errorf("capacity must be non-negative")
	}
	return nil
}

func (s *DeliveryVehicleService) publishVehicleEvent(ctx context.Context, eventType string, vehicle deliverytypes.DeliveryVehicle) {
	if s.eventBus == nil {
		return
	}

	if eventBus, ok := s.eventBus.(interface {
		Publish(eventType string, event interface{}) error
	}); ok {
		eventData := map[string]interface{}{
			"id":                  vehicle.ID,
			"organization_id":     vehicle.OrganizationID,
			"name":                vehicle.Name,
			"registration_number": vehicle.RegistrationNumber,
			"vehicle_type":        vehicle.VehicleType,
			"active":              vehicle.Active,
			"capacity":            vehicle.Capacity,
			"metadata":            vehicle.Metadata,
			"created_at":          vehicle.CreatedAt,
			"updated_at":          vehicle.UpdatedAt,
		}

		_ = eventBus.Publish(eventType, eventData)
	}
}
