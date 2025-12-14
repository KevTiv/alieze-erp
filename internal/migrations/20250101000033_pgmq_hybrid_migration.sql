-- =====================================================
-- PGMQ HYBRID MIGRATION
-- =====================================================
-- Migrates from custom queue system to pgmq while maintaining
-- the existing API for backward compatibility
--
-- Strategy:
-- 1. Enable pgmq extension
-- 2. Create wrapper functions that match old API
-- 3. Keep job handlers and business logic intact
-- 4. Maintain statistics and monitoring capabilities
-- 5. Support gradual migration path
-- =====================================================

-- =====================================================
-- STEP 1: ENABLE PGMQ EXTENSION
-- =====================================================

CREATE EXTENSION IF NOT EXISTS pgmq CASCADE;

-- Grant access to pgmq schema
GRANT USAGE ON SCHEMA pgmq TO postgres, authenticated, service_role;
GRANT ALL ON ALL TABLES IN SCHEMA pgmq TO postgres, service_role;
GRANT SELECT ON ALL TABLES IN SCHEMA pgmq TO authenticated;

-- =====================================================
-- STEP 2: CREATE METADATA TABLES
-- =====================================================
-- pgmq handles the message queue, but we need additional
-- metadata for monitoring, stats, and organization tracking

CREATE TABLE IF NOT EXISTS public.queue_metadata (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    msg_id bigint NOT NULL,
    queue_name text NOT NULL,
    organization_id uuid REFERENCES public.organizations(id) ON DELETE CASCADE,

    -- Job classification
    job_type text NOT NULL,

    -- Status tracking
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    attempt_count int NOT NULL DEFAULT 0,
    max_attempts int NOT NULL DEFAULT 3,

    -- Worker tracking
    worker_id text,

    -- Results and errors
    result jsonb,
    error_message text,

    -- Timestamps
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),

    -- Additional metadata
    metadata jsonb DEFAULT '{}'::jsonb,

    -- Link to pgmq message
    UNIQUE(queue_name, msg_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_queue_metadata_msg_id ON public.queue_metadata(msg_id);
CREATE INDEX IF NOT EXISTS idx_queue_metadata_status ON public.queue_metadata(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_queue_metadata_organization ON public.queue_metadata(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_queue_metadata_queue_name ON public.queue_metadata(queue_name, status);
CREATE INDEX IF NOT EXISTS idx_queue_metadata_worker ON public.queue_metadata(worker_id) WHERE worker_id IS NOT NULL;

-- Dead letter queue metadata (pgmq archives, we track metadata)
CREATE TABLE IF NOT EXISTS public.queue_dead_letter_metadata (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    original_msg_id bigint NOT NULL,
    queue_name text NOT NULL,
    organization_id uuid REFERENCES public.organizations(id) ON DELETE CASCADE,

    job_type text NOT NULL,
    payload jsonb NOT NULL,

    -- Failure details
    error_message text NOT NULL,
    attempt_count int NOT NULL,
    failed_at timestamptz NOT NULL DEFAULT now(),

    metadata jsonb DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_dlq_metadata_organization ON public.queue_dead_letter_metadata(organization_id, failed_at DESC);
CREATE INDEX IF NOT EXISTS idx_dlq_metadata_queue ON public.queue_dead_letter_metadata(queue_name, failed_at DESC);

-- Keep the existing queue_stats table structure
-- (already exists from previous migration, just ensure it's there)
CREATE TABLE IF NOT EXISTS public.queue_stats (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_name text NOT NULL,
    date date NOT NULL DEFAULT CURRENT_DATE,

    jobs_enqueued int NOT NULL DEFAULT 0,
    jobs_completed int NOT NULL DEFAULT 0,
    jobs_failed int NOT NULL DEFAULT 0,
    total_processing_time_ms bigint NOT NULL DEFAULT 0,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    UNIQUE(queue_name, date)
);

-- =====================================================
-- STEP 3: CREATE DEFAULT QUEUES
-- =====================================================

-- Create common queues
SELECT pgmq.create('default');
SELECT pgmq.create('embeddings');
SELECT pgmq.create('emails');
SELECT pgmq.create('invoices');
SELECT pgmq.create('exports');
SELECT pgmq.create('notifications');

-- =====================================================
-- STEP 4: WRAPPER FUNCTIONS (BACKWARD COMPATIBLE API)
-- =====================================================

-- Wrapper: enqueue_job (maintains exact same API as before)
CREATE OR REPLACE FUNCTION public.enqueue_job(
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
    v_msg_id bigint;
    v_metadata_id uuid;
    v_delay_seconds int;
    v_message jsonb;
BEGIN
    -- Create queue if it doesn't exist
    BEGIN
        PERFORM pgmq.create(p_queue_name);
    EXCEPTION
        WHEN duplicate_table THEN
            NULL; -- Queue already exists
    END;

    -- Calculate delay for scheduled jobs
    v_delay_seconds := GREATEST(0, EXTRACT(EPOCH FROM (p_scheduled_at - now()))::int);

    -- Build message payload with all context
    v_message := jsonb_build_object(
        'organization_id', p_organization_id,
        'job_type', p_job_type,
        'payload', p_payload,
        'priority', p_priority,
        'max_attempts', p_max_attempts,
        'metadata', p_metadata,
        'enqueued_at', now()
    );

    -- Send to pgmq
    v_msg_id := pgmq.send(
        queue_name := p_queue_name,
        msg := v_message,
        delay := v_delay_seconds
    );

    -- Store metadata
    INSERT INTO public.queue_metadata (
        msg_id,
        queue_name,
        organization_id,
        job_type,
        status,
        max_attempts,
        metadata,
        created_at
    ) VALUES (
        v_msg_id,
        p_queue_name,
        p_organization_id,
        p_job_type,
        'pending',
        p_max_attempts,
        p_metadata,
        now()
    )
    RETURNING id INTO v_metadata_id;

    -- Update stats
    INSERT INTO public.queue_stats (queue_name, jobs_enqueued)
    VALUES (p_queue_name, 1)
    ON CONFLICT (queue_name, date)
    DO UPDATE SET
        jobs_enqueued = public.queue_stats.jobs_enqueued + 1,
        updated_at = now();

    RETURN v_metadata_id;
END;
$$;

-- Wrapper: dequeue_job (maintains similar API, adapted for pgmq)
DROP FUNCTION IF EXISTS public.dequeue_job(text, text[], integer);
CREATE OR REPLACE FUNCTION public.dequeue_job(
    p_worker_id text,
    p_queue_names text[] DEFAULT ARRAY['default'],
    p_lock_duration_seconds int DEFAULT 300
)
RETURNS TABLE (
    job_id uuid,
    msg_id bigint,
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
    v_queue_name text;
    v_message RECORD;
    v_metadata RECORD;
BEGIN
    -- Try each queue in order
    FOREACH v_queue_name IN ARRAY p_queue_names
    LOOP
        -- Read from pgmq with visibility timeout
        SELECT * INTO v_message
        FROM pgmq.read(
            queue_name := v_queue_name,
            vt := p_lock_duration_seconds,
            qty := 1
        )
        LIMIT 1;

        -- If we got a message, process it
        IF FOUND THEN
            -- Get or create metadata
            SELECT * INTO v_metadata
            FROM public.queue_metadata
            WHERE queue_name = v_queue_name
                AND msg_id = v_message.msg_id;

            IF NOT FOUND THEN
                -- Create metadata if it doesn't exist (shouldn't happen)
                INSERT INTO public.queue_metadata (
                    msg_id,
                    queue_name,
                    organization_id,
                    job_type,
                    status,
                    max_attempts
                ) VALUES (
                    v_message.msg_id,
                    v_queue_name,
                    (v_message.message->>'organization_id')::uuid,
                    v_message.message->>'job_type',
                    'processing',
                    COALESCE((v_message.message->>'max_attempts')::int, 3)
                )
                RETURNING * INTO v_metadata;
            END IF;

            -- Update metadata to processing
            UPDATE public.queue_metadata
            SET
                status = 'processing',
                worker_id = p_worker_id,
                started_at = COALESCE(started_at, now()),
                attempt_count = attempt_count + 1,
                updated_at = now()
            WHERE id = v_metadata.id;

            -- Return job details
            RETURN QUERY
            SELECT
                v_metadata.id,
                v_message.msg_id,
                (v_message.message->>'organization_id')::uuid,
                v_queue_name,
                v_message.message->>'job_type',
                v_message.message->'payload',
                v_metadata.attempt_count + 1,
                COALESCE(v_message.message->'metadata', '{}'::jsonb);

            RETURN;
        END IF;
    END LOOP;

    -- No jobs found in any queue
    RETURN;
END;
$$;

-- Wrapper: complete_job
CREATE OR REPLACE FUNCTION public.complete_job(
    p_job_id uuid,
    p_result jsonb DEFAULT NULL
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_metadata RECORD;
    v_processing_time_ms bigint;
BEGIN
    -- Get job metadata
    SELECT * INTO v_metadata
    FROM public.queue_metadata
    WHERE id = p_job_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Calculate processing time
    v_processing_time_ms := EXTRACT(EPOCH FROM (now() - v_metadata.started_at)) * 1000;

    -- Archive in pgmq (keeps message for audit)
    PERFORM pgmq.archive(v_metadata.queue_name, v_metadata.msg_id);

    -- Update metadata
    UPDATE public.queue_metadata
    SET
        status = 'completed',
        result = p_result,
        completed_at = now(),
        updated_at = now(),
        worker_id = NULL
    WHERE id = p_job_id;

    -- Update stats
    INSERT INTO public.queue_stats (queue_name, jobs_completed, total_processing_time_ms)
    VALUES (v_metadata.queue_name, 1, COALESCE(v_processing_time_ms, 0))
    ON CONFLICT (queue_name, date)
    DO UPDATE SET
        jobs_completed = public.queue_stats.jobs_completed + 1,
        total_processing_time_ms = public.queue_stats.total_processing_time_ms + COALESCE(v_processing_time_ms, 0),
        updated_at = now();

    RETURN true;
END;
$$;

-- Wrapper: fail_job
CREATE OR REPLACE FUNCTION public.fail_job(
    p_job_id uuid,
    p_error_message text,
    p_retry boolean DEFAULT true
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_metadata RECORD;
    v_message RECORD;
    v_next_delay int;
BEGIN
    -- Get job metadata
    SELECT * INTO v_metadata
    FROM public.queue_metadata
    WHERE id = p_job_id;

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Check if we should retry
    IF p_retry AND v_metadata.attempt_count < v_metadata.max_attempts THEN
        -- Calculate exponential backoff: 2^attempt * 60 seconds
        v_next_delay := power(2, v_metadata.attempt_count)::int * 60;

        -- Get the original message from archive
        SELECT * INTO v_message
        FROM pgmq.read(v_metadata.queue_name, 0, 1)
        WHERE msg_id = v_metadata.msg_id;

        IF FOUND THEN
            -- Delete from current queue
            PERFORM pgmq.delete(v_metadata.queue_name, v_metadata.msg_id);

            -- Re-enqueue with delay
            PERFORM pgmq.send(
                queue_name := v_metadata.queue_name,
                msg := v_message.message,
                delay := v_next_delay
            );
        END IF;

        -- Update metadata
        UPDATE public.queue_metadata
        SET
            status = 'pending',
            error_message = p_error_message,
            updated_at = now(),
            worker_id = NULL
        WHERE id = p_job_id;

        RETURN true;
    ELSE
        -- Max attempts reached - move to dead letter queue

        -- Archive the message in pgmq
        PERFORM pgmq.archive(v_metadata.queue_name, v_metadata.msg_id);

        -- Update metadata
        UPDATE public.queue_metadata
        SET
            status = 'failed',
            error_message = p_error_message,
            completed_at = now(),
            updated_at = now(),
            worker_id = NULL
        WHERE id = p_job_id;

        -- Insert into dead letter metadata
        INSERT INTO public.queue_dead_letter_metadata (
            original_msg_id,
            queue_name,
            organization_id,
            job_type,
            payload,
            error_message,
            attempt_count,
            metadata
        )
        SELECT
            v_metadata.msg_id,
            v_metadata.queue_name,
            v_metadata.organization_id,
            v_metadata.job_type,
            v_metadata.result, -- Payload stored in result for now
            p_error_message,
            v_metadata.attempt_count,
            v_metadata.metadata
        FROM public.queue_metadata
        WHERE id = p_job_id;

        -- Update stats
        UPDATE public.queue_stats
        SET
            jobs_failed = jobs_failed + 1,
            updated_at = now()
        WHERE queue_name = v_metadata.queue_name
            AND date = CURRENT_DATE;

        RETURN true;
    END IF;
END;
$$;

-- Wrapper: extend_job_lock (pgmq manages visibility automatically)
-- This becomes a no-op since pgmq handles visibility timeout
CREATE OR REPLACE FUNCTION public.extend_job_lock(
    p_job_id uuid,
    p_worker_id text,
    p_extend_seconds int DEFAULT 300
)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_metadata RECORD;
BEGIN
    -- Get metadata
    SELECT * INTO v_metadata
    FROM public.queue_metadata
    WHERE id = p_job_id
        AND worker_id = p_worker_id
        AND status = 'processing';

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Extend visibility in pgmq
    PERFORM pgmq.set_vt(
        queue_name := v_metadata.queue_name,
        msg_id := v_metadata.msg_id,
        vt_offset := p_extend_seconds
    );

    -- Update metadata timestamp
    UPDATE public.queue_metadata
    SET updated_at = now()
    WHERE id = p_job_id;

    RETURN true;
END;
$$;

-- =====================================================
-- STEP 5: MONITORING FUNCTIONS
-- =====================================================

-- Get queue status (updated for pgmq)
CREATE OR REPLACE FUNCTION public.get_queue_status(p_queue_name text DEFAULT NULL)
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
        COALESCE(qm.queue_name, qs.queue_name) as queue_name,
        COUNT(*) FILTER (WHERE qm.status = 'pending') as pending_count,
        COUNT(*) FILTER (WHERE qm.status = 'processing') as processing_count,
        COALESCE(qs.jobs_completed, 0) as completed_today,
        COALESCE(qs.jobs_failed, 0) as failed_today,
        CASE
            WHEN COALESCE(qs.jobs_completed, 0) > 0
            THEN ROUND(qs.total_processing_time_ms::numeric / qs.jobs_completed, 2)
            ELSE 0
        END as avg_processing_time_ms,
        MIN(qm.created_at) FILTER (WHERE qm.status = 'pending') as oldest_pending_job
    FROM public.queue_metadata qm
    FULL OUTER JOIN public.queue_stats qs ON qm.queue_name = qs.queue_name AND qs.date = CURRENT_DATE
    WHERE (p_queue_name IS NULL OR qm.queue_name = p_queue_name OR qs.queue_name = p_queue_name)
    GROUP BY COALESCE(qm.queue_name, qs.queue_name), qs.jobs_completed, qs.jobs_failed, qs.total_processing_time_ms;
$$;

-- Get pgmq metrics (native pgmq stats)
CREATE OR REPLACE FUNCTION public.get_pgmq_metrics(p_queue_name text DEFAULT NULL)
RETURNS TABLE (
    queue_name text,
    queue_length bigint,
    newest_msg_age_sec int,
    oldest_msg_age_sec int,
    total_messages bigint
)
LANGUAGE sql
STABLE
AS $$
    SELECT
        queue_name,
        queue_length,
        newest_msg_age_sec,
        oldest_msg_age_sec,
        total_messages
    FROM pgmq.metrics_all()
    WHERE p_queue_name IS NULL OR queue_name = p_queue_name;
$$;

-- Cancel a job
CREATE OR REPLACE FUNCTION public.cancel_job(p_job_id uuid)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_metadata RECORD;
BEGIN
    SELECT * INTO v_metadata
    FROM public.queue_metadata
    WHERE id = p_job_id
        AND status IN ('pending', 'processing');

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Delete from pgmq
    PERFORM pgmq.delete(v_metadata.queue_name, v_metadata.msg_id);

    -- Update metadata
    UPDATE public.queue_metadata
    SET
        status = 'cancelled',
        completed_at = now(),
        updated_at = now(),
        worker_id = NULL
    WHERE id = p_job_id;

    RETURN true;
END;
$$;

-- Retry a failed job
CREATE OR REPLACE FUNCTION public.retry_failed_job(p_job_id uuid)
RETURNS boolean
LANGUAGE plpgsql
AS $$
DECLARE
    v_metadata RECORD;
    v_message jsonb;
BEGIN
    SELECT * INTO v_metadata
    FROM public.queue_metadata
    WHERE id = p_job_id
        AND status = 'failed';

    IF NOT FOUND THEN
        RETURN false;
    END IF;

    -- Rebuild message
    v_message := jsonb_build_object(
        'organization_id', v_metadata.organization_id,
        'job_type', v_metadata.job_type,
        'payload', v_metadata.result, -- Original payload
        'metadata', v_metadata.metadata,
        'max_attempts', v_metadata.max_attempts,
        'enqueued_at', now()
    );

    -- Send to pgmq
    PERFORM pgmq.send(v_metadata.queue_name, v_message, 0);

    -- Update metadata
    UPDATE public.queue_metadata
    SET
        status = 'pending',
        attempt_count = 0,
        error_message = NULL,
        completed_at = NULL,
        updated_at = now(),
        worker_id = NULL
    WHERE id = p_job_id;

    RETURN true;
END;
$$;

-- =====================================================
-- STEP 6: ROW LEVEL SECURITY
-- =====================================================

ALTER TABLE public.queue_metadata ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.queue_dead_letter_metadata ENABLE ROW LEVEL SECURITY;

-- Service role full access
CREATE POLICY queue_metadata_service_role ON public.queue_metadata
    FOR ALL TO service_role
    USING (true) WITH CHECK (true);

CREATE POLICY dlq_metadata_service_role ON public.queue_dead_letter_metadata
    FOR ALL TO service_role
    USING (true) WITH CHECK (true);

-- Authenticated users see their org's jobs
CREATE POLICY queue_metadata_org_access ON public.queue_metadata
    FOR ALL TO authenticated
    USING (organization_id = (current_setting('app.current_organization_id', true)::uuid))
    WITH CHECK (organization_id = (current_setting('app.current_organization_id', true)::uuid));

CREATE POLICY dlq_metadata_org_access ON public.queue_dead_letter_metadata
    FOR ALL TO authenticated
    USING (organization_id = (current_setting('app.current_organization_id', true)::uuid))
    WITH CHECK (organization_id = (current_setting('app.current_organization_id', true)::uuid));

-- =====================================================
-- STEP 7: VIEWS
-- =====================================================

-- Recent jobs view
DROP VIEW IF EXISTS public.recent_jobs;
CREATE OR REPLACE VIEW public.recent_jobs AS
SELECT
    qm.id,
    qm.organization_id,
    qm.queue_name,
    qm.job_type,
    qm.status,
    qm.attempt_count,
    qm.max_attempts,
    qm.error_message,
    qm.worker_id,
    qm.created_at,
    qm.started_at,
    qm.completed_at,
    CASE
        WHEN qm.completed_at IS NOT NULL AND qm.started_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (qm.completed_at - qm.started_at))
        ELSE NULL
    END as processing_time_seconds
FROM public.queue_metadata qm
WHERE qm.created_at > now() - interval '24 hours'
ORDER BY qm.created_at DESC;

-- Queue health view
DROP VIEW IF EXISTS public.queue_health;
CREATE OR REPLACE VIEW public.queue_health AS
SELECT
    qm.queue_name,
    COUNT(*) FILTER (WHERE qm.status = 'pending') as pending_jobs,
    COUNT(*) FILTER (WHERE qm.status = 'processing') as processing_jobs,
    COALESCE(pm.queue_length, 0) as pgmq_queue_length,
    COALESCE(pm.oldest_msg_age_sec, 0) as oldest_msg_age_sec,
    AVG(EXTRACT(EPOCH FROM (qm.completed_at - qm.started_at))) FILTER (
        WHERE qm.status = 'completed' AND qm.completed_at > now() - interval '1 hour'
    ) as avg_processing_time_1h
FROM public.queue_metadata qm
LEFT JOIN pgmq.metrics_all() pm ON qm.queue_name = pm.queue_name
GROUP BY qm.queue_name, pm.queue_length, pm.oldest_msg_age_sec;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON EXTENSION pgmq IS 'Postgres Message Queue - reliable message queue built on Postgres';
COMMENT ON TABLE public.queue_metadata IS 'Metadata for jobs in pgmq queues (org tracking, stats, etc)';
COMMENT ON TABLE public.queue_dead_letter_metadata IS 'Metadata for failed jobs that exceeded max retries';

COMMENT ON FUNCTION public.enqueue_job(uuid, text, text, jsonb, int, int, timestamptz, jsonb) IS 'Wrapper: Add job to pgmq with metadata tracking';
COMMENT ON FUNCTION public.dequeue_job(text, text[], int) IS 'Wrapper: Get next job from pgmq (worker-safe)';
COMMENT ON FUNCTION public.complete_job(uuid, jsonb) IS 'Wrapper: Mark pgmq job as completed';
COMMENT ON FUNCTION public.fail_job(uuid, text, boolean) IS 'Wrapper: Mark pgmq job as failed with retry logic';
COMMENT ON FUNCTION public.extend_job_lock(uuid, text, int) IS 'Wrapper: Extend pgmq message visibility timeout';
COMMENT ON FUNCTION public.get_queue_status(text) IS 'Get current status of queues (combines metadata + pgmq)';
COMMENT ON FUNCTION public.get_pgmq_metrics(text) IS 'Get native pgmq queue metrics';
