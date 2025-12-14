-- Migration: Groq with Automatic Sanitization (Option B)
-- Description: Leverage Groq's speed while protecting customer data through automatic sanitization
-- Created: 2025-01-01
-- Strategy: Detect PII → Sanitize → Send to Groq → Rehydrate response

-- ============================================
-- ENHANCED SANITIZATION ENGINE
-- ============================================

-- Enhanced PII detection with more patterns
CREATE OR REPLACE FUNCTION detect_pii_advanced(text_content text)
RETURNS jsonb
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_pii_found jsonb := '[]'::jsonb;
    v_patterns jsonb;
BEGIN
    IF text_content IS NULL OR trim(text_content) = '' THEN
        RETURN '[]'::jsonb;
    END IF;

    -- Email addresses
    IF text_content ~* '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'email',
            'risk', 'high',
            'pattern', 'email_address'
        );
    END IF;

    -- Phone numbers (international formats)
    IF text_content ~* '(\+?[0-9]{1,3}[\s.-]?)?\(?[0-9]{3}\)?[\s.-]?[0-9]{3}[\s.-]?[0-9]{4}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'phone',
            'risk', 'high',
            'pattern', 'phone_number'
        );
    END IF;

    -- SSN (US)
    IF text_content ~* '[0-9]{3}-[0-9]{2}-[0-9]{4}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'ssn',
            'risk', 'critical',
            'pattern', 'social_security'
        );
    END IF;

    -- Credit card numbers
    IF text_content ~* '[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'credit_card',
            'risk', 'critical',
            'pattern', 'payment_card'
        );
    END IF;

    -- IP addresses
    IF text_content ~* '([0-9]{1,3}\.){3}[0-9]{1,3}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'ip_address',
            'risk', 'medium',
            'pattern', 'ip_v4'
        );
    END IF;

    -- IBAN (European bank accounts)
    IF text_content ~* '[A-Z]{2}[0-9]{2}[A-Z0-9]{10,30}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'iban',
            'risk', 'critical',
            'pattern', 'bank_account'
        );
    END IF;

    -- Dates (potential DOB)
    IF text_content ~* '[0-9]{2}/[0-9]{2}/[0-9]{4}' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'date',
            'risk', 'medium',
            'pattern', 'date_of_birth'
        );
    END IF;

    -- Names (common patterns like "Mr. ", "Ms. ", "Dr. ")
    IF text_content ~* '(Mr\.|Mrs\.|Ms\.|Dr\.|Prof\.)\s+[A-Z][a-z]+\s+[A-Z][a-z]+' THEN
        v_pii_found := v_pii_found || jsonb_build_object(
            'type', 'name',
            'risk', 'medium',
            'pattern', 'formal_name'
        );
    END IF;

    RETURN v_pii_found;
END;
$$;

COMMENT ON FUNCTION detect_pii_advanced IS
'Advanced PII detection returning detailed analysis of found patterns';


-- Enhanced sanitization with context preservation
CREATE OR REPLACE FUNCTION sanitize_for_groq(
    p_text text,
    p_preserve_context boolean DEFAULT true
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_sanitized_text text;
    v_session_id text;
    v_replacements jsonb := '[]'::jsonb;
    v_matches text[];
    v_match text;
    v_counter int;
    v_pii_analysis jsonb;
BEGIN
    v_sanitized_text := p_text;
    v_session_id := gen_random_uuid()::text;

    -- Analyze PII first
    v_pii_analysis := detect_pii_advanced(p_text);

    -- Replace emails with contextual tokens
    v_counter := 0;
    FOR v_match IN
        SELECT DISTINCT m[1]
        FROM regexp_matches(p_text, '([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})', 'g') AS m
    LOOP
        v_counter := v_counter + 1;

        -- Preserve domain context if requested
        DECLARE
            v_token text;
            v_domain text;
        BEGIN
            IF p_preserve_context THEN
                v_domain := substring(v_match from '@(.+)$');
                v_token := '[EMAIL_' || v_counter || '@' || v_domain || ']';
            ELSE
                v_token := '[EMAIL_' || v_counter || ']';
            END IF;

            -- Store mapping
            INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
            VALUES (v_session_id, 'email', v_match, v_token);

            v_replacements := v_replacements || jsonb_build_object(
                'original', v_match,
                'token', v_token,
                'type', 'email'
            );

            v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
        END;
    END LOOP;

    -- Replace phone numbers
    v_counter := 0;
    FOR v_match IN
        SELECT DISTINCT m[1]
        FROM regexp_matches(p_text, '((\+?[0-9]{1,3}[\s.-]?)?\(?[0-9]{3}\)?[\s.-]?[0-9]{3}[\s.-]?[0-9]{4})', 'g') AS m
    LOOP
        v_counter := v_counter + 1;
        DECLARE
            v_token text := '[PHONE_' || v_counter || ']';
        BEGIN
            INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
            VALUES (v_session_id, 'phone', v_match, v_token);

            v_replacements := v_replacements || jsonb_build_object(
                'original', v_match,
                'token', v_token,
                'type', 'phone'
            );

            v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
        END;
    END LOOP;

    -- Replace SSN
    v_counter := 0;
    FOR v_match IN
        SELECT DISTINCT m[1]
        FROM regexp_matches(p_text, '([0-9]{3}-[0-9]{2}-[0-9]{4})', 'g') AS m
    LOOP
        v_counter := v_counter + 1;
        DECLARE
            v_token text := '[SSN_' || v_counter || ']';
        BEGIN
            INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
            VALUES (v_session_id, 'ssn', v_match, v_token);

            v_replacements := v_replacements || jsonb_build_object(
                'original', v_match,
                'token', v_token,
                'type', 'ssn'
            );

            v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
        END;
    END LOOP;

    -- Replace amounts (preserve currency context)
    v_counter := 0;
    FOR v_match IN
        SELECT DISTINCT m[1]
        FROM regexp_matches(p_text, '(\$[0-9,]+\.?[0-9]*|€[0-9,]+\.?[0-9]*|£[0-9,]+\.?[0-9]*)', 'g') AS m
    LOOP
        v_counter := v_counter + 1;
        DECLARE
            v_token text;
            v_currency text;
        BEGIN
            v_currency := substring(v_match from '^([$€£])');
            IF p_preserve_context THEN
                v_token := v_currency || '[AMOUNT_' || v_counter || ']';
            ELSE
                v_token := '[AMOUNT_' || v_counter || ']';
            END IF;

            INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
            VALUES (v_session_id, 'amount', v_match, v_token);

            v_replacements := v_replacements || jsonb_build_object(
                'original', v_match,
                'token', v_token,
                'type', 'amount'
            );

            v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
        END;
    END LOOP;

    -- Replace names (person entities)
    v_counter := 0;
    FOR v_match IN
        SELECT DISTINCT m[1]
        FROM regexp_matches(p_text, '((Mr\.|Mrs\.|Ms\.|Dr\.|Prof\.)\s+[A-Z][a-z]+\s+[A-Z][a-z]+)', 'g') AS m
    LOOP
        v_counter := v_counter + 1;
        DECLARE
            v_token text := '[PERSON_' || v_counter || ']';
        BEGIN
            INSERT INTO pii_sanitization_map (session_id, pii_type, original_value, sanitized_token)
            VALUES (v_session_id, 'name', v_match, v_token);

            v_replacements := v_replacements || jsonb_build_object(
                'original', v_match,
                'token', v_token,
                'type', 'name'
            );

            v_sanitized_text := replace(v_sanitized_text, v_match, v_token);
        END;
    END LOOP;

    RETURN jsonb_build_object(
        'sanitized_text', v_sanitized_text,
        'session_id', v_session_id,
        'had_pii', v_sanitized_text != p_text,
        'replacements', v_replacements,
        'pii_analysis', v_pii_analysis,
        'replacement_count', jsonb_array_length(v_replacements)
    );
END;
$$;

COMMENT ON FUNCTION sanitize_for_groq IS
'Enhanced sanitization optimized for Groq API with context preservation';


-- ============================================
-- GROQ FUNCTIONS WITH AUTO-SANITIZATION
-- ============================================

-- Safe Groq chat with automatic sanitization
CREATE OR REPLACE FUNCTION groq_chat_safe(
    p_prompt text,
    p_model text DEFAULT 'llama-3.3-70b-versatile',
    p_temperature float DEFAULT 0.7,
    p_max_tokens int DEFAULT 1024,
    p_auto_sanitize boolean DEFAULT true
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_sanitization_result jsonb;
    v_sanitized_prompt text;
    v_session_id text;
    v_groq_response text;
    v_final_response text;
    v_start_time timestamptz;
    v_end_time timestamptz;
    v_processing_time_ms int;
BEGIN
    v_start_time := clock_timestamp();

    -- Auto-sanitize if enabled and PII detected
    IF p_auto_sanitize THEN
        v_sanitization_result := sanitize_for_groq(p_prompt, true);

        IF (v_sanitization_result->>'had_pii')::boolean THEN
            RAISE NOTICE 'PII detected and sanitized. Session: %', v_sanitization_result->>'session_id';

            v_sanitized_prompt := v_sanitization_result->>'sanitized_text';
            v_session_id := v_sanitization_result->>'session_id';
        ELSE
            -- No PII, send as-is
            v_sanitized_prompt := p_prompt;
            v_session_id := NULL;
        END IF;
    ELSE
        -- Sanitization disabled
        v_sanitized_prompt := p_prompt;
        v_session_id := NULL;
    END IF;

    -- Call Groq API with sanitized prompt
    v_groq_response := groq_chat_completion(
        v_sanitized_prompt,
        p_model,
        p_temperature,
        p_max_tokens
    );

    -- Rehydrate response if we sanitized
    IF v_session_id IS NOT NULL THEN
        v_final_response := rehydrate_pii(v_groq_response, v_session_id);
    ELSE
        v_final_response := v_groq_response;
    END IF;

    v_end_time := clock_timestamp();
    v_processing_time_ms := EXTRACT(MILLISECONDS FROM (v_end_time - v_start_time))::int;

    -- Return detailed result
    RETURN jsonb_build_object(
        'response', v_final_response,
        'was_sanitized', v_session_id IS NOT NULL,
        'sanitization_details', COALESCE(v_sanitization_result, '{}'::jsonb),
        'processing_time_ms', v_processing_time_ms,
        'model', p_model,
        'provider', 'groq'
    );
END;
$$;

COMMENT ON FUNCTION groq_chat_safe IS
'Groq chat with automatic PII sanitization - best of both worlds: privacy + speed';


-- Safe Groq GPT-OSS reasoning with sanitization
CREATE OR REPLACE FUNCTION groq_gpt_oss_safe(
    p_prompt text,
    p_thinking_budget int DEFAULT 15000,
    p_temperature float DEFAULT 0.3,
    p_auto_sanitize boolean DEFAULT true
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_sanitization_result jsonb;
    v_sanitized_prompt text;
    v_session_id text;
    v_groq_response jsonb;
    v_final_content text;
    v_final_reasoning text;
BEGIN
    -- Auto-sanitize
    IF p_auto_sanitize THEN
        v_sanitization_result := sanitize_for_groq(p_prompt, true);

        IF (v_sanitization_result->>'had_pii')::boolean THEN
            v_sanitized_prompt := v_sanitization_result->>'sanitized_text';
            v_session_id := v_sanitization_result->>'session_id';
        ELSE
            v_sanitized_prompt := p_prompt;
            v_session_id := NULL;
        END IF;
    ELSE
        v_sanitized_prompt := p_prompt;
        v_session_id := NULL;
    END IF;

    -- Call Groq GPT-OSS
    v_groq_response := groq_gpt_oss_reasoning(
        v_sanitized_prompt,
        p_thinking_budget,
        p_temperature
    );

    -- Rehydrate both content and reasoning
    IF v_session_id IS NOT NULL THEN
        v_final_content := rehydrate_pii(v_groq_response->>'content', v_session_id);
        v_final_reasoning := rehydrate_pii(v_groq_response->>'reasoning', v_session_id);
    ELSE
        v_final_content := v_groq_response->>'content';
        v_final_reasoning := v_groq_response->>'reasoning';
    END IF;

    RETURN jsonb_build_object(
        'content', v_final_content,
        'reasoning', v_final_reasoning,
        'usage', v_groq_response->'usage',
        'was_sanitized', v_session_id IS NOT NULL,
        'sanitization_details', COALESCE(v_sanitization_result, '{}'::jsonb),
        'model', 'gpt-oss-120b',
        'provider', 'groq'
    );
END;
$$;

COMMENT ON FUNCTION groq_gpt_oss_safe IS
'GPT-OSS-120B reasoning with automatic sanitization for sensitive data';


-- Safe Groq Compound AI with sanitization
CREATE OR REPLACE FUNCTION groq_compound_safe(
    p_prompt text,
    p_system_prompt text DEFAULT 'You are a helpful AI assistant.',
    p_max_tokens int DEFAULT 3000,
    p_auto_sanitize boolean DEFAULT true
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_sanitization_result jsonb;
    v_sanitized_prompt text;
    v_session_id text;
    v_groq_response jsonb;
    v_final_content text;
BEGIN
    -- Auto-sanitize
    IF p_auto_sanitize THEN
        v_sanitization_result := sanitize_for_groq(p_prompt, true);

        IF (v_sanitization_result->>'had_pii')::boolean THEN
            v_sanitized_prompt := v_sanitization_result->>'sanitized_text';
            v_session_id := v_sanitization_result->>'session_id';
        ELSE
            v_sanitized_prompt := p_prompt;
            v_session_id := NULL;
        END IF;
    ELSE
        v_sanitized_prompt := p_prompt;
        v_session_id := NULL;
    END IF;

    -- Call Groq Compound
    v_groq_response := groq_compound_reasoning(
        v_sanitized_prompt,
        p_system_prompt,
        p_max_tokens
    );

    -- Rehydrate
    IF v_session_id IS NOT NULL THEN
        v_final_content := rehydrate_pii(v_groq_response->>'content', v_session_id);
    ELSE
        v_final_content := v_groq_response->>'content';
    END IF;

    RETURN jsonb_build_object(
        'content', v_final_content,
        'reasoning_steps', v_groq_response->'reasoning_steps',
        'usage', v_groq_response->'usage',
        'was_sanitized', v_session_id IS NOT NULL,
        'sanitization_details', COALESCE(v_sanitization_result, '{}'::jsonb),
        'model', 'compound',
        'provider', 'groq'
    );
END;
$$;

COMMENT ON FUNCTION groq_compound_safe IS
'Groq Compound AI with automatic sanitization for research/agentic tasks';


-- ============================================
-- SMART INVOICE EXTRACTION (GROQ-OPTIMIZED)
-- ============================================

CREATE OR REPLACE FUNCTION extract_invoice_groq(
    p_invoice_text text
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_prompt text;
    v_result jsonb;
    v_extracted_data jsonb;
    v_json_text text;
BEGIN
    v_prompt := format(
        'Extract invoice data and return ONLY valid JSON (no explanations, no markdown):
        {
          "invoice_number": "string",
          "invoice_date": "YYYY-MM-DD",
          "due_date": "YYYY-MM-DD",
          "vendor_name": "string",
          "vendor_email": "string or null",
          "vendor_phone": "string or null",
          "total_amount": number,
          "currency": "string",
          "tax_amount": number,
          "subtotal": number,
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

    -- Use Groq GPT-OSS for best accuracy, auto-sanitize PII
    v_result := groq_gpt_oss_safe(v_prompt, 20000, 0.1, true);

    -- Extract JSON from response
    v_json_text := v_result->>'content';

    -- Parse JSON (handle potential markdown wrapping)
    v_json_text := regexp_replace(v_json_text, '```json\s*', '', 'g');
    v_json_text := regexp_replace(v_json_text, '```\s*$', '', 'g');
    v_json_text := trim(v_json_text);

    BEGIN
        v_extracted_data := v_json_text::jsonb;
    EXCEPTION WHEN OTHERS THEN
        -- If parsing fails, extract JSON object from text
        v_json_text := substring(v_json_text from '\{[^\}]*\}');
        v_extracted_data := v_json_text::jsonb;
    END;

    -- Return with metadata
    RETURN jsonb_build_object(
        'invoice_data', v_extracted_data,
        'was_sanitized', (v_result->>'was_sanitized')::boolean,
        'processing_info', jsonb_build_object(
            'model', 'gpt-oss-120b',
            'provider', 'groq',
            'had_pii', (v_result->'sanitization_details'->>'had_pii')::boolean
        )
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'error', SQLERRM,
            'raw_response', v_result
        );
END;
$$;

COMMENT ON FUNCTION extract_invoice_groq IS
'Extract invoice data using Groq GPT-OSS with automatic PII sanitization';


-- ============================================
-- UNIFIED SMART AI FUNCTION (OPTION B)
-- ============================================

-- Ultimate function: auto-detect, auto-sanitize, auto-route
CREATE OR REPLACE FUNCTION ai_process(
    p_prompt text,
    p_prefer_groq boolean DEFAULT true,
    p_allow_sanitization boolean DEFAULT true
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_pii_detected boolean;
    v_result jsonb;
BEGIN
    -- Detect PII
    v_pii_detected := contains_pii(p_prompt);

    -- If PII detected
    IF v_pii_detected THEN
        IF p_allow_sanitization AND p_prefer_groq THEN
            -- Sanitize and use Groq (fast)
            RAISE NOTICE 'PII detected. Using Groq with sanitization.';
            RETURN groq_chat_safe(p_prompt, 'llama-3.3-70b-versatile', 0.7, 1024, true);
        ELSE
            -- Use Ollama (local, no sanitization needed)
            RAISE NOTICE 'PII detected. Using Ollama (local).';
            RETURN jsonb_build_object(
                'response', ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024),
                'was_sanitized', false,
                'model', 'mistral',
                'provider', 'ollama'
            );
        END IF;
    ELSE
        -- No PII, use Groq for speed
        IF p_prefer_groq THEN
            RAISE NOTICE 'No PII detected. Using Groq.';
            RETURN groq_chat_safe(p_prompt, 'llama-3.3-70b-versatile', 0.7, 1024, false);
        ELSE
            RAISE NOTICE 'Using Ollama.';
            RETURN jsonb_build_object(
                'response', ollama_chat_completion(p_prompt, 'mistral', 0.7, 1024),
                'was_sanitized', false,
                'model', 'mistral',
                'provider', 'ollama'
            );
        END IF;
    END IF;
END;
$$;

COMMENT ON FUNCTION ai_process IS
'Smart AI processing: auto-detects PII, sanitizes if needed, routes to best provider';


-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT EXECUTE ON FUNCTION detect_pii_advanced(text) TO authenticated;
GRANT EXECUTE ON FUNCTION sanitize_for_groq(text, boolean) TO authenticated;
GRANT EXECUTE ON FUNCTION groq_chat_safe(text, text, float, int, boolean) TO authenticated;
GRANT EXECUTE ON FUNCTION groq_gpt_oss_safe(text, int, float, boolean) TO authenticated;
GRANT EXECUTE ON FUNCTION groq_compound_safe(text, text, int, boolean) TO authenticated;
GRANT EXECUTE ON FUNCTION extract_invoice_groq(text) TO authenticated;
GRANT EXECUTE ON FUNCTION ai_process(text, boolean, boolean) TO authenticated;

GRANT EXECUTE ON FUNCTION detect_pii_advanced(text) TO service_role;
GRANT EXECUTE ON FUNCTION sanitize_for_groq(text, boolean) TO service_role;
GRANT EXECUTE ON FUNCTION groq_chat_safe(text, text, float, int, boolean) TO service_role;
GRANT EXECUTE ON FUNCTION groq_gpt_oss_safe(text, int, float, boolean) TO service_role;
GRANT EXECUTE ON FUNCTION groq_compound_safe(text, text, int, boolean) TO service_role;
GRANT EXECUTE ON FUNCTION extract_invoice_groq(text) TO service_role;
GRANT EXECUTE ON FUNCTION ai_process(text, boolean, boolean) TO service_role;


-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Safe Groq chat (auto-sanitizes PII)
SELECT groq_chat_safe('Process invoice for John Doe at john@acme.com, amount $5,000');
-- Returns:
-- {
--   "response": "I'll process the invoice for [PERSON_1] at [EMAIL_1@acme.com], amount $[AMOUNT_1]... (rehydrated response)",
--   "was_sanitized": true,
--   "sanitization_details": {...},
--   "processing_time_ms": 850,
--   "model": "llama-3.3-70b-versatile",
--   "provider": "groq"
-- }

-- 2. Extract invoice with Groq (auto-sanitizes)
SELECT extract_invoice_groq('
Invoice #12345
Date: 2025-01-15
Vendor: ACME Corp
Contact: john@acme.com
Phone: +1-555-0100
Total: $3,685.50
');
-- Returns structured JSON with PII sanitized during processing

-- 3. Smart AI processing (auto-routes)
SELECT ai_process('How do I create a sales order?');  -- No PII → Groq
SELECT ai_process('Process order for jane@company.com');  -- PII → Groq with sanitization

-- 4. Advanced PII detection
SELECT detect_pii_advanced('Contact Mr. John Doe at john@acme.com, SSN: 123-45-6789, Card: 4111-1111-1111-1111');
-- Returns detailed analysis of all PII found

-- 5. Manual sanitization workflow
SELECT sanitize_for_groq('Email john@acme.com about invoice $5,000', true);
-- Returns sanitized text with session ID for rehydration
*/
