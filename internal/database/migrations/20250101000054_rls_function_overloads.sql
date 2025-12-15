-- Migration: RLS Function Overloads
-- Description: Add overloaded versions of user_has_org_access and user_has_role for RLS policies
-- Created: 2025-01-26
-- Dependencies: 20250101000053_pos_rls_policies.sql

-- =====================================================
-- CREATE OVERLOADED user_has_org_access FUNCTION
-- =====================================================

-- Create overloaded version that accepts organization_id parameter
-- This is needed for RLS policies that want to check access to a specific organization
CREATE OR REPLACE FUNCTION public.user_has_org_access(p_organization_id uuid)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path TO ''
AS $function$
DECLARE
    v_user_id uuid;
BEGIN
    -- Get current authenticated user
    v_user_id := (SELECT auth.uid());

    -- Return false if no user is authenticated or no organization provided
    IF v_user_id IS NULL OR p_organization_id IS NULL THEN
        RETURN false;
    END IF;

    -- Check if user has active membership in the specified organization
    RETURN EXISTS (
        SELECT 1
        FROM public.organization_users
        WHERE user_id = v_user_id
          AND organization_id = p_organization_id
          AND is_active = true
    );
END;
$function$;

-- =====================================================
-- CREATE OVERLOADED user_has_role FUNCTION
-- =====================================================

-- Create overloaded version that only takes role_name and uses current user context
-- This is needed for RLS policies that want to check roles without specifying user/org
CREATE OR REPLACE FUNCTION public.user_has_role(p_role_name character varying)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path TO ''
AS $function$
DECLARE
    v_user_id uuid;
    v_org_id uuid;
BEGIN
    -- Get current authenticated user
    v_user_id := (SELECT auth.uid());

    -- Get current organization
    v_org_id := get_current_organization_id();

    -- Return false if no user is authenticated or no organization context
    IF v_user_id IS NULL OR v_org_id IS NULL THEN
        RETURN false;
    END IF;

    -- Use the existing 3-parameter function
    RETURN user_has_role(v_user_id, v_org_id, p_role_name);
END;
$function$;

-- =====================================================
-- COMMENTS AND DOCUMENTATION
-- =====================================================

COMMENT ON FUNCTION public.user_has_org_access(uuid) IS
'Overloaded version of user_has_org_access that checks if the current authenticated user has access to a specific organization. Used primarily in RLS policies.';

COMMENT ON FUNCTION public.user_has_role(character varying) IS
'Overloaded version of user_has_role that checks if the current authenticated user has a specific role in their current organization context. Used primarily in RLS policies.';

-- =====================================================
-- VERIFY ALL FUNCTIONS EXIST
-- =====================================================

-- Verify all function overloads exist
DO $$
BEGIN
    -- Check user_has_org_access() without parameters exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
          AND p.proname = 'user_has_org_access'
          AND p.pronargs = 0
    ) THEN
        RAISE EXCEPTION 'Function user_has_org_access() without parameters not found';
    END IF;

    -- Check user_has_org_access(uuid) with parameter exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
          AND p.proname = 'user_has_org_access'
          AND p.pronargs = 1
    ) THEN
        RAISE EXCEPTION 'Function user_has_org_access(uuid) with parameter not found';
    END IF;

    -- Check user_has_role(uuid, uuid, varchar) with 3 parameters exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
          AND p.proname = 'user_has_role'
          AND p.pronargs = 3
    ) THEN
        RAISE EXCEPTION 'Function user_has_role(uuid, uuid, varchar) with 3 parameters not found';
    END IF;

    -- Check user_has_role(varchar) with 1 parameter exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
          AND p.proname = 'user_has_role'
          AND p.pronargs = 1
    ) THEN
        RAISE EXCEPTION 'Function user_has_role(varchar) with 1 parameter not found';
    END IF;

    RAISE NOTICE 'All RLS function overloads are available and verified';
END $$;

-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

-- Grant execute permissions to authenticated users for RLS policy use
GRANT EXECUTE ON FUNCTION public.user_has_org_access(uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION public.user_has_role(character varying) TO authenticated;

-- Grant execute permissions to service role for background operations
GRANT EXECUTE ON FUNCTION public.user_has_org_access(uuid) TO service_role;
GRANT EXECUTE ON FUNCTION public.user_has_role(character varying) TO service_role;
