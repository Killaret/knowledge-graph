-- NLP-2: keyword becomes the canonical lemma; the surface form (as written
-- in the text) and the extractor name are stored alongside.
ALTER TABLE note_keywords
    ADD COLUMN surface TEXT NOT NULL DEFAULT '',
    ADD COLUMN extractor TEXT NOT NULL DEFAULT 'yake-0.4.8';

-- Existing rows were produced by yake and store surface forms; until the
-- recompute runs, surface = keyword for them.
UPDATE note_keywords SET surface = keyword WHERE surface = '';
