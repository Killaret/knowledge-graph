-- Add deleted_at and updated_at columns to links table
-- This fixes 500 errors when creating links

ALTER TABLE links 
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now(),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Create trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_links_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_links_updated_at ON links;
CREATE TRIGGER trigger_links_updated_at
    BEFORE UPDATE ON links
    FOR EACH ROW
    EXECUTE FUNCTION update_links_updated_at();
