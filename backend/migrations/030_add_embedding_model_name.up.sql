ALTER TABLE note_embeddings
    ADD COLUMN model_name VARCHAR(255) NOT NULL DEFAULT 'all-MiniLM-L6-v2';

-- Existing rows have already been written by the old (only) model.
-- New code must set the column explicitly, so the default is removed.
ALTER TABLE note_embeddings
    ALTER COLUMN model_name DROP DEFAULT;
