package com.alximac.knowledgegraph.texthandler.infrastructure.http;

public record CreateNoteResponse(
        NoteData data,
        String message
) {
    public record NoteData(
            String id
    ) {
    }
}


