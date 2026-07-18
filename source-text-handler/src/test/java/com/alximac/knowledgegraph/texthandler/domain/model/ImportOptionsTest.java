package com.alximac.knowledgegraph.texthandler.domain.model;


import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import static org.assertj.core.api.Assertions.*;

class ImportOptionsTest {

    @Test
    @DisplayName("Must be created with valid values")
    void mustCreateWithValidValues() {
        ImportOptions opts = new ImportOptions(500, 50, 100, 3, true);
        assertThat(opts.chunkSize()).isEqualTo(500);
        assertThat(opts.overlap()).isEqualTo(50);
        assertThat(opts.minChunkLength()).isEqualTo(100);
        assertThat(opts.maxRetries()).isEqualTo(3);
        assertThat(opts.createLinks()).isTrue();
    }

    @Test
    @DisplayName("Must apply defaults when nulls presented")
    void mustApplyDefaults() {
        ImportOptions opts = new ImportOptions(500, 50, null, null, null);
        assertThat(opts.maxRetries()).isEqualTo(3);
        assertThat(opts.createLinks()).isFalse();
        assertThat(opts.minChunkLength()).isEqualTo(50);
    }

    @Test
    @DisplayName("Must be thrown when overlap >= chunkSize")
    void mustThrowWhenOverlapTooBig() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportOptions(100, 100, 50, 3, false))
                .withMessageContaining("overlap is bigger or equal to chunkSize");
    }

    @Test
    @DisplayName("Must be thrown when minChunkLength > chunkSize")
    void mustThrowWhenMinLengthTooBig() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new ImportOptions(100, 50, 500, 3, false))
                .withMessageContaining("minChunkLength can't be bigger than chunkSize");
    }
}