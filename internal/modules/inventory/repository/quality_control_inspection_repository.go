package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type qualityControlInspectionRepository struct {
	db *sql.DB
}

func NewQualityControlInspectionRepository(db *sql.DB) QualityControlInspectionRepository {
	return &qualityControlInspectionRepository{db: db}
}

func (r *qualityControlInspectionRepository) Create(ctx context.Context, inspection types.QualityControlInspection) (*types.QualityControlInspection, error) {
	query := `
		INSERT INTO quality_control_inspections
		(id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
		 $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32)
		RETURNING id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
	`

	if inspection.ID == uuid.Nil {
		inspection.ID = uuid.New()
	}
	if inspection.Reference == "" {
		inspection.Reference = fmt.Sprintf("QC-%s-%d", time.Now().Format("20060102"), time.Now().Unix())
	}
	if inspection.InspectionDate.IsZero() {
		inspection.InspectionDate = time.Now()
	}
	if inspection.CreatedAt.IsZero() {
		inspection.CreatedAt = time.Now()
	}
	if inspection.UpdatedAt.IsZero() {
		inspection.UpdatedAt = time.Now()
	}

	// Handle metadata
	metadataBytes, err := json.Marshal(inspection.Metadata)
	if err != nil {
		metadataBytes = []byte("{}")
	}

	var created types.QualityControlInspection
	err = r.db.QueryRowContext(ctx, query,
		inspection.ID, inspection.OrganizationID, inspection.CompanyID, inspection.Reference, inspection.InspectionType,
		inspection.SourceDocumentID, inspection.SourceType, inspection.ProductID, inspection.ProductName,
		inspection.LotID, inspection.SerialNumber, inspection.Quantity, inspection.UOMID, inspection.LocationID,
		inspection.LocationName, inspection.InspectionDate, inspection.InspectorID, inspection.InspectionMethod,
		inspection.SampleSize, inspection.Status, inspection.DefectType, inspection.DefectDescription,
		inspection.DefectQuantity, inspection.QualityRating, inspection.ComplianceNotes, inspection.Disposition,
		inspection.DispositionDate, inspection.DispositionBy, inspection.CreatedAt, inspection.UpdatedAt, metadataBytes,
	).Scan(
		&created.ID, &created.OrganizationID, &created.CompanyID, &created.Reference, &created.InspectionType,
		&created.SourceDocumentID, &created.SourceType, &created.ProductID, &created.ProductName,
		&created.LotID, &created.SerialNumber, &created.Quantity, &created.UOMID, &created.LocationID,
		&created.LocationName, &created.InspectionDate, &created.InspectorID, &created.InspectionMethod,
		&created.SampleSize, &created.Status, &created.DefectType, &created.DefectDescription,
		&created.DefectQuantity, &created.QualityRating, &created.ComplianceNotes, &created.Disposition,
		&created.DispositionDate, &created.DispositionBy, &created.CreatedAt, &created.UpdatedAt, &metadataBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality control inspection: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataBytes, &created.Metadata); err != nil {
		created.Metadata = make(map[string]interface{})
	}

	return &created, nil
}

func (r *qualityControlInspectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections WHERE id = $1 AND deleted_at IS NULL
	`

	var inspection types.QualityControlInspection
	var metadataBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
		&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
		&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
		&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
		&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
		&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
		&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspection: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
		inspection.Metadata = make(map[string]interface{})
	}

	return &inspection, nil
}

func (r *qualityControlInspectionRepository) FindAll(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY inspection_date DESC, created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspections: %w", err)
	}
	defer rows.Close()

	var inspections []types.QualityControlInspection
	for rows.Next() {
		var inspection types.QualityControlInspection
		var metadataBytes []byte
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
			&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
			&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
			&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
			&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
			&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
			&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
			inspection.Metadata = make(map[string]interface{})
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (r *qualityControlInspectionRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections
		WHERE organization_id = $1 AND product_id = $2 AND deleted_at IS NULL
		ORDER BY inspection_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspections by product: %w", err)
	}
	defer rows.Close()

	var inspections []types.QualityControlInspection
	for rows.Next() {
		var inspection types.QualityControlInspection
		var metadataBytes []byte
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
			&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
			&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
			&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
			&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
			&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
			&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
			inspection.Metadata = make(map[string]interface{})
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (r *qualityControlInspectionRepository) FindByLot(ctx context.Context, organizationID uuid.UUID, lotID uuid.UUID) ([]types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections
		WHERE organization_id = $1 AND lot_id = $2 AND deleted_at IS NULL
		ORDER BY inspection_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, lotID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspections by lot: %w", err)
	}
	defer rows.Close()

	var inspections []types.QualityControlInspection
	for rows.Next() {
		var inspection types.QualityControlInspection
		var metadataBytes []byte
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
			&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
			&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
			&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
			&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
			&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
			&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
			inspection.Metadata = make(map[string]interface{})
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (r *qualityControlInspectionRepository) FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections
		WHERE organization_id = $1 AND location_id = $2 AND deleted_at IS NULL
		ORDER BY inspection_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspections by location: %w", err)
	}
	defer rows.Close()

	var inspections []types.QualityControlInspection
	for rows.Next() {
		var inspection types.QualityControlInspection
		var metadataBytes []byte
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
			&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
			&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
			&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
			&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
			&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
			&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
			inspection.Metadata = make(map[string]interface{})
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (r *qualityControlInspectionRepository) FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections
		WHERE organization_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY inspection_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspections by status: %w", err)
	}
	defer rows.Close()

	var inspections []types.QualityControlInspection
	for rows.Next() {
		var inspection types.QualityControlInspection
		var metadataBytes []byte
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
			&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
			&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
			&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
			&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
			&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
			&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
			inspection.Metadata = make(map[string]interface{})
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (r *qualityControlInspectionRepository) FindByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]types.QualityControlInspection, error) {
	query := `
		SELECT id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
		FROM quality_control_inspections
		WHERE organization_id = $1 AND inspection_date BETWEEN $2 AND $3 AND deleted_at IS NULL
		ORDER BY inspection_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control inspections by date range: %w", err)
	}
	defer rows.Close()

	var inspections []types.QualityControlInspection
	for rows.Next() {
		var inspection types.QualityControlInspection
		var metadataBytes []byte
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.CompanyID, &inspection.Reference, &inspection.InspectionType,
			&inspection.SourceDocumentID, &inspection.SourceType, &inspection.ProductID, &inspection.ProductName,
			&inspection.LotID, &inspection.SerialNumber, &inspection.Quantity, &inspection.UOMID, &inspection.LocationID,
			&inspection.LocationName, &inspection.InspectionDate, &inspection.InspectorID, &inspection.InspectionMethod,
			&inspection.SampleSize, &inspection.Status, &inspection.DefectType, &inspection.DefectDescription,
			&inspection.DefectQuantity, &inspection.QualityRating, &inspection.ComplianceNotes, &inspection.Disposition,
			&inspection.DispositionDate, &inspection.DispositionBy, &inspection.CreatedAt, &inspection.UpdatedAt, &metadataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control inspection: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &inspection.Metadata); err != nil {
			inspection.Metadata = make(map[string]interface{})
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (r *qualityControlInspectionRepository) Update(ctx context.Context, inspection types.QualityControlInspection) (*types.QualityControlInspection, error) {
	query := `
		UPDATE quality_control_inspections
		SET company_id = $2, reference = $3, inspection_type = $4, source_document_id = $5, source_type = $6,
		 product_name = $7, serial_number = $8, quantity = $9, uom_id = $10, location_name = $11,
		 inspection_date = $12, inspector_id = $13, inspection_method = $14, sample_size = $15,
		 status = $16, defect_type = $17, defect_description = $18, defect_quantity = $19,
		 quality_rating = $20, compliance_notes = $21, disposition = $22, disposition_date = $23,
		 disposition_by = $24, updated_at = $25, metadata = $26
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, company_id, reference, inspection_type, source_document_id, source_type,
		 product_id, product_name, lot_id, serial_number, quantity, uom_id, location_id, location_name,
		 inspection_date, inspector_id, inspection_method, sample_size, status, defect_type,
		 defect_description, defect_quantity, quality_rating, compliance_notes, disposition,
		 disposition_date, disposition_by, created_at, updated_at, metadata
	`

	inspection.UpdatedAt = time.Now()
	metadataBytes, err := json.Marshal(inspection.Metadata)
	if err != nil {
		metadataBytes = []byte("{}")
	}

	var updated types.QualityControlInspection
	err = r.db.QueryRowContext(ctx, query,
		inspection.ID, inspection.CompanyID, inspection.Reference, inspection.InspectionType, inspection.SourceDocumentID,
		inspection.SourceType, inspection.ProductName, inspection.SerialNumber, inspection.Quantity, inspection.UOMID,
		inspection.LocationName, inspection.InspectionDate, inspection.InspectorID, inspection.InspectionMethod,
		inspection.SampleSize, inspection.Status, inspection.DefectType, inspection.DefectDescription,
		inspection.DefectQuantity, inspection.QualityRating, inspection.ComplianceNotes, inspection.Disposition,
		inspection.DispositionDate, inspection.DispositionBy, inspection.UpdatedAt, metadataBytes,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.CompanyID, &updated.Reference, &updated.InspectionType,
		&updated.SourceDocumentID, &updated.SourceType, &updated.ProductID, &updated.ProductName,
		&updated.LotID, &updated.SerialNumber, &updated.Quantity, &updated.UOMID, &updated.LocationID,
		&updated.LocationName, &updated.InspectionDate, &updated.InspectorID, &updated.InspectionMethod,
		&updated.SampleSize, &updated.Status, &updated.DefectType, &updated.DefectDescription,
		&updated.DefectQuantity, &updated.QualityRating, &updated.ComplianceNotes, &updated.Disposition,
		&updated.DispositionDate, &updated.DispositionBy, &updated.CreatedAt, &updated.UpdatedAt, &metadataBytes,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quality control inspection not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update quality control inspection: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataBytes, &updated.Metadata); err != nil {
		updated.Metadata = make(map[string]interface{})
	}

	return &updated, nil
}

func (r *qualityControlInspectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE quality_control_inspections SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete quality control inspection: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality control inspection not found")
	}
	return nil
}

func (r *qualityControlInspectionRepository) CreateFromStockMove(ctx context.Context, stockMoveID, inspectorID uuid.UUID, checklistID *uuid.UUID, inspectionMethod string, sampleSize *int) (*types.QualityControlInspection, error) {
	query := `SELECT * FROM create_qc_inspection_from_stock_move($1, $2, $3, $4, $5)`

	var inspectionID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, stockMoveID, inspectorID, checklistID, inspectionMethod, sampleSize).Scan(&inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality control inspection from stock move: %w", err)
	}

	// Retrieve the created inspection
	inspection, err := r.FindByID(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created inspection: %w", err)
	}
	if inspection == nil {
		return nil, fmt.Errorf("created inspection not found")
	}

	return inspection, nil
}

func (r *qualityControlInspectionRepository) UpdateStatus(ctx context.Context, inspectionID uuid.UUID, status, defectType, defectDescription string, defectQuantity *float64, qualityRating *int, complianceNotes, disposition *string) error {
	query := `SELECT update_qc_inspection_status($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query, inspectionID, status, defectType, defectDescription, defectQuantity, qualityRating, complianceNotes, disposition)
	if err != nil {
		return fmt.Errorf("failed to update quality control inspection status: %w", err)
	}

	return nil
}

func (r *qualityControlInspectionRepository) CompleteInspection(ctx context.Context, inspectionID uuid.UUID, status string, results []types.QualityControlInspectionItem) error {
	// Convert results to JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal inspection results: %w", err)
	}

	query := `SELECT complete_qc_inspection($1, $2, $3)`

	_, err = r.db.ExecContext(ctx, query, inspectionID, status, resultsJSON)
	if err != nil {
		return fmt.Errorf("failed to complete quality control inspection: %w", err)
	}

	return nil
}

func (r *qualityControlInspectionRepository) GetStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, productID *uuid.UUID) (types.QualityControlStatistics, error) {
	query := `SELECT get_quality_control_statistics($1, $2, $3, $4)`

	var statsJSON []byte
	err := r.db.QueryRowContext(ctx, query, organizationID, fromTime, toTime, productID).Scan(&statsJSON)
	if err != nil {
		return types.QualityControlStatistics{}, fmt.Errorf("failed to get quality control statistics: %w", err)
	}

	var stats types.QualityControlStatistics
	if err := json.Unmarshal(statsJSON, &stats); err != nil {
		return types.QualityControlStatistics{}, fmt.Errorf("failed to unmarshal quality control statistics: %w", err)
	}

	return stats, nil
}
