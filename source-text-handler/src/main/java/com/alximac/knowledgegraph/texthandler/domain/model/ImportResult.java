package com.alximac.knowledgegraph.texthandler.domain.model;

import java.time.Instant;
import java.util.List;
//итог работы всей задачи: статус (COMPLETED, FAILED, PARTIAL),
// списки созданных noteIds и связей, список ошибок, время завершения. Именно этот объект летит в ответную очередь.
public record ImportResult(
        String correlationId,
        Status status,
        List<String> noteIds,
        List<Link> links,
        List<String> errors,
        Instant processedAt
) {
    public ImportResult {
        if (correlationId == null || correlationId.isBlank()) throw new IllegalArgumentException(
                "Correlation id must not be null or blank");

        if (status == null) throw new IllegalArgumentException("Status must not be null");

        if (noteIds == null)throw new IllegalArgumentException("NoteIds must not be null");
        if (links == null)throw new IllegalArgumentException("Links must not be null");
        if (errors == null)throw new IllegalArgumentException("Errors must not be null");

        if (processedAt == null) throw new IllegalArgumentException("ProcessedAt must not be null");


        noteIds = List.copyOf(noteIds);
        links = List.copyOf(links);
        errors = List.copyOf(errors);
    }


    public enum Status {
        COMPLETED,
        PARTIAL,
        FAILED
    }


}

