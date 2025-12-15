-- Migration: Employee-Centric Analytics
-- Description: Performance, productivity, and workforce analytics
-- Created: 2025-01-01

-- =====================================================
-- EMPLOYEE PERFORMANCE ANALYTICS
-- =====================================================

-- Employee productivity summary
CREATE OR REPLACE FUNCTION analytics_employee_productivity(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    employee_id uuid,
    employee_name varchar,
    department_name varchar,
    -- Sales metrics
    sales_orders_count integer,
    sales_total_amount numeric,
    -- Time tracking
    hours_logged numeric,
    days_worked integer,
    avg_hours_per_day numeric,
    -- Activities
    activities_completed integer,
    tasks_completed integer,
    -- Performance score
    productivity_score numeric
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH employee_sales AS (
        SELECT
            e.id as emp_id,
            COUNT(so.id) as orders_count,
            COALESCE(SUM(so.amount_total), 0) as total_amount
        FROM employees e
        LEFT JOIN sales_orders so ON so.user_id = e.user_id
            AND so.organization_id = p_organization_id
            AND so.date_order::date BETWEEN p_date_from AND p_date_to
            AND so.state IN ('sale', 'done')
        WHERE e.organization_id = p_organization_id
          AND e.deleted_at IS NULL
          AND e.active = true
        GROUP BY e.id
    ),
    employee_time AS (
        SELECT
            e.id as emp_id,
            COALESCE(SUM(t.unit_amount), 0) as total_hours,
            COUNT(DISTINCT t.date) as days_worked
        FROM employees e
        LEFT JOIN timesheets t ON t.employee_id = e.id
            AND t.date BETWEEN p_date_from AND p_date_to
        WHERE e.organization_id = p_organization_id
          AND e.deleted_at IS NULL
          AND e.active = true
        GROUP BY e.id
    ),
    employee_activities AS (
        SELECT
            e.id as emp_id,
            COUNT(a.id) FILTER (WHERE a.state = 'done') as activities_done,
            COUNT(t.id) FILTER (WHERE t.kanban_state = 'done') as tasks_done
        FROM employees e
        LEFT JOIN activities a ON a.assigned_to = e.user_id
            AND a.organization_id = p_organization_id
            AND a.done_date::date BETWEEN p_date_from AND p_date_to
        LEFT JOIN tasks t ON e.user_id = ANY(t.user_ids)
            AND t.organization_id = p_organization_id
            AND t.date_end::date BETWEEN p_date_from AND p_date_to
        WHERE e.organization_id = p_organization_id
          AND e.deleted_at IS NULL
          AND e.active = true
        GROUP BY e.id
    )
    SELECT
        e.id,
        e.name,
        d.name,
        COALESCE(es.orders_count, 0)::integer,
        COALESCE(es.total_amount, 0),
        COALESCE(et.total_hours, 0),
        COALESCE(et.days_worked, 0)::integer,
        CASE
            WHEN et.days_worked > 0 THEN ROUND(et.total_hours / et.days_worked, 2)
            ELSE 0
        END,
        COALESCE(ea.activities_done, 0)::integer,
        COALESCE(ea.tasks_done, 0)::integer,
        -- Simple productivity score (0-100)
        LEAST(100,
            (COALESCE(es.orders_count, 0) * 10) +
            (COALESCE(ea.activities_done, 0) * 2) +
            (COALESCE(ea.tasks_done, 0) * 5) +
            (LEAST(40, COALESCE(et.total_hours, 0) / 4)) -- Max 40 points for hours
        )
    FROM employees e
    LEFT JOIN departments d ON e.department_id = d.id
    LEFT JOIN employee_sales es ON e.id = es.emp_id
    LEFT JOIN employee_time et ON e.id = et.emp_id
    LEFT JOIN employee_activities ea ON e.id = ea.emp_id
    WHERE e.organization_id = p_organization_id
      AND e.deleted_at IS NULL
      AND e.active = true
    ORDER BY productivity_score DESC;
END;
$$;

-- Employee attendance tracking
CREATE OR REPLACE FUNCTION analytics_employee_attendance(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    employee_id uuid,
    employee_name varchar,
    department_name varchar,
    total_days integer,
    days_worked integer,
    days_on_leave integer,
    days_absent integer,
    attendance_rate numeric,
    status varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH date_range AS (
        SELECT (p_date_to - p_date_from + 1) as total_days
    ),
    employee_worked AS (
        SELECT
            employee_id,
            COUNT(DISTINCT date) as worked_days
        FROM timesheets
        WHERE organization_id = p_organization_id
          AND date BETWEEN p_date_from AND p_date_to
        GROUP BY employee_id
    ),
    employee_leave AS (
        SELECT
            employee_id,
            SUM(number_of_days)::integer as leave_days
        FROM leave_requests
        WHERE organization_id = p_organization_id
          AND state IN ('validate', 'validate1')
          AND date_from::date <= p_date_to
          AND date_to::date >= p_date_from
        GROUP BY employee_id
    )
    SELECT
        e.id,
        e.name,
        d.name,
        dr.total_days,
        COALESCE(ew.worked_days, 0)::integer,
        COALESCE(el.leave_days, 0)::integer,
        (dr.total_days - COALESCE(ew.worked_days, 0) - COALESCE(el.leave_days, 0))::integer,
        ROUND((COALESCE(ew.worked_days, 0)::numeric / NULLIF(dr.total_days, 0)) * 100, 2),
        CASE
            WHEN COALESCE(ew.worked_days, 0)::numeric / NULLIF(dr.total_days, 0) >= 0.9 THEN 'excellent'
            WHEN COALESCE(ew.worked_days, 0)::numeric / NULLIF(dr.total_days, 0) >= 0.75 THEN 'good'
            WHEN COALESCE(ew.worked_days, 0)::numeric / NULLIF(dr.total_days, 0) >= 0.5 THEN 'fair'
            ELSE 'poor'
        END::varchar
    FROM employees e
    CROSS JOIN date_range dr
    LEFT JOIN departments d ON e.department_id = d.id
    LEFT JOIN employee_worked ew ON e.id = ew.employee_id
    LEFT JOIN employee_leave el ON e.id = el.employee_id
    WHERE e.organization_id = p_organization_id
      AND e.deleted_at IS NULL
      AND e.active = true
      AND e.employment_type != 'contract' -- Exclude contractors
    ORDER BY attendance_rate DESC NULLS LAST;
END;
$$;

-- Employee workload analysis
CREATE OR REPLACE FUNCTION analytics_employee_workload(
    p_organization_id uuid
)
RETURNS TABLE (
    employee_id uuid,
    employee_name varchar,
    department_name varchar,
    open_tasks integer,
    overdue_tasks integer,
    total_hours_planned numeric,
    total_hours_logged numeric,
    utilization_rate numeric,
    workload_status varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH employee_tasks AS (
        SELECT
            e.id as emp_id,
            COUNT(t.id) FILTER (WHERE t.kanban_state != 'done' AND t.deleted_at IS NULL) as open_count,
            COUNT(t.id) FILTER (WHERE t.date_deadline < CURRENT_DATE AND t.kanban_state != 'done' AND t.deleted_at IS NULL) as overdue_count,
            COALESCE(SUM(t.planned_hours) FILTER (WHERE t.kanban_state != 'done' AND t.deleted_at IS NULL), 0) as planned_hrs,
            COALESCE(SUM(t.effective_hours), 0) as logged_hrs
        FROM employees e
        LEFT JOIN tasks t ON e.user_id = ANY(t.user_ids)
            AND t.organization_id = p_organization_id
        WHERE e.organization_id = p_organization_id
          AND e.deleted_at IS NULL
          AND e.active = true
        GROUP BY e.id
    )
    SELECT
        e.id,
        e.name,
        d.name,
        COALESCE(et.open_count, 0)::integer,
        COALESCE(et.overdue_count, 0)::integer,
        COALESCE(et.planned_hrs, 0),
        COALESCE(et.logged_hrs, 0),
        CASE
            WHEN et.planned_hrs > 0 THEN ROUND((et.logged_hrs / et.planned_hrs) * 100, 2)
            ELSE 0
        END,
        CASE
            WHEN et.overdue_count > 5 THEN 'critical'
            WHEN et.open_count > 20 THEN 'overloaded'
            WHEN et.open_count > 10 THEN 'busy'
            WHEN et.open_count > 5 THEN 'normal'
            ELSE 'light'
        END::varchar
    FROM employees e
    LEFT JOIN departments d ON e.department_id = d.id
    LEFT JOIN employee_tasks et ON e.id = et.emp_id
    WHERE e.organization_id = p_organization_id
      AND e.deleted_at IS NULL
      AND e.active = true
    ORDER BY et.overdue_count DESC NULLS LAST, et.open_count DESC NULLS LAST;
END;
$$;

-- Top performers leaderboard
CREATE OR REPLACE FUNCTION analytics_top_performers(
    p_organization_id uuid,
    p_metric varchar DEFAULT 'sales', -- sales, productivity, attendance
    p_limit integer DEFAULT 10,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    rank integer,
    employee_id uuid,
    employee_name varchar,
    department_name varchar,
    metric_value numeric,
    metric_label varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    IF p_metric = 'sales' THEN
        RETURN QUERY
        SELECT
            ROW_NUMBER() OVER (ORDER BY COALESCE(SUM(so.amount_total), 0) DESC)::integer,
            e.id,
            e.name,
            d.name,
            COALESCE(SUM(so.amount_total), 0),
            'Total Sales'::varchar
        FROM employees e
        LEFT JOIN departments d ON e.department_id = d.id
        LEFT JOIN sales_orders so ON so.user_id = e.user_id
            AND so.organization_id = p_organization_id
            AND so.date_order::date BETWEEN p_date_from AND p_date_to
            AND so.state IN ('sale', 'done')
        WHERE e.organization_id = p_organization_id
          AND e.deleted_at IS NULL
          AND e.active = true
        GROUP BY e.id, e.name, d.name
        ORDER BY metric_value DESC
        LIMIT p_limit;

    ELSIF p_metric = 'productivity' THEN
        RETURN QUERY
        SELECT
            ROW_NUMBER() OVER (ORDER BY productivity_score DESC)::integer,
            employee_id,
            employee_name,
            department_name,
            productivity_score,
            'Productivity Score'::varchar
        FROM analytics_employee_productivity(p_organization_id, p_date_from, p_date_to)
        ORDER BY productivity_score DESC
        LIMIT p_limit;

    ELSIF p_metric = 'attendance' THEN
        RETURN QUERY
        SELECT
            ROW_NUMBER() OVER (ORDER BY attendance_rate DESC)::integer,
            employee_id,
            employee_name,
            department_name,
            attendance_rate,
            'Attendance Rate'::varchar
        FROM analytics_employee_attendance(p_organization_id, p_date_from, p_date_to)
        ORDER BY attendance_rate DESC
        LIMIT p_limit;
    END IF;
END;
$$;

-- Department performance comparison
CREATE OR REPLACE FUNCTION analytics_department_performance(
    p_organization_id uuid,
    p_date_from date DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_date_to date DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    department_id uuid,
    department_name varchar,
    employee_count integer,
    total_sales numeric,
    avg_sales_per_employee numeric,
    total_hours_logged numeric,
    avg_productivity_score numeric,
    avg_attendance_rate numeric,
    performance_grade varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    WITH dept_stats AS (
        SELECT
            d.id,
            d.name,
            COUNT(DISTINCT e.id) as emp_count,
            COALESCE(SUM(so.amount_total), 0) as sales,
            COALESCE(SUM(t.unit_amount), 0) as hours,
            COALESCE(AVG(ep.productivity_score), 0) as avg_prod,
            COALESCE(AVG(ea.attendance_rate), 0) as avg_att
        FROM departments d
        LEFT JOIN employees e ON e.department_id = d.id
            AND e.organization_id = p_organization_id
            AND e.deleted_at IS NULL
            AND e.active = true
        LEFT JOIN sales_orders so ON so.user_id = e.user_id
            AND so.organization_id = p_organization_id
            AND so.date_order::date BETWEEN p_date_from AND p_date_to
            AND so.state IN ('sale', 'done')
        LEFT JOIN timesheets t ON t.employee_id = e.id
            AND t.date BETWEEN p_date_from AND p_date_to
        LEFT JOIN analytics_employee_productivity(p_organization_id, p_date_from, p_date_to) ep ON ep.employee_id = e.id
        LEFT JOIN analytics_employee_attendance(p_organization_id, p_date_from, p_date_to) ea ON ea.employee_id = e.id
        WHERE d.organization_id = p_organization_id
          AND d.deleted_at IS NULL
        GROUP BY d.id, d.name
    )
    SELECT
        ds.id,
        ds.name,
        ds.emp_count::integer,
        ds.sales,
        CASE WHEN ds.emp_count > 0 THEN ROUND(ds.sales / ds.emp_count, 2) ELSE 0 END,
        ds.hours,
        ROUND(ds.avg_prod, 2),
        ROUND(ds.avg_att, 2),
        CASE
            WHEN ds.avg_prod >= 70 AND ds.avg_att >= 85 THEN 'A'
            WHEN ds.avg_prod >= 50 AND ds.avg_att >= 75 THEN 'B'
            WHEN ds.avg_prod >= 30 AND ds.avg_att >= 60 THEN 'C'
            ELSE 'D'
        END::varchar
    FROM dept_stats ds
    WHERE ds.emp_count > 0
    ORDER BY ds.avg_prod DESC, ds.avg_att DESC;
END;
$$;

-- Employee leave balance
CREATE OR REPLACE FUNCTION analytics_employee_leave_balance(
    p_organization_id uuid
)
RETURNS TABLE (
    employee_id uuid,
    employee_name varchar,
    leave_type_id uuid,
    leave_type_name varchar,
    total_allocation numeric,
    days_taken numeric,
    days_remaining numeric,
    pending_requests integer,
    status varchar
)
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN QUERY
    SELECT
        e.id,
        e.name,
        lt.id,
        lt.name,
        lt.max_leaves,
        COALESCE(SUM(lr.number_of_days) FILTER (WHERE lr.state IN ('validate', 'validate1')), 0),
        lt.max_leaves - COALESCE(SUM(lr.number_of_days) FILTER (WHERE lr.state IN ('validate', 'validate1')), 0),
        COUNT(lr.id) FILTER (WHERE lr.state = 'confirm')::integer,
        CASE
            WHEN lt.max_leaves - COALESCE(SUM(lr.number_of_days) FILTER (WHERE lr.state IN ('validate', 'validate1')), 0) < 2 THEN 'critical'
            WHEN lt.max_leaves - COALESCE(SUM(lr.number_of_days) FILTER (WHERE lr.state IN ('validate', 'validate1')), 0) < 5 THEN 'low'
            ELSE 'normal'
        END::varchar
    FROM employees e
    CROSS JOIN leave_types lt
    LEFT JOIN leave_requests lr ON lr.employee_id = e.id
        AND lr.holiday_status_id = lt.id
        AND lr.organization_id = p_organization_id
    WHERE e.organization_id = p_organization_id
      AND lt.organization_id = p_organization_id
      AND e.deleted_at IS NULL
      AND e.active = true
      AND lt.active = true
    GROUP BY e.id, e.name, lt.id, lt.name, lt.max_leaves
    ORDER BY e.name, lt.name;
END;
$$;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON FUNCTION analytics_employee_productivity IS 'Employee productivity metrics including sales, hours, and tasks';
COMMENT ON FUNCTION analytics_employee_attendance IS 'Employee attendance tracking and rates';
COMMENT ON FUNCTION analytics_employee_workload IS 'Current workload and task analysis per employee';
COMMENT ON FUNCTION analytics_top_performers IS 'Leaderboard of top performing employees';
COMMENT ON FUNCTION analytics_department_performance IS 'Department-level performance comparison';
COMMENT ON FUNCTION analytics_employee_leave_balance IS 'Leave balance and allocation tracking';
