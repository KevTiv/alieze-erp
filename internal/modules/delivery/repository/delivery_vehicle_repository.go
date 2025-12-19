package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	deliverytypes "alieze-erp/internal/modules/delivery/types"

	"github.com/google/uuid"
)

type DeliveryVehicleRepository interface {
	Create(ctx context.Context, vehicle deliverytypes.DeliveryVehicle) (*deliverytypes.DeliveryVehicle, error)
	FindByID(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryVehicle, error)
	FindAll(ctx context.Context, filters DeliveryVehicleFilter) ([]deliverytypes.DeliveryVehicle, error)
	Update(ctx context.Context, vehicle deliverytypes.DeliveryVehicle) (*deliverytypes.DeliveryVehicle, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]deliverytypes.DeliveryVehicle, error)
	FindActiveByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]deliverytypes.DeliveryVehicle, error)
}

type DeliveryVehicleFilter struct {
	OrganizationID *uuid.UUID
	Active         *bool
	VehicleType    *deliverytypes.VehicleType
	Limit          int
	Offset         int
}

type deliveryVehicleRepository struct {
	db *sql.DB
}

func NewDeliveryVehicleRepository(db *sql.DB) DeliveryVehicleRepository {
	return &deliveryVehicleRepository{db: db}
}

func (r *deliveryVehicleRepository) Create(ctx context.Context, vehicle deliverytypes.DeliveryVehicle) (*deliverytypes.DeliveryVehicle, error) {
	query := `
		INSERT INTO delivery_vehicles (
			organization_id, name, registration_number, vehicle_identifier, vehicle_type,
			capacity, capacity_uom_id, active, last_service_at, service_interval_days, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		vehicle.OrganizationID,
		vehicle.Name,
		vehicle.RegistrationNumber,
		vehicle.VehicleIdentifier,
		vehicle.VehicleType,
		vehicle.Capacity,
		vehicle.CapacityUOMID,
		vehicle.Active,
		vehicle.LastServiceAt,
		vehicle.ServiceIntervalDays,
		vehicle.Metadata,
	).Scan(&vehicle.ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery vehicle: %w", err)
	}

	vehicle.CreatedAt = createdAt
	vehicle.UpdatedAt = updatedAt
	return &vehicle, nil
}

func (r *deliveryVehicleRepository) FindByID(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryVehicle, error) {
	query := `
		SELECT
			id, organization_id, name, registration_number, vehicle_identifier, vehicle_type,
			capacity, capacity_uom_id, active, last_service_at, service_interval_days, metadata,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_vehicles
		WHERE id = $1 AND deleted_at IS NULL
	`

	var vehicle deliverytypes.DeliveryVehicle
	var lastServiceAt, deletedAt sql.NullTime
	var capacityUOMID, createdBy, updatedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&vehicle.ID,
		&vehicle.OrganizationID,
		&vehicle.Name,
		&vehicle.RegistrationNumber,
		&vehicle.VehicleIdentifier,
		&vehicle.VehicleType,
		&vehicle.Capacity,
		&capacityUOMID,
		&vehicle.Active,
		&lastServiceAt,
		&vehicle.ServiceIntervalDays,
		&vehicle.Metadata,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
		&createdBy,
		&updatedBy,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find delivery vehicle: %w", err)
	}

	if capacityUOMID.Valid {
		parsedID, err := uuid.Parse(capacityUOMID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid capacity_uom_id: %w", err)
		}
		vehicle.CapacityUOMID = &parsedID
	}

	if lastServiceAt.Valid {
		vehicle.LastServiceAt = &lastServiceAt.Time
	}

	if createdBy.Valid {
		parsedID, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		vehicle.CreatedBy = &parsedID
	}

	if updatedBy.Valid {
		parsedID, err := uuid.Parse(updatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by: %w", err)
		}
		vehicle.UpdatedBy = &parsedID
	}

	if deletedAt.Valid {
		vehicle.DeletedAt = &deletedAt.Time
	}

	return &vehicle, nil
}

func (r *deliveryVehicleRepository) FindAll(ctx context.Context, filters DeliveryVehicleFilter) ([]deliverytypes.DeliveryVehicle, error) {
	query := `
		SELECT
			id, organization_id, name, registration_number, vehicle_identifier, vehicle_type,
			capacity, capacity_uom_id, active, last_service_at, service_interval_days, metadata,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_vehicles
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	if filters.OrganizationID != nil {
		query += fmt.Sprintf(" AND organization_id = $%d", argIndex)
		args = append(args, *filters.OrganizationID)
		argIndex++
	}

	if filters.Active != nil {
		query += fmt.Sprintf(" AND active = $%d", argIndex)
		args = append(args, *filters.Active)
		argIndex++
	}

	if filters.VehicleType != nil {
		query += fmt.Sprintf(" AND vehicle_type = $%d", argIndex)
		args = append(args, *filters.VehicleType)
		argIndex++
	}

	query += " ORDER BY name"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []deliverytypes.DeliveryVehicle
	for rows.Next() {
		var vehicle deliverytypes.DeliveryVehicle
		var lastServiceAt, deletedAt sql.NullTime
		var capacityUOMID, createdBy, updatedBy sql.NullString

		err := rows.Scan(
			&vehicle.ID,
			&vehicle.OrganizationID,
			&vehicle.Name,
			&vehicle.RegistrationNumber,
			&vehicle.VehicleIdentifier,
			&vehicle.VehicleType,
			&vehicle.Capacity,
			&capacityUOMID,
			&vehicle.Active,
			&lastServiceAt,
			&vehicle.ServiceIntervalDays,
			&vehicle.Metadata,
			&vehicle.CreatedAt,
			&vehicle.UpdatedAt,
			&createdBy,
			&updatedBy,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery vehicle: %w", err)
		}

		if capacityUOMID.Valid {
			parsedID, err := uuid.Parse(capacityUOMID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid capacity_uom_id: %w", err)
			}
			vehicle.CapacityUOMID = &parsedID
		}

		if lastServiceAt.Valid {
			vehicle.LastServiceAt = &lastServiceAt.Time
		}

		if createdBy.Valid {
			parsedID, err := uuid.Parse(createdBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid created_by: %w", err)
			}
			vehicle.CreatedBy = &parsedID
		}

		if updatedBy.Valid {
			parsedID, err := uuid.Parse(updatedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid updated_by: %w", err)
			}
			vehicle.UpdatedBy = &parsedID
		}

		if deletedAt.Valid {
			vehicle.DeletedAt = &deletedAt.Time
		}

		vehicles = append(vehicles, vehicle)
	}

	return vehicles, nil
}

func (r *deliveryVehicleRepository) Update(ctx context.Context, vehicle deliverytypes.DeliveryVehicle) (*deliverytypes.DeliveryVehicle, error) {
	query := `
		UPDATE delivery_vehicles SET
			name = $1,
			registration_number = $2,
			vehicle_identifier = $3,
			vehicle_type = $4,
			capacity = $5,
			capacity_uom_id = $6,
			active = $7,
			last_service_at = $8,
			service_interval_days = $9,
			metadata = $10,
			updated_at = NOW()
		WHERE id = $11 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		vehicle.Name,
		vehicle.RegistrationNumber,
		vehicle.VehicleIdentifier,
		vehicle.VehicleType,
		vehicle.Capacity,
		vehicle.CapacityUOMID,
		vehicle.Active,
		vehicle.LastServiceAt,
		vehicle.ServiceIntervalDays,
		vehicle.Metadata,
		vehicle.ID,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update delivery vehicle: %w", err)
	}

	vehicle.UpdatedAt = updatedAt
	return &vehicle, nil
}

func (r *deliveryVehicleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE delivery_vehicles
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery vehicle: %w", err)
	}

	return nil
}

func (r *deliveryVehicleRepository) FindByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]deliverytypes.DeliveryVehicle, error) {
	filters := DeliveryVehicleFilter{
		OrganizationID: &orgID,
	}
	return r.FindAll(ctx, filters)
}

func (r *deliveryVehicleRepository) FindActiveByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]deliverytypes.DeliveryVehicle, error) {
	active := true
	filters := DeliveryVehicleFilter{
		OrganizationID: &orgID,
		Active:         &active,
	}
	return r.FindAll(ctx, filters)
}
