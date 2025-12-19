package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"alieze-erp/internal/modules/common/repository"
	"alieze-erp/internal/modules/common/types"
	"alieze-erp/pkg/events"
	"alieze-erp/pkg/storage"

	"github.com/google/uuid"
)

type AttachmentService struct {
	repo     repository.AttachmentRepository
	storage  storage.Storage
	eventBus *events.Bus
}

func NewAttachmentService(repo repository.AttachmentRepository, storage storage.Storage) *AttachmentService {
	return &AttachmentService{
		repo:    repo,
		storage: storage,
	}
}

func NewAttachmentServiceWithEventBus(repo repository.AttachmentRepository, storage storage.Storage, eventBus *events.Bus) *AttachmentService {
	service := NewAttachmentService(repo, storage)
	service.eventBus = eventBus
	return service
}

// Upload uploads a file and creates an attachment record
func (s *AttachmentService) Upload(ctx context.Context, req types.AttachmentUploadRequest, uploadedBy uuid.UUID) (*types.Attachment, error) {
	// Calculate checksum
	checksum := calculateChecksum(req.FileData)

	// Check for existing file with same checksum (deduplication)
	existing, err := s.repo.FindByChecksum(ctx, checksum, req.ResID) // Using ResID as org for now
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	var storageKey string
	if existing != nil {
		// File already exists, reuse storage key
		storageKey = existing.StorageKey
	} else {
		// Upload new file to storage
		storageKey, err = s.uploadToStorage(ctx, req.FileData, req.Name, req.MimeType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload to storage: %w", err)
		}
	}

	// Determine attachment type
	attachmentType := types.GetAttachmentType(req.MimeType)

	// Create attachment record
	attachment := types.Attachment{
		ID:             uuid.New(),
		OrganizationID: req.ResID, // This should be actual org ID
		Name:           req.Name,
		Description:    req.Description,
		ResModel:       req.ResModel,
		ResID:          req.ResID,
		FileSize:       req.FileSize,
		MimeType:       req.MimeType,
		AttachmentType: attachmentType,
		Checksum:       checksum,
		StorageKey:     storageKey,
		AccessType:     req.AccessType,
		UploadedBy:     uploadedBy,
		Metadata:       req.Metadata,
	}

	// Generate access token if public
	if req.AccessType == types.AttachmentAccessPublic {
		token, expiresAt, err := s.generateAccessToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate access token: %w", err)
		}
		attachment.AccessToken = &token
		attachment.TokenExpiresAt = &expiresAt
	}

	created, err := s.repo.Create(ctx, attachment)
	if err != nil {
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "attachment.uploaded", map[string]interface{}{
		"attachment_id": created.ID,
		"res_model":     created.ResModel,
		"res_id":        created.ResID,
		"file_size":     created.FileSize,
		"is_duplicate":  existing != nil,
	})

	return created, nil
}

// Download downloads a file from storage
func (s *AttachmentService) Download(ctx context.Context, id uuid.UUID, accessedBy *uuid.UUID) (*types.AttachmentDownloadResponse, error) {
	// Get attachment record
	attachment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment == nil {
		return nil, fmt.Errorf("attachment not found")
	}

	// Check if deleted
	if attachment.IsDeleted {
		return nil, fmt.Errorf("attachment has been deleted")
	}

	// Download from storage
	file, err := s.storage.Download(ctx, attachment.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download from storage: %w", err)
	}

	// Log access
	accessLog := types.AttachmentAccessLog{
		AttachmentID: attachment.ID,
		AccessedBy:   accessedBy,
		AccessMethod: "download",
	}
	if err := s.repo.LogAccess(ctx, accessLog); err != nil {
		// Log error but don't fail download
		fmt.Printf("Failed to log attachment access: %v\n", err)
	}

	// Publish event
	s.publishEvent(ctx, "attachment.downloaded", map[string]interface{}{
		"attachment_id": attachment.ID,
		"accessed_by":   accessedBy,
	})

	return &types.AttachmentDownloadResponse{
		Attachment: attachment,
		FileData:   file.Data,
		MimeType:   attachment.MimeType,
		Filename:   attachment.Name,
	}, nil
}

// GetPublicURL generates a time-limited public URL for an attachment
func (s *AttachmentService) GetPublicURL(ctx context.Context, id uuid.UUID, expiry time.Duration) (string, error) {
	// Get attachment record
	attachment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment == nil {
		return "", fmt.Errorf("attachment not found")
	}

	// Get presigned URL from storage
	url, err := s.storage.GetURL(ctx, attachment.StorageKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to get public URL: %w", err)
	}

	return url, nil
}

// ListByResource lists all attachments for a specific resource
func (s *AttachmentService) ListByResource(ctx context.Context, resModel string, resID uuid.UUID) ([]types.Attachment, error) {
	attachments, err := s.repo.FindByResource(ctx, resModel, resID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attachments: %w", err)
	}

	return attachments, nil
}

// List lists attachments with filters
func (s *AttachmentService) List(ctx context.Context, filters types.AttachmentFilter) ([]types.Attachment, error) {
	attachments, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list attachments: %w", err)
	}

	return attachments, nil
}

// Update updates an attachment metadata
func (s *AttachmentService) Update(ctx context.Context, attachment types.Attachment) (*types.Attachment, error) {
	updated, err := s.repo.Update(ctx, attachment)
	if err != nil {
		return nil, fmt.Errorf("failed to update attachment: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "attachment.updated", map[string]interface{}{
		"attachment_id": updated.ID,
	})

	return updated, nil
}

// Delete soft deletes an attachment
func (s *AttachmentService) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// Get attachment to publish event
	attachment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment == nil {
		return fmt.Errorf("attachment not found")
	}

	err = s.repo.SoftDelete(ctx, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "attachment.deleted", map[string]interface{}{
		"attachment_id": id,
		"storage_key":   attachment.StorageKey,
	})

	return nil
}

// HardDelete permanently deletes an attachment and its file
func (s *AttachmentService) HardDelete(ctx context.Context, id uuid.UUID) error {
	// Get attachment to get storage key
	attachment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment == nil {
		return fmt.Errorf("attachment not found")
	}

	// Check if other attachments use the same storage key
	duplicates, err := s.repo.FindByChecksum(ctx, attachment.Checksum, attachment.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}

	// Delete from database
	err = s.repo.HardDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete attachment from database: %w", err)
	}

	// Only delete from storage if no other attachments reference it
	if duplicates == nil || duplicates.ID == id {
		if err := s.storage.Delete(ctx, attachment.StorageKey); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to delete file from storage: %v\n", err)
		}
	}

	return nil
}

// GetStats retrieves attachment statistics
func (s *AttachmentService) GetStats(ctx context.Context, organizationID uuid.UUID) (*types.AttachmentStats, error) {
	stats, err := s.repo.GetStats(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment stats: %w", err)
	}

	return stats, nil
}

// FindDuplicates finds duplicate files based on checksum
func (s *AttachmentService) FindDuplicates(ctx context.Context, organizationID uuid.UUID, minCount int) ([]types.DuplicateAttachment, error) {
	duplicates, err := s.repo.FindDuplicates(ctx, organizationID, minCount)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}

	return duplicates, nil
}

// CleanupOrphanedFiles removes files from storage that have no attachment records
func (s *AttachmentService) CleanupOrphanedFiles(ctx context.Context, organizationID uuid.UUID) (int, error) {
	// This is a maintenance operation that would:
	// 1. List all files in storage
	// 2. Check if each file has a corresponding attachment record
	// 3. Delete files without records
	// For now, returning 0 as placeholder
	return 0, nil
}

// RegenerateAccessToken generates a new access token for an attachment
func (s *AttachmentService) RegenerateAccessToken(ctx context.Context, id uuid.UUID) (*types.Attachment, error) {
	attachment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment == nil {
		return nil, fmt.Errorf("attachment not found")
	}

	token, expiresAt, err := s.generateAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	attachment.AccessToken = &token
	attachment.TokenExpiresAt = &expiresAt

	updated, err := s.repo.Update(ctx, *attachment)
	if err != nil {
		return nil, fmt.Errorf("failed to update attachment: %w", err)
	}

	return updated, nil
}

// Helper functions

func (s *AttachmentService) uploadToStorage(ctx context.Context, data []byte, filename, mimeType string) (string, error) {
	// Generate storage key
	storageKey := fmt.Sprintf("attachments/%s/%s", time.Now().Format("2006/01/02"), uuid.New().String())

	// Upload to storage
	_, err := s.storage.Upload(ctx, storage.UploadOptions{
		Key:         storageKey,
		Data:        data,
		ContentType: mimeType,
		Metadata: map[string]string{
			"original_filename": filename,
			"uploaded_at":       time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", err
	}

	return storageKey, nil
}

func (s *AttachmentService) generateAccessToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	return token, expiresAt, nil
}

func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *AttachmentService) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if s.eventBus != nil {
		if err := s.eventBus.Publish(ctx, eventType, payload); err != nil {
			fmt.Printf("Failed to publish event %s: %v\n", eventType, err)
		}
	}
}
