-- Migration: UI Views with Permission Integration
-- Description: Comprehensive views for frontend UI with built-in permission filtering
-- Created: 2025-01-01
-- Module: UI Views

-- =====================================================
-- HELPER FUNCTIONS FOR VIEWS
-- =====================================================

-- Get current user's organization_id
CREATE OR REPLACE FUNCTION get_current_user_organization_id()
RETURNS uuid
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    RETURN (
        SELECT organization_id
        FROM organization_users
        WHERE user_id = auth.uid()
          AND is_active = true
        LIMIT 1
    );
END;
$$;

COMMENT ON FUNCTION get_current_user_organization_id IS 'Get the current authenticated user''s organization ID';

-- Check if current user can view table
CREATE OR REPLACE FUNCTION can_view_table(p_table_name text)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    RETURN check_table_permission(
        auth.uid(),
        get_current_user_organization_id(),
        p_table_name,
        'select'::permission_action
    );
END;
$$;

COMMENT ON FUNCTION can_view_table IS 'Check if current user has SELECT permission on a table';

-- =====================================================
-- CRM & SALES VIEWS
-- =====================================================

-- Complete contact information with relationships
CREATE OR REPLACE VIEW view_contacts_full AS
SELECT
    c.id,
    c.organization_id,
    c.company_id,
    c.contact_type,
    c.name,
    c.display_name,
    c.email,
    c.phone,
    c.mobile,
    c.fax,
    c.website,
    c.title,
    c.job_position,
    c.is_company,
    c.is_customer,
    c.is_vendor,
    c.is_employee,
    -- Address information
    c.street,
    c.street2,
    c.city,
    c.zip,
    co.name as country_name,
    co.code as country_code,
    s.name as state_name,
    s.code as state_code,
    -- Relationships
    parent.name as parent_company_name,
    parent.id as parent_company_id,
    i.name as industry_name,
    pt.name as payment_term_name,
    st.name as sales_team_name,
    -- Metadata
    c.language,
    c.timezone,
    c.tax_id,
    c.reference,
    c.color,
    c.comment,
    c.created_at,
    c.updated_at,
    c.created_by,
    c.custom_fields,
    -- Computed fields
    CASE
        WHEN c.is_company THEN c.name
        WHEN c.parent_id IS NOT NULL THEN parent.name || ' / ' || c.name
        ELSE c.name
    END as full_name,
    CASE
        WHEN c.email IS NOT NULL THEN true
        ELSE false
    END as has_email,
    CASE
        WHEN c.phone IS NOT NULL OR c.mobile IS NOT NULL THEN true
        ELSE false
    END as has_phone
FROM contacts c
LEFT JOIN countries co ON c.country_id = co.id
LEFT JOIN states s ON c.state_id = s.id
LEFT JOIN contacts parent ON c.parent_id = parent.id
LEFT JOIN industries i ON c.industry_id = i.id
LEFT JOIN payment_terms pt ON c.payment_term_id = pt.id
LEFT JOIN sales_teams st ON c.team_id = st.id
WHERE c.deleted_at IS NULL
  AND c.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_contacts_full IS 'Complete contact information with all relationships - permission filtered';

-- Lead/Opportunity pipeline view
CREATE OR REPLACE VIEW view_leads_pipeline AS
SELECT
    l.id,
    l.organization_id,
    l.company_id,
    l.name,
    l.contact_name,
    l.email,
    l.phone,
    l.mobile,
    l.lead_type,
    l.priority,
    l.active,
    l.won_status,
    -- Financial
    l.expected_revenue,
    l.probability,
    l.recurring_revenue,
    l.recurring_plan,
    -- Stage and team
    ls.name as stage_name,
    ls.sequence as stage_sequence,
    ls.probability as stage_probability,
    ls.is_won as stage_is_won,
    st.name as sales_team_name,
    st.code as sales_team_code,
    -- Contact and source
    c.name as contact_company_name,
    c.email as contact_company_email,
    c.is_customer,
    lsrc.name as source_name,
    utm_m.name as medium_name,
    utm_c.name as campaign_name,
    -- Lost reason
    lr.name as lost_reason,
    -- Dates
    l.date_open,
    l.date_closed,
    l.date_deadline,
    l.date_last_stage_update,
    l.created_at,
    l.updated_at,
    -- Computed fields
    CASE
        WHEN l.date_deadline IS NOT NULL AND l.date_deadline < CURRENT_DATE AND l.won_status != 'won'
        THEN true
        ELSE false
    END as is_overdue,
    CASE
        WHEN l.date_last_stage_update IS NOT NULL
        THEN EXTRACT(EPOCH FROM (now() - l.date_last_stage_update))::integer / 86400
        ELSE NULL
    END as days_in_stage,
    CASE
        WHEN l.probability > 0
        THEN l.expected_revenue * (l.probability / 100.0)
        ELSE 0
    END as weighted_revenue
FROM leads l
LEFT JOIN lead_stages ls ON l.stage_id = ls.id
LEFT JOIN sales_teams st ON l.team_id = st.id
LEFT JOIN contacts c ON l.contact_id = c.id
LEFT JOIN lead_sources lsrc ON l.source_id = lsrc.id
LEFT JOIN utm_mediums utm_m ON l.medium_id = utm_m.id
LEFT JOIN utm_campaigns utm_c ON l.campaign_id = utm_c.id
LEFT JOIN lost_reasons lr ON l.lost_reason_id = lr.id
WHERE l.deleted_at IS NULL
  AND l.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_leads_pipeline IS 'CRM pipeline view with stage, team, and source information - permission filtered';

-- Sales orders summary
CREATE OR REPLACE VIEW view_sales_orders_summary AS
SELECT
    so.id,
    so.organization_id,
    so.company_id,
    so.name,
    so.date_order,
    so.validity_date,
    so.confirmation_date,
    so.state,
    so.invoice_status,
    so.delivery_status,
    -- Amounts
    so.amount_untaxed,
    so.amount_tax,
    so.amount_total,
    so.amount_discount,
    -- Customer information
    c.name as customer_name,
    c.email as customer_email,
    c.phone as customer_phone,
    c.is_company as customer_is_company,
    -- Shipping address
    ship.name as shipping_contact_name,
    ship.street as shipping_street,
    ship.city as shipping_city,
    ship.zip as shipping_zip,
    -- Sales team
    st.name as sales_team_name,
    -- Currency
    curr.name as currency_name,
    curr.symbol as currency_symbol,
    -- Payment
    pt.name as payment_term_name,
    -- References
    so.client_order_ref,
    so.origin,
    -- Metadata
    so.note,
    so.created_at,
    so.updated_at,
    so.created_by,
    -- Computed fields
    (SELECT COUNT(*) FROM sales_order_lines WHERE order_id = so.id AND deleted_at IS NULL) as line_count,
    (SELECT SUM(product_uom_qty) FROM sales_order_lines WHERE order_id = so.id AND deleted_at IS NULL) as total_qty,
    CASE
        WHEN so.state = 'draft' THEN 'Draft'
        WHEN so.state = 'sent' THEN 'Quotation Sent'
        WHEN so.state = 'sale' THEN 'Sales Order'
        WHEN so.state = 'done' THEN 'Locked'
        WHEN so.state = 'cancel' THEN 'Cancelled'
    END as state_label,
    CASE
        WHEN so.invoice_status = 'to invoice' THEN true
        ELSE false
    END as needs_invoice,
    CASE
        WHEN so.delivery_status = 'to deliver' THEN true
        ELSE false
    END as needs_delivery
FROM sales_orders so
JOIN contacts c ON so.partner_id = c.id
LEFT JOIN contacts ship ON so.partner_shipping_id = ship.id
LEFT JOIN sales_teams st ON so.team_id = st.id
LEFT JOIN currencies curr ON so.currency_id = curr.id
LEFT JOIN payment_terms pt ON so.payment_term_id = pt.id
WHERE so.deleted_at IS NULL
  AND so.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_sales_orders_summary IS 'Sales orders with customer and line item counts - permission filtered';

-- =====================================================
-- INVENTORY & PRODUCTS VIEWS
-- =====================================================

-- Complete product information
CREATE OR REPLACE VIEW view_products_full AS
SELECT
    p.id,
    p.organization_id,
    p.name,
    p.display_name,
    p.default_code,
    p.barcode,
    p.product_type,
    -- Category
    pc.name as category_name,
    pc.complete_name as category_path,
    -- Pricing
    p.list_price,
    p.standard_price,
    curr.name as currency_name,
    curr.symbol as currency_symbol,
    -- Sales and purchase
    p.sale_ok,
    p.can_be_sold,
    p.purchase_ok,
    p.can_be_purchased,
    p.invoice_policy,
    -- Inventory
    p.tracking,
    p.weight,
    p.volume,
    -- UoM
    uom.name as uom_name,
    uom_po.name as uom_purchase_name,
    -- Description
    p.description,
    p.description_sale,
    p.description_purchase,
    -- Metadata
    p.image_url,
    p.active,
    p.created_at,
    p.updated_at,
    p.custom_fields,
    -- Computed fields
    CASE
        WHEN p.product_type = 'storable' THEN
            (SELECT COALESCE(SUM(quantity), 0)
             FROM stock_quants sq
             JOIN stock_locations sl ON sq.location_id = sl.id
             WHERE sq.product_id = p.id AND sl.usage = 'internal')
        ELSE NULL
    END as qty_available,
    CASE
        WHEN p.sale_ok AND p.purchase_ok THEN 'Can be Sold & Purchased'
        WHEN p.sale_ok THEN 'Can be Sold'
        WHEN p.purchase_ok THEN 'Can be Purchased'
        ELSE 'Not for Sale/Purchase'
    END as availability_status,
    (p.list_price - p.standard_price) as margin,
    CASE
        WHEN p.standard_price > 0
        THEN ((p.list_price - p.standard_price) / p.standard_price * 100)
        ELSE 0
    END as margin_percentage
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN currencies curr ON p.currency_id = curr.id
LEFT JOIN uom_units uom ON p.uom_id = uom.id
LEFT JOIN uom_units uom_po ON p.uom_po_id = uom_po.id
WHERE p.deleted_at IS NULL
  AND p.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_products_full IS 'Complete product information with pricing and inventory - permission filtered';

-- Real-time inventory levels by location
CREATE OR REPLACE VIEW view_stock_levels AS
SELECT
    sq.organization_id,
    sq.product_id,
    p.name as product_name,
    p.default_code as product_code,
    p.barcode,
    sl.id as location_id,
    sl.name as location_name,
    sl.complete_name as location_path,
    sl.usage as location_usage,
    w.name as warehouse_name,
    w.code as warehouse_code,
    -- Quantities
    SUM(sq.quantity) as on_hand_qty,
    SUM(sq.reserved_quantity) as reserved_qty,
    SUM(sq.quantity - sq.reserved_quantity) as available_qty,
    -- Lot/Serial tracking
    p.tracking,
    lot.name as lot_number,
    lot.expiration_date as lot_expiration,
    -- Value
    p.standard_price,
    SUM(sq.quantity * p.standard_price) as total_value,
    -- Metadata
    MIN(sq.in_date) as oldest_stock_date,
    MAX(sq.in_date) as newest_stock_date
FROM stock_quants sq
JOIN products p ON sq.product_id = p.id
JOIN stock_locations sl ON sq.location_id = sl.id
LEFT JOIN warehouses w ON sl.id = w.lot_stock_id
LEFT JOIN stock_lots lot ON sq.lot_id = lot.id
WHERE sl.usage = 'internal'
  AND sq.organization_id = get_current_user_organization_id()
GROUP BY
    sq.organization_id, sq.product_id, p.name, p.default_code, p.barcode,
    sl.id, sl.name, sl.complete_name, sl.usage,
    w.name, w.code, p.tracking, lot.name, lot.expiration_date, p.standard_price
HAVING SUM(sq.quantity) != 0;

COMMENT ON VIEW view_stock_levels IS 'Real-time inventory quantities by product and location - permission filtered';

-- Stock transfer details
CREATE OR REPLACE VIEW view_stock_pickings_detailed AS
SELECT
    sp.id,
    sp.organization_id,
    sp.company_id,
    sp.name,
    sp.state,
    sp.priority,
    -- Dates
    sp.date,
    sp.scheduled_date,
    sp.date_deadline,
    sp.date_done,
    -- Operation type
    spt.name as operation_type,
    spt.code as operation_code,
    spt.color as operation_color,
    -- Locations
    src.name as source_location,
    src.complete_name as source_location_path,
    dest.name as dest_location,
    dest.complete_name as dest_location_path,
    -- Partner
    partner.name as partner_name,
    partner.email as partner_email,
    -- Warehouse
    w.name as warehouse_name,
    -- Origin
    sp.origin,
    sp.note,
    -- Metadata
    sp.created_at,
    sp.updated_at,
    -- Computed fields
    (SELECT COUNT(*) FROM stock_moves WHERE picking_id = sp.id AND deleted_at IS NULL) as move_count,
    (SELECT SUM(product_uom_qty) FROM stock_moves WHERE picking_id = sp.id AND deleted_at IS NULL) as total_qty,
    CASE
        WHEN sp.state = 'draft' THEN 'Draft'
        WHEN sp.state = 'waiting' THEN 'Waiting'
        WHEN sp.state = 'confirmed' THEN 'Confirmed'
        WHEN sp.state = 'assigned' THEN 'Ready'
        WHEN sp.state = 'done' THEN 'Done'
        WHEN sp.state = 'cancel' THEN 'Cancelled'
    END as state_label,
    CASE
        WHEN sp.date_deadline IS NOT NULL AND sp.date_deadline < CURRENT_DATE AND sp.state NOT IN ('done', 'cancel')
        THEN true
        ELSE false
    END as is_late
FROM stock_pickings sp
JOIN stock_picking_types spt ON sp.picking_type_id = spt.id
JOIN stock_locations src ON sp.location_id = src.id
JOIN stock_locations dest ON sp.location_dest_id = dest.id
LEFT JOIN contacts partner ON sp.partner_id = partner.id
LEFT JOIN warehouses w ON spt.warehouse_id = w.id
WHERE sp.deleted_at IS NULL
  AND sp.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_stock_pickings_detailed IS 'Stock transfers with source/destination and move counts - permission filtered';

-- =====================================================
-- ACCOUNTING & FINANCE VIEWS
-- =====================================================

-- Complete invoice information
CREATE OR REPLACE VIEW view_invoices_full AS
SELECT
    i.id,
    i.organization_id,
    i.company_id,
    i.name,
    i.move_type,
    i.state,
    i.payment_state,
    -- Dates
    i.date,
    i.invoice_date,
    i.invoice_date_due,
    -- Amounts
    i.amount_untaxed,
    i.amount_tax,
    i.amount_total,
    i.amount_residual,
    -- Partner information
    partner.name as customer_name,
    partner.email as customer_email,
    partner.phone as customer_phone,
    partner.tax_id as customer_tax_id,
    -- Journal
    j.name as journal_name,
    j.code as journal_code,
    j.type as journal_type,
    -- Currency
    curr.name as currency_name,
    curr.symbol as currency_symbol,
    -- Payment terms
    pt.name as payment_term_name,
    -- References
    i.ref,
    i.invoice_origin,
    i.invoice_payment_ref,
    -- Metadata
    i.narration,
    i.created_at,
    i.updated_at,
    i.created_by,
    -- Computed fields
    (SELECT COUNT(*) FROM invoice_lines WHERE move_id = i.id AND deleted_at IS NULL) as line_count,
    CASE
        WHEN i.move_type = 'out_invoice' THEN 'Customer Invoice'
        WHEN i.move_type = 'out_refund' THEN 'Customer Credit Note'
        WHEN i.move_type = 'in_invoice' THEN 'Vendor Bill'
        WHEN i.move_type = 'in_refund' THEN 'Vendor Refund'
        WHEN i.move_type = 'entry' THEN 'Journal Entry'
    END as type_label,
    CASE
        WHEN i.payment_state = 'paid' THEN 'green'
        WHEN i.payment_state = 'partial' THEN 'orange'
        WHEN i.payment_state = 'not_paid' AND i.invoice_date_due < CURRENT_DATE THEN 'red'
        WHEN i.payment_state = 'not_paid' THEN 'gray'
    END as payment_status_color,
    CASE
        WHEN i.invoice_date_due IS NOT NULL AND i.invoice_date_due < CURRENT_DATE AND i.payment_state != 'paid'
        THEN (CURRENT_DATE - i.invoice_date_due)::integer
        ELSE 0
    END as days_overdue,
    CASE
        WHEN i.move_type IN ('out_invoice', 'in_refund') THEN i.amount_total
        ELSE -i.amount_total
    END as signed_amount
FROM invoices i
JOIN contacts partner ON i.partner_id = partner.id
JOIN account_journals j ON i.journal_id = j.id
LEFT JOIN currencies curr ON i.currency_id = curr.id
LEFT JOIN payment_terms pt ON i.invoice_payment_term_id = pt.id
WHERE i.deleted_at IS NULL
  AND i.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_invoices_full IS 'Complete invoice information with partner and payment details - permission filtered';

-- Payments with invoice allocations
CREATE OR REPLACE VIEW view_payments_with_allocations AS
SELECT
    p.id,
    p.organization_id,
    p.company_id,
    p.name,
    p.payment_date,
    p.payment_type,
    p.partner_type,
    p.amount,
    p.state,
    -- Partner
    partner.name as partner_name,
    partner.email as partner_email,
    -- Journal
    j.name as journal_name,
    j.type as journal_type,
    -- Currency
    curr.name as currency_name,
    curr.symbol as currency_symbol,
    -- Payment method
    pm.name as payment_method_name,
    -- References
    p.communication,
    p.ref,
    -- Metadata
    p.created_at,
    p.updated_at,
    -- Allocated invoices
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'invoice_id', inv.id,
                'invoice_name', inv.name,
                'allocated_amount', pia.amount
            ) ORDER BY pia.created_at
        )
        FROM payment_invoice_allocation pia
        JOIN invoices inv ON pia.invoice_id = inv.id
        WHERE pia.payment_id = p.id
    ) as allocated_invoices,
    (SELECT COALESCE(SUM(amount), 0) FROM payment_invoice_allocation WHERE payment_id = p.id) as total_allocated,
    -- Computed fields
    CASE
        WHEN p.payment_type = 'inbound' THEN 'Receive Money'
        WHEN p.payment_type = 'outbound' THEN 'Send Money'
    END as payment_type_label,
    p.amount - (SELECT COALESCE(SUM(amount), 0) FROM payment_invoice_allocation WHERE payment_id = p.id) as unallocated_amount
FROM payments p
LEFT JOIN contacts partner ON p.partner_id = partner.id
LEFT JOIN account_journals j ON p.journal_id = j.id
LEFT JOIN currencies curr ON p.currency_id = curr.id
LEFT JOIN payment_methods pm ON p.payment_method_id = pm.id
WHERE p.deleted_at IS NULL
  AND p.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_payments_with_allocations IS 'Payments with invoice allocation details - permission filtered';

-- Account Receivable Aging Report
CREATE OR REPLACE VIEW view_ar_aging AS
SELECT
    partner.id as partner_id,
    partner.name as customer_name,
    partner.email as customer_email,
    partner.phone as customer_phone,
    -- Aging buckets
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due <= 0 THEN i.amount_residual ELSE 0 END) as current,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due BETWEEN 1 AND 30 THEN i.amount_residual ELSE 0 END) as days_1_30,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due BETWEEN 31 AND 60 THEN i.amount_residual ELSE 0 END) as days_31_60,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due BETWEEN 61 AND 90 THEN i.amount_residual ELSE 0 END) as days_61_90,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due > 90 THEN i.amount_residual ELSE 0 END) as over_90,
    SUM(i.amount_residual) as total_due,
    -- Metadata
    COUNT(i.id) as invoice_count,
    MIN(i.invoice_date_due) as oldest_due_date,
    MAX(i.invoice_date_due) as newest_due_date
FROM invoices i
JOIN contacts partner ON i.partner_id = partner.id
WHERE i.move_type = 'out_invoice'
  AND i.state = 'posted'
  AND i.payment_state IN ('not_paid', 'partial')
  AND i.deleted_at IS NULL
  AND i.organization_id = get_current_user_organization_id()
GROUP BY partner.id, partner.name, partner.email, partner.phone
HAVING SUM(i.amount_residual) > 0
ORDER BY SUM(i.amount_residual) DESC;

COMMENT ON VIEW view_ar_aging IS 'Accounts receivable aging report by customer - permission filtered';

-- Account Payable Aging Report
CREATE OR REPLACE VIEW view_ap_aging AS
SELECT
    vendor.id as vendor_id,
    vendor.name as vendor_name,
    vendor.email as vendor_email,
    vendor.phone as vendor_phone,
    -- Aging buckets
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due <= 0 THEN i.amount_residual ELSE 0 END) as current,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due BETWEEN 1 AND 30 THEN i.amount_residual ELSE 0 END) as days_1_30,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due BETWEEN 31 AND 60 THEN i.amount_residual ELSE 0 END) as days_31_60,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due BETWEEN 61 AND 90 THEN i.amount_residual ELSE 0 END) as days_61_90,
    SUM(CASE WHEN CURRENT_DATE - i.invoice_date_due > 90 THEN i.amount_residual ELSE 0 END) as over_90,
    SUM(i.amount_residual) as total_due,
    -- Metadata
    COUNT(i.id) as bill_count,
    MIN(i.invoice_date_due) as oldest_due_date,
    MAX(i.invoice_date_due) as newest_due_date
FROM invoices i
JOIN contacts vendor ON i.partner_id = vendor.id
WHERE i.move_type = 'in_invoice'
  AND i.state = 'posted'
  AND i.payment_state IN ('not_paid', 'partial')
  AND i.deleted_at IS NULL
  AND i.organization_id = get_current_user_organization_id()
GROUP BY vendor.id, vendor.name, vendor.email, vendor.phone
HAVING SUM(i.amount_residual) > 0
ORDER BY SUM(i.amount_residual) DESC;

COMMENT ON VIEW view_ap_aging IS 'Accounts payable aging report by vendor - permission filtered';

-- =====================================================
-- PURCHASING VIEWS
-- =====================================================

-- Purchase orders summary
CREATE OR REPLACE VIEW view_purchase_orders_summary AS
SELECT
    po.id,
    po.organization_id,
    po.company_id,
    po.name,
    po.date_order,
    po.date_approve,
    po.state,
    po.invoice_status,
    po.receipt_status,
    -- Amounts
    po.amount_untaxed,
    po.amount_tax,
    po.amount_total,
    -- Vendor information
    vendor.name as vendor_name,
    vendor.email as vendor_email,
    vendor.phone as vendor_phone,
    -- Currency
    curr.name as currency_name,
    curr.symbol as currency_symbol,
    -- References
    po.partner_ref as vendor_reference,
    po.origin,
    -- Metadata
    po.notes,
    po.created_at,
    po.updated_at,
    -- Computed fields
    (SELECT COUNT(*) FROM purchase_order_lines WHERE order_id = po.id AND deleted_at IS NULL) as line_count,
    (SELECT SUM(product_qty) FROM purchase_order_lines WHERE order_id = po.id AND deleted_at IS NULL) as total_qty,
    CASE
        WHEN po.state = 'draft' THEN 'Draft'
        WHEN po.state = 'sent' THEN 'RFQ Sent'
        WHEN po.state = 'to approve' THEN 'To Approve'
        WHEN po.state = 'purchase' THEN 'Purchase Order'
        WHEN po.state = 'done' THEN 'Locked'
        WHEN po.state = 'cancel' THEN 'Cancelled'
    END as state_label
FROM purchase_orders po
JOIN contacts vendor ON po.partner_id = vendor.id
LEFT JOIN currencies curr ON po.currency_id = curr.id
WHERE po.deleted_at IS NULL
  AND po.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_purchase_orders_summary IS 'Purchase orders with vendor information - permission filtered';

-- =====================================================
-- MANUFACTURING VIEWS
-- =====================================================

-- Manufacturing orders with details
CREATE OR REPLACE VIEW view_manufacturing_orders_full AS
SELECT
    mo.id,
    mo.organization_id,
    mo.company_id,
    mo.name,
    mo.state,
    mo.origin,
    -- Product
    p.name as product_name,
    p.default_code as product_code,
    mo.product_qty,
    uom.name as uom_name,
    -- BOM
    bom.code as bom_code,
    -- Locations
    src.name as source_location,
    dest.name as dest_location,
    -- Dates
    mo.date_planned_start,
    mo.date_planned_finished,
    mo.date_start,
    mo.date_finished,
    mo.date_deadline,
    -- Progress
    mo.qty_producing,
    mo.reservation_state,
    mo.priority,
    -- Metadata
    mo.created_at,
    mo.updated_at,
    -- Computed fields
    (SELECT COUNT(*) FROM work_orders WHERE production_id = mo.id) as work_order_count,
    (SELECT COUNT(*) FROM work_orders WHERE production_id = mo.id AND state = 'done') as completed_work_orders,
    CASE
        WHEN mo.product_qty > 0 THEN (mo.qty_producing / mo.product_qty * 100)
        ELSE 0
    END as progress_percentage,
    CASE
        WHEN mo.state = 'draft' THEN 'Draft'
        WHEN mo.state = 'confirmed' THEN 'Confirmed'
        WHEN mo.state = 'progress' THEN 'In Progress'
        WHEN mo.state = 'to_close' THEN 'To Close'
        WHEN mo.state = 'done' THEN 'Done'
        WHEN mo.state = 'cancel' THEN 'Cancelled'
    END as state_label
FROM manufacturing_orders mo
JOIN products p ON mo.product_id = p.id
LEFT JOIN uom_units uom ON mo.product_uom_id = uom.id
LEFT JOIN bom_bills bom ON mo.bom_id = bom.id
LEFT JOIN stock_locations src ON mo.location_src_id = src.id
LEFT JOIN stock_locations dest ON mo.location_dest_id = dest.id
WHERE mo.deleted_at IS NULL
  AND mo.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_manufacturing_orders_full IS 'Manufacturing orders with product and progress details - permission filtered';

-- =====================================================
-- HR & EMPLOYEE VIEWS
-- =====================================================

-- Complete employee information
CREATE OR REPLACE VIEW view_employees_full AS
SELECT
    e.id,
    e.organization_id,
    e.company_id,
    e.name,
    e.employee_number,
    e.work_email,
    e.work_phone,
    e.mobile_phone,
    e.job_title,
    e.employment_type,
    -- Position and department
    jp.name as job_position,
    dept.name as department_name,
    dept.complete_name as department_path,
    -- Manager and coach
    mgr.name as manager_name,
    mgr.work_email as manager_email,
    coach.name as coach_name,
    -- Company
    comp.name as company_name,
    -- Dates
    e.date_hired,
    e.date_terminated,
    -- Personal info (permission controlled)
    e.gender,
    e.birthday,
    e.marital,
    -- Contact
    e.work_location,
    contact.name as work_contact_name,
    contact.phone as work_contact_phone,
    -- Metadata
    e.barcode,
    e.image_url,
    e.active,
    e.created_at,
    e.updated_at,
    e.custom_fields,
    -- Computed fields
    CASE
        WHEN e.date_hired IS NOT NULL
        THEN EXTRACT(YEAR FROM AGE(COALESCE(e.date_terminated, CURRENT_DATE), e.date_hired))
        ELSE NULL
    END as years_of_service,
    CASE
        WHEN e.employment_type = 'full_time' THEN 'Full Time'
        WHEN e.employment_type = 'part_time' THEN 'Part Time'
        WHEN e.employment_type = 'contract' THEN 'Contract'
        WHEN e.employment_type = 'intern' THEN 'Intern'
    END as employment_type_label,
    CASE
        WHEN e.date_terminated IS NOT NULL THEN false
        ELSE true
    END as is_currently_employed
FROM employees e
LEFT JOIN job_positions jp ON e.job_id = jp.id
LEFT JOIN departments dept ON e.department_id = dept.id
LEFT JOIN employees mgr ON e.parent_id = mgr.id
LEFT JOIN employees coach ON e.coach_id = coach.id
LEFT JOIN companies comp ON e.company_id = comp.id
LEFT JOIN contacts contact ON e.work_contact_id = contact.id
WHERE e.deleted_at IS NULL
  AND e.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_employees_full IS 'Complete employee information with hierarchy - permission filtered';

-- Timesheet summary
CREATE OR REPLACE VIEW view_timesheets_summary AS
SELECT
    ts.id,
    ts.organization_id,
    ts.company_id,
    ts.date,
    ts.name,
    ts.unit_amount,
    ts.validated,
    -- Employee
    e.name as employee_name,
    e.employee_number,
    dept.name as department_name,
    -- Project and task
    proj.name as project_name,
    t.name as task_name,
    t.planned_hours as task_planned_hours,
    -- Analytic account
    aa.name as analytic_account_name,
    -- Metadata
    ts.created_at,
    ts.updated_at,
    -- Computed fields
    CASE
        WHEN ts.validated THEN 'Validated'
        ELSE 'Draft'
    END as status_label,
    CASE
        WHEN ts.validated THEN 'green'
        ELSE 'gray'
    END as status_color
FROM timesheets ts
JOIN employees e ON ts.employee_id = e.id
LEFT JOIN departments dept ON e.department_id = dept.id
LEFT JOIN projects proj ON ts.project_id = proj.id
LEFT JOIN tasks t ON ts.task_id = t.id
LEFT JOIN analytic_accounts aa ON ts.account_id = aa.id
WHERE ts.deleted_at IS NULL
  AND ts.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_timesheets_summary IS 'Timesheet entries with employee and project details - permission filtered';

-- =====================================================
-- PROJECT MANAGEMENT VIEWS
-- =====================================================

-- Tasks with full context (kanban view)
CREATE OR REPLACE VIEW view_tasks_kanban AS
SELECT
    t.id,
    t.organization_id,
    t.company_id,
    t.name,
    t.description,
    t.priority,
    t.kanban_state,
    t.sequence,
    -- Project and stage
    proj.name as project_name,
    proj.color as project_color,
    stage.name as stage_name,
    stage.sequence as stage_sequence,
    stage.fold as stage_fold,
    -- Customer
    partner.name as customer_name,
    -- Parent task
    parent.name as parent_task_name,
    -- Time tracking
    t.planned_hours,
    t.effective_hours,
    t.remaining_hours,
    t.progress,
    -- Dates
    t.date_deadline,
    t.date_end,
    t.date_assign,
    t.date_last_stage_update,
    -- Metadata
    t.active,
    t.created_at,
    t.updated_at,
    t.custom_fields,
    -- Assigned users (simplified for display)
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', e.id,
                'name', e.name,
                'image_url', e.image_url
            )
        )
        FROM employees e
        WHERE e.user_id = ANY(t.user_ids)
    ) as assigned_users,
    -- Computed fields
    CASE
        WHEN t.date_deadline IS NOT NULL AND t.date_deadline < CURRENT_DATE AND t.kanban_state != 'done'
        THEN true
        ELSE false
    END as is_overdue,
    CASE
        WHEN t.planned_hours > 0 THEN (t.effective_hours / t.planned_hours * 100)
        ELSE 0
    END as time_spent_percentage,
    CASE
        WHEN t.priority = '0' THEN 'Normal'
        WHEN t.priority = '1' THEN 'Low'
        WHEN t.priority = '2' THEN 'High'
        WHEN t.priority = '3' THEN 'Urgent'
        ELSE 'Normal'
    END as priority_label,
    CASE
        WHEN t.kanban_state = 'normal' THEN 'In Progress'
        WHEN t.kanban_state = 'done' THEN 'Done'
        WHEN t.kanban_state = 'blocked' THEN 'Blocked'
    END as status_label,
    (SELECT COUNT(*) FROM timesheets WHERE task_id = t.id) as timesheet_count,
    (SELECT COALESCE(SUM(unit_amount), 0) FROM timesheets WHERE task_id = t.id) as total_hours_logged
FROM tasks t
JOIN projects proj ON t.project_id = proj.id
LEFT JOIN task_stages stage ON t.stage_id = stage.id
LEFT JOIN contacts partner ON t.partner_id = partner.id
LEFT JOIN tasks parent ON t.parent_id = parent.id
WHERE t.deleted_at IS NULL
  AND t.active = true
  AND t.organization_id = get_current_user_organization_id();

COMMENT ON VIEW view_tasks_kanban IS 'Tasks with full context for kanban board display - permission filtered';

-- =====================================================
-- DASHBOARD & ANALYTICS VIEWS
-- =====================================================

-- Sales dashboard metrics
CREATE OR REPLACE VIEW view_sales_dashboard AS
SELECT
    DATE_TRUNC('month', so.date_order) as month,
    st.name as sales_team,
    st.id as sales_team_id,
    -- Metrics
    COUNT(so.id) as order_count,
    SUM(so.amount_total) as total_revenue,
    AVG(so.amount_total) as avg_order_value,
    COUNT(DISTINCT so.partner_id) as unique_customers,
    SUM(CASE WHEN so.state IN ('sale', 'done') THEN 1 ELSE 0 END) as confirmed_orders,
    SUM(CASE WHEN so.state = 'draft' THEN 1 ELSE 0 END) as draft_orders,
    SUM(CASE WHEN so.invoice_status = 'to invoice' THEN 1 ELSE 0 END) as orders_to_invoice,
    -- Computed
    SUM(so.amount_total) FILTER (WHERE so.state IN ('sale', 'done')) as confirmed_revenue,
    COUNT(DISTINCT so.partner_id) FILTER (WHERE so.state IN ('sale', 'done')) as paying_customers
FROM sales_orders so
LEFT JOIN sales_teams st ON so.team_id = st.id
WHERE so.deleted_at IS NULL
  AND so.organization_id = get_current_user_organization_id()
  AND so.date_order >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', so.date_order), st.name, st.id
ORDER BY month DESC, total_revenue DESC;

COMMENT ON VIEW view_sales_dashboard IS 'Sales metrics by month and team for dashboards - permission filtered';

-- Inventory valuation
CREATE OR REPLACE VIEW view_inventory_valuation AS
SELECT
    p.id as product_id,
    p.name as product_name,
    p.default_code as product_code,
    pc.name as category_name,
    SUM(sq.quantity) as total_qty,
    p.standard_price as unit_cost,
    p.list_price as sale_price,
    SUM(sq.quantity * p.standard_price) as total_cost_value,
    SUM(sq.quantity * p.list_price) as total_sale_value,
    SUM(sq.quantity * (p.list_price - p.standard_price)) as potential_margin,
    -- Location breakdown
    jsonb_object_agg(
        sl.name,
        jsonb_build_object(
            'quantity', sq.quantity,
            'value', sq.quantity * p.standard_price
        )
    ) as location_breakdown
FROM stock_quants sq
JOIN products p ON sq.product_id = p.id
LEFT JOIN product_categories pc ON p.category_id = pc.id
JOIN stock_locations sl ON sq.location_id = sl.id
WHERE sl.usage = 'internal'
  AND p.product_type = 'storable'
  AND sq.organization_id = get_current_user_organization_id()
GROUP BY p.id, p.name, p.default_code, pc.name, p.standard_price, p.list_price
HAVING SUM(sq.quantity) > 0
ORDER BY SUM(sq.quantity * p.standard_price) DESC;

COMMENT ON VIEW view_inventory_valuation IS 'Inventory valuation by product and location - permission filtered';

-- =====================================================
-- ROW LEVEL SECURITY FOR VIEWS
-- =====================================================

-- Enable RLS on all views
ALTER VIEW view_contacts_full SET (security_invoker = on);
ALTER VIEW view_leads_pipeline SET (security_invoker = on);
ALTER VIEW view_sales_orders_summary SET (security_invoker = on);
ALTER VIEW view_products_full SET (security_invoker = on);
ALTER VIEW view_stock_levels SET (security_invoker = on);
ALTER VIEW view_stock_pickings_detailed SET (security_invoker = on);
ALTER VIEW view_invoices_full SET (security_invoker = on);
ALTER VIEW view_payments_with_allocations SET (security_invoker = on);
ALTER VIEW view_ar_aging SET (security_invoker = on);
ALTER VIEW view_ap_aging SET (security_invoker = on);
ALTER VIEW view_purchase_orders_summary SET (security_invoker = on);
ALTER VIEW view_manufacturing_orders_full SET (security_invoker = on);
ALTER VIEW view_employees_full SET (security_invoker = on);
ALTER VIEW view_timesheets_summary SET (security_invoker = on);
ALTER VIEW view_tasks_kanban SET (security_invoker = on);
ALTER VIEW view_sales_dashboard SET (security_invoker = on);
ALTER VIEW view_inventory_valuation SET (security_invoker = on);

-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

GRANT SELECT ON view_contacts_full TO authenticated;
GRANT SELECT ON view_leads_pipeline TO authenticated;
GRANT SELECT ON view_sales_orders_summary TO authenticated;
GRANT SELECT ON view_products_full TO authenticated;
GRANT SELECT ON view_stock_levels TO authenticated;
GRANT SELECT ON view_stock_pickings_detailed TO authenticated;
GRANT SELECT ON view_invoices_full TO authenticated;
GRANT SELECT ON view_payments_with_allocations TO authenticated;
GRANT SELECT ON view_ar_aging TO authenticated;
GRANT SELECT ON view_ap_aging TO authenticated;
GRANT SELECT ON view_purchase_orders_summary TO authenticated;
GRANT SELECT ON view_manufacturing_orders_full TO authenticated;
GRANT SELECT ON view_employees_full TO authenticated;
GRANT SELECT ON view_timesheets_summary TO authenticated;
GRANT SELECT ON view_tasks_kanban TO authenticated;
GRANT SELECT ON view_sales_dashboard TO authenticated;
GRANT SELECT ON view_inventory_valuation TO authenticated;

GRANT EXECUTE ON FUNCTION get_current_user_organization_id() TO authenticated;
GRANT EXECUTE ON FUNCTION can_view_table(text) TO authenticated;

-- =====================================================
-- INDEXES FOR BETTER VIEW PERFORMANCE
-- =====================================================

-- These indexes support the views' WHERE clauses and JOINs
-- They're already mostly covered by existing indexes, but adding a few more:

CREATE INDEX IF NOT EXISTS idx_sales_orders_date_order_org
    ON sales_orders(organization_id, date_order DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_invoices_date_org
    ON invoices(organization_id, invoice_date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_stock_quants_org_product_location
    ON stock_quants(organization_id, product_id, location_id);

CREATE INDEX IF NOT EXISTS idx_timesheets_org_date
    ON timesheets(organization_id, date DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_org_active
    ON tasks(organization_id, active, project_id)
    WHERE deleted_at IS NULL;

-- =====================================================
-- USAGE EXAMPLES
-- =====================================================

COMMENT ON SCHEMA public IS 'UI Views Usage Examples:

-- Frontend: Fetch all contacts with full details
SELECT * FROM view_contacts_full
WHERE is_customer = true
ORDER BY name;

-- Frontend: CRM Pipeline
SELECT * FROM view_leads_pipeline
WHERE stage_name = ''Qualified''
ORDER BY weighted_revenue DESC;

-- Frontend: Sales orders needing action
SELECT * FROM view_sales_orders_summary
WHERE needs_invoice = true OR needs_delivery = true;

-- Frontend: Low stock alerts
SELECT * FROM view_stock_levels
WHERE available_qty < 10
ORDER BY product_name;

-- Frontend: Overdue invoices
SELECT * FROM view_invoices_full
WHERE payment_state != ''paid''
  AND days_overdue > 0
ORDER BY days_overdue DESC;

-- Frontend: AR Aging Report
SELECT * FROM view_ar_aging
ORDER BY total_due DESC;

-- Frontend: Task Kanban Board
SELECT * FROM view_tasks_kanban
WHERE project_name = ''Website Redesign''
ORDER BY stage_sequence, sequence;

-- Frontend: Sales Dashboard
SELECT * FROM view_sales_dashboard
WHERE month >= CURRENT_DATE - INTERVAL ''6 months''
ORDER BY month DESC;

-- Frontend: Inventory Valuation
SELECT * FROM view_inventory_valuation
WHERE category_name = ''Electronics''
ORDER BY total_cost_value DESC;

Note: All views automatically filter by the current user''s organization
and respect the permission system through security_invoker setting.
';
