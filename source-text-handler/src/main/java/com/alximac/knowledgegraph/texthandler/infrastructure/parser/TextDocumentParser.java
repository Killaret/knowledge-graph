package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;
import com.alximac.knowledgegraph.texthandler.infrastructure.validation.UrlValidator;

import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.Map;

public class TextDocumentParser implements DocumentParser {
    @Override
    public ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException {
        String text = new String(data, StandardCharsets.UTF_8);
        return new ParsedDocument(text, sourceMetadata != null ? sourceMetadata : Map.of());

    }

    @Override
    public ParsedDocument parseFromUrl(String url) throws DocumentParserException {
        UrlValidator.validate(url);
        try (InputStream stream = URI.create(url).toURL().openStream()) {
            String text = new String(stream.readAllBytes(),StandardCharsets.UTF_8);
            return  new ParsedDocument(text,Map.of());
        }catch (IOException|IllegalArgumentException e){
            throw new DocumentParserException("Failed to read text from URL: " + url, e);
        }
    }
}
