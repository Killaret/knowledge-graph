package com.alximac.knowledgegraph.texthandler.domain.service;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;


import java.util.Map;
// извлекает текст и метаданные из бинарников/URL.
public interface DocumentParser {
    ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException;
    ParsedDocument parseFromUrl(String url) throws DocumentParserException;

}

