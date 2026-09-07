-- Add source_type field to links table for tracking link origin (user-created vs gamma/recommended)
ALTER TABLE links ADD COLUMN source_type TEXT DEFAULT 'user' CHECK (source_type IN ('user', 'gamma'));

-- Add index for filtering by source_type
CREATE INDEX idx_links_source_type ON links(source_type);
