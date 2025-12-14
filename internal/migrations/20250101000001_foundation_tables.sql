-- Migration: Foundation Tables
-- Description: Core multi-tenant infrastructure tables
-- Created: 2025-01-01

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For text search

-- =====================================================
-- FOUNDATION TABLES
-- =====================================================

-- Organizations (Tenants)
CREATE TABLE organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    slug varchar(100) UNIQUE NOT NULL,
    subscription_tier varchar(50) DEFAULT 'free',
    subscription_status varchar(50) DEFAULT 'active',
    max_users integer DEFAULT 5,
    max_companies integer DEFAULT 1,
    features jsonb DEFAULT '{}'::jsonb,
    settings jsonb DEFAULT '{}'::jsonb,
    logo_url varchar(500),
    timezone varchar(100) DEFAULT 'UTC',
    date_format varchar(50) DEFAULT 'MM/DD/YYYY',
    language varchar(10) DEFAULT 'en_US',
    currency_id uuid, -- Will be linked after currencies table is created
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    CONSTRAINT organizations_slug_check CHECK (slug ~ '^[a-z0-9-]+$')
);

-- Organization Users (Membership & Roles)
CREATE TABLE organization_users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    role varchar(50) NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb,
    is_active boolean DEFAULT true,
    invited_at timestamptz,
    joined_at timestamptz DEFAULT now(),
    invited_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT organization_users_unique UNIQUE(organization_id, user_id),
    CONSTRAINT organization_users_role_check CHECK (role IN ('owner', 'admin', 'manager', 'user', 'viewer'))
);

-- Companies (Multi-company support within org)
CREATE TABLE companies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    legal_name varchar(255),
    tax_id varchar(100),
    registration_number varchar(100),
    email varchar(255),
    phone varchar(50),
    website varchar(255),
    logo_url varchar(500),
    currency_id uuid, -- Will be linked after currencies table
    parent_company_id uuid REFERENCES companies(id),
    is_default boolean DEFAULT false,
    -- Address fields
    street varchar(255),
    street2 varchar(255),
    city varchar(100),
    state_id uuid, -- Will be linked after states table
    zip varchar(20),
    country_id uuid, -- Will be linked after countries table
    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb
);

-- Sequences (For auto-numbering SO001, INV001, etc.)
CREATE TABLE sequences (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    sequence_code varchar(50) NOT NULL,
    name varchar(255) NOT NULL,
    prefix varchar(20),
    suffix varchar(20),
    padding integer DEFAULT 5,
    next_number integer DEFAULT 1,
    increment integer DEFAULT 1,
    implementation varchar(20) DEFAULT 'standard',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    CONSTRAINT sequences_unique UNIQUE(organization_id, sequence_code, company_id),
    CONSTRAINT sequences_implementation_check CHECK (implementation IN ('standard', 'no_gap'))
);

-- =====================================================
-- HELPER FUNCTIONS
-- =====================================================

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to generate sequence numbers
CREATE OR REPLACE FUNCTION generate_sequence_number(
    p_sequence_code text,
    p_organization_id uuid,
    p_company_id uuid DEFAULT NULL
)
RETURNS text AS $$
DECLARE
    v_sequence record;
    v_number integer;
    v_result text;
BEGIN
    -- Get and lock the sequence
    SELECT * INTO v_sequence
    FROM sequences
    WHERE organization_id = p_organization_id
      AND sequence_code = p_sequence_code
      AND (company_id = p_company_id OR (company_id IS NULL AND p_company_id IS NULL))
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Sequence % not found for organization %', p_sequence_code, p_organization_id;
    END IF;

    -- Get current number and increment
    v_number := v_sequence.next_number;

    -- Update next number
    UPDATE sequences
    SET next_number = next_number + v_sequence.increment
    WHERE id = v_sequence.id;

    -- Format the result
    v_result := COALESCE(v_sequence.prefix, '') ||
                LPAD(v_number::text, v_sequence.padding, '0') ||
                COALESCE(v_sequence.suffix, '');

    RETURN v_result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_organization_users_updated_at
    BEFORE UPDATE ON organization_users
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_sequences_updated_at
    BEFORE UPDATE ON sequences
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_organization_users_org ON organization_users(organization_id);
CREATE INDEX idx_organization_users_user ON organization_users(user_id);
CREATE INDEX idx_organization_users_role ON organization_users(organization_id, role);

CREATE INDEX idx_companies_org ON companies(organization_id);
CREATE INDEX idx_companies_parent ON companies(parent_company_id) WHERE parent_company_id IS NOT NULL;
CREATE INDEX idx_companies_deleted_at ON companies(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_sequences_org_code ON sequences(organization_id, sequence_code);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE organizations IS 'Multi-tenant root table - each organization is a separate tenant';
COMMENT ON TABLE organization_users IS 'User membership and roles within organizations';
COMMENT ON TABLE companies IS 'Multi-company support - organizations can have multiple legal entities';
COMMENT ON TABLE sequences IS 'Auto-numbering configuration for documents (SO001, INV001, etc)';

COMMENT ON FUNCTION generate_sequence_number IS 'Generate next sequence number for documents';
COMMENT ON FUNCTION trigger_set_updated_at IS 'Automatically update updated_at timestamp on row changes';
