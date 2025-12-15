-- =====================================================
-- QUEUE JOB HANDLERS
-- =====================================================
-- Pre-built job handlers for common ERP operations
-- These functions can be called directly or enqueued as jobs
-- =====================================================

-- =====================================================
-- EMBEDDING GENERATION JOB
-- =====================================================

CREATE OR REPLACE FUNCTION handle_generate_embedding_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_table_name text;
    v_record_id uuid;
    v_text text;
    v_embedding vector(768);
    v_sql text;
BEGIN
    -- Extract parameters from payload
    v_table_name := p_payload->>'table_name';
    v_record_id := (p_payload->>'record_id')::uuid;
    v_text := p_payload->>'text';

    -- Validate table name to prevent SQL injection
    IF v_table_name NOT IN ('contacts', 'products', 'sales_orders', 'invoices', 'purchase_orders') THEN
        RAISE EXCEPTION 'Invalid table name: %', v_table_name;
    END IF;

    -- Generate embedding
    v_embedding := generate_embedding_ollama_768(v_text);

    -- Update the record with the embedding
    v_sql := format(
        'UPDATE %I SET search_embedding_768 = $1, updated_at = now() WHERE id = $2',
        v_table_name
    );

    EXECUTE v_sql USING v_embedding, v_record_id;

    RETURN jsonb_build_object(
        'success', true,
        'table_name', v_table_name,
        'record_id', v_record_id,
        'embedding_dimensions', array_length(v_embedding::float[], 1)
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$;

-- Helper function to enqueue embedding generation
CREATE OR REPLACE FUNCTION enqueue_generate_embedding(
    p_organization_id uuid,
    p_table_name text,
    p_record_id uuid,
    p_text text
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN enqueue_job(
        p_organization_id := p_organization_id,
        p_queue_name := 'embeddings',
        p_job_type := 'generate_embedding',
        p_payload := jsonb_build_object(
            'table_name', p_table_name,
            'record_id', p_record_id,
            'text', p_text
        ),
        p_priority := 5,
        p_max_attempts := 3
    );
END;
$$;

-- =====================================================
-- INVOICE PROCESSING JOB
-- =====================================================

CREATE OR REPLACE FUNCTION handle_process_invoice_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_invoice_id uuid;
    v_invoice_record RECORD;
    v_result jsonb;
BEGIN
    v_invoice_id := (p_payload->>'invoice_id')::uuid;

    -- Get invoice details
    SELECT * INTO v_invoice_record
    FROM invoices
    WHERE id = v_invoice_id;

    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Invoice not found'
        );
    END IF;

    -- Calculate totals
    PERFORM invoice_compute_totals(v_invoice_id);

    -- Generate embedding for searchability
    PERFORM enqueue_generate_embedding(
        v_invoice_record.organization_id,
        'invoices',
        v_invoice_id,
        COALESCE(v_invoice_record.name, '') || ' ' ||
        COALESCE(v_invoice_record.reference, '')
    );

    RETURN jsonb_build_object(
        'success', true,
        'invoice_id', v_invoice_id,
        'invoice_number', v_invoice_record.name
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
-- DUPLICATE DETECTION JOB
-- =====================================================

CREATE OR REPLACE FUNCTION handle_duplicate_detection_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_organization_id uuid;
    v_entity_type text;
    v_threshold numeric;
    v_duplicates_found int := 0;
    v_duplicate_record RECORD;
    v_temp_table text;
    v_result jsonb;
BEGIN
    v_organization_id := (p_payload->>'organization_id')::uuid;
    v_entity_type := p_payload->>'entity_type'; -- 'contact' or 'product'
    v_threshold := COALESCE((p_payload->>'threshold')::numeric, 0.85);

    -- Create temporary table to store results
    v_temp_table := 'temp_duplicates_' || gen_random_uuid()::text;

    IF v_entity_type = 'contact' THEN
        -- Find duplicate contacts
        EXECUTE format('
            CREATE TEMP TABLE %I AS
            SELECT DISTINCT ON (c1.id)
                c1.id as record_id,
                c2.id as duplicate_id,
                c1.name as record_name,
                c2.name as duplicate_name,
                1 - (c1.search_embedding_768 <=> c2.search_embedding_768) as similarity
            FROM contacts c1
            JOIN contacts c2 ON c1.organization_id = c2.organization_id
                AND c1.id < c2.id
                AND c1.search_embedding_768 IS NOT NULL
                AND c2.search_embedding_768 IS NOT NULL
            WHERE c1.organization_id = %L
                AND (1 - (c1.search_embedding_768 <=> c2.search_embedding_768)) >= %L
            ORDER BY c1.id, (1 - (c1.search_embedding_768 <=> c2.search_embedding_768)) DESC
        ', v_temp_table, v_organization_id, v_threshold);

        EXECUTE format('SELECT COUNT(*) FROM %I', v_temp_table) INTO v_duplicates_found;

    ELSIF v_entity_type = 'product' THEN
        -- Find duplicate products
        EXECUTE format('
            CREATE TEMP TABLE %I AS
            SELECT DISTINCT ON (p1.id)
                p1.id as record_id,
                p2.id as duplicate_id,
                p1.name as record_name,
                p2.name as duplicate_name,
                1 - (p1.search_embedding_768 <=> p2.search_embedding_768) as similarity
            FROM products p1
            JOIN products p2 ON p1.organization_id = p2.organization_id
                AND p1.id < p2.id
                AND p1.search_embedding_768 IS NOT NULL
                AND p2.search_embedding_768 IS NOT NULL
            WHERE p1.organization_id = %L
                AND (1 - (p1.search_embedding_768 <=> p2.search_embedding_768)) >= %L
            ORDER BY p1.id, (1 - (p1.search_embedding_768 <=> p2.search_embedding_768)) DESC
        ', v_temp_table, v_organization_id, v_threshold);

        EXECUTE format('SELECT COUNT(*) FROM %I', v_temp_table) INTO v_duplicates_found;
    END IF;

    -- Build result
    v_result := jsonb_build_object(
        'success', true,
        'entity_type', v_entity_type,
        'duplicates_found', v_duplicates_found,
        'threshold', v_threshold
    );

    -- Cleanup
    EXECUTE format('DROP TABLE IF EXISTS %I', v_temp_table);

    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$;

-- =====================================================
-- EMAIL NOTIFICATION JOB
-- =====================================================

CREATE OR REPLACE FUNCTION handle_send_email_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_to_email text;
    v_subject text;
    v_body text;
    v_from_email text;
BEGIN
    v_to_email := p_payload->>'to_email';
    v_subject := p_payload->>'subject';
    v_body := p_payload->>'body';
    v_from_email := COALESCE(p_payload->>'from_email', 'noreply@pluto-boarding.local');

    -- Note: This is a placeholder. In production, integrate with:
    -- - SMTP server
    -- - SendGrid/Mailgun API
    -- - Supabase Edge Functions
    -- - n8n webhook

    -- For now, just log the email
    RAISE NOTICE 'Email queued: To: %, Subject: %', v_to_email, v_subject;

    RETURN jsonb_build_object(
        'success', true,
        'to_email', v_to_email,
        'subject', v_subject,
        'message', 'Email logged (no actual sending configured)'
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$;

-- Helper function to enqueue email
CREATE OR REPLACE FUNCTION enqueue_send_email(
    p_organization_id uuid,
    p_to_email text,
    p_subject text,
    p_body text,
    p_from_email text DEFAULT 'noreply@pluto-boarding.local'
)
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN enqueue_job(
        p_organization_id := p_organization_id,
        p_queue_name := 'emails',
        p_job_type := 'send_email',
        p_payload := jsonb_build_object(
            'to_email', p_to_email,
            'subject', p_subject,
            'body', p_body,
            'from_email', p_from_email
        ),
        p_priority := 10, -- High priority for emails
        p_max_attempts := 5
    );
END;
$$;

-- =====================================================
-- BATCH DATA EXPORT JOB
-- =====================================================

CREATE OR REPLACE FUNCTION handle_export_data_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_organization_id uuid;
    v_entity_type text;
    v_format text;
    v_filters jsonb;
    v_record_count int;
BEGIN
    v_organization_id := (p_payload->>'organization_id')::uuid;
    v_entity_type := p_payload->>'entity_type';
    v_format := COALESCE(p_payload->>'format', 'json');
    v_filters := COALESCE(p_payload->'filters', '{}'::jsonb);

    -- Count records
    EXECUTE format('
        SELECT COUNT(*)
        FROM %I
        WHERE organization_id = %L
            AND deleted_at IS NULL
    ', v_entity_type, v_organization_id) INTO v_record_count;

    -- In production, this would:
    -- 1. Query the data with filters
    -- 2. Transform to requested format (CSV, Excel, JSON)
    -- 3. Upload to storage
    -- 4. Generate download link
    -- 5. Send email with link

    RETURN jsonb_build_object(
        'success', true,
        'entity_type', v_entity_type,
        'record_count', v_record_count,
        'format', v_format,
        'message', 'Export job completed (placeholder)'
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
-- STOCK REORDER NOTIFICATION JOB
-- =====================================================

CREATE OR REPLACE FUNCTION handle_stock_reorder_check_job(p_payload jsonb)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_organization_id uuid;
    v_low_stock_products RECORD;
    v_products_count int := 0;
    v_notification_sent boolean := false;
BEGIN
    v_organization_id := (p_payload->>'organization_id')::uuid;

    -- Find products below reorder level
    FOR v_low_stock_products IN
        SELECT
            p.id,
            p.name,
            p.default_code,
            SUM(sq.quantity) as current_stock,
            p.metadata->>'reorder_level' as reorder_level
        FROM products p
        LEFT JOIN stock_quants sq ON p.id = sq.product_id
        WHERE p.organization_id = v_organization_id
            AND p.deleted_at IS NULL
            AND p.type = 'product'
            AND p.metadata->>'reorder_level' IS NOT NULL
        GROUP BY p.id, p.name, p.default_code, p.metadata
        HAVING SUM(COALESCE(sq.quantity, 0)) <= (p.metadata->>'reorder_level')::numeric
    LOOP
        v_products_count := v_products_count + 1;

        -- Enqueue notification email
        -- In production, would send actual email to purchasing team
        RAISE NOTICE 'Low stock alert: Product % (%) - Current: %, Reorder at: %',
            v_low_stock_products.name,
            v_low_stock_products.default_code,
            v_low_stock_products.current_stock,
            v_low_stock_products.reorder_level;
    END LOOP;

    IF v_products_count > 0 THEN
        v_notification_sent := true;
    END IF;

    RETURN jsonb_build_object(
        'success', true,
        'low_stock_products', v_products_count,
        'notification_sent', v_notification_sent
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
-- GENERIC JOB DISPATCHER
-- =====================================================

CREATE OR REPLACE FUNCTION dispatch_job(
    p_job_id uuid,
    p_job_type text,
    p_payload jsonb
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_result jsonb;
BEGIN
    -- Route to appropriate handler based on job type
    CASE p_job_type
        WHEN 'generate_embedding' THEN
            v_result := handle_generate_embedding_job(p_payload);

        WHEN 'process_invoice' THEN
            v_result := handle_process_invoice_job(p_payload);

        WHEN 'duplicate_detection' THEN
            v_result := handle_duplicate_detection_job(p_payload);

        WHEN 'send_email' THEN
            v_result := handle_send_email_job(p_payload);

        WHEN 'export_data' THEN
            v_result := handle_export_data_job(p_payload);

        WHEN 'stock_reorder_check' THEN
            v_result := handle_stock_reorder_check_job(p_payload);

        ELSE
            RAISE EXCEPTION 'Unknown job type: %', p_job_type;
    END CASE;

    RETURN v_result;
END;
$$;

-- =====================================================
-- AUTOMATIC TRIGGER EXAMPLES
-- =====================================================

-- Auto-enqueue embedding generation when contact is created/updated
CREATE OR REPLACE FUNCTION trigger_contact_embedding()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    -- Only enqueue if the relevant fields changed
    IF (TG_OP = 'INSERT') OR
       (TG_OP = 'UPDATE' AND (
           OLD.name IS DISTINCT FROM NEW.name OR
           OLD.email IS DISTINCT FROM NEW.email OR
           OLD.phone IS DISTINCT FROM NEW.phone OR
           OLD.comment IS DISTINCT FROM NEW.comment
       )) THEN

        PERFORM enqueue_generate_embedding(
            NEW.organization_id,
            'contacts',
            NEW.id,
            COALESCE(NEW.name, '') || ' ' ||
            COALESCE(NEW.email, '') || ' ' ||
            COALESCE(NEW.phone, '') || ' ' ||
            COALESCE(NEW.comment, '')
        );
    END IF;

    RETURN NEW;
END;
$$;

-- Apply trigger (commented out by default - enable when ready)
-- DROP TRIGGER IF EXISTS contact_embedding_trigger ON contacts;
-- CREATE TRIGGER contact_embedding_trigger
--     AFTER INSERT OR UPDATE ON contacts
--     FOR EACH ROW
--     EXECUTE FUNCTION trigger_contact_embedding();

-- =====================================================
-- HELPER: Bulk enqueue existing records
-- =====================================================

CREATE OR REPLACE FUNCTION bulk_enqueue_embeddings(
    p_organization_id uuid,
    p_table_name text,
    p_batch_size int DEFAULT 100
)
RETURNS jsonb
LANGUAGE plpgsql
AS $$
DECLARE
    v_enqueued_count int := 0;
    v_record RECORD;
    v_text text;
    v_query text;
BEGIN
    -- Validate table name
    IF p_table_name NOT IN ('contacts', 'products', 'sales_orders', 'invoices', 'purchase_orders') THEN
        RAISE EXCEPTION 'Invalid table name: %', p_table_name;
    END IF;

    -- Build query based on table
    v_query := format('
        SELECT id,
               CASE
                   WHEN %L = ''contacts'' THEN
                       COALESCE(name, '''') || '' '' ||
                       COALESCE(email, '''') || '' '' ||
                       COALESCE(phone, '''') || '' '' ||
                       COALESCE(comment, '''')
                   WHEN %L = ''products'' THEN
                       COALESCE(name, '''') || '' '' ||
                       COALESCE(default_code, '''') || '' '' ||
                       COALESCE(description, '''')
                   ELSE name::text
               END as search_text
        FROM %I
        WHERE organization_id = %L
          AND search_embedding_768 IS NULL
          AND deleted_at IS NULL
        LIMIT %s
    ', p_table_name, p_table_name, p_table_name, p_organization_id, p_batch_size);

    FOR v_record IN EXECUTE v_query
    LOOP
        PERFORM enqueue_generate_embedding(
            p_organization_id,
            p_table_name,
            v_record.id,
            v_record.search_text
        );

        v_enqueued_count := v_enqueued_count + 1;
    END LOOP;

    RETURN jsonb_build_object(
        'success', true,
        'table_name', p_table_name,
        'enqueued_count', v_enqueued_count
    );
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION handle_generate_embedding_job IS 'Job handler: Generate embedding for a record';
COMMENT ON FUNCTION handle_process_invoice_job IS 'Job handler: Process and validate invoice';
COMMENT ON FUNCTION handle_duplicate_detection_job IS 'Job handler: Find duplicate records';
COMMENT ON FUNCTION handle_send_email_job IS 'Job handler: Send email notification';
COMMENT ON FUNCTION handle_export_data_job IS 'Job handler: Export data to file';
COMMENT ON FUNCTION handle_stock_reorder_check_job IS 'Job handler: Check stock levels and notify';
COMMENT ON FUNCTION dispatch_job IS 'Routes jobs to appropriate handlers';
COMMENT ON FUNCTION bulk_enqueue_embeddings IS 'Bulk enqueue embedding generation for existing records';
