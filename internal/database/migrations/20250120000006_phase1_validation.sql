-- Phase 1 CRM: Contact Validation Rules
-- Custom validation rules per organization

CREATE TABLE contact_validation_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    field varchar(100) NOT NULL, -- email, phone, name, etc.
    rule_type varchar(50) NOT NULL, -- required, format, custom, length, enum
    validation_config jsonb NOT NULL, -- rule-specific configuration
    error_message varchar(500),
    is_active boolean DEFAULT true,
    severity varchar(20) DEFAULT 'error', -- error, warning, info
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT severity_check CHECK (severity IN ('error', 'warning', 'info')),
    CONSTRAINT rule_type_check CHECK (rule_type IN ('required', 'format', 'custom', 'length', 'enum', 'pattern', 'range'))
);

CREATE INDEX idx_validation_rules_org ON contact_validation_rules(organization_id);
CREATE INDEX idx_validation_rules_active ON contact_validation_rules(organization_id, is_active) WHERE is_active = true;
CREATE INDEX idx_validation_rules_field ON contact_validation_rules(organization_id, field, is_active);
CREATE INDEX idx_validation_rules_severity ON contact_validation_rules(organization_id, severity, is_active);

CREATE TRIGGER set_contact_validation_rules_updated_at
    BEFORE UPDATE ON contact_validation_rules
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Seed some default validation rules for all organizations
INSERT INTO contact_validation_rules (organization_id, name, field, rule_type, validation_config, error_message, is_active, severity)
SELECT
    id as organization_id,
    'Email Format Validation' as name,
    'email' as field,
    'format' as rule_type,
    '{"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"}'::jsonb as validation_config,
    'Email must be in valid format' as error_message,
    true as is_active,
    'error' as severity
FROM organizations;

INSERT INTO contact_validation_rules (organization_id, name, field, rule_type, validation_config, error_message, is_active, severity)
SELECT
    id as organization_id,
    'Name Required' as name,
    'name' as field,
    'required' as rule_type,
    '{}'::jsonb as validation_config,
    'Contact name is required' as error_message,
    true as is_active,
    'error' as severity
FROM organizations;

INSERT INTO contact_validation_rules (organization_id, name, field, rule_type, validation_config, error_message, is_active, severity)
SELECT
    id as organization_id,
    'Phone Format (Optional)' as name,
    'phone' as field,
    'pattern' as rule_type,
    '{"pattern": "^[+]?[(]?[0-9]{1,4}[)]?[-\\s.]?[(]?[0-9]{1,4}[)]?[-\\s.]?[0-9]{1,9}$", "allow_empty": true}'::jsonb as validation_config,
    'Phone number format is invalid' as error_message,
    true as is_active,
    'warning' as severity
FROM organizations;

-- Comments
COMMENT ON TABLE contact_validation_rules IS 'Organization-specific contact validation rules';
COMMENT ON COLUMN contact_validation_rules.field IS 'Contact field to validate (email, phone, name, etc.)';
COMMENT ON COLUMN contact_validation_rules.rule_type IS 'Validation type: required, format, custom, length, enum, pattern, range';
COMMENT ON COLUMN contact_validation_rules.validation_config IS 'JSONB configuration specific to the rule type';
COMMENT ON COLUMN contact_validation_rules.severity IS 'Severity: error (blocks), warning (allows), info (informational)';
