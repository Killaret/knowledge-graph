-- Add multilingual full-text search support (Russian + English)
-- This migration updates the search trigger to support both Russian and English

-- Drop existing trigger and function
DROP TRIGGER IF EXISTS notes_search_vector_trigger ON notes;
DROP FUNCTION IF EXISTS notes_search_vector_update();

-- Create updated function that supports both Russian and English
-- Uses 'simple' configuration for English (no stemming) and 'russian' for Russian
CREATE OR REPLACE FUNCTION notes_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        -- Russian text with stemming
        setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.content, '')), 'B') ||
        -- English text with simple configuration (preserves original words)
        setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(NEW.content, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate trigger
CREATE TRIGGER notes_search_vector_trigger
    BEFORE INSERT OR UPDATE ON notes
    FOR EACH ROW EXECUTE FUNCTION notes_search_vector_update();

-- Update existing records with new multilingual search vectors
UPDATE notes SET search_vector =
    setweight(to_tsvector('russian', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(content, '')), 'B') ||
    setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(content, '')), 'B');
