-- Migration: Advanced CRM Analytics Views
-- Description: Conversion funnels, engagement scoring, velocity metrics, and quote analysis
-- Created: 2025-01-01
-- Module: Advanced CRM

-- =====================================================
-- LEAD CONVERSION FUNNEL
-- =====================================================

-- Stage-by-stage conversion analysis
CREATE OR REPLACE VIEW view_lead_conversion_funnel AS
WITH stage_stats AS (
    SELECT
        ls.id as stage_id,
        ls.name as stage_name,
        ls.sequence,
        ls.probability,
        -- Current stage metrics
        COUNT(DISTINCT l.id) as leads_in_stage,
        COALESCE(SUM(l.expected_revenue), 0) as stage_value,
        -- Time in stage metrics
        AVG(EXTRACT(EPOCH FROM (COALESCE(l.date_closed, CURRENT_DATE::timestamptz) - COALESCE(l.date_last_stage_update, l.created_at)))::integer / 86400) as avg_days_in_stage,
        PERCENTILE_CONT(0.5) WITHIN GROUP (
            ORDER BY EXTRACT(EPOCH FROM (COALESCE(l.date_closed, CURRENT_DATE::timestamptz) - COALESCE(l.date_last_stage_update, l.created_at)))::integer / 86400
        ) as median_days_in_stage,
        PERCENTILE_CONT(0.9) WITHIN GROUP (
            ORDER BY EXTRACT(EPOCH FROM (COALESCE(l.date_closed, CURRENT_DATE::timestamptz) - COALESCE(l.date_last_stage_update, l.created_at)))::integer / 86400
        ) as p90_days_in_stage,
        -- Outcomes from this stage
        COUNT(DISTINCT l.id) FILTER (WHERE l.won_status = 'won') as won_from_stage,
        COUNT(DISTINCT l.id) FILTER (WHERE l.won_status = 'lost') as lost_from_stage,
        COALESCE(SUM(l.expected_revenue) FILTER (WHERE l.won_status = 'won'), 0) as revenue_from_stage
    FROM lead_stages ls
    LEFT JOIN leads l ON l.stage_id = ls.id
        AND l.deleted_at IS NULL
        AND l.organization_id = get_current_user_organization_id()
    WHERE ls.organization_id = get_current_user_organization_id()
    GROUP BY ls.id, ls.name, ls.sequence, ls.probability
),
monthly_comparison AS (
    SELECT
        ls.id as stage_id,
        -- Current month
        COUNT(DISTINCT l.id) FILTER (
            WHERE l.created_at >= DATE_TRUNC('month', CURRENT_DATE)
        ) as leads_this_month,
        -- Previous month
        COUNT(DISTINCT l.id) FILTER (
            WHERE l.created_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
            AND l.created_at < DATE_TRUNC('month', CURRENT_DATE)
        ) as leads_last_month
    FROM lead_stages ls
    LEFT JOIN leads l ON l.stage_id = ls.id AND l.deleted_at IS NULL
    WHERE ls.organization_id = get_current_user_organization_id()
    GROUP BY ls.id
)
SELECT
    ss.stage_id,
    ss.stage_name,
    ss.sequence,
    ss.probability as stage_probability,
    ss.leads_in_stage,
    ss.stage_value,
    -- Conversion metrics
    CASE
        WHEN ss.leads_in_stage > 0 THEN
            (ss.won_from_stage::numeric / ss.leads_in_stage * 100)
        ELSE 0
    END as stage_conversion_rate,
    CASE
        WHEN (ss.won_from_stage + ss.lost_from_stage) > 0 THEN
            (ss.won_from_stage::numeric / (ss.won_from_stage + ss.lost_from_stage) * 100)
        ELSE 0
    END as stage_win_rate,
    -- Time metrics
    ROUND(ss.avg_days_in_stage::numeric, 1) as avg_days_in_stage,
    ROUND(ss.median_days_in_stage::numeric, 1) as median_days_in_stage,
    ROUND(ss.p90_days_in_stage::numeric, 1) as p90_days_in_stage,
    -- Outcomes
    ss.won_from_stage,
    ss.lost_from_stage,
    ss.revenue_from_stage,
    -- Trends
    mc.leads_this_month,
    mc.leads_last_month,
    CASE
        WHEN mc.leads_last_month > 0 THEN
            ((mc.leads_this_month - mc.leads_last_month)::numeric / mc.leads_last_month * 100)
        ELSE 0
    END as month_over_month_change,
    -- Health indicators
    CASE
        WHEN ss.avg_days_in_stage > 60 THEN 'Stagnant'
        WHEN ss.avg_days_in_stage > 30 THEN 'Slow'
        WHEN ss.avg_days_in_stage > 14 THEN 'Normal'
        ELSE 'Fast'
    END as stage_velocity,
    CASE
        WHEN ss.leads_in_stage = 0 THEN 'empty'
        WHEN ss.avg_days_in_stage > 60 THEN 'red'
        WHEN ss.avg_days_in_stage > 30 THEN 'orange'
        ELSE 'green'
    END as health_color
FROM stage_stats ss
LEFT JOIN monthly_comparison mc ON mc.stage_id = ss.stage_id
ORDER BY ss.sequence;

COMMENT ON VIEW view_lead_conversion_funnel IS 'Detailed conversion funnel with stage velocity, conversion rates, and trends';

-- =====================================================
-- CUSTOMER ENGAGEMENT SCORE
-- =====================================================

CREATE OR REPLACE VIEW view_customer_engagement_score AS
WITH customer_interactions AS (
    SELECT
        c.id as customer_id,
        c.name as customer_name,
        c.email,
        -- Order activity (0-25 points)
        CASE
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.date_order >= CURRENT_DATE - INTERVAL '30 days') >= 3 THEN 25
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.date_order >= CURRENT_DATE - INTERVAL '30 days') = 2 THEN 20
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.date_order >= CURRENT_DATE - INTERVAL '30 days') = 1 THEN 15
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.date_order >= CURRENT_DATE - INTERVAL '90 days') >= 1 THEN 10
            ELSE 0
        END as order_activity_score,
        -- Quote requests (0-20 points)
        CASE
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.state = 'sent' AND so.date_order >= CURRENT_DATE - INTERVAL '30 days') >= 2 THEN 20
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.state = 'sent' AND so.date_order >= CURRENT_DATE - INTERVAL '30 days') = 1 THEN 15
            WHEN COUNT(DISTINCT so.id) FILTER (WHERE so.state = 'sent' AND so.date_order >= CURRENT_DATE - INTERVAL '90 days') >= 1 THEN 10
            ELSE 0
        END as quote_activity_score,
        -- Lead/opportunity activity (0-20 points)
        CASE
            WHEN COUNT(DISTINCT l.id) FILTER (WHERE l.date_last_stage_update >= CURRENT_DATE - INTERVAL '14 days') >= 1 THEN 20
            WHEN COUNT(DISTINCT l.id) FILTER (WHERE l.date_last_stage_update >= CURRENT_DATE - INTERVAL '30 days') >= 1 THEN 15
            WHEN COUNT(DISTINCT l.id) FILTER (WHERE l.active = true) >= 1 THEN 10
            ELSE 0
        END as pipeline_engagement_score,
        -- Communication frequency (0-20 points) - based on activities
        CASE
            WHEN COUNT(DISTINCT a.id) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '14 days') >= 3 THEN 20
            WHEN COUNT(DISTINCT a.id) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '30 days') >= 2 THEN 15
            WHEN COUNT(DISTINCT a.id) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '60 days') >= 1 THEN 10
            ELSE 0
        END as communication_score,
        -- Payment behavior (0-15 points)
        CASE
            WHEN AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid') <= 0 THEN 15
            WHEN AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid') <= 7 THEN 12
            WHEN AVG((i.date - i.invoice_date_due)::integer) FILTER (WHERE i.payment_state = 'paid') <= 30 THEN 8
            ELSE 5
        END as payment_behavior_score,
        -- Supporting metrics
        COUNT(DISTINCT so.id) FILTER (WHERE so.date_order >= CURRENT_DATE - INTERVAL '30 days') as orders_last_30d,
        COUNT(DISTINCT l.id) FILTER (WHERE l.active = true) as active_opportunities,
        COUNT(DISTINCT a.id) FILTER (WHERE a.created_at >= CURRENT_DATE - INTERVAL '30 days') as activities_last_30d,
        MAX(so.date_order) as last_order_date,
        MAX(a.created_at) as last_activity_date
    FROM contacts c
    LEFT JOIN sales_orders so ON so.partner_id = c.id AND so.deleted_at IS NULL
    LEFT JOIN leads l ON l.contact_id = c.id AND l.deleted_at IS NULL
    LEFT JOIN activities a ON a.res_model = 'contacts' AND a.res_id = c.id
    LEFT JOIN invoices i ON i.partner_id = c.id AND i.deleted_at IS NULL AND i.move_type = 'out_invoice'
    WHERE c.organization_id = get_current_user_organization_id()
      AND c.is_customer = true
      AND c.deleted_at IS NULL
    GROUP BY c.id, c.name, c.email
)
SELECT
    customer_id,
    customer_name,
    email,
    -- Component scores
    order_activity_score,
    quote_activity_score,
    pipeline_engagement_score,
    communication_score,
    payment_behavior_score,
    -- Total engagement score (0-100)
    (order_activity_score + quote_activity_score + pipeline_engagement_score +
     communication_score + payment_behavior_score) as total_engagement_score,
    -- Engagement level
    CASE
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 80 THEN 'Highly Engaged'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 60 THEN 'Engaged'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 40 THEN 'Moderately Engaged'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 20 THEN 'Low Engagement'
        ELSE 'Disengaged'
    END as engagement_level,
    -- Color coding
    CASE
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 80 THEN 'green'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 60 THEN 'blue'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 40 THEN 'yellow'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 20 THEN 'orange'
        ELSE 'red'
    END as engagement_color,
    -- Supporting metrics
    orders_last_30d,
    active_opportunities,
    activities_last_30d,
    last_order_date,
    last_activity_date,
    EXTRACT(EPOCH FROM (CURRENT_DATE - last_activity_date))::integer / 86400 as days_since_last_activity,
    -- Recommended action
    CASE
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 80 THEN 'Nurture & Expand'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 60 THEN 'Maintain Relationship'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 40 THEN 'Increase Touchpoints'
        WHEN (order_activity_score + quote_activity_score + pipeline_engagement_score +
              communication_score + payment_behavior_score) >= 20 THEN 'Re-engagement Campaign'
        ELSE 'Win-back or Archive'
    END as recommended_action
FROM customer_interactions
ORDER BY total_engagement_score DESC;

COMMENT ON VIEW view_customer_engagement_score IS 'Multi-dimensional customer engagement scoring with recommended actions';

-- =====================================================
-- ENABLE SECURITY AND GRANT PERMISSIONS
-- =====================================================

ALTER VIEW view_lead_conversion_funnel SET (security_invoker = on);
ALTER VIEW view_customer_engagement_score SET (security_invoker = on);

GRANT SELECT ON view_lead_conversion_funnel TO authenticated;
GRANT SELECT ON view_customer_engagement_score TO authenticated;
