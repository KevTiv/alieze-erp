# ERP Implementation Roadmap

This document outlines the implementation plan for making the Alieze ERP system usable as a basic ERP application.

## Current State

### What's Already Implemented
- ✅ **Database Schema**: Complete with 60+ migration files covering all ERP modules
- ✅ **Core Infrastructure**: Go API server with database connectivity
- ✅ **Migration System**: Automatic database migration handling
- ✅ **Basic API**: HTTP server with routing and middleware
- ✅ **Frontend Foundation**: React/Vite setup with basic backend connectivity

### What's Missing for Basic ERP Usability
- ❌ **Authentication/Authorization**: User management, JWT, permissions
- ❌ **API Endpoints**: Module-specific REST endpoints
- ✅ **Business Logic**: Service layer implementation (CRM & Products completed)
- ✅ **Data Access**: Repository patterns and database queries (CRM & Products completed with pgx)
- ❌ **Frontend UI**: ERP-specific components and views

## Implementation Priority Order

### Phase 1: Foundation Services (Critical - 2-3 weeks)
**Goal**: Basic authentication and core infrastructure for all modules

1. **Authentication System** (Highest Priority)
   - User registration and login endpoints
   - JWT token generation and validation
   - Password hashing and security
   - Session management
   - Basic user profile management

2. **Authorization System**
   - Role-based access control (RBAC)
   - Permission middleware for API endpoints
   - Organization/user context handling
   - Basic permission policies implementation

3. **Core Service Layer**
   - Base service interface and implementations
   - Error handling and logging framework
   - Request validation framework
   - API response standardization

### Phase 2: Core ERP Modules (High Priority - 4-6 weeks)
**Goal**: Implement basic CRUD operations for essential business modules

1. **CRM Module**
   - Contacts management (create, read, update, delete)
   - Leads management with status tracking
   - Basic activity logging
   - Contact search and filtering

2. **Products & Inventory Module**
   - Product catalog management
   - Inventory locations and stock tracking
   - Basic stock movements
   - Product search and categorization

3. **Sales Module**
   - Sales order creation and management
   - Pricelist management
   - Order status tracking
   - Basic sales reporting

4. **Accounting Module**
   - Chart of accounts management
   - Basic invoice creation
   - Payment processing
   - Simple financial reporting

### Phase 3: Business Logic & Integration (Medium Priority - 3-4 weeks)
**Goal**: Implement business workflows and module integration

1. **Business Workflows**
   - Lead to opportunity conversion
   - Sales order to invoice workflow
   - Inventory reservation and allocation
   - Basic accounting journal entries

2. **Module Integration**
   - CRM to Sales integration
   - Sales to Accounting integration
   - Inventory to Sales integration
   - Basic data consistency checks

3. **Validation Rules**
   - Business rule validation
   - Data integrity constraints
   - Workflow state transitions
   - Permission-based validation

### Phase 4: Frontend Implementation (Medium Priority - 4-6 weeks)
**Goal**: Basic usable UI for core ERP functions

1. **Authentication UI**
   - Login/registration forms
   - User profile management
   - Session handling

2. **Module-Specific UI**
   - CRM: Contacts list, lead management
   - Products: Catalog management, inventory views
   - Sales: Order creation, order lists
   - Accounting: Invoice management, basic reports

3. **Navigation & Layout**
   - Main dashboard layout
   - Module navigation
   - Responsive design foundation

### Phase 5: Basic Analytics & Reporting (Low Priority - 2-3 weeks)
**Goal**: Essential business insights

1. **Dashboard Widgets**
   - Sales overview
   - Inventory status
   - CRM activity summary
   - Financial snapshot

2. **Basic Reports**
   - Sales by period
   - Inventory valuation
   - Aging receivables
   - Customer activity

## Detailed Implementation Plan

### Phase 1: Foundation Services

#### 1.1 Authentication System Implementation
- **Backend**:
  - User registration endpoint (`POST /auth/register`)
  - User login endpoint (`POST /auth/login`)
  - JWT token generation and validation middleware
  - Password hashing with bcrypt
  - User profile endpoints (`GET/PUT /auth/profile`)
  - Session management with refresh tokens

- **Database**:
  - Utilize existing `organization_users` table
  - Extend with password hashing fields if needed
  - Session tracking table

- **Frontend**:
  - Login form component
  - Registration form component
  - Auth context for token management
  - Protected route wrapper

#### 1.2 Authorization System Implementation
- **Backend**:
  - Role-based permission middleware
  - Organization context middleware
  - Permission checking functions
  - Basic RBAC policy implementation

- **Database**:
  - Utilize existing permission tables
  - Implement basic role assignments
  - Organization membership validation

- **Frontend**:
  - Permission-based UI rendering
  - Role-aware navigation
  - Access denied handling

### Phase 2: Core ERP Modules

#### 2.1 CRM Module Implementation
- **Backend Services**:
  - ContactService: CRUD operations for contacts
  - LeadService: Lead management with status workflow
  - ActivityService: Activity logging
  - SearchService: Contact/lead search functionality

- **API Endpoints**:
  - `GET/POST/PUT/DELETE /crm/contacts`
  - `GET/POST/PUT /crm/contacts/:id`
  - `GET/POST/PUT /crm/leads`
  - `GET/POST /crm/activities`
  - `GET /crm/search`

- **Frontend Components**:
  - ContactList: Table view with filtering
  - ContactForm: Create/edit contact
  - LeadPipeline: Kanban-style lead management
  - ActivityFeed: Recent activities view

#### 2.2 Products & Inventory Module Implementation
- **Backend Services**:
  - ProductService: Product catalog management
  - InventoryService: Stock tracking and movements
  - CategoryService: Product categorization
  - LocationService: Warehouse/location management

- **API Endpoints**:
  - `GET/POST/PUT/DELETE /products`
  - `GET/POST/PUT /products/:id`
  - `GET/POST /inventory/movements`
  - `GET /inventory/stock-levels`
  - `GET/POST/PUT /products/categories`

- **Frontend Components**:
  - ProductCatalog: Grid/list view with search
  - ProductForm: Detailed product editing
  - InventoryDashboard: Stock overview
  - StockMovementForm: Inventory adjustments

#### 2.3 Sales Module Implementation
- **Backend Services**:
  - SalesOrderService: Order creation and management
  - PricelistService: Pricing management
  - OrderLineService: Line item management
  - OrderWorkflowService: Status transitions

- **API Endpoints**:
  - `GET/POST/PUT /sales/orders`
  - `GET/POST/PUT /sales/orders/:id`
  - `GET/POST/PUT /sales/pricelists`
  - `POST /sales/orders/:id/confirm`
  - `GET /sales/orders/:id/pdf` (future)

- **Frontend Components**:
  - SalesOrderList: Order overview table
  - SalesOrderForm: Multi-step order creation
  - PricelistManager: Price configuration
  - OrderStatusTracker: Workflow visualization

#### 2.4 Accounting Module Implementation
- **Backend Services**:
  - ChartOfAccountsService: Account management
  - InvoiceService: Invoice creation and management
  - PaymentService: Payment processing
  - JournalEntryService: Basic accounting entries

- **API Endpoints**:
  - `GET/POST/PUT /accounting/accounts`
  - `GET/POST/PUT /accounting/invoices`
  - `GET/POST /accounting/payments`
  - `GET /accounting/journal-entries`
  - `GET /accounting/balances`

- **Frontend Components**:
  - ChartOfAccounts: Hierarchical account view
  - InvoiceList: Invoice overview
  - InvoiceForm: Invoice creation/editing
  - PaymentProcessing: Payment recording
  - TrialBalance: Basic financial report

## Technical Implementation Approach

### Backend Architecture

```
┌─────────────────────────────────────────────────┐
│                 API Layer (HTTP)                 │
└─────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────┐
│               Service Layer (Business Logic)     │
└─────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────┐
│             Repository Layer (Data Access)       │
└─────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────┐
│               Database (PostgreSQL)               │
└─────────────────────────────────────────────────┘
```

### Service Layer Pattern

Each module will follow this pattern:

```go
type ContactService struct {
    repo        ContactRepository
    authService AuthService
    logger      *log.Logger
}

func NewContactService(repo ContactRepository, authService AuthService) *ContactService {
    return &ContactService{
        repo:        repo,
        authService: authService,
        logger:      log.New(os.Stdout, "contact-service: ", log.LstdFlags),
    }
}

func (s *ContactService) CreateContact(ctx context.Context, contact Contact) (*Contact, error) {
    // Validation
    if err := s.validateContact(contact); err != nil {
        return nil, fmt.Errorf("invalid contact: %w", err)
    }

    // Permission check
    if err := s.authService.CheckPermission(ctx, "contacts:create"); err != nil {
        return nil, fmt.Errorf("permission denied: %w", err)
    }

    // Business logic
    contact.CreatedAt = time.Now()
    contact.OrganizationID = s.authService.GetOrganizationID(ctx)

    // Data access
    created, err := s.repo.Create(ctx, contact)
    if err != nil {
        return nil, fmt.Errorf("failed to create contact: %w", err)
    }

    return created, nil
}
```

### Repository Layer Pattern

```go
type ContactRepository interface {
    Create(ctx context.Context, contact Contact) (*Contact, error)
    FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)
    FindAll(ctx context.Context, filters ContactFilter) ([]Contact, error)
    Update(ctx context.Context, contact Contact) (*Contact, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

type contactRepository struct {
    db *sql.DB
}

func (r *contactRepository) Create(ctx context.Context, contact Contact) (*Contact, error) {
    query := `
        INSERT INTO contacts
        (id, organization_id, name, email, phone, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, organization_id, name, email, phone, created_at, updated_at
    `

    var created Contact
    err := r.db.QueryRowContext(ctx, query,
        contact.ID, contact.OrganizationID, contact.Name,
        contact.Email, contact.Phone, contact.CreatedAt, contact.UpdatedAt,
    ).Scan(
        &created.ID, &created.OrganizationID, &created.Name,
        &created.Email, &created.Phone, &created.CreatedAt, &created.UpdatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to create contact: %w", err)
    }

    return &created, nil
}
```

## Implementation Timeline

### Week 1-2: Foundation Services
- **Day 1-3**: Authentication system (backend + basic frontend)
- **Day 4-5**: Authorization middleware and permission system
- **Day 6-7**: Core service layer infrastructure
- **Day 8-10**: Testing and integration of foundation services

### Week 3-4: CRM Module
- **Day 11-12**: CRM backend services and repositories
- **Day 13-14**: CRM API endpoints with validation
- **Day 15-17**: CRM frontend components
- **Day 18-20**: CRM testing and bug fixing

### Week 5-6: Products & Inventory Module
- **Day 21-22**: Products backend services
- **Day 23-24**: Inventory backend services
- **Day 25-26**: Module API endpoints
- **Day 27-29**: Frontend implementation
- **Day 30**: Integration testing

### Week 7-8: Sales Module
- **Day 31-33**: Sales order services and workflow
- **Day 34-35**: Pricelist management
- **Day 36-37**: API endpoints implementation
- **Day 38-40**: Frontend components
- **Day 41-42**: Integration with CRM and Inventory

### Week 9-10: Accounting Module
- **Day 43-45**: Chart of accounts and basic accounting services
- **Day 46-48**: Invoice and payment services
- **Day 49-50**: API endpoints
- **Day 51-53**: Frontend implementation
- **Day 54-56**: Integration with Sales module

### Week 11-12: Basic Analytics & Polish
- **Day 57-60**: Dashboard widgets and basic reports
- **Day 61-63**: UI/UX improvements
- **Day 64-66**: Performance optimization
- **Day 67-70**: Comprehensive testing and bug fixing

## Success Criteria

### Minimum Viable Product (MVP) Definition
A basic usable ERP system should include:

1. **User Authentication**: Secure login/logout functionality
2. **CRM**: Contact and lead management with basic workflow
3. **Products**: Product catalog with inventory tracking
4. **Sales**: Order creation and management
5. **Accounting**: Basic invoicing and payment processing
6. **Integration**: Workflow between CRM → Sales → Accounting
7. **UI**: Functional (if not beautiful) interface for all core functions

### Quality Standards
- All API endpoints properly authenticated and authorized
- Comprehensive input validation
- Proper error handling and logging
- Basic unit and integration tests
- Responsive design for core workflows
- Reasonable performance for typical use cases

## Next Steps

This roadmap provides a comprehensive plan for implementing the core ERP functionality. The implementation should proceed in the priority order outlined, with each phase building upon the foundation of the previous ones.
