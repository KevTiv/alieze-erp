-- Migration: POS Core Module
-- Description: Point of Sales core tables for retail/store operations
-- Created: 2025-01-26
-- Dependencies: sales_module, products_inventory_module, accounting_module

-- =====================================================
-- POS PAYMENT METHODS
-- =====================================================

CREATE TABLE pos_payment_methods (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),

    -- Identity
    name varchar(255) NOT NULL,
    code varchar(50) NOT NULL,

    -- Payment Type
    type varchar(20) NOT NULL, -- 'cash', 'card', 'bank', 'mobile', 'voucher', 'pay_later'

    -- Integration Configuration
    use_payment_terminal boolean DEFAULT false,
    payment_terminal_type varchar(50), -- 'stripe', 'square', 'adyen', 'paypal', null for manual
    payment_terminal_config jsonb DEFAULT '{}'::jsonb,

    -- Accounting Integration
    journal_id uuid REFERENCES account_journals(id),
    receivable_account_id uuid REFERENCES account_accounts(id),
    outstanding_account_id uuid REFERENCES account_accounts(id),

    -- Behavior
    split_transactions boolean DEFAULT true,
    requires_authorization boolean DEFAULT false,

    -- Display
    icon_name varchar(50), -- For UI: 'cash', 'credit-card', 'smartphone', etc.
    sequence integer DEFAULT 10,

    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    CONSTRAINT pos_payment_methods_type_check CHECK (type IN ('cash', 'card', 'bank', 'mobile', 'voucher', 'pay_later')),
    CONSTRAINT pos_payment_methods_unique UNIQUE(organization_id, code)
);

-- =====================================================
-- POS CONFIGURATION (Terminal/Register Setup)
-- =====================================================

CREATE TABLE pos_config (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),

    -- Identity
    name varchar(255) NOT NULL,
    code varchar(50) NOT NULL,

    -- Location & Warehouse
    warehouse_id uuid NOT NULL REFERENCES warehouses(id),
    stock_location_id uuid REFERENCES stock_locations(id),

    -- Pricing
    pricelist_id uuid NOT NULL REFERENCES pricelists(id),
    currency_id uuid REFERENCES currencies(id),
    available_pricelist_ids uuid[] DEFAULT '{}',

    -- Payment Methods (allowed methods for this terminal)
    payment_method_ids uuid[] NOT NULL DEFAULT '{}',

    -- Session Settings
    cash_control boolean DEFAULT true,
    set_maximum_difference boolean DEFAULT false,
    maximum_difference numeric(10,2) DEFAULT 0,

    -- Product Selection
    iface_available_categ_ids uuid[] DEFAULT '{}', -- Restrict to specific categories
    limit_categories boolean DEFAULT false,

    -- Receipt Settings
    receipt_header text,
    receipt_footer text,
    receipt_template_config jsonb DEFAULT '{}'::jsonb,
    auto_print_receipt boolean DEFAULT true,
    email_receipt_enabled boolean DEFAULT true,

    -- Behavior & Validation
    manual_discount boolean DEFAULT true,
    discount_limit numeric(5,2) DEFAULT 100, -- Max % discount without approval
    manager_discount_limit numeric(5,2), -- Max % with manager approval
    require_customer boolean DEFAULT false,
    allow_negative_stock boolean DEFAULT false, -- If false, block sales when out of stock
    flag_negative_stock boolean DEFAULT true, -- If true, allow but flag for review

    -- Margin Visibility
    show_margin_to_cashier boolean DEFAULT false,
    show_cost_price boolean DEFAULT false,
    warn_below_margin_threshold boolean DEFAULT true,
    minimum_margin_threshold numeric(5,2) DEFAULT 0, -- Warn if margin below X%

    -- UI Preferences
    iface_customer_display boolean DEFAULT false,
    iface_scan_via_camera boolean DEFAULT true,
    iface_big_scrollbars boolean DEFAULT false,

    -- Future: Restaurant Module
    module_pos_restaurant boolean DEFAULT false,
    floor_ids uuid[] DEFAULT '{}', -- Link to future pos_floors table

    -- Accounting Integration
    journal_id uuid REFERENCES account_journals(id), -- POS journal for session entries
    invoice_journal_id uuid REFERENCES account_journals(id), -- Customer invoice journal

    -- State
    active boolean DEFAULT true,
    sequence integer DEFAULT 10,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT pos_config_unique UNIQUE(organization_id, code)
);

-- =====================================================
-- POS SESSIONS (Cash Drawer Sessions)
-- =====================================================

CREATE TABLE pos_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),

    -- Session Identity
    name varchar(255) NOT NULL, -- Auto-generated: "POS/YYYY-MM-DD/NNN"
    pos_config_id uuid NOT NULL REFERENCES pos_config(id),
    user_id uuid NOT NULL, -- Cashier/user who opened session

    -- Session Lifecycle
    state varchar(20) DEFAULT 'opening_control',
    start_at timestamptz,
    stop_at timestamptz,

    -- Cash Management
    cash_register_balance_start numeric(15,2) DEFAULT 0, -- Expected cash at start
    cash_register_balance_end numeric(15,2), -- Expected cash at close (calculated)
    cash_register_balance_end_real numeric(15,2), -- Actual counted cash at close
    cash_register_difference numeric(15,2), -- Difference (overage/shortage)

    -- Session Summary (updated via triggers/functions)
    total_orders_count integer DEFAULT 0,
    total_amount numeric(15,2) DEFAULT 0,
    total_tax_amount numeric(15,2) DEFAULT 0,

    -- Payment Method Breakdowns
    cash_payment_amount numeric(15,2) DEFAULT 0,
    card_payment_amount numeric(15,2) DEFAULT 0,
    other_payment_amount numeric(15,2) DEFAULT 0,

    -- Reconciliation & Accounting
    move_id uuid, -- Link to account_move after closing
    closing_notes text,
    reconciliation_data jsonb DEFAULT '{}'::jsonb, -- Detailed breakdown by payment method

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT pos_sessions_state_check CHECK (state IN ('opening_control', 'opened', 'closing_control', 'closed', 'posted'))
);

-- =====================================================
-- CASH MOVEMENTS (Cash In/Out During Session)
-- =====================================================

CREATE TABLE pos_cash_movements (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    session_id uuid NOT NULL REFERENCES pos_sessions(id) ON DELETE CASCADE,

    -- Movement Details
    name varchar(255) NOT NULL, -- Description: "Bank deposit", "Change order", "Petty cash"
    type varchar(10) NOT NULL, -- 'in', 'out'
    amount numeric(15,2) NOT NULL,

    -- Reason & Authorization
    reason text,
    authorized_by uuid, -- Manager who approved

    -- Accounting
    account_id uuid REFERENCES account_accounts(id),

    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,

    CONSTRAINT pos_cash_movements_type_check CHECK (type IN ('in', 'out'))
);

-- =====================================================
-- EXTEND SALES_ORDERS FOR POS
-- =====================================================

-- Add columns without defaults first to avoid table rewrites
ALTER TABLE sales_orders
ADD COLUMN IF NOT EXISTS pos_session_id uuid,
ADD COLUMN IF NOT EXISTS pos_order_ref varchar(100),
ADD COLUMN IF NOT EXISTS is_pos_order boolean,
ADD COLUMN IF NOT EXISTS pos_validated_at timestamptz,
ADD COLUMN IF NOT EXISTS pos_offline_uuid uuid, -- For offline draft sync
ADD COLUMN IF NOT EXISTS pos_synced_at timestamptz,
ADD COLUMN IF NOT EXISTS pos_draft_data jsonb,
ADD COLUMN IF NOT EXISTS pos_table_id uuid, -- Future: restaurant table reference
ADD COLUMN IF NOT EXISTS pos_order_type varchar(20);

-- Set defaults separately (no table rewrite)
ALTER TABLE sales_orders
ALTER COLUMN is_pos_order SET DEFAULT false,
ALTER COLUMN pos_draft_data SET DEFAULT '{}'::jsonb,
ALTER COLUMN pos_order_type SET DEFAULT 'retail';

-- Add foreign key constraint separately (can be validated later if needed)
ALTER TABLE sales_orders
ADD CONSTRAINT sales_orders_pos_session_id_fkey
FOREIGN KEY (pos_session_id) REFERENCES pos_sessions(id)
NOT VALID;

-- Optionally validate the constraint in a separate transaction to avoid long locks
-- ALTER TABLE sales_orders VALIDATE CONSTRAINT sales_orders_pos_session_id_fkey; -- 'retail', 'restaurant', 'offer_response'

-- Add constraint for pos_order_type
ALTER TABLE sales_orders
DROP CONSTRAINT IF EXISTS sales_orders_pos_order_type_check,
ADD CONSTRAINT sales_orders_pos_order_type_check
CHECK (pos_order_type IN ('retail', 'restaurant', 'offer_response', 'service'));

-- =====================================================
-- POS PAYMENTS (Individual Payment Records)
-- =====================================================

CREATE TABLE pos_payments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Links
    pos_session_id uuid NOT NULL REFERENCES pos_sessions(id) ON DELETE CASCADE,
    order_id uuid NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    payment_method_id uuid NOT NULL REFERENCES pos_payment_methods(id),

    -- Payment Details
    amount numeric(15,2) NOT NULL,
    payment_date timestamptz DEFAULT now(),

    -- Transaction Info (for card/mobile payments)
    transaction_id varchar(255), -- External payment processor reference
    card_last_four varchar(4),
    card_type varchar(20), -- 'visa', 'mastercard', 'amex', 'discover'
    authorization_code varchar(50),

    -- Special Cases
    is_change boolean DEFAULT false,

    -- Accounting Link
    account_payment_id uuid, -- Link to account_payments after session close

    -- Offline Sync
    offline_payment_uuid uuid, -- For offline payment tracking
    synced_at timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    metadata jsonb DEFAULT '{}'::jsonb
);

-- =====================================================
-- POS ORDER DISCOUNTS (Track Discounts Applied)
-- =====================================================

CREATE TABLE pos_order_discounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    order_id uuid NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,

    -- Discount Details
    discount_type varchar(20) NOT NULL, -- 'percentage', 'fixed_amount', 'coupon', 'loyalty'
    discount_value numeric(15,4) NOT NULL, -- Percentage or amount
    discount_amount numeric(15,2) NOT NULL, -- Calculated discount amount

    -- Applies To
    applied_to varchar(20) DEFAULT 'order', -- 'order', 'line'
    order_line_id uuid REFERENCES sales_order_lines(id),

    -- Reason/Source
    reason varchar(255),
    coupon_code varchar(100),
    promotion_id uuid, -- Link to future promotions/campaigns table
    loyalty_program_id uuid, -- Link to future loyalty programs

    -- Authorization
    authorized_by uuid, -- Required if discount exceeds limit

    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,

    CONSTRAINT pos_discounts_type_check CHECK (discount_type IN ('percentage', 'fixed_amount', 'coupon', 'loyalty')),
    CONSTRAINT pos_discounts_applied_to_check CHECK (applied_to IN ('order', 'line'))
);

-- =====================================================
-- POS INVENTORY ALERTS (Flag Low Stock/Negative Stock)
-- =====================================================

CREATE TABLE pos_inventory_alerts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Order & Product Reference
    order_id uuid NOT NULL REFERENCES sales_orders(id),
    order_line_id uuid REFERENCES sales_order_lines(id),
    product_id uuid NOT NULL REFERENCES products(id),

    -- Alert Details
    alert_type varchar(20) NOT NULL, -- 'negative_stock', 'low_stock', 'out_of_stock'
    quantity_sold numeric(15,4) NOT NULL,
    quantity_available numeric(15,4), -- Stock level at time of sale
    warehouse_id uuid REFERENCES warehouses(id),
    location_id uuid REFERENCES stock_locations(id),

    -- Resolution
    state varchar(20) DEFAULT 'open', -- 'open', 'acknowledged', 'resolved', 'ignored'
    acknowledged_by uuid,
    acknowledged_at timestamptz,
    resolution_notes text,

    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,

    CONSTRAINT pos_inventory_alerts_type_check CHECK (alert_type IN ('negative_stock', 'low_stock', 'out_of_stock')),
    CONSTRAINT pos_inventory_alerts_state_check CHECK (state IN ('open', 'acknowledged', 'resolved', 'ignored'))
);

-- =====================================================
-- POS PRICING OVERRIDES (Track Manual Price Changes)
-- =====================================================

CREATE TABLE pos_pricing_overrides (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Order Line Reference
    order_id uuid NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    order_line_id uuid NOT NULL REFERENCES sales_order_lines(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products(id),

    -- Pricing Details
    original_price numeric(15,2) NOT NULL,
    override_price numeric(15,2) NOT NULL,
    reason varchar(255),

    -- Margin Impact
    original_margin numeric(15,2),
    new_margin numeric(15,2),
    margin_loss numeric(15,2), -- Negative if margin decreased

    -- Authorization
    authorized_by uuid, -- Required if price below threshold

    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid
);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_pos_payment_methods_updated_at
    BEFORE UPDATE ON pos_payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_pos_config_updated_at
    BEFORE UPDATE ON pos_config
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_pos_sessions_updated_at
    BEFORE UPDATE ON pos_sessions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

-- Payment Methods
CREATE INDEX idx_pos_payment_methods_org ON pos_payment_methods(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pos_payment_methods_active ON pos_payment_methods(organization_id, active) WHERE active = true;

-- POS Config
CREATE INDEX idx_pos_config_org ON pos_config(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pos_config_warehouse ON pos_config(warehouse_id);
CREATE INDEX idx_pos_config_active ON pos_config(organization_id, active) WHERE active = true;

-- POS Sessions
CREATE INDEX idx_pos_sessions_org ON pos_sessions(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pos_sessions_config ON pos_sessions(pos_config_id);
CREATE INDEX idx_pos_sessions_user ON pos_sessions(user_id);
CREATE INDEX idx_pos_sessions_state ON pos_sessions(organization_id, state);
CREATE INDEX idx_pos_sessions_date ON pos_sessions(organization_id, start_at);

-- Cash Movements
CREATE INDEX idx_pos_cash_movements_session ON pos_cash_movements(session_id);
CREATE INDEX idx_pos_cash_movements_date ON pos_cash_movements(created_at);

-- Sales Orders - POS Extensions
CREATE INDEX idx_sales_orders_pos_session ON sales_orders(pos_session_id) WHERE pos_session_id IS NOT NULL;
CREATE INDEX idx_sales_orders_is_pos ON sales_orders(organization_id, is_pos_order, state) WHERE is_pos_order = true;
CREATE INDEX idx_sales_orders_pos_offline ON sales_orders(organization_id, pos_offline_uuid) WHERE pos_offline_uuid IS NOT NULL;

-- POS Payments
CREATE INDEX idx_pos_payments_session ON pos_payments(pos_session_id);
CREATE INDEX idx_pos_payments_order ON pos_payments(order_id);
CREATE INDEX idx_pos_payments_method ON pos_payments(payment_method_id);
CREATE INDEX idx_pos_payments_date ON pos_payments(payment_date);
CREATE INDEX idx_pos_payments_offline ON pos_payments(offline_payment_uuid) WHERE offline_payment_uuid IS NOT NULL;

-- POS Discounts
CREATE INDEX idx_pos_discounts_order ON pos_order_discounts(order_id);
CREATE INDEX idx_pos_discounts_type ON pos_order_discounts(organization_id, discount_type);
CREATE INDEX idx_pos_discounts_authorized ON pos_order_discounts(authorized_by) WHERE authorized_by IS NOT NULL;

-- POS Inventory Alerts
CREATE INDEX idx_pos_inventory_alerts_org ON pos_inventory_alerts(organization_id);
CREATE INDEX idx_pos_inventory_alerts_order ON pos_inventory_alerts(order_id);
CREATE INDEX idx_pos_inventory_alerts_product ON pos_inventory_alerts(product_id);
CREATE INDEX idx_pos_inventory_alerts_state ON pos_inventory_alerts(organization_id, state) WHERE state = 'open';
CREATE INDEX idx_pos_inventory_alerts_type ON pos_inventory_alerts(organization_id, alert_type);

-- POS Pricing Overrides
CREATE INDEX idx_pos_pricing_overrides_order ON pos_pricing_overrides(order_id);
CREATE INDEX idx_pos_pricing_overrides_product ON pos_pricing_overrides(product_id);
CREATE INDEX idx_pos_pricing_overrides_created ON pos_pricing_overrides(created_at);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE pos_payment_methods IS 'Available payment methods for POS terminals (cash, card, mobile, etc.)';
COMMENT ON TABLE pos_config IS 'POS terminal/register configuration linking warehouse, pricelist, and payment methods';
COMMENT ON TABLE pos_sessions IS 'Cash drawer sessions with opening/closing reconciliation';
COMMENT ON TABLE pos_cash_movements IS 'Cash in/out operations during a POS session (deposits, withdrawals)';
COMMENT ON TABLE pos_payments IS 'Individual payment records for POS orders (supports split payments)';
COMMENT ON TABLE pos_order_discounts IS 'Track discounts and promotions applied to POS orders';
COMMENT ON TABLE pos_inventory_alerts IS 'Flag inventory issues during POS sales (negative stock, low stock)';
COMMENT ON TABLE pos_pricing_overrides IS 'Audit trail for manual price changes at POS with margin impact tracking';

COMMENT ON COLUMN sales_orders.pos_session_id IS 'Link to POS session if this is a POS order';
COMMENT ON COLUMN sales_orders.is_pos_order IS 'Flag to identify POS orders vs traditional sales orders';
COMMENT ON COLUMN sales_orders.pos_offline_uuid IS 'UUID for offline draft orders before sync';
COMMENT ON COLUMN sales_orders.pos_order_type IS 'Type of POS order: retail, restaurant, offer_response, service';
