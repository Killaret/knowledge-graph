package com.alximac.knowledgegraph.texthandler.domain.model;

import java.util.Collections;
import java.util.Map;

// один смысловой фрагмент текста. Хранит текст, индекс, метаданные и noteId (после создания заметки).
public record DocumentChunk(
        String text,
        int index,//index of chunk
        Map<String, Object> metadata,

        String noteId,
        String type
) {
    public static final int MAX_TEXT_LENGTH = 10_000;

    public DocumentChunk {
        if (text == null || text.isBlank()) throw new IllegalArgumentException("Text must not be null or empty");


        if (index < 0) throw new IllegalArgumentException("Index of the chunk must be positive, got " + index);

        if (text.length() > MAX_TEXT_LENGTH)
            throw new IllegalArgumentException("Chunk text exceeds 10000 characters, got: " + text.length());

        metadata = metadata != null ? Collections.unmodifiableMap(metadata) : Map.of();

        if (type == null || type.isBlank()) type = "unknown";
    }

    public DocumentChunk withNoteId(String noteId) {
        return new DocumentChunk(text, index, metadata, noteId, type);
    }
}
