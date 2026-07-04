package com.alximac.knowledgegraph.texthandler.domain.model;

public record Link(
        String sourceNoteId,
        String targetNoteId,
        double weight // link strength
) {
    public Link{
        if (sourceNoteId == null || sourceNoteId.isBlank())
            throw new IllegalArgumentException("SourceNoteId must not be null or blank");

        if (targetNoteId == null || targetNoteId.isBlank())
            throw new IllegalArgumentException("TargetNoteId must not be null or blank");

        if (weight < 0.0 || weight > 1.0)
            throw new IllegalArgumentException("Weight must be between 0 and 1, got: " + weight);
    }
}
