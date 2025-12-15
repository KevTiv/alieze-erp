# Action Plan: Complete Event System & Architecture Implementation

## Current Status: Event System Foundation Complete ✅

The event publishing and handling infrastructure is now operational. This document outlines the action plan to complete all remaining TODOs and implement the full event-driven architecture.

---

## Phase 1: Complete Event Handler Implementations (Week 1)

**Priority:** HIGH
**Effort:** Medium
**Dependencies:** None (can start immediately)

### 1.1 CRM Module Event Handlers

**File:** `internal/modules/crm/module.go`

#### TODO: Update contact last activity, sales stats
```go
func (m *CRMModule) handleOrderCreated(ctx context.Context, event interface{}) error {
    // Extract order data from event
    // Update contact's last_activity_date
    // Increment contact's order_count
    // Update contact's total_revenue
}
```

**Action Items:**
- [ ] Add fields to Contact type: `last_activity_date`, `order_count`, `total_revenue`
- [ ] Create migration to add these fields to contacts table
- [ ] Implement logic to extract order data from event payload
- [ ] Update contact record via contact repository
- [ ] Add error handling and logging

#### TODO: Update contact as active customer
```go
func (m *CRMModule) handleOrderConfirmed(ctx context.Context, event interface{}) error {
    // Extract customer_id from order
    // Update contact.is_customer = true
    // Update contact.customer_since_date if first order
}
```

**Action Items:**
- [ ] Add field to Contact type: `customer_since_date`
- [ ] Create migration to add field to contacts table
- [ ] Implement customer activation logic
- [ ] Handle case where contact is already a customer
- [ ] Add logging for audit trail

#### TODO: Track customer invoicing activity
```go
func (m *CRMModule) handleInvoiceCreated(ctx context.Context, event interface{}) error {
    // Extract partner_id from invoice
    // Update contact's invoice_count
    // Update contact's total_invoiced_amount
    // Update contact's last_invoice_date
}
```

**Action Items:**
- [ ] Add fields to Contact type: `invoice_count`, `total_invoiced_amount`, `last_invoice_date`
- [ ] Create migration
- [ ] Implement invoice tracking logic
- [ ] Consider creating a separate contact_stats table for better performance

---

### 1.2 Accounting Module Event Handlers

**File:** `internal/modules/accounting/module.go`

#### TODO: Implement invoice generation logic when order is confirmed
```go
func (m *AccountingModule) handleOrderConfirmed(ctx context.Context, event interface{}) error {
    // Extract order details from event
    // Create invoice from order data
    // Copy order lines to invoice lines
    // Link invoice to order
    // Publish invoice.created event
}
```

**Action Items:**
- [ ] Design order-to-invoice mapping logic
- [ ] Add `order_id` field to invoice type (if not exists)
- [ ] Create migration to link invoices to orders
- [ ] Implement auto-invoice creation service method
- [ ] Add configuration option to enable/disable auto-invoicing
- [ ] Handle currency, pricelist, and tax calculations
- [ ] Add comprehensive error handling

#### TODO: Implement partner sync logic
```go
func (m *AccountingModule) handleContactUpdated(ctx context.Context, event interface{}) error {
    // Extract contact data from event
    // Find invoices/payments for this partner
    // Update partner name/address on invoices (optional)
    // Or just log the sync for audit purposes
}
```

**Action Items:**
- [ ] Decide on sync strategy (update vs. immutable invoices)
- [ ] Implement contact data extraction from event
- [ ] Consider creating a partner cache/snapshot system
- [ ] Add logging for compliance/audit

---

### 1.3 Sales Module Event Handlers

**File:** `internal/modules/sales/module.go`

#### TODO: Implement customer sync logic
```go
func (m *SalesModule) handleContactCreated(ctx context.Context, event interface{}) error {
    // Extract contact data
    // Check if contact.is_customer = true
    // Potentially pre-create customer-specific records
    // Or just cache customer data for order creation
}
```

**Action Items:**
- [ ] Define customer sync requirements
- [ ] Implement contact data extraction
- [ ] Consider creating customer cache or using contact ID directly
- [ ] Add validation and error handling

#### TODO: Implement customer update sync logic
```go
func (m *SalesModule) handleContactUpdated(ctx context.Context, event interface{}) error {
    // Extract updated contact data
    // Invalidate any cached customer data
    // Update customer references in draft orders (if needed)
}
```

**Action Items:**
- [ ] Implement cache invalidation if using customer cache
- [ ] Decide on update strategy for existing orders
- [ ] Add logging

#### TODO: Mark related orders as fully invoiced/paid
```go
func (m *SalesModule) handleInvoicePaid(ctx context.Context, event interface{}) error {
    // Extract invoice data
    // Find related sales order via order_id
    // Update order.invoicing_status = "invoiced"
    // Update order.payment_status = "paid"
    // Consider triggering order.completed event
}
```

**Action Items:**
- [ ] Add fields to SalesOrder: `invoicing_status`, `payment_status`
- [ ] Create migration
- [ ] Implement order status update logic
- [ ] Add new event: `order.completed` or `order.fully_paid`
- [ ] Consider workflow state transition

---

## Phase 2: Auth & Permission System Integration (Week 2)

**Priority:** HIGH
**Effort:** High
**Dependencies:** None

### 2.1 Context-Based Organization & User Retrieval

**File:** `internal/modules/crm/module.go` (PolicyAuthServiceAdapter)

#### TODO: Implement organization ID retrieval from context
```go
func (a *PolicyAuthServiceAdapter) GetOrganizationID(ctx context.Context) (uuid.UUID, error) {
    // Extract from JWT claims in context
    // Middleware should have already set this
}
```

**Action Items:**
- [ ] Review auth middleware to ensure it sets organization_id in context
- [ ] Define context key constants (e.g., `ContextKeyOrganizationID`)
- [ ] Implement extraction logic
- [ ] Add fallback/error handling for missing context
- [ ] Update all services to use context-based org retrieval

#### TODO: Implement user ID retrieval from context
```go
func (a *PolicyAuthServiceAdapter) GetUserID(ctx context.Context) (uuid.UUID, error) {
    // Extract from JWT claims in context
}
```

**Action Items:**
- [ ] Define context key for user_id
- [ ] Implement extraction from auth context
- [ ] Add error handling
- [ ] Update services to use real user IDs

### 2.2 Real Policy Engine Implementation

**Current:** Allows all permissions with warning fallback
**Goal:** Enforce actual Casbin policies

**Action Items:**
- [ ] Create policies database table
- [ ] Seed default role-based policies (admin, accountant, sales, etc.)
- [ ] Load policies from database on startup
- [ ] Implement actual CheckPermission enforcement
- [ ] Remove "allow all" fallback
- [ ] Add policy management API (CRUD operations for policies)
- [ ] Create admin UI for policy management (future)

---

## Phase 3: Tax Calculation System (Week 2-3)

**Priority:** MEDIUM
**Effort:** Medium
**Dependencies:** None

### 3.1 Invoice Tax Calculation

**File:** `internal/modules/accounting/service/invoice_service.go`

#### TODO: Implement proper tax calculation based on tax rates
```go
// Current placeholder
line.PriceTax = 0
if line.TaxID != nil {
    // TODO: Implement proper tax calculation based on tax rates
}
```

**Action Items:**
- [ ] Create tax_rates table (id, name, rate, type, country_id, etc.)
- [ ] Create tax repository and service
- [ ] Implement tax lookup by TaxID
- [ ] Calculate tax amount based on rate and tax type (inclusive/exclusive)
- [ ] Support multiple taxes per line (tax groups)
- [ ] Handle tax rounding
- [ ] Add tax configuration (rounding rules, compound taxes)

### 3.2 Sales Order Tax Calculation

**File:** `internal/modules/sales/service/sales_order_service.go`

**Action Items:**
- [ ] Same as invoice tax calculation
- [ ] Ensure consistency between orders and invoices
- [ ] Consider creating shared tax calculation utility

---

## Phase 4: Complete Remaining Module Event Handlers (Week 3)

### 4.1 Auth Module Event Handlers

**File:** `internal/modules/auth/module.go`

#### TODO: Implement event handlers when event system is integrated

**Potential Events to Handle:**
- `organization.created` → Create default users/roles
- `user.created` → Send welcome email, setup defaults
- `user.login` → Track login activity
- `permission.denied` → Log security events

**Action Items:**
- [ ] Define which events auth module should handle
- [ ] Implement event handlers
- [ ] Consider security event logging

### 4.2 Products Module Event Handlers

**File:** `internal/modules/products/module.go`

#### TODO: Implement event handlers when event system is integrated

**Potential Events to Handle:**
- `order.confirmed` → Decrease stock levels
- `order.cancelled` → Restore stock levels
- `invoice.confirmed` → Trigger stock movement

**Action Items:**
- [ ] Define inventory/stock management requirements
- [ ] Implement stock adjustment event handlers
- [ ] Add product stock fields if not exists
- [ ] Consider creating inventory module

---

## Phase 5: Rule Engine Migration (Week 4)

**Goal:** Replace hardcoded validation with YAML-based rules

### 5.1 Create Rule Configuration Files

**Files to Create:**
- `config/rules/crm.yaml`
- `config/rules/accounting.yaml`
- `config/rules/sales.yaml`
- `config/rules/products.yaml`

**Example Structure:**
```yaml
modules:
  crm:
    validation:
      contact_create:
        rules:
          - name: "require_name"
            validator: "require_string"
            field: "name"
          - name: "validate_email"
            validator: "email_format"
            field: "email"
            required: false
          - name: "validate_phone"
            validator: "phone_format"
            field: "phone"
            required: false
```

**Action Items:**
- [ ] Audit all hardcoded validation in services
- [ ] Create YAML configs for each module
- [ ] Implement missing validators in `pkg/rules/validators/`
- [ ] Test rule engine with configs
- [ ] Remove hardcoded validation fallbacks
- [ ] Update services to rely solely on rule engine

---

## Phase 6: Workflow/State Machine Configuration (Week 5)

**Goal:** Externalize state transitions to YAML configs

### 6.1 Create Workflow Configuration Files

**Files to Create:**
- `config/workflows/invoice_states.yaml`
- `config/workflows/order_states.yaml`
- `config/workflows/payment_states.yaml`

**Example Structure:**
```yaml
workflow_id: accounting.invoice
model: accounting.invoice
initial: draft

states:
  - draft
  - open
  - paid
  - cancelled

transitions:
  - name: confirm
    from: draft
    to: open
    validator: invoice_confirmable
    permission: invoices:confirm

  - name: pay
    from: open
    to: paid
    validator: invoice_payable
    permission: invoices:pay

  - name: cancel
    from: [draft, open]
    to: cancelled
    permission: invoices:cancel
```

**Action Items:**
- [ ] Create workflow YAML files
- [ ] Implement transition validators
- [ ] Update services to use state machine for all transitions
- [ ] Remove hardcoded state transition logic
- [ ] Add state machine unit tests

---

## Phase 7: Database Schema Updates (Week 6)

**Goal:** Add fields needed for full event system functionality

### 7.1 Contact Enhancements
```sql
ALTER TABLE contacts ADD COLUMN last_activity_date TIMESTAMPTZ;
ALTER TABLE contacts ADD COLUMN customer_since_date TIMESTAMPTZ;
ALTER TABLE contacts ADD COLUMN order_count INTEGER DEFAULT 0;
ALTER TABLE contacts ADD COLUMN invoice_count INTEGER DEFAULT 0;
ALTER TABLE contacts ADD COLUMN total_revenue DECIMAL(15,2) DEFAULT 0;
ALTER TABLE contacts ADD COLUMN total_invoiced_amount DECIMAL(15,2) DEFAULT 0;
ALTER TABLE contacts ADD COLUMN last_invoice_date TIMESTAMPTZ;
```

### 7.2 Order Enhancements
```sql
ALTER TABLE sales_orders ADD COLUMN invoicing_status VARCHAR(50) DEFAULT 'not_invoiced';
ALTER TABLE sales_orders ADD COLUMN payment_status VARCHAR(50) DEFAULT 'not_paid';
```

### 7.3 Invoice Enhancements
```sql
ALTER TABLE invoices ADD COLUMN order_id UUID REFERENCES sales_orders(id);
CREATE INDEX idx_invoices_order_id ON invoices(order_id);
```

### 7.4 Tax System
```sql
CREATE TABLE tax_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    rate DECIMAL(5,2) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'exclusive', 'inclusive'
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 7.5 Policies System
```sql
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id),
    subject VARCHAR(255) NOT NULL,
    object VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    effect VARCHAR(10) DEFAULT 'allow', -- 'allow' or 'deny'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_policies_subject ON policies(subject);
CREATE INDEX idx_policies_org ON policies(organization_id);
```

**Action Items:**
- [ ] Create migration files for each change
- [ ] Test migrations on development database
- [ ] Update domain types to include new fields
- [ ] Update repositories to handle new fields

---

## Summary Checklist

### Immediate Actions (This Week)
- [ ] Implement all CRM event handlers (3 handlers)
- [ ] Implement all Accounting event handlers (2 handlers)
- [ ] Implement all Sales event handlers (3 handlers)
- [ ] Fix context-based auth (GetOrganizationID, GetUserID)

### Short Term (Weeks 2-3)
- [ ] Implement real policy enforcement with Casbin
- [ ] Create policies database and seed data
- [ ] Implement tax calculation system
- [ ] Create tax_rates table and tax service

### Medium Term (Weeks 4-5)
- [ ] Migrate all validation to rule engine
- [ ] Create YAML rule configs for all modules
- [ ] Create workflow YAML configs
- [ ] Wire up state machines in all services

### Long Term (Week 6+)
- [ ] Complete all database migrations
- [ ] Remove all hardcoded validation
- [ ] Remove all state machine fallbacks
- [ ] Create admin UI for policies/rules/workflows
- [ ] Add event persistence for audit trail
- [ ] Implement async event processing

---

## Success Metrics

**Event System Maturity:**
- [ ] All business operations publish events
- [ ] All modules have functional event handlers
- [ ] Cross-module communication works without direct dependencies
- [ ] Events include complete business context

**Validation Maturity:**
- [ ] 0% hardcoded validation (all in YAML)
- [ ] All validators implemented and tested
- [ ] Easy to customize validation per tenant

**Workflow Maturity:**
- [ ] All state transitions in YAML
- [ ] No hardcoded status checks
- [ ] Workflow can be changed without code deploy

**Security Maturity:**
- [ ] Real policy enforcement (no fallbacks)
- [ ] Role-based access control functional
- [ ] Per-organization policy customization
- [ ] Audit trail of permission checks

---

## Estimated Timeline

- **Week 1:** Event handler implementations (HIGH PRIORITY)
- **Week 2:** Auth/permissions + tax system
- **Week 3:** Tax completion + products events
- **Week 4:** Rule engine migration
- **Week 5:** Workflow configs
- **Week 6:** Database migrations + cleanup
- **Week 7+:** Testing, documentation, polish

**Total Estimated Effort:** 6-8 weeks for complete implementation

---

## Risk Mitigation

**Risk 1: Breaking changes during migration**
- Mitigation: Keep fallbacks until new system is tested
- Strategy: Gradual migration module-by-module

**Risk 2: Performance impact of event system**
- Mitigation: Use synchronous events initially
- Strategy: Profile and optimize, switch to async if needed

**Risk 3: Complex tax calculation requirements**
- Mitigation: Start with simple tax model
- Strategy: Iterate based on actual business needs

**Risk 4: Policy misconfiguration**
- Mitigation: Provide sane defaults
- Strategy: Add policy testing tools and admin UI

---

**Document Created:** 2025-12-15
**Status:** Ready for implementation
**Next Action:** Begin Phase 1 - Event Handler Implementation
