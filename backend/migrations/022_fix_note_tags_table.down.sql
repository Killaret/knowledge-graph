-- Migration 022 down: Revert note_tags to old structure
-- WARNING: This will lose tag associations since old structure used TEXT tags

BEGIN;

DROP TABLE IF EXISTS note_tags CASCADE;

CREATE TABLE note_tags (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (note_id, tag)
);

CREATE INDEX idx_note_tags_tag ON note_tags(tag);

COMMIT;
