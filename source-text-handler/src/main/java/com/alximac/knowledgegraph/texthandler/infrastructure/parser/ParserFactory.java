package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.TaskType;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;

public class ParserFactory {

    private final DocumentParser tikaParser;
    private final DocumentParser textParser;

    public ParserFactory(DocumentParser tikaParser, DocumentParser textParser) {
        this.tikaParser = tikaParser;
        this.textParser = textParser;
    }

    public DocumentParser getSelectedParser (TaskType type){
        return switch (type){
            case FILE, URL -> tikaParser;
            case TEXT -> textParser;
        };
    }
}
