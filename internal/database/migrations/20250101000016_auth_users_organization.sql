-- Migration: User Metadata Management
-- Description: Helper functions to manage user metadata via auth.users app_metadata
-- Created: 2025-01-01

-- =====================================================
-- ORGANIZATION-SPECIFIC HELPERS
-- =====================================================

-- Helper function to get current user's organization_id from raw_app_metadata
CREATE OR REPLACE FUNCTION public.get_current_user_organization_id()
RETURNS uuid AS $$
BEGIN
    RETURN (
        SELECT (raw_app_metadata->>'organization_id')::uuid
        FROM auth.users
        WHERE id = (SELECT auth.uid())
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.get_current_user_organization_id IS 'Get the organization_id of the currently authenticated user from app metadata';

-- Helper function to check if user belongs to organization
CREATE OR REPLACE FUNCTION public.user_belongs_to_organization(p_organization_id uuid)
RETURNS boolean AS $$
BEGIN
    RETURN (
        SELECT (raw_app_metadata->>'organization_id')::uuid = p_organization_id
        FROM auth.users
        WHERE id = (SELECT auth.uid())
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.user_belongs_to_organization IS 'Check if the current user belongs to the specified organization';

-- =====================================================
-- GENERIC METADATA HELPERS
-- =====================================================

-- Get a specific metadata value for current user
CREATE OR REPLACE FUNCTION public.get_user_metadata(p_key text)
RETURNS text AS $$
BEGIN
    RETURN (
        SELECT raw_app_metadata->>p_key
        FROM auth.users
        WHERE id = (SELECT auth.uid())
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.get_user_metadata IS 'Get a specific metadata value for the current user';

-- Get all metadata for current user
CREATE OR REPLACE FUNCTION public.get_all_user_metadata()
RETURNS jsonb AS $$
BEGIN
    RETURN (
        SELECT raw_app_metadata
        FROM auth.users
        WHERE id = (SELECT auth.uid())
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.get_all_user_metadata IS 'Get all metadata for the current user';

-- Set a single metadata key-value pair for a user (requires service role)
CREATE OR REPLACE FUNCTION public.set_user_metadata(
    p_user_id uuid,
    p_key text,
    p_value text
)
RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET raw_app_metadata =
        COALESCE(raw_app_metadata, '{}'::jsonb) ||
        jsonb_build_object(p_key, p_value)
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.set_user_metadata IS 'Set a single metadata key-value pair (use from backend with service role)';

-- Set multiple metadata key-value pairs for a user (requires service role)
CREATE OR REPLACE FUNCTION public.set_user_metadata_bulk(
    p_user_id uuid,
    p_metadata jsonb
)
RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET raw_app_metadata =
        COALESCE(raw_app_metadata, '{}'::jsonb) || p_metadata
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.set_user_metadata_bulk IS 'Set multiple metadata key-value pairs at once (use from backend with service role)';

-- Remove a metadata key for a user (requires service role)
CREATE OR REPLACE FUNCTION public.remove_user_metadata(
    p_user_id uuid,
    p_key text
)
RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET raw_app_metadata = raw_app_metadata - p_key
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION public.remove_user_metadata IS 'Remove a metadata key (use from backend with service role)';

-- =====================================================
-- USAGE EXAMPLES
-- =====================================================

-- Example 1: Set organization_id for a user (from backend with service role)
-- SELECT public.set_user_metadata(
--     'user-uuid'::uuid,
--     'organization_id',
--     'org-uuid'
-- );

-- Example 2: Set multiple metadata fields at once
-- SELECT public.set_user_metadata_bulk(
--     'user-uuid'::uuid,
--     '{"organization_id": "org-uuid", "tenant_id": "tenant-uuid", "role": "admin"}'::jsonb
-- );

-- Example 3: Get a specific metadata value
-- SELECT public.get_user_metadata('organization_id');

-- Example 4: Get all metadata for current user
-- SELECT public.get_all_user_metadata();

-- Example 5: Use in RLS policies
-- CREATE POLICY "Users see only their org data"
-- ON some_table
-- FOR SELECT
-- USING (organization_id = public.get_current_user_organization_id());

-- Example 6: Check metadata in queries
-- SELECT * FROM some_table
-- WHERE tenant_id = public.get_user_metadata('tenant_id')::uuid;

-- JavaScript/TypeScript Example (using Supabase Admin API):
-- const { data, error } = await supabase.auth.admin.updateUserById(
--   userId,
--   {
--     app_metadata: {
--       organization_id: 'org-uuid',
--       tenant_id: 'tenant-uuid',
--       custom_field: 'custom-value'
--     }
--   }
-- );
