-- Add missing columns to existing logs table
-- Run this if you have an existing logs table without details and created_at columns

-- Add the details column (optional)
ALTER TABLE logs 
ADD COLUMN IF NOT EXISTS details TEXT;

-- Add the created_at column with default value
ALTER TABLE logs 
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Update existing records to have created_at = timestamp if created_at is NULL
UPDATE logs 
SET created_at = timestamp 
WHERE created_at IS NULL;

-- Verify the migration
SELECT id, timestamp, level, component, job_id, backup_id, message, details, created_at 
FROM logs 
LIMIT 5; 