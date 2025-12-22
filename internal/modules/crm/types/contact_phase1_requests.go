package types

import (
	"github.com/google/uuid"
)

// ========== Merge Requests ==========

// ContactMergeRequest represents a request to merge contacts
type ContactMergeRequest struct {
	MasterContactID uuid.UUID            `json:"master_contact_id"`
	MergeContactIDs []uuid.UUID          `json:"merge_contact_ids"`
	Strategy        string               `json:"strategy"`                   // keep_master, keep_latest, custom
	FieldSelections map[string]uuid.UUID `json:"field_selections,omitempty"` // field -> source contact ID
	Notes           *string              `json:"notes,omitempty"`
}

// ContactMergeResponse represents the result of a merge operation
type ContactMergeResponse struct {
	MasterContact  *Contact  `json:"master_contact"`
	MergeHistoryID uuid.UUID `json:"merge_history_id"`
	MergedCount    int       `json:"merged_count"`
	CanUndo        bool      `json:"can_undo"`
}

// DuplicateResolutionRequest represents a request to resolve a duplicate
type DuplicateResolutionRequest struct {
	DuplicateID uuid.UUID `json:"duplicate_id"`
	Action      string    `json:"action"` // merge, ignore, false_positive
	Notes       *string   `json:"notes,omitempty"`
}

// ========== Duplicate Detection Responses ==========

// DetectDuplicatesResponse represents the result of duplicate detection
type DetectDuplicatesResponse struct {
	TotalFound   int                 `json:"total_found"`
	TotalCreated int                 `json:"total_created"`
	Threshold    int                 `json:"threshold"`
	Duplicates   []*ContactDuplicate `json:"duplicates"`
}

// CalculateSimilarityResponse represents the result of similarity calculation
type CalculateSimilarityResponse struct {
	Contact1ID      uuid.UUID `json:"contact1_id"`
	Contact2ID      uuid.UUID `json:"contact2_id"`
	SimilarityScore float64   `json:"similarity_score"`
	IsDuplicate     bool      `json:"is_duplicate"`
}

// ListDuplicatesResponse represents a list of duplicates with pagination
type ListDuplicatesResponse struct {
	Duplicates []*ContactDuplicate `json:"duplicates"`
	Total      int                 `json:"total"`
	Limit      *int                `json:"limit,omitempty"`
	Offset     *int                `json:"offset,omitempty"`
}
