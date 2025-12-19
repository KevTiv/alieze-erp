package repository

import (
	"context"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// QualityControlInspectionRepository interface
type QualityControlInspectionRepository interface {
	Create(ctx context.Context, inspection types.QualityControlInspection) (*types.QualityControlInspection, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.QualityControlInspection, error)
	FindAll(ctx context.Context, organizationID uuid.UUID, limit int) ([]types.QualityControlInspection, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.QualityControlInspection, error)
	FindByLot(ctx context.Context, organizationID uuid.UUID, lotID uuid.UUID) ([]types.QualityControlInspection, error)
	FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]types.QualityControlInspection, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.QualityControlInspection, error)
	FindByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]types.QualityControlInspection, error)
	Update(ctx context.Context, inspection types.QualityControlInspection) (*types.QualityControlInspection, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Business logic methods
	CreateFromStockMove(ctx context.Context, stockMoveID, inspectorID uuid.UUID, checklistID *uuid.UUID, inspectionMethod string, sampleSize *int) (*types.QualityControlInspection, error)
	UpdateStatus(ctx context.Context, inspectionID uuid.UUID, status, defectType, defectDescription string, defectQuantity *float64, qualityRating *int, complianceNotes, disposition *string) error
	CompleteInspection(ctx context.Context, inspectionID uuid.UUID, status string, results []types.QualityControlInspectionItem) error
	GetStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, productID *uuid.UUID) (types.QualityControlStatistics, error)
}

// QualityControlChecklistRepository interface
type QualityControlChecklistRepository interface {
	Create(ctx context.Context, checklist types.QualityControlChecklist) (*types.QualityControlChecklist, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.QualityControlChecklist, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlChecklist, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.QualityControlChecklist, error)
	FindByCategory(ctx context.Context, organizationID, categoryID uuid.UUID) ([]types.QualityControlChecklist, error)
	FindByInspectionType(ctx context.Context, organizationID uuid.UUID, inspectionType string) ([]types.QualityControlChecklist, error)
	FindActive(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlChecklist, error)
	Update(ctx context.Context, checklist types.QualityControlChecklist) (*types.QualityControlChecklist, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// QualityChecklistItemRepository interface
type QualityChecklistItemRepository interface {
	Create(ctx context.Context, item types.QualityChecklistItem) (*types.QualityChecklistItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.QualityChecklistItem, error)
	FindByChecklist(ctx context.Context, checklistID uuid.UUID) ([]types.QualityChecklistItem, error)
	FindActiveByChecklist(ctx context.Context, checklistID uuid.UUID) ([]types.QualityChecklistItem, error)
	Update(ctx context.Context, item types.QualityChecklistItem) (*types.QualityChecklistItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByChecklist(ctx context.Context, checklistID uuid.UUID) error
}

// QualityControlInspectionItemRepository interface
type QualityControlInspectionItemRepository interface {
	Create(ctx context.Context, item types.QualityControlInspectionItem) (*types.QualityControlInspectionItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.QualityControlInspectionItem, error)
	FindByInspection(ctx context.Context, inspectionID uuid.UUID) ([]types.QualityControlInspectionItem, error)
	Update(ctx context.Context, item types.QualityControlInspectionItem) (*types.QualityControlInspectionItem, error)
	UpdateResult(ctx context.Context, itemID uuid.UUID, result, notes string) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByInspection(ctx context.Context, inspectionID uuid.UUID) error
}

// QualityControlAlertRepository interface
type QualityControlAlertRepository interface {
	Create(ctx context.Context, alert types.QualityControlAlert) (*types.QualityControlAlert, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.QualityControlAlert, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlAlert, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]types.QualityControlAlert, error)
	FindBySeverity(ctx context.Context, organizationID uuid.UUID, severity string) ([]types.QualityControlAlert, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]types.QualityControlAlert, error)
	FindOpen(ctx context.Context, organizationID uuid.UUID) ([]types.QualityControlAlert, error)
	Update(ctx context.Context, alert types.QualityControlAlert) (*types.QualityControlAlert, error)
	UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedBy *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Business logic methods
	CreateFromInspection(ctx context.Context, inspectionID uuid.UUID, alertType, severity, title, message string) (*types.QualityControlAlert, error)
}
