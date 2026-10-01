-- COMET-1 stage A: a comet is a task ("дело") — optional due date, optional
-- reminder offset, and a done marker. Real columns (not metadata) so the
-- "upcoming" list can index due_at. Reminder offset is stored in seconds to
-- keep the column interval-free and language-neutral.
ALTER TABLE notes ADD COLUMN IF NOT EXISTS due_at timestamptz;
ALTER TABLE notes ADD COLUMN IF NOT EXISTS remind_before_seconds bigint;
ALTER TABLE notes ADD COLUMN IF NOT EXISTS done_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_notes_comet_upcoming
    ON notes(creator_id, due_at)
    WHERE deleted_at IS NULL AND done_at IS NULL AND due_at IS NOT NULL;
