-- Add enabled column to existing postgresql_instances table
-- Run this if you have an existing table without the enabled column

-- Add the enabled column with default value true
ALTER TABLE postgresql_instances 
ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT true;

-- Update existing records to be enabled by default
UPDATE postgresql_instances 
SET enabled = true 
WHERE enabled IS NULL;

-- Verify the migration
SELECT id, name, enabled FROM postgresql_instances; 