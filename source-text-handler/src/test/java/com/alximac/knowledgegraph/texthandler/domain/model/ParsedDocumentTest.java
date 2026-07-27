package com.alximac.knowledgegraph.texthandler.domain.model;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import java.util.Map;
import static org.assertj.core.api.Assertions.*;

class ParsedDocumentTest {

    @Test
    @DisplayName("Must create with valid text and metadata")
    void mustCreateWithValidValues() {
        ParsedDocument doc = new ParsedDocument("some text", Map.of("lang", "ru"));
        assertThat(doc.text()).isEqualTo("some text");
        assertThat(doc.metaData()).containsEntry("lang", "ru");
    }

    @Test
    @DisplayName("Must reject null or blank text")
    void mustRejectNullOrBlankText() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ParsedDocument(null, Map.of()))
                .withMessageContaining("Text must not be null or blank");
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ParsedDocument("  ", Map.of()))
                .withMessageContaining("Text must not be null or blank");
    }

    @Test
    @DisplayName("Must reject null metadata")
    void mustRejectNullMetadata() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ParsedDocument("text", null))
                .withMessageContaining("MetaData must not be null");
    }

    @Test
    @DisplayName("Must make metadata unmodifiable")
    void mustMakeMetadataUnmodifiable() {
        Map<String, Object> meta = new java.util.HashMap<>(Map.of("k", "v"));
        ParsedDocument doc = new ParsedDocument("text", meta);
        assertThatThrownBy(() -> doc.metaData().put("new", "val"))
                .isInstanceOf(UnsupportedOperationException.class);
    }
}