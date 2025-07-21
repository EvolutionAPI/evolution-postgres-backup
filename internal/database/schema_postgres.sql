-- PostgreSQL Schema for Backup Service
-- Adapted from SQLite schema with PostgreSQL-specific optimizations

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- PostgreSQL instances table
CREATE TABLE IF NOT EXISTS postgresql_instances (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 5432,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    databases JSONB DEFAULT '["postgres"]'::jsonb, -- Array of database names
    enabled BOOLEAN NOT NULL DEFAULT true, -- Whether instance is enabled for backups
    ssl_mode TEXT NOT NULL DEFAULT 'prefer' CHECK(ssl_mode IN ('disable', 'allow', 'prefer', 'require')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Backups table
CREATE TABLE IF NOT EXISTS backups (
    id TEXT PRIMARY KEY,
    postgresql_id TEXT NOT NULL,
    database_name TEXT NOT NULL,
    backup_type TEXT NOT NULL CHECK(backup_type IN ('hourly', 'daily', 'weekly', 'monthly', 'manual')),
    status TEXT NOT NULL CHECK(status IN ('pending', 'in_progress', 'completed', 'failed')),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    file_path TEXT,
    file_size BIGINT DEFAULT 0,
    s3_key TEXT,
    error_message TEXT,
    job_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (postgresql_id) REFERENCES postgresql_instances(id) ON DELETE CASCADE
);

-- Logs table 
CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    level TEXT NOT NULL CHECK(level IN ('DEBUG', 'INFO', 'WARN', 'ERROR')),
    component TEXT NOT NULL,
    job_id TEXT,
    backup_id TEXT,
    message TEXT NOT NULL,
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Backup schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    postgresql_id TEXT NOT NULL,
    database_name TEXT NOT NULL,
    backup_type TEXT NOT NULL CHECK(backup_type IN ('hourly', 'daily', 'weekly', 'monthly')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    cron_expression TEXT NOT NULL,
    retention_days INTEGER NOT NULL DEFAULT 30,
    last_run TIMESTAMP WITH TIME ZONE,
    next_run TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (postgresql_id) REFERENCES postgresql_instances(id) ON DELETE CASCADE
);

-- Jobs table (for API-Worker communication)
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('backup', 'restore', 'cleanup')),
    postgres_id TEXT NOT NULL,
    database_name TEXT NOT NULL,
    backup_id TEXT,
    priority INTEGER NOT NULL DEFAULT 5,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    payload JSONB,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Configuration table
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_backups_postgresql_id ON backups(postgresql_id);
CREATE INDEX IF NOT EXISTS idx_backups_created_at ON backups(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status);
CREATE INDEX IF NOT EXISTS idx_backups_type ON backups(backup_type);
CREATE INDEX IF NOT EXISTS idx_backups_database ON backups(database_name);
CREATE INDEX IF NOT EXISTS idx_backups_job_id ON backups(job_id);

CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_job_id ON logs(job_id);
CREATE INDEX IF NOT EXISTS idx_logs_backup_id ON logs(backup_id);
CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_component ON logs(component);

CREATE INDEX IF NOT EXISTS idx_schedules_postgresql_id ON schedules(postgresql_id);
CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run);

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_priority ON jobs(priority DESC);

-- Insert default configuration
INSERT INTO config (key, value, description) VALUES 
('app_version', '2.0.0', 'Application version')
ON CONFLICT (key) DO NOTHING;

INSERT INTO config (key, value, description) VALUES 
('db_schema_version', '1.0.0', 'Database schema version')
ON CONFLICT (key) DO NOTHING;

INSERT INTO config (key, value, description) VALUES 
('retention_default_days', '30', 'Default backup retention in days')
ON CONFLICT (key) DO NOTHING; 