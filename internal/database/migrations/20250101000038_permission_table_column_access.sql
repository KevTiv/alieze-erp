-- Migration: Permission System - Table & Column Level Access
-- Description: Granular table and column-level permission controls
-- Created: 2025-01-01
-- Module: Table & Column Permissions

-- =====================================================
-- TABLE-LEVEL PERMISSIONS
-- =====================================================

-- Permission policies for tables
CREATE TABLE permission_table_policies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id uuid REFERENCES permission_roles(id) ON DELETE CASCADE,
    group_id uuid REFERENCES permission_groups(id) ON DELETE CASCADE,
    user_id uuid, -- For user-specific overrides

    -- Resource identification
    table_name text NOT NULL,
    schema_name text DEFAULT 'public',

    -- Permissions
    actions permission_action[] NOT NULL,
    effect permission_effect DEFAULT 'allow',
    scope permission_scope DEFAULT 'organization',

    -- Conditions
    conditions jsonb, -- Additional SQL conditions as JSON
    priority integer DEFAULT 100, -- Higher priority policies override lower ones

    -- Temporal constraints
    valid_from timestamptz DEFAULT now(),
    valid_until timestamptz,

    -- Metadata
    description text,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT permission_table_policies_role_or_group_or_user CHECK (
        (role_id IS NOT NULL AND group_id IS NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NOT NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NULL AND user_id IS NOT NULL)
    ),
    CONSTRAINT permission_table_policies_priority_check CHECK (priority BETWEEN 1 AND 1000),
    CONSTRAINT permission_table_policies_valid_period CHECK (
        valid_until IS NULL OR valid_until > valid_from
    )
);

CREATE INDEX idx_permission_table_policies_org ON permission_table_policies(organization_id, table_name)
    WHERE is_active = true;
CREATE INDEX idx_permission_table_policies_role ON permission_table_policies(role_id)
    WHERE is_active = true AND role_id IS NOT NULL;
CREATE INDEX idx_permission_table_policies_group ON permission_table_policies(group_id)
    WHERE is_active = true AND group_id IS NOT NULL;
CREATE INDEX idx_permission_table_policies_user ON permission_table_policies(user_id, organization_id)
    WHERE is_active = true AND user_id IS NOT NULL;
CREATE INDEX idx_permission_table_policies_temporal ON permission_table_policies(valid_from, valid_until)
    WHERE is_active = true;

COMMENT ON TABLE permission_table_policies IS 'Defines table-level access permissions for roles, groups, or users';

-- =====================================================
-- COLUMN-LEVEL PERMISSIONS
-- =====================================================

-- Column visibility and access control
CREATE TABLE permission_column_policies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id uuid REFERENCES permission_roles(id) ON DELETE CASCADE,
    group_id uuid REFERENCES permission_groups(id) ON DELETE CASCADE,
    user_id uuid, -- For user-specific overrides

    -- Resource identification
    table_name text NOT NULL,
    column_name text NOT NULL,
    schema_name text DEFAULT 'public',

    -- Permissions
    can_view boolean DEFAULT true,    -- Column visible in SELECT
    can_read boolean DEFAULT true,    -- Can read actual values
    can_write boolean DEFAULT false,  -- Can UPDATE this column
    can_export boolean DEFAULT true,  -- Can export column data

    -- Special behaviors
    mask_type varchar(50), -- 'partial', 'full', 'hash', 'redact'
    mask_pattern text,      -- e.g., 'XXX-XX-####' for SSN
    default_value text,     -- Value to show if no read permission

    -- Conditions
    conditions jsonb, -- Additional conditions for access
    priority integer DEFAULT 100,

    -- Metadata
    description text,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT permission_column_policies_role_or_group_or_user CHECK (
        (role_id IS NOT NULL AND group_id IS NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NOT NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NULL AND user_id IS NOT NULL)
    ),
    CONSTRAINT permission_column_policies_mask_type CHECK (
        mask_type IS NULL OR mask_type IN ('partial', 'full', 'hash', 'redact', 'custom')
    ),
    CONSTRAINT permission_column_policies_priority_check CHECK (priority BETWEEN 1 AND 1000)
);

CREATE INDEX idx_permission_column_policies_org ON permission_column_policies(organization_id, table_name)
    WHERE is_active = true;
CREATE INDEX idx_permission_column_policies_role ON permission_column_policies(role_id)
    WHERE is_active = true AND role_id IS NOT NULL;
CREATE INDEX idx_permission_column_policies_group ON permission_column_policies(group_id)
    WHERE is_active = true AND group_id IS NOT NULL;
CREATE INDEX idx_permission_column_policies_user ON permission_column_policies(user_id, organization_id)
    WHERE is_active = true AND user_id IS NOT NULL;
CREATE INDEX idx_permission_column_policies_column ON permission_column_policies(table_name, column_name)
    WHERE is_active = true;

COMMENT ON TABLE permission_column_policies IS 'Defines column-level visibility and access permissions';

-- =====================================================
-- ROW-LEVEL FILTERS
-- =====================================================

-- Custom row-level security filters
CREATE TABLE permission_row_filters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id uuid REFERENCES permission_roles(id) ON DELETE CASCADE,
    group_id uuid REFERENCES permission_groups(id) ON DELETE CASCADE,
    user_id uuid,

    -- Resource identification
    table_name text NOT NULL,
    schema_name text DEFAULT 'public',

    -- Filter configuration
    filter_name varchar(100) NOT NULL,
    filter_type varchar(50) NOT NULL, -- 'sql', 'jsonb', 'function'
    filter_sql text,                  -- SQL WHERE clause
    filter_jsonb jsonb,               -- JSONB-based conditions
    filter_function text,             -- Function name to call

    -- Filter scope
    applies_to permission_action[] DEFAULT ARRAY['select']::permission_action[],
    scope permission_scope DEFAULT 'organization',

    -- Conditions
    priority integer DEFAULT 100,
    combine_with varchar(20) DEFAULT 'AND', -- How to combine with other filters

    -- Metadata
    description text,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT permission_row_filters_role_or_group_or_user CHECK (
        (role_id IS NOT NULL AND group_id IS NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NOT NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NULL AND user_id IS NOT NULL)
    ),
    CONSTRAINT permission_row_filters_type CHECK (
        filter_type IN ('sql', 'jsonb', 'function')
    ),
    CONSTRAINT permission_row_filters_combine CHECK (
        combine_with IN ('AND', 'OR')
    ),
    CONSTRAINT permission_row_filters_priority_check CHECK (priority BETWEEN 1 AND 1000),
    CONSTRAINT permission_row_filters_has_filter CHECK (
        (filter_type = 'sql' AND filter_sql IS NOT NULL) OR
        (filter_type = 'jsonb' AND filter_jsonb IS NOT NULL) OR
        (filter_type = 'function' AND filter_function IS NOT NULL)
    )
);

CREATE INDEX idx_permission_row_filters_org ON permission_row_filters(organization_id, table_name)
    WHERE is_active = true;
CREATE INDEX idx_permission_row_filters_role ON permission_row_filters(role_id)
    WHERE is_active = true AND role_id IS NOT NULL;
CREATE INDEX idx_permission_row_filters_group ON permission_row_filters(group_id)
    WHERE is_active = true AND group_id IS NOT NULL;
CREATE INDEX idx_permission_row_filters_user ON permission_row_filters(user_id, organization_id)
    WHERE is_active = true AND user_id IS NOT NULL;

COMMENT ON TABLE permission_row_filters IS 'Defines custom row-level security filters for fine-grained access control';

-- =====================================================
-- PERMISSION EVALUATION FUNCTIONS
-- =====================================================

-- Check if user has permission on a table
CREATE OR REPLACE FUNCTION check_table_permission(
    p_user_id uuid,
    p_organization_id uuid,
    p_table_name text,
    p_action permission_action,
    p_company_id uuid DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_has_permission boolean := false;
    v_explicit_deny boolean := false;
    v_user_roles uuid[];
    v_user_groups uuid[];
BEGIN
    -- Get all user roles
    SELECT array_agg(role_id) INTO v_user_roles
    FROM get_user_roles(p_user_id, p_organization_id, p_company_id);

    -- Get all user groups (through roles)
    SELECT array_agg(DISTINCT group_id) INTO v_user_groups
    FROM role_permission_groups
    WHERE role_id = ANY(v_user_roles);

    -- Check for explicit deny (highest priority)
    SELECT EXISTS (
        SELECT 1
        FROM permission_table_policies
        WHERE organization_id = p_organization_id
          AND table_name = p_table_name
          AND p_action = ANY(actions)
          AND effect = 'deny'
          AND is_active = true
          AND (
              (user_id = p_user_id) OR
              (role_id = ANY(v_user_roles)) OR
              (group_id = ANY(v_user_groups))
          )
          AND (valid_from IS NULL OR valid_from <= now())
          AND (valid_until IS NULL OR valid_until > now())
    ) INTO v_explicit_deny;

    IF v_explicit_deny THEN
        RETURN false;
    END IF;

    -- Check for explicit allow
    SELECT EXISTS (
        SELECT 1
        FROM permission_table_policies
        WHERE organization_id = p_organization_id
          AND table_name = p_table_name
          AND p_action = ANY(actions)
          AND effect = 'allow'
          AND is_active = true
          AND (
              (user_id = p_user_id) OR
              (role_id = ANY(v_user_roles)) OR
              (group_id = ANY(v_user_groups))
          )
          AND (valid_from IS NULL OR valid_from <= now())
          AND (valid_until IS NULL OR valid_until > now())
        ORDER BY priority DESC
        LIMIT 1
    ) INTO v_has_permission;

    RETURN v_has_permission;
END;
$$;

COMMENT ON FUNCTION check_table_permission IS 'Check if user has specific permission on a table';

-- Get column permissions for a user
CREATE OR REPLACE FUNCTION get_column_permissions(
    p_user_id uuid,
    p_organization_id uuid,
    p_table_name text,
    p_column_name text DEFAULT NULL
)
RETURNS TABLE (
    column_name text,
    can_view boolean,
    can_read boolean,
    can_write boolean,
    can_export boolean,
    mask_type varchar,
    mask_pattern text
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_user_roles uuid[];
    v_user_groups uuid[];
BEGIN
    -- Get all user roles
    SELECT array_agg(role_id) INTO v_user_roles
    FROM get_user_roles(p_user_id, p_organization_id);

    -- Get all user groups
    SELECT array_agg(DISTINCT group_id) INTO v_user_groups
    FROM role_permission_groups
    WHERE role_id = ANY(v_user_roles);

    RETURN QUERY
    SELECT DISTINCT ON (cp.column_name)
        cp.column_name,
        COALESCE(cp.can_view, true),
        COALESCE(cp.can_read, true),
        COALESCE(cp.can_write, false),
        COALESCE(cp.can_export, true),
        cp.mask_type,
        cp.mask_pattern
    FROM permission_column_policies cp
    WHERE cp.organization_id = p_organization_id
      AND cp.table_name = p_table_name
      AND (p_column_name IS NULL OR cp.column_name = p_column_name)
      AND cp.is_active = true
      AND (
          (cp.user_id = p_user_id) OR
          (cp.role_id = ANY(v_user_roles)) OR
          (cp.group_id = ANY(v_user_groups))
      )
    ORDER BY cp.column_name, cp.priority DESC;
END;
$$;

COMMENT ON FUNCTION get_column_permissions IS 'Get column-level permissions for a user on a table';

-- Build row filter for a user
CREATE OR REPLACE FUNCTION build_row_filter(
    p_user_id uuid,
    p_organization_id uuid,
    p_table_name text,
    p_action permission_action DEFAULT 'select'
)
RETURNS text
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_user_roles uuid[];
    v_user_groups uuid[];
    v_filters text[];
    v_filter record;
    v_combined_filter text;
BEGIN
    -- Get all user roles
    SELECT array_agg(role_id) INTO v_user_roles
    FROM get_user_roles(p_user_id, p_organization_id);

    -- Get all user groups
    SELECT array_agg(DISTINCT group_id) INTO v_user_groups
    FROM role_permission_groups
    WHERE role_id = ANY(v_user_roles);

    -- Collect all applicable filters
    FOR v_filter IN
        SELECT filter_sql, filter_type, combine_with
        FROM permission_row_filters
        WHERE organization_id = p_organization_id
          AND table_name = p_table_name
          AND p_action = ANY(applies_to)
          AND is_active = true
          AND filter_type = 'sql'
          AND filter_sql IS NOT NULL
          AND (
              (user_id = p_user_id) OR
              (role_id = ANY(v_user_roles)) OR
              (group_id = ANY(v_user_groups))
          )
        ORDER BY priority DESC
    LOOP
        v_filters := array_append(v_filters, '(' || v_filter.filter_sql || ')');
    END LOOP;

    -- Combine filters
    IF array_length(v_filters, 1) > 0 THEN
        v_combined_filter := array_to_string(v_filters, ' AND ');
    ELSE
        v_combined_filter := 'TRUE'; -- No filters means full access
    END IF;

    RETURN v_combined_filter;
END;
$$;

COMMENT ON FUNCTION build_row_filter IS 'Build SQL WHERE clause for row-level filtering';

-- =====================================================
-- HELPER FUNCTIONS FOR DATA MASKING
-- =====================================================

-- Mask sensitive data based on pattern
CREATE OR REPLACE FUNCTION mask_data(
    p_value text,
    p_mask_type varchar,
    p_mask_pattern text DEFAULT NULL
)
RETURNS text
LANGUAGE plpgsql
IMMUTABLE
AS $$
BEGIN
    IF p_value IS NULL THEN
        RETURN NULL;
    END IF;

    CASE p_mask_type
        WHEN 'full' THEN
            RETURN repeat('*', length(p_value));

        WHEN 'partial' THEN
            -- Show first and last 2 characters
            IF length(p_value) <= 4 THEN
                RETURN repeat('*', length(p_value));
            ELSE
                RETURN substr(p_value, 1, 2) || repeat('*', length(p_value) - 4) || substr(p_value, -2);
            END IF;

        WHEN 'hash' THEN
            RETURN 'HASH:' || substr(md5(p_value), 1, 8);

        WHEN 'redact' THEN
            RETURN '[REDACTED]';

        WHEN 'custom' THEN
            -- Use provided pattern (e.g., 'XXX-XX-####' for SSN)
            IF p_mask_pattern IS NOT NULL THEN
                -- Simple pattern replacement (can be enhanced)
                RETURN p_mask_pattern;
            ELSE
                RETURN '[MASKED]';
            END IF;

        ELSE
            RETURN p_value; -- No masking
    END CASE;
END;
$$;

COMMENT ON FUNCTION mask_data IS 'Mask sensitive data according to specified pattern';

-- =====================================================
-- DEFAULT POLICIES FOR COMMON SCENARIOS
-- =====================================================

-- Function to create default policies for a role
CREATE OR REPLACE FUNCTION create_default_table_policies(
    p_organization_id uuid,
    p_role_id uuid,
    p_role_type varchar
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    CASE p_role_type
        WHEN 'viewer' THEN
            -- Viewers can only SELECT from most tables
            INSERT INTO permission_table_policies (
                organization_id, role_id, table_name, actions, effect, scope
            )
            SELECT p_organization_id, p_role_id, table_name, ARRAY['select']::permission_action[], 'allow', 'organization'
            FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_type = 'BASE TABLE'
              AND table_name NOT IN ('audit_logs', 'permission_table_policies', 'permission_column_policies');

        WHEN 'user' THEN
            -- Users can SELECT and INSERT/UPDATE their own records
            INSERT INTO permission_table_policies (
                organization_id, role_id, table_name, actions, effect, scope
            )
            SELECT p_organization_id, p_role_id, table_name,
                   ARRAY['select', 'insert', 'update']::permission_action[], 'allow', 'self'
            FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_type = 'BASE TABLE'
              AND table_name NOT LIKE 'permission%';

        WHEN 'manager' THEN
            -- Managers have broader access
            INSERT INTO permission_table_policies (
                organization_id, role_id, table_name, actions, effect, scope
            )
            SELECT p_organization_id, p_role_id, table_name,
                   ARRAY['select', 'insert', 'update', 'delete']::permission_action[], 'allow', 'team'
            FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_type = 'BASE TABLE';

        ELSE
            -- Admin/Owner have full access by default
            NULL;
    END CASE;
END;
$$;

COMMENT ON FUNCTION create_default_table_policies IS 'Create default table policies for standard roles';

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER update_permission_table_policies_updated_at
    BEFORE UPDATE ON permission_table_policies
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

CREATE TRIGGER update_permission_column_policies_updated_at
    BEFORE UPDATE ON permission_column_policies
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

CREATE TRIGGER update_permission_row_filters_updated_at
    BEFORE UPDATE ON permission_row_filters
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

-- =====================================================
-- RLS POLICIES
-- =====================================================

ALTER TABLE permission_table_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE permission_column_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE permission_row_filters ENABLE ROW LEVEL SECURITY;

-- Table policies
CREATE POLICY "Users can view table policies in their org"
    ON permission_table_policies FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage table policies"
    ON permission_table_policies FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- Column policies
CREATE POLICY "Users can view column policies in their org"
    ON permission_column_policies FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage column policies"
    ON permission_column_policies FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- Row filters
CREATE POLICY "Users can view row filters in their org"
    ON permission_row_filters FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage row filters"
    ON permission_row_filters FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- =====================================================
-- SAMPLE ROW FILTERS
-- =====================================================

COMMENT ON TABLE permission_table_policies IS 'Example filters:
-- Department filter:
{
  "filter_name": "department_filter",
  "filter_sql": "department_id IN (SELECT department_id FROM employees WHERE user_id = $USER_ID)",
  "filter_type": "sql"
}

-- Date range filter:
{
  "filter_name": "date_range_filter",
  "filter_sql": "created_at >= CURRENT_DATE - INTERVAL ''30 days''",
  "filter_type": "sql"
}

-- Status filter:
{
  "filter_name": "active_only",
  "filter_sql": "status = ''active''",
  "filter_type": "sql"
}

-- Hierarchical filter:
{
  "filter_name": "subordinates_only",
  "filter_sql": "created_by IN (SELECT user_id FROM employees WHERE reports_to = $USER_ID OR user_id = $USER_ID)",
  "filter_type": "sql"
}';
