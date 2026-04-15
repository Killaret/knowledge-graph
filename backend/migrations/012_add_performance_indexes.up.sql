-- Performance indexes for large datasets
-- These indexes significantly improve query performance for 10k+ records

-- Indexes for links table - speeds up graph traversal queries
CREATE INDEX IF NOT EXISTS idx_links_source_id ON links(source_id);
CREATE INDEX IF NOT EXISTS idx_links_target_id ON links(target_id);

-- Index for notes sorting - speeds up ORDER BY created_at queries
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

-- Index for notes type filtering - speeds up type-based filtering
CREATE INDEX IF NOT EXISTS idx_notes_type ON notes(type);
