-- Remove deleted_at and updated_at columns from links table

DROP TRIGGER IF EXISTS trigger_links_updated_at ON links;
DROP FUNCTION IF EXISTS update_links_updated_at();

ALTER TABLE links 
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS deleted_at;
