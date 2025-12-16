-- Initialize development database
-- This script runs when the container starts

-- Create database if it doesn't exist
CREATE DATABASE alieze_erp_dev;

-- Connect to the database
\c alieze_erp_dev

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create test organization
INSERT INTO organizations (id, name, code, active)
VALUES ('d4e8f3b2-1a7a-4b1b-8c9d-0e1f2a3b4c5d', 'Test Organization', 'TEST-ORG', true)
ON CONFLICT (id) DO NOTHING;

-- Create test company
INSERT INTO companies (id, organization_id, name, code, active)
VALUES ('a1b2c3d4-5e6f-7g8h-9i0j-1k2l3m4n5o6p', 'd4e8f3b2-1a7a-4b1b-8c9d-0e1f2a3b4c5d', 'Test Company', 'TEST-COMP', true)
ON CONFLICT (id) DO NOTHING;
