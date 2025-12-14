-- Migration: Unified Business Insights Function
-- Description: Aggregated insight payload that combines analytics across modules
-- Created: 2025-01-01

-- =====================================================
-- BUSINESS INSIGHTS AGGREGATOR
-- =====================================================

CREATE OR REPLACE FUNCTION analytics_business_insights(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_date_to date DEFAULT CURRENT_DATE,
    p_cash_interval varchar DEFAULT 'month'
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_data_quality jsonb := '{}'::jsonb;
    v_pipeline jsonb := '[]'::jsonb;
    v_win_loss jsonb := '{}'::jsonb;
    v_top_customers jsonb := '[]'::jsonb;
    v_product_profitability jsonb := '[]'::jsonb;
    v_low_stock jsonb := '[]'::jsonb;
    v_inventory_turnover jsonb := '[]'::jsonb;
    v_top_performers jsonb := '[]'::jsonb;
    v_department jsonb := '[]'::jsonb;
    v_leave_risks jsonb := '[]'::jsonb;
    v_cash_periods jsonb := '[]'::jsonb;
    v_cash_summary jsonb := '{}'::jsonb;
    v_receivables jsonb := '{}'::jsonb;
    v_payables jsonb := '{}'::jsonb;
    v_ar_total numeric := 0;
    v_ar_over_90 numeric := 0;
    v_ar_accounts jsonb := '[]'::jsonb;
    v_ap_total numeric := 0;
    v_ap_over_90 numeric := 0;
    v_ap_accounts jsonb := '[]'::jsonb;
BEGIN
    -- Data quality summary (leverages existing quality detectors)
    v_data_quality := analytics_data_quality_summary(p_organization_id);

    -- Pipeline health snapshot
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'stage_id', stage_id,
                'stage_name', stage_name,
                'probability', stage_probability,
                'opportunity_count', opportunity_count,
                'total_expected_revenue', total_expected_revenue,
                'weighted_revenue', weighted_revenue,
                'avg_days_in_stage', avg_days_in_stage
            )
        ),
        '[]'::jsonb
    )
    INTO v_pipeline
    FROM (
        SELECT *
        FROM analytics_sales_pipeline(p_organization_id)
        ORDER BY stage_probability DESC, stage_name
    ) pipeline_data;

    -- Win/loss performance
    v_win_loss := analytics_win_loss_analysis(p_organization_id, p_date_from, p_date_to);

    -- Top customers by revenue (limit 5)
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'rank', rank,
                'customer_id', customer_id,
                'customer_name', customer_name,
                'total_revenue', total_revenue,
                'invoice_count', invoice_count,
                'avg_invoice_amount', avg_invoice_amount,
                'last_order_date', last_order_date
            )
            ORDER BY rank
        ),
        '[]'::jsonb
    )
    INTO v_top_customers
    FROM analytics_top_customers(p_organization_id, p_date_from, p_date_to, 5);

    -- Product profitability leaders (top 5 by profit)
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'product_id', product_id,
                'product_name', product_name,
                'category_name', category_name,
                'qty_sold', qty_sold,
                'revenue', revenue,
                'cost', cost,
                'profit', profit,
                'profit_margin_pct', profit_margin_pct
            )
            ORDER BY profit DESC, product_name
        ),
        '[]'::jsonb
    )
    INTO v_product_profitability
    FROM (
        SELECT *
        FROM analytics_product_profitability(
            p_organization_id,
            p_date_from,
            p_date_to
        )
        ORDER BY profit DESC
        LIMIT 5
    ) profit_data;

    -- Low stock alerts (top 5 most critical)
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'product_id', product_id,
                'product_name', product_name,
                'product_code', product_code,
                'location_id', location_id,
                'location_name', location_name,
                'current_qty', current_qty,
                'reserved_qty', reserved_qty,
                'available_qty', available_qty,
                'status', status
            )
            ORDER BY available_qty ASC, product_name
        ),
        '[]'::jsonb
    )
    INTO v_low_stock
    FROM (
        SELECT *
        FROM analytics_low_stock_alerts(p_organization_id, 10)
        ORDER BY available_qty ASC
        LIMIT 5
    ) stock_data;

    -- Inventory turnover (highlight products that lag)
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'product_id', product_id,
                'product_name', product_name,
                'category_name', category_name,
                'avg_inventory', avg_inventory,
                'cost_of_goods_sold', cost_of_goods_sold,
                'turnover_rate', turnover_rate,
                'days_to_sell', days_to_sell
            )
            ORDER BY days_to_sell DESC, product_name
        ),
        '[]'::jsonb
    )
    INTO v_inventory_turnover
    FROM (
        SELECT *
        FROM analytics_inventory_turnover(
            p_organization_id,
            p_date_from,
            p_date_to
        )
        ORDER BY days_to_sell DESC
        LIMIT 5
    ) turnover_data;

    -- Team performance (top performers)
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'employee_id', employee_id,
                'employee_name', employee_name,
                'department_name', department_name,
                'sales_orders_count', sales_orders_count,
                'sales_total_amount', sales_total_amount,
                'hours_logged', hours_logged,
                'activities_completed', activities_completed,
                'tasks_completed', tasks_completed,
                'productivity_score', productivity_score
            )
            ORDER BY productivity_score DESC, employee_name
        ),
        '[]'::jsonb
    )
    INTO v_top_performers
    FROM (
        SELECT *
        FROM analytics_employee_productivity(
            p_organization_id,
            p_date_from,
            p_date_to
        )
        ORDER BY productivity_score DESC
        LIMIT 5
    ) productivity_data;

    -- Department health
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'department_id', department_id,
                'department_name', department_name,
                'employee_count', employee_count,
                'total_sales', total_sales,
                'avg_sales_per_employee', avg_sales_per_employee,
                'total_hours_logged', total_hours_logged,
                'avg_productivity_score', avg_productivity_score,
                'avg_attendance_rate', avg_attendance_rate,
                'performance_grade', performance_grade
            )
            ORDER BY avg_productivity_score DESC, department_name
        ),
        '[]'::jsonb
    )
    INTO v_department
    FROM (
        SELECT *
        FROM analytics_department_performance(
            p_organization_id,
            p_date_from,
            p_date_to
        )
        ORDER BY avg_productivity_score DESC
        LIMIT 5
    ) department_data;

    -- Leave balance risks
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'employee_id', employee_id,
                'employee_name', employee_name,
                'leave_type_id', leave_type_id,
                'leave_type_name', leave_type_name,
                'total_allocation', total_allocation,
                'days_taken', days_taken,
                'days_remaining', days_remaining,
                'pending_requests', pending_requests,
                'status', status
            )
            ORDER BY days_remaining ASC, employee_name
        ),
        '[]'::jsonb
    )
    INTO v_leave_risks
    FROM (
        SELECT *
        FROM analytics_employee_leave_balance(p_organization_id)
        WHERE status IN ('critical', 'low')
        ORDER BY
            CASE status
                WHEN 'critical' THEN 1
                WHEN 'low' THEN 2
                ELSE 3
            END,
            days_remaining ASC
        LIMIT 5
    ) leave_data;

    -- Cash flow periods and summary
    SELECT
        COALESCE(
            jsonb_agg(
                jsonb_build_object(
                    'period_start', period_start,
                    'period_end', period_end,
                    'cash_in', cash_in,
                    'cash_out', cash_out,
                    'net_cash_flow', net_cash_flow,
                    'cumulative_cash_flow', cumulative_cash_flow
                )
                ORDER BY period_start
            ),
            '[]'::jsonb
        ),
        jsonb_build_object(
            'total_cash_in', COALESCE(SUM(cash_in), 0),
            'total_cash_out', COALESCE(SUM(cash_out), 0),
            'net_cash_flow', COALESCE(SUM(net_cash_flow), 0),
            'ending_cash', COALESCE(MAX(cumulative_cash_flow), 0)
        )
    INTO v_cash_periods, v_cash_summary
    FROM analytics_cash_flow(
        p_organization_id,
        p_date_from,
        p_date_to,
        p_cash_interval
    );

    -- Accounts receivable snapshot
    SELECT
        COALESCE(SUM(total_due), 0),
        COALESCE(SUM(over_90), 0),
        COALESCE(
            jsonb_agg(account_info) FILTER (WHERE row_number <= 5),
            '[]'::jsonb
        )
    INTO v_ar_total, v_ar_over_90, v_ar_accounts
    FROM (
        SELECT
            ROW_NUMBER() OVER (ORDER BY total_due DESC) AS row_number,
            total_due,
            over_90,
            jsonb_build_object(
                'partner_id', partner_id,
                'partner_name', partner_name,
                'current_amount', current_amount,
                'days_30', days_30,
                'days_60', days_60,
                'days_90', days_90,
                'over_90', over_90,
                'total_due', total_due
            ) AS account_info
        FROM analytics_ar_aging(p_organization_id, p_date_to)
    ) ar_data;

    v_receivables := jsonb_build_object(
        'total_due', v_ar_total,
        'over_90', v_ar_over_90,
        'top_accounts', v_ar_accounts
    );

    -- Accounts payable snapshot
    SELECT
        COALESCE(SUM(total_due), 0),
        COALESCE(SUM(over_90), 0),
        COALESCE(
            jsonb_agg(account_info) FILTER (WHERE row_number <= 5),
            '[]'::jsonb
        )
    INTO v_ap_total, v_ap_over_90, v_ap_accounts
    FROM (
        SELECT
            ROW_NUMBER() OVER (ORDER BY total_due DESC) AS row_number,
            total_due,
            over_90,
            jsonb_build_object(
                'partner_id', partner_id,
                'partner_name', partner_name,
                'current_amount', current_amount,
                'days_30', days_30,
                'days_60', days_60,
                'days_90', days_90,
                'over_90', over_90,
                'total_due', total_due
            ) AS account_info
        FROM analytics_ap_aging(p_organization_id, p_date_to)
    ) ap_data;

    v_payables := jsonb_build_object(
        'total_due', v_ap_total,
        'over_90', v_ap_over_90,
        'top_accounts', v_ap_accounts
    );

    RETURN jsonb_build_object(
        'generated_at', now(),
        'organization_id', p_organization_id,
        'period', jsonb_build_object(
            'start', p_date_from,
            'end', p_date_to
        ),
        'data_quality', v_data_quality,
        'sales', jsonb_build_object(
            'pipeline', v_pipeline,
            'win_loss', v_win_loss,
            'top_customers', v_top_customers,
            'product_profitability', v_product_profitability
        ),
        'finance', jsonb_build_object(
            'cash_flow', jsonb_build_object(
                'summary', v_cash_summary,
                'periods', v_cash_periods
            ),
            'receivables', v_receivables,
            'payables', v_payables
        ),
        'inventory', jsonb_build_object(
            'low_stock', v_low_stock,
            'inventory_turnover', v_inventory_turnover
        ),
        'people', jsonb_build_object(
            'top_performers', v_top_performers,
            'department_performance', v_department,
            'leave_risks', v_leave_risks
        )
    );
END;
$$;

COMMENT ON FUNCTION analytics_business_insights IS 'Aggregate cross-functional metrics into a single insights payload';
