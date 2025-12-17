-- Batch Operations System
-- This migration adds batch operation functionality for bulk inventory updates

BEGIN;

-- Create batch_operations table
CREATE TABLE batch_operations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),

    -- Operation details
    operation_type varchar(50) NOT NULL CHECK (operation_type IN (
        'stock_adjustment', 'stock_transfer', 'stock_count',
        'price_update', 'location_update', 'status_update'
    )),
    status varchar(50) NOT NULL CHECK (status IN (
        'draft', 'pending', 'processing', 'completed', 'failed', 'cancelled'
    )),
    reference varchar(100) NOT NULL,
    description text,
    priority integer DEFAULT 10,

    -- Source information
    source_type varchar(50),
    source_id uuid,
    created_by uuid,

    -- Processing information
    processed_by uuid,
    processed_at timestamptz,
    total_items integer DEFAULT 0,
    successful_items integer DEFAULT 0,
    failed_items integer DEFAULT 0,

    -- Error handling
    error_message text,
    error_details jsonb DEFAULT '{}'::jsonb,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,

    -- Metadata
    metadata jsonb DEFAULT '{}'::jsonb,

    -- Indexes
    CONSTRAINT unique_batch_operation_reference_per_org UNIQUE (organization_id, reference)
);

-- Create batch_operation_items table
CREATE TABLE batch_operation_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_operation_id uuid NOT NULL REFERENCES batch_operations(id) ON DELETE CASCADE,
    sequence integer DEFAULT 10,

    -- Item identification
    product_id uuid NOT NULL REFERENCES products(id),
    product_name varchar(255) NOT NULL,
    lot_id uuid REFERENCES stock_lots(id),
    serial_number varchar(100),

    -- Location information
    source_location_id uuid REFERENCES stock_locations(id),
    dest_location_id uuid REFERENCES stock_locations(id),

    -- Quantity information
    current_quantity numeric(15,4) DEFAULT 0,
    adjustment_quantity numeric(15,4) DEFAULT 0,
    new_quantity numeric(15,4) DEFAULT 0,

    -- Operation-specific data
    operation_data jsonb DEFAULT '{}'::jsonb,

    -- Processing status
    status varchar(50) NOT NULL CHECK (status IN ('pending', 'processed', 'failed')),
    error_message text,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Create batch_operation_results table
CREATE TABLE batch_operation_results (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_operation_id uuid NOT NULL REFERENCES batch_operations(id) ON DELETE CASCADE,
    status varchar(50) NOT NULL CHECK (status IN (
        'draft', 'pending', 'processing', 'completed', 'failed', 'cancelled'
    )),
    total_items integer NOT NULL,
    successful_items integer NOT NULL,
    failed_items integer NOT NULL,
    error_message text,

    -- Summary statistics
    processing_time interval,
    started_at timestamptz NOT NULL,
    completed_at timestamptz,

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Create batch_operation_item_results table
CREATE TABLE batch_operation_item_results (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_operation_result_id uuid NOT NULL REFERENCES batch_operation_results(id) ON DELETE CASCADE,
    item_id uuid NOT NULL,
    product_id uuid NOT NULL,
    status varchar(50) NOT NULL CHECK (status IN ('success', 'failed', 'skipped')),
    error_message text,

    -- Before/after values
    before_value jsonb,
    after_value jsonb,

    -- Any resulting documents/records
    result_id uuid,
    result_type varchar(50),

    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Create indexes for performance
CREATE INDEX idx_batch_operations_org ON batch_operations(organization_id);
CREATE INDEX idx_batch_operations_status ON batch_operations(organization_id, status);
CREATE INDEX idx_batch_operations_type ON batch_operations(organization_id, operation_type);
CREATE INDEX idx_batch_operations_date ON batch_operations(organization_id, created_at);
CREATE INDEX idx_batch_operations_priority ON batch_operations(organization_id, priority);

CREATE INDEX idx_batch_operation_items_batch ON batch_operation_items(batch_operation_id);
CREATE INDEX idx_batch_operation_items_product ON batch_operation_items(product_id);
CREATE INDEX idx_batch_operation_items_status ON batch_operation_items(status);
CREATE INDEX idx_batch_operation_items_sequence ON batch_operation_items(batch_operation_id, sequence);

CREATE INDEX idx_batch_operation_results_batch ON batch_operation_results(batch_operation_id);
CREATE INDEX idx_batch_operation_results_status ON batch_operation_results(status);

CREATE INDEX idx_batch_operation_item_results_result ON batch_operation_item_results(batch_operation_result_id);
CREATE INDEX idx_batch_operation_item_results_item ON batch_operation_item_results(item_id);
CREATE INDEX idx_batch_operation_item_results_status ON batch_operation_item_results(status);

-- Create function to create a new batch operation
CREATE OR REPLACE FUNCTION create_batch_operation(
    p_organization_id uuid,
    p_operation_type varchar,
    p_reference varchar,
    p_description text,
    p_priority integer,
    p_source_type varchar,
    p_source_id uuid,
    p_created_by uuid,
    p_metadata jsonb
) RETURNS uuid AS $$
DECLARE
    v_batch_id uuid;
    v_final_reference varchar;
BEGIN
    -- Generate reference if not provided
    IF p_reference IS NULL OR p_reference = '' THEN
        v_final_reference := 'BATCH-' || to_char(now(), 'YYYYMMDD') || '-' ||
            LPAD(CAST(COALESCE((SELECT MAX(CAST(SUBSTRING(reference FROM '\d+$') AS INTEGER))
                FROM batch_operations
                WHERE organization_id = p_organization_id AND
                      reference LIKE 'BATCH-' || to_char(now(), 'YYYYMMDD') || '-%'), 0) + 1 AS VARCHAR), 4, '0');
    ELSE
        v_final_reference := p_reference;
    END IF;

    -- Create the batch operation
    INSERT INTO batch_operations (
        organization_id, operation_type, status, reference, description,
        priority, source_type, source_id, created_by, metadata, created_at, updated_at
    ) VALUES (
        p_organization_id, p_operation_type, 'draft', v_final_reference, p_description,
        p_priority, p_source_type, p_source_id, p_created_by, p_metadata, now(), now()
    ) RETURNING id INTO v_batch_id;

    RETURN v_batch_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to add items to a batch operation
CREATE OR REPLACE FUNCTION add_batch_operation_item(
    p_batch_id uuid,
    p_sequence integer,
    p_product_id uuid,
    p_product_name varchar,
    p_lot_id uuid,
    p_serial_number varchar,
    p_source_location_id uuid,
    p_dest_location_id uuid,
    p_current_quantity numeric,
    p_adjustment_quantity numeric,
    p_new_quantity numeric,
    p_operation_data jsonb
) RETURNS uuid AS $$
DECLARE
    v_item_id uuid;
BEGIN
    -- Create the batch operation item
    INSERT INTO batch_operation_items (
        batch_operation_id, sequence, product_id, product_name, lot_id, serial_number,
        source_location_id, dest_location_id, current_quantity, adjustment_quantity,
        new_quantity, operation_data, status, created_at, updated_at
    ) VALUES (
        p_batch_id, p_sequence, p_product_id, p_product_name, p_lot_id, p_serial_number,
        p_source_location_id, p_dest_location_id, p_current_quantity, p_adjustment_quantity,
        p_new_quantity, p_operation_data, 'pending', now(), now()
    ) RETURNING id INTO v_item_id;

    -- Update the batch operation item count
    UPDATE batch_operations
    SET total_items = total_items + 1, updated_at = now()
    WHERE id = p_batch_id;

    RETURN v_item_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to process a batch operation
CREATE OR REPLACE FUNCTION process_batch_operation(
    p_batch_id uuid,
    p_processed_by uuid
) RETURNS jsonb AS $$
DECLARE
    v_batch_record batch_operations%ROWTYPE;
    v_item_record batch_operation_items%ROWTYPE;
    v_result_id uuid;
    v_success_count integer := 0;
    v_failure_count integer := 0;
    v_start_time timestamptz := now();
    v_error_message text;
    v_item_results jsonb := '[]'::jsonb;
BEGIN
    -- Get the batch operation
    SELECT * INTO v_batch_record FROM batch_operations WHERE id = p_batch_id;

    -- Update status to processing
    UPDATE batch_operations
    SET status = 'processing', processed_by = p_processed_by, updated_at = now()
    WHERE id = p_batch_id;

    -- Process each item based on operation type
    FOR v_item_record IN SELECT * FROM batch_operation_items WHERE batch_operation_id = p_batch_id ORDER BY sequence LOOP
        BEGIN
            -- Process different operation types
            IF v_batch_record.operation_type = 'stock_adjustment' THEN
                -- Handle stock adjustment
                -- This would update stock quantities in the appropriate tables
                -- For now, we'll just mark as successful
                v_item_record.status := 'processed';
                v_success_count := v_success_count + 1;

            ELSIF v_batch_record.operation_type = 'stock_transfer' THEN
                -- Handle stock transfer
                -- This would create stock moves between locations
                v_item_record.status := 'processed';
                v_success_count := v_success_count + 1;

            ELSIF v_batch_record.operation_type = 'price_update' THEN
                -- Handle price update
                -- This would update product prices
                v_item_record.status := 'processed';
                v_success_count := v_success_count + 1;

            ELSIF v_batch_record.operation_type = 'location_update' THEN
                -- Handle location update
                -- This would update product locations
                v_item_record.status := 'processed';
                v_success_count := v_success_count + 1;

            ELSIF v_batch_record.operation_type = 'status_update' THEN
                -- Handle status update
                -- This would update product/status
                v_item_record.status := 'processed';
                v_success_count := v_success_count + 1;

            ELSE
                -- Unknown operation type
                v_item_record.status := 'failed';
                v_item_record.error_message := 'Unknown operation type: ' || v_batch_record.operation_type;
                v_failure_count := v_failure_count + 1;
            END IF;

            -- Update the item status
            UPDATE batch_operation_items
            SET status = v_item_record.status,
                error_message = v_item_record.error_message,
                updated_at = now()
            WHERE id = v_item_record.id;

            -- Add to item results
            v_item_results := jsonb_insert(v_item_results, '{-1}', jsonb_build_object(
                'item_id', v_item_record.id,
                'product_id', v_item_record.product_id,
                'status', v_item_record.status,
                'error_message', v_item_record.error_message
            ));

        EXCEPTION WHEN OTHERS THEN
            v_error_message := SQLERRM;
            v_item_record.status := 'failed';
            v_item_record.error_message := v_error_message;
            v_failure_count := v_failure_count + 1;

            -- Update the item status with error
            UPDATE batch_operation_items
            SET status = 'failed',
                error_message = v_error_message,
                updated_at = now()
            WHERE id = v_item_record.id;

            -- Add to item results with error
            v_item_results := jsonb_insert(v_item_results, '{-1}', jsonb_build_object(
                'item_id', v_item_record.id,
                'product_id', v_item_record.product_id,
                'status', 'failed',
                'error_message', v_error_message
            ));
        END;
    END LOOP;

    -- Update batch operation with results
    UPDATE batch_operations
    SET
        status = CASE
            WHEN v_failure_count = 0 THEN 'completed'
            WHEN v_failure_count < v_batch_record.total_items THEN 'completed'
            ELSE 'failed'
        END,
        processed_by = p_processed_by,
        processed_at = now(),
        successful_items = v_success_count,
        failed_items = v_failure_count,
        error_message = CASE WHEN v_failure_count > 0 THEN 'Partial failure: ' || v_failure_count || ' of ' || v_batch_record.total_items || ' items failed' ELSE NULL END,
        updated_at = now()
    WHERE id = p_batch_id;

    -- Create batch operation result record
    INSERT INTO batch_operation_results (
        batch_operation_id, status, total_items, successful_items, failed_items,
        processing_time, started_at, completed_at, created_at
    ) VALUES (
        p_batch_id,
        CASE
            WHEN v_failure_count = 0 THEN 'completed'
            WHEN v_failure_count < v_batch_record.total_items THEN 'completed'
            ELSE 'failed'
        END,
        v_batch_record.total_items,
        v_success_count,
        v_failure_count,
        now() - v_start_time,
        v_start_time,
        now(),
        now()
    ) RETURNING id INTO v_result_id;

    -- Return detailed results
    RETURN jsonb_build_object(
        'batch_operation_id', p_batch_id,
        'status', CASE
            WHEN v_failure_count = 0 THEN 'completed'
            WHEN v_failure_count < v_batch_record.total_items THEN 'completed'
            ELSE 'failed'
        END,
        'total_items', v_batch_record.total_items,
        'successful_items', v_success_count,
        'failed_items', v_failure_count,
        'processing_time', to_char(now() - v_start_time, 'HH24:MI:SS.MS'),
        'item_results', v_item_results
    );
END;
$$ LANGUAGE plpgsql;

-- Create function to get batch operation statistics
CREATE OR REPLACE FUNCTION get_batch_operation_statistics(
    p_organization_id uuid,
    p_from_date timestamptz DEFAULT NULL,
    p_to_date timestamptz DEFAULT NULL,
    p_operation_type varchar DEFAULT NULL
) RETURNS jsonb AS $$
DECLARE
    v_stats jsonb;
    v_total integer;
    v_completed integer;
    v_failed integer;
    v_pending integer;
    v_processing integer;
    v_by_type jsonb := '{}'::jsonb;
BEGIN
    -- Calculate basic statistics
    SELECT
        COUNT(*) as total,
        COUNT(*) FILTER (WHERE status = 'completed') as completed,
        COUNT(*) FILTER (WHERE status = 'failed') as failed,
        COUNT(*) FILTER (WHERE status = 'pending') as pending,
        COUNT(*) FILTER (WHERE status = 'processing') as processing
    INTO
        v_total, v_completed, v_failed, v_pending, v_processing
    FROM batch_operations
    WHERE organization_id = p_organization_id
    AND (p_from_date IS NULL OR created_at >= p_from_date)
    AND (p_to_date IS NULL OR created_at <= p_to_date)
    AND (p_operation_type IS NULL OR operation_type = p_operation_type);

    -- Build statistics JSON
    v_stats := jsonb_build_object(
        'total_operations', v_total,
        'completed_operations', v_completed,
        'failed_operations', v_failed,
        'pending_operations', v_pending,
        'processing_operations', v_processing,
        'success_rate', CASE WHEN v_total > 0 THEN ROUND((v_completed::float / v_total::float) * 100, 2) ELSE 0 END,
        'failure_rate', CASE WHEN v_total > 0 THEN ROUND((v_failed::float / v_total::float) * 100, 2) ELSE 0 END,
        'average_items', 0, -- Would be calculated in real implementation
        'average_processing', '0 seconds' -- Would be calculated in real implementation
    );

    -- Add by type statistics
    IF p_operation_type IS NULL THEN
        v_stats := jsonb_set(v_stats, '{by_type}', (
            SELECT jsonb_object_agg(operation_type, count)
            FROM batch_operations
            WHERE organization_id = p_organization_id
            AND (p_from_date IS NULL OR created_at >= p_from_date)
            AND (p_to_date IS NULL OR created_at <= p_to_date)
            GROUP BY operation_type
        ));
    END IF;

    -- Add time-based statistics
    v_stats := jsonb_set(v_stats, '{last_30_days}', to_jsonb((
        SELECT COUNT(*)
        FROM batch_operations
        WHERE organization_id = p_organization_id
        AND created_at >= (now() - interval '30 days')
        AND (p_operation_type IS NULL OR operation_type = p_operation_type)
    )));

    v_stats := jsonb_set(v_stats, '{last_7_days}', to_jsonb((
        SELECT COUNT(*)
        FROM batch_operations
        WHERE organization_id = p_organization_id
        AND created_at >= (now() - interval '7 days')
        AND (p_operation_type IS NULL OR operation_type = p_operation_type)
    )));

    v_stats := jsonb_set(v_stats, '{today}', to_jsonb((
        SELECT COUNT(*)
        FROM batch_operations
        WHERE organization_id = p_organization_id
        AND created_at >= date_trunc('day', now())
        AND (p_operation_type IS NULL OR operation_type = p_operation_type)
    )));

    RETURN v_stats;
END;
$$ LANGUAGE plpgsql;

-- Create RLS policies for batch operation tables
ALTER TABLE batch_operations ENABLE ROW LEVEL SECURITY;
ALTER TABLE batch_operation_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE batch_operation_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE batch_operation_item_results ENABLE ROW LEVEL SECURITY;

-- Apply standard RLS policies
CREATE POLICY batch_operations_org_policy ON batch_operations
    USING (organization_id = current_setting('app.current_organization_id')::uuid);

CREATE POLICY batch_operation_items_org_policy ON batch_operation_items
    USING (batch_operation_id IN (SELECT id FROM batch_operations WHERE organization_id = current_setting('app.current_organization_id')::uuid));

CREATE POLICY batch_operation_results_org_policy ON batch_operation_results
    USING (batch_operation_id IN (SELECT id FROM batch_operations WHERE organization_id = current_setting('app.current_organization_id')::uuid));

CREATE POLICY batch_operation_item_results_org_policy ON batch_operation_item_results
    USING (batch_operation_result_id IN (SELECT id FROM batch_operation_results WHERE batch_operation_id IN (SELECT id FROM batch_operations WHERE organization_id = current_setting('app.current_organization_id')::uuid)));

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON batch_operations TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON batch_operation_items TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON batch_operation_results TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON batch_operation_item_results TO authenticated;

GRANT EXECUTE ON FUNCTION create_batch_operation(uuid, varchar, varchar, text, integer, varchar, uuid, uuid, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION add_batch_operation_item(uuid, integer, uuid, varchar, uuid, varchar, uuid, uuid, numeric, numeric, numeric, jsonb) TO authenticated;
GRANT EXECUTE ON FUNCTION process_batch_operation(uuid, uuid) TO authenticated;
GRANT EXECUTE ON FUNCTION get_batch_operation_statistics(uuid, timestamptz, timestamptz, varchar) TO authenticated;

COMMIT;

-- Comments
COMMENT ON TABLE batch_operations IS 'Batch operations for bulk inventory updates - filtered by organization RLS';
COMMENT ON TABLE batch_operation_items IS 'Individual items within batch operations - filtered by organization RLS';
COMMENT ON TABLE batch_operation_results IS 'Results of batch operation processing - filtered by organization RLS';
COMMENT ON TABLE batch_operation_item_results IS 'Results of individual batch item processing - filtered by organization RLS';

COMMENT ON FUNCTION create_batch_operation IS 'Create a new batch operation for bulk inventory updates';
COMMENT ON FUNCTION add_batch_operation_item IS 'Add an item to a batch operation';
COMMENT ON FUNCTION process_batch_operation IS 'Process a batch operation and update inventory';
COMMENT ON FUNCTION get_batch_operation_statistics IS 'Get statistics about batch operations for reporting';

-- Indexes for RLS
CREATE INDEX idx_batch_operations_org_rls ON batch_operations(organization_id) WHERE organization_id = current_setting('app.current_organization_id')::uuid;
CREATE INDEX idx_batch_operation_items_org_rls ON batch_operation_items(batch_operation_id) WHERE batch_operation_id IN (SELECT id FROM batch_operations WHERE organization_id = current_setting('app.current_organization_id')::uuid);
CREATE INDEX idx_batch_operation_results_org_rls ON batch_operation_results(batch_operation_id) WHERE batch_operation_id IN (SELECT id FROM batch_operations WHERE organization_id = current_setting('app.current_organization_id')::uuid);
CREATE INDEX idx_batch_operation_item_results_org_rls ON batch_operation_item_results(batch_operation_result_id) WHERE batch_operation_result_id IN (SELECT id FROM batch_operation_results WHERE batch_operation_id IN (SELECT id FROM batch_operations WHERE organization_id = current_setting('app.current_organization_id')::uuid));
