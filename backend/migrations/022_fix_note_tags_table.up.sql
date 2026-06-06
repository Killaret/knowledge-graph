-- Migration 022: Fix note_tags table structure
-- Problem: Old migration created note_tags(tag TEXT), but model expects note_tags(tag_id UUID)
-- Solution: Recreate table with correct structure

BEGIN;

-- Drop old table if exists
DROP TABLE IF EXISTS note_tags CASCADE;

-- Recreate with correct structure (many-to-many: notes <-> tags)
CREATE TABLE note_tags (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (note_id, tag_id)
);

CREATE INDEX idx_note_tags_note ON note_tags(note_id);
CREATE INDEX idx_note_tags_tag ON note_tags(tag_id);

COMMIT;
