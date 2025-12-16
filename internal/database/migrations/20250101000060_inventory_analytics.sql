-- Migration: Inventory Analytics Module
-- Description: Advanced inventory analytics with per-organization materialized views
-- Created: 2025-01-01

-- =====================================================
-- INVENTORY ANALYTICS MODULE
-- =====================================================

-- Add valuation method to products (if not already present)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'valuation_method'
    ) THEN
        ALTER TABLE products ADD COLUMN valuation_method VARCHAR(20) DEFAULT 'fifo';
        ALTER TABLE products ADD CONSTRAINT products_valuation_check
            CHECK (valuation_method IN ('fifo', 'lifo', 'average', 'standard'));
    END IF;
END $$;

-- Add inventory tracking fields to products
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'last_movement_date'
    ) THEN
        ALTER TABLE products ADD COLUMN last_movement_date TIMESTAMPTZ;
        ALTER TABLE products ADD COLUMN days_since_movement INTEGER
            GENERATED ALWAYS AS (CASE
                WHEN last_movement_date IS NULL THEN NULL
                ELSE CURRENT_DATE - DATE(last_movement_date)
            END) STORED;
    END IF;
END $$;

-- Add reorder and safety stock fields
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'reorder_point'
    ) THEN
        ALTER TABLE products ADD COLUMN reorder_point NUMERIC(15,4);
        ALTER TABLE products ADD COLUMN safety_stock NUMERIC(15,4);
        ALTER TABLE products ADD COLUMN lead_time_days INTEGER DEFAULT 7;
    END IF;
END $$;

-- Create inventory valuation view (PER ORGANIZATION)
CREATE OR REPLACE VIEW inventory_valuation AS
SELECT
    p.organization_id,
    p.id as product_id,
    p.name as product_name,
    p.default_code,
    p.category_id,
    p.valuation_method,
    SUM(sq.quantity) as total_quantity,
    SUM(sq.quantity * p.standard_price) as current_value,
    SUM(sq.quantity * p.list_price) as retail_value,
    SUM(sq.quantity * p.standard_price) - SUM(sq.quantity * p.standard_price) as unrealized_gain_loss,
    p.currency_id,
    p.uom_id,
    p.active,
    p.created_at,
    p.updated_at
FROM products p
JOIN stock_quants sq ON p.id = sq.product_id AND p.organization_id = sq.organization_id
WHERE sq.quantity > 0 AND sq.deleted_at IS NULL
GROUP BY p.organization_id, p.id;

-- Create inventory turnover view (PER ORGANIZATION)
CREATE OR REPLACE VIEW inventory_turnover AS
SELECT
    p.organization_id,
    p.id as product_id,
    p.name as product_name,
    p.category_id,
    -- Calculate COGS (Cost of Goods Sold) for the period
    COALESCE(SUM(
        CASE
            WHEN sm.state = 'done' AND sm.date >= DATE_TRUNC('year', CURRENT_DATE)
            THEN sm.product_uom_qty * p.standard_price
            ELSE 0
        END
    ), 0) as annual_cogs,
    -- Calculate average inventory
    COALESCE(AVG(sq.quantity), 0) as average_inventory,
    -- Calculate turnover ratio
    CASE
        WHEN AVG(sq.quantity) > 0
        THEN COALESCE(SUM(
            CASE
                WHEN sm.state = 'done' AND sm.date >= DATE_TRUNC('year', CURRENT_DATE)
                THEN sm.product_uom_qty * p.standard_price
                ELSE 0
            END
        ), 0) / NULLIF(AVG(sq.quantity), 0)
        ELSE 0
    END as turnover_ratio,
    -- Days of supply
    CASE
        WHEN AVG(sq.quantity) > 0
        THEN 365 / NULLIF(CASE
            WHEN AVG(sq.quantity) > 0
            THEN COALESCE(SUM(
                CASE
                    WHEN sm.state = 'done' AND sm.date >= DATE_TRUNC('year', CURRENT_DATE)
                    THEN sm.product_uom_qty * p.standard_price
                    ELSE 0
                END
            ), 0) / NULLIF(AVG(sq.quantity), 0)
            ELSE 0
        END, 0)
        ELSE 999
    END as days_of_supply
FROM products p
JOIN stock_quants sq ON p.id = sq.product_id AND p.organization_id = sq.organization_id
LEFT JOIN stock_moves sm ON p.id = sm.product_id AND p.organization_id = sm.organization_id
WHERE sq.deleted_at IS NULL AND (sm.deleted_at IS NULL OR sm.deleted_at IS NOT NULL)
GROUP BY p.organization_id, p.id;

-- Create stock aging view (PER ORGANIZATION)
CREATE OR REPLACE VIEW inventory_aging AS
SELECT
    p.organization_id,
    p.id as product_id,
    p.name as product_name,
    p.category_id,
    p.default_code,
    sq.location_id,
    sl.name as location_name,
    sq.lot_id,
    slots.name as lot_name,
    sq.quantity,
    sq.in_date,
    -- Age brackets
    CASE
        WHEN sq.in_date >= CURRENT_DATE - INTERVAL '30 days' THEN '0-30_days'
        WHEN sq.in_date >= CURRENT_DATE - INTERVAL '90 days' THEN '31-90_days'
        WHEN sq.in_date >= CURRENT_DATE - INTERVAL '180 days' THEN '91-180_days'
        WHEN sq.in_date >= CURRENT_DATE - INTERVAL '365 days' THEN '181-365_days'
        ELSE '365+_days'
    END as age_bracket,
    CURRENT_DATE - DATE(sq.in_date) as days_in_stock,
    sq.quantity * p.standard_price as value
FROM stock_quants sq
JOIN products p ON sq.product_id = p.id AND sq.organization_id = p.organization_id
JOIN stock_locations sl ON sq.location_id = sl.id AND sq.organization_id = sl.organization_id
LEFT JOIN stock_lots slots ON sq.lot_id = slots.id AND sq.organization_id = slots.organization_id
WHERE sq.quantity > 0 AND sq.deleted_at IS NULL
ORDER BY p.organization_id, p.id, days_in_stock DESC;

-- Create dead stock analysis view (PER ORGANIZATION)
CREATE OR REPLACE VIEW inventory_dead_stock AS
SELECT
    p.organization_id,
    p.id as product_id,
    p.name as product_name,
    p.default_code,
    p.category_id,
    p.last_movement_date,
    p.days_since_movement,
    SUM(sq.quantity) as total_quantity,
    SUM(sq.quantity) * p.standard_price as total_value,
    CASE
        WHEN p.days_since_movement IS NULL THEN 'never_moved'
        WHEN p.days_since_movement >= 365 THEN '365+_days'
        WHEN p.days_since_movement >= 180 THEN '180-364_days'
        WHEN p.days_since_movement >= 90 THEN '90-179_days'
        WHEN p.days_since_movement >= 30 THEN '30-89_days'
        ELSE 'active'
    END as dead_stock_category
FROM products p
JOIN stock_quants sq ON p.id = sq.product_id AND p.organization_id = sq.organization_id
WHERE sq.quantity > 0 AND sq.deleted_at IS NULL
GROUP BY p.organization_id, p.id
HAVING p.days_since_movement >= 30 OR p.days_since_movement IS NULL
ORDER BY total_value DESC;

-- Create inventory movement summary view (PER ORGANIZATION)
CREATE OR REPLACE VIEW inventory_movement_summary AS
SELECT
    sm.organization_id,
    DATE_TRUNC('month', sm.date) as month,
    p.id as product_id,
    p.name as product_name,
    p.category_id,
    sl.name as location_name,
    COUNT(sm.id) as move_count,
    SUM(sm.product_uom_qty) as total_quantity,
    SUM(sm.product_uom_qty * p.standard_price) as total_value,
    AVG(sm.product_uom_qty) as avg_move_quantity
FROM stock_moves sm
JOIN products p ON sm.product_id = p.id AND sm.organization_id = p.organization_id
JOIN stock_locations sl ON sm.location_id = sl.id AND sm.organization_id = sl.organization_id
WHERE sm.state = 'done' AND sm.deleted_at IS NULL
GROUP BY sm.organization_id, DATE_TRUNC('month', sm.date), p.id, sl.id
ORDER BY sm.organization_id, month DESC, total_value DESC;

-- Create inventory reorder analysis view (PER ORGANIZATION)
CREATE OR REPLACE VIEW inventory_reorder_analysis AS
SELECT
    p.organization_id,
    p.id as product_id,
    p.name as product_name,
    p.default_code,
    p.category_id,
    SUM(sq.quantity) as current_stock,
    p.reorder_point,
    p.safety_stock,
    p.lead_time_days,
    -- Calculate daily consumption (30-day average)
    COALESCE(
        (SELECT SUM(sm.product_uom_qty) / 30
         FROM stock_moves sm
         WHERE sm.product_id = p.id
         AND sm.organization_id = p.organization_id
         AND sm.state = 'done'
         AND sm.date >= CURRENT_DATE - INTERVAL '30 days'
         AND sm.deleted_at IS NULL),
        0
    ) as daily_consumption,
    -- Days until reorder point
    CASE
        WHEN daily_consumption > 0
        THEN (SUM(sq.quantity) - p.reorder_point) / NULLIF(daily_consumption, 0)
        ELSE 999
    END as days_until_reorder,
    -- Reorder recommendation
    CASE
        WHEN SUM(sq.quantity) <= p.reorder_point THEN 'reorder_now'
        WHEN (SUM(sq.quantity) - p.reorder_point) / NULLIF(daily_consumption, 0) <= p.lead_time_days THEN 'reorder_soon'
        ELSE 'stock_ok'
    END as reorder_status,
    -- Recommended order quantity
    CASE
        WHEN SUM(sq.quantity) <= p.reorder_point
        THEN GREATEST(p.reorder_point - SUM(sq.quantity) + p.safety_stock, 0)
        ELSE 0
    END as recommended_order_quantity
FROM products p
JOIN stock_quants sq ON p.id = sq.product_id AND p.organization_id = sq.organization_id
WHERE sq.deleted_at IS NULL
GROUP BY p.organization_id, p.id
HAVING SUM(sq.quantity) > 0
ORDER BY
    CASE reorder_status
        WHEN 'reorder_now' THEN 1
        WHEN 'reorder_soon' THEN 2
        ELSE 3
    END,
    days_until_reorder;

-- Create indexes for analytics views
CREATE INDEX idx_inventory_valuation_org ON inventory_valuation(organization_id);
CREATE INDEX idx_inventory_turnover_org ON inventory_turnover(organization_id);
CREATE INDEX idx_inventory_aging_org ON inventory_aging(organization_id);
CREATE INDEX idx_inventory_dead_stock_org ON inventory_dead_stock(organization_id);
CREATE INDEX idx_inventory_movement_summary_org ON inventory_movement_summary(organization_id);
CREATE INDEX idx_inventory_reorder_analysis_org ON inventory_reorder_analysis(organization_id);

-- Create function to refresh analytics for a specific organization
CREATE OR REPLACE FUNCTION refresh_organization_analytics(org_id uuid)
RETURNS void AS $$
BEGIN
    -- Refresh materialized views for the specific organization
    -- Note: In PostgreSQL, we can't refresh materialized views by organization
    -- So we'll use regular views with proper indexing
    -- This function can be used to trigger recalculations if needed

    -- Update last movement dates for products
    UPDATE products p
    SET last_movement_date = (
        SELECT MAX(sm.date)
        FROM stock_moves sm
        WHERE sm.product_id = p.id
        AND sm.organization_id = p.organization_id
        AND sm.state = 'done'
        AND sm.deleted_at IS NULL
    )
    WHERE p.organization_id = org_id
    AND EXISTS (
        SELECT 1 FROM stock_moves sm
        WHERE sm.product_id = p.id
        AND sm.organization_id = p.organization_id
        AND sm.state = 'done'
        AND sm.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql;

-- Create function to get inventory snapshot for an organization
CREATE OR REPLACE FUNCTION get_inventory_snapshot(org_id uuid)
RETURNS TABLE (
    product_id uuid,
    product_name varchar,
    category_id uuid,
    current_stock numeric,
    reorder_point numeric,
    safety_stock numeric,
    reorder_status varchar,
    days_until_reorder integer,
    current_value numeric,
    retail_value numeric
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id as product_id,
        p.name as product_name,
        p.category_id,
        SUM(sq.quantity) as current_stock,
        p.reorder_point,
        p.safety_stock,
        CASE
            WHEN SUM(sq.quantity) <= p.reorder_point THEN 'reorder_now'
            WHEN (SUM(sq.quantity) - p.reorder_point) / NULLIF(
                (SELECT COALESCE(SUM(sm.product_uom_qty) / 30, 0)
                 FROM stock_moves sm
                 WHERE sm.product_id = p.id
                 AND sm.organization_id = p.organization_id
                 AND sm.state = 'done'
                 AND sm.date >= CURRENT_DATE - INTERVAL '30 days'
                 AND sm.deleted_at IS NULL), 0) <= p.lead_time_days THEN 'reorder_soon'
            ELSE 'stock_ok'
        END as reorder_status,
        CASE
            WHEN (SELECT COALESCE(SUM(sm.product_uom_qty) / 30, 0)
                  FROM stock_moves sm
                  WHERE sm.product_id = p.id
                  AND sm.organization_id = p.organization_id
                  AND sm.state = 'done'
                  AND sm.date >= CURRENT_DATE - INTERVAL '30 days'
                  AND sm.deleted_at IS NULL) > 0
            THEN (SUM(sq.quantity) - p.reorder_point) / NULLIF(
                (SELECT COALESCE(SUM(sm.product_uom_qty) / 30, 0)
                 FROM stock_moves sm
                 WHERE sm.product_id = p.id
                 AND sm.organization_id = p.organization_id
                 AND sm.state = 'done'
                 AND sm.date >= CURRENT_DATE - INTERVAL '30 days'
                 AND sm.deleted_at IS NULL), 0)
            ELSE 999
        END as days_until_reorder,
        SUM(sq.quantity) * p.standard_price as current_value,
        SUM(sq.quantity) * p.list_price as retail_value
    FROM products p
    JOIN stock_quants sq ON p.id = sq.product_id AND p.organization_id = sq.organization_id
    WHERE p.organization_id = org_id AND sq.deleted_at IS NULL
    GROUP BY p.id
    HAVING SUM(sq.quantity) > 0
    ORDER BY current_value DESC;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Grant permissions for analytics views
-- These will be handled by the RLS policies in the main migration
COMMENT ON VIEW inventory_valuation IS 'Inventory valuation by product - filtered by organization RLS';
COMMENT ON VIEW inventory_turnover IS 'Inventory turnover analysis - filtered by organization RLS';
COMMENT ON VIEW inventory_aging IS 'Stock aging analysis - filtered by organization RLS';
COMMENT ON VIEW inventory_dead_stock IS 'Dead stock identification - filtered by organization RLS';
COMMENT ON VIEW inventory_movement_summary IS 'Inventory movement summary - filtered by organization RLS';
COMMENT ON VIEW inventory_reorder_analysis IS 'Reorder analysis and recommendations - filtered by organization RLS';

-- =====================================================
-- END OF MIGRATION
-- =====================================================
