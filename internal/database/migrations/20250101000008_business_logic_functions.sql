-- Migration: Core Business Logic Functions
-- Description: PostgreSQL functions for common business operations
-- Created: 2025-01-01

-- =====================================================
-- SALES ORDER FUNCTIONS
-- =====================================================

-- Compute sales order totals
CREATE OR REPLACE FUNCTION sales_order_compute_totals(p_order_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_result jsonb;
    v_amount_untaxed numeric(15,2);
    v_amount_tax numeric(15,2);
    v_amount_total numeric(15,2);
BEGIN
    -- Calculate totals from lines
    SELECT
        COALESCE(SUM(price_subtotal), 0),
        COALESCE(SUM(price_tax), 0),
        COALESCE(SUM(price_total), 0)
    INTO v_amount_untaxed, v_amount_tax, v_amount_total
    FROM sales_order_lines
    WHERE order_id = p_order_id
      AND deleted_at IS NULL;

    -- Update sales order
    UPDATE sales_orders
    SET
        amount_untaxed = v_amount_untaxed,
        amount_tax = v_amount_tax,
        amount_total = v_amount_total,
        updated_at = now()
    WHERE id = p_order_id;

    -- Return result
    v_result := jsonb_build_object(
        'amount_untaxed', v_amount_untaxed,
        'amount_tax', v_amount_tax,
        'amount_total', v_amount_total
    );

    RETURN v_result;
END;
$$;

-- Confirm sales order
CREATE OR REPLACE FUNCTION sales_order_confirm(p_order_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_order record;
    v_result jsonb;
BEGIN
    -- Get order
    SELECT * INTO v_order
    FROM sales_orders
    WHERE id = p_order_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Sales order not found: %', p_order_id;
    END IF;

    IF v_order.state != 'draft' AND v_order.state != 'sent' THEN
        RAISE EXCEPTION 'Cannot confirm sales order in state: %', v_order.state;
    END IF;

    -- Update state
    UPDATE sales_orders
    SET
        state = 'sale',
        confirmation_date = now(),
        updated_at = now()
    WHERE id = p_order_id;

    v_result := jsonb_build_object(
        'success', true,
        'order_id', p_order_id,
        'state', 'sale',
        'confirmation_date', now()
    );

    RETURN v_result;
END;
$$;

-- =====================================================
-- INVOICE FUNCTIONS
-- =====================================================

-- Compute invoice totals
CREATE OR REPLACE FUNCTION invoice_compute_totals(p_invoice_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_result jsonb;
    v_amount_untaxed numeric(15,2);
    v_amount_tax numeric(15,2);
    v_amount_total numeric(15,2);
BEGIN
    -- Calculate totals from lines
    SELECT
        COALESCE(SUM(price_subtotal), 0),
        COALESCE(SUM(CASE WHEN tax_line_id IS NOT NULL THEN debit - credit ELSE 0 END), 0),
        COALESCE(SUM(price_total), 0)
    INTO v_amount_untaxed, v_amount_tax, v_amount_total
    FROM invoice_lines
    WHERE move_id = p_invoice_id
      AND deleted_at IS NULL
      AND exclude_from_invoice_tab = false;

    -- Update invoice
    UPDATE invoices
    SET
        amount_untaxed = v_amount_untaxed,
        amount_tax = v_amount_tax,
        amount_total = v_amount_total,
        amount_residual = v_amount_total, -- Will be updated by payments
        updated_at = now()
    WHERE id = p_invoice_id;

    v_result := jsonb_build_object(
        'amount_untaxed', v_amount_untaxed,
        'amount_tax', v_amount_tax,
        'amount_total', v_amount_total
    );

    RETURN v_result;
END;
$$;

-- Post invoice (draft -> posted)
CREATE OR REPLACE FUNCTION invoice_post(p_invoice_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_invoice record;
    v_result jsonb;
    v_sequence_name text;
BEGIN
    -- Get invoice
    SELECT * INTO v_invoice
    FROM invoices
    WHERE id = p_invoice_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Invoice not found: %', p_invoice_id;
    END IF;

    IF v_invoice.state != 'draft' THEN
        RAISE EXCEPTION 'Cannot post invoice in state: %', v_invoice.state;
    END IF;

    -- Generate invoice number if not set
    IF v_invoice.name IS NULL THEN
        IF v_invoice.move_type = 'out_invoice' OR v_invoice.move_type = 'out_refund' THEN
            v_sequence_name := 'invoice';
        ELSE
            v_sequence_name := 'bill';
        END IF;

        UPDATE invoices
        SET name = generate_sequence_number(v_sequence_name, v_invoice.organization_id, v_invoice.company_id)
        WHERE id = p_invoice_id;
    END IF;

    -- Update state
    UPDATE invoices
    SET
        state = 'posted',
        updated_at = now()
    WHERE id = p_invoice_id;

    v_result := jsonb_build_object(
        'success', true,
        'invoice_id', p_invoice_id,
        'state', 'posted'
    );

    RETURN v_result;
END;
$$;

-- Register payment on invoice
CREATE OR REPLACE FUNCTION invoice_register_payment(
    p_invoice_id uuid,
    p_payment_amount numeric,
    p_payment_date date DEFAULT CURRENT_DATE,
    p_payment_method_id uuid DEFAULT NULL,
    p_reference text DEFAULT NULL
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
DECLARE
    v_invoice record;
    v_payment_id uuid;
    v_payment_type text;
    v_partner_type text;
BEGIN
    -- Get invoice
    SELECT * INTO v_invoice
    FROM invoices
    WHERE id = p_invoice_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Invoice not found: %', p_invoice_id;
    END IF;

    -- Determine payment type
    IF v_invoice.move_type IN ('out_invoice', 'out_refund') THEN
        v_payment_type := 'inbound';
        v_partner_type := 'customer';
    ELSE
        v_payment_type := 'outbound';
        v_partner_type := 'supplier';
    END IF;

    -- Create payment
    INSERT INTO payments (
        organization_id,
        company_id,
        payment_type,
        partner_type,
        partner_id,
        amount,
        currency_id,
        payment_date,
        journal_id,
        payment_method_id,
        ref,
        state
    ) VALUES (
        v_invoice.organization_id,
        v_invoice.company_id,
        v_payment_type,
        v_partner_type,
        v_invoice.partner_id,
        p_payment_amount,
        v_invoice.currency_id,
        p_payment_date,
        v_invoice.journal_id,
        p_payment_method_id,
        p_reference,
        'posted'
    )
    RETURNING id INTO v_payment_id;

    -- Allocate payment to invoice
    INSERT INTO payment_invoice_allocation (
        organization_id,
        payment_id,
        invoice_id,
        amount
    ) VALUES (
        v_invoice.organization_id,
        v_payment_id,
        p_invoice_id,
        p_payment_amount
    );

    -- Update invoice amount_residual and payment_state
    UPDATE invoices
    SET
        amount_residual = amount_residual - p_payment_amount,
        payment_state = CASE
            WHEN amount_residual - p_payment_amount <= 0 THEN 'paid'
            WHEN amount_residual - p_payment_amount < amount_total THEN 'partial'
            ELSE 'not_paid'
        END,
        updated_at = now()
    WHERE id = p_invoice_id;

    RETURN v_payment_id;
END;
$$;

-- =====================================================
-- INVENTORY FUNCTIONS
-- =====================================================

-- Update product stock
CREATE OR REPLACE FUNCTION product_update_stock(
    p_product_id uuid,
    p_location_id uuid,
    p_quantity_delta numeric,
    p_reference text DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_quant record;
    v_organization_id uuid;
    v_result jsonb;
BEGIN
    -- Get organization from product
    SELECT organization_id INTO v_organization_id
    FROM products
    WHERE id = p_product_id;

    -- Get or create quant
    SELECT * INTO v_quant
    FROM stock_quants
    WHERE product_id = p_product_id
      AND location_id = p_location_id
      AND lot_id IS NULL
      AND package_id IS NULL
      AND owner_id IS NULL
    FOR UPDATE;

    IF FOUND THEN
        -- Update existing quant
        UPDATE stock_quants
        SET
            quantity = quantity + p_quantity_delta,
            updated_at = now()
        WHERE id = v_quant.id;
    ELSE
        -- Create new quant
        INSERT INTO stock_quants (
            organization_id,
            product_id,
            location_id,
            quantity
        ) VALUES (
            v_organization_id,
            p_product_id,
            p_location_id,
            p_quantity_delta
        );
    END IF;

    v_result := jsonb_build_object(
        'success', true,
        'product_id', p_product_id,
        'location_id', p_location_id,
        'quantity_delta', p_quantity_delta
    );

    RETURN v_result;
END;
$$;

-- Get current stock for product
CREATE OR REPLACE FUNCTION product_get_stock(
    p_product_id uuid,
    p_location_id uuid DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_result jsonb;
    v_quantity_on_hand numeric;
    v_quantity_available numeric;
    v_quantity_reserved numeric;
BEGIN
    IF p_location_id IS NOT NULL THEN
        -- Get stock for specific location
        SELECT
            COALESCE(SUM(quantity), 0),
            COALESCE(SUM(quantity - reserved_quantity), 0),
            COALESCE(SUM(reserved_quantity), 0)
        INTO v_quantity_on_hand, v_quantity_available, v_quantity_reserved
        FROM stock_quants
        WHERE product_id = p_product_id
          AND location_id = p_location_id;
    ELSE
        -- Get stock across all internal locations
        SELECT
            COALESCE(SUM(q.quantity), 0),
            COALESCE(SUM(q.quantity - q.reserved_quantity), 0),
            COALESCE(SUM(q.reserved_quantity), 0)
        INTO v_quantity_on_hand, v_quantity_available, v_quantity_reserved
        FROM stock_quants q
        JOIN stock_locations l ON q.location_id = l.id
        WHERE q.product_id = p_product_id
          AND l.usage = 'internal';
    END IF;

    v_result := jsonb_build_object(
        'quantity_on_hand', v_quantity_on_hand,
        'quantity_available', v_quantity_available,
        'quantity_reserved', v_quantity_reserved
    );

    RETURN v_result;
END;
$$;

-- =====================================================
-- LEAD/OPPORTUNITY FUNCTIONS
-- =====================================================

-- Convert lead to opportunity
CREATE OR REPLACE FUNCTION lead_convert_to_opportunity(
    p_lead_id uuid,
    p_expected_revenue numeric DEFAULT NULL,
    p_probability integer DEFAULT NULL
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
DECLARE
    v_lead record;
BEGIN
    -- Get lead
    SELECT * INTO v_lead
    FROM leads
    WHERE id = p_lead_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Lead not found: %', p_lead_id;
    END IF;

    IF v_lead.lead_type != 'lead' THEN
        RAISE EXCEPTION 'Record is already an opportunity';
    END IF;

    -- Update to opportunity
    UPDATE leads
    SET
        lead_type = 'opportunity',
        expected_revenue = COALESCE(p_expected_revenue, expected_revenue),
        probability = COALESCE(p_probability, probability, 10),
        date_open = now(),
        updated_at = now()
    WHERE id = p_lead_id;

    RETURN p_lead_id;
END;
$$;

-- Mark opportunity as won
CREATE OR REPLACE FUNCTION opportunity_mark_won(
    p_opportunity_id uuid,
    p_create_sales_order boolean DEFAULT false
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_opportunity record;
    v_result jsonb;
    v_sales_order_id uuid;
BEGIN
    -- Get opportunity
    SELECT * INTO v_opportunity
    FROM leads
    WHERE id = p_opportunity_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Opportunity not found: %', p_opportunity_id;
    END IF;

    -- Update opportunity
    UPDATE leads
    SET
        won_status = 'won',
        probability = 100,
        date_closed = now(),
        active = false,
        updated_at = now()
    WHERE id = p_opportunity_id;

    v_result := jsonb_build_object(
        'success', true,
        'opportunity_id', p_opportunity_id,
        'won_status', 'won'
    );

    -- TODO: Create sales order if requested
    -- This would require additional implementation

    RETURN v_result;
END;
$$;

-- =====================================================
-- UTILITY FUNCTIONS
-- =====================================================

-- Archive/Unarchive record (soft delete toggle)
CREATE OR REPLACE FUNCTION toggle_archive(
    p_table_name text,
    p_record_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_query text;
    v_current_deleted_at timestamptz;
    v_new_deleted_at timestamptz;
BEGIN
    -- Validate table name (security)
    IF p_table_name !~ '^[a-z_]+$' THEN
        RAISE EXCEPTION 'Invalid table name: %', p_table_name;
    END IF;

    -- Get current deleted_at value
    v_query := format('SELECT deleted_at FROM %I WHERE id = %L', p_table_name, p_record_id);
    EXECUTE v_query INTO v_current_deleted_at;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Record not found in table %: %', p_table_name, p_record_id;
    END IF;

    -- Toggle deleted_at
    IF v_current_deleted_at IS NULL THEN
        v_new_deleted_at := now();
    ELSE
        v_new_deleted_at := NULL;
    END IF;

    -- Update record
    v_query := format('UPDATE %I SET deleted_at = %L, updated_at = now() WHERE id = %L',
                      p_table_name, v_new_deleted_at, p_record_id);
    EXECUTE v_query;

    RETURN v_new_deleted_at IS NOT NULL;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION sales_order_compute_totals IS 'Calculate and update sales order totals from lines';
COMMENT ON FUNCTION sales_order_confirm IS 'Confirm a sales order (draft/sent -> sale)';
COMMENT ON FUNCTION invoice_compute_totals IS 'Calculate and update invoice totals from lines';
COMMENT ON FUNCTION invoice_post IS 'Post an invoice (draft -> posted)';
COMMENT ON FUNCTION invoice_register_payment IS 'Create payment and allocate to invoice';
COMMENT ON FUNCTION product_update_stock IS 'Update product stock quantity at location';
COMMENT ON FUNCTION product_get_stock IS 'Get current stock levels for a product';
COMMENT ON FUNCTION lead_convert_to_opportunity IS 'Convert lead to opportunity';
COMMENT ON FUNCTION opportunity_mark_won IS 'Mark opportunity as won';
COMMENT ON FUNCTION toggle_archive IS 'Archive or unarchive a record (soft delete)';
