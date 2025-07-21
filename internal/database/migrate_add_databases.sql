-- Add databases column to existing postgresql_instances table
-- Run this if you have an existing table without the databases column

-- Add the databases column with default value
ALTER TABLE postgresql_instances 
ADD COLUMN IF NOT EXISTS databases JSONB DEFAULT '["postgres"]'::jsonb;

-- Update existing records to have default databases if NULL
UPDATE postgresql_instances 
SET databases = '["postgres"]'::jsonb 
WHERE databases IS NULL;

-- Verify the migration
SELECT id, name, databases FROM postgresql_instances; 