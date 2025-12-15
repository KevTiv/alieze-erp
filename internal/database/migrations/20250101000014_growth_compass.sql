-- Migration: Growth Compass Insights
-- Description: Entrepreneur-friendly momentum score and guidance
-- Created: 2025-01-01

-- =====================================================
-- GROWTH COMPASS FUNCTION
-- =====================================================

CREATE OR REPLACE FUNCTION analytics_growth_compass(
    p_organization_id uuid,
    p_target_revenue numeric DEFAULT 0,
    p_date_from date DEFAULT date_trunc('month', CURRENT_DATE),
    p_date_to date DEFAULT (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month' - INTERVAL '1 day')
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_target numeric := COALESCE(p_target_revenue, 0);
    v_actual_revenue numeric := 0;
    v_pipeline_total numeric := 0;
    v_weighted_pipeline numeric := 0;
    v_projected_total numeric := 0;
    v_remaining_target numeric := 0;
    v_pipeline_coverage numeric := 0;
    v_pipeline_needed numeric := 0;
    v_period_days integer;
    v_days_elapsed integer;
    v_days_left integer;
    v_time_progress numeric := 0;
    v_expected_to_date numeric := 0;
    v_pace_ratio numeric;
    v_pace_ratio_percent numeric;
    v_pace_gap numeric := 0;
    v_win_loss jsonb := '{}'::jsonb;
    v_win_rate numeric := 0;
    v_avg_sales_cycle numeric := 0;
    v_avg_deal_size numeric := 0;
    v_projected_pct numeric := 0;
    v_gap numeric := 0;
    v_ar_overdue numeric := 0;
    v_ar_over_90 numeric := 0;
    v_cash_in numeric := 0;
    v_cash_out numeric := 0;
    v_cash_net numeric := 0;
    v_cash_ending numeric := 0;
    v_avg_weekly_out numeric := 0;
    v_cash_runway_weeks numeric := NULL;
    v_period_weeks numeric := 0;
    v_score numeric := 0;
    v_score_pace numeric := 0;
    v_score_projection numeric := 0;
    v_score_pipeline numeric := 0;
    v_score_win numeric := 0;
    v_status text;
    v_headline text;
    v_actions text[] := ARRAY[]::text[];
    v_projected_text text;
    v_target_text text;
    v_gap_text text;
    v_pct_text text;
    v_days_left_text text;
BEGIN
    IF p_date_to < p_date_from THEN
        RAISE EXCEPTION 'analytics_growth_compass: p_date_to must be on or after p_date_from';
    END IF;

    v_period_days := GREATEST(1, (p_date_to - p_date_from) + 1);

    v_days_elapsed := CASE
        WHEN CURRENT_DATE < p_date_from THEN 0
        WHEN CURRENT_DATE > p_date_to THEN v_period_days
        ELSE (CURRENT_DATE - p_date_from) + 1
    END;

    v_days_left := GREATEST(v_period_days - v_days_elapsed, 0);
    v_time_progress := LEAST(1, GREATEST(0, v_days_elapsed::numeric / v_period_days::numeric));

    -- Actual booked revenue (posted invoices)
    SELECT COALESCE(SUM(amount_total), 0)
    INTO v_actual_revenue
    FROM invoices
    WHERE organization_id = p_organization_id
      AND move_type = 'out_invoice'
      AND state = 'posted'
      AND invoice_date BETWEEN p_date_from AND p_date_to
      AND deleted_at IS NULL;

    -- Pipeline health
    SELECT
        COALESCE(SUM(total_expected_revenue), 0),
        COALESCE(SUM(weighted_revenue), 0)
    INTO v_pipeline_total, v_weighted_pipeline
    FROM analytics_sales_pipeline(p_organization_id);

    -- Adjust target if not provided (fallback to projected total so ratios make sense)
    IF v_target <= 0 THEN
        v_target := GREATEST(v_actual_revenue + v_weighted_pipeline, 1);
    END IF;

    v_expected_to_date := v_target * v_time_progress;
    v_pace_ratio := CASE
        WHEN v_expected_to_date > 0 THEN v_actual_revenue / v_expected_to_date
        ELSE NULL
    END;
    v_pace_ratio_percent := CASE
        WHEN v_pace_ratio IS NOT NULL THEN ROUND(v_pace_ratio * 100, 1)
        ELSE NULL
    END;
    v_pace_gap := GREATEST(v_expected_to_date - v_actual_revenue, 0);

    v_remaining_target := GREATEST(v_target - v_actual_revenue, 0);
    v_pipeline_coverage := CASE
        WHEN v_remaining_target > 0 THEN v_weighted_pipeline / v_remaining_target
        ELSE 1
    END;
    v_pipeline_needed := GREATEST(v_remaining_target - v_weighted_pipeline, 0);

    v_projected_total := v_actual_revenue + v_weighted_pipeline;
    v_projected_pct := CASE
        WHEN v_target > 0 THEN ROUND((v_projected_total / v_target) * 100, 1)
        ELSE NULL
    END;
    v_gap := GREATEST(v_target - v_projected_total, 0);

    -- Win/loss trends (look back 90 days to smooth)
    v_win_loss := analytics_win_loss_analysis(
        p_organization_id,
        p_date_from - INTERVAL '90 days',
        p_date_to
    );

    v_win_rate := COALESCE((v_win_loss->>'win_rate_pct')::numeric, 0);
    v_avg_sales_cycle := COALESCE((v_win_loss->>'avg_sales_cycle_days')::numeric, 0);
    v_avg_deal_size := COALESCE((v_win_loss->>'avg_deal_size')::numeric, 0);

    -- Collections risk
    SELECT
        COALESCE(SUM(days_30 + days_60 + days_90 + over_90), 0),
        COALESCE(SUM(over_90), 0)
    INTO v_ar_overdue, v_ar_over_90
    FROM analytics_ar_aging(p_organization_id, p_date_to);

    -- Cash flow trends
    SELECT
        COALESCE(SUM(cash_in), 0),
        COALESCE(SUM(cash_out), 0),
        COALESCE(SUM(net_cash_flow), 0),
        COALESCE(MAX(cumulative_cash_flow), 0)
    INTO v_cash_in, v_cash_out, v_cash_net, v_cash_ending
    FROM analytics_cash_flow(
        p_organization_id,
        p_date_from,
        p_date_to,
        'week'
    );

    v_period_weeks := CEILING(v_period_days::numeric / 7.0);
    IF v_period_weeks < 1 THEN
        v_period_weeks := 1;
    END IF;

    v_avg_weekly_out := CASE
        WHEN v_period_weeks > 0 THEN v_cash_out / v_period_weeks
        ELSE 0
    END;

    IF v_avg_weekly_out > 0 THEN
        v_cash_runway_weeks := ROUND(v_cash_ending / v_avg_weekly_out, 1);
    ELSE
        v_cash_runway_weeks := NULL;
    END IF;

    -- Momentum scoring (0-100)
    v_score_pace := CASE
        WHEN v_expected_to_date > 0 THEN LEAST(1, v_actual_revenue / v_expected_to_date)
        ELSE 1
    END;

    v_score_projection := CASE
        WHEN v_target > 0 THEN LEAST(1, v_projected_total / v_target)
        ELSE 1
    END;

    v_score_pipeline := LEAST(1, v_pipeline_coverage);
    v_score_win := LEAST(1, v_win_rate / 100);

    v_score := ROUND(
        (v_score_pace * 35) +
        (v_score_projection * 40) +
        (v_score_pipeline * 15) +
        (v_score_win * 10)
    );

    v_score := LEAST(100, GREATEST(0, v_score));

    -- Status bands
    IF v_projected_total >= v_target THEN
        v_status := 'on_track';
    ELSIF v_projected_total >= v_target * 0.85 THEN
        v_status := 'monitor';
    ELSE
        v_status := 'at_risk';
    END IF;

    v_projected_text := to_char(v_projected_total, 'FM$999G999G990D00');
    v_target_text := to_char(v_target, 'FM$999G999G990D00');
    v_gap_text := to_char(v_gap, 'FM$999G999G990D00');
    v_pct_text := CASE
        WHEN v_projected_pct IS NOT NULL THEN to_char(v_projected_pct, 'FM999D0')
        ELSE 'N/A'
    END;
    v_days_left_text := CASE
        WHEN v_days_left = 1 THEN '1 day'
        ELSE format('%s days', v_days_left)
    END;

    IF v_status = 'on_track' THEN
        v_headline := format(
            'On track: projected %s (%s%% of goal) with %s left.',
            v_projected_text,
            v_pct_text,
            v_days_left_text
        );
    ELSIF v_status = 'monitor' THEN
        v_headline := format(
            'Within reach: projected %s (%s%% of goal) — gap of %s with %s left.',
            v_projected_text,
            v_pct_text,
            v_gap_text,
            v_days_left_text
        );
    ELSE
        v_headline := format(
            'At risk: short by %s against goal %s — %s remain.',
            v_gap_text,
            v_target_text,
            v_days_left_text
        );
    END IF;

    -- Recommended actions
    IF v_pace_ratio IS NOT NULL AND v_pace_ratio < 0.95 AND v_pace_gap > 0 THEN
        v_actions := array_append(
            v_actions,
            format(
                'Close %s more revenue to get back on pace this period.',
                to_char(v_pace_gap, 'FM$999G999G990D00')
            )
        );
    END IF;

    IF v_pipeline_coverage < 1 AND v_pipeline_needed > 0 THEN
        v_actions := array_append(
            v_actions,
            format(
                'Add %s in weighted pipeline to cover the remaining target.',
                to_char(v_pipeline_needed, 'FM$999G999G990D00')
            )
        );
    END IF;

    IF v_ar_overdue > 0 THEN
        v_actions := array_append(
            v_actions,
            format(
                'Collect %s in overdue invoices (%s over 90 days).',
                to_char(v_ar_overdue, 'FM$999G999G990D00'),
                to_char(v_ar_over_90, 'FM$999G999G990D00')
            )
        );
    END IF;

    IF v_cash_net < 0 THEN
        v_actions := array_append(
            v_actions,
            format(
                'Cash flow is negative %s for the period; runway ≈ %s weeks.',
                to_char(ABS(v_cash_net), 'FM$999G999G990D00'),
                COALESCE(to_char(v_cash_runway_weeks, 'FM999D0'), 'N/A')
            )
        );
    END IF;

    IF v_cash_runway_weeks IS NOT NULL AND v_cash_runway_weeks < 6 THEN
        v_actions := array_append(
            v_actions,
            format(
                'Cash runway is about %s weeks — consider trimming spend or accelerating collections.',
                to_char(v_cash_runway_weeks, 'FM999D0')
            )
        );
    END IF;

    IF v_win_rate < 25 THEN
        v_actions := array_append(
            v_actions,
            format(
                'Win rate is %s%%; reinforce deal coaching and qualification.',
                to_char(v_win_rate, 'FM999D0')
            )
        );
    END IF;

    RETURN jsonb_build_object(
        'generated_at', now(),
        'organization_id', p_organization_id,
        'status', v_status,
        'headline', v_headline,
        'momentum', jsonb_build_object(
            'score', v_score,
            'breakdown', jsonb_build_object(
                'pace', ROUND(v_score_pace * 35, 1),
                'projection', ROUND(v_score_projection * 40, 1),
                'pipeline', ROUND(v_score_pipeline * 15, 1),
                'win_rate', ROUND(v_score_win * 10, 1)
            ),
            'weights', jsonb_build_object(
                'pace', 35,
                'projection', 40,
                'pipeline', 15,
                'win_rate', 10
            )
        ),
        'metrics', jsonb_build_object(
            'revenue', jsonb_build_object(
                'actual', v_actual_revenue,
                'projected', v_projected_total,
                'target', v_target,
                'projected_pct', v_projected_pct,
                'pace_ratio', v_pace_ratio,
                'pace_ratio_percent', v_pace_ratio_percent,
                'expected_to_date', v_expected_to_date
            ),
            'pipeline', jsonb_build_object(
                'total_expected', v_pipeline_total,
                'weighted', v_weighted_pipeline,
                'coverage', v_pipeline_coverage,
                'remaining_target', v_remaining_target
            ),
            'sales_performance', jsonb_build_object(
                'win_rate_pct', v_win_rate,
                'avg_sales_cycle_days', v_avg_sales_cycle,
                'avg_deal_size', v_avg_deal_size
            ),
            'cash', jsonb_build_object(
                'inflow', v_cash_in,
                'outflow', v_cash_out,
                'net_flow', v_cash_net,
                'ending_cash', v_cash_ending,
                'runway_weeks', v_cash_runway_weeks
            ),
            'collections', jsonb_build_object(
                'overdue', v_ar_overdue,
                'over_90', v_ar_over_90
            ),
            'period', jsonb_build_object(
                'start', p_date_from,
                'end', p_date_to,
                'days_elapsed', v_days_elapsed,
                'days_remaining', v_days_left,
                'time_progress', v_time_progress
            )
        ),
        'actions', COALESCE(to_jsonb(v_actions), '[]'::jsonb)
    );
END;
$$;

COMMENT ON FUNCTION analytics_growth_compass IS 'Summarize revenue pace, pipeline coverage, cash runway, and recommended actions for entrepreneurs';
