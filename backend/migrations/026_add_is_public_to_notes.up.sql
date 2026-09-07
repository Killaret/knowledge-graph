-- Migration 026: Add is_public column to notes for public graph visibility
ALTER TABLE notes ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_notes_is_public ON notes(is_public) WHERE is_public = true;
