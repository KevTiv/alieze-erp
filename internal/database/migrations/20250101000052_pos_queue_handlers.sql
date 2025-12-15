-- Migration: POS Queue Handlers
-- Description: Queue job handlers for POS offline sync, receipts, stock moves, and accounting
-- Created: 2025-01-26
-- Dependencies: 20250101000051_pos_analytics_views.sql, queue_system

-- =====================================================
-- OFFLINE SYNC HANDLER
-- =====================================================

CREATE OR REPLACE FUNCTION handle_pos_offline_sync_job(
    p_job_id uuid,
    p_payload jsonb
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_organization_id uuid;
    v_offline_orders jsonb;
    v_order jsonb;
    v_order_id uuid;
    v_synced_count integer := 0;
    v_failed_count integer := 0;
    v_errors jsonb := '[]'::jsonb;
BEGIN
    v_organization_id := (p_payload->>'organization_id')::uuid;
    v_offline_orders := p_payload->'orders';

    -- Process each offline order
    FOR v_order IN SELECT * FROM jsonb_array_elements(v_offline_orders)
    LOOP
        BEGIN
            -- Check if already synced
            IF EXISTS (
                SELECT 1 FROM public.sales_orders
                WHERE pos_offline_uuid = (v_order->>'offline_uuid')::uuid
                  AND organization_id = v_organization_id
            ) THEN
                CONTINUE;  -- Skip already synced orders
            END IF;

            -- Create order
            SELECT (public.pos_order_create(
                p_organization_id := v_organization_id,
                p_session_id := (v_order->>'session_id')::uuid,
                p_customer_id := (v_order->>'customer_id')::uuid,
                p_lines := v_order->'lines',
                p_pricelist_id := (v_order->>'pricelist_id')::uuid,
                p_offline_uuid := (v_order->>'offline_uuid')::uuid
            )->>'order_id')::uuid INTO v_order_id;

            -- Validate if payments provided
            IF v_order ? 'payments' THEN
                PERFORM public.pos_order_validate(
                    p_order_id := v_order_id,
                    p_payments := v_order->'payments'
                );
            END IF;

            -- Mark as synced
            UPDATE public.sales_orders
            SET pos_synced_at = now(),
                pos_draft_data = v_order
            WHERE id = v_order_id;

            v_synced_count := v_synced_count + 1;

        EXCEPTION WHEN OTHERS THEN
            v_failed_count := v_failed_count + 1;
            v_errors := v_errors || jsonb_build_object(
                'offline_uuid', v_order->>'offline_uuid',
                'error', SQLERRM
            );
        END;
    END LOOP;

    RETURN jsonb_build_object(
        'synced_count', v_synced_count,
        'failed_count', v_failed_count,
        'errors', v_errors
    );
END;
$$;

-- =====================================================
-- RECEIPT GENERATION HANDLER
-- =====================================================

CREATE OR REPLACE FUNCTION handle_pos_receipt_generation_job(
    p_job_id uuid,
    p_payload jsonb
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_order_id uuid;
    v_order record;
    v_config record;
    v_lines jsonb;
    v_payments jsonb;
    v_receipt_data jsonb;
BEGIN
    v_order_id := (p_payload->>'order_id')::uuid;

    -- Get order details
    SELECT
        so.*,
        c.name as customer_name,
        c.email as customer_email
    INTO v_order
    FROM public.sales_orders so
    JOIN public.contacts c ON c.id = so.partner_id
    WHERE so.id = v_order_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Order not found';
    END IF;

    -- Get config for receipt template
    SELECT pc.*
    INTO v_config
    FROM public.pos_config pc
    JOIN public.pos_sessions ps ON ps.pos_config_id = pc.id
    WHERE ps.id = v_order.pos_session_id;

    -- Get order lines
    SELECT jsonb_agg(
        jsonb_build_object(
            'name', sol.name,
            'quantity', sol.product_uom_qty,
            'price_unit', sol.price_unit,
            'discount', sol.discount,
            'subtotal', sol.price_subtotal
        ) ORDER BY sol.sequence
    ) INTO v_lines
    FROM public.sales_order_lines sol
    WHERE sol.order_id = v_order_id;

    -- Get payments
    SELECT jsonb_agg(
        jsonb_build_object(
            'method', pm.name,
            'amount', pp.amount,
            'card_last_four', pp.card_last_four
        )
    ) INTO v_payments
    FROM public.pos_payments pp
    JOIN public.pos_payment_methods pm ON pm.id = pp.payment_method_id
    WHERE pp.order_id = v_order_id
      AND NOT pp.is_change;

    -- Build receipt data
    v_receipt_data := jsonb_build_object(
        'order_ref', v_order.pos_order_ref,
        'date', v_order.date_order,
        'customer_name', v_order.customer_name,
        'customer_email', v_order.customer_email,
        'lines', v_lines,
        'subtotal', v_order.amount_untaxed,
        'tax', v_order.amount_tax,
        'total', v_order.amount_total,
        'payments', v_payments,
        'header', v_config.receipt_header,
        'footer', v_config.receipt_footer,
        'template_config', v_config.receipt_template_config
    );

    -- Email receipt if enabled and customer has email
    IF v_config.email_receipt_enabled AND v_order.customer_email IS NOT NULL THEN
        -- Enqueue email job (assuming you have an email queue handler)
        PERFORM public.enqueue_job(
            p_organization_id := v_order.organization_id,
            p_job_type := 'send_email',
            p_payload := jsonb_build_object(
                'to', v_order.customer_email,
                'subject', 'Receipt: ' || v_order.pos_order_ref,
                'template', 'pos_receipt',
                'data', v_receipt_data
            ),
            p_priority := 5
        );
    END IF;

    RETURN v_receipt_data;
END;
$$;

-- =====================================================
-- STOCK MOVE CREATION HANDLER
-- =====================================================

CREATE OR REPLACE FUNCTION handle_pos_stock_move_create_job(
    p_job_id uuid,
    p_payload jsonb
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_order_id uuid;
    v_order record;
    v_config record;
    v_line record;
    v_picking_id uuid;
    v_move_ids uuid[] := '{}';
    v_move_id uuid;
BEGIN
    v_order_id := (p_payload->>'order_id')::uuid;

    -- Get order and config
    SELECT
        so.*,
        s.pos_config_id
    INTO v_order
    FROM public.sales_orders so
    JOIN public.pos_sessions s ON s.id = so.pos_session_id
    WHERE so.id = v_order_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Order not found';
    END IF;

    -- Get config for location details
    SELECT * INTO v_config
    FROM public.pos_config
    WHERE id = v_order.pos_config_id;

    -- Create stock picking
    INSERT INTO public.stock_pickings (
        organization_id, company_id, name, origin,
        partner_id, picking_type_id, location_id, location_dest_id,
        scheduled_date, state,
        created_by
    )
    SELECT
        v_order.organization_id,
        v_order.company_id,
        'POS/' || v_order.pos_order_ref,
        v_order.pos_order_ref,
        v_order.partner_id,
        (SELECT id FROM public.stock_picking_types
         WHERE code = 'outgoing' AND organization_id = v_order.organization_id LIMIT 1),
        v_config.stock_location_id,
        (SELECT id FROM public.stock_locations
         WHERE usage = 'customer' AND organization_id = v_order.organization_id LIMIT 1),
        v_order.date_order,
        'assigned',
        v_order.created_by
    RETURNING id INTO v_picking_id;

    -- Create stock moves for each line
    FOR v_line IN
        SELECT * FROM public.sales_order_lines
        WHERE order_id = v_order_id
          AND product_id IS NOT NULL
    LOOP
        INSERT INTO public.stock_moves (
            organization_id, name, product_id, product_uom_qty, product_uom,
            picking_id, location_id, location_dest_id,
            state, origin,
            created_by
        ) VALUES (
            v_order.organization_id,
            v_line.name,
            v_line.product_id,
            v_line.product_uom_qty,
            v_line.product_uom,
            v_picking_id,
            v_config.stock_location_id,
            (SELECT id FROM public.stock_locations
             WHERE usage = 'customer' AND organization_id = v_order.organization_id LIMIT 1),
            'assigned',
            v_order.pos_order_ref,
            v_order.created_by
        ) RETURNING id INTO v_move_id;

        v_move_ids := array_append(v_move_ids, v_move_id);
    END LOOP;

    -- Validate picking immediately (POS orders are instant delivery)
    UPDATE public.stock_pickings
    SET state = 'done',
        date_done = now()
    WHERE id = v_picking_id;

    -- Validate moves
    UPDATE public.stock_moves
    SET state = 'done',
        date = now()
    WHERE id = ANY(v_move_ids);

    -- Update stock quantities
    FOR v_line IN
        SELECT * FROM public.sales_order_lines
        WHERE order_id = v_order_id
          AND product_id IS NOT NULL
    LOOP
        -- Decrease stock
        UPDATE public.stock_quant
        SET quantity = quantity - v_line.product_uom_qty
        WHERE product_id = v_line.product_id
          AND location_id = v_config.stock_location_id;

        -- Create stock quant if doesn't exist (for negative stock scenario)
        IF NOT FOUND THEN
            INSERT INTO public.stock_quant (
                organization_id, product_id, location_id, quantity
            ) VALUES (
                v_order.organization_id,
                v_line.product_id,
                v_config.stock_location_id,
                -v_line.product_uom_qty
            );
        END IF;
    END LOOP;

    RETURN jsonb_build_object(
        'picking_id', v_picking_id,
        'move_ids', v_move_ids,
        'state', 'done'
    );
END;
$$;

-- =====================================================
-- SESSION ACCOUNTING HANDLER
-- =====================================================

CREATE OR REPLACE FUNCTION handle_pos_session_accounting_job(
    p_job_id uuid,
    p_payload jsonb
) RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_session_id uuid;
    v_session record;
    v_config record;
    v_move_id uuid;
    v_line_id uuid;
    v_payment_method record;
BEGIN
    v_session_id := (p_payload->>'session_id')::uuid;

    -- Get session and config
    SELECT
        s.*,
        c.journal_id
    INTO v_session
    FROM public.pos_sessions s
    JOIN public.pos_config c ON c.id = s.pos_config_id
    WHERE s.id = v_session_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Session not found';
    END IF;

    -- Skip if already posted
    IF v_session.state = 'posted' THEN
        RETURN jsonb_build_object('status', 'already_posted');
    END IF;

    -- Create journal entry
    INSERT INTO public.account_move (
        organization_id, company_id, name, ref,
        move_type, journal_id, date,
        state,
        created_by
    ) VALUES (
        v_session.organization_id,
        v_session.company_id,
        'POS/' || v_session.name,
        v_session.name,
        'entry',
        v_session.journal_id,
        v_session.stop_at::date,
        'draft',
        v_session.user_id
    ) RETURNING id INTO v_move_id;

    -- Create journal lines for each payment method
    FOR v_payment_method IN
        SELECT
            pm.id as method_id,
            pm.name as method_name,
            pm.receivable_account_id,
            SUM(pp.amount) as total_amount
        FROM public.pos_payments pp
        JOIN public.pos_payment_methods pm ON pm.id = pp.payment_method_id
        WHERE pp.pos_session_id = v_session_id
          AND NOT pp.is_change
        GROUP BY pm.id, pm.name, pm.receivable_account_id
    LOOP
        -- Debit line (receivable)
        INSERT INTO public.account_move_line (
            organization_id, move_id, name,
            account_id, debit, credit,
            created_by
        ) VALUES (
            v_session.organization_id,
            v_move_id,
            'POS Payment - ' || v_payment_method.method_name,
            v_payment_method.receivable_account_id,
            v_payment_method.total_amount,
            0,
            v_session.user_id
        );

        -- Credit line (income - simplified, should match product income accounts)
        INSERT INTO public.account_move_line (
            organization_id, move_id, name,
            account_id, debit, credit,
            created_by
        )
        SELECT
            v_session.organization_id,
            v_move_id,
            'POS Sales - ' || v_payment_method.method_name,
            (SELECT id FROM public.account_accounts
             WHERE organization_id = v_session.organization_id
               AND code = '400000' LIMIT 1),  -- Sales income account
            0,
            v_payment_method.total_amount,
            v_session.user_id;
    END LOOP;

    -- Post the journal entry
    UPDATE public.account_move
    SET state = 'posted'
    WHERE id = v_move_id;

    -- Update session
    UPDATE public.pos_sessions
    SET state = 'posted',
        move_id = v_move_id
    WHERE id = v_session_id;

    RETURN jsonb_build_object(
        'move_id', v_move_id,
        'status', 'posted'
    );
END;
$$;

-- =====================================================
-- ANALYTICS REFRESH HANDLER
-- =====================================================

CREATE OR REPLACE FUNCTION handle_pos_analytics_refresh_job(
    p_job_id uuid,
    p_payload jsonb
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM public.refresh_pos_analytics();

    RETURN jsonb_build_object(
        'status', 'completed',
        'refreshed_at', now()
    );
END;
$$;

-- =====================================================
-- REGISTER JOB HANDLERS
-- =====================================================

-- Register POS job handlers (if you have a job handler registry)
DO $$
BEGIN
    -- These would integrate with your existing queue_job_handlers pattern
    -- Just documenting the job types that need to be recognized by workers

    -- Job type: pos_offline_sync
    -- Handler: handle_pos_offline_sync_job

    -- Job type: pos_receipt_generation
    -- Handler: handle_pos_receipt_generation_job

    -- Job type: pos_stock_move_create
    -- Handler: handle_pos_stock_move_create_job

    -- Job type: pos_session_accounting
    -- Handler: handle_pos_session_accounting_job

    -- Job type: pos_analytics_refresh
    -- Handler: handle_pos_analytics_refresh_job
END $$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION handle_pos_offline_sync_job IS 'Sync offline POS orders when connection restored';
COMMENT ON FUNCTION handle_pos_receipt_generation_job IS 'Generate and email POS receipts';
COMMENT ON FUNCTION handle_pos_stock_move_create_job IS 'Create and validate stock moves for POS orders';
COMMENT ON FUNCTION handle_pos_session_accounting_job IS 'Create accounting journal entries for closed POS sessions';
COMMENT ON FUNCTION handle_pos_analytics_refresh_job IS 'Refresh POS analytics materialized views';
