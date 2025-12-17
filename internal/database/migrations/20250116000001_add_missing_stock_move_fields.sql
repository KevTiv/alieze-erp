-- Add missing fields to stock_moves table to support enhanced inventory operations

BEGIN;

-- Add scheduled_date field for planned move timing
ALTER TABLE stock_moves ADD COLUMN IF NOT EXISTS scheduled_date timestamptz;

-- Add quantity field for actual quantity moved (redundant with product_uom_qty but useful for reporting)
ALTER TABLE stock_moves ADD COLUMN IF NOT EXISTS quantity numeric(15,4) DEFAULT 0;

-- Add reserved_quantity field for tracking reserved quantities
ALTER TABLE stock_moves ADD COLUMN IF NOT EXISTS reserved_quantity numeric(15,4) DEFAULT 0;

-- Add product_uom_id as an alias/alternative to product_uom for API compatibility
ALTER TABLE stock_moves ADD COLUMN IF NOT EXISTS product_uom_id uuid;

-- Create index on scheduled_date for performance
CREATE INDEX IF NOT EXISTS idx_stock_moves_scheduled_date ON stock_moves(scheduled_date);

-- Update existing records to set quantity = product_uom_qty for data consistency
UPDATE stock_moves SET quantity = product_uom_qty WHERE quantity = 0;

-- Create a trigger to keep quantity in sync with product_uom_qty
CREATE OR REPLACE FUNCTION sync_stock_move_quantity()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.product_uom_qty IS DISTINCT FROM OLD.product_uom_qty THEN
        NEW.quantity = NEW.product_uom_qty;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_stock_move_quantity ON stock_moves;
CREATE TRIGGER trg_sync_stock_move_quantity
BEFORE UPDATE OR INSERT ON stock_moves
FOR EACH ROW
EXECUTE FUNCTION sync_stock_move_quantity();

COMMIT;
