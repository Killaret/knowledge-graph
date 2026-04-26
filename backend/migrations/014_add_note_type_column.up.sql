-- Add type column to notes table for astronomical object classification
-- Types: star, planet, comet, galaxy, asteroid, satellite, debris, nebula

ALTER TABLE notes ADD COLUMN IF NOT EXISTS type VARCHAR(50) NOT NULL DEFAULT 'star';

-- Index for type filtering
CREATE INDEX IF NOT EXISTS idx_notes_type ON notes(type);
