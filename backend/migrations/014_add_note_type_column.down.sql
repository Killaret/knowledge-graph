-- Remove type column from notes table

DROP INDEX IF EXISTS idx_notes_type;
ALTER TABLE notes DROP COLUMN IF EXISTS type;
