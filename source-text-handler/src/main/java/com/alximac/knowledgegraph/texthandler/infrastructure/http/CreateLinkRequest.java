package com.alximac.knowledgegraph.texthandler.infrastructure.http;


public record CreateLinkRequest(
        String sourceNoteId,
        String targetNoteId,
        double weight
) {
}
