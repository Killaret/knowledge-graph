DROP INDEX IF EXISTS idx_links_deleted_via_note;
ALTER TABLE links DROP COLUMN IF EXISTS deleted_via_note_id;
