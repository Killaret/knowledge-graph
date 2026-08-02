package com.alximac.knowledgegraph.texthandler.application;

import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.application.port.OutboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.model.*;
import com.alximac.knowledgegraph.texthandler.domain.repository.ImportStateRepository;
import com.alximac.knowledgegraph.texthandler.domain.service.*;
import com.alximac.knowledgegraph.texthandler.infrastructure.parser.ParserFactory;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.BDDMockito.*;

@ExtendWith(MockitoExtension.class)
class ImportDocumentHandlerTest {

    @Mock private ChunkingStrategy chunkingStrategy;
    @Mock private ParserFactory parserFactory;
    @Mock private DocumentParser documentParser; // заглушка, которую вернёт фабрика
    @Mock private LinkDetector linkDetector;
    @Mock private NoteCreatorPort noteCreatorPort;
    @Mock private OutboundQueuePort outboundQueuePort;
    @Mock private ImportStateRepository stateRepository;

    private ImportDocumentHandler handler;

    private final ImportTask validTask = new ImportTask(
            "evt-1", "corr-1", TaskType.TEXT, "valid content",
            null, new ImportOptions(500, 50, 100, 3,
            false), null
    );

    private final DocumentChunk sampleChunk = new DocumentChunk("sample text", 0, Map.of(), null);

    @BeforeEach
    void setUp() {
        handler = new ImportDocumentHandler(
                chunkingStrategy, parserFactory, linkDetector,
                noteCreatorPort, outboundQueuePort, stateRepository
        );
    }

    @Test
    @DisplayName("Must complete successfully when all notes created")
    void mustCompleteSuccessfully() throws Exception {
        given(stateRepository.tryClaim("evt-1")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(byte[].class), any())).willReturn(
                new ParsedDocument("text", Map.of())
        );
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of(sampleChunk));
        given(noteCreatorPort.createNote(sampleChunk)).willReturn("note-1");


        handler.handle(validTask);

        ArgumentCaptor<ImportResult> resultCaptor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-1"), resultCaptor.capture());
        ImportResult result = resultCaptor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.COMPLETED);
        assertThat(result.noteIds()).containsExactly("note-1");
        assertThat(result.errors()).isEmpty();

        then(outboundQueuePort).should().publish(result);
    }

    @Test
    @DisplayName("Must set PARTIAL status when some notes fail")
    void mustSetPartialStatusWhenSomeNotesFail() throws Exception {
        DocumentChunk chunk1 = new DocumentChunk("text1", 0, Map.of(), null);
        DocumentChunk chunk2 = new DocumentChunk("text2", 1, Map.of(), null);

        given(stateRepository.tryClaim("evt-1")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(byte[].class), any())).willReturn(
                new ParsedDocument("text", Map.of())
        );
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of(chunk1, chunk2));
        given(noteCreatorPort.createNote(chunk1)).willReturn("note-1");
        given(noteCreatorPort.createNote(chunk2)).willThrow(new RemoteServiceException("Fail"));
        // linkDetector не используется

        handler.handle(validTask);

        ArgumentCaptor<ImportResult> resultCaptor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-1"), resultCaptor.capture());
        ImportResult result = resultCaptor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.PARTIAL);
        assertThat(result.noteIds()).containsExactly("note-1");
        assertThat(result.errors()).hasSize(1).first().asString().contains("Fail");

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must set FAILED status when all notes fail")
    void mustSetFailedStatusWhenAllNotesFail() throws Exception {
        DocumentChunk chunk1 = new DocumentChunk("text1", 0, Map.of(), null);

        given(stateRepository.tryClaim("evt-1")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(byte[].class), any())).willReturn(
                new ParsedDocument("text", Map.of())
        );
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of(chunk1));
        given(noteCreatorPort.createNote(chunk1)).willThrow(new RemoteServiceException("Fail"));

        handler.handle(validTask);

        ArgumentCaptor<ImportResult> resultCaptor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-1"), resultCaptor.capture());
        ImportResult result = resultCaptor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.FAILED);
        assertThat(result.noteIds()).isEmpty();
        assertThat(result.errors()).hasSize(1);

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must skip processing when task already processed")
    void mustSkipWhenAlreadyProcessed() {
        given(stateRepository.tryClaim("evt-1")).willReturn(false);

        handler.handle(validTask);

        // Никакие порты не должны быть вызваны
        then(parserFactory).shouldHaveNoInteractions();
        then(chunkingStrategy).shouldHaveNoInteractions();
        then(noteCreatorPort).shouldHaveNoInteractions();
    }

    @Test
    @DisplayName("Must catch DocumentParserException and set FAILED")
    void mustCatchParseExceptionAndSetFailed() throws Exception {
        given(stateRepository.tryClaim("evt-1")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(byte[].class), any()))
                .willThrow(new DocumentParserException("Parse failed"));

        handler.handle(validTask);

        ArgumentCaptor<ImportResult> resultCaptor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-1"), resultCaptor.capture());
        ImportResult result = resultCaptor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.FAILED);
        assertThat(result.errors()).first().asString().contains("Parse failed");

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must create links when enabled and more than one chunk")
    void mustCreateLinksWhenEnabled() throws Exception {
        ImportTask taskWithLinks = new ImportTask(
                "evt-2", "corr-2", TaskType.TEXT, "content",
                null, new ImportOptions(500, 50, 100, 3, true), null
        );

        DocumentChunk chunk1 = new DocumentChunk("text1", 0, Map.of(), null);
        DocumentChunk chunk2 = new DocumentChunk("text2", 1, Map.of(), null);
        Link link = new Link("note-1", "note-2", 0.95); // предположим, конструктор Link

        given(stateRepository.tryClaim("evt-2")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(), any())).willReturn(new ParsedDocument("text", Map.of()));
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of(chunk1, chunk2));
        given(noteCreatorPort.createNote(chunk1)).willReturn("note-1");
        given(noteCreatorPort.createNote(chunk2)).willReturn("note-2");
        given(linkDetector.detectLinks(any())).willReturn(List.of(link));

        handler.handle(taskWithLinks);

        ArgumentCaptor<ImportResult> captor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-2"), captor.capture());
        ImportResult result = captor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.COMPLETED);
        assertThat(result.links()).containsExactly(link);
        then(linkDetector).should().detectLinks(any()); // проверяем, что детектор вызывался

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must set PARTIAL if link creation fails")
    void mustSetPartialWhenLinkCreationFails() throws Exception {
        ImportTask task = new ImportTask(
                "evt-3", "corr-3", TaskType.TEXT, "content",
                null, new ImportOptions(500, 50, 100, 3, true), null
        );
        DocumentChunk chunk1 = new DocumentChunk("t1", 0, Map.of(), null);
        DocumentChunk chunk2 = new DocumentChunk("t2", 1, Map.of(), null);
        Link link = new Link("note-1", "note-2", 0.65);

        given(stateRepository.tryClaim("evt-3")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(), any())).willReturn(new ParsedDocument("text", Map.of()));
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of(chunk1, chunk2));
        given(noteCreatorPort.createNote(chunk1)).willReturn("note-1");
        given(noteCreatorPort.createNote(chunk2)).willReturn("note-2");
        given(linkDetector.detectLinks(any())).willReturn(List.of(link));
        willThrow(new RemoteServiceException("Link error")).given(noteCreatorPort).createLink(link);

        handler.handle(task);

        ArgumentCaptor<ImportResult> captor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-3"), captor.capture());
        ImportResult result = captor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.PARTIAL);
        assertThat(result.errors()).anyMatch(e -> e.contains("Link error"));
        assertThat(result.noteIds()).containsExactly("note-1", "note-2");

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must set FAILED when chunking returns empty list")
    void mustFailWhenNoChunksProduced() throws Exception {
        given(stateRepository.tryClaim("evt-1")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(), any())).willReturn(new ParsedDocument("text", Map.of()));
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of());

        handler.handle(validTask);

        ArgumentCaptor<ImportResult> captor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-1"), captor.capture());
        ImportResult result = captor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.FAILED);
        assertThat(result.errors()).first().asString().contains("No chunks produced");

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must not call linkDetector when createLinks=true but only one chunk")
    void mustNotCallLinkDetectorWhenOnlyOneChunk() throws Exception {
        ImportTask taskWithLinks = new ImportTask(
                "evt-4", "corr-4", TaskType.TEXT, "content",
                null, new ImportOptions(500, 50, 100, 3, true), null
        );

        DocumentChunk singleChunk = new DocumentChunk("text1", 0, Map.of(), null);

        given(stateRepository.tryClaim("evt-4")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(byte[].class), any())).willReturn(
                new ParsedDocument("text", Map.of())
        );
        given(chunkingStrategy.chunk(any(), any())).willReturn(List.of(singleChunk));
        given(noteCreatorPort.createNote(singleChunk)).willReturn("note-1");
        // linkDetector не должен вызваться поэтому не настраиваем

        handler.handle(taskWithLinks);

        ArgumentCaptor<ImportResult> captor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-4"), captor.capture());
        ImportResult result = captor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.COMPLETED);
        assertThat(result.noteIds()).containsExactly("note-1");
        // linkDetector не должен был вызываться
        then(linkDetector).should(never()).detectLinks(any());

        then(outboundQueuePort).should().publish(any());
    }

    @Test
    @DisplayName("Must set FAILED on unexpected exception")
    void mustFailOnUnexpectedException() throws Exception {
        given(stateRepository.tryClaim("evt-1")).willReturn(true);
        given(parserFactory.getSelectedParser(TaskType.TEXT)).willReturn(documentParser);
        given(documentParser.parse(any(), any())).willThrow(new RuntimeException("Boom!"));

        handler.handle(validTask);

        ArgumentCaptor<ImportResult> captor = ArgumentCaptor.forClass(ImportResult.class);
        then(stateRepository).should().markProcessed(eq("evt-1"), captor.capture());
        ImportResult result = captor.getValue();
        assertThat(result.status()).isEqualTo(ImportResult.Status.FAILED);
        assertThat(result.errors()).first().asString().contains("Unexpected error");

        then(outboundQueuePort).should().publish(any());
    }
}