# API-First ERP Implementation Roadmap

## Current Implementation Status

### ✅ COMPLETED (Phase 1 & 2 from Roadmap)
- **Authentication System**: Full JWT implementation, user registration/login
- **Authorization Framework**: RBAC structure, Casbin integration (enforcement needs work)
- **CRM Module API**: Complete CRUD for contacts, partial leads
- **Products Module API**: Complete CRUD for products, categories
- **Sales Module API**: Complete CRUD for orders with workflow, pricelists
- **Accounting Module API**: Complete CRUD for invoices with workflow, payments
- **Event System**: Full pub/sub architecture operational
- **Module Registry**: Dependency injection and module lifecycle

### ⚠️ PARTIAL IMPLEMENTATION
- **Event Handlers**: All registered but contain TODO placeholders (13 handlers need business logic)
- **Permission Enforcement**: Casbin framework exists but currently allows all requests with warning
- **Tax Calculation**: Tables exist, no calculation logic implemented
- **Lead Management**: Repository exists, handler incomplete

### ❌ MISSING API COVERAGE
**60+ migration files exist with 145+ database tables, only 10 tables have API endpoints**

Critical gaps for MVP:
1. **Inventory Management** (stock_locations, stock_moves, stock_quants, warehouses)
2. **Purchase Orders** (purchase_orders, purchase_order_lines)
3. **Chart of Accounts** (account_accounts, account_journals)
4. **HR/Employees** (employees, departments, job_positions)
5. **Projects & Tasks** (projects, tasks, task_stages)
6. **Reference Data** (countries, states, currencies, UOM units)

## API-First Development Roadmap (Frontend Discarded)

### PRIORITY 1: Complete Event-Driven Architecture (Week 1)
**Goal**: Make existing modules fully functional with business logic

#### 1.1 Implement Event Handler Business Logic
**Files to modify**:
- `/internal/modules/crm/crm_event_handlers.go` (3 handlers)
- `/internal/modules/sales/sales_event_handlers.go` (3 handlers)
- `/internal/modules/accounting/accounting_event_handlers.go` (2 handlers)

**Tasks**:
1. **CRM Handlers**:
   - `order.created` → Create activity record linking contact to order
   - `order.confirmed` → Update contact's `is_customer` flag to true
   - `invoice.created` → Create activity record for invoice

2. **Sales Handlers**:
   - `contact.created` → Sync customer data to sales order customer cache
   - `contact.updated` → Update existing orders with new customer info
   - `invoice.paid` → Mark related orders as fully paid, update payment status

3. **Accounting Handlers**:
   - `order.confirmed` → Auto-generate invoice from confirmed order
   - `contact.updated` → Sync partner information to invoice records

**Database additions needed**:
- `activities` table usage for CRM event tracking
- Order-to-invoice linking field in `invoices` table

---

### PRIORITY 2: Enforce Permission Policies (Week 1-2)
**Goal**: Replace "allow all" fallback with real Casbin policy enforcement

#### 2.1 Complete Authorization System
**Files to modify**:
- `/pkg/policy/manager.go` - Remove mock fallback, implement real policy loading
- `/internal/modules/auth/auth_middleware.go` - Add actual permission checks per endpoint
- `/internal/migrations/20250115000002_seed_initial_policies.sql` - Verify/enhance policy seeds

**Tasks**:
1. Load Casbin policies from database instead of using mock allow-all
2. Implement per-endpoint permission checks:
   - `crm:contacts:read`, `crm:contacts:create`, `crm:contacts:update`, `crm:contacts:delete`
   - `products:read`, `products:create`, `products:update`, `products:delete`
   - `sales:orders:read`, `sales:orders:create`, `sales:orders:confirm`, etc.
   - `accounting:invoices:read`, `accounting:invoices:create`, `accounting:invoices:confirm`, etc.
3. Implement organization-scoped data access (ensure users only see their org's data)
4. Add permission audit logging

---

### PRIORITY 3: Inventory Management API (Week 2-3)
**Goal**: Expose inventory and stock management tables

#### 3.1 Create Inventory Module
**New files to create**:
- `/internal/modules/inventory/inventory_service.go`
- `/internal/modules/inventory/inventory_repository.go`
- `/internal/modules/inventory/inventory_handler.go`
- `/internal/modules/inventory/inventory_module.go`
- `/internal/modules/inventory/inventory_event_handlers.go`

**Database tables to expose**:
- `stock_locations` - Warehouse locations
- `stock_quants` - Current stock levels per location/product
- `stock_moves` - Stock movement history
- `stock_pickings` - Picking operations
- `warehouses` - Warehouse management

**API Endpoints to implement**:
```
GET    /api/inventory/locations                  - List locations
POST   /api/inventory/locations                  - Create location
GET    /api/inventory/locations/:id              - Get location
PUT    /api/inventory/locations/:id              - Update location

GET    /api/inventory/stock-levels               - Get stock levels (filterable)
GET    /api/inventory/stock-levels/:product_id   - Get stock for product

POST   /api/inventory/stock-moves                - Create stock movement
GET    /api/inventory/stock-moves                - List movements
GET    /api/inventory/stock-moves/:id            - Get movement

GET    /api/inventory/warehouses                 - List warehouses
POST   /api/inventory/warehouses                 - Create warehouse
```

**Events to publish**:
- `stock.moved`, `stock.adjusted`, `warehouse.created`

**Integration**:
- Sales order confirmation should reserve inventory
- Product queries should include stock availability
- Link to delivery/shipping workflows

---

### PRIORITY 4: Purchase Orders API (Week 3)
**Goal**: Enable procurement workflow

#### 4.1 Create Purchase Module
**New files to create**:
- `/internal/modules/purchase/purchase_service.go`
- `/internal/modules/purchase/purchase_repository.go`
- `/internal/modules/purchase/purchase_handler.go`
- `/internal/modules/purchase/purchase_module.go`

**Database tables to expose**:
- `purchase_orders` - PO management
- `purchase_order_lines` - PO line items

**API Endpoints**:
```
POST   /api/purchase/orders                      - Create PO
GET    /api/purchase/orders                      - List POs
GET    /api/purchase/orders/:id                  - Get PO
PUT    /api/purchase/orders/:id                  - Update PO
DELETE /api/purchase/orders/:id                  - Cancel PO
POST   /api/purchase/orders/:id/confirm          - Confirm PO
POST   /api/purchase/orders/:id/receive          - Mark as received
GET    /api/purchase/orders/supplier/:id         - Get supplier POs
```

**Workflow states**: `draft` → `sent` → `purchase` → `done`

**Events**: `po.created`, `po.confirmed`, `po.received`

**Integration**:
- Receipt creates stock moves
- Links to supplier contacts
- Auto-generate bills/invoices

---

### PRIORITY 5: Accounting - Chart of Accounts API (Week 4)
**Goal**: Enable financial structure configuration

#### 4.2 Extend Accounting Module
**Files to modify**:
- `/internal/modules/accounting/accounting_service.go`
- `/internal/modules/accounting/accounting_repository.go`
- `/internal/modules/accounting/accounting_handler.go`

**Database tables to expose**:
- `account_accounts` - Chart of accounts
- `account_journals` - Accounting journals
- `account_taxes` - Tax configuration

**New API Endpoints**:
```
GET    /api/accounting/accounts                  - List accounts
POST   /api/accounting/accounts                  - Create account
GET    /api/accounting/accounts/:id              - Get account
PUT    /api/accounting/accounts/:id              - Update account

GET    /api/accounting/journals                  - List journals
POST   /api/accounting/journals                  - Create journal
GET    /api/accounting/journals/:id              - Get journal

GET    /api/accounting/taxes                     - List tax rates
POST   /api/accounting/taxes                     - Create tax
PUT    /api/accounting/taxes/:id                 - Update tax

GET    /api/accounting/balances                  - Trial balance report
GET    /api/accounting/balances/:account_id      - Account balance
```

**Tax Calculation Enhancement**:
- Implement tax calculation in invoice service
- Support multiple taxes per line item
- Tax totals on orders and invoices

---

### PRIORITY 6: Reference Data APIs (Week 4)
**Goal**: Expose configuration and reference tables

#### 6.1 Create Reference Module
**New files**:
- `/internal/modules/reference/reference_service.go`
- `/internal/modules/reference/reference_repository.go`
- `/internal/modules/reference/reference_handler.go`

**Tables to expose**:
- `countries`, `states` - Geographic data
- `currencies` - Currency management
- `uom_units`, `uom_categories` - Units of measure
- `payment_terms`, `payment_methods` - Payment configuration
- `fiscal_positions` - Tax positions

**API Endpoints**:
```
GET    /api/reference/countries
GET    /api/reference/states/:country_code
GET    /api/reference/currencies
GET    /api/reference/uom-units
GET    /api/reference/payment-terms
GET    /api/reference/payment-methods
```

Most endpoints are read-only reference data with occasional admin updates.

---

### PRIORITY 7: CRM - Complete Lead Management (Week 5)
**Goal**: Finish lead-to-opportunity workflow

#### 7.1 Complete CRM Module
**Files to modify**:
- `/internal/modules/crm/crm_service.go` - Add lead methods
- `/internal/modules/crm/crm_repository.go` - Complete lead repo
- `/internal/modules/crm/crm_handler.go` - Add lead endpoints

**Database tables**:
- `leads` (already exists)
- `lead_stages`, `lead_sources`, `lost_reasons`

**New API Endpoints**:
```
GET    /api/crm/leads                            - List leads
POST   /api/crm/leads                            - Create lead
GET    /api/crm/leads/:id                        - Get lead
PUT    /api/crm/leads/:id                        - Update lead
DELETE /api/crm/leads/:id                        - Delete lead
POST   /api/crm/leads/:id/convert                - Convert to opportunity/contact
POST   /api/crm/leads/:id/mark-lost              - Mark as lost

GET    /api/crm/lead-stages                      - List stages
GET    /api/crm/lead-sources                     - List sources
```

**Workflow**: `new` → `qualified` → `proposition` → `won`/`lost`

**Events**: `lead.created`, `lead.converted`, `lead.lost`

---

### PRIORITY 8: HR/Employee Module API (Week 5-6)
**Goal**: Basic employee and department management

#### 8.1 Create HR Module
**New files**:
- `/internal/modules/hr/hr_service.go`
- `/internal/modules/hr/hr_repository.go`
- `/internal/modules/hr/hr_handler.go`
- `/internal/modules/hr/hr_module.go`

**Tables**:
- `employees` - Employee records
- `departments` - Organizational departments
- `job_positions` - Job titles/positions

**API Endpoints**:
```
GET    /api/hr/employees
POST   /api/hr/employees
GET    /api/hr/employees/:id
PUT    /api/hr/employees/:id
DELETE /api/hr/employees/:id

GET    /api/hr/departments
POST   /api/hr/departments
GET    /api/hr/departments/:id

GET    /api/hr/positions
POST   /api/hr/positions
```

---

### PRIORITY 9: Project Management API (Week 6-7)
**Goal**: Basic project and task tracking

#### 9.1 Create Projects Module
**New files**:
- `/internal/modules/projects/projects_service.go`
- `/internal/modules/projects/projects_repository.go`
- `/internal/modules/projects/projects_handler.go`
- `/internal/modules/projects/projects_module.go`

**Tables**:
- `projects` - Project records
- `tasks` - Task management
- `task_stages` - Task workflow stages

**API Endpoints**:
```
GET    /api/projects
POST   /api/projects
GET    /api/projects/:id
PUT    /api/projects/:id

GET    /api/projects/:id/tasks
POST   /api/projects/:id/tasks
GET    /api/tasks/:id
PUT    /api/tasks/:id
POST   /api/tasks/:id/move                       - Change stage
```

---

### PRIORITY 10: Analytics & Reporting APIs (Week 7-8)
**Goal**: Business intelligence endpoints

#### 10.1 Create Analytics Module
**New files**:
- `/internal/modules/analytics/analytics_service.go`
- `/internal/modules/analytics/analytics_handler.go`

**Leverage existing views** (migrations 44-47):
- Financial intelligence views
- CRM analytics views
- Operations excellence views
- Insight analytics views

**API Endpoints**:
```
GET    /api/analytics/sales/overview             - Sales KPIs
GET    /api/analytics/sales/by-period            - Sales by time period
GET    /api/analytics/crm/pipeline               - Lead/opportunity pipeline
GET    /api/analytics/inventory/valuation        - Inventory value
GET    /api/analytics/accounting/receivables     - Aging receivables
GET    /api/analytics/accounting/payables        - Aging payables
GET    /api/analytics/financial/snapshot         - Financial overview
```

---

## Implementation Timeline Summary

| Week | Priority | Module | Status |
|------|----------|--------|--------|
| 1 | P1 | Event Handler Logic | Complete existing TODOs |
| 1-2 | P2 | Permission Enforcement | Remove mock, enforce policies |
| 2-3 | P3 | Inventory Management | New module, 8+ endpoints |
| 3 | P4 | Purchase Orders | New module, 7+ endpoints |
| 4 | P5 | Chart of Accounts | Extend accounting, 10+ endpoints |
| 4 | P6 | Reference Data | New module, 6+ endpoints |
| 5 | P7 | Lead Management | Complete CRM, 7+ endpoints |
| 5-6 | P8 | HR/Employees | New module, 8+ endpoints |
| 6-7 | P9 | Project Management | New module, 8+ endpoints |
| 7-8 | P10 | Analytics/Reporting | New module, 6+ endpoints |

**Total**: ~8 weeks to API-complete core ERP functionality

---

## Success Criteria for API-Complete ERP

### Minimum Viable API Coverage
- ✅ Authentication/Authorization with real enforcement
- ✅ CRM: Contacts + Leads with full workflow
- ✅ Products: Catalog with inventory tracking
- ✅ Sales: Orders with pricelist and workflow
- ✅ Accounting: Invoices, payments, chart of accounts, taxes
- ✅ Inventory: Stock locations, movements, warehouses
- ✅ Purchasing: Purchase orders with workflow
- ✅ HR: Employees, departments, positions
- ✅ Projects: Project and task management
- ✅ Analytics: Core business intelligence reports
- ✅ Reference Data: Countries, currencies, UOM, payment terms

### Event-Driven Integration
- All modules publish domain events
- Cross-module workflows via event handlers:
  - Lead → Contact → Order → Invoice → Payment
  - Order → Inventory Reservation → Picking → Delivery
  - Purchase Order → Receipt → Stock Move → Bill

### Quality Standards
- All endpoints authenticated and authorized
- Comprehensive input validation
- Proper error handling with standard responses
- Organization-scoped data access
- Audit logging for critical operations
- Event publishing for all state changes

---

## Module Architecture Pattern

Every new module follows this structure:

```
/internal/modules/{module}/
├── {module}_module.go          - Module registration and lifecycle
├── {module}_service.go         - Business logic layer
├── {module}_repository.go      - Data access layer (pgx)
├── {module}_handler.go         - HTTP handlers and routing
├── {module}_event_handlers.go  - Event subscriptions
└── {module}_models.go          - Domain models/DTOs
```

**Service Layer Responsibilities**:
- Validation of business rules
- Permission checking via auth context
- Event publishing on state changes
- Orchestration of repository calls
- Error handling and logging

**Repository Layer**:
- Use pgx for database access (not GORM)
- Organization-scoped queries
- Transaction management
- Prepared statement usage

**Handler Layer**:
- Request parsing and validation
- Auth middleware application
- Response serialization
- HTTP status codes

---

## Critical Files Reference

### Core Infrastructure
- `/cmd/server/main.go` - Server initialization
- `/internal/server/server.go` - HTTP server setup
- `/pkg/registry/registry.go` - Module registry
- `/pkg/events/bus.go` - Event bus
- `/pkg/policy/manager.go` - Casbin policy engine

### Existing Modules
- `/internal/modules/auth/` - Authentication module
- `/internal/modules/crm/` - CRM module
- `/internal/modules/products/` - Products module
- `/internal/modules/sales/` - Sales module
- `/internal/modules/accounting/` - Accounting module

### Database
- `/internal/migrations/*.sql` - 60+ migration files
- Key migrations:
  - `20250101000001_foundation_tables.sql`
  - `20250101000003_crm_module.sql`
  - `20250101000004_products_inventory_module.sql`
  - `20250101000005_sales_module.sql`
  - `20250101000006_accounting_module.sql`
  - `20250101000007_remaining_modules.sql` (HR, Projects, Purchase)

---

## Next Immediate Actions

1. ✅ Confirm with user: Is this API-first roadmap aligned with expectations?
2. Start with Priority 1: Event handler implementation (quick win, completes existing work)
3. Move to Priority 2: Real permission enforcement (critical for production)
4. Then systematically add missing modules per priority order
