-- Migration: AI Agent Triggers
-- Description: Creates automatic triggers for AI agents to be proactive
-- Created: 2025-01-01
-- Purpose: Make AI agents automatically respond to business events

-- ============================================
-- SALES AGENT TRIGGERS (Alex - Sales Assistant)
-- ============================================

-- Trigger: Create follow-up task when lead status changes
CREATE OR REPLACE FUNCTION create_sales_followup_task()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_sales_agent_id uuid;
    v_days_since_contact integer;
BEGIN
    -- Only for status changes to 'ongoing'
    IF NEW.status = 'ongoing' AND (OLD.status IS DISTINCT FROM NEW.status OR OLD.status IS NULL) THEN
        -- Get sales agent ID
        SELECT id INTO v_sales_agent_id
        FROM ai_team_members
        WHERE organization_id = NEW.organization_id
          AND role = 'sales_assistant'
        LIMIT 1;
        
        -- Calculate days since last contact
        IF NEW.last_contact_date IS NOT NULL THEN
            v_days_since_contact := CURRENT_DATE - NEW.last_contact_date::date;
        ELSE
            v_days_since_contact := 30; -- Default if no contact date
        END IF;
        
        -- Create follow-up task if no recent contact
        IF v_days_since_contact > 3 THEN
            INSERT INTO ai_agent_tasks (
                organization_id, agent_id, task_type, title, description,
                related_entity_type, related_entity_id, priority, scheduled_for
            ) VALUES (
                NEW.organization_id,
                v_sales_agent_id,
                'reminder',
                'Follow up with lead: ' || NEW.name,
                'Lead ' || NEW.name || ' status changed to ongoing. Last contact was ' || v_days_since_contact || ' days ago.',
                'lead',
                NEW.id,
                7,
                CASE 
                    WHEN v_days_since_contact > 14 THEN CURRENT_TIMESTAMP + INTERVAL '1 hour'
                    WHEN v_days_since_contact > 7 THEN CURRENT_TIMESTAMP + INTERVAL '6 hours'
                    ELSE CURRENT_TIMESTAMP + INTERVAL '24 hours'
                END
            );
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_sales_followup
    AFTER UPDATE ON leads
    FOR EACH ROW
    EXECUTE FUNCTION create_sales_followup_task();

-- Trigger: Create task for high-value leads
CREATE OR REPLACE FUNCTION create_high_value_lead_task()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_sales_agent_id uuid;
BEGIN
    -- Check if this is a high-value lead (expected revenue > $5000)
    IF NEW.expected_revenue > 5000 AND NEW.status = 'ongoing' THEN
        -- Get sales agent ID
        SELECT id INTO v_sales_agent_id
        FROM ai_team_members
        WHERE organization_id = NEW.organization_id
          AND role = 'sales_assistant'
        LIMIT 1;
        
        -- Create high-priority task
        INSERT INTO ai_agent_tasks (
            organization_id, agent_id, task_type, title, description,
            related_entity_type, related_entity_id, priority
        ) VALUES (
            NEW.organization_id,
            v_sales_agent_id,
            'alert',
            'High-value lead needs attention: ' || NEW.name || ' ($' || NEW.expected_revenue || ')',
            'High-value lead ' || NEW.name || ' with expected revenue of $' || NEW.expected_revenue || ' needs immediate attention.',
            'lead',
            NEW.id,
            9
        );
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_high_value_lead
    AFTER INSERT OR UPDATE ON leads
    FOR EACH ROW
    EXECUTE FUNCTION create_high_value_lead_task();

-- ============================================
-- INVENTORY AGENT TRIGGERS (Taylor - Inventory Manager)
-- ============================================

-- Trigger: Low stock alert
CREATE OR REPLACE FUNCTION create_low_stock_alert()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_inventory_agent_id uuid;
    v_product_name varchar(255);
BEGIN
    -- Check if stock is below threshold (10 units)
    IF NEW.quantity < 10 THEN
        -- Get product name
        SELECT name INTO v_product_name
        FROM products
        WHERE id = NEW.product_id
        LIMIT 1;
        
        -- Get inventory agent ID
        SELECT id INTO v_inventory_agent_id
        FROM ai_team_members
        WHERE organization_id = NEW.organization_id
          AND role = 'inventory_manager'
        LIMIT 1;
        
        -- Create low stock task
        INSERT INTO ai_agent_tasks (
            organization_id, agent_id, task_type, title, description,
            related_entity_type, related_entity_id, priority
        ) VALUES (
            NEW.organization_id,
            v_inventory_agent_id,
            'alert',
            'Low stock: ' || v_product_name || ' (' || NEW.quantity || ' units)',
            'Product ' || v_product_name || ' is low on stock with only ' || NEW.quantity || ' units remaining.',
            'product',
            NEW.product_id,
            8
        );
        
        -- Also create an insight about reordering
        INSERT INTO ai_agent_insights (
            organization_id, agent_id, insight_type, title, summary,
            confidence_score, supporting_data, suggested_actions
        ) VALUES (
            NEW.organization_id,
            v_inventory_agent_id,
            'recommendation',
            'Reorder suggestion for ' || v_product_name,
            'Product ' || v_product_name || ' is running low. Consider reordering to avoid stockouts.',
            90,
            jsonb_build_object(
                'current_stock', NEW.quantity,
                'product_name', v_product_name,
                'last_30_days_sales', 0  -- Would be populated with actual data
            ),
            jsonb_build_array(
                jsonb_build_object(
                    'action', 'Place reorder for ' || v_product_name,
                    'estimated_quantity', 50,
                    'urgency', 'high'
                )
            )
        );
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_low_stock_alert
    AFTER INSERT OR UPDATE ON stock_quants
    FOR EACH ROW
    EXECUTE FUNCTION create_low_stock_alert();

-- Trigger: Fast-moving product detection
CREATE OR REPLACE FUNCTION detect_fast_moving_products()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_inventory_agent_id uuid;
    v_product_name varchar(255);
    v_sales_count integer;
BEGIN
    -- This would be enhanced with actual sales data analysis
    -- For now, simple example based on stock changes
    
    IF NEW.quantity < (OLD.quantity - 5) AND OLD.quantity IS NOT NULL THEN
        SELECT name INTO v_product_name
        FROM products
        WHERE id = NEW.product_id
        LIMIT 1;
        
        SELECT id INTO v_inventory_agent_id
        FROM ai_team_members
        WHERE organization_id = NEW.organization_id
          AND role = 'inventory_manager'
        LIMIT 1;
        
        INSERT INTO ai_agent_insights (
            organization_id, agent_id, insight_type, title, summary,
            confidence_score, supporting_data
        ) VALUES (
            NEW.organization_id,
            v_inventory_agent_id,
            'trend',
            v_product_name || ' is selling fast!',
            'Product ' || v_product_name || ' has seen significant sales activity recently.',
            75,
            jsonb_build_object(
                'product_name', v_product_name,
                'recent_activity', 'high',
                'current_stock', NEW.quantity
            )
        );
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_fast_moving_product
    AFTER UPDATE ON stock_quants
    FOR EACH ROW
    EXECUTE FUNCTION detect_fast_moving_products();

-- ============================================
-- FINANCIAL AGENT TRIGGERS (Jordan - Financial Advisor)
-- ============================================

-- Trigger: Large expense detection
CREATE OR REPLACE FUNCTION detect_large_expense()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_financial_agent_id uuid;
    v_contact_name varchar(255);
BEGIN
    -- Check if this is a large expense (invoice over $1000)
    IF NEW.move_type = 'in_invoice' AND NEW.amount_total > 1000 THEN
        -- Get contact name
        SELECT name INTO v_contact_name
        FROM contacts
        WHERE id = NEW.partner_id
        LIMIT 1;
        
        -- Get financial agent ID
        SELECT id INTO v_financial_agent_id
        FROM ai_team_members
        WHERE organization_id = NEW.organization_id
          AND role = 'financial_advisor'
        LIMIT 1;
        
        -- Create alert for large expense
        INSERT INTO ai_agent_tasks (
            organization_id, agent_id, task_type, title, description,
            related_entity_type, related_entity_id, priority
        ) VALUES (
            NEW.organization_id,
            v_financial_agent_id,
            'alert',
            'Large expense detected: $' || NEW.amount_total || ' to ' || v_contact_name,
            'Invoice #' || NEW.name || ' for $' || NEW.amount_total || ' to ' || v_contact_name || ' requires review.',
            'invoice',
            NEW.id,
            7
        );
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_large_expense
    AFTER INSERT ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION detect_large_expense();

-- Trigger: Cash flow analysis
CREATE OR REPLACE FUNCTION analyze_cash_flow()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_financial_agent_id uuid;
    v_recent_invoices integer;
    v_recent_payments integer;
BEGIN
    -- Simple cash flow analysis (would be enhanced with real calculations)
    -- For now, just create a periodic insight
    
    IF NEW.move_type = 'out_invoice' AND NEW.amount_total > 500 THEN
        SELECT id INTO v_financial_agent_id
        FROM ai_team_members
        WHERE organization_id = NEW.organization_id
          AND role = 'financial_advisor'
        LIMIT 1;
        
        -- Get count of recent invoices and payments (simplified)
        SELECT COUNT(*) INTO v_recent_invoices
        FROM invoices
        WHERE organization_id = NEW.organization_id
          AND move_type = 'out_invoice'
          AND created_at > CURRENT_TIMESTAMP - INTERVAL '30 days';
        
        SELECT COUNT(*) INTO v_recent_payments
        FROM payments
        WHERE organization_id = NEW.organization_id
          AND created_at > CURRENT_TIMESTAMP - INTERVAL '30 days';
        
        INSERT INTO ai_agent_insights (
            organization_id, agent_id, insight_type, title, summary,
            confidence_score, supporting_data, suggested_actions
        ) VALUES (
            NEW.organization_id,
            v_financial_agent_id,
            'tip',
            'Cash flow update',
            'You have ' || v_recent_invoices || ' invoices sent and ' || v_recent_payments || ' payments received in the last 30 days.',
            80,
            jsonb_build_object(
                'invoices_sent', v_recent_invoices,
                'payments_received', v_recent_payments,
                'outstanding_invoices', v_recent_invoices - v_recent_payments
            ),
            jsonb_build_array(
                jsonb_build_object(
                    'action', 'Review outstanding invoices',
                    'priority', 'medium'
                )
            )
        );
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_cash_flow_analysis
    AFTER INSERT ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION analyze_cash_flow();

-- ============================================
-- BUSINESS STRATEGIST TRIGGERS (Casey - Business Strategist)
-- ============================================

-- Trigger: Customer purchase pattern detection
CREATE OR REPLACE FUNCTION detect_purchase_patterns()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_strategist_agent_id uuid;
    v_customer_name varchar(255);
    v_recent_orders integer;
BEGIN
    -- Simple pattern detection (would be enhanced with real analysis)
    -- For now, detect customers with multiple recent orders
    
    IF NEW.state = 'sale' THEN
        SELECT name INTO v_customer_name
        FROM contacts
        WHERE id = NEW.partner_id
        LIMIT 1;
        
        -- Count recent orders for this customer
        SELECT COUNT(*) INTO v_recent_orders
        FROM sales_orders
        WHERE partner_id = NEW.partner_id
          AND created_at > CURRENT_TIMESTAMP - INTERVAL '60 days'
          AND state = 'sale';
        
        -- If customer has multiple recent orders, suggest upsell
        IF v_recent_orders >= 2 THEN
            SELECT id INTO v_strategist_agent_id
            FROM ai_team_members
            WHERE organization_id = NEW.organization_id
              AND role = 'business_strategist'
            LIMIT 1;
            
            INSERT INTO ai_agent_insights (
                organization_id, agent_id, insight_type, title, summary,
                confidence_score, supporting_data, suggested_actions
            ) VALUES (
                NEW.organization_id,
                v_strategist_agent_id,
                'opportunity',
                'Upsell opportunity: ' || v_customer_name,
                v_customer_name || ' has placed ' || v_recent_orders || ' orders in the last 60 days. Consider upsell opportunities.',
                85,
                jsonb_build_object(
                    'customer_name', v_customer_name,
                    'recent_orders', v_recent_orders,
                    'last_order_date', NEW.created_at
                ),
                jsonb_build_array(
                    jsonb_build_object(
                        'action', 'Contact ' || v_customer_name || ' about premium offerings',
                        'estimated_impact', '10-20% revenue increase'
                    ),
                    jsonb_build_object(
                        'action', 'Offer loyalty discount for next purchase',
                        'estimated_impact', 'Improved customer retention'
                    )
                )
            );
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$;

CREATE TRIGGER trigger_purchase_patterns
    AFTER INSERT ON sales_orders
    FOR EACH ROW
    EXECUTE FUNCTION detect_purchase_patterns();

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON FUNCTION create_sales_followup_task IS
'Automatically creates follow-up tasks for sales leads';

COMMENT ON FUNCTION create_high_value_lead_task IS
'Flags high-value leads for immediate attention';

COMMENT ON FUNCTION create_low_stock_alert IS
'Creates alerts when inventory levels are low';

COMMENT ON FUNCTION detect_fast_moving_products IS
'Identifies products that are selling quickly';

COMMENT ON FUNCTION detect_large_expense IS
'Alerts about significant expenses that need review';

COMMENT ON FUNCTION analyze_cash_flow IS
'Provides cash flow insights and suggestions';

COMMENT ON FUNCTION detect_purchase_patterns IS
'Identifies upsell and cross-sell opportunities';

-- ============================================
-- EXAMPLE USAGE
-- ============================================

/*
-- These triggers will automatically create AI tasks and insights
-- when business events occur. Examples:

1. When a lead status changes to 'ongoing', Alex (sales assistant) 
   will create a follow-up task if there's been no recent contact.

2. When stock levels drop below 10 units, Taylor (inventory manager)
   will create a low stock alert and reorder suggestion.

3. When a large invoice is created, Jordan (financial advisor)
   will flag it for review.

4. When a customer places multiple orders, Casey (business strategist)
   will suggest upsell opportunities.

The AI agents work automatically in the background, creating tasks
and insights that appear in the entrepreneur's daily briefing.
*/
