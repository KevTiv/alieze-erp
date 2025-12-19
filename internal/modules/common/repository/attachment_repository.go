package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmi

	"alieze-erp/internal/modules/common/types"

	"github.com/google/uuid"
)

type AttachmentRepository interface {
	Create(ctx context.Context, attachment types.Attachment) (*types.Attachment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.Attachment, error)
	FindByChecksum(ctx context.Context, checksum string, organizationID uuid.UUID) (*types.Attachment, error)
	FindAll(ctx context.Context, filters types.AttachmentFilter) ([]types.Attachment, error)
	FindByResource(ctx context.Context, resModel string, resID uuid.UUID) ([]types.Attachment, error)
	Update(ctx context.Context, attachment types.Attachment) (*types.Attachment, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
	LogAccess(ctx context.Context, log types.AttachmentAccessLog) error
	GetStats(ctx context.Context, organizationID uuid.UUID) (*types.AttachmentStats, error)
	FindDuplicates(ctx context.Context, organizationID uuid.UUID, minCount int) ([]types.DuplicateAttachment, error)
}

type attachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(ctx context.Context, attachment types.Attachment) (*types.Attachment, error) {
	// Set defaults
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}
	if attachment.UploadedAt.IsZero() {
		attachment.UploadedAt = time.Now()
	}
	attachment.UpdatedAt = time.Now()
	attachment.Version = 1
	attachment.IsDeleted = false

	// Serialize metadata to JSON
	var metadataJSON []byte
	var err error
	if attachment.Metadata != nil {
		metadataJSON, err = json.Marshal(attachment.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO attachments
		(id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
	`

	var created types.Attachment
	var metadataStr sql.NullString
	err = r.db.QueryRowContext(ctx, query,
		attachment.ID, attachment.OrganizationID, attachment.Name, attachment.Description,
		attachment.ResModel, attachment.ResID, attachment.ResName, attachment.FileSize,
		attachment.MimeType, attachment.AttachmentType, attachment.Checksum, attachment.StorageKey,
		attachment.AccessType, attachment.AccessToken, attachment.TokenExpiresAt,
		attachment.UploadedBy, attachment.UploadedAt, attachment.UpdatedAt, attachment.Version,
		attachment.IsDeleted, metadataJSON,
	).Scan(
		&created.ID, &created.OrganizationID, &created.Name, &created.Description,
		&created.ResModel, &created.ResID, &created.ResName, &created.FileSize,
		&created.MimeType, &created.AttachmentType, &created.Checksum, &created.StorageKey,
		&created.AccessType, &created.AccessToken, &created.TokenExpiresAt,
		&created.UploadedBy, &created.UploadedAt, &created.UpdatedAt, &created.Version,
		&created.IsDeleted, &created.DeletedAt, &created.DeletedBy, &metadataStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	// Deserialize metadata
	if metadataStr.Valid && metadataStr.String != "" {
		if err := json.Unmarshal([]byte(metadataStr.String), &created.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &created, nil
}

func (r *attachmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Attachment, error) {
	query := `
		SELECT id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
		FROM attachments
		WHERE id = $1
	`

	var attachment types.Attachment
	var metadataStr sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&attachment.ID, &attachment.OrganizationID, &attachment.Name, &attachment.Description,
		&attachment.ResModel, &attachment.ResID, &attachment.ResName, &attachment.FileSize,
		&attachment.MimeType, &attachment.AttachmentType, &attachment.Checksum, &attachment.StorageKey,
		&attachment.AccessType, &attachment.AccessToken, &attachment.TokenExpiresAt,
		&attachment.UploadedBy, &attachment.UploadedAt, &attachment.UpdatedAt, &attachment.Version,
		&attachment.IsDeleted, &attachment.DeletedAt, &attachment.DeletedBy, &metadataStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find attachment: %w", err)
	}

	// Deserialize metadata
	if metadataStr.Valid && metadataStr.String != "" {
		if err := json.Unmarshal([]byte(metadataStr.String), &attachment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &attachment, nil
}

func (r *attachmentRepository) FindByChecksum(ctx context.Context, checksum string, organizationID uuid.UUID) (*types.Attachment, error) {
	query := `
		SELECT id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
		FROM attachments
		WHERE checksum = $1 AND organization_id = $2 AND is_deleted = false
		LIMIT 1
	`

	var attachment types.Attachment
	var metadataStr sql.NullString
	err := r.db.QueryRowContext(ctx, query, checksum, organizationID).Scan(
		&attachment.ID, &attachment.OrganizationID, &attachment.Name, &attachment.Description,
		&attachment.ResModel, &attachment.ResID, &attachment.ResName, &attachment.FileSize,
		&attachment.MimeType, &attachment.AttachmentType, &attachment.Checksum, &attachment.StorageKey,
		&attachment.AccessType, &attachment.AccessToken, &attachment.TokenExpiresAt,
		&attachment.UploadedBy, &attachment.UploadedAt, &attachment.UpdatedAt, &attachment.Version,
		&attachment.IsDeleted, &attachment.DeletedAt, &attachment.DeletedBy, &metadataStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find attachment by checksum: %w", err)
	}

	// Deserialize metadata
	if metadataStr.Valid && metadataStr.String != "" {
		if err := json.Unmarshal([]byte(metadataStr.String), &attachment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &attachment, nil
}

func (r *attachmentRepository) FindAll(ctx context.Context, filters types.AttachmentFilter) ([]types.Attachment, error) {
	query := `
		SELECT id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
		FROM attachments
		WHERE 1=1
	`

	var params []interface{}
	paramIndex := 1

	// Build WHERE clause
	if filters.OrganizationID != nil {
		query += fmt.Sprintf(" AND organization_id = $%d", paramIndex)
		params = append(params, *filters.OrganizationID)
		paramIndex++
	}

	if filters.ResModel != nil {
		query += fmt.Sprintf(" AND res_model = $%d", paramIndex)
		params = append(params, *filters.ResModel)
		paramIndex++
	}

	if filters.ResID != nil {
		query += fmt.Sprintf(" AND res_id = $%d", paramIndex)
		params = append(params, *filters.ResID)
		paramIndex++
	}

	if filters.AttachmentType != nil {
		query += fmt.Sprintf(" AND attachment_type = $%d", paramIndex)
		params = append(params, *filters.AttachmentType)
		paramIndex++
	}

	if filters.AccessType != nil {
		query += fmt.Sprintf(" AND access_type = $%d", paramIndex)
		params = append(params, *filters.AccessType)
		paramIndex++
	}

	if filters.UploadedBy != nil {
		query += fmt.Sprintf(" AND uploaded_by = $%d", paramIndex)
		params = append(params, *filters.UploadedBy)
		paramIndex++
	}

	if !filters.IncludeDeleted {
		query += " AND is_deleted = false"
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND uploaded_at >= $%d", paramIndex)
		params = append(params, *filters.DateFrom)
		paramIndex++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND uploaded_at <= $%d", paramIndex)
		params = append(params, *filters.DateTo)
		paramIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY uploaded_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		params = append(params, filters.Limit)
		paramIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramIndex)
		params = append(params, filters.Offset)
		paramIndex++
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var attachments []types.Attachment
	for rows.Next() {
		var attachment types.Attachment
		var metadataStr sql.NullString
		err = rows.Scan(
			&attachment.ID, &attachment.OrganizationID, &attachment.Name, &attachment.Description,
			&attachment.ResModel, &attachment.ResID, &attachment.ResName, &attachment.FileSize,
			&attachment.MimeType, &attachment.AttachmentType, &attachment.Checksum, &attachment.StorageKey,
			&attachment.AccessType, &attachment.AccessToken, &attachment.TokenExpiresAt,
			&attachment.UploadedBy, &attachment.UploadedAt, &attachment.UpdatedAt, &attachment.Version,
			&attachment.IsDeleted, &attachment.DeletedAt, &attachment.DeletedBy, &metadataStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}

		// Deserialize metadata
		if metadataStr.Valid && metadataStr.String != "" {
			if err := json.Unmarshal([]byte(metadataStr.String), &attachment.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (r *attachmentRepository) FindByResource(ctx context.Context, resModel string, resID uuid.UUID) ([]types.Attachment, error) {
	query := `
		SELECT id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
		FROM attachments
		WHERE res_model = $1 AND res_id = $2 AND is_deleted = false
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, resModel, resID)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments by resource: %w", err)
	}
	defer rows.Close()

	var attachments []types.Attachment
	for rows.Next() {
		var attachment types.Attachment
		var metadataStr sql.NullString
		err = rows.Scan(
			&attachment.ID, &attachment.OrganizationID, &attachment.Name, &attachment.Description,
			&attachment.ResModel, &attachment.ResID, &attachment.ResName, &attachment.FileSize,
			&attachment.MimeType, &attachment.AttachmentType, &attachment.Checksum, &attachment.StorageKey,
			&attachment.AccessType, &attachment.AccessToken, &attachment.TokenExpiresAt,
			&attachment.UploadedBy, &attachment.UploadedAt, &attachment.UpdatedAt, &attachment.Version,
			&attachment.IsDeleted, &attachment.DeletedAt, &attachment.DeletedBy, &metadataStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}

		// Deserialize metadata
		if metadataStr.Valid && metadataStr.String != "" {
			if err := json.Unmarshal([]byte(metadataStr.String), &attachment.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (r *attachmentRepository) Update(ctx context.Context, attachment types.Attachment) (*types.Attachment, error) {
	attachment.UpdatedAt = time.Now()
	attachment.Version++

	// Serialize metadata
	var metadataJSON []byte
	var err error
	if attachment.Metadata != nil {
		metadataJSON, err = json.Marshal(attachment.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE attachments
		SET name = $1, description = $2, res_model = $3, res_id = $4, res_name = $5,
		    access_type = $6, access_token = $7, token_expires_at = $8,
		    updated_at = $9, version = $10, metadata = $11
		WHERE id = $12
		RETURNING id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
	`

	var updated types.Attachment
	var metadataStr sql.NullString
	err = r.db.QueryRowContext(ctx, query,
		attachment.Name, attachment.Description, attachment.ResModel, attachment.ResID,
		attachment.ResName, attachment.AccessType, attachment.AccessToken,
		attachment.TokenExpiresAt, attachment.UpdatedAt, attachment.Version,
		metadataJSON, attachment.ID,
	).Scan(
		&updated.ID, &updated.OrganizationID, &updated.Name, &updated.Description,
		&updated.ResModel, &updated.ResID, &updated.ResName, &updated.FileSize,
		&updated.MimeType, &updated.AttachmentType, &updated.Checksum, &updated.StorageKey,
		&updated.AccessType, &updated.AccessToken, &updated.TokenExpiresAt,
		&updated.UploadedBy, &updated.UploadedAt, &updated.UpdatedAt, &updated.Version,
		&updated.IsDeleted, &updated.DeletedAt, &updated.DeletedBy, &metadataStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update attachment: %w", err)
	}

	// Deserialize metadata
	if metadataStr.Valid && metadataStr.String != "" {
		if err := json.Unmarshal([]byte(metadataStr.String), &updated.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &updated, nil
}

func (r *attachmentRepository) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE attachments
		SET is_deleted = true, deleted_at = $1, deleted_by = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, now, deletedBy, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete attachment: %w", err)
	}

	return nil
}

func (r *attachmentRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM attachments WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete attachment: %w", err)
	}

	return nil
}

func (r *attachmentRepository) LogAccess(ctx context.Context, log types.AttachmentAccessLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.AccessedAt.IsZero() {
		log.AccessedAt = time.Now()
	}

	query := `
		INSERT INTO attachment_access_logs
		(id, attachment_id, accessed_by, accessed_at, access_method, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.AttachmentID, log.AccessedBy, log.AccessedAt,
		log.AccessMethod, log.IPAddress, log.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to log attachment access: %w", err)
	}

	return nil
}

func (r *attachmentRepository) GetStats(ctx context.Context, organizationID uuid.UUID) (*types.AttachmentStats, error) {
	stats := &types.AttachmentStats{
		ByType:  make(map[string]int64),
		ByModel: make(map[string]int64),
	}

	// Total attachments and size
	query := `
		SELECT COUNT(*), COALESCE(SUM(file_size), 0)
		FROM attachments
		WHERE organization_id = $1 AND is_deleted = false
	`
	err := r.db.QueryRowContext(ctx, query, organizationID).Scan(&stats.TotalAttachments, &stats.TotalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get total stats: %w", err)
	}

	// By type
	query = `
		SELECT attachment_type, COUNT(*)
		FROM attachments
		WHERE organization_id = $1 AND is_deleted = false
		GROUP BY attachment_type
	`
	rows, err := r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var attachmentType string
		var count int64
		if err := rows.Scan(&attachmentType, &count); err != nil {
			return nil, err
		}
		stats.ByType[attachmentType] = count
	}

	// By model
	query = `
		SELECT res_model, COUNT(*)
		FROM attachments
		WHERE organization_id = $1 AND is_deleted = false
		GROUP BY res_model
	`
	rows, err = r.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by model: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var resModel string
		var count int64
		if err := rows.Scan(&resModel, &count); err != nil {
			return nil, err
		}
		stats.ByModel[resModel] = count
	}

	// Recent uploads (last 24 hours)
	query = `
		SELECT COUNT(*)
		FROM attachments
		WHERE organization_id = $1 AND is_deleted = false
		  AND uploaded_at >= NOW() - INTERVAL '24 hours'
	`
	err = r.db.QueryRowContext(ctx, query, organizationID).Scan(&stats.RecentUploads)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent uploads: %w", err)
	}

	// Unique files and duplicates based on checksum
	query = `
		SELECT
			COUNT(DISTINCT checksum) as unique_files,
			COUNT(*) - COUNT(DISTINCT checksum) as duplicates
		FROM attachments
		WHERE organization_id = $1 AND is_deleted = false
	`
	err = r.db.QueryRowContext(ctx, query, organizationID).Scan(&stats.UniqueFiles, &stats.DuplicateFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get deduplication stats: %w", err)
	}

	// Space saved from deduplication
	query = `
		SELECT COALESCE(SUM(file_size * (count - 1)), 0) as space_saved
		FROM (
			SELECT file_size, COUNT(*) as count
			FROM attachments
			WHERE organization_id = $1 AND is_deleted = false
			GROUP BY checksum, file_size
			HAVING COUNT(*) > 1
		) duplicates
	`
	err = r.db.QueryRowContext(ctx, query, organizationID).Scan(&stats.SpaceSaved)
	if err != nil {
		return nil, fmt.Errorf("failed to get space saved: %w", err)
	}

	return stats, nil
}

func (r *attachmentRepository) FindDuplicates(ctx context.Context, organizationID uuid.UUID, minCount int) ([]types.DuplicateAttachment, error) {
	query := `
		SELECT
			checksum,
			COUNT(*) as count,
			SUM(file_size) as total_size,
			SUM(file_size) - MAX(file_size) as space_saved
		FROM attachments
		WHERE organization_id = $1 AND is_deleted = false
		GROUP BY checksum
		HAVING COUNT(*) >= $2
		ORDER BY count DESC, total_size DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID, minCount)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}
	defer rows.Close()

	var duplicates []types.DuplicateAttachment
	for rows.Next() {
		var dup types.DuplicateAttachment
		if err := rows.Scan(&dup.Checksum, &dup.Count, &dup.TotalSize, &dup.SpaceSaved); err != nil {
			return nil, err
		}

		// Get the actual attachments for this checksum
		attachments, err := r.findByChecksumAll(ctx, dup.Checksum, organizationID)
		if err != nil {
			return nil, err
		}
		dup.Attachments = attachments

		duplicates = append(duplicates, dup)
	}

	return duplicates, nil
}

func (r *attachmentRepository) findByChecksumAll(ctx context.Context, checksum string, organizationID uuid.UUID) ([]types.Attachment, error) {
	query := `
		SELECT id, organization_id, name, description, res_model, res_id, res_name,
		 file_size, mime_type, attachment_type, checksum, storage_key,
		 access_type, access_token, token_expires_at, uploaded_by, uploaded_at,
		 updated_at, version, is_deleted, deleted_at, deleted_by, metadata
		FROM attachments
		WHERE checksum = $1 AND organization_id = $2 AND is_deleted = false
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, checksum, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments by checksum: %w", err)
	}
	defer rows.Close()

	var attachments []types.Attachment
	for rows.Next() {
		var attachment types.Attachment
		var metadataStr sql.NullString
		err = rows.Scan(
			&attachment.ID, &attachment.OrganizationID, &attachment.Name, &attachment.Description,
			&attachment.ResModel, &attachment.ResID, &attachment.ResName, &attachment.FileSize,
			&attachment.MimeType, &attachment.AttachmentType, &attachment.Checksum, &attachment.StorageKey,
			&attachment.AccessType, &attachment.AccessToken, &attachment.TokenExpiresAt,
			&attachment.UploadedBy, &attachment.UploadedAt, &attachment.UpdatedAt, &attachment.Version,
			&attachment.IsDeleted, &attachment.DeletedAt, &attachment.DeletedBy, &metadataStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}

		// Deserialize metadata
		if metadataStr.Valid && metadataStr.String != "" {
			if err := json.Unmarshal([]byte(metadataStr.String), &attachment.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}
