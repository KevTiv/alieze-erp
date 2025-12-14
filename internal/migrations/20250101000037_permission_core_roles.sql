-- Migration: Permission System - Core Roles & Templates
-- Description: Extensible role-based permission system foundation
-- Created: 2025-01-01
-- Module: Core Roles

-- =====================================================
-- ENABLE EXTENSIONS
-- =====================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "btree_gist"; -- For exclusion constraints

-- =====================================================
-- ENUMS & TYPES
-- =====================================================

-- Permission action types
CREATE TYPE permission_action AS ENUM (
    'view',      -- Can see resource exists
    'select',    -- Can read data
    'insert',    -- Can create new records
    'update',    -- Can modify existing records
    'delete',    -- Can remove records
    'execute',   -- Can execute functions/procedures
    'truncate',  -- Can remove all records
    'references', -- Can create foreign keys
    'export',    -- Can export data
    'import'     -- Can import data
);

-- Permission scope
CREATE TYPE permission_scope AS ENUM (
    'organization', -- Entire organization
    'company',      -- Specific company within org
    'department',   -- Department level
    'team',         -- Team level
    'self'          -- Only own records
);

-- Permission effect
CREATE TYPE permission_effect AS ENUM (
    'allow',  -- Explicitly allow
    'deny'    -- Explicitly deny (takes precedence)
);

-- =====================================================
-- CORE PERMISSION TABLES
-- =====================================================

-- Permission Roles (Custom roles beyond the basic 5)
CREATE TABLE permission_roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    display_name varchar(255) NOT NULL,
    description text,
    base_role varchar(50), -- Inherits from base role (owner, admin, manager, user, viewer)
    is_system boolean DEFAULT false, -- System roles can't be deleted
    is_active boolean DEFAULT true,
    priority integer DEFAULT 100, -- Higher priority roles override lower ones

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT permission_roles_unique UNIQUE(organization_id, name),
    CONSTRAINT permission_roles_base_check CHECK (
        base_role IS NULL OR
        base_role IN ('owner', 'admin', 'manager', 'user', 'viewer')
    ),
    CONSTRAINT permission_roles_priority_check CHECK (priority BETWEEN 1 AND 1000)
);

CREATE INDEX idx_permission_roles_org ON permission_roles(organization_id) WHERE is_active = true;
CREATE INDEX idx_permission_roles_priority ON permission_roles(organization_id, priority DESC);

COMMENT ON TABLE permission_roles IS 'Custom roles that organizations can create for fine-grained access control';

-- Role Inheritance (roles can inherit from other roles)
CREATE TABLE permission_role_inheritance (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    child_role_id uuid NOT NULL REFERENCES permission_roles(id) ON DELETE CASCADE,
    parent_role_id uuid NOT NULL REFERENCES permission_roles(id) ON DELETE CASCADE,
    inherit_all boolean DEFAULT true, -- If false, specific permissions must be defined

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,

    -- Constraints
    CONSTRAINT permission_role_inheritance_unique UNIQUE(child_role_id, parent_role_id),
    CONSTRAINT permission_role_inheritance_no_self CHECK (child_role_id != parent_role_id)
);

CREATE INDEX idx_permission_role_inheritance_child ON permission_role_inheritance(child_role_id);
CREATE INDEX idx_permission_role_inheritance_parent ON permission_role_inheritance(parent_role_id);

COMMENT ON TABLE permission_role_inheritance IS 'Defines role hierarchy and inheritance relationships';

-- Permission Templates (pre-built role templates)
CREATE TABLE permission_role_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(50) UNIQUE NOT NULL,
    name varchar(255) NOT NULL,
    category varchar(100) NOT NULL, -- 'sales', 'finance', 'hr', 'operations', etc.
    description text,
    base_role varchar(50),
    permissions jsonb NOT NULL DEFAULT '[]'::jsonb, -- Array of permission definitions
    is_active boolean DEFAULT true,

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Constraints
    CONSTRAINT permission_role_templates_category_check CHECK (
        category IN ('general', 'sales', 'finance', 'hr', 'operations', 'support', 'custom')
    )
);

CREATE INDEX idx_permission_role_templates_category ON permission_role_templates(category) WHERE is_active = true;

COMMENT ON TABLE permission_role_templates IS 'Pre-built role templates that organizations can use as starting points';

-- User Role Assignments (beyond the single role in organization_users)
CREATE TABLE user_role_assignments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL REFERENCES permission_roles(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id) ON DELETE CASCADE, -- Optional: role for specific company
    scope permission_scope DEFAULT 'organization',

    -- Temporal permissions
    valid_from timestamptz DEFAULT now(),
    valid_until timestamptz, -- NULL means no expiration

    -- Assignment metadata
    assigned_by uuid,
    assignment_reason text,
    is_active boolean DEFAULT true,

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Constraints
    CONSTRAINT user_role_assignments_unique UNIQUE(user_id, role_id, company_id),
    CONSTRAINT user_role_assignments_valid_period CHECK (
        valid_until IS NULL OR valid_until > valid_from
    )
);

CREATE INDEX idx_user_role_assignments_user ON user_role_assignments(user_id, organization_id)
    WHERE is_active = true;
CREATE INDEX idx_user_role_assignments_role ON user_role_assignments(role_id)
    WHERE is_active = true;
CREATE INDEX idx_user_role_assignments_temporal ON user_role_assignments(valid_from, valid_until)
    WHERE is_active = true;

COMMENT ON TABLE user_role_assignments IS 'Assigns custom roles to users with optional temporal and scope constraints';

-- Permission Groups (bundle permissions for easier management)
CREATE TABLE permission_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    display_name varchar(255) NOT NULL,
    description text,
    is_active boolean DEFAULT true,

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT permission_groups_unique UNIQUE(organization_id, name)
);

CREATE INDEX idx_permission_groups_org ON permission_groups(organization_id) WHERE is_active = true;

COMMENT ON TABLE permission_groups IS 'Bundle related permissions together for easier management';

-- Role to Permission Group Mapping
CREATE TABLE role_permission_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id uuid NOT NULL REFERENCES permission_roles(id) ON DELETE CASCADE,
    group_id uuid NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    effect permission_effect DEFAULT 'allow',

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,

    -- Constraints
    CONSTRAINT role_permission_groups_unique UNIQUE(role_id, group_id)
);

CREATE INDEX idx_role_permission_groups_role ON role_permission_groups(role_id);
CREATE INDEX idx_role_permission_groups_group ON role_permission_groups(group_id);

COMMENT ON TABLE role_permission_groups IS 'Maps roles to permission groups';

-- =====================================================
-- HELPER FUNCTIONS
-- =====================================================

-- Get all roles for a user (including inherited)
CREATE OR REPLACE FUNCTION get_user_roles(
    p_user_id uuid,
    p_organization_id uuid,
    p_company_id uuid DEFAULT NULL
)
RETURNS TABLE (
    role_id uuid,
    role_name varchar,
    base_role varchar,
    priority integer,
    source text -- 'direct', 'inherited', 'base'
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE role_hierarchy AS (
        -- Direct role assignments
        SELECT
            pr.id AS role_id,
            pr.name AS role_name,
            pr.base_role,
            pr.priority,
            'direct'::text AS source
        FROM user_role_assignments ura
        JOIN permission_roles pr ON pr.id = ura.role_id
        WHERE ura.user_id = p_user_id
          AND ura.organization_id = p_organization_id
          AND ura.is_active = true
          AND pr.is_active = true
          AND (p_company_id IS NULL OR ura.company_id = p_company_id OR ura.company_id IS NULL)
          AND (ura.valid_from IS NULL OR ura.valid_from <= now())
          AND (ura.valid_until IS NULL OR ura.valid_until > now())

        UNION

        -- Inherited roles
        SELECT
            pr.id AS role_id,
            pr.name AS role_name,
            pr.base_role,
            pr.priority,
            'inherited'::text AS source
        FROM role_hierarchy rh
        JOIN permission_role_inheritance pri ON pri.child_role_id = rh.role_id
        JOIN permission_roles pr ON pr.id = pri.parent_role_id
        WHERE pr.is_active = true
    ),
    base_role AS (
        -- Get base role from organization_users
        SELECT
            NULL::uuid AS role_id,
            ou.role AS role_name,
            ou.role AS base_role,
            CASE ou.role
                WHEN 'owner' THEN 1000
                WHEN 'admin' THEN 900
                WHEN 'manager' THEN 500
                WHEN 'user' THEN 100
                WHEN 'viewer' THEN 50
            END AS priority,
            'base'::text AS source
        FROM organization_users ou
        WHERE ou.user_id = p_user_id
          AND ou.organization_id = p_organization_id
          AND ou.is_active = true
    )
    SELECT * FROM role_hierarchy
    UNION ALL
    SELECT * FROM base_role
    ORDER BY priority DESC, source;
END;
$$;

COMMENT ON FUNCTION get_user_roles IS 'Returns all roles for a user including inherited and base roles';

-- Check if user has a specific role
CREATE OR REPLACE FUNCTION user_has_role(
    p_user_id uuid,
    p_organization_id uuid,
    p_role_name varchar
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM get_user_roles(p_user_id, p_organization_id)
        WHERE role_name = p_role_name
    );
END;
$$;

COMMENT ON FUNCTION user_has_role IS 'Check if a user has a specific role (directly or through inheritance)';

-- =====================================================
-- INITIAL TEMPLATES
-- =====================================================

INSERT INTO permission_role_templates (code, name, category, base_role, description, permissions) VALUES
-- Sales Templates
('sales_rep', 'Sales Representative', 'sales', 'user',
 'Can manage own leads, opportunities, and quotes',
 '[
    {"resource": "leads", "actions": ["select", "insert", "update"], "scope": "self"},
    {"resource": "contacts", "actions": ["select", "insert", "update"], "scope": "team"},
    {"resource": "sales_orders", "actions": ["select", "insert", "update"], "scope": "self"},
    {"resource": "products", "actions": ["select"], "scope": "organization"}
 ]'::jsonb),

('sales_manager', 'Sales Manager', 'sales', 'manager',
 'Can manage team sales activities and view reports',
 '[
    {"resource": "leads", "actions": ["select", "insert", "update", "delete"], "scope": "team"},
    {"resource": "sales_orders", "actions": ["select", "insert", "update", "delete"], "scope": "team"},
    {"resource": "sales_reports", "actions": ["select", "execute"], "scope": "team"}
 ]'::jsonb),

-- Finance Templates
('accountant', 'Accountant', 'finance', 'user',
 'Can manage invoices and basic accounting',
 '[
    {"resource": "invoices", "actions": ["select", "insert", "update"], "scope": "organization"},
    {"resource": "payments", "actions": ["select", "insert"], "scope": "organization"},
    {"resource": "account_accounts", "actions": ["select"], "scope": "organization"}
 ]'::jsonb),

('financial_controller', 'Financial Controller', 'finance', 'manager',
 'Full access to financial data and reports',
 '[
    {"resource": "invoices", "actions": ["select", "insert", "update", "delete"], "scope": "organization"},
    {"resource": "payments", "actions": ["select", "insert", "update", "delete"], "scope": "organization"},
    {"resource": "financial_reports", "actions": ["select", "execute", "export"], "scope": "organization"}
 ]'::jsonb),

-- HR Templates
('hr_specialist', 'HR Specialist', 'hr', 'user',
 'Can manage employee records and leave requests',
 '[
    {"resource": "employees", "actions": ["select", "insert", "update"], "scope": "organization"},
    {"resource": "leave_requests", "actions": ["select", "update"], "scope": "organization"},
    {"resource": "departments", "actions": ["select"], "scope": "organization"}
 ]'::jsonb),

-- Operations Templates
('warehouse_operator', 'Warehouse Operator', 'operations', 'user',
 'Can manage stock movements and picking',
 '[
    {"resource": "stock_moves", "actions": ["select", "insert", "update"], "scope": "company"},
    {"resource": "stock_pickings", "actions": ["select", "update"], "scope": "company"},
    {"resource": "products", "actions": ["select"], "scope": "organization"}
 ]'::jsonb),

-- Support Templates
('support_agent', 'Support Agent', 'support', 'user',
 'Can view customer data and manage tickets',
 '[
    {"resource": "contacts", "actions": ["select"], "scope": "organization"},
    {"resource": "tickets", "actions": ["select", "insert", "update"], "scope": "team"},
    {"resource": "knowledge_base", "actions": ["select"], "scope": "organization"}
 ]'::jsonb);

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_permission_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_permission_roles_updated_at
    BEFORE UPDATE ON permission_roles
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

CREATE TRIGGER update_permission_groups_updated_at
    BEFORE UPDATE ON permission_groups
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

CREATE TRIGGER update_user_role_assignments_updated_at
    BEFORE UPDATE ON user_role_assignments
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

-- =====================================================
-- RLS POLICIES
-- =====================================================

ALTER TABLE permission_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE permission_role_inheritance ENABLE ROW LEVEL SECURITY;
ALTER TABLE permission_role_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_role_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE permission_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_permission_groups ENABLE ROW LEVEL SECURITY;

-- Permission roles policies
CREATE POLICY "Users can view roles in their org"
    ON permission_roles FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage roles"
    ON permission_roles FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- User role assignments policies
CREATE POLICY "Users can view role assignments in their org"
    ON user_role_assignments FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage role assignments"
    ON user_role_assignments FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- Templates are viewable by all authenticated users
CREATE POLICY "All users can view role templates"
    ON permission_role_templates FOR SELECT
    USING (is_active = true);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

CREATE INDEX idx_user_role_assignments_active
    ON user_role_assignments(user_id, organization_id, is_active, valid_from, valid_until);

CREATE INDEX idx_permission_roles_org_active
    ON permission_roles(organization_id, is_active, priority);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TYPE permission_action IS 'Available actions that can be performed on resources';
COMMENT ON TYPE permission_scope IS 'Scope level for permission application';
COMMENT ON TYPE permission_effect IS 'Whether to allow or deny the permission (deny takes precedence)';
