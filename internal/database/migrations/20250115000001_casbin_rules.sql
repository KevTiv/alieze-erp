-- Migration: Create Casbin rules table for RBAC policy enforcement
-- Created: 2025-01-15
-- Description: This table stores Casbin policies for role-based access control

-- Create casbin_rules table
CREATE TABLE IF NOT EXISTS casbin_rules (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100),
    CONSTRAINT unique_key_casbin_rules UNIQUE(ptype, v0, v1, v2, v3, v4, v5)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_casbin_rules_ptype ON casbin_rules(ptype);
CREATE INDEX IF NOT EXISTS idx_casbin_rules_v0 ON casbin_rules(v0);
CREATE INDEX IF NOT EXISTS idx_casbin_rules_v1 ON casbin_rules(v1);

-- Insert default admin policies
-- Format: p, role, resource, action
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES
    ('p', 'role:admin', 'contacts', 'read'),
    ('p', 'role:admin', 'contacts', 'create'),
    ('p', 'role:admin', 'contacts', 'update'),
    ('p', 'role:admin', 'contacts', 'delete'),
    ('p', 'role:admin', 'invoices', 'read'),
    ('p', 'role:admin', 'invoices', 'create'),
    ('p', 'role:admin', 'invoices', 'update'),
    ('p', 'role:admin', 'invoices', 'delete'),
    ('p', 'role:admin', 'invoices', 'confirm'),
    ('p', 'role:admin', 'invoices', 'pay'),
    ('p', 'role:admin', 'orders', 'read'),
    ('p', 'role:admin', 'orders', 'create'),
    ('p', 'role:admin', 'orders', 'update'),
    ('p', 'role:admin', 'orders', 'delete'),
    ('p', 'role:admin', 'orders', 'confirm'),
    ('p', 'role:admin', 'products', 'read'),
    ('p', 'role:admin', 'products', 'create'),
    ('p', 'role:admin', 'products', 'update'),
    ('p', 'role:admin', 'products', 'delete')
ON CONFLICT DO NOTHING;

-- Insert accountant role policies
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES
    ('p', 'role:accountant', 'contacts', 'read'),
    ('p', 'role:accountant', 'invoices', 'read'),
    ('p', 'role:accountant', 'invoices', 'create'),
    ('p', 'role:accountant', 'invoices', 'update'),
    ('p', 'role:accountant', 'invoices', 'confirm'),
    ('p', 'role:accountant', 'invoices', 'pay'),
    ('p', 'role:accountant', 'orders', 'read')
ON CONFLICT DO NOTHING;

-- Insert sales role policies
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES
    ('p', 'role:sales', 'contacts', 'read'),
    ('p', 'role:sales', 'contacts', 'create'),
    ('p', 'role:sales', 'contacts', 'update'),
    ('p', 'role:sales', 'orders', 'read'),
    ('p', 'role:sales', 'orders', 'create'),
    ('p', 'role:sales', 'orders', 'update'),
    ('p', 'role:sales', 'orders', 'confirm'),
    ('p', 'role:sales', 'products', 'read')
ON CONFLICT DO NOTHING;

-- Insert viewer role policies (read-only)
INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES
    ('p', 'role:viewer', 'contacts', 'read'),
    ('p', 'role:viewer', 'invoices', 'read'),
    ('p', 'role:viewer', 'orders', 'read'),
    ('p', 'role:viewer', 'products', 'read')
ON CONFLICT DO NOTHING;

-- Example grouping policies (users to roles) - these would be created dynamically
-- Format: g, user, role
-- INSERT INTO casbin_rules (ptype, v0, v1) VALUES
--     ('g', 'user:alice', 'role:admin'),
--     ('g', 'user:bob', 'role:accountant'),
--     ('g', 'user:charlie', 'role:sales');

COMMENT ON TABLE casbin_rules IS 'Stores Casbin RBAC policies for authorization';
COMMENT ON COLUMN casbin_rules.ptype IS 'Policy type: p (policy), g (grouping/role)';
COMMENT ON COLUMN casbin_rules.v0 IS 'For p: subject (role), For g: user';
COMMENT ON COLUMN casbin_rules.v1 IS 'For p: object (resource), For g: role';
COMMENT ON COLUMN casbin_rules.v2 IS 'For p: action, For g: unused';
