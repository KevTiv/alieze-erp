-- Migration: POS Business Logic Functions
-- Description: Core POS functions for session management, orders, payments, and pricing
-- Created: 2025-01-26
-- Dependencies: 20250101000049_pos_core_module.sql

-- =====================================================
-- SESSION MANAGEMENT FUNCTIONS
-- =====================================================

-- Open a new POS session
CREATE OR REPLACE FUNCTION pos_session_open(
    p_organization_id uuid,
    p_config_id uuid,
    p_user_id uuid,
    p_opening_balance numeric DEFAULT 0
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_session_id uuid;
    v_session_name varchar;
    v_config_record record;
BEGIN
    -- Validate config exists and is active
    SELECT * INTO v_config_record
    FROM public.pos_config
    WHERE id = p_config_id
      AND organization_id = p_organization_id
      AND active = true
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'POS configuration not found or inactive';
    END IF;

    -- Check for already open session for this config
    IF EXISTS (
        SELECT 1 FROM public.pos_sessions
        WHERE pos_config_id = p_config_id
          AND state IN ('opening_control', 'opened')
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'A session is already open for this POS configuration';
    END IF;

    -- Generate session name: POS/YYYY-MM-DD/NNN
    SELECT 'POS/' || to_char(now(), 'YYYY-MM-DD') || '/' ||
           lpad((COALESCE(COUNT(*) FILTER (WHERE DATE(created_at) = CURRENT_DATE), 0) + 1)::text, 3, '0')
    INTO v_session_name
    FROM public.pos_sessions
    WHERE organization_id = p_organization_id;

    -- Create session
    INSERT INTO public.pos_sessions (
        organization_id, company_id, name, pos_config_id, user_id,
        state, start_at, cash_register_balance_start,
        created_by
    ) VALUES (
        p_organization_id, v_config_record.company_id, v_session_name, p_config_id, p_user_id,
        CASE WHEN v_config_record.cash_control THEN 'opening_control' ELSE 'opened' END,
        now(), p_opening_balance, p_user_id
    ) RETURNING id INTO v_session_id;

    RETURN jsonb_build_object(
        'session_id', v_session_id,
        'session_name', v_session_name,
        'state', CASE WHEN v_config_record.cash_control THEN 'opening_control' ELSE 'opened' END,
        'config_id', p_config_id,
        'user_id', p_user_id,
        'opening_balance', p_opening_balance
    );
END;
$$;

-- Validate session opening (confirm counted cash)
CREATE OR REPLACE FUNCTION pos_session_validate_opening(
    p_session_id uuid,
    p_counted_cash numeric
) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
    UPDATE public.pos_sessions
    SET state = 'opened',
        cash_register_balance_start = p_counted_cash
    WHERE id = p_session_id
      AND state = 'opening_control'
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Session not found or not in opening_control state';
    END IF;
END;
$$;

-- Record cash in/out movement
CREATE OR REPLACE FUNCTION pos_cash_movement_record(
    p_organization_id uuid,
    p_session_id uuid,
    p_type varchar,
    p_amount numeric,
    p_name varchar,
    p_reason text DEFAULT NULL,
    p_authorized_by uuid DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_movement_id uuid;
    v_session_state varchar;
BEGIN
    -- Validate session is open
    SELECT state INTO v_session_state
    FROM public.pos_sessions
    WHERE id = p_session_id
      AND organization_id = p_organization_id
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Session not found';
    END IF;

    IF v_session_state NOT IN ('opened', 'closing_control') THEN
        RAISE EXCEPTION 'Session is not open';
    END IF;

    -- Record movement
    INSERT INTO public.pos_cash_movements (
        organization_id, session_id, type, amount, name, reason, authorized_by, created_by
    ) VALUES (
        p_organization_id, p_session_id, p_type, p_amount, p_name, p_reason, p_authorized_by, p_authorized_by
    ) RETURNING id INTO v_movement_id;

    -- Update session cash balance
    UPDATE public.pos_sessions
    SET cash_register_balance_start = cash_register_balance_start +
        CASE WHEN p_type = 'in' THEN p_amount ELSE -p_amount END
    WHERE id = p_session_id;

    RETURN v_movement_id;
END;
$$;

-- Close POS session with reconciliation
CREATE OR REPLACE FUNCTION pos_session_close(
    p_session_id uuid,
    p_counted_cash numeric,
    p_closing_notes text DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_session record;
    v_config record;
    v_expected_cash numeric;
    v_cash_in numeric := 0;
    v_cash_out numeric := 0;
    v_difference numeric;
    v_report jsonb;
    v_payment_breakdown jsonb;
BEGIN
    -- Get session with config
    SELECT
        s.*,
        c.cash_control,
        c.set_maximum_difference,
        c.maximum_difference
    INTO v_session
    FROM public.pos_sessions s
    JOIN public.pos_config c ON c.id = s.pos_config_id
    WHERE s.id = p_session_id
      AND s.deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Session not found';
    END IF;

    IF v_session.state NOT IN ('opened', 'closing_control') THEN
        RAISE EXCEPTION 'Session cannot be closed from state: %', v_session.state;
    END IF;

    -- Calculate cash movements
    SELECT
        COALESCE(SUM(amount) FILTER (WHERE type = 'in'), 0),
        COALESCE(SUM(amount) FILTER (WHERE type = 'out'), 0)
    INTO v_cash_in, v_cash_out
    FROM public.pos_cash_movements
    WHERE session_id = p_session_id;

    -- Calculate expected cash from payments
    WITH payment_totals AS (
        SELECT
            pm.type,
            pm.name,
            COALESCE(SUM(p.amount), 0) as total_amount,
            COUNT(*) as transaction_count
        FROM public.pos_payments p
        JOIN public.pos_payment_methods pm ON pm.id = p.payment_method_id
        WHERE p.pos_session_id = p_session_id
        GROUP BY pm.type, pm.name, pm.id
        ORDER BY pm.type
    )
    SELECT
        jsonb_object_agg(name, jsonb_build_object(
            'type', type,
            'amount', total_amount,
            'count', transaction_count
        ))
    INTO v_payment_breakdown
    FROM payment_totals;

    -- Calculate expected cash
    SELECT COALESCE(SUM(p.amount), 0)
    INTO v_expected_cash
    FROM public.pos_payments p
    JOIN public.pos_payment_methods pm ON pm.id = p.payment_method_id
    WHERE p.pos_session_id = p_session_id
      AND pm.type = 'cash';

    v_expected_cash := v_session.cash_register_balance_start + v_expected_cash + v_cash_in - v_cash_out;
    v_difference := p_counted_cash - v_expected_cash;

    -- Check maximum difference threshold
    IF v_session.set_maximum_difference
       AND ABS(v_difference) > v_session.maximum_difference THEN
        RAISE EXCEPTION 'Cash difference (%) exceeds maximum allowed (%)',
            v_difference, v_session.maximum_difference;
    END IF;

    -- Update session summary
    UPDATE public.pos_sessions
    SET state = 'closed',
        stop_at = now(),
        cash_register_balance_end = v_expected_cash,
        cash_register_balance_end_real = p_counted_cash,
        cash_register_difference = v_difference,
        closing_notes = p_closing_notes,
        reconciliation_data = jsonb_build_object(
            'payment_breakdown', v_payment_breakdown,
            'cash_movements', jsonb_build_object(
                'in', v_cash_in,
                'out', v_cash_out,
                'net', v_cash_in - v_cash_out
            ),
            'expected_cash', v_expected_cash,
            'counted_cash', p_counted_cash,
            'difference', v_difference
        )
    WHERE id = p_session_id;

    -- Enqueue accounting journal entry creation
    PERFORM public.enqueue_job(
        p_organization_id := v_session.organization_id,
        p_job_type := 'pos_session_accounting',
        p_payload := jsonb_build_object('session_id', p_session_id),
        p_priority := 5
    );

    -- Build report
    v_report := jsonb_build_object(
        'session_id', p_session_id,
        'session_name', v_session.name,
        'opened_at', v_session.start_at,
        'closed_at', now(),
        'cashier_id', v_session.user_id,
        'opening_balance', v_session.cash_register_balance_start,
        'expected_cash', v_expected_cash,
        'counted_cash', p_counted_cash,
        'difference', v_difference,
        'total_orders', v_session.total_orders_count,
        'total_sales', v_session.total_amount,
        'payment_breakdown', v_payment_breakdown,
        'cash_movements', jsonb_build_object(
            'in', v_cash_in,
            'out', v_cash_out
        )
    );

    RETURN v_report;
END;
$$;

-- =====================================================
-- ORDER MANAGEMENT FUNCTIONS
-- =====================================================

-- Create POS order (draft mode for offline support)
CREATE OR REPLACE FUNCTION pos_order_create(
    p_organization_id uuid,
    p_session_id uuid,
    p_customer_id uuid DEFAULT NULL,
    p_lines jsonb DEFAULT '[]'::jsonb,
    p_pricelist_id uuid DEFAULT NULL,
    p_offline_uuid uuid DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_order_id uuid;
    v_session record;
    v_order_ref varchar;
    v_line jsonb;
    v_line_id uuid;
    v_product record;
    v_subtotal numeric := 0;
    v_tax_total numeric := 0;
    v_total numeric := 0;
    v_default_customer_id uuid;
BEGIN
    -- Validate session
    SELECT s.*, c.pricelist_id as config_pricelist_id, c.require_customer, c.company_id
    INTO v_session
    FROM public.pos_sessions s
    JOIN public.pos_config c ON c.id = s.pos_config_id
    WHERE s.id = p_session_id
      AND s.organization_id = p_organization_id
      AND s.state = 'opened'
      AND s.deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Session not found or not open';
    END IF;

    -- Handle customer requirement
    IF p_customer_id IS NULL THEN
        IF v_session.require_customer THEN
            RAISE EXCEPTION 'Customer is required for this POS configuration';
        ELSE
            -- Get or create default walk-in customer
            SELECT id INTO v_default_customer_id
            FROM public.contacts
            WHERE organization_id = p_organization_id
              AND name = 'Walk-in Customer'
              AND is_customer = true
            LIMIT 1;

            IF v_default_customer_id IS NULL THEN
                INSERT INTO public.contacts (
                    organization_id, company_id, name, is_customer, is_company
                ) VALUES (
                    p_organization_id, v_session.company_id, 'Walk-in Customer', true, false
                ) RETURNING id INTO v_default_customer_id;
            END IF;
        END IF;
    END IF;

    -- Generate order reference
    v_order_ref := 'POS/' || v_session.name || '/' ||
                   lpad((v_session.total_orders_count + 1)::text, 4, '0');

    -- Create sales order
    INSERT INTO public.sales_orders (
        organization_id, company_id, name, pos_order_ref,
        date_order, partner_id, pricelist_id,
        pos_session_id, is_pos_order, pos_offline_uuid,
        state, invoice_status, delivery_status,
        created_by
    ) VALUES (
        p_organization_id, v_session.company_id, v_order_ref, v_order_ref,
        now(), COALESCE(p_customer_id, v_default_customer_id),
        COALESCE(p_pricelist_id, v_session.config_pricelist_id),
        p_session_id, true, p_offline_uuid,
        'draft', 'no', 'no',
        v_session.user_id
    ) RETURNING id INTO v_order_id;

    -- Add order lines if provided
    IF jsonb_array_length(p_lines) > 0 THEN
        FOR v_line IN SELECT * FROM jsonb_array_elements(p_lines)
        LOOP
            PERFORM public.pos_order_add_line(
                p_order_id := v_order_id,
                p_product_id := (v_line->>'product_id')::uuid,
                p_quantity := (v_line->>'qty')::numeric,
                p_price_unit := (v_line->>'price_unit')::numeric,
                p_discount := COALESCE((v_line->>'discount')::numeric, 0)
            );
        END LOOP;

        -- Recalculate totals
        PERFORM public.pos_order_calculate_totals(v_order_id);
    END IF;

    RETURN jsonb_build_object(
        'order_id', v_order_id,
        'order_ref', v_order_ref,
        'session_id', p_session_id,
        'customer_id', COALESCE(p_customer_id, v_default_customer_id),
        'state', 'draft'
    );
END;
$$;

-- Add line to POS order with stock check
CREATE OR REPLACE FUNCTION pos_order_add_line(
    p_order_id uuid,
    p_product_id uuid,
    p_quantity numeric,
    p_price_unit numeric DEFAULT NULL,
    p_discount numeric DEFAULT 0
) RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_line_id uuid;
    v_order record;
    v_product record;
    v_config record;
    v_stock_qty numeric;
    v_price numeric;
    v_cost numeric;
    v_margin numeric;
    v_margin_pct numeric;
BEGIN
    -- Get order and config
    SELECT o.*, s.pos_config_id, o.organization_id
    INTO v_order
    FROM public.sales_orders o
    JOIN public.pos_sessions s ON s.id = o.pos_session_id
    WHERE o.id = p_order_id
      AND o.is_pos_order = true
      AND o.state = 'draft';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'POS order not found or not in draft state';
    END IF;

    -- Get config
    SELECT * INTO v_config
    FROM public.pos_config
    WHERE id = v_order.pos_config_id;

    -- Get product details
    SELECT * INTO v_product
    FROM public.products
    WHERE id = p_product_id
      AND organization_id = v_order.organization_id
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Product not found';
    END IF;

    -- Check stock availability
    SELECT COALESCE(SUM(quantity), 0) INTO v_stock_qty
    FROM public.stock_quant
    WHERE product_id = p_product_id
      AND location_id IN (
          SELECT id FROM public.stock_locations
          WHERE usage = 'internal'
            AND (id = v_config.stock_location_id OR v_config.stock_location_id IS NULL)
      );

    -- Handle stock validation
    IF v_stock_qty < p_quantity THEN
        IF NOT v_config.allow_negative_stock AND NOT v_config.flag_negative_stock THEN
            RAISE EXCEPTION 'Insufficient stock for product: % (available: %, requested: %)',
                v_product.name, v_stock_qty, p_quantity;
        END IF;

        -- Flag for review if enabled
        IF v_config.flag_negative_stock THEN
            INSERT INTO public.pos_inventory_alerts (
                organization_id, order_id, product_id,
                alert_type, quantity_sold, quantity_available,
                warehouse_id, location_id, created_by
            ) VALUES (
                v_order.organization_id, p_order_id, p_product_id,
                CASE
                    WHEN v_stock_qty <= 0 THEN 'out_of_stock'
                    WHEN v_stock_qty < p_quantity THEN 'negative_stock'
                    ELSE 'low_stock'
                END,
                p_quantity, v_stock_qty,
                v_config.warehouse_id, v_config.stock_location_id, v_order.created_by
            );
        END IF;
    END IF;

    -- Determine price
    v_price := COALESCE(p_price_unit, v_product.list_price);
    v_cost := COALESCE(v_product.standard_price, 0);

    -- Calculate margin
    IF v_price > 0 THEN
        v_margin := (v_price - v_cost) * p_quantity;
        v_margin_pct := ((v_price - v_cost) / v_price) * 100;

        -- Warn if below minimum margin
        IF v_config.warn_below_margin_threshold
           AND v_margin_pct < v_config.minimum_margin_threshold THEN
            RAISE NOTICE 'Margin below threshold: % (threshold: %)',
                v_margin_pct, v_config.minimum_margin_threshold;
        END IF;
    END IF;

    -- Create order line
    INSERT INTO public.sales_order_lines (
        organization_id, order_id, sequence,
        name, product_id, product_uom_qty, product_uom,
        price_unit, discount, state,
        custom_fields, created_by
    ) VALUES (
        v_order.organization_id, p_order_id,
        (SELECT COALESCE(MAX(sequence), 0) + 10 FROM public.sales_order_lines WHERE order_id = p_order_id),
        v_product.name,
        v_product.id,
        p_quantity,
        v_product.uom_id,
        v_price,
        p_discount,
        'draft',
        jsonb_build_object(
            'cost_price', v_cost,
            'margin', v_margin,
            'margin_pct', v_margin_pct,
            'stock_available', v_stock_qty
        ),
        v_order.created_by
    ) RETURNING id INTO v_line_id;

    RETURN v_line_id;
END;
$$;

-- Calculate order totals
CREATE OR REPLACE FUNCTION pos_order_calculate_totals(
    p_order_id uuid
) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_subtotal numeric := 0;
    v_tax_total numeric := 0;
    v_discount_total numeric := 0;
BEGIN
    -- Calculate line totals
    UPDATE public.sales_order_lines
    SET price_subtotal = product_uom_qty * price_unit * (1 - discount / 100),
        price_total = product_uom_qty * price_unit * (1 - discount / 100)
    WHERE order_id = p_order_id;

    -- Sum order totals
    SELECT
        COALESCE(SUM(price_subtotal), 0),
        COALESCE(SUM(price_tax), 0)
    INTO v_subtotal, v_tax_total
    FROM public.sales_order_lines
    WHERE order_id = p_order_id;

    -- Get discount total
    SELECT COALESCE(SUM(discount_amount), 0)
    INTO v_discount_total
    FROM public.pos_order_discounts
    WHERE order_id = p_order_id;

    -- Update order
    UPDATE public.sales_orders
    SET amount_untaxed = v_subtotal - v_discount_total,
        amount_tax = v_tax_total,
        amount_total = v_subtotal - v_discount_total + v_tax_total,
        amount_discount = v_discount_total
    WHERE id = p_order_id;
END;
$$;

-- Validate POS order (complete transaction)
CREATE OR REPLACE FUNCTION pos_order_validate(
    p_order_id uuid,
    p_payments jsonb DEFAULT '[]'::jsonb
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_order record;
    v_config record;
    v_payment jsonb;
    v_total_paid numeric := 0;
    v_change numeric := 0;
    v_payment_id uuid;
BEGIN
    -- Get order with session and config
    SELECT o.*, s.pos_config_id, s.user_id as cashier_id
    INTO v_order
    FROM public.sales_orders o
    JOIN public.pos_sessions s ON s.id = o.pos_session_id
    WHERE o.id = p_order_id
      AND o.is_pos_order = true
      AND o.state = 'draft';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'POS order not found or already validated';
    END IF;

    -- Recalculate totals
    PERFORM public.pos_order_calculate_totals(p_order_id);

    -- Refresh order totals
    SELECT amount_total INTO v_order.amount_total
    FROM public.sales_orders WHERE id = p_order_id;

    -- Process payments
    FOR v_payment IN SELECT * FROM jsonb_array_elements(p_payments)
    LOOP
        INSERT INTO public.pos_payments (
            organization_id, pos_session_id, order_id,
            payment_method_id, amount, payment_date,
            transaction_id, card_last_four, card_type,
            authorization_code, created_by
        ) VALUES (
            v_order.organization_id,
            v_order.pos_session_id,
            p_order_id,
            (v_payment->>'payment_method_id')::uuid,
            (v_payment->>'amount')::numeric,
            now(),
            v_payment->>'transaction_id',
            v_payment->>'card_last_four',
            v_payment->>'card_type',
            v_payment->>'authorization_code',
            v_order.cashier_id
        ) RETURNING id INTO v_payment_id;

        v_total_paid := v_total_paid + (v_payment->>'amount')::numeric;
    END LOOP;

    -- Calculate change
    v_change := v_total_paid - v_order.amount_total;

    -- Verify sufficient payment
    IF v_total_paid < v_order.amount_total THEN
        RAISE EXCEPTION 'Insufficient payment: paid %, required %', v_total_paid, v_order.amount_total;
    END IF;

    -- Record change as negative payment if needed
    IF v_change > 0 THEN
        -- Find cash payment method for change
        INSERT INTO public.pos_payments (
            organization_id, pos_session_id, order_id,
            payment_method_id, amount, payment_date,
            is_change, created_by
        )
        SELECT
            v_order.organization_id,
            v_order.pos_session_id,
            p_order_id,
            pm.id,
            -v_change,
            now(),
            true,
            v_order.cashier_id
        FROM public.pos_payment_methods pm
        WHERE pm.organization_id = v_order.organization_id
          AND pm.type = 'cash'
        LIMIT 1;
    END IF;

    -- Update order state
    UPDATE public.sales_orders
    SET state = 'sale',
        confirmation_date = now(),
        pos_validated_at = now(),
        invoice_status = 'invoiced',
        delivery_status = 'delivered'
    WHERE id = p_order_id;

    -- Update session totals
    UPDATE public.pos_sessions
    SET total_orders_count = total_orders_count + 1,
        total_amount = total_amount + v_order.amount_total
    WHERE id = v_order.pos_session_id;

    -- Enqueue stock moves creation
    PERFORM public.enqueue_job(
        p_organization_id := v_order.organization_id,
        p_job_type := 'pos_stock_move_create',
        p_payload := jsonb_build_object('order_id', p_order_id),
        p_priority := 10
    );

    -- Enqueue receipt generation
    PERFORM public.enqueue_job(
        p_organization_id := v_order.organization_id,
        p_job_type := 'pos_receipt_generation',
        p_payload := jsonb_build_object('order_id', p_order_id),
        p_priority := 10
    );

    RETURN jsonb_build_object(
        'order_id', p_order_id,
        'order_ref', v_order.pos_order_ref,
        'total', v_order.amount_total,
        'paid', v_total_paid,
        'change', GREATEST(v_change, 0),
        'validated_at', now(),
        'receipt_queued', true
    );
END;
$$;

-- =====================================================
-- PRICING & DISCOUNT FUNCTIONS
-- =====================================================

-- Apply discount to order
CREATE OR REPLACE FUNCTION pos_discount_apply(
    p_order_id uuid,
    p_discount_type varchar,
    p_discount_value numeric,
    p_reason varchar DEFAULT NULL,
    p_authorized_by uuid DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_discount_id uuid;
    v_order record;
    v_config record;
    v_discount_amount numeric;
BEGIN
    -- Get order and config
    SELECT o.*, s.pos_config_id, o.amount_untaxed
    INTO v_order
    FROM public.sales_orders o
    JOIN public.pos_sessions s ON s.id = o.pos_session_id
    WHERE o.id = p_order_id
      AND o.is_pos_order = true
      AND o.state = 'draft';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Order not found or not in draft state';
    END IF;

    -- Get config for discount limits
    SELECT * INTO v_config
    FROM public.pos_config
    WHERE id = v_order.pos_config_id;

    IF NOT v_config.manual_discount THEN
        RAISE EXCEPTION 'Manual discounts are not allowed for this POS configuration';
    END IF;

    -- Calculate discount amount
    IF p_discount_type = 'percentage' THEN
        v_discount_amount := v_order.amount_untaxed * (p_discount_value / 100);

        -- Check discount limits
        IF p_discount_value > v_config.discount_limit THEN
            IF p_authorized_by IS NULL THEN
                RAISE EXCEPTION 'Discount of % requires manager authorization (limit: %)',
                    p_discount_value, v_config.discount_limit;
            ELSIF v_config.manager_discount_limit IS NOT NULL
                  AND p_discount_value > v_config.manager_discount_limit THEN
                RAISE EXCEPTION 'Discount of % exceeds maximum allowed (%)',
                    p_discount_value, v_config.manager_discount_limit;
            END IF;
        END IF;
    ELSE
        v_discount_amount := p_discount_value;
    END IF;

    -- Create discount record
    INSERT INTO public.pos_order_discounts (
        organization_id, order_id, discount_type, discount_value,
        discount_amount, reason, authorized_by, created_by
    ) VALUES (
        v_order.organization_id, p_order_id, p_discount_type, p_discount_value,
        v_discount_amount, p_reason, p_authorized_by, v_order.created_by
    ) RETURNING id INTO v_discount_id;

    -- Recalculate order totals
    PERFORM public.pos_order_calculate_totals(p_order_id);

    RETURN v_discount_id;
END;
$$;

-- Override line price with margin tracking
CREATE OR REPLACE FUNCTION pos_price_override(
    p_order_line_id uuid,
    p_new_price numeric,
    p_reason varchar DEFAULT NULL,
    p_authorized_by uuid DEFAULT NULL
) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_line record;
    v_product record;
    v_original_margin numeric;
    v_new_margin numeric;
BEGIN
    -- Get line and product
    SELECT
        l.*,
        p.standard_price as cost_price
    INTO v_line
    FROM public.sales_order_lines l
    JOIN public.products p ON p.id = l.product_id
    WHERE l.id = p_order_line_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Order line not found';
    END IF;

    -- Calculate margins
    v_original_margin := (v_line.price_unit - v_line.cost_price) * v_line.product_uom_qty;
    v_new_margin := (p_new_price - v_line.cost_price) * v_line.product_uom_qty;

    -- Record override
    INSERT INTO public.pos_pricing_overrides (
        organization_id, order_id, order_line_id, product_id,
        original_price, override_price, reason,
        original_margin, new_margin, margin_loss,
        authorized_by, created_by
    ) VALUES (
        v_line.organization_id, v_line.order_id, p_order_line_id, v_line.product_id,
        v_line.price_unit, p_new_price, p_reason,
        v_original_margin, v_new_margin, v_new_margin - v_original_margin,
        p_authorized_by, p_authorized_by
    );

    -- Update line price
    UPDATE public.sales_order_lines
    SET price_unit = p_new_price
    WHERE id = p_order_line_id;

    -- Recalculate totals
    PERFORM public.pos_order_calculate_totals(v_line.order_id);
END;
$$;

-- =====================================================
-- AI-POWERED PRODUCT SEARCH FOR POS
-- =====================================================

-- Search products with semantic search (leverages existing vector search)
CREATE OR REPLACE FUNCTION pos_product_search(
    p_organization_id uuid,
    p_query_embedding vector(768),
    p_config_id uuid DEFAULT NULL,
    p_limit integer DEFAULT 20
) RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    barcode varchar,
    default_code varchar,
    list_price numeric,
    standard_price numeric,
    stock_qty numeric,
    margin_pct numeric,
    category_id uuid,
    similarity_score float
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_category_ids uuid[];
    v_location_id uuid;
BEGIN
    -- Get category restrictions if config provided
    IF p_config_id IS NOT NULL THEN
        SELECT
            CASE WHEN limit_categories THEN iface_available_categ_ids ELSE NULL END,
            stock_location_id
        INTO v_category_ids, v_location_id
        FROM public.pos_config
        WHERE id = p_config_id;
    END IF;

    RETURN QUERY
    SELECT
        p.id,
        p.name,
        p.barcode,
        p.default_code,
        p.list_price,
        p.standard_price,
        COALESCE(sq.qty, 0) as stock_qty,
        CASE
            WHEN p.list_price > 0 THEN ((p.list_price - p.standard_price) / p.list_price) * 100
            ELSE 0
        END as margin_pct,
        p.category_id,
        1 - (p.search_embedding <=> p_query_embedding) as similarity_score
    FROM public.products p
    LEFT JOIN LATERAL (
        SELECT SUM(quantity) as qty
        FROM public.stock_quant
        WHERE product_id = p.id
          AND (v_location_id IS NULL OR location_id = v_location_id)
    ) sq ON true
    WHERE p.organization_id = p_organization_id
      AND p.deleted_at IS NULL
      AND p.active = true
      AND p.sale_ok = true
      AND p.search_embedding IS NOT NULL
      AND (v_category_ids IS NULL OR p.category_id = ANY(v_category_ids))
    ORDER BY similarity_score DESC
    LIMIT p_limit;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION pos_session_open IS 'Open a new POS session with cash control';
COMMENT ON FUNCTION pos_session_close IS 'Close POS session with reconciliation and accounting batch';
COMMENT ON FUNCTION pos_order_create IS 'Create draft POS order (supports offline mode)';
COMMENT ON FUNCTION pos_order_add_line IS 'Add product line with stock validation and margin calculation';
COMMENT ON FUNCTION pos_order_validate IS 'Complete POS transaction with payment validation and stock/accounting queue';
COMMENT ON FUNCTION pos_discount_apply IS 'Apply discount with authorization checks';
COMMENT ON FUNCTION pos_price_override IS 'Override line price with margin impact tracking';
COMMENT ON FUNCTION pos_product_search IS 'AI-powered semantic product search for POS';
