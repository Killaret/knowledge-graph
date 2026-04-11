-- Add full-text search support for notes table
-- This migration adds tsvector column and triggers for Russian language search

-- Add tsvector column for storing search tokens
ALTER TABLE notes ADD COLUMN search_vector tsvector;

-- Create GIN index for fast full-text search
-- GIN is optimal for tsvector and speeds up @@ operator significantly
CREATE INDEX notes_search_idx ON notes USING GIN(search_vector);

-- Function to automatically update search_vector on insert/update
-- setweight sets weights: 'A' for title (more important), 'B' for content
-- coalesce handles NULL values gracefully
CREATE OR REPLACE FUNCTION notes_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.content, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger that calls the function before insert or update
CREATE TRIGGER notes_search_vector_trigger
    BEFORE INSERT OR UPDATE ON notes
    FOR EACH ROW EXECUTE FUNCTION notes_search_vector_update();

-- Update existing records to populate search_vector
UPDATE notes SET search_vector =
    setweight(to_tsvector('russian', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(content, '')), 'B');
