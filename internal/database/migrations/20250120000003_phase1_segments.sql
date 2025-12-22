-- Phase 1 CRM: Contact Segmentation
-- Static and dynamic contact segmentation tables

-- Contact segments table (static and dynamic)
CREATE TABLE contact_segments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    segment_type varchar(20) NOT NULL DEFAULT 'static', -- static or dynamic

    -- Dynamic segment criteria (JSONB for flexible rule definitions)
    criteria jsonb DEFAULT '{}'::jsonb,

    -- Metadata
    color integer,
    icon varchar(50),

    -- Stats (updated periodically for dynamic segments)
    member_count integer DEFAULT 0,
    last_calculated_at timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,

    CONSTRAINT segment_type_check CHECK (segment_type IN ('static', 'dynamic'))
);

CREATE INDEX idx_contact_segments_org ON contact_segments(organization_id);
CREATE INDEX idx_contact_segments_type ON contact_segments(organization_id, segment_type);
CREATE INDEX idx_contact_segments_name ON contact_segments(organization_id, name);

CREATE TRIGGER set_contact_segments_updated_at
    BEFORE UPDATE ON contact_segments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Segment membership junction table
CREATE TABLE contact_segment_members (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    segment_id uuid NOT NULL REFERENCES contact_segments(id) ON DELETE CASCADE,
    contact_id uuid NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    added_at timestamptz NOT NULL DEFAULT now(),
    added_by uuid,

    CONSTRAINT unique_segment_member UNIQUE(organization_id, segment_id, contact_id)
);

CREATE INDEX idx_segment_members_org ON contact_segment_members(organization_id);
CREATE INDEX idx_segment_members_segment ON contact_segment_members(segment_id);
CREATE INDEX idx_segment_members_contact ON contact_segment_members(contact_id);
CREATE INDEX idx_segment_members_composite ON contact_segment_members(segment_id, contact_id);

-- Comments
COMMENT ON TABLE contact_segments IS 'Contact segmentation definitions (static and dynamic)';
COMMENT ON COLUMN contact_segments.segment_type IS 'Type: static (manual membership) or dynamic (criteria-based)';
COMMENT ON COLUMN contact_segments.criteria IS 'JSONB criteria for dynamic segment evaluation';
COMMENT ON COLUMN contact_segments.member_count IS 'Cached member count, updated on recalculation';
COMMENT ON TABLE contact_segment_members IS 'Junction table for segment membership';
