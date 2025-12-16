-- Migration: Barcode Scanning Integration
-- Description: Add barcode support for mobile inventory operations
-- Created: 2025-01-01

-- =====================================================
-- BARCODE SCANNING MODULE
-- =====================================================

-- Add barcode fields to products (if not already present)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'barcode'
    ) THEN
        ALTER TABLE products ADD COLUMN barcode VARCHAR(100);
        CREATE INDEX idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
    END IF;
END $$;

-- Add barcode fields to product variants (if not already present)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'product_variants' AND column_name = 'barcode'
    ) THEN
        ALTER TABLE product_variants ADD COLUMN barcode VARCHAR(100);
        CREATE INDEX idx_product_variants_barcode ON product_variants(barcode) WHERE barcode IS NOT NULL;
    END IF;
END $$;

-- Add barcode fields to stock locations (if not already present)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'stock_locations' AND column_name = 'barcode'
    ) THEN
        ALTER TABLE stock_locations ADD COLUMN barcode VARCHAR(100);
        CREATE INDEX idx_stock_locations_barcode ON stock_locations(barcode) WHERE barcode IS NOT NULL;
    END IF;
END $$;

-- Add barcode fields to stock lots (if not already present)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'stock_lots' AND column_name = 'barcode'
    ) THEN
        ALTER TABLE stock_lots ADD COLUMN barcode VARCHAR(100);
        CREATE INDEX idx_stock_lots_barcode ON stock_lots(barcode) WHERE barcode IS NOT NULL;
    END IF;
END $$;

-- Add barcode fields to stock packages (if not already present)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'stock_packages' AND column_name = 'barcode'
    ) THEN
        ALTER TABLE stock_packages ADD COLUMN barcode VARCHAR(100);
        CREATE INDEX idx_stock_packages_barcode ON stock_packages(barcode) WHERE barcode IS NOT NULL;
    END IF;
END $$;

-- Create barcode scanning log table
CREATE TABLE IF NOT EXISTS barcode_scans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid REFERENCES organization_users(id),
    scan_type VARCHAR(50) NOT NULL, -- product, location, lot, package, move
    scanned_barcode VARCHAR(100) NOT NULL,
    entity_id uuid, -- ID of the scanned entity
    entity_type VARCHAR(50), -- Type of entity (product, location, etc.)
    location_id uuid REFERENCES stock_locations(id),
    quantity NUMERIC(15,4),
    scan_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    device_info JSONB,
    ip_address VARCHAR(50),
    success boolean DEFAULT true,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for barcode scans
CREATE INDEX IF NOT EXISTS idx_barcode_scans_org ON barcode_scans(organization_id);
CREATE INDEX IF NOT EXISTS idx_barcode_scans_barcode ON barcode_scans(scanned_barcode);
CREATE INDEX IF NOT EXISTS idx_barcode_scans_type ON barcode_scans(scan_type);
CREATE INDEX IF NOT EXISTS idx_barcode_scans_time ON barcode_scans(scan_time);
CREATE INDEX IF NOT EXISTS idx_barcode_scans_user ON barcode_scans(user_id) WHERE user_id IS NOT NULL;

-- Create barcode format validation function
CREATE OR REPLACE FUNCTION validate_barcode_format(barcode_text VARCHAR)
RETURNS boolean AS $$
DECLARE
    -- Common barcode patterns
    ean13_pattern VARCHAR := '^[0-9]{13}$';
    upc_pattern VARCHAR := '^[0-9]{12}$';
    code128_pattern VARCHAR := '^[A-Za-z0-9\-\.\$\/\+\%]{6,20}$';
    qr_pattern VARCHAR := '^[A-Za-z0-9\-\_\.\:\/]{10,100}$';
    simple_pattern VARCHAR := '^[A-Za-z0-9\-]{4,50}$';
BEGIN
    -- Check against common patterns
    IF barcode_text ~ ean13_pattern THEN
        RETURN true;
    ELSIF barcode_text ~ upc_pattern THEN
        RETURN true;
    ELSIF barcode_text ~ code128_pattern THEN
        RETURN true;
    ELSIF barcode_text ~ qr_pattern THEN
        RETURN true;
    ELSIF barcode_text ~ simple_pattern THEN
        RETURN true;
    END IF;

    RETURN false;
END;
$$ LANGUAGE plpgsql;

-- Create barcode generation function
CREATE OR REPLACE FUNCTION generate_barcode(entity_type VARCHAR, org_id uuid, prefix VARCHAR)
RETURNS VARCHAR AS $$
DECLARE
    base_code VARCHAR;
    random_part VARCHAR;
    full_code VARCHAR;
    counter INT;
BEGIN
    -- Generate random part
    random_part := LPAD(CAST(FLOOR(RANDOM() * 1000000) AS VARCHAR), 6, '0');

    -- Create base code
    IF prefix IS NULL OR prefix = '' THEN
        base_code := SUBSTRING(MD5(org_id::text), 1, 4);
    ELSE
        base_code := prefix;
    END IF;

    -- Create full code
    full_code := base_code || '-' || random_part;

    -- Ensure uniqueness
    counter := 0;
    WHILE EXISTS (
        SELECT 1 FROM products WHERE barcode = full_code
        UNION ALL
        SELECT 1 FROM product_variants WHERE barcode = full_code
        UNION ALL
        SELECT 1 FROM stock_locations WHERE barcode = full_code
        UNION ALL
        SELECT 1 FROM stock_lots WHERE barcode = full_code
        UNION ALL
        SELECT 1 FROM stock_packages WHERE barcode = full_code
    ) AND counter < 100 LOOP
        random_part := LPAD(CAST(FLOOR(RANDOM() * 1000000) AS VARCHAR), 6, '0');
        full_code := base_code || '-' || random_part;
        counter := counter + 1;
    END LOOP;

    RETURN full_code;
END;
$$ LANGUAGE plpgsql;

-- Create function to find entity by barcode
CREATE OR REPLACE FUNCTION find_entity_by_barcode(barcode_text VARCHAR, org_id uuid)
RETURNS TABLE (
    entity_type VARCHAR,
    entity_id uuid,
    entity_name VARCHAR,
    additional_info JSONB
) AS $$
BEGIN
    -- Search in products
    RETURN QUERY
    SELECT
        'product'::VARCHAR as entity_type,
        p.id as entity_id,
        p.name as entity_name,
        jsonb_build_object(
            'default_code', p.default_code,
            'product_type', p.product_type,
            'category_id', p.category_id
        ) as additional_info
    FROM products p
    WHERE p.barcode = barcode_text AND p.organization_id = org_id
    LIMIT 1;

    -- Search in product variants
    RETURN QUERY
    SELECT
        'product_variant'::VARCHAR as entity_type,
        pv.id as entity_id,
        COALESCE(pv.name, p.name) as entity_name,
        jsonb_build_object(
            'product_id', pv.product_tmpl_id,
            'default_code', pv.default_code,
            'attribute_values', pv.attribute_values
        ) as additional_info
    FROM product_variants pv
    JOIN products p ON pv.product_tmpl_id = p.id
    WHERE pv.barcode = barcode_text AND p.organization_id = org_id
    LIMIT 1;

    -- Search in stock locations
    RETURN QUERY
    SELECT
        'location'::VARCHAR as entity_type,
        sl.id as entity_id,
        sl.name as entity_name,
        jsonb_build_object(
            'usage', sl.usage,
            'complete_name', sl.complete_name,
            'location_id', sl.location_id
        ) as additional_info
    FROM stock_locations sl
    WHERE sl.barcode = barcode_text AND sl.organization_id = org_id
    LIMIT 1;

    -- Search in stock lots
    RETURN QUERY
    SELECT
        'lot'::VARCHAR as entity_type,
        slots.id as entity_id,
        slots.name as entity_name,
        jsonb_build_object(
            'product_id', slots.product_id,
            'ref', slots.ref,
            'expiration_date', slots.expiration_date
        ) as additional_info
    FROM stock_lots slots
    WHERE slots.barcode = barcode_text AND slots.organization_id = org_id
    LIMIT 1;

    -- Search in stock packages
    RETURN QUERY
    SELECT
        'package'::VARCHAR as entity_type,
        sp.id as entity_id,
        sp.name as entity_name,
        jsonb_build_object(
            'location_id', sp.location_id,
            'created_at', sp.created_at
        ) as additional_info
    FROM stock_packages sp
    WHERE sp.barcode = barcode_text AND sp.organization_id = org_id
    LIMIT 1;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create mobile scanning sessions table
CREATE TABLE IF NOT EXISTS mobile_scanning_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid REFERENCES organization_users(id),
    device_id VARCHAR(100),
    session_type VARCHAR(50) NOT NULL, -- inventory, picking, receiving, counting
    status VARCHAR(20) DEFAULT 'active', -- active, completed, cancelled
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    location_id uuid REFERENCES stock_locations(id),
    reference VARCHAR(100),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create mobile scanning session lines table
CREATE TABLE IF NOT EXISTS mobile_scanning_session_lines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES mobile_scanning_sessions(id) ON DELETE CASCADE,
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    scan_id uuid REFERENCES barcode_scans(id),
    product_id uuid REFERENCES products(id),
    product_variant_id uuid REFERENCES product_variants(id),
    location_id uuid REFERENCES stock_locations(id),
    lot_id uuid REFERENCES stock_lots(id),
    package_id uuid REFERENCES stock_packages(id),
    quantity NUMERIC(15,4),
    scanned_quantity NUMERIC(15,4),
    uom_id uuid REFERENCES uom_units(id),
    status VARCHAR(20) DEFAULT 'scanned', -- scanned, verified, processed, error
    notes TEXT,
    sequence INTEGER DEFAULT 10,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for mobile scanning tables
CREATE INDEX IF NOT EXISTS idx_mobile_sessions_org ON mobile_scanning_sessions(organization_id);
CREATE INDEX IF NOT EXISTS idx_mobile_sessions_user ON mobile_scanning_sessions(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mobile_sessions_status ON mobile_scanning_sessions(status);
CREATE INDEX IF NOT EXISTS idx_mobile_session_lines_org ON mobile_scanning_session_lines(organization_id);
CREATE INDEX IF NOT EXISTS idx_mobile_session_lines_session ON mobile_scanning_session_lines(session_id);
CREATE INDEX IF NOT EXISTS idx_mobile_session_lines_product ON mobile_scanning_session_lines(product_id);

-- Create function to create mobile scanning session
CREATE OR REPLACE FUNCTION create_mobile_scanning_session(
    org_id uuid,
    user_id uuid,
    session_type VARCHAR,
    location_id uuid,
    reference VARCHAR
)
RETURNS uuid AS $$
DECLARE
    session_id uuid;
BEGIN
    session_id := gen_random_uuid();

    INSERT INTO mobile_scanning_sessions (
        id, organization_id, user_id, session_type, location_id, reference
    ) VALUES (
        session_id, org_id, user_id, session_type, location_id, reference
    );

    RETURN session_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to add scan to session
CREATE OR REPLACE FUNCTION add_scan_to_session(
    session_id uuid,
    org_id uuid,
    user_id uuid,
    scanned_barcode VARCHAR,
    quantity NUMERIC,
    location_id uuid,
    device_info JSONB
)
RETURNS TABLE (
    success boolean,
    message TEXT,
    scan_id uuid,
    entity_type VARCHAR,
    entity_id uuid,
    entity_name VARCHAR
) AS $$
DECLARE
    scan_id uuid;
    entity_record RECORD;
    product_id uuid;
    location_id_param uuid;
BEGIN
    -- Validate session
    IF NOT EXISTS (
        SELECT 1 FROM mobile_scanning_sessions
        WHERE id = session_id AND organization_id = org_id AND status = 'active'
    ) THEN
        RETURN QUERY SELECT false, 'Invalid or inactive session', NULL, NULL, NULL, NULL;
        RETURN;
    END IF;

    -- Find entity by barcode
    FOR entity_record IN
        SELECT * FROM find_entity_by_barcode(scanned_barcode, org_id)
    LOOP
        -- Create scan record
        scan_id := gen_random_uuid();

        INSERT INTO barcode_scans (
            id, organization_id, user_id, scan_type, scanned_barcode,
            entity_id, entity_type, location_id, quantity, device_info
        ) VALUES (
            scan_id, org_id, user_id, entity_record.entity_type, scanned_barcode,
            entity_record.entity_id, entity_record.entity_type, location_id, quantity, device_info
        );

        -- Add to session lines
        INSERT INTO mobile_scanning_session_lines (
            session_id, organization_id, scan_id, product_id,
            location_id, quantity, scanned_quantity, status
        ) VALUES (
            session_id, org_id, scan_id,
            CASE WHEN entity_record.entity_type = 'product' THEN entity_record.entity_id ELSE NULL END,
            location_id, quantity, quantity, 'scanned'
        );

        RETURN QUERY SELECT true, 'Scan successful', scan_id,
                      entity_record.entity_type, entity_record.entity_id, entity_record.entity_name;
        RETURN;
    END LOOP;

    -- If no entity found
    RETURN QUERY SELECT false, 'Barcode not found', NULL, NULL, NULL, NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to complete scanning session
CREATE OR REPLACE FUNCTION complete_scanning_session(
    session_id uuid,
    org_id uuid
)
RETURNS boolean AS $$
DECLARE
    session_record RECORD;
BEGIN
    -- Get session
    SELECT * INTO session_record
    FROM mobile_scanning_sessions
    WHERE id = session_id AND organization_id = org_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Update session status
    UPDATE mobile_scanning_sessions
    SET status = 'completed', end_time = NOW(), updated_at = NOW()
    WHERE id = session_id AND organization_id = org_id;

    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions (handled by RLS in main migration)
COMMENT ON TABLE barcode_scans IS 'Barcode scanning log - filtered by organization RLS';
COMMENT ON TABLE mobile_scanning_sessions IS 'Mobile scanning sessions - filtered by organization RLS';
COMMENT ON TABLE mobile_scanning_session_lines IS 'Mobile scanning session lines - filtered by organization RLS';

-- =====================================================
-- END OF MIGRATION
-- =====================================================
