-- Migration: Cycle Counting Module
-- Description: Inventory accuracy through systematic counting
-- Created: 2025-01-01

-- =====================================================
-- CYCLE COUNTING MODULE
-- =====================================================

-- Create cycle count plans table
CREATE TABLE IF NOT EXISTS inventory_cycle_count_plans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    frequency VARCHAR(20) CHECK (frequency IN ('daily', 'weekly', 'monthly', 'quarterly', 'custom')),
    abc_class VARCHAR(10) CHECK (abc_class IN ('A', 'B', 'C', 'all')),
    start_date DATE,
    end_date DATE,
    status VARCHAR(20) DEFAULT 'draft', -- draft, active, completed, cancelled
    priority INTEGER DEFAULT 10,
    created_by uuid,
    assigned_to uuid,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Create cycle count sessions table
CREATE TABLE IF NOT EXISTS inventory_cycle_count_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_id uuid REFERENCES inventory_cycle_count_plans(id),
    name VARCHAR(255) NOT NULL,
    location_id uuid REFERENCES stock_locations(id),
    user_id uuid REFERENCES organization_users(id),
    start_time TIMESTAMPTZ DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'in_progress', -- in_progress, completed, verified, cancelled
    count_method VARCHAR(20) DEFAULT 'manual', -- manual, barcode, mobile
    device_id VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create cycle count lines table
CREATE TABLE IF NOT EXISTS inventory_cycle_count_lines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES inventory_cycle_count_sessions(id) ON DELETE CASCADE,
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products(id),
    product_variant_id uuid REFERENCES product_variants(id),
    location_id uuid NOT NULL REFERENCES stock_locations(id),
    lot_id uuid REFERENCES stock_lots(id),
    package_id uuid REFERENCES stock_packages(id),
    counted_quantity NUMERIC(15,4) NOT NULL,
    system_quantity NUMERIC(15,4) NOT NULL,
    variance NUMERIC(15,4) GENERATED ALWAYS AS (counted_quantity - system_quantity) STORED,
    variance_percentage NUMERIC(10,2) GENERATED ALWAYS AS (
        CASE
            WHEN system_quantity = 0 THEN 0
            ELSE (variance / system_quantity) * 100
        END
    ) STORED,
    uom_id uuid REFERENCES uom_units(id),
    count_time TIMESTAMPTZ DEFAULT NOW(),
    counted_by uuid REFERENCES organization_users(id),
    verified_by uuid REFERENCES organization_users(id),
    verification_time TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'counted', -- counted, verified, adjusted, resolved
    notes TEXT,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create cycle count adjustments table
CREATE TABLE IF NOT EXISTS inventory_cycle_count_adjustments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    count_line_id uuid NOT NULL REFERENCES inventory_cycle_count_lines(id),
    product_id uuid NOT NULL REFERENCES products(id),
    location_id uuid NOT NULL REFERENCES stock_locations(id),
    lot_id uuid REFERENCES stock_lots(id),
    package_id uuid REFERENCES stock_packages(id),
    old_quantity NUMERIC(15,4) NOT NULL,
    new_quantity NUMERIC(15,4) NOT NULL,
    adjustment_quantity NUMERIC(15,4) GENERATED ALWAYS AS (new_quantity - old_quantity) STORED,
    adjustment_type VARCHAR(20) DEFAULT 'variance', -- variance, damage, theft, misplacement
    reason VARCHAR(100),
    adjustment_time TIMESTAMPTZ DEFAULT NOW(),
    adjusted_by uuid REFERENCES organization_users(id),
    approved_by uuid REFERENCES organization_users(id),
    approval_time TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create cycle count accuracy history table
CREATE TABLE IF NOT EXISTS inventory_cycle_count_accuracy (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id uuid REFERENCES products(id),
    location_id uuid REFERENCES stock_locations(id),
    count_date DATE NOT NULL,
    system_quantity NUMERIC(15,4) NOT NULL,
    counted_quantity NUMERIC(15,4) NOT NULL,
    variance NUMERIC(15,4) NOT NULL,
    variance_percentage NUMERIC(10,2) NOT NULL,
    accuracy_score NUMERIC(5,2) GENERATED ALWAYS AS (
        CASE
            WHEN system_quantity = 0 THEN 100
            ELSE 100 - ABS(variance_percentage)
        END
    ) STORED,
    count_method VARCHAR(20),
    counted_by uuid REFERENCES organization_users(id),
    verified_by uuid REFERENCES organization_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for cycle counting tables
CREATE INDEX IF NOT EXISTS idx_cycle_count_plans_org ON inventory_cycle_count_plans(organization_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_plans_status ON inventory_cycle_count_plans(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_cycle_count_sessions_org ON inventory_cycle_count_sessions(organization_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_sessions_status ON inventory_cycle_count_sessions(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_cycle_count_sessions_location ON inventory_cycle_count_sessions(organization_id, location_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_lines_org ON inventory_cycle_count_lines(organization_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_lines_session ON inventory_cycle_count_lines(session_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_lines_product ON inventory_cycle_count_lines(organization_id, product_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_lines_location ON inventory_cycle_count_lines(organization_id, location_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_adjustments_org ON inventory_cycle_count_adjustments(organization_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_adjustments_status ON inventory_cycle_count_adjustments(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_cycle_count_accuracy_org ON inventory_cycle_count_accuracy(organization_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_accuracy_product ON inventory_cycle_count_accuracy(organization_id, product_id);
CREATE INDEX IF NOT EXISTS idx_cycle_count_accuracy_date ON inventory_cycle_count_accuracy(organization_id, count_date);

-- Create function to create cycle count plan
CREATE OR REPLACE FUNCTION create_cycle_count_plan(
    org_id uuid,
    name VARCHAR,
    description TEXT,
    frequency VARCHAR,
    abc_class VARCHAR,
    start_date DATE,
    end_date DATE,
    created_by uuid
)
RETURNS uuid AS $$
DECLARE
    plan_id uuid;
BEGIN
    plan_id := gen_random_uuid();

    INSERT INTO inventory_cycle_count_plans (
        id, organization_id, name, description, frequency, abc_class,
        start_date, end_date, status, created_by, created_at, updated_at
    ) VALUES (
        plan_id, org_id, name, description, frequency, abc_class,
        start_date, end_date, 'draft', created_by, NOW(), NOW()
    );

    RETURN plan_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to start cycle count session
CREATE OR REPLACE FUNCTION start_cycle_count_session(
    org_id uuid,
    plan_id uuid,
    name VARCHAR,
    location_id uuid,
    user_id uuid,
    count_method VARCHAR
)
RETURNS uuid AS $$
DECLARE
    session_id uuid;
BEGIN
    session_id := gen_random_uuid();

    INSERT INTO inventory_cycle_count_sessions (
        id, organization_id, plan_id, name, location_id, user_id,
        status, count_method, created_at, updated_at
    ) VALUES (
        session_id, org_id, plan_id, name, location_id, user_id,
        'in_progress', count_method, NOW(), NOW()
    );

    RETURN session_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to add count line to session
CREATE OR REPLACE FUNCTION add_cycle_count_line(
    session_id uuid,
    org_id uuid,
    product_id uuid,
    location_id uuid,
    counted_quantity NUMERIC,
    counted_by uuid,
    lot_id uuid DEFAULT NULL,
    package_id uuid DEFAULT NULL
)
RETURNS TABLE (
    line_id uuid,
    success boolean,
    message TEXT,
    system_quantity NUMERIC,
    variance NUMERIC,
    variance_percentage NUMERIC
) AS $$
DECLARE
    line_id uuid;
    system_qty NUMERIC;
    variance_qty NUMERIC;
    variance_pct NUMERIC;
BEGIN
    -- Get current system quantity
    SELECT COALESCE(SUM(quantity), 0) INTO system_qty
    FROM stock_quants
    WHERE organization_id = org_id
    AND product_id = product_id
    AND location_id = location_id
    AND (lot_id = lot_id OR (lot_id IS NULL AND COALESCE(lot_id, '00000000-0000-0000-0000-000000000000'::uuid) = '00000000-0000-0000-0000-000000000000'::uuid))
    AND (package_id = package_id OR (package_id IS NULL AND COALESCE(package_id, '00000000-0000-0000-0000-000000000000'::uuid) = '00000000-0000-0000-0000-000000000000'::uuid));

    -- Calculate variance
    variance_qty := counted_quantity - system_qty;

    IF system_qty = 0 THEN
        variance_pct := 0;
    ELSE
        variance_pct := (variance_qty / system_qty) * 100;
    END IF;

    -- Create count line
    line_id := gen_random_uuid();

    INSERT INTO inventory_cycle_count_lines (
        id, session_id, organization_id, product_id, location_id,
        counted_quantity, system_quantity, counted_by, lot_id, package_id,
        count_time, created_at, updated_at
    ) VALUES (
        line_id, session_id, org_id, product_id, location_id,
        counted_quantity, system_qty, counted_by, lot_id, package_id,
        NOW(), NOW(), NOW()
    );

    RETURN QUERY SELECT
        line_id,
        true,
        'Count line added successfully',
        system_qty,
        variance_qty,
        variance_pct;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to complete cycle count session
CREATE OR REPLACE FUNCTION complete_cycle_count_session(
    session_id uuid,
    org_id uuid,
    end_time TIMESTAMPTZ DEFAULT NULL
)
RETURNS boolean AS $$
DECLARE
    session_record RECORD;
    line_count INT;
BEGIN
    -- Get session
    SELECT * INTO session_record
    FROM inventory_cycle_count_sessions
    WHERE id = session_id AND organization_id = org_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Count lines in session
    SELECT COUNT(*) INTO line_count
    FROM inventory_cycle_count_lines
    WHERE session_id = session_id AND organization_id = org_id;

    -- Update session status
    UPDATE inventory_cycle_count_sessions
    SET
        status = 'completed',
        end_time = COALESCE(end_time, NOW()),
        updated_at = NOW()
    WHERE id = session_id AND organization_id = org_id;

    -- Record accuracy history for each line
    INSERT INTO inventory_cycle_count_accuracy (
        id, organization_id, product_id, location_id, count_date,
        system_quantity, counted_quantity, variance, variance_percentage,
        count_method, counted_by, created_at
    )
    SELECT
        gen_random_uuid(),
        org_id,
        product_id,
        location_id,
        DATE(NOW()),
        system_quantity,
        counted_quantity,
        variance,
        variance_percentage,
        'manual',
        counted_by,
        NOW()
    FROM inventory_cycle_count_lines
    WHERE session_id = session_id AND organization_id = org_id;

    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Create function to verify cycle count line
CREATE OR REPLACE FUNCTION verify_cycle_count_line(
    line_id uuid,
    org_id uuid,
    verified_by uuid,
    status VARCHAR
)
RETURNS boolean AS $$
DECLARE
    line_record RECORD;
BEGIN
    -- Get line
    SELECT * INTO line_record
    FROM inventory_cycle_count_lines
    WHERE id = line_id AND organization_id = org_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Update line status
    UPDATE inventory_cycle_count_lines
    SET
        status = status,
        verified_by = verified_by,
        verification_time = NOW(),
        updated_at = NOW()
    WHERE id = line_id AND organization_id = org_id;

    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Create function to create adjustment from count variance
CREATE OR REPLACE FUNCTION create_adjustment_from_variance(
    line_id uuid,
    org_id uuid,
    adjustment_type VARCHAR,
    reason VARCHAR,
    adjusted_by uuid
)
RETURNS uuid AS $$
DECLARE
    adjustment_id uuid;
    line_record RECORD;
BEGIN
    -- Get count line
    SELECT * INTO line_record
    FROM inventory_cycle_count_lines
    WHERE id = line_id AND organization_id = org_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Count line not found';
    END IF;

    -- Create adjustment
    adjustment_id := gen_random_uuid();

    INSERT INTO inventory_cycle_count_adjustments (
        id, organization_id, count_line_id, product_id, location_id,
        old_quantity, new_quantity, adjustment_type, reason,
        adjusted_by, status, created_at, updated_at
    ) VALUES (
        adjustment_id, org_id, line_id, line_record.product_id, line_record.location_id,
        line_record.system_quantity, line_record.counted_quantity, adjustment_type, reason,
        adjusted_by, 'pending', NOW(), NOW()
    );

    RETURN adjustment_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to approve cycle count adjustment
CREATE OR REPLACE FUNCTION approve_cycle_count_adjustment(
    adjustment_id uuid,
    org_id uuid,
    approved_by uuid
)
RETURNS boolean AS $$
DECLARE
    adjustment_record RECORD;
    product_id uuid;
    location_id uuid;
    lot_id uuid;
    package_id uuid;
    new_quantity NUMERIC;
BEGIN
    -- Get adjustment
    SELECT * INTO adjustment_record
    FROM inventory_cycle_count_adjustments
    WHERE id = adjustment_id AND organization_id = org_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Update adjustment status
    UPDATE inventory_cycle_count_adjustments
    SET
        status = 'approved',
        approved_by = approved_by,
        approval_time = NOW(),
        updated_at = NOW()
    WHERE id = adjustment_id AND organization_id = org_id;

    -- Apply adjustment to stock quantities
    product_id := adjustment_record.product_id;
    location_id := adjustment_record.location_id;
    new_quantity := adjustment_record.new_quantity;

    -- Update stock quantity
    INSERT INTO stock_quants (
        id, organization_id, product_id, location_id, quantity,
        reserved_quantity, created_at, updated_at
    ) VALUES (
        gen_random_uuid(), org_id, product_id, location_id, new_quantity, 0, NOW(), NOW()
    )
    ON CONFLICT (product_id, location_id, organization_id)
    DO UPDATE SET
        quantity = EXCLUDED.quantity,
        updated_at = NOW();

    -- Update count line status
    UPDATE inventory_cycle_count_lines
    SET
        status = 'adjusted',
        updated_at = NOW()
    WHERE id = adjustment_record.count_line_id AND organization_id = org_id;

    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Create function to get cycle count accuracy metrics
CREATE OR REPLACE FUNCTION get_cycle_count_accuracy_metrics(
    org_id uuid,
    date_from DATE DEFAULT NULL,
    date_to DATE DEFAULT NULL
)
RETURNS TABLE (
    total_counts BIGINT,
    accurate_counts BIGINT,
    variance_counts BIGINT,
    accuracy_percentage NUMERIC(5,2),
    average_variance_percentage NUMERIC(10,2),
    total_variance_quantity NUMERIC(15,4)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_counts,
        SUM(CASE WHEN variance_percentage = 0 THEN 1 ELSE 0 END) as accurate_counts,
        SUM(CASE WHEN variance_percentage != 0 THEN 1 ELSE 0 END) as variance_counts,
        COALESCE(SUM(CASE WHEN variance_percentage = 0 THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as accuracy_percentage,
        COALESCE(AVG(ABS(variance_percentage)), 0) as average_variance_percentage,
        COALESCE(SUM(variance), 0) as total_variance_quantity
    FROM inventory_cycle_count_accuracy
    WHERE organization_id = org_id
    AND (date_from IS NULL OR count_date >= date_from)
    AND (date_to IS NULL OR count_date <= date_to);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get products needing cycle count
CREATE OR REPLACE FUNCTION get_products_needing_cycle_count(
    org_id uuid,
    days_since_last_count INT DEFAULT 30,
    min_variance_percentage NUMERIC DEFAULT 5
)
RETURNS TABLE (
    product_id uuid,
    product_name VARCHAR,
    default_code VARCHAR,
    category_id uuid,
    last_count_date DATE,
    days_since_count INT,
    last_variance_percentage NUMERIC(10,2),
    average_variance_percentage NUMERIC(10,2),
    count_priority INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id as product_id,
        p.name as product_name,
        p.default_code,
        p.category_id,
        MAX(cca.count_date) as last_count_date,
        COALESCE(DATE_PART('day', NOW() - MAX(cca.count_date)), 999) as days_since_count,
        MAX(cca.variance_percentage) as last_variance_percentage,
        COALESCE(AVG(cca.variance_percentage), 0) as average_variance_percentage,
        CASE
            WHEN MAX(cca.count_date) IS NULL THEN 1
            WHEN DATE_PART('day', NOW() - MAX(cca.count_date)) > days_since_last_count THEN 2
            WHEN MAX(cca.variance_percentage) > min_variance_percentage THEN 3
            ELSE 4
        END as count_priority
    FROM products p
    LEFT JOIN inventory_cycle_count_accuracy cca ON
        p.id = cca.product_id AND
        p.organization_id = cca.organization_id
    WHERE p.organization_id = org_id
    GROUP BY p.id
    HAVING
        MAX(cca.count_date) IS NULL OR
        DATE_PART('day', NOW() - MAX(cca.count_date)) > days_since_last_count OR
        MAX(cca.variance_percentage) > min_variance_percentage
    ORDER BY count_priority, days_since_count DESC;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Grant permissions (handled by RLS in main migration)
COMMENT ON TABLE inventory_cycle_count_plans IS 'Cycle count plans - filtered by organization RLS';
COMMENT ON TABLE inventory_cycle_count_sessions IS 'Cycle count sessions - filtered by organization RLS';
COMMENT ON TABLE inventory_cycle_count_lines IS 'Cycle count lines - filtered by organization RLS';
COMMENT ON TABLE inventory_cycle_count_adjustments IS 'Cycle count adjustments - filtered by organization RLS';
COMMENT ON TABLE inventory_cycle_count_accuracy IS 'Cycle count accuracy history - filtered by organization RLS';

-- =====================================================
-- END OF MIGRATION
-- =====================================================
