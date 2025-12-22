-- Phase 1 CRM: Contact Table Enhancements
-- Add scoring fields and enrichment tracking to contacts table

ALTER TABLE contacts ADD COLUMN IF NOT EXISTS engagement_score integer DEFAULT 50 CHECK (engagement_score >= 0 AND engagement_score <= 100);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS lead_score integer DEFAULT 0 CHECK (lead_score >= 0 AND lead_score <= 100);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS data_quality_score integer DEFAULT 0 CHECK (data_quality_score >= 0 AND data_quality_score <= 100);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS last_contacted_at timestamptz;
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS last_activity_type varchar(50);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS next_recommended_action text;
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS is_duplicate boolean DEFAULT false;
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS enrichment_status varchar(20) DEFAULT 'pending';
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS enrichment_last_attempt timestamptz;

-- Create indexes for new fields
CREATE INDEX IF NOT EXISTS idx_contacts_engagement_score ON contacts(organization_id, engagement_score DESC);
CREATE INDEX IF NOT EXISTS idx_contacts_lead_score ON contacts(organization_id, lead_score DESC);
CREATE INDEX IF NOT EXISTS idx_contacts_data_quality_score ON contacts(organization_id, data_quality_score DESC);
CREATE INDEX IF NOT EXISTS idx_contacts_last_contacted ON contacts(last_contacted_at DESC) WHERE last_contacted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_contacts_duplicates ON contacts(organization_id, is_duplicate) WHERE is_duplicate = true;
CREATE INDEX IF NOT EXISTS idx_contacts_enrichment_status ON contacts(enrichment_status) WHERE enrichment_status != 'completed';

-- Add comments for documentation
COMMENT ON COLUMN contacts.engagement_score IS 'Contact engagement level (0-100) based on interactions';
COMMENT ON COLUMN contacts.lead_score IS 'Lead scoring for sales prioritization (0-100)';
COMMENT ON COLUMN contacts.data_quality_score IS 'Data completeness and accuracy score (0-100)';
COMMENT ON COLUMN contacts.last_contacted_at IS 'Timestamp of last contact/activity';
COMMENT ON COLUMN contacts.last_activity_type IS 'Type of last activity (call, email, meeting, etc.)';
COMMENT ON COLUMN contacts.next_recommended_action IS 'AI-suggested next action for this contact';
COMMENT ON COLUMN contacts.is_duplicate IS 'Flag indicating this contact might be a duplicate';
COMMENT ON COLUMN contacts.enrichment_status IS 'Status of data enrichment (pending, in_progress, completed, failed)';
COMMENT ON COLUMN contacts.enrichment_last_attempt IS 'Last attempt to enrich contact data';
