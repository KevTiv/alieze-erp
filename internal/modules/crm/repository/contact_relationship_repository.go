package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
)

// ContactRelationshipRepository handles relationship type and relationship operations
type ContactRelationshipRepository interface {
	// Relationship Types
	CreateRelationshipType(ctx context.Context, relType *types.RelationshipType) error
	GetRelationshipType(ctx context.Context, id uuid.UUID) (*types.RelationshipType, error)
	UpdateRelationshipType(ctx context.Context, relType *types.RelationshipType) error
	DeleteRelationshipType(ctx context.Context, id uuid.UUID) error
	ListRelationshipTypes(ctx context.Context, filter types.RelationshipTypeFilter) ([]*types.RelationshipType, error)

	// Relationships
	CreateRelationship(ctx context.Context, rel *types.ContactRelationship) error
	GetRelationship(ctx context.Context, id uuid.UUID) (*types.ContactRelationship, error)
	UpdateRelationship(ctx context.Context, rel *types.ContactRelationship) error
	DeleteRelationship(ctx context.Context, id uuid.UUID) error
	ListRelationships(ctx context.Context, contactID uuid.UUID) ([]*types.ContactRelationship, error)

	// Relationship Strength & Interactions
	UpdateRelationshipStrength(ctx context.Context, id uuid.UUID, strength int) error
	RecordInteraction(ctx context.Context, id uuid.UUID) error

	// Network Analysis
	GetRelationshipNetwork(ctx context.Context, contactID uuid.UUID, depth int) (*types.RelationshipNetwork, error)
}

type contactRelationshipRepository struct {
	db *sql.DB
}

func NewContactRelationshipRepository(db *sql.DB) ContactRelationshipRepository {
	return &contactRelationshipRepository{db: db}
}

// CreateRelationshipType creates a new custom relationship type
func (r *contactRelationshipRepository) CreateRelationshipType(ctx context.Context, relType *types.RelationshipType) error {
	query := `
		INSERT INTO contact_relationship_types (
			id, organization_id, name, description,
			is_bidirectional, is_system, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		relType.ID,
		relType.OrganizationID,
		relType.Name,
		relType.Description,
		relType.IsBidirectional,
		relType.IsSystem,
		now,
		now,
	).Scan(&relType.CreatedAt, &relType.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create relationship type: %w", err)
	}

	return nil
}

// GetRelationshipType retrieves a relationship type by ID
func (r *contactRelationshipRepository) GetRelationshipType(ctx context.Context, id uuid.UUID) (*types.RelationshipType, error) {
	query := `
		SELECT id, organization_id, name, description,
		       is_bidirectional, is_system, created_at, updated_at
		FROM contact_relationship_types
		WHERE id = $1 AND deleted_at IS NULL
	`

	relType := &types.RelationshipType{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&relType.ID,
		&relType.OrganizationID,
		&relType.Name,
		&relType.Description,
		&relType.IsBidirectional,
		&relType.IsSystem,
		&relType.CreatedAt,
		&relType.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("relationship type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	return relType, nil
}

// UpdateRelationshipType updates a relationship type
func (r *contactRelationshipRepository) UpdateRelationshipType(ctx context.Context, relType *types.RelationshipType) error {
	query := `
		UPDATE contact_relationship_types
		SET name = $1, description = $2, is_bidirectional = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		relType.Name,
		relType.Description,
		relType.IsBidirectional,
		time.Now(),
		relType.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update relationship type: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("relationship type not found")
	}

	return nil
}

// DeleteRelationshipType soft deletes a relationship type
func (r *contactRelationshipRepository) DeleteRelationshipType(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE contact_relationship_types SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship type: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("relationship type not found")
	}

	return nil
}

// ListRelationshipTypes retrieves relationship types with filtering
func (r *contactRelationshipRepository) ListRelationshipTypes(ctx context.Context, filter types.RelationshipTypeFilter) ([]*types.RelationshipType, error) {
	query := `
		SELECT id, organization_id, name, description,
		       is_bidirectional, is_system, created_at, updated_at
		FROM contact_relationship_types
		WHERE organization_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.IsSystem != nil {
		query += fmt.Sprintf(" AND is_system = $%d", argPos)
		args = append(args, *filter.IsSystem)
		argPos++
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list relationship types: %w", err)
	}
	defer rows.Close()

	var relTypes []*types.RelationshipType
	for rows.Next() {
		relType := &types.RelationshipType{}
		err := rows.Scan(
			&relType.ID,
			&relType.OrganizationID,
			&relType.Name,
			&relType.Description,
			&relType.IsBidirectional,
			&relType.IsSystem,
			&relType.CreatedAt,
			&relType.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship type: %w", err)
		}
		relTypes = append(relTypes, relType)
	}

	return relTypes, nil
}

// CreateRelationship creates a new relationship between contacts
func (r *contactRelationshipRepository) CreateRelationship(ctx context.Context, rel *types.ContactRelationship) error {
	query := `
		INSERT INTO contact_relationships (
			id, organization_id, from_contact_id, to_contact_id,
			relationship_type_id, strength_score, last_interaction_date,
			interaction_count, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		rel.ID,
		rel.OrganizationID,
		rel.ContactID,
		rel.RelatedContactID,
		rel.Type,
		rel.Notes,
		now,
		now,
	).Scan(&rel.CreatedAt, &rel.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	return nil
}

// GetRelationship retrieves a relationship by ID
func (r *contactRelationshipRepository) GetRelationship(ctx context.Context, id uuid.UUID) (*types.ContactRelationship, error) {
	query := `
		SELECT id, organization_id, from_contact_id, to_contact_id,
		       relationship_type_id, strength_score, last_interaction_date,
		       interaction_count, metadata, created_at, updated_at
		FROM contact_relationships
		WHERE id = $1 AND deleted_at IS NULL
	`

	rel := &types.ContactRelationship{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rel.ID,
		&rel.OrganizationID,
		&rel.ContactID,
		&rel.RelatedContactID,
		&rel.Type,
		&rel.Notes,
		&rel.CreatedAt,
		&rel.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("relationship not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}

	return rel, nil
}

// UpdateRelationship updates a relationship
func (r *contactRelationshipRepository) UpdateRelationship(ctx context.Context, rel *types.ContactRelationship) error {
	query := `
		UPDATE contact_relationships
		SET relationship_type_id = $1, strength_score = $2,
		    last_interaction_date = $3, interaction_count = $4,
		    metadata = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		rel.RelationshipTypeID,
		rel.StrengthScore,
		rel.LastInteractionDate,
		rel.InteractionCount,
		rel.Metadata,
		time.Now(),
		rel.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("relationship not found")
	}

	return nil
}

// DeleteRelationship soft deletes a relationship
func (r *contactRelationshipRepository) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE contact_relationships SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("relationship not found")
	}

	return nil
}

// ListRelationships retrieves all relationships for a contact
func (r *contactRelationshipRepository) ListRelationships(ctx context.Context, contactID uuid.UUID) ([]*types.ContactRelationship, error) {
	query := `
		SELECT id, organization_id, from_contact_id, to_contact_id,
		       relationship_type_id, strength_score, last_interaction_date,
		       interaction_count, metadata, created_at, updated_at
		FROM contact_relationships
		WHERE (from_contact_id = $1 OR to_contact_id = $1) AND deleted_at IS NULL
		ORDER BY strength_score DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to list relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*types.ContactRelationship
	for rows.Next() {
		rel := &types.ContactRelationship{}
		err := rows.Scan(
			&rel.ID,
			&rel.OrganizationID,
			&rel.FromContactID,
			&rel.ToContactID,
			&rel.RelationshipTypeID,
			&rel.StrengthScore,
			&rel.LastInteractionDate,
			&rel.InteractionCount,
			&rel.Metadata,
			&rel.CreatedAt,
			&rel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

// UpdateRelationshipStrength updates the strength score of a relationship
func (r *contactRelationshipRepository) UpdateRelationshipStrength(ctx context.Context, id uuid.UUID, strength int) error {
	query := `
		UPDATE contact_relationships
		SET strength_score = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, strength, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update relationship strength: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("relationship not found")
	}

	return nil
}

// RecordInteraction increments interaction count and updates last interaction date
func (r *contactRelationshipRepository) RecordInteraction(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE contact_relationships
		SET interaction_count = interaction_count + 1,
		    last_interaction_date = $1,
		    updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to record interaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("relationship not found")
	}

	return nil
}

// GetRelationshipNetwork performs graph traversal to build a relationship network
func (r *contactRelationshipRepository) GetRelationshipNetwork(ctx context.Context, contactID uuid.UUID, depth int) (*types.RelationshipNetwork, error) {
	if depth < 1 || depth > 3 {
		return nil, fmt.Errorf("depth must be between 1 and 3")
	}

	network := &types.RelationshipNetwork{
		CenterContactID: contactID,
		Nodes:           make([]types.NetworkNode, 0),
		Edges:           make([]types.ContactRelationship, 0),
		Depth:           depth,
	}

	// Track visited contacts to avoid cycles
	visited := make(map[uuid.UUID]bool)
	visited[contactID] = true

	// Add center node
	network.Nodes = append(network.Nodes, types.NetworkNode{
		ContactID: contactID,
		Depth:     0,
	})

	// Perform BFS traversal
	currentLevel := []uuid.UUID{contactID}

	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		var nextLevel []uuid.UUID

		for _, currentContactID := range currentLevel {
			// Get all relationships for this contact
			relationships, err := r.ListRelationships(ctx, currentContactID)
			if err != nil {
				return nil, fmt.Errorf("failed to get relationships at depth %d: %w", currentDepth, err)
			}

			for _, rel := range relationships {
				// Determine the connected contact
				var connectedContactID uuid.UUID
				if rel.FromContactID == currentContactID {
					connectedContactID = rel.ToContactID
				} else {
					connectedContactID = rel.FromContactID
				}

				// Skip if already visited
				if visited[connectedContactID] {
					continue
				}

				// Mark as visited
				visited[connectedContactID] = true

				// Add node
				network.Nodes = append(network.Nodes, types.NetworkNode{
					ContactID: connectedContactID,
					Depth:     currentDepth,
				})

				// Add edge
				network.Edges = append(network.Edges, *rel)

				// Add to next level for continued traversal
				nextLevel = append(nextLevel, connectedContactID)
			}
		}

		// Move to next level
		currentLevel = nextLevel
	}

	network.TotalNodes = len(network.Nodes)
	network.TotalEdges = len(network.Edges)

	return network, nil
}
