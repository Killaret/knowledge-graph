CREATE TABLE IF NOT EXISTS note_recommendations (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    recommended_note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    score REAL NOT NULL CHECK (score >= 0 AND score <= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (note_id, recommended_note_id)
);

CREATE INDEX idx_note_recs_note ON note_recommendations(note_id);
CREATE INDEX idx_note_recs_recommended ON note_recommendations(recommended_note_id);
CREATE INDEX idx_note_recs_score ON note_recommendations(note_id, score DESC);
CREATE INDEX idx_note_recs_updated ON note_recommendations(updated_at);
