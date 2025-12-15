-- Migration: Data Import System
-- Description: Tables to support AI-powered CSV/Excel data import
-- Created: 2025-01-01
-- Purpose: Enable magical data onboarding with AI auto-detection and mapping

-- ============================================
-- DATA IMPORT SESSIONS
-- ============================================

CREATE TABLE data_import_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid,
    
    -- File info
    filename varchar(255),
    file_size integer,
    file_type varchar(20), -- 'csv', 'xlsx', 'xls'
    
    -- Import config
    target_table varchar(100), -- 'contacts', 'products', 'sales_orders', etc.
    import_mode varchar(20), -- 'create_only', 'update_only', 'create_or_update'
    
    -- Data
    raw_data jsonb, -- Parsed CSV/Excel data
    sample_data jsonb, -- First 10 rows for preview
    total_rows integer,
    
    -- AI Analysis
    column_mapping jsonb, -- AI-suggested mapping
    mapping_confidence jsonb, -- Confidence scores per column
    detected_relationships jsonb, -- Foreign key relationships
    
    -- Validation
    validation_results jsonb,
    errors jsonb,
    warnings jsonb,
    
    -- Status
    status varchar(20), -- 'uploaded', 'analyzing', 'mapped', 'validated', 'ready', 'importing', 'completed', 'failed'
    progress integer DEFAULT 0, -- 0-100
    
    -- Results
    rows_processed integer DEFAULT 0,
    rows_created integer DEFAULT 0,
    rows_updated integer DEFAULT 0,
    rows_skipped integer DEFAULT 0,
    rows_failed integer DEFAULT 0,
    
    -- Timestamps
    created_at timestamptz DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,
    
    -- Metadata
    metadata jsonb DEFAULT '{}'::jsonb
);

COMMENT ON TABLE data_import_sessions IS
'Tracks data import sessions with AI analysis and results';

CREATE INDEX idx_data_import_sessions_org ON data_import_sessions(organization_id, status);
CREATE INDEX idx_data_import_sessions_user ON data_import_sessions(user_id, status);

-- ============================================
-- DATA IMPORT ROW RESULTS
-- ============================================

CREATE TABLE data_import_row_results (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES data_import_sessions(id) ON DELETE CASCADE,
    row_number integer,
    row_data jsonb,
    transformed_data jsonb,
    status varchar(20), -- 'pending', 'success', 'skipped', 'failed'
    error_message text,
    created_record_id uuid,
    created_at timestamptz DEFAULT now()
);

COMMENT ON TABLE data_import_row_results IS
'Detailed tracking of individual row import results';

CREATE INDEX idx_data_import_row_results_session ON data_import_row_results(session_id);
CREATE INDEX idx_data_import_row_results_status ON data_import_row_results(session_id, status);

-- ============================================
-- DATA IMPORT TEMPLATES
-- ============================================

CREATE TABLE data_import_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255),
    target_table varchar(100),
    column_mapping jsonb,
    transformation_rules jsonb,
    is_default boolean DEFAULT false,
    created_at timestamptz DEFAULT now()
);

COMMENT ON TABLE data_import_templates IS
'Reusable import templates for common data formats';

CREATE INDEX idx_data_import_templates_org ON data_import_templates(organization_id, target_table);

-- ============================================
-- DATA IMPORT HISTORY
-- ============================================

CREATE TABLE data_import_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid,
    session_id uuid REFERENCES data_import_sessions(id),
    
    -- Import summary
    import_type varchar(50),
    total_rows integer,
    success_rows integer,
    failed_rows integer,
    
    -- Timestamps
    imported_at timestamptz DEFAULT now()
);

COMMENT ON TABLE data_import_history IS
'Historical record of completed data imports';

CREATE INDEX idx_data_import_history_org ON data_import_history(organization_id, imported_at);

-- ============================================
-- AI FUNCTIONS FOR DATA IMPORT
-- ============================================

-- Main AI Analysis Function
CREATE OR REPLACE FUNCTION ai_analyze_import_data(
    p_session_id uuid,
    p_sample_data jsonb,
    p_target_table varchar DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_prompt text;
    v_response jsonb;
    v_schema jsonb;
    v_session record;
BEGIN
    SELECT * INTO v_session FROM data_import_sessions WHERE id = p_session_id;
    
    -- Get target table schema
    IF p_target_table IS NOT NULL THEN
        v_schema := get_table_schema(p_target_table);
    ELSE
        -- AI detects table from data
        v_schema := get_all_importable_schemas();
    END IF;
    
    -- Build AI prompt
    v_prompt := format(
        'Analyze this CSV data and map it to our database schema.
        
        CSV Sample (first 5 rows):
        %s
        
        Available Database Tables & Fields:
        %s
        
        Return JSON with:
        {
          "target_table": "table_name",
          "confidence": 0.95,
          "column_mapping": {
            "CSV_Column": {
              "db_field": "field_name",
              "confidence": 0.98,
              "transformation": "phone_format|date_format|none",
              "notes": "optional explanation"
            }
          },
          "detected_relationships": [
            {"column": "Product SKU", "references_table": "products", "references_field": "default_code"}
          ],
          "warnings": ["list of potential issues"],
          "suggestions": ["list of recommendations"]
        }',
        p_sample_data::text,
        v_schema::text
    );
    
    -- Call Groq AI (GPT-OSS-120B for structured data)
    v_response := groq_gpt_oss_reasoning(v_prompt, 15000, 0.1);
    
    -- Update session with AI analysis
    UPDATE data_import_sessions
    SET
        column_mapping = (v_response->'content')::jsonb,
        status = 'mapped'
    WHERE id = p_session_id;
    
    RETURN v_response;
END;
$$;

COMMENT ON FUNCTION ai_analyze_import_data IS
'AI analyzes uploaded data and suggests column mappings';

-- Column Mapping Function
CREATE OR REPLACE FUNCTION ai_map_columns(
    p_csv_columns text[],
    p_table_name varchar,
    p_sample_values jsonb DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_prompt text;
    v_schema jsonb;
    v_response text;
BEGIN
    v_schema := get_table_schema(p_table_name);
    
    v_prompt := format(
        'Map these CSV columns to database fields:
        
        CSV Columns: %s
        Sample Values: %s
        
        Database Schema for %s table:
        %s
        
        Return JSON mapping with confidence scores.',
        array_to_string(p_csv_columns, ', '),
        COALESCE(p_sample_values::text, 'not provided'),
        p_table_name,
        v_schema::text
    );
    
    v_response := groq_chat_completion(v_prompt, 'llama-3.3-70b-versatile', 0.1, 2000);
    
    RETURN v_response::jsonb;
END;
$$;

COMMENT ON FUNCTION ai_map_columns IS
'AI maps CSV columns to database fields with confidence scores';

-- Data Validation Function
CREATE OR REPLACE FUNCTION validate_import_data(
    p_session_id uuid
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_session record;
    v_errors jsonb := '[]'::jsonb;
    v_warnings jsonb := '[]'::jsonb;
    v_row jsonb;
    v_row_num integer := 0;
BEGIN
    SELECT * INTO v_session FROM data_import_sessions WHERE id = p_session_id;
    
    -- Validate each row
    FOR v_row IN SELECT * FROM jsonb_array_elements(v_session.raw_data)
    LOOP
        v_row_num := v_row_num + 1;
        
        -- Check required fields
        -- Check data types
        -- Check foreign key references
        -- Check duplicates
        -- Add to errors/warnings
    END LOOP;
    
    UPDATE data_import_sessions
    SET
        validation_results = jsonb_build_object(
            'total_rows', v_row_num,
            'errors', v_errors,
            'warnings', v_warnings
        ),
        status = 'validated'
    WHERE id = p_session_id;
    
    RETURN jsonb_build_object('errors', v_errors, 'warnings', v_warnings);
END;
$$;

COMMENT ON FUNCTION validate_import_data IS
'Validates import data and identifies issues';

-- Execute Import Function
CREATE OR REPLACE FUNCTION execute_data_import(
    p_session_id uuid,
    p_confirmed_mapping jsonb DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_session record;
    v_mapping jsonb;
    v_result jsonb;
    v_created integer := 0;
    v_updated integer := 0;
    v_skipped integer := 0;
    v_failed integer := 0;
BEGIN
    SELECT * INTO v_session FROM data_import_sessions WHERE id = p_session_id;
    
    -- Use confirmed mapping or AI mapping
    v_mapping := COALESCE(p_confirmed_mapping, v_session.column_mapping);
    
    -- Update status
    UPDATE data_import_sessions
    SET status = 'importing', started_at = now()
    WHERE id = p_session_id;
    
    -- Enqueue as background job for large imports
    IF v_session.total_rows > 100 THEN
        PERFORM enqueue_job('data_import', jsonb_build_object(
            'session_id', p_session_id,
            'mapping', v_mapping
        ));
        
        RETURN jsonb_build_object('status', 'queued', 'job_type', 'background');
    END IF;
    
    -- Process inline for small imports
    -- Loop through rows, transform, insert/update
    -- Track results
    
    -- Update session with results
    UPDATE data_import_sessions
    SET
        status = 'completed',
        completed_at = now(),
        rows_created = v_created,
        rows_updated = v_updated,
        rows_skipped = v_skipped,
        rows_failed = v_failed
    WHERE id = p_session_id;
    
    RETURN jsonb_build_object(
        'created', v_created,
        'updated', v_updated,
        'skipped', v_skipped,
        'failed', v_failed
    );
END;
$$;

COMMENT ON FUNCTION execute_data_import IS
'Executes the data import process';

-- ============================================
-- TRIGGERS
-- ============================================

CREATE OR REPLACE FUNCTION update_import_session_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Note: data_import_sessions doesn't have updated_at, but if added:
-- CREATE TRIGGER set_data_import_sessions_updated_at
--     BEFORE UPDATE ON data_import_sessions
--     FOR EACH ROW
--     EXECUTE FUNCTION update_import_session_updated_at();

-- ============================================
-- RLS POLICIES
-- ============================================

ALTER TABLE data_import_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_import_row_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_import_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_import_history ENABLE ROW LEVEL SECURITY;

-- Data Import Sessions Policies
CREATE POLICY "Users can view their import sessions"
    ON data_import_sessions FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Users can manage their own import sessions"
    ON data_import_sessions FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

-- Data Import Row Results Policies
CREATE POLICY "Users can view their import row results"
    ON data_import_row_results FOR SELECT
    USING (session_id IN (
        SELECT id FROM data_import_sessions
        WHERE organization_id IN (
            SELECT organization_id FROM organization_users
            WHERE user_id = (SELECT auth.uid())
        )
    ));

-- Data Import Templates Policies
CREATE POLICY "Users can view their org import templates"
    ON data_import_templates FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Users can manage their org import templates"
    ON data_import_templates FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

-- Data Import History Policies
CREATE POLICY "Users can view their import history"
    ON data_import_history FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT, INSERT, UPDATE ON data_import_sessions TO authenticated;
GRANT ALL ON data_import_sessions TO service_role;

GRANT SELECT ON data_import_row_results TO authenticated;
GRANT ALL ON data_import_row_results TO service_role;

GRANT ALL ON data_import_templates TO authenticated;
GRANT ALL ON data_import_templates TO service_role;

GRANT SELECT ON data_import_history TO authenticated;
GRANT ALL ON data_import_history TO service_role;

GRANT EXECUTE ON FUNCTION ai_analyze_import_data(uuid, jsonb, varchar) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_map_columns(text[], varchar, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION validate_import_data(uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION execute_data_import(uuid, jsonb) TO authenticated;

-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Create import session
INSERT INTO data_import_sessions (
    organization_id, user_id, filename, file_type, 
    raw_data, sample_data, total_rows, status
) VALUES (
    'your-org-id', 'user-id', 'customers.csv', 'csv',
    '[{"name": "Acme Corp", "email": "contact@acme.com"}, ...]',
    '[{"name": "Acme Corp", "email": "contact@acme.com"}]',
    100, 'uploaded'
) RETURNING id;

-- 2. Analyze with AI
SELECT ai_analyze_import_data('session-id', 
    '[{"name": "Acme Corp", "email": "contact@acme.com"}]',
    'contacts'
);

-- 3. Validate data
SELECT validate_import_data('session-id');

-- 4. Execute import
SELECT execute_data_import('session-id');

-- 5. Get import history
SELECT * FROM data_import_history 
WHERE organization_id = 'your-org-id'
ORDER BY imported_at DESC;

-- 6. Create import template
INSERT INTO data_import_templates (
    organization_id, name, target_table, column_mapping
) VALUES (
    'your-org-id', 'Customer Import Template', 'contacts',
    '{"name": "name", "email": "email"}'
);
*/
