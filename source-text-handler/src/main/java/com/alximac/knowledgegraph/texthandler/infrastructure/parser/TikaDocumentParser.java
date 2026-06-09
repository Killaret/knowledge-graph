package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;
import org.apache.tika.exception.TikaException;
import org.apache.tika.metadata.Metadata;
import org.apache.tika.parser.AutoDetectParser;
import org.apache.tika.sax.BodyContentHandler;
import org.xml.sax.SAXException;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.util.HashMap;
import java.util.Map;

public class TikaDocumentParser implements DocumentParser {
    @Override
    public ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException {
        return parseFromStream(new ByteArrayInputStream(data), sourceMetadata);
    }

    @Override
    public ParsedDocument parseFromUrl(String url) throws DocumentParserException {
        try (InputStream stream = URI.create(url).toURL().openStream()) {
            return parseFromStream(stream, Map.of());
        } catch (IOException | IllegalArgumentException e) {
            throw new DocumentParserException("Failed to open URL: " + url, e);
        }
    }

    private ParsedDocument parseFromStream(InputStream inputStream, Map<String, Object> sourceMetadata)
            throws DocumentParserException {

        final Map<String, Object> meta = sourceMetadata != null ? sourceMetadata : Map.of();

        AutoDetectParser autoDetectParser = new AutoDetectParser();
        BodyContentHandler bodyContentHandler = new BodyContentHandler(-1);

        Metadata metadata = new Metadata();
        meta.forEach((key, value) -> {
            if (value != null) {
                metadata.set(key, value.toString());
            }
        });

        try {
            autoDetectParser.parse(inputStream, bodyContentHandler, metadata);
        } catch (IOException | SAXException | TikaException e) {
            throw new DocumentParserException("Failed to parse given content " + e);
        }

        String text = bodyContentHandler.toString().strip();

        Map<String, Object> combinedMetadata = new HashMap<>();
        for (String name : metadata.names()) {
            combinedMetadata.put(name, metadata.get(name));
        }
        combinedMetadata.putAll(meta);


        return new ParsedDocument(text, combinedMetadata);
    }
}
