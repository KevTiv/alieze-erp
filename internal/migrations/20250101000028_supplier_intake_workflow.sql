-- Migration: Supplier Intake Workflow Helper
-- Description: Stored procedure used by n8n to convert supplier deliveries into purchase orders
-- Created: 2025-01-01

CREATE OR REPLACE FUNCTION workflow_supplier_merch_intake(
    p_payload jsonb
)
RETURNS TABLE (
    purchase_order_id uuid,
    purchase_order_number text,
    supplier_id uuid,
    supplier_name text,
    total_items integer,
    total_quantity numeric,
    total_cost numeric,
    products jsonb
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_org_id uuid;
    v_supplier jsonb;
    v_items jsonb;
    v_supplier_id uuid;
    v_purchase_order_id uuid;
    v_purchase_order_number text;
    v_now timestamptz := now();
    v_total_items integer := 0;
    v_total_quantity numeric := 0;
    v_total_cost numeric := 0;
    v_products jsonb := '[]'::jsonb;
    v_item jsonb;
    v_qty numeric;
    v_unit_cost numeric;
    v_list_price numeric;
    v_product_id uuid;
    v_action text;
    v_line_total numeric;
    v_currency_id uuid;
    v_company_id uuid;
    v_reference text;
    v_note text;
BEGIN
    IF p_payload IS NULL THEN
        RAISE EXCEPTION 'payload is required';
    END IF;

    v_org_id := NULLIF(p_payload->>'organization_id', '')::uuid;
    IF v_org_id IS NULL THEN
        RAISE EXCEPTION 'organization_id is required';
    END IF;

    v_supplier := COALESCE(p_payload->'supplier', '{}'::jsonb);
    v_items := COALESCE(p_payload->'items', '[]'::jsonb);

    IF jsonb_typeof(v_items) <> 'array' OR jsonb_array_length(v_items) = 0 THEN
        RAISE EXCEPTION 'items array is required';
    END IF;

    v_reference := NULLIF(p_payload->>'reference', '');
    v_note := COALESCE(p_payload->>'note', '');

    SELECT id
    INTO v_supplier_id
    FROM contacts
    WHERE organization_id = v_org_id
      AND (
        (
            v_supplier->>'name' IS NOT NULL
            AND lower(name) = lower(v_supplier->>'name')
        )
        OR (
            v_supplier->>'email' IS NOT NULL
            AND lower(email) = lower(v_supplier->>'email')
        )
      )
    ORDER BY updated_at DESC
    LIMIT 1;

    IF v_supplier_id IS NULL THEN
        INSERT INTO contacts (
            organization_id,
            contact_type,
            name,
            display_name,
            email,
            phone,
            street,
            city,
            is_company,
            is_vendor,
            comment,
            metadata
        )
        VALUES (
            v_org_id,
            'company',
            COALESCE(v_supplier->>'name', 'Supplier'),
            v_supplier->>'name',
            NULLIF(v_supplier->>'email', ''),
            NULLIF(v_supplier->>'phone', ''),
            NULLIF(v_supplier->>'street', ''),
            NULLIF(v_supplier->>'city', ''),
            true,
            true,
            COALESCE(v_supplier->>'notes', ''),
            jsonb_build_object(
                'source', 'n8n_supplier_intake',
                'payload', v_supplier
            )
        )
        RETURNING id INTO v_supplier_id;
    ELSE
        UPDATE contacts
        SET
            phone = COALESCE(NULLIF(v_supplier->>'phone', ''), phone),
            email = COALESCE(NULLIF(v_supplier->>'email', ''), email),
            street = COALESCE(NULLIF(v_supplier->>'street', ''), street),
            city = COALESCE(NULLIF(v_supplier->>'city', ''), city),
            is_vendor = true,
            updated_at = v_now,
            comment = COALESCE(NULLIF(v_supplier->>'notes', ''), comment),
            metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object(
                'last_n8n_intake', v_now
            )
        WHERE id = v_supplier_id;
    END IF;

    BEGIN
        v_purchase_order_number := generate_sequence_number('purchase_order', v_org_id, NULL);
    EXCEPTION
        WHEN others THEN
            v_purchase_order_number := 'PO-' || to_char(v_now, 'YYYYMMDDHH24MISS');
    END;

    v_currency_id := NULLIF(p_payload->>'currency_id', '')::uuid;
    v_company_id := NULLIF(p_payload->>'company_id', '')::uuid;

    INSERT INTO purchase_orders (
        organization_id,
        company_id,
        name,
        state,
        date_order,
        partner_id,
        partner_ref,
        currency_id,
        notes,
        origin,
        metadata
    )
    VALUES (
        v_org_id,
        v_company_id,
        v_purchase_order_number,
        'draft',
        v_now,
        v_supplier_id,
        NULLIF(v_supplier->>'reference', ''),
        v_currency_id,
        v_note,
        v_reference,
        jsonb_build_object(
            'source', 'n8n_supplier_intake',
            'supplier', v_supplier
        )
    )
    RETURNING id INTO v_purchase_order_id;

    FOR v_item IN SELECT value FROM jsonb_array_elements(v_items) AS t(value) LOOP
        v_total_items := v_total_items + 1;
        v_qty := COALESCE((v_item->>'quantity')::numeric, 0);
        IF v_qty <= 0 THEN
            v_qty := 1;
        END IF;

        v_unit_cost := COALESCE((v_item->>'unit_cost')::numeric, 0);
        v_list_price := COALESCE((v_item->>'list_price')::numeric, v_unit_cost);
        v_line_total := v_unit_cost * v_qty;

        v_total_quantity := v_total_quantity + v_qty;
        v_total_cost := v_total_cost + v_line_total;

        SELECT id
        INTO v_product_id
        FROM products
        WHERE organization_id = v_org_id
          AND (
            (
                v_item->>'sku' IS NOT NULL
                AND lower(default_code) = lower(v_item->>'sku')
            )
            OR (
                v_item->>'name' IS NOT NULL
                AND lower(name) = lower(v_item->>'name')
            )
          )
        ORDER BY updated_at DESC
        LIMIT 1;

        IF v_product_id IS NULL THEN
            INSERT INTO products (
                organization_id,
                name,
                display_name,
                default_code,
                barcode,
                product_type,
                list_price,
                standard_price,
                purchase_ok,
                can_be_purchased,
                sale_ok,
                can_be_sold,
                description_purchase,
                description_sale,
                metadata
            )
            VALUES (
                v_org_id,
                COALESCE(v_item->>'name', 'New Item'),
                COALESCE(v_item->>'display_name', v_item->>'name'),
                NULLIF(v_item->>'sku', ''),
                NULLIF(v_item->>'barcode', ''),
                'storable',
                v_list_price,
                v_unit_cost,
                true,
                true,
                true,
                true,
                COALESCE(v_item->>'description', v_item->>'description_purchase'),
                COALESCE(v_item->>'description_sale', v_item->>'description'),
                jsonb_build_object(
                    'supplier_name', v_supplier->>'name',
                    'supplier_sku', v_item->>'supplier_sku',
                    'warehouse_code', v_item->>'warehouse_code',
                    'source', 'n8n_supplier_intake'
                )
            )
            RETURNING id INTO v_product_id;
            v_action := 'inserted';
        ELSE
            UPDATE products
            SET
                list_price = v_list_price,
                standard_price = v_unit_cost,
                description_purchase = COALESCE(v_item->>'description', description_purchase),
                description_sale = COALESCE(v_item->>'description_sale', description_sale),
                metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object(
                    'last_supplier_refresh', v_now,
                    'supplier_name', v_supplier->>'name',
                    'warehouse_code', v_item->>'warehouse_code'
                ),
                updated_at = v_now
            WHERE id = v_product_id;
            v_action := 'updated';
        END IF;

        INSERT INTO purchase_order_lines (
            organization_id,
            order_id,
            name,
            product_id,
            product_qty,
            price_unit,
            price_subtotal,
            price_total,
            state,
            metadata
        )
        VALUES (
            v_org_id,
            v_purchase_order_id,
            COALESCE(v_item->>'line_name', v_item->>'name'),
            v_product_id,
            v_qty,
            v_unit_cost,
            v_line_total,
            v_line_total,
            'draft',
            jsonb_build_object(
                'supplier_sku', v_item->>'sku',
                'warehouse_code', v_item->>'warehouse_code',
                'unit_cost', v_unit_cost
            )
        );

        v_products := v_products || jsonb_build_array(
            jsonb_build_object(
                'product_id', v_product_id,
                'name', v_item->>'name',
                'sku', v_item->>'sku',
                'action', v_action,
                'quantity', v_qty,
                'unit_cost', v_unit_cost
            )
        );
    END LOOP;

    UPDATE purchase_orders
    SET
        amount_untaxed = v_total_cost,
        amount_total = v_total_cost,
        amount_tax = 0,
        metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object(
            'intake_total_items', v_total_items,
            'intake_total_quantity', v_total_quantity,
            'intake_source', 'n8n_supplier_intake'
        )
    WHERE id = v_purchase_order_id;

    RETURN QUERY
    SELECT
        v_purchase_order_id,
        v_purchase_order_number,
        v_supplier_id,
        COALESCE(v_supplier->>'name', ''),
        v_total_items,
        v_total_quantity,
        v_total_cost,
        v_products;
END;
$$;
