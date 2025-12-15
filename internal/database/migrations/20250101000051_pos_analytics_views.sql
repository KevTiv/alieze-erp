-- Migration: POS Analytics and Reporting Views
-- Description: Materialized views and functions for POS analytics, pricing, margins, and insights
-- Created: 2025-01-26
-- Dependencies: 20250101000050_pos_business_functions.sql

-- =====================================================
-- POS SESSION ANALYTICS VIEW
-- =====================================================

CREATE MATERIALIZED VIEW pos_session_analytics_mv AS
SELECT
    s.id as session_id,
    s.organization_id,
    s.name as session_name,
    s.pos_config_id,
    c.name as config_name,
    s.user_id as cashier_id,
    s.state,
    s.start_at,
    s.stop_at,
    EXTRACT(EPOCH FROM (COALESCE(s.stop_at, now()) - s.start_at)) / 3600 as duration_hours,

    -- Cash Management
    s.cash_register_balance_start,
    s.cash_register_balance_end,
    s.cash_register_balance_end_real,
    s.cash_register_difference,

    -- Sales Summary
    s.total_orders_count,
    s.total_amount,
    s.total_tax_amount,
    CASE WHEN s.total_orders_count > 0 THEN s.total_amount / s.total_orders_count ELSE 0 END as avg_order_value,

    -- Payment Breakdown
    s.cash_payment_amount,
    s.card_payment_amount,
    s.other_payment_amount,

    -- Performance Metrics
    CASE
        WHEN s.stop_at IS NOT NULL AND s.start_at IS NOT NULL
        THEN s.total_orders_count::numeric / NULLIF(EXTRACT(EPOCH FROM (s.stop_at - s.start_at)) / 3600, 0)
        ELSE 0
    END as orders_per_hour,

    -- Margin Analysis
    margin_stats.total_cost,
    margin_stats.total_margin,
    margin_stats.avg_margin_pct,

    -- Discount Analysis
    discount_stats.total_discounts,
    discount_stats.discount_count,

    -- Stock Alerts
    alert_stats.inventory_alerts_count,

    s.created_at
FROM pos_sessions s
JOIN pos_config c ON c.id = s.pos_config_id
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM((sol.custom_fields->>'cost_price')::numeric * sol.product_uom_qty), 0) as total_cost,
        COALESCE(SUM((sol.custom_fields->>'margin')::numeric), 0) as total_margin,
        CASE
            WHEN COUNT(*) > 0 THEN AVG((sol.custom_fields->>'margin_pct')::numeric)
            ELSE 0
        END as avg_margin_pct
    FROM sales_orders so
    JOIN sales_order_lines sol ON sol.order_id = so.id
    WHERE so.pos_session_id = s.id
) margin_stats ON true
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM(discount_amount), 0) as total_discounts,
        COUNT(*) as discount_count
    FROM pos_order_discounts pod
    JOIN sales_orders so ON so.id = pod.order_id
    WHERE so.pos_session_id = s.id
) discount_stats ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*) as inventory_alerts_count
    FROM pos_inventory_alerts pia
    JOIN sales_orders so ON so.id = pia.order_id
    WHERE so.pos_session_id = s.id
      AND pia.state = 'open'
) alert_stats ON true
WHERE s.deleted_at IS NULL;

CREATE INDEX idx_pos_session_analytics_org ON pos_session_analytics_mv(organization_id);
CREATE INDEX idx_pos_session_analytics_config ON pos_session_analytics_mv(pos_config_id);
CREATE INDEX idx_pos_session_analytics_date ON pos_session_analytics_mv(start_at);
CREATE INDEX idx_pos_session_analytics_state ON pos_session_analytics_mv(organization_id, state);

-- =====================================================
-- POS PRODUCT PERFORMANCE VIEW
-- =====================================================

CREATE MATERIALIZED VIEW pos_product_performance_mv AS
WITH product_sales AS (
    SELECT
        sol.organization_id,
        sol.product_id,
        COUNT(DISTINCT so.id) as orders_count,
        SUM(sol.product_uom_qty) as total_qty_sold,
        SUM(sol.price_subtotal) as total_revenue,
        SUM(sol.price_tax) as total_tax,
        SUM((sol.custom_fields->>'cost_price')::numeric * sol.product_uom_qty) as total_cost,
        SUM((sol.custom_fields->>'margin')::numeric) as total_margin,
        AVG((sol.custom_fields->>'margin_pct')::numeric) as avg_margin_pct,
        COUNT(*) FILTER (WHERE (sol.custom_fields->>'stock_available')::numeric < sol.product_uom_qty) as low_stock_occurrences,
        MIN(so.date_order) as first_sale_date,
        MAX(so.date_order) as last_sale_date
    FROM sales_order_lines sol
    JOIN sales_orders so ON so.id = sol.order_id
    WHERE so.is_pos_order = true
      AND so.state = 'sale'
      AND sol.product_id IS NOT NULL
    GROUP BY sol.organization_id, sol.product_id
)
SELECT
    ps.organization_id,
    ps.product_id,
    p.name as product_name,
    p.default_code,
    p.barcode,
    p.category_id,
    pc.name as category_name,

    -- Sales Metrics
    ps.orders_count,
    ps.total_qty_sold,
    ps.total_revenue,
    ps.total_tax,
    ps.total_cost,
    ps.total_margin,
    ps.avg_margin_pct,

    -- Performance
    CASE WHEN ps.total_revenue > 0 THEN (ps.total_margin / ps.total_revenue) * 100 ELSE 0 END as margin_pct_of_revenue,
    ps.total_revenue / NULLIF(ps.orders_count, 0) as avg_revenue_per_order,
    ps.total_qty_sold / NULLIF(ps.orders_count, 0) as avg_qty_per_order,

    -- Current State
    p.list_price,
    p.standard_price,
    CASE WHEN p.list_price > 0 THEN ((p.list_price - p.standard_price) / p.list_price) * 100 ELSE 0 END as current_margin_pct,

    -- Inventory Status
    ps.low_stock_occurrences,

    -- Sales History
    ps.first_sale_date,
    ps.last_sale_date,
    EXTRACT(EPOCH FROM (COALESCE(ps.last_sale_date, now()) - ps.first_sale_date)) / 86400 as days_on_sale,

    -- Current Stock (from latest available)
    COALESCE((
        SELECT SUM(quantity)
        FROM stock_quants sq
        JOIN stock_locations sl ON sl.id = sq.location_id
        WHERE sq.product_id = ps.product_id
          AND sl.usage = 'internal'
    ), 0) as current_stock_qty

FROM product_sales ps
JOIN products p ON p.id = ps.product_id
LEFT JOIN product_categories pc ON pc.id = p.category_id
WHERE p.deleted_at IS NULL;

CREATE INDEX idx_pos_product_performance_org ON pos_product_performance_mv(organization_id);
CREATE INDEX idx_pos_product_performance_product ON pos_product_performance_mv(product_id);
CREATE INDEX idx_pos_product_performance_category ON pos_product_performance_mv(category_id);
CREATE INDEX idx_pos_product_performance_revenue ON pos_product_performance_mv(organization_id, total_revenue DESC);
CREATE INDEX idx_pos_product_performance_margin ON pos_product_performance_mv(organization_id, total_margin DESC);

-- =====================================================
-- POS PRICING INSIGHTS FUNCTION
-- =====================================================

CREATE OR REPLACE FUNCTION pos_pricing_insights(
    p_organization_id uuid,
    p_product_id uuid DEFAULT NULL,
    p_category_id uuid DEFAULT NULL,
    p_date_from timestamptz DEFAULT (now() - interval '30 days'),
    p_date_to timestamptz DEFAULT now()
) RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    current_price numeric,
    current_cost numeric,
    current_margin_pct numeric,
    avg_selling_price numeric,
    min_selling_price numeric,
    max_selling_price numeric,
    price_override_count bigint,
    discount_count bigint,
    avg_discount_pct numeric,
    total_revenue numeric,
    total_margin numeric,
    qty_sold numeric,
    recommended_price numeric,
    price_elasticity_indicator varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH pricing_data AS (
        SELECT
            sol.product_id,
            p.name as product_name,
            p.list_price as current_price,
            p.standard_price as current_cost,
            CASE WHEN p.list_price > 0 THEN ((p.list_price - p.standard_price) / p.list_price) * 100 ELSE 0 END as current_margin_pct,

            AVG(sol.price_unit) as avg_selling_price,
            MIN(sol.price_unit) as min_selling_price,
            MAX(sol.price_unit) as max_selling_price,

            COUNT(DISTINCT po.id) as price_override_count,
            COUNT(DISTINCT pod.id) as discount_count,
            AVG(pod.discount_value) FILTER (WHERE pod.discount_type = 'percentage') as avg_discount_pct,

            SUM(sol.price_subtotal) as total_revenue,
            SUM((sol.custom_fields->>'margin')::numeric) as total_margin,
            SUM(sol.product_uom_qty) as qty_sold
        FROM sales_order_lines sol
        JOIN sales_orders so ON so.id = sol.order_id
        JOIN products p ON p.id = sol.product_id
        LEFT JOIN pos_pricing_overrides po ON po.order_line_id = sol.id
        LEFT JOIN pos_order_discounts pod ON pod.order_id = so.id
        WHERE so.organization_id = p_organization_id
          AND so.is_pos_order = true
          AND so.state = 'sale'
          AND so.date_order BETWEEN p_date_from AND p_date_to
          AND (p_product_id IS NULL OR sol.product_id = p_product_id)
          AND (p_category_id IS NULL OR p.category_id = p_category_id)
        GROUP BY sol.product_id, p.name, p.list_price, p.standard_price
    )
    SELECT
        pd.product_id,
        pd.product_name,
        pd.current_price,
        pd.current_cost,
        pd.current_margin_pct,
        pd.avg_selling_price,
        pd.min_selling_price,
        pd.max_selling_price,
        pd.price_override_count,
        pd.discount_count,
        pd.avg_discount_pct,
        pd.total_revenue,
        pd.total_margin,
        pd.qty_sold,

        -- Recommended price based on target margin and market data
        CASE
            WHEN pd.current_margin_pct < 20 THEN pd.current_cost * 1.25  -- 25% margin
            WHEN pd.price_override_count > 5 THEN pd.avg_selling_price  -- Market price
            ELSE pd.current_price
        END as recommended_price,

        -- Price elasticity indicator
        CASE
            WHEN pd.price_override_count > 10 THEN 'High Elasticity'
            WHEN pd.discount_count > 5 THEN 'Price Sensitive'
            WHEN pd.current_margin_pct < 10 THEN 'Low Margin Risk'
            ELSE 'Stable'
        END as price_elasticity_indicator
    FROM pricing_data pd
    ORDER BY pd.total_revenue DESC;
END;
$$;

-- =====================================================
-- POS MARGIN ANALYSIS FUNCTION
-- =====================================================

CREATE OR REPLACE FUNCTION pos_margin_analysis(
    p_organization_id uuid,
    p_session_id uuid DEFAULT NULL,
    p_config_id uuid DEFAULT NULL,
    p_date_from timestamptz DEFAULT (now() - interval '7 days'),
    p_date_to timestamptz DEFAULT now()
) RETURNS TABLE (
    analysis_period varchar,
    total_revenue numeric,
    total_cost numeric,
    total_margin numeric,
    margin_pct numeric,
    orders_count bigint,
    avg_margin_per_order numeric,
    low_margin_orders_count bigint,
    high_margin_orders_count bigint,
    price_overrides_count bigint,
    discount_impact numeric,
    margin_by_category jsonb
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH order_margins AS (
        SELECT
            so.id as order_id,
            so.pos_session_id,
            s.pos_config_id,
            so.amount_total as revenue,
            SUM((sol.custom_fields->>'cost_price')::numeric * sol.product_uom_qty) as cost,
            SUM((sol.custom_fields->>'margin')::numeric) as margin,
            CASE
                WHEN SUM(sol.price_subtotal) > 0
                THEN (SUM((sol.custom_fields->>'margin')::numeric) / SUM(sol.price_subtotal)) * 100
                ELSE 0
            END as margin_pct,
            so.amount_discount
        FROM sales_orders so
        JOIN pos_sessions s ON s.id = so.pos_session_id
        JOIN sales_order_lines sol ON sol.order_id = so.id
        WHERE so.organization_id = p_organization_id
          AND so.is_pos_order = true
          AND so.state = 'sale'
          AND so.date_order BETWEEN p_date_from AND p_date_to
          AND (p_session_id IS NULL OR so.pos_session_id = p_session_id)
          AND (p_config_id IS NULL OR s.pos_config_id = p_config_id)
        GROUP BY so.id, so.pos_session_id, s.pos_config_id, so.amount_total, so.amount_discount
    ),
    category_margins AS (
        SELECT
            jsonb_object_agg(
                COALESCE(pc.name, 'Uncategorized'),
                jsonb_build_object(
                    'revenue', COALESCE(SUM(sol.price_subtotal), 0),
                    'margin', COALESCE(SUM((sol.custom_fields->>'margin')::numeric), 0),
                    'margin_pct', CASE
                        WHEN SUM(sol.price_subtotal) > 0
                        THEN (SUM((sol.custom_fields->>'margin')::numeric) / SUM(sol.price_subtotal)) * 100
                        ELSE 0
                    END
                )
            ) as category_data
        FROM sales_orders so
        JOIN sales_order_lines sol ON sol.order_id = so.id
        JOIN products p ON p.id = sol.product_id
        LEFT JOIN product_categories pc ON pc.id = p.category_id
        WHERE so.organization_id = p_organization_id
          AND so.is_pos_order = true
          AND so.state = 'sale'
          AND so.date_order BETWEEN p_date_from AND p_date_to
        GROUP BY pc.name
    )
    SELECT
        to_char(p_date_from, 'YYYY-MM-DD') || ' to ' || to_char(p_date_to, 'YYYY-MM-DD') as analysis_period,
        COALESCE(SUM(om.revenue), 0) as total_revenue,
        COALESCE(SUM(om.cost), 0) as total_cost,
        COALESCE(SUM(om.margin), 0) as total_margin,
        CASE
            WHEN SUM(om.revenue) > 0 THEN (SUM(om.margin) / SUM(om.revenue)) * 100
            ELSE 0
        END as margin_pct,
        COUNT(om.order_id) as orders_count,
        CASE WHEN COUNT(om.order_id) > 0 THEN SUM(om.margin) / COUNT(om.order_id) ELSE 0 END as avg_margin_per_order,
        COUNT(*) FILTER (WHERE om.margin_pct < 15) as low_margin_orders_count,
        COUNT(*) FILTER (WHERE om.margin_pct > 40) as high_margin_orders_count,
        (SELECT COUNT(*) FROM pos_pricing_overrides WHERE created_at BETWEEN p_date_from AND p_date_to) as price_overrides_count,
        COALESCE(SUM(om.amount_discount), 0) as discount_impact,
        (SELECT category_data FROM category_margins LIMIT 1) as margin_by_category
    FROM order_margins om;
END;
$$;

-- =====================================================
-- POS CASHIER PERFORMANCE FUNCTION
-- =====================================================

CREATE OR REPLACE FUNCTION pos_cashier_performance(
    p_organization_id uuid,
    p_date_from timestamptz DEFAULT (now() - interval '30 days'),
    p_date_to timestamptz DEFAULT now()
) RETURNS TABLE (
    cashier_id uuid,
    sessions_count bigint,
    total_orders bigint,
    total_revenue numeric,
    total_margin numeric,
    avg_order_value numeric,
    avg_margin_pct numeric,
    discounts_given_count bigint,
    price_overrides_count bigint,
    cash_discrepancy_total numeric,
    inventory_alerts_count bigint,
    performance_score numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.user_id as cashier_id,
        COUNT(DISTINCT s.id) as sessions_count,
        COALESCE(SUM(s.total_orders_count), 0) as total_orders,
        COALESCE(SUM(s.total_amount), 0) as total_revenue,
        COALESCE(SUM(margin_data.total_margin), 0) as total_margin,
        CASE
            WHEN SUM(s.total_orders_count) > 0
            THEN SUM(s.total_amount) / SUM(s.total_orders_count)
            ELSE 0
        END as avg_order_value,
        AVG(margin_data.avg_margin_pct) as avg_margin_pct,
        COALESCE(SUM(discount_data.discount_count), 0) as discounts_given_count,
        COALESCE(SUM(override_data.override_count), 0) as price_overrides_count,
        COALESCE(SUM(ABS(s.cash_register_difference)), 0) as cash_discrepancy_total,
        COALESCE(SUM(alert_data.alert_count), 0) as inventory_alerts_count,

        -- Performance score (0-100)
        GREATEST(0, LEAST(100,
            100
            - (COALESCE(SUM(ABS(s.cash_register_difference)), 0) * 2)  -- Penalize cash errors
            - (COALESCE(SUM(alert_data.alert_count), 0) * 1)  -- Penalize inventory issues
            + (CASE WHEN AVG(margin_data.avg_margin_pct) > 25 THEN 10 ELSE 0 END)  -- Bonus for good margins
        )) as performance_score
    FROM pos_sessions s
    LEFT JOIN LATERAL (
        SELECT
            AVG((sol.custom_fields->>'margin_pct')::numeric) as avg_margin_pct,
            SUM((sol.custom_fields->>'margin')::numeric) as total_margin
        FROM sales_orders so
        JOIN sales_order_lines sol ON sol.order_id = so.id
        WHERE so.pos_session_id = s.id
    ) margin_data ON true
    LEFT JOIN LATERAL (
        SELECT COUNT(*) as discount_count
        FROM sales_orders so
        JOIN pos_order_discounts pod ON pod.order_id = so.id
        WHERE so.pos_session_id = s.id
    ) discount_data ON true
    LEFT JOIN LATERAL (
        SELECT COUNT(*) as override_count
        FROM sales_orders so
        JOIN sales_order_lines sol ON sol.order_id = so.id
        JOIN pos_pricing_overrides po ON po.order_line_id = sol.id
        WHERE so.pos_session_id = s.id
    ) override_data ON true
    LEFT JOIN LATERAL (
        SELECT COUNT(*) as alert_count
        FROM sales_orders so
        JOIN pos_inventory_alerts pia ON pia.order_id = so.id
        WHERE so.pos_session_id = s.id
    ) alert_data ON true
    WHERE s.organization_id = p_organization_id
      AND s.start_at BETWEEN p_date_from AND p_date_to
      AND s.deleted_at IS NULL
    GROUP BY s.user_id
    ORDER BY total_revenue DESC;
END;
$$;

-- =====================================================
-- REFRESH FUNCTIONS
-- =====================================================

CREATE OR REPLACE FUNCTION refresh_pos_analytics()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY pos_session_analytics_mv;
    REFRESH MATERIALIZED VIEW CONCURRENTLY pos_product_performance_mv;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON MATERIALIZED VIEW pos_session_analytics_mv IS 'Comprehensive POS session analytics with margins, performance metrics';
COMMENT ON MATERIALIZED VIEW pos_product_performance_mv IS 'Product-level sales performance with margin analysis';
COMMENT ON FUNCTION pos_pricing_insights IS 'Analyze pricing trends and get AI-powered pricing recommendations';
COMMENT ON FUNCTION pos_margin_analysis IS 'Detailed margin analysis by session, config, or time period';
COMMENT ON FUNCTION pos_cashier_performance IS 'Cashier performance metrics including accuracy and efficiency';
COMMENT ON FUNCTION refresh_pos_analytics IS 'Refresh all POS materialized views';
