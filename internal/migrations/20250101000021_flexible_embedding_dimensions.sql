-- Migration: Flexible Embedding Dimensions
-- Description: Support multiple embedding providers with different vector dimensions
-- Created: 2025-01-01

-- =====================================================
-- EMBEDDING PROVIDER CONFIGURATION
-- =====================================================

-- Track which embedding provider/model each organization uses
CREATE TABLE embedding_providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider_type text NOT NULL, -- 'openai', 'ollama', 'local', 'transformers'
    model_name text NOT NULL,
    dimensions integer NOT NULL,
    api_endpoint text, -- For Ollama/custom endpoints
    api_key_encrypted text, -- For OpenAI/commercial APIs
    is_active boolean DEFAULT true,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT embedding_providers_type_check
        CHECK (provider_type IN ('openai', 'ollama', 'local', 'transformers', 'cohere', 'custom')),
    CONSTRAINT embedding_providers_dimensions_check
        CHECK (dimensions IN (384, 512, 768, 1024, 1536, 3072))
);

-- Only one active provider per organization
CREATE UNIQUE INDEX idx_embedding_providers_active
ON embedding_providers(organization_id)
WHERE is_active = true;

CREATE INDEX idx_embedding_providers_org ON embedding_providers(organization_id);

COMMENT ON TABLE embedding_providers IS 'Track embedding provider and model per organization for multi-provider support';

-- =====================================================
-- ADD MULTI-DIMENSIONAL EMBEDDING COLUMNS
-- =====================================================

-- Contacts - add multiple embedding columns
ALTER TABLE contacts
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Products
ALTER TABLE products
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Sales Orders
ALTER TABLE sales_orders
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Invoices
ALTER TABLE invoices
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Knowledge Entries - add new dimensions (768 already exists from earlier migration)
ALTER TABLE knowledge_entries
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Payments
ALTER TABLE payments
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Purchase Orders
ALTER TABLE purchase_orders
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Stock Pickings
ALTER TABLE stock_pickings
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Stock Moves
ALTER TABLE stock_moves
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Manufacturing Orders
ALTER TABLE manufacturing_orders
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- Work Orders
ALTER TABLE work_orders
ADD COLUMN IF NOT EXISTS search_embedding_384 vector(384),
ADD COLUMN IF NOT EXISTS search_embedding_512 vector(512),
ADD COLUMN IF NOT EXISTS search_embedding_768 vector(768),
ADD COLUMN IF NOT EXISTS search_embedding_1024 vector(1024),
ADD COLUMN IF NOT EXISTS search_embedding_1536 vector(1536);

-- =====================================================
-- CREATE INDEXES FOR ALL DIMENSIONS
-- =====================================================

-- Contacts
CREATE INDEX IF NOT EXISTS idx_contacts_embedding_384
ON contacts USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_contacts_embedding_512
ON contacts USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_contacts_embedding_768
ON contacts USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_contacts_embedding_1024
ON contacts USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_contacts_embedding_1536
ON contacts USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Products
CREATE INDEX IF NOT EXISTS idx_products_embedding_384
ON products USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_products_embedding_512
ON products USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_products_embedding_768
ON products USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_products_embedding_1024
ON products USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_products_embedding_1536
ON products USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Sales Orders
CREATE INDEX IF NOT EXISTS idx_sales_orders_embedding_384
ON sales_orders USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_sales_orders_embedding_512
ON sales_orders USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_sales_orders_embedding_768
ON sales_orders USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_sales_orders_embedding_1024
ON sales_orders USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_sales_orders_embedding_1536
ON sales_orders USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Invoices
CREATE INDEX IF NOT EXISTS idx_invoices_embedding_384
ON invoices USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_invoices_embedding_512
ON invoices USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_invoices_embedding_768
ON invoices USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_invoices_embedding_1024
ON invoices USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_invoices_embedding_1536
ON invoices USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Knowledge Entries
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_embedding_384
ON knowledge_entries USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_embedding_512
ON knowledge_entries USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_embedding_768
ON knowledge_entries USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_embedding_1024
ON knowledge_entries USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_embedding_1536
ON knowledge_entries USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Payments
CREATE INDEX IF NOT EXISTS idx_payments_embedding_384
ON payments USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_payments_embedding_512
ON payments USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_payments_embedding_768
ON payments USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_payments_embedding_1024
ON payments USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_payments_embedding_1536
ON payments USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Purchase Orders
CREATE INDEX IF NOT EXISTS idx_purchase_orders_embedding_384
ON purchase_orders USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_embedding_512
ON purchase_orders USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_embedding_768
ON purchase_orders USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_embedding_1024
ON purchase_orders USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_embedding_1536
ON purchase_orders USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Stock Pickings
CREATE INDEX IF NOT EXISTS idx_stock_pickings_embedding_384
ON stock_pickings USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_pickings_embedding_512
ON stock_pickings USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_pickings_embedding_768
ON stock_pickings USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_pickings_embedding_1024
ON stock_pickings USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_pickings_embedding_1536
ON stock_pickings USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Stock Moves
CREATE INDEX IF NOT EXISTS idx_stock_moves_embedding_384
ON stock_moves USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_moves_embedding_512
ON stock_moves USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_moves_embedding_768
ON stock_moves USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_moves_embedding_1024
ON stock_moves USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_stock_moves_embedding_1536
ON stock_moves USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Manufacturing Orders
CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_embedding_384
ON manufacturing_orders USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_embedding_512
ON manufacturing_orders USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_embedding_768
ON manufacturing_orders USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_embedding_1024
ON manufacturing_orders USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_embedding_1536
ON manufacturing_orders USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- Work Orders
CREATE INDEX IF NOT EXISTS idx_work_orders_embedding_384
ON work_orders USING ivfflat (search_embedding_384 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_work_orders_embedding_512
ON work_orders USING ivfflat (search_embedding_512 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_work_orders_embedding_768
ON work_orders USING ivfflat (search_embedding_768 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_work_orders_embedding_1024
ON work_orders USING ivfflat (search_embedding_1024 vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_work_orders_embedding_1536
ON work_orders USING ivfflat (search_embedding_1536 vector_cosine_ops) WITH (lists = 100);

-- =====================================================
-- HELPER FUNCTIONS
-- =====================================================

-- Get active provider for organization
CREATE OR REPLACE FUNCTION get_active_embedding_provider(p_organization_id uuid)
RETURNS TABLE (
    provider_type text,
    model_name text,
    dimensions integer,
    api_endpoint text
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ep.provider_type,
        ep.model_name,
        ep.dimensions,
        ep.api_endpoint
    FROM embedding_providers ep
    WHERE ep.organization_id = p_organization_id
        AND ep.is_active = true
    LIMIT 1;
END;
$$;

-- Get embedding column name based on dimensions
CREATE OR REPLACE FUNCTION get_embedding_column_name(p_dimensions integer)
RETURNS text
LANGUAGE plpgsql
IMMUTABLE
AS $$
BEGIN
    RETURN 'search_embedding_' || p_dimensions::text;
END;
$$;

COMMENT ON FUNCTION get_active_embedding_provider IS 'Get active embedding provider configuration for organization';
COMMENT ON FUNCTION get_embedding_column_name IS 'Get column name for specific embedding dimension';

-- =====================================================
-- DYNAMIC GLOBAL SEARCH (DIMENSION-AGNOSTIC)
-- =====================================================

CREATE OR REPLACE FUNCTION global_semantic_search_dynamic(
    p_organization_id uuid,
    p_query_embedding jsonb, -- Pass as JSONB array to handle any dimension
    p_limit integer DEFAULT 20,
    p_include_contacts boolean DEFAULT true,
    p_include_products boolean DEFAULT true,
    p_include_sales_orders boolean DEFAULT true,
    p_include_invoices boolean DEFAULT true,
    p_include_payments boolean DEFAULT true,
    p_include_purchase_orders boolean DEFAULT true,
    p_include_stock_pickings boolean DEFAULT true,
    p_include_stock_moves boolean DEFAULT true,
    p_include_manufacturing_orders boolean DEFAULT true,
    p_include_work_orders boolean DEFAULT true
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
STABLE
AS $$
DECLARE
    v_provider record;
    v_embedding_column text;
    v_query text;
    v_vector_string text;
BEGIN
    -- Get active provider
    SELECT * INTO v_provider
    FROM get_active_embedding_provider(p_organization_id);

    IF NOT FOUND THEN
        RAISE EXCEPTION 'No active embedding provider configured for organization %', p_organization_id;
    END IF;

    -- Get column name
    v_embedding_column := get_embedding_column_name(v_provider.dimensions);

    -- Convert JSONB array to vector string
    v_vector_string := '[' || (
        SELECT string_agg(value::text, ',')
        FROM jsonb_array_elements_text(p_query_embedding)
    ) || ']';

    -- Build dynamic query with proper dimension
    v_query := format($sql$
        WITH ranked_results AS (
            -- Contacts
            SELECT
                'contact'::text as record_type,
                c.id as record_id,
                c.name as record_name,
                c.reference as record_reference,
                COALESCE(c.comment, '') as record_description,
                1 - (c.%I <=> $1::vector(%s)) as similarity_score,
                jsonb_build_object(
                    'email', c.email,
                    'phone', c.phone,
                    'is_customer', c.is_customer,
                    'is_vendor', c.is_vendor,
                    'city', c.city
                ) as metadata,
                c.created_at,
                c.updated_at
            FROM contacts c
            WHERE c.organization_id = $2
                AND c.deleted_at IS NULL
                AND c.%I IS NOT NULL
                AND $3 = true

            UNION ALL

            -- Products
            SELECT
                'product'::text,
                p.id,
                p.name,
                p.default_code,
                COALESCE(p.description, ''),
                1 - (p.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'barcode', p.barcode,
                    'list_price', p.list_price,
                    'product_type', p.product_type
                ),
                p.created_at,
                p.updated_at
            FROM products p
            WHERE p.organization_id = $2
                AND p.deleted_at IS NULL
                AND p.%I IS NOT NULL
                AND $4 = true

            UNION ALL

            -- Sales Orders
            SELECT
                'sales_order'::text,
                so.id,
                so.name,
                so.client_order_ref,
                COALESCE(so.note, ''),
                1 - (so.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'partner_id', so.partner_id,
                    'state', so.state,
                    'amount_total', so.amount_total,
                    'date_order', so.date_order
                ),
                so.created_at,
                so.updated_at
            FROM sales_orders so
            WHERE so.organization_id = $2
                AND so.deleted_at IS NULL
                AND so.%I IS NOT NULL
                AND $5 = true

            UNION ALL

        -- Invoices
            SELECT
                'invoice'::text,
                inv.id,
                inv.name,
                inv.ref,
                COALESCE(inv.narration, ''),
                1 - (inv.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'partner_id', inv.partner_id,
                    'state', inv.state,
                    'move_type', inv.move_type,
                    'amount_total', inv.amount_total
                ),
                inv.created_at,
                inv.updated_at
            FROM invoices inv
            WHERE inv.organization_id = $2
                AND inv.%I IS NOT NULL
                AND $6 = true

            UNION ALL

            -- Payments
            SELECT
                'payment'::text,
                pay.id,
                COALESCE(pay.name, pay.ref),
                pay.communication,
                COALESCE(pay.ref, ''),
                1 - (pay.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'partner_id', pay.partner_id,
                    'payment_type', pay.payment_type,
                    'state', pay.state,
                    'amount', pay.amount,
                    'payment_date', pay.payment_date,
                    'journal_id', pay.journal_id
                ),
                pay.created_at,
                pay.updated_at
            FROM payments pay
            WHERE pay.organization_id = $2
                AND pay.%I IS NOT NULL
                AND $7 = true

            UNION ALL

            -- Purchase Orders
            SELECT
                'purchase_order'::text,
                po.id,
                po.name,
                po.partner_ref,
                COALESCE(po.notes, ''),
                1 - (po.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'partner_id', po.partner_id,
                    'state', po.state,
                    'amount_total', po.amount_total,
                    'date_order', po.date_order,
                    'invoice_status', po.invoice_status
                ),
                po.created_at,
                po.updated_at
            FROM purchase_orders po
            WHERE po.organization_id = $2
                AND po.deleted_at IS NULL
                AND po.%I IS NOT NULL
                AND $8 = true

            UNION ALL

            -- Stock Pickings
            SELECT
                'stock_picking'::text,
                sp.id,
                sp.name,
                sp.origin,
                COALESCE(sp.note, ''),
                1 - (sp.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'state', sp.state,
                    'picking_type_id', sp.picking_type_id,
                    'scheduled_date', sp.scheduled_date,
                    'location_id', sp.location_id,
                    'location_dest_id', sp.location_dest_id,
                    'partner_id', sp.partner_id
                ),
                sp.created_at,
                sp.updated_at
            FROM stock_pickings sp
            WHERE sp.organization_id = $2
                AND sp.deleted_at IS NULL
                AND sp.%I IS NOT NULL
                AND $9 = true

            UNION ALL

            -- Stock Moves
            SELECT
                'stock_move'::text,
                sm.id,
                sm.name,
                sm.reference,
                COALESCE(sm.note, ''),
                1 - (sm.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'product_id', sm.product_id,
                    'state', sm.state,
                    'product_uom_qty', sm.product_uom_qty,
                    'location_id', sm.location_id,
                    'location_dest_id', sm.location_dest_id,
                    'picking_id', sm.picking_id
                ),
                sm.created_at,
                sm.updated_at
            FROM stock_moves sm
            WHERE sm.organization_id = $2
                AND sm.deleted_at IS NULL
                AND sm.%I IS NOT NULL
                AND $10 = true

            UNION ALL

            -- Manufacturing Orders
            SELECT
                'manufacturing_order'::text,
                mo.id,
                mo.name,
                mo.origin,
                COALESCE(mo.metadata->>'routing_notes', '') || ' ' ||
                COALESCE(mo.state, ''),
                1 - (mo.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'state', mo.state,
                    'product_id', mo.product_id,
                    'qty', mo.product_qty,
                    'date_planned_start', mo.date_planned_start,
                    'date_planned_finished', mo.date_planned_finished,
                    'workcenter_ids', mo.metadata->'workcenter_ids'
                ),
                mo.created_at,
                mo.updated_at
            FROM manufacturing_orders mo
            WHERE mo.organization_id = $2
                AND mo.deleted_at IS NULL
                AND mo.%I IS NOT NULL
                AND $11 = true

            UNION ALL

            -- Work Orders
            SELECT
                'work_order'::text,
                wo.id,
                wo.name,
                wo.production_id::text,
                COALESCE(wo.state, '') || ' ' ||
                COALESCE(wo.date_planned_start::text, '') || ' ' ||
                COALESCE(wo.date_planned_finished::text, ''),
                1 - (wo.%I <=> $1::vector(%s)),
                jsonb_build_object(
                    'state', wo.state,
                    'workcenter_id', wo.workcenter_id,
                    'production_id', wo.production_id,
                    'date_planned_start', wo.date_planned_start,
                    'date_planned_finished', wo.date_planned_finished
                ),
                wo.created_at,
                wo.updated_at
            FROM work_orders wo
            WHERE wo.organization_id = $2
                AND wo.%I IS NOT NULL
                AND $12 = true
        )
        SELECT * FROM ranked_results
        WHERE similarity_score > 0.3
        ORDER BY similarity_score DESC
        LIMIT $13
    $sql$,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column,
        v_embedding_column, v_provider.dimensions, v_embedding_column
    );

    -- Execute dynamic query
    RETURN QUERY EXECUTE v_query
        USING v_vector_string::vector, p_organization_id,
              p_include_contacts, p_include_products,
              p_include_sales_orders, p_include_invoices,
              p_include_payments, p_include_purchase_orders,
              p_include_stock_pickings, p_include_stock_moves,
              p_include_manufacturing_orders, p_include_work_orders,
              p_limit;
END;
$$;

COMMENT ON FUNCTION global_semantic_search_dynamic IS 'Dimension-agnostic semantic search using active provider settings';

-- =====================================================
-- UPDATED CONTEXTUAL HELP (DIMENSION-AGNOSTIC)
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
STABLE
AS $$
DECLARE
    v_provider record;
    v_embedding_column text;
    v_screen_info record;
    v_query text;
    v_combined_embedding_json jsonb;
BEGIN
    -- Get provider
    SELECT * INTO v_provider
    FROM get_active_embedding_provider(p_organization_id);

    IF NOT FOUND THEN
        RAISE EXCEPTION 'No active embedding provider for organization %', p_organization_id;
    END IF;

    v_embedding_column := get_embedding_column_name(v_provider.dimensions);

    -- Get screen info
    SELECT * INTO v_screen_info
    FROM app_screens
    WHERE screen_key = p_screen_key;

    -- For now, just search by screen context
    -- In production, you'd generate embeddings for screen + query

    v_query := format($sql$
        SELECT
            ke.id,
            ke.title,
            ke.summary,
            ke.body_markdown,
            1 - (ke.%I <=> $1::vector(%s)) as relevance_score,
            ks.name as space_name,
            ARRAY(
                SELECT kt.name
                FROM knowledge_entry_tags ket
                JOIN knowledge_tags kt ON kt.id = ket.tag_id
                WHERE ket.entry_id = ke.id
            ) as tags
        FROM knowledge_entries ke
        JOIN knowledge_spaces ks ON ks.id = ke.space_id
        WHERE ke.organization_id = $2
            AND ke.is_published = true
            AND ke.%I IS NOT NULL
        ORDER BY relevance_score DESC
        LIMIT $3
    $sql$, v_embedding_column, v_provider.dimensions, v_embedding_column);

    -- This is a placeholder - in production you'd generate embedding for screen + query
    RETURN QUERY EXECUTE v_query
        USING array_fill(0, ARRAY[v_provider.dimensions])::vector,
              p_organization_id, p_limit;

    -- Log usage
    IF p_user_id IS NOT NULL THEN
        INSERT INTO contextual_help_usage (
            organization_id, user_id, screen_context, query_text
        ) VALUES (p_organization_id, p_user_id, p_screen_key, p_user_query);
    END IF;
END;
$$;

-- =====================================================
-- UPDATED DUPLICATE DETECTION (DIMENSION-AGNOSTIC)
-- =====================================================

CREATE OR REPLACE FUNCTION find_duplicate_contacts_dynamic(
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
    created_at timestamptz
)
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_provider record;
    v_embedding_column text;
    v_query text;
BEGIN
    -- Get provider
    SELECT * INTO v_provider
    FROM get_active_embedding_provider(p_organization_id);

    IF NOT FOUND THEN
        -- Fall back to non-vector matching if no provider
        v_embedding_column := NULL;
    ELSE
        v_embedding_column := get_embedding_column_name(v_provider.dimensions);
    END IF;

    -- Use traditional matching (email, phone, name similarity)
    -- Vector similarity is bonus if available
    RETURN QUERY
    WITH similarity_matches AS (
        SELECT
            c.id,
            c.name,
            c.email,
            c.phone,
            c.created_at,
            -- Exact matches
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
            -- Name similarity
            CASE
                WHEN p_name IS NOT NULL THEN
                    similarity(LOWER(c.name), LOWER(p_name))
                ELSE 0.0
            END as name_similarity
        FROM contacts c
        WHERE c.organization_id = p_organization_id
            AND c.deleted_at IS NULL
            AND (p_contact_id IS NULL OR c.id != p_contact_id)
    ),
    scored_matches AS (
        SELECT
            *,
            -- Weighted score: email (40%), phone (30%), name (30%)
            (email_match * 0.4 + phone_match * 0.3 + name_similarity * 0.3) as combined_score,
            CASE
                WHEN email_match = 1.0 THEN 'exact_email_match'
                WHEN phone_match = 1.0 THEN 'exact_phone_match'
                WHEN name_similarity > 0.8 THEN 'similar_name'
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
        sm.created_at
    FROM scored_matches sm
    WHERE sm.combined_score >= p_similarity_threshold
    ORDER BY sm.combined_score DESC, sm.created_at DESC
    LIMIT p_limit;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON COLUMN contacts.search_embedding_384 IS 'Vector embedding (384 dim) for models like all-MiniLM';
COMMENT ON COLUMN contacts.search_embedding_512 IS 'Vector embedding (512 dim) for OpenAI small models';
COMMENT ON COLUMN contacts.search_embedding_768 IS 'Vector embedding (768 dim) for nomic-embed-text, OpenAI default';
COMMENT ON COLUMN contacts.search_embedding_1024 IS 'Vector embedding (1024 dim) for mxbai-embed-large';
COMMENT ON COLUMN contacts.search_embedding_1536 IS 'Vector embedding (1536 dim) for OpenAI large models';
COMMENT ON COLUMN products.search_embedding_768 IS 'Vector embedding (768 dim) for semantic product search';
COMMENT ON COLUMN sales_orders.search_embedding_768 IS 'Vector embedding (768 dim) to power natural language order lookup';
COMMENT ON COLUMN invoices.search_embedding_768 IS 'Vector embedding (768 dim) for invoice and billing insights';
COMMENT ON COLUMN payments.search_embedding_768 IS 'Vector embedding (768 dim) for payment reconciliation search';
COMMENT ON COLUMN purchase_orders.search_embedding_768 IS 'Vector embedding (768 dim) for procurement intelligence';
COMMENT ON COLUMN stock_pickings.search_embedding_768 IS 'Vector embedding (768 dim) for warehouse transfer search';
COMMENT ON COLUMN stock_moves.search_embedding_768 IS 'Vector embedding (768 dim) for inventory movement search';
COMMENT ON COLUMN manufacturing_orders.search_embedding_768 IS 'Vector embedding (768 dim) for production planning insights';
COMMENT ON COLUMN work_orders.search_embedding_768 IS 'Vector embedding (768 dim) for shop floor work order tracking';
