DROP INDEX IF EXISTS idx_notes_is_public;
ALTER TABLE notes DROP COLUMN IF EXISTS is_public;
