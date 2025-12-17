package types

import (
	"time"

	"github.com/google/uuid"
)

// CycleCountPlan represents a cycle counting plan
type CycleCountPlan struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	Frequency   string     `json:"frequency" db:"frequency"` // daily, weekly, monthly, quarterly, custom
	ABCClass    string     `json:"abc_class" db:"abc_class"`   // A, B, C, all
	StartDate   *time.Time `json:"start_date,omitempty" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	Status      string     `json:"status" db:"status"` // draft, active, completed, cancelled
	Priority    int        `json:"priority" db:"priority"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Metadata    *string    `json:"metadata,omitempty" db:"metadata"`
}

// CycleCountSession represents a cycle counting session
type CycleCountSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	PlanID       *uuid.UUID `json:"plan_id,omitempty" db:"plan_id"`
	Name         string     `json:"name" db:"name"`
	LocationID   *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	UserID       *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	StartTime    time.Time  `json:"start_time" db:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty" db:"end_time"`
	Status       string     `json:"status" db:"status"` // in_progress, completed, verified, cancelled
	CountMethod  string     `json:"count_method" db:"count_method"` // manual, barcode, mobile
	DeviceID     *string    `json:"device_id,omitempty" db:"device_id"`
	Notes        *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// CycleCountLine represents an individual count line in a session
type CycleCountLine struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	SessionID       uuid.UUID  `json:"session_id" db:"session_id"`
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID       uuid.UUID  `json:"product_id" db:"product_id"`
	ProductVariantID *uuid.UUID `json:"product_variant_id,omitempty" db:"product_variant_id"`
	LocationID      uuid.UUID  `json:"location_id" db:"location_id"`
	LotID           *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	PackageID       *uuid.UUID `json:"package_id,omitempty" db:"package_id"`
	CountedQuantity float64    `json:"counted_quantity" db:"counted_quantity"`
	SystemQuantity  float64    `json:"system_quantity" db:"system_quantity"`
	Variance        float64    `json:"variance" db:"variance"`
	VariancePercentage float64 `json:"variance_percentage" db:"variance_percentage"`
	UOMID          *uuid.UUID `json:"uom_id,omitempty" db:"uom_id"`
	CountTime      time.Time  `json:"count_time" db:"count_time"`
	CountedBy      *uuid.UUID `json:"counted_by,omitempty" db:"counted_by"`
	VerifiedBy     *uuid.UUID `json:"verified_by,omitempty" db:"verified_by"`
	VerificationTime *time.Time `json:"verification_time,omitempty" db:"verification_time"`
	Status         string     `json:"status" db:"status"` // counted, verified, adjusted, resolved
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	ResolutionNotes *string   `json:"resolution_notes,omitempty" db:"resolution_notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// CycleCountAdjustment represents an inventory adjustment from cycle counting
type CycleCountAdjustment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	CountLineID     uuid.UUID  `json:"count_line_id" db:"count_line_id"`
	ProductID       uuid.UUID  `json:"product_id" db:"product_id"`
	LocationID      uuid.UUID  `json:"location_id" db:"location_id"`
	LotID           *uuid.UUID `json:"lot_id,omitempty" db:"lot_id"`
	PackageID       *uuid.UUID `json:"package_id,omitempty" db:"package_id"`
	OldQuantity     float64    `json:"old_quantity" db:"old_quantity"`
	NewQuantity     float64    `json:"new_quantity" db:"new_quantity"`
	AdjustmentQuantity float64 `json:"adjustment_quantity" db:"adjustment_quantity"`
	AdjustmentType  string     `json:"adjustment_type" db:"adjustment_type"` // variance, damage, theft, misplacement
	Reason          *string    `json:"reason,omitempty" db:"reason"`
	AdjustmentTime  time.Time  `json:"adjustment_time" db:"adjustment_time"`
	AdjustedBy      *uuid.UUID `json:"adjusted_by,omitempty" db:"adjusted_by"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
	ApprovalTime    *time.Time `json:"approval_time,omitempty" db:"approval_time"`
	Status          string     `json:"status" db:"status"` // pending, approved, rejected
	Notes           *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// CycleCountAccuracy represents accuracy history for cycle counting
type CycleCountAccuracy struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID       *uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	LocationID      *uuid.UUID `json:"location_id,omitempty" db:"location_id"`
	CountDate       time.Time  `json:"count_date" db:"count_date"`
	SystemQuantity  float64    `json:"system_quantity" db:"system_quantity"`
	CountedQuantity float64    `json:"counted_quantity" db:"counted_quantity"`
	Variance        float64    `json:"variance" db:"variance"`
	VariancePercentage float64 `json:"variance_percentage" db:"variance_percentage"`
	AccuracyScore   float64    `json:"accuracy_score" db:"accuracy_score"`
	CountMethod     string     `json:"count_method" db:"count_method"`
	CountedBy       *uuid.UUID `json:"counted_by,omitempty" db:"counted_by"`
	VerifiedBy      *uuid.UUID `json:"verified_by,omitempty" db:"verified_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// CycleCountMetrics represents accuracy metrics
type CycleCountMetrics struct {
	TotalCounts            int     `json:"total_counts"`
	AccurateCounts         int     `json:"accurate_counts"`
	VarianceCounts         int     `json:"variance_counts"`
	AccuracyPercentage     float64 `json:"accuracy_percentage"`
	AverageVariancePercentage float64 `json:"average_variance_percentage"`
	TotalVarianceQuantity  float64 `json:"total_variance_quantity"`
}

// ProductNeedingCycleCount represents a product that needs cycle counting
type ProductNeedingCycleCount struct {
	ProductID              uuid.UUID  `json:"product_id"`
	ProductName            string     `json:"product_name"`
	DefaultCode            *string    `json:"default_code,omitempty"`
	CategoryID             *uuid.UUID `json:"category_id,omitempty"`
	LastCountDate          *time.Time `json:"last_count_date,omitempty"`
	DaysSinceCount         int        `json:"days_since_count"`
	LastVariancePercentage float64    `json:"last_variance_percentage"`
	AverageVariancePercentage float64 `json:"average_variance_percentage"`
	CountPriority          int        `json:"count_priority"`
}

// CreateCycleCountPlanRequest represents a request to create a cycle count plan
type CreateCycleCountPlanRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Frequency   string    `json:"frequency"`
	ABCClass    string    `json:"abc_class"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Priority    int       `json:"priority"`
	CreatedBy   uuid.UUID `json:"created_by"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	Metadata    *string   `json:"metadata,omitempty"`
}

// CreateCycleCountSessionRequest represents a request to create a cycle count session
type CreateCycleCountSessionRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	PlanID       *uuid.UUID `json:"plan_id,omitempty"`
	Name         string    `json:"name"`
	LocationID   *uuid.UUID `json:"location_id,omitempty"`
	UserID       uuid.UUID `json:"user_id"`
	CountMethod  string    `json:"count_method"`
	DeviceID     *string   `json:"device_id,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
}

// AddCycleCountLineRequest represents a request to add a count line
type AddCycleCountLineRequest struct {
	SessionID      uuid.UUID  `json:"session_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ProductID      uuid.UUID `json:"product_id"`
	LocationID     uuid.UUID `json:"location_id"`
	CountedQuantity float64  `json:"counted_quantity"`
	CountedBy      uuid.UUID `json:"counted_by"`
	LotID          *uuid.UUID `json:"lot_id,omitempty"`
	PackageID      *uuid.UUID `json:"package_id,omitempty"`
}

// VerifyCycleCountLineRequest represents a request to verify a count line
type VerifyCycleCountLineRequest struct {
	LineID      uuid.UUID `json:"line_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	VerifiedBy  uuid.UUID `json:"verified_by"`
	Status      string    `json:"status"`
}

// CreateAdjustmentRequest represents a request to create an adjustment
type CreateAdjustmentRequest struct {
	LineID          uuid.UUID `json:"line_id"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	AdjustmentType  string    `json:"adjustment_type"`
	Reason          *string   `json:"reason,omitempty"`
	AdjustedBy      uuid.UUID `json:"adjusted_by"`
}

// ApproveAdjustmentRequest represents a request to approve an adjustment
type ApproveAdjustmentRequest struct {
	AdjustmentID  uuid.UUID `json:"adjustment_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ApprovedBy    uuid.UUID `json:"approved_by"`
}

// GetAccuracyMetricsRequest represents a request for accuracy metrics
type GetAccuracyMetricsRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	DateFrom       *time.Time `json:"date_from,omitempty"`
	DateTo         *time.Time `json:"date_to,omitempty"`
}

// GetProductsNeedingCountRequest represents a request for products needing counting
type GetProductsNeedingCountRequest struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	DaysSinceLastCount *int `json:"days_since_last_count,omitempty"`
	MinVariancePercentage *float64 `json:"min_variance_percentage,omitempty"`
}
