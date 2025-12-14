-- Migration: CRM Module Tables
-- Description: Customer Relationship Management tables
-- Created: 2025-01-01

-- =====================================================
-- CRM MODULE TABLES
-- =====================================================

-- Contact Tags
CREATE TABLE contact_tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    color integer DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Sales Teams
CREATE TABLE sales_teams (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(50),
    team_leader_id uuid,
    member_ids uuid[],
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Contacts (res.partner equivalent - customers, vendors, etc.)
CREATE TABLE contacts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    contact_type varchar(20) DEFAULT 'person',
    name varchar(255) NOT NULL,
    display_name varchar(255),
    email varchar(255),
    phone varchar(50),
    mobile varchar(50),
    fax varchar(50),
    website varchar(255),
    title varchar(50),
    job_position varchar(100),
    -- Relationship
    parent_id uuid REFERENCES contacts(id),
    is_company boolean DEFAULT false,
    -- Contact classification
    is_customer boolean DEFAULT false,
    is_vendor boolean DEFAULT false,
    is_employee boolean DEFAULT false,
    -- Address
    street varchar(255),
    street2 varchar(255),
    city varchar(100),
    state_id uuid REFERENCES states(id),
    zip varchar(20),
    country_id uuid REFERENCES countries(id),
    -- Business info
    language varchar(10) DEFAULT 'en_US',
    timezone varchar(100),
    tax_id varchar(100),
    reference varchar(100),
    industry_id uuid REFERENCES industries(id),
    -- Image and color
    image_url varchar(500),
    color integer,
    -- Communication tracking
    last_contact_date timestamptz,
    next_contact_date timestamptz,
    -- Sales info
    user_id uuid,
    team_id uuid REFERENCES sales_teams(id),
    -- Payment info
    payment_term_id uuid REFERENCES payment_terms(id),
    -- Misc
    comment text,
    tags uuid[],
    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT contacts_type_check CHECK (contact_type IN ('person', 'company'))
);

-- Update analytic_accounts foreign key
ALTER TABLE analytic_accounts
    ADD CONSTRAINT analytic_accounts_partner_fk
    FOREIGN KEY (partner_id) REFERENCES contacts(id) NOT VALID;

ALTER TABLE analytic_accounts
    VALIDATE CONSTRAINT analytic_accounts_partner_fk;

-- Update bank_accounts foreign key
ALTER TABLE bank_accounts
    ADD CONSTRAINT bank_accounts_partner_fk
    FOREIGN KEY (partner_id) REFERENCES contacts(id) NOT VALID;

ALTER TABLE bank_accounts
    VALIDATE CONSTRAINT bank_accounts_partner_fk;

-- Lead Stages
CREATE TABLE lead_stages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    sequence integer DEFAULT 10,
    probability integer DEFAULT 0,
    fold boolean DEFAULT false,
    is_won boolean DEFAULT false,
    requirements text,
    team_id uuid REFERENCES sales_teams(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT lead_stages_probability_check CHECK (probability >= 0 AND probability <= 100)
);

-- Lead Sources
CREATE TABLE lead_sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Lost Reasons
CREATE TABLE lost_reasons (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Leads/Opportunities (crm.lead)
CREATE TABLE leads (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    contact_name varchar(255),
    email varchar(255),
    phone varchar(50),
    mobile varchar(50),
    contact_id uuid REFERENCES contacts(id),
    user_id uuid,
    team_id uuid REFERENCES sales_teams(id),
    -- Lead details
    lead_type varchar(20) DEFAULT 'lead',
    stage_id uuid REFERENCES lead_stages(id),
    priority varchar(20) DEFAULT 'medium',
    source_id uuid REFERENCES lead_sources(id),
    medium_id uuid REFERENCES utm_mediums(id),
    campaign_id uuid REFERENCES utm_campaigns(id),
    -- Revenue
    expected_revenue numeric(15,2),
    probability integer DEFAULT 0,
    recurring_revenue numeric(15,2),
    recurring_plan varchar(50),
    -- Dates
    date_open timestamptz,
    date_closed timestamptz,
    date_deadline timestamptz,
    date_last_stage_update timestamptz,
    -- Status
    active boolean DEFAULT true,
    won_status varchar(20),
    lost_reason_id uuid REFERENCES lost_reasons(id),
    -- Contact info
    street varchar(255),
    street2 varchar(255),
    city varchar(100),
    state_id uuid REFERENCES states(id),
    zip varchar(20),
    country_id uuid REFERENCES countries(id),
    website varchar(255),
    -- Description
    description text,
    tag_ids uuid[],
    color integer,
    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT leads_type_check CHECK (lead_type IN ('lead', 'opportunity')),
    CONSTRAINT leads_priority_check CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT leads_won_status_check CHECK (won_status IN ('won', 'lost', 'ongoing')),
    CONSTRAINT leads_probability_check CHECK (probability >= 0 AND probability <= 100)
);

-- Activities (Scheduled activities on records)
CREATE TABLE activities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    activity_type varchar(50) NOT NULL,
    summary text NOT NULL,
    note text,
    date_deadline date,
    user_id uuid,
    assigned_to uuid,
    -- Polymorphic reference to any record
    res_model varchar(100),
    res_id uuid,
    -- State
    state varchar(20) DEFAULT 'planned',
    done_date timestamptz,
    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    CONSTRAINT activities_type_check CHECK (activity_type IN ('call', 'meeting', 'email', 'todo', 'note')),
    CONSTRAINT activities_state_check CHECK (state IN ('planned', 'done', 'cancelled'))
);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_sales_teams_updated_at
    BEFORE UPDATE ON sales_teams
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_contacts_updated_at
    BEFORE UPDATE ON contacts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_lead_stages_updated_at
    BEFORE UPDATE ON lead_stages
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_leads_updated_at
    BEFORE UPDATE ON leads
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_activities_updated_at
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_contact_tags_org ON contact_tags(organization_id);

CREATE INDEX idx_sales_teams_org ON sales_teams(organization_id);
CREATE INDEX idx_sales_teams_leader ON sales_teams(team_leader_id) WHERE team_leader_id IS NOT NULL;

CREATE INDEX idx_contacts_org ON contacts(organization_id);
CREATE INDEX idx_contacts_parent ON contacts(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_contacts_customer ON contacts(organization_id, is_customer) WHERE is_customer = true;
CREATE INDEX idx_contacts_vendor ON contacts(organization_id, is_vendor) WHERE is_vendor = true;
CREATE INDEX idx_contacts_email ON contacts(email) WHERE email IS NOT NULL;
CREATE INDEX idx_contacts_name_trgm ON contacts USING gin(name gin_trgm_ops);
CREATE INDEX idx_contacts_deleted_at ON contacts(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_lead_stages_org ON lead_stages(organization_id);
CREATE INDEX idx_lead_sources_org ON lead_sources(organization_id);
CREATE INDEX idx_lost_reasons_org ON lost_reasons(organization_id);

CREATE INDEX idx_leads_org ON leads(organization_id);
CREATE INDEX idx_leads_contact ON leads(contact_id) WHERE contact_id IS NOT NULL;
CREATE INDEX idx_leads_user ON leads(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_leads_team ON leads(team_id) WHERE team_id IS NOT NULL;
CREATE INDEX idx_leads_stage ON leads(stage_id) WHERE stage_id IS NOT NULL;
CREATE INDEX idx_leads_type ON leads(organization_id, lead_type);
CREATE INDEX idx_leads_deleted_at ON leads(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_activities_org ON activities(organization_id);
CREATE INDEX idx_activities_assigned ON activities(assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX idx_activities_res ON activities(res_model, res_id);
CREATE INDEX idx_activities_deadline ON activities(date_deadline) WHERE date_deadline IS NOT NULL;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE contact_tags IS 'Tags for categorizing contacts';
COMMENT ON TABLE sales_teams IS 'Sales team organization and hierarchy';
COMMENT ON TABLE contacts IS 'Unified contact management (customers, vendors, employees) - equivalent to res.partner in Odoo';
COMMENT ON TABLE lead_stages IS 'CRM pipeline stages configuration';
COMMENT ON TABLE lead_sources IS 'Lead source tracking (website, referral, etc)';
COMMENT ON TABLE lost_reasons IS 'Reasons for lost opportunities';
COMMENT ON TABLE leads IS 'CRM leads and opportunities tracking';
COMMENT ON TABLE activities IS 'Scheduled activities on any record (calls, meetings, todos)';
