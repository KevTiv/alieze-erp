-- Migration: Redis Integration via redis_fdw
-- Description: Enables Redis access from PostgreSQL for caching and session management
-- Created: 2025-01-25

-- ============================================
-- REDIS FOREIGN DATA WRAPPER SETUP
-- ============================================

-- Install redis_fdw extension (if available)
-- Note: This requires redis_fdw to be installed in the Postgres container
-- The supabase/postgres image may not include this by default
-- You can alternatively use HTTP calls from edge functions to access Redis

-- For now, we'll create a schema for Redis-related utilities
CREATE SCHEMA IF NOT EXISTS redis;

-- Grant access to the schema
GRANT USAGE ON SCHEMA redis TO postgres, anon, authenticated, service_role;

-- ============================================
-- REDIS CONNECTION HELPERS
-- ============================================

-- Create a function to store Redis connection info
-- This can be used by Edge Functions to connect to Redis
CREATE TABLE IF NOT EXISTS redis.config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert Redis connection details
INSERT INTO redis.config (key, value, description)
VALUES
    ('redis_host', 'redis', 'Redis hostname (Docker service name)'),
    ('redis_port', '6379', 'Redis port'),
    ('redis_db', '0', 'Default Redis database number')
ON CONFLICT (key) DO NOTHING;

-- Grant read access to Redis config
GRANT SELECT ON redis.config TO postgres, authenticated, service_role;

-- ============================================
-- CACHE MANAGEMENT UTILITIES
-- ============================================

-- Create a table to track what's cached in Redis
CREATE TABLE IF NOT EXISTS redis.cache_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key TEXT NOT NULL UNIQUE,
    description TEXT,
    ttl_seconds INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed TIMESTAMPTZ DEFAULT NOW()
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_cache_keys_key ON redis.cache_keys(cache_key);

-- Grant access
GRANT SELECT, INSERT, UPDATE, DELETE ON redis.cache_keys TO postgres, authenticated, service_role;

-- Function to register a cache key
CREATE OR REPLACE FUNCTION redis.register_cache_key(
    p_cache_key TEXT,
    p_description TEXT DEFAULT NULL,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS UUID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_id UUID;
BEGIN
    INSERT INTO redis.cache_keys (cache_key, description, ttl_seconds)
    VALUES (p_cache_key, p_description, p_ttl_seconds)
    ON CONFLICT (cache_key) DO UPDATE
    SET
        description = COALESCE(EXCLUDED.description, redis.cache_keys.description),
        ttl_seconds = COALESCE(EXCLUDED.ttl_seconds, redis.cache_keys.ttl_seconds),
        last_accessed = NOW()
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$$;

ALTER FUNCTION redis.register_cache_key(text, text, integer) SET search_path TO '';

-- ============================================
-- SESSION STORAGE HELPERS
-- ============================================

-- Table to track active sessions stored in Redis
CREATE TABLE IF NOT EXISTS redis.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    session_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    last_activity TIMESTAMPTZ DEFAULT NOW()
);

-- Index for faster user lookups
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON redis.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON redis.sessions(expires_at);

-- Grant access
GRANT SELECT, INSERT, UPDATE, DELETE ON redis.sessions TO postgres, authenticated, service_role;

-- Function to clean up expired sessions
CREATE OR REPLACE FUNCTION redis.cleanup_expired_sessions()
RETURNS INTEGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM redis.sessions
    WHERE expires_at < NOW();

    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;

    RETURN v_deleted_count;
END;
$$;

ALTER FUNCTION redis.cleanup_expired_sessions() SET search_path TO '';

-- ============================================
-- EMBEDDING CACHE TRACKING
-- ============================================

-- Track which embeddings are cached in Redis
CREATE TABLE IF NOT EXISTS redis.embedding_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash TEXT NOT NULL UNIQUE,
    cache_key TEXT NOT NULL,
    model_name TEXT NOT NULL,
    dimension INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    hit_count INTEGER DEFAULT 0,
    last_hit TIMESTAMPTZ
);

-- Index for lookups
CREATE INDEX IF NOT EXISTS idx_embedding_cache_hash ON redis.embedding_cache(content_hash);
CREATE INDEX IF NOT EXISTS idx_embedding_cache_key ON redis.embedding_cache(cache_key);

-- Grant access
GRANT SELECT, INSERT, UPDATE ON redis.embedding_cache TO postgres, authenticated, service_role;

-- Function to record cache hit
CREATE OR REPLACE FUNCTION redis.record_embedding_cache_hit(p_content_hash TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    UPDATE redis.embedding_cache
    SET
        hit_count = hit_count + 1,
        last_hit = NOW()
    WHERE content_hash = p_content_hash;

    RETURN FOUND;
END;
$$;

ALTER FUNCTION redis.record_embedding_cache_hit(text) SET search_path TO '';

-- ============================================
-- RATE LIMITING HELPERS
-- ============================================

-- Table to track rate limit keys in Redis
CREATE TABLE IF NOT EXISTS redis.rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    limit_key TEXT NOT NULL UNIQUE,
    description TEXT,
    max_requests INTEGER NOT NULL,
    window_seconds INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Grant access
GRANT SELECT, INSERT ON redis.rate_limits TO postgres, authenticated, service_role;

-- ============================================
-- REDIS STATISTICS VIEW
-- ============================================

-- Create a view for Redis usage statistics
CREATE OR REPLACE VIEW redis.usage_stats AS
SELECT
    'cache_keys' AS category,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE last_accessed > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(*) FILTER (WHERE last_accessed > NOW() - INTERVAL '1 day') AS active_last_day
FROM redis.cache_keys
UNION ALL
SELECT
    'sessions' AS category,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE last_activity > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(*) FILTER (WHERE last_activity > NOW() - INTERVAL '1 day') AS active_last_day
FROM redis.sessions
UNION ALL
SELECT
    'embedding_cache' AS category,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE last_hit > NOW() - INTERVAL '1 hour') AS active_last_hour,
    COUNT(*) FILTER (WHERE last_hit > NOW() - INTERVAL '1 day') AS active_last_day
FROM redis.embedding_cache;

-- Grant access to view
GRANT SELECT ON redis.usage_stats TO postgres, authenticated, service_role;

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON SCHEMA redis IS 'Redis integration utilities and cache management';
COMMENT ON TABLE redis.config IS 'Redis connection configuration';
COMMENT ON TABLE redis.cache_keys IS 'Tracks cache keys stored in Redis';
COMMENT ON TABLE redis.sessions IS 'Tracks user sessions stored in Redis';
COMMENT ON TABLE redis.embedding_cache IS 'Tracks embedding vectors cached in Redis';
COMMENT ON TABLE redis.rate_limits IS 'Defines rate limiting rules stored in Redis';
COMMENT ON VIEW redis.usage_stats IS 'Statistics on Redis usage across different categories';

-- ============================================
-- ENABLE ROW LEVEL SECURITY
-- ============================================

ALTER TABLE redis.config ENABLE ROW LEVEL SECURITY;
ALTER TABLE redis.cache_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE redis.sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE redis.embedding_cache ENABLE ROW LEVEL SECURITY;
ALTER TABLE redis.rate_limits ENABLE ROW LEVEL SECURITY;

-- Policies for redis.config (read-only for authenticated users)
CREATE POLICY "Allow read access to config for authenticated users"
    ON redis.config FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Allow full access to config for service role"
    ON redis.config FOR ALL
    TO service_role
    USING (true);

-- Policies for redis.sessions (users can only see their own sessions)
CREATE POLICY "Users can view their own sessions"
    ON redis.sessions FOR SELECT
    TO authenticated
    USING ((SELECT auth.uid()) = user_id);

CREATE POLICY "Users can create their own sessions"
    ON redis.sessions FOR INSERT
    TO authenticated
    WITH CHECK ((SELECT auth.uid()) = user_id);

CREATE POLICY "Users can update their own sessions"
    ON redis.sessions FOR UPDATE
    TO authenticated
    USING ((SELECT auth.uid()) = user_id);

CREATE POLICY "Users can delete their own sessions"
    ON redis.sessions FOR DELETE
    TO authenticated
    USING ((SELECT auth.uid()) = user_id);

CREATE POLICY "Service role has full access to sessions"
    ON redis.sessions FOR ALL
    TO service_role
    USING (true);

-- Policies for cache_keys (service role only)
CREATE POLICY "Service role has full access to cache_keys"
    ON redis.cache_keys FOR ALL
    TO service_role
    USING (true);

-- Policies for embedding_cache (service role only)
CREATE POLICY "Service role has full access to embedding_cache"
    ON redis.embedding_cache FOR ALL
    TO service_role
    USING (true);

-- Policies for rate_limits (read for authenticated, write for service role)
CREATE POLICY "Allow read access to rate_limits for authenticated users"
    ON redis.rate_limits FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Service role has full access to rate_limits"
    ON redis.rate_limits FOR ALL
    TO service_role
    USING (true);
