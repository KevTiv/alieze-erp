-- Migration: Financial Intelligence Views
-- Description: Profitability analysis, working capital metrics, revenue concentration, and subscription metrics
-- Created: 2025-01-01
-- Module: Financial Intelligence

-- =====================================================
-- CUSTOMER PROFITABILITY ANALYSIS
-- =====================================================

CREATE OR REPLACE VIEW view_profitability_by_customer AS
WITH customer_revenue AS (
    SELECT
        c.id as customer_id,
        c.name as customer_name,
        -- Revenue
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_revenue,
        COALESCE(SUM(so.amount_discount), 0) as total_discounts_given,
        COUNT(DISTINCT so.id) FILTER (WHERE so.state IN ('sale', 'done')) as order_count
    FROM contacts c
    LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL
    WHERE c.organization_id = get_current_user_organization_id()
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name
),
customer_cogs AS (
    SELECT
        c.id as customer_id,
        -- COGS from products sold
        COALESCE(SUM(sol.product_uom_qty * p.standard_price), 0) as total_cogs
    FROM contacts c
    LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL AND so.state IN ('sale', 'done')
    LEFT JOIN sales_order_lines sol ON sol.order_id = so.id AND sol.deleted_at IS NULL
    LEFT JOIN products p ON sol.product_id = p.id
    WHERE c.organization_id = get_current_user_organization_id()
    GROUP BY c.id
),
customer_service_costs AS (
    SELECT
        c.id as customer_id,
        -- Service/support costs from timesheets (assuming $100/hour default rate)
        COALESCE(SUM(ts.unit_amount * 100), 0) as total_service_costs,
        COUNT(DISTINCT ts.id) as service_hours_count
    FROM contacts c
    LEFT JOIN projects proj ON proj.partner_id = c.id AND proj.deleted_at IS NULL
    LEFT JOIN tasks t ON t.project_id = proj.id AND t.deleted_at IS NULL
    LEFT JOIN timesheets ts ON ts.task_id = t.id AND ts.deleted_at IS NULL
    WHERE c.organization_id = get_current_user_organization_id()
    GROUP BY c.id
),
customer_returns AS (
    SELECT
        c.id as customer_id,
        -- Returns and refunds
        COALESCE(SUM(i.amount_total) FILTER (WHERE i.move_type = 'out_refund'), 0) as total_refunds
    FROM contacts c
    LEFT JOIN invoices i ON i.partner_id = c.id AND i.deleted_at IS NULL AND i.state = 'posted'
    WHERE c.organization_id = get_current_user_organization_id()
    GROUP BY c.id
)
SELECT
    cr.customer_id,
    cr.customer_name,
    cr.order_count,
    -- Revenue breakdown
    cr.total_revenue,
    cr.total_discounts_given,
    (cr.total_revenue - cr.total_discounts_given) as net_revenue,
    -- Cost breakdown
    COALESCE(cc.total_cogs, 0) as product_costs,
    COALESCE(csc.total_service_costs, 0) as service_costs,
    COALESCE(cret.total_refunds, 0) as refund_costs,
    (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0)) as total_costs,
    -- Profitability
    ((cr.total_revenue - cr.total_discounts_given) -
     (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) as net_profit,
    -- Profit margin
    CASE
        WHEN (cr.total_revenue - cr.total_discounts_given) > 0 THEN
            (((cr.total_revenue - cr.total_discounts_given) -
              (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) /
             NULLIF((cr.total_revenue - cr.total_discounts_given), 0) * 100)
        ELSE 0
    END as profit_margin_percentage,
    -- Profitability rating
    CASE
        WHEN ((cr.total_revenue - cr.total_discounts_given) -
              (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) < 0
             THEN 'Unprofitable'
        WHEN (((cr.total_revenue - cr.total_discounts_given) -
               (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) /
              NULLIF((cr.total_revenue - cr.total_discounts_given), 0) * 100) >= 40
             THEN 'Highly Profitable'
        WHEN (((cr.total_revenue - cr.total_discounts_given) -
               (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) /
              NULLIF((cr.total_revenue - cr.total_discounts_given), 0) * 100) >= 20
             THEN 'Profitable'
        WHEN (((cr.total_revenue - cr.total_discounts_given) -
               (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) /
              NULLIF((cr.total_revenue - cr.total_discounts_given), 0) * 100) >= 10
             THEN 'Marginally Profitable'
        ELSE 'Low Profit'
    END as profitability_tier,
    -- Color coding
    CASE
        WHEN ((cr.total_revenue - cr.total_discounts_given) -
              (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) < 0
             THEN 'red'
        WHEN (((cr.total_revenue - cr.total_discounts_given) -
               (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) /
              NULLIF((cr.total_revenue - cr.total_discounts_given), 0) * 100) >= 40
             THEN 'green'
        WHEN (((cr.total_revenue - cr.total_discounts_given) -
               (COALESCE(cc.total_cogs, 0) + COALESCE(csc.total_service_costs, 0) + COALESCE(cret.total_refunds, 0))) /
              NULLIF((cr.total_revenue - cr.total_discounts_given), 0) * 100) >= 20
             THEN 'blue'
        ELSE 'yellow'
    END as profitability_color
FROM customer_revenue cr
LEFT JOIN customer_cogs cc ON cc.customer_id = cr.customer_id
LEFT JOIN customer_service_costs csc ON csc.customer_id = cr.customer_id
LEFT JOIN customer_returns cret ON cret.customer_id = cr.customer_id
WHERE cr.total_revenue > 0
ORDER BY net_profit DESC;

COMMENT ON VIEW view_profitability_by_customer IS 'True customer profitability including revenue, COGS, service costs, and returns';

-- =====================================================
-- WORKING CAPITAL METRICS
-- =====================================================

CREATE OR REPLACE VIEW view_working_capital_metrics AS
WITH receivables_metrics AS (
    SELECT
        -- Days Sales Outstanding (DSO)
        CASE
            WHEN COALESCE(SUM(amount_total) FILTER (WHERE date >= CURRENT_DATE - INTERVAL '90 days'), 0) > 0 THEN
                (COALESCE(SUM(amount_residual) FILTER (WHERE payment_state IN ('not_paid', 'partial')), 0) /
                 (COALESCE(SUM(amount_total) FILTER (WHERE date >= CURRENT_DATE - INTERVAL '90 days'), 0) / 90.0))
            ELSE 0
        END as dso_days,
        COALESCE(SUM(amount_residual) FILTER (WHERE payment_state IN ('not_paid', 'partial')), 0) as total_ar
    FROM invoices
    WHERE organization_id = get_current_user_organization_id()
      AND move_type = 'out_invoice'
      AND state = 'posted'
      AND deleted_at IS NULL
),
inventory_metrics AS (
    SELECT
        -- Days Inventory Outstanding (DIO)
        CASE
            WHEN COALESCE(SUM(sm.product_uom_qty * p.standard_price) FILTER (
                WHERE sm.date >= CURRENT_DATE - INTERVAL '90 days'
                AND sl_src.usage = 'internal'
            ), 0) > 0 THEN
                (COALESCE(SUM(sq.quantity * p.standard_price), 0) /
                 (COALESCE(SUM(sm.product_uom_qty * p.standard_price) FILTER (
                     WHERE sm.date >= CURRENT_DATE - INTERVAL '90 days'
                     AND sl_src.usage = 'internal'
                 ), 0) / 90.0))
            ELSE 0
        END as dio_days,
        COALESCE(SUM(sq.quantity * p.standard_price), 0) as total_inventory_value
    FROM stock_quants sq
    JOIN products p ON sq.product_id = p.id
    JOIN stock_locations sl ON sq.location_id = sl.id AND sl.usage = 'internal'
    LEFT JOIN stock_moves sm ON sm.product_id = p.id AND sm.state = 'done' AND sm.deleted_at IS NULL
    LEFT JOIN stock_locations sl_src ON sm.location_id = sl_src.id
    WHERE sq.organization_id = get_current_user_organization_id()
),
payables_metrics AS (
    SELECT
        -- Days Payable Outstanding (DPO)
        CASE
            WHEN COALESCE(SUM(amount_total) FILTER (WHERE date >= CURRENT_DATE - INTERVAL '90 days'), 0) > 0 THEN
                (COALESCE(SUM(amount_residual) FILTER (WHERE payment_state IN ('not_paid', 'partial')), 0) /
                 (COALESCE(SUM(amount_total) FILTER (WHERE date >= CURRENT_DATE - INTERVAL '90 days'), 0) / 90.0))
            ELSE 0
        END as dpo_days,
        COALESCE(SUM(amount_residual) FILTER (WHERE payment_state IN ('not_paid', 'partial')), 0) as total_ap
    FROM invoices
    WHERE organization_id = get_current_user_organization_id()
      AND move_type = 'in_invoice'
      AND state = 'posted'
      AND deleted_at IS NULL
)
SELECT
    ROUND(rm.dso_days::numeric, 1) as days_sales_outstanding,
    ROUND(im.dio_days::numeric, 1) as days_inventory_outstanding,
    ROUND(pm.dpo_days::numeric, 1) as days_payable_outstanding,
    -- Cash Conversion Cycle
    ROUND((rm.dso_days + im.dio_days - pm.dpo_days)::numeric, 1) as cash_conversion_cycle_days,
    -- Supporting values
    rm.total_ar as accounts_receivable,
    im.total_inventory_value,
    pm.total_ap as accounts_payable,
    (rm.total_ar + im.total_inventory_value - pm.total_ap) as working_capital,
    -- Performance indicators
    CASE
        WHEN (rm.dso_days + im.dio_days - pm.dpo_days) < 30 THEN 'Excellent'
        WHEN (rm.dso_days + im.dio_days - pm.dpo_days) < 60 THEN 'Good'
        WHEN (rm.dso_days + im.dio_days - pm.dpo_days) < 90 THEN 'Average'
        ELSE 'Needs Improvement'
    END as ccc_rating,
    CASE
        WHEN rm.dso_days < 30 THEN 'Excellent'
        WHEN rm.dso_days < 45 THEN 'Good'
        WHEN rm.dso_days < 60 THEN 'Average'
        ELSE 'Slow'
    END as collection_efficiency,
    CASE
        WHEN im.dio_days < 45 THEN 'Fast Turnover'
        WHEN im.dio_days < 90 THEN 'Normal Turnover'
        WHEN im.dio_days < 180 THEN 'Slow Turnover'
        ELSE 'Very Slow'
    END as inventory_efficiency,
    CASE
        WHEN pm.dpo_days > 45 THEN 'Optimized'
        WHEN pm.dpo_days > 30 THEN 'Good'
        ELSE 'Could Extend'
    END as payment_optimization
FROM receivables_metrics rm, inventory_metrics im, payables_metrics pm;

COMMENT ON VIEW view_working_capital_metrics IS 'Working capital efficiency metrics: DSO, DIO, DPO, and Cash Conversion Cycle';

-- =====================================================
-- REVENUE CONCENTRATION RISK
-- =====================================================

CREATE OR REPLACE VIEW view_revenue_concentration_risk AS
WITH customer_revenue AS (
    SELECT
        c.id as customer_id,
        c.name as customer_name,
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as customer_revenue,
        c.industry_id
    FROM contacts c
    LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL
    WHERE c.organization_id = get_current_user_organization_id()
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name, c.industry_id
),
total_revenue AS (
    SELECT SUM(customer_revenue) as total
    FROM customer_revenue
),
top_customers AS (
    SELECT
        cr.*,
        tr.total as total_company_revenue,
        (cr.customer_revenue / NULLIF(tr.total, 0) * 100) as revenue_percentage,
        ROW_NUMBER() OVER (ORDER BY cr.customer_revenue DESC) as revenue_rank
    FROM customer_revenue cr, total_revenue tr
),
industry_concentration AS (
    SELECT
        i.id as industry_id,
        i.name as industry_name,
        COUNT(DISTINCT cr.customer_id) as customer_count,
        SUM(cr.customer_revenue) as industry_revenue,
        (SUM(cr.customer_revenue) / NULLIF((SELECT total FROM total_revenue), 0) * 100) as industry_percentage
    FROM customer_revenue cr
    LEFT JOIN industries i ON cr.industry_id = i.id
    GROUP BY i.id, i.name
)
SELECT
    -- Top 10 customers
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'customer_name', customer_name,
                'revenue', customer_revenue,
                'percentage', ROUND(revenue_percentage::numeric, 2)
            ) ORDER BY revenue_rank
        )
        FROM top_customers
        WHERE revenue_rank <= 10
    ) as top_10_customers,
    -- Concentration metrics
    (SELECT SUM(revenue_percentage) FROM top_customers WHERE revenue_rank <= 5) as top_5_concentration_pct,
    (SELECT SUM(revenue_percentage) FROM top_customers WHERE revenue_rank <= 10) as top_10_concentration_pct,
    -- Herfindahl-Hirschman Index (HHI) - sum of squared market shares
    (
        SELECT SUM(POWER(revenue_percentage, 2))
        FROM top_customers
    ) as hhi_index,
    -- Industry concentration
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'industry', industry_name,
                'revenue', industry_revenue,
                'percentage', ROUND(industry_percentage::numeric, 2)
            ) ORDER BY industry_revenue DESC
        )
        FROM industry_concentration
        WHERE industry_revenue > 0
        LIMIT 5
    ) as top_industries,
    -- Risk assessment
    CASE
        WHEN (SELECT SUM(revenue_percentage) FROM top_customers WHERE revenue_rank <= 5) > 50 THEN 'High Risk'
        WHEN (SELECT SUM(revenue_percentage) FROM top_customers WHERE revenue_rank <= 5) > 30 THEN 'Moderate Risk'
        ELSE 'Low Risk'
    END as concentration_risk_level,
    CASE
        WHEN (SELECT SUM(POWER(revenue_percentage, 2)) FROM top_customers) > 2500 THEN 'Highly Concentrated'
        WHEN (SELECT SUM(POWER(revenue_percentage, 2)) FROM top_customers) > 1500 THEN 'Moderately Concentrated'
        ELSE 'Diversified'
    END as diversification_status,
    -- Total revenue for reference
    (SELECT total FROM total_revenue) as total_revenue;

COMMENT ON VIEW view_revenue_concentration_risk IS 'Revenue concentration analysis with HHI index and risk assessment';

-- =====================================================
-- SUBSCRIPTION METRICS (for recurring revenue)
-- =====================================================

CREATE OR REPLACE VIEW view_subscription_metrics AS
WITH subscription_customers AS (
    -- Assuming recurring products or orders are identified
    -- This is a simplified version - adjust based on your subscription model
    SELECT
        c.id as customer_id,
        c.name as customer_name,
        c.created_at as customer_since,
        MIN(so.date_order) FILTER (WHERE so.state IN ('sale', 'done')) as first_purchase_date,
        MAX(so.date_order) FILTER (WHERE so.state IN ('sale', 'done')) as last_purchase_date,
        COUNT(DISTINCT so.id) FILTER (WHERE so.state IN ('sale', 'done')) as purchase_count,
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as total_lifetime_value,
        -- Monthly recurring revenue (if you have recurring orders)
        COALESCE(AVG(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as avg_order_value
    FROM contacts c
    LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL
    WHERE c.organization_id = get_current_user_organization_id()
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name, c.created_at
),
monthly_metrics AS (
    SELECT
        DATE_TRUNC('month', so.date_order) as month,
        COUNT(DISTINCT so.partner_id) FILTER (WHERE so.state IN ('sale', 'done')) as active_customers,
        COUNT(DISTINCT so.partner_id) FILTER (
            WHERE so.state IN ('sale', 'done')
            AND NOT EXISTS (
                SELECT 1 FROM sales_orders so2
                WHERE so2.partner_id = so.partner_id
                AND so2.date_order < so.date_order
                AND so2.state IN ('sale', 'done')
                AND so2.deleted_at IS NULL
            )
        ) as new_customers,
        COALESCE(SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')), 0) as monthly_revenue
    FROM sales_orders so
    WHERE so.organization_id = get_current_user_organization_id()
      AND so.deleted_at IS NULL
      AND so.date_order >= CURRENT_DATE - INTERVAL '12 months'
    GROUP BY DATE_TRUNC('month', so.date_order)
),
current_month AS (
    SELECT * FROM monthly_metrics WHERE month = DATE_TRUNC('month', CURRENT_DATE)
),
previous_month AS (
    SELECT * FROM monthly_metrics WHERE month = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
)
SELECT
    -- Current MRR (Monthly Recurring Revenue)
    (SELECT monthly_revenue FROM current_month) as current_mrr,
    -- Annual Recurring Revenue
    ((SELECT monthly_revenue FROM current_month) * 12) as arr,
    -- Customer metrics
    (SELECT active_customers FROM current_month) as active_customers,
    (SELECT new_customers FROM current_month) as new_customers_this_month,
    -- Retention metrics
    CASE
        WHEN (SELECT active_customers FROM previous_month) > 0 THEN
            (((SELECT active_customers FROM current_month)::numeric -
              (SELECT new_customers FROM current_month)) /
             (SELECT active_customers FROM previous_month) * 100)
        ELSE 100
    END as gross_retention_rate,
    -- Revenue retention
    CASE
        WHEN (SELECT monthly_revenue FROM previous_month) > 0 THEN
            ((SELECT monthly_revenue FROM current_month)::numeric /
             (SELECT monthly_revenue FROM previous_month) * 100)
        ELSE 100
    END as revenue_retention_rate,
    -- Average metrics
    (
        SELECT AVG(total_lifetime_value)
        FROM subscription_customers
        WHERE purchase_count > 1
    ) as avg_ltv,
    -- Growth rate
    CASE
        WHEN (SELECT monthly_revenue FROM previous_month) > 0 THEN
            (((SELECT monthly_revenue FROM current_month) - (SELECT monthly_revenue FROM previous_month))::numeric /
             (SELECT monthly_revenue FROM previous_month) * 100)
        ELSE 0
    END as mrr_growth_rate,
    -- Customer churn (simplified)
    CASE
        WHEN (SELECT active_customers FROM previous_month) > 0 THEN
            (((SELECT active_customers FROM previous_month) -
              ((SELECT active_customers FROM current_month) - (SELECT new_customers FROM current_month)))::numeric /
             (SELECT active_customers FROM previous_month) * 100)
        ELSE 0
    END as customer_churn_rate;

COMMENT ON VIEW view_subscription_metrics IS 'Subscription business metrics: MRR, ARR, retention, churn, and LTV';

-- =====================================================
-- ENABLE SECURITY AND GRANT PERMISSIONS
-- =====================================================

ALTER VIEW view_profitability_by_customer SET (security_invoker = on);
ALTER VIEW view_working_capital_metrics SET (security_invoker = on);
ALTER VIEW view_revenue_concentration_risk SET (security_invoker = on);
ALTER VIEW view_subscription_metrics SET (security_invoker = on);

GRANT SELECT ON view_profitability_by_customer TO authenticated;
GRANT SELECT ON view_working_capital_metrics TO authenticated;
GRANT SELECT ON view_revenue_concentration_risk TO authenticated;
GRANT SELECT ON view_subscription_metrics TO authenticated;
