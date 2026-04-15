-- Rollback performance indexes

DROP INDEX IF EXISTS idx_links_source_id;
DROP INDEX IF EXISTS idx_links_target_id;
DROP INDEX IF EXISTS idx_notes_created_at;
DROP INDEX IF EXISTS idx_notes_type;
