-- Migration: Delivery Tracking Module
-- Description: Delivery routes, shipments, tracking events, GPS ingestion, and queue integration
-- Created: 2025-01-01

-- =====================================================
-- EXTEND STOCK PICKINGS WITH DELIVERY METADATA
-- =====================================================

-- Add columns without DEFAULT to avoid table rewrite
ALTER TABLE stock_pickings
    ADD COLUMN delivery_status varchar(20),
    ADD COLUMN shipment_priority varchar(20),
    ADD COLUMN planned_departure_at timestamptz,
    ADD COLUMN planned_arrival_at timestamptz,
    ADD COLUMN delivery_metadata jsonb;

-- Set defaults separately to avoid table rewrite
ALTER TABLE stock_pickings
    ALTER COLUMN delivery_status SET DEFAULT 'draft',
    ALTER COLUMN shipment_priority SET DEFAULT 'normal',
    ALTER COLUMN delivery_metadata SET DEFAULT '{}'::jsonb;

-- Update existing rows with default values
UPDATE stock_pickings
SET delivery_status = 'draft'
WHERE delivery_status IS NULL;

UPDATE stock_pickings
SET shipment_priority = 'normal'
WHERE shipment_priority IS NULL;

UPDATE stock_pickings
SET delivery_metadata = '{}'::jsonb
WHERE delivery_metadata IS NULL;

-- Add NOT NULL constraints after setting values
-- Add NOT NULL constraints using CHECK constraints first (non-blocking)
ALTER TABLE stock_pickings
    ADD CONSTRAINT stock_pickings_delivery_status_not_null CHECK (delivery_status IS NOT NULL) NOT VALID,
    ADD CONSTRAINT stock_pickings_shipment_priority_not_null CHECK (shipment_priority IS NOT NULL) NOT VALID,
    ADD CONSTRAINT stock_pickings_delivery_metadata_not_null CHECK (delivery_metadata IS NOT NULL) NOT VALID;

-- Validate the constraints separately (can be done concurrently)
ALTER TABLE stock_pickings VALIDATE CONSTRAINT stock_pickings_delivery_status_not_null;
ALTER TABLE stock_pickings VALIDATE CONSTRAINT stock_pickings_shipment_priority_not_null;
ALTER TABLE stock_pickings VALIDATE CONSTRAINT stock_pickings_delivery_metadata_not_null;

ALTER TABLE stock_pickings
    ADD CONSTRAINT stock_pickings_delivery_status_check
        CHECK (delivery_status IN ('draft', 'scheduled', 'in_transit', 'delivered', 'failed', 'cancelled')),
    ADD CONSTRAINT stock_pickings_shipment_priority_check
        CHECK (shipment_priority IN ('low', 'normal', 'high', 'critical'));

-- =====================================================
-- CORE DELIVERY TABLES
-- =====================================================

CREATE TABLE delivery_vehicles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    registration_number varchar(100),
    vehicle_identifier varchar(100),
    vehicle_type varchar(50) DEFAULT 'truck',
    capacity numeric(12,2),
    capacity_uom_id uuid REFERENCES uom_units(id),
    active boolean DEFAULT true,
    last_service_at date,
    service_interval_days integer,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    CONSTRAINT delivery_vehicles_type_check CHECK (vehicle_type IN ('truck', 'van', 'bike', 'car', 'drone', 'other')),
    CONSTRAINT delivery_vehicles_registration_unique UNIQUE (organization_id, registration_number)
);

CREATE TABLE delivery_routes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    warehouse_id uuid REFERENCES warehouses(id),
    name varchar(255) NOT NULL,
    route_code varchar(50),
    transport_mode varchar(20) DEFAULT 'road',
    status varchar(20) NOT NULL DEFAULT 'draft',
    scheduled_start_at timestamptz,
    scheduled_end_at timestamptz,
    actual_start_at timestamptz,
    actual_end_at timestamptz,
    origin_location_id uuid REFERENCES stock_locations(id),
    destination_location_id uuid REFERENCES stock_locations(id),
    notes text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    CONSTRAINT delivery_routes_status_check CHECK (status IN ('draft', 'scheduled', 'in_progress', 'completed', 'cancelled')),
    CONSTRAINT delivery_routes_mode_check CHECK (transport_mode IN ('road', 'air', 'sea', 'rail', 'other')),
    CONSTRAINT delivery_routes_unique_code UNIQUE (organization_id, route_code)
);

CREATE TABLE delivery_route_assignments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    route_id uuid NOT NULL REFERENCES delivery_routes(id) ON DELETE CASCADE,
    vehicle_id uuid REFERENCES delivery_vehicles(id),
    driver_employee_id uuid REFERENCES employees(id),
    driver_contact_id uuid REFERENCES contacts(id),
    assignment_status varchar(20) NOT NULL DEFAULT 'assigned',
    assigned_at timestamptz NOT NULL DEFAULT now(),
    acknowledged_at timestamptz,
    released_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    CONSTRAINT delivery_route_assignments_status_check CHECK (assignment_status IN ('assigned', 'accepted', 'declined', 'released', 'completed'))
);

CREATE TABLE delivery_shipments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    picking_id uuid NOT NULL UNIQUE REFERENCES stock_pickings(id) ON DELETE CASCADE,
    route_id uuid REFERENCES delivery_routes(id) ON DELETE SET NULL,
    assignment_id uuid REFERENCES delivery_route_assignments(id) ON DELETE SET NULL,
    tracking_number varchar(100),
    carrier_name varchar(120),
    carrier_code varchar(60),
    carrier_service_level varchar(60),
    shipment_type varchar(20) NOT NULL DEFAULT 'outbound',
    status varchar(20) NOT NULL DEFAULT 'draft',
    requires_signature boolean DEFAULT false,
    estimated_departure_at timestamptz,
    estimated_arrival_at timestamptz,
    departed_at timestamptz,
    arrived_at timestamptz,
    last_event_at timestamptz,
    last_latitude numeric(10,6),
    last_longitude numeric(10,6),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    CONSTRAINT delivery_shipments_status_check CHECK (status IN ('draft', 'scheduled', 'in_transit', 'delivered', 'failed', 'cancelled')),
    CONSTRAINT delivery_shipments_type_check CHECK (shipment_type IN ('outbound', 'inbound', 'internal'))
);

CREATE UNIQUE INDEX delivery_shipments_tracking_uidx
    ON delivery_shipments (organization_id, tracking_number)
    WHERE tracking_number IS NOT NULL;

CREATE TABLE delivery_route_stops (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    route_id uuid NOT NULL REFERENCES delivery_routes(id) ON DELETE CASCADE,
    assignment_id uuid REFERENCES delivery_route_assignments(id) ON DELETE SET NULL,
    shipment_id uuid REFERENCES delivery_shipments(id) ON DELETE SET NULL,
    stop_sequence integer NOT NULL,
    contact_id uuid REFERENCES contacts(id),
    location_id uuid REFERENCES stock_locations(id),
    address jsonb,
    planned_arrival_at timestamptz,
    planned_departure_at timestamptz,
    actual_arrival_at timestamptz,
    actual_departure_at timestamptz,
    status varchar(20) NOT NULL DEFAULT 'planned',
    notes text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    CONSTRAINT delivery_route_stops_status_check CHECK (status IN ('planned', 'en_route', 'arrived', 'completed', 'skipped', 'failed')),
    CONSTRAINT delivery_route_stops_sequence_unique UNIQUE (route_id, stop_sequence)
);

CREATE TABLE delivery_tracking_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shipment_id uuid NOT NULL REFERENCES delivery_shipments(id) ON DELETE CASCADE,
    stop_id uuid REFERENCES delivery_route_stops(id) ON DELETE SET NULL,
    event_type varchar(50) NOT NULL,
    status varchar(20),
    event_time timestamptz NOT NULL DEFAULT now(),
    source varchar(50),
    message text,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    latitude numeric(10,6),
    longitude numeric(10,6),
    altitude numeric(10,2),
    speed_kph numeric(10,2),
    heading numeric(10,2),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    CONSTRAINT delivery_tracking_events_status_check CHECK (
        status IS NULL OR status IN ('draft', 'scheduled', 'in_transit', 'delivered', 'failed', 'cancelled')
    ),
    CONSTRAINT delivery_tracking_events_latitude CHECK (latitude IS NULL OR (latitude BETWEEN -90 AND 90)),
    CONSTRAINT delivery_tracking_events_longitude CHECK (longitude IS NULL OR (longitude BETWEEN -180 AND 180))
);

CREATE TABLE delivery_route_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    route_id uuid NOT NULL REFERENCES delivery_routes(id) ON DELETE CASCADE,
    assignment_id uuid REFERENCES delivery_route_assignments(id) ON DELETE SET NULL,
    vehicle_id uuid REFERENCES delivery_vehicles(id),
    recorded_at timestamptz NOT NULL DEFAULT now(),
    latitude numeric(10,6) NOT NULL,
    longitude numeric(10,6) NOT NULL,
    altitude numeric(10,2),
    speed_kph numeric(10,2),
    heading numeric(10,2),
    source varchar(50),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT delivery_route_positions_latitude CHECK (latitude BETWEEN -90 AND 90),
    CONSTRAINT delivery_route_positions_longitude CHECK (longitude BETWEEN -180 AND 180)
);

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_delivery_vehicles_org
    ON delivery_vehicles (organization_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_delivery_routes_org_status
    ON delivery_routes (organization_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_delivery_route_assignments_route
    ON delivery_route_assignments (route_id, assignment_status);

CREATE INDEX idx_delivery_route_assignments_employee
    ON delivery_route_assignments (driver_employee_id)
    WHERE driver_employee_id IS NOT NULL;

CREATE INDEX idx_delivery_shipments_org_status
    ON delivery_shipments (organization_id, status);

CREATE INDEX idx_delivery_shipments_route
    ON delivery_shipments (route_id);

CREATE INDEX idx_delivery_route_stops_route
    ON delivery_route_stops (route_id, stop_sequence);

CREATE INDEX idx_delivery_route_stops_shipment
    ON delivery_route_stops (shipment_id)
    WHERE shipment_id IS NOT NULL;

CREATE INDEX idx_delivery_tracking_events_shipment_time
    ON delivery_tracking_events (shipment_id, event_time DESC);

CREATE INDEX idx_delivery_tracking_events_status
    ON delivery_tracking_events (status)
    WHERE status IS NOT NULL;

CREATE INDEX idx_delivery_route_positions_route_time
    ON delivery_route_positions (route_id, recorded_at DESC);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_delivery_vehicles_updated_at
    BEFORE UPDATE ON delivery_vehicles
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_delivery_routes_updated_at
    BEFORE UPDATE ON delivery_routes
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_delivery_route_assignments_updated_at
    BEFORE UPDATE ON delivery_route_assignments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_delivery_shipments_updated_at
    BEFORE UPDATE ON delivery_shipments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_delivery_route_stops_updated_at
    BEFORE UPDATE ON delivery_route_stops
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_delivery_tracking_events_updated_at
    BEFORE UPDATE ON delivery_tracking_events
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_delivery_route_positions_updated_at
    BEFORE UPDATE ON delivery_route_positions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- CORE DELIVERY FUNCTIONS
-- =====================================================

CREATE OR REPLACE FUNCTION create_delivery_route(
    p_organization_id uuid,
    p_name text,
    p_route_code text DEFAULT NULL,
    p_transport_mode text DEFAULT 'road',
    p_company_id uuid DEFAULT NULL,
    p_warehouse_id uuid DEFAULT NULL,
    p_scheduled_start_at timestamptz DEFAULT NULL,
    p_scheduled_end_at timestamptz DEFAULT NULL,
    p_metadata jsonb DEFAULT '{}'::jsonb
)
RETURNS delivery_routes
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_route delivery_routes;
BEGIN
    INSERT INTO delivery_routes (
        organization_id,
        company_id,
        warehouse_id,
        name,
        route_code,
        transport_mode,
        status,
        scheduled_start_at,
        scheduled_end_at,
        metadata
    )
    VALUES (
        p_organization_id,
        p_company_id,
        p_warehouse_id,
        p_name,
        p_route_code,
        COALESCE(p_transport_mode, 'road'),
        'draft',
        p_scheduled_start_at,
        p_scheduled_end_at,
        COALESCE(p_metadata, '{}'::jsonb)
    )
    RETURNING * INTO v_route;

    RETURN v_route;
END;
$$;

CREATE OR REPLACE FUNCTION set_shipment_status_from_event(
    p_shipment_id uuid,
    p_status text,
    p_stop_id uuid DEFAULT NULL,
    p_event_time timestamptz DEFAULT now()
)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    v_status text;
    v_picking_id uuid;
BEGIN
    IF p_status IS NULL THEN
        RETURN;
    END IF;

    v_status := lower(p_status);

    UPDATE delivery_shipments
    SET status = v_status,
        last_event_at = GREATEST(COALESCE(last_event_at, p_event_time), p_event_time),
        updated_at = now()
    WHERE id = p_shipment_id;

    SELECT picking_id INTO v_picking_id
    FROM delivery_shipments
    WHERE id = p_shipment_id;

    IF v_picking_id IS NOT NULL THEN
        UPDATE stock_pickings
        SET delivery_status = v_status,
            date_done = CASE
                WHEN v_status = 'delivered' AND date_done IS NULL THEN p_event_time
                ELSE date_done
            END,
            updated_at = now()
        WHERE id = v_picking_id;
    END IF;

    IF p_stop_id IS NOT NULL THEN
        UPDATE delivery_route_stops
        SET status = CASE
            WHEN v_status = 'delivered' THEN 'completed'
            WHEN v_status = 'in_transit' THEN 'en_route'
            WHEN v_status = 'failed' THEN 'failed'
            ELSE status
        END,
        actual_arrival_at = CASE
            WHEN v_status IN ('delivered', 'in_transit') AND actual_arrival_at IS NULL THEN p_event_time
            ELSE actual_arrival_at
        END,
        actual_departure_at = CASE
            WHEN v_status = 'delivered' THEN p_event_time
            ELSE actual_departure_at
        END,
        updated_at = now()
        WHERE id = p_stop_id;
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION assign_picking_to_route(
    p_picking_id uuid,
    p_route_id uuid,
    p_assignment_id uuid DEFAULT NULL,
    p_stop_sequence integer DEFAULT NULL,
    p_contact_id uuid DEFAULT NULL,
    p_location_id uuid DEFAULT NULL,
    p_metadata jsonb DEFAULT '{}'::jsonb
)
RETURNS delivery_shipments
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_picking RECORD;
    v_shipment delivery_shipments;
BEGIN
    SELECT
        sp.id,
        sp.organization_id,
        sp.company_id,
        sp.scheduled_date,
        sp.date_deadline,
        sp.delivery_status
    INTO v_picking
    FROM stock_pickings sp
    WHERE sp.id = p_picking_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Picking % not found', p_picking_id;
    END IF;

    INSERT INTO delivery_shipments (
        organization_id,
        company_id,
        picking_id,
        route_id,
        assignment_id,
        status,
        estimated_departure_at,
        estimated_arrival_at,
        metadata
    )
    VALUES (
        v_picking.organization_id,
        v_picking.company_id,
        p_picking_id,
        p_route_id,
        p_assignment_id,
        'scheduled',
        v_picking.scheduled_date,
        COALESCE(v_picking.date_deadline, v_picking.scheduled_date),
        COALESCE(p_metadata, '{}'::jsonb)
    )
    ON CONFLICT (picking_id) DO UPDATE
    SET
        route_id = COALESCE(EXCLUDED.route_id, delivery_shipments.route_id),
        assignment_id = COALESCE(EXCLUDED.assignment_id, delivery_shipments.assignment_id),
        estimated_departure_at = COALESCE(EXCLUDED.estimated_departure_at, delivery_shipments.estimated_departure_at),
        estimated_arrival_at = COALESCE(EXCLUDED.estimated_arrival_at, delivery_shipments.estimated_arrival_at),
        metadata = delivery_shipments.metadata || EXCLUDED.metadata,
        status = CASE
            WHEN delivery_shipments.status = 'draft' THEN 'scheduled'
            ELSE delivery_shipments.status
        END,
        updated_at = now()
    RETURNING * INTO v_shipment;

    UPDATE stock_pickings
    SET delivery_status = 'scheduled',
        planned_departure_at = COALESCE(planned_departure_at, v_picking.scheduled_date),
        planned_arrival_at = COALESCE(planned_arrival_at, v_picking.date_deadline),
        updated_at = now()
    WHERE id = p_picking_id;

    IF p_stop_sequence IS NOT NULL THEN
        INSERT INTO delivery_route_stops (
            organization_id,
            route_id,
            assignment_id,
            shipment_id,
            stop_sequence,
            contact_id,
            location_id,
            planned_arrival_at,
            planned_departure_at,
            status,
            metadata
        )
        VALUES (
            v_picking.organization_id,
            p_route_id,
            p_assignment_id,
            v_shipment.id,
            p_stop_sequence,
            p_contact_id,
            p_location_id,
            v_picking.date_deadline,
            v_picking.date_deadline,
            'planned',
            COALESCE(p_metadata, '{}'::jsonb)
        )
        ON CONFLICT (route_id, stop_sequence) DO UPDATE
        SET
            shipment_id = EXCLUDED.shipment_id,
            assignment_id = EXCLUDED.assignment_id,
            contact_id = COALESCE(EXCLUDED.contact_id, delivery_route_stops.contact_id),
            location_id = COALESCE(EXCLUDED.location_id, delivery_route_stops.location_id),
            planned_arrival_at = COALESCE(EXCLUDED.planned_arrival_at, delivery_route_stops.planned_arrival_at),
            planned_departure_at = COALESCE(EXCLUDED.planned_departure_at, delivery_route_stops.planned_departure_at),
            metadata = delivery_route_stops.metadata || EXCLUDED.metadata,
            updated_at = now();
    END IF;

    RETURN v_shipment;
END;
$$;

CREATE OR REPLACE FUNCTION log_delivery_tracking_event(
    p_organization_id uuid,
    p_shipment_id uuid,
    p_event_type text,
    p_status text DEFAULT NULL,
    p_event_time timestamptz DEFAULT now(),
    p_message text DEFAULT NULL,
    p_stop_id uuid DEFAULT NULL,
    p_source text DEFAULT NULL,
    p_raw_payload jsonb DEFAULT '{}'::jsonb,
    p_latitude numeric DEFAULT NULL,
    p_longitude numeric DEFAULT NULL,
    p_speed_kph numeric DEFAULT NULL,
    p_heading numeric DEFAULT NULL
)
RETURNS delivery_tracking_events
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_event delivery_tracking_events;
BEGIN
    INSERT INTO delivery_tracking_events (
        organization_id,
        shipment_id,
        stop_id,
        event_type,
        status,
        event_time,
        source,
        message,
        raw_payload,
        latitude,
        longitude,
        speed_kph,
        heading
    )
    VALUES (
        p_organization_id,
        p_shipment_id,
        p_stop_id,
        p_event_type,
        p_status,
        COALESCE(p_event_time, now()),
        p_source,
        p_message,
        COALESCE(p_raw_payload, '{}'::jsonb),
        p_latitude,
        p_longitude,
        p_speed_kph,
        p_heading
    )
    RETURNING * INTO v_event;

    PERFORM set_shipment_status_from_event(
        p_shipment_id := p_shipment_id,
        p_status := COALESCE(p_status, v_event.status),
        p_stop_id := p_stop_id,
        p_event_time := v_event.event_time
    );

    IF p_latitude IS NOT NULL AND p_longitude IS NOT NULL THEN
        UPDATE delivery_shipments
        SET last_latitude = p_latitude,
            last_longitude = p_longitude,
            updated_at = now()
        WHERE id = p_shipment_id;
    END IF;

    RETURN v_event;
END;
$$;

CREATE OR REPLACE FUNCTION get_delivery_movement_timeline(
    p_shipment_id uuid
)
RETURNS TABLE (
    entry_type text,
    entry_id uuid,
    occurred_at timestamptz,
    status text,
    description text,
    payload jsonb
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        'shipment'::text AS entry_type,
        ds.id AS entry_id,
        ds.created_at AS occurred_at,
        ds.status,
        'Shipment created' AS description,
        to_jsonb(ds) - 'organization_id'
    FROM delivery_shipments ds
    WHERE ds.id = p_shipment_id

    UNION ALL

    SELECT
        'event'::text,
        dte.id,
        dte.event_time,
        dte.status,
        dte.event_type,
        (to_jsonb(dte) - 'organization_id') || jsonb_build_object('stop_id', dte.stop_id)
    FROM delivery_tracking_events dte
    WHERE dte.shipment_id = p_shipment_id

    UNION ALL

    SELECT
        'stop'::text,
        drs.id,
        COALESCE(drs.actual_arrival_at, drs.planned_arrival_at, drs.created_at),
        drs.status,
        'Stop #' || drs.stop_sequence,
        to_jsonb(drs) - 'organization_id'
    FROM delivery_route_stops drs
    WHERE drs.shipment_id = p_shipment_id

    UNION ALL

    SELECT
        'stock_move'::text,
        sm.id,
        sm.date,
        sm.state,
        sm.name,
        jsonb_build_object(
            'product_id', sm.product_id,
            'product_uom_qty', sm.product_uom_qty,
            'location_id', sm.location_id,
            'location_dest_id', sm.location_dest_id
        )
    FROM stock_moves sm
    JOIN delivery_shipments ds ON ds.picking_id = sm.picking_id
    WHERE ds.id = p_shipment_id
    ORDER BY occurred_at;
END;
$$;

CREATE OR REPLACE FUNCTION get_live_route_snapshot(
    p_route_id uuid
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_route delivery_routes;
    v_snapshot jsonb;
BEGIN
    SELECT * INTO v_route
    FROM delivery_routes
    WHERE id = p_route_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Route % not found', p_route_id;
    END IF;

    SELECT jsonb_build_object(
        'route', to_jsonb(v_route) - 'organization_id',
        'latest_position',
            (
                SELECT to_jsonb(rp) - 'organization_id'
                FROM delivery_route_positions rp
                WHERE rp.route_id = p_route_id
                ORDER BY rp.recorded_at DESC
                LIMIT 1
            ),
        'assignments',
            (
                SELECT COALESCE(
                    jsonb_agg(to_jsonb(ra) - 'organization_id'),
                    '[]'::jsonb
                )
                FROM delivery_route_assignments ra
                WHERE ra.route_id = p_route_id
            ),
        'shipments',
            (
                SELECT COALESCE(
                    jsonb_agg(
                        jsonb_build_object(
                            'id', ds.id,
                            'picking_id', ds.picking_id,
                            'status', ds.status,
                            'tracking_number', ds.tracking_number,
                            'estimated_arrival_at', ds.estimated_arrival_at,
                            'last_event_at', ds.last_event_at
                        )
                    ),
                    '[]'::jsonb
                )
                FROM delivery_shipments ds
                WHERE ds.route_id = p_route_id
            ),
        'stops',
            (
                SELECT COALESCE(
                    jsonb_agg(
                        jsonb_build_object(
                            'id', drs.id,
                            'sequence', drs.stop_sequence,
                            'status', drs.status,
                            'planned_arrival_at', drs.planned_arrival_at,
                            'actual_arrival_at', drs.actual_arrival_at,
                            'shipment_id', drs.shipment_id
                        )
                        ORDER BY drs.stop_sequence
                    ),
                    '[]'::jsonb
                )
                FROM delivery_route_stops drs
                WHERE drs.route_id = p_route_id
            )
    ) INTO v_snapshot;

    RETURN v_snapshot;
END;
$$;

-- =====================================================
-- MATERIALIZED VIEW FOR DELIVERY ANALYTICS
-- =====================================================

CREATE MATERIALIZED VIEW IF NOT EXISTS delivery_movement_log_mv AS
SELECT
    ds.organization_id,
    ds.id AS shipment_id,
    ds.route_id,
    ds.status AS shipment_status,
    ds.created_at AS occurred_at,
    'shipment_created'::text AS entry_type,
    jsonb_build_object(
        'picking_id', ds.picking_id,
        'tracking_number', ds.tracking_number,
        'carrier_name', ds.carrier_name,
        'status', ds.status
    ) AS payload
FROM delivery_shipments ds

UNION ALL

SELECT
    dte.organization_id,
    dte.shipment_id,
    ds.route_id,
    ds.status,
    dte.event_time,
    'tracking_event'::text,
    jsonb_build_object(
        'event_type', dte.event_type,
        'status', dte.status,
        'message', dte.message,
        'source', dte.source,
        'latitude', dte.latitude,
        'longitude', dte.longitude
    )
FROM delivery_tracking_events dte
JOIN delivery_shipments ds ON ds.id = dte.shipment_id

UNION ALL

SELECT
    drs.organization_id,
    COALESCE(drs.shipment_id, ds.id),
    drs.route_id,
    ds.status,
    COALESCE(drs.actual_arrival_at, drs.planned_arrival_at, drs.created_at),
    'route_stop'::text,
    jsonb_build_object(
        'stop_sequence', drs.stop_sequence,
        'status', drs.status,
        'planned_arrival_at', drs.planned_arrival_at,
        'actual_arrival_at', drs.actual_arrival_at
    )
FROM delivery_route_stops drs
LEFT JOIN delivery_shipments ds ON ds.id = drs.shipment_id;

CREATE INDEX IF NOT EXISTS delivery_movement_log_idx
    ON delivery_movement_log_mv (organization_id, shipment_id, occurred_at DESC);

CREATE OR REPLACE FUNCTION refresh_delivery_movement_log()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW delivery_movement_log_mv;
END;
$$;

-- =====================================================
-- QUEUE INTEGRATION FOR DELIVERY TRACKING
-- =====================================================

CREATE OR REPLACE FUNCTION enqueue_delivery_webhook_job(
    p_organization_id uuid,
    p_payload jsonb,
    p_priority int DEFAULT 0,
    p_scheduled_at timestamptz DEFAULT now()
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN enqueue_job(
        p_organization_id := p_organization_id,
        p_queue_name := 'delivery_webhooks',
        p_job_type := 'delivery_webhook_ingest',
        p_payload := COALESCE(p_payload, '{}'::jsonb),
        p_priority := p_priority,
        p_max_attempts := 5,
        p_scheduled_at := p_scheduled_at
    );
END;
$$;

CREATE OR REPLACE FUNCTION handle_delivery_tracking_webhook_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_organization_id uuid;
    v_tracking_number text;
    v_status text;
    v_event_type text;
    v_event_time timestamptz;
    v_shipment_id uuid;
    v_stop_id uuid;
    v_message text;
    v_latitude numeric;
    v_longitude numeric;
    v_event delivery_tracking_events;
BEGIN
    v_organization_id := (p_payload->>'organization_id')::uuid;
    v_tracking_number := p_payload->>'tracking_number';
    v_status := p_payload->>'status';
    v_event_type := COALESCE(p_payload->>'event_type', 'status_update');
    v_event_time := COALESCE((p_payload->>'event_time')::timestamptz, now());
    v_stop_id := CASE
        WHEN (p_payload->>'stop_id') IS NULL THEN NULL
        ELSE (p_payload->>'stop_id')::uuid
    END;
    v_message := p_payload->>'message';
    v_latitude := (p_payload->>'latitude')::numeric;
    v_longitude := (p_payload->>'longitude')::numeric;

    IF v_organization_id IS NULL OR v_tracking_number IS NULL THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'organization_id and tracking_number are required'
        );
    END IF;

    SELECT id INTO v_shipment_id
    FROM delivery_shipments
    WHERE organization_id = v_organization_id
      AND tracking_number = v_tracking_number
    LIMIT 1;

    IF v_shipment_id IS NULL THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Shipment not found for tracking number',
            'tracking_number', v_tracking_number
        );
    END IF;

    v_event := log_delivery_tracking_event(
        p_organization_id := v_organization_id,
        p_shipment_id := v_shipment_id,
        p_event_type := v_event_type,
        p_status := v_status,
        p_event_time := v_event_time,
        p_message := v_message,
        p_stop_id := v_stop_id,
        p_source := p_payload->>'source',
        p_raw_payload := p_payload,
        p_latitude := v_latitude,
        p_longitude := v_longitude,
        p_speed_kph := (p_payload->>'speed_kph')::numeric,
        p_heading := (p_payload->>'heading')::numeric
    );

    RETURN jsonb_build_object(
        'success', true,
        'shipment_id', v_shipment_id,
        'event_id', v_event.id,
        'status', v_event.status
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$;

CREATE OR REPLACE FUNCTION enqueue_delivery_gps_snapshot(
    p_organization_id uuid,
    p_route_id uuid,
    p_assignment_id uuid,
    p_payload jsonb,
    p_priority int DEFAULT 0
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN enqueue_job(
        p_organization_id := p_organization_id,
        p_queue_name := 'delivery_tracking',
        p_job_type := 'delivery_gps_ingest',
        p_payload := jsonb_build_object(
            'organization_id', p_organization_id,
            'route_id', p_route_id,
            'assignment_id', p_assignment_id
        ) || COALESCE(p_payload, '{}'::jsonb),
        p_priority := p_priority,
        p_max_attempts := 3
    );
END;
$$;

CREATE OR REPLACE FUNCTION handle_delivery_gps_ingest_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_organization_id uuid;
    v_route_id uuid;
    v_assignment_id uuid;
    v_vehicle_id uuid;
    v_recorded_at timestamptz;
    v_latitude numeric;
    v_longitude numeric;
    v_position_id uuid;
BEGIN
    v_organization_id := (p_payload->>'organization_id')::uuid;
    v_route_id := (p_payload->>'route_id')::uuid;
    v_assignment_id := CASE
        WHEN (p_payload->>'assignment_id') IS NULL THEN NULL
        ELSE (p_payload->>'assignment_id')::uuid
    END;
    v_vehicle_id := CASE
        WHEN (p_payload->>'vehicle_id') IS NULL THEN NULL
        ELSE (p_payload->>'vehicle_id')::uuid
    END;
    v_recorded_at := COALESCE((p_payload->>'recorded_at')::timestamptz, now());
    v_latitude := (p_payload->>'latitude')::numeric;
    v_longitude := (p_payload->>'longitude')::numeric;

    IF v_organization_id IS NULL OR v_route_id IS NULL OR v_latitude IS NULL OR v_longitude IS NULL THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'organization_id, route_id, latitude and longitude are required'
        );
    END IF;

    INSERT INTO delivery_route_positions (
        organization_id,
        route_id,
        assignment_id,
        vehicle_id,
        recorded_at,
        latitude,
        longitude,
        altitude,
        speed_kph,
        heading,
        source,
        metadata
    )
    VALUES (
        v_organization_id,
        v_route_id,
        v_assignment_id,
        v_vehicle_id,
        v_recorded_at,
        v_latitude,
        v_longitude,
        (p_payload->>'altitude')::numeric,
        (p_payload->>'speed_kph')::numeric,
        (p_payload->>'heading')::numeric,
        COALESCE(p_payload->>'source', 'gps'),
        COALESCE(p_payload->'metadata', '{}'::jsonb)
    )
    RETURNING id INTO v_position_id;

    UPDATE delivery_routes
    SET status = CASE
        WHEN status = 'scheduled' THEN 'in_progress'
        ELSE status
    END,
    actual_start_at = COALESCE(actual_start_at, v_recorded_at),
    actual_end_at = CASE
        WHEN status = 'completed' THEN COALESCE(actual_end_at, v_recorded_at)
        ELSE actual_end_at
    END,
    updated_at = now()
    WHERE id = v_route_id;

    RETURN jsonb_build_object(
        'success', true,
        'position_id', v_position_id
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$;

CREATE OR REPLACE FUNCTION enqueue_refresh_delivery_log(
    p_organization_id uuid,
    p_priority int DEFAULT -5
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN enqueue_job(
        p_organization_id := p_organization_id,
        p_queue_name := 'delivery_tracking',
        p_job_type := 'delivery_log_refresh',
        p_payload := jsonb_build_object('organization_id', p_organization_id),
        p_priority := p_priority,
        p_max_attempts := 2
    );
END;
$$;

CREATE OR REPLACE FUNCTION handle_refresh_delivery_log_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    PERFORM refresh_delivery_movement_log();

    RETURN jsonb_build_object(
        'success', true,
        'message', 'Delivery movement log refreshed'
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$;

-- =====================================================
-- ROW LEVEL SECURITY
-- =====================================================

ALTER TABLE delivery_vehicles ENABLE ROW LEVEL SECURITY;
ALTER TABLE delivery_routes ENABLE ROW LEVEL SECURITY;
ALTER TABLE delivery_route_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE delivery_shipments ENABLE ROW LEVEL SECURITY;
ALTER TABLE delivery_route_stops ENABLE ROW LEVEL SECURITY;
ALTER TABLE delivery_tracking_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE delivery_route_positions ENABLE ROW LEVEL SECURITY;

DO $$
DECLARE
    table_name text;
    tables_list text[] := ARRAY[
        'delivery_vehicles',
        'delivery_routes',
        'delivery_route_assignments',
        'delivery_shipments',
        'delivery_route_stops',
        'delivery_tracking_events',
        'delivery_route_positions'
    ];
BEGIN
    FOREACH table_name IN ARRAY tables_list
    LOOP
        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR SELECT
            USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_select', table_name);

        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR INSERT
            WITH CHECK (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_insert', table_name);

        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR UPDATE
            USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_update', table_name);

        EXECUTE format('
            CREATE POLICY %I ON %I
            FOR DELETE
            USING (organization_id = (SELECT get_current_organization_id()) AND (SELECT user_has_org_access()))
        ', table_name || '_delete', table_name);
    END LOOP;
END $$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE delivery_vehicles IS 'Registered vehicles used for delivery routes';
COMMENT ON TABLE delivery_routes IS 'Delivery routes, internal transfers, and distribution runs';
COMMENT ON TABLE delivery_route_assignments IS 'Assignments linking routes to drivers and vehicles';
COMMENT ON TABLE delivery_shipments IS 'Shipments associated with stock pickings and delivery tracking';
COMMENT ON TABLE delivery_route_stops IS 'Individual stops within a route, linked to shipments when applicable';
COMMENT ON TABLE delivery_tracking_events IS 'Events received from carriers, drivers, or systems updating shipment status';
COMMENT ON TABLE delivery_route_positions IS 'GPS or telemetry positions captured for a delivery route';
COMMENT ON MATERIALIZED VIEW delivery_movement_log_mv IS 'Flattened timeline of shipment lifecycle events for analytics';

COMMENT ON FUNCTION create_delivery_route IS 'Create a new delivery route draft for an organization';
COMMENT ON FUNCTION assign_picking_to_route IS 'Attach a stock picking to a delivery route and create associated shipment';
COMMENT ON FUNCTION log_delivery_tracking_event IS 'Persist an external tracking event and sync shipment status';
COMMENT ON FUNCTION get_delivery_movement_timeline IS 'Return aggregated timeline entries for a shipment';
COMMENT ON FUNCTION get_live_route_snapshot IS 'Return current status, positions, stops, and shipments for a route';
COMMENT ON FUNCTION enqueue_delivery_webhook_job IS 'Queue a delivery webhook payload for asynchronous processing';
COMMENT ON FUNCTION handle_delivery_tracking_webhook_job IS 'Process external carrier webhook payloads into tracking events';
COMMENT ON FUNCTION enqueue_delivery_gps_snapshot IS 'Queue GPS telemetry for ingestion into delivery route positions';
COMMENT ON FUNCTION handle_delivery_gps_ingest_job IS 'Persist GPS telemetry and update related route metadata';
COMMENT ON FUNCTION enqueue_refresh_delivery_log IS 'Queue a background refresh of the delivery movement materialized view';
COMMENT ON FUNCTION handle_refresh_delivery_log_job IS 'Handle refresh jobs for delivery movement log';
COMMENT ON FUNCTION refresh_delivery_movement_log IS 'Refresh the materialized view used for delivery analytics';

-- =====================================================
-- SECURITY HARDENING FOR SECURITY DEFINER FUNCTIONS
-- =====================================================

ALTER FUNCTION create_delivery_route(uuid, text, text, text, uuid, uuid, timestamptz, timestamptz, jsonb) SET search_path TO '';
ALTER FUNCTION assign_picking_to_route(uuid, uuid, uuid, integer, uuid, uuid, jsonb) SET search_path TO '';
ALTER FUNCTION log_delivery_tracking_event(uuid, uuid, text, text, timestamptz, text, uuid, text, jsonb, numeric, numeric, numeric, numeric) SET search_path TO '';
ALTER FUNCTION handle_delivery_tracking_webhook_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_delivery_gps_ingest_job(jsonb) SET search_path TO '';
ALTER FUNCTION handle_refresh_delivery_log_job(jsonb) SET search_path TO '';
