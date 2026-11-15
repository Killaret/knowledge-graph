-- NOTE-DELETE-1: links removed together with a soft-deleted note carry the
-- note's id so that restore can revive exactly those links — links deleted
-- on their own keep deleted_via_note_id NULL and never come back.
ALTER TABLE links ADD COLUMN IF NOT EXISTS deleted_via_note_id uuid;
CREATE INDEX IF NOT EXISTS idx_links_deleted_via_note ON links(deleted_via_note_id) WHERE deleted_via_note_id IS NOT NULL;
