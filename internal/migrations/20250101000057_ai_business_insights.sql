-- Migration: AI Business Insights Framework
-- Description: Tables to support AI-powered business insights and analytics
-- Created: 2025-01-01
-- Purpose: Enable smart business insights without requiring sensitive data

-- ============================================
-- BUSINESS INSIGHTS CACHE
-- ============================================

CREATE TABLE business_insights_cache (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Insight categories
    growth_compass jsonb,
    data_quality_summary jsonb,
    inventory_insights jsonb,
    customer_insights jsonb,
    financial_insights jsonb,
    team_performance jsonb,
    cash_flow_analysis jsonb,
    workflow_efficiency jsonb,
    
    -- Metadata
    generated_at timestamptz DEFAULT now(),
    expires_at timestamptz DEFAULT now() + INTERVAL '24 hours',
    
    -- Status
    is_active boolean DEFAULT true
);

COMMENT ON TABLE business_insights_cache IS
'Cached AI-generated business insights for fast dashboard loading';

CREATE INDEX idx_business_insights_cache_org ON business_insights_cache(organization_id);
CREATE INDEX idx_business_insights_cache_active ON business_insights_cache(organization_id, is_active)
    WHERE is_active = true;

-- ============================================
-- AI INSIGHT HISTORY
-- ============================================

CREATE TABLE ai_insight_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid,
    
    -- Insight details
    insight_type varchar(50) NOT NULL,
    title varchar(255) NOT NULL,
    summary text NOT NULL,
    detailed_analysis text,
    confidence_score integer,
    supporting_data jsonb,
    suggested_actions jsonb,
    
    -- Status and tracking
    status varchar(20) DEFAULT 'new',
    is_favorite boolean DEFAULT false,
    action_taken boolean DEFAULT false,
    action_taken_at timestamptz,
    
    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_insight_history_type_check CHECK (insight_type IN (
        'growth', 'data_quality', 'inventory', 'customer', 'financial',
        'team_performance', 'cash_flow', 'workflow', 'risk', 'opportunity'
    ))
);

COMMENT ON TABLE ai_insight_history IS
'Historical record of AI-generated insights and user actions';

CREATE INDEX idx_ai_insight_history_org ON ai_insight_history(organization_id, created_at);
CREATE INDEX idx_ai_insight_history_type ON ai_insight_history(organization_id, insight_type);
CREATE INDEX idx_ai_insight_history_favorite ON ai_insight_history(organization_id, is_favorite)
    WHERE is_favorite = true;

-- ============================================
-- AI INSIGHT FEEDBACK
-- ============================================

CREATE TABLE ai_insight_feedback (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    insight_id uuid REFERENCES ai_insight_history(id) ON DELETE CASCADE,
    
    -- Feedback details
    feedback_type varchar(20) NOT NULL,
    user_comment text,
    ai_response text,
    
    -- Impact tracking
    was_useful boolean,
    led_to_action boolean,
    estimated_impact varchar(50),
    
    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_insight_feedback_type_check CHECK (feedback_type IN (
        'useful', 'not_useful', 'incorrect', 'confusing', 'action_taken', 'ignored'
    ))
);

COMMENT ON TABLE ai_insight_feedback IS
'User feedback on AI insights to improve future recommendations';

CREATE INDEX idx_ai_insight_feedback_org ON ai_insight_feedback(organization_id);
CREATE INDEX idx_ai_insight_feedback_insight ON ai_insight_feedback(insight_id);

-- ============================================
-- AI INSIGHT NOTIFICATIONS
-- ============================================

CREATE TABLE ai_insight_notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    insight_id uuid REFERENCES ai_insight_history(id) ON DELETE CASCADE,
    
    -- Notification details
    notification_type varchar(20) NOT NULL,
    message text NOT NULL,
    priority integer DEFAULT 5,
    
    -- Delivery tracking
    delivery_method varchar(20),
    delivered_at timestamptz,
    read_at timestamptz,
    action_taken_at timestamptz,
    
    -- Status
    status varchar(20) DEFAULT 'pending',
    
    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_insight_notifications_type_check CHECK (notification_type IN (
        'alert', 'recommendation', 'opportunity', 'risk', 'summary'
    ))
);

COMMENT ON TABLE ai_insight_notifications IS
'Notification system for important AI insights';

CREATE INDEX idx_ai_insight_notifications_user ON ai_insight_notifications(user_id, status);
CREATE INDEX idx_ai_insight_notifications_priority ON ai_insight_notifications(user_id, priority, status)
    WHERE status = 'pending';

-- ============================================
-- BUSINESS METRICS TRACKING
-- ============================================

CREATE TABLE business_metrics (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Metric identification
    metric_type varchar(50) NOT NULL,
    metric_name varchar(255) NOT NULL,
    metric_code varchar(50),
    
    -- Metric values
    current_value numeric(15,2),
    previous_value numeric(15,2),
    target_value numeric(15,2),
    
    -- Analysis
    trend varchar(20),
    trend_percentage numeric(10,2),
    ai_analysis text,
    ai_recommendations jsonb,
    
    -- Time period
    period_type varchar(20) DEFAULT 'daily',
    period_start date,
    period_end date,
    
    -- Status
    is_active boolean DEFAULT true,
    
    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT business_metrics_type_check CHECK (metric_type IN (
        'revenue', 'expenses', 'profit', 'cash_flow', 'inventory',
        'customer_acquisition', 'customer_retention', 'sales_conversion',
        'order_value', 'team_productivity', 'data_quality'
    ))
);

COMMENT ON TABLE business_metrics IS
'Track key business metrics with AI analysis and recommendations';

CREATE INDEX idx_business_metrics_org ON business_metrics(organization_id, metric_type);
CREATE INDEX idx_business_metrics_period ON business_metrics(organization_id, period_type, period_end);

-- ============================================
-- TRIGGERS
-- ============================================

CREATE OR REPLACE FUNCTION update_ai_insight_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_ai_insight_history_updated_at
    BEFORE UPDATE ON ai_insight_history
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_insight_updated_at();

CREATE TRIGGER set_ai_insight_notifications_updated_at
    BEFORE UPDATE ON ai_insight_notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_insight_updated_at();

CREATE TRIGGER set_business_metrics_updated_at
    BEFORE UPDATE ON business_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_insight_updated_at();

-- ============================================
-- RLS POLICIES
-- ============================================

ALTER TABLE business_insights_cache ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_insight_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_insight_feedback ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_insight_notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE business_metrics ENABLE ROW LEVEL SECURITY;

-- Business Insights Cache Policies
CREATE POLICY "Users can view their org business insights"
    ON business_insights_cache FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

-- AI Insight History Policies
CREATE POLICY "Users can view their AI insights"
    ON ai_insight_history FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Users can manage their own AI insights"
    ON ai_insight_history FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

-- AI Insight Feedback Policies
CREATE POLICY "Users can provide feedback on insights"
    ON ai_insight_feedback FOR ALL
    USING (user_id = (SELECT auth.uid()));

-- AI Insight Notifications Policies
CREATE POLICY "Users can view their notifications"
    ON ai_insight_notifications FOR SELECT
    USING (user_id = (SELECT auth.uid()));

CREATE POLICY "Users can manage their notifications"
    ON ai_insight_notifications FOR ALL
    USING (user_id = (SELECT auth.uid()));

-- Business Metrics Policies
CREATE POLICY "Users can view their org metrics"
    ON business_metrics FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Generate business insights for an organization
CREATE OR REPLACE FUNCTION generate_business_insights(
    p_organization_id uuid
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_insights jsonb;
    v_growth jsonb;
    v_data_quality jsonb;
    v_inventory jsonb;
BEGIN
    -- Get existing insights or generate new ones
    SELECT insights INTO v_insights
    FROM business_insights_cache
    WHERE organization_id = p_organization_id
      AND is_active = true
      AND expires_at > now()
    LIMIT 1;
    
    IF v_insights IS NULL THEN
        -- Generate fresh insights
        v_growth := analytics_growth_compass(p_organization_id);
        v_data_quality := analytics_data_quality_summary(p_organization_id);
        v_inventory := analytics_inventory_turnover(p_organization_id);
        
        v_insights := jsonb_build_object(
            'growth_compass', v_growth,
            'data_quality', v_data_quality,
            'inventory', v_inventory
        );
        
        -- Cache the results
        INSERT INTO business_insights_cache (
            organization_id, growth_compass, data_quality_summary, 
            inventory_insights, generated_at
        ) VALUES (
            p_organization_id, v_growth, v_data_quality, v_inventory, now()
        );
    END IF;
    
    RETURN v_insights;
END;
$$;

COMMENT ON FUNCTION generate_business_insights IS
'Generate comprehensive business insights for an organization';

-- Create AI insight with analysis
CREATE OR REPLACE FUNCTION create_ai_insight(
    p_organization_id uuid,
    p_user_id uuid DEFAULT NULL,
    p_insight_type varchar,
    p_title varchar,
    p_summary text,
    p_confidence integer DEFAULT 80,
    p_supporting_data jsonb DEFAULT NULL,
    p_suggested_actions jsonb DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql
AS $$
DECLARE
    v_insight_id uuid;
BEGIN
    INSERT INTO ai_insight_history (
        organization_id, user_id, insight_type, title, summary,
        confidence_score, supporting_data, suggested_actions
    ) VALUES (
        p_organization_id, p_user_id, p_insight_type, p_title, p_summary,
        p_confidence, p_supporting_data, p_suggested_actions
    ) RETURNING id INTO v_insight_id;
    
    RETURN v_insight_id;
END;
$$;

COMMENT ON FUNCTION create_ai_insight IS
'Create a new AI-generated insight record';

-- Get recent insights for dashboard
CREATE OR REPLACE FUNCTION get_recent_ai_insights(
    p_organization_id uuid,
    p_limit integer DEFAULT 10
) RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_insights jsonb;
BEGIN
    SELECT jsonb_agg(
        jsonb_build_object(
            'id', id,
            'type', insight_type,
            'title', title,
            'summary', summary,
            'confidence', confidence_score,
            'created_at', created_at,
            'is_favorite', is_favorite
        )
    ) INTO v_insights
    FROM ai_insight_history
    WHERE organization_id = p_organization_id
      AND status = 'new'
    ORDER BY confidence_score DESC, created_at DESC
    LIMIT p_limit;
    
    RETURN v_insights;
END;
$$;

COMMENT ON FUNCTION get_recent_ai_insights IS
'Get recent AI insights for dashboard display';

-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT ON business_insights_cache TO authenticated;
GRANT ALL ON business_insights_cache TO service_role;

GRANT SELECT, INSERT, UPDATE ON ai_insight_history TO authenticated;
GRANT ALL ON ai_insight_history TO service_role;

GRANT ALL ON ai_insight_feedback TO authenticated;
GRANT ALL ON ai_insight_feedback TO service_role;

GRANT ALL ON ai_insight_notifications TO authenticated;
GRANT ALL ON ai_insight_notifications TO service_role;

GRANT SELECT ON business_metrics TO authenticated;
GRANT ALL ON business_metrics TO service_role;

GRANT EXECUTE ON FUNCTION generate_business_insights(uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION create_ai_insight(uuid, uuid, varchar, varchar, text, integer, jsonb, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION get_recent_ai_insights(uuid, integer) TO authenticated;

-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Generate business insights for an organization
SELECT generate_business_insights('your-org-id');

-- 2. Create a specific AI insight
SELECT create_ai_insight(
    'your-org-id',
    'user-id',
    'opportunity',
    'Upsell opportunity with Acme Corp',
    'Acme Corp has placed 3 orders in 60 days - suggest premium offering',
    85,
    '{"customer": "Acme Corp", "recent_orders": 3, "last_order": "2025-01-15"}',
    '[{"action": "Contact Acme Corp", "impact": "Potential $5K upsell"}]'
);

-- 3. Get recent insights for dashboard
SELECT get_recent_ai_insights('your-org-id', 5);

-- 4. Get cached insights (fast)
SELECT * FROM business_insights_cache 
WHERE organization_id = 'your-org-id' AND is_active = true;

-- 5. Track user feedback on insights
INSERT INTO ai_insight_feedback (
    organization_id, user_id, insight_id, feedback_type, was_useful
) VALUES (
    'your-org-id', 'user-id', 'insight-id', 'useful', true
);
*/
