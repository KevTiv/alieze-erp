-- Migration: Products & Inventory Module Tables
-- Description: Product management and inventory control tables
-- Created: 2025-01-01

-- =====================================================
-- PRODUCTS & INVENTORY MODULE TABLES
-- =====================================================

-- Product Categories
CREATE TABLE product_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    complete_name varchar(500),
    parent_id uuid REFERENCES product_categories(id),
    parent_path varchar(1000),
    sequence integer DEFAULT 10,
    removal_strategy varchar(20) DEFAULT 'fifo',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    CONSTRAINT product_categories_removal_check CHECK (removal_strategy IN ('fifo', 'lifo', 'nearest'))
);

-- Products (product.template)
CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    display_name varchar(500),
    default_code varchar(100),
    barcode varchar(100),
    product_type varchar(20) DEFAULT 'storable',
    category_id uuid REFERENCES product_categories(id),
    -- Pricing
    list_price numeric(15,2) DEFAULT 0,
    standard_price numeric(15,2) DEFAULT 0,
    currency_id uuid REFERENCES currencies(id),
    -- Sales
    sale_ok boolean DEFAULT true,
    can_be_sold boolean DEFAULT true,
    invoice_policy varchar(20) DEFAULT 'order',
    -- Purchase
    purchase_ok boolean DEFAULT true,
    can_be_purchased boolean DEFAULT true,
    -- Inventory
    tracking varchar(20) DEFAULT 'none',
    weight numeric(10,4),
    volume numeric(10,4),
    -- UoM
    uom_id uuid REFERENCES uom_units(id),
    uom_po_id uuid REFERENCES uom_units(id),
    -- Tax
    tax_ids uuid[],
    supplier_tax_ids uuid[],
    -- Description
    description text,
    description_purchase text,
    description_sale text,
    -- Image
    image_url varchar(500),
    image_variant_urls jsonb,
    -- Misc
    active boolean DEFAULT true,
    company_id uuid REFERENCES companies(id),
    responsible_id uuid,
    -- Standard fields
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT products_type_check CHECK (product_type IN ('consumable', 'service', 'storable')),
    CONSTRAINT products_tracking_check CHECK (tracking IN ('none', 'lot', 'serial')),
    CONSTRAINT products_invoice_policy_check CHECK (invoice_policy IN ('order', 'delivery'))
);

-- Product Variants (product.product)
CREATE TABLE product_variants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_tmpl_id uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name varchar(255),
    default_code varchar(100),
    barcode varchar(100),
    -- Price override
    list_price numeric(15,2),
    standard_price numeric(15,2),
    -- Variant attributes
    attribute_values jsonb,
    combination_indices varchar(255),
    -- Image
    image_variant_url varchar(500),
    -- Volume/weight override
    weight numeric(10,4),
    volume numeric(10,4),
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Warehouses (stock.warehouse)
CREATE TABLE warehouses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(10) NOT NULL,
    partner_id uuid REFERENCES contacts(id),
    -- Locations (will be linked after stock_locations table)
    lot_stock_id uuid,
    wh_input_stock_loc_id uuid,
    wh_qc_stock_loc_id uuid,
    wh_output_stock_loc_id uuid,
    wh_pack_stock_loc_id uuid,
    -- Routes
    reception_steps varchar(20) DEFAULT 'one_step',
    delivery_steps varchar(20) DEFAULT 'ship_only',
    -- Misc
    active boolean DEFAULT true,
    sequence integer DEFAULT 10,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb
);

-- Stock Locations (stock.location)
CREATE TABLE stock_locations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    complete_name varchar(500),
    location_id uuid REFERENCES stock_locations(id),
    usage varchar(20) DEFAULT 'internal',
    barcode varchar(100),
    removal_strategy varchar(20) DEFAULT 'fifo',
    comment text,
    posx integer,
    posy integer,
    posz integer,
    active boolean DEFAULT true,
    scrap_location boolean DEFAULT false,
    return_location boolean DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,

    CONSTRAINT stock_locations_usage_check CHECK (usage IN ('supplier', 'view', 'internal', 'customer', 'inventory', 'production', 'transit')),
    CONSTRAINT stock_locations_removal_check CHECK (removal_strategy IN ('fifo', 'lifo', 'nearest'))
);

-- Update warehouses with foreign keys to stock_locations
-- Add foreign key constraints with NOT VALID to avoid table locks during migration
ALTER TABLE warehouses
    ADD CONSTRAINT warehouses_lot_stock_fk FOREIGN KEY (lot_stock_id) REFERENCES stock_locations(id) NOT VALID,
    ADD CONSTRAINT warehouses_input_stock_fk FOREIGN KEY (wh_input_stock_loc_id) REFERENCES stock_locations(id) NOT VALID,
    ADD CONSTRAINT warehouses_qc_stock_fk FOREIGN KEY (wh_qc_stock_loc_id) REFERENCES stock_locations(id) NOT VALID,
    ADD CONSTRAINT warehouses_output_stock_fk FOREIGN KEY (wh_output_stock_loc_id) REFERENCES stock_locations(id) NOT VALID,
    ADD CONSTRAINT warehouses_pack_stock_fk FOREIGN KEY (wh_pack_stock_loc_id) REFERENCES stock_locations(id) NOT VALID;

-- Validate constraints separately to avoid prolonged locks
ALTER TABLE warehouses VALIDATE CONSTRAINT warehouses_lot_stock_fk;
ALTER TABLE warehouses VALIDATE CONSTRAINT warehouses_input_stock_fk;
ALTER TABLE warehouses VALIDATE CONSTRAINT warehouses_qc_stock_fk;
ALTER TABLE warehouses VALIDATE CONSTRAINT warehouses_output_stock_fk;
ALTER TABLE warehouses VALIDATE CONSTRAINT warehouses_pack_stock_fk;

-- Stock Packages
CREATE TABLE stock_packages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    location_id uuid REFERENCES stock_locations(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Stock Lots (Lot/Serial numbers)
CREATE TABLE stock_lots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    ref varchar(255),
    product_id uuid NOT NULL REFERENCES products(id),
    expiration_date date,
    use_date date,
    removal_date date,
    alert_date date,
    note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT stock_lots_unique UNIQUE(organization_id, company_id, product_id, name)
);

-- Stock Quants (Real-time inventory)
CREATE TABLE stock_quants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    product_id uuid NOT NULL REFERENCES products(id),
    location_id uuid NOT NULL REFERENCES stock_locations(id),
    lot_id uuid REFERENCES stock_lots(id),
    package_id uuid REFERENCES stock_packages(id),
    owner_id uuid REFERENCES contacts(id),
    quantity numeric(15,4) DEFAULT 0,
    reserved_quantity numeric(15,4) DEFAULT 0,
    in_date timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT stock_quants_unique UNIQUE(product_id, location_id, lot_id, package_id, owner_id, organization_id)
);

-- Procurement Groups
CREATE TABLE procurement_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    partner_id uuid REFERENCES contacts(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Stock Rules
CREATE TABLE stock_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    action varchar(50) NOT NULL,
    location_src_id uuid REFERENCES stock_locations(id),
    location_dest_id uuid REFERENCES stock_locations(id),
    sequence integer DEFAULT 10,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Stock Picking Types (Operation types)
CREATE TABLE stock_picking_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    code varchar(20),
    sequence integer DEFAULT 10,
    sequence_id uuid REFERENCES sequences(id),
    default_location_src_id uuid REFERENCES stock_locations(id),
    default_location_dest_id uuid REFERENCES stock_locations(id),
    warehouse_id uuid REFERENCES warehouses(id),
    color integer,
    barcode varchar(100),
    use_create_lots boolean DEFAULT true,
    use_existing_lots boolean DEFAULT true,
    show_entire_packs boolean DEFAULT false,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT stock_picking_types_code_check CHECK (code IN ('incoming', 'outgoing', 'internal'))
);

-- Stock Pickings (Transfers/Deliveries)
CREATE TABLE stock_pickings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    sequence_code varchar(20),
    picking_type_id uuid REFERENCES stock_picking_types(id),
    location_id uuid REFERENCES stock_locations(id),
    location_dest_id uuid REFERENCES stock_locations(id),
    partner_id uuid REFERENCES contacts(id),
    date timestamptz DEFAULT now(),
    scheduled_date timestamptz,
    date_deadline timestamptz,
    date_done timestamptz,
    origin varchar(255),
    state varchar(20) DEFAULT 'draft',
    priority varchar(10) DEFAULT '1',
    user_id uuid,
    owner_id uuid REFERENCES contacts(id),
    note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT stock_pickings_state_check CHECK (state IN ('draft', 'waiting', 'confirmed', 'assigned', 'done', 'cancel'))
);

-- Stock Moves (Inventory movements)
CREATE TABLE stock_moves (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    sequence integer DEFAULT 10,
    priority varchar(10) DEFAULT '1',
    date timestamptz DEFAULT now(),
    date_deadline timestamptz,
    product_id uuid NOT NULL REFERENCES products(id),
    product_uom_qty numeric(15,4) NOT NULL,
    product_uom uuid REFERENCES uom_units(id),
    location_id uuid NOT NULL REFERENCES stock_locations(id),
    location_dest_id uuid NOT NULL REFERENCES stock_locations(id),
    partner_id uuid REFERENCES contacts(id),
    picking_id uuid REFERENCES stock_pickings(id),
    state varchar(20) DEFAULT 'draft',
    procure_method varchar(20) DEFAULT 'make_to_stock',
    origin varchar(255),
    group_id uuid REFERENCES procurement_groups(id),
    rule_id uuid REFERENCES stock_rules(id),
    lot_ids uuid[],
    note text,
    reference varchar(255),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT stock_moves_state_check CHECK (state IN ('draft', 'confirmed', 'assigned', 'done', 'cancel'))
);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_product_categories_updated_at
    BEFORE UPDATE ON product_categories
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_product_variants_updated_at
    BEFORE UPDATE ON product_variants
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_warehouses_updated_at
    BEFORE UPDATE ON warehouses
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_locations_updated_at
    BEFORE UPDATE ON stock_locations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_lots_updated_at
    BEFORE UPDATE ON stock_lots
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_quants_updated_at
    BEFORE UPDATE ON stock_quants
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_rules_updated_at
    BEFORE UPDATE ON stock_rules
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_picking_types_updated_at
    BEFORE UPDATE ON stock_picking_types
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_pickings_updated_at
    BEFORE UPDATE ON stock_pickings
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_stock_moves_updated_at
    BEFORE UPDATE ON stock_moves
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_product_categories_org ON product_categories(organization_id);
CREATE INDEX idx_product_categories_parent ON product_categories(parent_id) WHERE parent_id IS NOT NULL;

CREATE INDEX idx_products_org ON products(organization_id);
CREATE INDEX idx_products_category ON products(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX idx_products_default_code ON products(default_code) WHERE default_code IS NOT NULL;
CREATE INDEX idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_products_name_trgm ON products USING gin(name gin_trgm_ops);
CREATE INDEX idx_products_deleted_at ON products(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX idx_product_variants_tmpl ON product_variants(product_tmpl_id);
CREATE INDEX idx_product_variants_barcode ON product_variants(barcode) WHERE barcode IS NOT NULL;

CREATE INDEX idx_warehouses_org ON warehouses(organization_id);
CREATE INDEX idx_stock_locations_org ON stock_locations(organization_id);
CREATE INDEX idx_stock_locations_parent ON stock_locations(location_id) WHERE location_id IS NOT NULL;
CREATE INDEX idx_stock_locations_usage ON stock_locations(organization_id, usage);

CREATE INDEX idx_stock_lots_org ON stock_lots(organization_id);
CREATE INDEX idx_stock_lots_product ON stock_lots(product_id);

CREATE INDEX idx_stock_quants_org ON stock_quants(organization_id);
CREATE INDEX idx_stock_quants_product ON stock_quants(product_id);
CREATE INDEX idx_stock_quants_location ON stock_quants(location_id);
CREATE INDEX idx_stock_quants_product_location ON stock_quants(product_id, location_id);

CREATE INDEX idx_stock_pickings_org ON stock_pickings(organization_id);
CREATE INDEX idx_stock_pickings_partner ON stock_pickings(partner_id) WHERE partner_id IS NOT NULL;
CREATE INDEX idx_stock_pickings_state ON stock_pickings(organization_id, state);
CREATE INDEX idx_stock_pickings_date ON stock_pickings(date);

CREATE INDEX idx_stock_moves_org ON stock_moves(organization_id);
CREATE INDEX idx_stock_moves_product ON stock_moves(product_id);
CREATE INDEX idx_stock_moves_picking ON stock_moves(picking_id) WHERE picking_id IS NOT NULL;
CREATE INDEX idx_stock_moves_state ON stock_moves(organization_id, state);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE product_categories IS 'Hierarchical product categorization';
COMMENT ON TABLE products IS 'Product templates - equivalent to product.template in Odoo';
COMMENT ON TABLE product_variants IS 'Product variants with attributes - equivalent to product.product in Odoo';
COMMENT ON TABLE warehouses IS 'Warehouse configuration and locations';
COMMENT ON TABLE stock_locations IS 'Stock location hierarchy (warehouses, shelves, bins, etc)';
COMMENT ON TABLE stock_lots IS 'Lot and serial number tracking';
COMMENT ON TABLE stock_quants IS 'Real-time inventory quantities per product/location';
COMMENT ON TABLE stock_pickings IS 'Transfer orders (receipts, deliveries, internal transfers)';
COMMENT ON TABLE stock_moves IS 'Individual stock movements between locations';
COMMENT ON TABLE stock_picking_types IS 'Operation type configuration (incoming, outgoing, internal)';
