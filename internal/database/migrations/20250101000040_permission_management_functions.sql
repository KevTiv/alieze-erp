-- Migration: Permission System - Management Functions & Audit
-- Description: Helper functions, views, and audit capabilities for the permission system
-- Created: 2025-01-01
-- Module: Permission Management

-- =====================================================
-- AUDIT LOG
-- =====================================================

-- Permission change audit log
CREATE TABLE permission_audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,

    -- Change details
    action varchar(50) NOT NULL, -- 'create', 'update', 'delete', 'grant', 'revoke'
    target_type varchar(50) NOT NULL, -- 'role', 'permission', 'assignment', etc.
    target_id uuid,
    target_name text,

    -- What changed
    old_values jsonb,
    new_values jsonb,
    changed_fields text[],

    -- Context
    reason text,
    ip_address inet,
    user_agent text,
    session_id text,

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_permission_audit_log_org ON permission_audit_log(organization_id, created_at DESC);
CREATE INDEX idx_permission_audit_log_user ON permission_audit_log(user_id, created_at DESC);
CREATE INDEX idx_permission_audit_log_target ON permission_audit_log(target_type, target_id);
CREATE INDEX idx_permission_audit_log_action ON permission_audit_log(action, created_at DESC);

COMMENT ON TABLE permission_audit_log IS 'Comprehensive audit trail of all permission-related changes';

-- =====================================================
-- PERMISSION VIEWS
-- =====================================================

-- Consolidated view of user permissions
CREATE OR REPLACE VIEW user_effective_permissions AS
WITH user_roles AS (
    -- Get all roles (base + custom) for each user
    SELECT DISTINCT
        ou.user_id,
        ou.organization_id,
        ou.role AS base_role,
        pr.id AS custom_role_id,
        pr.name AS custom_role_name,
        COALESCE(pr.priority,
            CASE ou.role
                WHEN 'owner' THEN 1000
                WHEN 'admin' THEN 900
                WHEN 'manager' THEN 500
                WHEN 'user' THEN 100
                WHEN 'viewer' THEN 50
            END
        ) AS priority
    FROM organization_users ou
    LEFT JOIN user_role_assignments ura ON ura.user_id = ou.user_id
        AND ura.organization_id = ou.organization_id
        AND ura.is_active = true
        AND (ura.valid_from IS NULL OR ura.valid_from <= now())
        AND (ura.valid_until IS NULL OR ura.valid_until > now())
    LEFT JOIN permission_roles pr ON pr.id = ura.role_id AND pr.is_active = true
    WHERE ou.is_active = true
),
table_perms AS (
    -- Aggregate table permissions
    SELECT
        ur.user_id,
        ur.organization_id,
        ptp.table_name,
        ptp.actions,
        ptp.effect,
        ptp.scope,
        MAX(ptp.priority) AS priority
    FROM user_roles ur
    LEFT JOIN permission_table_policies ptp ON
        (ptp.role_id = ur.custom_role_id OR ptp.user_id = ur.user_id)
        AND ptp.organization_id = ur.organization_id
        AND ptp.is_active = true
    WHERE ptp.id IS NOT NULL
    GROUP BY ur.user_id, ur.organization_id, ptp.table_name, ptp.actions, ptp.effect, ptp.scope
)
SELECT
    user_id,
    organization_id,
    table_name,
    array_agg(DISTINCT action) AS allowed_actions,
    scope,
    MAX(priority) AS effective_priority
FROM table_perms tp
CROSS JOIN LATERAL unnest(tp.actions) AS action
WHERE effect = 'allow'
  AND NOT EXISTS (
      -- Check for explicit deny
      SELECT 1 FROM table_perms tp2
      WHERE tp2.user_id = tp.user_id
        AND tp2.organization_id = tp.organization_id
        AND tp2.table_name = tp.table_name
        AND tp2.effect = 'deny'
        AND tp2.priority >= tp.priority
  )
GROUP BY user_id, organization_id, table_name, scope;

COMMENT ON VIEW user_effective_permissions IS 'Consolidated view of effective permissions for each user';

-- View of AI data access permissions
CREATE OR REPLACE VIEW user_ai_permissions AS
SELECT DISTINCT ON (ou.user_id, ou.organization_id, adp.table_name, adp.column_name)
    ou.user_id,
    ou.organization_id,
    adp.table_name,
    adp.column_name,
    adp.exposure_level,
    adp.allowed_providers,
    adp.requires_sanitization,
    adp.sanitization_method,
    adp.include_in_context,
    adp.max_rows_in_context
FROM organization_users ou
LEFT JOIN user_role_assignments ura ON ura.user_id = ou.user_id
    AND ura.organization_id = ou.organization_id
    AND ura.is_active = true
LEFT JOIN ai_data_permissions adp ON
    (adp.user_id = ou.user_id OR
     adp.role_id = ura.role_id OR
     adp.role_id IN (SELECT pr.id FROM permission_roles pr
                      WHERE pr.organization_id = ou.organization_id
                      AND pr.base_role = ou.role))
    AND adp.organization_id = ou.organization_id
    AND adp.is_active = true
WHERE ou.is_active = true
ORDER BY ou.user_id, ou.organization_id, adp.table_name, adp.column_name, adp.priority DESC;

COMMENT ON VIEW user_ai_permissions IS 'View of effective AI data access permissions for users';

-- =====================================================
-- PERMISSION MANAGEMENT FUNCTIONS
-- =====================================================

-- Grant role to user
CREATE OR REPLACE FUNCTION grant_role_to_user(
    p_organization_id uuid,
    p_user_id uuid,
    p_role_name varchar,
    p_granted_by uuid,
    p_reason text DEFAULT NULL,
    p_valid_until timestamptz DEFAULT NULL
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
DECLARE
    v_role_id uuid;
    v_assignment_id uuid;
BEGIN
    -- Find the role
    SELECT id INTO v_role_id
    FROM permission_roles
    WHERE organization_id = p_organization_id
      AND name = p_role_name
      AND is_active = true;

    IF v_role_id IS NULL THEN
        RAISE EXCEPTION 'Role % not found in organization', p_role_name;
    END IF;

    -- Create assignment
    INSERT INTO user_role_assignments (
        organization_id, user_id, role_id, assigned_by,
        assignment_reason, valid_until
    ) VALUES (
        p_organization_id, p_user_id, v_role_id, p_granted_by,
        p_reason, p_valid_until
    ) RETURNING id INTO v_assignment_id;

    -- Audit log
    INSERT INTO permission_audit_log (
        organization_id, user_id, action, target_type,
        target_id, target_name, new_values, reason
    ) VALUES (
        p_organization_id, p_granted_by, 'grant', 'role',
        v_role_id, p_role_name,
        jsonb_build_object(
            'user_id', p_user_id,
            'role_id', v_role_id,
            'valid_until', p_valid_until
        ),
        p_reason
    );

    RETURN v_assignment_id;
END;
$$;

COMMENT ON FUNCTION grant_role_to_user IS 'Grant a custom role to a user with audit logging';

-- Revoke role from user
CREATE OR REPLACE FUNCTION revoke_role_from_user(
    p_organization_id uuid,
    p_user_id uuid,
    p_role_name varchar,
    p_revoked_by uuid,
    p_reason text DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_role_id uuid;
    v_assignment_id uuid;
BEGIN
    -- Find the role
    SELECT id INTO v_role_id
    FROM permission_roles
    WHERE organization_id = p_organization_id
      AND name = p_role_name;

    IF v_role_id IS NULL THEN
        RETURN false;
    END IF;

    -- Find and deactivate assignment
    UPDATE user_role_assignments
    SET is_active = false,
        updated_at = now()
    WHERE organization_id = p_organization_id
      AND user_id = p_user_id
      AND role_id = v_role_id
      AND is_active = true
    RETURNING id INTO v_assignment_id;

    IF v_assignment_id IS NULL THEN
        RETURN false;
    END IF;

    -- Audit log
    INSERT INTO permission_audit_log (
        organization_id, user_id, action, target_type,
        target_id, target_name, old_values, reason
    ) VALUES (
        p_organization_id, p_revoked_by, 'revoke', 'role',
        v_role_id, p_role_name,
        jsonb_build_object(
            'user_id', p_user_id,
            'role_id', v_role_id,
            'assignment_id', v_assignment_id
        ),
        p_reason
    );

    RETURN true;
END;
$$;

COMMENT ON FUNCTION revoke_role_from_user IS 'Revoke a custom role from a user with audit logging';

-- Create role from template
CREATE OR REPLACE FUNCTION create_role_from_template(
    p_organization_id uuid,
    p_template_code varchar,
    p_role_name varchar,
    p_created_by uuid
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
DECLARE
    v_template record;
    v_role_id uuid;
    v_permission record;
BEGIN
    -- Get template
    SELECT * INTO v_template
    FROM permission_role_templates
    WHERE code = p_template_code
      AND is_active = true;

    IF v_template IS NULL THEN
        RAISE EXCEPTION 'Template % not found', p_template_code;
    END IF;

    -- Create role
    INSERT INTO permission_roles (
        organization_id, name, display_name, description,
        base_role, created_by
    ) VALUES (
        p_organization_id, p_role_name, v_template.name,
        v_template.description, v_template.base_role, p_created_by
    ) RETURNING id INTO v_role_id;

    -- Create permissions from template
    FOR v_permission IN
        SELECT * FROM jsonb_array_elements(v_template.permissions)
    LOOP
        INSERT INTO permission_table_policies (
            organization_id, role_id, table_name, actions, scope
        ) VALUES (
            p_organization_id, v_role_id,
            v_permission->>'resource',
            ARRAY(SELECT jsonb_array_elements_text(v_permission->'actions'))::permission_action[],
            (v_permission->>'scope')::permission_scope
        );
    END LOOP;

    -- Audit log
    INSERT INTO permission_audit_log (
        organization_id, user_id, action, target_type,
        target_id, target_name, new_values
    ) VALUES (
        p_organization_id, p_created_by, 'create', 'role',
        v_role_id, p_role_name,
        jsonb_build_object(
            'template_code', p_template_code,
            'permissions', v_template.permissions
        )
    );

    RETURN v_role_id;
END;
$$;

COMMENT ON FUNCTION create_role_from_template IS 'Create a new role from a predefined template';

-- Check user access to specific data
CREATE OR REPLACE FUNCTION check_user_access(
    p_user_id uuid,
    p_organization_id uuid,
    p_table_name text,
    p_action permission_action,
    p_record_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_has_table_permission boolean;
    v_row_filter text;
    v_column_permissions jsonb;
    v_ai_permission jsonb;
    v_result jsonb;
BEGIN
    -- Check table permission
    v_has_table_permission := check_table_permission(
        p_user_id, p_organization_id, p_table_name, p_action
    );

    IF NOT v_has_table_permission THEN
        RETURN jsonb_build_object(
            'allowed', false,
            'reason', 'No table permission'
        );
    END IF;

    -- Get row filter
    v_row_filter := build_row_filter(
        p_user_id, p_organization_id, p_table_name, p_action
    );

    -- Get column permissions
    SELECT jsonb_agg(
        jsonb_build_object(
            'column', column_name,
            'can_view', can_view,
            'can_read', can_read,
            'can_write', can_write,
            'mask_type', mask_type
        )
    ) INTO v_column_permissions
    FROM get_column_permissions(
        p_user_id, p_organization_id, p_table_name
    );

    -- Get AI permissions
    v_ai_permission := check_ai_data_permission(
        p_user_id, p_organization_id, p_table_name
    );

    -- Build result
    v_result := jsonb_build_object(
        'allowed', true,
        'table_permission', true,
        'row_filter', v_row_filter,
        'column_permissions', COALESCE(v_column_permissions, '[]'::jsonb),
        'ai_permission', v_ai_permission
    );

    RETURN v_result;
END;
$$;

COMMENT ON FUNCTION check_user_access IS 'Comprehensive access check for user on specific data';

-- Bulk permission assignment
CREATE OR REPLACE FUNCTION bulk_assign_permissions(
    p_organization_id uuid,
    p_assignments jsonb,
    p_assigned_by uuid
)
RETURNS integer
LANGUAGE plpgsql
AS $$
DECLARE
    v_assignment jsonb;
    v_count integer := 0;
BEGIN
    FOR v_assignment IN SELECT * FROM jsonb_array_elements(p_assignments)
    LOOP
        -- Assign based on type
        IF v_assignment->>'type' = 'role' THEN
            PERFORM grant_role_to_user(
                p_organization_id,
                (v_assignment->>'user_id')::uuid,
                v_assignment->>'role_name',
                p_assigned_by,
                v_assignment->>'reason'
            );
            v_count := v_count + 1;
        ELSIF v_assignment->>'type' = 'table' THEN
            INSERT INTO permission_table_policies (
                organization_id,
                user_id,
                table_name,
                actions,
                effect,
                scope
            ) VALUES (
                p_organization_id,
                (v_assignment->>'user_id')::uuid,
                v_assignment->>'table_name',
                ARRAY(SELECT jsonb_array_elements_text(v_assignment->'actions'))::permission_action[],
                COALESCE((v_assignment->>'effect')::permission_effect, 'allow'),
                COALESCE((v_assignment->>'scope')::permission_scope, 'organization')
            );
            v_count := v_count + 1;
        END IF;
    END LOOP;

    RETURN v_count;
END;
$$;

COMMENT ON FUNCTION bulk_assign_permissions IS 'Assign multiple permissions in a single transaction';

-- Get permission summary for organization
CREATE OR REPLACE FUNCTION get_permission_summary(
    p_organization_id uuid
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_summary jsonb;
BEGIN
    v_summary := jsonb_build_object(
        'total_users', (
            SELECT COUNT(*) FROM organization_users
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'custom_roles', (
            SELECT COUNT(*) FROM permission_roles
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'role_assignments', (
            SELECT COUNT(*) FROM user_role_assignments
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'table_policies', (
            SELECT COUNT(*) FROM permission_table_policies
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'column_policies', (
            SELECT COUNT(*) FROM permission_column_policies
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'row_filters', (
            SELECT COUNT(*) FROM permission_row_filters
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'ai_permissions', (
            SELECT COUNT(*) FROM ai_data_permissions
            WHERE organization_id = p_organization_id AND is_active = true
        ),
        'recent_changes', (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'action', action,
                    'target_type', target_type,
                    'target_name', target_name,
                    'user_id', user_id,
                    'created_at', created_at
                ) ORDER BY created_at DESC
            ) FROM (
                SELECT * FROM permission_audit_log
                WHERE organization_id = p_organization_id
                ORDER BY created_at DESC
                LIMIT 10
            ) recent
        )
    );

    RETURN v_summary;
END;
$$;

COMMENT ON FUNCTION get_permission_summary IS 'Get comprehensive permission system summary for organization';

-- =====================================================
-- CLEANUP FUNCTIONS
-- =====================================================

-- Clean up expired permissions
CREATE OR REPLACE FUNCTION cleanup_expired_permissions()
RETURNS integer
LANGUAGE plpgsql
AS $$
DECLARE
    v_count integer := 0;
BEGIN
    -- Deactivate expired role assignments
    UPDATE user_role_assignments
    SET is_active = false
    WHERE is_active = true
      AND valid_until IS NOT NULL
      AND valid_until < now();

    GET DIAGNOSTICS v_count = ROW_COUNT;

    -- Clean up expired table policies
    UPDATE permission_table_policies
    SET is_active = false
    WHERE is_active = true
      AND valid_until IS NOT NULL
      AND valid_until < now();

    RETURN v_count;
END;
$$;

COMMENT ON FUNCTION cleanup_expired_permissions IS 'Clean up expired temporal permissions';

-- =====================================================
-- RLS POLICIES
-- =====================================================

ALTER TABLE permission_audit_log ENABLE ROW LEVEL SECURITY;

-- Audit log is viewable by admins only
CREATE POLICY "Admins can view permission audit log"
    ON permission_audit_log FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Additional performance indexes
CREATE INDEX idx_permission_audit_created ON permission_audit_log(created_at DESC);
CREATE INDEX idx_user_role_assignments_valid ON user_role_assignments(valid_until)
    WHERE is_active = true AND valid_until IS NOT NULL;

-- =====================================================
-- SCHEDULED JOBS (To be run via pg_cron or external scheduler)
-- =====================================================

COMMENT ON FUNCTION cleanup_expired_permissions IS
'Run this function periodically (e.g., daily) to clean up expired permissions:
SELECT cron.schedule(
    ''cleanup-expired-permissions'',
    ''0 2 * * *'', -- Run at 2 AM daily
    $$SELECT cleanup_expired_permissions();$$
);';
