-- Migration: Duplicate Detection System
-- Description: AI-powered duplicate detection using vector similarity for contacts, products, and more
-- Created: 2025-01-01

-- =====================================================
-- DUPLICATE DETECTION FUNCTIONS
-- =====================================================

-- Find duplicate contacts
CREATE OR REPLACE FUNCTION find_duplicate_contacts(
    p_organization_id uuid,
    p_contact_id uuid DEFAULT NULL,
    p_name text DEFAULT NULL,
    p_email text DEFAULT NULL,
    p_phone text DEFAULT NULL,
    p_similarity_threshold float DEFAULT 0.85,
    p_limit integer DEFAULT 10
)
RETURNS TABLE (
    contact_id uuid,
    contact_name text,
    contact_email text,
    contact_phone text,
    similarity_score float,
    match_reason text,
    is_customer boolean,
    is_vendor boolean,
    created_at timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_search_embedding vector(768);
    v_search_text text;
BEGIN
    -- Build search text from provided parameters
    v_search_text := COALESCE(p_name, '') || ' ' ||
                     COALESCE(p_email, '') || ' ' ||
                     COALESCE(p_phone, '');

    -- Generate embedding for search
    v_search_embedding := generate_search_embedding(v_search_text);

    RETURN QUERY
    WITH similarity_matches AS (
        SELECT
            c.id,
            c.name,
            c.email,
            c.phone,
            1 - (c.search_embedding <=> v_search_embedding) as vector_similarity,
            c.is_customer,
            c.is_vendor,
            c.created_at,
            -- Calculate exact match scores
            CASE
                WHEN p_email IS NOT NULL AND LOWER(c.email) = LOWER(p_email) THEN 1.0
                ELSE 0.0
            END as email_match,
            CASE
                WHEN p_phone IS NOT NULL AND
                     regexp_replace(c.phone, '[^0-9]', '', 'g') =
                     regexp_replace(p_phone, '[^0-9]', '', 'g')
                THEN 1.0
                ELSE 0.0
            END as phone_match,
            -- Levenshtein distance for name similarity (if pg_trgm available)
            CASE
                WHEN p_name IS NOT NULL THEN
                    similarity(LOWER(c.name), LOWER(p_name))
                ELSE 0.0
            END as name_similarity
        FROM contacts c
        WHERE c.organization_id = p_organization_id
            AND c.deleted_at IS NULL
            AND c.search_embedding IS NOT NULL
            AND (p_contact_id IS NULL OR c.id != p_contact_id) -- Exclude self
    ),
    scored_matches AS (
        SELECT
            *,
            -- Weighted score: email (40%), phone (30%), name (20%), vector (10%)
            (email_match * 0.4 +
             phone_match * 0.3 +
             name_similarity * 0.2 +
             vector_similarity * 0.1) as combined_score,
            -- Determine primary match reason
            CASE
                WHEN email_match = 1.0 THEN 'exact_email_match'
                WHEN phone_match = 1.0 THEN 'exact_phone_match'
                WHEN name_similarity > 0.8 THEN 'similar_name'
                WHEN vector_similarity > p_similarity_threshold THEN 'semantic_similarity'
                ELSE 'low_confidence'
            END as reason
        FROM similarity_matches
    )
    SELECT
        sm.id,
        sm.name,
        sm.email,
        sm.phone,
        sm.combined_score,
        sm.reason,
        sm.is_customer,
        sm.is_vendor,
        sm.created_at
    FROM scored_matches sm
    WHERE sm.combined_score >= p_similarity_threshold
    ORDER BY sm.combined_score DESC, sm.created_at DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION find_duplicate_contacts IS 'Find duplicate contacts using multiple matching strategies (email, phone, name, semantic)';

-- Find duplicate products
CREATE OR REPLACE FUNCTION find_duplicate_products(
    p_organization_id uuid,
    p_product_id uuid DEFAULT NULL,
    p_name text DEFAULT NULL,
    p_default_code text DEFAULT NULL,
    p_barcode text DEFAULT NULL,
    p_similarity_threshold float DEFAULT 0.85,
    p_limit integer DEFAULT 10
)
RETURNS TABLE (
    product_id uuid,
    product_name text,
    default_code text,
    barcode text,
    similarity_score float,
    match_reason text,
    list_price numeric,
    active boolean,
    created_at timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_search_embedding vector(768);
    v_search_text text;
BEGIN
    -- Build search text
    v_search_text := COALESCE(p_name, '') || ' ' ||
                     COALESCE(p_default_code, '') || ' ' ||
                     COALESCE(p_barcode, '');

    -- Generate embedding
    v_search_embedding := generate_search_embedding(v_search_text);

    RETURN QUERY
    WITH similarity_matches AS (
        SELECT
            p.id,
            p.name,
            p.default_code,
            p.barcode,
            1 - (p.search_embedding <=> v_search_embedding) as vector_similarity,
            p.list_price,
            p.active,
            p.created_at,
            -- Exact matches
            CASE
                WHEN p_barcode IS NOT NULL AND p.barcode = p_barcode THEN 1.0
                ELSE 0.0
            END as barcode_match,
            CASE
                WHEN p_default_code IS NOT NULL AND
                     LOWER(p.default_code) = LOWER(p_default_code)
                THEN 1.0
                ELSE 0.0
            END as code_match,
            -- Name similarity
            CASE
                WHEN p_name IS NOT NULL THEN
                    similarity(LOWER(p.name), LOWER(p_name))
                ELSE 0.0
            END as name_similarity
        FROM products p
        WHERE p.organization_id = p_organization_id
            AND p.deleted_at IS NULL
            AND p.search_embedding IS NOT NULL
            AND (p_product_id IS NULL OR p.id != p_product_id)
    ),
    scored_matches AS (
        SELECT
            *,
            -- Weighted: barcode (50%), code (30%), name (15%), vector (5%)
            (barcode_match * 0.5 +
             code_match * 0.3 +
             name_similarity * 0.15 +
             vector_similarity * 0.05) as combined_score,
            CASE
                WHEN barcode_match = 1.0 THEN 'exact_barcode_match'
                WHEN code_match = 1.0 THEN 'exact_code_match'
                WHEN name_similarity > 0.8 THEN 'similar_name'
                WHEN vector_similarity > p_similarity_threshold THEN 'semantic_similarity'
                ELSE 'low_confidence'
            END as reason
        FROM similarity_matches
    )
    SELECT
        sm.id,
        sm.name,
        sm.default_code,
        sm.barcode,
        sm.combined_score,
        sm.reason,
        sm.list_price,
        sm.active,
        sm.created_at
    FROM scored_matches sm
    WHERE sm.combined_score >= p_similarity_threshold
    ORDER BY sm.combined_score DESC, sm.created_at DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION find_duplicate_products IS 'Find duplicate products using barcode, code, name, and semantic similarity';

-- =====================================================
-- AUTOMATIC DUPLICATE DETECTION
-- =====================================================

-- Trigger to check for duplicates when inserting new contact
CREATE OR REPLACE FUNCTION check_contact_duplicates_on_insert()
RETURNS TRIGGER AS $$
DECLARE
    v_duplicate_count integer;
    v_config record;
BEGIN
    -- Get duplicate detection config
    SELECT * INTO v_config
    FROM duplicate_detection_config
    WHERE organization_id = NEW.organization_id
        AND entity_type = 'contact'
        AND auto_merge = false -- Only alert, don't auto-merge on insert
    LIMIT 1;

    -- If no config found, skip check
    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    -- Check for duplicates
    WITH duplicates AS (
        SELECT contact_id, similarity_score
        FROM find_duplicate_contacts(
            NEW.organization_id,
            NULL, -- Don't exclude any ID since this is new
            NEW.name,
            NEW.email,
            NEW.phone,
            v_config.similarity_threshold,
            5
        )
    )
    SELECT COUNT(*) INTO v_duplicate_count FROM duplicates;

    -- If duplicates found, record them
    IF v_duplicate_count > 0 THEN
        INSERT INTO duplicate_candidates (
            organization_id,
            entity_type,
            record_id_1,
            record_id_2,
            similarity_score,
            status
        )
        SELECT
            NEW.organization_id,
            'contact',
            NEW.id,
            d.contact_id,
            d.similarity_score,
            'pending'
        FROM duplicates d;

        -- Optionally raise a notice (won't block insert)
        RAISE NOTICE 'Found % potential duplicate(s) for contact: %', v_duplicate_count, NEW.name;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_contact_duplicates
    AFTER INSERT ON contacts
    FOR EACH ROW
    EXECUTE FUNCTION check_contact_duplicates_on_insert();

-- Similar trigger for products
CREATE OR REPLACE FUNCTION check_product_duplicates_on_insert()
RETURNS TRIGGER AS $$
DECLARE
    v_duplicate_count integer;
    v_config record;
BEGIN
    SELECT * INTO v_config
    FROM duplicate_detection_config
    WHERE organization_id = NEW.organization_id
        AND entity_type = 'product'
        AND auto_merge = false
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    WITH duplicates AS (
        SELECT product_id, similarity_score
        FROM find_duplicate_products(
            NEW.organization_id,
            NULL,
            NEW.name,
            NEW.default_code,
            NEW.barcode,
            v_config.similarity_threshold,
            5
        )
    )
    SELECT COUNT(*) INTO v_duplicate_count FROM duplicates;

    IF v_duplicate_count > 0 THEN
        INSERT INTO duplicate_candidates (
            organization_id,
            entity_type,
            record_id_1,
            record_id_2,
            similarity_score,
            status
        )
        SELECT
            NEW.organization_id,
            'product',
            NEW.id,
            d.product_id,
            d.similarity_score,
            'pending'
        FROM duplicates d;

        RAISE NOTICE 'Found % potential duplicate(s) for product: %', v_duplicate_count, NEW.name;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_product_duplicates
    AFTER INSERT ON products
    FOR EACH ROW
    EXECUTE FUNCTION check_product_duplicates_on_insert();

-- =====================================================
-- DUPLICATE MANAGEMENT FUNCTIONS
-- =====================================================

-- Get pending duplicates for review
CREATE OR REPLACE FUNCTION get_pending_duplicates(
    p_organization_id uuid,
    p_entity_type text DEFAULT NULL,
    p_limit integer DEFAULT 20
)
RETURNS TABLE (
    duplicate_id uuid,
    entity_type text,
    record_1_id uuid,
    record_1_name text,
    record_2_id uuid,
    record_2_name text,
    similarity_score float,
    created_at timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        dc.id,
        dc.entity_type,
        dc.record_id_1,
        CASE dc.entity_type
            WHEN 'contact' THEN (SELECT name FROM contacts WHERE id = dc.record_id_1)
            WHEN 'product' THEN (SELECT name FROM products WHERE id = dc.record_id_1)
            ELSE NULL
        END,
        dc.record_id_2,
        CASE dc.entity_type
            WHEN 'contact' THEN (SELECT name FROM contacts WHERE id = dc.record_id_2)
            WHEN 'product' THEN (SELECT name FROM products WHERE id = dc.record_id_2)
            ELSE NULL
        END,
        dc.similarity_score,
        dc.created_at
    FROM duplicate_candidates dc
    WHERE dc.organization_id = p_organization_id
        AND dc.status = 'pending'
        AND (p_entity_type IS NULL OR dc.entity_type = p_entity_type)
    ORDER BY dc.similarity_score DESC, dc.created_at DESC
    LIMIT p_limit;
END;
$$;

-- Mark duplicates as reviewed
CREATE OR REPLACE FUNCTION mark_duplicate_reviewed(
    p_duplicate_id uuid,
    p_user_id uuid,
    p_action text -- 'merged', 'ignored', 'confirmed_different'
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE duplicate_candidates
    SET
        status = p_action,
        reviewed_by = p_user_id,
        reviewed_at = now()
    WHERE id = p_duplicate_id;
END;
$$;

-- Batch check all existing records for duplicates
CREATE OR REPLACE FUNCTION batch_check_duplicates(
    p_organization_id uuid,
    p_entity_type text,
    p_batch_size integer DEFAULT 100
)
RETURNS TABLE (
    total_checked integer,
    duplicates_found integer
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_total_checked integer := 0;
    v_duplicates_found integer := 0;
    v_record record;
    v_dups integer;
BEGIN
    -- Check contacts
    IF p_entity_type = 'contact' THEN
        FOR v_record IN
            SELECT id, name, email, phone
            FROM contacts
            WHERE organization_id = p_organization_id
                AND deleted_at IS NULL
            LIMIT p_batch_size
        LOOP
            v_total_checked := v_total_checked + 1;

            WITH duplicates AS (
                SELECT contact_id
                FROM find_duplicate_contacts(
                    p_organization_id,
                    v_record.id,
                    v_record.name,
                    v_record.email,
                    v_record.phone,
                    0.85,
                    5
                )
            )
            SELECT COUNT(*) INTO v_dups FROM duplicates;

            v_duplicates_found := v_duplicates_found + v_dups;
        END LOOP;
    END IF;

    -- Check products
    IF p_entity_type = 'product' THEN
        FOR v_record IN
            SELECT id, name, default_code, barcode
            FROM products
            WHERE organization_id = p_organization_id
                AND deleted_at IS NULL
            LIMIT p_batch_size
        LOOP
            v_total_checked := v_total_checked + 1;

            WITH duplicates AS (
                SELECT product_id
                FROM find_duplicate_products(
                    p_organization_id,
                    v_record.id,
                    v_record.name,
                    v_record.default_code,
                    v_record.barcode,
                    0.85,
                    5
                )
            )
            SELECT COUNT(*) INTO v_dups FROM duplicates;

            v_duplicates_found := v_duplicates_found + v_dups;
        END LOOP;
    END IF;

    RETURN QUERY SELECT v_total_checked, v_duplicates_found;
END;
$$;

COMMENT ON FUNCTION batch_check_duplicates IS 'Batch process to check existing records for duplicates (run as background job)';

-- =====================================================
-- DEFAULT CONFIGURATION
-- =====================================================

-- Insert default duplicate detection configs for common entity types
-- These can be customized per organization
COMMENT ON TABLE duplicate_detection_config IS 'Organization-specific duplicate detection settings';

-- =====================================================
-- ANALYTICS
-- =====================================================

-- Get duplicate detection statistics
CREATE OR REPLACE FUNCTION get_duplicate_stats(
    p_organization_id uuid,
    p_days_back integer DEFAULT 30
)
RETURNS TABLE (
    entity_type text,
    total_detected integer,
    pending integer,
    merged integer,
    ignored integer,
    confirmed_different integer
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        dc.entity_type,
        COUNT(*)::integer as total_detected,
        COUNT(*) FILTER (WHERE dc.status = 'pending')::integer,
        COUNT(*) FILTER (WHERE dc.status = 'merged')::integer,
        COUNT(*) FILTER (WHERE dc.status = 'ignored')::integer,
        COUNT(*) FILTER (WHERE dc.status = 'confirmed_different')::integer
    FROM duplicate_candidates dc
    WHERE dc.organization_id = p_organization_id
        AND dc.created_at > now() - (p_days_back || ' days')::interval
    GROUP BY dc.entity_type;
END;
$$;

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_duplicate_candidates_entity_status
ON duplicate_candidates(organization_id, entity_type, status);

CREATE INDEX idx_duplicate_candidates_records
ON duplicate_candidates(record_id_1, record_id_2);
