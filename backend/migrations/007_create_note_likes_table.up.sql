CREATE TABLE note_likes (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    like_type TEXT NOT NULL CHECK (like_type IN ('like', 'dislike')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, note_id)
);

CREATE INDEX idx_note_likes_note ON note_likes(note_id);