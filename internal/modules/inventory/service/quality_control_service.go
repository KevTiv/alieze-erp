package service

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// QualityControlService handles business logic for quality control operations
type QualityControlService struct {
	inspectionRepo      repository.QualityControlInspectionRepository
	checklistRepo       repository.QualityControlChecklistRepository
	checklistItemRepo   repository.QualityChecklistItemRepository
	inspectionItemRepo  repository.QualityControlInspectionItemRepository
	alertRepo           repository.QualityControlAlertRepository
	inventoryRepo      repository.InventoryRepository
}

// NewQualityControlService creates a new QualityControlService instance
func NewQualityControlService(
	inspectionRepo repository.QualityControlInspectionRepository,
	checklistRepo repository.QualityControlChecklistRepository,
	checklistItemRepo repository.QualityChecklistItemRepository,
	inspectionItemRepo repository.QualityControlInspectionItemRepository,
	alertRepo repository.QualityControlAlertRepository,
	inventoryRepo repository.InventoryRepository,
) *QualityControlService {
	return &QualityControlService{
		inspectionRepo:      inspectionRepo,
		checklistRepo:       checklistRepo,
		checklistItemRepo:   checklistItemRepo,
		inspectionItemRepo:  inspectionItemRepo,
		alertRepo:           alertRepo,
		inventoryRepo:      inventoryRepo,
	}
}

// Inspection Management

func (s *QualityControlService) CreateInspection(ctx context.Context, inspection types.QualityControlInspection) (*types.QualityControlInspection, error) {
	if inspection.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if inspection.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product_id is required")
	}
	if inspection.LocationID == uuid.Nil {
		return nil, fmt.Errorf("location_id is required")
	}
	if inspection.InspectionType == "" {
		inspection.InspectionType = "incoming"
	}
	if inspection.InspectionMethod == "" {
		inspection.InspectionMethod = "visual"
	}
	if inspection.Status == "" {
		inspection.Status = "pending"
	}
	if inspection.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Get product and location names for the inspection
	product, err := s.inventoryRepo.GetProductByID(ctx, inspection.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}
	inspection.ProductName = product.Name

	location, err := s.inventoryRepo.GetLocation(ctx, inspection.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	if location == nil {
		return nil, fmt.Errorf("location not found")
	}
	inspection.LocationName = location.Name

	return s.inspectionRepo.Create(ctx, inspection)
}

func (s *QualityControlService) GetInspection(ctx context.Context, id uuid.UUID) (*types.QualityControlInspection, error) {
	inspection, err := s.inspectionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inspection == nil {
		return nil, fmt.Errorf("quality control inspection not found")
	}
	return inspection, nil
}

func (s *QualityControlService) ListInspections(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.QualityControlInspection, error) {
	return s.inspectionRepo.FindAll(ctx, organizationID, limit)
}

func (s *QualityControlService) ListInspectionsByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.QualityControlInspection, error) {
	return s.inspectionRepo.FindByProduct(ctx, organizationID, productID)
}

func (s *QualityControlService) ListInspectionsByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.QualityControlInspection, error) {
	return s.inspectionRepo.FindByStatus(ctx, organizationID, status)
}

func (s *QualityControlService) UpdateInspection(ctx context.Context, inspection types.QualityControlInspection) (*types.QualityControlInspection, error) {
	return s.inspectionRepo.Update(ctx, inspection)
}

func (s *QualityControlService) DeleteInspection(ctx context.Context, id uuid.UUID) error {
	return s.inspectionRepo.Delete(ctx, id)
}

// Checklist Management

func (s *QualityControlService) CreateChecklist(ctx context.Context, checklist types.QualityControlChecklist) (*types.QualityControlChecklist, error) {
	if checklist.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if checklist.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if checklist.InspectionType == "" {
		checklist.InspectionType = "incoming"
	}
	if checklist.Priority == 0 {
		checklist.Priority = 10
	}

	return s.checklistRepo.Create(ctx, checklist)
}

func (s *QualityControlService) GetChecklist(ctx context.Context, id uuid.UUID) (*types.QualityControlChecklist, error) {
	checklist, err := s.checklistRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if checklist == nil {
		return nil, fmt.Errorf("quality control checklist not found")
	}
	return checklist, nil
}

func (s *QualityControlService) ListChecklists(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlChecklist, error) {
	return s.checklistRepo.FindAll(ctx, organizationID)
}

func (s *QualityControlService) ListActiveChecklists(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlChecklist, error) {
	return s.checklistRepo.FindActive(ctx, organizationID)
}

func (s *QualityControlService) UpdateChecklist(ctx context.Context, checklist types.QualityControlChecklist) (*types.QualityControlChecklist, error) {
	return s.checklistRepo.Update(ctx, checklist)
}

func (s *QualityControlService) DeleteChecklist(ctx context.Context, id uuid.UUID) error {
	return s.checklistRepo.Delete(ctx, id)
}

// Checklist Item Management

func (s *QualityControlService) CreateChecklistItem(ctx context.Context, item types.QualityChecklistItem) (*types.QualityChecklistItem, error) {
	if item.ChecklistID == uuid.Nil {
		return nil, fmt.Errorf("checklist_id is required")
	}
	if item.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if item.Sequence == 0 {
		item.Sequence = 10
	}

	return s.checklistItemRepo.Create(ctx, item)
}

func (s *QualityControlService) GetChecklistItem(ctx context.Context, id uuid.UUID) (*types.QualityChecklistItem, error) {
	item, err := s.checklistItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("quality control checklist item not found")
	}
	return item, nil
}

func (s *QualityControlService) ListChecklistItems(ctx context.Context, checklistID uuid.UUID) ([]types.QualityChecklistItem, error) {
	return s.checklistItemRepo.FindByChecklist(ctx, checklistID)
}

func (s *QualityControlService) ListActiveChecklistItems(ctx context.Context, checklistID uuid.UUID) ([]types.QualityChecklistItem, error) {
	return s.checklistItemRepo.FindActiveByChecklist(ctx, checklistID)
}

func (s *QualityControlService) UpdateChecklistItem(ctx context.Context, item types.QualityChecklistItem) (*types.QualityChecklistItem, error) {
	return s.checklistItemRepo.Update(ctx, item)
}

func (s *QualityControlService) DeleteChecklistItem(ctx context.Context, id uuid.UUID) error {
	return s.checklistItemRepo.Delete(ctx, id)
}

// Inspection Item Management

func (s *QualityControlService) CreateInspectionItem(ctx context.Context, item types.QualityControlInspectionItem) (*types.QualityControlInspectionItem, error) {
	if item.InspectionID == uuid.Nil {
		return nil, fmt.Errorf("inspection_id is required")
	}
	if item.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if item.Result == "" {
		item.Result = "pending"
	}

	return s.inspectionItemRepo.Create(ctx, item)
}

func (s *QualityControlService) GetInspectionItem(ctx context.Context, id uuid.UUID) (*types.QualityControlInspectionItem, error) {
	item, err := s.inspectionItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("quality control inspection item not found")
	}
	return item, nil
}

func (s *QualityControlService) ListInspectionItems(ctx context.Context, inspectionID uuid.UUID) ([]types.QualityControlInspectionItem, error) {
	return s.inspectionItemRepo.FindByInspection(ctx, inspectionID)
}

func (s *QualityControlService) UpdateInspectionItem(ctx context.Context, item types.QualityControlInspectionItem) (*types.QualityControlInspectionItem, error) {
	return s.inspectionItemRepo.Update(ctx, item)
}

func (s *QualityControlService) UpdateInspectionItemResult(ctx context.Context, itemID uuid.UUID, result, notes string) error {
	return s.inspectionItemRepo.UpdateResult(ctx, itemID, result, notes)
}

func (s *QualityControlService) DeleteInspectionItem(ctx context.Context, id uuid.UUID) error {
	return s.inspectionItemRepo.Delete(ctx, id)
}

// Alert Management

func (s *QualityControlService) CreateAlert(ctx context.Context, alert types.QualityControlAlert) (*types.QualityControlAlert, error) {
	if alert.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if alert.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if alert.Message == "" {
		return nil, fmt.Errorf("message is required")
	}
	if alert.AlertType == "" {
		return nil, fmt.Errorf("alert_type is required")
	}
	if alert.Severity == "" {
		return nil, fmt.Errorf("severity is required")
	}

	return s.alertRepo.Create(ctx, alert)
}

func (s *QualityControlService) GetAlert(ctx context.Context, id uuid.UUID) (*types.QualityControlAlert, error) {
	alert, err := s.alertRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, fmt.Errorf("quality control alert not found")
	}
	return alert, nil
}

func (s *QualityControlService) ListAlerts(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlAlert, error) {
	return s.alertRepo.FindAll(ctx, organizationID)
}

func (s *QualityControlService) ListOpenAlerts(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlAlert, error) {
	return s.alertRepo.FindOpen(ctx, organizationID)
}

func (s *QualityControlService) UpdateAlert(ctx context.Context, alert types.QualityControlAlert) (*types.QualityControlAlert, error) {
	return s.alertRepo.Update(ctx, alert)
}

func (s *QualityControlService) UpdateAlertStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedBy *uuid.UUID) error {
	return s.alertRepo.UpdateStatus(ctx, alertID, status, resolvedBy)
}

func (s *QualityControlService) DeleteAlert(ctx context.Context, id uuid.UUID) error {
	return s.alertRepo.Delete(ctx, id)
}

// Business Logic Methods

func (s *QualityControlService) CreateInspectionFromStockMove(ctx context.Context, stockMoveID, inspectorID uuid.UUID, checklistID *uuid.UUID, inspectionMethod string, sampleSize *int) (*types.QualityControlInspection, error) {
	// Validate stock move exists
	stockMove, err := s.inventoryRepo.GetMove(ctx, stockMoveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock move: %w", err)
	}
	if stockMove == nil {
		return nil, fmt.Errorf("stock move not found")
	}

	return s.inspectionRepo.CreateFromStockMove(ctx, stockMoveID, inspectorID, checklistID, inspectionMethod, sampleSize)
}

func (s *QualityControlService) UpdateInspectionStatus(ctx context.Context, inspectionID uuid.UUID, status, defectType, defectDescription string, defectQuantity *float64, qualityRating *int, complianceNotes, disposition *string) error {
	// Validate inspection exists
	inspection, err := s.GetInspection(ctx, inspectionID)
	if err != nil {
		return fmt.Errorf("failed to get inspection: %w", err)
	}

	// If status is changing to failed or quarantined, create an alert
	if (status == "failed" || status == "quarantined") && inspection.Status != status {
		alertTitle := fmt.Sprintf("Quality Issue: %s - %s", inspection.ProductName, status)
		alertMessage := fmt.Sprintf("Product %s (Lot: %v) has quality status: %s",
			inspection.ProductName, inspection.LotID, status)

		if defectType != "" {
			alertMessage += fmt.Sprintf("\nDefect Type: %s", defectType)
		}
		if defectDescription != "" {
			alertMessage += fmt.Sprintf("\nDescription: %s", defectDescription)
		}

		severity := "medium"
		if status == "failed" {
			severity = "high"
		}

		_, err = s.alertRepo.CreateFromInspection(ctx, inspectionID, "defect", severity, alertTitle, alertMessage)
		if err != nil {
			// Log error but don't fail the status update
			fmt.Printf("Failed to create quality alert: %v\n", err)
		}
	}

	return s.inspectionRepo.UpdateStatus(ctx, inspectionID, status, defectType, defectDescription, defectQuantity, qualityRating, complianceNotes, disposition)
}

func (s *QualityControlService) CompleteInspection(ctx context.Context, inspectionID uuid.UUID, status string, results []types.QualityControlInspectionItem) error {
	// Validate inspection exists
	inspection, err := s.GetInspection(ctx, inspectionID)
	if err != nil {
		return fmt.Errorf("failed to get inspection: %w", err)
	}

	// Complete the inspection
	err = s.inspectionRepo.CompleteInspection(ctx, inspectionID, status, results)
	if err != nil {
		return fmt.Errorf("failed to complete inspection: %w", err)
	}

	// If inspection failed, create an alert
	if status == "failed" || status == "quarantined" {
		failedCount := 0
		for _, result := range results {
			if result.Result == "fail" {
				failedCount++
			}
		}

		alertTitle := fmt.Sprintf("Quality Inspection %s: %d/%d items failed", status, failedCount, len(results))
		alertMessage := fmt.Sprintf("Product %s quality inspection %s with %d failed items out of %d total",
			inspection.ProductName, status, failedCount, len(results))

		severity := "medium"
		if status == "failed" {
			severity = "high"
		}

		_, err = s.alertRepo.CreateFromInspection(ctx, inspectionID, "defect", severity, alertTitle, alertMessage)
		if err != nil {
			// Log error but don't fail the completion
			fmt.Printf("Failed to create quality alert: %v\n", err)
		}
	}

	return nil
}

func (s *QualityControlService) GetQualityControlStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, productID *uuid.UUID) (types.QualityControlStatistics, error) {
	return s.inspectionRepo.GetStatistics(ctx, organizationID, fromTime, toTime, productID)
}

func (s *QualityControlService) CreateAlertFromInspection(ctx context.Context, inspectionID uuid.UUID, alertType, severity, title, message string) (*types.QualityControlAlert, error) {
	// Validate inspection exists
	_, err := s.GetInspection(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}

	return s.alertRepo.CreateFromInspection(ctx, inspectionID, alertType, severity, title, message)
}

// Quality Control Workflow Methods

func (s *QualityControlService) StartQualityControlWorkflow(ctx context.Context, stockMoveID, inspectorID uuid.UUID) (*types.QualityControlInspection, error) {
	// 1. Get stock move details
	stockMove, err := s.inventoryRepo.GetMove(ctx, stockMoveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock move: %w", err)
	}
	if stockMove == nil {
		return nil, fmt.Errorf("stock move not found")
	}

	// 2. Find appropriate checklist for this product
	checklists, err := s.checklistRepo.FindByProduct(ctx, stockMove.OrganizationID, stockMove.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to find checklists: %w", err)
	}

	// 3. Use the highest priority active checklist, or nil if none found
	var checklistID *uuid.UUID
	for _, checklist := range checklists {
		if checklist.Active && checklist.InspectionType == "incoming" {
			checklistID = &checklist.ID
			break // Use the first active one (they're ordered by priority)
		}
	}

	// 4. Create inspection from stock move
	return s.CreateInspectionFromStockMove(ctx, stockMoveID, inspectorID, checklistID, "visual", nil)
}

func (s *QualityControlService) ProcessQualityControlResult(ctx context.Context, inspectionID uuid.UUID, results []types.QualityControlInspectionItem) (*types.QualityControlInspection, error) {
	// 1. Update all inspection items with results
	for _, result := range results {
		err := s.inspectionItemRepo.UpdateResult(ctx, result.ID, result.Result, result.Notes)
		if err != nil {
			return nil, fmt.Errorf("failed to update inspection item result: %w", err)
		}
	}

	// 2. Determine overall status based on results
	failedCount := 0
	for _, result := range results {
		if result.Result == "fail" {
			failedCount++
		}
	}

	status := "passed"
	if failedCount > 0 {
		if failedCount <= len(results)/10 { // 10% or less failed
			status = "quarantined"
		} else {
			status = "failed"
		}
	}

	// 3. Complete the inspection
	err := s.CompleteInspection(ctx, inspectionID, status, results)
	if err != nil {
		return nil, fmt.Errorf("failed to complete inspection: %w", err)
	}

	// 4. Return the updated inspection
	return s.GetInspection(ctx, inspectionID)
}

func (s *QualityControlService) HandleQualityControlDisposition(ctx context.Context, inspectionID uuid.UUID, disposition string) (*types.QualityControlInspection, error) {
	// Validate disposition
	switch disposition {
	case "accept", "reject", "rework", "scrap", "return":
		// Valid dispositions
	default:
		return nil, fmt.Errorf("invalid disposition: %s", disposition)
	}

	// Get current inspection
	inspection, err := s.GetInspection(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}

	// Update status based on disposition
	status := "passed"
	if disposition == "reject" || disposition == "scrap" {
		status = "rejected"
	} else if disposition == "rework" {
		status = "quarantined"
	}

	// Update the inspection
	err = s.UpdateInspectionStatus(ctx, inspectionID, status, inspection.DefectType, inspection.DefectDescription,
		inspection.DefectQuantity, inspection.QualityRating, inspection.ComplianceNotes, &disposition)
	if err != nil {
		return nil, fmt.Errorf("failed to update inspection status: %w", err)
	}

	// Handle stock movement based on disposition
	err = s.handleStockDisposition(ctx, inspection, disposition)
	if err != nil {
		return nil, fmt.Errorf("failed to handle stock disposition: %w", err)
	}

	// Return updated inspection
	return s.GetInspection(ctx, inspectionID)
}

func (s *QualityControlService) handleStockDisposition(ctx context.Context, inspection *types.QualityControlInspection, disposition string) error {
	// This would handle actual stock movements based on QC results
	// For now, we'll just log the action
	fmt.Printf("Quality control disposition handled: Inspection %s -> %s\n", inspection.ID, disposition)

	// In a real implementation, this would:
	// 1. Move stock to appropriate locations (quarantine, scrap, etc.)
	// 2. Update stock quantities
	// 3. Create appropriate stock moves
	// 4. Handle any financial transactions (write-offs, etc.)

	return nil
}

// Quality Control Monitoring

func (s *QualityControlService) GetQualityControlDashboard(ctx context.Context, organizationID uuid.UUID) (map[string]interface{}, error) {
	// Get statistics
	stats, err := s.GetQualityControlStatistics(ctx, organizationID, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Get recent inspections
	recentInspections, err := s.ListInspections(ctx, organizationID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent inspections: %w", err)
	}

	// Get open alerts
	openAlerts, err := s.ListOpenAlerts(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get open alerts: %w", err)
	}

	// Get pending inspections
	pendingInspections, err := s.ListInspectionsByStatus(ctx, organizationID, "pending")
	if err != nil {
		return nil, fmt.Errorf("failed to get pending inspections: %w", err)
	}

	return map[string]interface{}{
		"statistics":           stats,
		"recent_inspections":   recentInspections,
		"open_alerts":          openAlerts,
		"pending_inspections":  pendingInspections,
		"quality_trend":        "stable", // Would be calculated in real implementation
		"last_updated":         time.Now(),
	}, nil
}
