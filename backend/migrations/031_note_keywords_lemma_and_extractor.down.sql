-- The keyword column is NOT reverted: the recompute replaced surface forms
-- with lemmas and there is no inverse operation.
ALTER TABLE note_keywords
    DROP COLUMN extractor,
    DROP COLUMN surface;
