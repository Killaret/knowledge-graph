package com.alximac.knowledgegraph.texthandler.domain.service;

public class DocumentParserException extends Exception{

    public DocumentParserException(String message) {
        super(message);
    }

    public DocumentParserException(String message, Throwable cause) {
        super(message, cause);
    }
}
