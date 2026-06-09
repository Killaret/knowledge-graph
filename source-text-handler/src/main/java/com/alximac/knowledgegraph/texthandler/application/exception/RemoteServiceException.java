package com.alximac.knowledgegraph.texthandler.application.exception;

public class RemoteServiceException extends Exception{
    public RemoteServiceException(String message) {
        super(message);
    }

    public RemoteServiceException(String message, Throwable cause) {
        super(message, cause);
    }
}
