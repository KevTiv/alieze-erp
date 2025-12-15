-- Migration: Data Privacy & Sanitization Layer
-- Description: Implements data classification, PII detection, and safe AI routing
-- Created: 2025-01-01
-- Purpose: Ensure customer data never leaks to external AI providers

-- ============================================
-- DATA CLASSIFICATION
-- ============================================

-- Data sensitivity levels
CREATE TYPE data_sensitivity AS ENUM (
    'public',      -- Can be shared externally (product descriptions)
    'internal',    -- Company internal (sales data, inventory)
    'confidential',-- Customer data, financial records
    'restricted'   -- Highly sensitive (PII, PHI, payment info)
);

-- Table to classify data types
CREATE TABLE data_classification_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name text NOT NULL,
    column_name text NOT NULL,
    sensitivity data_sensitivity NOT NULL,
    contains_pii boolean DEFAULT false,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE(table_name, column_name)
);

COMMENT ON TABLE data_classification_rules IS
'Defines sensitivity level for each table/column to enforce privacy routing';

-- Populate initial classification rules
INSERT INTO data_classification_rules (table_name, column_name, sensitivity, contains_pii, description) VALUES
-- RESTRICTED (Never send to external APIs)
('contacts', 'email', 'restricted', true, 'Customer email addresses'),
('contacts', 'phone', 'restricted', true, 'Customer phone numbers'),
('contacts', 'mobile', 'restricted', true, 'Customer mobile numbers'),
('invoices', 'amount_total', 'restricted', false, 'Invoice amounts'),
('invoices', 'payment_reference', 'restricted', true, 'Payment references'),

-- CONFIDENTIAL (Ollama only, unless sanitized)
('contacts', 'name', 'confidential', true, 'Customer names'),
('contacts', 'comment', 'confidential', false, 'Customer notes'),
('sales_orders', 'client_order_ref', 'confidential', false, 'Client references'),
('products', 'default_code', 'confidential', false, 'Product codes'),

-- INTERNAL (Prefer Ollama, Groq acceptable if sanitized)
('products', 'name', 'internal', false, 'Product names'),
('products', 'list_price', 'internal', false, 'Product prices'),
('sales_orders', 'amount_total', 'internal', false, 'Order totals'),

-- PUBLIC (Can use Groq for speed)
('products', 'description', 'public', false, 'Product descriptions (public)'),
('knowledge_entries', 'body_markdown', 'public', false, 'Public knowledge base');


-- ============================================
-- PII DETECTION
-- ============================================

-- Function to detect PII in text using regex patterns
CREATE OR REPLACE FUNCTION contains_pii(text_content text)
RETURNS boolean
LANGUAGE plpgsql
IMMUTABLE
AS $$
BEGIN
    IF text_content IS NULL THEN
        RETURN false;
    END IF;

    -- Check for common PII patterns
    RETURN text_content ~* ANY(ARRAY[
        -- Email addresses
        '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}',
        -- Phone numbers (various formats)
        '(\+?[0-9]{1,3}[\s.-]?)?\(?[0-9]{3}\)?[\s.-]?[0-9]{3}[\s.-]?[0-9]{4}',
        -- SSN (US)
        '[0-9]{3}-[0-9]{2}-[0-9]{4}',
        -- Credit card (basic pattern)
        '[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}',
        -- IP addresses
        '([0-9]{1,3}\.){3}[0-9]{1,3}',
        -- Dates (potential DOB)
        '[0-9]{2}/[0-9]{2}/[0-9]{4}',
        -- IBAN
        '[A-Z]{2}[0-9]{2}[A-Z0-9]{10,30}'
    ]);
END;
$$;

COMMENT ON FUNCTION contains_pii IS
'Detects common PII patterns in text using regex. Returns true if PII found.';


-- Function to classify text sensitivity
CREATE OR REPLACE FUNCTION classify_text_sensitivity(
    p_table_name text,
    p_column_name text,
    p_text_content text
)
RETURNS data_sensitivity
LANGUAGE plpgsql
AS $$
DECLARE
    v_sensitivity data_sensitivity;
    v_marked_pii boolean;
BEGIN
    -- Check classification rules first
    SELECT sensitivity, contains_pii
    INTO v_sensitivity, v_marked_pii
    FROM data_classification_rules
    WHERE table_name = p_table_name
      AND column_name = p_column_name;

    -- If column is marked as containing PII, return restricted
    IF v_marked_pii THEN
        RETURN 'restricted';
    END IF;

    -- If found in rules, return that sensitivity
    IF v_sensitivity IS NOT NULL THEN
        RETURN v_sensitivity;
    END IF;

    -- Otherwise, scan content for PII
    IF contains_pii(p_text_content) THEN
        RETURN 'restricted';
    END IF;

    -- Default to confidential (safe choice)
    RETURN 'confidential';
END;
$$;

COMMENT ON FUNCTION classify_text_sensitivity IS
'Classifies text sensitivity based on rules and PII detection';


-- ============================================
-- PII SANITIZATION
-- ============================================

-- Sanitization mapping table (temporary, session-scoped)
CREATE TABLE pii_sanitization_map (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id text NOT NULL,
    pii_type text NOT NULL,
    original_value text NOT NULL,
    sanitized_token text NOT NULL,
    created_at timestamptz DEFAULT now(),
    expires_at timestamptz DEFAULT now() + interval '1 hour'
);

CREATE INDEX idx_pii_sanitization_session ON pii_sanitization_map(session_id);

COMMENT ON TABLE pii_sanitization_map IS
'Temporary mapping for PII sanitization (auto-expires after 1 hour)';


-- Function to sanitize text by replacing PII with tokens
CREATE OR REPLACE FUNCTION sanitize_pii(
    p_text text,
    p_session_id text DEFAULT gen_random_uuid()::text
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_sanitized_text text;
    v_matches text[];
    v_match text;
    v_token text;
    v_token_counter int := 0;
    v_pii_type text;
BEGIN
    v_sanitized_text := p_text;

    -- Replace emails
    FOR v_match IN
        SELECT regexp_matches[1]
        FROM regexp_matches(p_text, '([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})', 'g') AS regexp_matches
    LOOP
        v_token_counter := v_token_counter + 1;
        v_token := '[EMAIL_' || v_token_counter || ']';

        INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
        VALUES (p_session_id, 'email', v_match, v_token);

        v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
    END LOOP;

    -- Replace phone numbers
    v_token_counter := 0;
    FOR v_match IN
        SELECT regexp_matches[1]
        FROM regexp_matches(p_text, '((\+?[0-9]{1,3}[\s.-]?)?\(?[0-9]{3}\)?[\s.-]?[0-9]{3}[\s.-]?[0-9]{4})', 'g') AS regexp_matches
    LOOP
        v_token_counter := v_token_counter + 1;
        v_token := '[PHONE_' || v_token_counter || ']';

        INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
        VALUES (p_session_id, 'phone', v_match, v_token);

        v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
    END LOOP;

    -- Return both sanitized text and session ID for rehydration
    RETURN jsonb_build_object(
        'sanitized_text', v_sanitized_text,
        'session_id', p_session_id,
        'had_pii', v_sanitized_text != p_text
    );
END;
$$;

COMMENT ON FUNCTION sanitize_pii IS
'Replaces PII in text with tokens. Returns sanitized text and session ID for rehydration.';


-- Function to rehydrate sanitized text
CREATE OR REPLACE FUNCTION rehydrate_pii(
    p_sanitized_text text,
    p_session_id text
)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
    v_rehydrated_text text;
    v_mapping record;
BEGIN
    v_rehydrated_text := p_sanitized_text;

    -- Replace all tokens with original values
    FOR v_mapping IN
        SELECT sanitized_token, original_value
        FROM pii_sanitization_map
        WHERE session_id = p_session_id
    LOOP
        v_rehydrated_text := replace(v_rehydrated_text, v_mapping.sanitized_token, v_mapping.original_value);
    END LOOP;

    -- Clean up mapping (optional, expires automatically)
    DELETE FROM pii_sanitization_map WHERE session_id = p_session_id;

    RETURN v_rehydrated_text;
END;
$$;

COMMENT ON FUNCTION rehydrate_pii IS
'Replaces tokens with original PII values using session mapping';


-- ============================================
-- SAFE AI ROUTING
-- ============================================

-- Determine safe provider for given data sensitivity
CREATE OR REPLACE FUNCTION get_safe_ai_provider(
    p_sensitivity data_sensitivity,
    p_allow_external boolean DEFAULT false
)
RETURNS text
LANGUAGE plpgsql
AS $$
BEGIN
    -- RESTRICTED data: Ollama ONLY (never external)
    IF p_sensitivity = 'restricted' THEN
        RETURN 'ollama';
    END IF;

    -- CONFIDENTIAL data: Ollama preferred, external only if explicitly allowed
    IF p_sensitivity = 'confidential' THEN
        IF p_allow_external THEN
            RETURN 'ollama_or_groq_sanitized';
        ELSE
            RETURN 'ollama';
        END IF;
    END IF;

    -- INTERNAL data: Ollama preferred, Groq acceptable
    IF p_sensitivity = 'internal' THEN
        RETURN 'ollama_or_groq';
    END IF;

    -- PUBLIC data: Groq for speed
    IF p_sensitivity = 'public' THEN
        RETURN 'groq';
    END IF;

    -- Default: safe choice
    RETURN 'ollama';
END;
$$;

COMMENT ON FUNCTION get_safe_ai_provider IS
'Returns safe AI provider based on data sensitivity level';


-- Safe chat function with automatic routing
CREATE OR REPLACE FUNCTION ai_chat_safe(
    p_prompt text,
    p_table_name text DEFAULT NULL,
    p_column_name text DEFAULT NULL,
    p_force_local boolean DEFAULT false
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_sensitivity data_sensitivity;
    v_provider text;
    v_sanitization_result jsonb;
    v_sanitized_prompt text;
    v_response text;
    v_session_id text;
BEGIN
    -- Classify data sensitivity
    IF p_table_name IS NOT NULL AND p_column_name IS NOT NULL THEN
        v_sensitivity := classify_text_sensitivity(p_table_name, p_column_name, p_prompt);
    ELSIF contains_pii(p_prompt) THEN
        v_sensitivity := 'restricted';
    ELSE
        v_sensitivity := 'internal';
    END IF;

    -- Force local if requested or if restricted
    IF p_force_local OR v_sensitivity IN ('restricted', 'confidential') THEN
        RETURN ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024);
    END IF;

    -- Get safe provider
    v_provider := get_safe_ai_provider(v_sensitivity, false);

    -- Route based on provider
    IF v_provider = 'ollama' THEN
        RETURN ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024);

    ELSIF v_provider = 'groq' THEN
        -- Still check for PII one more time
        IF contains_pii(p_prompt) THEN
            RAISE NOTICE 'PII detected in prompt marked as public. Routing to Ollama.';
            RETURN ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024);
        END IF;
        RETURN groq_chat_completion(p_prompt, 'llama-3.3-70b-versatile', 0.7, 1024);

    ELSIF v_provider = 'ollama_or_groq' THEN
        -- Prefer Ollama, fallback to Groq only if Ollama unavailable
        BEGIN
            RETURN ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Ollama unavailable, falling back to Groq';
            RETURN groq_chat_completion(p_prompt, 'llama-3.3-70b-versatile', 0.7, 1024);
        END;

    ELSIF v_provider = 'ollama_or_groq_sanitized' THEN
        -- Sanitize first, then send to Groq
        v_sanitization_result := sanitize_pii(p_prompt);
        v_sanitized_prompt := v_sanitization_result->>'sanitized_text';
        v_session_id := v_sanitization_result->>'session_id';

        -- Call external API with sanitized data
        v_response := groq_chat_completion(v_sanitized_prompt, 'llama-3.3-70b-versatile', 0.7, 1024);

        -- Rehydrate response
        RETURN rehydrate_pii(v_response, v_session_id);

    ELSE
        -- Fallback to Ollama (safe choice)
        RETURN ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024);
    END IF;
END;
$$;

COMMENT ON FUNCTION ai_chat_safe IS
'Privacy-first AI chat that automatically routes based on data sensitivity and PII detection';


-- ============================================
-- AUDIT LOGGING
-- ============================================

CREATE TABLE ai_usage_audit (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations(id),
    user_id uuid,
    provider text NOT NULL,
    model text,
    data_classification data_sensitivity,
    contained_pii boolean,
    was_sanitized boolean DEFAULT false,
    prompt_hash text,  -- SHA256 hash, never store actual prompt
    response_hash text,
    tokens_used int,
    processing_time_ms int,
    error text,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_ai_usage_audit_org ON ai_usage_audit(organization_id, created_at);
CREATE INDEX idx_ai_usage_audit_user ON ai_usage_audit(user_id, created_at);

COMMENT ON TABLE ai_usage_audit IS
'Audit log for all AI API calls (stores hashes only, not actual content)';

-- Enable RLS
ALTER TABLE ai_usage_audit ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their org AI usage"
    ON ai_usage_audit FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));


-- ============================================
-- CLEANUP JOB
-- ============================================

-- Auto-cleanup expired PII mappings
CREATE OR REPLACE FUNCTION cleanup_expired_pii_mappings()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM pii_sanitization_map
    WHERE expires_at < now();
END;
$$;

-- Schedule cleanup (call from cron or manually)
-- SELECT cleanup_expired_pii_mappings();


-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT ON data_classification_rules TO authenticated;
GRANT ALL ON data_classification_rules TO service_role;

GRANT EXECUTE ON FUNCTION contains_pii(text) TO authenticated;
GRANT EXECUTE ON FUNCTION classify_text_sensitivity(text, text, text) TO authenticated;
GRANT EXECUTE ON FUNCTION sanitize_pii(text, text) TO authenticated;
GRANT EXECUTE ON FUNCTION rehydrate_pii(text, text) TO authenticated;
GRANT EXECUTE ON FUNCTION get_safe_ai_provider(data_sensitivity, boolean) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_chat_safe(text, text, text, boolean) TO authenticated;

GRANT SELECT ON ai_usage_audit TO authenticated;


-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Check if text contains PII
SELECT contains_pii('Contact John Doe at john@acme.com or call +1-555-0100');
-- Returns: true

-- 2. Classify data sensitivity
SELECT classify_text_sensitivity('contacts', 'email', 'john@acme.com');
-- Returns: 'restricted'

-- 3. Sanitize PII
SELECT sanitize_pii('Contact John at john@acme.com or +1-555-0100');
-- Returns: {
--   "sanitized_text": "Contact John at [EMAIL_1] or [PHONE_1]",
--   "session_id": "uuid-here",
--   "had_pii": true
-- }

-- 4. Safe AI chat (automatic routing)
SELECT ai_chat_safe('How do I process an invoice for john@acme.com?');
-- Automatically detects PII and routes to Ollama

SELECT ai_chat_safe('What are the features of our ERP system?');
-- No PII, can use Groq for speed

-- 5. Force local processing
SELECT ai_chat_safe('Analyze this contract...', NULL, NULL, true);
-- Always uses Ollama regardless of content

-- 6. View audit log
SELECT * FROM ai_usage_audit
WHERE organization_id = 'your-org-id'
ORDER BY created_at DESC
LIMIT 100;
*/
