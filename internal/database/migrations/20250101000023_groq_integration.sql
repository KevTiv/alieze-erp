-- Migration: Groq API Integration
-- Description: Integrates Groq API for high-speed LLM inference (Compound AI, GPT-OSS-120B)
-- Created: 2025-01-01
-- Dependencies: Requires http extension and Groq API key

-- Enable HTTP extension (should already be enabled)
CREATE EXTENSION IF NOT EXISTS http;

-- ============================================
-- GROQ API CONFIGURATION TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS ai_provider_config (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations(id) ON DELETE CASCADE,
    provider_type text NOT NULL CHECK (provider_type IN ('ollama', 'groq', 'openai', 'anthropic')),
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_active boolean DEFAULT true,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE(organization_id, provider_type)
);

COMMENT ON TABLE ai_provider_config IS
'Configuration for AI providers per organization';

-- Enable RLS
ALTER TABLE ai_provider_config ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY "Users can view their org AI provider config"
    ON ai_provider_config FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage AI provider config"
    ON ai_provider_config FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid()) AND role IN ('admin', 'owner')
    ));


-- ============================================
-- GROQ CHAT COMPLETION FUNCTIONS
-- ============================================

-- Function to call Groq API for chat completions
CREATE OR REPLACE FUNCTION groq_chat_completion(
    p_prompt text,
    p_model text DEFAULT 'llama-3.3-70b-versatile',
    p_temperature float DEFAULT 0.7,
    p_max_tokens int DEFAULT 1024,
    p_api_key text DEFAULT NULL
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_response http_response;
    v_response_json jsonb;
    v_groq_url text := 'https://api.groq.com/openai/v1/chat/completions';
    v_api_key text;
BEGIN
    -- Get API key from parameter or environment
    v_api_key := COALESCE(p_api_key, current_setting('app.groq_api_key', true));

    IF v_api_key IS NULL OR v_api_key = '' THEN
        RAISE EXCEPTION 'Groq API key not configured';
    END IF;

    -- Validate input
    IF p_prompt IS NULL OR trim(p_prompt) = '' THEN
        RAISE EXCEPTION 'Prompt cannot be empty';
    END IF;

    -- Call Groq API
    v_response := http((
        'POST',
        v_groq_url,
        ARRAY[
            http_header('Content-Type', 'application/json'),
            http_header('Authorization', 'Bearer ' || v_api_key)
        ],
        'application/json',
        jsonb_build_object(
            'model', p_model,
            'messages', jsonb_build_array(
                jsonb_build_object(
                    'role', 'user',
                    'content', p_prompt
                )
            ),
            'temperature', p_temperature,
            'max_tokens', p_max_tokens
        )::text
    )::http_request);

    -- Check response status
    IF v_response.status != 200 THEN
        RAISE EXCEPTION 'Groq API error: % - %', v_response.status, v_response.content;
    END IF;

    -- Parse response
    v_response_json := v_response.content::jsonb;

    -- Return the generated response
    RETURN v_response_json->'choices'->0->'message'->>'content';
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Failed to generate completion via Groq: %', SQLERRM;
        RETURN NULL;
END;
$$;

COMMENT ON FUNCTION groq_chat_completion IS
'Generate chat completions using Groq API (Llama 3.3, GPT-OSS-120B)';


-- Function for Groq Compound AI (agentic reasoning)
CREATE OR REPLACE FUNCTION groq_compound_reasoning(
    p_prompt text,
    p_system_prompt text DEFAULT 'You are a helpful AI assistant that can use tools and research to provide accurate answers.',
    p_max_tokens int DEFAULT 2048,
    p_api_key text DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_response http_response;
    v_response_json jsonb;
    v_groq_url text := 'https://api.groq.com/openai/v1/chat/completions';
    v_api_key text;
BEGIN
    v_api_key := COALESCE(p_api_key, current_setting('app.groq_api_key', true));

    IF v_api_key IS NULL OR v_api_key = '' THEN
        RAISE EXCEPTION 'Groq API key not configured';
    END IF;

    IF p_prompt IS NULL OR trim(p_prompt) = '' THEN
        RAISE EXCEPTION 'Prompt cannot be empty';
    END IF;

    -- Call Groq Compound API
    v_response := http((
        'POST',
        v_groq_url,
        ARRAY[
            http_header('Content-Type', 'application/json'),
            http_header('Authorization', 'Bearer ' || v_api_key)
        ],
        'application/json',
        jsonb_build_object(
            'model', 'compound',
            'messages', jsonb_build_array(
                jsonb_build_object('role', 'system', 'content', p_system_prompt),
                jsonb_build_object('role', 'user', 'content', p_prompt)
            ),
            'max_tokens', p_max_tokens
        )::text
    )::http_request);

    IF v_response.status != 200 THEN
        RAISE EXCEPTION 'Groq Compound API error: % - %', v_response.status, v_response.content;
    END IF;

    v_response_json := v_response.content::jsonb;

    -- Return full response with reasoning steps
    RETURN jsonb_build_object(
        'content', v_response_json->'choices'->0->'message'->>'content',
        'reasoning_steps', v_response_json->'choices'->0->'message'->'reasoning_content',
        'usage', v_response_json->'usage',
        'model', v_response_json->>'model'
    );
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Failed to generate Compound reasoning via Groq: %', SQLERRM;
        RETURN jsonb_build_object('error', SQLERRM);
END;
$$;

COMMENT ON FUNCTION groq_compound_reasoning IS
'Use Groq Compound AI for agentic reasoning with tool use and research capabilities';


-- Function for GPT-OSS-120B reasoning
CREATE OR REPLACE FUNCTION groq_gpt_oss_reasoning(
    p_prompt text,
    p_thinking_budget int DEFAULT 10000,
    p_temperature float DEFAULT 0.3,
    p_api_key text DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_response http_response;
    v_response_json jsonb;
    v_groq_url text := 'https://api.groq.com/openai/v1/chat/completions';
    v_api_key text;
BEGIN
    v_api_key := COALESCE(p_api_key, current_setting('app.groq_api_key', true));

    IF v_api_key IS NULL OR v_api_key = '' THEN
        RAISE EXCEPTION 'Groq API key not configured';
    END IF;

    -- Call Groq GPT-OSS-120B
    v_response := http((
        'POST',
        v_groq_url,
        ARRAY[
            http_header('Content-Type', 'application/json'),
            http_header('Authorization', 'Bearer ' || v_api_key)
        ],
        'application/json',
        jsonb_build_object(
            'model', 'openai/gpt-oss-120b',
            'messages', jsonb_build_array(
                jsonb_build_object('role', 'user', 'content', p_prompt)
            ),
            'temperature', p_temperature,
            'thinking_budget', p_thinking_budget,
            'include_reasoning', true
        )::text
    )::http_request);

    IF v_response.status != 200 THEN
        RAISE EXCEPTION 'Groq GPT-OSS API error: % - %', v_response.status, v_response.content;
    END IF;

    v_response_json := v_response.content::jsonb;

    RETURN jsonb_build_object(
        'content', v_response_json->'choices'->0->'message'->>'content',
        'reasoning', v_response_json->'choices'->0->'message'->>'reasoning_content',
        'usage', v_response_json->'usage'
    );
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Failed to generate GPT-OSS reasoning via Groq: %', SQLERRM;
        RETURN jsonb_build_object('error', SQLERRM);
END;
$$;

COMMENT ON FUNCTION groq_gpt_oss_reasoning IS
'Use GPT-OSS-120B via Groq for advanced reasoning tasks';


-- ============================================
-- MULTI-PROVIDER ROUTING FUNCTION
-- ============================================

-- Smart routing function that chooses the best provider
CREATE OR REPLACE FUNCTION ai_chat(
    p_prompt text,
    p_organization_id uuid DEFAULT NULL,
    p_provider text DEFAULT 'auto',
    p_temperature float DEFAULT 0.7,
    p_max_tokens int DEFAULT 1024
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_provider text;
    v_config jsonb;
BEGIN
    -- Determine provider
    IF p_provider = 'auto' THEN
        -- Get org preference or use environment default
        IF p_organization_id IS NOT NULL THEN
            SELECT config->>'chat_provider' INTO v_provider
            FROM ai_provider_config
            WHERE organization_id = p_organization_id
              AND provider_type = 'groq'
              AND is_active = true
            LIMIT 1;
        END IF;

        v_provider := COALESCE(
            v_provider,
            current_setting('app.ai_chat_provider', true),
            'ollama'
        );
    ELSE
        v_provider := p_provider;
    END IF;

    -- Route to appropriate provider
    CASE v_provider
        WHEN 'groq' THEN
            RETURN groq_chat_completion(p_prompt, 'llama-3.3-70b-versatile', p_temperature, p_max_tokens);
        WHEN 'ollama' THEN
            RETURN ollama_chat_completion(p_prompt, 'mistral', p_temperature, p_max_tokens);
        ELSE
            RAISE EXCEPTION 'Unknown AI provider: %', v_provider;
    END CASE;
END;
$$;

COMMENT ON FUNCTION ai_chat IS
'Smart routing function for chat completions across multiple AI providers';


-- ============================================
-- BUSINESS LOGIC HELPERS
-- ============================================

-- Invoice data extraction with choice of provider
CREATE OR REPLACE FUNCTION extract_invoice_data(
    p_invoice_text text,
    p_provider text DEFAULT 'groq'
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_prompt text;
    v_response text;
    v_json_match text;
BEGIN
    v_prompt := format(
        'Extract the following information from this invoice and return ONLY valid JSON (no explanations):
        {
          "invoice_number": "string",
          "invoice_date": "YYYY-MM-DD",
          "due_date": "YYYY-MM-DD",
          "vendor_name": "string",
          "total_amount": number,
          "currency": "string",
          "line_items": [
            {
              "description": "string",
              "quantity": number,
              "unit_price": number,
              "total": number
            }
          ]
        }

        Invoice text:
        %s', p_invoice_text
    );

    -- Use Groq GPT-OSS for better accuracy on structured extraction
    IF p_provider = 'groq' THEN
        v_response := groq_chat_completion(v_prompt, 'openai/gpt-oss-120b', 0.1, 2000);
    ELSE
        v_response := ollama_chat_completion(v_prompt, 'mistral', 0.1, 2000);
    END IF;

    -- Extract JSON from response
    v_json_match := substring(v_response from '\{[^\}]*\}');

    IF v_json_match IS NULL THEN
        RAISE EXCEPTION 'Could not extract JSON from AI response';
    END IF;

    RETURN v_json_match::jsonb;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object('error', SQLERRM, 'raw_response', v_response);
END;
$$;

COMMENT ON FUNCTION extract_invoice_data IS
'Extract structured data from invoice text using AI (Groq GPT-OSS-120B recommended)';


-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT ON ai_provider_config TO authenticated;
GRANT ALL ON ai_provider_config TO service_role;

GRANT EXECUTE ON FUNCTION groq_chat_completion(text, text, float, int, text) TO authenticated;
GRANT EXECUTE ON FUNCTION groq_compound_reasoning(text, text, int, text) TO authenticated;
GRANT EXECUTE ON FUNCTION groq_gpt_oss_reasoning(text, int, float, text) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_chat(text, uuid, text, float, int) TO authenticated;
GRANT EXECUTE ON FUNCTION extract_invoice_data(text, text) TO authenticated;

GRANT EXECUTE ON FUNCTION groq_chat_completion(text, text, float, int, text) TO service_role;
GRANT EXECUTE ON FUNCTION groq_compound_reasoning(text, text, int, text) TO service_role;
GRANT EXECUTE ON FUNCTION groq_gpt_oss_reasoning(text, int, float, text) TO service_role;
GRANT EXECUTE ON FUNCTION ai_chat(text, uuid, text, float, int) TO service_role;
GRANT EXECUTE ON FUNCTION extract_invoice_data(text, text) TO service_role;


-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Configure Groq API key (set in app config or .env)
-- ALTER DATABASE postgres SET app.groq_api_key = 'gsk_...';

-- 2. Simple chat with Groq Llama 3.3
SELECT groq_chat_completion('What are the key benefits of an ERP system?');

-- 3. Advanced reasoning with GPT-OSS-120B
SELECT groq_gpt_oss_reasoning(
    'Analyze the financial health of a company with revenue $5M, expenses $4.5M, and growth rate 20% YoY',
    15000  -- thinking budget
);

-- 4. Compound AI for research tasks
SELECT groq_compound_reasoning(
    'Research the current best practices for inventory management in 2025',
    'You are an ERP consultant. Provide detailed, actionable insights.',
    3000
);

-- 5. Smart routing (auto-selects best provider)
SELECT ai_chat('Explain how to create a sales order in 3 steps');

-- 6. Extract invoice data
SELECT extract_invoice_data('
Invoice #12345
Date: January 15, 2025
Due: February 15, 2025
Vendor: ACME Corporation

Items:
- Server Rack 42U x2 @ $1,200 = $2,400
- Network Switch x1 @ $800 = $800

Total: $3,200
', 'groq');

-- 7. Configure org AI preferences
INSERT INTO ai_provider_config (organization_id, provider_type, config)
VALUES (
    'your-org-id',
    'groq',
    '{"chat_provider": "groq", "reasoning_model": "openai/gpt-oss-120b"}'::jsonb
);
*/
