package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;

import java.util.Map;

public class TikaDocumentParser implements DocumentParser {
    @Override
    public ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException {
        return null;
    }

    @Override
    public ParsedDocument parseFromUrl(String url) throws DocumentParserException {
        return null;
    }
}
