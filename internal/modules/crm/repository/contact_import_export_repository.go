package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
)

// ContactImportExportRepository handles import and export job tracking
type ContactImportExportRepository interface {
	// Import Jobs
	CreateImportJob(ctx context.Context, job *types.ContactImportJob) error
	GetImportJob(ctx context.Context, id uuid.UUID) (*types.ContactImportJob, error)
	UpdateImportJob(ctx context.Context, job *types.ContactImportJob) error
	ListImportJobs(ctx context.Context, filter types.ImportJobFilter) ([]*types.ContactImportJob, error)
	CountImportJobs(ctx context.Context, filter types.ImportJobFilter) (int, error)

	// Export Jobs
	CreateExportJob(ctx context.Context, job *types.ContactExportJob) error
	GetExportJob(ctx context.Context, id uuid.UUID) (*types.ContactExportJob, error)
	UpdateExportJob(ctx context.Context, job *types.ContactExportJob) error
	ListExportJobs(ctx context.Context, filter types.ExportJobFilter) ([]*types.ContactExportJob, error)
	CountExportJobs(ctx context.Context, filter types.ExportJobFilter) (int, error)
}

type contactImportExportRepository struct {
	db *sql.DB
}

func NewContactImportExportRepository(db *sql.DB) ContactImportExportRepository {
	return &contactImportExportRepository{db: db}
}

// CreateImportJob creates a new import job
func (r *contactImportExportRepository) CreateImportJob(ctx context.Context, job *types.ContactImportJob) error {
	query := `
		INSERT INTO contact_import_jobs (
			id, organization_id, job_id, filename, file_size, file_type,
			field_mapping, options, total_rows, processed_rows,
			successful_rows, failed_rows, status, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		job.ID,
		job.OrganizationID,
		job.JobID,
		job.Filename,
		job.FileSize,
		job.FileType,
		job.FieldMapping,
		job.Options,
		job.TotalRows,
		job.ProcessedRows,
		job.SuccessfulRows,
		job.FailedRows,
		job.Status,
		job.CreatedBy,
		now,
	).Scan(&job.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create import job: %w", err)
	}

	return nil
}

// GetImportJob retrieves an import job by ID
func (r *contactImportExportRepository) GetImportJob(ctx context.Context, id uuid.UUID) (*types.ContactImportJob, error) {
	query := `
		SELECT id, organization_id, job_id, filename, file_size, file_type,
		       field_mapping, options, total_rows, processed_rows,
		       successful_rows, failed_rows, status, error_message,
		       started_at, completed_at, created_at, created_by
		FROM contact_import_jobs
		WHERE id = $1
	`

	job := &types.ContactImportJob{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.OrganizationID,
		&job.JobID,
		&job.Filename,
		&job.FileSize,
		&job.FileType,
		&job.FieldMapping,
		&job.Options,
		&job.TotalRows,
		&job.ProcessedRows,
		&job.SuccessfulRows,
		&job.FailedRows,
		&job.Status,
		&job.ErrorMessage,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("import job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get import job: %w", err)
	}

	return job, nil
}

// UpdateImportJob updates an import job
func (r *contactImportExportRepository) UpdateImportJob(ctx context.Context, job *types.ContactImportJob) error {
	query := `
		UPDATE contact_import_jobs
		SET status = $1, total_rows = $2, processed_rows = $3,
		    successful_rows = $4, failed_rows = $5, error_message = $6,
		    started_at = $7, completed_at = $8
		WHERE id = $9
	`

	result, err := r.db.ExecContext(ctx, query,
		job.Status,
		job.TotalRows,
		job.ProcessedRows,
		job.SuccessfulRows,
		job.FailedRows,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update import job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("import job not found")
	}

	return nil
}

// ListImportJobs retrieves import jobs with filtering
func (r *contactImportExportRepository) ListImportJobs(ctx context.Context, filter types.ImportJobFilter) ([]*types.ContactImportJob, error) {
	query := `
		SELECT id, organization_id, job_id, filename, file_size, file_type,
		       field_mapping, options, total_rows, processed_rows,
		       successful_rows, failed_rows, status, error_message,
		       started_at, completed_at, created_at, created_by
		FROM contact_import_jobs
		WHERE organization_id = $1
	`

	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list import jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*types.ContactImportJob
	for rows.Next() {
		job := &types.ContactImportJob{}
		err := rows.Scan(
			&job.ID,
			&job.OrganizationID,
			&job.JobID,
			&job.Filename,
			&job.FileSize,
			&job.FileType,
			&job.FieldMapping,
			&job.Options,
			&job.TotalRows,
			&job.ProcessedRows,
			&job.SuccessfulRows,
			&job.FailedRows,
			&job.Status,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CountImportJobs counts import jobs matching the filter
func (r *contactImportExportRepository) CountImportJobs(ctx context.Context, filter types.ImportJobFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contact_import_jobs WHERE organization_id = $1`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count import jobs: %w", err)
	}

	return count, nil
}

// CreateExportJob creates a new export job
func (r *contactImportExportRepository) CreateExportJob(ctx context.Context, job *types.ContactExportJob) error {
	query := `
		INSERT INTO contact_export_jobs (
			id, organization_id, job_id, filter_criteria, selected_fields,
			format, total_contacts, status, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		job.ID,
		job.OrganizationID,
		job.JobID,
		job.FilterCriteria,
		job.SelectedFields,
		job.Format,
		job.TotalContacts,
		job.Status,
		job.CreatedBy,
		now,
	).Scan(&job.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create export job: %w", err)
	}

	return nil
}

// GetExportJob retrieves an export job by ID
func (r *contactImportExportRepository) GetExportJob(ctx context.Context, id uuid.UUID) (*types.ContactExportJob, error) {
	query := `
		SELECT id, organization_id, job_id, filter_criteria, selected_fields,
		       format, total_contacts, status, error_message, file_url,
		       started_at, completed_at, created_at, created_by
		FROM contact_export_jobs
		WHERE id = $1
	`

	job := &types.ContactExportJob{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.OrganizationID,
		&job.JobID,
		&job.FilterCriteria,
		&job.SelectedFields,
		&job.Format,
		&job.TotalContacts,
		&job.Status,
		&job.ErrorMessage,
		&job.FileURL,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export job: %w", err)
	}

	return job, nil
}

// UpdateExportJob updates an export job
func (r *contactImportExportRepository) UpdateExportJob(ctx context.Context, job *types.ContactExportJob) error {
	query := `
		UPDATE contact_export_jobs
		SET status = $1, total_contacts = $2, file_url = $3,
		    error_message = $4, started_at = $5, completed_at = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		job.Status,
		job.TotalContacts,
		job.FileURL,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update export job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("export job not found")
	}

	return nil
}

// ListExportJobs retrieves export jobs with filtering
func (r *contactImportExportRepository) ListExportJobs(ctx context.Context, filter types.ExportJobFilter) ([]*types.ContactExportJob, error) {
	query := `
		SELECT id, organization_id, job_id, filter_criteria, selected_fields,
		       format, total_contacts, status, error_message, file_url,
		       started_at, completed_at, created_at, created_by
		FROM contact_export_jobs
		WHERE organization_id = $1
	`

	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list export jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*types.ContactExportJob
	for rows.Next() {
		job := &types.ContactExportJob{}
		err := rows.Scan(
			&job.ID,
			&job.OrganizationID,
			&job.JobID,
			&job.FilterCriteria,
			&job.SelectedFields,
			&job.Format,
			&job.TotalContacts,
			&job.Status,
			&job.ErrorMessage,
			&job.FileURL,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CountExportJobs counts export jobs matching the filter
func (r *contactImportExportRepository) CountExportJobs(ctx context.Context, filter types.ExportJobFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contact_export_jobs WHERE organization_id = $1`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count export jobs: %w", err)
	}

	return count, nil
}
