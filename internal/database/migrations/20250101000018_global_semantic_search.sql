-- Migration: Global Semantic Search
-- Description: Add vector embeddings across all modules for intelligent search
-- Created: 2025-01-01

-- =====================================================
-- ADD EMBEDDING COLUMNS TO KEY TABLES
-- =====================================================

-- CRM Module
ALTER TABLE contacts
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- Products Module
ALTER TABLE products
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- Sales Module
ALTER TABLE sales_orders
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- Accounting Module
ALTER TABLE invoices
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

ALTER TABLE payments
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- Procurement
ALTER TABLE purchase_orders
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- Inventory & Fulfillment
ALTER TABLE stock_pickings
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

ALTER TABLE stock_moves
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- Manufacturing
ALTER TABLE manufacturing_orders
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

ALTER TABLE work_orders
ADD COLUMN IF NOT EXISTS search_embedding vector(768);

-- =====================================================
-- CREATE VECTOR INDEXES
-- =====================================================

-- Use IVFFlat index for faster similarity search
CREATE INDEX IF NOT EXISTS idx_contacts_embedding
ON contacts USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_products_embedding
ON products USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_sales_orders_embedding
ON sales_orders USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_invoices_embedding
ON invoices USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_payments_embedding
ON payments USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_purchase_orders_embedding
ON purchase_orders USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_stock_pickings_embedding
ON stock_pickings USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_stock_moves_embedding
ON stock_moves USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_manufacturing_orders_embedding
ON manufacturing_orders USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_work_orders_embedding
ON work_orders USING ivfflat (search_embedding vector_cosine_ops)
WITH (lists = 100);

-- =====================================================
-- GLOBAL SEMANTIC SEARCH FUNCTION
-- =====================================================

CREATE OR REPLACE FUNCTION global_semantic_search(
    p_organization_id uuid,
    p_query_embedding vector(768),
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
BEGIN
    RETURN QUERY
    WITH ranked_results AS (
        -- Contacts/Customers
        SELECT
            'contact'::text as record_type,
            c.id as record_id,
            c.name as record_name,
            c.reference as record_reference,
            COALESCE(c.comment, '') as record_description,
            1 - (c.search_embedding <=> p_query_embedding) as similarity_score,
            jsonb_build_object(
                'email', c.email,
                'phone', c.phone,
                'is_customer', c.is_customer,
                'is_vendor', c.is_vendor,
                'city', c.city,
                'country_id', c.country_id
            ) as metadata,
            c.created_at,
            c.updated_at
        FROM contacts c
        WHERE c.organization_id = p_organization_id
            AND c.deleted_at IS NULL
            AND c.search_embedding IS NOT NULL
            AND p_include_contacts = true

        UNION ALL

        -- Products
        SELECT
            'product'::text,
            p.id,
            p.name,
            p.default_code,
            COALESCE(p.description, ''),
            1 - (p.search_embedding <=> p_query_embedding),
            jsonb_build_object(
                'barcode', p.barcode,
                'list_price', p.list_price,
                'product_type', p.product_type,
                'category_id', p.category_id,
                'active', p.active
            ),
            p.created_at,
            p.updated_at
        FROM products p
        WHERE p.organization_id = p_organization_id
            AND p.deleted_at IS NULL
            AND p.search_embedding IS NOT NULL
            AND p_include_products = true

        UNION ALL

        -- Sales Orders
        SELECT
            'sales_order'::text,
            so.id,
            so.name,
            so.client_order_ref,
            COALESCE(so.note, ''),
            1 - (so.search_embedding <=> p_query_embedding),
            jsonb_build_object(
                'partner_id', so.partner_id,
                'state', so.state,
                'amount_total', so.amount_total,
                'date_order', so.date_order,
                'user_id', so.user_id,
                'team_id', so.team_id
            ),
            so.created_at,
            so.updated_at
        FROM sales_orders so
        WHERE so.organization_id = p_organization_id
            AND so.deleted_at IS NULL
            AND so.search_embedding IS NOT NULL
            AND p_include_sales_orders = true

        UNION ALL

        -- Invoices/Journal Entries
        SELECT
            'invoice'::text,
            inv.id,
            inv.name,
            inv.ref,
            COALESCE(inv.narration, ''),
            1 - (inv.search_embedding <=> p_query_embedding),
            jsonb_build_object(
                'partner_id', inv.partner_id,
                'state', inv.state,
                'move_type', inv.move_type,
                'amount_total', inv.amount_total,
                'date', inv.date,
                'invoice_date', inv.invoice_date
            ),
            inv.created_at,
            inv.updated_at
        FROM invoices inv
        WHERE inv.organization_id = p_organization_id
            AND inv.search_embedding IS NOT NULL
            AND p_include_invoices = true

        UNION ALL

        -- Payments
        SELECT
            'payment'::text,
            pay.id,
            COALESCE(pay.name, pay.ref),
            pay.communication,
            COALESCE(pay.ref, ''),
            1 - (pay.search_embedding <=> p_query_embedding),
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
        WHERE pay.organization_id = p_organization_id
            AND pay.search_embedding IS NOT NULL
            AND p_include_payments = true

        UNION ALL

        -- Purchase Orders
        SELECT
            'purchase_order'::text,
            po.id,
            po.name,
            po.partner_ref,
            COALESCE(po.notes, ''),
            1 - (po.search_embedding <=> p_query_embedding),
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
        WHERE po.organization_id = p_organization_id
            AND po.deleted_at IS NULL
            AND po.search_embedding IS NOT NULL
            AND p_include_purchase_orders = true

        UNION ALL

        -- Stock Pickings (Transfers)
        SELECT
            'stock_picking'::text,
            sp.id,
            sp.name,
            sp.origin,
            COALESCE(sp.note, ''),
            1 - (sp.search_embedding <=> p_query_embedding),
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
        WHERE sp.organization_id = p_organization_id
            AND sp.deleted_at IS NULL
            AND sp.search_embedding IS NOT NULL
            AND p_include_stock_pickings = true

        UNION ALL

        -- Stock Moves
        SELECT
            'stock_move'::text,
            sm.id,
            sm.name,
            sm.reference,
            COALESCE(sm.note, ''),
            1 - (sm.search_embedding <=> p_query_embedding),
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
        WHERE sm.organization_id = p_organization_id
            AND sm.deleted_at IS NULL
            AND sm.search_embedding IS NOT NULL
            AND p_include_stock_moves = true

        UNION ALL

        -- Manufacturing Orders
        SELECT
            'manufacturing_order'::text,
            mo.id,
            mo.name,
            mo.origin,
            COALESCE(mo.metadata->>'routing_notes', '') || ' ' ||
            COALESCE(mo.state, ''),
            1 - (mo.search_embedding <=> p_query_embedding),
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
        WHERE mo.organization_id = p_organization_id
            AND mo.deleted_at IS NULL
            AND mo.search_embedding IS NOT NULL
            AND p_include_manufacturing_orders = true

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
            1 - (wo.search_embedding <=> p_query_embedding),
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
        WHERE wo.organization_id = p_organization_id
            AND wo.search_embedding IS NOT NULL
            AND p_include_work_orders = true
    )
    SELECT *
    FROM ranked_results
    ORDER BY similarity_score DESC
    LIMIT p_limit;
END;
$$;

COMMENT ON FUNCTION global_semantic_search IS 'Search across all modules using vector similarity';

-- =====================================================
-- TEXT-TO-EMBEDDING HELPER (PLACEHOLDER)
-- =====================================================

-- Note: In production, you'll call an external embedding service (OpenAI, Cohere, etc.)
-- This is a placeholder that you'll replace with actual embedding generation
CREATE OR REPLACE FUNCTION generate_search_embedding(text_content text)
RETURNS vector(768)
LANGUAGE plpgsql
AS $$
BEGIN
    -- TODO: Replace with actual embedding generation
    -- For now, return a zero vector
    -- In production, call your embedding API (OpenAI, Sentence Transformers, etc.)
    RAISE NOTICE 'Generate embedding for: %', text_content;
    RETURN array_fill(0, ARRAY[768])::vector(768);
END;
$$;

COMMENT ON FUNCTION generate_search_embedding IS 'Placeholder - integrate with embedding service (OpenAI, Cohere, etc.)';

-- =====================================================
-- AUTO-UPDATE EMBEDDINGS ON INSERT/UPDATE
-- =====================================================

-- Contacts trigger
CREATE OR REPLACE FUNCTION update_contact_embedding()
RETURNS TRIGGER AS $$
BEGIN
    -- Combine relevant text fields for embedding
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.email, '') || ' ' ||
        COALESCE(NEW.phone, '') || ' ' ||
        COALESCE(NEW.comment, '') || ' ' ||
        COALESCE(NEW.job_position, '') || ' ' ||
        COALESCE(NEW.reference, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_contact_embedding
    BEFORE INSERT OR UPDATE OF name, email, phone, comment, job_position, reference
    ON contacts
    FOR EACH ROW
    EXECUTE FUNCTION update_contact_embedding();

-- Products trigger
CREATE OR REPLACE FUNCTION update_product_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.default_code, '') || ' ' ||
        COALESCE(NEW.barcode, '') || ' ' ||
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(NEW.description_sale, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_product_embedding
    BEFORE INSERT OR UPDATE OF name, default_code, barcode, description, description_sale
    ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_product_embedding();

-- Sales Orders trigger
CREATE OR REPLACE FUNCTION update_sales_order_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.client_order_ref, '') || ' ' ||
        COALESCE(NEW.note, '') || ' ' ||
        COALESCE(NEW.origin, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_sales_order_embedding
    BEFORE INSERT OR UPDATE OF name, client_order_ref, note, origin
    ON sales_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_sales_order_embedding();

-- Invoices trigger
CREATE OR REPLACE FUNCTION update_invoice_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.ref, '') || ' ' ||
        COALESCE(NEW.narration, '') || ' ' ||
        COALESCE(NEW.invoice_origin, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_invoice_embedding
    BEFORE INSERT OR UPDATE OF name, ref, narration, invoice_origin
    ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_invoice_embedding();

-- Payments trigger
CREATE OR REPLACE FUNCTION update_payment_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.ref, '') || ' ' ||
        COALESCE(NEW.communication, '') || ' ' ||
        COALESCE(NEW.payment_type, '') || ' ' ||
        COALESCE(NEW.state, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_payment_embedding
    BEFORE INSERT OR UPDATE OF name, ref, communication, payment_type, state
    ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_embedding();

-- Purchase Orders trigger
CREATE OR REPLACE FUNCTION update_purchase_order_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.partner_ref, '') || ' ' ||
        COALESCE(NEW.notes, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.state, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_purchase_order_embedding
    BEFORE INSERT OR UPDATE OF name, partner_ref, notes, origin, state
    ON purchase_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_purchase_order_embedding();

-- Stock Pickings trigger
CREATE OR REPLACE FUNCTION update_stock_picking_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.note, '') || ' ' ||
        COALESCE(NEW.state, '') || ' ' ||
        COALESCE(NEW.sequence_code, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_stock_picking_embedding
    BEFORE INSERT OR UPDATE OF name, origin, note, state, sequence_code
    ON stock_pickings
    FOR EACH ROW
    EXECUTE FUNCTION update_stock_picking_embedding();

-- Stock Moves trigger
CREATE OR REPLACE FUNCTION update_stock_move_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.reference, '') || ' ' ||
        COALESCE(NEW.note, '') || ' ' ||
        COALESCE(NEW.state, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_stock_move_embedding
    BEFORE INSERT OR UPDATE OF name, origin, reference, note, state
    ON stock_moves
    FOR EACH ROW
    EXECUTE FUNCTION update_stock_move_embedding();

-- Manufacturing Orders trigger
CREATE OR REPLACE FUNCTION update_manufacturing_order_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.origin, '') || ' ' ||
        COALESCE(NEW.state, '') || ' ' ||
        COALESCE(NEW.metadata::text, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_mo_embedding
    BEFORE INSERT OR UPDATE OF name, origin, state, metadata
    ON manufacturing_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_manufacturing_order_embedding();

-- Work Orders trigger
CREATE OR REPLACE FUNCTION update_work_order_embedding()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_embedding := generate_search_embedding(
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.state, '') || ' ' ||
        COALESCE(NEW.production_id::text, '') || ' ' ||
        COALESCE(NEW.workcenter_id::text, '') || ' ' ||
        COALESCE(NEW.date_planned_start::text, '') || ' ' ||
        COALESCE(NEW.date_planned_finished::text, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_work_order_embedding
    BEFORE INSERT OR UPDATE OF name, state, production_id, workcenter_id, date_planned_start, date_planned_finished
    ON work_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_work_order_embedding();


-- =====================================================
-- CONVENIENCE SEARCH FUNCTIONS
-- =====================================================

-- Search by text query (converts text to embedding first)
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
DECLARE
    v_query_embedding vector(768);
BEGIN
    -- Generate embedding for search query
    v_query_embedding := generate_search_embedding(p_query_text);

    -- Call global search
    RETURN QUERY
    SELECT
        gs.record_type,
        gs.record_id,
        gs.record_name,
        gs.record_reference,
        gs.similarity_score
    FROM global_semantic_search(
        p_organization_id,
        v_query_embedding,
        p_limit
    ) gs;
END;
$$;

COMMENT ON FUNCTION search_by_text IS 'Convenience function to search using text instead of embedding vector';

-- =====================================================
-- PREPARE SCHEMAS FOR FUTURE FEATURES
-- =====================================================

-- Table for storing duplicate detection configurations
CREATE TABLE duplicate_detection_config (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type text NOT NULL, -- 'contact', 'product', etc.
    similarity_threshold float NOT NULL DEFAULT 0.85,
    fields_to_compare text[] NOT NULL,
    auto_merge boolean DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT duplicate_detection_entity_type_check
        CHECK (entity_type IN (
            'contact',
            'product',
            'sales_order',
            'invoice',
            'payment',
            'purchase_order',
            'stock_picking',
            'stock_move',
            'manufacturing_order',
            'work_order'
        ))
);

-- Table for tracking detected duplicates
CREATE TABLE duplicate_candidates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type text NOT NULL,
    record_id_1 uuid NOT NULL,
    record_id_2 uuid NOT NULL,
    similarity_score float NOT NULL,
    status text DEFAULT 'pending', -- 'pending', 'merged', 'ignored', 'confirmed_different'
    reviewed_by uuid,
    reviewed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT duplicate_candidates_status_check
        CHECK (status IN ('pending', 'merged', 'ignored', 'confirmed_different'))
);

CREATE INDEX idx_duplicate_candidates_org_status
ON duplicate_candidates(organization_id, status);

-- Table for contextual help tracking (analytics)
CREATE TABLE contextual_help_usage (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    screen_context text NOT NULL,
    query_text text,
    knowledge_entry_id uuid REFERENCES knowledge_entries(id),
    was_helpful boolean,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_contextual_help_org_user
ON contextual_help_usage(organization_id, user_id);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON COLUMN contacts.search_embedding IS 'Vector embedding for semantic search across contact data';
COMMENT ON COLUMN products.search_embedding IS 'Vector embedding for semantic search across product data';
COMMENT ON COLUMN sales_orders.search_embedding IS 'Vector embedding for semantic search across sales orders';
COMMENT ON COLUMN invoices.search_embedding IS 'Vector embedding for semantic search across invoices/journal entries';
COMMENT ON COLUMN payments.search_embedding IS 'Vector embedding for semantic search across payment records';
COMMENT ON COLUMN purchase_orders.search_embedding IS 'Vector embedding for semantic search across purchase orders';
COMMENT ON COLUMN stock_pickings.search_embedding IS 'Vector embedding for semantic search across stock transfers';
COMMENT ON COLUMN stock_moves.search_embedding IS 'Vector embedding for semantic search across inventory movements';
COMMENT ON COLUMN manufacturing_orders.search_embedding IS 'Vector embedding for semantic search across manufacturing orders';
COMMENT ON COLUMN work_orders.search_embedding IS 'Vector embedding for semantic search across work orders';

COMMENT ON TABLE duplicate_detection_config IS 'Configuration for automatic duplicate detection per entity type';
COMMENT ON TABLE duplicate_candidates IS 'Tracks potential duplicate records detected by similarity matching';
COMMENT ON TABLE contextual_help_usage IS 'Analytics for contextual help feature usage and effectiveness';
