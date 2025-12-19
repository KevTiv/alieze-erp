package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	deliverytypes "github.com/KevTiv/alieze-erp/internal/modules/delivery/types"

	"github.com/google/uuid"
)

type DeliveryRouteRepository interface {
	Create(ctx context.Context, route deliverytypes.DeliveryRoute) (*deliverytypes.DeliveryRoute, error)
	FindByID(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryRoute, error)
	FindAll(ctx context.Context, filters DeliveryRouteFilter) ([]deliverytypes.DeliveryRoute, error)
	Update(ctx context.Context, route deliverytypes.DeliveryRoute) (*deliverytypes.DeliveryRoute, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]deliverytypes.DeliveryRoute, error)
	FindByStatus(ctx context.Context, orgID uuid.UUID, status deliverytypes.RouteStatus) ([]deliverytypes.DeliveryRoute, error)
}

type DeliveryRouteFilter struct {
	OrganizationID *uuid.UUID
	Status         *deliverytypes.RouteStatus
	TransportMode  *deliverytypes.TransportMode
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}

type deliveryRouteRepository struct {
	db *sql.DB
}

func NewDeliveryRouteRepository(db *sql.DB) DeliveryRouteRepository {
	return &deliveryRouteRepository{db: db}
}

func (r *deliveryRouteRepository) Create(ctx context.Context, route deliverytypes.DeliveryRoute) (*deliverytypes.DeliveryRoute, error) {
	query := `
		INSERT INTO delivery_routes (
			organization_id, company_id, warehouse_id, name, route_code, transport_mode,
			status, scheduled_start_at, scheduled_end_at, origin_location_id, destination_location_id,
			notes, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		route.OrganizationID,
		route.CompanyID,
		route.WarehouseID,
		route.Name,
		route.RouteCode,
		route.TransportMode,
		route.Status,
		route.ScheduledStartAt,
		route.ScheduledEndAt,
		route.OriginLocationID,
		route.DestinationLocationID,
		route.Notes,
		route.Metadata,
	).Scan(&route.ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery route: %w", err)
	}

	route.CreatedAt = createdAt
	route.UpdatedAt = updatedAt
	return &route, nil
}

func (r *deliveryRouteRepository) FindByID(ctx context.Context, id uuid.UUID) (*deliverytypes.DeliveryRoute, error) {
	query := `
		SELECT
			id, organization_id, company_id, warehouse_id, name, route_code, transport_mode,
			status, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at,
			origin_location_id, destination_location_id, notes, metadata,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_routes
		WHERE id = $1 AND deleted_at IS NULL
	`

	var route deliverytypes.DeliveryRoute
	var companyID, warehouseID, originLocationID, destinationLocationID, createdBy, updatedBy sql.NullString
	var scheduledStartAt, scheduledEndAt, actualStartAt, actualEndAt, deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID,
		&route.OrganizationID,
		&companyID,
		&warehouseID,
		&route.Name,
		&route.RouteCode,
		&route.TransportMode,
		&route.Status,
		&scheduledStartAt,
		&scheduledEndAt,
		&actualStartAt,
		&actualEndAt,
		&originLocationID,
		&destinationLocationID,
		&route.Notes,
		&route.Metadata,
		&route.CreatedAt,
		&route.UpdatedAt,
		&createdBy,
		&updatedBy,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find delivery route: %w", err)
	}

	if companyID.Valid {
		parsedID, err := uuid.Parse(companyID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid company_id: %w", err)
		}
		route.CompanyID = &parsedID
	}

	if warehouseID.Valid {
		parsedID, err := uuid.Parse(warehouseID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid warehouse_id: %w", err)
		}
		route.WarehouseID = &parsedID
	}

	if originLocationID.Valid {
		parsedID, err := uuid.Parse(originLocationID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid origin_location_id: %w", err)
		}
		route.OriginLocationID = &parsedID
	}

	if destinationLocationID.Valid {
		parsedID, err := uuid.Parse(destinationLocationID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid destination_location_id: %w", err)
		}
		route.DestinationLocationID = &parsedID
	}

	if scheduledStartAt.Valid {
		route.ScheduledStartAt = &scheduledStartAt.Time
	}

	if scheduledEndAt.Valid {
		route.ScheduledEndAt = &scheduledEndAt.Time
	}

	if actualStartAt.Valid {
		route.ActualStartAt = &actualStartAt.Time
	}

	if actualEndAt.Valid {
		route.ActualEndAt = &actualEndAt.Time
	}

	if createdBy.Valid {
		parsedID, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		route.CreatedBy = &parsedID
	}

	if updatedBy.Valid {
		parsedID, err := uuid.Parse(updatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by: %w", err)
		}
		route.UpdatedBy = &parsedID
	}

	if deletedAt.Valid {
		route.DeletedAt = &deletedAt.Time
	}

	return &route, nil
}

func (r *deliveryRouteRepository) FindAll(ctx context.Context, filters DeliveryRouteFilter) ([]deliverytypes.DeliveryRoute, error) {
	query := `
		SELECT
			id, organization_id, company_id, warehouse_id, name, route_code, transport_mode,
			status, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at,
			origin_location_id, destination_location_id, notes, metadata,
			created_at, updated_at, created_by, updated_by, deleted_at
		FROM delivery_routes
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	if filters.OrganizationID != nil {
		query += fmt.Sprintf(" AND organization_id = $%d", argIndex)
		args = append(args, *filters.OrganizationID)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.TransportMode != nil {
		query += fmt.Sprintf(" AND transport_mode = $%d", argIndex)
		args = append(args, *filters.TransportMode)
		argIndex++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND scheduled_start_at >= $%d", argIndex)
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND scheduled_start_at <= $%d", argIndex)
		args = append(args, *filters.DateTo)
		argIndex++
	}

	query += " ORDER BY scheduled_start_at DESC"

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
		return nil, fmt.Errorf("failed to query delivery routes: %w", err)
	}
	defer rows.Close()

	var routes []deliverytypes.DeliveryRoute
	for rows.Next() {
		var route deliverytypes.DeliveryRoute
		var companyID, warehouseID, originLocationID, destinationLocationID, createdBy, updatedBy sql.NullString
		var scheduledStartAt, scheduledEndAt, actualStartAt, actualEndAt, deletedAt sql.NullTime

		err := rows.Scan(
			&route.ID,
			&route.OrganizationID,
			&companyID,
			&warehouseID,
			&route.Name,
			&route.RouteCode,
			&route.TransportMode,
			&route.Status,
			&scheduledStartAt,
			&scheduledEndAt,
			&actualStartAt,
			&actualEndAt,
			&originLocationID,
			&destinationLocationID,
			&route.Notes,
			&route.Metadata,
			&route.CreatedAt,
			&route.UpdatedAt,
			&createdBy,
			&updatedBy,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery route: %w", err)
		}

		if companyID.Valid {
			parsedID, err := uuid.Parse(companyID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid company_id: %w", err)
			}
			route.CompanyID = &parsedID
		}

		if warehouseID.Valid {
			parsedID, err := uuid.Parse(warehouseID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid warehouse_id: %w", err)
			}
			route.WarehouseID = &parsedID
		}

		if originLocationID.Valid {
			parsedID, err := uuid.Parse(originLocationID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid origin_location_id: %w", err)
			}
			route.OriginLocationID = &parsedID
		}

		if destinationLocationID.Valid {
			parsedID, err := uuid.Parse(destinationLocationID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid destination_location_id: %w", err)
			}
			route.DestinationLocationID = &parsedID
		}

		if scheduledStartAt.Valid {
			route.ScheduledStartAt = &scheduledStartAt.Time
		}

		if scheduledEndAt.Valid {
			route.ScheduledEndAt = &scheduledEndAt.Time
		}

		if actualStartAt.Valid {
			route.ActualStartAt = &actualStartAt.Time
		}

		if actualEndAt.Valid {
			route.ActualEndAt = &actualEndAt.Time
		}

		if createdBy.Valid {
			parsedID, err := uuid.Parse(createdBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid created_by: %w", err)
			}
			route.CreatedBy = &parsedID
		}

		if updatedBy.Valid {
			parsedID, err := uuid.Parse(updatedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid updated_by: %w", err)
			}
			route.UpdatedBy = &parsedID
		}

		if deletedAt.Valid {
			route.DeletedAt = &deletedAt.Time
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func (r *deliveryRouteRepository) Update(ctx context.Context, route deliverytypes.DeliveryRoute) (*deliverytypes.DeliveryRoute, error) {
	query := `
		UPDATE delivery_routes SET
			name = $1,
			route_code = $2,
			transport_mode = $3,
			status = $4,
			scheduled_start_at = $5,
			scheduled_end_at = $6,
			actual_start_at = $7,
			actual_end_at = $8,
			origin_location_id = $9,
			destination_location_id = $10,
			notes = $11,
			metadata = $12,
			updated_at = NOW()
		WHERE id = $13 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		route.Name,
		route.RouteCode,
		route.TransportMode,
		route.Status,
		route.ScheduledStartAt,
		route.ScheduledEndAt,
		route.ActualStartAt,
		route.ActualEndAt,
		route.OriginLocationID,
		route.DestinationLocationID,
		route.Notes,
		route.Metadata,
		route.ID,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update delivery route: %w", err)
	}

	route.UpdatedAt = updatedAt
	return &route, nil
}

func (r *deliveryRouteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE delivery_routes
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery route: %w", err)
	}

	return nil
}

func (r *deliveryRouteRepository) FindByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]deliverytypes.DeliveryRoute, error) {
	filters := DeliveryRouteFilter{
		OrganizationID: &orgID,
	}
	return r.FindAll(ctx, filters)
}

func (r *deliveryRouteRepository) FindByStatus(ctx context.Context, orgID uuid.UUID, status deliverytypes.RouteStatus) ([]deliverytypes.DeliveryRoute, error) {
	filters := DeliveryRouteFilter{
		OrganizationID: &orgID,
		Status:         &status,
	}
	return r.FindAll(ctx, filters)
}
