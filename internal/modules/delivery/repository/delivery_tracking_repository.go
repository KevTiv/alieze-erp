package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	deliverytypes "github.com/KevTiv/alieze-erp/internal/modules/delivery/types"

	"github.com/google/uuid"
)

type DeliveryTrackingRepository interface {
	// Shipment operations
	CreateShipment(ctx context.Context, shipment deliverytypes.DeliveryShipment) (*deliverytypes.DeliveryShipment, error)
	FindShipmentByID(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryShipment, error)
	FindShipmentsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryShipment, error)
	FindShipmentsByPickingID(ctx context.Context, pickingID uuid.UUID) (*deliverytypes.DeliveryShipment, error)
	UpdateShipment(ctx context.Context, shipment deliverytypes.DeliveryShipment) (*deliverytypes.DeliveryShipment, error)

	// Tracking event operations
	CreateTrackingEvent(ctx context.Context, event deliverytypes.DeliveryTrackingEvent) (*deliverytypes.DeliveryTrackingEvent, error)
	FindTrackingEventsByShipmentID(ctx context.Context, shipmentID uuid.UUID) ([]deliverytypes.DeliveryTrackingEvent, error)
	FindLatestTrackingEventByShipmentID(ctx context.Context, shipmentID uuid.UUID) (*deliverytypes.DeliveryTrackingEvent, error)

	// Route position operations
	CreateRoutePosition(ctx context.Context, position deliverytypes.DeliveryRoutePosition) (*deliverytypes.DeliveryRoutePosition, error)
	FindRoutePositionsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRoutePosition, error)
	FindLatestRoutePositionByRouteID(ctx context.Context, routeID uuid.UUID) (*deliverytypes.DeliveryRoutePosition, error)

	// Route assignment operations
	CreateRouteAssignment(ctx context.Context, assignment deliverytypes.DeliveryRouteAssignment) (*deliverytypes.DeliveryRouteAssignment, error)
	FindRouteAssignmentsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRouteAssignment, error)

	// Route stop operations
	CreateRouteStop(ctx context.Context, stop deliverytypes.DeliveryRouteStop) (*deliverytypes.DeliveryRouteStop, error)
	FindRouteStopsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRouteStop, error)
	FindRouteStopByShipmentID(ctx context.Context, shipmentID uuid.UUID) (*deliverytypes.DeliveryRouteStop, error)
	UpdateRouteStop(ctx context.Context, stop deliverytypes.DeliveryRouteStop) (*deliverytypes.DeliveryRouteStop, error)
}

type deliveryTrackingRepository struct {
	db *sql.DB
}

func NewDeliveryTrackingRepository(db *sql.DB) DeliveryTrackingRepository {
	return &deliveryTrackingRepository{db: db}
}

func (r *deliveryTrackingRepository) CreateShipment(ctx context.Context, shipment deliverytypes.DeliveryShipment) (*deliverytypes.DeliveryShipment, error) {
	query := `
		INSERT INTO delivery_shipments (
			organization_id, company_id, picking_id, route_id, assignment_id,
			tracking_number, carrier_name, carrier_code, carrier_service_level, shipment_type,
			status, requires_signature, estimated_departure_at, estimated_arrival_at,
			metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		shipment.OrganizationID,
		shipment.CompanyID,
		shipment.PickingID,
		shipment.RouteID,
		shipment.AssignmentID,
		shipment.TrackingNumber,
		shipment.CarrierName,
		shipment.CarrierCode,
		shipment.CarrierServiceLevel,
		shipment.ShipmentType,
		shipment.Status,
		shipment.RequiresSignature,
		shipment.EstimatedDepartureAt,
		shipment.EstimatedArrivalAt,
		shipment.Metadata,
	).Scan(&shipment.ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery shipment: %w", err)
	}

	shipment.CreatedAt = createdAt
	shipment.UpdatedAt = updatedAt
	return &shipment, nil
}

func (r *deliveryTrackingRepository) FindShipmentByID(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryShipment, error) {
	query := `
		SELECT
			id, organization_id, company_id, picking_id, route_id, assignment_id,
			tracking_number, carrier_name, carrier_code, carrier_service_level, shipment_type,
			status, requires_signature, estimated_departure_at, estimated_arrival_at,
			departed_at, arrived_at, last_event_at, last_latitude, last_longitude,
			metadata, created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_shipments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var shipment deliverytypes.DeliveryShipment
	var companyID, routeID, assignmentID, createdBy, updatedBy sql.NullString
	var estimatedDepartureAt, estimatedArrivalAt, departedAt, arrivedAt, lastEventAt, deletedAt sql.NullTime
	var lastLatitude, lastLongitude sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shipment.ID,
		&shipment.OrganizationID,
		&companyID,
		&shipment.PickingID,
		&routeID,
		&assignmentID,
		&shipment.TrackingNumber,
		&shipment.CarrierName,
		&shipment.CarrierCode,
		&shipment.CarrierServiceLevel,
		&shipment.ShipmentType,
		&shipment.Status,
		&shipment.RequiresSignature,
		&estimatedDepartureAt,
		&estimatedArrivalAt,
		&departedAt,
		&arrivedAt,
		&lastEventAt,
		&lastLatitude,
		&lastLongitude,
		&shipment.Metadata,
		&shipment.CreatedAt,
		&shipment.UpdatedAt,
		&createdBy,
		&updatedBy,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find delivery shipment: %w", err)
	}

	if companyID.Valid {
		parsedID, err := uuid.Parse(companyID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid company_id: %w", err)
		}
		shipment.CompanyID = &parsedID
	}

	if routeID.Valid {
		parsedID, err := uuid.Parse(routeID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid route_id: %w", err)
		}
		shipment.RouteID = &parsedID
	}

	if assignmentID.Valid {
		parsedID, err := uuid.Parse(assignmentID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid assignment_id: %w", err)
		}
		shipment.AssignmentID = &parsedID
	}

	if estimatedDepartureAt.Valid {
		time := estimatedDepartureAt.Time
		shipment.EstimatedDepartureAt = &time
	}

	if estimatedArrivalAt.Valid {
		time := estimatedArrivalAt.Time
		shipment.EstimatedArrivalAt = &time
	}

	if departedAt.Valid {
		time := departedAt.Time
		shipment.DepartedAt = &time
	}

	if arrivedAt.Valid {
		time := arrivedAt.Time
		shipment.ArrivedAt = &time
	}

	if lastEventAt.Valid {
		time := lastEventAt.Time
		shipment.LastEventAt = &time
	}

	if lastLatitude.Valid {
		lat := lastLatitude.Float64
		shipment.LastLatitude = &lat
	}

	if lastLongitude.Valid {
		lon := lastLongitude.Float64
		shipment.LastLongitude = &lon
	}

	if createdBy.Valid {
		parsedID, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		shipment.CreatedBy = &parsedID
	}

	if updatedBy.Valid {
		parsedID, err := uuid.Parse(updatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by: %w", err)
		}
		shipment.UpdatedBy = &parsedID
	}

	if deletedAt.Valid {
		time := deletedAt.Time
		shipment.DeletedAt = &time
	}

	return &shipment, nil
}

func (r *deliveryTrackingRepository) FindShipmentsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryShipment, error) {
	query := `
		SELECT
			id, organization_id, company_id, picking_id, route_id, assignment_id,
			tracking_number, carrier_name, carrier_code, carrier_service_level, shipment_type,
			status, requires_signature, estimated_departure_at, estimated_arrival_at,
			departed_at, arrived_at, last_event_at, last_latitude, last_longitude,
			metadata, created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_shipments
		WHERE route_id = $1 AND deleted_at IS NULL
		ORDER BY estimated_departure_at
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery shipments: %w", err)
	}
	defer rows.Close()

	var shipments []deliverytypes.DeliveryShipment
	for rows.Next() {
		var shipment deliverytypes.DeliveryShipment
		var companyID, assignmentID, createdBy, updatedBy sql.NullString
		var estimatedDepartureAt, estimatedArrivalAt, departedAt, arrivedAt, lastEventAt, deletedAt sql.NullTime
		var lastLatitude, lastLongitude sql.NullFloat64

		err := rows.Scan(
			&shipment.ID,
			&shipment.OrganizationID,
			&companyID,
			&shipment.PickingID,
			&shipment.RouteID,
			&assignmentID,
			&shipment.TrackingNumber,
			&shipment.CarrierName,
			&shipment.CarrierCode,
			&shipment.CarrierServiceLevel,
			&shipment.ShipmentType,
			&shipment.Status,
			&shipment.RequiresSignature,
			&estimatedDepartureAt,
			&estimatedArrivalAt,
			&departedAt,
			&arrivedAt,
			&lastEventAt,
			&lastLatitude,
			&lastLongitude,
			&shipment.Metadata,
			&shipment.CreatedAt,
			&shipment.UpdatedAt,
			&createdBy,
			&updatedBy,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery shipment: %w", err)
		}

		if companyID.Valid {
			parsedID, err := uuid.Parse(companyID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid company_id: %w", err)
			}
			shipment.CompanyID = &parsedID
		}

		if assignmentID.Valid {
			parsedID, err := uuid.Parse(assignmentID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid assignment_id: %w", err)
			}
			shipment.AssignmentID = &parsedID
		}

		if estimatedDepartureAt.Valid {
			time := estimatedDepartureAt.Time
			shipment.EstimatedDepartureAt = &time
		}

		if estimatedArrivalAt.Valid {
			time := estimatedArrivalAt.Time
			shipment.EstimatedArrivalAt = &time
		}

		if departedAt.Valid {
			time := departedAt.Time
			shipment.DepartedAt = &time
		}

		if arrivedAt.Valid {
			time := arrivedAt.Time
			shipment.ArrivedAt = &time
		}

		if lastEventAt.Valid {
			time := lastEventAt.Time
			shipment.LastEventAt = &time
		}

		if lastLatitude.Valid {
			lat := lastLatitude.Float64
			shipment.LastLatitude = &lat
		}

		if lastLongitude.Valid {
			lon := lastLongitude.Float64
			shipment.LastLongitude = &lon
		}

		if createdBy.Valid {
			parsedID, err := uuid.Parse(createdBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid created_by: %w", err)
			}
			shipment.CreatedBy = &parsedID
		}

		if updatedBy.Valid {
			parsedID, err := uuid.Parse(updatedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid updated_by: %w", err)
			}
			shipment.UpdatedBy = &parsedID
		}

		if deletedAt.Valid {
			time := deletedAt.Time
			shipment.DeletedAt = &time
		}

		shipments = append(shipments, shipment)
	}

	return shipments, nil
}

func (r *deliveryTrackingRepository) FindShipmentsByPickingID(ctx context.Context, pickingID uuid.UUID) (*deliverytypes.DeliveryShipment, error) {
	query := `
		SELECT
			id, organization_id, company_id, picking_id, route_id, assignment_id,
			tracking_number, carrier_name, carrier_code, carrier_service_level, shipment_type,
			status, requires_signature, estimated_departure_at, estimated_arrival_at,
			departed_at, arrived_at, last_event_at, last_latitude, last_longitude,
			metadata, created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_shipments
		WHERE picking_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var shipment deliverytypes.DeliveryShipment
	var companyID, routeID, assignmentID, createdBy, updatedBy sql.NullString
	var estimatedDepartureAt, estimatedArrivalAt, departedAt, arrivedAt, lastEventAt, deletedAt sql.NullTime
	var lastLatitude, lastLongitude sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, pickingID).Scan(
		&shipment.ID,
		&shipment.OrganizationID,
		&companyID,
		&shipment.PickingID,
		&routeID,
		&assignmentID,
		&shipment.TrackingNumber,
		&shipment.CarrierName,
		&shipment.CarrierCode,
		&shipment.CarrierServiceLevel,
		&shipment.ShipmentType,
		&shipment.Status,
		&shipment.RequiresSignature,
		&estimatedDepartureAt,
		&estimatedArrivalAt,
		&departedAt,
		&arrivedAt,
		&lastEventAt,
		&lastLatitude,
		&lastLongitude,
		&shipment.Metadata,
		&shipment.CreatedAt,
		&shipment.UpdatedAt,
		&createdBy,
		&updatedBy,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find delivery shipment by picking: %w", err)
	}

	if companyID.Valid {
		parsedID, err := uuid.Parse(companyID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid company_id: %w", err)
		}
		shipment.CompanyID = &parsedID
	}

	if routeID.Valid {
		parsedID, err := uuid.Parse(routeID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid route_id: %w", err)
		}
		shipment.RouteID = &parsedID
	}

	if assignmentID.Valid {
		parsedID, err := uuid.Parse(assignmentID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid assignment_id: %w", err)
		}
		shipment.AssignmentID = &parsedID
	}

	if estimatedDepartureAt.Valid {
		time := estimatedDepartureAt.Time
		shipment.EstimatedDepartureAt = &time
	}

	if estimatedArrivalAt.Valid {
		time := estimatedArrivalAt.Time
		shipment.EstimatedArrivalAt = &time
	}

	if departedAt.Valid {
		time := departedAt.Time
		shipment.DepartedAt = &time
	}

	if arrivedAt.Valid {
		time := arrivedAt.Time
		shipment.ArrivedAt = &time
	}

	if lastEventAt.Valid {
		time := lastEventAt.Time
		shipment.LastEventAt = &time
	}

	if lastLatitude.Valid {
		lat := lastLatitude.Float64
		shipment.LastLatitude = &lat
	}

	if lastLongitude.Valid {
		lon := lastLongitude.Float64
		shipment.LastLongitude = &lon
	}

	if createdBy.Valid {
		parsedID, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		shipment.CreatedBy = &parsedID
	}

	if updatedBy.Valid {
		parsedID, err := uuid.Parse(updatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by: %w", err)
		}
		shipment.UpdatedBy = &parsedID
	}

	if deletedAt.Valid {
		time := deletedAt.Time
		shipment.DeletedAt = &time
	}

	return &shipment, nil
}

func (r *deliveryTrackingRepository) UpdateShipment(ctx context.Context, shipment deliverytypes.DeliveryShipment) (*deliverytypes.DeliveryShipment, error) {
	query := `
		UPDATE delivery_shipments SET
			tracking_number = $1,
			carrier_name = $2,
			carrier_code = $3,
			carrier_service_level = $4,
			status = $5,
			requires_signature = $6,
			estimated_departure_at = $7,
			estimated_arrival_at = $8,
			departed_at = $9,
			arrived_at = $10,
			last_event_at = $11,
			last_latitude = $12,
			last_longitude = $13,
			metadata = $14,
			updated_at = NOW()
		WHERE id = $15 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		shipment.TrackingNumber,
		shipment.CarrierName,
		shipment.CarrierCode,
		shipment.CarrierServiceLevel,
		shipment.Status,
		shipment.RequiresSignature,
		shipment.EstimatedDepartureAt,
		shipment.EstimatedArrivalAt,
		shipment.DepartedAt,
		shipment.ArrivedAt,
		shipment.LastEventAt,
		shipment.LastLatitude,
		shipment.LastLongitude,
		shipment.Metadata,
		shipment.ID,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update delivery shipment: %w", err)
	}

	shipment.UpdatedAt = updatedAt
	return &shipment, nil
}

func (r *deliveryTrackingRepository) CreateTrackingEvent(ctx context.Context, event deliverytypes.DeliveryTrackingEvent) (*deliverytypes.DeliveryTrackingEvent, error) {
	query := `
		INSERT INTO delivery_tracking_events (
			organization_id, shipment_id, stop_id, event_type, status,
			event_time, source, message, raw_payload, latitude, longitude,
			altitude, speed_kph, heading
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		event.OrganizationID,
		event.ShipmentID,
		event.StopID,
		event.EventType,
		event.Status,
		event.EventTime,
		event.Source,
		event.Message,
		event.RawPayload,
		event.Latitude,
		event.Longitude,
		event.Altitude,
		event.SpeedKPH,
		event.Heading,
	).Scan(&event.ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery tracking event: %w", err)
	}

	event.CreatedAt = createdAt
	event.UpdatedAt = updatedAt
	return &event, nil
}

func (r *deliveryTrackingRepository) FindTrackingEventsByShipmentID(ctx context.Context, shipmentID uuid.UUID) ([]deliverytypes.DeliveryTrackingEvent, error) {
	query := `
		SELECT
			id, organization_id, shipment_id, stop_id, event_type, status,
			event_time, source, message, raw_payload, latitude, longitude,
			altitude, speed_kph, heading, created_at, updated_at, created_by, updated_by
		FROM delivery_tracking_events
		WHERE shipment_id = $1
		ORDER BY event_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery tracking events: %w", err)
	}
	defer rows.Close()

	var events []deliverytypes.DeliveryTrackingEvent
	for rows.Next() {
		var event deliverytypes.DeliveryTrackingEvent
		var stopID, createdBy, updatedBy sql.NullString
		var latitude, longitude, altitude, speedKPH, heading sql.NullFloat64

		err := rows.Scan(
			&event.ID,
			&event.OrganizationID,
			&event.ShipmentID,
			&stopID,
			&event.EventType,
			&event.Status,
			&event.EventTime,
			&event.Source,
			&event.Message,
			&event.RawPayload,
			&latitude,
			&longitude,
			&altitude,
			&speedKPH,
			&heading,
			&event.CreatedAt,
			&event.UpdatedAt,
			&createdBy,
			&updatedBy,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery tracking event: %w", err)
		}

		if stopID.Valid {
			parsedID, err := uuid.Parse(stopID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid stop_id: %w", err)
			}
			event.StopID = &parsedID
		}

		if latitude.Valid {
			lat := latitude.Float64
			event.Latitude = &lat
		}

		if longitude.Valid {
			lon := longitude.Float64
			event.Longitude = &lon
		}

		if altitude.Valid {
			alt := altitude.Float64
			event.Altitude = &alt
		}

		if speedKPH.Valid {
			speed := speedKPH.Float64
			event.SpeedKPH = &speed
		}

		if heading.Valid {
			head := heading.Float64
			event.Heading = &head
		}

		if createdBy.Valid {
			parsedID, err := uuid.Parse(createdBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid created_by: %w", err)
			}
			event.CreatedBy = &parsedID
		}

		if updatedBy.Valid {
			parsedID, err := uuid.Parse(updatedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid updated_by: %w", err)
			}
			event.UpdatedBy = &parsedID
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *deliveryTrackingRepository) FindLatestTrackingEventByShipmentID(ctx context.Context, shipmentID uuid.UUID) (*deliverytypes.DeliveryTrackingEvent, error) {
	query := `
		SELECT
			id, organization_id, shipment_id, stop_id, event_type, status,
			event_time, source, message, raw_payload, latitude, longitude,
			altitude, speed_kph, heading, created_at, updated_at, created_by, updated_by
		FROM delivery_tracking_events
		WHERE shipment_id = $1
		ORDER BY event_time DESC
		LIMIT 1
	`

	var event deliverytypes.DeliveryTrackingEvent
	var stopID, createdBy, updatedBy sql.NullString
	var latitude, longitude, altitude, speedKPH, heading sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, shipmentID).Scan(
		&event.ID,
		&event.OrganizationID,
		&event.ShipmentID,
		&stopID,
		&event.EventType,
		&event.Status,
		&event.EventTime,
		&event.Source,
		&event.Message,
		&event.RawPayload,
		&latitude,
		&longitude,
		&altitude,
		&speedKPH,
		&heading,
		&event.CreatedAt,
		&event.UpdatedAt,
		&createdBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest delivery tracking event: %w", err)
	}

	if stopID.Valid {
		parsedID, err := uuid.Parse(stopID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid stop_id: %w", err)
		}
		event.StopID = &parsedID
	}

	if latitude.Valid {
		lat := latitude.Float64
		event.Latitude = &lat
	}

	if longitude.Valid {
		lon := longitude.Float64
		event.Longitude = &lon
	}

	if altitude.Valid {
		alt := altitude.Float64
		event.Altitude = &alt
	}

	if speedKPH.Valid {
		speed := speedKPH.Float64
		event.SpeedKPH = &speed
	}

	if heading.Valid {
		head := heading.Float64
		event.Heading = &head
	}

	if createdBy.Valid {
		parsedID, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		event.CreatedBy = &parsedID
	}

	if updatedBy.Valid {
		parsedID, err := uuid.Parse(updatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by: %w", err)
		}
		event.UpdatedBy = &parsedID
	}

	return &event, nil
}

func (r *deliveryTrackingRepository) CreateRoutePosition(ctx context.Context, position deliverytypes.DeliveryRoutePosition) (*deliverytypes.DeliveryRoutePosition, error) {
	query := `
		INSERT INTO delivery_route_positions (
			organization_id, route_id, assignment_id, vehicle_id,
			recorded_at, latitude, longitude, altitude, speed_kph, heading,
			source, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		position.OrganizationID,
		position.RouteID,
		position.AssignmentID,
		position.VehicleID,
		position.RecordedAt,
		position.Latitude,
		position.Longitude,
		position.Altitude,
		position.SpeedKPH,
		position.Heading,
		position.Source,
		position.Metadata,
	).Scan(&position.ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery route position: %w", err)
	}

	position.CreatedAt = createdAt
	position.UpdatedAt = updatedAt
	return &position, nil
}

func (r *deliveryTrackingRepository) FindRoutePositionsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRoutePosition, error) {
	query := `
		SELECT
			id, organization_id, route_id, assignment_id, vehicle_id,
			recorded_at, latitude, longitude, altitude, speed_kph, heading,
			source, metadata, created_at, updated_at
		FROM delivery_route_positions
		WHERE route_id = $1
		ORDER BY recorded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery route positions: %w", err)
	}
	defer rows.Close()

	var positions []deliverytypes.DeliveryRoutePosition
	for rows.Next() {
		var position deliverytypes.DeliveryRoutePosition
		var assignmentID, vehicleID sql.NullString
		var altitude, speedKPH, heading sql.NullFloat64

		err := rows.Scan(
			&position.ID,
			&position.OrganizationID,
			&position.RouteID,
			&assignmentID,
			&vehicleID,
			&position.RecordedAt,
			&position.Latitude,
			&position.Longitude,
			&altitude,
			&speedKPH,
			&heading,
			&position.Source,
			&position.Metadata,
			&position.CreatedAt,
			&position.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery route position: %w", err)
		}

		if assignmentID.Valid {
			parsedID, err := uuid.Parse(assignmentID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid assignment_id: %w", err)
			}
			position.AssignmentID = &parsedID
		}

		if vehicleID.Valid {
			parsedID, err := uuid.Parse(vehicleID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid vehicle_id: %w", err)
			}
			position.VehicleID = &parsedID
		}

		if altitude.Valid {
			alt := altitude.Float64
			position.Altitude = &alt
		}

		if speedKPH.Valid {
			speed := speedKPH.Float64
			position.SpeedKPH = &speed
		}

		if heading.Valid {
			head := heading.Float64
			position.Heading = &head
		}

		positions = append(positions, position)
	}

	return positions, nil
}

func (r *deliveryTrackingRepository) FindLatestRoutePositionByRouteID(ctx context.Context, routeID uuid.UUID) (*deliverytypes.DeliveryRoutePosition, error) {
	query := `
		SELECT
			id, organization_id, route_id, assignment_id, vehicle_id,
			recorded_at, latitude, longitude, altitude, speed_kph, heading,
			source, metadata, created_at, updated_at
		FROM delivery_route_positions
		WHERE route_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`

	var position deliverytypes.DeliveryRoutePosition
	var assignmentID, vehicleID sql.NullString
	var altitude, speedKPH, heading sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, routeID).Scan(
		&position.ID,
		&position.OrganizationID,
		&position.RouteID,
		&assignmentID,
		&vehicleID,
		&position.RecordedAt,
		&position.Latitude,
		&position.Longitude,
		&altitude,
		&speedKPH,
		&heading,
		&position.Source,
		&position.Metadata,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest delivery route position: %w", err)
	}

	if assignmentID.Valid {
		parsedID, err := uuid.Parse(assignmentID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid assignment_id: %w", err)
		}
		position.AssignmentID = &parsedID
	}

	if vehicleID.Valid {
		parsedID, err := uuid.Parse(vehicleID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid vehicle_id: %w", err)
		}
		position.VehicleID = &parsedID
	}

	if altitude.Valid {
		alt := altitude.Float64
		position.Altitude = &alt
	}

	if speedKPH.Valid {
		speed := speedKPH.Float64
		position.SpeedKPH = &speed
	}

	if heading.Valid {
		head := heading.Float64
		position.Heading = &head
	}

	return &position, nil
}

func (r *deliveryTrackingRepository) CreateRouteAssignment(ctx context.Context, assignment deliverytypes.DeliveryRouteAssignment) (*deliverytypes.DeliveryRouteAssignment, error) {
	query := `
		INSERT INTO delivery_route_assignments (
			organization_id, route_id, vehicle_id, driver_employee_id, driver_contact_id,
			assignment_status, acknowledged_at, released_at, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, assigned_at, created_at, updated_at
	`

	var assignedAt, createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		assignment.OrganizationID,
		assignment.RouteID,
		assignment.VehicleID,
		assignment.DriverEmployeeID,
		assignment.DriverContactID,
		assignment.AssignmentStatus,
		assignment.AcknowledgedAt,
		assignment.ReleasedAt,
		assignment.Metadata,
	).Scan(&assignment.ID, &assignedAt, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery route assignment: %w", err)
	}

	assignment.AssignedAt = assignedAt
	assignment.CreatedAt = createdAt
	assignment.UpdatedAt = updatedAt
	return &assignment, nil
}

func (r *deliveryTrackingRepository) FindRouteAssignmentsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRouteAssignment, error) {
	query := `
		SELECT
			id, organization_id, route_id, vehicle_id, driver_employee_id, driver_contact_id,
			assignment_status, assigned_at, acknowledged_at, released_at, metadata,
			created_at, updated_at, created_by, updated_by
		FROM delivery_route_assignments
		WHERE route_id = $1
		ORDER BY assigned_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery route assignments: %w", err)
	}
	defer rows.Close()

	var assignments []deliverytypes.DeliveryRouteAssignment
	for rows.Next() {
		var assignment deliverytypes.DeliveryRouteAssignment
		var vehicleID, driverEmployeeID, driverContactID, createdBy, updatedBy sql.NullString
		var acknowledgedAt, releasedAt sql.NullTime

		err := rows.Scan(
			&assignment.ID,
			&assignment.OrganizationID,
			&assignment.RouteID,
			&vehicleID,
			&driverEmployeeID,
			&driverContactID,
			&assignment.AssignmentStatus,
			&assignment.AssignedAt,
			&acknowledgedAt,
			&releasedAt,
			&assignment.Metadata,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
			&createdBy,
			&updatedBy,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery route assignment: %w", err)
		}

		if vehicleID.Valid {
			parsedID, err := uuid.Parse(vehicleID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid vehicle_id: %w", err)
			}
			assignment.VehicleID = &parsedID
		}

		if driverEmployeeID.Valid {
			parsedID, err := uuid.Parse(driverEmployeeID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid driver_employee_id: %w", err)
			}
			assignment.DriverEmployeeID = &parsedID
		}

		if driverContactID.Valid {
			parsedID, err := uuid.Parse(driverContactID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid driver_contact_id: %w", err)
			}
			assignment.DriverContactID = &parsedID
		}

		if acknowledgedAt.Valid {
			assignment.AcknowledgedAt = &acknowledgedAt.Time
		}

		if releasedAt.Valid {
			assignment.ReleasedAt = &releasedAt.Time
		}

		if createdBy.Valid {
			parsedID, err := uuid.Parse(createdBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid created_by: %w", err)
			}
			assignment.CreatedBy = &parsedID
		}

		if updatedBy.Valid {
			parsedID, err := uuid.Parse(updatedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid updated_by: %w", err)
			}
			assignment.UpdatedBy = &parsedID
		}

		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

func (r *deliveryTrackingRepository) CreateRouteStop(ctx context.Context, stop deliverytypes.DeliveryRouteStop) (*deliverytypes.DeliveryRouteStop, error) {
	query := `
		INSERT INTO delivery_route_stops (
			organization_id, route_id, assignment_id, shipment_id, stop_sequence,
			contact_id, location_id, address, planned_arrival_at, planned_departure_at,
			status, notes, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		stop.OrganizationID,
		stop.RouteID,
		stop.AssignmentID,
		stop.ShipmentID,
		stop.StopSequence,
		stop.ContactID,
		stop.LocationID,
		stop.Address,
		stop.PlannedArrivalAt,
		stop.PlannedDepartureAt,
		stop.Status,
		stop.Notes,
		stop.Metadata,
	).Scan(&stop.ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery route stop: %w", err)
	}

	stop.CreatedAt = createdAt
	stop.UpdatedAt = updatedAt
	return &stop, nil
}

func (r *deliveryTrackingRepository) FindRouteStopsByRouteID(ctx context.Context, routeID uuid.UUID) ([]deliverytypes.DeliveryRouteStop, error) {
	query := `
		SELECT
			id, organization_id, route_id, assignment_id, shipment_id, stop_sequence,
			contact_id, location_id, address, planned_arrival_at, planned_departure_at,
			actual_arrival_at, actual_departure_at, status, notes, metadata,
			created_at, updated_at, created_by, updated_by
		FROM delivery_route_stops
		WHERE route_id = $1
		ORDER BY stop_sequence
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery route stops: %w", err)
	}
	defer rows.Close()

	var stops []deliverytypes.DeliveryRouteStop
	for rows.Next() {
		var stop deliverytypes.DeliveryRouteStop
		var assignmentID, shipmentID, contactID, locationID, createdBy, updatedBy sql.NullString
		var plannedArrivalAt, plannedDepartureAt, actualArrivalAt, actualDepartureAt sql.NullTime

		err := rows.Scan(
			&stop.ID,
			&stop.OrganizationID,
			&stop.RouteID,
			&assignmentID,
			&shipmentID,
			&stop.StopSequence,
			&contactID,
			&locationID,
			&stop.Address,
			&plannedArrivalAt,
			&plannedDepartureAt,
			&actualArrivalAt,
			&actualDepartureAt,
			&stop.Status,
			&stop.Notes,
			&stop.Metadata,
			&stop.CreatedAt,
			&stop.UpdatedAt,
			&createdBy,
			&updatedBy,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery route stop: %w", err)
		}

		if assignmentID.Valid {
			parsedID, err := uuid.Parse(assignmentID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid assignment_id: %w", err)
			}
			stop.AssignmentID = &parsedID
		}

		if shipmentID.Valid {
			parsedID, err := uuid.Parse(shipmentID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid shipment_id: %w", err)
			}
			stop.ShipmentID = &parsedID
		}

		if contactID.Valid {
			parsedID, err := uuid.Parse(contactID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid contact_id: %w", err)
			}
			stop.ContactID = &parsedID
		}

		if locationID.Valid {
			parsedID, err := uuid.Parse(locationID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid location_id: %w", err)
			}
			stop.LocationID = &parsedID
		}

		if plannedArrivalAt.Valid {
			time := plannedArrivalAt.Time
			stop.PlannedArrivalAt = &time
		}

		if plannedDepartureAt.Valid {
			time := plannedDepartureAt.Time
			stop.PlannedDepartureAt = &time
		}

		if actualArrivalAt.Valid {
			time := actualArrivalAt.Time
			stop.ActualArrivalAt = &time
		}

		if actualDepartureAt.Valid {
			time := actualDepartureAt.Time
			stop.ActualDepartureAt = &time
		}

		if createdBy.Valid {
			parsedID, err := uuid.Parse(createdBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid created_by: %w", err)
			}
			stop.CreatedBy = &parsedID
		}

		if updatedBy.Valid {
			parsedID, err := uuid.Parse(updatedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid updated_by: %w", err)
			}
			stop.UpdatedBy = &parsedID
		}

		stops = append(stops, stop)
	}

	return stops, nil
}

func (r *deliveryTrackingRepository) FindRouteStopByShipmentID(ctx context.Context, shipmentID uuid.UUID) (*deliverytypes.DeliveryRouteStop, error) {
	query := `
		SELECT
			id, organization_id, route_id, assignment_id, shipment_id, stop_sequence,
			contact_id, location_id, address, planned_arrival_at, planned_departure_at,
			actual_arrival_at, actual_departure_at, status, notes, metadata,
			created_at, updated_at, created_by, updated_by
		FROM delivery_route_stops
		WHERE shipment_id = $1
		LIMIT 1
	`

	var stop deliverytypes.DeliveryRouteStop
	var assignmentID, contactID, locationID, createdBy, updatedBy sql.NullString
	var plannedArrivalAt, plannedDepartureAt, actualArrivalAt, actualDepartureAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, shipmentID).Scan(
		&stop.ID,
		&stop.OrganizationID,
		&stop.RouteID,
		&assignmentID,
		&stop.ShipmentID,
		&stop.StopSequence,
		&contactID,
		&locationID,
		&stop.Address,
		&plannedArrivalAt,
		&plannedDepartureAt,
		&actualArrivalAt,
		&actualDepartureAt,
		&stop.Status,
		&stop.Notes,
		&stop.Metadata,
		&stop.CreatedAt,
		&stop.UpdatedAt,
		&createdBy,
		&updatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find delivery route stop by shipment: %w", err)
	}

	if assignmentID.Valid {
		parsedID, err := uuid.Parse(assignmentID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid assignment_id: %w", err)
		}
		stop.AssignmentID = &parsedID
	}

	if contactID.Valid {
		parsedID, err := uuid.Parse(contactID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid contact_id: %w", err)
		}
		stop.ContactID = &parsedID
	}

	if locationID.Valid {
		parsedID, err := uuid.Parse(locationID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid location_id: %w", err)
		}
		stop.LocationID = &parsedID
	}

	if plannedArrivalAt.Valid {
		time := plannedArrivalAt.Time
		stop.PlannedArrivalAt = &time
	}

	if plannedDepartureAt.Valid {
		time := plannedDepartureAt.Time
		stop.PlannedDepartureAt = &time
	}

	if actualArrivalAt.Valid {
		time := actualArrivalAt.Time
		stop.ActualArrivalAt = &time
	}

	if actualDepartureAt.Valid {
		time := actualDepartureAt.Time
		stop.ActualDepartureAt = &time
	}

	if createdBy.Valid {
		parsedID, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		stop.CreatedBy = &parsedID
	}

	if updatedBy.Valid {
		parsedID, err := uuid.Parse(updatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by: %w", err)
		}
		stop.UpdatedBy = &parsedID
	}

	return &stop, nil
}

func (r *deliveryTrackingRepository) UpdateRouteStop(ctx context.Context, stop deliverytypes.DeliveryRouteStop) (*deliverytypes.DeliveryRouteStop, error) {
	query := `
		UPDATE delivery_route_stops SET
			assignment_id = $1,
			contact_id = $2,
			location_id = $3,
			address = $4,
			planned_arrival_at = $5,
			planned_departure_at = $6,
			actual_arrival_at = $7,
			actual_departure_at = $8,
			status = $9,
			notes = $10,
			metadata = $11,
			updated_at = NOW()
		WHERE id = $12
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		stop.AssignmentID,
		stop.ContactID,
		stop.LocationID,
		stop.Address,
		stop.PlannedArrivalAt,
		stop.PlannedDepartureAt,
		stop.ActualArrivalAt,
		stop.ActualDepartureAt,
		stop.Status,
		stop.Notes,
		stop.Metadata,
		stop.ID,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update delivery route stop: %w", err)
	}

	stop.UpdatedAt = updatedAt
	return &stop, nil
}
