-- Migration: Operations Excellence Views
-- Description: Production efficiency, quality metrics, and fulfillment performance
-- Created: 2025-01-01
-- Module: Operations

-- =====================================================
-- PRODUCTION EFFICIENCY METRICS
-- =====================================================

CREATE OR REPLACE VIEW view_production_efficiency AS
WITH production_time_metrics AS (
    SELECT
        mo.id as manufacturing_order_id,
        p.id as product_id,
        p.name as product_name,
        mo.product_qty as planned_qty,
        mo.qty_producing as actual_qty,
        -- Time metrics
        EXTRACT(EPOCH FROM (mo.date_finished - mo.date_start)) / 3600.0 as actual_hours,
        EXTRACT(EPOCH FROM (mo.date_planned_finished - mo.date_planned_start)) / 3600.0 as planned_hours,
        -- Efficiency calculation
        CASE
            WHEN EXTRACT(EPOCH FROM (mo.date_finished - mo.date_start)) > 0 THEN
                (EXTRACT(EPOCH FROM (mo.date_planned_finished - mo.date_planned_start)) /
                 NULLIF(EXTRACT(EPOCH FROM (mo.date_finished - mo.date_start)), 0) * 100)
            ELSE 100
        END as time_efficiency_pct,
        -- Yield
        CASE
            WHEN mo.product_qty > 0 THEN
                (mo.qty_producing / NULLIF(mo.product_qty, 0) * 100)
            ELSE 100
        END as yield_pct,
        mo.state,
        mo.created_at
    FROM manufacturing_orders mo
    JOIN products p ON mo.product_id = p.id
    WHERE mo.organization_id = get_current_user_organization_id()
      AND mo.deleted_at IS NULL
      AND mo.state IN ('done', 'progress')
),
workcenter_metrics AS (
    SELECT
        wc.id as workcenter_id,
        wc.name as workcenter_name,
        COUNT(DISTINCT wo.id) as total_work_orders,
        COUNT(DISTINCT wo.id) FILTER (WHERE wo.state = 'done') as completed_work_orders,
        AVG(wo.duration) FILTER (WHERE wo.state = 'done') as avg_actual_duration,
        AVG(wo.duration_expected) as avg_expected_duration,
        -- Utilization
        CASE
            WHEN SUM(wo.duration_expected) > 0 THEN
                (SUM(wo.duration) / NULLIF(SUM(wo.duration_expected), 0) * 100)
            ELSE 0
        END as utilization_rate
    FROM workcenters wc
    LEFT JOIN work_orders wo ON wo.workcenter_id = wc.id
    WHERE wc.organization_id = get_current_user_organization_id()
      AND wc.active = true
    GROUP BY wc.id, wc.name
)
SELECT
    -- Overall metrics
    COUNT(DISTINCT ptm.manufacturing_order_id) as total_production_orders,
    COUNT(DISTINCT ptm.manufacturing_order_id) FILTER (WHERE ptm.state = 'done') as completed_orders,
    -- Time efficiency
    ROUND(AVG(ptm.time_efficiency_pct)::numeric, 1) as avg_time_efficiency_pct,
    ROUND(AVG(ptm.yield_pct)::numeric, 1) as avg_yield_pct,
    -- OEE components (simplified)
    ROUND((
        AVG(ptm.time_efficiency_pct) *
        AVG(ptm.yield_pct) *
        100 / 10000
    )::numeric, 1) as overall_equipment_effectiveness,
    -- Scrap rate (100 - yield)
    ROUND((100 - AVG(ptm.yield_pct))::numeric, 1) as avg_scrap_rate_pct,
    -- Top/bottom performers by product
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'product_name', product_name,
                'avg_efficiency', ROUND(AVG(time_efficiency_pct)::numeric, 1),
                'avg_yield', ROUND(AVG(yield_pct)::numeric, 1)
            ) ORDER BY AVG(time_efficiency_pct) DESC
        )
        FROM (
            SELECT
                product_name,
                AVG(time_efficiency_pct) as avg_time_eff,
                AVG(yield_pct)
            FROM production_time_metrics
            GROUP BY product_name
            ORDER BY AVG(time_efficiency_pct) DESC
            LIMIT 5
        ) top_products
    ) as top_performing_products,
    -- Workcenter utilization
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'workcenter', workcenter_name,
                'utilization_rate', ROUND(utilization_rate::numeric, 1),
                'completed_orders', completed_work_orders
            ) ORDER BY utilization_rate DESC
        )
        FROM workcenter_metrics
        WHERE total_work_orders > 0
    ) as workcenter_performance,
    -- Performance rating
    CASE
        WHEN AVG(ptm.time_efficiency_pct) >= 90 AND AVG(ptm.yield_pct) >= 95 THEN 'Excellent'
        WHEN AVG(ptm.time_efficiency_pct) >= 80 AND AVG(ptm.yield_pct) >= 90 THEN 'Good'
        WHEN AVG(ptm.time_efficiency_pct) >= 70 AND AVG(ptm.yield_pct) >= 85 THEN 'Average'
        ELSE 'Needs Improvement'
    END as overall_performance_rating
FROM production_time_metrics ptm
WHERE ptm.created_at >= CURRENT_DATE - INTERVAL '90 days';

COMMENT ON VIEW view_production_efficiency IS 'Manufacturing efficiency metrics including OEE, yield, and workcenter utilization';

-- =====================================================
-- QUALITY METRICS
-- =====================================================

CREATE OR REPLACE VIEW view_quality_metrics AS
WITH product_returns AS (
    SELECT
        p.id as product_id,
        p.name as product_name,
        pc.name as category_name,
        -- Sales and returns
        COUNT(DISTINCT sol.order_id) as times_sold,
        COALESCE(SUM(sol.product_uom_qty), 0) as total_qty_sold,
        -- Returns (from credit notes)
        COUNT(DISTINCT il.move_id) FILTER (WHERE i.move_type = 'out_refund') as return_count,
        COALESCE(SUM(il.quantity) FILTER (WHERE i.move_type = 'out_refund'), 0) as total_qty_returned,
        -- Return rate
        CASE
            WHEN COALESCE(SUM(sol.product_uom_qty), 0) > 0 THEN
                (COALESCE(SUM(il.quantity) FILTER (WHERE i.move_type = 'out_refund'), 0) /
                 NULLIF(COALESCE(SUM(sol.product_uom_qty), 0), 0) * 100)
            ELSE 0
        END as return_rate_pct
    FROM products p
    LEFT JOIN product_categories pc ON p.category_id = pc.id
    LEFT JOIN sales_order_lines sol ON sol.product_id = p.id AND sol.deleted_at IS NULL
    LEFT JOIN sales_orders so ON so.id = sol.order_id AND so.deleted_at IS NULL AND so.state IN ('sale', 'done')
    LEFT JOIN invoice_lines il ON il.product_id = p.id AND il.deleted_at IS NULL
    LEFT JOIN invoices i ON i.id = il.move_id AND i.deleted_at IS NULL AND i.state = 'posted'
    WHERE p.organization_id = get_current_user_organization_id()
      AND p.deleted_at IS NULL
      AND p.product_type IN ('storable', 'consumable')
    GROUP BY p.id, p.name, pc.name
),
supplier_quality AS (
    SELECT
        c.id as supplier_id,
        c.name as supplier_name,
        COUNT(DISTINCT po.id) as purchase_orders,
        COUNT(DISTINCT pol.id) as items_ordered,
        -- Quality issues would typically be tracked in a quality table
        -- This is a simplified version using returns to vendor
        COUNT(DISTINCT i.id) FILTER (WHERE i.move_type = 'in_refund') as vendor_returns,
        COALESCE(SUM(i.amount_total) FILTER (WHERE i.move_type = 'in_refund'), 0) as return_value
    FROM contacts c
    LEFT JOIN purchase_orders po ON po.partner_id = c.id AND po.deleted_at IS NULL
    LEFT JOIN purchase_order_lines pol ON pol.order_id = po.id AND pol.deleted_at IS NULL
    LEFT JOIN invoices i ON i.partner_id = c.id AND i.deleted_at IS NULL AND i.state = 'posted'
    WHERE c.organization_id = get_current_user_organization_id()
      AND c.is_vendor = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name
)
SELECT
    -- Product quality metrics
    COUNT(DISTINCT pr.product_id) as total_products_tracked,
    ROUND(AVG(pr.return_rate_pct)::numeric, 2) as avg_product_return_rate,
    -- Top quality issues by product
    (
        SELECT jsonb_agg(prod_obj)
        FROM (
            SELECT
                jsonb_build_object(
                    'product_name', product_name,
                    'return_rate', ROUND(return_rate_pct::numeric, 2),
                    'qty_returned', total_qty_returned
                ) as prod_obj
            FROM product_returns
            WHERE return_rate_pct > 0
            ORDER BY return_rate_pct DESC
            LIMIT 10
        ) top_issues
    ) as products_with_quality_issues,
    -- Supplier quality
    (
        SELECT jsonb_agg(supp_obj)
        FROM (
            SELECT
                jsonb_build_object(
                    'supplier_name', supplier_name,
                    'vendor_returns', vendor_returns,
                    'return_value', return_value
                ) as supp_obj
            FROM supplier_quality
            WHERE vendor_returns > 0
            ORDER BY vendor_returns DESC
            LIMIT 10
        ) top_suppliers
    ) as suppliers_with_issues,
    -- Quality rating
    CASE
        WHEN AVG(pr.return_rate_pct) < 1 THEN 'Excellent'
        WHEN AVG(pr.return_rate_pct) < 3 THEN 'Good'
        WHEN AVG(pr.return_rate_pct) < 5 THEN 'Average'
        ELSE 'Poor'
    END as overall_quality_rating,
    -- Cost of poor quality (returns value)
    (
        SELECT COALESCE(SUM(i.amount_total), 0)
        FROM invoices i
        WHERE i.organization_id = get_current_user_organization_id()
          AND i.move_type = 'out_refund'
          AND i.state = 'posted'
          AND i.deleted_at IS NULL
          AND i.date >= CURRENT_DATE - INTERVAL '12 months'
    ) as cost_of_poor_quality_last_12m
FROM product_returns pr;

COMMENT ON VIEW view_quality_metrics IS 'Product and supplier quality metrics with defect rates and cost of poor quality';

-- =====================================================
-- FULFILLMENT PERFORMANCE
-- =====================================================

CREATE OR REPLACE VIEW view_fulfillment_performance AS
WITH order_fulfillment_metrics AS (
    SELECT
        so.id as order_id,
        so.name as order_number,
        so.date_order,
        so.confirmation_date,
        c.name as customer_name,
        -- Fulfillment time
        EXTRACT(EPOCH FROM (sp.date_done - so.confirmation_date))::integer / 86400 as fulfillment_days,
        -- On-time delivery
        CASE
            WHEN sp.date_done IS NOT NULL AND sp.date_deadline IS NOT NULL THEN
                CASE WHEN sp.date_done <= sp.date_deadline THEN 1 ELSE 0 END
            ELSE NULL
        END as is_on_time,
        -- Order accuracy (simplified - assumes all picked = accurate)
        CASE
            WHEN sp.state = 'done' THEN 1
            ELSE 0
        END as is_complete,
        sp.state as shipment_state,
        so.state as order_state
    FROM sales_orders so
    JOIN contacts c ON so.partner_id = c.id
    LEFT JOIN stock_pickings sp ON sp.origin = so.name AND sp.deleted_at IS NULL
    WHERE so.organization_id = get_current_user_organization_id()
      AND so.deleted_at IS NULL
      AND so.state IN ('sale', 'done')
      AND so.date_order >= CURRENT_DATE - INTERVAL '90 days'
),
backorder_metrics AS (
    SELECT
        COUNT(DISTINCT so.id) as total_orders_with_backorders,
        COUNT(DISTINCT sol.id) as total_backordered_lines
    FROM sales_orders so
    JOIN sales_order_lines sol ON sol.order_id = so.id AND sol.deleted_at IS NULL
    WHERE so.organization_id = get_current_user_organization_id()
      AND so.deleted_at IS NULL
      AND so.delivery_status = 'no'
      AND so.state IN ('sale', 'done')
),
shipping_cost_metrics AS (
    SELECT
        AVG(
            COALESCE(
                (so.custom_fields->>'shipping_cost')::numeric,
                0
            )
        ) as avg_shipping_cost_per_order
    FROM sales_orders so
    WHERE so.organization_id = get_current_user_organization_id()
      AND so.deleted_at IS NULL
      AND so.state IN ('sale', 'done')
      AND so.date_order >= CURRENT_DATE - INTERVAL '90 days'
)
SELECT
    -- Order fulfillment metrics
    COUNT(DISTINCT ofm.order_id) as total_orders,
    COUNT(DISTINCT ofm.order_id) FILTER (WHERE ofm.shipment_state = 'done') as completed_shipments,
    -- On-time delivery
    ROUND((
        COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
        NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
    )::numeric, 1) as on_time_delivery_rate,
    -- Average fulfillment time
    ROUND(AVG(ofm.fulfillment_days)::numeric, 1) as avg_fulfillment_days,
    -- Perfect order rate (on-time + complete)
    ROUND((
        COUNT(*) FILTER (WHERE ofm.is_on_time = 1 AND ofm.is_complete = 1)::numeric /
        NULLIF(COUNT(*), 0) * 100
    )::numeric, 1) as perfect_order_rate,
    -- Backorder metrics
    (SELECT total_orders_with_backorders FROM backorder_metrics) as orders_with_backorders,
    ROUND((
        (SELECT total_orders_with_backorders FROM backorder_metrics)::numeric /
        NULLIF(COUNT(DISTINCT ofm.order_id), 0) * 100
    )::numeric, 1) as backorder_rate,
    -- Shipping costs
    ROUND((SELECT avg_shipping_cost_per_order FROM shipping_cost_metrics)::numeric, 2) as avg_shipping_cost,
    -- Performance rating
    CASE
        WHEN (
            COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
            NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
        ) >= 95 THEN 'Excellent'
        WHEN (
            COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
            NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
        ) >= 90 THEN 'Good'
        WHEN (
            COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
            NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
        ) >= 80 THEN 'Average'
        ELSE 'Needs Improvement'
    END as delivery_performance_rating,
    -- Color coding
    CASE
        WHEN (
            COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
            NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
        ) >= 95 THEN 'green'
        WHEN (
            COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
            NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
        ) >= 90 THEN 'blue'
        WHEN (
            COUNT(*) FILTER (WHERE ofm.is_on_time = 1)::numeric /
            NULLIF(COUNT(*) FILTER (WHERE ofm.is_on_time IS NOT NULL), 0) * 100
        ) >= 80 THEN 'yellow'
        ELSE 'red'
    END as performance_color
FROM order_fulfillment_metrics ofm;

COMMENT ON VIEW view_fulfillment_performance IS 'Order fulfillment metrics: on-time delivery, perfect order rate, and shipping costs';

-- =====================================================
-- ENABLE SECURITY AND GRANT PERMISSIONS
-- =====================================================

ALTER VIEW view_production_efficiency SET (security_invoker = on);
ALTER VIEW view_quality_metrics SET (security_invoker = on);
ALTER VIEW view_fulfillment_performance SET (security_invoker = on);

GRANT SELECT ON view_production_efficiency TO authenticated;
GRANT SELECT ON view_quality_metrics TO authenticated;
GRANT SELECT ON view_fulfillment_performance TO authenticated;
