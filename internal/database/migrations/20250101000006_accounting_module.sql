-- Migration: Accounting Module Tables
-- Description: Financial accounting and invoicing tables
-- Created: 2025-01-01

-- =====================================================
-- ACCOUNTING MODULE TABLES
-- =====================================================

-- Account Types
CREATE TABLE account_account_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    type varchar(50) NOT NULL,
    internal_group varchar(50),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Account Groups
CREATE TABLE account_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    code varchar(50),
    parent_id uuid REFERENCES account_groups(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Chart of Accounts
CREATE TABLE account_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(50) NOT NULL,
    deprecated boolean DEFAULT false,
    account_type varchar(50) NOT NULL,
    internal_type varchar(50),
    internal_group varchar(50),
    user_type_id uuid REFERENCES account_account_types(id),
    reconcile boolean DEFAULT false,
    currency_id uuid REFERENCES currencies(id),
    group_id uuid REFERENCES account_groups(id),
    tax_ids uuid[],
    note text,
    tag_ids uuid[],
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    CONSTRAINT account_accounts_unique UNIQUE(organization_id, company_id, code)
);

-- Account Journals
CREATE TABLE account_journals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(10) NOT NULL,
    type varchar(20) NOT NULL,
    default_account_id uuid REFERENCES account_accounts(id),
    refund_sequence boolean DEFAULT false,
    sequence_id uuid REFERENCES sequences(id),
    currency_id uuid REFERENCES currencies(id),
    bank_account_id uuid REFERENCES bank_accounts(id),
    color integer,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT account_journals_unique UNIQUE(organization_id, company_id, code),
    CONSTRAINT account_journals_type_check CHECK (type IN ('sale', 'purchase', 'cash', 'bank', 'general'))
);

-- Tax Groups
CREATE TABLE account_tax_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    sequence integer DEFAULT 10,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Taxes
CREATE TABLE account_taxes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    type_tax_use varchar(20),
    amount_type varchar(20) DEFAULT 'percent',
    amount numeric(15,4) DEFAULT 0,
    price_include boolean DEFAULT false,
    include_base_amount boolean DEFAULT false,
    is_base_affected boolean DEFAULT false,
    description varchar(50),
    sequence integer DEFAULT 10,
    active boolean DEFAULT true,
    tax_group_id uuid REFERENCES account_tax_groups(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT account_taxes_use_check CHECK (type_tax_use IN ('sale', 'purchase', 'none')),
    CONSTRAINT account_taxes_amount_type_check CHECK (amount_type IN ('group', 'fixed', 'percent', 'division'))
);

-- Invoices/Account Moves (account.move)
CREATE TABLE invoices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255),
    move_type varchar(30) NOT NULL,
    date date NOT NULL DEFAULT CURRENT_DATE,
    invoice_date date,
    invoice_date_due date,
    ref varchar(255),
    state varchar(20) DEFAULT 'draft',
    -- Partner
    partner_id uuid NOT NULL REFERENCES contacts(id),
    commercial_partner_id uuid REFERENCES contacts(id),
    partner_shipping_id uuid REFERENCES contacts(id),
    partner_bank_id uuid REFERENCES bank_accounts(id),
    -- Amounts
    amount_untaxed numeric(15,2) DEFAULT 0,
    amount_tax numeric(15,2) DEFAULT 0,
    amount_total numeric(15,2) DEFAULT 0,
    amount_residual numeric(15,2) DEFAULT 0,
    amount_untaxed_signed numeric(15,2) DEFAULT 0,
    amount_total_signed numeric(15,2) DEFAULT 0,
    -- Currency
    currency_id uuid REFERENCES currencies(id),
    -- Journal
    journal_id uuid NOT NULL REFERENCES account_journals(id),
    -- User
    user_id uuid,
    -- Invoice details
    invoice_origin varchar(255),
    invoice_payment_ref varchar(255),
    invoice_payment_term_id uuid REFERENCES payment_terms(id),
    fiscal_position_id uuid REFERENCES fiscal_positions(id),
    -- Payment
    payment_state varchar(20) DEFAULT 'not_paid',
    payment_reference varchar(255),
    -- Auto post
    auto_post varchar(20) DEFAULT 'no',
    to_check boolean DEFAULT false,
    -- Source documents
    invoice_source_email varchar(255),
    invoice_partner_display_name varchar(255),
    -- Reversal
    reversed_entry_id uuid REFERENCES invoices(id),
    -- Narration
    narration text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT invoices_type_check CHECK (move_type IN ('out_invoice', 'out_refund', 'in_invoice', 'in_refund', 'entry')),
    CONSTRAINT invoices_state_check CHECK (state IN ('draft', 'posted', 'cancel')),
    CONSTRAINT invoices_payment_state_check CHECK (payment_state IN ('not_paid', 'in_payment', 'paid', 'partial', 'reversed', 'invoicing_legacy'))
);

-- Account Full Reconcile
CREATE TABLE account_full_reconcile (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Invoice Lines/Account Move Lines (account.move.line)
CREATE TABLE invoice_lines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    move_id uuid NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    sequence integer DEFAULT 10,
    name text,
    -- Account
    account_id uuid NOT NULL REFERENCES account_accounts(id),
    -- Product
    product_id uuid REFERENCES products(id),
    product_uom_id uuid REFERENCES uom_units(id),
    quantity numeric(15,4) DEFAULT 1,
    price_unit numeric(15,2) DEFAULT 0,
    discount numeric(5,2) DEFAULT 0,
    -- Amounts
    debit numeric(15,2) DEFAULT 0,
    credit numeric(15,2) DEFAULT 0,
    balance numeric(15,2) DEFAULT 0,
    amount_currency numeric(15,2) DEFAULT 0,
    price_subtotal numeric(15,2) DEFAULT 0,
    price_total numeric(15,2) DEFAULT 0,
    -- Tax
    tax_ids uuid[],
    tax_line_id uuid REFERENCES account_taxes(id),
    tax_base_amount numeric(15,2) DEFAULT 0,
    -- Analytics
    analytic_account_id uuid REFERENCES analytic_accounts(id),
    analytic_tag_ids uuid[],
    -- Partner
    partner_id uuid REFERENCES contacts(id),
    -- Currency
    currency_id uuid REFERENCES currencies(id),
    -- Reconciliation
    reconciled boolean DEFAULT false,
    full_reconcile_id uuid REFERENCES account_full_reconcile(id),
    -- Misc
    date_maturity date,
    blocked boolean DEFAULT false,
    display_type varchar(20),
    exclude_from_invoice_tab boolean DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    CONSTRAINT invoice_lines_display_type_check CHECK (display_type IN ('line_section', 'line_note', 'product'))
);

-- Payments (account.payment)
CREATE TABLE payments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    payment_type varchar(20) NOT NULL,
    partner_type varchar(20) NOT NULL,
    partner_id uuid REFERENCES contacts(id),
    amount numeric(15,2) NOT NULL,
    currency_id uuid REFERENCES currencies(id),
    payment_date date NOT NULL DEFAULT CURRENT_DATE,
    communication varchar(255),
    journal_id uuid REFERENCES account_journals(id),
    payment_method_id uuid REFERENCES payment_methods(id),
    destination_account_id uuid REFERENCES account_accounts(id),
    state varchar(20) DEFAULT 'draft',
    name varchar(255),
    ref varchar(255),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT payments_payment_type_check CHECK (payment_type IN ('inbound', 'outbound')),
    CONSTRAINT payments_partner_type_check CHECK (partner_type IN ('customer', 'supplier')),
    CONSTRAINT payments_state_check CHECK (state IN ('draft', 'posted', 'sent', 'reconciled', 'cancelled'))
);

-- Payment-Invoice Allocation
CREATE TABLE payment_invoice_allocation (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payment_id uuid NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    invoice_id uuid NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount numeric(15,2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_account_accounts_updated_at
    BEFORE UPDATE ON account_accounts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_account_journals_updated_at
    BEFORE UPDATE ON account_journals
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_account_taxes_updated_at
    BEFORE UPDATE ON account_taxes
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_invoice_lines_updated_at
    BEFORE UPDATE ON invoice_lines
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_account_accounts_org ON account_accounts(organization_id);
CREATE INDEX idx_account_accounts_code ON account_accounts(organization_id, code);

CREATE INDEX idx_account_journals_org ON account_journals(organization_id);
CREATE INDEX idx_account_journals_type ON account_journals(organization_id, type);

CREATE INDEX idx_account_taxes_org ON account_taxes(organization_id);

CREATE INDEX idx_invoices_org ON invoices(organization_id);
CREATE INDEX idx_invoices_partner ON invoices(partner_id);
CREATE INDEX idx_invoices_state ON invoices(organization_id, state);
CREATE INDEX idx_invoices_payment_state ON invoices(organization_id, payment_state);
CREATE INDEX idx_invoices_date ON invoices(date);
CREATE INDEX idx_invoices_type ON invoices(organization_id, move_type);
CREATE INDEX idx_invoices_deleted_at ON invoices(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_invoice_lines_move ON invoice_lines(move_id);
CREATE INDEX idx_invoice_lines_account ON invoice_lines(account_id);
CREATE INDEX idx_invoice_lines_product ON invoice_lines(product_id) WHERE product_id IS NOT NULL;
CREATE INDEX idx_invoice_lines_partner ON invoice_lines(partner_id) WHERE partner_id IS NOT NULL;

CREATE INDEX idx_payments_org ON payments(organization_id);
CREATE INDEX idx_payments_partner ON payments(partner_id) WHERE partner_id IS NOT NULL;
CREATE INDEX idx_payments_state ON payments(organization_id, state);
CREATE INDEX idx_payments_date ON payments(payment_date);

CREATE INDEX idx_payment_invoice_alloc_payment ON payment_invoice_allocation(payment_id);
CREATE INDEX idx_payment_invoice_alloc_invoice ON payment_invoice_allocation(invoice_id);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE account_accounts IS 'Chart of accounts - GL accounts';
COMMENT ON TABLE account_journals IS 'Accounting journals (sales, purchase, bank, etc)';
COMMENT ON TABLE account_taxes IS 'Tax configuration (VAT, sales tax, etc)';
COMMENT ON TABLE invoices IS 'Invoices and bills - equivalent to account.move in Odoo';
COMMENT ON TABLE invoice_lines IS 'Invoice line items and journal entries - equivalent to account.move.line in Odoo';
COMMENT ON TABLE payments IS 'Payment records - equivalent to account.payment in Odoo';
COMMENT ON TABLE payment_invoice_allocation IS 'Links payments to invoices for reconciliation';
