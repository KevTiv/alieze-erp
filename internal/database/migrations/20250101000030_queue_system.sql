-- =====================================================
-- QUEUE MANAGEMENT SYSTEM
-- =====================================================
-- A PostgreSQL-based job queue system for background processing
-- Features:
--   - Priority queues
--   - Retry logic with exponential backoff
--   - Dead letter queue
--   - Job scheduling (delayed execution)
--   - Multi-tenancy support
--   - Monitoring and metrics
-- =====================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- QUEUE TABLES
-- =====================================================

-- Main queue table
CREATE TABLE IF NOT EXISTS job_queue (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations(id) ON DELETE CASCADE,

    -- Queue classification
    queue_name text NOT NULL DEFAULT 'default',
    job_type text NOT NULL, -- e.g., 'generate_embedding', 'process_invoice', 'send_email'

    -- Job data
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    result jsonb,
    error_message text,

    -- Priority and scheduling
    priority int NOT NULL DEFAULT 0, -- Higher = more priority
    scheduled_at timestamptz NOT NULL DEFAULT now(), -- When job should run

    -- Status tracking
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    attempt_count int NOT NULL DEFAULT 0,
    max_attempts int NOT NULL DEFAULT 3,

    -- Worker tracking
    worker_id text, -- ID of worker processing this job
    locked_at timestamptz,
    locked_until timestamptz,

    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Metadata
    metadata jsonb DEFAULT '{}'::jsonb,

    -- Indexes for performance
    CONSTRAINT valid_scheduled_at CHECK (scheduled_at >= created_at)
);

-- Dead letter queue for failed jobs
CREATE TABLE IF NOT EXISTS job_dead_letter_queue (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    original_job_id uuid NOT NULL,
    organization_id uuid REFERENCES organizations(id) ON DELETE CASCADE,

    queue_name text NOT NULL,
    job_type text NOT NULL,
    payload jsonb NOT NULL,

    -- Failure details
    error_message text NOT NULL,
    attempt_count int NOT NULL,
    failed_at timestamptz NOT NULL DEFAULT now(),

    -- Original job metadata
    metadata jsonb DEFAULT '{}'::jsonb,

    created_at timestamptz NOT NULL DEFAULT now()
);

-- Queue statistics table
CREATE TABLE IF NOT EXISTS queue_stats (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_name text NOT NULL,
    date date NOT NULL DEFAULT CURRENT_DATE,

    -- Metrics
    jobs_enqueued int NOT NULL DEFAULT 0,
    jobs_completed int NOT NULL DEFAULT 0,
    jobs_failed int NOT NULL DEFAULT 0,
    total_processing_time_ms bigint NOT NULL DEFAULT 0,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    UNIQUE(queue_name, date)
);

-- =====================================================
-- INDEXES
-- =====================================================

-- Primary query indexes
CREATE INDEX IF NOT EXISTS idx_job_queue_status_priority
    ON job_queue(status, priority DESC, scheduled_at ASC)
    WHERE status IN ('pending', 'processing');

CREATE INDEX IF NOT EXISTS idx_job_queue_organization
    ON job_queue(organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_job_queue_queue_name
    ON job_queue(queue_name, status);

CREATE INDEX IF NOT EXISTS idx_job_queue_scheduled
    ON job_queue(scheduled_at)
    WHERE status = 'pending';

-- Worker query indexes
CREATE INDEX IF NOT EXISTS idx_job_queue_worker
    ON job_queue(worker_id, status);

-- Cleanup indexes
CREATE INDEX IF NOT EXISTS idx_job_queue_completed_at
    ON job_queue(completed_at)
    WHERE status IN ('completed', 'failed');

-- Dead letter queue indexes
CREATE INDEX IF NOT EXISTS idx_dlq_organization
    ON job_dead_letter_queue(organization_id, failed_at DESC);

CREATE INDEX IF NOT EXISTS idx_dlq_queue_name
    ON job_dead_letter_queue(queue_name, failed_at DESC);

-- =====================================================
-- CORE QUEUE FUNCTIONS
-- =====================================================

-- Enqueue a job
CREATE OR REPLACE FUNCTION enqueue_job(
    p_organization_id uuid,
    p_queue_name text,
    p_job_type text,
    p_payload jsonb DEFAULT '{}'::jsonb,
    p_priority int DEFAULT 0,
    p_max_attempts int DEFAULT 3,
    p_scheduled_at timestamptz DEFAULT now(),
    p_metadata jsonb DEFAULT '{}'::jsonb
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_job_id uuid;
BEGIN
    INSERT INTO job_queue (
        organization_id,
        queue_name,
        job_type,
        payload,
        priority,
        max_attempts,
        scheduled_at,
        metadata,
        status
    ) VALUES (
        p_organization_id,
        p_queue_name,
        p_job_type,
        p_payload,
        p_priority,
        p_max_attempts,
        p_scheduled_at,
        p_metadata,
        'pending'
    )
    RETURNING id INTO v_job_id;

    -- Update stats
    INSERT INTO queue_stats (queue_name, jobs_enqueued)
    VALUES (p_queue_name, 1)
    ON CONFLICT (queue_name, date)
    DO UPDATE SET
        jobs_enqueued = queue_stats.jobs_enqueued + 1,
        updated_at = now();

    RETURN v_job_id;
END;
$$;

-- Dequeue a job for processing
CREATE OR REPLACE FUNCTION dequeue_job(
    p_worker_id text,
    p_queue_names text[] DEFAULT ARRAY['default'],
    p_lock_duration_seconds int DEFAULT 300
)
RETURNS TABLE (
    job_id uuid,
    organization_id uuid,
    queue_name text,
    job_type text,
    payload jsonb,
    attempt_count int,
    metadata jsonb
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_job_id uuid;
    v_lock_until timestamptz;
BEGIN
    v_lock_until := now() + (p_lock_duration_seconds || ' seconds')::interval;

    -- Find and lock the highest priority job
    SELECT jq.id INTO v_job_id
    FROM job_queue jq
    WHERE jq.queue_name = ANY(p_queue_names)
        AND jq.status = 'pending'
        AND jq.scheduled_at <= now()
        AND (jq.locked_until IS NULL OR jq.locked_until < now())
    ORDER BY jq.priority DESC, jq.scheduled_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED;

    -- If no job found, return empty
    IF v_job_id IS NULL THEN
        RETURN;
    END IF;

    -- Update job status and lock it
    UPDATE job_queue
    SET
        status = 'processing',
        worker_id = p_worker_id,
        locked_at = now(),
        locked_until = v_lock_until,
        started_at = COALESCE(started_at, now()),
        attempt_count = attempt_count + 1,
        updated_at = now()
    WHERE id = v_job_id;

    -- Return job details
    RETURN QUERY
    SELECT
        jq.id,
        jq.organization_id,
        jq.queue_name,
        jq.job_type,
        jq.payload,
        jq.attempt_count,
        jq.metadata
    FROM job_queue jq
    WHERE jq.id = v_job_id;
END;
$$;

-- Mark job as completed
CREATE OR REPLACE FUNCTION complete_job(
    p_job_id uuid,
    p_result jsonb DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_queue_name text;
    v_processing_time_ms bigint;
BEGIN
    -- Get queue name and calculate processing time
    SELECT
        queue_name,
        EXTRACT(EPOCH FROM (now() - started_at)) * 1000
    INTO v_queue_name, v_processing_time_ms
    FROM job_queue
    WHERE id = p_job_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Update job
    UPDATE job_queue
    SET
        status = 'completed',
        result = p_result,
        completed_at = now(),
        updated_at = now(),
        worker_id = NULL,
        locked_at = NULL,
        locked_until = NULL
    WHERE id = p_job_id;

    -- Update stats
    INSERT INTO queue_stats (queue_name, jobs_completed, total_processing_time_ms)
    VALUES (v_queue_name, 1, COALESCE(v_processing_time_ms, 0))
    ON CONFLICT (queue_name, date)
    DO UPDATE SET
        jobs_completed = queue_stats.jobs_completed + 1,
        total_processing_time_ms = queue_stats.total_processing_time_ms + COALESCE(v_processing_time_ms, 0),
        updated_at = now();

    RETURN true;
END;
$$;

-- Mark job as failed
CREATE OR REPLACE FUNCTION fail_job(
    p_job_id uuid,
    p_error_message text,
    p_retry boolean DEFAULT true
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_job job_queue%ROWTYPE;
    v_next_attempt_delay interval;
BEGIN
    -- Get job details
    SELECT * INTO v_job
    FROM job_queue
    WHERE id = p_job_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Check if we should retry
    IF p_retry AND v_job.attempt_count < v_job.max_attempts THEN
        -- Calculate exponential backoff: 2^attempt * 60 seconds
        v_next_attempt_delay := (power(2, v_job.attempt_count) * 60)::text || ' seconds';

        -- Reset to pending with scheduled retry
        UPDATE job_queue
        SET
            status = 'pending',
            error_message = p_error_message,
            scheduled_at = now() + v_next_attempt_delay,
            updated_at = now(),
            worker_id = NULL,
            locked_at = NULL,
            locked_until = NULL
        WHERE id = p_job_id;

        RETURN true;
    ELSE
        -- Max attempts reached or retry disabled - mark as failed
        UPDATE job_queue
        SET
            status = 'failed',
            error_message = p_error_message,
            completed_at = now(),
            updated_at = now(),
            worker_id = NULL,
            locked_at = NULL,
            locked_until = NULL
        WHERE id = p_job_id;

        -- Move to dead letter queue
        INSERT INTO job_dead_letter_queue (
            original_job_id,
            organization_id,
            queue_name,
            job_type,
            payload,
            error_message,
            attempt_count,
            metadata
        )
        SELECT
            id,
            organization_id,
            queue_name,
            job_type,
            payload,
            p_error_message,
            attempt_count,
            metadata
        FROM job_queue
        WHERE id = p_job_id;

        -- Update stats
        UPDATE queue_stats
        SET
            jobs_failed = jobs_failed + 1,
            updated_at = now()
        WHERE queue_name = v_job.queue_name
            AND date = CURRENT_DATE;

        RETURN true;
    END IF;
END;
$$;

-- Extend job lock (heartbeat)
CREATE OR REPLACE FUNCTION extend_job_lock(
    p_job_id uuid,
    p_worker_id text,
    p_extend_seconds int DEFAULT 300
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE job_queue
    SET
        locked_until = now() + (p_extend_seconds || ' seconds')::interval,
        updated_at = now()
    WHERE id = p_job_id
        AND worker_id = p_worker_id
        AND status = 'processing';

    RETURN FOUND;
END;
$$;

-- =====================================================
-- QUEUE MONITORING FUNCTIONS
-- =====================================================

-- Get queue status
CREATE OR REPLACE FUNCTION get_queue_status(p_queue_name text DEFAULT NULL)
RETURNS TABLE (
    queue_name text,
    pending_count bigint,
    processing_count bigint,
    completed_today bigint,
    failed_today bigint,
    avg_processing_time_ms numeric,
    oldest_pending_job timestamptz
)
LANGUAGE sql
STABLE
AS $$
    SELECT
        COALESCE(jq.queue_name, qs.queue_name) as queue_name,
        COUNT(*) FILTER (WHERE jq.status = 'pending') as pending_count,
        COUNT(*) FILTER (WHERE jq.status = 'processing') as processing_count,
        COALESCE(qs.jobs_completed, 0) as completed_today,
        COALESCE(qs.jobs_failed, 0) as failed_today,
        CASE
            WHEN COALESCE(qs.jobs_completed, 0) > 0
            THEN ROUND(qs.total_processing_time_ms::numeric / qs.jobs_completed, 2)
            ELSE 0
        END as avg_processing_time_ms,
        MIN(jq.created_at) FILTER (WHERE jq.status = 'pending') as oldest_pending_job
    FROM job_queue jq
    FULL OUTER JOIN queue_stats qs ON jq.queue_name = qs.queue_name AND qs.date = CURRENT_DATE
    WHERE (p_queue_name IS NULL OR jq.queue_name = p_queue_name OR qs.queue_name = p_queue_name)
    GROUP BY COALESCE(jq.queue_name, qs.queue_name), qs.jobs_completed, qs.jobs_failed, qs.total_processing_time_ms;
$$;

-- Get stuck jobs (locked but expired)
CREATE OR REPLACE FUNCTION get_stuck_jobs()
RETURNS TABLE (
    job_id uuid,
    queue_name text,
    job_type text,
    worker_id text,
    locked_at timestamptz,
    locked_until timestamptz,
    stuck_duration interval
)
LANGUAGE sql
STABLE
AS $$
    SELECT
        id,
        queue_name,
        job_type,
        worker_id,
        locked_at,
        locked_until,
        now() - locked_until as stuck_duration
    FROM job_queue
    WHERE status = 'processing'
        AND locked_until < now()
    ORDER BY locked_until ASC;
$$;

-- Release stuck jobs
CREATE OR REPLACE FUNCTION release_stuck_jobs()
RETURNS int
LANGUAGE plpgsql
AS $$
DECLARE
    v_released_count int;
BEGIN
    UPDATE job_queue
    SET
        status = 'pending',
        worker_id = NULL,
        locked_at = NULL,
        locked_until = NULL,
        scheduled_at = now(),
        updated_at = now()
    WHERE status = 'processing'
        AND locked_until < now();

    GET DIAGNOSTICS v_released_count = ROW_COUNT;

    RETURN v_released_count;
END;
$$;

-- =====================================================
-- QUEUE MAINTENANCE FUNCTIONS
-- =====================================================

-- Cleanup old completed jobs
CREATE OR REPLACE FUNCTION cleanup_old_jobs(
    p_days_old int DEFAULT 7,
    p_batch_size int DEFAULT 1000
)
RETURNS int
LANGUAGE plpgsql
AS $$
DECLARE
    v_deleted_count int := 0;
    v_batch_deleted int;
BEGIN
    LOOP
        DELETE FROM job_queue
        WHERE id IN (
            SELECT id
            FROM job_queue
            WHERE status IN ('completed', 'cancelled')
                AND completed_at < now() - (p_days_old || ' days')::interval
            LIMIT p_batch_size
        );

        GET DIAGNOSTICS v_batch_deleted = ROW_COUNT;
        v_deleted_count := v_deleted_count + v_batch_deleted;

        EXIT WHEN v_batch_deleted = 0;

        -- Avoid long-running transaction
        COMMIT;
    END LOOP;

    RETURN v_deleted_count;
END;
$$;

-- Cancel a job
CREATE OR REPLACE FUNCTION cancel_job(p_job_id uuid)
RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE job_queue
    SET
        status = 'cancelled',
        completed_at = now(),
        updated_at = now(),
        worker_id = NULL,
        locked_at = NULL,
        locked_until = NULL
    WHERE id = p_job_id
        AND status IN ('pending', 'processing');

    RETURN FOUND;
END;
$$;

-- Retry a failed job
CREATE OR REPLACE FUNCTION retry_failed_job(p_job_id uuid)
RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE job_queue
    SET
        status = 'pending',
        attempt_count = 0,
        error_message = NULL,
        scheduled_at = now(),
        completed_at = NULL,
        updated_at = now(),
        worker_id = NULL,
        locked_at = NULL,
        locked_until = NULL
    WHERE id = p_job_id
        AND status = 'failed';

    RETURN FOUND;
END;
$$;

-- =====================================================
-- ROW LEVEL SECURITY
-- =====================================================

ALTER TABLE job_queue ENABLE ROW LEVEL SECURITY;
ALTER TABLE job_dead_letter_queue ENABLE ROW LEVEL SECURITY;
ALTER TABLE queue_stats ENABLE ROW LEVEL SECURITY;

-- Allow service role full access
CREATE POLICY job_queue_service_role ON job_queue
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

CREATE POLICY dlq_service_role ON job_dead_letter_queue
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

CREATE POLICY queue_stats_service_role ON queue_stats
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

-- Allow authenticated users to access their organization's jobs
CREATE POLICY job_queue_org_access ON job_queue
    FOR ALL
    TO authenticated
    USING (organization_id = (current_setting('app.current_organization_id', true)::uuid))
    WITH CHECK (organization_id = (current_setting('app.current_organization_id', true)::uuid));

CREATE POLICY dlq_org_access ON job_dead_letter_queue
    FOR ALL
    TO authenticated
    USING (organization_id = (current_setting('app.current_organization_id', true)::uuid))
    WITH CHECK (organization_id = (current_setting('app.current_organization_id', true)::uuid));

-- Allow authenticated users to read queue stats
CREATE POLICY queue_stats_read ON queue_stats
    FOR SELECT
    TO authenticated
    USING (true);

-- =====================================================
-- HELPFUL VIEWS
-- =====================================================

-- Recent jobs view
CREATE OR REPLACE VIEW recent_jobs AS
SELECT
    id,
    organization_id,
    queue_name,
    job_type,
    status,
    priority,
    attempt_count,
    max_attempts,
    error_message,
    worker_id,
    created_at,
    started_at,
    completed_at,
    CASE
        WHEN completed_at IS NOT NULL AND started_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (completed_at - started_at))
        ELSE NULL
    END as processing_time_seconds
FROM job_queue
WHERE created_at > now() - interval '24 hours'
ORDER BY created_at DESC;

-- Queue health view
CREATE OR REPLACE VIEW queue_health AS
SELECT
    queue_name,
    COUNT(*) FILTER (WHERE status = 'pending' AND scheduled_at <= now()) as ready_jobs,
    COUNT(*) FILTER (WHERE status = 'pending' AND scheduled_at > now()) as scheduled_jobs,
    COUNT(*) FILTER (WHERE status = 'processing') as processing_jobs,
    COUNT(*) FILTER (WHERE status = 'processing' AND locked_until < now()) as stuck_jobs,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) FILTER (WHERE status = 'completed' AND completed_at > now() - interval '1 hour') as avg_processing_time_1h
FROM job_queue
GROUP BY queue_name;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE job_queue IS 'Main job queue for background processing';
COMMENT ON TABLE job_dead_letter_queue IS 'Failed jobs that exceeded max retry attempts';
COMMENT ON TABLE queue_stats IS 'Daily statistics for queue monitoring';

COMMENT ON FUNCTION enqueue_job IS 'Add a new job to the queue';
COMMENT ON FUNCTION dequeue_job IS 'Get the next job for processing (worker-safe)';
COMMENT ON FUNCTION complete_job IS 'Mark a job as successfully completed';
COMMENT ON FUNCTION fail_job IS 'Mark a job as failed and handle retry logic';
COMMENT ON FUNCTION extend_job_lock IS 'Extend the lock on a processing job (heartbeat)';
COMMENT ON FUNCTION get_queue_status IS 'Get current status of queues';
COMMENT ON FUNCTION get_stuck_jobs IS 'Find jobs that are locked but their lock has expired';
COMMENT ON FUNCTION release_stuck_jobs IS 'Release stuck jobs back to pending';
COMMENT ON FUNCTION cleanup_old_jobs IS 'Delete old completed/cancelled jobs';
COMMENT ON FUNCTION cancel_job IS 'Cancel a pending or processing job';
COMMENT ON FUNCTION retry_failed_job IS 'Reset a failed job to retry it';
