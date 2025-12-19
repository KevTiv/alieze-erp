package repository_test

import (
	"context"
	"testing"

	deliverytypes "github.com/KevTiv/alieze-erp/internal/modules/delivery/types"
	deliveryrepository "github.com/KevTiv/alieze-erp/internal/modules/delivery/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockDB is a simple mock database for testing
type MockDB struct{}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) interface{} {
	return nil
}

func TestDeliveryVehicleRepository_Create(t *testing.T) {
	// This is a simple test to verify the repository can be instantiated
	// In a real scenario, you would use a test database

	mockDB := &MockDB{}
	repo := deliveryrepository.NewDeliveryVehicleRepository(nil)

	assert.NotNil(t, repo, "Repository should not be nil")
}

func TestDeliveryRouteRepository_Create(t *testing.T) {
	// Simple instantiation test
	repo := deliveryrepository.NewDeliveryRouteRepository(nil)

	assert.NotNil(t, repo, "Repository should not be nil")
}

func TestDeliveryTrackingRepository_Create(t *testing.T) {
	// Simple instantiation test
	repo := deliveryrepository.NewDeliveryTrackingRepository(nil)

	assert.NotNil(t, repo, "Repository should not be nil")
}

func TestDeliveryVehicleTypes(t *testing.T) {
	// Test that our types work correctly
	vehicle := deliverytypes.DeliveryVehicle{
		ID:               uuid.New(),
		OrganizationID:   uuid.New(),
		Name:             "Test Vehicle",
		VehicleType:      deliverytypes.VehicleTypeTruck,
		Active:           true,
		Metadata:         make(map[string]interface{}),
	}

	assert.NotEqual(t, uuid.Nil, vehicle.ID)
	assert.Equal(t, "Test Vehicle", vehicle.Name)
	assert.Equal(t, deliverytypes.VehicleTypeTruck, vehicle.VehicleType)
	assert.True(t, vehicle.Active)
}

func TestDeliveryRouteTypes(t *testing.T) {
	// Test that our types work correctly
	route := deliverytypes.DeliveryRoute{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "Test Route",
		TransportMode:  deliverytypes.TransportModeRoad,
		Status:         deliverytypes.RouteStatusDraft,
		Metadata:       make(map[string]interface{}),
	}

	assert.NotEqual(t, uuid.Nil, route.ID)
	assert.Equal(t, "Test Route", route.Name)
	assert.Equal(t, deliverytypes.TransportModeRoad, route.TransportMode)
	assert.Equal(t, deliverytypes.RouteStatusDraft, route.Status)
}
