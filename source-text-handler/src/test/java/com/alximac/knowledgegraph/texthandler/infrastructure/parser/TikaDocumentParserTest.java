package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.InputStream;
import java.net.URL;
import java.util.Arrays;
import java.util.Map;

import static org.assertj.core.api.Assertions.*;

class TikaDocumentParserTest {

    private TikaDocumentParser parser;

    @BeforeEach
    void setUp() {
        parser = new TikaDocumentParser();
    }

    @Test
    @DisplayName("Must extract text from PDF file")
    void mustExtractTextFromPdf() throws IOException, DocumentParserException {

        byte[] pdfBytes;

        try (InputStream is = getClass().getClassLoader().getResourceAsStream("sample.pdf")) {
            assertThat(is).as("sample.pdf not found in test resources").isNotNull();
            pdfBytes = is.readAllBytes();
        }

        Map<String, Object> sourceMetadata = Map.of("source", "test");
        ParsedDocument doc = parser.parse(pdfBytes, sourceMetadata);

        assertThat(doc.text()).isNotBlank();

        assertThat(doc.metaData()).containsEntry("source", "test");

        assertThatThrownBy(() -> doc.metaData().put("new_key", "value"))
                .isInstanceOf(UnsupportedOperationException.class);
        System.out.println(doc.text());
    }

    @Test
    @DisplayName("Must throw DocumentParserException for  corrupted PDF")
    void mustThrowExceptionForCorruptedPdf() throws IOException {
        byte[] pdfBytes;
        try (InputStream is = getClass().getClassLoader().getResourceAsStream("corrupted.pdf")) {
            assertThat(is).isNotNull();
            pdfBytes = is.readAllBytes();
        }
        byte[] broken = Arrays.copyOf(pdfBytes, pdfBytes.length / 2);

        assertThatThrownBy(() -> parser.parse(broken, Map.of()))
                .isInstanceOf(DocumentParserException.class)
                .hasMessageContaining("Failed to parse");
    }

    @Test
    @DisplayName("Must parse from file URL")
    void mustParseFromFileUrl() throws Exception {
        URL resource = getClass().getClassLoader().getResource("sample.pdf");
        assertThat(resource).isNotNull();
        String fileUrl = resource.toExternalForm();

        ParsedDocument doc = parser.parseFromUrl(fileUrl);

        assertThat(doc.text()).isNotBlank();
        assertThat(doc.metaData()).isNotNull();
        System.out.println(doc.text());
    }

    @Test
    @DisplayName("Must throw DocumentParserException for invalid URL")
    void mustThrowExceptionForInvalidUrl() {
        assertThatThrownBy(() -> parser.parseFromUrl("not-a-valid-url"))
                .isInstanceOf(DocumentParserException.class)
                .hasMessageContaining("Failed to open URL");
    }


    @Test
    @DisplayName("Must merge Tika metadata with source metadata, source have priority")
    void mustMergeMetadataWithSourcePriority() throws IOException, DocumentParserException {
        byte[] pdfBytes;
        try (InputStream is = getClass().getClassLoader().getResourceAsStream("sample.pdf")) {
            assertThat(is).isNotNull();
            pdfBytes = is.readAllBytes();
        }

        Map<String, Object> sourceMeta = Map.of("Author", "MyAuthor");
        ParsedDocument doc = parser.parse(pdfBytes, sourceMeta);

        assertThat(doc.metaData()).containsEntry("Author", "MyAuthor");
        assertThat(doc.metaData()).containsKey("Content-Type");
    }


}