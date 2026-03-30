CREATE TABLE note_embeddings (
    note_id UUID PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
    embedding vector(384),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Индекс для быстрого поиска по косинусному сходству
CREATE INDEX idx_note_embeddings_vector ON note_embeddings USING ivfflat (embedding vector_cosine_ops);