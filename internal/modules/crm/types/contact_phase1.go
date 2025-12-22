package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSONBMap represents a JSONB field as a map
type JSONBMap map[string]interface{}

// Value implements the driver.Valuer interface for JSONBMap
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONBMap
func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONBMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ========== Relationship Types ==========

// ContactRelationshipType represents a custom relationship type definition
type CustomRelationshipType struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	Name            string    `json:"name" db:"name"`
	Code            string    `json:"code" db:"code"`
	Description     *string   `json:"description,omitempty" db:"description"`
	IsBidirectional bool      `json:"is_bidirectional" db:"is_bidirectional"`
	ReverseName     *string   `json:"reverse_name,omitempty" db:"reverse_name"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	IsSystem        bool      `json:"is_system" db:"is_system"`
	Color           *int      `json:"color,omitempty" db:"color"`
	Metadata        JSONBMap  `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ContactRelationshipEnhanced extends ContactRelationship with strength tracking
type ContactRelationshipEnhanced struct {
	ContactRelationship
	StrengthScore       int        `json:"strength_score" db:"strength_score"`
	LastInteractionDate *time.Time `json:"last_interaction_date,omitempty" db:"last_interaction_date"`
	InteractionCount    int        `json:"interaction_count" db:"interaction_count"`
	Metadata            JSONBMap   `json:"metadata,omitempty" db:"metadata"`
}

// ========== Segmentation ==========

// ContactSegment represents a contact segment (static or dynamic)
type ContactSegment struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name             string     `json:"name" db:"name"`
	Description      *string    `json:"description,omitempty" db:"description"`
	SegmentType      string     `json:"segment_type" db:"segment_type"` // static, dynamic
	Criteria         JSONBMap   `json:"criteria,omitempty" db:"criteria"`
	Color            *int       `json:"color,omitempty" db:"color"`
	Icon             *string    `json:"icon,omitempty" db:"icon"`
	MemberCount      int        `json:"member_count" db:"member_count"`
	LastCalculatedAt *time.Time `json:"last_calculated_at,omitempty" db:"last_calculated_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy        *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
}

// ContactSegmentMember represents segment membership
type ContactSegmentMember struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	SegmentID      uuid.UUID  `json:"segment_id" db:"segment_id"`
	ContactID      uuid.UUID  `json:"contact_id" db:"contact_id"`
	AddedAt        time.Time  `json:"added_at" db:"added_at"`
	AddedBy        *uuid.UUID `json:"added_by,omitempty" db:"added_by"`
}

// ========== Duplicate Detection ==========

// ContactDuplicate represents a potential duplicate pair
type ContactDuplicate struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id" db:"organization_id"`
	ContactID1      uuid.UUID  `json:"contact_id_1" db:"contact_id_1"`
	ContactID2      uuid.UUID  `json:"contact_id_2" db:"contact_id_2"`
	SimilarityScore float64    `json:"similarity_score" db:"similarity_score"`
	MatchingFields  []string   `json:"matching_fields" db:"matching_fields"`
	Status          string     `json:"status" db:"status"` // pending, merged, ignored, false_positive
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	Notes           *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// ContactMergeHistory represents merge audit trail
type ContactMergeHistory struct {
	ID               uuid.UUID   `json:"id" db:"id"`
	OrganizationID   uuid.UUID   `json:"organization_id" db:"organization_id"`
	MasterContactID  uuid.UUID   `json:"master_contact_id" db:"master_contact_id"`
	MergedContactIDs []uuid.UUID `json:"merged_contact_ids" db:"merged_contact_ids"`
	MergeStrategy    string      `json:"merge_strategy" db:"merge_strategy"`
	FieldSelections  JSONBMap    `json:"field_selections" db:"field_selections"`
	MergedBy         *uuid.UUID  `json:"merged_by,omitempty" db:"merged_by"`
	MergedAt         time.Time   `json:"merged_at" db:"merged_at"`
	CanUndo          bool        `json:"can_undo" db:"can_undo"`
	Notes            *string     `json:"notes,omitempty" db:"notes"`
}

// ========== Import/Export ==========

// ContactImportJob represents an import job
type ContactImportJob struct {
	ID                 uuid.UUID   `json:"id" db:"id"`
	OrganizationID     uuid.UUID   `json:"organization_id" db:"organization_id"`
	JobID              uuid.UUID   `json:"job_id" db:"job_id"`
	Filename           string      `json:"filename" db:"filename"`
	FileSize           *int64      `json:"file_size,omitempty" db:"file_size"`
	FileType           string      `json:"file_type" db:"file_type"`
	FieldMapping       JSONBMap    `json:"field_mapping" db:"field_mapping"`
	Options            JSONBMap    `json:"options" db:"options"`
	TotalRows          int         `json:"total_rows" db:"total_rows"`
	ProcessedRows      int         `json:"processed_rows" db:"processed_rows"`
	SuccessfulRows     int         `json:"successful_rows" db:"successful_rows"`
	FailedRows         int         `json:"failed_rows" db:"failed_rows"`
	DuplicateRows      int         `json:"duplicate_rows" db:"duplicate_rows"`
	Status             string      `json:"status" db:"status"`
	ErrorMessage       *string     `json:"error_message,omitempty" db:"error_message"`
	ErrorDetails       JSONBMap    `json:"error_details,omitempty" db:"error_details"`
	ImportedContactIDs []uuid.UUID `json:"imported_contact_ids" db:"imported_contact_ids"`
	FailedRowsData     JSONBMap    `json:"failed_rows_data,omitempty" db:"failed_rows_data"`
	CreatedBy          *uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
	StartedAt          *time.Time  `json:"started_at,omitempty" db:"started_at"`
	CompletedAt        *time.Time  `json:"completed_at,omitempty" db:"completed_at"`
}

// ContactExportJob represents an export job
type ContactExportJob struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	JobID          uuid.UUID  `json:"job_id" db:"job_id"`
	FilterCriteria JSONBMap   `json:"filter_criteria" db:"filter_criteria"`
	SelectedFields JSONBMap   `json:"selected_fields,omitempty" db:"selected_fields"`
	Format         string     `json:"format" db:"format"`
	TotalContacts  int        `json:"total_contacts" db:"total_contacts"`
	Status         string     `json:"status" db:"status"`
	ErrorMessage   *string    `json:"error_message,omitempty" db:"error_message"`
	FileKey        *string    `json:"file_key,omitempty" db:"file_key"`
	FileURL        *string    `json:"file_url,omitempty" db:"file_url"`
	FileExpiresAt  *time.Time `json:"file_expires_at,omitempty" db:"file_expires_at"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// ========== Validation ==========

// ContactValidationRule represents a validation rule
type ContactValidationRule struct {
	ID               uuid.UUID `json:"id" db:"id"`
	OrganizationID   uuid.UUID `json:"organization_id" db:"organization_id"`
	Name             string    `json:"name" db:"name"`
	Field            string    `json:"field" db:"field"`
	RuleType         string    `json:"rule_type" db:"rule_type"`
	ValidationConfig JSONBMap  `json:"validation_config" db:"validation_config"`
	ErrorMessage     *string   `json:"error_message,omitempty" db:"error_message"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	Severity         string    `json:"severity" db:"severity"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ContactValidationResult represents validation results
type ContactValidationResult struct {
	IsValid      bool                `json:"is_valid"`
	Errors       []ValidationError   `json:"errors,omitempty"`
	Warnings     []ValidationWarning `json:"warnings,omitempty"`
	Suggestions  []DataSuggestion    `json:"suggestions,omitempty"`
	QualityScore int                 `json:"quality_score"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	RuleID  string `json:"rule_id,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DataSuggestion represents a data completion suggestion
type DataSuggestion struct {
	Field          string  `json:"field"`
	CurrentValue   string  `json:"current_value,omitempty"`
	SuggestedValue string  `json:"suggested_value"`
	Confidence     float64 `json:"confidence"`
	Source         string  `json:"source"` // AI, duplicate, pattern
}

// ========== Relationship Network ==========

// ContactRelationshipNetwork represents a contact's relationship network
type ContactRelationshipNetwork struct {
	CenterContactID uuid.UUID            `json:"center_contact_id"`
	Nodes           []ContactNetworkNode `json:"nodes"`
	Edges           []ContactNetworkEdge `json:"edges"`
	Statistics      NetworkStatistics    `json:"statistics"`
}

// NetworkNode represents a node in the relationship network
type ContactNetworkNode struct {
	ContactID       uuid.UUID `json:"contact_id"`
	Name            string    `json:"name"`
	Email           *string   `json:"email,omitempty"`
	Company         *string   `json:"company,omitempty"`
	NodeType        string    `json:"node_type"` // center, direct, indirect
	EngagementScore int       `json:"engagement_score"`
}

// ContactNetworkEdge represents an edge in the relationship network
type ContactNetworkEdge struct {
	FromContactID    uuid.UUID `json:"from_contact_id"`
	ToContactID      uuid.UUID `json:"to_contact_id"`
	RelationshipType string    `json:"relationship_type"`
	StrengthScore    int       `json:"strength_score"`
	IsBidirectional  bool      `json:"is_bidirectional"`
}

// NetworkStatistics represents network statistics
type NetworkStatistics struct {
	TotalConnections    int                 `json:"total_connections"`
	DirectConnections   int                 `json:"direct_connections"`
	IndirectConnections int                 `json:"indirect_connections"`
	AverageStrength     float64             `json:"average_strength"`
	StrongestConnection *ContactNetworkNode `json:"strongest_connection,omitempty"`
}
