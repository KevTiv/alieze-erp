-- Migration: PostgreSQL Native Caching with UNLOGGED Tables
-- Description: Uses PostgreSQL UNLOGGED tables for caching and session management
-- Created: 2025-01-25
-- Updated: 2025-12-15 - Refactored to use native PostgreSQL instead of Redis

-- ============================================
-- CACHE SCHEMA SETUP
-- ============================================

CREATE SCHEMA IF NOT EXISTS cache;

-- Grant access to the schema
GRANT USAGE ON SCHEMA cache TO postgres, anon, authenticated, service_role;

-- ============================================
-- KEY-VALUE CACHE STORE
-- ============================================

-- Main cache table using UNLOGGED for performance
-- UNLOGGED tables are not written to WAL, making them faster for temporary data
-- Data is lost on crash, but that's acceptable for cache
CREATE UNLOGGED TABLE IF NOT EXISTS cache.store (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed TIMESTAMPTZ DEFAULT NOW()
);

-- Index for expiration cleanup
CREATE INDEX IF NOT EXISTS idx_cache_store_expires_at ON cache.store(expires_at) WHERE expires_at IS NOT NULL;

-- Grant access
GRANT SELECT, INSERT, UPDATE, DELETE ON cache.store TO postgres, authenticated, service_role;

-- Function to get cached value
CREATE OR REPLACE FUNCTION cache.get(p_key TEXT)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_value JSONB;
BEGIN
    -- Get value and update last_accessed
    UPDATE cache.store
    SET last_accessed = NOW()
    WHERE key = p_key
      AND (expires_at IS NULL OR expires_at > NOW())
    RETURNING value INTO v_value;

    RETURN v_value;
END;
$$;

ALTER FUNCTION cache.get(text) SET search_path TO '';

-- Function to set cached value with optional TTL
CREATE OR REPLACE FUNCTION cache.set(
    p_key TEXT,
    p_value JSONB,
    p_ttl_seconds INTEGER DEFAULT NULL
)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    INSERT INTO cache.store (key, value, expires_at)
    VALUES (
        p_key,
        p_value,
        CASE WHEN p_ttl_seconds IS NOT NULL
             THEN NOW() + (p_ttl_seconds || ' seconds')::INTERVAL
             ELSE NULL
        END
    )
    ON CONFLICT (key) DO UPDATE
    SET
        value = EXCLUDED.value,
        expires_at = EXCLUDED.expires_at,
        last_accessed = NOW();
END;
$$;

ALTER FUNCTION cache.set(text, jsonb, integer) SET search_path TO '';

-- Function to delete cached value
CREATE OR REPLACE FUNCTION cache.delete(p_key TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    DELETE FROM cache.store WHERE key = p_key;
    RETURN FOUND;
END;
$$;

ALTER FUNCTION cache.delete(text) SET search_path TO '';

-- Function to check if key exists and is not expired
CREATE OR REPLACE FUNCTION cache.exists(p_key TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_exists BOOLEAN;
BEGIN
    SELECT EXISTS(
        SELECT 1 FROM cache.store
        WHERE key = p_key
          AND (expires_at IS NULL OR expires_at > NOW())
    ) INTO v_exists;

    RETURN v_exists;
END;
$$;

ALTER FUNCTION cache.exists(text) SET search_path TO '';

-- Function to clean up expired cache entries
CREATE OR REPLACE FUNCTION cache.cleanup_expired()
RETURNS INTEGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM cache.store
    WHERE expires_at IS NOT NULL AND expires_at < NOW();

    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;

    RETURN v_deleted_count;
END;
$$;

ALTER FUNCTION cache.cleanup_expired() SET search_path TO '';

-- ============================================
-- SESSION STORAGE
-- ============================================

-- UNLOGGED table for session storage
CREATE UNLOGGED TABLE IF NOT EXISTS cache.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    session_key TEXT NOT NULL UNIQUE,
    session_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_activity TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON cache.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON cache.sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_last_activity ON cache.sessions(last_activity);

-- Grant access
GRANT SELECT, INSERT, UPDATE, DELETE ON cache.sessions TO postgres, authenticated, service_role;

-- Function to create or update session
CREATE OR REPLACE FUNCTION cache.upsert_session(
    p_user_id UUID,
    p_session_key TEXT,
    p_session_data JSONB DEFAULT NULL,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS UUID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_session_id UUID;
BEGIN
    INSERT INTO cache.sessions (user_id, session_key, session_data, expires_at)
    VALUES (
        p_user_id,
        p_session_key,
        p_session_data,
        NOW() + (p_ttl_seconds || ' seconds')::INTERVAL
    )
    ON CONFLICT (session_key) DO UPDATE
    SET
        session_data = COALESCE(EXCLUDED.session_data, cache.sessions.session_data),
        expires_at = EXCLUDED.expires_at,
        last_activity = NOW()
    RETURNING id INTO v_session_id;

    RETURN v_session_id;
END;
$$;

ALTER FUNCTION cache.upsert_session(uuid, text, jsonb, integer) SET search_path TO '';

-- Function to get session
CREATE OR REPLACE FUNCTION cache.get_session(p_session_key TEXT)
RETURNS TABLE (
    id UUID,
    user_id UUID,
    session_data JSONB,
    created_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    last_activity TIMESTAMPTZ
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    RETURN QUERY
    UPDATE cache.sessions
    SET last_activity = NOW()
    WHERE session_key = p_session_key
      AND expires_at > NOW()
    RETURNING
        cache.sessions.id,
        cache.sessions.user_id,
        cache.sessions.session_data,
        cache.sessions.created_at,
        cache.sessions.expires_at,
        cache.sessions.last_activity;
END;
$$;

ALTER FUNCTION cache.get_session(text) SET search_path TO '';

-- Function to clean up expired sessions
CREATE OR REPLACE FUNCTION cache.cleanup_expired_sessions()
RETURNS INTEGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM cache.sessions
    WHERE expires_at < NOW();

    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;

    RETURN v_deleted_count;
END;
$$;

ALTER FUNCTION cache.cleanup_expired_sessions() SET search_path TO '';

-- ============================================
-- EMBEDDING CACHE
-- ============================================

-- UNLOGGED table for embedding cache
CREATE UNLOGGED TABLE IF NOT EXISTS cache.embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash TEXT NOT NULL UNIQUE,
    model_name TEXT NOT NULL,
    embedding vector(1536), -- Adjust dimension as needed
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    hit_count INTEGER DEFAULT 0,
    last_hit TIMESTAMPTZ
);

-- Indexes for lookups
CREATE INDEX IF NOT EXISTS idx_embedding_cache_hash ON cache.embeddings(content_hash);
CREATE INDEX IF NOT EXISTS idx_embedding_cache_model ON cache.embeddings(model_name);
CREATE INDEX IF NOT EXISTS idx_embedding_cache_expires_at ON cache.embeddings(expires_at) WHERE expires_at IS NOT NULL;

-- Grant access
GRANT SELECT, INSERT, UPDATE, DELETE ON cache.embeddings TO postgres, authenticated, service_role;

-- Function to get cached embedding
CREATE OR REPLACE FUNCTION cache.get_embedding(
    p_content_hash TEXT,
    p_model_name TEXT DEFAULT NULL
)
RETURNS vector
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_embedding vector;
BEGIN
    UPDATE cache.embeddings
    SET
        hit_count = hit_count + 1,
        last_hit = NOW()
    WHERE content_hash = p_content_hash
      AND (p_model_name IS NULL OR model_name = p_model_name)
      AND (expires_at IS NULL OR expires_at > NOW())
    RETURNING embedding INTO v_embedding;

    RETURN v_embedding;
END;
$$;

ALTER FUNCTION cache.get_embedding(text, text) SET search_path TO '';

-- Function to set cached embedding
CREATE OR REPLACE FUNCTION cache.set_embedding(
    p_content_hash TEXT,
    p_model_name TEXT,
    p_embedding vector,
    p_ttl_seconds INTEGER DEFAULT NULL
)
RETURNS UUID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_id UUID;
BEGIN
    INSERT INTO cache.embeddings (content_hash, model_name, embedding, expires_at)
    VALUES (
        p_content_hash,
        p_model_name,
        p_embedding,
        CASE WHEN p_ttl_seconds IS NOT NULL
             THEN NOW() + (p_ttl_seconds || ' seconds')::INTERVAL
             ELSE NULL
        END
    )
    ON CONFLICT (content_hash) DO UPDATE
    SET
        embedding = EXCLUDED.embedding,
        model_name = EXCLUDED.model_name,
        expires_at = EXCLUDED.expires_at,
        hit_count = 0,
        last_hit = NULL
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$$;

ALTER FUNCTION cache.set_embedding(text, text, vector, integer) SET search_path TO '';

-- ============================================
-- RATE LIMITING
-- ============================================

-- UNLOGGED table for rate limiting
CREATE UNLOGGED TABLE IF NOT EXISTS cache.rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    limit_key TEXT NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    request_count INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(limit_key, window_start)
);

-- Index for cleanup
CREATE INDEX IF NOT EXISTS idx_rate_limits_window_start ON cache.rate_limits(window_start);
CREATE INDEX IF NOT EXISTS idx_rate_limits_key ON cache.rate_limits(limit_key, window_start);

-- Grant access
GRANT SELECT, INSERT, UPDATE, DELETE ON cache.rate_limits TO postgres, authenticated, service_role;

-- Function to check and increment rate limit
CREATE OR REPLACE FUNCTION cache.check_rate_limit(
    p_limit_key TEXT,
    p_max_requests INTEGER,
    p_window_seconds INTEGER DEFAULT 60
)
RETURNS TABLE (
    allowed BOOLEAN,
    current_count INTEGER,
    reset_at TIMESTAMPTZ
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_window_start TIMESTAMPTZ;
    v_current_count INTEGER;
    v_reset_at TIMESTAMPTZ;
BEGIN
    -- Calculate the current window start
    v_window_start := date_trunc('minute', NOW()) -
                      ((EXTRACT(EPOCH FROM NOW())::INTEGER % p_window_seconds) || ' seconds')::INTERVAL;
    v_reset_at := v_window_start + (p_window_seconds || ' seconds')::INTERVAL;

    -- Insert or update the counter
    INSERT INTO cache.rate_limits (limit_key, window_start, request_count)
    VALUES (p_limit_key, v_window_start, 1)
    ON CONFLICT (limit_key, window_start) DO UPDATE
    SET request_count = cache.rate_limits.request_count + 1
    RETURNING cache.rate_limits.request_count INTO v_current_count;

    -- Return whether the request is allowed
    RETURN QUERY SELECT
        v_current_count <= p_max_requests,
        v_current_count,
        v_reset_at;
END;
$$;

ALTER FUNCTION cache.check_rate_limit(text, integer, integer) SET search_path TO '';

-- Function to clean up old rate limit entries
CREATE OR REPLACE FUNCTION cache.cleanup_rate_limits(p_window_seconds INTEGER DEFAULT 3600)
RETURNS INTEGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM cache.rate_limits
    WHERE window_start < NOW() - (p_window_seconds || ' seconds')::INTERVAL;

    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;

    RETURN v_deleted_count;
END;
$$;

ALTER FUNCTION cache.cleanup_rate_limits(integer) SET search_path TO '';

-- ============================================
-- CACHE STATISTICS VIEW
-- ============================================

CREATE OR REPLACE VIEW cache.usage_stats AS
SELECT
    'key_value_cache' AS category,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE last_accessed > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(*) FILTER (WHERE last_accessed > NOW() - INTERVAL '1 day') AS active_last_day,
    COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at < NOW()) AS expired_count
FROM cache.store
UNION ALL
SELECT
    'sessions' AS category,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE last_activity > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(*) FILTER (WHERE last_activity > NOW() - INTERVAL '1 day') AS active_last_day,
    COUNT(*) FILTER (WHERE expires_at < NOW()) AS expired_count
FROM cache.sessions
UNION ALL
SELECT
    'embeddings' AS category,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE last_hit > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(*) FILTER (WHERE last_hit > NOW() - INTERVAL '1 day') AS active_last_day,
    COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at < NOW()) AS expired_count
FROM cache.embeddings
UNION ALL
SELECT
    'rate_limits' AS category,
    COUNT(DISTINCT limit_key) AS total_count,
    COUNT(DISTINCT limit_key) FILTER (WHERE window_start > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(DISTINCT limit_key) FILTER (WHERE window_start > NOW() - INTERVAL '1 day') AS active_last_day,
    0 AS expired_count
FROM cache.rate_limits;

-- Grant access to view
GRANT SELECT ON cache.usage_stats TO postgres, authenticated, service_role;

-- ============================================
-- SCHEDULED CLEANUP JOB
-- ============================================

-- Create a function to run all cleanup tasks
CREATE OR REPLACE FUNCTION cache.cleanup_all()
RETURNS TABLE (
    category TEXT,
    deleted_count INTEGER
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_cache_deleted INTEGER;
    v_sessions_deleted INTEGER;
    v_embeddings_deleted INTEGER;
    v_rate_limits_deleted INTEGER;
BEGIN
    -- Cleanup expired cache entries
    SELECT cache.cleanup_expired() INTO v_cache_deleted;

    -- Cleanup expired sessions
    SELECT cache.cleanup_expired_sessions() INTO v_sessions_deleted;

    -- Cleanup expired embeddings
    DELETE FROM cache.embeddings
    WHERE expires_at IS NOT NULL AND expires_at < NOW();
    GET DIAGNOSTICS v_embeddings_deleted = ROW_COUNT;

    -- Cleanup old rate limits (keep last hour)
    SELECT cache.cleanup_rate_limits(3600) INTO v_rate_limits_deleted;

    RETURN QUERY
    SELECT 'cache_store'::TEXT, v_cache_deleted
    UNION ALL
    SELECT 'sessions'::TEXT, v_sessions_deleted
    UNION ALL
    SELECT 'embeddings'::TEXT, v_embeddings_deleted
    UNION ALL
    SELECT 'rate_limits'::TEXT, v_rate_limits_deleted;
END;
$$;

ALTER FUNCTION cache.cleanup_all() SET search_path TO '';

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON SCHEMA cache IS 'PostgreSQL native caching using UNLOGGED tables for performance';
COMMENT ON TABLE cache.store IS 'Key-value cache store using UNLOGGED table for fast access';
COMMENT ON TABLE cache.sessions IS 'User session storage using UNLOGGED table';
COMMENT ON TABLE cache.embeddings IS 'Vector embedding cache using UNLOGGED table';
COMMENT ON TABLE cache.rate_limits IS 'Rate limiting counters using UNLOGGED table';
COMMENT ON VIEW cache.usage_stats IS 'Statistics on cache usage across different categories';

-- ============================================
-- ENABLE ROW LEVEL SECURITY
-- ============================================

ALTER TABLE cache.store ENABLE ROW LEVEL SECURITY;
ALTER TABLE cache.sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE cache.embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE cache.rate_limits ENABLE ROW LEVEL SECURITY;

-- Policies for cache.store (service role only for direct access)
CREATE POLICY "Service role has full access to cache store"
    ON cache.store FOR ALL
    TO service_role
    USING (true);

-- Policies for cache.sessions (users can only see their own sessions)
CREATE POLICY "Users can view their own sessions"
    ON cache.sessions FOR SELECT
    TO authenticated
    USING ((SELECT auth.uid()) = user_id);

CREATE POLICY "Users can create their own sessions"
    ON cache.sessions FOR INSERT
    TO authenticated
    WITH CHECK ((SELECT auth.uid()) = user_id);

CREATE POLICY "Users can update their own sessions"
    ON cache.sessions FOR UPDATE
    TO authenticated
    USING ((SELECT auth.uid()) = user_id);

CREATE POLICY "Users can delete their own sessions"
    ON cache.sessions FOR DELETE
    TO authenticated
    USING ((SELECT auth.uid()) = user_id);

CREATE POLICY "Service role has full access to sessions"
    ON cache.sessions FOR ALL
    TO service_role
    USING (true);

-- Policies for cache.embeddings (service role only)
CREATE POLICY "Service role has full access to embeddings cache"
    ON cache.embeddings FOR ALL
    TO service_role
    USING (true);

-- Policies for cache.rate_limits (service role and authenticated can check limits)
CREATE POLICY "Authenticated users can read rate limits"
    ON cache.rate_limits FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Service role has full access to rate limits"
    ON cache.rate_limits FOR ALL
    TO service_role
    USING (true);
