# CRM Service Migration Guide

## Overview

This guide provides step-by-step instructions for migrating existing CRM services to the new standardized architecture. The migration is designed to be incremental, maintaining backward compatibility while improving code quality and consistency.

## Migration Strategy

### Principles
1. **Incremental Migration**: Migrate one service at a time
2. **Backward Compatibility**: Maintain existing API contracts during transition
3. **Test-Driven**: Ensure all tests pass before and after migration
4. **Zero Downtime**: Services continue to function during migration

### Phases
1. **Adapter Creation**: Bridge old and new interfaces
2. **Service Refactoring**: Migrate service implementation
3. **Test Migration**: Update and extend test coverage
4. **Cleanup**: Remove old implementation and adapters

## Step-by-Step Migration Process

### Phase 1: Prepare Repository Adapter

#### 1.1 Create Repository Adapter
```go
// internal/modules/crm/repository/[entity]_adapter.go
package repository

import (
    "context"
    "github.com/google/uuid"
    "alieze-erp/pkg/crm/base"
    "alieze-erp/internal/modules/crm/types"
)

type [Entity]RepositoryAdapter struct {
    repo *[Entity]Repository // Old repository
}

func New[Entity]RepositoryAdapter(repo *[Entity]Repository) *ContactRepositoryAdapter {
    return &[Entity]RepositoryAdapter{repo: repo}
}

// Implement base.Repository[Entity, Filter] interface
func (a *[Entity]RepositoryAdapter) Create(ctx context.Context, entity types.[Entity]) (*types.[Entity], error) {
    return a.repo.Create(ctx, entity)
}

func (a *[Entity]RepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*types.[Entity], error) {
    return a.repo.FindByID(ctx, id)
}

func (a *[Entity]RepositoryAdapter) FindAll(ctx context.Context, filter types.[Entity]Filter) ([]*types.[Entity], error) {
    return a.repo.FindAll(ctx, filter)
}

func (a *[Entity]RepositoryAdapter) Update(ctx context.Context, entity types.[Entity]) (*types.[Entity], error) {
    return a.repo.Update(ctx, entity)
}

func (a *[Entity]RepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
    return a.repo.Delete(ctx, id)
}

func (a *[Entity]RepositoryAdapter) Count(ctx context.Context, filter types.[Entity]Filter) (int, error) {
    return a.repo.Count(ctx, filter)
}
```

#### 1.2 Verify Interface Compliance
```go
// Add compile-time check
var _ base.Repository[types.[Entity], types.[Entity]Filter] = (*[Entity]RepositoryAdapter)(nil)
```

### Phase 2: Create New Service Implementation

#### 2.1 Create New Service File
```go
// internal/modules/crm/service/[entity]_service_v2.go
package service

import (
    "context"
    "github.com/google/uuid"
    "log/slog"
    "alieze-erp/pkg/crm/base"
    "alieze-erp/pkg/crm/errors"
    "alieze-erp/pkg/crm/validation"
    "alieze-erp/internal/modules/crm/types"
)

type [Entity]ServiceV2 struct {
    *base.CRUDService[types.[Entity], types.[Entity]Request, types.[Entity]UpdateRequest, types.[Entity]Filter]
}

func New[Entity]ServiceV2(
    repo base.Repository[types.[Entity], types.[Entity]Filter],
    authService base.AuthService,
    logger *slog.Logger,
) *[Entity]ServiceV2 {
    return &[Entity]ServiceV2{
        CRUDService: base.NewCRUDService[types.[Entity], types.[Entity]Request, types.[Entity]UpdateRequest, types.[Entity]Filter](
            repo, authService, base.ServiceOptions{Logger: logger},
        ),
    }
}

// Create[Entity] implements the standardized create operation
func (s *[Entity]ServiceV2) Create[Entity](ctx context.Context, req types.[Entity]Request) (*types.[Entity], error) {
    // Validate input
    if err := s.validate[Entity]Request(req); err != nil {
        return nil, err
    }
    
    // Convert request to entity
    entity := s.requestTo[Entity](req)
    
    // Check authorization
    if err := s.GetAuthService().CheckOrganizationAccess(ctx, entity.OrganizationID); err != nil {
        return nil, errors.ErrOrganizationAccess
    }
    
    // Create entity
    result, err := s.GetRepository().Create(ctx, entity)
    if err != nil {
        return nil, errors.Wrap(err, "CREATE_FAILED", "failed to create [entity]")
    }
    
    // Log operation
    s.LogOperation(ctx, "create_[entity]", result.ID, map[string]interface{}{
        "organization_id": entity.OrganizationID,
        "name":           entity.Name, // or other identifier
    })
    
    // Publish event
    s.PublishEvent(ctx, "[entity].created", result)
    
    return result, nil
}

// Get[Entity] implements the standardized get operation
func (s *[Entity]ServiceV2) Get[Entity](ctx context.Context, id uuid.UUID) (*types.[Entity], error) {
    result, err := s.GetRepository().FindByID(ctx, id)
    if err != nil {
        return nil, errors.Wrap(err, "GET_FAILED", "failed to get [entity]")
    }
    
    if result == nil {
        return nil, errors.ErrNotFound
    }
    
    // Check authorization
    if err := s.GetAuthService().CheckOrganizationAccess(ctx, result.OrganizationID); err != nil {
        return nil, errors.ErrOrganizationAccess
    }
    
    return result, nil
}

// Update[Entity] implements the standardized update operation
func (s *[Entity]ServiceV2) Update[Entity](ctx context.Context, id uuid.UUID, req types.[Entity]UpdateRequest) (*types.[Entity], error) {
    // Get existing entity
    existing, err := s.Get[Entity](ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Validate update request
    if err := s.validate[Entity]UpdateRequest(req); err != nil {
        return nil, err
    }
    
    // Apply updates
    updated := s.apply[Entity]Update(existing, req)
    
    // Update entity
    result, err := s.GetRepository().Update(ctx, *updated)
    if err != nil {
        return nil, errors.Wrap(err, "UPDATE_FAILED", "failed to update [entity]")
    }
    
    // Log operation
    s.LogOperation(ctx, "update_[entity]", result.ID, map[string]interface{}{
        "organization_id": result.OrganizationID,
        "changes":        s.getChanges(existing, result),
    })
    
    // Publish event
    s.PublishEvent(ctx, "[entity].updated", result)
    
    return result, nil
}

// Delete[Entity] implements the standardized delete operation
func (s *[Entity]ServiceV2) Delete[Entity](ctx context.Context, id uuid.UUID) error {
    // Get existing entity for authorization
    existing, err := s.Get[Entity](ctx, id)
    if err != nil {
        return err
    }
    
    // Delete entity
    err = s.GetRepository().Delete(ctx, id)
    if err != nil {
        return errors.Wrap(err, "DELETE_FAILED", "failed to delete [entity]")
    }
    
    // Log operation
    s.LogOperation(ctx, "delete_[entity]", id, map[string]interface{}{
        "organization_id": existing.OrganizationID,
    })
    
    // Publish event
    s.PublishEvent(ctx, "[entity].deleted", map[string]interface{}{
        "id": id,
        "organization_id": existing.OrganizationID,
    })
    
    return nil
}

// List[Entities] implements the standardized list operation
func (s *[Entity]ServiceV2) List[Entities](ctx context.Context, filter types.[Entity]Filter) ([]*types.[Entity], int, error) {
    // Validate filter
    if err := s.validate[Entity]Filter(filter); err != nil {
        return nil, 0, err
    }
    
    // Check organization access
    if filter.OrganizationID != nil {
        if err := s.GetAuthService().CheckOrganizationAccess(ctx, *filter.OrganizationID); err != nil {
            return nil, 0, errors.ErrOrganizationAccess
        }
    }
    
    // Get entities
    entities, err := s.GetRepository().FindAll(ctx, filter)
    if err != nil {
        return nil, 0, errors.Wrap(err, "LIST_FAILED", "failed to list [entities]")
    }
    
    // Get count
    count, err := s.GetRepository().Count(ctx, filter)
    if err != nil {
        return nil, 0, errors.Wrap(err, "COUNT_FAILED", "failed to count [entities]")
    }
    
    return entities, count, nil
}

// Helper methods
func (s *[Entity]ServiceV2) validate[Entity]Request(req types.[Entity]Request) error {
    return validation.ValidateMultiple(
        func() error { return validation.ValidateRequired("name", req.Name) },
        func() error { return validation.ValidateLength("name", req.Name, 1, 255) },
        // Add other validation rules
    )
}

func (s *[Entity]ServiceV2) requestTo[Entity](req types.[Entity]Request) types.[Entity] {
    return types.[Entity]{
        // Map request fields to entity
    }
}
```

### Phase 3: Update Tests

#### 3.1 Create New Test Suite
```go
// internal/modules/crm/service_test/[entity]_service_v2_test.go
package service_test

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    "alieze-erp/pkg/crm/base"
    "alieze-erp/pkg/crm/errors"
    testutils "alieze-erp/internal/testutils"
    "alieze-erp/internal/modules/crm/service"
    "alieze-erp/internal/modules/crm/types"
)

type [Entity]ServiceV2TestSuite struct {
    suite.Suite
    service   *service.[Entity]ServiceV2
    repo      *testutils.MockCRUDRepository[types.[Entity], types.[Entity]Filter]
    auth      *testutils.MockAuthService
    ctx       context.Context
    orgID     uuid.UUID
    userID    uuid.UUID
}

func (s *[Entity]ServiceV2TestSuite) SetupTest() {
    s.ctx = context.Background()
    s.orgID = uuid.New()
    s.userID = uuid.New()
    
    s.repo = testutils.NewMockCRUDRepository[types.[Entity], types.[Entity]Filter]()
    s.auth = testutils.NewMockAuthService()
    
    s.service = service.New[Entity]ServiceV2(s.repo, s.auth, nil)
    
    // Setup default auth expectations
    s.auth.On("CheckOrganizationAccess", mock.Anything, s.orgID).Return(nil)
    s.auth.On("GetCurrentUser", mock.Anything).Return(&types.User{ID: s.userID, OrganizationID: s.orgID}, nil)
}

func (s *[Entity]ServiceV2TestSuite) TestCreate[Entity]_Success() {
    // Arrange
    req := types.[Entity]Request{
        Name: "Test [Entity]",
        OrganizationID: s.orgID,
    }
    
    expected := &types.[Entity]{
        ID: uuid.New(),
        Name: req.Name,
        OrganizationID: req.OrganizationID,
    }
    
    s.repo.WithCreateFunc(func(ctx context.Context, entity types.[Entity]) (*types.[Entity], error) {
        s.Equal(req.Name, entity.Name)
        s.Equal(req.OrganizationID, entity.OrganizationID)
        return expected, nil
    })
    
    // Act
    result, err := s.service.Create[Entity](s.ctx, req)
    
    // Assert
    s.NoError(err)
    s.Equal(expected.ID, result.ID)
    s.Equal(expected.Name, result.Name)
    s.repo.AssertExpectations(s.T())
}

func (s *[Entity]ServiceV2TestSuite) TestCreate[Entity]_ValidationError() {
    // Arrange
    req := types.[Entity]Request{
        Name: "", // Invalid: empty name
        OrganizationID: s.orgID,
    }
    
    // Act
    result, err := s.service.Create[Entity](s.ctx, req)
    
    // Assert
    s.Error(err)
    s.Nil(result)
    s.Contains(err.Error(), "validation error")
}

func (s *[Entity]ServiceV2TestSuite) TestGet[Entity]_Success() {
    // Arrange
    expected := &types.[Entity]{
        ID: uuid.New(),
        Name: "Test [Entity]",
        OrganizationID: s.orgID,
    }
    
    s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.[Entity], error) {
        return expected, nil
    })
    
    // Act
    result, err := s.service.Get[Entity](s.ctx, expected.ID)
    
    // Assert
    s.NoError(err)
    s.Equal(expected.ID, result.ID)
    s.repo.AssertExpectations(s.T())
}

func (s *[Entity]ServiceV2TestSuite) TestGet[Entity]_NotFound() {
    // Arrange
    id := uuid.New()
    s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.[Entity], error) {
        return nil, nil
    })
    
    // Act
    result, err := s.service.Get[Entity](s.ctx, id)
    
    // Assert
    s.Error(err)
    s.Nil(result)
    s.Equal(errors.ErrNotFound, err)
}

func Test[Entity]ServiceV2TestSuite(t *testing.T) {
    suite.Run(t, new([Entity]ServiceV2TestSuite))
}
```

### Phase 4: Update Module Registration

#### 4.1 Update Module Constructor
```go
// internal/modules/crm/module.go
func (m *CRMModule) initializeServices() {
    // Create repository adapters
    contactAdapter := repository.NewContactRepositoryAdapter(m.contactRepo)
    
    // Create new services
    m.contactServiceV2 = service.NewContactServiceV2(
        contactAdapter,
        m.authService,
        m.logger,
    )
    
    // Keep old service for backward compatibility
    m.contactService = service.NewContactService(m.contactRepo, m.authService)
}
```

### Phase 5: Gradual Transition

#### 5.1 Feature Flag Implementation
```go
// pkg/crm/feature_flags.go
package crm

type FeatureFlags struct {
    UseContactServiceV2 bool
    // Add other feature flags
}

func (m *CRMModule) GetContactService() interface{} {
    if m.featureFlags.UseContactServiceV2 {
        return m.contactServiceV2
    }
    return m.contactService
}
```

#### 5.2 Handler Updates
```go
// internal/modules/crm/handler/contact_handler.go
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
    // Check feature flag
    var contact interface{}
    var err error
    
    if h.useV2Service {
        // Use new service
        contact, err = h.serviceV2.CreateContact(r.Context(), req)
    } else {
        // Use old service
        contact, err = h.service.CreateContact(r.Context(), req)
    }
    
    // Handle response...
}
```

### Phase 6: Cleanup

#### 6.1 Remove Old Implementation
Once the new service has been tested and is stable:

1. **Remove Old Service File**: Delete the old service implementation
2. **Remove Adapters**: Delete repository adapter files
3. **Update Imports**: Remove old imports throughout the codebase
4. **Remove Feature Flags**: Clean up feature flag code
5. **Update Documentation**: Update API documentation

## Testing Strategy

### 1. Unit Tests
- Ensure 100% test coverage for new service implementation
- Test all CRUD operations
- Test validation scenarios
- Test error handling
- Test authorization

### 2. Integration Tests
- Test service with real database
- Test transaction scenarios
- Test concurrent operations
- Test event publishing

### 3. Performance Tests
- Compare performance between old and new services
- Ensure no regression in response times
- Test with large datasets

### 4. Compatibility Tests
- Test old and new services return same results
- Test API contracts remain unchanged
- Test error responses are consistent

## Common Migration Patterns

### Handling Extended Repository Methods
For repositories with methods beyond standard CRUD:

```go
type ExtendedContactRepositoryAdapter struct {
    *ContactRepositoryAdapter
    repo *repository.ContactRepository
}

func (a *ExtendedContactRepositoryAdapter) CreateRelationship(ctx context.Context, relationship *types.ContactRelationship) error {
    return a.repo.CreateRelationship(ctx, relationship)
}
```

### Handling Complex Business Logic
For services with complex domain-specific operations:

```go
func (s *ContactServiceV2) CreateRelationship(ctx context.Context, req types.CreateRelationshipRequest) error {
    // Use base CRUD service for standard operations
    // Implement domain-specific logic here
    // Delegate to old service for complex operations during migration
    return s.oldService.CreateRelationship(ctx, req)
}
```

### Handling Event Bus Integration
For services that publish events:

```go
func New[Entity]ServiceV2(
    repo base.Repository[types.[Entity], types.[Entity]Filter],
    authService base.AuthService,
    eventBus *events.Bus,
    logger *slog.Logger,
) *[Entity]ServiceV2 {
    return &[Entity]ServiceV2{
        CRUDService: base.NewCRUDService[types.[Entity], types.[Entity]Request, types.[Entity]UpdateRequest, types.[Entity]Filter](
            repo, authService, base.ServiceOptions{
                EventBus: eventBus,
                Logger:  logger,
            },
        ),
    }
}
```

## Rollback Plan

### If Migration Fails
1. **Revert Feature Flags**: Disable new service implementation
2. **Restore Old Service**: Ensure old service is fully functional
3. **Fix Issues**: Address migration problems
4. **Retry Migration**: Resume migration process

### Data Consistency
- Ensure no data corruption during migration
- Validate database integrity
- Test backup and restore procedures

## Success Criteria

### Technical Success
- All tests pass (>90% coverage)
- Performance meets or exceeds baseline
- No breaking changes to public APIs
- Zero downtime during migration

### Business Success
- Improved code maintainability
- Reduced development time for new features
- Better error handling and logging
- Enhanced testing capabilities

## Timeline

| Week | Activities |
|------|------------|
| 1 | Phase 1-2: Preparation and infrastructure setup |
| 2 | Phase 3: Pilot migration (ContactService) |
| 3-4 | Phase 4: Gradual service migration |
| 5 | Phase 5: Testing and validation |
| 6 | Phase 6: Cleanup and documentation |

## Support and Resources

### Documentation
- CRM Service Standardization Guide
- API Documentation
- Database Schema Documentation

### Tools
- Migration scripts
- Test data generators
- Performance benchmarking tools

### Team Support
- Code review process
- Pair programming sessions
- Knowledge transfer sessions