package com.alximac.knowledgegraph.texthandler.domain.model;

import java.util.Collections;
import java.util.Map;
// One semantic text fragment. Stores text, index, metadata, and noteId (after note creation).
public record DocumentChunk(
        String text,
        int index,//index of chunk
        Map<String, Object> metadata,

        String noteId
) {

    public DocumentChunk {
        if (text == null || text.isBlank()) throw new IllegalArgumentException("Text must not be null or empty");


        if (index < 0) throw new IllegalArgumentException("Index of the chunk must be positive, got " + index);

        metadata = metadata != null ? Collections.unmodifiableMap(metadata) : Map.of();
    }

    public DocumentChunk withNoteId(String noteId) {
        return new DocumentChunk(text, index, metadata, noteId);
    }
}
