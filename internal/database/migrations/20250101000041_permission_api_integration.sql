-- Migration: Permission System - API Integration & Helper Functions
-- Description: User-friendly API functions for permission management
-- Created: 2025-01-01
-- Module: API Integration

-- =====================================================
-- PUBLIC API FUNCTIONS
-- =====================================================

-- Get current user's permissions
CREATE OR REPLACE FUNCTION get_my_permissions(
    p_table_name text DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
DECLARE
    v_user_id uuid;
    v_org_id uuid;
    v_result jsonb;
BEGIN
    v_user_id := (SELECT auth.uid());
    v_org_id := (SELECT get_current_user_organization_id());

    IF v_user_id IS NULL OR v_org_id IS NULL THEN
        RETURN jsonb_build_object(
            'error', 'Not authenticated or no organization'
        );
    END IF;

    IF p_table_name IS NULL THEN
        -- Return all permissions
        SELECT jsonb_build_object(
            'user_id', v_user_id,
            'organization_id', v_org_id,
            'base_role', ou.role,
            'custom_roles', (
                SELECT jsonb_agg(jsonb_build_object(
                    'role_id', pr.id,
                    'role_name', pr.name,
                    'display_name', pr.display_name
                ))
                FROM user_role_assignments ura
                JOIN permission_roles pr ON pr.id = ura.role_id
                WHERE ura.user_id = v_user_id
                  AND ura.organization_id = v_org_id
                  AND ura.is_active = true
            ),
            'table_permissions', (
                SELECT jsonb_object_agg(
                    table_name,
                    jsonb_build_object(
                        'actions', allowed_actions,
                        'scope', scope
                    )
                )
                FROM user_effective_permissions
                WHERE user_id = v_user_id
                  AND organization_id = v_org_id
            )
        ) INTO v_result
        FROM organization_users ou
        WHERE ou.user_id = v_user_id
          AND ou.organization_id = v_org_id;
    ELSE
        -- Return permissions for specific table
        v_result := check_user_access(
            v_user_id,
            v_org_id,
            p_table_name,
            'select'::permission_action
        );
    END IF;

    RETURN v_result;
END;
$$;

COMMENT ON FUNCTION get_my_permissions IS 'Get current user permissions (all or for specific table)';

-- Check if current user can perform action
CREATE OR REPLACE FUNCTION can_i(
    p_action text,
    p_table_name text,
    p_column_name text DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
DECLARE
    v_user_id uuid;
    v_org_id uuid;
    v_can_do boolean;
BEGIN
    v_user_id := (SELECT auth.uid());
    v_org_id := (SELECT get_current_user_organization_id());

    IF v_user_id IS NULL OR v_org_id IS NULL THEN
        RETURN false;
    END IF;

    -- Check table-level permission
    v_can_do := check_table_permission(
        v_user_id,
        v_org_id,
        p_table_name,
        p_action::permission_action
    );

    -- If checking column-specific permission
    IF v_can_do AND p_column_name IS NOT NULL THEN
        SELECT
            CASE p_action
                WHEN 'select' THEN can_read
                WHEN 'update' THEN can_write
                ELSE can_view
            END INTO v_can_do
        FROM get_column_permissions(
            v_user_id,
            v_org_id,
            p_table_name,
            p_column_name
        )
        LIMIT 1;
    END IF;

    RETURN COALESCE(v_can_do, false);
END;
$$;

COMMENT ON FUNCTION can_i IS 'Simple permission check - returns true/false for specific action';

-- Get AI-safe data for current user
CREATE OR REPLACE FUNCTION get_ai_safe_data(
    p_table_name text,
    p_limit integer DEFAULT 100
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
DECLARE
    v_user_id uuid;
    v_org_id uuid;
    v_ai_permission jsonb;
    v_columns text[];
    v_query text;
    v_result jsonb;
BEGIN
    v_user_id := (SELECT auth.uid());
    v_org_id := (SELECT get_current_user_organization_id());

    -- Check AI permission
    v_ai_permission := check_ai_data_permission(
        v_user_id,
        v_org_id,
        p_table_name
    );

    IF NOT (v_ai_permission->>'allowed')::boolean THEN
        RETURN jsonb_build_object(
            'error', 'No AI access to this table',
            'exposure_level', v_ai_permission->>'exposure_level'
        );
    END IF;

    -- Get allowed columns
    SELECT array_agg(column_name)
    INTO v_columns
    FROM get_column_permissions(v_user_id, v_org_id, p_table_name)
    WHERE can_read = true;

    -- Build safe query based on exposure level
    CASE v_ai_permission->>'exposure_level'
        WHEN 'metadata' THEN
            -- Return only structure, no data
            RETURN jsonb_build_object(
                'table', p_table_name,
                'columns', v_columns,
                'exposure_level', 'metadata',
                'data', '[]'::jsonb
            );

        WHEN 'aggregated' THEN
            -- Return only aggregated data
            v_query := format(
                'SELECT COUNT(*) as total_rows,
                        MIN(created_at) as earliest_record,
                        MAX(created_at) as latest_record
                 FROM %I
                 WHERE organization_id = %L',
                p_table_name, v_org_id
            );

        WHEN 'sanitized' THEN
            -- Return sanitized data
            v_query := format(
                'SELECT %s FROM %I
                 WHERE organization_id = %L
                 LIMIT %s',
                array_to_string(
                    ARRAY(
                        SELECT format(
                            'sanitize_for_ai(%L, %L, %I, %I::text) AS %I',
                            v_org_id, p_table_name, col, col, col
                        )
                        FROM unnest(v_columns) AS col
                    ),
                    ', '
                ),
                p_table_name, v_org_id, p_limit
            );

        WHEN 'full' THEN
            -- Return full data (but still respect column permissions)
            v_query := format(
                'SELECT %s FROM %I
                 WHERE organization_id = %L
                 LIMIT %s',
                array_to_string(v_columns, ', '),
                p_table_name, v_org_id, p_limit
            );

        ELSE
            RETURN jsonb_build_object(
                'error', 'Unknown exposure level'
            );
    END CASE;

    -- Execute query and return result
    EXECUTE format('SELECT jsonb_agg(row_to_json(t)) FROM (%s) t', v_query) INTO v_result;

    RETURN jsonb_build_object(
        'table', p_table_name,
        'exposure_level', v_ai_permission->>'exposure_level',
        'was_sanitized', (v_ai_permission->>'requires_sanitization')::boolean,
        'row_count', COALESCE(jsonb_array_length(v_result), 0),
        'data', COALESCE(v_result, '[]'::jsonb)
    );
END;
$$;

COMMENT ON FUNCTION get_ai_safe_data IS 'Get data that is safe to expose to AI based on permissions';

-- Quick permission setup for new organization
CREATE OR REPLACE FUNCTION setup_default_permissions(
    p_organization_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_org_id uuid;
    v_roles_created integer := 0;
    v_policies_created integer := 0;
BEGIN
    v_org_id := COALESCE(p_organization_id, (SELECT get_current_user_organization_id()));

    IF v_org_id IS NULL THEN
        RETURN jsonb_build_object(
            'error', 'No organization specified'
        );
    END IF;

    -- Create default custom roles from templates
    INSERT INTO permission_roles (organization_id, name, display_name, base_role, description)
    SELECT
        v_org_id,
        code || '_' || substr(md5(random()::text), 1, 6),
        name,
        base_role,
        description
    FROM permission_role_templates
    WHERE category IN ('sales', 'finance', 'operations')
      AND is_active = true
    LIMIT 5;

    GET DIAGNOSTICS v_roles_created = ROW_COUNT;

    -- Create default AI permissions
    PERFORM create_default_ai_permissions(v_org_id);

    -- Create basic table policies for viewer role
    INSERT INTO permission_table_policies (
        organization_id,
        role_id,
        table_name,
        actions,
        effect,
        scope
    )
    SELECT
        v_org_id,
        pr.id,
        ist.table_name,
        ARRAY['select']::permission_action[],
        'allow'::permission_effect,
        'organization'::permission_scope
    FROM permission_roles pr
    CROSS JOIN information_schema.tables ist
    WHERE pr.organization_id = v_org_id
      AND pr.base_role = 'viewer'
      AND ist.table_schema = 'public'
      AND ist.table_type = 'BASE TABLE'
      AND ist.table_name NOT LIKE 'permission%'
      AND ist.table_name NOT LIKE '%audit%'
    LIMIT 20;

    GET DIAGNOSTICS v_policies_created = ROW_COUNT;

    RETURN jsonb_build_object(
        'success', true,
        'organization_id', v_org_id,
        'roles_created', v_roles_created,
        'policies_created', v_policies_created,
        'message', 'Default permissions setup complete'
    );
END;
$$;

COMMENT ON FUNCTION setup_default_permissions IS 'Quick setup of default permissions for an organization';

-- =====================================================
-- DYNAMIC RLS POLICY GENERATOR
-- =====================================================

-- Generate RLS policy SQL for a table
CREATE OR REPLACE FUNCTION generate_rls_policy(
    p_table_name text,
    p_policy_name text DEFAULT NULL
)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
    v_policy_name text;
    v_policy_sql text;
BEGIN
    v_policy_name := COALESCE(p_policy_name, p_table_name || '_dynamic_rls');

    v_policy_sql := format('
-- Drop existing policy if exists
DROP POLICY IF EXISTS %I ON %I;

-- Create new dynamic RLS policy
CREATE POLICY %I ON %I
FOR ALL
USING (
    EXISTS (
        SELECT 1
        FROM user_effective_permissions uep
        WHERE uep.user_id = (SELECT auth.uid())
          AND uep.organization_id = (SELECT get_current_user_organization_id())
          AND uep.table_name = %L
          AND (
              -- Check if user has the required action
              CASE TG_OP
                  WHEN ''SELECT'' THEN ''select'' = ANY(uep.allowed_actions)
                  WHEN ''INSERT'' THEN ''insert'' = ANY(uep.allowed_actions)
                  WHEN ''UPDATE'' THEN ''update'' = ANY(uep.allowed_actions)
                  WHEN ''DELETE'' THEN ''delete'' = ANY(uep.allowed_actions)
              END
          )
          AND (
              -- Apply scope-based filtering
              CASE uep.scope
                  WHEN ''organization'' THEN %I.organization_id = (SELECT get_current_user_organization_id())
                  WHEN ''company'' THEN %I.company_id IN (
                      SELECT company_id FROM user_company_access
                      WHERE user_id = (SELECT auth.uid())
                  )
                  WHEN ''team'' THEN %I.team_id IN (
                      SELECT team_id FROM user_team_members
                      WHERE user_id = (SELECT auth.uid())
                  )
                  WHEN ''self'' THEN %I.created_by = (SELECT auth.uid())
                  ELSE false
              END
          )
    )
    -- Apply row filters
    AND (
        SELECT build_row_filter(
            (SELECT auth.uid()),
            (SELECT get_current_user_organization_id()),
            %L,
            CASE TG_OP
                WHEN ''SELECT'' THEN ''select''::permission_action
                WHEN ''INSERT'' THEN ''insert''::permission_action
                WHEN ''UPDATE'' THEN ''update''::permission_action
                WHEN ''DELETE'' THEN ''delete''::permission_action
            END
        )
    )::boolean
);

-- Enable RLS on the table
ALTER TABLE %I ENABLE ROW LEVEL SECURITY;

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON %I TO authenticated;
',
        v_policy_name, p_table_name,
        v_policy_name, p_table_name,
        p_table_name,
        p_table_name, p_table_name, p_table_name, p_table_name,
        p_table_name,
        p_table_name,
        p_table_name
    );

    RETURN v_policy_sql;
END;
$$;

COMMENT ON FUNCTION generate_rls_policy IS 'Generate dynamic RLS policy SQL for a table based on permission system';

-- =====================================================
-- PERMISSION MIGRATION HELPERS
-- =====================================================

-- Migrate existing simple roles to permission system
CREATE OR REPLACE FUNCTION migrate_existing_permissions(
    p_organization_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_org_id uuid;
    v_migrated_count integer := 0;
    v_user record;
BEGIN
    v_org_id := COALESCE(p_organization_id, (SELECT get_current_user_organization_id()));

    -- For each user in the organization
    FOR v_user IN
        SELECT user_id, role
        FROM organization_users
        WHERE organization_id = v_org_id
          AND is_active = true
    LOOP
        -- Create default table policies based on their base role
        CASE v_user.role
            WHEN 'owner', 'admin' THEN
                -- Full access to everything
                INSERT INTO permission_table_policies (
                    organization_id, user_id, table_name, actions, effect, scope
                )
                SELECT
                    v_org_id,
                    v_user.user_id,
                    table_name,
                    ARRAY['select', 'insert', 'update', 'delete']::permission_action[],
                    'allow'::permission_effect,
                    'organization'::permission_scope
                FROM information_schema.tables
                WHERE table_schema = 'public'
                  AND table_type = 'BASE TABLE'
                ON CONFLICT DO NOTHING;

            WHEN 'manager' THEN
                -- Read/write to most tables
                INSERT INTO permission_table_policies (
                    organization_id, user_id, table_name, actions, effect, scope
                )
                SELECT
                    v_org_id,
                    v_user.user_id,
                    table_name,
                    ARRAY['select', 'insert', 'update']::permission_action[],
                    'allow'::permission_effect,
                    'team'::permission_scope
                FROM information_schema.tables
                WHERE table_schema = 'public'
                  AND table_type = 'BASE TABLE'
                  AND table_name NOT LIKE '%audit%'
                ON CONFLICT DO NOTHING;

            WHEN 'user' THEN
                -- Limited access
                INSERT INTO permission_table_policies (
                    organization_id, user_id, table_name, actions, effect, scope
                )
                SELECT
                    v_org_id,
                    v_user.user_id,
                    table_name,
                    ARRAY['select', 'insert', 'update']::permission_action[],
                    'allow'::permission_effect,
                    'self'::permission_scope
                FROM information_schema.tables
                WHERE table_schema = 'public'
                  AND table_type = 'BASE TABLE'
                  AND table_name NOT LIKE 'permission%'
                  AND table_name NOT LIKE '%audit%'
                ON CONFLICT DO NOTHING;

            WHEN 'viewer' THEN
                -- Read-only access
                INSERT INTO permission_table_policies (
                    organization_id, user_id, table_name, actions, effect, scope
                )
                SELECT
                    v_org_id,
                    v_user.user_id,
                    table_name,
                    ARRAY['select']::permission_action[],
                    'allow'::permission_effect,
                    'organization'::permission_scope
                FROM information_schema.tables
                WHERE table_schema = 'public'
                  AND table_type = 'BASE TABLE'
                  AND table_name NOT LIKE 'permission%'
                  AND table_name NOT LIKE '%audit%'
                ON CONFLICT DO NOTHING;
        END CASE;

        v_migrated_count := v_migrated_count + 1;
    END LOOP;

    RETURN jsonb_build_object(
        'success', true,
        'organization_id', v_org_id,
        'users_migrated', v_migrated_count,
        'message', 'Existing permissions migrated to new system'
    );
END;
$$;

COMMENT ON FUNCTION migrate_existing_permissions IS 'Migrate existing role-based permissions to the new granular system';

-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

-- Grant execute permissions on API functions
GRANT EXECUTE ON FUNCTION get_my_permissions(text) TO authenticated;
GRANT EXECUTE ON FUNCTION can_i(text, text, text) TO authenticated;
GRANT EXECUTE ON FUNCTION get_ai_safe_data(text, integer) TO authenticated;
GRANT EXECUTE ON FUNCTION setup_default_permissions(uuid) TO authenticated;

-- Admin-only functions
GRANT EXECUTE ON FUNCTION grant_role_to_user(uuid, uuid, varchar, uuid, text, timestamptz) TO authenticated;
GRANT EXECUTE ON FUNCTION revoke_role_from_user(uuid, uuid, varchar, uuid, text) TO authenticated;
GRANT EXECUTE ON FUNCTION create_role_from_template(uuid, varchar, varchar, uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION bulk_assign_permissions(uuid, jsonb, uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION migrate_existing_permissions(uuid) TO authenticated;

-- =====================================================
-- API EXAMPLES
-- =====================================================

COMMENT ON SCHEMA public IS '
Permission System API Examples:

-- Check what I can do
SELECT get_my_permissions(); -- All permissions
SELECT get_my_permissions(''invoices''); -- Specific table
SELECT can_i(''update'', ''invoices'', ''amount''); -- Specific action

-- Get AI-safe data
SELECT get_ai_safe_data(''products'', 100);
SELECT get_ai_safe_data(''contacts'', 10); -- Will be sanitized

-- Setup permissions (Admin only)
SELECT setup_default_permissions();
SELECT grant_role_to_user(
    org_id, user_id, ''sales_manager'',
    granted_by_id, ''Promoted to sales manager''
);

-- Migrate from old system
SELECT migrate_existing_permissions();

-- Generate RLS policy
SELECT generate_rls_policy(''my_custom_table'');
';
