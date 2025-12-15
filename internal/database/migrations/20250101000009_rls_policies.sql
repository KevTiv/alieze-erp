-- Migration: Row Level Security Policies
-- Description: Multi-tenant RLS policies for all tables
-- Created: 2025-01-01

-- =====================================================
-- ENABLE RLS ON ALL TABLES
-- =====================================================

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequences ENABLE ROW LEVEL SECURITY;

-- Shared reference tables
ALTER TABLE countries ENABLE ROW LEVEL SECURITY;
ALTER TABLE states ENABLE ROW LEVEL SECURITY;
ALTER TABLE currencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE uom_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE uom_units ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_account_types ENABLE ROW LEVEL SECURITY;

-- Reference tables (shared or org-specific)
ALTER TABLE payment_terms ENABLE ROW LEVEL SECURITY;
ALTER TABLE fiscal_positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE analytic_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE utm_campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE utm_mediums ENABLE ROW LEVEL SECURITY;
ALTER TABLE utm_sources ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_methods ENABLE ROW LEVEL SECURITY;
ALTER TABLE bank_accounts ENABLE ROW LEVEL SECURITY;

-- CRM
ALTER TABLE contact_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE lead_stages ENABLE ROW LEVEL SECURITY;
ALTER TABLE lead_sources ENABLE ROW LEVEL SECURITY;
ALTER TABLE lost_reasons ENABLE ROW LEVEL SECURITY;
ALTER TABLE leads ENABLE ROW LEVEL SECURITY;
ALTER TABLE activities ENABLE ROW LEVEL SECURITY;

-- Products & Inventory
ALTER TABLE product_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_variants ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_packages ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_lots ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_quants ENABLE ROW LEVEL SECURITY;
ALTER TABLE procurement_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_picking_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_pickings ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_moves ENABLE ROW LEVEL SECURITY;

-- Sales
ALTER TABLE pricelists ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_lines ENABLE ROW LEVEL SECURITY;

-- Accounting
ALTER TABLE account_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_journals ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_tax_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_taxes ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_full_reconcile ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoice_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_invoice_allocation ENABLE ROW LEVEL SECURITY;

-- Purchasing
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_order_lines ENABLE ROW LEVEL SECURITY;

-- Manufacturing
ALTER TABLE workcenters ENABLE ROW LEVEL SECURITY;
ALTER TABLE bom_bills ENABLE ROW LEVEL SECURITY;
ALTER TABLE bom_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE manufacturing_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE work_orders ENABLE ROW LEVEL SECURITY;

-- HR
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;
ALTER TABLE job_positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE timesheets ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_requests ENABLE ROW LEVEL SECURITY;

-- Projects
ALTER TABLE task_stages ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;

-- =====================================================
-- HELPER FUNCTION FOR RLS
-- =====================================================

-- Get current user's organization ID from JWT or session
CREATE OR REPLACE FUNCTION get_current_organization_id()
RETURNS uuid
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    -- Try to get from JWT claim first
    RETURN (current_setting('request.jwt.claims', true)::json->>'organization_id')::uuid;
EXCEPTION
    WHEN OTHERS THEN
        -- Fallback to session variable
        RETURN current_setting('app.current_organization_id', true)::uuid;
END;
$$;

-- Check if user has permission in organization
CREATE OR REPLACE FUNCTION user_has_org_access()
RETURNS boolean
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_user_id uuid;
    v_org_id uuid;
BEGIN
    v_user_id := (SELECT auth.uid());
    v_org_id := get_current_organization_id();

    IF v_user_id IS NULL OR v_org_id IS NULL THEN
        RETURN false;
    END IF;

    RETURN EXISTS (
        SELECT 1
        FROM organization_users
        WHERE user_id = v_user_id
          AND organization_id = v_org_id
          AND is_active = true
    );
END;
$$;

-- =====================================================
-- RLS POLICIES - FOUNDATION TABLES
-- =====================================================

-- Organizations: users can see their own organizations
CREATE POLICY organizations_select ON organizations
    FOR SELECT
    USING (
        id IN (
            SELECT organization_id
            FROM organization_users
            WHERE user_id = (SELECT auth.uid())
              AND is_active = true
        )
    );

CREATE POLICY organizations_update ON organizations
    FOR UPDATE
    USING (
        id IN (
            SELECT organization_id
            FROM organization_users
            WHERE user_id = (SELECT auth.uid())
              AND role IN ('owner', 'admin')
              AND is_active = true
        )
    );

-- Organization Users
CREATE POLICY organization_users_select ON organization_users
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id
            FROM organization_users ou
            WHERE ou.user_id = (SELECT auth.uid())
              AND ou.is_active = true
        )
    );

-- Companies
CREATE POLICY companies_select ON companies
    FOR SELECT
    USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()));

CREATE POLICY companies_insert ON companies
    FOR INSERT
    WITH CHECK (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()));

CREATE POLICY companies_update ON companies
    FOR UPDATE
    USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()));

CREATE POLICY companies_delete ON companies
    FOR DELETE
    USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()));

-- =====================================================
-- GENERIC RLS POLICY TEMPLATE
-- For all tables with organization_id
-- =====================================================

-- This macro creates standard CRUD policies for organization-scoped tables
DO $$
DECLARE
    table_name text;
    tables_list text[] := ARRAY[
        'sequences', 'payment_terms', 'fiscal_positions', 'analytic_accounts',
        'utm_campaigns', 'utm_mediums', 'utm_sources', 'payment_methods', 'bank_accounts',
        'contact_tags', 'sales_teams', 'contacts', 'lead_stages', 'lead_sources',
        'lost_reasons', 'leads', 'activities',
        'product_categories', 'products', 'product_variants', 'warehouses',
        'stock_locations', 'stock_packages', 'stock_lots', 'stock_quants',
        'procurement_groups', 'stock_rules', 'stock_picking_types', 'stock_pickings', 'stock_moves',
        'pricelists', 'sales_orders', 'sales_order_lines',
        'account_groups', 'account_accounts', 'account_journals', 'account_tax_groups',
        'account_taxes', 'invoices', 'account_full_reconcile', 'invoice_lines',
        'payments', 'payment_invoice_allocation',
        'purchase_orders', 'purchase_order_lines',
        'workcenters', 'bom_bills', 'bom_lines', 'manufacturing_orders', 'work_orders',
        'resources', 'departments', 'job_positions', 'employees', 'timesheets',
        'leave_types', 'leave_requests',
        'task_stages', 'projects', 'tasks'
    ];
BEGIN
    FOREACH table_name IN ARRAY tables_list
    LOOP
        -- SELECT policy
        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR SELECT
            USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_select', table_name);

        -- INSERT policy
        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR INSERT
            WITH CHECK (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_insert', table_name);

        -- UPDATE policy
        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR UPDATE
            USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_update', table_name);

        -- DELETE policy
        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR DELETE
            USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_delete', table_name);
    END LOOP;
END $$;

-- =====================================================
-- REFERENCE DATA POLICIES (Shared tables)
-- =====================================================

-- Countries, states, currencies, etc. are readable by all authenticated users
CREATE POLICY countries_select ON countries FOR SELECT USING (true);
CREATE POLICY states_select ON states FOR SELECT USING (true);
CREATE POLICY currencies_select ON currencies FOR SELECT USING (true);
CREATE POLICY uom_categories_select ON uom_categories FOR SELECT USING (true);
CREATE POLICY uom_units_select ON uom_units FOR SELECT USING (true);
CREATE POLICY industries_select ON industries FOR SELECT USING (true);
CREATE POLICY account_account_types_select ON account_account_types FOR SELECT USING (true);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION get_current_organization_id IS 'Get current user organization ID from JWT or session';
COMMENT ON FUNCTION user_has_org_access IS 'Check if current user has access to current organization';
