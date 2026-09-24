package com.alximac.knowledgegraph.texthandler.infrastructure.http;

import java.util.Map;

public record CreateNoteRequest(
        String title,
        String content,
        String type,
        Map<String,Object> metadata
) {

}
