-- Migration: Enhanced Contextual Help System
-- Description: AI-powered contextual help that understands user context and provides relevant guidance
-- Created: 2025-01-01

-- =====================================================
-- SCREEN CONTEXT DEFINITIONS
-- =====================================================

-- Define all screens/routes in the application for context-aware help
CREATE TABLE app_screens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    screen_key text NOT NULL UNIQUE, -- e.g., 'sales_order_create', 'invoice_list'
    screen_name text NOT NULL,
    screen_path text, -- Route path
    module text NOT NULL, -- 'crm', 'sales', 'accounting', 'inventory'
    description text,
    common_tasks jsonb DEFAULT '[]'::jsonb, -- List of common tasks on this screen
    related_screens text[], -- Related screen keys
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT app_screens_module_check
        CHECK (module IN ('crm', 'sales', 'accounting', 'inventory', 'hr', 'project', 'settings'))
);

-- Pre-populate common screens
INSERT INTO app_screens (screen_key, screen_name, screen_path, module, description, common_tasks) VALUES
('sales_order_create', 'Create Sales Order', '/sales/orders/new', 'sales', 'Create a new sales order',
    '["Select customer", "Add products", "Set pricing", "Confirm order"]'::jsonb),
('sales_order_list', 'Sales Orders', '/sales/orders', 'sales', 'View and manage sales orders',
    '["Filter orders", "Export data", "Bulk actions", "Create new order"]'::jsonb),
('invoice_create', 'Create Invoice', '/accounting/invoices/new', 'accounting', 'Create a new invoice',
    '["Select customer", "Add line items", "Apply taxes", "Set payment terms"]'::jsonb),
('customer_create', 'Create Customer', '/crm/customers/new', 'crm', 'Add a new customer',
    '["Enter contact info", "Set payment terms", "Add tags", "Assign sales team"]'::jsonb),
('product_create', 'Create Product', '/inventory/products/new', 'inventory', 'Add a new product',
    '["Set product type", "Configure pricing", "Set inventory tracking", "Upload images"]'::jsonb),
('dashboard', 'Dashboard', '/dashboard', 'sales', 'Main dashboard overview',
    '["View analytics", "Quick actions", "Recent activity"]'::jsonb);

-- =====================================================
-- ENHANCED KNOWLEDGE BASE FUNCTIONS
-- =====================================================

-- Get contextual help based on current screen and optional user query
CREATE OR REPLACE FUNCTION get_contextual_help(
    p_organization_id uuid,
    p_screen_key text,
    p_user_query text DEFAULT NULL,
    p_user_id uuid DEFAULT NULL,
    p_limit integer DEFAULT 5
)
RETURNS TABLE (
    entry_id uuid,
    title text,
    summary text,
    body_markdown text,
    relevance_score float,
    space_name text,
    tags text[]
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_screen_info record;
    v_query_embedding vector(768);
    v_screen_embedding vector(768);
    v_combined_embedding vector(768);
BEGIN
    -- Get screen information
    SELECT * INTO v_screen_info
    FROM app_screens
    WHERE screen_key = p_screen_key;

    IF NOT FOUND THEN
        -- If screen not found, just search by query if provided
        IF p_user_query IS NOT NULL THEN
            v_query_embedding := generate_search_embedding(p_user_query);

            RETURN QUERY
            SELECT
                ke.id,
                ke.title,
                ke.summary,
                ke.body_markdown,
                1 - (ke.search_embedding <=> v_query_embedding) as relevance_score,
                ks.name as space_name,
                ARRAY(
                    SELECT kt.name
                    FROM knowledge_entry_tags ket
                    JOIN knowledge_tags kt ON kt.id = ket.tag_id
                    WHERE ket.entry_id = ke.id
                ) as tags
            FROM knowledge_entries ke
            JOIN knowledge_spaces ks ON ks.id = ke.space_id
            WHERE ke.organization_id = p_organization_id
                AND ke.is_published = true
                AND ke.search_embedding IS NOT NULL
            ORDER BY relevance_score DESC
            LIMIT p_limit;
            RETURN;
        END IF;

        RAISE EXCEPTION 'Screen key % not found and no query provided', p_screen_key;
    END IF;

    -- Generate embeddings for screen context
    v_screen_embedding := generate_search_embedding(
        COALESCE(v_screen_info.screen_name, '') || ' ' ||
        COALESCE(v_screen_info.description, '') || ' ' ||
        COALESCE(v_screen_info.module, '')
    );

    -- If user provided a query, combine screen context with query
    IF p_user_query IS NOT NULL AND p_user_query != '' THEN
        v_query_embedding := generate_search_embedding(p_user_query);

        -- Weight: 70% query, 30% screen context
        v_combined_embedding := (
            SELECT array_agg(
                (v_query_val * 0.7 + v_screen_val * 0.3)::real
            )::vector(768)
            FROM unnest(v_query_embedding::real[]) WITH ORDINALITY AS t1(v_query_val, idx)
            JOIN unnest(v_screen_embedding::real[]) WITH ORDINALITY AS t2(v_screen_val, idx2)
                ON t1.idx = t2.idx2
        );
    ELSE
        -- No query, use screen context only
        v_combined_embedding := v_screen_embedding;
    END IF;

    -- Search knowledge base
    RETURN QUERY
    SELECT
        ke.id,
        ke.title,
        ke.summary,
        ke.body_markdown,
        1 - (ke.search_embedding <=> v_combined_embedding) as relevance_score,
        ks.name as space_name,
        ARRAY(
            SELECT kt.name
            FROM knowledge_entry_tags ket
            JOIN knowledge_tags kt ON kt.id = ket.tag_id
            WHERE ket.entry_id = ke.id
        ) as tags
    FROM knowledge_entries ke
    JOIN knowledge_spaces ks ON ks.id = ke.space_id
    WHERE ke.organization_id = p_organization_id
        AND ke.is_published = true
        AND ke.search_embedding IS NOT NULL
        -- Optional: Filter by module
        AND (
            ke.metadata->>'module' = v_screen_info.module
            OR ke.metadata->>'module' IS NULL
        )
    ORDER BY relevance_score DESC
    LIMIT p_limit;

    -- Log usage for analytics (async, don't block)
    IF p_user_id IS NOT NULL THEN
        INSERT INTO contextual_help_usage (
            organization_id,
            user_id,
            screen_context,
            query_text
        ) VALUES (
            p_organization_id,
            p_user_id,
            p_screen_key,
            p_user_query
        );
    END IF;
END;
$$;

COMMENT ON FUNCTION get_contextual_help IS 'Get context-aware help based on current screen and optional user query';

-- =====================================================
-- QUICK ACTIONS / GUIDED FLOWS
-- =====================================================

-- Store guided walkthroughs for complex tasks
CREATE TABLE guided_flows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    flow_key text NOT NULL,
    title text NOT NULL,
    description text,
    screen_key text REFERENCES app_screens(screen_key),
    steps jsonb NOT NULL, -- Array of step objects
    prerequisites jsonb DEFAULT '[]'::jsonb,
    estimated_duration_minutes integer,
    difficulty text DEFAULT 'beginner',
    is_active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT guided_flows_unique UNIQUE(organization_id, flow_key),
    CONSTRAINT guided_flows_difficulty_check
        CHECK (difficulty IN ('beginner', 'intermediate', 'advanced'))
);

-- Track user progress through guided flows
CREATE TABLE user_flow_progress (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    flow_id uuid NOT NULL REFERENCES guided_flows(id) ON DELETE CASCADE,
    current_step_index integer DEFAULT 0,
    completed_steps integer[] DEFAULT ARRAY[]::integer[],
    status text DEFAULT 'in_progress',
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    abandoned_at timestamptz,

    CONSTRAINT user_flow_progress_status_check
        CHECK (status IN ('in_progress', 'completed', 'abandoned'))
);

CREATE INDEX idx_user_flow_progress_user
ON user_flow_progress(organization_id, user_id, status);

-- =====================================================
-- SMART SUGGESTIONS BASED ON CONTEXT
-- =====================================================

-- Suggest next actions based on current context
CREATE OR REPLACE FUNCTION get_smart_suggestions(
    p_organization_id uuid,
    p_screen_key text,
    p_user_id uuid,
    p_current_record_type text DEFAULT NULL,
    p_current_record_id uuid DEFAULT NULL,
    p_limit integer DEFAULT 5
)
RETURNS TABLE (
    suggestion_type text,
    suggestion_title text,
    suggestion_description text,
    action_type text,
    action_data jsonb,
    priority integer
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_screen_info record;
BEGIN
    -- Get screen info
    SELECT * INTO v_screen_info
    FROM app_screens
    WHERE screen_key = p_screen_key;

    -- Return suggestions based on context
    -- This is a framework - customize based on your business logic

    RETURN QUERY
    SELECT *
    FROM (
        -- Suggest incomplete flows
        (
            SELECT
                'guided_flow'::text,
                gf.title,
                'Continue where you left off: ' || gf.description,
                'resume_flow'::text,
                jsonb_build_object(
                    'flow_id', gf.id,
                    'flow_key', gf.flow_key,
                    'current_step', ufp.current_step_index
                ),
                1 as priority
            FROM user_flow_progress ufp
            JOIN guided_flows gf ON gf.id = ufp.flow_id
            WHERE ufp.organization_id = p_organization_id
                AND ufp.user_id = p_user_id
                AND ufp.status = 'in_progress'
                AND (gf.screen_key = p_screen_key OR gf.screen_key IS NULL)
            ORDER BY ufp.started_at DESC
            LIMIT 2
        )

        UNION ALL

        -- Suggest relevant knowledge articles
        (
            SELECT
                'knowledge_article'::text,
                ke.title,
                COALESCE(ke.summary, 'Learn more about this topic'),
                'open_article'::text,
                jsonb_build_object(
                    'entry_id', ke.id,
                    'space_id', ke.space_id
                ),
                2 as priority
            FROM knowledge_entries ke
            WHERE ke.organization_id = p_organization_id
                AND ke.is_published = true
                AND ke.metadata->>'module' = v_screen_info.module
            ORDER BY ke.updated_at DESC
            LIMIT 2
        )
    ) AS suggestions
    ORDER BY priority, suggestion_title
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION get_smart_suggestions IS 'Get context-aware smart suggestions for next actions';

-- =====================================================
-- FEEDBACK & ANALYTICS
-- =====================================================

-- Track help article feedback
CREATE OR REPLACE FUNCTION record_help_feedback(
    p_organization_id uuid,
    p_user_id uuid,
    p_knowledge_entry_id uuid,
    p_was_helpful boolean,
    p_screen_context text DEFAULT NULL
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO contextual_help_usage (
        organization_id,
        user_id,
        screen_context,
        knowledge_entry_id,
        was_helpful
    ) VALUES (
        p_organization_id,
        p_user_id,
        p_screen_context,
        p_knowledge_entry_id,
        p_was_helpful
    );
END;
$$;

-- Get help effectiveness metrics
CREATE OR REPLACE FUNCTION get_help_effectiveness_metrics(
    p_organization_id uuid,
    p_days_back integer DEFAULT 30
)
RETURNS TABLE (
    entry_id uuid,
    entry_title text,
    total_views integer,
    helpful_count integer,
    not_helpful_count integer,
    helpfulness_ratio float
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ke.id,
        ke.title,
        COUNT(chu.id)::integer as total_views,
        COUNT(chu.id) FILTER (WHERE chu.was_helpful = true)::integer,
        COUNT(chu.id) FILTER (WHERE chu.was_helpful = false)::integer,
        CASE
            WHEN COUNT(chu.id) FILTER (WHERE chu.was_helpful IS NOT NULL) > 0
            THEN COUNT(chu.id) FILTER (WHERE chu.was_helpful = true)::float /
                 COUNT(chu.id) FILTER (WHERE chu.was_helpful IS NOT NULL)::float
            ELSE NULL
        END as helpfulness_ratio
    FROM knowledge_entries ke
    LEFT JOIN contextual_help_usage chu ON chu.knowledge_entry_id = ke.id
        AND chu.created_at > now() - (p_days_back || ' days')::interval
    WHERE ke.organization_id = p_organization_id
    GROUP BY ke.id, ke.title
    HAVING COUNT(chu.id) > 0
    ORDER BY total_views DESC, helpfulness_ratio DESC NULLS LAST;
END;
$$;

COMMENT ON FUNCTION get_help_effectiveness_metrics IS 'Analyze which help articles are most viewed and helpful';

-- =====================================================
-- EXAMPLE GUIDED FLOWS (Sample Data)
-- =====================================================

-- This would be populated by your application or seeded
COMMENT ON TABLE guided_flows IS 'Step-by-step guided walkthroughs for complex tasks';
COMMENT ON TABLE user_flow_progress IS 'Track user progress through guided flows';
COMMENT ON TABLE app_screens IS 'Registry of all application screens for contextual help routing';

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_app_screens_module ON app_screens(module);
CREATE INDEX idx_guided_flows_org_screen ON guided_flows(organization_id, screen_key);
CREATE INDEX idx_guided_flows_active ON guided_flows(organization_id) WHERE is_active = true;
