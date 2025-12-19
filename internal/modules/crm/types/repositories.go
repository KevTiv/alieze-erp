package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Core interfaces for basic operations
type CRUDRepository[T any, F any] interface {
	Create(ctx context.Context, entity T) (*T, error)
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	FindAll(ctx context.Context, filter F) ([]T, error)
	Update(ctx context.Context, entity T) (*T, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter F) (int, error)
}

// ContactRepository extends CRUD with contact-specific methods
type ContactRepository interface {
	CRUDRepository[Contact, ContactFilter]

	// Relationship methods
	CreateRelationship(ctx context.Context, relationship *ContactRelationship) error
	FindRelationships(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, relationshipType string, limit int) ([]*ContactRelationship, error)
	ContactExists(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) (bool, error)

	// Segment and tag methods
	AddContactToSegments(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, segmentIDs []string) error
	AddContactTags(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID, tags []string) error
}

// ContactTagRepository
type ContactTagRepository interface {
	CRUDRepository[ContactTag, ContactTagFilter]
	FindByContact(ctx context.Context, contactID uuid.UUID) ([]ContactTag, error)
}

// LeadRepository extends CRUD with lead-specific methods
type LeadRepository interface {
	CRUDRepository[Lead, LeadFilter]

	// Date range queries
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]Lead, error)
	FindByDeadlineRange(ctx context.Context, startDate, endDate time.Time) ([]Lead, error)

	// Utility methods
	CountByStage(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error)
	FindOverdue(ctx context.Context) ([]Lead, error)
	FindHighValue(ctx context.Context, minValue float64) ([]Lead, error)
	FindBySearchTerm(ctx context.Context, searchTerm string) ([]Lead, error)
}

// Other domain repositories
type LeadStageRepository interface {
	CRUDRepository[LeadStage, LeadStageFilter]
}

type LeadSourceRepository interface {
	CRUDRepository[LeadSource, LeadSourceFilter]
}

type LostReasonRepository interface {
	CRUDRepository[LostReason, LostReasonFilter]
}

type SalesTeamRepository interface {
	CRUDRepository[SalesTeam, SalesTeamFilter]
	FindByMember(ctx context.Context, userID uuid.UUID) ([]SalesTeam, error)
}

type ActivityRepository interface {
	CRUDRepository[Activity, ActivityFilter]
	FindByContact(ctx context.Context, contactID uuid.UUID) ([]Activity, error)
	FindByLead(ctx context.Context, leadID uuid.UUID) ([]Activity, error)
}

type AssignmentRuleRepository interface {
	Create(ctx context.Context, rule AssignmentRule) (*AssignmentRule, error)
	FindByID(ctx context.Context, id uuid.UUID) (*AssignmentRule, error)
	FindAll(ctx context.Context, limit, offset int) ([]AssignmentRule, error)
	Update(ctx context.Context, rule AssignmentRule) (*AssignmentRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByTargetModel(ctx context.Context, targetModel AssignmentTargetModel) ([]AssignmentRule, error)
	FindActiveRules(ctx context.Context, targetModel AssignmentTargetModel) ([]AssignmentRule, error)
	// Legacy methods for backward compatibility
	UpdateAssignmentRule(ctx context.Context, rule *AssignmentRule) error
	DeleteAssignmentRule(ctx context.Context, id uuid.UUID) error
	ListAssignmentRules(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*AssignmentRule, error)
	CreateTerritory(ctx context.Context, territory *Territory) error
	GetTerritory(ctx context.Context, id uuid.UUID) (*Territory, error)
	UpdateTerritory(ctx context.Context, territory *Territory) error
	DeleteTerritory(ctx context.Context, id uuid.UUID) error
	ListTerritories(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*Territory, error)
	GetNextAssignee(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error)
	GetAssignmentStatsByUser(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*AssignmentStatsByUser, error)
	GetAssignmentRuleEffectiveness(ctx context.Context, orgID uuid.UUID) ([]*AssignmentRuleEffectiveness, error)
	AssignLead(ctx context.Context, leadID uuid.UUID, assigneeID uuid.UUID, reason string) error
	GetLead(ctx context.Context, leadID uuid.UUID) (*Lead, error)
}
