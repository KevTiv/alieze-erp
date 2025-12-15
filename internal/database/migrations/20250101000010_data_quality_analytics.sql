-- Migration: Data Quality & Completeness Analytics
-- Description: Functions to identify incomplete or problematic records
-- Created: 2025-01-01

-- =====================================================
-- DATA QUALITY FUNCTIONS
-- =====================================================

-- Get incomplete contacts (missing critical information)
CREATE OR REPLACE FUNCTION analytics_incomplete_contacts(
    p_organization_id uuid
)
RETURNS TABLE (
    contact_id uuid,
    contact_name varchar,
    contact_type varchar,
    missing_fields text[],
    is_customer boolean,
    is_vendor boolean,
    severity varchar,
    last_updated timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.name,
        c.contact_type,
        ARRAY_REMOVE(ARRAY[
            CASE WHEN c.email IS NULL THEN 'email' END,
            CASE WHEN c.phone IS NULL AND c.mobile IS NULL THEN 'phone' END,
            CASE WHEN c.street IS NULL THEN 'address' END,
            CASE WHEN c.city IS NULL THEN 'city' END,
            CASE WHEN c.country_id IS NULL THEN 'country' END,
            CASE WHEN c.is_customer AND c.payment_term_id IS NULL THEN 'payment_terms' END,
            CASE WHEN c.is_vendor AND c.payment_term_id IS NULL THEN 'payment_terms' END,
            CASE WHEN c.is_company AND c.tax_id IS NULL THEN 'tax_id' END
        ], NULL) as missing_fields,
        c.is_customer,
        c.is_vendor,
        CASE
            WHEN c.email IS NULL AND c.phone IS NULL AND c.mobile IS NULL THEN 'high'
            WHEN c.street IS NULL OR c.country_id IS NULL THEN 'medium'
            ELSE 'low'
        END as severity,
        c.updated_at
    FROM contacts c
    WHERE c.organization_id = p_organization_id
      AND c.deleted_at IS NULL
      AND (
          c.email IS NULL OR
          (c.phone IS NULL AND c.mobile IS NULL) OR
          c.street IS NULL OR
          c.city IS NULL OR
          c.country_id IS NULL OR
          (c.is_customer AND c.payment_term_id IS NULL) OR
          (c.is_vendor AND c.payment_term_id IS NULL) OR
          (c.is_company AND c.tax_id IS NULL)
      )
    ORDER BY
        CASE
            WHEN c.email IS NULL AND c.phone IS NULL AND c.mobile IS NULL THEN 1
            WHEN c.street IS NULL OR c.country_id IS NULL THEN 2
            ELSE 3
        END,
        c.updated_at DESC;
END;
$$;

-- Get incomplete products (missing critical information)
CREATE OR REPLACE FUNCTION analytics_incomplete_products(
    p_organization_id uuid
)
RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    product_type varchar,
    missing_fields text[],
    severity varchar,
    last_updated timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id,
        p.name,
        p.product_type,
        ARRAY_REMOVE(ARRAY[
            CASE WHEN p.default_code IS NULL THEN 'internal_reference' END,
            CASE WHEN p.category_id IS NULL THEN 'category' END,
            CASE WHEN p.uom_id IS NULL THEN 'unit_of_measure' END,
            CASE WHEN p.list_price = 0 THEN 'sale_price' END,
            CASE WHEN p.standard_price = 0 THEN 'cost_price' END,
            CASE WHEN p.product_type = 'storable' AND p.tracking IS NULL THEN 'tracking_method' END,
            CASE WHEN p.image_url IS NULL THEN 'image' END,
            CASE WHEN p.description IS NULL OR p.description = '' THEN 'description' END
        ], NULL) as missing_fields,
        CASE
            WHEN p.list_price = 0 OR p.uom_id IS NULL THEN 'high'
            WHEN p.category_id IS NULL OR p.default_code IS NULL THEN 'medium'
            ELSE 'low'
        END as severity,
        p.updated_at
    FROM products p
    WHERE p.organization_id = p_organization_id
      AND p.deleted_at IS NULL
      AND p.active = true
      AND (
          p.default_code IS NULL OR
          p.category_id IS NULL OR
          p.uom_id IS NULL OR
          p.list_price = 0 OR
          p.standard_price = 0 OR
          (p.product_type = 'storable' AND p.tracking IS NULL) OR
          p.image_url IS NULL OR
          p.description IS NULL OR
          p.description = ''
      )
    ORDER BY
        CASE
            WHEN p.list_price = 0 OR p.uom_id IS NULL THEN 1
            WHEN p.category_id IS NULL OR p.default_code IS NULL THEN 2
            ELSE 3
        END,
        p.updated_at DESC;
END;
$$;

-- Get draft/stuck records (records in draft state too long)
CREATE OR REPLACE FUNCTION analytics_stuck_records(
    p_organization_id uuid,
    p_days_threshold integer DEFAULT 7
)
RETURNS TABLE (
    record_type varchar,
    record_id uuid,
    record_name varchar,
    state varchar,
    days_in_state integer,
    assigned_user_id uuid,
    created_at timestamptz,
    severity varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    -- Sales Orders in draft
    SELECT
        'sales_order'::varchar,
        so.id,
        so.name,
        so.state,
        EXTRACT(EPOCH FROM (now() - so.created_at))::integer / 86400,
        so.user_id,
        so.created_at,
        CASE
            WHEN EXTRACT(EPOCH FROM (now() - so.created_at))::integer / 86400 > p_days_threshold * 2 THEN 'high'
            WHEN EXTRACT(EPOCH FROM (now() - so.created_at))::integer / 86400 > p_days_threshold THEN 'medium'
            ELSE 'low'
        END::varchar
    FROM sales_orders so
    WHERE so.organization_id = p_organization_id
      AND so.state IN ('draft', 'sent')
      AND so.deleted_at IS NULL
      AND EXTRACT(EPOCH FROM (now() - so.created_at))::integer / 86400 > p_days_threshold

    UNION ALL

    -- Invoices in draft
    SELECT
        'invoice'::varchar,
        i.id,
        i.name,
        i.state,
        EXTRACT(EPOCH FROM (now() - i.created_at))::integer / 86400,
        i.user_id,
        i.created_at,
        CASE
            WHEN EXTRACT(EPOCH FROM (now() - i.created_at))::integer / 86400 > p_days_threshold * 2 THEN 'high'
            WHEN EXTRACT(EPOCH FROM (now() - i.created_at))::integer / 86400 > p_days_threshold THEN 'medium'
            ELSE 'low'
        END::varchar
    FROM invoices i
    WHERE i.organization_id = p_organization_id
      AND i.state = 'draft'
      AND i.deleted_at IS NULL
      AND EXTRACT(EPOCH FROM (now() - i.created_at))::integer / 86400 > p_days_threshold

    UNION ALL

    -- Purchase Orders in draft
    SELECT
        'purchase_order'::varchar,
        po.id,
        po.name,
        po.state,
        EXTRACT(EPOCH FROM (now() - po.created_at))::integer / 86400,
        po.user_id,
        po.created_at,
        CASE
            WHEN EXTRACT(EPOCH FROM (now() - po.created_at))::integer / 86400 > p_days_threshold * 2 THEN 'high'
            WHEN EXTRACT(EPOCH FROM (now() - po.created_at))::integer / 86400 > p_days_threshold THEN 'medium'
            ELSE 'low'
        END::varchar
    FROM purchase_orders po
    WHERE po.organization_id = p_organization_id
      AND po.state IN ('draft', 'sent')
      AND po.deleted_at IS NULL
      AND EXTRACT(EPOCH FROM (now() - po.created_at))::integer / 86400 > p_days_threshold

    ORDER BY 6 DESC, 5 DESC;
END;
$$;

-- Get overdue tasks and activities
CREATE OR REPLACE FUNCTION analytics_overdue_items(
    p_organization_id uuid
)
RETURNS TABLE (
    item_type varchar,
    item_id uuid,
    item_name varchar,
    assigned_to uuid,
    due_date date,
    days_overdue integer,
    priority varchar,
    related_record varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    -- Overdue activities
    SELECT
        'activity'::varchar,
        a.id,
        a.summary,
        a.assigned_to,
        a.date_deadline,
        EXTRACT(day FROM CURRENT_DATE - a.date_deadline)::integer,
        'medium'::varchar,
        COALESCE(a.res_model || ':' || a.res_id::text, '')::varchar
    FROM activities a
    WHERE a.organization_id = p_organization_id
      AND a.state = 'planned'
      AND a.date_deadline < CURRENT_DATE

    UNION ALL

    -- Overdue tasks
    SELECT
        'task'::varchar,
        t.id,
        t.name,
        t.user_ids[1], -- First assigned user
        t.date_deadline,
        EXTRACT(day FROM CURRENT_DATE - t.date_deadline)::integer,
        t.priority,
        'project:' || t.project_id::text
    FROM tasks t
    WHERE t.organization_id = p_organization_id
      AND t.date_deadline < CURRENT_DATE
      AND t.kanban_state != 'done'
      AND t.deleted_at IS NULL

    UNION ALL

    -- Overdue invoices
    SELECT
        'invoice'::varchar,
        i.id,
        i.name,
        i.user_id,
        i.invoice_date_due,
        EXTRACT(day FROM CURRENT_DATE - i.invoice_date_due)::integer,
        CASE
            WHEN EXTRACT(day FROM CURRENT_DATE - i.invoice_date_due) > 90 THEN 'urgent'
            WHEN EXTRACT(day FROM CURRENT_DATE - i.invoice_date_due) > 30 THEN 'high'
            ELSE 'medium'
        END::varchar,
        'partner:' || i.partner_id::text
    FROM invoices i
    WHERE i.organization_id = p_organization_id
      AND i.state = 'posted'
      AND i.payment_state IN ('not_paid', 'partial')
      AND i.invoice_date_due < CURRENT_DATE
      AND i.move_type IN ('out_invoice', 'out_refund')
      AND i.deleted_at IS NULL

    ORDER BY 6 DESC, 7 DESC;
END;
$$;

-- Detect duplicate contacts
CREATE OR REPLACE FUNCTION analytics_duplicate_contacts(
    p_organization_id uuid,
    p_similarity_threshold float DEFAULT 0.7
)
RETURNS TABLE (
    contact_id_1 uuid,
    contact_name_1 varchar,
    contact_id_2 uuid,
    contact_name_2 varchar,
    similarity_score float,
    matching_fields text[]
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        c1.id,
        c1.name,
        c2.id,
        c2.name,
        similarity(c1.name, c2.name) as sim_score,
        ARRAY_REMOVE(ARRAY[
            CASE WHEN c1.email = c2.email AND c1.email IS NOT NULL THEN 'email' END,
            CASE WHEN c1.phone = c2.phone AND c1.phone IS NOT NULL THEN 'phone' END,
            CASE WHEN c1.mobile = c2.mobile AND c1.mobile IS NOT NULL THEN 'mobile' END,
            CASE WHEN similarity(c1.name, c2.name) > p_similarity_threshold THEN 'name' END
        ], NULL) as matching
    FROM contacts c1
    JOIN contacts c2 ON c1.organization_id = c2.organization_id
        AND c1.id < c2.id  -- Avoid duplicate pairs
    WHERE c1.organization_id = p_organization_id
      AND c1.deleted_at IS NULL
      AND c2.deleted_at IS NULL
      AND (
          (c1.email = c2.email AND c1.email IS NOT NULL) OR
          (c1.phone = c2.phone AND c1.phone IS NOT NULL) OR
          (c1.mobile = c2.mobile AND c1.mobile IS NOT NULL) OR
          similarity(c1.name, c2.name) > p_similarity_threshold
      )
    ORDER BY similarity(c1.name, c2.name) DESC;
END;
$$;

-- Data quality summary dashboard
CREATE OR REPLACE FUNCTION analytics_data_quality_summary(
    p_organization_id uuid
)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result jsonb;
    v_incomplete_contacts integer;
    v_incomplete_products integer;
    v_stuck_orders integer;
    v_overdue_invoices integer;
    v_duplicate_contacts integer;
BEGIN
    -- Count incomplete contacts
    SELECT COUNT(*) INTO v_incomplete_contacts
    FROM analytics_incomplete_contacts(p_organization_id);

    -- Count incomplete products
    SELECT COUNT(*) INTO v_incomplete_products
    FROM analytics_incomplete_products(p_organization_id);

    -- Count stuck sales orders
    SELECT COUNT(*) INTO v_stuck_orders
    FROM analytics_stuck_records(p_organization_id)
    WHERE record_type = 'sales_order';

    -- Count overdue invoices
    SELECT COUNT(*) INTO v_overdue_invoices
    FROM analytics_overdue_items(p_organization_id)
    WHERE item_type = 'invoice';

    -- Count potential duplicates
    SELECT COUNT(*) INTO v_duplicate_contacts
    FROM analytics_duplicate_contacts(p_organization_id);

    v_result := jsonb_build_object(
        'incomplete_contacts', v_incomplete_contacts,
        'incomplete_products', v_incomplete_products,
        'stuck_orders', v_stuck_orders,
        'overdue_invoices', v_overdue_invoices,
        'duplicate_contacts', v_duplicate_contacts,
        'total_issues', v_incomplete_contacts + v_incomplete_products + v_stuck_orders + v_overdue_invoices + v_duplicate_contacts,
        'generated_at', now()
    );

    RETURN v_result;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION analytics_incomplete_contacts IS 'Find contacts with missing critical information';
COMMENT ON FUNCTION analytics_incomplete_products IS 'Find products with missing critical information';
COMMENT ON FUNCTION analytics_stuck_records IS 'Find records stuck in draft state for too long';
COMMENT ON FUNCTION analytics_overdue_items IS 'Find overdue tasks, activities, and invoices';
COMMENT ON FUNCTION analytics_duplicate_contacts IS 'Detect potential duplicate contacts';
COMMENT ON FUNCTION analytics_data_quality_summary IS 'Summary of all data quality issues';
