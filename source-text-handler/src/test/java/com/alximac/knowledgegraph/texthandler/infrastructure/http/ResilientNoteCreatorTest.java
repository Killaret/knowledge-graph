package com.alximac.knowledgegraph.texthandler.infrastructure.http;

import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;
import io.github.resilience4j.circuitbreaker.CircuitBreaker;
import io.github.resilience4j.circuitbreaker.CircuitBreakerConfig;
import io.github.resilience4j.retry.Retry;
import io.github.resilience4j.retry.RetryConfig;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.time.Duration;
import java.util.Map;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

class ResilientNoteCreatorTest {

    private NoteCreatorPort delegate;
    private Retry retry;
    private CircuitBreaker circuitBreaker;
    private ResilientNoteCreator resilient;

    @BeforeEach
    void setUp() {
        delegate = mock(NoteCreatorPort.class);

        // Retry: 3 попытки, задержка 100 мс, ретраить только RemoteServiceException
        retry = Retry.of("test-retry", RetryConfig.custom()
                .maxAttempts(3)
                .waitDuration(Duration.ofMillis(100))
                .retryExceptions(RemoteServiceException.class)
                .build());

        // CircuitBreaker: откроется после 2 неудачных вызовов
        circuitBreaker = CircuitBreaker.of("test-cb", CircuitBreakerConfig.custom()
                .failureRateThreshold(50)
                .slidingWindowSize(2)
                .minimumNumberOfCalls(2)
                .waitDurationInOpenState(Duration.ofMillis(50))
                .permittedNumberOfCallsInHalfOpenState(1)
                .build());

        resilient = new ResilientNoteCreator(delegate, retry, circuitBreaker);
    }

    @Test
    @DisplayName("Must retry on RemoteServiceException and then succeed")
    void mustRetryAndThenSucceed() throws Exception {
        DocumentChunk chunk = new DocumentChunk("text", 0, Map.of(), null);

        // Первые два вызова падают, третий успешен
        when(delegate.createNote(chunk))
                .thenThrow(new RemoteServiceException("first fail"))
                .thenThrow(new RemoteServiceException("second fail"))
                .thenReturn("note-123");

        String noteId = resilient.createNote(chunk);

        assertThat(noteId).isEqualTo("note-123");
        // вызывался ровно 3 раза
        verify(delegate, times(3)).createNote(chunk);
    }

    @Test
    @DisplayName("Must open circuit after repeated failures and fail fast")
    void mustOpenCircuitAfterFailures() throws Exception {
        DocumentChunk chunk = new DocumentChunk("text", 0, Map.of(), null);

        // всегда падение происхдит
        when(delegate.createNote(chunk)).thenThrow(new RemoteServiceException("fail"));

        // Первый вызов: цепь до сих пор закрыта, должны пройти 3 попытки
        assertThatThrownBy(() -> resilient.createNote(chunk))
                .isInstanceOf(RemoteServiceException.class);

        // Второй вызов: цепь должна записать достаточно ошибок и быть открыта
        assertThatThrownBy(() -> resilient.createNote(chunk))
                .isInstanceOf(RemoteServiceException.class);

        // После двух вызовов цепь должна быть открыта
        assertThat(circuitBreaker.getState()).isEqualTo(CircuitBreaker.State.OPEN);

        //  цепь открыта, сразу должно быть исключение
        assertThatThrownBy(() -> resilient.createNote(chunk))
                .isInstanceOf(RemoteServiceException.class);

        // Вызвался 6 раз - 2 на 3
        verify(delegate, times(6)).createNote(chunk);
    }

    @Test
    @DisplayName("Must retry link creation and then succeed")
    void mustRetryLinkAndThenSucceed() throws Exception {
        Link link = new Link("s", "t", 0.5);

        // Первая попытка падает, вторая успешна
        doThrow(new RemoteServiceException("link fail"))
                .doNothing()
                .when(delegate).createLink(link);

        resilient.createLink(link);

        verify(delegate, times(2)).createLink(link);
    }

    @Test
    @DisplayName("Must throw RemoteServiceException when link retries are exhausted")
    void mustThrowWhenLinkRetriesExhausted() throws Exception {
        Link link = new Link("s", "t", 0.5);
        doThrow(new RemoteServiceException("link fail"))
                .when(delegate).createLink(link);

        assertThatThrownBy(() -> resilient.createLink(link))
                .isInstanceOf(RemoteServiceException.class);

        verify(delegate, times(3)).createLink(link);
    }

    @Test
    @DisplayName("Must not retry on non-RemoteServiceException and wrap it")
    void mustNotRetryOnOtherException() throws Exception {
        DocumentChunk chunk = new DocumentChunk("text", 0, Map.of(), null);
        when(delegate.createNote(chunk)).thenThrow(new RuntimeException("unexpected"));

        assertThatThrownBy(() -> resilient.createNote(chunk))
                .isInstanceOf(RemoteServiceException.class)
                .hasMessageContaining("Failed to create note")
                .hasCauseInstanceOf(RuntimeException.class);

        verify(delegate, times(1)).createNote(chunk); // только одна попытка
    }

    @Test
    @DisplayName("Must open circuit after link failures and fail fast")
    void mustOpenCircuitAfterLinkFailures() throws Exception {
        Link link = new Link("s", "t", 0.5);
        doThrow(new RemoteServiceException("link fail")).when(delegate).createLink(link);

        // первый вызов – цепь закрыта 3 попытки
        assertThatThrownBy(() -> resilient.createLink(link))
                .isInstanceOf(RemoteServiceException.class);

        // второй вызов – закрыта, 3 попытки, после чего CB открывается
        assertThatThrownBy(() -> resilient.createLink(link))
                .isInstanceOf(RemoteServiceException.class);

        assertThat(circuitBreaker.getState()).isEqualTo(CircuitBreaker.State.OPEN);

        // третий вызов – делегат не должен вызываться
        assertThatThrownBy(() -> resilient.createLink(link))
                .isInstanceOf(RemoteServiceException.class);

        verify(delegate, times(6)).createLink(link); // 2 * 3 попытки
    }

    @Test
    @DisplayName("Must allow calls after circuit half-open wait duration")
    void mustAllowAfterOpenStateTimeout() throws Exception {
        DocumentChunk chunk = new DocumentChunk("text", 0, Map.of(), null);
        when(delegate.createNote(chunk)).thenThrow(new RemoteServiceException("fail"));

        // два вызова, чтобы открыть цепь
        assertThatThrownBy(() -> resilient.createNote(chunk)).isInstanceOf(RemoteServiceException.class);
        assertThatThrownBy(() -> resilient.createNote(chunk)).isInstanceOf(RemoteServiceException.class);
        assertThat(circuitBreaker.getState()).isEqualTo(CircuitBreaker.State.OPEN);

        // ждём waitDurationInOpenState
        Thread.sleep(200);

        //на успех
        doReturn("success").when(delegate).createNote(chunk);

        String id = resilient.createNote(chunk);
        assertThat(id).isEqualTo("success");

        // После успеха цепь должна закрываться
        assertThat(circuitBreaker.getState()).isEqualTo(CircuitBreaker.State.CLOSED);
    }
}