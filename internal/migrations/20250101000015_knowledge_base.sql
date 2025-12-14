-- Migration: Knowledge Base Schema
-- Description: Structured knowledge hub linked to operational records
-- Created: 2025-01-01

-- =====================================================
-- EXTENSIONS
-- =====================================================

CREATE EXTENSION IF NOT EXISTS "vector";

-- =====================================================
-- TABLES
-- =====================================================

-- Spaces group knowledge within an organization (e.g., Sales Playbooks, Product Docs)
CREATE TABLE knowledge_spaces (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    visibility text NOT NULL DEFAULT 'internal',
    created_by uuid NOT NULL REFERENCES organization_users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT knowledge_spaces_visibility_chk
        CHECK (visibility IN ('internal', 'public', 'restricted'))
);

-- Core entries store SOPs, how-to guides, playbooks, etc.
CREATE TABLE knowledge_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    space_id uuid NOT NULL REFERENCES knowledge_spaces(id) ON DELETE CASCADE,
    title text NOT NULL,
    body_markdown text,
    summary text,
    owner_id uuid REFERENCES organization_users(id),
    visibility text NOT NULL DEFAULT 'space',
    is_published boolean NOT NULL DEFAULT true,
    search_embedding vector(768),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    CONSTRAINT knowledge_entries_visibility_chk
        CHECK (visibility IN ('space', 'org', 'restricted'))
);

-- Revision history for auditability
CREATE TABLE knowledge_entry_revisions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id uuid NOT NULL REFERENCES knowledge_entries(id) ON DELETE CASCADE,
    revision_number integer NOT NULL,
    author_id uuid REFERENCES organization_users(id),
    title text,
    body_markdown text,
    summary text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT knowledge_entry_revisions_unique UNIQUE(entry_id, revision_number)
);

-- Tags for lightweight classification
CREATE TABLE knowledge_tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    color text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT knowledge_tags_unique UNIQUE(organization_id, name)
);

CREATE TABLE knowledge_entry_tags (
    entry_id uuid NOT NULL REFERENCES knowledge_entries(id) ON DELETE CASCADE,
    tag_id uuid NOT NULL REFERENCES knowledge_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, tag_id)
);

-- Polymorphic links to operational data (products, processes, opportunities, etc.)
CREATE TABLE knowledge_context_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id uuid NOT NULL REFERENCES knowledge_entries(id) ON DELETE CASCADE,
    context_type text NOT NULL,
    context_id uuid NOT NULL,
    context_table text NOT NULL,
    relationship jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid REFERENCES organization_users(id),
    CONSTRAINT knowledge_context_type_chk CHECK (context_type ~ '^[a-z_]+$')
);

-- Attachments for files stored in Supabase Storage
CREATE TABLE knowledge_entry_assets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id uuid NOT NULL REFERENCES knowledge_entries(id) ON DELETE CASCADE,
    storage_path text NOT NULL,
    filename text,
    mimetype text,
    size_bytes integer,
    uploaded_by uuid REFERENCES organization_users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_knowledge_spaces_org ON knowledge_spaces(organization_id);
CREATE INDEX idx_knowledge_entries_org_space ON knowledge_entries(organization_id, space_id);
CREATE INDEX idx_knowledge_entries_owner ON knowledge_entries(owner_id);
CREATE INDEX idx_knowledge_entries_embedding ON knowledge_entries USING ivfflat (search_embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_knowledge_context_links_entry ON knowledge_context_links(entry_id);
CREATE INDEX idx_knowledge_context_links_context ON knowledge_context_links(context_type, context_id);
CREATE INDEX idx_knowledge_tags_org ON knowledge_tags(organization_id);

-- =====================================================
-- GRAINED ACCESS (RLS STUBS)
-- =====================================================

ALTER TABLE knowledge_spaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE knowledge_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE knowledge_entry_revisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE knowledge_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE knowledge_entry_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE knowledge_context_links ENABLE ROW LEVEL SECURITY;
ALTER TABLE knowledge_entry_assets ENABLE ROW LEVEL SECURITY;

COMMENT ON TABLE knowledge_spaces IS 'Knowledge space containers (e.g., Sales Playbook, Product Docs)';
COMMENT ON TABLE knowledge_entries IS 'Core knowledge articles stored as markdown with metadata and embeddings';
COMMENT ON TABLE knowledge_entry_revisions IS 'Historical revisions for knowledge entries';
COMMENT ON TABLE knowledge_tags IS 'Organization-level tags for knowledge entries';
COMMENT ON TABLE knowledge_entry_tags IS 'Many-to-many join table linking knowledge entries and tags';
COMMENT ON TABLE knowledge_context_links IS 'Associates knowledge entries with operational records (product, process, etc.)';
COMMENT ON TABLE knowledge_entry_assets IS 'File attachments for knowledge entries';

-- =====================================================
-- HELPERS
-- =====================================================

CREATE OR REPLACE FUNCTION knowledge_entries_for_context(
    p_org_id uuid,
    p_context_type text,
    p_context_id uuid,
    p_limit integer DEFAULT 10
)
RETURNS TABLE (
    entry_id uuid,
    title text,
    summary text,
    context_relationship jsonb,
    updated_at timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ke.id,
        ke.title,
        ke.summary,
        kcl.relationship,
        ke.updated_at
    FROM knowledge_context_links kcl
    JOIN knowledge_entries ke ON ke.id = kcl.entry_id
    WHERE ke.organization_id = p_org_id
      AND kcl.context_type = p_context_type
      AND kcl.context_id = p_context_id
      AND ke.is_published = true
    ORDER BY ke.updated_at DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION knowledge_entries_for_context IS 'Fetch published knowledge entries for a given operational context';
