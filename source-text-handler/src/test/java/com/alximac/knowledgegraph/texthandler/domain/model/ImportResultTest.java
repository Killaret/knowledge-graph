package com.alximac.knowledgegraph.texthandler.domain.model;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.time.Instant;
import java.util.List;

import static org.assertj.core.api.Assertions.*;

class ImportResultTest {

    private static final String VALID_CORR_ID = "corr-1";
    private static final Instant NOW = Instant.now();

    @Test
    @DisplayName("Must create with COMPLETED status")
    void mustCreateCompleted() {
        ImportResult result = new ImportResult(
                VALID_CORR_ID, ImportResult.Status.COMPLETED,
                List.of("n1", "n2"), List.of(), List.of(), NOW
        );
        assertThat(result.status()).isEqualTo(ImportResult.Status.COMPLETED);
        assertThat(result.noteIds()).containsExactly("n1", "n2");
        assertThat(result.links()).isEmpty();
        assertThat(result.errors()).isEmpty();
    }

    @Test
    @DisplayName("Must create with PARTIAL status and errors")
    void mustCreatePartial() {
        ImportResult result = new ImportResult(
                VALID_CORR_ID, ImportResult.Status.PARTIAL,
                List.of("n1"), List.of(), List.of("Failed chunk 2"), NOW
        );
        assertThat(result.status()).isEqualTo(ImportResult.Status.PARTIAL);
        assertThat(result.errors()).containsExactly("Failed chunk 2");
    }

    @Test
    @DisplayName("Must reject null fields")
    void mustRejectNullFields() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(null, ImportResult.Status.COMPLETED, List.of(), List.of(), List.of(), NOW))
                .withMessageContaining("Correlation id must not be null or blank");

        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(VALID_CORR_ID, null, List.of(), List.of(), List.of(), NOW))
                .withMessageContaining("Status must not be null");

        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(VALID_CORR_ID, ImportResult.Status.COMPLETED, null, List.of(), List.of(), NOW))
                .withMessageContaining("NoteIds must not be null");

        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(VALID_CORR_ID, ImportResult.Status.COMPLETED, List.of(), null, List.of(), NOW))
                .withMessageContaining("Links must not be null");

        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(VALID_CORR_ID, ImportResult.Status.COMPLETED, List.of(), List.of(), null, NOW))
                .withMessageContaining("Errors must not be null");

        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(VALID_CORR_ID, ImportResult.Status.COMPLETED, List.of(), List.of(), List.of(), null))
                .withMessageContaining("ProcessedAt must not be null");
    }

    @Test
    @DisplayName("Must reject blank correlationId")
    void mustRejectBlankCorrelationId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportResult(" ", ImportResult.Status.COMPLETED, List.of(), List.of(), List.of(), NOW))
                .withMessageContaining("Correlation id must not be null or blank");
    }

    @Test
    @DisplayName("Must make collections unmodifiable")
    void mustMakeCollectionsUnmodifiable() {
        ImportResult result = new ImportResult(
                VALID_CORR_ID, ImportResult.Status.COMPLETED,
                List.of("n1"), List.of(), List.of(), NOW
        );
        assertThatThrownBy(() -> result.noteIds().add("n2"))
                .isInstanceOf(UnsupportedOperationException.class);
        assertThatThrownBy(() -> result.errors().add("error"))
                .isInstanceOf(UnsupportedOperationException.class);
    }
}