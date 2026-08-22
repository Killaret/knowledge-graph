package com.alximac.knowledgegraph.texthandler.infrastructure.http;

import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;
import io.github.resilience4j.circuitbreaker.CircuitBreaker;
import io.github.resilience4j.retry.Retry;

import java.util.concurrent.Callable;


public class ResilientNoteCreator implements NoteCreatorPort {

    private final NoteCreatorPort noteCreatorPort;
    private final Retry retry;
    private final CircuitBreaker circuitBreaker;

    public ResilientNoteCreator(NoteCreatorPort noteCreatorPort, Retry retry, CircuitBreaker circuitBreaker) {
        this.noteCreatorPort = noteCreatorPort;
        this.retry = retry;
        this.circuitBreaker = circuitBreaker;
    }

    @Override
    public String createNote(DocumentChunk chunk) throws RemoteServiceException {
        Callable<String> callable = () -> noteCreatorPort.createNote(chunk);
        Callable<String> decorated = Retry.decorateCallable(retry, callable);
        decorated = CircuitBreaker.decorateCallable(circuitBreaker, decorated);

        try {
            return decorated.call();
        } catch (RemoteServiceException e) {
            throw e;
        } catch (Exception e) {
            throw new RemoteServiceException("Failed to create note", e);
        }
    }

    @Override
    public void createLink(Link link) throws RemoteServiceException {
        Callable<Void> callable = () -> {
            noteCreatorPort.createLink(link);
            return null;
        };

        Callable<Void> decorated = Retry.decorateCallable(retry, callable);
        decorated = CircuitBreaker.decorateCallable(circuitBreaker, decorated);

        try {
            decorated.call();
        } catch (RemoteServiceException e) {
            throw e;
        } catch (Exception e) {
            throw new RemoteServiceException("Failed to create link", e);
        }
    }
}
