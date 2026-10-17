package com.alximac.knowledgegraph.texthandler.infrastructure.validation;

import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;

public class UrlValidationException extends DocumentParserException {
    public UrlValidationException(String message) {
        super(message);
    }

    public UrlValidationException(String message, Throwable cause) {
        super(message, cause);
    }
}
