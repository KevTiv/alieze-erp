-- Migration: Move Extensions from Public Schema
-- Description: Move all extensions to dedicated extensions schema
-- Created: 2025-01-01
-- Issue: Extensions should not be installed in the public schema for security and organization

-- =====================================================
-- CREATE EXTENSIONS SCHEMA
-- =====================================================

-- Create a dedicated schema for extensions
CREATE SCHEMA IF NOT EXISTS extensions;

-- Grant usage on the extensions schema
GRANT USAGE ON SCHEMA extensions TO postgres, anon, authenticated, service_role;

-- =====================================================
-- MOVE EXTENSIONS TO EXTENSIONS SCHEMA
-- =====================================================

-- Note: Extensions cannot be "moved" directly in PostgreSQL
-- We need to drop and recreate them in the correct schema

-- Drop all extensions from public schema (if they exist)
DROP EXTENSION IF EXISTS pg_trgm CASCADE;
DROP EXTENSION IF EXISTS http CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
DROP EXTENSION IF EXISTS vector CASCADE;
DROP EXTENSION IF EXISTS pgmq CASCADE;

-- Recreate extensions in the extensions schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS pg_trgm SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS vector SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS http SCHEMA extensions;
CREATE EXTENSION IF NOT EXISTS pgmq SCHEMA extensions CASCADE;

-- =====================================================
-- UPDATE SEARCH PATH (Optional but recommended)
-- =====================================================

-- Add extensions schema to the default search path so functions can find the extensions
-- This is especially important for pg_trgm operators and http functions
ALTER DATABASE postgres SET search_path TO public, extensions;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON SCHEMA extensions IS 'Dedicated schema for PostgreSQL extensions';
COMMENT ON EXTENSION "uuid-ossp" IS 'Generate universally unique identifiers (UUIDs)';
COMMENT ON EXTENSION pg_trgm IS 'Text similarity measurement and index searching based on trigrams';
COMMENT ON EXTENSION vector IS 'Vector similarity search for AI embeddings';
COMMENT ON EXTENSION http IS 'HTTP client for PostgreSQL to make HTTP requests from SQL';
COMMENT ON EXTENSION pgmq IS 'PostgreSQL Message Queue for async job processing';
