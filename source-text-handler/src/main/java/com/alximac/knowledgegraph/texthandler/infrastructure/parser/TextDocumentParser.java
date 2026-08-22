package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;

import java.nio.charset.StandardCharsets;
import java.util.Map;

public class TextDocumentParser implements DocumentParser {
    @Override
    public ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException {
        String text = new String(data, StandardCharsets.UTF_8);

        return new ParsedDocument(text,sourceMetadata);
    }

    @Override
    public ParsedDocument parseFromUrl(String url) throws DocumentParserException { // stub for now
        throw new DocumentParserException("TextDocumentParser doesn't support links ");
    }
}
