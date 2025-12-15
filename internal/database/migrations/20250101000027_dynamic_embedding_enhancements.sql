-- Migration: Dynamic Embedding Enhancements & ERP Seeds
-- Description: Provider-aware embedding helpers, dynamic search wrappers, and Odoo-inspired knowledge base seeds
-- Created: 2025-01-01

-- =====================================================
-- PROVIDER-AWARE EMBEDDING HELPERS
-- =====================================================

CREATE OR REPLACE FUNCTION generate_dynamic_embedding(
    p_organization_id uuid,
    p_text text
)
RETURNS vector
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_provider RECORD;
BEGIN
    IF p_text IS NULL OR length(trim(p_text)) = 0 THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_provider
    FROM get_active_embedding_provider(p_organization_id);

    IF v_provider.provider_type IS NULL THEN
        RETURN NULL;
    END IF;

    IF v_provider.provider_type = 'ollama' THEN
        IF v_provider.dimensions = 768 THEN
            RETURN generate_embedding_ollama_768(
                p_text,
                COALESCE(v_provider.model_name, 'nomic-embed-text')
            );
        ELSIF v_provider.dimensions = 384 THEN
            RETURN generate_embedding_ollama_384(
                p_text,
                COALESCE(v_provider.model_name, 'all-minilm')
            );
        ELSE
            RAISE NOTICE 'generate_dynamic_embedding: Unsupported Ollama dimension %', v_provider.dimensions;
            RETURN NULL;
        END IF;
    ELSE
        RAISE NOTICE 'generate_dynamic_embedding: Unsupported provider type %', v_provider.provider_type;
        RETURN NULL;
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION map_embedding_to_columns(
    p_dimensions integer,
    p_vector vector
)
RETURNS TABLE (
    search_embedding vector(768),
    search_embedding_384 vector(384),
    search_embedding_512 vector(512),
    search_embedding_768 vector(768),
    search_embedding_1024 vector(1024),
    search_embedding_1536 vector(1536)
)
LANGUAGE plpgsql
AS $$
BEGIN
    search_embedding := NULL;
    search_embedding_384 := NULL;
    search_embedding_512 := NULL;
    search_embedding_768 := NULL;
    search_embedding_1024 := NULL;
    search_embedding_1536 := NULL;

    IF p_vector IS NOT NULL THEN
        IF p_dimensions = 768 THEN
            search_embedding := p_vector::vector(768);
            search_embedding_768 := p_vector::vector(768);
        ELSIF p_dimensions = 384 THEN
            search_embedding_384 := p_vector::vector(384);
        ELSIF p_dimensions = 512 THEN
            search_embedding_512 := p_vector::vector(512);
        ELSIF p_dimensions = 1024 THEN
            search_embedding_1024 := p_vector::vector(1024);
        ELSIF p_dimensions = 1536 THEN
            search_embedding_1536 := p_vector::vector(1536);
        ELSE
            RAISE NOTICE 'map_embedding_to_columns: Unsupported dimension %', p_dimensions;
        END IF;
    END IF;

    RETURN NEXT;
END;
$$;

CREATE OR REPLACE FUNCTION vector_to_jsonb(p_vector vector)
RETURNS jsonb
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_text text;
    v_array text[];
    v_float_array float8[];
BEGIN
    -- Convert vector to text format: [0.1, 0.2, 0.3]
    v_text := p_vector::text;

    -- Remove brackets and split by comma
    v_text := trim(both '[]' from v_text);
    v_array := string_to_array(v_text, ',');

    -- Convert to float8 array
    v_float_array := ARRAY(
        SELECT trim(elem)::float8
        FROM unnest(v_array) AS elem
    );

    -- Convert to JSONB
    RETURN to_jsonb(v_float_array);
END;
$$;

CREATE OR REPLACE FUNCTION weighted_combine_vectors(
    p_dimensions integer,
    p_primary vector,
    p_primary_weight double precision,
    p_secondary vector,
    p_secondary_weight double precision
)
RETURNS vector
LANGUAGE plpgsql
AS $$
DECLARE
    v_primary float8[];
    v_secondary float8[];
    v_combined float8[];
BEGIN
    IF p_primary IS NULL AND p_secondary IS NULL THEN
        RETURN NULL;
    ELSIF p_primary IS NULL THEN
        RETURN p_secondary;
    ELSIF p_secondary IS NULL THEN
        RETURN p_primary;
    END IF;

    v_primary := p_primary::float8[];
    v_secondary := p_secondary::float8[];

    IF array_length(v_primary, 1) IS NULL THEN
        v_primary := ARRAY_FILL(0::float8, ARRAY[p_dimensions]);
    END IF;

    IF array_length(v_secondary, 1) IS NULL THEN
        v_secondary := ARRAY_FILL(0::float8, ARRAY[p_dimensions]);
    END IF;

    v_combined := ARRAY(
        SELECT
            (COALESCE(v_primary[i], 0)::double precision * p_primary_weight) +
            (COALESCE(v_secondary[i], 0)::double precision * p_secondary_weight)
        FROM generate_series(1, p_dimensions) AS i
    );

    RETURN v_combined::vector(p_dimensions);
END;
$$;

-- =====================================================
-- UPDATE EMBEDDING TRIGGERS TO USE PROVIDER HELPERS
-- =====================================================

CREATE OR REPLACE FUNCTION update_contact_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.email, '') || ' ' ||
        COALESCE(NEW.phone, '') || ' ' ||
        COALESCE(NEW.comment, '') || ' ' ||
        COALESCE(NEW.job_position, '') || ' ' ||
        COALESCE(NEW.reference, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_product_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.default_code, '') || ' ' ||
        COALESCE(NEW.barcode, '') || ' ' ||
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(NEW.description_sale, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_sales_order_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.client_order_ref, '') || ' ' ||
        COALESCE(NEW.note, '') || ' ' ||
        COALESCE(NEW.origin, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_invoice_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.ref, '') || ' ' ||
        COALESCE(NEW.narration, '') || ' ' ||
        COALESCE(NEW.invoice_origin, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_payment_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.ref, '') || ' ' ||
        COALESCE(NEW.communication, '') || ' ' ||
        COALESCE(NEW.payment_type, '') || ' ' ||
        COALESCE(NEW.state, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_purchase_order_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.partner_ref, '') || ' ' ||
        COALESCE(NEW.notes, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.state, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_stock_picking_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.note, '') || ' ' ||
        COALESCE(NEW.state, '') || ' ' ||
        COALESCE(NEW.sequence_code, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_stock_move_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.reference, '') || ' ' ||
        COALESCE(NEW.note, '') || ' ' ||
        COALESCE(NEW.state, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_manufacturing_order_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.state, '') || ' ' ||
        COALESCE(NEW.metadata::text, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_work_order_embedding()
RETURNS TRIGGER AS $$
DECLARE
    v_embedding vector;
    v_dimensions integer;
BEGIN
    v_embedding := generate_dynamic_embedding(
        NEW.organization_id,
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.state, '') || ' ' ||
        COALESCE(NEW.production_id::text, '') || ' ' ||
        COALESCE(NEW.workcenter_id::text, '') || ' ' ||
        COALESCE(NEW.date_planned_start::text, '') || ' ' ||
        COALESCE(NEW.date_planned_finished::text, '')
    );

    IF v_embedding IS NULL THEN
        RETURN NEW;
    END IF;

    v_dimensions := vector_dims(v_embedding);

    SELECT
        mapped.search_embedding,
        mapped.search_embedding_384,
        mapped.search_embedding_512,
        mapped.search_embedding_768,
        mapped.search_embedding_1024,
        mapped.search_embedding_1536
    INTO
        NEW.search_embedding,
        NEW.search_embedding_384,
        NEW.search_embedding_512,
        NEW.search_embedding_768,
        NEW.search_embedding_1024,
        NEW.search_embedding_1536
    FROM map_embedding_to_columns(v_dimensions, v_embedding) mapped;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- DYNAMIC SEARCH WRAPPERS
-- =====================================================

CREATE OR REPLACE FUNCTION search_by_text_dynamic(
    p_organization_id uuid,
    p_query_text text,
    p_limit integer DEFAULT 20
)
RETURNS TABLE (
    record_type text,
    record_id uuid,
    record_name text,
    record_reference text,
    record_description text,
    similarity_score float,
    metadata jsonb,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_embedding vector;
    v_embedding_json jsonb;
BEGIN
    v_embedding := generate_dynamic_embedding(p_organization_id, p_query_text);

    IF v_embedding IS NULL THEN
        RETURN;
    END IF;

    v_embedding_json := vector_to_jsonb(v_embedding);

    RETURN QUERY
    SELECT *
    FROM global_semantic_search_dynamic(
        p_organization_id,
        v_embedding_json,
        p_limit
    );
END;
$$;

CREATE OR REPLACE FUNCTION search_by_text(
    p_organization_id uuid,
    p_query_text text,
    p_limit integer DEFAULT 20
)
RETURNS TABLE (
    record_type text,
    record_id uuid,
    record_name text,
    record_reference text,
    similarity_score float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        result.record_type,
        result.record_id,
        result.record_name,
        result.record_reference,
        result.similarity_score
    FROM search_by_text_dynamic(p_organization_id, p_query_text, p_limit) result;
END;
$$;

-- =====================================================
-- UPDATED CONTEXTUAL HELP (REAL EMBEDDINGS)
-- =====================================================

CREATE OR REPLACE FUNCTION get_contextual_help_dynamic(
    p_organization_id uuid,
    p_screen_key text,
    p_user_query text DEFAULT NULL,
    p_user_id uuid DEFAULT NULL,
    p_limit integer DEFAULT 5
)
RETURNS TABLE (
    entry_id uuid,
    title text,
    summary text,
    body_markdown text,
    relevance_score float,
    space_name text,
    tags text[]
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_provider RECORD;
    v_embedding_column text;
    v_screen_info RECORD;
    v_screen_context text;
    v_screen_embedding vector;
    v_query_embedding vector;
    v_final_embedding vector;
    v_query text;
BEGIN
    SELECT * INTO v_provider
    FROM get_active_embedding_provider(p_organization_id);

    IF v_provider.provider_type IS NULL THEN
        RETURN;
    END IF;

    v_embedding_column := get_embedding_column_name(v_provider.dimensions);

    SELECT * INTO v_screen_info
    FROM app_screens
    WHERE screen_key = p_screen_key;

    IF NOT FOUND THEN
        IF p_user_query IS NULL OR length(trim(p_user_query)) = 0 THEN
            RAISE EXCEPTION 'Screen key % not found and no query provided', p_screen_key;
        END IF;

        v_query_embedding := generate_dynamic_embedding(p_organization_id, p_user_query);

        IF v_query_embedding IS NULL THEN
            RETURN;
        END IF;

        v_query := format($sql$
            SELECT
                ke.id,
                ke.title,
                ke.summary,
                ke.body_markdown,
                1 - (ke.%I <=> $1::vector(%s)) AS relevance_score,
                ks.name AS space_name,
                ARRAY(
                    SELECT kt.name
                    FROM knowledge_entry_tags ket
                    JOIN knowledge_tags kt ON kt.id = ket.tag_id
                    WHERE ket.entry_id = ke.id
                ) AS tags
            FROM knowledge_entries ke
            JOIN knowledge_spaces ks ON ks.id = ke.space_id
            WHERE ke.organization_id = $2
                AND ke.is_published = true
                AND ke.%I IS NOT NULL
            ORDER BY relevance_score DESC
            LIMIT $3
        $sql$, v_embedding_column, v_provider.dimensions, v_embedding_column);

        RETURN QUERY EXECUTE v_query
            USING v_query_embedding::vector, p_organization_id, p_limit;

        IF p_user_id IS NOT NULL THEN
            INSERT INTO contextual_help_usage (
                organization_id,
                user_id,
                screen_context,
                query_text
            ) VALUES (
                p_organization_id,
                p_user_id,
                p_screen_key,
                p_user_query
            );
        END IF;

        RETURN;
    END IF;

    v_screen_context := COALESCE(v_screen_info.screen_name, '') || ' ' ||
                        COALESCE(v_screen_info.description, '') || ' ' ||
                        COALESCE(v_screen_info.module, '');

    v_screen_embedding := generate_dynamic_embedding(p_organization_id, v_screen_context);

    IF p_user_query IS NOT NULL AND length(trim(p_user_query)) > 0 THEN
        v_query_embedding := generate_dynamic_embedding(p_organization_id, p_user_query);
    END IF;

    IF v_query_embedding IS NOT NULL AND v_screen_embedding IS NOT NULL THEN
        v_final_embedding := weighted_combine_vectors(
            v_provider.dimensions,
            v_query_embedding,
            0.7,
            v_screen_embedding,
            0.3
        );
    ELSE
        v_final_embedding := COALESCE(v_query_embedding, v_screen_embedding);
    END IF;

    IF v_final_embedding IS NULL THEN
        RETURN;
    END IF;

    v_query := format($sql$
        SELECT
            ke.id,
            ke.title,
            ke.summary,
            ke.body_markdown,
            1 - (ke.%I <=> $1::vector(%s)) AS relevance_score,
            ks.name AS space_name,
            ARRAY(
                SELECT kt.name
                FROM knowledge_entry_tags ket
                JOIN knowledge_tags kt ON kt.id = ket.tag_id
                WHERE ket.entry_id = ke.id
            ) AS tags
        FROM knowledge_entries ke
        JOIN knowledge_spaces ks ON ks.id = ke.space_id
        WHERE ke.organization_id = $2
            AND ke.is_published = true
            AND ke.%I IS NOT NULL
        ORDER BY relevance_score DESC
        LIMIT $3
    $sql$, v_embedding_column, v_provider.dimensions, v_embedding_column);

    RETURN QUERY EXECUTE v_query
        USING v_final_embedding::vector, p_organization_id, p_limit;

    IF p_user_id IS NOT NULL THEN
        INSERT INTO contextual_help_usage (
            organization_id,
            user_id,
            screen_context,
            query_text
        ) VALUES (
            p_organization_id,
            p_user_id,
            p_screen_key,
            p_user_query
        );
    END IF;
END;
$$;

-- =====================================================
-- SEED DEFAULT PROVIDER & ERP KNOWLEDGE CONTENT
-- =====================================================

INSERT INTO embedding_providers (
    organization_id,
    provider_type,
    model_name,
    dimensions,
    api_endpoint,
    metadata
)
SELECT
    org.id,
    'ollama',
    'nomic-embed-text',
    768,
    'http://ollama:11434/api/embeddings',
    jsonb_build_object('source', 'dynamic_embedding_seed')
FROM organizations org
WHERE NOT EXISTS (
    SELECT 1
    FROM embedding_providers ep
    WHERE ep.organization_id = org.id
      AND ep.is_active = true
);

WITH org_primary_user AS (
    SELECT DISTINCT ON (ou.organization_id)
        ou.organization_id,
        ou.id AS organization_user_id
    FROM organization_users ou
    WHERE ou.is_active = true
    ORDER BY ou.organization_id,
             CASE WHEN ou.role IN ('owner', 'admin') THEN 0 ELSE 1 END,
             ou.joined_at
),
inserted_spaces AS (
    INSERT INTO knowledge_spaces (
        organization_id,
        name,
        description,
        visibility,
        created_by
    )
    SELECT
        opu.organization_id,
        'Sales Operations Playbook',
        'Migration handbook for Odoo-style sales workflows',
        'internal',
        opu.organization_user_id
    FROM org_primary_user opu
    WHERE NOT EXISTS (
        SELECT 1
        FROM knowledge_spaces ks
        WHERE ks.organization_id = opu.organization_id
          AND ks.name = 'Sales Operations Playbook'
    )
    RETURNING id, organization_id, created_by
),
target_spaces AS (
    SELECT id, organization_id, created_by
    FROM inserted_spaces
    UNION
    SELECT ks.id, ks.organization_id, ks.created_by
    FROM knowledge_spaces ks
    WHERE ks.name = 'Sales Operations Playbook'
)
INSERT INTO knowledge_entries (
    organization_id,
    space_id,
    title,
    body_markdown,
    summary,
    owner_id,
    metadata,
    is_published
)
SELECT
    ts.organization_id,
    ts.id,
    'How to add a bundle with delivery charges',
    $$
    ## Add Packaged Products with Shipping Fees

    1. Open the Sales Order and click **Add Product**.
    2. Choose the predefined bundle (`Server Rack Deployment`).
    3. Switch to the **Extras** tab and select **Add Delivery**.
    4. Confirm to auto-create the delivery picking, similar to Odoo's "Validate Delivery" step.

    > Tip: Use the **Smart Buttons** on the top-right to jump to the delivery or invoice just like in Odoo.
    $$,
    'Guide for sales reps migrating from Odoo quotations',
    ts.created_by,
    jsonb_build_object(
        'module', 'sales',
        'screens', ARRAY['sales_order_create'],
        'tags', ARRAY['bundles', 'delivery', 'odoo-migration']
    ),
    true
FROM target_spaces ts
WHERE NOT EXISTS (
    SELECT 1
    FROM knowledge_entries ke
    WHERE ke.space_id = ts.id
      AND ke.title = 'How to add a bundle with delivery charges'
);

WITH target_spaces AS (
    SELECT ks.id, ks.organization_id, ks.created_by
    FROM knowledge_spaces ks
    WHERE ks.name = 'Sales Operations Playbook'
)
INSERT INTO knowledge_entries (
    organization_id,
    space_id,
    title,
    body_markdown,
    summary,
    owner_id,
    metadata,
    is_published
)
SELECT
    ts.organization_id,
    ts.id,
    'Configure fiscal positions during customer onboarding',
    $$
    ## Fiscal Position Checklist

    - Enable **Auto-detect fiscal position** in the customer form (Settings tab).
    - Map state-based taxes via **Accounting -> Configuration -> Fiscal Positions**.
    - When creating a Sales Order, the system applies the fiscal position just like Odoo's tax engine.
    $$,
    'Mirror Odoo tax auto-mapping with contextual help',
    ts.created_by,
    jsonb_build_object(
        'module', 'sales',
        'screens', ARRAY['customer_create', 'sales_order_create'],
        'tags', ARRAY['fiscal-position', 'taxes', 'odoo-migration']
    ),
    true
FROM target_spaces ts
WHERE NOT EXISTS (
    SELECT 1
    FROM knowledge_entries ke
    WHERE ke.space_id = ts.id
      AND ke.title = 'Configure fiscal positions during customer onboarding'
);

INSERT INTO duplicate_detection_config (
    organization_id,
    entity_type,
    similarity_threshold,
    fields_to_compare,
    auto_merge
)
SELECT
    org.id,
    'contact',
    0.85,
    ARRAY['name', 'email', 'phone'],
    false
FROM organizations org
WHERE NOT EXISTS (
    SELECT 1
    FROM duplicate_detection_config ddc
    WHERE ddc.organization_id = org.id
      AND ddc.entity_type = 'contact'
);

INSERT INTO duplicate_detection_config (
    organization_id,
    entity_type,
    similarity_threshold,
    fields_to_compare,
    auto_merge
)
SELECT
    org.id,
    'product',
    0.90,
    ARRAY['name', 'default_code', 'barcode'],
    false
FROM organizations org
WHERE NOT EXISTS (
    SELECT 1
    FROM duplicate_detection_config ddc
    WHERE ddc.organization_id = org.id
      AND ddc.entity_type = 'product'
);
