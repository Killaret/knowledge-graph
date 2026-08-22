package com.alximac.knowledgegraph.texthandler.domain.model;

import java.util.Collections;
import java.util.Map;

public record ImportTask { // incoming task from queue. Transfers event from Go backend.
        String eventId,
        String correlationId,
        TaskType type,
        String content,
        String contentType,
        ImportOptions importOptions,
        Map<String, Object> metadata) {

    private static final int MAX_CONTENT_LENGTH = 30_000_000;

    public ImportTask {
        if (eventId == null || eventId.isBlank()) throw new IllegalArgumentException("" +
                "EventId must not be null or empty");

        if (correlationId == null || correlationId.isBlank()) throw new IllegalArgumentException(
                "CorrelationId is null or empty");

        if (type == null) throw new IllegalArgumentException("TaskType is null");

        if (content == null || content.isBlank()) throw new IllegalArgumentException("Content is null or empty");

        if (importOptions == null) throw new IllegalArgumentException("ImportOptions is null" +
                ", no chunk|overlap validation");

        if (type == TaskType.FILE && contentType == null) throw new IllegalArgumentException(
                "For FILE type contentType can't be null ");

        if (content.length() > MAX_CONTENT_LENGTH) {
            throw new IllegalArgumentException(
                    "Content too large: " + content.length() + " chars, maximum is " + MAX_CONTENT_LENGTH);
        }

        metadata = metadata != null ? Collections.unmodifiableMap(metadata) : Map.of();//defence copying

    }
}



