package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type BarcodeService struct {
	barcodeRepo repository.BarcodeRepository
}

func NewBarcodeService(barcodeRepo repository.BarcodeRepository) *BarcodeService {
	return &BarcodeService{
		barcodeRepo: barcodeRepo,
	}
}

// ScanBarcode performs a barcode scan and returns the result
func (s *BarcodeService) ScanBarcode(ctx context.Context, request domain.BarcodeScanRequest) (*domain.BarcodeScanResponse, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.UserID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	if request.Barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}

	return s.barcodeRepo.ScanBarcode(ctx, request)
}

// GetScanByID retrieves a specific barcode scan
func (s *BarcodeService) GetScanByID(ctx context.Context, orgID uuid.UUID, scanID uuid.UUID) (*domain.BarcodeScan, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if scanID == uuid.Nil {
		return nil, fmt.Errorf("scan_id is required")
	}

	return s.barcodeRepo.GetScanByID(ctx, orgID, scanID)
}

// ListScans retrieves barcode scans for an organization
func (s *BarcodeService) ListScans(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.BarcodeScan, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.barcodeRepo.ListScans(ctx, orgID, limit, offset)
}

// CreateScanningSession creates a new mobile scanning session
func (s *BarcodeService) CreateScanningSession(ctx context.Context, request domain.CreateScanningSessionRequest) (*domain.MobileScanningSession, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.UserID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	if request.SessionType == "" {
		return nil, fmt.Errorf("session_type is required")
	}

	return s.barcodeRepo.CreateScanningSession(ctx, request)
}

// GetScanningSession retrieves a scanning session
func (s *BarcodeService) GetScanningSession(ctx context.Context, orgID, sessionID uuid.UUID) (*domain.MobileScanningSession, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}

	return s.barcodeRepo.GetScanningSession(ctx, orgID, sessionID)
}

// ListScanningSessions retrieves scanning sessions for an organization
func (s *BarcodeService) ListScanningSessions(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]domain.MobileScanningSession, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.barcodeRepo.ListScanningSessions(ctx, orgID, status, limit, offset)
}

// AddScanToSession adds a scan to an existing session
func (s *BarcodeService) AddScanToSession(ctx context.Context, request domain.AddScanToSessionRequest) (*domain.BarcodeScanResponse, error) {
	if request.SessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.UserID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	if request.Barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}

	return s.barcodeRepo.AddScanToSession(ctx, request)
}

// CompleteScanningSession marks a session as completed
func (s *BarcodeService) CompleteScanningSession(ctx context.Context, request domain.CompleteSessionRequest) (bool, error) {
	if request.SessionID == uuid.Nil {
		return false, fmt.Errorf("session_id is required")
	}
	if request.OrganizationID == uuid.Nil {
		return false, fmt.Errorf("organization_id is required")
	}

	return s.barcodeRepo.CompleteScanningSession(ctx, request)
}

// GetSessionLines retrieves all lines for a scanning session
func (s *BarcodeService) GetSessionLines(ctx context.Context, orgID, sessionID uuid.UUID) ([]domain.MobileScanningSessionLine, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}

	return s.barcodeRepo.GetSessionLines(ctx, orgID, sessionID)
}

// GenerateBarcode generates a barcode for an entity
func (s *BarcodeService) GenerateBarcode(ctx context.Context, request domain.BarcodeGenerationRequest) (*domain.BarcodeGenerationResponse, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.EntityID == uuid.Nil {
		return nil, fmt.Errorf("entity_id is required")
	}
	if request.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}

	return s.barcodeRepo.GenerateBarcode(ctx, request)
}

// GenerateBarcodesForProducts generates barcodes for multiple products
func (s *BarcodeService) GenerateBarcodesForProducts(ctx context.Context, orgID uuid.UUID, productIDs []uuid.UUID, prefix *string) (map[uuid.UUID]string, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if len(productIDs) == 0 {
		return nil, fmt.Errorf("product_ids is required")
	}

	return s.barcodeRepo.GenerateBarcodesForProducts(ctx, orgID, productIDs, prefix)
}

// FindEntityByBarcode finds an entity by its barcode
func (s *BarcodeService) FindEntityByBarcode(ctx context.Context, orgID uuid.UUID, barcode string) (*domain.BarcodeEntity, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}

	return s.barcodeRepo.FindEntityByBarcode(ctx, orgID, barcode)
}

// ValidateBarcodeFormat validates a barcode format
func (s *BarcodeService) ValidateBarcodeFormat(ctx context.Context, barcode string) (bool, error) {
	if barcode == "" {
		return false, fmt.Errorf("barcode is required")
	}

	return s.barcodeRepo.ValidateBarcodeFormat(ctx, barcode)
}

// GenerateBarcodeLabel generates a printable barcode label
func (s *BarcodeService) GenerateBarcodeLabel(ctx context.Context, orgID uuid.UUID, entityType, entityID, format string) (string, error) {
	// This would integrate with a barcode label printing service
	// For now, return a simple JSON representation
	if orgID == uuid.Nil {
		return "", fmt.Errorf("organization_id is required")
	}
	if entityID == "" {
		return "", fmt.Errorf("entity_id is required")
	}
	if entityType == "" {
		return "", fmt.Errorf("entity_type is required")
	}

	// Get entity details
	entity, err := s.FindEntityByBarcode(ctx, orgID, entityID)
	if err != nil {
		return "", err
	}

	if entity == nil {
		return "", fmt.Errorf("entity not found")
	}

	// Generate label data (would be sent to printing service)
	labelData := map[string]interface{}{
		"organization_id": orgID.String(),
		"entity_type":     entityType,
		"entity_id":       entityID,
		"entity_name":     entity.EntityName,
		"barcode":         entityID, // In real implementation, this would be the actual barcode
		"format":          format,
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	// Return JSON representation
	return fmt.Sprintf("Label data for %s %s: %v", entityType, entityID, labelData), nil
}
