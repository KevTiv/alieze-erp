-- Migration: Reference/Shared Tables
-- Description: Reference data tables shared across tenants or organizations
-- Created: 2025-01-01

-- =====================================================
-- REFERENCE TABLES (Shared Data)
-- =====================================================

-- Countries (ISO 3166-1)
CREATE TABLE countries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(2) NOT NULL UNIQUE,
    name varchar(255) NOT NULL,
    phone_code varchar(10),
    currency_id uuid,
    address_format text,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- States/Provinces (ISO 3166-2)
CREATE TABLE states (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    country_id uuid NOT NULL REFERENCES countries(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    code varchar(10) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT states_unique UNIQUE(country_id, code)
);

-- Currencies (ISO 4217)
CREATE TABLE currencies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL,
    symbol varchar(10) NOT NULL,
    code varchar(3) NOT NULL UNIQUE,
    rounding numeric(12,6) DEFAULT 0.01,
    decimal_places integer DEFAULT 2,
    position varchar(10) DEFAULT 'before',
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT currencies_position_check CHECK (position IN ('before', 'after'))
);

-- Unit of Measure Categories
CREATE TABLE uom_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Units of Measure
CREATE TABLE uom_units (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id uuid NOT NULL REFERENCES uom_categories(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    uom_type varchar(20) DEFAULT 'reference',
    factor numeric(12,6) DEFAULT 1.0,
    factor_inv numeric(12,6) DEFAULT 1.0,
    rounding numeric(12,6) DEFAULT 0.01,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT uom_units_type_check CHECK (uom_type IN ('reference', 'bigger', 'smaller'))
);

-- Payment Terms (Organization-specific)
CREATE TABLE payment_terms (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    note text,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Fiscal Positions (Tax mapping rules)
CREATE TABLE fiscal_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    auto_apply boolean DEFAULT false,
    vat_required boolean DEFAULT false,
    country_id uuid REFERENCES countries(id),
    state_ids uuid[],
    zip_from integer,
    zip_to integer,
    note text,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Analytic Accounts (Cost centers)
CREATE TABLE analytic_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(50),
    partner_id uuid, -- Will reference contacts later
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Industries
CREATE TABLE industries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    code varchar(50),
    full_name varchar(500),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- UTM Campaigns (Marketing tracking)
CREATE TABLE utm_campaigns (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- UTM Mediums
CREATE TABLE utm_mediums (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- UTM Sources
CREATE TABLE utm_sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Payment Methods
CREATE TABLE payment_methods (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    code varchar(50),
    payment_type varchar(20), -- inbound, outbound
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT payment_methods_type_check CHECK (payment_type IN ('inbound', 'outbound'))
);

-- Bank Accounts
CREATE TABLE bank_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    partner_id uuid, -- Will reference contacts
    acc_number varchar(255) NOT NULL,
    bank_name varchar(255),
    bank_bic varchar(11),
    sequence integer DEFAULT 10,
    currency_id uuid REFERENCES currencies(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- =====================================================
-- UPDATE FOREIGN KEYS IN FOUNDATION TABLES
-- =====================================================

-- Add foreign key constraints to organizations
ALTER TABLE organizations
    ADD CONSTRAINT organizations_currency_fk
    FOREIGN KEY (currency_id) REFERENCES currencies(id) NOT VALID;

ALTER TABLE organizations
    VALIDATE CONSTRAINT organizations_currency_fk;

-- Add foreign key constraints to companies
ALTER TABLE companies
    ADD CONSTRAINT companies_currency_fk
    FOREIGN KEY (currency_id) REFERENCES currencies(id) NOT VALID;

ALTER TABLE companies
    VALIDATE CONSTRAINT companies_currency_fk;

ALTER TABLE companies
    ADD CONSTRAINT companies_state_fk
    FOREIGN KEY (state_id) REFERENCES states(id) NOT VALID;

ALTER TABLE companies
    VALIDATE CONSTRAINT companies_state_fk;

ALTER TABLE companies
    ADD CONSTRAINT companies_country_fk
    FOREIGN KEY (country_id) REFERENCES countries(id) NOT VALID;

ALTER TABLE companies
    VALIDATE CONSTRAINT companies_country_fk;

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_payment_terms_updated_at
    BEFORE UPDATE ON payment_terms
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_fiscal_positions_updated_at
    BEFORE UPDATE ON fiscal_positions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_analytic_accounts_updated_at
    BEFORE UPDATE ON analytic_accounts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_bank_accounts_updated_at
    BEFORE UPDATE ON bank_accounts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_countries_code ON countries(code);
CREATE INDEX idx_states_country ON states(country_id);
CREATE INDEX idx_currencies_code ON currencies(code);
CREATE INDEX idx_currencies_active ON currencies(active) WHERE active = true;

CREATE INDEX idx_uom_units_category ON uom_units(category_id);
CREATE INDEX idx_uom_units_active ON uom_units(active) WHERE active = true;

CREATE INDEX idx_payment_terms_org ON payment_terms(organization_id);
CREATE INDEX idx_fiscal_positions_org ON fiscal_positions(organization_id);
CREATE INDEX idx_analytic_accounts_org ON analytic_accounts(organization_id);

CREATE INDEX idx_utm_campaigns_org ON utm_campaigns(organization_id);
CREATE INDEX idx_utm_mediums_org ON utm_mediums(organization_id);
CREATE INDEX idx_utm_sources_org ON utm_sources(organization_id);

CREATE INDEX idx_payment_methods_org ON payment_methods(organization_id);
CREATE INDEX idx_bank_accounts_org ON bank_accounts(organization_id);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE countries IS 'ISO 3166-1 country codes and data';
COMMENT ON TABLE states IS 'ISO 3166-2 state/province codes';
COMMENT ON TABLE currencies IS 'ISO 4217 currency codes and formatting';
COMMENT ON TABLE uom_categories IS 'Unit of measure categories (Weight, Length, Time, etc)';
COMMENT ON TABLE uom_units IS 'Units of measure (kg, m, hours, etc)';
COMMENT ON TABLE payment_terms IS 'Payment terms configuration (Net 30, Net 60, etc)';
COMMENT ON TABLE fiscal_positions IS 'Tax mapping rules for different regions';
COMMENT ON TABLE analytic_accounts IS 'Cost centers and analytic accounting';
COMMENT ON TABLE industries IS 'Industry classifications for contacts';
