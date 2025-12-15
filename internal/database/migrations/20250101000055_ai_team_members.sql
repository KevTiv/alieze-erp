-- Migration: AI Team Members Framework
-- Description: Creates the foundation for AI agents that act as virtual team members
-- Created: 2025-01-01
-- Purpose: Give solo entrepreneurs the feeling of having a supportive team

-- ============================================
-- AI TEAM MEMBERS (Virtual Colleagues)
-- ============================================

CREATE TABLE ai_team_members (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(50) NOT NULL,
    role varchar(50) NOT NULL,
    description text,
    personality_traits jsonb DEFAULT '{"tone": "friendly", "style": "concise"}'::jsonb,
    avatar_url varchar(255),
    capabilities jsonb DEFAULT '[]'::jsonb,
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_team_members_role_check CHECK (role IN (
        'sales_assistant', 'inventory_manager', 'financial_advisor', 
        'business_strategist', 'customer_support', 'marketing_assistant'
    ))
);

COMMENT ON TABLE ai_team_members IS
'Virtual team members that provide AI-powered assistance to entrepreneurs';

CREATE INDEX idx_ai_team_members_org ON ai_team_members(organization_id);
CREATE INDEX idx_ai_team_members_role ON ai_team_members(organization_id, role);

-- ============================================
-- AI AGENT TASKS (Simple, Actionable Items)
-- ============================================

CREATE TABLE ai_agent_tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    agent_id uuid NOT NULL REFERENCES ai_team_members(id),
    user_id uuid,
    task_type varchar(50) NOT NULL,
    title varchar(255) NOT NULL,
    description text,
    status varchar(20) DEFAULT 'pending',
    priority integer DEFAULT 5,
    related_entity_type varchar(50),
    related_entity_id uuid,
    scheduled_for timestamptz,
    completed_at timestamptz,
    result jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_agent_tasks_status_check CHECK (status IN (
        'pending', 'in_progress', 'completed', 'failed', 'dismissed'
    ))
);

COMMENT ON TABLE ai_agent_tasks IS
'Simple, actionable tasks created by AI agents for the entrepreneur';

CREATE INDEX idx_ai_agent_tasks_org ON ai_agent_tasks(organization_id, status);
CREATE INDEX idx_ai_agent_tasks_agent ON ai_agent_tasks(agent_id, status);
CREATE INDEX idx_ai_agent_tasks_user ON ai_agent_tasks(user_id, status);
CREATE INDEX idx_ai_agent_tasks_priority ON ai_agent_tasks(organization_id, priority, status)
    WHERE status = 'pending';

-- ============================================
-- AI AGENT INSIGHTS (Proactive Suggestions)
-- ============================================

CREATE TABLE ai_agent_insights (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    agent_id uuid NOT NULL REFERENCES ai_team_members(id),
    insight_type varchar(50) NOT NULL,
    title varchar(255) NOT NULL,
    summary text NOT NULL,
    confidence_score integer,
    supporting_data jsonb,
    suggested_actions jsonb,
    status varchar(20) DEFAULT 'new',
    is_favorite boolean DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_agent_insights_type_check CHECK (insight_type IN (
        'opportunity', 'risk', 'trend', 'tip', 'alert', 'recommendation'
    ))
);

COMMENT ON TABLE ai_agent_insights IS
'Proactive insights and suggestions from AI agents to help grow the business';

CREATE INDEX idx_ai_agent_insights_org ON ai_agent_insights(organization_id, status);
CREATE INDEX idx_ai_agent_insights_agent ON ai_agent_insights(agent_id, status);
CREATE INDEX idx_ai_agent_insights_favorite ON ai_agent_insights(organization_id, is_favorite)
    WHERE is_favorite = true;

-- ============================================
-- AI USER PREFERENCES (Personalization)
-- ============================================

CREATE TABLE ai_user_preferences (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    preferred_agents jsonb DEFAULT '[]'::jsonb,
    notification_frequency varchar(20) DEFAULT 'daily',
    communication_style varchar(20) DEFAULT 'friendly',
    insight_depth varchar(20) DEFAULT 'balanced',
    auto_approve_simple_tasks boolean DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT ai_user_preferences_unique UNIQUE(organization_id, user_id),
    CONSTRAINT ai_user_preferences_frequency_check CHECK (notification_frequency IN (
        'real_time', 'daily', 'weekly', 'manual'
    ))
);

COMMENT ON TABLE ai_user_preferences IS
'User preferences for how AI agents interact and communicate';

CREATE INDEX idx_ai_user_preferences_user ON ai_user_preferences(user_id);

-- ============================================
-- TRIGGERS FOR AUTOMATIC UPDATES
-- ============================================

CREATE OR REPLACE FUNCTION update_ai_team_members_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_ai_team_members_updated_at
    BEFORE UPDATE ON ai_team_members
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_team_members_updated_at();

CREATE TRIGGER set_ai_agent_tasks_updated_at
    BEFORE UPDATE ON ai_agent_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_team_members_updated_at();

CREATE TRIGGER set_ai_agent_insights_updated_at
    BEFORE UPDATE ON ai_agent_insights
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_team_members_updated_at();

CREATE TRIGGER set_ai_user_preferences_updated_at
    BEFORE UPDATE ON ai_user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_ai_team_members_updated_at();

-- ============================================
-- RLS POLICIES (Security)
-- ============================================

ALTER TABLE ai_team_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_agent_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_agent_insights ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_user_preferences ENABLE ROW LEVEL SECURITY;

-- AI Team Members Policies
CREATE POLICY "Users can view their org AI team members"
    ON ai_team_members FOR SELECT
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid())
    ));

CREATE POLICY "Admins can manage AI team members"
    ON ai_team_members FOR ALL
    USING (organization_id IN (
        SELECT organization_id FROM organization_users
        WHERE user_id = (SELECT auth.uid()) AND role IN ('owner', 'admin')
    ));

-- AI Agent Tasks Policies
CREATE POLICY "Users can view their AI tasks"
    ON ai_agent_tasks FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM organization_users
            WHERE user_id = (SELECT auth.uid())
        )
    );

CREATE POLICY "Users can manage their own AI tasks"
    ON ai_agent_tasks FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM organization_users
            WHERE user_id = (SELECT auth.uid())
        ) AND (
            user_id = (SELECT auth.uid()) OR
            user_id IS NULL
        )
    );

-- AI Agent Insights Policies
CREATE POLICY "Users can view their AI insights"
    ON ai_agent_insights FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM organization_users
            WHERE user_id = (SELECT auth.uid())
        )
    );

CREATE POLICY "Users can manage their AI insights"
    ON ai_agent_insights FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM organization_users
            WHERE user_id = (SELECT auth.uid())
        )
    );

-- AI User Preferences Policies
CREATE POLICY "Users can manage their own AI preferences"
    ON ai_user_preferences FOR ALL
    USING (
        user_id = (SELECT auth.uid())
    );

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Create default AI team members for new organizations
CREATE OR REPLACE FUNCTION create_default_ai_team(p_organization_id uuid)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    -- Alex: Sales Assistant
    INSERT INTO ai_team_members (
        organization_id, name, role, description, personality_traits, capabilities
    ) VALUES (
        p_organization_id,
        'Alex',
        'sales_assistant',
        'Your friendly sales assistant who helps manage customer relationships',
        '{"tone": "friendly", "style": "proactive", "focus": "customer_success"}',
        '["lead_followup", "customer_analysis", "sales_opportunities"]'::jsonb
    );
    
    -- Taylor: Inventory Manager
    INSERT INTO ai_team_members (
        organization_id, name, role, description, personality_traits, capabilities
    ) VALUES (
        p_organization_id,
        'Taylor',
        'inventory_manager',
        'Your detail-oriented inventory manager who keeps stock optimized',
        '{"tone": "helpful", "style": "analytical", "focus": "efficiency"}',
        '["stock_alerts", "demand_forecasting", "bundle_suggestions"]'::jsonb
    );
    
    -- Jordan: Financial Advisor
    INSERT INTO ai_team_members (
        organization_id, name, role, description, personality_traits, capabilities
    ) VALUES (
        p_organization_id,
        'Jordan',
        'financial_advisor',
        'Your conservative financial advisor who watches cash flow',
        '{"tone": "professional", "style": "cautious", "focus": "financial_health"}',
        '["cash_flow_analysis", "expense_monitoring", "tax_tips"]'::jsonb
    );
    
    -- Casey: Business Strategist
    INSERT INTO ai_team_members (
        organization_id, name, role, description, personality_traits, capabilities
    ) VALUES (
        p_organization_id,
        'Casey',
        'business_strategist',
        'Your visionary strategist who helps grow your business',
        '{"tone": "inspiring", "style": "visionary", "focus": "growth"}',
        '["market_opportunities", "product_strategy", "competitive_analysis"]'::jsonb
    );
END;
$$;

COMMENT ON FUNCTION create_default_ai_team IS
'Creates default AI team members for a new organization';

-- Get daily briefing from AI team
CREATE OR REPLACE FUNCTION get_ai_daily_briefing(p_user_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_org_id uuid;
    v_briefing jsonb;
BEGIN
    -- Get user's organization
    SELECT organization_id INTO v_org_id
    FROM organization_users
    WHERE user_id = p_user_id
    LIMIT 1;
    
    -- Build simple briefing
    v_briefing := jsonb_build_object(
        'team_updates', (
            SELECT jsonb_agg(jsonb_build_object(
                'agent_name', am.name,
                'agent_role', am.role,
                'message', at.title,
                'priority', at.priority,
                'task_id', at.id
            ))
            FROM ai_agent_tasks at
            JOIN ai_team_members am ON at.agent_id = am.id
            WHERE at.organization_id = v_org_id
              AND at.status = 'pending'
              AND at.user_id IS NULL
            ORDER BY at.priority DESC, at.created_at
            LIMIT 5
        ),
        'high_priority_tasks', (
            SELECT jsonb_agg(jsonb_build_object(
                'agent_name', am.name,
                'message', at.title,
                'due_date', at.scheduled_for
            ))
            FROM ai_agent_tasks at
            JOIN ai_team_members am ON at.agent_id = am.id
            WHERE at.organization_id = v_org_id
              AND at.status = 'pending'
              AND at.priority >= 8
        ),
        'new_insights', (
            SELECT jsonb_agg(jsonb_build_object(
                'agent_name', am.name,
                'insight_type', ai.insight_type,
                'title', ai.title,
                'confidence', ai.confidence_score
            ))
            FROM ai_agent_insights ai
            JOIN ai_team_members am ON ai.agent_id = am.id
            WHERE ai.organization_id = v_org_id
              AND ai.status = 'new'
            ORDER BY ai.confidence_score DESC
            LIMIT 3
        )
    );
    
    RETURN v_briefing;
END;
$$;

COMMENT ON FUNCTION get_ai_daily_briefing IS
'Get a simple daily briefing from your AI team members';

-- ============================================
-- GRANT PERMISSIONS
-- ============================================

GRANT SELECT, INSERT, UPDATE ON ai_team_members TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON ai_agent_tasks TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON ai_agent_insights TO authenticated;
GRANT ALL ON ai_user_preferences TO authenticated;

GRANT ALL ON ai_team_members TO service_role;
GRANT ALL ON ai_agent_tasks TO service_role;
GRANT ALL ON ai_agent_insights TO service_role;
GRANT ALL ON ai_user_preferences TO service_role;

GRANT EXECUTE ON FUNCTION create_default_ai_team(uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION get_ai_daily_briefing(uuid) TO authenticated;

-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- 1. Create default AI team for a new organization
SELECT create_default_ai_team('your-org-id');

-- 2. Get daily briefing
SELECT get_ai_daily_briefing('your-user-id');

-- 3. Create a simple task for an agent
INSERT INTO ai_agent_tasks (
    organization_id, agent_id, task_type, title, description, 
    related_entity_type, related_entity_id, priority
) VALUES (
    'your-org-id',
    (SELECT id FROM ai_team_members WHERE role = 'sales_assistant' LIMIT 1),
    'reminder',
    'Follow up with John Smith',
    'John hasn''t responded to the proposal sent 5 days ago',
    'contact',
    'contact-id-here',
    7
);

-- 4. Create an insight
INSERT INTO ai_agent_insights (
    organization_id, agent_id, insight_type, title, summary, 
    confidence_score, suggested_actions
) VALUES (
    'your-org-id',
    (SELECT id FROM ai_team_members WHERE role = 'inventory_manager' LIMIT 1),
    'opportunity',
    'Bundle opportunity detected',
    'Customers who buy Product A often buy Product B within 30 days',
    85,
    '[{"action": "Create bundle: Product A + Product B", "estimated_impact": "15% revenue increase"}]'::jsonb
);

-- 5. Set user preferences
INSERT INTO ai_user_preferences (
    organization_id, user_id, notification_frequency, communication_style
) VALUES (
    'your-org-id',
    'your-user-id',
    'daily',
    'friendly'
);
*/
