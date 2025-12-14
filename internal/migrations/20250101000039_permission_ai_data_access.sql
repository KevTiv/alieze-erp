-- Migration: Permission System - AI Data Access Control
-- Description: Control which data can be exposed to AI models and how
-- Created: 2025-01-01
-- Module: AI Data Access

-- =====================================================
-- AI PERMISSION TYPES
-- =====================================================

-- AI provider types
CREATE TYPE ai_provider_type AS ENUM (
    'internal',    -- Local/self-hosted models (Ollama)
    'external',    -- External APIs (OpenAI, Groq, etc.)
    'hybrid'       -- Can use either based on rules
);

-- AI data exposure levels
CREATE TYPE ai_exposure_level AS ENUM (
    'none',        -- Never expose to AI
    'metadata',    -- Only table/column names, no data
    'aggregated',  -- Only aggregated/statistical data
    'sanitized',   -- Individual data but sanitized
    'full'         -- Full access (use with caution)
);

-- =====================================================
-- AI DATA PERMISSIONS
-- =====================================================

-- Define which tables/columns AI can access
CREATE TABLE ai_data_permissions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id uuid REFERENCES permission_roles(id) ON DELETE CASCADE,
    group_id uuid REFERENCES permission_groups(id) ON DELETE CASCADE,
    user_id uuid,

    -- Resource identification
    table_name text NOT NULL,
    column_name text, -- NULL means entire table
    schema_name text DEFAULT 'public',

    -- AI exposure settings
    exposure_level ai_exposure_level DEFAULT 'none',
    allowed_providers ai_provider_type[] DEFAULT ARRAY['internal']::ai_provider_type[],

    -- Sanitization rules
    requires_sanitization boolean DEFAULT true,
    sanitization_method varchar(50), -- 'pii_removal', 'hash', 'tokenize', 'aggregate'
    sanitization_config jsonb, -- Method-specific configuration

    -- Context settings
    include_in_context boolean DEFAULT false, -- Include in AI context window
    context_description text, -- Human-readable description for AI
    max_rows_in_context integer DEFAULT 10, -- Limit rows sent to AI

    -- Usage restrictions
    allowed_operations text[], -- 'query', 'summarize', 'analyze', 'generate'
    rate_limit_per_hour integer,
    requires_approval boolean DEFAULT false,

    -- Conditions
    conditions jsonb, -- Additional conditions for AI access
    priority integer DEFAULT 100,

    -- Metadata
    description text,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT ai_data_permissions_role_or_group_or_user CHECK (
        (role_id IS NOT NULL AND group_id IS NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NOT NULL AND user_id IS NULL) OR
        (role_id IS NULL AND group_id IS NULL AND user_id IS NOT NULL)
    ),
    CONSTRAINT ai_data_permissions_priority_check CHECK (priority BETWEEN 1 AND 1000)
);

CREATE INDEX idx_ai_data_permissions_org ON ai_data_permissions(organization_id, table_name)
    WHERE is_active = true;
CREATE INDEX idx_ai_data_permissions_role ON ai_data_permissions(role_id)
    WHERE is_active = true AND role_id IS NOT NULL;
CREATE INDEX idx_ai_data_permissions_exposure ON ai_data_permissions(exposure_level)
    WHERE is_active = true;

COMMENT ON TABLE ai_data_permissions IS 'Controls which data can be exposed to AI models and under what conditions';

-- =====================================================
-- AI CONTEXT RULES
-- =====================================================

-- Define what context AI receives for different query types
CREATE TABLE ai_context_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Context identification
    context_name varchar(100) NOT NULL,
    query_pattern text, -- Regex pattern to match queries
    operation_type varchar(50), -- 'query', 'report', 'analysis', etc.

    -- Tables to include
    included_tables text[],
    excluded_tables text[],

    -- Data scope
    max_rows_per_table integer DEFAULT 100,
    time_window_days integer, -- Only include recent data

    -- Aggregation rules
    prefer_aggregation boolean DEFAULT true,
    aggregation_functions text[], -- 'count', 'sum', 'avg', etc.
    group_by_columns text[],

    -- Sanitization
    auto_sanitize boolean DEFAULT true,
    sanitization_level varchar(50) DEFAULT 'standard', -- 'minimal', 'standard', 'strict'

    -- Provider routing
    preferred_provider ai_provider_type DEFAULT 'internal',
    fallback_provider ai_provider_type,

    -- Metadata
    description text,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT ai_context_rules_unique UNIQUE(organization_id, context_name)
);

CREATE INDEX idx_ai_context_rules_org ON ai_context_rules(organization_id)
    WHERE is_active = true;
CREATE INDEX idx_ai_context_rules_pattern ON ai_context_rules USING gin(query_pattern gin_trgm_ops)
    WHERE is_active = true;

COMMENT ON TABLE ai_context_rules IS 'Defines what context and data AI models receive for different types of queries';

-- =====================================================
-- AI SANITIZATION RULES
-- =====================================================

-- Custom sanitization rules per table/column
CREATE TABLE ai_sanitization_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Target identification
    table_name text NOT NULL,
    column_name text,
    data_type varchar(50), -- 'email', 'phone', 'ssn', 'credit_card', etc.

    -- Sanitization method
    method varchar(50) NOT NULL, -- 'remove', 'mask', 'hash', 'tokenize', 'generalize'
    pattern text, -- Regex pattern for detection
    replacement text, -- Replacement pattern or value

    -- Advanced options
    preserve_format boolean DEFAULT false, -- Keep format (e.g., XXX-XX-1234 for SSN)
    preserve_domain boolean DEFAULT false, -- For emails, keep @domain.com
    generalization_level integer, -- For generalization method (1-5)

    -- Conditions
    apply_condition jsonb, -- When to apply this rule
    priority integer DEFAULT 100,

    -- Metadata
    description text,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    -- Constraints
    CONSTRAINT ai_sanitization_rules_method CHECK (
        method IN ('remove', 'mask', 'hash', 'tokenize', 'generalize', 'replace')
    ),
    CONSTRAINT ai_sanitization_rules_priority_check CHECK (priority BETWEEN 1 AND 1000)
);

CREATE INDEX idx_ai_sanitization_rules_org ON ai_sanitization_rules(organization_id, table_name)
    WHERE is_active = true;
CREATE INDEX idx_ai_sanitization_rules_column ON ai_sanitization_rules(table_name, column_name)
    WHERE is_active = true;

COMMENT ON TABLE ai_sanitization_rules IS 'Custom sanitization rules for specific data types and columns before AI exposure';

-- =====================================================
-- AI USAGE TRACKING
-- =====================================================

-- Track AI data access for compliance and auditing
CREATE TABLE ai_data_access_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,

    -- Request details
    request_id uuid NOT NULL,
    query_type varchar(50),
    query_hash text, -- SHA256 of query

    -- Data accessed
    tables_accessed text[],
    columns_accessed jsonb, -- {"table": ["col1", "col2"]}
    row_count integer,

    -- AI details
    provider ai_provider_type,
    model_name varchar(100),
    exposure_level ai_exposure_level,
    was_sanitized boolean,
    sanitization_methods text[],

    -- Performance
    processing_time_ms integer,
    tokens_used integer,

    -- Results
    response_hash text, -- SHA256 of response
    error_message text,

    -- Metadata
    client_ip inet,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_data_access_log_org ON ai_data_access_log(organization_id, created_at DESC);
CREATE INDEX idx_ai_data_access_log_user ON ai_data_access_log(user_id, created_at DESC);
CREATE INDEX idx_ai_data_access_log_request ON ai_data_access_log(request_id);

COMMENT ON TABLE ai_data_access_log IS 'Audit log of all AI data access for compliance and monitoring';

-- =====================================================
-- HELPER FUNCTIONS
-- =====================================================

-- Check if data can be exposed to AI
CREATE OR REPLACE FUNCTION check_ai_data_permission(
    p_user_id uuid,
    p_organization_id uuid,
    p_table_name text,
    p_column_name text DEFAULT NULL,
    p_provider ai_provider_type DEFAULT 'internal'
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_user_roles uuid[];
    v_user_groups uuid[];
    v_permission record;
    v_result jsonb;
BEGIN
    -- Get user roles and groups
    SELECT array_agg(role_id) INTO v_user_roles
    FROM get_user_roles(p_user_id, p_organization_id);

    SELECT array_agg(DISTINCT group_id) INTO v_user_groups
    FROM role_permission_groups
    WHERE role_id = ANY(v_user_roles);

    -- Find the highest priority matching permission
    SELECT * INTO v_permission
    FROM ai_data_permissions
    WHERE organization_id = p_organization_id
      AND table_name = p_table_name
      AND (column_name IS NULL OR column_name = p_column_name)
      AND is_active = true
      AND p_provider = ANY(allowed_providers)
      AND (
          (user_id = p_user_id) OR
          (role_id = ANY(v_user_roles)) OR
          (group_id = ANY(v_user_groups))
      )
    ORDER BY
        CASE WHEN user_id IS NOT NULL THEN 1
             WHEN role_id IS NOT NULL THEN 2
             WHEN group_id IS NOT NULL THEN 3
        END,
        priority DESC
    LIMIT 1;

    -- Build result
    IF v_permission IS NULL THEN
        v_result := jsonb_build_object(
            'allowed', false,
            'exposure_level', 'none',
            'reason', 'No permission found'
        );
    ELSE
        v_result := jsonb_build_object(
            'allowed', v_permission.exposure_level != 'none',
            'exposure_level', v_permission.exposure_level,
            'requires_sanitization', v_permission.requires_sanitization,
            'sanitization_method', v_permission.sanitization_method,
            'max_rows', v_permission.max_rows_in_context,
            'requires_approval', v_permission.requires_approval
        );
    END IF;

    RETURN v_result;
END;
$$;

COMMENT ON FUNCTION check_ai_data_permission IS 'Check if specific data can be exposed to AI and under what conditions';

-- Build AI context based on rules
CREATE OR REPLACE FUNCTION build_ai_context(
    p_user_id uuid,
    p_organization_id uuid,
    p_query text,
    p_operation_type varchar DEFAULT 'query'
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_context_rule record;
    v_context jsonb := '[]'::jsonb;
    v_table_data jsonb;
    v_table text;
BEGIN
    -- Find matching context rule
    SELECT * INTO v_context_rule
    FROM ai_context_rules
    WHERE organization_id = p_organization_id
      AND is_active = true
      AND (
          operation_type = p_operation_type OR
          (query_pattern IS NOT NULL AND p_query ~ query_pattern)
      )
    ORDER BY
        CASE WHEN query_pattern IS NOT NULL THEN 1 ELSE 2 END
    LIMIT 1;

    IF v_context_rule IS NULL THEN
        RETURN jsonb_build_object(
            'error', 'No matching context rule found',
            'tables', '[]'::jsonb
        );
    END IF;

    -- Build context for each included table
    FOREACH v_table IN ARRAY v_context_rule.included_tables
    LOOP
        -- Check AI permission for this table
        IF (check_ai_data_permission(
            p_user_id,
            p_organization_id,
            v_table,
            NULL,
            v_context_rule.preferred_provider
        )->>'allowed')::boolean THEN

            v_table_data := jsonb_build_object(
                'table', v_table,
                'max_rows', v_context_rule.max_rows_per_table,
                'aggregation', v_context_rule.prefer_aggregation,
                'time_window_days', v_context_rule.time_window_days
            );

            v_context := v_context || v_table_data;
        END IF;
    END LOOP;

    RETURN jsonb_build_object(
        'context_name', v_context_rule.context_name,
        'provider', v_context_rule.preferred_provider,
        'auto_sanitize', v_context_rule.auto_sanitize,
        'tables', v_context
    );
END;
$$;

COMMENT ON FUNCTION build_ai_context IS 'Build appropriate context for AI based on query and permissions';

-- Sanitize data before AI exposure
CREATE OR REPLACE FUNCTION sanitize_for_ai(
    p_organization_id uuid,
    p_table_name text,
    p_column_name text,
    p_value text,
    p_method varchar DEFAULT 'auto'
)
RETURNS text
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_rule record;
    v_sanitized text;
BEGIN
    -- Find applicable sanitization rule
    SELECT * INTO v_rule
    FROM ai_sanitization_rules
    WHERE organization_id = p_organization_id
      AND table_name = p_table_name
      AND (column_name IS NULL OR column_name = p_column_name)
      AND is_active = true
    ORDER BY
        CASE WHEN column_name IS NOT NULL THEN 1 ELSE 2 END,
        priority DESC
    LIMIT 1;

    IF v_rule IS NULL AND p_method = 'auto' THEN
        -- Default sanitization if no rule found
        IF p_value ~ '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}' THEN
            RETURN regexp_replace(p_value, '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}', '[EMAIL]', 'g');
        ELSIF p_value ~ '\d{3}-\d{2}-\d{4}' THEN
            RETURN '[SSN]';
        ELSIF p_value ~ '\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}' THEN
            RETURN '[CREDIT_CARD]';
        ELSE
            RETURN p_value;
        END IF;
    END IF;

    -- Apply sanitization based on method
    CASE v_rule.method
        WHEN 'remove' THEN
            v_sanitized := regexp_replace(p_value, v_rule.pattern, '', 'g');
        WHEN 'mask' THEN
            v_sanitized := regexp_replace(p_value, v_rule.pattern, v_rule.replacement, 'g');
        WHEN 'hash' THEN
            v_sanitized := 'HASH:' || substr(md5(p_value), 1, 8);
        WHEN 'tokenize' THEN
            v_sanitized := '[TOKEN_' || substr(md5(p_value), 1, 6) || ']';
        WHEN 'generalize' THEN
            -- Example: generalize age to ranges
            v_sanitized := CASE
                WHEN p_value::int < 18 THEN 'Under 18'
                WHEN p_value::int < 30 THEN '18-29'
                WHEN p_value::int < 50 THEN '30-49'
                ELSE '50+'
            END;
        ELSE
            v_sanitized := p_value;
    END CASE;

    RETURN v_sanitized;
END;
$$;

COMMENT ON FUNCTION sanitize_for_ai IS 'Sanitize data according to configured rules before AI exposure';

-- =====================================================
-- DEFAULT AI PERMISSIONS
-- =====================================================

-- Function to create default AI permissions for common scenarios
CREATE OR REPLACE FUNCTION create_default_ai_permissions(
    p_organization_id uuid
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    -- Never expose these tables to AI
    INSERT INTO ai_data_permissions (
        organization_id, table_name, exposure_level, allowed_providers
    ) VALUES
    (p_organization_id, 'auth_users', 'none', ARRAY[]::ai_provider_type[]),
    (p_organization_id, 'payment_methods', 'none', ARRAY[]::ai_provider_type[]),
    (p_organization_id, 'bank_accounts', 'none', ARRAY[]::ai_provider_type[]),
    (p_organization_id, 'api_keys', 'none', ARRAY[]::ai_provider_type[]);

    -- Limited exposure with sanitization
    INSERT INTO ai_data_permissions (
        organization_id, table_name, exposure_level, allowed_providers, requires_sanitization
    ) VALUES
    (p_organization_id, 'contacts', 'sanitized', ARRAY['internal']::ai_provider_type[], true),
    (p_organization_id, 'employees', 'sanitized', ARRAY['internal']::ai_provider_type[], true),
    (p_organization_id, 'invoices', 'aggregated', ARRAY['internal']::ai_provider_type[], true);

    -- Safe for AI with internal providers
    INSERT INTO ai_data_permissions (
        organization_id, table_name, exposure_level, allowed_providers, requires_sanitization
    ) VALUES
    (p_organization_id, 'products', 'full', ARRAY['internal', 'external']::ai_provider_type[], false),
    (p_organization_id, 'product_categories', 'full', ARRAY['internal', 'external']::ai_provider_type[], false),
    (p_organization_id, 'knowledge_entries', 'full', ARRAY['internal', 'external']::ai_provider_type[], false);
END;
$$;

COMMENT ON FUNCTION create_default_ai_permissions IS 'Create default AI permission rules for common tables';

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER update_ai_data_permissions_updated_at
    BEFORE UPDATE ON ai_data_permissions
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

CREATE TRIGGER update_ai_context_rules_updated_at
    BEFORE UPDATE ON ai_context_rules
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

CREATE TRIGGER update_ai_sanitization_rules_updated_at
    BEFORE UPDATE ON ai_sanitization_rules
    FOR EACH ROW EXECUTE FUNCTION update_permission_updated_at();

-- =====================================================
-- RLS POLICIES
-- =====================================================

ALTER TABLE ai_data_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_context_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_sanitization_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_data_access_log ENABLE ROW LEVEL SECURITY;

-- AI data permissions policies
CREATE POLICY "Users can view AI permissions in their org"
    ON ai_data_permissions FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage AI permissions"
    ON ai_data_permissions FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- Context rules policies
CREATE POLICY "Users can view AI context rules in their org"
    ON ai_context_rules FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage AI context rules"
    ON ai_context_rules FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- Sanitization rules policies
CREATE POLICY "Users can view sanitization rules in their org"
    ON ai_sanitization_rules FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage sanitization rules"
    ON ai_sanitization_rules FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- Access log policies
CREATE POLICY "Users can view their own AI access logs"
    ON ai_data_access_log FOR SELECT
    USING (
        user_id = (SELECT auth.uid()) OR
        organization_id IN (
            SELECT organization_id FROM organization_users
            WHERE user_id = (SELECT auth.uid())
            AND role IN ('owner', 'admin')
        )
    );

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TYPE ai_provider_type IS 'Types of AI providers (internal/external)';
COMMENT ON TYPE ai_exposure_level IS 'Levels of data exposure allowed for AI';
