DROP INDEX IF EXISTS idx_notes_comet_upcoming;
ALTER TABLE notes DROP COLUMN IF EXISTS done_at;
ALTER TABLE notes DROP COLUMN IF EXISTS remind_before_seconds;
ALTER TABLE notes DROP COLUMN IF EXISTS due_at;
