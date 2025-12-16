package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Warehouse Repository

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse domain.Warehouse) (*domain.Warehouse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Warehouse, error)
	Update(ctx context.Context, warehouse domain.Warehouse) (*domain.Warehouse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type warehouseRepository struct {
	db *sql.DB
}

func NewWarehouseRepository(db *sql.DB) WarehouseRepository {
	return &warehouseRepository{db: db}
}

func (r *warehouseRepository) Create(ctx context.Context, wh domain.Warehouse) (*domain.Warehouse, error) {
	query := `
		INSERT INTO warehouses
		(id, organization_id, company_id, name, code, partner_id, reception_steps,
		 delivery_steps, active, sequence, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, organization_id, company_id, name, code, partner_id,
		 reception_steps, delivery_steps, active, sequence, created_at, updated_at
	`

	if wh.ID == uuid.Nil {
		wh.ID = uuid.New()
	}
	now := time.Now()
	if wh.CreatedAt.IsZero() {
		wh.CreatedAt = now
	}
	if wh.UpdatedAt.IsZero() {
		wh.UpdatedAt = now
	}

	var created domain.Warehouse
	err := r.db.QueryRowContext(ctx, query,
		wh.ID, wh.OrganizationID, wh.CompanyID, wh.Name, wh.Code, wh.PartnerID,
		wh.ReceptionSteps, wh.DeliverySteps, wh.Active, wh.Sequence, wh.CreatedAt, wh.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Name, &created.Code,
		&created.PartnerID, &created.ReceptionSteps, &created.DeliverySteps, &created.Active,
		&created.Sequence, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}
	return &created, nil
}

func (r *warehouseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, partner_id,
		 reception_steps, delivery_steps, active, sequence, created_at, updated_at
		FROM warehouses WHERE id = $1 AND deleted_at IS NULL
	`

	var wh domain.Warehouse
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wh.ID, &wh.OrganizationID, &wh.CompanyID, &wh.Name, &wh.Code, &wh.PartnerID,
		&wh.ReceptionSteps, &wh.DeliverySteps, &wh.Active, &wh.Sequence, &wh.CreatedAt, &wh.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find warehouse: %w", err)
	}
	return &wh, nil
}

func (r *warehouseRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Warehouse, error) {
	query := `
		SELECT id, organization_id, company_id, name, code, partner_id,
		 reception_steps, delivery_steps, active, sequence, created_at, updated_at
		FROM warehouses WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY sequence ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var wh domain.Warehouse
		err := rows.Scan(
			&wh.ID, &wh.OrganizationID, &wh.CompanyID, &wh.Name, &wh.Code, &wh.PartnerID,
			&wh.ReceptionSteps, &wh.DeliverySteps, &wh.Active, &wh.Sequence, &wh.CreatedAt, &wh.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan warehouse: %w", err)
		}
		warehouses = append(warehouses, wh)
	}
	return warehouses, nil
}

func (r *warehouseRepository) Update(ctx context.Context, wh domain.Warehouse) (*domain.Warehouse, error) {
	query := `
		UPDATE warehouses
		SET name = $2, code = $3, partner_id = $4, reception_steps = $5,
		    delivery_steps = $6, active = $7, sequence = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, name, code, partner_id,
		 reception_steps, delivery_steps, active, sequence, created_at, updated_at
	`

	wh.UpdatedAt = time.Now()
	var updated domain.Warehouse
	err := r.db.QueryRowContext(ctx, query,
		wh.ID, wh.Name, wh.Code, wh.PartnerID, wh.ReceptionSteps, wh.DeliverySteps,
		wh.Active, wh.Sequence, wh.UpdatedAt,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.Name, &updated.Code,
		&updated.PartnerID, &updated.ReceptionSteps, &updated.DeliverySteps, &updated.Active,
		&updated.Sequence, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("warehouse not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}
	return &updated, nil
}

func (r *warehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE warehouses SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete warehouse: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("warehouse not found")
	}
	return nil
}

// Stock Location Repository

type StockLocationRepository interface {
	Create(ctx context.Context, location domain.StockLocation) (*domain.StockLocation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.StockLocation, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.StockLocation, error)
	Update(ctx context.Context, location domain.StockLocation) (*domain.StockLocation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type stockLocationRepository struct {
	db *sql.DB
}

func NewStockLocationRepository(db *sql.DB) StockLocationRepository {
	return &stockLocationRepository{db: db}
}

func (r *stockLocationRepository) Create(ctx context.Context, loc domain.StockLocation) (*domain.StockLocation, error) {
	query := `
		INSERT INTO stock_locations
		(id, organization_id, company_id, name, location_id, usage, removal_strategy,
		 active, scrap_location, return_location, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, organization_id, company_id, name, location_id, usage, removal_strategy,
		 active, scrap_location, return_location, created_at, updated_at
	`

	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	now := time.Now()
	if loc.CreatedAt.IsZero() {
		loc.CreatedAt = now
	}
	if loc.UpdatedAt.IsZero() {
		loc.UpdatedAt = now
	}

	var created domain.StockLocation
	err := r.db.QueryRowContext(ctx, query,
		loc.ID, loc.OrganizationID, loc.CompanyID, loc.Name, loc.LocationID, loc.Usage,
		loc.RemovalStrategy, loc.Active, loc.ScrapLocation, loc.ReturnLocation,
		loc.CreatedAt, loc.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Name, &created.LocationID,
		&created.Usage, &created.RemovalStrategy, &created.Active, &created.ScrapLocation,
		&created.ReturnLocation, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock location: %w", err)
	}
	return &created, nil
}

func (r *stockLocationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockLocation, error) {
	query := `
		SELECT id, organization_id, company_id, name, location_id, usage, removal_strategy,
		 active, scrap_location, return_location, created_at, updated_at
		FROM stock_locations WHERE id = $1 AND deleted_at IS NULL
	`

	var loc domain.StockLocation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&loc.ID, &loc.OrganizationID, &loc.CompanyID, &loc.Name, &loc.LocationID, &loc.Usage,
		&loc.RemovalStrategy, &loc.Active, &loc.ScrapLocation, &loc.ReturnLocation,
		&loc.CreatedAt, &loc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find stock location: %w", err)
	}
	return &loc, nil
}

func (r *stockLocationRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.StockLocation, error) {
	query := `
		SELECT id, organization_id, company_id, name, location_id, usage, removal_strategy,
		 active, scrap_location, return_location, created_at, updated_at
		FROM stock_locations WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find stock locations: %w", err)
	}
	defer rows.Close()

	var locations []domain.StockLocation
	for rows.Next() {
		var loc domain.StockLocation
		err := rows.Scan(
			&loc.ID, &loc.OrganizationID, &loc.CompanyID, &loc.Name, &loc.LocationID, &loc.Usage,
			&loc.RemovalStrategy, &loc.Active, &loc.ScrapLocation, &loc.ReturnLocation,
			&loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock location: %w", err)
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

func (r *stockLocationRepository) Update(ctx context.Context, loc domain.StockLocation) (*domain.StockLocation, error) {
	query := `
		UPDATE stock_locations
		SET name = $2, location_id = $3, usage = $4, removal_strategy = $5,
		    active = $6, scrap_location = $7, return_location = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, name, location_id, usage, removal_strategy,
		 active, scrap_location, return_location, created_at, updated_at
	`

	loc.UpdatedAt = time.Now()
	var updated domain.StockLocation
	err := r.db.QueryRowContext(ctx, query,
		loc.ID, loc.Name, loc.LocationID, loc.Usage, loc.RemovalStrategy,
		loc.Active, loc.ScrapLocation, loc.ReturnLocation, loc.UpdatedAt,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.Name, &updated.LocationID,
		&updated.Usage, &updated.RemovalStrategy, &updated.Active, &updated.ScrapLocation,
		&updated.ReturnLocation, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stock location not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update stock location: %w", err)
	}
	return &updated, nil
}

func (r *stockLocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE stock_locations SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete stock location: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("stock location not found")
	}
	return nil
}

// Stock Quant Repository

type StockQuantRepository interface {
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.StockQuant, error)
	FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]domain.StockQuant, error)
	FindAvailable(ctx context.Context, organizationID, productID, locationID uuid.UUID) (float64, error)
	UpdateQuantity(ctx context.Context, organizationID, productID, locationID uuid.UUID, deltaQty float64) error
}

type stockQuantRepository struct {
	db *sql.DB
}

func NewStockQuantRepository(db *sql.DB) StockQuantRepository {
	return &stockQuantRepository{db: db}
}

func (r *stockQuantRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.StockQuant, error) {
	query := `
		SELECT id, organization_id, company_id, product_id, location_id, lot_id, package_id,
		 owner_id, quantity, reserved_quantity, in_date, created_at, updated_at
		FROM stock_quants
		WHERE organization_id = $1 AND product_id = $2 AND quantity > 0
		ORDER BY location_id, in_date
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find stock quants: %w", err)
	}
	defer rows.Close()

	var quants []domain.StockQuant
	for rows.Next() {
		var q domain.StockQuant
		err := rows.Scan(
			&q.ID, &q.OrganizationID, &q.CompanyID, &q.ProductID, &q.LocationID, &q.LotID,
			&q.PackageID, &q.OwnerID, &q.Quantity, &q.ReservedQuantity, &q.InDate,
			&q.CreatedAt, &q.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock quant: %w", err)
		}
		quants = append(quants, q)
	}
	return quants, nil
}

func (r *stockQuantRepository) FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]domain.StockQuant, error) {
	query := `
		SELECT id, organization_id, company_id, product_id, location_id, lot_id, package_id,
		 owner_id, quantity, reserved_quantity, in_date, created_at, updated_at
		FROM stock_quants
		WHERE organization_id = $1 AND location_id = $2 AND quantity > 0
		ORDER BY product_id
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find stock quants: %w", err)
	}
	defer rows.Close()

	var quants []domain.StockQuant
	for rows.Next() {
		var q domain.StockQuant
		err := rows.Scan(
			&q.ID, &q.OrganizationID, &q.CompanyID, &q.ProductID, &q.LocationID, &q.LotID,
			&q.PackageID, &q.OwnerID, &q.Quantity, &q.ReservedQuantity, &q.InDate,
			&q.CreatedAt, &q.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock quant: %w", err)
		}
		quants = append(quants, q)
	}
	return quants, nil
}

func (r *stockQuantRepository) FindAvailable(ctx context.Context, organizationID, productID, locationID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(quantity - reserved_quantity), 0)
		FROM stock_quants
		WHERE organization_id = $1 AND product_id = $2 AND location_id = $3
	`

	var available float64
	err := r.db.QueryRowContext(ctx, query, organizationID, productID, locationID).Scan(&available)
	if err != nil {
		return 0, fmt.Errorf("failed to get available quantity: %w", err)
	}
	return available, nil
}

func (r *stockQuantRepository) UpdateQuantity(ctx context.Context, organizationID, productID, locationID uuid.UUID, deltaQty float64) error {
	// Simplified: Try to update existing quant, or insert new one
	query := `
		INSERT INTO stock_quants (id, organization_id, product_id, location_id, quantity, reserved_quantity, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 0, now(), now())
		ON CONFLICT (product_id, location_id, COALESCE(lot_id, '00000000-0000-0000-0000-000000000000'::uuid),
					 COALESCE(package_id, '00000000-0000-0000-0000-000000000000'::uuid),
					 COALESCE(owner_id, '00000000-0000-0000-0000-000000000000'::uuid), organization_id)
		DO UPDATE SET quantity = stock_quants.quantity + $4, updated_at = now()
	`

	_, err := r.db.ExecContext(ctx, query, organizationID, productID, locationID, deltaQty)
	if err != nil {
		return fmt.Errorf("failed to update quantity: %w", err)
	}
	return nil
}

// Stock Move Repository

type StockMoveRepository interface {
	Create(ctx context.Context, move domain.StockMove) (*domain.StockMove, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.StockMove, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.StockMove, error)
	UpdateState(ctx context.Context, id uuid.UUID, state string) error
}

type stockMoveRepository struct {
	db *sql.DB
}

func NewStockMoveRepository(db *sql.DB) StockMoveRepository {
	return &stockMoveRepository{db: db}
}

func (r *stockMoveRepository) Create(ctx context.Context, move domain.StockMove) (*domain.StockMove, error) {
	query := `
		INSERT INTO stock_moves
		(id, organization_id, company_id, name, sequence, priority, date, product_id,
		 product_uom_qty, location_id, location_dest_id, partner_id, state, procure_method,
		 origin, lot_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, organization_id, company_id, name, sequence, priority, date, product_id,
		 product_uom_qty, location_id, location_dest_id, partner_id, state, procure_method,
		 origin, lot_ids, created_at, updated_at
	`

	if move.ID == uuid.Nil {
		move.ID = uuid.New()
	}
	now := time.Now()
	if move.CreatedAt.IsZero() {
		move.CreatedAt = now
	}
	if move.UpdatedAt.IsZero() {
		move.UpdatedAt = now
	}
	if move.Date.IsZero() {
		move.Date = now
	}

	var created domain.StockMove
	err := r.db.QueryRowContext(ctx, query,
		move.ID, move.OrganizationID, move.CompanyID, move.Name, move.Sequence, move.Priority,
		move.Date, move.ProductID, move.ProductUOMQty, move.LocationID, move.LocationDestID,
		move.PartnerID, move.State, move.ProcureMethod, move.Origin, pq.Array(move.LotIDs),
		move.CreatedAt, move.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Name, &created.Sequence,
		&created.Priority, &created.Date, &created.ProductID, &created.ProductUOMQty, &created.LocationID,
		&created.LocationDestID, &created.PartnerID, &created.State, &created.ProcureMethod,
		&created.Origin, pq.Array(&created.LotIDs), &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock move: %w", err)
	}
	return &created, nil
}

func (r *stockMoveRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StockMove, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence, priority, date, product_id,
		 product_uom_qty, location_id, location_dest_id, partner_id, state, procure_method,
		 origin, lot_ids, created_at, updated_at
		FROM stock_moves WHERE id = $1 AND deleted_at IS NULL
	`

	var move domain.StockMove
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority,
		&move.Date, &move.ProductID, &move.ProductUOMQty, &move.LocationID, &move.LocationDestID,
		&move.PartnerID, &move.State, &move.ProcureMethod, &move.Origin, pq.Array(&move.LotIDs),
		&move.CreatedAt, &move.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find stock move: %w", err)
	}
	return &move, nil
}

func (r *stockMoveRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.StockMove, error) {
	query := `
		SELECT id, organization_id, company_id, name, sequence, priority, date, product_id,
		 product_uom_qty, location_id, location_dest_id, partner_id, state, procure_method,
		 origin, lot_ids, created_at, updated_at
		FROM stock_moves WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY date DESC, sequence ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find stock moves: %w", err)
	}
	defer rows.Close()

	var moves []domain.StockMove
	for rows.Next() {
		var move domain.StockMove
		err := rows.Scan(
			&move.ID, &move.OrganizationID, &move.CompanyID, &move.Name, &move.Sequence, &move.Priority,
			&move.Date, &move.ProductID, &move.ProductUOMQty, &move.LocationID, &move.LocationDestID,
			&move.PartnerID, &move.State, &move.ProcureMethod, &move.Origin, pq.Array(&move.LotIDs),
			&move.CreatedAt, &move.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock move: %w", err)
		}
		moves = append(moves, move)
	}
	return moves, nil
}

func (r *stockMoveRepository) UpdateState(ctx context.Context, id uuid.UUID, state string) error {
	query := `UPDATE stock_moves SET state = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, state, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update stock move state: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("stock move not found")
	}
	return nil
}
