-- Migration: Ollama Embedding Integration
-- Description: Integrates Ollama API for generating embeddings
-- Created: 2025-01-01
-- Dependencies: Requires Ollama container running with nomic-embed-text model

-- Enable HTTP extension for calling Ollama API
CREATE EXTENSION IF NOT EXISTS http;

-- ============================================
-- OLLAMA EMBEDDING FUNCTION
-- ============================================

-- Function to generate 768-dim embeddings using Ollama
CREATE OR REPLACE FUNCTION generate_embedding_ollama_768(
    p_text text,
    p_model text DEFAULT 'nomic-embed-text'
)
RETURNS vector(768)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_response http_response;
    v_response_json jsonb;
    v_embedding float[];
    v_ollama_url text := 'http://ollama:11434/api/embeddings';
BEGIN
    -- Validate input
    IF p_text IS NULL OR trim(p_text) = '' THEN
        RAISE EXCEPTION 'Text cannot be empty';
    END IF;

    -- Call Ollama API
    v_response := http((
        'POST',
        v_ollama_url,
        ARRAY[http_header('Content-Type', 'application/json')],
        'application/json',
        jsonb_build_object(
            'model', p_model,
            'prompt', p_text
        )::text
    )::http_request);

    -- Check response status
    IF v_response.status != 200 THEN
        RAISE EXCEPTION 'Ollama API error: % - %', v_response.status, v_response.content;
    END IF;

    -- Parse response
    v_response_json := v_response.content::jsonb;

    -- Extract embedding array
    SELECT ARRAY(
        SELECT (jsonb_array_elements_text(v_response_json->'embedding'))::float
    ) INTO v_embedding;

    -- Verify dimension
    IF array_length(v_embedding, 1) != 768 THEN
        RAISE EXCEPTION 'Invalid embedding dimension: expected 768, got %', array_length(v_embedding, 1);
    END IF;

    RETURN v_embedding::vector(768);
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Failed to generate embedding via Ollama: %', SQLERRM;
        RETURN NULL;
END;
$$;

COMMENT ON FUNCTION generate_embedding_ollama_768 IS
'Generate 768-dimensional embeddings using Ollama nomic-embed-text model';


-- Function to generate 384-dim embeddings using Ollama (lightweight)
CREATE OR REPLACE FUNCTION generate_embedding_ollama_384(
    p_text text,
    p_model text DEFAULT 'all-minilm'
)
RETURNS vector(384)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_response http_response;
    v_response_json jsonb;
    v_embedding float[];
    v_ollama_url text := 'http://ollama:11434/api/embeddings';
BEGIN
    -- Validate input
    IF p_text IS NULL OR trim(p_text) = '' THEN
        RAISE EXCEPTION 'Text cannot be empty';
    END IF;

    -- Call Ollama API
    v_response := http((
        'POST',
        v_ollama_url,
        ARRAY[http_header('Content-Type', 'application/json')],
        'application/json',
        jsonb_build_object(
            'model', p_model,
            'prompt', p_text
        )::text
    )::http_request);

    -- Check response status
    IF v_response.status != 200 THEN
        RAISE EXCEPTION 'Ollama API error: % - %', v_response.status, v_response.content;
    END IF;

    -- Parse response
    v_response_json := v_response.content::jsonb;

    -- Extract embedding array
    SELECT ARRAY(
        SELECT (jsonb_array_elements_text(v_response_json->'embedding'))::float
    ) INTO v_embedding;

    -- Verify dimension
    IF array_length(v_embedding, 1) != 384 THEN
        RAISE EXCEPTION 'Invalid embedding dimension: expected 384, got %', array_length(v_embedding, 1);
    END IF;

    RETURN v_embedding::vector(384);
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Failed to generate embedding via Ollama: %', SQLERRM;
        RETURN NULL;
END;
$$;

COMMENT ON FUNCTION generate_embedding_ollama_384 IS
'Generate 384-dimensional embeddings using Ollama all-minilm model (lightweight)';


-- ============================================
-- UPDATE EXISTING PLACEHOLDER FUNCTIONS
-- ============================================

-- Replace the placeholder function from previous migrations
CREATE OR REPLACE FUNCTION generate_search_embedding(text_content text)
RETURNS vector(768)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    -- Use Ollama for embedding generation
    RETURN generate_embedding_ollama_768(text_content, 'nomic-embed-text');
END;
$$;

COMMENT ON FUNCTION generate_search_embedding IS
'Generate search embeddings using Ollama (768 dimensions)';


-- ============================================
-- OLLAMA CHAT COMPLETION FUNCTION (OPTIONAL)
-- ============================================

-- Function to use Mistral 7B for chat completions
CREATE OR REPLACE FUNCTION ollama_chat_completion(
    p_prompt text,
    p_model text DEFAULT 'mistral',
    p_temperature float DEFAULT 0.7,
    p_max_tokens int DEFAULT 500
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_response http_response;
    v_response_json jsonb;
    v_ollama_url text := 'http://ollama:11434/api/generate';
BEGIN
    -- Validate input
    IF p_prompt IS NULL OR trim(p_prompt) = '' THEN
        RAISE EXCEPTION 'Prompt cannot be empty';
    END IF;

    -- Call Ollama API
    v_response := http((
        'POST',
        v_ollama_url,
        ARRAY[http_header('Content-Type', 'application/json')],
        'application/json',
        jsonb_build_object(
            'model', p_model,
            'prompt', p_prompt,
            'stream', false,
            'options', jsonb_build_object(
                'temperature', p_temperature,
                'num_predict', p_max_tokens
            )
        )::text
    )::http_request);

    -- Check response status
    IF v_response.status != 200 THEN
        RAISE EXCEPTION 'Ollama API error: % - %', v_response.status, v_response.content;
    END IF;

    -- Parse response
    v_response_json := v_response.content::jsonb;

    -- Return the generated response
    RETURN v_response_json->>'response';
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Failed to generate completion via Ollama: %', SQLERRM;
        RETURN NULL;
END;
$$;

COMMENT ON FUNCTION ollama_chat_completion IS
'Generate chat completions using Ollama Mistral 7B model';


-- ============================================
-- UTILITY FUNCTIONS
-- ============================================

-- Function to check Ollama health
CREATE OR REPLACE FUNCTION ollama_health_check()
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_response http_response;
    v_ollama_url text := 'http://ollama:11434/api/tags';
BEGIN
    v_response := http((
        'GET',
        v_ollama_url,
        ARRAY[]::http_header[],
        NULL,
        NULL
    )::http_request);

    IF v_response.status = 200 THEN
        RETURN jsonb_build_object(
            'status', 'healthy',
            'models', v_response.content::jsonb
        );
    ELSE
        RETURN jsonb_build_object(
            'status', 'unhealthy',
            'error', v_response.content
        );
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'message', SQLERRM
        );
END;
$$;

COMMENT ON FUNCTION ollama_health_check IS
'Check Ollama service health and list available models';


-- ============================================
-- GRANT PERMISSIONS
-- ============================================

-- Grant execute permissions to authenticated users
GRANT EXECUTE ON FUNCTION generate_embedding_ollama_768(text, text) TO authenticated;
GRANT EXECUTE ON FUNCTION generate_embedding_ollama_384(text, text) TO authenticated;
GRANT EXECUTE ON FUNCTION generate_search_embedding(text) TO authenticated;
GRANT EXECUTE ON FUNCTION ollama_chat_completion(text, text, float, int) TO authenticated;
GRANT EXECUTE ON FUNCTION ollama_health_check() TO authenticated;

-- Grant execute to service role
GRANT EXECUTE ON FUNCTION generate_embedding_ollama_768(text, text) TO service_role;
GRANT EXECUTE ON FUNCTION generate_embedding_ollama_384(text, text) TO service_role;
GRANT EXECUTE ON FUNCTION generate_search_embedding(text) TO service_role;
GRANT EXECUTE ON FUNCTION ollama_chat_completion(text, text, float, int) TO service_role;
GRANT EXECUTE ON FUNCTION ollama_health_check() TO service_role;


-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- Generate embedding
SELECT generate_embedding_ollama_768('This is a test document about sales orders');

-- Update existing contact with embedding
UPDATE contacts
SET search_embedding_768 = generate_embedding_ollama_768(
    COALESCE(name, '') || ' ' || COALESCE(email, '') || ' ' || COALESCE(phone, '')
)
WHERE id = 'some-contact-id';

-- Use Mistral for AI completion
SELECT ollama_chat_completion(
    'Summarize the key points of this invoice: ' || invoice_description,
    'mistral',
    0.7,
    200
);

-- Check Ollama health
SELECT ollama_health_check();
*/
