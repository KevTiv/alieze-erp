-- =====================================================
-- CONVERT QUEUE TABLES TO UNLOGGED
-- =====================================================
-- Description: Convert job queue tables to UNLOGGED for better performance
--
-- UNLOGGED tables are faster because they skip WAL (Write-Ahead Log) but
-- data is lost on server crash - acceptable for queue/cache systems
-- where jobs can be re-enqueued if needed.
--
-- Note: This migration assumes you're okay with losing queue jobs on crash.
-- If you need persistence, keep tables as regular tables.
-- =====================================================

-- Convert job_queue to UNLOGGED
-- Note: Cannot directly ALTER, must recreate
BEGIN;

-- Create new UNLOGGED table
CREATE UNLOGGED TABLE job_queue_new (LIKE job_queue INCLUDING ALL);

-- Copy data if exists
INSERT INTO job_queue_new SELECT * FROM job_queue;

-- Drop old table and rename
DROP TABLE job_queue CASCADE;
ALTER TABLE job_queue_new RENAME TO job_queue;

-- Recreate foreign key constraints
ALTER TABLE job_queue
    ADD CONSTRAINT job_queue_organization_fk
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE CASCADE;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_job_queue_status ON job_queue(status) WHERE status IN ('pending', 'processing');
CREATE INDEX IF NOT EXISTS idx_job_queue_scheduled ON job_queue(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_job_queue_queue_priority ON job_queue(queue_name, priority DESC, scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_job_queue_worker ON job_queue(worker_id) WHERE worker_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_job_queue_org ON job_queue(organization_id);

COMMIT;

-- Convert job_dead_letter_queue to UNLOGGED
BEGIN;

CREATE UNLOGGED TABLE job_dead_letter_queue_new (LIKE job_dead_letter_queue INCLUDING ALL);

INSERT INTO job_dead_letter_queue_new SELECT * FROM job_dead_letter_queue;

DROP TABLE job_dead_letter_queue CASCADE;
ALTER TABLE job_dead_letter_queue_new RENAME TO job_dead_letter_queue;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_job_dlq_date ON job_dead_letter_queue(failed_at DESC);
CREATE INDEX IF NOT EXISTS idx_job_dlq_type ON job_dead_letter_queue(job_type);

COMMIT;

-- Convert queue_stats to UNLOGGED
BEGIN;

CREATE UNLOGGED TABLE queue_stats_new (LIKE queue_stats INCLUDING ALL);

INSERT INTO queue_stats_new SELECT * FROM queue_stats;

DROP TABLE queue_stats CASCADE;
ALTER TABLE queue_stats_new RENAME TO queue_stats;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_queue_stats_date ON queue_stats(queue_name, date DESC);

COMMIT;

-- Add comment documenting UNLOGGED status
COMMENT ON TABLE job_queue IS 'UNLOGGED: Job queue for background processing. Data lost on crash but performance is better. Jobs can be re-enqueued.';
COMMENT ON TABLE job_dead_letter_queue IS 'UNLOGGED: Dead letter queue for failed jobs. Data lost on crash.';
COMMENT ON TABLE queue_stats IS 'UNLOGGED: Queue statistics. Data lost on crash but can be recalculated.';
