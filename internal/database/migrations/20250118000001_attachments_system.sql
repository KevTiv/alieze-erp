-- =====================================================
-- ATTACHMENTS SYSTEM
-- =====================================================
-- Description: File attachment system with polymorphic references
-- Odoo-compatible (ir.attachment model)
-- Features:
--   - Checksum-based deduplication
--   - Polymorphic references (res_model/res_id pattern)
--   - Public/private access control
--   - Access tracking and analytics
--   - Version support (future enhancement)
-- =====================================================

-- Main attachments table
CREATE TABLE IF NOT EXISTS attachments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- File metadata
    name varchar(255) NOT NULL,
    description text,
    file_name varchar(255) NOT NULL,
    file_size bigint NOT NULL CHECK (file_size >= 0),
    mimetype varchar(100) NOT NULL,
    checksum varchar(64), -- SHA256 hash for deduplication

    -- Storage information
    storage_provider varchar(50) NOT NULL DEFAULT 's3',
    storage_key varchar(500) NOT NULL,  -- S3 key or file path
    storage_url varchar(1000),          -- Public URL if applicable

    -- Polymorphic relationship (Odoo-compatible pattern)
    res_model varchar(100),  -- e.g., 'leads', 'contacts', 'sales_orders', 'invoices'
    res_id uuid,             -- Foreign key to the related record
    res_name varchar(255),   -- Display name of the linked record

    -- Access control
    public boolean DEFAULT false,
    access_token varchar(100),  -- For public sharing with time-limited URLs

    -- Versioning (optional, for future use)
    version integer DEFAULT 1,
    previous_version_id uuid REFERENCES attachments(id) ON DELETE SET NULL,

    -- Audit fields
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid,

    -- Constraints
    CONSTRAINT attachments_checksum_unique UNIQUE(organization_id, checksum),
    CONSTRAINT attachments_res_check CHECK (
        (res_model IS NULL AND res_id IS NULL) OR
        (res_model IS NOT NULL AND res_id IS NOT NULL)
    )
);

-- Indexes for performance
CREATE INDEX idx_attachments_org ON attachments(organization_id);
CREATE INDEX idx_attachments_res ON attachments(res_model, res_id) WHERE res_model IS NOT NULL;
CREATE INDEX idx_attachments_checksum ON attachments(checksum) WHERE checksum IS NOT NULL;
CREATE INDEX idx_attachments_storage_key ON attachments(storage_key);
CREATE INDEX idx_attachments_created_at ON attachments(created_at DESC);
CREATE INDEX idx_attachments_public ON attachments(public) WHERE public = true;
CREATE INDEX idx_attachments_access_token ON attachments(access_token) WHERE access_token IS NOT NULL;

-- Attachment access log for analytics
CREATE TABLE IF NOT EXISTS attachment_access_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    attachment_id uuid NOT NULL REFERENCES attachments(id) ON DELETE CASCADE,
    user_id uuid,
    ip_address inet,
    user_agent text,
    accessed_at timestamptz NOT NULL DEFAULT now(),
    access_type varchar(20) DEFAULT 'view' CHECK (access_type IN ('view', 'download', 'share'))
);

CREATE INDEX idx_attachment_access_logs_attachment ON attachment_access_logs(attachment_id);
CREATE INDEX idx_attachment_access_logs_user ON attachment_access_logs(user_id);
CREATE INDEX idx_attachment_access_logs_date ON attachment_access_logs(accessed_at DESC);

-- Function to log attachment access
CREATE OR REPLACE FUNCTION log_attachment_access(
    p_attachment_id uuid,
    p_user_id uuid DEFAULT NULL,
    p_ip_address inet DEFAULT NULL,
    p_user_agent text DEFAULT NULL,
    p_access_type varchar DEFAULT 'view'
) RETURNS void AS $$
BEGIN
    INSERT INTO attachment_access_logs (
        attachment_id, user_id, ip_address, user_agent, access_type
    ) VALUES (
        p_attachment_id, p_user_id, p_ip_address, p_user_agent, p_access_type
    );
END;
$$ LANGUAGE plpgsql;

-- Function to find duplicate file by checksum
CREATE OR REPLACE FUNCTION find_duplicate_attachment(
    p_organization_id uuid,
    p_checksum varchar
) RETURNS TABLE (
    id uuid,
    name varchar,
    file_name varchar,
    file_size bigint,
    storage_key varchar,
    created_at timestamptz
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.name,
        a.file_name,
        a.file_size,
        a.storage_key,
        a.created_at
    FROM attachments a
    WHERE a.organization_id = p_organization_id
      AND a.checksum = p_checksum
    ORDER BY a.created_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to get attachments for a resource
CREATE OR REPLACE FUNCTION get_resource_attachments(
    p_organization_id uuid,
    p_res_model varchar,
    p_res_id uuid
) RETURNS TABLE (
    id uuid,
    name varchar,
    file_name varchar,
    file_size bigint,
    mimetype varchar,
    storage_url varchar,
    public boolean,
    created_at timestamptz,
    created_by uuid
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.id,
        a.name,
        a.file_name,
        a.file_size,
        a.mimetype,
        a.storage_url,
        a.public,
        a.created_at,
        a.created_by
    FROM attachments a
    WHERE a.organization_id = p_organization_id
      AND a.res_model = p_res_model
      AND a.res_id = p_res_id
    ORDER BY a.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- View for attachment statistics
CREATE OR REPLACE VIEW attachment_stats AS
SELECT
    organization_id,
    COUNT(*) as total_attachments,
    SUM(file_size) as total_size_bytes,
    SUM(file_size) / (1024.0 * 1024.0) as total_size_mb,
    SUM(file_size) / (1024.0 * 1024.0 * 1024.0) as total_size_gb,
    COUNT(DISTINCT res_model) as unique_models,
    COUNT(*) FILTER (WHERE public = true) as public_attachments,
    COUNT(*) FILTER (WHERE public = false) as private_attachments,
    MAX(created_at) as last_upload
FROM attachments
GROUP BY organization_id;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_attachment_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_attachment_timestamp
    BEFORE UPDATE ON attachments
    FOR EACH ROW
    EXECUTE FUNCTION update_attachment_timestamp();

-- Comments for documentation
COMMENT ON TABLE attachments IS 'File attachments with polymorphic references (Odoo ir.attachment compatible)';
COMMENT ON COLUMN attachments.res_model IS 'Model name of the related record (e.g., leads, contacts, sales_orders)';
COMMENT ON COLUMN attachments.res_id IS 'ID of the related record';
COMMENT ON COLUMN attachments.checksum IS 'SHA256 hash for deduplication';
COMMENT ON COLUMN attachments.access_token IS 'Token for time-limited public access';
COMMENT ON TABLE attachment_access_logs IS 'Tracks when and by whom attachments are accessed';
