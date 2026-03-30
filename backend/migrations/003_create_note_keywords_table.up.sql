CREATE TABLE note_keywords (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    keyword TEXT NOT NULL,
    weight FLOAT,
    PRIMARY KEY (note_id, keyword)
);