package types

import (
	"time"

	"github.com/google/uuid"
)

// DuplicateFilter represents filtering criteria for duplicates
type DuplicateFilter struct {
	OrganizationID uuid.UUID
	Status         *string // pending, merged, ignored, false_positive
	MinSimilarity  *float64
	ContactID      *uuid.UUID
	Limit          int
	Offset         int
}

// SegmentFilter represents filtering criteria for segments
type SegmentFilter struct {
	OrganizationID uuid.UUID
	SegmentType    *string // static, dynamic
	Name           *string
	Limit          int
	Offset         int
}

// ImportJobFilter represents filtering criteria for import jobs
type ImportJobFilter struct {
	OrganizationID uuid.UUID
	Status         *string
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}

// ExportJobFilter represents filtering criteria for export jobs
type ExportJobFilter struct {
	OrganizationID uuid.UUID
	Status         *string
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}

// CustomRelationshipTypeFilter represents filtering criteria for relationship types
type CustomRelationshipTypeFilter struct {
	OrganizationID uuid.UUID
	IsActive       *bool
	IsSystem       *bool
	Limit          int
	Offset         int
}

// ValidationRuleFilter represents filtering criteria for validation rules
type ValidationRuleFilter struct {
	OrganizationID uuid.UUID
	Field          *string
	IsActive       *bool
	Severity       *string
	Limit          int
	Offset         int
}
