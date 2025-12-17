-- Replenishment Rules and Orders
-- This migration adds automatic stock replenishment functionality

BEGIN;

-- Create replenishment_rules table
CREATE TABLE replenishment_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    description text,
    product_id uuid REFERENCES products(id),
    product_category_id uuid REFERENCES product_categories(id),
    warehouse_id uuid REFERENCES warehouses(id),
    location_id uuid REFERENCES stock_locations(id),

    -- Replenishment triggers
    trigger_type varchar(50) NOT NULL CHECK (trigger_type IN ('reorder_point', 'safety_stock', 'manual')),
    min_quantity numeric(15,4),
    max_quantity numeric(15,4),
    reorder_point numeric(15,4),
    safety_stock numeric(15,4),

    -- Procurement settings
    procure_method varchar(50) NOT NULL CHECK (procure_method IN ('make_to_stock', 'make_to_order')),
    order_quantity numeric(15,4),
    multiple_of numeric(15,4),
    lead_time_days integer,

    -- Scheduling
    check_frequency varchar(50) NOT NULL CHECK (check_frequency IN ('daily', 'weekly', 'monthly', 'real_time')),
    last_checked_at timestamptz,
    next_check_at timestamptz,

    -- Source and destination
    source_location_id uuid REFERENCES stock_locations(id),
    dest_location_id uuid REFERENCES stock_locations(id),

    -- Status
    active boolean DEFAULT true,
    priority integer DEFAULT 10,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    -- Indexes
    CONSTRAINT unique_replenishment_rule_name_per_org UNIQUE (organization_id, name)
);

-- Create replenishment_orders table
CREATE TABLE replenishment_orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    rule_id uuid NOT NULL REFERENCES replenishment_rules(id),
    product_id uuid NOT NULL REFERENCES products(id),
    product_name varchar(255) NOT NULL,
    quantity numeric(15,4) NOT NULL,
    uom_id uuid REFERENCES uom_units(id),

    -- Source and destination
    source_location_id uuid REFERENCES stock_locations(id),
    dest_location_id uuid REFERENCES stock_locations(id),

    -- Status
    status varchar(50) NOT NULL CHECK (status IN ('draft', 'confirmed', 'processed', 'cancelled')),
    priority integer DEFAULT 10,
    scheduled_date timestamptz,

    -- Procurement details
    procure_method varchar(50) NOT NULL CHECK (procure_method IN ('make_to_stock', 'make_to_order')),
    reference varchar(255),
    notes text,

    -- Resulting documents
    purchase_order_id uuid REFERENCES purchase_orders(id),
    manufacturing_order_id uuid REFERENCES manufacturing_orders(id),
    transfer_id uuid REFERENCES stock_pickings(id),

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Create indexes for performance
CREATE INDEX idx_replenishment_rules_org ON replenishment_rules(organization_id);
CREATE INDEX idx_replenishment_rules_product ON replenishment_rules(product_id);
CREATE INDEX idx_replenishment_rules_category ON replenishment_rules(product_category_id);
CREATE INDEX idx_replenishment_rules_warehouse ON replenishment_rules(warehouse_id);
CREATE INDEX idx_replenishment_rules_location ON replenishment_rules(location_id);
CREATE INDEX idx_replenishment_rules_active ON replenishment_rules(organization_id, active) WHERE active = true;

CREATE INDEX idx_replenishment_orders_org ON replenishment_orders(organization_id);
CREATE INDEX idx_replenishment_orders_product ON replenishment_orders(product_id);
CREATE INDEX idx_replenishment_orders_rule ON replenishment_orders(rule_id);
CREATE INDEX idx_replenishment_orders_status ON replenishment_orders(organization_id, status);
CREATE INDEX idx_replenishment_orders_priority ON replenishment_orders(organization_id, priority);

-- Create function to check and create replenishment orders
CREATE OR REPLACE FUNCTION check_and_create_replenishment_orders(
    p_organization_id uuid,
    p_limit integer DEFAULT 100
) RETURNS TABLE (
    order_id uuid,
    product_id uuid,
    product_name varchar,
    quantity numeric,
    status varchar,
    rule_name varchar
) AS $$
BEGIN
    RETURN QUERY
    WITH eligible_rules AS (
        SELECT r.id, r.name, r.product_id, r.product_category_id, r.warehouse_id, r.location_id,
               r.trigger_type, r.reorder_point, r.safety_stock, r.procure_method,
               r.order_quantity, r.multiple_of, r.source_location_id, r.dest_location_id,
               r.active, r.priority
        FROM replenishment_rules r
        WHERE r.organization_id = p_organization_id
        AND r.active = true
        AND (r.check_frequency = 'real_time' OR
             (r.check_frequency = 'daily' AND (r.next_check_at IS NULL OR r.next_check_at <= now())) OR
             (r.check_frequency = 'weekly' AND (r.next_check_at IS NULL OR r.next_check_at <= now())) OR
             (r.check_frequency = 'monthly' AND (r.next_check_at IS NULL OR r.next_check_at <= now())))
        ORDER BY r.priority ASC
        LIMIT p_limit
    ),
    products_needing_replenishment AS (
        SELECT
            er.id as rule_id,
            er.name as rule_name,
            COALESCE(er.product_id, p.id) as product_id,
            p.name as product_name,
            SUM(sq.quantity) as current_quantity,
            COALESCE(p.reorder_point, er.reorder_point, 0) as reorder_point,
            COALESCE(p.safety_stock, er.safety_stock, 0) as safety_stock,
            COALESCE(er.order_quantity,
                     GREATEST(COALESCE(p.reorder_point, er.reorder_point, 0) - SUM(sq.quantity) + COALESCE(p.safety_stock, er.safety_stock, 0), 0),
                     0) as recommended_quantity,
            er.procure_method,
            er.source_location_id,
            er.dest_location_id,
            er.priority
        FROM eligible_rules er
        LEFT JOIN products p ON
            (er.product_id = p.id OR
             (er.product_id IS NULL AND er.product_category_id = p.category_id AND p.organization_id = p_organization_id))
        LEFT JOIN stock_quants sq ON p.id = sq.product_id AND sq.organization_id = p_organization_id
            AND (er.location_id IS NULL OR sq.location_id = er.location_id)
        WHERE p.id IS NOT NULL
        AND p.organization_id = p_organization_id
        AND p.active = true
        AND sq.deleted_at IS NULL
        GROUP BY er.id, er.name, COALESCE(er.product_id, p.id), p.name, er.procure_method,
                 er.source_location_id, er.dest_location_id, er.priority,
                 COALESCE(p.reorder_point, er.reorder_point),
                 COALESCE(p.safety_stock, er.safety_stock),
                 er.order_quantity
        HAVING
            (er.trigger_type = 'reorder_point' AND SUM(sq.quantity) <= COALESCE(p.reorder_point, er.reorder_point, 0)) OR
            (er.trigger_type = 'safety_stock' AND SUM(sq.quantity) <= COALESCE(p.safety_stock, er.safety_stock, 0)) OR
            (er.trigger_type = 'manual' AND er.order_quantity IS NOT NULL)
    )
    INSERT INTO replenishment_orders (
        organization_id, rule_id, product_id, product_name, quantity, uom_id,
        source_location_id, dest_location_id, status, priority, procure_method,
        scheduled_date, created_at, updated_at
    )
    SELECT
        p_organization_id,
        pnr.rule_id,
        pnr.product_id,
        pnr.product_name,
        pnr.recommended_quantity,
        p.uom_id,
        pnr.source_location_id,
        pnr.dest_location_id,
        'draft',
        pnr.priority,
        pnr.procure_method,
        CASE
            WHEN pnr.procure_method = 'make_to_order' THEN now() + (COALESCE(p.lead_time_days, 7) || ' days')::interval
            ELSE now()
        END,
        now(),
        now()
    FROM products_needing_replenishment pnr
    JOIN products p ON pnr.product_id = p.id
    ON CONFLICT (organization_id, product_id, status)
    DO UPDATE SET
        quantity = EXCLUDED.quantity,
        updated_at = now()
    RETURNING
        id as order_id,
        product_id,
        product_name,
        quantity,
        status,
        (SELECT name FROM replenishment_rules WHERE id = rule_id) as rule_name;
END;
$$ LANGUAGE plpgsql;

-- Create function to update next check times for replenishment rules
CREATE OR REPLACE FUNCTION update_replenishment_rule_check_times(
    p_organization_id uuid
) RETURNS void AS $$
BEGIN
    -- Update next check times based on frequency
    UPDATE replenishment_rules
    SET
        last_checked_at = now(),
        next_check_at =
            CASE check_frequency
                WHEN 'daily' THEN now() + interval '1 day'
                WHEN 'weekly' THEN now() + interval '7 days'
                WHEN 'monthly' THEN now() + interval '30 days'
                ELSE NULL -- real_time rules don't need next_check_at
            END,
        updated_at = now()
    WHERE organization_id = p_organization_id
    AND active = true
    AND check_frequency != 'real_time';
END;
$$ LANGUAGE plpgsql;

-- Create function to process replenishment orders
CREATE OR REPLACE FUNCTION process_replenishment_orders(
    p_organization_id uuid,
    p_limit integer DEFAULT 10
) RETURNS TABLE (
    order_id uuid,
    product_id uuid,
    product_name varchar,
    quantity numeric,
    status varchar,
    procure_method varchar
) AS $$
DECLARE
    v_order_id uuid;
    v_product_id uuid;
    v_quantity numeric;
    v_procure_method varchar;
    v_source_location_id uuid;
    v_dest_location_id uuid;
    v_reference text;
BEGIN
    FOR v_order_id, v_product_id, v_quantity, v_procure_method, v_source_location_id, v_dest_location_id IN
        SELECT id, product_id, quantity, procure_method, source_location_id, dest_location_id
        FROM replenishment_orders
        WHERE organization_id = p_organization_id
        AND status = 'draft'
        ORDER BY priority ASC, created_at ASC
        LIMIT p_limit
    LOOP
        -- Generate reference
        v_reference := 'AUTO-REPLENISH-' || to_char(now(), 'YYYYMMDD-HH24MISS') || '-' || v_product_id::text;

        -- Update order status to confirmed
        UPDATE replenishment_orders
        SET
            status = 'confirmed',
            reference = v_reference,
            updated_at = now()
        WHERE id = v_order_id;

        -- For make_to_stock, we would typically create a purchase order or manufacturing order
        -- For now, we'll just create a stock move for internal transfers
        IF v_procure_method = 'make_to_stock' AND v_source_location_id IS NOT NULL AND v_dest_location_id IS NOT NULL THEN
            -- Create a stock move for internal transfer
            INSERT INTO stock_moves (
                organization_id, name, sequence, priority, date, product_id,
                product_uom_qty, location_id, location_dest_id, state, procure_method,
                origin, created_at, updated_at
            )
            SELECT
                p_organization_id,
                'Replenishment: ' || p.name,
                10,
                '1',
                now(),
                v_product_id,
                v_quantity,
                v_source_location_id,
                v_dest_location_id,
                'draft',
                v_procure_method,
                'Auto-replenishment order: ' || v_reference,
                now(),
                now()
            FROM products p
            WHERE p.id = v_product_id;

            -- Update the replenishment order with the transfer reference
            UPDATE replenishment_orders
            SET
                transfer_id = currval('stock_moves_id_seq'),
                status = 'processed',
                updated_at = now()
            WHERE id = v_order_id;
        END IF;

        RETURN NEXT;
        RETURN QUERY SELECT v_order_id, v_product_id, p.name, v_quantity, 'processed', v_procure_method
        FROM products p WHERE p.id = v_product_id;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Create function to run complete replenishment cycle
CREATE OR REPLACE FUNCTION run_replenishment_cycle(
    p_organization_id uuid
) RETURNS TABLE (
    action varchar,
    count integer,
    message text
) AS $$
DECLARE
    v_check_count integer;
    v_process_count integer;
BEGIN
    -- Check for products needing replenishment and create orders
    INSERT INTO replenishment_orders (
        organization_id, rule_id, product_id, product_name, quantity, uom_id,
        source_location_id, dest_location_id, status, priority, procure_method,
        scheduled_date, created_at, updated_at
    )
    SELECT * FROM check_and_create_replenishment_orders(p_organization_id, 100);

    GET DIAGNOSTICS v_check_count = ROW_COUNT;

    -- Process the created orders
    PERFORM * FROM process_replenishment_orders(p_organization_id, 20);
    GET DIAGNOSTICS v_process_count = ROW_COUNT;

    -- Update check times for rules
    PERFORM update_replenishment_rule_check_times(p_organization_id);

    -- Return results
    RETURN QUERY VALUES
        ('replenishment_check', v_check_count, 'Products checked for replenishment'),
        ('orders_created', v_check_count, 'Replenishment orders created'),
        ('orders_processed', v_process_count, 'Replenishment orders processed'),
        ('rules_updated', (SELECT count(*) FROM replenishment_rules WHERE organization_id = p_organization_id AND active = true), 'Replenishment rules updated');
END;
$$ LANGUAGE plpgsql;

-- Create RLS policies for replenishment tables
ALTER TABLE replenishment_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE replenishment_orders ENABLE ROW LEVEL SECURITY;

-- Apply standard RLS policies
CREATE POLICY replenishment_rules_org_policy ON replenishment_rules
    USING (organization_id = current_setting('app.current_organization_id')::uuid);

CREATE POLICY replenishment_orders_org_policy ON replenishment_orders
    USING (organization_id = current_setting('app.current_organization_id')::uuid);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON replenishment_rules TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON replenishment_orders TO authenticated;
GRANT EXECUTE ON FUNCTION check_and_create_replenishment_orders(uuid, integer) TO authenticated;
GRANT EXECUTE ON FUNCTION update_replenishment_rule_check_times(uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION process_replenishment_orders(uuid, integer) TO authenticated;
GRANT EXECUTE ON FUNCTION run_replenishment_cycle(uuid) TO authenticated;

COMMIT;

-- Comments
COMMENT ON TABLE replenishment_rules IS 'Automatic stock replenishment rules - filtered by organization RLS';
COMMENT ON TABLE replenishment_orders IS 'Automatically generated replenishment orders - filtered by organization RLS';
COMMENT ON FUNCTION check_and_create_replenishment_orders IS 'Check replenishment rules and create orders for products needing stock';
COMMENT ON FUNCTION process_replenishment_orders IS 'Process draft replenishment orders and create stock moves';
COMMENT ON FUNCTION run_replenishment_cycle IS 'Run complete replenishment cycle: check, create, and process orders';
COMMENT ON FUNCTION update_replenishment_rule_check_times IS 'Update next check times for replenishment rules based on frequency';

-- Indexes for RLS
CREATE INDEX idx_replenishment_rules_org_rls ON replenishment_rules(organization_id) WHERE organization_id = current_setting('app.current_organization_id')::uuid;
CREATE INDEX idx_replenishment_orders_org_rls ON replenishment_orders(organization_id) WHERE organization_id = current_setting('app.current_organization_id')::uuid;
