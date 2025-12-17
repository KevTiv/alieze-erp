package domain

import (
	"time"

	"github.com/google/uuid"
)

// QualityControlInspection represents a quality control inspection record
type QualityControlInspection struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID `json:"company_id,omitempty" db:"company_id"`

	// Reference information
	Reference       string `json:"reference" db:"reference"` // Inspection reference number
	InspectionType  string `json:"inspection_type" db:"inspection_type"` // "incoming", "outgoing", "internal", "return"
	SourceDocumentID *uuid.UUID `json:"source_document_id,omitempty" db:"source_document_id"` // Related document (purchase order, stock move, etc.)
	SourceType      *string `json:"source_type,omitempty" db:"source_type"` // Type of source document

	// Product information
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	ProductName     string `json:"product_name" db:"product_name"`
	LotID           *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	SerialNumber    *string `json:"serial_number,omitempty" db:"serial_number"`
	Quantity        float64 `json:"quantity" db:"quantity"`
	UOMID           *uuid.UUID `json:"uom_id,omitempty" db:"uom_id"`

	// Location information
	LocationID      uuid.UUID `json:"location_id" db:"location_id"`
	LocationName    string `json:"location_name" db:"location_name"`

	// Inspection details
	InspectionDate  time.Time `json:"inspection_date" db:"inspection_date"`
	InspectorID     *uuid.UUID `json:"inspector_id,omitempty" db:"inspector_id"`
	InspectionMethod string `json:"inspection_method" db:"inspection_method"` // "visual", "measurement", "testing", "sampling"
	SampleSize      *int `json:"sample_size,omitempty" db:"sample_size"`

	// Quality status
	Status          string `json:"status" db:"status"` // "pending", "passed", "failed", "quarantined", "rejected"
	DefectType      *string `json:"defect_type,omitempty" db:"defect_type"` // Type of defect if failed
	DefectDescription *string `json:"defect_description,omitempty" db:"defect_description"`
	DefectQuantity  *float64 `json:"defect_quantity,omitempty" db:"defect_quantity"`

	// Quality metrics
	QualityRating   *int `json:"quality_rating,omitempty" db:"quality_rating"` // 1-100 scale
	ComplianceNotes  *string `json:"compliance_notes,omitempty" db:"compliance_notes"`

	// Disposition
	Disposition     *string `json:"disposition,omitempty" db:"disposition"` // "accept", "reject", "rework", "scrap", "return"
	DispositionDate *time.Time `json:"disposition_date,omitempty" db:"disposition_date"`
	DispositionBy   *uuid.UUID `json:"disposition_by,omitempty" db:"disposition_by"`

	// Standard fields
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Metadata
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// QualityControlChecklist represents a quality control checklist template
type QualityControlChecklist struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyID       *uuid.UUID `json:"company_id,omitempty" db:"company_id"`
	Name            string `json:"name" db:"name"`
	Description     *string `json:"description,omitempty" db:"description"`
	ProductID       *uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	ProductCategoryID *uuid.UUID `json:"product_category_id,omitempty" db:"product_category_id"`
	InspectionType  string `json:"inspection_type" db:"inspection_type"` // "incoming", "outgoing", "internal"
	ChecklistItems  []QualityChecklistItem `json:"checklist_items" db:"checklist_items"`
	Active          bool `json:"active" db:"active"`
	Priority        int `json:"priority" db:"priority"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// QualityChecklistItem represents an individual item in a quality control checklist
type QualityChecklistItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ChecklistID uuid.UUID `json:"checklist_id" db:"checklist_id"`
	Description string `json:"description" db:"description"`
	Criteria    *string `json:"criteria,omitempty" db:"criteria"`
	Sequence    int `json:"sequence" db:"sequence"`
	Active      bool `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// QualityControlInspectionItem represents the result of a specific checklist item inspection
type QualityControlInspectionItem struct {
	ID              uuid.UUID `json:"id" db:"id"`
	InspectionID    uuid.UUID `json:"inspection_id" db:"inspection_id"`
	ChecklistItemID uuid.UUID `json:"checklist_item_id" db:"checklist_item_id"`
	Description     string `json:"description" db:"description"`
	Result          string `json:"result" db:"result"` // "pass", "fail", "na"
	Notes           *string `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// QualityControlStatistics represents quality control metrics and statistics
type QualityControlStatistics struct {
	TotalInspections      int     `json:"total_inspections"`
	PassedInspections     int     `json:"passed_inspections"`
	FailedInspections     int     `json:"failed_inspections"`
	QuarantinedItems      int     `json:"quarantined_items"`
	RejectedItems         int     `json:"rejected_items"`
	PassRate              float64 `json:"pass_rate"`
	FailRate              float64 `json:"fail_rate"`
	AverageQualityRating  float64 `json:"average_quality_rating"`
	DefectRate            float64 `json:"defect_rate"`
	InspectionTime        string  `json:"inspection_time"` // Average time per inspection
	TopDefectTypes        []DefectTypeSummary `json:"top_defect_types"`
	QualityTrend          string  `json:"quality_trend"` // "improving", "stable", "declining"
}

// DefectTypeSummary represents a summary of defect types
type DefectTypeSummary struct {
	DefectType  string  `json:"defect_type"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// QualityControlAlert represents a quality control alert or notification
type QualityControlAlert struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	AlertType       string `json:"alert_type" db:"alert_type"` // "defect", "quarantine", "rejection", "threshold"
	Severity        string `json:"severity" db:"severity"` // "low", "medium", "high", "critical"
	Title           string `json:"title" db:"title"`
	Message         string `json:"message" db:"message"`
	RelatedInspectionID *uuid.UUID `json:"related_inspection_id,omitempty" db:"related_inspection_id"`
	ProductID       *uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	Status          string `json:"status" db:"status"` // "open", "acknowledged", "resolved", "closed"
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy      *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
}
