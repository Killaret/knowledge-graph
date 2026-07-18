package com.alximac.knowledgegraph.texthandler.domain.model;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import java.util.Map;
import static org.assertj.core.api.Assertions.*;

class ImportTaskTest {

    private static final String VALID_EVENT_ID = "evt-123";
    private static final String VALID_CORRELATION_ID = "corr-123";
    private static final TaskType VALID_TYPE = TaskType.TEXT;
    private static final String VALID_CONTENT = "Valid content";
    private static final ImportOptions VALID_OPTIONS = new ImportOptions(500, 50, 100, 3, false);

    @Test
    @DisplayName("Must create with valid values")
    void mustCreateWithValidValues() {
        ImportTask task = new ImportTask(
                VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE,
                VALID_CONTENT, null, VALID_OPTIONS, null
        );
        assertThat(task.eventId()).isEqualTo(VALID_EVENT_ID);
        assertThat(task.correlationId()).isEqualTo(VALID_CORRELATION_ID);
        assertThat(task.type()).isEqualTo(VALID_TYPE);
        assertThat(task.content()).isEqualTo(VALID_CONTENT);
        assertThat(task.importOptions()).isSameAs(VALID_OPTIONS);
        assertThat(task.metadata()).isEqualTo(Map.of());
    }

    @Test
    @DisplayName("Must reject null eventId")
    void mustRejectNullEventId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(null, VALID_CORRELATION_ID, VALID_TYPE, VALID_CONTENT, null, VALID_OPTIONS, null))
                .withMessageContaining("EventId must not be null or empty");
    }

    @Test
    @DisplayName("Must reject blank eventId")
    void mustRejectBlankEventId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask("  ", VALID_CORRELATION_ID, VALID_TYPE, VALID_CONTENT, null, VALID_OPTIONS, null))
                .withMessageContaining("EventId must not be null or empty");
    }

    @Test
    @DisplayName("Must reject null correlationId")
    void mustRejectNullCorrelationId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, null, VALID_TYPE, VALID_CONTENT, null, VALID_OPTIONS, null))
                .withMessageContaining("CorrelationId is null or empty");
    }

    @Test
    @DisplayName("Must reject blank correlationId")
    void mustRejectBlankCorrelationId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, " ", VALID_TYPE, VALID_CONTENT, null, VALID_OPTIONS, null))
                .withMessageContaining("CorrelationId is null or empty");
    }

    @Test
    @DisplayName("Must reject null type")
    void mustRejectNullType() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, null, VALID_CONTENT, null, VALID_OPTIONS, null))
                .withMessageContaining("TaskType is null");
    }

    @Test
    @DisplayName("Must reject null content")
    void mustRejectNullContent() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE, null, null, VALID_OPTIONS, null))
                .withMessageContaining("Content is null or empty");
    }

    @Test
    @DisplayName("Must reject blank content")
    void mustRejectBlankContent() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE, " ", null, VALID_OPTIONS, null))
                .withMessageContaining("Content is null or empty");
    }

    @Test
    @DisplayName("Must reject null importOptions")
    void mustRejectNullImportOptions() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE, VALID_CONTENT, null, null, null))
                .withMessageContaining("ImportOptions is null");
    }

    @Test
    @DisplayName("Must require contentType when type is FILE")
    void mustRequireContentTypeForFile() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, TaskType.FILE, VALID_CONTENT, null, VALID_OPTIONS, null))
                .withMessageContaining("For FILE type contentType can't be null");
    }

    @Test
    @DisplayName("Must reject too long content")
    void mustRejectTooLongContent() {
        String longContent = "x".repeat(30_000_001); // больше MAX_CONTENT_LENGTH (30_000_000)
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE, longContent, null, VALID_OPTIONS, null))
                .withMessageContaining("Content too large");
    }

    @Test
    @DisplayName("Must allow content of exactly MAX_CONTENT_LENGTH chars")
    void mustAllowMaxContentLength() {
        String maxContent = "x".repeat(30_000_000);
        assertThatCode(() -> new ImportTask(
                VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE,
                maxContent, null, VALID_OPTIONS, null))
                .doesNotThrowAnyException();
    }

    @Test
    @DisplayName("Must copy metadata to unmodifiable map")
    void mustCopyMetadataToUnmodifiable() {
        Map<String, Object> meta = new java.util.HashMap<>(Map.of("key1", "value1"));
        ImportTask task = new ImportTask(VALID_EVENT_ID, VALID_CORRELATION_ID, VALID_TYPE, VALID_CONTENT, null, VALID_OPTIONS, meta);
        // Проверяем, что вернулась не та же ссылка и карта защищена
        assertThat(task.metadata()).containsEntry("key1", "value1");
        assertThatThrownBy(() -> task.metadata().put("key2", "value2"))
                .isInstanceOf(UnsupportedOperationException.class);
    }
}