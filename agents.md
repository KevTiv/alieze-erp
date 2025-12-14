# Agent Request Handling Guide

## Migration Files Quick View

Here's a comprehensive overview of the migration files and the tables they contain:

### Foundation Tables (20250101000001)
- `organizations` - Multi-tenant organizations
- `organization_users` - User organization memberships
- `companies` - Business entities
- `sequences` - Number sequence generation

### Reference Data (20250101000002)
- `countries`, `states` - Geographic data
- `currencies` - Currency information
- `uom_categories`, `uom_units` - Units of measure
- `payment_terms`, `fiscal_positions` - Financial terms
- `analytic_accounts` - Analytical accounting
- `industries`, `utm_campaigns` - Marketing data

### CRM Module (20250101000003)
- `contact_tags`, `sales_teams` - CRM organization
- `contacts` - Customer contacts
- `lead_stages`, `lead_sources`, `lost_reasons` - Lead management
- `leads` - Sales leads
- `activities` - CRM activities

### Products & Inventory (20250101000004)
- `product_categories`, `products`, `product_variants` - Product catalog
- `warehouses`, `stock_locations` - Inventory locations
- `stock_packages`, `stock_lots`, `stock_quants` - Inventory tracking
- `procurement_groups`, `stock_rules` - Procurement logic

### Sales Module (20250101000005)
- `pricelists` - Product pricing
- `sales_orders`, `sales_order_lines` - Sales orders

### Accounting Module (20250101000006)
- `account_account_types`, `account_groups`, `account_accounts` - Chart of accounts
- `account_journals` - Accounting journals
- `account_tax_groups`, `account_taxes` - Tax management
- `invoices`, `invoice_lines` - Invoicing
- `payments` - Payment processing
- `account_full_reconcile` - Account reconciliation

### Remaining Modules (20250101000007)
- `purchase_orders`, `purchase_order_lines` - Purchasing
- `workcenters`, `bom_bills`, `bom_lines` - Manufacturing BOMs
- `manufacturing_orders`, `work_orders` - Production
- `resources`, `departments`, `job_positions` - HR structure

### Knowledge Base (20250101000015)
- `knowledge_spaces`, `knowledge_entries` - Knowledge articles
- `knowledge_entry_revisions` - Article versions
- `knowledge_tags`, `knowledge_entry_tags` - Tagging system
- `knowledge_context_links` - Contextual links
- `knowledge_entry_assets` - Attachments

### AI & Advanced Features
- `ai_team_members`, `ai_agent_tasks`, `ai_agent_insights` - AI agents
- `ai_user_preferences` - User AI settings
- `business_insights_cache`, `ai_insight_history` - Business intelligence
- `data_import_sessions`, `data_import_row_results` - Data import system
- `ai_model_config`, `ai_provider_routing_rules` - AI configuration

### Delivery Tracking (20250101000048)
- `delivery_vehicles`, `delivery_routes` - Delivery infrastructure
- `delivery_route_assignments`, `delivery_shipments` - Shipments
- `delivery_route_stops`, `delivery_tracking_events` - Tracking
- `delivery_route_positions` - Route positions

### Point of Sale (20250101000049)
- `pos_payment_methods`, `pos_config` - POS configuration
- `pos_sessions`, `pos_cash_movements` - Cash management
- `pos_payments`, `pos_order_discounts` - Transactions
- `pos_inventory_alerts`, `pos_pricing_overrides` - Inventory control

### Permission System (20250101000037-42)
- `permission_roles`, `permission_role_inheritance` - Role management
- `permission_role_templates`, `user_role_assignments` - Role assignments
- `permission_groups`, `role_permission_groups` - Permission groups
- `permission_table_policies`, `permission_column_policies` - Data access control
- `permission_row_filters`, `ai_data_permissions` - Row-level security
- `ai_context_rules`, `ai_sanitization_rules` - AI data governance
- `ai_data_access_log`, `permission_audit_log` - Audit trails

### Functional Migrations (No Tables)
These migrations contain functions, indexes, policies, and other database objects:
- `20250101000008_business_logic_functions.sql` - Business logic functions
- `20250101000009_rls_policies.sql` - Row-level security policies
- `20250101000010-14_*analytics*.sql` - Analytics views and functions
- `20250101000016_auth_users_organization.sql` - Auth functions
- `20250101000017_fix_reference_tables_rls.sql` - RLS fixes
- `20250101000020_duplicate_detection.sql` - Duplicate detection logic
- `20250101000022-27_*integration*.sql` - AI integrations
- `20250101000030-34_*queue*.sql` - Queue system
- `20250101000041-47_*views*.sql` - Database views
- `20250101000050-54_*functions*.sql` - Business functions

## Request Handling Priorities

1. **Database Schema Questions**: High priority - foundation of the system
2. **Core Module Implementation**: High priority - CRM, Products, Sales, Accounting
3. **Authentication/Authorization**: High priority - security critical
4. **Analytics Implementation**: Medium priority - business intelligence
5. **Enhancement Features**: Medium-Low priority - nice-to-have features
6. **Experimental Features**: Low priority - developmental features

## Common Request Patterns

- Schema explanations and relationships
- Module implementation guidance
- API endpoint creation
- Frontend component development
- Integration between modules
- Testing strategies

## Response Strategy

For future requests:
1. First check if the request relates to existing migration files
2. Prioritize based on the module importance (foundation > core > analytics > enhancements)
3. Provide clear, actionable guidance
4. Reference specific migration files when applicable
5. Offer to implement changes if within scope
