-- Rollback full-text search migration

DROP TRIGGER IF EXISTS notes_search_vector_trigger ON notes;
DROP FUNCTION IF EXISTS notes_search_vector_update();
DROP INDEX IF EXISTS notes_search_idx;
ALTER TABLE notes DROP COLUMN IF EXISTS search_vector;
