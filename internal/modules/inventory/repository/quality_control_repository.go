package repository

import (
	"context"
	"time"

	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

// QualityControlInspectionRepository interface
type QualityControlInspectionRepository interface {
	Create(ctx context.Context, inspection domain.QualityControlInspection) (*domain.QualityControlInspection, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityControlInspection, error)
	FindAll(ctx context.Context, organizationID uuid.UUID, limit int) ([]domain.QualityControlInspection, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.QualityControlInspection, error)
	FindByLot(ctx context.Context, organizationID uuid.UUID, lotID uuid.UUID) ([]domain.QualityControlInspection, error)
	FindByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]domain.QualityControlInspection, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]domain.QualityControlInspection, error)
	FindByDateRange(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) ([]domain.QualityControlInspection, error)
	Update(ctx context.Context, inspection domain.QualityControlInspection) (*domain.QualityControlInspection, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Business logic methods
	CreateFromStockMove(ctx context.Context, stockMoveID, inspectorID uuid.UUID, checklistID *uuid.UUID, inspectionMethod string, sampleSize *int) (*domain.QualityControlInspection, error)
	UpdateStatus(ctx context.Context, inspectionID uuid.UUID, status, defectType, defectDescription string, defectQuantity *float64, qualityRating *int, complianceNotes, disposition *string) error
	CompleteInspection(ctx context.Context, inspectionID uuid.UUID, status string, results []domain.QualityControlInspectionItem) error
	GetStatistics(ctx context.Context, organizationID uuid.UUID, fromTime, toTime *time.Time, productID *uuid.UUID) (domain.QualityControlStatistics, error)
}

// QualityControlChecklistRepository interface
type QualityControlChecklistRepository interface {
	Create(ctx context.Context, checklist domain.QualityControlChecklist) (*domain.QualityControlChecklist, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityControlChecklist, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.QualityControlChecklist, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.QualityControlChecklist, error)
	FindByCategory(ctx context.Context, organizationID, categoryID uuid.UUID) ([]domain.QualityControlChecklist, error)
	FindByInspectionType(ctx context.Context, organizationID uuid.UUID, inspectionType string) ([]domain.QualityControlChecklist, error)
	FindActive(ctx context.Context, organizationID uuid.UUID) ([]domain.QualityControlChecklist, error)
	Update(ctx context.Context, checklist domain.QualityControlChecklist) (*domain.QualityControlChecklist, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// QualityChecklistItemRepository interface
type QualityChecklistItemRepository interface {
	Create(ctx context.Context, item domain.QualityChecklistItem) (*domain.QualityChecklistItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityChecklistItem, error)
	FindByChecklist(ctx context.Context, checklistID uuid.UUID) ([]domain.QualityChecklistItem, error)
	FindActiveByChecklist(ctx context.Context, checklistID uuid.UUID) ([]domain.QualityChecklistItem, error)
	Update(ctx context.Context, item domain.QualityChecklistItem) (*domain.QualityChecklistItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByChecklist(ctx context.Context, checklistID uuid.UUID) error
}

// QualityControlInspectionItemRepository interface
type QualityControlInspectionItemRepository interface {
	Create(ctx context.Context, item domain.QualityControlInspectionItem) (*domain.QualityControlInspectionItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityControlInspectionItem, error)
	FindByInspection(ctx context.Context, inspectionID uuid.UUID) ([]domain.QualityControlInspectionItem, error)
	Update(ctx context.Context, item domain.QualityControlInspectionItem) (*domain.QualityControlInspectionItem, error)
	UpdateResult(ctx context.Context, itemID uuid.UUID, result, notes string) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByInspection(ctx context.Context, inspectionID uuid.UUID) error
}

// QualityControlAlertRepository interface
type QualityControlAlertRepository interface {
	Create(ctx context.Context, alert domain.QualityControlAlert) (*domain.QualityControlAlert, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.QualityControlAlert, error)
	FindAll(ctx context.Context, organizationID uuid.UUID) ([]domain.QualityControlAlert, error)
	FindByStatus(ctx context.Context, organizationID uuid.UUID, status string) ([]domain.QualityControlAlert, error)
	FindBySeverity(ctx context.Context, organizationID uuid.UUID, severity string) ([]domain.QualityControlAlert, error)
	FindByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]domain.QualityControlAlert, error)
	FindOpen(ctx context.Context, organizationID uuid.UUID) ([]domain.QualityControlAlert, error)
	Update(ctx context.Context, alert domain.QualityControlAlert) (*domain.QualityControlAlert, error)
	UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedBy *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Business logic methods
	CreateFromInspection(ctx context.Context, inspectionID uuid.UUID, alertType, severity, title, message string) (*domain.QualityControlAlert, error)
}
