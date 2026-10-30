-- Link suppressions: explicit human rejection "these notes are not related".
-- The pair is normalized (smaller uuid first) — a rejection applies to the
-- pair, not to a direction. link_type NULL means "no link between them at all".
CREATE TABLE link_suppressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_a_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    note_b_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    link_type TEXT,
    creator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- UNIQUE (note_a_id, note_b_id, link_type) with NULL treated as a value:
-- plain UNIQUE would allow duplicate NULL rows, so the index uses COALESCE.
CREATE UNIQUE INDEX idx_link_suppressions_pair
    ON link_suppressions (note_a_id, note_b_id, COALESCE(link_type, ''));

CREATE INDEX idx_link_suppressions_note_a ON link_suppressions(note_a_id);
CREATE INDEX idx_link_suppressions_note_b ON link_suppressions(note_b_id);
