-- Migration: Advanced Insight & Analytics Views
-- Description: Business intelligence views for user insights and analytics
-- Created: 2025-01-01
-- Module: Analytics & Insights

-- =====================================================
-- CUSTOMER INSIGHTS
-- =====================================================

-- Customer Lifetime Value (CLV) Analysis
CREATE OR REPLACE VIEW view_customer_lifetime_value AS
SELECT
    c.id as customer_id,
    c.name as customer_name,
    c.email as customer_email,
    c.is_customer,
    -- Order metrics
    COUNT(DISTINCT so.id) as total_orders,
    COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_revenue,
    COALESCE(AVG(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as avg_order_value,
    -- Time metrics
    MIN(so.date_order) as first_order_date,
    MAX(so.date_order) as last_order_date,
    EXTRACT(EPOCH FROM (MAX(so.date_order) - MIN(so.date_order)))::integer / 86400 as customer_lifetime_days,
    -- Invoice metrics
    COUNT(DISTINCT i.id) FILTER (WHERE i.state = 'posted') as total_invoices,
    COALESCE(SUM(i.amount_total) FILTER (WHERE i.state = 'posted'), 0) as total_invoiced,
    COALESCE(SUM(i.amount_residual) FILTER (WHERE i.payment_state IN ('not_paid', 'partial')), 0) as outstanding_balance,
    -- Payment behavior
    COALESCE(AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid'), 0) as avg_days_to_pay,
    -- Segmentation
    CASE
        WHEN COUNT(DISTINCT so.id) >= 10 AND COALESCE(SUM(so.amount_total), 0) >= 10000 THEN 'VIP'
        WHEN COUNT(DISTINCT so.id) >= 5 AND COALESCE(SUM(so.amount_total), 0) >= 5000 THEN 'High Value'
        WHEN COUNT(DISTINCT so.id) >= 2 THEN 'Regular'
        WHEN COUNT(DISTINCT so.id) = 1 THEN 'One-time'
        ELSE 'Prospect'
    END as customer_segment,
    -- Recency
    EXTRACT(EPOCH FROM (CURRENT_DATE - MAX(so.date_order)))::integer / 86400 as days_since_last_order,
    CASE
        WHEN MAX(so.date_order) >= CURRENT_DATE - INTERVAL '30 days' THEN 'Active'
        WHEN MAX(so.date_order) >= CURRENT_DATE - INTERVAL '90 days' THEN 'At Risk'
        WHEN MAX(so.date_order) >= CURRENT_DATE - INTERVAL '180 days' THEN 'Dormant'
        WHEN MAX(so.date_order) < CURRENT_DATE - INTERVAL '180 days' THEN 'Churned'
        ELSE 'Prospect'
    END as customer_status,
    -- Computed metrics
    CASE
        WHEN EXTRACT(EPOCH FROM (MAX(so.date_order) - MIN(so.date_order)))::integer / 86400 > 0 THEN
            COALESCE(SUM(so.amount_total), 0) / NULLIF(EXTRACT(EPOCH FROM (MAX(so.date_order) - MIN(so.date_order)))::integer / 86400, 0) * 30
        ELSE 0
    END as monthly_revenue_rate
FROM contacts c
LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL
LEFT JOIN invoices i ON i.partner_id = c.id AND i.deleted_at IS NULL AND i.move_type = 'out_invoice'
WHERE c.organization_id = get_current_user_organization_id()
  AND c.deleted_at IS NULL
GROUP BY c.id, c.name, c.email, c.is_customer
ORDER BY total_revenue DESC;

COMMENT ON VIEW view_customer_lifetime_value IS 'Customer lifetime value analysis with segmentation and churn prediction';

-- Customer Purchase Patterns
CREATE OR REPLACE VIEW view_customer_purchase_patterns AS
SELECT
    c.id as customer_id,
    c.name as customer_name,
    -- Product preferences
    (
        SELECT jsonb_agg(prod_obj)
        FROM (
            SELECT
                jsonb_build_object(
                    'product_name', p.name,
                    'product_id', p.id,
                    'times_purchased', COUNT(*),
                    'total_qty', SUM(sol.product_uom_qty),
                    'total_spent', SUM(sol.price_total)
                ) as prod_obj
            FROM sales_order_lines sol
            JOIN sales_orders so ON sol.order_id = so.id
            JOIN products p ON sol.product_id = p.id
            WHERE so.partner_id = c.id
              AND so.state IN ('sale', 'done')
              AND sol.deleted_at IS NULL
            GROUP BY p.id, p.name
            ORDER BY COUNT(*) DESC
            LIMIT 10
        ) top_prods
    ) as top_products,
    -- Category preferences
    (
        SELECT jsonb_agg(cat_obj)
        FROM (
            SELECT
                jsonb_build_object(
                    'category_name', pc.name,
                    'purchase_count', COUNT(*),
                    'total_spent', SUM(sol.price_total)
                ) as cat_obj
            FROM sales_order_lines sol
            JOIN sales_orders so ON sol.order_id = so.id
            JOIN products p ON sol.product_id = p.id
            LEFT JOIN product_categories pc ON p.category_id = pc.id
            WHERE so.partner_id = c.id
              AND so.state IN ('sale', 'done')
              AND sol.deleted_at IS NULL
            GROUP BY pc.name
            ORDER BY COUNT(*) DESC
            LIMIT 5
        ) top_cats
    ) as top_categories,
    -- Purchase frequency
    COUNT(DISTINCT so.id) as total_orders,
    CASE
        WHEN COUNT(DISTINCT so.id) > 0 THEN
            EXTRACT(EPOCH FROM (MAX(so.date_order) - MIN(so.date_order)))::integer / 86400 / NULLIF(COUNT(DISTINCT so.id) - 1, 0)
        ELSE NULL
    END as avg_days_between_orders,
    -- Preferred channels (if you have campaign/source tracking)
    (
        SELECT jsonb_build_object(
            'source', lsrc.name,
            'count', COUNT(*)
        )
        FROM leads l
        LEFT JOIN lead_sources lsrc ON l.source_id = lsrc.id
        WHERE l.contact_id = c.id
          AND l.deleted_at IS NULL
        GROUP BY lsrc.name
        ORDER BY COUNT(*) DESC
        LIMIT 1
    ) as primary_acquisition_source
FROM contacts c
LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL
WHERE c.organization_id = get_current_user_organization_id()
  AND c.is_customer = true
  AND c.deleted_at IS NULL
GROUP BY c.id, c.name;

COMMENT ON VIEW view_customer_purchase_patterns IS 'Detailed customer purchase patterns and product preferences';

-- Customer Health Score
CREATE OR REPLACE VIEW view_customer_health_score AS
WITH customer_metrics AS (
    SELECT
        c.id as customer_id,
        c.name as customer_name,
        -- Recency (0-30 points)
        CASE
            WHEN MAX(so.date_order) >= CURRENT_DATE - INTERVAL '30 days' THEN 30
            WHEN MAX(so.date_order) >= CURRENT_DATE - INTERVAL '60 days' THEN 20
            WHEN MAX(so.date_order) >= CURRENT_DATE - INTERVAL '90 days' THEN 10
            ELSE 0
        END as recency_score,
        -- Frequency (0-30 points)
        CASE
            WHEN COUNT(DISTINCT so.id) >= 10 THEN 30
            WHEN COUNT(DISTINCT so.id) >= 5 THEN 20
            WHEN COUNT(DISTINCT so.id) >= 2 THEN 10
            ELSE 5
        END as frequency_score,
        -- Monetary (0-30 points)
        CASE
            WHEN COALESCE(SUM(so.amount_total), 0) >= 50000 THEN 30
            WHEN COALESCE(SUM(so.amount_total), 0) >= 10000 THEN 20
            WHEN COALESCE(SUM(so.amount_total), 0) >= 1000 THEN 10
            ELSE 5
        END as monetary_score,
        -- Payment behavior (0-10 points)
        CASE
            WHEN AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid') <= 0 THEN 10
            WHEN AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid') <= 15 THEN 7
            WHEN AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid') <= 30 THEN 5
            ELSE 0
        END as payment_score
    FROM contacts c
    LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL AND so.state IN ('sale', 'done')
    LEFT JOIN invoices i ON i.partner_id = c.id AND i.deleted_at IS NULL AND i.move_type = 'out_invoice'
    WHERE c.organization_id = get_current_user_organization_id()
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name
)
SELECT
    customer_id,
    customer_name,
    recency_score,
    frequency_score,
    monetary_score,
    payment_score,
    (recency_score + frequency_score + monetary_score + payment_score) as total_health_score,
    CASE
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 80 THEN 'Excellent'
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 60 THEN 'Good'
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 40 THEN 'Fair'
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 20 THEN 'At Risk'
        ELSE 'Critical'
    END as health_status,
    CASE
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 80 THEN 'green'
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 60 THEN 'blue'
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 40 THEN 'yellow'
        WHEN (recency_score + frequency_score + monetary_score + payment_score) >= 20 THEN 'orange'
        ELSE 'red'
    END as health_color
FROM customer_metrics
ORDER BY total_health_score DESC;

COMMENT ON VIEW view_customer_health_score IS 'Customer health scoring based on RFM analysis and payment behavior';

-- =====================================================
-- SALES PERFORMANCE INSIGHTS
-- =====================================================

-- Sales Team Performance Comparison
CREATE OR REPLACE VIEW view_sales_team_performance AS
SELECT
    st.id as team_id,
    st.name as team_name,
    st.code as team_code,
    -- Team members
    COALESCE(array_length(st.member_ids, 1), 0) as team_size,
    -- Lead metrics
    COUNT(DISTINCT l.id) as total_leads,
    COUNT(DISTINCT l.id) FILTER (WHERE l.won_status = 'won') as won_leads,
    COUNT(DISTINCT l.id) FILTER (WHERE l.won_status = 'lost') as lost_leads,
    CASE
        WHEN COUNT(DISTINCT l.id) > 0 THEN
            COUNT(DISTINCT l.id) FILTER (WHERE l.won_status = 'won')::numeric / COUNT(DISTINCT l.id) * 100
        ELSE 0
    END as win_rate,
    -- Sales metrics
    COUNT(DISTINCT so.id) as total_orders,
    COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_revenue,
    COALESCE(AVG(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as avg_deal_size,
    -- Pipeline metrics
    COALESCE(SUM(l.expected_revenue), 0) as pipeline_value,
    COALESCE(SUM(l.expected_revenue * l.probability / 100.0), 0) as weighted_pipeline,
    -- Time metrics
    AVG(EXTRACT(EPOCH FROM (l.date_closed - l.date_open))::integer / 86400) FILTER (WHERE l.won_status = 'won') as avg_days_to_close,
    -- Per member metrics
    CASE
        WHEN COALESCE(array_length(st.member_ids, 1), 0) > 0 THEN
            COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) / array_length(st.member_ids, 1)
        ELSE 0
    END as revenue_per_member,
    -- Monthly trend (last 30 days vs previous 30 days)
    COALESCE(SUM(so.amount_total) FILTER (
        WHERE so.state IN ('sale', 'done')
        AND so.date_order >= CURRENT_DATE - INTERVAL '30 days'
    ), 0) as revenue_last_30_days,
    COALESCE(SUM(so.amount_total) FILTER (
        WHERE so.state IN ('sale', 'done')
        AND so.date_order >= CURRENT_DATE - INTERVAL '60 days'
        AND so.date_order < CURRENT_DATE - INTERVAL '30 days'
    ), 0) as revenue_previous_30_days,
    -- Growth calculation
    CASE
        WHEN COALESCE(SUM(so.amount_total) FILTER (
            WHERE so.state IN ('sale', 'done')
            AND so.date_order >= CURRENT_DATE - INTERVAL '60 days'
            AND so.date_order < CURRENT_DATE - INTERVAL '30 days'
        ), 0) > 0 THEN
            ((COALESCE(SUM(so.amount_total) FILTER (
                WHERE so.state IN ('sale', 'done')
                AND so.date_order >= CURRENT_DATE - INTERVAL '30 days'
            ), 0) - COALESCE(SUM(so.amount_total) FILTER (
                WHERE so.state IN ('sale', 'done')
                AND so.date_order >= CURRENT_DATE - INTERVAL '60 days'
                AND so.date_order < CURRENT_DATE - INTERVAL '30 days'
            ), 0)) / NULLIF(COALESCE(SUM(so.amount_total) FILTER (
                WHERE so.state IN ('sale', 'done')
                AND so.date_order >= CURRENT_DATE - INTERVAL '60 days'
                AND so.date_order < CURRENT_DATE - INTERVAL '30 days'
            ), 0), 0)) * 100
        ELSE 0
    END as growth_percentage
FROM sales_teams st
LEFT JOIN leads l ON l.team_id = st.id AND l.deleted_at IS NULL
LEFT JOIN sales_orders so ON so.team_id = st.id AND so.deleted_at IS NULL
WHERE st.organization_id = get_current_user_organization_id()
  AND st.is_active = true
GROUP BY st.id, st.name, st.code, st.member_ids
ORDER BY total_revenue DESC;

COMMENT ON VIEW view_sales_team_performance IS 'Comprehensive sales team performance metrics and comparisons';

-- Sales Pipeline Health
CREATE OR REPLACE VIEW view_sales_pipeline_health AS
WITH stage_metrics AS (
    SELECT
        ls.id as stage_id,
        ls.name as stage_name,
        ls.sequence,
        ls.probability,
        COUNT(l.id) as lead_count,
        COALESCE(SUM(l.expected_revenue), 0) as total_value,
        COALESCE(SUM(l.expected_revenue * l.probability / 100.0), 0) as weighted_value,
        AVG(EXTRACT(EPOCH FROM (CURRENT_DATE - l.date_last_stage_update))::integer / 86400) as avg_days_in_stage,
        -- Stagnation detection
        COUNT(l.id) FILTER (WHERE EXTRACT(EPOCH FROM (CURRENT_DATE - l.date_last_stage_update))::integer / 86400 > 30) as stagnant_leads,
        -- Conversion tracking
        (
            SELECT COUNT(*)
            FROM leads l2
            WHERE l2.stage_id = ls.id
              AND l2.won_status = 'won'
              AND l2.deleted_at IS NULL
        ) as converted_from_stage
    FROM lead_stages ls
    LEFT JOIN leads l ON l.stage_id = ls.id AND l.deleted_at IS NULL AND l.active = true
    WHERE ls.organization_id = get_current_user_organization_id()
    GROUP BY ls.id, ls.name, ls.sequence, ls.probability
)
SELECT
    stage_id,
    stage_name,
    sequence,
    probability,
    lead_count,
    total_value,
    weighted_value,
    avg_days_in_stage,
    stagnant_leads,
    converted_from_stage,
    -- Health indicators
    CASE
        WHEN avg_days_in_stage <= 14 THEN 'Healthy'
        WHEN avg_days_in_stage <= 30 THEN 'Normal'
        WHEN avg_days_in_stage <= 60 THEN 'Slow'
        ELSE 'Stagnant'
    END as stage_health,
    CASE
        WHEN stagnant_leads::numeric / NULLIF(lead_count, 0) > 0.5 THEN 'red'
        WHEN stagnant_leads::numeric / NULLIF(lead_count, 0) > 0.25 THEN 'orange'
        ELSE 'green'
    END as health_color,
    -- Conversion rate from this stage
    CASE
        WHEN lead_count > 0 THEN
            converted_from_stage::numeric / lead_count * 100
        ELSE 0
    END as stage_conversion_rate
FROM stage_metrics
ORDER BY sequence;

COMMENT ON VIEW view_sales_pipeline_health IS 'Pipeline health metrics by stage with stagnation detection';

-- =====================================================
-- PRODUCT PERFORMANCE INSIGHTS
-- =====================================================

-- Product Performance Analysis
CREATE OR REPLACE VIEW view_product_performance AS
SELECT
    p.id as product_id,
    p.name as product_name,
    p.default_code,
    pc.name as category_name,
    -- Sales metrics
    COUNT(DISTINCT sol.order_id) as times_sold,
    COALESCE(SUM(sol.product_uom_qty) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_qty_sold,
    COALESCE(SUM(sol.price_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_revenue,
    COALESCE(AVG(sol.price_unit) FILTER (WHERE so.state IN ('sale', 'done')), 0) as avg_selling_price,
    -- Margin analysis
    p.standard_price as cost_price,
    p.list_price as list_price,
    (p.list_price - p.standard_price) as unit_margin,
    CASE
        WHEN p.standard_price > 0 THEN
            ((p.list_price - p.standard_price) / p.standard_price * 100)
        ELSE 0
    END as margin_percentage,
    COALESCE(SUM(sol.product_uom_qty * (sol.price_unit - p.standard_price)) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_margin,
    -- Inventory metrics
    CASE
        WHEN p.product_type = 'storable' THEN
            (SELECT COALESCE(SUM(quantity), 0) FROM stock_quants sq
             JOIN stock_locations sl ON sq.location_id = sl.id
             WHERE sq.product_id = p.id AND sl.usage = 'internal')
        ELSE NULL
    END as current_stock,
    -- Purchase metrics
    COUNT(DISTINCT pol.order_id) as times_purchased,
    COALESCE(AVG(pol.price_unit) FILTER (WHERE po.state IN ('purchase', 'done')), 0) as avg_purchase_price,
    -- Time metrics
    MIN(so.date_order) FILTER (WHERE so.state IN ('sale', 'done')) as first_sale_date,
    MAX(so.date_order) FILTER (WHERE so.state IN ('sale', 'done')) as last_sale_date,
    EXTRACT(EPOCH FROM (CURRENT_DATE - MAX(so.date_order)))::integer / 86400 as days_since_last_sale,
    -- Velocity
    CASE
        WHEN EXTRACT(EPOCH FROM (MAX(so.date_order) - MIN(so.date_order)))::integer / 86400 > 0 THEN
            COALESCE(SUM(sol.product_uom_qty) FILTER (WHERE so.state IN ('sale', 'done')), 0) /
            NULLIF(EXTRACT(EPOCH FROM (MAX(so.date_order) - MIN(so.date_order)))::integer / 86400, 0) * 30
        ELSE 0
    END as avg_monthly_sales,
    -- Performance rating
    CASE
        WHEN COALESCE(SUM(sol.price_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) >= 10000
             AND COUNT(DISTINCT sol.order_id) >= 10 THEN 'Star Product'
        WHEN COALESCE(SUM(sol.price_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) >= 5000 THEN 'High Performer'
        WHEN COUNT(DISTINCT sol.order_id) >= 5 THEN 'Regular Seller'
        WHEN COUNT(DISTINCT sol.order_id) >= 1 THEN 'Slow Mover'
        ELSE 'Dead Stock'
    END as performance_category
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN sales_order_lines sol ON sol.product_id = p.id AND sol.deleted_at IS NULL
LEFT JOIN sales_orders so ON so.id = sol.order_id AND so.deleted_at IS NULL
LEFT JOIN purchase_order_lines pol ON pol.product_id = p.id AND pol.deleted_at IS NULL
LEFT JOIN purchase_orders po ON po.id = pol.order_id AND po.deleted_at IS NULL
WHERE p.organization_id = get_current_user_organization_id()
  AND p.deleted_at IS NULL
GROUP BY p.id, p.name, p.default_code, pc.name, p.standard_price, p.list_price, p.product_type
ORDER BY total_revenue DESC;

COMMENT ON VIEW view_product_performance IS 'Comprehensive product performance with sales, margin, and velocity metrics';

-- Product Recommendations (Cross-sell/Up-sell)
CREATE OR REPLACE VIEW view_product_recommendations AS
WITH product_pairs AS (
    SELECT
        sol1.product_id as product_a_id,
        sol2.product_id as product_b_id,
        COUNT(DISTINCT sol1.order_id) as times_bought_together
    FROM sales_order_lines sol1
    JOIN sales_order_lines sol2 ON sol1.order_id = sol2.order_id AND sol1.product_id != sol2.product_id
    JOIN sales_orders so ON so.id = sol1.order_id
    WHERE so.state IN ('sale', 'done')
      AND sol1.deleted_at IS NULL
      AND sol2.deleted_at IS NULL
      AND so.deleted_at IS NULL
      AND so.organization_id = get_current_user_organization_id()
    GROUP BY sol1.product_id, sol2.product_id
    HAVING COUNT(DISTINCT sol1.order_id) >= 2
)
SELECT
    p1.id as product_id,
    p1.name as product_name,
    p1.default_code as product_code,
    jsonb_agg(
        jsonb_build_object(
            'recommended_product_id', p2.id,
            'recommended_product_name', p2.name,
            'recommended_product_code', p2.default_code,
            'times_bought_together', pp.times_bought_together,
            'confidence_score', ROUND((pp.times_bought_together::numeric /
                NULLIF((SELECT COUNT(DISTINCT order_id)
                        FROM sales_order_lines
                        WHERE product_id = p1.id), 0)) * 100, 2)
        ) ORDER BY pp.times_bought_together DESC
    ) FILTER (WHERE p2.id IS NOT NULL) as recommended_products
FROM products p1
LEFT JOIN product_pairs pp ON pp.product_a_id = p1.id
LEFT JOIN products p2 ON p2.id = pp.product_b_id
WHERE p1.organization_id = get_current_user_organization_id()
  AND p1.deleted_at IS NULL
  AND p1.sale_ok = true
GROUP BY p1.id, p1.name, p1.default_code
HAVING jsonb_agg(p2.id) FILTER (WHERE p2.id IS NOT NULL) IS NOT NULL;

COMMENT ON VIEW view_product_recommendations IS 'Product cross-sell and up-sell recommendations based on purchase patterns';

-- =====================================================
-- INVENTORY INSIGHTS
-- =====================================================

-- Stock Movement Analysis
CREATE OR REPLACE VIEW view_stock_movement_analysis AS
SELECT
    p.id as product_id,
    p.name as product_name,
    p.default_code,
    pc.name as category_name,
    -- Current inventory
    COALESCE(SUM(sq.quantity), 0) as current_stock,
    COALESCE(SUM(sq.reserved_quantity), 0) as reserved_stock,
    COALESCE(SUM(sq.quantity - sq.reserved_quantity), 0) as available_stock,
    -- Movement metrics (last 90 days)
    (
        SELECT COUNT(*)
        FROM stock_moves sm
        WHERE sm.product_id = p.id
          AND sm.state = 'done'
          AND sm.date >= CURRENT_DATE - INTERVAL '90 days'
    ) as total_moves_90d,
    (
        SELECT COALESCE(SUM(product_uom_qty), 0)
        FROM stock_moves sm
        JOIN stock_locations sl ON sm.location_dest_id = sl.id
        WHERE sm.product_id = p.id
          AND sm.state = 'done'
          AND sl.usage = 'internal'
          AND sm.date >= CURRENT_DATE - INTERVAL '90 days'
    ) as qty_received_90d,
    (
        SELECT COALESCE(SUM(product_uom_qty), 0)
        FROM stock_moves sm
        JOIN stock_locations sl ON sm.location_id = sl.id
        WHERE sm.product_id = p.id
          AND sm.state = 'done'
          AND sl.usage = 'internal'
          AND sm.date >= CURRENT_DATE - INTERVAL '90 days'
    ) as qty_shipped_90d,
    -- Turnover calculation
    CASE
        WHEN COALESCE(SUM(sq.quantity), 0) > 0 THEN
            (SELECT COALESCE(SUM(product_uom_qty), 0)
             FROM stock_moves sm
             JOIN stock_locations sl ON sm.location_id = sl.id
             WHERE sm.product_id = p.id
               AND sm.state = 'done'
               AND sl.usage = 'internal'
               AND sm.date >= CURRENT_DATE - INTERVAL '90 days') /
            NULLIF(COALESCE(SUM(sq.quantity), 0), 0) * 4 -- Annualized
        ELSE 0
    END as annual_turnover_rate,
    -- Stock days
    CASE
        WHEN (SELECT COALESCE(SUM(product_uom_qty), 0)
              FROM stock_moves sm
              JOIN stock_locations sl ON sm.location_id = sl.id
              WHERE sm.product_id = p.id
                AND sm.state = 'done'
                AND sl.usage = 'internal'
                AND sm.date >= CURRENT_DATE - INTERVAL '90 days') > 0 THEN
            COALESCE(SUM(sq.quantity), 0) /
            NULLIF((SELECT SUM(product_uom_qty)
                    FROM stock_moves sm
                    JOIN stock_locations sl ON sm.location_id = sl.id
                    WHERE sm.product_id = p.id
                      AND sm.state = 'done'
                      AND sl.usage = 'internal'
                      AND sm.date >= CURRENT_DATE - INTERVAL '90 days') / 90.0, 0)
        ELSE NULL
    END as days_of_stock,
    -- Stock status
    CASE
        WHEN COALESCE(SUM(sq.quantity - sq.reserved_quantity), 0) <= 0 THEN 'Out of Stock'
        WHEN COALESCE(SUM(sq.quantity - sq.reserved_quantity), 0) < 10 THEN 'Low Stock'
        WHEN COALESCE(SUM(sq.quantity - sq.reserved_quantity), 0) > 100 THEN 'Overstock'
        ELSE 'Normal'
    END as stock_status,
    -- Value
    COALESCE(SUM(sq.quantity), 0) * p.standard_price as inventory_value
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN stock_quants sq ON sq.product_id = p.id
LEFT JOIN stock_locations sl ON sq.location_id = sl.id AND sl.usage = 'internal'
WHERE p.organization_id = get_current_user_organization_id()
  AND p.product_type = 'storable'
  AND p.deleted_at IS NULL
GROUP BY p.id, p.name, p.default_code, pc.name, p.standard_price
ORDER BY inventory_value DESC;

COMMENT ON VIEW view_stock_movement_analysis IS 'Stock movement analysis with turnover rates and stock status';

-- Reorder Point Recommendations
CREATE OR REPLACE VIEW view_reorder_recommendations AS
WITH product_velocity AS (
    SELECT
        sm.product_id,
        -- Average daily consumption (last 90 days)
        COALESCE(SUM(sm.product_uom_qty) FILTER (
            WHERE sl.usage = 'internal' AND sm.state = 'done'
        ), 0) / 90.0 as avg_daily_consumption,
        -- Lead time (average days from PO to receipt)
        COALESCE(AVG(EXTRACT(EPOCH FROM (sp.date_done - po.date_order))::integer / 86400) FILTER (
            WHERE sp.state = 'done' AND po.state IN ('purchase', 'done')
        ), 7) as avg_lead_time_days
    FROM stock_moves sm
    JOIN stock_locations sl ON sm.location_id = sl.id
    LEFT JOIN stock_pickings sp ON sm.picking_id = sp.id
    LEFT JOIN purchase_order_lines pol ON pol.product_id = sm.product_id
    LEFT JOIN purchase_orders po ON po.id = pol.order_id
    WHERE sm.date >= CURRENT_DATE - INTERVAL '90 days'
      AND sm.organization_id = get_current_user_organization_id()
    GROUP BY sm.product_id
)
SELECT
    p.id as product_id,
    p.name as product_name,
    p.default_code,
    -- Current stock
    COALESCE((
        SELECT SUM(quantity - reserved_quantity)
        FROM stock_quants sq
        JOIN stock_locations sl ON sq.location_id = sl.id
        WHERE sq.product_id = p.id AND sl.usage = 'internal'
    ), 0) as current_available_stock,
    -- Velocity metrics
    pv.avg_daily_consumption,
    pv.avg_lead_time_days,
    -- Reorder calculations (assuming safety stock = 1.5x lead time demand)
    ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days) as lead_time_demand,
    ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days * 1.5) as safety_stock,
    ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days * 2.5) as reorder_point,
    ROUND(pv.avg_daily_consumption * 30) as recommended_order_qty,
    -- Days until stockout
    CASE
        WHEN pv.avg_daily_consumption > 0 THEN
            COALESCE((
                SELECT SUM(quantity - reserved_quantity)
                FROM stock_quants sq
                JOIN stock_locations sl ON sq.location_id = sl.id
                WHERE sq.product_id = p.id AND sl.usage = 'internal'
            ), 0) / NULLIF(pv.avg_daily_consumption, 0)
        ELSE NULL
    END as days_until_stockout,
    -- Action needed
    CASE
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days * 2.5) THEN 'Order Now'
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days * 3) THEN 'Monitor Closely'
        ELSE 'OK'
    END as action_required,
    -- Priority
    CASE
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= 0 THEN 'Critical'
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days) THEN 'High'
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days * 2.5) THEN 'Medium'
        ELSE 'Low'
    END as priority
FROM products p
JOIN product_velocity pv ON pv.product_id = p.id
WHERE p.organization_id = get_current_user_organization_id()
  AND p.product_type = 'storable'
  AND p.purchase_ok = true
  AND p.deleted_at IS NULL
  AND pv.avg_daily_consumption > 0
ORDER BY
    CASE
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= 0 THEN 1
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days) THEN 2
        WHEN COALESCE((
            SELECT SUM(quantity - reserved_quantity)
            FROM stock_quants sq
            JOIN stock_locations sl ON sq.location_id = sl.id
            WHERE sq.product_id = p.id AND sl.usage = 'internal'
        ), 0) <= ROUND(pv.avg_daily_consumption * pv.avg_lead_time_days * 2.5) THEN 3
        ELSE 4
    END,
    pv.avg_daily_consumption DESC;

COMMENT ON VIEW view_reorder_recommendations IS 'Intelligent reorder point recommendations based on consumption velocity and lead times';

-- =====================================================
-- FINANCIAL INSIGHTS
-- =====================================================

-- Cash Flow Forecast (30/60/90 days)
CREATE OR REPLACE VIEW view_cash_flow_forecast AS
WITH receivables AS (
    SELECT
        'Receivables' as flow_type,
        COALESCE(SUM(amount_residual) FILTER (
            WHERE invoice_date_due BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
        ), 0) as next_30_days,
        COALESCE(SUM(amount_residual) FILTER (
            WHERE invoice_date_due BETWEEN CURRENT_DATE + INTERVAL '31 days' AND CURRENT_DATE + INTERVAL '60 days'
        ), 0) as days_31_60,
        COALESCE(SUM(amount_residual) FILTER (
            WHERE invoice_date_due BETWEEN CURRENT_DATE + INTERVAL '61 days' AND CURRENT_DATE + INTERVAL '90 days'
        ), 0) as days_61_90
    FROM invoices
    WHERE move_type = 'out_invoice'
      AND state = 'posted'
      AND payment_state IN ('not_paid', 'partial')
      AND organization_id = get_current_user_organization_id()
),
payables AS (
    SELECT
        'Payables' as flow_type,
        -COALESCE(SUM(amount_residual) FILTER (
            WHERE invoice_date_due BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
        ), 0) as next_30_days,
        -COALESCE(SUM(amount_residual) FILTER (
            WHERE invoice_date_due BETWEEN CURRENT_DATE + INTERVAL '31 days' AND CURRENT_DATE + INTERVAL '60 days'
        ), 0) as days_31_60,
        -COALESCE(SUM(amount_residual) FILTER (
            WHERE invoice_date_due BETWEEN CURRENT_DATE + INTERVAL '61 days' AND CURRENT_DATE + INTERVAL '90 days'
        ), 0) as days_61_90
    FROM invoices
    WHERE move_type = 'in_invoice'
      AND state = 'posted'
      AND payment_state IN ('not_paid', 'partial')
      AND organization_id = get_current_user_organization_id()
)
SELECT
    flow_type,
    next_30_days,
    days_31_60,
    days_61_90,
    next_30_days + days_31_60 + days_61_90 as total_90_days
FROM receivables
UNION ALL
SELECT
    flow_type,
    next_30_days,
    days_31_60,
    days_61_90,
    next_30_days + days_31_60 + days_61_90 as total_90_days
FROM payables
UNION ALL
SELECT
    'Net Cash Flow' as flow_type,
    r.next_30_days + p.next_30_days as next_30_days,
    r.days_31_60 + p.days_31_60 as days_31_60,
    r.days_61_90 + p.days_61_90 as days_61_90,
    (r.next_30_days + p.next_30_days) + (r.days_31_60 + p.days_31_60) + (r.days_61_90 + p.days_61_90) as total_90_days
FROM receivables r, payables p;

COMMENT ON VIEW view_cash_flow_forecast IS 'Cash flow forecast for next 90 days based on receivables and payables';

-- Revenue Recognition Timeline
CREATE OR REPLACE VIEW view_revenue_recognition AS
WITH order_months AS (
    SELECT
        DATE_TRUNC('month', so.date_order) as revenue_month,
        COUNT(DISTINCT so.id) as order_count,
        -- Order values
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state = 'draft'), 0) as draft_value,
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state = 'sent'), 0) as quoted_value,
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as confirmed_value,
        -- Invoice status
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.invoice_status = 'to invoice'), 0) as to_invoice_value,
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.invoice_status = 'invoiced'), 0) as invoiced_value
    FROM sales_orders so
    WHERE so.organization_id = get_current_user_organization_id()
      AND so.deleted_at IS NULL
      AND so.date_order >= CURRENT_DATE - INTERVAL '12 months'
    GROUP BY DATE_TRUNC('month', so.date_order)
),
invoice_amounts AS (
    SELECT
        DATE_TRUNC('month', so.date_order) as revenue_month,
        COALESCE(SUM(i.amount_total), 0) as actual_invoiced,
        COALESCE(SUM(i.amount_total - i.amount_residual), 0) as collected_value
    FROM sales_orders so
    JOIN invoices i ON i.invoice_origin = so.name
    WHERE so.organization_id = get_current_user_organization_id()
      AND so.deleted_at IS NULL
      AND so.date_order >= CURRENT_DATE - INTERVAL '12 months'
      AND i.state = 'posted'
      AND i.move_type = 'out_invoice'
    GROUP BY DATE_TRUNC('month', so.date_order)
)
SELECT
    om.revenue_month,
    om.order_count,
    om.draft_value,
    om.quoted_value,
    om.confirmed_value,
    om.to_invoice_value,
    om.invoiced_value,
    COALESCE(ia.actual_invoiced, 0) as actual_invoiced,
    COALESCE(ia.collected_value, 0) as collected_value
FROM order_months om
LEFT JOIN invoice_amounts ia ON om.revenue_month = ia.revenue_month
ORDER BY om.revenue_month DESC;

COMMENT ON VIEW view_revenue_recognition IS 'Revenue recognition timeline from quote to collection';

-- =====================================================
-- OPERATIONAL INSIGHTS
-- =====================================================

-- Employee Productivity Metrics
CREATE OR REPLACE VIEW view_employee_productivity AS
SELECT
    e.id as employee_id,
    e.name as employee_name,
    e.employee_number,
    dept.name as department_name,
    jp.name as job_position,
    -- Timesheet metrics (last 30 days)
    COALESCE(SUM(ts.unit_amount) FILTER (
        WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
    ), 0) as hours_logged_30d,
    COUNT(DISTINCT ts.project_id) FILTER (
        WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
    ) as projects_worked_30d,
    COUNT(DISTINCT ts.task_id) FILTER (
        WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
    ) as tasks_worked_30d,
    -- Task completion
    COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'done') as tasks_completed_total,
    COUNT(DISTINCT t.id) FILTER (
        WHERE t.kanban_state = 'done'
        AND t.date_end >= CURRENT_DATE - INTERVAL '30 days'
    ) as tasks_completed_30d,
    -- Sales performance (if applicable)
    COUNT(DISTINCT so.id) FILTER (
        WHERE so.state IN ('sale', 'done')
        AND so.date_order >= CURRENT_DATE - INTERVAL '30 days'
    ) as sales_orders_30d,
    COALESCE(SUM(so.amount_total) FILTER (
        WHERE so.state IN ('sale', 'done')
        AND so.date_order >= CURRENT_DATE - INTERVAL '30 days'
    ), 0) as sales_revenue_30d,
    -- Utilization rate (assuming 160 hours/month)
    CASE
        WHEN COALESCE(SUM(ts.unit_amount) FILTER (
            WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
        ), 0) > 0 THEN
            COALESCE(SUM(ts.unit_amount) FILTER (
                WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
            ), 0) / 160.0 * 100
        ELSE 0
    END as utilization_rate,
    -- Performance score
    CASE
        WHEN COALESCE(SUM(ts.unit_amount) FILTER (
            WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
        ), 0) >= 160
        AND COUNT(DISTINCT t.id) FILTER (
            WHERE t.kanban_state = 'done'
            AND t.date_end >= CURRENT_DATE - INTERVAL '30 days'
        ) >= 10 THEN 'High Performer'
        WHEN COALESCE(SUM(ts.unit_amount) FILTER (
            WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
        ), 0) >= 120 THEN 'Good Performer'
        WHEN COALESCE(SUM(ts.unit_amount) FILTER (
            WHERE ts.date >= CURRENT_DATE - INTERVAL '30 days'
        ), 0) >= 80 THEN 'Average'
        ELSE 'Below Average'
    END as performance_level
FROM employees e
LEFT JOIN departments dept ON e.department_id = dept.id
LEFT JOIN job_positions jp ON e.job_id = jp.id
LEFT JOIN timesheets ts ON ts.employee_id = e.id AND ts.deleted_at IS NULL
LEFT JOIN tasks t ON t.user_ids @> ARRAY[e.user_id] AND t.deleted_at IS NULL
LEFT JOIN sales_orders so ON so.user_id = e.user_id AND so.deleted_at IS NULL
WHERE e.organization_id = get_current_user_organization_id()
  AND e.active = true
  AND e.deleted_at IS NULL
GROUP BY e.id, e.name, e.employee_number, dept.name, jp.name
ORDER BY hours_logged_30d DESC;

COMMENT ON VIEW view_employee_productivity IS 'Employee productivity metrics including utilization and task completion';

-- Project Health Dashboard
CREATE OR REPLACE VIEW view_project_health AS
SELECT
    proj.id as project_id,
    proj.name as project_name,
    proj.date_start,
    proj.date as project_deadline,
    partner.name as customer_name,
    -- Task metrics
    COUNT(DISTINCT t.id) as total_tasks,
    COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'done') as completed_tasks,
    COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'blocked') as blocked_tasks,
    COUNT(DISTINCT t.id) FILTER (
        WHERE t.date_deadline < CURRENT_DATE AND t.kanban_state != 'done'
    ) as overdue_tasks,
    -- Progress
    CASE
        WHEN COUNT(DISTINCT t.id) > 0 THEN
            COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'done')::numeric / COUNT(DISTINCT t.id) * 100
        ELSE 0
    END as completion_percentage,
    -- Time tracking
    COALESCE(SUM(t.planned_hours), 0) as total_planned_hours,
    COALESCE(SUM(t.effective_hours), 0) as total_actual_hours,
    COALESCE(SUM(t.remaining_hours), 0) as total_remaining_hours,
    -- Variance
    CASE
        WHEN COALESCE(SUM(t.planned_hours), 0) > 0 THEN
            (COALESCE(SUM(t.effective_hours), 0) - COALESCE(SUM(t.planned_hours), 0)) /
            NULLIF(COALESCE(SUM(t.planned_hours), 0), 0) * 100
        ELSE 0
    END as time_variance_percentage,
    -- Team size
    (
        SELECT COUNT(DISTINCT user_id)
        FROM tasks t2
        CROSS JOIN LATERAL unnest(t2.user_ids) AS user_id
        WHERE t2.project_id = proj.id
          AND t2.deleted_at IS NULL
    ) as team_member_count,
    -- Health score
    CASE
        WHEN COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'blocked')::numeric /
             NULLIF(COUNT(DISTINCT t.id), 0) > 0.3 THEN 'Critical'
        WHEN COUNT(DISTINCT t.id) FILTER (WHERE t.date_deadline < CURRENT_DATE AND t.kanban_state != 'done')::numeric /
             NULLIF(COUNT(DISTINCT t.id), 0) > 0.2 THEN 'At Risk'
        WHEN COALESCE(SUM(t.effective_hours), 0) > COALESCE(SUM(t.planned_hours), 0) * 1.2 THEN 'Over Budget'
        WHEN COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'done')::numeric /
             NULLIF(COUNT(DISTINCT t.id), 0) > 0.8 THEN 'On Track'
        ELSE 'Normal'
    END as health_status,
    -- Timeline status
    CASE
        WHEN proj.date IS NOT NULL AND proj.date < CURRENT_DATE THEN 'Overdue'
        WHEN proj.date IS NOT NULL AND proj.date < CURRENT_DATE + INTERVAL '7 days' THEN 'Due Soon'
        ELSE 'On Schedule'
    END as timeline_status
FROM projects proj
LEFT JOIN contacts partner ON proj.partner_id = partner.id
LEFT JOIN tasks t ON t.project_id = proj.id AND t.deleted_at IS NULL
WHERE proj.organization_id = get_current_user_organization_id()
  AND proj.active = true
  AND proj.deleted_at IS NULL
GROUP BY proj.id, proj.name, proj.date_start, proj.date, partner.name
ORDER BY
    CASE
        WHEN COUNT(DISTINCT t.id) FILTER (WHERE t.kanban_state = 'blocked')::numeric /
             NULLIF(COUNT(DISTINCT t.id), 0) > 0.3 THEN 1
        WHEN COUNT(DISTINCT t.id) FILTER (WHERE t.date_deadline < CURRENT_DATE AND t.kanban_state != 'done')::numeric /
             NULLIF(COUNT(DISTINCT t.id), 0) > 0.2 THEN 2
        ELSE 3
    END,
    proj.date;

COMMENT ON VIEW view_project_health IS 'Project health dashboard with task progress and timeline tracking';

-- =====================================================
-- ENABLE SECURITY AND GRANT PERMISSIONS
-- =====================================================

ALTER VIEW view_customer_lifetime_value SET (security_invoker = on);
ALTER VIEW view_customer_purchase_patterns SET (security_invoker = on);
ALTER VIEW view_customer_health_score SET (security_invoker = on);
ALTER VIEW view_sales_team_performance SET (security_invoker = on);
ALTER VIEW view_sales_pipeline_health SET (security_invoker = on);
ALTER VIEW view_product_performance SET (security_invoker = on);
ALTER VIEW view_product_recommendations SET (security_invoker = on);
ALTER VIEW view_stock_movement_analysis SET (security_invoker = on);
ALTER VIEW view_reorder_recommendations SET (security_invoker = on);
ALTER VIEW view_cash_flow_forecast SET (security_invoker = on);
ALTER VIEW view_revenue_recognition SET (security_invoker = on);
ALTER VIEW view_employee_productivity SET (security_invoker = on);
ALTER VIEW view_project_health SET (security_invoker = on);

GRANT SELECT ON view_customer_lifetime_value TO authenticated;
GRANT SELECT ON view_customer_purchase_patterns TO authenticated;
GRANT SELECT ON view_customer_health_score TO authenticated;
GRANT SELECT ON view_sales_team_performance TO authenticated;
GRANT SELECT ON view_sales_pipeline_health TO authenticated;
GRANT SELECT ON view_product_performance TO authenticated;
GRANT SELECT ON view_product_recommendations TO authenticated;
GRANT SELECT ON view_stock_movement_analysis TO authenticated;
GRANT SELECT ON view_reorder_recommendations TO authenticated;
GRANT SELECT ON view_cash_flow_forecast TO authenticated;
GRANT SELECT ON view_revenue_recognition TO authenticated;
GRANT SELECT ON view_employee_productivity TO authenticated;
GRANT SELECT ON view_project_health TO authenticated;

-- =====================================================
-- USAGE EXAMPLES
-- =====================================================

COMMENT ON SCHEMA public IS 'Advanced Analytics Views - Usage Examples:

-- Find VIP customers who need attention
SELECT * FROM view_customer_lifetime_value
WHERE customer_segment = ''VIP''
  AND customer_status IN (''At Risk'', ''Dormant'')
ORDER BY total_revenue DESC;

-- Get customer health scores
SELECT * FROM view_customer_health_score
WHERE health_status IN (''At Risk'', ''Critical'')
ORDER BY total_health_score ASC;

-- Compare sales team performance
SELECT * FROM view_sales_team_performance
ORDER BY total_revenue DESC;

-- Find stagnant pipeline stages
SELECT * FROM view_sales_pipeline_health
WHERE stage_health = ''Stagnant''
ORDER BY stagnant_leads DESC;

-- Top performing products
SELECT * FROM view_product_performance
WHERE performance_category IN (''Star Product'', ''High Performer'')
ORDER BY total_revenue DESC;

-- Products that need reordering
SELECT * FROM view_reorder_recommendations
WHERE action_required = ''Order Now''
ORDER BY priority, days_until_stockout;

-- Forecast cash flow
SELECT * FROM view_cash_flow_forecast;

-- Employee productivity rankings
SELECT * FROM view_employee_productivity
WHERE performance_level IN (''High Performer'', ''Good Performer'')
ORDER BY hours_logged_30d DESC;

-- Projects at risk
SELECT * FROM view_project_health
WHERE health_status IN (''Critical'', ''At Risk'')
ORDER BY overdue_tasks DESC;
';
