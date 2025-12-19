package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// BarcodeRepository interface for barcode scanning operations
type BarcodeRepository interface {
	// Barcode scanning operations
	ScanBarcode(ctx context.Context, request types.BarcodeScanRequest) (*types.BarcodeScanResponse, error)
	GetScanByID(ctx context.Context, orgID uuid.UUID, scanID uuid.UUID) (*types.BarcodeScan, error)
	ListScans(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]types.BarcodeScan, error)

	// Mobile scanning session operations
	CreateScanningSession(ctx context.Context, request types.CreateScanningSessionRequest) (*types.MobileScanningSession, error)
	GetScanningSession(ctx context.Context, orgID, sessionID uuid.UUID) (*types.MobileScanningSession, error)
	ListScanningSessions(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.MobileScanningSession, error)
	AddScanToSession(ctx context.Context, request types.AddScanToSessionRequest) (*types.BarcodeScanResponse, error)
	CompleteScanningSession(ctx context.Context, request types.CompleteSessionRequest) (bool, error)
	GetSessionLines(ctx context.Context, orgID, sessionID uuid.UUID) ([]types.MobileScanningSessionLine, error)

	// Barcode generation operations
	GenerateBarcode(ctx context.Context, request types.BarcodeGenerationRequest) (*types.BarcodeGenerationResponse, error)
	GenerateBarcodesForProducts(ctx context.Context, orgID uuid.UUID, productIDs []uuid.UUID, prefix *string) (map[uuid.UUID]string, error)

	// Barcode lookup operations
	FindEntityByBarcode(ctx context.Context, orgID uuid.UUID, barcode string) (*types.BarcodeEntity, error)
	ValidateBarcodeFormat(ctx context.Context, barcode string) (bool, error)
}

type barcodeRepository struct {
	db *sql.DB
}

func NewBarcodeRepository(db *sql.DB) BarcodeRepository {
	return &barcodeRepository{db: db}
}

// ScanBarcode performs a barcode scan and returns the result
func (r *barcodeRepository) ScanBarcode(ctx context.Context, request types.BarcodeScanRequest) (*types.BarcodeScanResponse, error) {
	// Validate barcode format
	valid, err := r.ValidateBarcodeFormat(ctx, request.Barcode)
	if err != nil {
		return nil, err
	}
	if !valid {
		return &types.BarcodeScanResponse{
			Success:   false,
			Message:   "Invalid barcode format",
			Timestamp: time.Now(),
		}, nil
	}

	// Find entity by barcode
	entity, err := r.FindEntityByBarcode(ctx, request.OrganizationID, request.Barcode)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		// Log failed scan
		scanID := uuid.New()
		query := `
			INSERT INTO barcode_scans (
				id, organization_id, user_id, scan_type, scanned_barcode,
				success, error_message, scan_time, created_at
			) VALUES (
				$1, $2, $3, 'unknown', $4, false, 'Barcode not found', NOW(), NOW()
			)
		`
		_, err := r.db.ExecContext(ctx, query, scanID, request.OrganizationID, request.UserID, request.Barcode)
		if err != nil {
			return nil, err
		}

		return &types.BarcodeScanResponse{
			Success:   false,
			Message:   "Barcode not found",
			ScanID:    scanID,
			Timestamp: time.Now(),
		}, nil
	}

	// Log successful scan
	scanID := uuid.New()
	query := `
		INSERT INTO barcode_scans (
			id, organization_id, user_id, scan_type, scanned_barcode,
			entity_id, entity_type, location_id, quantity, success,
			scan_time, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW(), NOW()
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		scanID, request.OrganizationID, request.UserID, entity.EntityType,
		request.Barcode, entity.EntityID, entity.EntityType, request.LocationID, request.Quantity,
	)
	if err != nil {
		return nil, err
	}

	return &types.BarcodeScanResponse{
		Success:   true,
		Message:   "Scan successful",
		ScanID:    scanID,
		Entity: &types.BarcodeEntity{
			EntityType:    entity.EntityType,
			EntityID:      entity.EntityID,
			EntityName:    entity.EntityName,
			AdditionalInfo: entity.AdditionalInfo,
		},
		Timestamp: time.Now(),
	}, nil
}

// GetScanByID retrieves a specific barcode scan
func (r *barcodeRepository) GetScanByID(ctx context.Context, orgID uuid.UUID, scanID uuid.UUID) (*types.BarcodeScan, error) {
	query := `
		SELECT
			id, organization_id, user_id, scan_type, scanned_barcode,
			entity_id, entity_type, location_id, quantity, scan_time,
			device_info, ip_address, success, error_message, metadata, created_at
		FROM barcode_scans
		WHERE organization_id = $1 AND id = $2
	`

	var scan types.BarcodeScan
	err := r.db.QueryRowContext(ctx, query, orgID, scanID).Scan(
		&scan.ID, &scan.OrganizationID, &scan.UserID, &scan.ScanType, &scan.ScannedBarcode,
		&scan.EntityID, &scan.EntityType, &scan.LocationID, &scan.Quantity, &scan.ScanTime,
		&scan.DeviceInfo, &scan.IPAddress, &scan.Success, &scan.ErrorMessage, &scan.Metadata, &scan.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &scan, nil
}

// ListScans retrieves barcode scans for an organization
func (r *barcodeRepository) ListScans(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]types.BarcodeScan, error) {
	query := `
		SELECT
			id, organization_id, user_id, scan_type, scanned_barcode,
			entity_id, entity_type, location_id, quantity, scan_time,
			device_info, ip_address, success, error_message, metadata, created_at
		FROM barcode_scans
		WHERE organization_id = $1
		ORDER BY scan_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []types.BarcodeScan
	for rows.Next() {
		var scan types.BarcodeScan
		err := rows.Scan(
			&scan.ID, &scan.OrganizationID, &scan.UserID, &scan.ScanType, &scan.ScannedBarcode,
			&scan.EntityID, &scan.EntityType, &scan.LocationID, &scan.Quantity, &scan.ScanTime,
			&scan.DeviceInfo, &scan.IPAddress, &scan.Success, &scan.ErrorMessage, &scan.Metadata, &scan.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		scans = append(scans, scan)
	}

	return scans, nil
}

// CreateScanningSession creates a new mobile scanning session
func (r *barcodeRepository) CreateScanningSession(ctx context.Context, request types.CreateScanningSessionRequest) (*types.MobileScanningSession, error) {
	query := `
		INSERT INTO mobile_scanning_sessions (
			id, organization_id, user_id, device_id, session_type,
			location_id, reference, metadata, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
		) RETURNING
			id, organization_id, user_id, device_id, session_type,
			status, start_time, end_time, location_id, reference, metadata, created_at, updated_at
	`

	var session types.MobileScanningSession
	err := r.db.QueryRowContext(ctx, query,
		request.OrganizationID, request.UserID, request.DeviceID, request.SessionType,
		request.LocationID, request.Reference, request.Metadata,
	).Scan(
		&session.ID, &session.OrganizationID, &session.UserID, &session.DeviceID, &session.SessionType,
		&session.Status, &session.StartTime, &session.EndTime, &session.LocationID, &session.Reference, &session.Metadata, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetScanningSession retrieves a scanning session
func (r *barcodeRepository) GetScanningSession(ctx context.Context, orgID, sessionID uuid.UUID) (*types.MobileScanningSession, error) {
	query := `
		SELECT
			id, organization_id, user_id, device_id, session_type,
			status, start_time, end_time, location_id, reference, metadata, created_at, updated_at
		FROM mobile_scanning_sessions
		WHERE organization_id = $1 AND id = $2
	`

	var session types.MobileScanningSession
	err := r.db.QueryRowContext(ctx, query, orgID, sessionID).Scan(
		&session.ID, &session.OrganizationID, &session.UserID, &session.DeviceID, &session.SessionType,
		&session.Status, &session.StartTime, &session.EndTime, &session.LocationID, &session.Reference, &session.Metadata, &session.CreatedAt, &session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// ListScanningSessions retrieves scanning sessions for an organization
func (r *barcodeRepository) ListScanningSessions(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]types.MobileScanningSession, error) {
	query := `
		SELECT
			id, organization_id, user_id, device_id, session_type,
			status, start_time, end_time, location_id, reference, metadata, created_at, updated_at
		FROM mobile_scanning_sessions
		WHERE organization_id = $1
	`

	var params []interface{}
	params = append(params, orgID)

	if status != nil {
		query += ` AND status = $` + string(len(params)+1)
		params = append(params, *status)
	}

	query += ` ORDER BY start_time DESC LIMIT $` + string(len(params)+1) + ` OFFSET $` + string(len(params)+2)
	params = append(params, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []types.MobileScanningSession
	for rows.Next() {
		var session types.MobileScanningSession
		err := rows.Scan(
			&session.ID, &session.OrganizationID, &session.UserID, &session.DeviceID, &session.SessionType,
			&session.Status, &session.StartTime, &session.EndTime, &session.LocationID, &session.Reference, &session.Metadata, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// AddScanToSession adds a scan to an existing session
func (r *barcodeRepository) AddScanToSession(ctx context.Context, request types.AddScanToSessionRequest) (*types.BarcodeScanResponse, error) {
	// First, perform the scan
	scanRequest := types.BarcodeScanRequest{
		OrganizationID: request.OrganizationID,
		UserID:        request.UserID,
		Barcode:       request.Barcode,
		Quantity:      request.Quantity,
		LocationID:    request.LocationID,
		DeviceInfo:    request.DeviceInfo,
	}

	scanResponse, err := r.ScanBarcode(ctx, scanRequest)
	if err != nil {
		return nil, err
	}

	if !scanResponse.Success {
		return scanResponse, nil
	}

	// Add to session lines
	query := `
		INSERT INTO mobile_scanning_session_lines (
			id, session_id, organization_id, scan_id, product_id,
			location_id, quantity, scanned_quantity, status, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3,
			CASE WHEN $4 = 'product' THEN $5 ELSE NULL END,
			$6, $7, $7, 'scanned', NOW(), NOW()
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		request.SessionID, request.OrganizationID, scanResponse.ScanID,
		scanResponse.Entity.EntityType, scanResponse.Entity.EntityID,
		request.LocationID, request.Quantity,
	)
	if err != nil {
		return nil, err
	}

	return scanResponse, nil
}

// CompleteScanningSession marks a session as completed
func (r *barcodeRepository) CompleteScanningSession(ctx context.Context, request types.CompleteSessionRequest) (bool, error) {
	query := `
		UPDATE mobile_scanning_sessions
		SET status = 'completed', end_time = NOW(), updated_at = NOW()
		WHERE organization_id = $1 AND id = $2 AND status = 'active'
	`

	result, err := r.db.ExecContext(ctx, query, request.OrganizationID, request.SessionID)
	if err != nil {
		return false, err
	}

	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// GetSessionLines retrieves all lines for a scanning session
func (r *barcodeRepository) GetSessionLines(ctx context.Context, orgID, sessionID uuid.UUID) ([]types.MobileScanningSessionLine, error) {
	query := `
		SELECT
			id, session_id, organization_id, scan_id, product_id,
			product_variant_id, location_id, lot_id, package_id,
			quantity, scanned_quantity, uom_id, status, notes,
			sequence, created_at, updated_at
		FROM mobile_scanning_session_lines
		WHERE organization_id = $1 AND session_id = $2
		ORDER BY sequence, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []types.MobileScanningSessionLine
	for rows.Next() {
		var line types.MobileScanningSessionLine
		err := rows.Scan(
			&line.ID, &line.SessionID, &line.OrganizationID, &line.ScanID, &line.ProductID,
			&line.ProductVariantID, &line.LocationID, &line.LotID, &line.PackageID,
			&line.Quantity, &line.ScannedQuantity, &line.UOMID, &line.Status, &line.Notes,
			&line.Sequence, &line.CreatedAt, &line.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// GenerateBarcode generates a barcode for an entity
func (r *barcodeRepository) GenerateBarcode(ctx context.Context, request types.BarcodeGenerationRequest) (*types.BarcodeGenerationResponse, error) {
	query := `SELECT generate_barcode($1, $2, $3)`

	var barcode string
	err := r.db.QueryRowContext(ctx, query, request.EntityType, request.OrganizationID, request.Prefix).Scan(&barcode)
	if err != nil {
		return nil, err
	}

	// Update the entity with the generated barcode
	var updateQuery string
	var entityName string

	switch request.EntityType {
	case "product":
		updateQuery = `UPDATE products SET barcode = $1 WHERE id = $2 AND organization_id = $3`
		entityName = "product"
	case "location":
		updateQuery = `UPDATE stock_locations SET barcode = $1 WHERE id = $2 AND organization_id = $3`
		entityName = "location"
	case "lot":
		updateQuery = `UPDATE stock_lots SET barcode = $1 WHERE id = $2 AND organization_id = $3`
		entityName = "lot"
	case "package":
		updateQuery = `UPDATE stock_packages SET barcode = $1 WHERE id = $2 AND organization_id = $3`
		entityName = "package"
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", request.EntityType)
	}

	_, err = r.db.ExecContext(ctx, updateQuery, barcode, request.EntityID, request.OrganizationID)
	if err != nil {
		return nil, err
	}

	return &types.BarcodeGenerationResponse{
		Success:    true,
		Barcode:    barcode,
		EntityType: entityName,
		EntityID:   request.EntityID,
	}, nil
}

// GenerateBarcodesForProducts generates barcodes for multiple products
func (r *barcodeRepository) GenerateBarcodesForProducts(ctx context.Context, orgID uuid.UUID, productIDs []uuid.UUID, prefix *string) (map[uuid.UUID]string, error) {
	barcodes := make(map[uuid.UUID]string)

	for _, productID := range productIDs {
		request := types.BarcodeGenerationRequest{
			OrganizationID: orgID,
			EntityType:    "product",
			EntityID:      productID,
			Prefix:        prefix,
		}

		response, err := r.GenerateBarcode(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to generate barcode for product %s: %w", productID, err)
		}

		barcodes[productID] = response.Barcode
	}

	return barcodes, nil
}

// FindEntityByBarcode finds an entity by its barcode
func (r *barcodeRepository) FindEntityByBarcode(ctx context.Context, orgID uuid.UUID, barcode string) (*types.BarcodeEntity, error) {
	query := `
		SELECT
			entity_type, entity_id, entity_name, additional_info
		FROM find_entity_by_barcode($1, $2)
		LIMIT 1
	`

	var entity types.BarcodeEntity
	var additionalInfo string
	err := r.db.QueryRowContext(ctx, query, barcode, orgID).Scan(
		&entity.EntityType, &entity.EntityID, &entity.EntityName, &additionalInfo,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	entity.AdditionalInfo = &additionalInfo
	return &entity, nil
}

// ValidateBarcodeFormat validates a barcode format
func (r *barcodeRepository) ValidateBarcodeFormat(ctx context.Context, barcode string) (bool, error) {
	query := `SELECT validate_barcode_format($1)`

	var valid bool
	err := r.db.QueryRowContext(ctx, query, barcode).Scan(&valid)
	if err != nil {
		return false, err
	}

	return valid, nil
}
