# CRM Service Standardization Guide

## Overview

This guide defines the standards and patterns for CRM services in the ERP system. The goal is to ensure consistency, maintainability, and testability across all CRM modules.

## Architecture Principles

### 1. Layered Architecture
- **Handler Layer**: HTTP request/response handling
- **Service Layer**: Business logic and orchestration
- **Repository Layer**: Data access and persistence
- **Base Layer**: Shared infrastructure and utilities

### 2. Interface-First Design
- All repositories implement standardized interfaces
- Services depend on interfaces, not concrete implementations
- Easy testing with mock implementations

### 3. Generic CRUD Patterns
- Use generic base service for common CRUD operations
- Extend with domain-specific methods when needed
- Consistent method signatures across services

## Service Standards

### Constructor Pattern
```go
// Standard constructor with options
func New[Entity]Service(
    repo base.Repository[Entity, Filter],
    authService auth.Service,
    opts base.ServiceOptions,
) *[Entity]Service {
    return &[Entity]Service{
        CRUDService: base.NewCRUDService[Entity, Request, Update, Filter](repo, authService, opts),
    }
}
```

### Method Naming Conventions
- **Create**: `Create[Entity](ctx context.Context, req Request) (*Entity, error)`
- **Read**: `Get[Entity](ctx context.Context, id uuid.UUID) (*Entity, error)`
- **Update**: `Update[Entity](ctx context.Context, id uuid.UUID, req UpdateRequest) (*Entity, error)`
- **Delete**: `Delete[Entity](ctx context.Context, id uuid.UUID) error`
- **List**: `List[Entities](ctx context.Context, filter Filter) ([]*Entity, int, error)`

### Error Handling Standards
- Use custom error types from `pkg/crm/errors`
- Wrap errors with context and operation codes
- Return appropriate HTTP status codes
- Log errors at service layer with structured data

### Validation Standards
- Validate at service layer before repository operations
- Use shared validation utilities from `pkg/crm/validation`
- Return validation errors with field-level details
- Support both rule engine and hardcoded validation

## Repository Standards

### Interface Contract
```go
type Repository[Entity, Filter] interface {
    Create(ctx context.Context, entity Entity) (*Entity, error)
    FindByID(ctx context.Context, id uuid.UUID) (*Entity, error)
    FindAll(ctx context.Context, filter Filter) ([]*Entity, error)
    Update(ctx context.Context, entity Entity) (*Entity, error)
    Delete(ctx context.Context, id uuid.UUID) error
    Count(ctx context.Context, filter Filter) (int, error)
}
```

### Extended Interface (Optional)
```go
type ExtendedRepository[Entity, Filter] interface {
    Repository[Entity, Filter]
    FindByOrganization(ctx context.Context, orgID uuid.UUID, filter Filter) ([]*Entity, error)
    Exists(ctx context.Context, orgID, id uuid.UUID) (bool, error)
}
```

### Implementation Standards
- Use context for all database operations
- Support organization-scoped queries
- Implement soft deletes where applicable
- Use parameterized queries for security
- Handle database errors appropriately

## Testing Standards

### Test Structure
```go
type [Entity]ServiceTestSuite struct {
    suite.Suite
    service   *[Entity]Service
    repo      *testutils.MockCRUDRepository[Entity, Filter]
    auth      *testutils.MockAuthService
    ctx       context.Context
    orgID     uuid.UUID
    userID    uuid.UUID
}
```

### Test Coverage Requirements
- **Success Scenarios**: Happy path testing
- **Validation Errors**: Input validation testing
- **Permission Errors**: Authorization testing
- **Organization Access**: Multi-tenancy testing
- **Repository Errors**: Database error handling
- **Minimum Coverage**: 90%

### Test Data Standards
- Use test data generators for consistent test data
- Generate realistic data with proper relationships
- Support both valid and invalid test data
- Use factories for complex object creation

## Event Standards

### Event Publishing
- Publish events for all significant state changes
- Use consistent event naming: `[entity].[action]`
- Include relevant context data in events
- Handle event publishing failures gracefully

### Event Types
- `contact.created`, `contact.updated`, `contact.deleted`
- `lead.created`, `lead.converted`, `lead.assigned`
- `activity.created`, `activity.completed`
- `sales_team.created`, `sales_team.updated`

## Migration Standards

### Backward Compatibility
- Maintain existing API contracts during migration
- Use adapter pattern to bridge old and new implementations
- Deprecate old methods before removal
- Provide migration period for consumers

### Incremental Migration
- Migrate one service at a time
- Maintain test coverage throughout migration
- Use feature flags for gradual rollout
- Monitor performance during transition

## Code Quality Standards

### Linting and Formatting
- Use `gofmt` for code formatting
- Follow `golangci-lint` rules
- Enforce consistent naming conventions
- Use proper package documentation

### Documentation
- Document all public interfaces
- Provide usage examples
- Document error conditions
- Include performance considerations

## Security Standards

### Authorization
- Check organization access for all operations
- Validate user permissions at service layer
- Use principle of least privilege
- Audit access patterns

### Data Validation
- Validate all input data
- Sanitize data before persistence
- Prevent SQL injection with parameterized queries
- Handle sensitive data appropriately

## Performance Standards

### Database Operations
- Use connection pooling
- Implement proper indexing
- Avoid N+1 query patterns
- Use transactions for complex operations

### Caching Strategy
- Cache frequently accessed data
- Implement cache invalidation
- Use appropriate cache TTL
- Monitor cache hit rates

## Monitoring Standards

### Logging
- Use structured logging with context
- Log at appropriate levels (debug, info, warn, error)
- Include correlation IDs for request tracing
- Avoid logging sensitive data

### Metrics
- Track service response times
- Monitor error rates
- Measure database performance
- Track business metrics

## Compliance Standards

### Audit Trail
- Log all data modifications
- Track user actions
- Maintain immutable audit logs
- Support audit queries

### Data Privacy
- Implement data retention policies
- Support data deletion requests
- Handle personal data appropriately
- Comply with privacy regulations