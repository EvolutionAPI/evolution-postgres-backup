-- Migration: Add jobs table for worker system
-- Date: 2025-07-21

-- Jobs table for worker queue system
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('backup', 'restore', 'cleanup')),
    postgres_id TEXT,
    database_name TEXT,
    backup_id TEXT,
    priority INTEGER NOT NULL DEFAULT 5,
    payload TEXT, -- JSON payload (optional)
    status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed', 'retrying')),
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    error_message TEXT,
    worker_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Index for efficient job queries
CREATE INDEX IF NOT EXISTS idx_jobs_status_priority ON jobs(status, priority DESC, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at); 