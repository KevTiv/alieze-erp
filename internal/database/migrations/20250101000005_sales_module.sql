-- Migration: Sales Module Tables
-- Description: Sales order management tables
-- Created: 2025-01-01

-- =====================================================
-- SALES MODULE TABLES
-- =====================================================

-- Pricelists
CREATE TABLE pricelists (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    currency_id uuid REFERENCES currencies(id),
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Sales Orders (sale.order)
CREATE TABLE sales_orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    date_order timestamptz DEFAULT now(),
    validity_date date,
    confirmation_date timestamptz,
    partner_id uuid NOT NULL REFERENCES contacts(id),
    partner_invoice_id uuid REFERENCES contacts(id),
    partner_shipping_id uuid REFERENCES contacts(id),
    -- Amounts
    amount_untaxed numeric(15,2) DEFAULT 0,
    amount_tax numeric(15,2) DEFAULT 0,
    amount_total numeric(15,2) DEFAULT 0,
    amount_discount numeric(15,2) DEFAULT 0,
    -- State
    state varchar(20) DEFAULT 'draft',
    invoice_status varchar(20) DEFAULT 'no',
    delivery_status varchar(20) DEFAULT 'no',
    -- Sales team
    user_id uuid,
    team_id uuid REFERENCES sales_teams(id),
    -- Payment
    payment_term_id uuid REFERENCES payment_terms(id),
    fiscal_position_id uuid REFERENCES fiscal_positions(id),
    pricelist_id uuid REFERENCES pricelists(id),
    currency_id uuid REFERENCES currencies(id),
    -- References
    client_order_ref varchar(255),
    origin varchar(255),
    campaign_id uuid REFERENCES utm_campaigns(id),
    medium_id uuid REFERENCES utm_mediums(id),
    source_id uuid REFERENCES utm_sources(id),
    -- Misc
    note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT sales_orders_state_check CHECK (state IN ('draft', 'sent', 'sale', 'done', 'cancel')),
    CONSTRAINT sales_orders_invoice_status_check CHECK (invoice_status IN ('no', 'to invoice', 'invoiced')),
    CONSTRAINT sales_orders_delivery_status_check CHECK (delivery_status IN ('no', 'to deliver', 'delivered'))
);

-- Sales Order Lines (sale.order.line)
CREATE TABLE sales_order_lines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    order_id uuid NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    sequence integer DEFAULT 10,
    name text NOT NULL,
    product_id uuid REFERENCES products(id),
    product_uom_qty numeric(15,4) DEFAULT 1,
    product_uom uuid REFERENCES uom_units(id),
    -- Pricing
    price_unit numeric(15,2) DEFAULT 0,
    discount numeric(5,2) DEFAULT 0,
    price_subtotal numeric(15,2) DEFAULT 0,
    price_tax numeric(15,2) DEFAULT 0,
    price_total numeric(15,2) DEFAULT 0,
    tax_ids uuid[],
    -- Delivery
    qty_delivered numeric(15,4) DEFAULT 0,
    qty_delivered_method varchar(20),
    qty_to_invoice numeric(15,4) DEFAULT 0,
    qty_invoiced numeric(15,4) DEFAULT 0,
    -- State
    state varchar(20) DEFAULT 'draft',
    invoice_status varchar(20) DEFAULT 'no',
    -- Misc
    customer_lead integer DEFAULT 0,
    display_type varchar(20),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT sales_order_lines_display_type_check CHECK (display_type IN ('line_section', 'line_note'))
);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_pricelists_updated_at
    BEFORE UPDATE ON pricelists
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_sales_orders_updated_at
    BEFORE UPDATE ON sales_orders
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_sales_order_lines_updated_at
    BEFORE UPDATE ON sales_order_lines
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_pricelists_org ON pricelists(organization_id);

CREATE INDEX idx_sales_orders_org ON sales_orders(organization_id);
CREATE INDEX idx_sales_orders_partner ON sales_orders(partner_id);
CREATE INDEX idx_sales_orders_user ON sales_orders(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_sales_orders_team ON sales_orders(team_id) WHERE team_id IS NOT NULL;
CREATE INDEX idx_sales_orders_state ON sales_orders(organization_id, state);
CREATE INDEX idx_sales_orders_date ON sales_orders(date_order);
CREATE INDEX idx_sales_orders_deleted_at ON sales_orders(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_sales_order_lines_order ON sales_order_lines(order_id);
CREATE INDEX idx_sales_order_lines_product ON sales_order_lines(product_id) WHERE product_id IS NOT NULL;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE pricelists IS 'Pricing rules and lists for customers';
COMMENT ON TABLE sales_orders IS 'Sales orders - equivalent to sale.order in Odoo';
COMMENT ON TABLE sales_order_lines IS 'Sales order line items - equivalent to sale.order.line in Odoo';
