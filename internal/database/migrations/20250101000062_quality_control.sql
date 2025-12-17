-- Quality Control System
-- This migration adds comprehensive quality control functionality to inventory management

BEGIN;

-- Create quality_control_inspections table
CREATE TABLE quality_control_inspections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),

    -- Reference information
    reference varchar(100) NOT NULL,
    inspection_type varchar(50) NOT NULL CHECK (inspection_type IN ('incoming', 'outgoing', 'internal', 'return')),
    source_document_id uuid,
    source_type varchar(50),

    -- Product information
    product_id uuid NOT NULL REFERENCES products(id),
    product_name varchar(255) NOT NULL,
    lot_id uuid REFERENCES stock_lots(id),
    serial_number varchar(100),
    quantity numeric(15,4) NOT NULL,
    uom_id uuid REFERENCES uom_units(id),

    -- Location information
    location_id uuid NOT NULL REFERENCES stock_locations(id),
    location_name varchar(255) NOT NULL,

    -- Inspection details
    inspection_date timestamptz NOT NULL,
    inspector_id uuid,
    inspection_method varchar(50) NOT NULL CHECK (inspection_method IN ('visual', 'measurement', 'testing', 'sampling')),
    sample_size integer,

    -- Quality status
    status varchar(50) NOT NULL CHECK (status IN ('pending', 'passed', 'failed', 'quarantined', 'rejected')),
    defect_type varchar(100),
    defect_description text,
    defect_quantity numeric(15,4),

    -- Quality metrics
    quality_rating integer,
    compliance_notes text,

    -- Disposition
    disposition varchar(50) CHECK (disposition IN ('accept', 'reject', 'rework', 'scrap', 'return')),
    disposition_date timestamptz,
    disposition_by uuid,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    -- Metadata
    metadata jsonb DEFAULT '{}'::jsonb,

    -- Indexes
    CONSTRAINT unique_qc_inspection_reference_per_org UNIQUE (organization_id, reference)
);

-- Create quality_control_checklists table
CREATE TABLE quality_control_checklists (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    description text,
    product_id uuid REFERENCES products(id),
    product_category_id uuid REFERENCES product_categories(id),
    inspection_type varchar(50) NOT NULL CHECK (inspection_type IN ('incoming', 'outgoing', 'internal')),
    active boolean DEFAULT true,
    priority integer DEFAULT 10,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    -- Indexes
    CONSTRAINT unique_qc_checklist_name_per_org UNIQUE (organization_id, name)
);

-- Create quality_checklist_items table
CREATE TABLE quality_checklist_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    checklist_id uuid NOT NULL REFERENCES quality_control_checklists(id) ON DELETE CASCADE,
    description text NOT NULL,
    criteria text,
    sequence integer DEFAULT 10,
    active boolean DEFAULT true,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Create quality_control_inspection_items table
CREATE TABLE quality_control_inspection_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    inspection_id uuid NOT NULL REFERENCES quality_control_inspections(id) ON DELETE CASCADE,
    checklist_item_id uuid REFERENCES quality_checklist_items(id),
    description text NOT NULL,
    result varchar(50) NOT NULL CHECK (result IN ('pass', 'fail', 'na')),
    notes text,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Create quality_control_alerts table
CREATE TABLE quality_control_alerts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    alert_type varchar(50) NOT NULL CHECK (alert_type IN ('defect', 'quarantine', 'rejection', 'threshold')),
    severity varchar(50) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    title varchar(255) NOT NULL,
    message text NOT NULL,
    related_inspection_id uuid REFERENCES quality_control_inspections(id),
    product_id uuid REFERENCES products(id),
    status varchar(50) NOT NULL CHECK (status IN ('open', 'acknowledged', 'resolved', 'closed')),

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    resolved_at timestamptz,
    resolved_by uuid,

    -- Indexes
    CONSTRAINT unique_qc_alert_title_per_org UNIQUE (organization_id, title)
);

-- Create indexes for performance
CREATE INDEX idx_qc_inspections_org ON quality_control_inspections(organization_id);
CREATE INDEX idx_qc_inspections_product ON quality_control_inspections(product_id);
CREATE INDEX idx_qc_inspections_location ON quality_control_inspections(location_id);
CREATE INDEX idx_qc_inspections_status ON quality_control_inspections(organization_id, status);
CREATE INDEX idx_qc_inspections_date ON quality_control_inspections(organization_id, inspection_date);
CREATE INDEX idx_qc_inspections_lot ON quality_control_inspections(lot_id);

CREATE INDEX idx_qc_checklists_org ON quality_control_checklists(organization_id);
CREATE INDEX idx_qc_checklists_product ON quality_control_checklists(product_id);
CREATE INDEX idx_qc_checklists_category ON quality_control_checklists(product_category_id);
CREATE INDEX idx_qc_checklists_active ON quality_control_checklists(organization_id, active) WHERE active = true;

CREATE INDEX idx_qc_checklist_items_checklist ON quality_checklist_items(checklist_id);
CREATE INDEX idx_qc_checklist_items_active ON quality_checklist_items(active) WHERE active = true;

CREATE INDEX idx_qc_inspection_items_inspection ON quality_control_inspection_items(inspection_id);
CREATE INDEX idx_qc_inspection_items_checklist_item ON quality_control_inspection_items(checklist_item_id);

CREATE INDEX idx_qc_alerts_org ON quality_control_alerts(organization_id);
CREATE INDEX idx_qc_alerts_product ON quality_control_alerts(product_id);
CREATE INDEX idx_qc_alerts_status ON quality_control_alerts(organization_id, status);
CREATE INDEX idx_qc_alerts_severity ON quality_control_alerts(organization_id, severity);

-- Create function to create quality control inspection from stock move
CREATE OR REPLACE FUNCTION create_qc_inspection_from_stock_move(
    p_stock_move_id uuid,
    p_inspector_id uuid,
    p_checklist_id uuid DEFAULT NULL,
    p_inspection_method varchar DEFAULT 'visual',
    p_sample_size integer DEFAULT NULL
) RETURNS uuid AS $$
DECLARE
    v_inspection_id uuid;
    v_reference varchar;
    v_product_id uuid;
    v_product_name varchar;
    v_location_id uuid;
    v_location_name varchar;
    v_quantity numeric;
    v_uom_id uuid;
    v_lot_id uuid;
BEGIN
    -- Get stock move details
    SELECT
        sm.product_id,
        p.name,
        sm.location_dest_id,
        sl.name,
        sm.product_uom_qty,
        sm.product_uom,
        sm.lot_ids[1] as lot_id
    INTO
        v_product_id, v_product_name, v_location_id, v_location_name, v_quantity, v_uom_id, v_lot_id
    FROM stock_moves sm
    JOIN products p ON sm.product_id = p.id
    JOIN stock_locations sl ON sm.location_dest_id = sl.id
    WHERE sm.id = p_stock_move_id AND sm.organization_id = current_setting('app.current_organization_id')::uuid;

    -- Generate reference
    v_reference := 'QC-' || to_char(now(), 'YYYYMMDD') || '-' || LPAD(CAST(COALESCE((SELECT MAX(CAST(SUBSTRING(reference FROM '\d+$') AS INTEGER))
        FROM quality_control_inspections
        WHERE organization_id = current_setting('app.current_organization_id')::uuid AND
              reference LIKE 'QC-' || to_char(now(), 'YYYYMMDD') || '-%'), 0) + 1 AS VARCHAR), 4, '0');

    -- Create inspection record
    INSERT INTO quality_control_inspections (
        organization_id, reference, inspection_type, source_document_id, source_type,
        product_id, product_name, lot_id, quantity, uom_id, location_id, location_name,
        inspection_date, inspector_id, inspection_method, sample_size, status, created_at, updated_at
    ) VALUES (
        current_setting('app.current_organization_id')::uuid,
        v_reference,
        'incoming',
        p_stock_move_id,
        'stock_move',
        v_product_id, v_product_name, v_lot_id, v_quantity, v_uom_id, v_location_id, v_location_name,
        now(), p_inspector_id, p_inspection_method, p_sample_size, 'pending', now(), now()
    ) RETURNING id INTO v_inspection_id;

    -- If checklist is provided, create inspection items from checklist
    IF p_checklist_id IS NOT NULL THEN
        INSERT INTO quality_control_inspection_items (
            inspection_id, checklist_item_id, description, result, created_at
        )
        SELECT
            v_inspection_id,
            qci.id,
            qci.description,
            'pending',
            now()
        FROM quality_checklist_items qci
        WHERE qci.checklist_id = p_checklist_id AND qci.active = true
        ORDER BY qci.sequence;
    END IF;

    RETURN v_inspection_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to update quality control inspection status
CREATE OR REPLACE FUNCTION update_qc_inspection_status(
    p_inspection_id uuid,
    p_status varchar,
    p_defect_type varchar DEFAULT NULL,
    p_defect_description text DEFAULT NULL,
    p_defect_quantity numeric DEFAULT NULL,
    p_quality_rating integer DEFAULT NULL,
    p_compliance_notes text DEFAULT NULL,
    p_disposition varchar DEFAULT NULL
) RETURNS void AS $$
BEGIN
    UPDATE quality_control_inspections
    SET
        status = p_status,
        defect_type = p_defect_type,
        defect_description = p_defect_description,
        defect_quantity = p_defect_quantity,
        quality_rating = p_quality_rating,
        compliance_notes = p_compliance_notes,
        disposition = p_disposition,
        disposition_date = CASE WHEN p_disposition IS NOT NULL THEN now() ELSE NULL END,
        updated_at = now()
    WHERE id = p_inspection_id AND organization_id = current_setting('app.current_organization_id')::uuid;
END;
$$ LANGUAGE plpgsql;

-- Create function to complete quality control inspection
CREATE OR REPLACE FUNCTION complete_qc_inspection(
    p_inspection_id uuid,
    p_status varchar,
    p_results jsonb
) RETURNS void AS $$
DECLARE
    v_inspection_record quality_control_inspections%ROWTYPE;
    v_passed_count integer := 0;
    v_failed_count integer := 0;
    v_total_count integer := 0;
    v_result_record jsonb;
BEGIN
    -- Get inspection record
    SELECT * INTO v_inspection_record
    FROM quality_control_inspections
    WHERE id = p_inspection_id AND organization_id = current_setting('app.current_organization_id')::uuid;

    -- Update inspection items from results
    FOR v_result_record IN SELECT * FROM jsonb_array_elements(p_results) LOOP
        UPDATE quality_control_inspection_items
        SET
            result = v_result_record->>'result',
            notes = v_result_record->>'notes'
        WHERE id = (v_result_record->>'id')::uuid;

        -- Count results
        v_total_count := v_total_count + 1;
        IF v_result_record->>'result' = 'pass' THEN
            v_passed_count := v_passed_count + 1;
        ELSIF v_result_record->>'result' = 'fail' THEN
            v_failed_count := v_failed_count + 1;
        END IF;
    END LOOP;

    -- Determine overall status and quality rating
    DECLARE
        v_final_status varchar;
        v_quality_rating integer;
        v_defect_type varchar;
        v_defect_description text;
    BEGIN
        -- Calculate quality rating based on pass/fail ratio
        IF v_total_count > 0 THEN
            v_quality_rating := ROUND((v_passed_count::float / v_total_count::float) * 100);
        ELSE
            v_quality_rating := 100;
        END IF;

        -- Determine final status
        IF v_failed_count = 0 THEN
            v_final_status := 'passed';
        ELSIF v_failed_count > 0 AND v_failed_count <= (v_total_count * 0.1) THEN
            v_final_status := 'quarantined';
            v_defect_type := 'minor_defects';
            v_defect_description := 'Minor defects detected - quarantined for review';
        ELSE
            v_final_status := 'failed';
            v_defect_type := 'major_defects';
            v_defect_description := 'Major defects detected - failed quality control';
        END IF;

        -- Update inspection status
        CALL update_qc_inspection_status(
            p_inspection_id,
            v_final_status,
            v_defect_type,
            v_defect_description,
            NULL, -- defect_quantity
            v_quality_rating,
            NULL, -- compliance_notes
            CASE
                WHEN v_final_status = 'passed' THEN 'accept'
                WHEN v_final_status = 'quarantined' THEN 'rework'
                ELSE 'reject'
            END
        );
    END;
END;
$$ LANGUAGE plpgsql;

-- Create function to get quality control statistics
CREATE OR REPLACE FUNCTION get_quality_control_statistics(
    p_organization_id uuid,
    p_date_from timestamptz DEFAULT NULL,
    p_date_to timestamptz DEFAULT NULL,
    p_product_id uuid DEFAULT NULL
) RETURNS jsonb AS $$
DECLARE
    v_stats jsonb;
    v_total_inspections integer;
    v_passed_inspections integer;
    v_failed_inspections integer;
    v_quarantined_items integer;
    v_rejected_items integer;
    v_avg_quality_rating numeric;
    v_defect_rate numeric;
    v_avg_inspection_time interval;
BEGIN
    -- Calculate basic statistics
    SELECT
        COUNT(*) as total_inspections,
        COUNT(*) FILTER (WHERE status = 'passed') as passed_inspections,
        COUNT(*) FILTER (WHERE status = 'failed') as failed_inspections,
        COUNT(*) FILTER (WHERE status = 'quarantined') as quarantined_items,
        COUNT(*) FILTER (WHERE disposition = 'reject') as rejected_items,
        AVG(quality_rating) as avg_quality_rating,
        AVG(defect_quantity) as defect_rate,
        AVG(AGE(updated_at, created_at)) as avg_inspection_time
    INTO
        v_total_inspections, v_passed_inspections, v_failed_inspections,
        v_quarantined_items, v_rejected_items, v_avg_quality_rating,
        v_defect_rate, v_avg_inspection_time
    FROM quality_control_inspections
    WHERE organization_id = p_organization_id
    AND (p_date_from IS NULL OR inspection_date >= p_date_from)
    AND (p_date_to IS NULL OR inspection_date <= p_date_to)
    AND (p_product_id IS NULL OR product_id = p_product_id);

    -- Build statistics JSON
    v_stats := jsonb_build_object(
        'total_inspections', v_total_inspections,
        'passed_inspections', v_passed_inspections,
        'failed_inspections', v_failed_inspections,
        'quarantined_items', v_quarantined_items,
        'rejected_items', v_rejected_items,
        'pass_rate', CASE WHEN v_total_inspections > 0 THEN ROUND((v_passed_inspections::float / v_total_inspections::float) * 100, 2) ELSE 0 END,
        'fail_rate', CASE WHEN v_total_inspections > 0 THEN ROUND((v_failed_inspections::float / v_total_inspections::float) * 100, 2) ELSE 0 END,
        'average_quality_rating', COALESCE(ROUND(v_avg_quality_rating, 1), 0),
        'defect_rate', COALESCE(ROUND(v_defect_rate, 2), 0),
        'inspection_time', COALESCE(v_avg_inspection_time::text, '0 seconds'),
        'quality_trend', 'stable' -- Would be calculated based on trends in real implementation
    );

    -- Add top defect types
    v_stats := jsonb_set(v_stats, '{top_defect_types}',
        COALESCE(
            (SELECT jsonb_agg(jsonb_build_object(
                'defect_type', defect_type,
                'count', count,
                'percentage', ROUND((count::float / NULLIF(v_total_inspections, 0)::float) * 100, 1)
            ) ORDER BY count DESC LIMIT 5)
            FROM (
                SELECT
                    defect_type,
                    COUNT(*) as count
                FROM quality_control_inspections
                WHERE organization_id = p_organization_id
                AND defect_type IS NOT NULL
                AND (p_date_from IS NULL OR inspection_date >= p_date_from)
                AND (p_date_to IS NULL OR inspection_date <= p_date_to)
                AND (p_product_id IS NULL OR product_id = p_product_id)
                GROUP BY defect_type
            ) as defect_counts
        ), '[]'::jsonb));

    RETURN v_stats;
END;
$$ LANGUAGE plpgsql;

-- Create function to create quality control alert
CREATE OR REPLACE FUNCTION create_quality_control_alert(
    p_organization_id uuid,
    p_alert_type varchar,
    p_severity varchar,
    p_title varchar,
    p_message text,
    p_related_inspection_id uuid DEFAULT NULL,
    p_product_id uuid DEFAULT NULL
) RETURNS uuid AS $$
DECLARE
    v_alert_id uuid;
BEGIN
    INSERT INTO quality_control_alerts (
        organization_id, alert_type, severity, title, message,
        related_inspection_id, product_id, status, created_at, updated_at
    ) VALUES (
        p_organization_id, p_alert_type, p_severity, p_title, p_message,
        p_related_inspection_id, p_product_id, 'open', now(), now()
    ) RETURNING id INTO v_alert_id;

    RETURN v_alert_id;
END;
$$ LANGUAGE plpgsql;

-- Create RLS policies for quality control tables
ALTER TABLE quality_control_inspections ENABLE ROW LEVEL SECURITY;
ALTER TABLE quality_control_checklists ENABLE ROW LEVEL SECURITY;
ALTER TABLE quality_checklist_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE quality_control_inspection_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE quality_control_alerts ENABLE ROW LEVEL SECURITY;

-- Apply standard RLS policies
CREATE POLICY qc_inspections_org_policy ON quality_control_inspections
    USING (organization_id = current_setting('app.current_organization_id')::uuid);

CREATE POLICY qc_checklists_org_policy ON quality_control_checklists
    USING (organization_id = current_setting('app.current_organization_id')::uuid);

CREATE POLICY qc_checklist_items_org_policy ON quality_checklist_items
    USING (checklist_id IN (SELECT id FROM quality_control_checklists WHERE organization_id = current_setting('app.current_organization_id')::uuid));

CREATE POLICY qc_inspection_items_org_policy ON quality_control_inspection_items
    USING (inspection_id IN (SELECT id FROM quality_control_inspections WHERE organization_id = current_setting('app.current_organization_id')::uuid));

CREATE POLICY qc_alerts_org_policy ON quality_control_alerts
    USING (organization_id = current_setting('app.current_organization_id')::uuid);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON quality_control_inspections TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON quality_control_checklists TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON quality_checklist_items TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON quality_control_inspection_items TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON quality_control_alerts TO authenticated;

GRANT EXECUTE ON FUNCTION create_qc_inspection_from_stock_move(uuid, uuid, uuid, varchar, integer) TO authenticated;
GRANT EXECUTE ON FUNCTION update_qc_inspection_status(uuid, varchar, varchar, text, numeric, integer, text, varchar) TO authenticated;
GRANT EXECUTE ON FUNCTION complete_qc_inspection(uuid, varchar, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION get_quality_control_statistics(uuid, timestamptz, timestamptz, uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION create_quality_control_alert(uuid, varchar, varchar, varchar, text, uuid, uuid) TO authenticated;

COMMIT;

-- Comments
COMMENT ON TABLE quality_control_inspections IS 'Quality control inspection records - filtered by organization RLS';
COMMENT ON TABLE quality_control_checklists IS 'Quality control checklist templates - filtered by organization RLS';
COMMENT ON TABLE quality_checklist_items IS 'Quality control checklist items - filtered by organization RLS';
COMMENT ON TABLE quality_control_inspection_items IS 'Quality control inspection item results - filtered by organization RLS';
COMMENT ON TABLE quality_control_alerts IS 'Quality control alerts and notifications - filtered by organization RLS';

COMMENT ON FUNCTION create_qc_inspection_from_stock_move IS 'Create quality control inspection from a stock move';
COMMENT ON FUNCTION update_qc_inspection_status IS 'Update quality control inspection status and results';
COMMENT ON FUNCTION complete_qc_inspection IS 'Complete quality control inspection with results and determine final status';
COMMENT ON FUNCTION get_quality_control_statistics IS 'Get quality control statistics and metrics for reporting';
COMMENT ON FUNCTION create_quality_control_alert IS 'Create quality control alert for failed inspections or quality issues';

-- Indexes for RLS
CREATE INDEX idx_qc_inspections_org_rls ON quality_control_inspections(organization_id) WHERE organization_id = current_setting('app.current_organization_id')::uuid;
CREATE INDEX idx_qc_checklists_org_rls ON quality_control_checklists(organization_id) WHERE organization_id = current_setting('app.current_organization_id')::uuid;
CREATE INDEX idx_qc_alerts_org_rls ON quality_control_alerts(organization_id) WHERE organization_id = current_setting('app.current_organization_id')::uuid;
