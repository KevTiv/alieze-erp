package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type qualityControlAlertRepository struct {
	db *sql.DB
}

func NewQualityControlAlertRepository(db *sql.DB) QualityControlAlertRepository {
	return &qualityControlAlertRepository{db: db}
}

func (r *qualityControlAlertRepository) Create(ctx context.Context, alert domain.QualityControlAlert) (*domain.QualityControlAlert, error) {
	query := `
		INSERT INTO quality_control_alerts
		(id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, organization_id, alert_type, severity, title, message, related_inspection_id,
		 product_id, status, created_at, updated_at, resolved_at, resolved_by
	`

	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}
	if alert.UpdatedAt.IsZero() {
		alert.UpdatedAt = time.Now()
	}
	if alert.Status == "" {
		alert.Status = "open"
	}

	var created domain.QualityControlAlert
	err := r.db.QueryRowContext(ctx, query,
		alert.ID, alert.OrganizationID, alert.AlertType, alert.Severity, alert.Title, alert.Message,
		alert.RelatedInspectionID, alert.ProductID, alert.Status, alert.CreatedAt, alert.UpdatedAt,
		alert.ResolvedAt, alert.ResolvedBy,
	).Scan(
		&created.ID, &created.OrganizationID, &created.AlertType, &created.Severity, &created.Title,
		&created.Message, &created.RelatedInspectionID, &created.ProductID, &created.Status,
		&created.CreatedAt, &created.UpdatedAt, &created.ResolvedAt, &created.ResolvedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create quality control alert: %w", err)
	}

	return &created, nil
}

func (r *qualityControlAlertRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityControlAlert, error) {
	query := `
		SELECT id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by
		FROM quality_control_alerts WHERE id = $1
	`

	var alert domain.QualityControlAlert
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.OrganizationID, &alert.AlertType, &alert.Severity, &alert.Title,
		&alert.Message, &alert.RelatedInspectionID, &alert.ProductID, &alert.Status,
		&alert.CreatedAt, &alert.UpdatedAt, &alert.ResolvedAt, &alert.ResolvedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control alert: %w", err)
	}

	return &alert, nil
}

func (r *qualityControlAlertRepository) FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.QualityControlAlert, error) {
	query := `
		SELECT id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by
		FROM quality_control_alerts WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.QualityControlAlert
	for rows.Next() {
		var alert domain.QualityControlAlert
		err := rows.Scan(
			&alert.ID, &alert.OrganizationID, &alert.AlertType, &alert.Severity, &alert.Title,
			&alert.Message, &alert.RelatedInspectionID, &alert.ProductID, &alert.Status,
			&alert.CreatedAt, &alert.UpdatedAt, &alert.ResolvedAt, &alert.ResolvedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control alert: %w", err)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *qualityControlAlertRepository) FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]domain.QualityControlAlert, error) {
	query := `
		SELECT id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by
		FROM quality_control_alerts WHERE organization_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control alerts by status: %w", err)
	}
	defer rows.Close()

	var alerts []domain.QualityControlAlert
	for rows.Next() {
		var alert domain.QualityControlAlert
		err := rows.Scan(
			&alert.ID, &alert.OrganizationID, &alert.AlertType, &alert.Severity, &alert.Title,
			&alert.Message, &alert.RelatedInspectionID, &alert.ProductID, &alert.Status,
			&alert.CreatedAt, &alert.UpdatedAt, &alert.ResolvedAt, &alert.ResolvedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control alert: %w", err)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *qualityControlAlertRepository) FindBySeverity(ctx context.Context, organizationID uuid.UUID, severity string) ([]domain.QualityControlAlert, error) {
	query := `
		SELECT id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by
		FROM quality_control_alerts WHERE organization_id = $1 AND severity = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, severity)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control alerts by severity: %w", err)
	}
	defer rows.Close()

	var alerts []domain.QualityControlAlert
	for rows.Next() {
		var alert domain.QualityControlAlert
		err := rows.Scan(
			&alert.ID, &alert.OrganizationID, &alert.AlertType, &alert.Severity, &alert.Title,
			&alert.Message, &alert.RelatedInspectionID, &alert.ProductID, &alert.Status,
			&alert.CreatedAt, &alert.UpdatedAt, &alert.ResolvedAt, &alert.ResolvedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control alert: %w", err)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *qualityControlAlertRepository) FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.QualityControlAlert, error) {
	query := `
		SELECT id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by
		FROM quality_control_alerts WHERE organization_id = $1 AND product_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality control alerts by product: %w", err)
	}
	defer rows.Close()

	var alerts []domain.QualityControlAlert
	for rows.Next() {
		var alert domain.QualityControlAlert
		err := rows.Scan(
			&alert.ID, &alert.OrganizationID, &alert.AlertType, &alert.Severity, &alert.Title,
			&alert.Message, &alert.RelatedInspectionID, &alert.ProductID, &alert.Status,
			&alert.CreatedAt, &alert.UpdatedAt, &alert.ResolvedAt, &alert.ResolvedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control alert: %w", err)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *qualityControlAlertRepository) FindOpen(ctx context.Context, organizationID uuid.UUID) ([]domain.QualityControlAlert, error) {
	query := `
		SELECT id, organization_id, alert_type, severity, title, message, related_inspection_id, product_id,
		 status, created_at, updated_at, resolved_at, resolved_by
		FROM quality_control_alerts WHERE organization_id = $1 AND status = 'open'
		ORDER BY severity DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find open quality control alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.QualityControlAlert
	for rows.Next() {
		var alert domain.QualityControlAlert
		err := rows.Scan(
			&alert.ID, &alert.OrganizationID, &alert.AlertType, &alert.Severity, &alert.Title,
			&alert.Message, &alert.RelatedInspectionID, &alert.ProductID, &alert.Status,
			&alert.CreatedAt, &alert.UpdatedAt, &alert.ResolvedAt, &alert.ResolvedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quality control alert: %w", err)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *qualityControlAlertRepository) Update(ctx context.Context, alert domain.QualityControlAlert) (*domain.QualityControlAlert, error) {
	query := `
		UPDATE quality_control_alerts
		SET alert_type = $2, severity = $3, title = $4, message = $5, related_inspection_id = $6,
		 product_id = $7, status = $8, updated_at = $9, resolved_at = $10, resolved_by = $11
		WHERE id = $1
		RETURNING id, organization_id, alert_type, severity, title, message, related_inspection_id,
		 product_id, status, created_at, updated_at, resolved_at, resolved_by
	`

	alert.UpdatedAt = time.Now()
	var updated domain.QualityControlAlert
	err := r.db.QueryRowContext(ctx, query,
		alert.ID, alert.AlertType, alert.Severity, alert.Title, alert.Message, alert.RelatedInspectionID,
		alert.ProductID, alert.Status, alert.UpdatedAt, alert.ResolvedAt, alert.ResolvedBy,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.AlertType, &updated.Severity, &updated.Title,
		&updated.Message, &updated.RelatedInspectionID, &updated.ProductID, &updated.Status,
		&updated.CreatedAt, &updated.UpdatedAt, &updated.ResolvedAt, &updated.ResolvedBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quality control alert not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update quality control alert: %w", err)
	}

	return &updated, nil
}

func (r *qualityControlAlertRepository) UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedBy *uuid.UUID) error {
	query := `UPDATE quality_control_alerts SET status = $2, resolved_at = $3, resolved_by = $4, updated_at = $5 WHERE id = $1`

	resolvedAt := time.Now()
	if status != "resolved" && status != "closed" {
		resolvedAt = time.Time{}
	}

	result, err := r.db.ExecContext(ctx, query, alertID, status, resolvedAt, resolvedBy, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update quality control alert status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality control alert not found")
	}
	return nil
}

func (r *qualityControlAlertRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM quality_control_alerts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quality control alert: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality control alert not found")
	}
	return nil
}

func (r *qualityControlAlertRepository) CreateFromInspection(ctx context.Context, inspectionID uuid.UUID, alertType, severity, title, message string) (*domain.QualityControlAlert, error) {
	// Get inspection details to populate alert
	var productID uuid.UUID
	err := r.db.QueryRowContext(ctx, `
		SELECT product_id FROM quality_control_inspections WHERE id = $1
	`, inspectionID).Scan(&productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspection details: %w", err)
	}

	// Create the alert
	alert := domain.QualityControlAlert{
		OrganizationID:          uuid.Nil, // Will be set by the function
		AlertType:               alertType,
		Severity:                severity,
		Title:                   title,
		Message:                 message,
		RelatedInspectionID:     &inspectionID,
		ProductID:               &productID,
		Status:                  "open",
	}

	return r.Create(ctx, alert)
}
