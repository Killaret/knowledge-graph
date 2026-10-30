-- Remove source_type field from links table
DROP INDEX IF EXISTS idx_links_source_type;
ALTER TABLE links DROP COLUMN IF EXISTS source_type;
