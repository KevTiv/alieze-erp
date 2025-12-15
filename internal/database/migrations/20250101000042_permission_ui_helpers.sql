-- Migration: Permission System - UI Helpers & Dashboard Views
-- Description: Views and functions to support permission management UI
-- Created: 2025-01-01
-- Module: UI Helpers

-- =====================================================
-- DASHBOARD VIEWS
-- =====================================================

-- User permission matrix view
CREATE OR REPLACE VIEW permission_matrix AS
WITH user_permissions AS (
    SELECT
        ou.organization_id,
        ou.user_id,
        u.email as user_email,
        ou.role as base_role,
        array_agg(DISTINCT pr.display_name) FILTER (WHERE pr.id IS NOT NULL) as custom_roles,
        count(DISTINCT ptp.id) as table_policies_count,
        count(DISTINCT pcp.id) as column_policies_count,
        count(DISTINCT prf.id) as row_filters_count,
        count(DISTINCT adp.id) as ai_permissions_count
    FROM organization_users ou
    LEFT JOIN auth.users u ON u.id = ou.user_id
    LEFT JOIN user_role_assignments ura ON ura.user_id = ou.user_id
        AND ura.organization_id = ou.organization_id
        AND ura.is_active = true
    LEFT JOIN permission_roles pr ON pr.id = ura.role_id
    LEFT JOIN permission_table_policies ptp ON ptp.user_id = ou.user_id
        AND ptp.organization_id = ou.organization_id
        AND ptp.is_active = true
    LEFT JOIN permission_column_policies pcp ON pcp.user_id = ou.user_id
        AND pcp.organization_id = ou.organization_id
        AND pcp.is_active = true
    LEFT JOIN permission_row_filters prf ON prf.user_id = ou.user_id
        AND prf.organization_id = ou.organization_id
        AND prf.is_active = true
    LEFT JOIN ai_data_permissions adp ON adp.user_id = ou.user_id
        AND adp.organization_id = ou.organization_id
        AND adp.is_active = true
    WHERE ou.is_active = true
    GROUP BY ou.organization_id, ou.user_id, u.email, ou.role
)
SELECT
    organization_id,
    user_id,
    user_email,
    base_role,
    custom_roles,
    table_policies_count,
    column_policies_count,
    row_filters_count,
    ai_permissions_count,
    (table_policies_count + column_policies_count + row_filters_count + ai_permissions_count) as total_custom_policies
FROM user_permissions
ORDER BY base_role DESC, total_custom_policies DESC;

COMMENT ON VIEW permission_matrix IS 'Overview of all users and their permission counts for UI dashboards';

-- Table access summary view
CREATE OR REPLACE VIEW table_access_summary AS
WITH action_aggregation AS (
    SELECT
        ptp.organization_id,
        ptp.table_name,
        ptp.user_id,
        ptp.role_id,
        ptp.group_id,
        action
    FROM permission_table_policies ptp
    CROSS JOIN LATERAL unnest(ptp.actions) AS action
    WHERE ptp.is_active = true
)
SELECT
    aa.organization_id,
    aa.table_name,
    COUNT(DISTINCT aa.user_id) FILTER (WHERE aa.user_id IS NOT NULL) as direct_user_access,
    COUNT(DISTINCT aa.role_id) FILTER (WHERE aa.role_id IS NOT NULL) as role_based_access,
    COUNT(DISTINCT aa.group_id) FILTER (WHERE aa.group_id IS NOT NULL) as group_based_access,
    array_agg(DISTINCT aa.action) as all_actions,
    COUNT(DISTINCT pcp.column_name) as restricted_columns,
    COUNT(DISTINCT prf.id) as row_filters,
    MAX(adp.exposure_level) as max_ai_exposure
FROM action_aggregation aa
LEFT JOIN permission_column_policies pcp ON pcp.table_name = aa.table_name
    AND pcp.organization_id = aa.organization_id
    AND pcp.is_active = true
LEFT JOIN permission_row_filters prf ON prf.table_name = aa.table_name
    AND prf.organization_id = aa.organization_id
    AND prf.is_active = true
LEFT JOIN ai_data_permissions adp ON adp.table_name = aa.table_name
    AND adp.organization_id = aa.organization_id
    AND adp.is_active = true
GROUP BY aa.organization_id, aa.table_name
ORDER BY aa.table_name;

COMMENT ON VIEW table_access_summary IS 'Summary of access controls per table for UI display';

-- Role hierarchy view
CREATE OR REPLACE VIEW role_hierarchy AS
WITH RECURSIVE hierarchy AS (
    -- Base roles
    SELECT
        pr.id as role_id,
        pr.organization_id,
        pr.name as role_name,
        pr.display_name,
        pr.base_role,
        pr.priority,
        0 as depth,
        ARRAY[pr.id] as path,
        pr.id::text as hierarchy_key
    FROM permission_roles pr
    WHERE pr.is_active = true
      AND NOT EXISTS (
          SELECT 1 FROM permission_role_inheritance pri
          WHERE pri.child_role_id = pr.id
      )

    UNION ALL

    -- Recursive part: children
    SELECT
        pr.id as role_id,
        pr.organization_id,
        pr.name as role_name,
        pr.display_name,
        pr.base_role,
        pr.priority,
        h.depth + 1 as depth,
        h.path || pr.id as path,
        h.hierarchy_key || '.' || pr.id::text as hierarchy_key
    FROM permission_roles pr
    INNER JOIN permission_role_inheritance pri ON pri.parent_role_id = pr.id
    INNER JOIN hierarchy h ON h.role_id = pri.child_role_id
    WHERE pr.is_active = true
)
SELECT
    role_id,
    organization_id,
    role_name,
    display_name,
    base_role,
    priority,
    depth,
    path,
    hierarchy_key,
    repeat('  ', depth) || display_name as indented_name
FROM hierarchy
ORDER BY hierarchy_key;

COMMENT ON VIEW role_hierarchy IS 'Hierarchical view of roles with inheritance for tree display in UI';

-- AI permission status view
CREATE OR REPLACE VIEW ai_permission_status AS
SELECT
    adp.organization_id,
    adp.table_name,
    adp.column_name,
    adp.exposure_level,
    adp.allowed_providers,
    adp.requires_sanitization,
    adp.sanitization_method,
    COUNT(DISTINCT asr.id) as sanitization_rules_count,
    COUNT(DISTINCT acr.id) as context_rules_count,
    CASE
        WHEN adp.exposure_level = 'none' THEN 'Blocked'
        WHEN adp.exposure_level = 'metadata' THEN 'Metadata Only'
        WHEN adp.exposure_level = 'aggregated' THEN 'Aggregated Only'
        WHEN adp.exposure_level = 'sanitized' THEN 'Sanitized'
        WHEN adp.exposure_level = 'full' THEN 'Full Access'
        ELSE 'Unknown'
    END as status_label,
    CASE
        WHEN adp.exposure_level = 'none' THEN 'red'
        WHEN adp.exposure_level = 'metadata' THEN 'orange'
        WHEN adp.exposure_level = 'aggregated' THEN 'yellow'
        WHEN adp.exposure_level = 'sanitized' THEN 'blue'
        WHEN adp.exposure_level = 'full' THEN 'green'
        ELSE 'gray'
    END as status_color
FROM ai_data_permissions adp
LEFT JOIN ai_sanitization_rules asr ON asr.table_name = adp.table_name
    AND asr.organization_id = adp.organization_id
    AND (asr.column_name IS NULL OR asr.column_name = adp.column_name)
    AND asr.is_active = true
LEFT JOIN ai_context_rules acr ON adp.table_name = ANY(acr.included_tables)
    AND acr.organization_id = adp.organization_id
    AND acr.is_active = true
WHERE adp.is_active = true
GROUP BY adp.id, adp.organization_id, adp.table_name, adp.column_name,
         adp.exposure_level, adp.allowed_providers, adp.requires_sanitization,
         adp.sanitization_method
ORDER BY adp.table_name, adp.column_name;

COMMENT ON VIEW ai_permission_status IS 'AI permission status with UI-friendly labels and colors';

-- =====================================================
-- UI HELPER FUNCTIONS
-- =====================================================

-- Get permission templates for UI dropdown
CREATE OR REPLACE FUNCTION get_permission_templates(
    p_category varchar DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN jsonb_agg(
        jsonb_build_object(
            'id', id,
            'code', code,
            'name', name,
            'category', category,
            'description', description,
            'base_role', base_role,
            'permissions_count', jsonb_array_length(permissions)
        ) ORDER BY category, name
    )
    FROM permission_role_templates
    WHERE is_active = true
      AND (p_category IS NULL OR category = p_category);
END;
$$;

COMMENT ON FUNCTION get_permission_templates IS 'Get available role templates for UI selection';

-- Get tables for permission assignment UI
CREATE OR REPLACE FUNCTION get_assignable_tables()
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN jsonb_agg(
        jsonb_build_object(
            'table_name', table_name,
            'table_type', table_type,
            'columns', (
                SELECT jsonb_agg(column_name ORDER BY ordinal_position)
                FROM information_schema.columns
                WHERE table_name = t.table_name
                  AND table_schema = t.table_schema
            ),
            'has_org_id', EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_name = t.table_name
                  AND table_schema = t.table_schema
                  AND column_name = 'organization_id'
            ),
            'has_created_by', EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_name = t.table_name
                  AND table_schema = t.table_schema
                  AND column_name = 'created_by'
            )
        ) ORDER BY table_name
    )
    FROM information_schema.tables t
    WHERE table_schema = 'public'
      AND table_type = 'BASE TABLE'
      AND table_name NOT LIKE 'permission%'
      AND table_name NOT LIKE '%audit%';
END;
$$;

COMMENT ON FUNCTION get_assignable_tables IS 'Get list of tables available for permission assignment';

-- Search users for permission assignment
CREATE OR REPLACE FUNCTION search_users_for_permissions(
    p_search text,
    p_organization_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_org_id uuid;
BEGIN
    v_org_id := COALESCE(p_organization_id, (SELECT get_current_user_organization_id()));

    RETURN jsonb_agg(
        jsonb_build_object(
            'user_id', ou.user_id,
            'email', u.email,
            'full_name', u.raw_user_meta_data->>'full_name',
            'base_role', ou.role,
            'custom_roles', (
                SELECT jsonb_agg(pr.display_name)
                FROM user_role_assignments ura
                JOIN permission_roles pr ON pr.id = ura.role_id
                WHERE ura.user_id = ou.user_id
                  AND ura.organization_id = v_org_id
                  AND ura.is_active = true
            ),
            'last_seen', u.last_sign_in_at
        ) ORDER BY u.email
    )
    FROM organization_users ou
    JOIN auth.users u ON u.id = ou.user_id
    WHERE ou.organization_id = v_org_id
      AND ou.is_active = true
      AND (
          p_search IS NULL OR
          u.email ILIKE '%' || p_search || '%' OR
          u.raw_user_meta_data->>'full_name' ILIKE '%' || p_search || '%'
      );
END;
$$;

COMMENT ON FUNCTION search_users_for_permissions IS 'Search users for permission assignment UI';

-- Get permission changes for activity feed
CREATE OR REPLACE FUNCTION get_permission_activity(
    p_organization_id uuid DEFAULT NULL,
    p_limit integer DEFAULT 50
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_org_id uuid;
BEGIN
    v_org_id := COALESCE(p_organization_id, (SELECT get_current_user_organization_id()));

    RETURN jsonb_agg(
        jsonb_build_object(
            'id', pal.id,
            'action', pal.action,
            'target_type', pal.target_type,
            'target_name', pal.target_name,
            'user_email', u.email,
            'reason', pal.reason,
            'changed_fields', pal.changed_fields,
            'created_at', pal.created_at,
            'time_ago', (
                CASE
                    WHEN pal.created_at > now() - interval '1 minute' THEN 'Just now'
                    WHEN pal.created_at > now() - interval '1 hour' THEN
                        extract(minute from now() - pal.created_at) || ' minutes ago'
                    WHEN pal.created_at > now() - interval '1 day' THEN
                        extract(hour from now() - pal.created_at) || ' hours ago'
                    WHEN pal.created_at > now() - interval '1 week' THEN
                        extract(day from now() - pal.created_at) || ' days ago'
                    ELSE to_char(pal.created_at, 'Mon DD, YYYY')
                END
            ),
            'action_icon', (
                CASE pal.action
                    WHEN 'create' THEN 'plus-circle'
                    WHEN 'update' THEN 'edit'
                    WHEN 'delete' THEN 'trash'
                    WHEN 'grant' THEN 'user-plus'
                    WHEN 'revoke' THEN 'user-minus'
                    ELSE 'activity'
                END
            ),
            'action_color', (
                CASE pal.action
                    WHEN 'create' THEN 'green'
                    WHEN 'update' THEN 'blue'
                    WHEN 'delete' THEN 'red'
                    WHEN 'grant' THEN 'teal'
                    WHEN 'revoke' THEN 'orange'
                    ELSE 'gray'
                END
            )
        ) ORDER BY pal.created_at DESC
    )
    FROM permission_audit_log pal
    LEFT JOIN auth.users u ON u.id = pal.user_id
    WHERE pal.organization_id = v_org_id
    ORDER BY pal.created_at DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION get_permission_activity IS 'Get recent permission changes for activity feed';

-- Validate permission configuration
CREATE OR REPLACE FUNCTION validate_permissions(
    p_organization_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_org_id uuid;
    v_issues jsonb := '[]'::jsonb;
    v_warnings jsonb := '[]'::jsonb;
BEGIN
    v_org_id := COALESCE(p_organization_id, (SELECT get_current_user_organization_id()));

    -- Check for users with no roles
    IF EXISTS (
        SELECT 1 FROM organization_users
        WHERE organization_id = v_org_id
          AND is_active = true
          AND role IS NULL
    ) THEN
        v_issues := v_issues || jsonb_build_object(
            'type', 'error',
            'message', 'Some users have no base role assigned'
        );
    END IF;

    -- Check for circular role inheritance
    IF EXISTS (
        WITH RECURSIVE check_circular AS (
            SELECT child_role_id, parent_role_id, ARRAY[child_role_id] as path
            FROM permission_role_inheritance
            WHERE organization_id = v_org_id
            UNION ALL
            SELECT pri.child_role_id, pri.parent_role_id, cc.path || pri.child_role_id
            FROM permission_role_inheritance pri
            JOIN check_circular cc ON cc.parent_role_id = pri.child_role_id
            WHERE pri.child_role_id = ANY(cc.path)
        )
        SELECT 1 FROM check_circular
    ) THEN
        v_issues := v_issues || jsonb_build_object(
            'type', 'error',
            'message', 'Circular role inheritance detected'
        );
    END IF;

    -- Check for conflicting deny/allow policies
    IF EXISTS (
        SELECT 1
        FROM permission_table_policies p1
        JOIN permission_table_policies p2 ON
            p1.organization_id = p2.organization_id
            AND p1.table_name = p2.table_name
            AND p1.user_id = p2.user_id
        WHERE p1.organization_id = v_org_id
          AND p1.effect = 'allow'
          AND p2.effect = 'deny'
          AND p1.priority = p2.priority
    ) THEN
        v_warnings := v_warnings || jsonb_build_object(
            'type', 'warning',
            'message', 'Conflicting allow/deny policies with same priority detected'
        );
    END IF;

    -- Check for tables with no AI permissions
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables t
        WHERE t.table_schema = 'public'
          AND t.table_type = 'BASE TABLE'
          AND NOT EXISTS (
              SELECT 1 FROM ai_data_permissions adp
              WHERE adp.organization_id = v_org_id
                AND adp.table_name = t.table_name
          )
    ) THEN
        v_warnings := v_warnings || jsonb_build_object(
            'type', 'warning',
            'message', 'Some tables have no AI permission configuration'
        );
    END IF;

    RETURN jsonb_build_object(
        'valid', jsonb_array_length(v_issues) = 0,
        'issues', v_issues,
        'warnings', v_warnings,
        'checked_at', now()
    );
END;
$$;

COMMENT ON FUNCTION validate_permissions IS 'Validate permission configuration and detect issues';

-- =====================================================
-- PERMISSION COMPARISON
-- =====================================================

-- Compare permissions between two users
CREATE OR REPLACE FUNCTION compare_user_permissions(
    p_user1_id uuid,
    p_user2_id uuid,
    p_organization_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_org_id uuid;
    v_user1_perms jsonb;
    v_user2_perms jsonb;
BEGIN
    v_org_id := COALESCE(p_organization_id, (SELECT get_current_user_organization_id()));

    -- Get permissions for user 1
    SELECT jsonb_build_object(
        'user_id', p_user1_id,
        'email', u.email,
        'base_role', ou.role,
        'custom_roles', (
            SELECT jsonb_agg(pr.name)
            FROM user_role_assignments ura
            JOIN permission_roles pr ON pr.id = ura.role_id
            WHERE ura.user_id = p_user1_id
              AND ura.organization_id = v_org_id
              AND ura.is_active = true
        ),
        'table_permissions', (
            SELECT jsonb_object_agg(table_name, allowed_actions)
            FROM user_effective_permissions
            WHERE user_id = p_user1_id
              AND organization_id = v_org_id
        )
    ) INTO v_user1_perms
    FROM organization_users ou
    LEFT JOIN auth.users u ON u.id = ou.user_id
    WHERE ou.user_id = p_user1_id
      AND ou.organization_id = v_org_id;

    -- Get permissions for user 2
    SELECT jsonb_build_object(
        'user_id', p_user2_id,
        'email', u.email,
        'base_role', ou.role,
        'custom_roles', (
            SELECT jsonb_agg(pr.name)
            FROM user_role_assignments ura
            JOIN permission_roles pr ON pr.id = ura.role_id
            WHERE ura.user_id = p_user2_id
              AND ura.organization_id = v_org_id
              AND ura.is_active = true
        ),
        'table_permissions', (
            SELECT jsonb_object_agg(table_name, allowed_actions)
            FROM user_effective_permissions
            WHERE user_id = p_user2_id
              AND organization_id = v_org_id
        )
    ) INTO v_user2_perms
    FROM organization_users ou
    LEFT JOIN auth.users u ON u.id = ou.user_id
    WHERE ou.user_id = p_user2_id
      AND ou.organization_id = v_org_id;

    RETURN jsonb_build_object(
        'user1', v_user1_perms,
        'user2', v_user2_perms,
        'comparison', jsonb_build_object(
            'same_base_role', v_user1_perms->>'base_role' = v_user2_perms->>'base_role',
            'tables_in_common', (
                SELECT COUNT(*)
                FROM jsonb_object_keys(v_user1_perms->'table_permissions') k1
                WHERE k1 IN (
                    SELECT jsonb_object_keys(v_user2_perms->'table_permissions')
                )
            ),
            'unique_to_user1', (
                SELECT jsonb_agg(key)
                FROM jsonb_object_keys(v_user1_perms->'table_permissions') key
                WHERE key NOT IN (
                    SELECT jsonb_object_keys(v_user2_perms->'table_permissions')
                )
            ),
            'unique_to_user2', (
                SELECT jsonb_agg(key)
                FROM jsonb_object_keys(v_user2_perms->'table_permissions') key
                WHERE key NOT IN (
                    SELECT jsonb_object_keys(v_user1_perms->'table_permissions')
                )
            )
        )
    );
END;
$$;

COMMENT ON FUNCTION compare_user_permissions IS 'Compare permissions between two users for UI display';

-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

GRANT SELECT ON permission_matrix TO authenticated;
GRANT SELECT ON table_access_summary TO authenticated;
GRANT SELECT ON role_hierarchy TO authenticated;
GRANT SELECT ON ai_permission_status TO authenticated;

GRANT EXECUTE ON FUNCTION get_permission_templates(varchar) TO authenticated;
GRANT EXECUTE ON FUNCTION get_assignable_tables() TO authenticated;
GRANT EXECUTE ON FUNCTION search_users_for_permissions(text, uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION get_permission_activity(uuid, integer) TO authenticated;
GRANT EXECUTE ON FUNCTION validate_permissions(uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION compare_user_permissions(uuid, uuid, uuid) TO authenticated;
