-- Migration: Seed initial user-role assignments
-- Created: 2025-01-15
-- Description: Creates initial user-role mappings for testing

-- Note: In production, these would be created via the auth system
-- These are examples for development/testing

-- Assign example users to roles (using organization context)
-- Format: g, user:{user_id}:org:{org_id}, role:{role_name}

-- Example: Default admin user gets admin role
-- This should be replaced with actual user IDs from your auth system
INSERT INTO casbin_rules (ptype, v0, v1) VALUES
    ('g', 'user:00000000-0000-0000-0000-000000000001:org:00000000-0000-0000-0000-000000000001', 'role:admin')
ON CONFLICT DO NOTHING;

-- Add some test users for different roles (replace with actual user IDs)
-- INSERT INTO casbin_rules (ptype, v0, v1) VALUES
--     ('g', 'user:{user_id}:org:{org_id}', 'role:accountant'),
--     ('g', 'user:{user_id}:org:{org_id}', 'role:sales'),
--     ('g', 'user:{user_id}:org:{org_id}', 'role:viewer')
-- ON CONFLICT DO NOTHING;

COMMENT ON TABLE casbin_rules IS 'Updated with initial user-role assignments for development';
