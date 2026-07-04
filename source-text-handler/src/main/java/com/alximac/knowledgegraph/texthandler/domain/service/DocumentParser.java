package com.alximac.knowledgegraph.texthandler.domain.service;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;


import java.util.Map;

// Extracts text and metadata from binaries/URLs.
public interface DocumentParser {
    ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException;
    ParsedDocument parseFromUrl(String url) throws DocumentParserException;

}

