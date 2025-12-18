-- =====================================================
-- QUOTE/PROPOSAL ENHANCEMENTS
-- =====================================================
-- Description: Extends sales_orders table for quote functionality
-- Approach: Use existing sales_orders table (already has 'quotation' status)
-- This follows Odoo's pattern where quotes and orders are the same entity
-- =====================================================

-- Add quote-specific columns to sales_orders table
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_template varchar(100) DEFAULT 'standard';
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_sent_at timestamptz;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_viewed_at timestamptz;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_accepted_at timestamptz;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_expired_at timestamptz;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_access_token varchar(100);
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_public_url varchar(500);
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS quote_pdf_attachment_id uuid REFERENCES attachments(id);
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS validity_days integer DEFAULT 30;

-- Add index for quote access tokens
CREATE INDEX IF NOT EXISTS idx_sales_orders_quote_token ON sales_orders(quote_access_token) WHERE quote_access_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sales_orders_quote_status ON sales_orders(status) WHERE status = 'quotation';

-- Function to generate quote access token
CREATE OR REPLACE FUNCTION generate_quote_access_token()
RETURNS varchar AS $$
DECLARE
    token varchar;
    exists boolean;
BEGIN
    LOOP
        -- Generate random 32-character token
        token := encode(gen_random_bytes(24), 'base64');
        token := replace(replace(replace(token, '+', ''), '/', ''), '=', '');
        token := substring(token, 1, 32);

        -- Check if token already exists
        SELECT EXISTS(SELECT 1 FROM sales_orders WHERE quote_access_token = token) INTO exists;
        EXIT WHEN NOT exists;
    END LOOP;

    RETURN token;
END;
$$ LANGUAGE plpgsql;

-- Function to mark quote as sent
CREATE OR REPLACE FUNCTION mark_quote_sent(
    p_order_id uuid,
    p_access_token varchar DEFAULT NULL
) RETURNS void AS $$
DECLARE
    v_token varchar;
    v_validity_days integer;
    v_expired_at timestamptz;
BEGIN
    -- Get or generate access token
    IF p_access_token IS NULL THEN
        v_token := generate_quote_access_token();
    ELSE
        v_token := p_access_token;
    END IF;

    -- Get validity days
    SELECT validity_days INTO v_validity_days FROM sales_orders WHERE id = p_order_id;
    IF v_validity_days IS NULL THEN
        v_validity_days := 30;
    END IF;

    -- Calculate expiry date
    v_expired_at := now() + (v_validity_days || ' days')::interval;

    -- Update quote
    UPDATE sales_orders
    SET quote_sent_at = now(),
        quote_access_token = v_token,
        quote_expired_at = v_expired_at,
        updated_at = now()
    WHERE id = p_order_id;
END;
$$ LANGUAGE plpgsql;

-- Function to track quote view
CREATE OR REPLACE FUNCTION track_quote_view(
    p_access_token varchar,
    p_ip_address inet DEFAULT NULL,
    p_user_agent text DEFAULT NULL
) RETURNS uuid AS $$
DECLARE
    v_order_id uuid;
BEGIN
    -- Get order ID from token
    SELECT id INTO v_order_id
    FROM sales_orders
    WHERE quote_access_token = p_access_token
      AND status = 'quotation';

    IF v_order_id IS NULL THEN
        RAISE EXCEPTION 'Invalid or expired quote token';
    END IF;

    -- Update first view timestamp if not set
    UPDATE sales_orders
    SET quote_viewed_at = COALESCE(quote_viewed_at, now()),
        updated_at = now()
    WHERE id = v_order_id;

    -- Log the view (optional - could create a quote_views table for analytics)
    -- For now, just return the order ID

    RETURN v_order_id;
END;
$$ LANGUAGE plpgsql;

-- Function to accept quote (convert to order)
CREATE OR REPLACE FUNCTION accept_quote(
    p_access_token varchar,
    p_accepted_by uuid DEFAULT NULL
) RETURNS uuid AS $$
DECLARE
    v_order_id uuid;
    v_status varchar;
    v_expired_at timestamptz;
BEGIN
    -- Get order and check status
    SELECT id, status, quote_expired_at
    INTO v_order_id, v_status, v_expired_at
    FROM sales_orders
    WHERE quote_access_token = p_access_token;

    IF v_order_id IS NULL THEN
        RAISE EXCEPTION 'Invalid quote token';
    END IF;

    IF v_status != 'quotation' THEN
        RAISE EXCEPTION 'Quote has already been processed';
    END IF;

    IF v_expired_at IS NOT NULL AND v_expired_at < now() THEN
        RAISE EXCEPTION 'Quote has expired';
    END IF;

    -- Update quote to accepted
    UPDATE sales_orders
    SET quote_accepted_at = now(),
        status = 'confirmed',
        confirmation_date = now(),
        updated_at = now()
    WHERE id = v_order_id;

    RETURN v_order_id;
END;
$$ LANGUAGE plpgsql;

-- Function to check if quote is expired
CREATE OR REPLACE FUNCTION check_quote_expiry(p_order_id uuid)
RETURNS boolean AS $$
DECLARE
    v_expired_at timestamptz;
    v_status varchar;
BEGIN
    SELECT quote_expired_at, status
    INTO v_expired_at, v_status
    FROM sales_orders
    WHERE id = p_order_id;

    IF v_status != 'quotation' THEN
        RETURN false; -- Not a quote anymore
    END IF;

    IF v_expired_at IS NULL THEN
        RETURN false; -- No expiry set
    END IF;

    RETURN v_expired_at < now();
END;
$$ LANGUAGE plpgsql;

-- View for quote analytics
CREATE OR REPLACE VIEW quote_analytics AS
SELECT
    organization_id,
    COUNT(*) FILTER (WHERE status = 'quotation') as total_quotes,
    COUNT(*) FILTER (WHERE status = 'quotation' AND quote_sent_at IS NOT NULL) as sent_quotes,
    COUNT(*) FILTER (WHERE status = 'quotation' AND quote_viewed_at IS NOT NULL) as viewed_quotes,
    COUNT(*) FILTER (WHERE status = 'confirmed' AND quote_accepted_at IS NOT NULL) as accepted_quotes,
    COUNT(*) FILTER (WHERE status = 'quotation' AND quote_expired_at < now()) as expired_quotes,
    ROUND(
        COUNT(*) FILTER (WHERE quote_viewed_at IS NOT NULL)::numeric /
        NULLIF(COUNT(*) FILTER (WHERE quote_sent_at IS NOT NULL), 0) * 100,
        2
    ) as view_rate_percent,
    ROUND(
        COUNT(*) FILTER (WHERE status = 'confirmed' AND quote_accepted_at IS NOT NULL)::numeric /
        NULLIF(COUNT(*) FILTER (WHERE quote_sent_at IS NOT NULL), 0) * 100,
        2
    ) as conversion_rate_percent,
    AVG(
        EXTRACT(epoch FROM (quote_accepted_at - quote_sent_at)) / 3600
    ) FILTER (WHERE quote_accepted_at IS NOT NULL AND quote_sent_at IS NOT NULL) as avg_accept_time_hours
FROM sales_orders
WHERE status IN ('quotation', 'confirmed')
GROUP BY organization_id;

-- Comments for documentation
COMMENT ON COLUMN sales_orders.quote_template IS 'Template name for PDF generation (standard, modern, etc.)';
COMMENT ON COLUMN sales_orders.quote_sent_at IS 'When the quote was sent to customer';
COMMENT ON COLUMN sales_orders.quote_viewed_at IS 'First time customer viewed the quote';
COMMENT ON COLUMN sales_orders.quote_accepted_at IS 'When customer accepted the quote';
COMMENT ON COLUMN sales_orders.quote_expired_at IS 'When the quote expires';
COMMENT ON COLUMN sales_orders.quote_access_token IS 'Public access token for viewing quote';
COMMENT ON COLUMN sales_orders.quote_public_url IS 'Full public URL for quote access';
COMMENT ON COLUMN sales_orders.quote_pdf_attachment_id IS 'Reference to generated PDF in attachments table';
COMMENT ON COLUMN sales_orders.validity_days IS 'Number of days quote is valid (default 30)';
