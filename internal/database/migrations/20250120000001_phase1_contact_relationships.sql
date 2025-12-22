-- Phase 1 CRM: Contact Relationship Enhancements
-- Add custom relationship types and relationship strength tracking

-- Create contact_relationship_types table for custom relationship definitions
CREATE TABLE IF NOT EXISTS contact_relationship_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    code varchar(50) NOT NULL,
    description text,
    is_bidirectional boolean DEFAULT false,
    reverse_name varchar(100), -- for bidirectional relationships
    is_active boolean DEFAULT true,
    is_system boolean DEFAULT false, -- system types cannot be deleted
    color integer,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT unique_relationship_type UNIQUE(organization_id, code)
);

CREATE INDEX idx_contact_relationship_types_org ON contact_relationship_types(organization_id);
CREATE INDEX idx_contact_relationship_types_active ON contact_relationship_types(organization_id, is_active) WHERE is_active = true;

CREATE TRIGGER set_contact_relationship_types_updated_at
    BEFORE UPDATE ON contact_relationship_types
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Enhance existing contact_relationships table (if it exists, add new columns)
-- If table doesn't exist, it will be created by the CRM module migration

DO $$
BEGIN
    -- Add strength_score if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'contact_relationships' AND column_name = 'strength_score') THEN
        ALTER TABLE contact_relationships ADD COLUMN strength_score integer DEFAULT 50 CHECK (strength_score >= 0 AND strength_score <= 100);
        CREATE INDEX idx_contact_relationships_strength ON contact_relationships(strength_score DESC);
    END IF;

    -- Add last_interaction_date if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'contact_relationships' AND column_name = 'last_interaction_date') THEN
        ALTER TABLE contact_relationships ADD COLUMN last_interaction_date timestamptz;
    END IF;

    -- Add interaction_count if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'contact_relationships' AND column_name = 'interaction_count') THEN
        ALTER TABLE contact_relationships ADD COLUMN interaction_count integer DEFAULT 0;
    END IF;

    -- Add metadata if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'contact_relationships' AND column_name = 'metadata') THEN
        ALTER TABLE contact_relationships ADD COLUMN metadata jsonb DEFAULT '{}'::jsonb;
    END IF;
END$$;

-- Seed default relationship types for all organizations
INSERT INTO contact_relationship_types (organization_id, name, code, is_bidirectional, reverse_name, is_system)
SELECT
    id as organization_id,
    'Colleague' as name,
    'colleague' as code,
    true as is_bidirectional,
    'Colleague' as reverse_name,
    true as is_system
FROM organizations
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO contact_relationship_types (organization_id, name, code, is_bidirectional, reverse_name, is_system)
SELECT
    id as organization_id,
    'Manager' as name,
    'manager' as code,
    true as is_bidirectional,
    'Reports To' as reverse_name,
    true as is_system
FROM organizations
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO contact_relationship_types (organization_id, name, code, is_bidirectional, reverse_name, is_system)
SELECT
    id as organization_id,
    'Family' as name,
    'family' as code,
    true as is_bidirectional,
    'Family' as reverse_name,
    true as is_system
FROM organizations
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO contact_relationship_types (organization_id, name, code, is_bidirectional, reverse_name, is_system)
SELECT
    id as organization_id,
    'Partner' as name,
    'partner' as code,
    true as is_bidirectional,
    'Partner' as reverse_name,
    true as is_system
FROM organizations
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO contact_relationship_types (organization_id, name, code, is_bidirectional, reverse_name, is_system)
SELECT
    id as organization_id,
    'Referral' as name,
    'referral' as code,
    false as is_bidirectional,
    NULL as reverse_name,
    true as is_system
FROM organizations
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO contact_relationship_types (organization_id, name, code, is_bidirectional, reverse_name, is_system)
SELECT
    id as organization_id,
    'Customer' as name,
    'customer' as code,
    true as is_bidirectional,
    'Vendor' as reverse_name,
    true as is_system
FROM organizations
ON CONFLICT (organization_id, code) DO NOTHING;

COMMENT ON TABLE contact_relationship_types IS 'Custom relationship type definitions for contacts';
COMMENT ON COLUMN contact_relationship_types.is_bidirectional IS 'Whether the relationship applies in both directions';
COMMENT ON COLUMN contact_relationship_types.reverse_name IS 'Name of the relationship when viewed from the related contact';
COMMENT ON COLUMN contact_relationship_types.is_system IS 'System types cannot be deleted, only deactivated';
COMMENT ON COLUMN contact_relationships.strength_score IS 'Relationship strength score (0-100)';
COMMENT ON COLUMN contact_relationships.interaction_count IS 'Number of recorded interactions between contacts';
