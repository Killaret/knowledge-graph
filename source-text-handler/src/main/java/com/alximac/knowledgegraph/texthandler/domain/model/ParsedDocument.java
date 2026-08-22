package com.alximac.knowledgegraph.texthandler.domain.model;

import java.util.Collections;
import java.util.Map;

public record ParsedDocument(
        String text,
        Map<String, Object> metaData
) {
    public ParsedDocument{
        if (text == null || text.isBlank()) throw new IllegalArgumentException("Text must not be null or blank");
        if (metaData == null) throw new IllegalArgumentException("MetaData must not be null");
        metaData = Collections.unmodifiableMap(metaData);
    }
}
