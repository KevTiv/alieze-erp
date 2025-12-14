-- Migration: Purchasing, Manufacturing, HR, and Project Management Modules
-- Description: Remaining core ERP modules
-- Created: 2025-01-01

-- =====================================================
-- PURCHASING MODULE
-- =====================================================

-- Purchase Orders
CREATE TABLE purchase_orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    state varchar(20) DEFAULT 'draft',
    date_order timestamptz DEFAULT now(),
    date_approve timestamptz,
    date_planned timestamptz,
    partner_id uuid NOT NULL REFERENCES contacts(id),
    partner_ref varchar(255),
    currency_id uuid REFERENCES currencies(id),
    amount_untaxed numeric(15,2) DEFAULT 0,
    amount_tax numeric(15,2) DEFAULT 0,
    amount_total numeric(15,2) DEFAULT 0,
    payment_term_id uuid REFERENCES payment_terms(id),
    fiscal_position_id uuid REFERENCES fiscal_positions(id),
    invoice_status varchar(20) DEFAULT 'no',
    receipt_status varchar(20) DEFAULT 'no',
    user_id uuid,
    dest_address_id uuid REFERENCES contacts(id),
    picking_type_id uuid REFERENCES stock_picking_types(id),
    notes text,
    origin varchar(255),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT purchase_orders_state_check CHECK (state IN ('draft', 'sent', 'to approve', 'purchase', 'done', 'cancel'))
);

-- Purchase Order Lines
CREATE TABLE purchase_order_lines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    order_id uuid NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    sequence integer DEFAULT 10,
    name text NOT NULL,
    product_id uuid REFERENCES products(id),
    product_qty numeric(15,4) DEFAULT 1,
    product_uom uuid REFERENCES uom_units(id),
    price_unit numeric(15,2) DEFAULT 0,
    price_subtotal numeric(15,2) DEFAULT 0,
    price_tax numeric(15,2) DEFAULT 0,
    price_total numeric(15,2) DEFAULT 0,
    tax_ids uuid[],
    date_planned timestamptz,
    qty_received numeric(15,4) DEFAULT 0,
    qty_invoiced numeric(15,4) DEFAULT 0,
    account_analytic_id uuid REFERENCES analytic_accounts(id),
    state varchar(20) DEFAULT 'draft',
    display_type varchar(20),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb
);

-- =====================================================
-- MANUFACTURING (MRP) MODULE
-- =====================================================

-- Workcenters
CREATE TABLE workcenters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(50),
    working_state varchar(20),
    capacity numeric(10,2) DEFAULT 1,
    sequence integer DEFAULT 10,
    color integer,
    time_efficiency numeric(5,2) DEFAULT 100,
    costs_hour numeric(15,2) DEFAULT 0,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Bill of Materials
CREATE TABLE bom_bills (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    product_tmpl_id uuid REFERENCES products(id),
    product_id uuid REFERENCES product_variants(id),
    product_qty numeric(15,4) DEFAULT 1,
    product_uom_id uuid REFERENCES uom_units(id),
    code varchar(255),
    type varchar(20) DEFAULT 'normal',
    sequence integer DEFAULT 10,
    ready_to_produce varchar(20) DEFAULT 'all_available',
    picking_type_id uuid REFERENCES stock_picking_types(id),
    consumption varchar(20) DEFAULT 'flexible',
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT bom_bills_type_check CHECK (type IN ('normal', 'phantom'))
);

-- BOM Lines
CREATE TABLE bom_lines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    bom_id uuid NOT NULL REFERENCES bom_bills(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products(id),
    product_qty numeric(15,4) DEFAULT 1,
    product_uom_id uuid REFERENCES uom_units(id),
    sequence integer DEFAULT 10,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Manufacturing Orders
CREATE TABLE manufacturing_orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    origin varchar(255),
    state varchar(20) DEFAULT 'draft',
    product_id uuid NOT NULL REFERENCES products(id),
    product_qty numeric(15,4) DEFAULT 1,
    product_uom_id uuid REFERENCES uom_units(id),
    bom_id uuid REFERENCES bom_bills(id),
    date_planned_start timestamptz,
    date_planned_finished timestamptz,
    date_deadline timestamptz,
    date_start timestamptz,
    date_finished timestamptz,
    location_src_id uuid REFERENCES stock_locations(id),
    location_dest_id uuid REFERENCES stock_locations(id),
    picking_type_id uuid REFERENCES stock_picking_types(id),
    qty_producing numeric(15,4) DEFAULT 0,
    reservation_state varchar(20),
    user_id uuid,
    priority varchar(10) DEFAULT '1',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT manufacturing_orders_state_check CHECK (state IN ('draft', 'confirmed', 'progress', 'to_close', 'done', 'cancel'))
);

-- Work Orders
CREATE TABLE work_orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    production_id uuid NOT NULL REFERENCES manufacturing_orders(id) ON DELETE CASCADE,
    workcenter_id uuid REFERENCES workcenters(id),
    product_id uuid REFERENCES products(id),
    state varchar(20) DEFAULT 'pending',
    date_planned_start timestamptz,
    date_planned_finished timestamptz,
    date_start timestamptz,
    date_finished timestamptz,
    duration_expected numeric(10,2),
    duration numeric(10,2),
    sequence integer DEFAULT 10,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,

    CONSTRAINT work_orders_state_check CHECK (state IN ('pending', 'ready', 'progress', 'done', 'cancel'))
);

-- =====================================================
-- HR MODULE
-- =====================================================

-- Resources (base for employees)
CREATE TABLE resources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    resource_type varchar(20) DEFAULT 'user',
    user_id uuid,
    time_efficiency numeric(5,2) DEFAULT 100,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Departments
CREATE TABLE departments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    complete_name varchar(500),
    parent_id uuid REFERENCES departments(id),
    manager_id uuid, -- References employees (created below)
    color integer,
    note text,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Job Positions
CREATE TABLE job_positions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    department_id uuid REFERENCES departments(id),
    expected_employees integer DEFAULT 1,
    no_of_employee integer DEFAULT 0,
    no_of_recruitment integer DEFAULT 0,
    no_of_hired_employee integer DEFAULT 0,
    description text,
    requirements text,
    state varchar(20) DEFAULT 'recruit',
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Employees
CREATE TABLE employees (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    user_id uuid,
    resource_id uuid REFERENCES resources(id),
    name varchar(255) NOT NULL,
    employee_number varchar(50),
    job_title varchar(100),
    job_id uuid REFERENCES job_positions(id),
    department_id uuid REFERENCES departments(id),
    parent_id uuid REFERENCES employees(id),
    coach_id uuid REFERENCES employees(id),
    work_contact_id uuid REFERENCES contacts(id),
    work_email varchar(255),
    work_phone varchar(50),
    mobile_phone varchar(50),
    work_location varchar(255),
    date_hired date,
    date_terminated date,
    employment_type varchar(20) DEFAULT 'full_time',
    gender varchar(20),
    birthday date,
    marital varchar(20),
    emergency_contact varchar(255),
    emergency_phone varchar(50),
    barcode varchar(100),
    pin varchar(100),
    badge_ids uuid[],
    image_url varchar(500),
    color integer,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT employees_employment_type_check CHECK (employment_type IN ('full_time', 'part_time', 'contract', 'intern'))
);

-- Add foreign key to departments for manager
ALTER TABLE departments
    ADD CONSTRAINT departments_manager_fk FOREIGN KEY (manager_id) REFERENCES employees(id);

-- Timesheets
CREATE TABLE timesheets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    employee_id uuid NOT NULL REFERENCES employees(id),
    user_id uuid,
    date date NOT NULL,
    name text NOT NULL,
    unit_amount numeric(10,2) NOT NULL,
    project_id uuid, -- References projects (created below)
    task_id uuid, -- References tasks (created below)
    account_id uuid REFERENCES analytic_accounts(id),
    validated boolean DEFAULT false,
    manager_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz
);

-- Leave Types
CREATE TABLE leave_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    code varchar(50),
    color_name varchar(50),
    allocation_type varchar(20) DEFAULT 'no',
    validity_start date,
    validity_stop date,
    max_leaves numeric(10,2) DEFAULT 0,
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Leave Requests
CREATE TABLE leave_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    employee_id uuid NOT NULL REFERENCES employees(id),
    holiday_status_id uuid NOT NULL REFERENCES leave_types(id),
    name varchar(255),
    state varchar(20) DEFAULT 'confirm',
    date_from timestamptz NOT NULL,
    date_to timestamptz NOT NULL,
    number_of_days numeric(10,2) NOT NULL,
    request_date_from date,
    request_date_to date,
    notes text,
    manager_id uuid,
    first_approver_id uuid,
    second_approver_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT leave_requests_state_check CHECK (state IN ('draft', 'confirm', 'refuse', 'validate', 'validate1'))
);

-- =====================================================
-- PROJECT MANAGEMENT MODULE
-- =====================================================

-- Task Stages
CREATE TABLE task_stages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    sequence integer DEFAULT 1,
    project_ids uuid[],
    fold boolean DEFAULT false,
    rating_template_id uuid,
    auto_validation_kanban_state boolean DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Projects
CREATE TABLE projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    name varchar(255) NOT NULL,
    sequence integer DEFAULT 10,
    active boolean DEFAULT true,
    partner_id uuid REFERENCES contacts(id),
    user_id uuid,
    date_start date,
    date date,
    color integer,
    privacy_visibility varchar(20) DEFAULT 'followers',
    rating_status varchar(20) DEFAULT 'stage',
    rating_status_period varchar(20) DEFAULT 'monthly',
    allow_timesheets boolean DEFAULT true,
    subtask_effective_hours numeric(10,2) DEFAULT 0,
    total_timesheet_time numeric(10,2) DEFAULT 0,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb
);

-- Tasks
CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    company_id uuid REFERENCES companies(id),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    sequence integer DEFAULT 10,
    stage_id uuid REFERENCES task_stages(id),
    partner_id uuid REFERENCES contacts(id),
    user_ids uuid[],
    date_deadline date,
    date_end date,
    date_assign date,
    date_last_stage_update timestamptz,
    priority varchar(10) DEFAULT '0',
    kanban_state varchar(20) DEFAULT 'normal',
    parent_id uuid REFERENCES tasks(id),
    planned_hours numeric(10,2),
    effective_hours numeric(10,2) DEFAULT 0,
    remaining_hours numeric(10,2),
    progress numeric(5,2) DEFAULT 0,
    color integer,
    tag_ids uuid[],
    active boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid,
    updated_by uuid,
    deleted_at timestamptz,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,

    CONSTRAINT tasks_kanban_state_check CHECK (kanban_state IN ('normal', 'done', 'blocked'))
);

-- Add foreign keys to timesheets
ALTER TABLE timesheets
    ADD CONSTRAINT timesheets_project_fk FOREIGN KEY (project_id) REFERENCES projects(id),
    ADD CONSTRAINT timesheets_task_fk FOREIGN KEY (task_id) REFERENCES tasks(id);

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER set_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_purchase_order_lines_updated_at BEFORE UPDATE ON purchase_order_lines FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_workcenters_updated_at BEFORE UPDATE ON workcenters FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_bom_bills_updated_at BEFORE UPDATE ON bom_bills FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_bom_lines_updated_at BEFORE UPDATE ON bom_lines FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_manufacturing_orders_updated_at BEFORE UPDATE ON manufacturing_orders FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_work_orders_updated_at BEFORE UPDATE ON work_orders FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_departments_updated_at BEFORE UPDATE ON departments FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_job_positions_updated_at BEFORE UPDATE ON job_positions FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_employees_updated_at BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_timesheets_updated_at BEFORE UPDATE ON timesheets FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_leave_types_updated_at BEFORE UPDATE ON leave_types FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_leave_requests_updated_at BEFORE UPDATE ON leave_requests FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_task_stages_updated_at BEFORE UPDATE ON task_stages FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_tasks_updated_at BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_purchase_orders_org ON purchase_orders(organization_id);
CREATE INDEX idx_purchase_orders_partner ON purchase_orders(partner_id);
CREATE INDEX idx_purchase_order_lines_order ON purchase_order_lines(order_id);

CREATE INDEX idx_bom_bills_org ON bom_bills(organization_id);
CREATE INDEX idx_bom_lines_bom ON bom_lines(bom_id);
CREATE INDEX idx_manufacturing_orders_org ON manufacturing_orders(organization_id);
CREATE INDEX idx_work_orders_production ON work_orders(production_id);

CREATE INDEX idx_employees_org ON employees(organization_id);
CREATE INDEX idx_employees_department ON employees(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_timesheets_org ON timesheets(organization_id);
CREATE INDEX idx_timesheets_employee ON timesheets(employee_id);
CREATE INDEX idx_leave_requests_org ON leave_requests(organization_id);
CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id);

CREATE INDEX idx_projects_org ON projects(organization_id);
CREATE INDEX idx_tasks_org ON tasks(organization_id);
CREATE INDEX idx_tasks_project ON tasks(project_id);

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE purchase_orders IS 'Purchase orders for vendor management';
COMMENT ON TABLE bom_bills IS 'Bill of materials for manufacturing';
COMMENT ON TABLE manufacturing_orders IS 'Manufacturing production orders';
COMMENT ON TABLE employees IS 'Employee records';
COMMENT ON TABLE timesheets IS 'Time tracking for projects and tasks';
COMMENT ON TABLE projects IS 'Project management';
COMMENT ON TABLE tasks IS 'Tasks within projects';
