package com.alximac.knowledgegraph.texthandler.domain.model;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.*;

class DocumentChunkTest {

    @Test
    @DisplayName("Must create with valid values")
    void mustCreateWithValidValues() {
        DocumentChunk chunk = new DocumentChunk("some text", 0, Map.of("lang", "ru"), null);
        assertThat(chunk.text()).isEqualTo("some text");
        assertThat(chunk.index()).isEqualTo(0);
        assertThat(chunk.metadata()).containsEntry("lang", "ru");
        assertThat(chunk.noteId()).isNull();
    }

    @Test
    @DisplayName("Must reject null or blank text")
    void mustRejectNullOrBlankText() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new DocumentChunk(null, 0, Map.of(), null))
                .withMessageContaining("Text must not be null or empty");
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new DocumentChunk("  ", 0, Map.of(), null))
                .withMessageContaining("Text must not be null or empty");
    }

    @Test
    @DisplayName("Must reject negative index")
    void mustRejectNegativeIndex() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new DocumentChunk("text", -1, Map.of(), null))
                .withMessageContaining("Index of the chunk must be positive");
    }

    @Test
    @DisplayName("Must copy metadata to unmodifiable map")
    void mustCopyMetadataToUnmodifiable() {
        Map<String, Object> meta = new java.util.HashMap<>(Map.of("k", "v"));
        DocumentChunk chunk = new DocumentChunk("text", 0, meta, null);
        assertThatThrownBy(() -> chunk.metadata().put("new", "value"))
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    @DisplayName("Must set noteId via withNoteId and keep other fields")
    void mustSetNoteIdViaWithNoteId() {
        DocumentChunk original = new DocumentChunk("text", 0, Map.of(), null);
        DocumentChunk updated = original.withNoteId("note-123");
        assertThat(updated.noteId()).isEqualTo("note-123");
        assertThat(updated.text()).isEqualTo("text");
        assertThat(updated.index()).isEqualTo(0);
        // Исходный чанк не изменился
        assertThat(original.noteId()).isNull();
    }
}