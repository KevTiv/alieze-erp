-- Migration: Provider-Agnostic AI Framework
-- Description: Refactors AI functions to be provider-agnostic for flexibility
-- Created: 2025-01-01
-- Purpose: Enable easy switching between AI providers and models

-- ============================================
-- AI PROVIDER CONFIGURATION
-- ============================================

-- Extend existing ai_provider_config table to support more providers
-- and add provider-agnostic routing

-- Add new provider types to support future expansion
ALTER TYPE ai_provider_type ADD VALUE IF NOT EXISTS 'internal';
ALTER TYPE ai_provider_type ADD VALUE IF NOT EXISTS 'external';
ALTER TYPE ai_provider_type ADD VALUE IF NOT EXISTS 'hybrid';

-- ============================================
-- AI MODEL CONFIGURATION
-- ============================================

CREATE TABLE ai_model_config (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations(id) ON DELETE CASCADE,

    -- Model identification
    provider_type ai_provider_type NOT NULL,
    model_name varchar(100) NOT NULL,
    model_type varchar(50) NOT NULL, -- 'chat', 'reasoning', 'analysis', 'embedding'

    -- Model capabilities
    capabilities jsonb DEFAULT '[]'::jsonb,
    max_tokens integer DEFAULT 4096,
    max_context_length integer DEFAULT 8192,

    -- Performance characteristics
    speed varchar(20) DEFAULT 'medium', -- 'slow', 'medium', 'fast', 'instant'
    cost_per_token numeric(10,6) DEFAULT 0.000002,
    quality_rating integer DEFAULT 85, -- 1-100

    -- Usage tracking
    total_calls integer DEFAULT 0,
    total_tokens_used integer DEFAULT 0,
    last_used_at timestamptz,

    -- Status
    is_active boolean DEFAULT true,
    is_recommended boolean DEFAULT false,

    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT ai_model_config_type_check CHECK (model_type IN (
        'chat', 'reasoning', 'analysis', 'embedding', 'classification'
    ))
);

COMMENT ON TABLE ai_model_config IS
'Configuration for AI models with performance and cost tracking';

CREATE INDEX idx_ai_model_config_org ON ai_model_config(organization_id);
CREATE INDEX idx_ai_model_config_provider ON ai_model_config(organization_id, provider_type);
CREATE INDEX idx_ai_model_config_type ON ai_model_config(organization_id, model_type);

-- ============================================
-- AI PROVIDER ROUTING RULES
-- ============================================

CREATE TABLE ai_provider_routing_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Rule identification
    rule_name varchar(100) NOT NULL,
    rule_type varchar(50) NOT NULL,

    -- Conditions
    conditions jsonb DEFAULT '{}'::jsonb,

    -- Routing logic
    preferred_provider ai_provider_type,
    fallback_provider ai_provider_type,

    -- Model selection
    preferred_model varchar(100),
    fallback_model varchar(100),

    -- Priority
    priority integer DEFAULT 100,

    -- Status
    is_active boolean DEFAULT true,

    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT ai_provider_routing_rules_type_check CHECK (rule_type IN (
        'data_sensitivity', 'cost_optimization', 'speed_optimization',
        'quality_optimization', 'default', 'custom'
    ))
);

COMMENT ON TABLE ai_provider_routing_rules IS
'Rules for routing AI requests to appropriate providers based on conditions';

CREATE INDEX idx_ai_provider_routing_rules_org ON ai_provider_routing_rules(organization_id, rule_type);
CREATE INDEX idx_ai_provider_routing_rules_priority ON ai_provider_routing_rules(organization_id, priority)
    WHERE is_active = true;

-- ============================================
-- AI USAGE TRACKING (Enhanced)
-- ============================================

-- Extend existing ai_usage_audit table
ALTER TABLE ai_usage_audit
    ADD COLUMN IF NOT EXISTS model_name varchar(100),
    ADD COLUMN IF NOT EXISTS model_type varchar(50),
    ADD COLUMN IF NOT EXISTS provider_type ai_provider_type,
    ADD COLUMN IF NOT EXISTS tokens_input integer,
    ADD COLUMN IF NOT EXISTS tokens_output integer,
    ADD COLUMN IF NOT EXISTS estimated_cost numeric(10,6);

-- ============================================
-- PROVIDER-AGNOSTIC AI FUNCTIONS
-- ============================================

-- Unified AI chat function with provider routing
CREATE OR REPLACE FUNCTION ai_chat_unified(
    p_prompt text,
    p_organization_id uuid DEFAULT NULL,
    p_context jsonb DEFAULT NULL,
    p_options jsonb DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_provider ai_provider_type;
    v_model varchar(100);
    v_response jsonb;
    v_routing_rule record;
    v_data_sensitivity data_sensitivity;
    v_usage_cost numeric(10,6);
    v_tokens_used integer;
BEGIN
    -- Determine appropriate provider based on context
    IF p_context IS NOT NULL AND p_context ? 'data_sensitivity' THEN
        v_data_sensitivity := p_context->>'data_sensitivity'::data_sensitivity;
    ELSE
        v_data_sensitivity := 'internal'::data_sensitivity;
    END IF;

    -- Find matching routing rule
    SELECT * INTO v_routing_rule
    FROM ai_provider_routing_rules
    WHERE organization_id = p_organization_id
      AND is_active = true
      AND (
          (rule_type = 'data_sensitivity' AND
           (conditions->>'sensitivity_level')::data_sensitivity = v_data_sensitivity) OR
          (rule_type = 'default')
      )
    ORDER BY priority ASC
    LIMIT 1;

    -- If no specific rule found, use default routing
    IF v_routing_rule IS NULL THEN
        -- Use organization preference or system default
        SELECT config->>'default_provider' INTO v_provider
        FROM ai_provider_config
        WHERE organization_id = p_organization_id
          AND provider_type = 'groq'
          AND is_active = true
        LIMIT 1;

        v_provider := COALESCE(v_provider, 'internal'::ai_provider_type);

        -- Get model based on provider
        CASE v_provider
            WHEN 'internal' THEN v_model := 'mistral';
            WHEN 'external' THEN v_model := 'llama-3.3-70b-versatile';
            ELSE v_model := 'llama-3.3-70b-versatile';
        END CASE;
    ELSE
        v_provider := v_routing_rule.preferred_provider;
        v_model := v_routing_rule.preferred_model;
    END IF;

    -- Route to appropriate provider
    CASE v_provider
        WHEN 'internal' THEN
            -- Use Ollama (local)
            v_response := jsonb_build_object(
                'response', ollama_chat_completion(p_prompt, v_model,
                    COALESCE(p_options->>'temperature', 0.7)::float,
                    COALESCE(p_options->>'max_tokens', 1024)::int),
                'provider', 'ollama',
                'model', v_model,
                'provider_type', 'internal'
            );
            v_tokens_used := 0; -- Would be calculated from response
            v_usage_cost := 0; -- Local models have minimal cost

        WHEN 'external' THEN
            -- Use Groq (external)
            v_response := jsonb_build_object(
                'response', groq_chat_completion(p_prompt, v_model,
                    COALESCE(p_options->>'temperature', 0.7)::float,
                    COALESCE(p_options->>'max_tokens', 1024)::int),
                'provider', 'groq',
                'model', v_model,
                'provider_type', 'external'
            );
            v_tokens_used := 0; -- Would be calculated from response
            v_usage_cost := 0.000002 * v_tokens_used; -- Example cost

        WHEN 'hybrid' THEN
            -- Try internal first, fallback to external
            BEGIN
                v_response := jsonb_build_object(
                    'response', ollama_chat_completion(p_prompt, v_model,
                        COALESCE(p_options->>'temperature', 0.7)::float,
                        COALESCE(p_options->>'max_tokens', 1024)::int),
                    'provider', 'ollama',
                    'model', v_model,
                    'provider_type', 'hybrid'
                );
                v_tokens_used := 0;
                v_usage_cost := 0;
            EXCEPTION WHEN OTHERS THEN
                v_response := jsonb_build_object(
                    'response', groq_chat_completion(p_prompt, v_model,
                        COALESCE(p_options->>'temperature', 0.7)::float,
                        COALESCE(p_options->>'max_tokens', 1024)::int),
                    'provider', 'groq',
                    'model', v_model,
                    'provider_type', 'hybrid',
                    'fallback_used', true
                );
                v_tokens_used := 0;
                v_usage_cost := 0.000002 * v_tokens_used;
            END;

        ELSE
            -- Default to internal for safety
            v_response := jsonb_build_object(
                'response', ollama_chat_completion(p_prompt, v_model,
                    COALESCE(p_options->>'temperature', 0.7)::float,
                    COALESCE(p_options->>'max_tokens', 1024)::int),
                'provider', 'ollama',
                'model', v_model,
                'provider_type', 'internal'
            );
            v_tokens_used := 0;
            v_usage_cost := 0;
    END CASE;

    -- Log usage for cost tracking
    INSERT INTO ai_usage_audit (
        organization_id, user_id, provider, model, data_classification,
        contained_pii, was_sanitized, prompt_hash, response_hash,
        tokens_used, processing_time_ms, error, created_at,
        model_name, model_type, provider_type, tokens_input, tokens_output, estimated_cost
    ) VALUES (
        p_organization_id,
        (SELECT auth.uid()),
        v_response->>'provider',
        v_model,
        v_data_sensitivity,
        false, -- Would be determined from context
        false, -- Would be determined from context
        encode(digest(p_prompt, 'sha256'), 'hex'),
        encode(digest(v_response->>'response', 'sha256'), 'hex'),
        v_tokens_used,
        0, -- Would be measured
        NULL,
        now(),
        v_model,
        'chat',
        v_provider,
        v_tokens_used / 2, -- Estimate input tokens
        v_tokens_used / 2, -- Estimate output tokens
        v_usage_cost
    );

    RETURN v_response;
END;
$$;

COMMENT ON FUNCTION ai_chat_unified IS
'Unified AI chat function with intelligent provider routing based on data sensitivity and cost';

-- Provider-agnostic insight generation
CREATE OR REPLACE FUNCTION ai_generate_insight(
    p_prompt text,
    p_organization_id uuid DEFAULT NULL,
    p_insight_type varchar DEFAULT 'general',
    p_options jsonb DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_context jsonb;
    v_response jsonb;
BEGIN
    -- Build context based on insight type
    CASE p_insight_type
        WHEN 'business_health' THEN
            v_context := jsonb_build_object(
                'data_sensitivity', 'internal',
                'insight_type', 'business_health',
                'description', 'Generate business health score and analysis'
            );

        WHEN 'data_quality' THEN
            v_context := jsonb_build_object(
                'data_sensitivity', 'internal',
                'insight_type', 'data_quality',
                'description', 'Analyze data quality issues'
            );

        WHEN 'inventory' THEN
            v_context := jsonb_build_object(
                'data_sensitivity', 'internal',
                'insight_type', 'inventory',
                'description', 'Provide inventory insights and recommendations'
            );

        WHEN 'customer' THEN
            v_context := jsonb_build_object(
                'data_sensitivity', 'confidential',
                'insight_type', 'customer',
                'description', 'Analyze customer behavior patterns'
            );

        WHEN 'financial' THEN
            v_context := jsonb_build_object(
                'data_sensitivity', 'restricted',
                'insight_type', 'financial',
                'description', 'Provide financial analysis and insights'
            );

        ELSE
            v_context := jsonb_build_object(
                'data_sensitivity', 'internal',
                'insight_type', 'general',
                'description', 'General business insight generation'
            );
    END CASE;

    -- Use unified chat function
    v_response := ai_chat_unified(p_prompt, p_organization_id, v_context, p_options);

    -- Add insight-specific metadata
    v_response := jsonb_set(v_response, '{insight_type}', to_jsonb(p_insight_type));
    v_response := jsonb_set(v_response, '{generated_at}', to_jsonb(now()));

    RETURN v_response;
END;
$$;

COMMENT ON FUNCTION ai_generate_insight IS
'Generate business insights with appropriate provider routing based on sensitivity';

-- Provider-agnostic data analysis
CREATE OR REPLACE FUNCTION ai_analyze_data(
    p_data jsonb,
    p_analysis_type varchar,
    p_organization_id uuid DEFAULT NULL,
    p_options jsonb DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_prompt text;
    v_context jsonb;
    v_response jsonb;
BEGIN
    -- Build analysis prompt based on type
    CASE p_analysis_type
        WHEN 'column_mapping' THEN
            v_prompt := format(
                'Analyze this data structure and map columns to our database schema:

                Data Sample: %s

                Return JSON with column mappings and confidence scores.',
                p_data::text
            );
            v_context := jsonb_build_object('data_sensitivity', 'metadata');

        WHEN 'data_validation' THEN
            v_prompt := format(
                'Validate this data and identify potential issues:

                Data: %s

                Return JSON with validation results, errors, and warnings.',
                p_data::text
            );
            v_context := jsonb_build_object('data_sensitivity', 'internal');

        WHEN 'trend_analysis' THEN
            v_prompt := format(
                'Analyze trends in this business data:

                Data: %s

                Return JSON with trend analysis, patterns, and recommendations.',
                p_data::text
            );
            v_context := jsonb_build_object('data_sensitivity', 'internal');

        WHEN 'risk_detection' THEN
            v_prompt := format(
                'Detect potential risks in this business data:

                Data: %s

                Return JSON with risk assessment and mitigation suggestions.',
                p_data::text
            );
            v_context := jsonb_build_object('data_sensitivity', 'confidential');

        ELSE
            v_prompt := format(
                'Analyze this business data:

                Data: %s

                Return JSON with insights and recommendations.',
                p_data::text
            );
            v_context := jsonb_build_object('data_sensitivity', 'internal');
    END CASE;

    -- Use unified chat function for analysis
    v_response := ai_chat_unified(v_prompt, p_organization_id, v_context, p_options);

    -- Add analysis-specific metadata
    v_response := jsonb_set(v_response, '{analysis_type}', to_jsonb(p_analysis_type));
    v_response := jsonb_set(v_response, '{analyzed_at}', to_jsonb(now()));

    RETURN v_response;
END;
$$;

COMMENT ON FUNCTION ai_analyze_data IS
'Analyze business data with appropriate provider routing based on sensitivity';

-- Provider-agnostic text generation
CREATE OR REPLACE FUNCTION ai_generate_text(
    p_template text,
    p_data jsonb,
    p_organization_id uuid DEFAULT NULL,
    p_options jsonb DEFAULT NULL
) RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
    v_prompt text;
    v_context jsonb;
    v_response jsonb;
BEGIN
    -- Build prompt from template and data
    v_prompt := format(p_template, p_data::text);

    -- Determine context based on template
    IF p_template LIKE '%customer%' OR p_template LIKE '%contact%' THEN
        v_context := jsonb_build_object('data_sensitivity', 'confidential');
    ELSIF p_template LIKE '%financial%' OR p_template LIKE '%payment%' THEN
        v_context := jsonb_build_object('data_sensitivity', 'restricted');
    ELSE
        v_context := jsonb_build_object('data_sensitivity', 'internal');
    END IF;

    -- Use unified chat function
    v_response := ai_chat_unified(v_prompt, p_organization_id, v_context, p_options);

    RETURN v_response->>'response';
END;
$$;

COMMENT ON FUNCTION ai_generate_text IS
'Generate text from templates with appropriate provider routing';

-- ============================================
-- PROVIDER ROUTING HELPER FUNCTIONS
-- ============================================

-- Get best provider for given data sensitivity
CREATE OR REPLACE FUNCTION get_best_ai_provider(
    p_organization_id uuid,
    p_data_sensitivity data_sensitivity,
    p_task_type varchar DEFAULT 'general'
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_rule record;
    v_provider ai_provider_type;
    v_model varchar(100);
    v_reason text;
BEGIN
    -- Find matching routing rule
    SELECT * INTO v_rule
    FROM ai_provider_routing_rules
    WHERE organization_id = p_organization_id
      AND is_active = true
      AND (
          (rule_type = 'data_sensitivity' AND
           (conditions->>'sensitivity_level')::data_sensitivity = p_data_sensitivity) OR
          (rule_type = 'default')
      )
    ORDER BY priority ASC
    LIMIT 1;

    IF v_rule IS NOT NULL THEN
        v_provider := v_rule.preferred_provider;
        v_model := v_rule.preferred_model;
        v_reason := 'Matching routing rule found';
    ELSE
        -- Use system defaults based on sensitivity
        CASE p_data_sensitivity
            WHEN 'restricted' THEN
                v_provider := 'internal';
                v_model := 'mistral';
                v_reason := 'Restricted data requires internal processing';

            WHEN 'confidential' THEN
                v_provider := 'internal';
                v_model := 'mistral';
                v_reason := 'Confidential data prefers internal processing';

            WHEN 'internal' THEN
                v_provider := 'hybrid';
                v_model := 'mistral';
                v_reason := 'Internal data can use hybrid approach';

            WHEN 'public' THEN
                v_provider := 'external';
                v_model := 'llama-3.3-70b-versatile';
                v_reason := 'Public data can use external providers';

            ELSE
                v_provider := 'internal';
                v_model := 'mistral';
                v_reason := 'Default to internal for safety';
        END CASE;
    END IF;

    RETURN jsonb_build_object(
        'provider', v_provider,
        'model', v_model,
        'reason', v_reason,
        'sensitivity', p_data_sensitivity,
        'task_type', p_task_type
    );
END;
$$;

COMMENT ON FUNCTION get_best_ai_provider IS
'Determine best AI provider based on data sensitivity and task type';

-- ============================================
-- DEFAULT ROUTING RULES SETUP
-- ============================================

CREATE OR REPLACE FUNCTION setup_default_ai_routing(
    p_organization_id uuid
) RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    -- Data sensitivity rules
    INSERT INTO ai_provider_routing_rules (
        organization_id, rule_name, rule_type, conditions,
        preferred_provider, fallback_provider, preferred_model, priority
    ) VALUES
    (
        p_organization_id, 'Restricted Data Rule', 'data_sensitivity',
        '{"sensitivity_level": "restricted"}',
        'internal', 'internal', 'mistral', 10
    ),
    (
        p_organization_id, 'Confidential Data Rule', 'data_sensitivity',
        '{"sensitivity_level": "confidential"}',
        'internal', 'internal', 'mistral', 20
    ),
    (
        p_organization_id, 'Internal Data Rule', 'data_sensitivity',
        '{"sensitivity_level": "internal"}',
        'hybrid', 'internal', 'mistral', 30
    ),
    (
        p_organization_id, 'Public Data Rule', 'data_sensitivity',
        '{"sensitivity_level": "public"}',
        'external', 'internal', 'llama-3.3-70b-versatile', 40
    ),
    (
        p_organization_id, 'Default Rule', 'default', '{}',
        'internal', 'internal', 'mistral', 100
    );

    -- AI model configurations
    INSERT INTO ai_model_config (
        organization_id, provider_type, model_name, model_type,
        capabilities, speed, cost_per_token, quality_rating, is_recommended
    ) VALUES
    (
        p_organization_id, 'internal', 'mistral', 'chat',
        '{"tasks": ["chat", "analysis", "general"]}', 'fast', 0.000000, 85, true
    ),
    (
        p_organization_id, 'internal', 'llama3', 'reasoning',
        '{"tasks": ["reasoning", "analysis", "complex"]}', 'medium', 0.000000, 88, true
    ),
    (
        p_organization_id, 'external', 'llama-3.3-70b-versatile', 'chat',
        '{"tasks": ["chat", "analysis", "general"]}', 'fast', 0.000002, 92, true
    ),
    (
        p_organization_id, 'external', 'openai/gpt-oss-120b', 'reasoning',
        '{"tasks": ["reasoning", "analysis", "complex"]}', 'medium', 0.000003, 95, true
    );
END;
$$;

COMMENT ON FUNCTION setup_default_ai_routing IS
'Set up default AI provider routing rules and model configurations';

-- ============================================
-- TRIGGERS
-- ============================================

CREATE OR REPLACE FUNCTION update_ai_model_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_ai_model_config_updated_at
    BEFORE UPDATE ON ai_model_config
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_model_updated_at();

CREATE TRIGGER set_ai_provider_routing_rules_updated_at
    BEFORE UPDATE ON ai_provider_routing_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_model_updated_at();

-- ============================================
-- RLS POLICIES
-- ============================================

ALTER TABLE ai_model_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_provider_routing_rules ENABLE ROW LEVEL SECURITY;

-- AI Model Config Policies
CREATE POLICY "Users can view AI models in their org"
    ON ai_model_config FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage AI models"
    ON ai_model_config FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- AI Provider Routing Policies
CREATE POLICY "Users can view AI routing rules in their org"
    ON ai_provider_routing_rules FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage AI routing rules"
    ON ai_provider_routing_rules FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
        AND role IN ('owner', 'admin')
    ));

-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT ON ai_model_config TO authenticated;
GRANT ALL ON ai_model_config TO service_role;

GRANT SELECT ON ai_provider_routing_rules TO authenticated;
GRANT ALL ON ai_provider_routing_rules TO service_role;

GRANT EXECUTE ON FUNCTION ai_chat_unified(text, uuid, jsonb, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_generate_insight(text, uuid, varchar, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_analyze_data(jsonb, varchar, uuid, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_generate_text(text, jsonb, uuid, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION get_best_ai_provider(uuid, data_sensitivity, varchar) TO authenticated;
GRANT EXECUTE ON FUNCTION setup_default_ai_routing(uuid) TO authenticated;

-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Set up default AI routing for organization
SELECT setup_default_ai_routing('your-org-id');

-- 2. Use unified AI chat (automatic provider selection)
SELECT ai_chat_unified(
    'What are the key benefits of an ERP system?',
    'your-org-id',
    '{"data_sensitivity": "public"}',
    '{"temperature": 0.7, "max_tokens": 1024}'
);

-- 3. Generate business insights
SELECT ai_generate_insight(
    'Analyze our current sales pipeline and provide recommendations',
    'your-org-id',
    'business_health'
);

-- 4. Analyze data
SELECT ai_analyze_data(
    '[{"name": "Acme Corp", "email": "contact@acme.com"}]',
    'column_mapping',
    'your-org-id'
);

-- 5. Generate text from template
SELECT ai_generate_text(
    'Write a friendly follow-up email to %s about %s',
    '{"contact@acme.com", "our new product"}',
    'your-org-id'
);

-- 6. Get best provider for specific task
SELECT get_best_ai_provider('your-org-id', 'confidential', 'customer_analysis');

-- 7. Add custom routing rule
INSERT INTO ai_provider_routing_rules (
    organization_id, rule_name, rule_type, conditions,
    preferred_provider, fallback_provider, preferred_model, priority
) VALUES (
    'your-org-id', 'High Quality Analysis', 'quality_optimization',
    '{"min_quality": 90}', 'external', 'internal', 'openai/gpt-oss-120b', 15
);

-- 8. Add new AI model
INSERT INTO ai_model_config (
    organization_id, provider_type, model_name, model_type,
    capabilities, speed, cost_per_token, quality_rating, is_recommended
) VALUES (
    'your-org-id', 'external', 'anthropic-claude-3', 'reasoning',
    '{"tasks": ["reasoning", "analysis", "complex"]}', 'medium', 0.000003, 96, true
);
*/
