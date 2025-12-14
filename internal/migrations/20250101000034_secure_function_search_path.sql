-- Migration: Secure Function Search Paths
-- Description: Set immutable search_path on ALL SECURITY DEFINER functions
-- Created: 2025-01-01
-- Issue: Functions have mutable search_path, which can lead to security vulnerabilities

-- =====================================================
-- SECURITY: SET SEARCH PATH ON ALL SECURITY DEFINER FUNCTIONS
-- =====================================================

-- All functions using SECURITY DEFINER must have an immutable search_path
-- to prevent search path hijacking attacks

-- Redis Functions
ALTER FUNCTION redis.register_cache_key(text, text, integer) SET search_path TO '';
ALTER FUNCTION redis.cleanup_expired_sessions() SET search_path TO '';
ALTER FUNCTION redis.record_embedding_cache_hit(text) SET search_path TO '';

-- Queue Job Handlers
ALTER FUNCTION handle_generate_embedding_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_process_invoice_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_duplicate_detection_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_send_email_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_export_data_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_stock_reorder_check_job(jsonb) SET search_path TO '';

-- PGMQ Functions
ALTER FUNCTION enqueue_job(uuid, text, text, jsonb, int, int, timestamptz, jsonb) SET search_path TO '';

-- Groq AI Functions
ALTER FUNCTION groq_chat_completion(text, text, float, int, text) SET search_path TO '';
ALTER FUNCTION groq_compound_reasoning(text, text, int, text) SET search_path TO '';
ALTER FUNCTION groq_gpt_oss_reasoning(text, int, float, text) SET search_path TO '';

-- Embedding Functions
-- get_or_create_embedding function does not exist, removing

-- Auth/User Metadata Functions
-- Only including functions that actually exist
ALTER FUNCTION user_belongs_to_organization(uuid) SET search_path TO '';
-- Other auth functions don't exist, removing them

-- Data Privacy Functions
-- encrypt_field function does not exist, removing

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION redis.register_cache_key(text, text, integer) IS
    'Registers a cache key with metadata - SECURITY DEFINER with immutable search_path';
COMMENT ON FUNCTION redis.cleanup_expired_sessions() IS
    'Removes expired sessions - SECURITY DEFINER with immutable search_path';
COMMENT ON FUNCTION redis.record_embedding_cache_hit(text) IS
    'Records a cache hit for embeddings - SECURITY DEFINER with immutable search_path';
