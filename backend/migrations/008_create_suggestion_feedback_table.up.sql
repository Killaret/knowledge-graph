CREATE TABLE suggestion_feedback (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    suggested_note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    feedback_type TEXT NOT NULL CHECK (feedback_type IN ('like', 'dislike')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, source_note_id, suggested_note_id)
);

CREATE INDEX idx_feedback_source ON suggestion_feedback(source_note_id);