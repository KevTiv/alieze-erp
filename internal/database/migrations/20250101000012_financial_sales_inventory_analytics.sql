-- Migration: Financial, Sales, and Inventory Analytics
-- Description: Comprehensive business intelligence functions
-- Created: 2025-01-01

-- =====================================================
-- FINANCIAL ANALYTICS
-- =====================================================

-- Accounts Receivable Aging Report
CREATE OR REPLACE FUNCTION analytics_ar_aging(
    p_organization_id uuid,
    p_as_of_date date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    partner_id uuid,
    partner_name varchar,
    current_amount numeric,
    days_30 numeric,
    days_60 numeric,
    days_90 numeric,
    over_90 numeric,
    total_due numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.name,
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due <= 0), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due BETWEEN 1 AND 30), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due BETWEEN 31 AND 60), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due BETWEEN 61 AND 90), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due > 90), 0),
        COALESCE(SUM(i.amount_residual), 0)
    FROM contacts c
    LEFT JOIN invoices i ON i.partner_id = c.id
        AND i.organization_id = p_organization_id
        AND i.move_type IN ('out_invoice', 'out_refund')
        AND i.state = 'posted'
        AND i.payment_state IN ('not_paid', 'partial')
        AND i.deleted_at IS NULL
    WHERE c.organization_id = p_organization_id
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name
    HAVING SUM(i.amount_residual) > 0
    ORDER BY total_due DESC;
END;
$$;

-- Accounts Payable Aging Report
CREATE OR REPLACE FUNCTION analytics_ap_aging(
    p_organization_id uuid,
    p_as_of_date date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    partner_id uuid,
    partner_name varchar,
    current_amount numeric,
    days_30 numeric,
    days_60 numeric,
    days_90 numeric,
    over_90 numeric,
    total_due numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.name,
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due <= 0), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due BETWEEN 1 AND 30), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due BETWEEN 31 AND 60), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due BETWEEN 61 AND 90), 0),
        COALESCE(SUM(i.amount_residual) FILTER (WHERE p_as_of_date - i.invoice_date_due > 90), 0),
        COALESCE(SUM(i.amount_residual), 0)
    FROM contacts c
    LEFT JOIN invoices i ON i.partner_id = c.id
        AND i.organization_id = p_organization_id
        AND i.move_type IN ('in_invoice', 'in_refund')
        AND i.state = 'posted'
        AND i.payment_state IN ('not_paid', 'partial')
        AND i.deleted_at IS NULL
    WHERE c.organization_id = p_organization_id
      AND c.is_vendor = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name
    HAVING SUM(i.amount_residual) > 0
    ORDER BY total_due DESC;
END;
$$;

-- Cash Flow Analysis
CREATE OR REPLACE FUNCTION analytics_cash_flow(
    p_organization_id uuid,
    p_date_from date,
    p_date_to date,
    p_interval varchar DEFAULT 'month' -- day, week, month
)
RETURNS TABLE (
    period_start date,
    period_end date,
    cash_in numeric,
    cash_out numeric,
    net_cash_flow numeric,
    cumulative_cash_flow numeric
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_interval_text text;
BEGIN
    v_interval_text := '1 ' || p_interval;

    RETURN QUERY
    WITH RECURSIVE periods AS (
        SELECT
            p_date_from as period_start,
            (p_date_from + v_interval_text::interval)::date - 1 as period_end
        UNION ALL
        SELECT
            (period_end + 1)::date,
            (period_end + v_interval_text::interval)::date
        FROM periods
        WHERE period_end < p_date_to
    ),
    cash_movements AS (
        SELECT
            p.period_start,
            p.period_end,
            -- Cash in (customer payments)
            COALESCE(SUM(pay.amount) FILTER (WHERE pay.payment_type = 'inbound'), 0) as cash_in,
            -- Cash out (vendor payments)
            COALESCE(SUM(pay.amount) FILTER (WHERE pay.payment_type = 'outbound'), 0) as cash_out
        FROM periods p
        LEFT JOIN payments pay ON pay.organization_id = p_organization_id
            AND pay.payment_date BETWEEN p.period_start AND p.period_end
            AND pay.state = 'posted'
            AND pay.deleted_at IS NULL
        GROUP BY p.period_start, p.period_end
    )
    SELECT
        cm.period_start,
        cm.period_end,
        cm.cash_in,
        cm.cash_out,
        cm.cash_in - cm.cash_out as net_flow,
        SUM(cm.cash_in - cm.cash_out) OVER (ORDER BY cm.period_start) as cumulative
    FROM cash_movements cm
    ORDER BY cm.period_start;
END;
$$;

-- Revenue by Period
CREATE OR REPLACE FUNCTION analytics_revenue_by_period(
    p_organization_id uuid,
    p_date_from date,
    p_date_to date,
    p_group_by varchar DEFAULT 'month' -- day, week, month, quarter, year
)
RETURNS TABLE (
    period varchar,
    period_start date,
    invoiced_amount numeric,
    paid_amount numeric,
    outstanding_amount numeric,
    invoice_count integer
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        TO_CHAR(i.invoice_date,
            CASE p_group_by
                WHEN 'day' THEN 'YYYY-MM-DD'
                WHEN 'week' THEN 'IYYY-IW'
                WHEN 'month' THEN 'YYYY-MM'
                WHEN 'quarter' THEN 'YYYY-Q'
                WHEN 'year' THEN 'YYYY'
            END
        ),
        DATE_TRUNC(p_group_by, i.invoice_date)::date,
        COALESCE(SUM(i.amount_total), 0),
        COALESCE(SUM(i.amount_total - i.amount_residual), 0),
        COALESCE(SUM(i.amount_residual), 0),
        COUNT(i.id)::integer
    FROM invoices i
    WHERE i.organization_id = p_organization_id
      AND i.move_type IN ('out_invoice', 'out_refund')
      AND i.state = 'posted'
      AND i.invoice_date BETWEEN p_date_from AND p_date_to
      AND i.deleted_at IS NULL
    GROUP BY
        TO_CHAR(i.invoice_date,
            CASE p_group_by
                WHEN 'day' THEN 'YYYY-MM-DD'
                WHEN 'week' THEN 'IYYY-IW'
                WHEN 'month' THEN 'YYYY-MM'
                WHEN 'quarter' THEN 'YYYY-Q'
                WHEN 'year' THEN 'YYYY'
            END
        ),
        DATE_TRUNC(p_group_by, i.invoice_date)
    ORDER BY period_start;
END;
$$;

-- Profitability by Product
CREATE OR REPLACE FUNCTION analytics_product_profitability(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    category_name varchar,
    qty_sold numeric,
    revenue numeric,
    cost numeric,
    profit numeric,
    profit_margin_pct numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id,
        p.name,
        pc.name,
        COALESCE(SUM(sol.product_uom_qty), 0),
        COALESCE(SUM(sol.price_subtotal), 0),
        COALESCE(SUM(sol.product_uom_qty * p.standard_price), 0),
        COALESCE(SUM(sol.price_subtotal) - SUM(sol.product_uom_qty * p.standard_price), 0),
        CASE
            WHEN SUM(sol.price_subtotal) > 0
            THEN ROUND(((SUM(sol.price_subtotal) - SUM(sol.product_uom_qty * p.standard_price)) / SUM(sol.price_subtotal)) * 100, 2)
            ELSE 0
        END
    FROM products p
    LEFT JOIN product_categories pc ON p.category_id = pc.id
    LEFT JOIN sales_order_lines sol ON sol.product_id = p.id
    LEFT JOIN sales_orders so ON sol.order_id = so.id
        AND so.organization_id = p_organization_id
        AND so.state IN ('sale', 'done')
        AND so.date_order::date BETWEEN p_date_from AND p_date_to
        AND so.deleted_at IS NULL
    WHERE p.organization_id = p_organization_id
      AND p.deleted_at IS NULL
      AND p.active = true
    GROUP BY p.id, p.name, pc.name
    HAVING SUM(sol.product_uom_qty) > 0
    ORDER BY profit DESC;
END;
$$;

-- =====================================================
-- SALES ANALYTICS
-- =====================================================

-- Sales Pipeline Health
CREATE OR REPLACE FUNCTION analytics_sales_pipeline(
    p_organization_id uuid
)
RETURNS TABLE (
    stage_id uuid,
    stage_name varchar,
    stage_probability integer,
    opportunity_count integer,
    total_expected_revenue numeric,
    weighted_revenue numeric,
    avg_days_in_stage numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ls.id,
        ls.name,
        ls.probability,
        COUNT(l.id)::integer,
        COALESCE(SUM(l.expected_revenue), 0),
        COALESCE(SUM(l.expected_revenue * l.probability / 100.0), 0),
        COALESCE(AVG(EXTRACT(day FROM now() - l.date_last_stage_update)), 0)
    FROM lead_stages ls
    LEFT JOIN leads l ON l.stage_id = ls.id
        AND l.organization_id = p_organization_id
        AND l.lead_type = 'opportunity'
        AND l.active = true
        AND l.won_status = 'ongoing'
        AND l.deleted_at IS NULL
    WHERE ls.organization_id = p_organization_id
    GROUP BY ls.id, ls.name, ls.probability, ls.sequence
    ORDER BY ls.sequence;
END;
$$;

-- Win/Loss Analysis
CREATE OR REPLACE FUNCTION analytics_win_loss_analysis(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '90 days',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result jsonb;
    v_total_opps integer;
    v_won_opps integer;
    v_lost_opps integer;
    v_win_rate numeric;
    v_total_won_revenue numeric;
    v_avg_deal_size numeric;
    v_avg_sales_cycle numeric;
BEGIN
    -- Count opportunities
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE won_status = 'won'),
        COUNT(*) FILTER (WHERE won_status = 'lost')
    INTO v_total_opps, v_won_opps, v_lost_opps
    FROM leads
    WHERE organization_id = p_organization_id
      AND lead_type = 'opportunity'
      AND date_closed BETWEEN p_date_from AND p_date_to;

    -- Calculate win rate
    v_win_rate := CASE
        WHEN v_total_opps > 0
        THEN ROUND((v_won_opps::numeric / v_total_opps) * 100, 2)
        ELSE 0
    END;

    -- Get won revenue and avg deal size
    SELECT
        COALESCE(SUM(expected_revenue), 0),
        COALESCE(AVG(expected_revenue), 0)
    INTO v_total_won_revenue, v_avg_deal_size
    FROM leads
    WHERE organization_id = p_organization_id
      AND lead_type = 'opportunity'
      AND won_status = 'won'
      AND date_closed BETWEEN p_date_from AND p_date_to;

    -- Calculate avg sales cycle
    SELECT COALESCE(AVG(EXTRACT(day FROM date_closed - date_open)), 0)
    INTO v_avg_sales_cycle
    FROM leads
    WHERE organization_id = p_organization_id
      AND lead_type = 'opportunity'
      AND won_status = 'won'
      AND date_closed BETWEEN p_date_from AND p_date_to;

    -- Build result
    v_result := jsonb_build_object(
        'period_start', p_date_from,
        'period_end', p_date_to,
        'total_opportunities', v_total_opps,
        'won_opportunities', v_won_opps,
        'lost_opportunities', v_lost_opps,
        'win_rate_pct', v_win_rate,
        'total_won_revenue', v_total_won_revenue,
        'avg_deal_size', v_avg_deal_size,
        'avg_sales_cycle_days', v_avg_sales_cycle
    );

    RETURN v_result;
END;
$$;

-- Top Customers by Revenue
CREATE OR REPLACE FUNCTION analytics_top_customers(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '12 months',
    p_date_to date DEFAULT CURRENT_DATE,
    p_limit integer DEFAULT 10
)
RETURNS TABLE (
    rank integer,
    customer_id uuid,
    customer_name varchar,
    total_revenue numeric,
    invoice_count integer,
    avg_invoice_amount numeric,
    last_order_date date
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ROW_NUMBER() OVER (ORDER BY SUM(i.amount_total) DESC)::integer,
        c.id,
        c.name,
        COALESCE(SUM(i.amount_total), 0),
        COUNT(i.id)::integer,
        COALESCE(AVG(i.amount_total), 0),
        MAX(i.invoice_date)
    FROM contacts c
    LEFT JOIN invoices i ON i.partner_id = c.id
        AND i.organization_id = p_organization_id
        AND i.move_type IN ('out_invoice')
        AND i.state = 'posted'
        AND i.invoice_date BETWEEN p_date_from AND p_date_to
        AND i.deleted_at IS NULL
    WHERE c.organization_id = p_organization_id
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name
    HAVING SUM(i.amount_total) > 0
    ORDER BY total_revenue DESC
    LIMIT p_limit;
END;
$$;

-- =====================================================
-- INVENTORY ANALYTICS
-- =====================================================

-- Stock Valuation by Location
CREATE OR REPLACE FUNCTION analytics_stock_valuation(
    p_organization_id uuid,
    p_location_id uuid DEFAULT NULL
)
RETURNS TABLE (
    location_id uuid,
    location_name varchar,
    product_id uuid,
    product_name varchar,
    quantity_on_hand numeric,
    unit_cost numeric,
    total_value numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        sl.id,
        sl.name,
        p.id,
        p.name,
        sq.quantity,
        p.standard_price,
        sq.quantity * p.standard_price
    FROM stock_quants sq
    JOIN stock_locations sl ON sq.location_id = sl.id
    JOIN products p ON sq.product_id = p.id
    WHERE sq.organization_id = p_organization_id
      AND sl.usage = 'internal'
      AND sq.quantity > 0
      AND (p_location_id IS NULL OR sl.id = p_location_id)
    ORDER BY sl.name, total_value DESC;
END;
$$;

-- Low Stock Alerts
CREATE OR REPLACE FUNCTION analytics_low_stock_alerts(
    p_organization_id uuid,
    p_threshold numeric DEFAULT 10
)
RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    product_code varchar,
    location_id uuid,
    location_name varchar,
    current_qty numeric,
    reserved_qty numeric,
    available_qty numeric,
    status varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id,
        p.name,
        p.default_code,
        sl.id,
        sl.name,
        sq.quantity,
        sq.reserved_quantity,
        sq.quantity - sq.reserved_quantity,
        CASE
            WHEN sq.quantity - sq.reserved_quantity <= 0 THEN 'out_of_stock'
            WHEN sq.quantity - sq.reserved_quantity < p_threshold * 0.5 THEN 'critical'
            WHEN sq.quantity - sq.reserved_quantity < p_threshold THEN 'low'
            ELSE 'normal'
        END::varchar
    FROM stock_quants sq
    JOIN products p ON sq.product_id = p.id
    JOIN stock_locations sl ON sq.location_id = sl.id
    WHERE sq.organization_id = p_organization_id
      AND p.product_type = 'storable'
      AND sl.usage = 'internal'
      AND sq.quantity - sq.reserved_quantity < p_threshold
      AND p.deleted_at IS NULL
      AND p.active = true
    ORDER BY sq.quantity - sq.reserved_quantity ASC;
END;
$$;

-- Inventory Turnover Rate
CREATE OR REPLACE FUNCTION analytics_inventory_turnover(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '12 months',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    category_name varchar,
    avg_inventory numeric,
    cost_of_goods_sold numeric,
    turnover_rate numeric,
    days_to_sell numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH product_movements AS (
        SELECT
            sm.product_id,
            AVG(sq.quantity) as avg_qty,
            SUM(sm.product_uom_qty * p.standard_price) FILTER (
                WHERE sl_dest.usage = 'customer'
            ) as cogs
        FROM stock_moves sm
        JOIN products p ON sm.product_id = p.id
        JOIN stock_locations sl_dest ON sm.location_dest_id = sl_dest.id
        LEFT JOIN stock_quants sq ON sq.product_id = sm.product_id
        WHERE sm.organization_id = p_organization_id
          AND sm.state = 'done'
          AND sm.date::date BETWEEN p_date_from AND p_date_to
          AND sm.deleted_at IS NULL
        GROUP BY sm.product_id
    )
    SELECT
        p.id,
        p.name,
        pc.name,
        COALESCE(pm.avg_qty, 0),
        COALESCE(pm.cogs, 0),
        CASE
            WHEN pm.avg_qty > 0 THEN ROUND(pm.cogs / (pm.avg_qty * p.standard_price), 2)
            ELSE 0
        END,
        CASE
            WHEN pm.cogs > 0 AND pm.avg_qty > 0
            THEN ROUND(365 / (pm.cogs / (pm.avg_qty * p.standard_price)), 0)
            ELSE 0
        END
    FROM products p
    LEFT JOIN product_categories pc ON p.category_id = pc.id
    LEFT JOIN product_movements pm ON p.id = pm.product_id
    WHERE p.organization_id = p_organization_id
      AND p.product_type = 'storable'
      AND p.deleted_at IS NULL
      AND p.active = true
      AND pm.cogs > 0
    ORDER BY turnover_rate DESC;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION analytics_ar_aging IS 'Accounts receivable aging report';
COMMENT ON FUNCTION analytics_ap_aging IS 'Accounts payable aging report';
COMMENT ON FUNCTION analytics_cash_flow IS 'Cash flow analysis over time';
COMMENT ON FUNCTION analytics_revenue_by_period IS 'Revenue breakdown by time period';
COMMENT ON FUNCTION analytics_product_profitability IS 'Product profitability analysis';
COMMENT ON FUNCTION analytics_sales_pipeline IS 'Sales pipeline health and metrics';
COMMENT ON FUNCTION analytics_win_loss_analysis IS 'Win/loss ratio and sales cycle analysis';
COMMENT ON FUNCTION analytics_top_customers IS 'Top customers by revenue';
COMMENT ON FUNCTION analytics_stock_valuation IS 'Stock valuation by location';
COMMENT ON FUNCTION analytics_low_stock_alerts IS 'Low stock and out-of-stock alerts';
COMMENT ON FUNCTION analytics_inventory_turnover IS 'Inventory turnover rate analysis';
