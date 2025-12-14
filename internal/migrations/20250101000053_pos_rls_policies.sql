-- Migration: POS RLS Policies
-- Description: Row-level security policies for all POS tables
-- Created: 2025-01-26
-- Dependencies: 20250101000052_pos_queue_handlers.sql

-- =====================================================
-- ENABLE RLS ON ALL POS TABLES
-- =====================================================

ALTER TABLE pos_payment_methods ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_cash_movements ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_order_discounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_inventory_alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE pos_pricing_overrides ENABLE ROW LEVEL SECURITY;

-- =====================================================
-- POS PAYMENT METHODS POLICIES
-- =====================================================

CREATE POLICY pos_payment_methods_org_isolation ON pos_payment_methods
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_payment_methods_select ON pos_payment_methods
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_payment_methods_insert ON pos_payment_methods
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_payment_methods_update ON pos_payment_methods
    FOR UPDATE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_payment_methods_delete ON pos_payment_methods
    FOR DELETE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

-- =====================================================
-- POS CONFIG POLICIES
-- =====================================================

CREATE POLICY pos_config_org_isolation ON pos_config
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_config_select ON pos_config
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_config_insert ON pos_config
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_config_update ON pos_config
    FOR UPDATE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_config_delete ON pos_config
    FOR DELETE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

-- =====================================================
-- POS SESSIONS POLICIES
-- =====================================================

CREATE POLICY pos_sessions_org_isolation ON pos_sessions
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_sessions_select ON pos_sessions
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        -- Optional: restrict to own sessions unless manager
        -- AND (user_id = auth.uid() OR user_has_role('pos_manager'))
    );

CREATE POLICY pos_sessions_insert ON pos_sessions
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_sessions_update ON pos_sessions
    FOR UPDATE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        -- Can only update own session unless manager
        AND (user_id = auth.uid() OR user_has_role('pos_manager'))
    );

CREATE POLICY pos_sessions_delete ON pos_sessions
    FOR DELETE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        AND user_has_role('pos_manager')  -- Only managers can delete sessions
    );

-- =====================================================
-- POS CASH MOVEMENTS POLICIES
-- =====================================================

CREATE POLICY pos_cash_movements_org_isolation ON pos_cash_movements
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_cash_movements_select ON pos_cash_movements
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        -- Can see movements from own sessions or if manager
        AND (
            EXISTS (
                SELECT 1 FROM pos_sessions
                WHERE id = pos_cash_movements.session_id
                  AND user_id = auth.uid()
            )
            OR user_has_role('pos_manager')
        )
    );

CREATE POLICY pos_cash_movements_insert ON pos_cash_movements
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        -- Can only add to own open sessions
        AND EXISTS (
            SELECT 1 FROM pos_sessions
            WHERE id = session_id
              AND user_id = auth.uid()
              AND state IN ('opened', 'closing_control')
        )
    );

-- =====================================================
-- POS PAYMENTS POLICIES
-- =====================================================

CREATE POLICY pos_payments_org_isolation ON pos_payments
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_payments_select ON pos_payments
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_payments_insert ON pos_payments
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_payments_update ON pos_payments
    FOR UPDATE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        AND user_has_role('pos_manager')  -- Only managers can modify payments
    );

-- =====================================================
-- POS ORDER DISCOUNTS POLICIES
-- =====================================================

CREATE POLICY pos_order_discounts_org_isolation ON pos_order_discounts
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_order_discounts_select ON pos_order_discounts
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_order_discounts_insert ON pos_order_discounts
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

-- =====================================================
-- POS INVENTORY ALERTS POLICIES
-- =====================================================

CREATE POLICY pos_inventory_alerts_org_isolation ON pos_inventory_alerts
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_inventory_alerts_select ON pos_inventory_alerts
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_inventory_alerts_insert ON pos_inventory_alerts
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_inventory_alerts_update ON pos_inventory_alerts
    FOR UPDATE
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
        -- Only managers can acknowledge/resolve alerts
        AND (state = 'open' OR user_has_role('inventory_manager'))
    );

-- =====================================================
-- POS PRICING OVERRIDES POLICIES
-- =====================================================

CREATE POLICY pos_pricing_overrides_org_isolation ON pos_pricing_overrides
    FOR ALL
    USING (organization_id = get_current_organization_id());

CREATE POLICY pos_pricing_overrides_select ON pos_pricing_overrides
    FOR SELECT
    USING (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

CREATE POLICY pos_pricing_overrides_insert ON pos_pricing_overrides
    FOR INSERT
    WITH CHECK (
        organization_id = get_current_organization_id()
        AND user_has_org_access(organization_id)
    );

-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

-- Grant access to authenticated users
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_payment_methods TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_config TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_sessions TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_cash_movements TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_payments TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_order_discounts TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_inventory_alerts TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON pos_pricing_overrides TO authenticated;

-- Grant access to service role for background jobs
GRANT ALL ON pos_payment_methods TO service_role;
GRANT ALL ON pos_config TO service_role;
GRANT ALL ON pos_sessions TO service_role;
GRANT ALL ON pos_cash_movements TO service_role;
GRANT ALL ON pos_payments TO service_role;
GRANT ALL ON pos_order_discounts TO service_role;
GRANT ALL ON pos_inventory_alerts TO service_role;
GRANT ALL ON pos_pricing_overrides TO service_role;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON POLICY pos_sessions_select ON pos_sessions IS 'Users can view sessions in their organization';
COMMENT ON POLICY pos_sessions_update ON pos_sessions IS 'Users can only update their own sessions unless they are managers';
COMMENT ON POLICY pos_cash_movements_insert ON pos_cash_movements IS 'Users can only add cash movements to their own open sessions';
COMMENT ON POLICY pos_inventory_alerts_update ON pos_inventory_alerts IS 'Only inventory managers can acknowledge or resolve alerts';
