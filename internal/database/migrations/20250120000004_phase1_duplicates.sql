-- Phase 1 CRM: Duplicate Detection and Merging
-- Track potential duplicates and merge history

-- Contact duplicates table for tracking potential duplicate pairs
CREATE TABLE contact_duplicates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contact_id_1 uuid NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    contact_id_2 uuid NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    similarity_score numeric(5,2) NOT NULL CHECK (similarity_score >= 0 AND similarity_score <= 100),
    matching_fields jsonb DEFAULT '[]'::jsonb, -- ["email", "phone", "name"]
    status varchar(20) DEFAULT 'pending', -- pending, merged, ignored, false_positive
    reviewed_by uuid,
    reviewed_at timestamptz,
    notes text,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT duplicate_order_check CHECK (contact_id_1 < contact_id_2),
    CONSTRAINT unique_duplicate_pair UNIQUE(organization_id, contact_id_1, contact_id_2),
    CONSTRAINT status_check CHECK (status IN ('pending', 'merged', 'ignored', 'false_positive'))
);

CREATE INDEX idx_contact_duplicates_org ON contact_duplicates(organization_id);
CREATE INDEX idx_contact_duplicates_status ON contact_duplicates(organization_id, status);
CREATE INDEX idx_contact_duplicates_score ON contact_duplicates(similarity_score DESC);
CREATE INDEX idx_contact_duplicates_contact1 ON contact_duplicates(contact_id_1);
CREATE INDEX idx_contact_duplicates_contact2 ON contact_duplicates(contact_id_2);
CREATE INDEX idx_contact_duplicates_pending ON contact_duplicates(organization_id, status, similarity_score DESC) WHERE status = 'pending';

-- Contact merge history for audit trail
CREATE TABLE contact_merge_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    master_contact_id uuid NOT NULL REFERENCES contacts(id),
    merged_contact_ids uuid[] NOT NULL,
    merge_strategy varchar(50) NOT NULL, -- keep_master, keep_latest, custom
    field_selections jsonb DEFAULT '{}'::jsonb, -- which fields were kept from which contact
    merged_by uuid,
    merged_at timestamptz NOT NULL DEFAULT now(),
    can_undo boolean DEFAULT true,
    notes text
);

CREATE INDEX idx_merge_history_org ON contact_merge_history(organization_id);
CREATE INDEX idx_merge_history_master ON contact_merge_history(master_contact_id);
CREATE INDEX idx_merge_history_merged_at ON contact_merge_history(merged_at DESC);
CREATE INDEX idx_merge_history_can_undo ON contact_merge_history(organization_id, can_undo) WHERE can_undo = true;

-- Comments
COMMENT ON TABLE contact_duplicates IS 'Potential duplicate contact pairs with similarity scores';
COMMENT ON COLUMN contact_duplicates.similarity_score IS 'Similarity score (0-100) based on matching algorithm';
COMMENT ON COLUMN contact_duplicates.matching_fields IS 'JSONB array of fields that matched (email, phone, name)';
COMMENT ON COLUMN contact_duplicates.status IS 'Resolution status: pending, merged, ignored, false_positive';
COMMENT ON TABLE contact_merge_history IS 'Audit trail of contact merge operations';
COMMENT ON COLUMN contact_merge_history.merge_strategy IS 'Strategy used: keep_master, keep_latest, custom';
COMMENT ON COLUMN contact_merge_history.field_selections IS 'JSONB mapping of field -> source contact ID for custom merges';
COMMENT ON COLUMN contact_merge_history.can_undo IS 'Whether the merge can be undone';
