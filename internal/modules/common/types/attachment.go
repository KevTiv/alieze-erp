package types

import (
	"time"

	"github.com/google/uuid"
)

// AttachmentType represents the type of attachment
type AttachmentType string

const (
	AttachmentTypeDocument   AttachmentType = "document"
	AttachmentTypeImage      AttachmentType = "image"
	AttachmentTypeSpreadsheet AttachmentType = "spreadsheet"
	AttachmentTypePDF        AttachmentType = "pdf"
	AttachmentTypeVideo      AttachmentType = "video"
	AttachmentTypeAudio      AttachmentType = "audio"
	AttachmentTypeArchive    AttachmentType = "archive"
	AttachmentTypeOther      AttachmentType = "other"
)

// AttachmentAccessType represents access control for attachments
type AttachmentAccessType string

const (
	AttachmentAccessPrivate AttachmentAccessType = "private"
	AttachmentAccessPublic  AttachmentAccessType = "public"
	AttachmentAccessShared  AttachmentAccessType = "shared"
)

// Attachment represents a file attachment in the system
// Uses Odoo-compatible polymorphic reference pattern (res_model/res_id)
type Attachment struct {
	ID             uuid.UUID            `json:"id" db:"id"`
	OrganizationID uuid.UUID            `json:"organization_id" db:"organization_id"`
	Name           string               `json:"name" db:"name"`
	Description    string               `json:"description" db:"description"`
	ResModel       string               `json:"res_model" db:"res_model"`         // e.g., 'leads', 'contacts', 'sales_orders'
	ResID          uuid.UUID            `json:"res_id" db:"res_id"`               // ID of the related record
	ResName        string               `json:"res_name,omitempty" db:"res_name"` // Optional name of related record for display
	FileSize       int64                `json:"file_size" db:"file_size"`
	MimeType       string               `json:"mime_type" db:"mime_type"`
	AttachmentType AttachmentType       `json:"attachment_type" db:"attachment_type"`
	Checksum       string               `json:"checksum" db:"checksum"`         // SHA256 for deduplication
	StorageKey     string               `json:"storage_key" db:"storage_key"`   // Key in storage system (S3/MinIO)
	AccessType     AttachmentAccessType `json:"access_type" db:"access_type"`
	AccessToken    *string              `json:"access_token,omitempty" db:"access_token"` // For public access
	TokenExpiresAt *time.Time           `json:"token_expires_at,omitempty" db:"token_expires_at"`
	UploadedBy     uuid.UUID            `json:"uploaded_by" db:"uploaded_by"`
	UploadedAt     time.Time            `json:"uploaded_at" db:"uploaded_at"`
	UpdatedAt      time.Time            `json:"updated_at" db:"updated_at"`
	Version        int                  `json:"version" db:"version"`      // For versioning support
	IsDeleted      bool                 `json:"is_deleted" db:"is_deleted"` // Soft delete
	DeletedAt      *time.Time           `json:"deleted_at,omitempty" db:"deleted_at"`
	DeletedBy      *uuid.UUID           `json:"deleted_by,omitempty" db:"deleted_by"`
	// Metadata stored as JSONB in database
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// AttachmentAccessLog tracks access to attachments for analytics
type AttachmentAccessLog struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	AttachmentID uuid.UUID  `json:"attachment_id" db:"attachment_id"`
	AccessedBy   *uuid.UUID `json:"accessed_by,omitempty" db:"accessed_by"` // Null if public access
	AccessedAt   time.Time  `json:"accessed_at" db:"accessed_at"`
	AccessMethod string     `json:"access_method" db:"access_method"` // 'download', 'view', 'preview'
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
}

// AttachmentFilter for querying attachments
type AttachmentFilter struct {
	OrganizationID *uuid.UUID
	ResModel       *string
	ResID          *uuid.UUID
	AttachmentType *AttachmentType
	AccessType     *AttachmentAccessType
	UploadedBy     *uuid.UUID
	IncludeDeleted bool
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}

// AttachmentUploadRequest represents a request to upload a file
type AttachmentUploadRequest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	ResModel       string                 `json:"res_model"`
	ResID          uuid.UUID              `json:"res_id"`
	AccessType     AttachmentAccessType   `json:"access_type"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	FileData       []byte                 `json:"-"` // Binary file data
	MimeType       string                 `json:"mime_type"`
	FileSize       int64                  `json:"file_size"`
}

// AttachmentDownloadResponse contains attachment data for download
type AttachmentDownloadResponse struct {
	Attachment *Attachment
	FileData   []byte
	MimeType   string
	Filename   string
}

// AttachmentStats contains statistics about attachments
type AttachmentStats struct {
	TotalAttachments int64              `json:"total_attachments"`
	TotalSize        int64              `json:"total_size_bytes"`
	ByType           map[string]int64   `json:"by_type"`
	ByModel          map[string]int64   `json:"by_model"`
	RecentUploads    int64              `json:"recent_uploads_24h"`
	UniqueFiles      int64              `json:"unique_files"` // Based on checksum
	DuplicateFiles   int64              `json:"duplicate_files"`
	SpaceSaved       int64              `json:"space_saved_bytes"` // From deduplication
}

// DuplicateAttachment represents a potential duplicate file
type DuplicateAttachment struct {
	Checksum      string        `json:"checksum"`
	Count         int           `json:"count"`
	TotalSize     int64         `json:"total_size_bytes"`
	SpaceSaved    int64         `json:"space_saved_bytes"`
	Attachments   []Attachment  `json:"attachments"`
}

// GetAttachmentType determines the attachment type from MIME type
func GetAttachmentType(mimeType string) AttachmentType {
	switch {
	case mimeType == "application/pdf":
		return AttachmentTypePDF
	case mimeType[:6] == "image/":
		return AttachmentTypeImage
	case mimeType[:6] == "video/":
		return AttachmentTypeVideo
	case mimeType[:6] == "audio/":
		return AttachmentTypeAudio
	case mimeType == "application/vnd.ms-excel" ||
		 mimeType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
		 mimeType == "text/csv":
		return AttachmentTypeSpreadsheet
	case mimeType == "application/msword" ||
		 mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
		 mimeType == "text/plain":
		return AttachmentTypeDocument
	case mimeType == "application/zip" ||
		 mimeType == "application/x-rar-compressed" ||
		 mimeType == "application/x-tar" ||
		 mimeType == "application/gzip":
		return AttachmentTypeArchive
	default:
		return AttachmentTypeOther
	}
}

// IsImage checks if the attachment is an image
func (a *Attachment) IsImage() bool {
	return a.AttachmentType == AttachmentTypeImage
}

// IsPDF checks if the attachment is a PDF
func (a *Attachment) IsPDF() bool {
	return a.AttachmentType == AttachmentTypePDF
}

// IsPublic checks if the attachment has public access
func (a *Attachment) IsPublic() bool {
	return a.AccessType == AttachmentAccessPublic
}

// HasValidToken checks if the attachment has a valid access token
func (a *Attachment) HasValidToken() bool {
	if a.AccessToken == nil || a.TokenExpiresAt == nil {
		return false
	}
	return time.Now().Before(*a.TokenExpiresAt)
}
