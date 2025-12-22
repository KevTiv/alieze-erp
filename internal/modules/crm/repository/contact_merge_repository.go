package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
)

// ContactMergeRepository handles duplicate detection and contact merging operations
type ContactMergeRepository interface {
	// Duplicate Detection
	FindPotentialDuplicates(ctx context.Context, orgID uuid.UUID, threshold int, limit int) ([]*types.ContactDuplicate, error)
	CreateDuplicate(ctx context.Context, duplicate *types.ContactDuplicate) error
	GetDuplicate(ctx context.Context, id uuid.UUID) (*types.ContactDuplicate, error)
	UpdateDuplicateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID) error
	ListDuplicates(ctx context.Context, filter types.DuplicateFilter) ([]*types.ContactDuplicate, error)
	CountDuplicates(ctx context.Context, filter types.DuplicateFilter) (int, error)

	// Merging
	MergeContacts(ctx context.Context, masterID, duplicateID uuid.UUID, fieldSelections map[string]string, mergedBy uuid.UUID, orgID uuid.UUID) error
	GetMergeHistory(ctx context.Context, contactID uuid.UUID) ([]*types.ContactMergeHistory, error)

	// Similarity Calculation
	CalculateSimilarity(contact1, contact2 *types.Contact) int
}

type contactMergeRepository struct {
	db *sql.DB
}

func NewContactMergeRepository(db *sql.DB) ContactMergeRepository {
	return &contactMergeRepository{db: db}
}

// FindPotentialDuplicates identifies potential duplicate contacts based on similarity scoring
func (r *contactMergeRepository) FindPotentialDuplicates(ctx context.Context, orgID uuid.UUID, threshold int, limit int) ([]*types.ContactDuplicate, error) {
	// Get all active contacts for the organization
	query := `
		SELECT id, email, phone, name
		FROM contacts
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*types.Contact
	for rows.Next() {
		contact := &types.Contact{}
		err := rows.Scan(
			&contact.ID,
			&contact.Email,
			&contact.Phone,
			&contact.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	// Compare all contact pairs to find duplicates
	var duplicates []*types.ContactDuplicate
	seen := make(map[string]bool) // Track unique pairs to avoid duplicates

	for i := 0; i < len(contacts); i++ {
		for j := i + 1; j < len(contacts); j++ {
			similarity := r.CalculateSimilarity(contacts[i], contacts[j])

			if similarity >= threshold {
				// Create unique key for pair (smaller UUID first)
				contact1ID := contacts[i].ID.String()
				contact2ID := contacts[j].ID.String()
				if contact1ID > contact2ID {
					contact1ID, contact2ID = contact2ID, contact1ID
				}
				pairKey := contact1ID + "-" + contact2ID

				if !seen[pairKey] {
					duplicate := &types.ContactDuplicate{
						ID:              uuid.New(),
						OrganizationID:  orgID,
						ContactID1:      contacts[i].ID,
						ContactID2:      contacts[j].ID,
						SimilarityScore: float64(similarity),
						Status:          "pending",
						MatchingFields:  r.getMatchingFields(contacts[i], contacts[j]),
					}
					duplicates = append(duplicates, duplicate)
					seen[pairKey] = true

					if limit > 0 && len(duplicates) >= limit {
						return duplicates, nil
					}
				}
			}
		}
	}

	return duplicates, nil
}

// CalculateSimilarity computes a similarity score (0-100) between two contacts
func (r *contactMergeRepository) CalculateSimilarity(contact1, contact2 *types.Contact) int {
	score := 0

	// Email exact match: +40 points
	if contact1.Email != nil && contact2.Email != nil && strings.EqualFold(*contact1.Email, *contact2.Email) {
		score += 40
	}

	// Phone exact match: +30 points
	if contact1.Phone != nil && contact2.Phone != nil {
		phone1 := normalizePhone(*contact1.Phone)
		phone2 := normalizePhone(*contact2.Phone)
		if phone1 == phone2 {
			score += 30
		}
	}

	// Name similarity: 0-30 points (using Levenshtein distance)
	name1 := strings.ToLower(strings.TrimSpace(contact1.Name))
	name2 := strings.ToLower(strings.TrimSpace(contact2.Name))

	if name1 != "" && name2 != "" {
		nameScore := calculateStringSimilarity(name1, name2)
		score += int(float64(nameScore) * 0.30) // Scale to 30 points max
	}

	return score
}

// getMatchingFields returns a list of fields that match between two contacts
func (r *contactMergeRepository) getMatchingFields(contact1, contact2 *types.Contact) []string {
	var matching []string

	if contact1.Email != nil && contact2.Email != nil && strings.EqualFold(*contact1.Email, *contact2.Email) {
		matching = append(matching, "email")
	}

	if contact1.Phone != nil && contact2.Phone != nil && normalizePhone(*contact1.Phone) == normalizePhone(*contact2.Phone) {
		matching = append(matching, "phone")
	}

	if strings.EqualFold(strings.TrimSpace(contact1.Name), strings.TrimSpace(contact2.Name)) {
		matching = append(matching, "name")
	}

	return matching
}

// CreateDuplicate stores a detected duplicate pair
func (r *contactMergeRepository) CreateDuplicate(ctx context.Context, duplicate *types.ContactDuplicate) error {
	query := `
		INSERT INTO contact_duplicates (
			id, organization_id, contact_id_1, contact_id_2,
			similarity_score, matching_fields, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		duplicate.ID,
		duplicate.OrganizationID,
		duplicate.ContactID1,
		duplicate.ContactID2,
		duplicate.SimilarityScore,
		pq.Array(duplicate.MatchingFields),
		duplicate.Status,
		now,
	).Scan(&duplicate.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create duplicate: %w", err)
	}

	return nil
}

// GetDuplicate retrieves a duplicate record by ID
func (r *contactMergeRepository) GetDuplicate(ctx context.Context, id uuid.UUID) (*types.ContactDuplicate, error) {
	query := `
		SELECT id, organization_id, contact_id_1, contact_id_2,
			   similarity_score, matching_fields, status, reviewed_by,
			   reviewed_at, notes, created_at
		FROM contact_duplicates
		WHERE id = $1
	`

	duplicate := &types.ContactDuplicate{}
	var matchingFields pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&duplicate.ID,
		&duplicate.OrganizationID,
		&duplicate.ContactID1,
		&duplicate.ContactID2,
		&duplicate.SimilarityScore,
		&matchingFields,
		&duplicate.Status,
		&duplicate.ReviewedBy,
		&duplicate.ReviewedAt,
		&duplicate.Notes,
		&duplicate.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("duplicate not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get duplicate: %w", err)
	}

	duplicate.MatchingFields = []string(matchingFields)
	return duplicate, nil
}

// UpdateDuplicateStatus updates the status of a duplicate record
func (r *contactMergeRepository) UpdateDuplicateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE contact_duplicates
		SET status = $1, reviewed_by = $2,
		    reviewed_at = $3
		WHERE id = $4
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, reviewedBy, now, id)
	if err != nil {
		return fmt.Errorf("failed to update duplicate status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("duplicate not found")
	}

	return nil
}

// ListDuplicates retrieves duplicates with filtering
func (r *contactMergeRepository) ListDuplicates(ctx context.Context, filter types.DuplicateFilter) ([]*types.ContactDuplicate, error) {
	query := `
		SELECT id, organization_id, contact_id_1, contact_id_2,
			   similarity_score, matching_fields, status, reviewed_by,
			   reviewed_at, notes, created_at
		FROM contact_duplicates
		WHERE organization_id = $1
	`

	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.MinSimilarity != nil {
		query += fmt.Sprintf(" AND similarity_score >= $%d", argPos)
		args = append(args, *filter.MinSimilarity)
		argPos++
	}

	query += " ORDER BY similarity_score DESC, created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list duplicates: %w", err)
	}
	defer rows.Close()

	var duplicates []*types.ContactDuplicate
	for rows.Next() {
		duplicate := &types.ContactDuplicate{}
		var matchingFields pq.StringArray

		err := rows.Scan(
			&duplicate.ID,
			&duplicate.OrganizationID,
			&duplicate.ContactID1,
			&duplicate.ContactID2,
			&duplicate.SimilarityScore,
			&matchingFields,
			&duplicate.Status,
			&duplicate.ReviewedBy,
			&duplicate.ReviewedAt,
			&duplicate.Notes,
			&duplicate.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate: %w", err)
		}

		duplicate.MatchingFields = []string(matchingFields)
		duplicates = append(duplicates, duplicate)
	}

	return duplicates, nil
}

// CountDuplicates counts duplicates matching the filter
func (r *contactMergeRepository) CountDuplicates(ctx context.Context, filter types.DuplicateFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contact_duplicates WHERE organization_id = $1`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.MinSimilarity != nil {
		query += fmt.Sprintf(" AND similarity_score >= $%d", argPos)
		args = append(args, *filter.MinSimilarity)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count duplicates: %w", err)
	}

	return count, nil
}

// MergeContacts merges two contacts, keeping the master and archiving the duplicate
func (r *contactMergeRepository) MergeContacts(ctx context.Context, masterID, duplicateID uuid.UUID, fieldSelections map[string]string, mergedBy uuid.UUID, orgID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get both contacts
	masterContact, err := r.getContactInTx(ctx, tx, masterID)
	if err != nil {
		return fmt.Errorf("failed to get master contact: %w", err)
	}

	duplicateContact, err := r.getContactInTx(ctx, tx, duplicateID)
	if err != nil {
		return fmt.Errorf("failed to get duplicate contact: %w", err)
	}

	// Build merged contact based on field selections
	mergedData := r.buildMergedContact(masterContact, duplicateContact, fieldSelections)

	// Update master contact with merged data
	err = r.updateContactInTx(ctx, tx, masterID, mergedData)
	if err != nil {
		return fmt.Errorf("failed to update master contact: %w", err)
	}

	// Soft delete duplicate contact
	_, err = tx.ExecContext(ctx, `UPDATE contacts SET deleted_at = $1 WHERE id = $2`, time.Now(), duplicateID)
	if err != nil {
		return fmt.Errorf("failed to delete duplicate contact: %w", err)
	}

	// Create merge history record
	historyID := uuid.New()
	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO contact_merge_history (
			id, organization_id, master_contact_id, merged_contact_ids,
			merge_strategy, merged_by, merged_at, field_selections, can_undo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		historyID, orgID, masterID, []uuid.UUID{duplicateID}, "manual",
		mergedBy, now, convertStringMapToJSONBMap(fieldSelections), true,
	)
	if err != nil {
		return fmt.Errorf("failed to create merge history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetMergeHistory retrieves merge history for a contact
func (r *contactMergeRepository) GetMergeHistory(ctx context.Context, contactID uuid.UUID) ([]*types.ContactMergeHistory, error) {
	query := `
		SELECT id, organization_id, master_contact_id, merged_contact_ids,
			   merge_strategy, merged_by, merged_at, field_selections, can_undo, notes
		FROM contact_merge_history
		WHERE master_contact_id = $1 OR $1 = ANY(merged_contact_ids)
		ORDER BY merged_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to query merge history: %w", err)
	}
	defer rows.Close()

	var history []*types.ContactMergeHistory
	for rows.Next() {
		h := &types.ContactMergeHistory{}
		err := rows.Scan(
			&h.ID,
			&h.OrganizationID,
			&h.MasterContactID,
			pq.Array(&h.MergedContactIDs),
			&h.MergeStrategy,
			&h.MergedBy,
			&h.MergedAt,
			&h.FieldSelections,
			&h.CanUndo,
			&h.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merge history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// Helper functions

func (r *contactMergeRepository) getContactInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT email, phone, name, street, city, state_id, country_id
		FROM contacts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var email, phone, name, street, city *string
	var stateID, countryID *uuid.UUID

	err := tx.QueryRowContext(ctx, query, id).Scan(
		&email, &phone, &name, &street, &city, &stateID, &countryID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contact not found")
	}
	if err != nil {
		return nil, err
	}

	contact := make(map[string]interface{})
	contact["email"] = email
	contact["phone"] = phone
	contact["name"] = name
	contact["street"] = street
	contact["city"] = city
	contact["state_id"] = stateID
	contact["country_id"] = countryID

	return contact, nil
}

func (r *contactMergeRepository) updateContactInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, data map[string]interface{}) error {
	query := `
		UPDATE contacts SET
			email = $1, phone = $2, name = $3, street = $4,
			city = $5, state_id = $6, country_id = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := tx.ExecContext(ctx, query,
		data["email"], data["phone"], data["name"], data["street"],
		data["city"], data["state_id"], data["country_id"], time.Now(), id,
	)

	return err
}

func (r *contactMergeRepository) buildMergedContact(master, duplicate map[string]interface{}, selections map[string]string) map[string]interface{} {
	merged := make(map[string]interface{})

	// Default strategy: keep master values unless selection specifies otherwise
	for key, value := range master {
		merged[key] = value
	}

	// Apply field selections
	for field, source := range selections {
		if source == "duplicate" {
			if val, ok := duplicate[field]; ok {
				merged[field] = val
			}
		}
		// "master" or default: already in merged from master copy
	}

	return merged
}

// Utility functions

func normalizePhone(phone string) string {
	// Remove all non-digit characters
	var digits strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	return digits.String()
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func calculateStringSimilarity(s1, s2 string) int {
	if s1 == s2 {
		return 100
	}

	distance := levenshtein.DistanceForStrings([]rune(s1), []rune(s2), levenshtein.DefaultOptions)
	maxLen := max(len(s1), len(s2))

	if maxLen == 0 {
		return 100
	}

	similarity := 100 - (distance * 100 / maxLen)
	if similarity < 0 {
		return 0
	}

	return similarity
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func convertMapToJSON(m map[string]interface{}) string {
	// Simple JSON serialization for snapshot storage
	// In production, use json.Marshal
	return fmt.Sprintf("%v", m)
}

func convertStringMapToJSONBMap(m map[string]string) types.JSONBMap {
	result := make(types.JSONBMap)
	for k, v := range m {
		result[k] = v
	}
	return result
}
