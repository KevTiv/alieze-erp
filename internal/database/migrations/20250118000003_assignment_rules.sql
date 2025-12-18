-- Migration: Assignment Rules System
-- Description: Lead assignment rules with round-robin, weighted, and territory-based assignment
-- Version: 20250118000003

-- ============================================================================
-- Assignment Rules Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS assignment_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(200) NOT NULL,
    description text,

    -- Rule type: round_robin, weighted, territory, custom
    rule_type varchar(50) NOT NULL CHECK (rule_type IN ('round_robin', 'weighted', 'territory', 'custom')),

    -- Target entity type: leads, contacts, opportunities
    target_model varchar(100) NOT NULL,

    -- Priority (higher = evaluated first)
    priority int DEFAULT 0,

    -- Active status
    is_active boolean DEFAULT true,

    -- Conditions for rule matching (JSONB)
    -- Example: [{"field": "country", "operator": "=", "value": "USA"}]
    conditions jsonb DEFAULT '[]'::jsonb,

    -- Assignment configuration (JSONB)
    -- Round Robin: {"users": ["uuid1", "uuid2"], "current_index": 0}
    -- Weighted: {"assignments": [{"user_id": "uuid", "weight": 5}, ...]}
    -- Territory: {"territories": [{"name": "West", "users": ["uuid"], "conditions": {...}}]}
    -- Custom: {"logic": "custom", "params": {...}}
    assignment_config jsonb NOT NULL,

    -- Assignment pool (users or teams eligible for assignment)
    assign_to_type varchar(50) NOT NULL CHECK (assign_to_type IN ('user', 'team')),

    -- Maximum assignments per user (0 = unlimited)
    max_assignments_per_user int DEFAULT 0,

    -- Time-based assignment window
    assignment_window_start time,
    assignment_window_end time,

    -- Days of week when rule is active (1=Monday, 7=Sunday)
    active_days int[] DEFAULT ARRAY[1,2,3,4,5,6,7],

    -- Metadata
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    created_by uuid NOT NULL REFERENCES users(id),
    updated_by uuid NOT NULL REFERENCES users(id)
);

-- Indexes
CREATE INDEX idx_assignment_rules_organization ON assignment_rules(organization_id);
CREATE INDEX idx_assignment_rules_target_model ON assignment_rules(target_model);
CREATE INDEX idx_assignment_rules_active ON assignment_rules(is_active) WHERE is_active = true;
CREATE INDEX idx_assignment_rules_priority ON assignment_rules(priority DESC);
CREATE INDEX idx_assignment_rules_conditions ON assignment_rules USING GIN (conditions);

-- ============================================================================
-- Assignment History Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS assignment_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- The assignment rule that was applied
    rule_id uuid REFERENCES assignment_rules(id) ON DELETE SET NULL,
    rule_name varchar(200), -- Stored for history even if rule deleted

    -- Target record that was assigned
    target_model varchar(100) NOT NULL,
    target_id uuid NOT NULL,
    target_name varchar(255), -- For display

    -- Assignment details
    assigned_to_type varchar(50) NOT NULL CHECK (assigned_to_type IN ('user', 'team')),
    assigned_to_id uuid NOT NULL,
    assigned_to_name varchar(255),

    -- Previous assignment (for reassignment tracking)
    previous_assigned_to_id uuid,
    previous_assigned_to_name varchar(255),

    -- Assignment reason/trigger
    assignment_reason varchar(100), -- 'auto', 'manual', 'reassignment', 'load_balancing'

    -- Additional context
    metadata jsonb DEFAULT '{}'::jsonb,

    -- Timestamps
    assigned_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    assigned_by uuid REFERENCES users(id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_assignment_history_organization ON assignment_history(organization_id);
CREATE INDEX idx_assignment_history_rule ON assignment_history(rule_id);
CREATE INDEX idx_assignment_history_target ON assignment_history(target_model, target_id);
CREATE INDEX idx_assignment_history_assigned_to ON assignment_history(assigned_to_id);
CREATE INDEX idx_assignment_history_date ON assignment_history(assigned_at DESC);

-- ============================================================================
-- User Assignment Load Table (for tracking current assignments)
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_assignment_load (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Target model this load applies to
    target_model varchar(100) NOT NULL,

    -- Current assignment counts
    active_assignments int DEFAULT 0,
    total_assignments int DEFAULT 0,

    -- Last assignment timestamp
    last_assigned_at timestamptz,

    -- Capacity and weights
    max_capacity int DEFAULT 0, -- 0 = unlimited
    weight int DEFAULT 1, -- For weighted assignment

    -- Availability
    is_available boolean DEFAULT true,
    unavailable_until timestamptz,

    -- Stats
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id, user_id, target_model)
);

-- Indexes
CREATE INDEX idx_user_assignment_load_user ON user_assignment_load(user_id);
CREATE INDEX idx_user_assignment_load_organization ON user_assignment_load(organization_id);
CREATE INDEX idx_user_assignment_load_model ON user_assignment_load(target_model);
CREATE INDEX idx_user_assignment_load_available ON user_assignment_load(is_available) WHERE is_available = true;

-- ============================================================================
-- Territory Definitions Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS territories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(200) NOT NULL,
    description text,

    -- Territory type: geographic, industry, size, custom
    territory_type varchar(50) NOT NULL,

    -- Territory matching conditions (JSONB)
    -- Example: {"country": "USA", "state": ["CA", "OR", "WA"]}
    conditions jsonb NOT NULL,

    -- Assigned users/teams
    assigned_users uuid[] DEFAULT ARRAY[]::uuid[],
    assigned_teams uuid[] DEFAULT ARRAY[]::uuid[],

    -- Priority (higher priority territories matched first)
    priority int DEFAULT 0,

    -- Active status
    is_active boolean DEFAULT true,

    -- Metadata
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    created_by uuid NOT NULL REFERENCES users(id),
    updated_by uuid NOT NULL REFERENCES users(id),

    UNIQUE(organization_id, name)
);

-- Indexes
CREATE INDEX idx_territories_organization ON territories(organization_id);
CREATE INDEX idx_territories_active ON territories(is_active) WHERE is_active = true;
CREATE INDEX idx_territories_conditions ON territories USING GIN (conditions);

-- ============================================================================
-- SQL Functions
-- ============================================================================

-- Function: Get next user for round-robin assignment
CREATE OR REPLACE FUNCTION get_next_round_robin_user(
    p_rule_id uuid
) RETURNS uuid AS $$
DECLARE
    v_config jsonb;
    v_users jsonb;
    v_current_index int;
    v_user_count int;
    v_next_user_id uuid;
BEGIN
    -- Get rule configuration
    SELECT assignment_config INTO v_config
    FROM assignment_rules
    WHERE id = p_rule_id;

    -- Extract users array and current index
    v_users := v_config->'users';
    v_current_index := COALESCE((v_config->>'current_index')::int, 0);
    v_user_count := jsonb_array_length(v_users);

    IF v_user_count = 0 THEN
        RETURN NULL;
    END IF;

    -- Get next user
    v_next_user_id := (v_users->v_current_index)::text::uuid;

    -- Update index (wrap around)
    v_current_index := (v_current_index + 1) % v_user_count;

    -- Update rule configuration
    UPDATE assignment_rules
    SET assignment_config = jsonb_set(
        assignment_config,
        '{current_index}',
        to_jsonb(v_current_index)
    ),
    updated_at = CURRENT_TIMESTAMP
    WHERE id = p_rule_id;

    RETURN v_next_user_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Get weighted user assignment
CREATE OR REPLACE FUNCTION get_weighted_user(
    p_rule_id uuid,
    p_target_model varchar(100)
) RETURNS uuid AS $$
DECLARE
    v_config jsonb;
    v_assignments jsonb;
    v_assignment jsonb;
    v_user_id uuid;
    v_weight int;
    v_current_load int;
    v_weighted_load numeric;
    v_best_user_id uuid;
    v_best_score numeric := 999999;
    v_score numeric;
BEGIN
    -- Get rule configuration
    SELECT assignment_config INTO v_config
    FROM assignment_rules
    WHERE id = p_rule_id;

    v_assignments := v_config->'assignments';

    -- Loop through assignments and find user with lowest weighted load
    FOR v_assignment IN SELECT * FROM jsonb_array_elements(v_assignments)
    LOOP
        v_user_id := (v_assignment->>'user_id')::uuid;
        v_weight := COALESCE((v_assignment->>'weight')::int, 1);

        -- Get current load
        SELECT COALESCE(active_assignments, 0) INTO v_current_load
        FROM user_assignment_load
        WHERE user_id = v_user_id
          AND target_model = p_target_model
          AND is_available = true;

        -- Calculate weighted score (lower is better)
        v_weighted_load := COALESCE(v_current_load, 0)::numeric / NULLIF(v_weight, 0)::numeric;

        IF v_weighted_load < v_best_score THEN
            v_best_score := v_weighted_load;
            v_best_user_id := v_user_id;
        END IF;
    END LOOP;

    RETURN v_best_user_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Match territory based on conditions
CREATE OR REPLACE FUNCTION match_territory(
    p_organization_id uuid,
    p_conditions jsonb
) RETURNS uuid AS $$
DECLARE
    v_territory_id uuid;
BEGIN
    -- Find matching territory with highest priority
    -- This is a simplified version - real implementation would need complex JSONB matching
    SELECT id INTO v_territory_id
    FROM territories
    WHERE organization_id = p_organization_id
      AND is_active = true
      AND conditions @> p_conditions -- JSONB containment operator
    ORDER BY priority DESC
    LIMIT 1;

    RETURN v_territory_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Record assignment in history
CREATE OR REPLACE FUNCTION record_assignment(
    p_organization_id uuid,
    p_rule_id uuid,
    p_target_model varchar(100),
    p_target_id uuid,
    p_assigned_to_id uuid,
    p_assigned_to_type varchar(50),
    p_reason varchar(100),
    p_assigned_by uuid DEFAULT NULL
) RETURNS uuid AS $$
DECLARE
    v_history_id uuid;
    v_rule_name varchar(200);
BEGIN
    -- Get rule name
    SELECT name INTO v_rule_name
    FROM assignment_rules
    WHERE id = p_rule_id;

    -- Insert history record
    INSERT INTO assignment_history (
        organization_id,
        rule_id,
        rule_name,
        target_model,
        target_id,
        assigned_to_type,
        assigned_to_id,
        assignment_reason,
        assigned_by
    ) VALUES (
        p_organization_id,
        p_rule_id,
        v_rule_name,
        p_target_model,
        p_target_id,
        p_assigned_to_type,
        p_assigned_to_id,
        p_reason,
        p_assigned_by
    ) RETURNING id INTO v_history_id;

    -- Update user assignment load
    INSERT INTO user_assignment_load (
        organization_id,
        user_id,
        target_model,
        active_assignments,
        total_assignments,
        last_assigned_at
    ) VALUES (
        p_organization_id,
        p_assigned_to_id,
        p_target_model,
        1,
        1,
        CURRENT_TIMESTAMP
    )
    ON CONFLICT (organization_id, user_id, target_model)
    DO UPDATE SET
        active_assignments = user_assignment_load.active_assignments + 1,
        total_assignments = user_assignment_load.total_assignments + 1,
        last_assigned_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP;

    RETURN v_history_id;
END;
$$ LANGUAGE plpgsql;

-- Function: Update assignment load when lead is closed/converted
CREATE OR REPLACE FUNCTION decrement_assignment_load(
    p_user_id uuid,
    p_target_model varchar(100)
) RETURNS void AS $$
BEGIN
    UPDATE user_assignment_load
    SET active_assignments = GREATEST(active_assignments - 1, 0),
        updated_at = CURRENT_TIMESTAMP
    WHERE user_id = p_user_id
      AND target_model = p_target_model;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Views
-- ============================================================================

-- View: Assignment statistics by user
CREATE OR REPLACE VIEW assignment_stats_by_user AS
SELECT
    u.id as user_id,
    u.name as user_name,
    u.email as user_email,
    ual.target_model,
    ual.active_assignments,
    ual.total_assignments,
    ual.last_assigned_at,
    ual.weight,
    ual.is_available,
    COUNT(ah.id) as assignments_today
FROM users u
LEFT JOIN user_assignment_load ual ON u.id = ual.user_id
LEFT JOIN assignment_history ah ON u.id = ah.assigned_to_id
    AND ah.assigned_at >= CURRENT_DATE
GROUP BY u.id, u.name, u.email, ual.target_model, ual.active_assignments,
         ual.total_assignments, ual.last_assigned_at, ual.weight, ual.is_available;

-- View: Assignment rule effectiveness
CREATE OR REPLACE VIEW assignment_rule_effectiveness AS
SELECT
    ar.id as rule_id,
    ar.name as rule_name,
    ar.rule_type,
    ar.target_model,
    ar.is_active,
    COUNT(ah.id) as total_assignments,
    COUNT(CASE WHEN ah.assigned_at >= CURRENT_DATE THEN 1 END) as assignments_today,
    COUNT(CASE WHEN ah.assigned_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as assignments_this_week,
    MAX(ah.assigned_at) as last_used_at,
    COUNT(DISTINCT ah.assigned_to_id) as unique_assignees
FROM assignment_rules ar
LEFT JOIN assignment_history ah ON ar.id = ah.rule_id
GROUP BY ar.id, ar.name, ar.rule_type, ar.target_model, ar.is_active;

-- ============================================================================
-- Triggers
-- ============================================================================

-- Trigger: Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_assignment_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_assignment_rules_updated_at
    BEFORE UPDATE ON assignment_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_assignment_rules_updated_at();

CREATE TRIGGER trg_territories_updated_at
    BEFORE UPDATE ON territories
    FOR EACH ROW
    EXECUTE FUNCTION update_assignment_rules_updated_at();

-- ============================================================================
-- Sample Data (for testing)
-- ============================================================================

-- Insert sample round-robin rule
INSERT INTO assignment_rules (
    organization_id,
    name,
    description,
    rule_type,
    target_model,
    priority,
    assignment_config,
    assign_to_type,
    created_by,
    updated_by
) VALUES (
    (SELECT id FROM organizations LIMIT 1),
    'USA Leads Round Robin',
    'Distribute USA leads evenly across sales team',
    'round_robin',
    'leads',
    10,
    '{"users": [], "current_index": 0}'::jsonb,
    'user',
    (SELECT id FROM users LIMIT 1),
    (SELECT id FROM users LIMIT 1)
) ON CONFLICT DO NOTHING;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE assignment_rules IS 'Configuration for automatic lead/contact assignment';
COMMENT ON TABLE assignment_history IS 'Audit trail of all assignments';
COMMENT ON TABLE user_assignment_load IS 'Current assignment load per user for load balancing';
COMMENT ON TABLE territories IS 'Territory definitions for territory-based assignment';

COMMENT ON COLUMN assignment_rules.rule_type IS 'round_robin, weighted, territory, or custom';
COMMENT ON COLUMN assignment_rules.conditions IS 'JSONB conditions for when rule applies';
COMMENT ON COLUMN assignment_rules.assignment_config IS 'Configuration specific to rule type';
COMMENT ON COLUMN user_assignment_load.weight IS 'Weight for weighted assignment (higher = more assignments)';
