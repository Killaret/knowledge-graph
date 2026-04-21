-- Удаляем поле type
DROP INDEX IF EXISTS idx_notes_type;
ALTER TABLE notes DROP COLUMN IF EXISTS type;
