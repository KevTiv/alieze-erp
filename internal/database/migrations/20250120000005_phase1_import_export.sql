-- Phase 1 CRM: Contact Import/Export Job Tracking
-- Track async import and export operations

-- Contact import jobs table
CREATE TABLE contact_import_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    job_id uuid NOT NULL, -- reference to queue job
    filename varchar(255) NOT NULL,
    file_size bigint,
    file_type varchar(20) NOT NULL, -- csv, xlsx

    -- Field mapping (source column -> contact field)
    field_mapping jsonb NOT NULL,

    -- Import options
    options jsonb DEFAULT '{}'::jsonb, -- skip_duplicates, update_existing, etc.

    -- Stats
    total_rows integer DEFAULT 0,
    processed_rows integer DEFAULT 0,
    successful_rows integer DEFAULT 0,
    failed_rows integer DEFAULT 0,
    duplicate_rows integer DEFAULT 0,

    -- Status
    status varchar(20) DEFAULT 'pending', -- pending, processing, completed, failed, cancelled
    error_message text,
    error_details jsonb,

    -- Results
    imported_contact_ids uuid[],
    failed_rows_data jsonb, -- store first 100 failed rows with error details

    created_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,

    CONSTRAINT import_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT import_file_type_check CHECK (file_type IN ('csv', 'xlsx'))
);

CREATE INDEX idx_import_jobs_org ON contact_import_jobs(organization_id);
CREATE INDEX idx_import_jobs_status ON contact_import_jobs(organization_id, status);
CREATE INDEX idx_import_jobs_created ON contact_import_jobs(created_at DESC);
CREATE INDEX idx_import_jobs_job_id ON contact_import_jobs(job_id);
CREATE INDEX idx_import_jobs_processing ON contact_import_jobs(organization_id, status, created_at DESC) WHERE status IN ('pending', 'processing');

-- Contact export jobs table
CREATE TABLE contact_export_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    job_id uuid NOT NULL, -- reference to queue job

    -- Export criteria
    filter_criteria jsonb NOT NULL,
    selected_fields jsonb, -- null = all fields
    format varchar(20) NOT NULL, -- csv, xlsx

    -- Stats
    total_contacts integer DEFAULT 0,

    -- Status
    status varchar(20) DEFAULT 'pending',
    error_message text,

    -- Result
    file_key varchar(500), -- storage key for download
    file_url varchar(1000), -- presigned URL
    file_expires_at timestamptz,

    created_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,

    CONSTRAINT export_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT export_format_check CHECK (format IN ('csv', 'xlsx'))
);

CREATE INDEX idx_export_jobs_org ON contact_export_jobs(organization_id);
CREATE INDEX idx_export_jobs_status ON contact_export_jobs(organization_id, status);
CREATE INDEX idx_export_jobs_created ON contact_export_jobs(created_at DESC);
CREATE INDEX idx_export_jobs_job_id ON contact_export_jobs(job_id);
CREATE INDEX idx_export_jobs_expires ON contact_export_jobs(file_expires_at) WHERE file_expires_at IS NOT NULL;

-- Comments
COMMENT ON TABLE contact_import_jobs IS 'Async contact import job tracking';
COMMENT ON COLUMN contact_import_jobs.field_mapping IS 'JSONB mapping of source columns to contact fields';
COMMENT ON COLUMN contact_import_jobs.options IS 'Import options: skip_duplicates, update_existing, duplicate_strategy';
COMMENT ON COLUMN contact_import_jobs.failed_rows_data IS 'JSONB array of failed rows (max 100) with error details';
COMMENT ON TABLE contact_export_jobs IS 'Async contact export job tracking';
COMMENT ON COLUMN contact_export_jobs.filter_criteria IS 'JSONB filter criteria for contact selection';
COMMENT ON COLUMN contact_export_jobs.selected_fields IS 'JSONB array of field names to export (null = all)';
COMMENT ON COLUMN contact_export_jobs.file_key IS 'Storage system key for the exported file';
COMMENT ON COLUMN contact_export_jobs.file_url IS 'Presigned download URL (temporary)';
