-- Migration: Fix RLS for Reference Tables
-- Description: Enable RLS and add proper policies for all global reference tables
-- Created: 2025-01-01

-- =====================================================
-- ENABLE RLS ON REFERENCE TABLES
-- =====================================================

ALTER TABLE countries ENABLE ROW LEVEL SECURITY;
ALTER TABLE states ENABLE ROW LEVEL SECURITY;
ALTER TABLE currencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE industries ENABLE ROW LEVEL SECURITY;
ALTER TABLE uom_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE account_account_types ENABLE ROW LEVEL SECURITY;

-- =====================================================
-- DROP EXISTING POLICIES (if they exist)
-- =====================================================

DROP POLICY IF EXISTS countries_select ON countries;
DROP POLICY IF EXISTS states_select ON states;
DROP POLICY IF EXISTS currencies_select ON currencies;
DROP POLICY IF EXISTS industries_select ON industries;
DROP POLICY IF EXISTS uom_categories_select ON uom_categories;
DROP POLICY IF EXISTS account_account_types_select ON account_account_types;

-- =====================================================
-- COUNTRIES POLICIES
-- =====================================================

-- Allow all authenticated users to read countries
CREATE POLICY countries_select
ON countries
FOR SELECT
TO authenticated
USING (true);

-- Only allow service role or admins to insert/update/delete
CREATE POLICY countries_insert
ON countries
FOR INSERT
TO authenticated
WITH CHECK (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY countries_update
ON countries
FOR UPDATE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY countries_delete
ON countries
FOR DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

-- =====================================================
-- STATES POLICIES
-- =====================================================

-- Allow all authenticated users to read states
CREATE POLICY states_select
ON states
FOR SELECT
TO authenticated
USING (true);

-- Only allow service role or admins to insert/update/delete
CREATE POLICY states_insert
ON states
FOR INSERT
TO authenticated
WITH CHECK (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY states_update
ON states
FOR UPDATE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY states_delete
ON states
FOR DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

-- =====================================================
-- CURRENCIES POLICIES
-- =====================================================

-- Allow all authenticated users to read currencies
CREATE POLICY currencies_select
ON currencies
FOR SELECT
TO authenticated
USING (true);

-- Only allow service role or admins to insert/update/delete
CREATE POLICY currencies_insert
ON currencies
FOR INSERT
TO authenticated
WITH CHECK (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY currencies_update
ON currencies
FOR UPDATE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY currencies_delete
ON currencies
FOR DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

-- =====================================================
-- INDUSTRIES POLICIES
-- =====================================================

-- Allow all authenticated users to read industries
CREATE POLICY industries_select
ON industries
FOR SELECT
TO authenticated
USING (true);

-- Only allow service role or admins to insert/update/delete
CREATE POLICY industries_insert
ON industries
FOR INSERT
TO authenticated
WITH CHECK (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY industries_update
ON industries
FOR UPDATE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY industries_delete
ON industries
FOR DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

-- =====================================================
-- UOM_CATEGORIES POLICIES
-- =====================================================

-- Allow all authenticated users to read uom_categories
CREATE POLICY uom_categories_select
ON uom_categories
FOR SELECT
TO authenticated
USING (true);

-- Only allow service role or admins to insert/update/delete
CREATE POLICY uom_categories_insert
ON uom_categories
FOR INSERT
TO authenticated
WITH CHECK (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY uom_categories_update
ON uom_categories
FOR UPDATE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY uom_categories_delete
ON uom_categories
FOR DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

-- =====================================================
-- ACCOUNT_ACCOUNT_TYPES POLICIES
-- =====================================================

-- Allow all authenticated users to read account_account_types
CREATE POLICY account_account_types_select
ON account_account_types
FOR SELECT
TO authenticated
USING (true);

-- Only allow service role or admins to insert/update/delete
CREATE POLICY account_account_types_insert
ON account_account_types
FOR INSERT
TO authenticated
WITH CHECK (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY account_account_types_update
ON account_account_types
FOR UPDATE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

CREATE POLICY account_account_types_delete
ON account_account_types
FOR DELETE
TO authenticated
USING (
    EXISTS (
        SELECT 1 FROM organization_users ou
        WHERE ou.user_id = (SELECT auth.uid())
        AND ou.role IN ('owner', 'admin')
    )
);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON POLICY countries_select ON countries IS 'All authenticated users can read countries';
COMMENT ON POLICY countries_insert ON countries IS 'Only owners and admins can create countries';
COMMENT ON POLICY countries_update ON countries IS 'Only owners and admins can update countries';
COMMENT ON POLICY countries_delete ON countries IS 'Only owners and admins can delete countries';

COMMENT ON POLICY states_select ON states IS 'All authenticated users can read states';
COMMENT ON POLICY states_insert ON states IS 'Only owners and admins can create states';
COMMENT ON POLICY states_update ON states IS 'Only owners and admins can update states';
COMMENT ON POLICY states_delete ON states IS 'Only owners and admins can delete states';

COMMENT ON POLICY currencies_select ON currencies IS 'All authenticated users can read currencies';
COMMENT ON POLICY currencies_insert ON currencies IS 'Only owners and admins can create currencies';
COMMENT ON POLICY currencies_update ON currencies IS 'Only owners and admins can update currencies';
COMMENT ON POLICY currencies_delete ON currencies IS 'Only owners and admins can delete currencies';

COMMENT ON POLICY industries_select ON industries IS 'All authenticated users can read industries';
COMMENT ON POLICY industries_insert ON industries IS 'Only owners and admins can create industries';
COMMENT ON POLICY industries_update ON industries IS 'Only owners and admins can update industries';
COMMENT ON POLICY industries_delete ON industries IS 'Only owners and admins can delete industries';

COMMENT ON POLICY uom_categories_select ON uom_categories IS 'All authenticated users can read UOM categories';
COMMENT ON POLICY uom_categories_insert ON uom_categories IS 'Only owners and admins can create UOM categories';
COMMENT ON POLICY uom_categories_update ON uom_categories IS 'Only owners and admins can update UOM categories';
COMMENT ON POLICY uom_categories_delete ON uom_categories IS 'Only owners and admins can delete UOM categories';

COMMENT ON POLICY account_account_types_select ON account_account_types IS 'All authenticated users can read account types';
COMMENT ON POLICY account_account_types_insert ON account_account_types IS 'Only owners and admins can create account types';
COMMENT ON POLICY account_account_types_update ON account_account_types IS 'Only owners and admins can update account types';
COMMENT ON POLICY account_account_types_delete ON account_account_types IS 'Only owners and admins can delete account types';
