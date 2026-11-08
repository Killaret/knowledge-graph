package com.alximac.knowledgegraph.texthandler.application;

import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.application.port.OutboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.model.*;
import com.alximac.knowledgegraph.texthandler.domain.repository.ImportStateRepository;
import com.alximac.knowledgegraph.texthandler.domain.service.ChunkingStrategy;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;
import com.alximac.knowledgegraph.texthandler.infrastructure.parser.ParserFactory;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

import static java.util.Base64.getDecoder;

public class ImportDocumentHandler {//бизнес процесс(парсинг, нарезка, заметка, связи и т.д. . Узнаем про задачу и выполнена ли она.

    private final ChunkingStrategy chunkingStrategy;
    private final ParserFactory parserFactory;
    private final NoteCreatorPort noteCreatorPort;
    private final OutboundQueuePort outboundQueuePort;
    private final ImportStateRepository stateRepository;

    public ImportDocumentHandler(
        ChunkingStrategy chunkingStrategy,
        ParserFactory parserFactory,
        NoteCreatorPort noteCreatorPort,
        OutboundQueuePort outboundQueuePort,
        ImportStateRepository importStateRepository) {

        this.chunkingStrategy = chunkingStrategy;
        this.parserFactory = parserFactory;
        this.noteCreatorPort = noteCreatorPort;
        this.outboundQueuePort = outboundQueuePort;
        this.stateRepository = importStateRepository;
    }

    public void handle(ImportTask task) {

        if (!stateRepository.tryClaim(task.eventId())) {
            return; // задача уже обрабатывается или обработана
        }

        ImportResult result;

        try {

            DocumentParser parser = parserFactory.getSelectedParser(task.type());
            ParsedDocument parsedDocument = switch (task.type()) {
                case FILE -> parser.parse(getDecoder().decode(task.content()), task.metadata());
                case URL -> parser.parseFromUrl(task.content());
                case TEXT -> parser.parse(task.content().getBytes(StandardCharsets.UTF_8), task.metadata());
            };

            List<DocumentChunk> chunks = chunkingStrategy.chunk(parsedDocument.text(), task.importOptions());

            if (chunks.isEmpty()) {
                throw new IllegalStateException("No chunks produced");
            }

            List<String> noteIds = new ArrayList<>();
            List<String> errors = new ArrayList<>();
            List<Link> links = List.of();

            for (DocumentChunk chunk : chunks) {
                try {
                    String noteId = noteCreatorPort.createNote(chunk, task.jwt());
                    noteIds.add(noteId);

                } catch (RemoteServiceException e) {

                    errors.add("Failed to create note for chunk " + chunk.index() + " " + e.getMessage());
                }
            }

            ImportResult.Status status;

            if (noteIds.isEmpty()) {
                status = ImportResult.Status.FAILED;
            } else if (!errors.isEmpty()) {
                status = ImportResult.Status.PARTIAL;
            } else {
                status = ImportResult.Status.COMPLETED;
            }

            result = new ImportResult(
                task.correlationId(),
                task.eventId(),
                task.userId(),
                status,
                noteIds,
                links,
                errors,
                Instant.now()
            );

        } catch (DocumentParserException e) {
            result = new ImportResult(
                task.correlationId(),
                task.eventId(),
                task.userId(),
                ImportResult.Status.FAILED,
                List.of(),
                List.of(),
                List.of("Parse error: " + e.getMessage()),
                Instant.now()
            );
        } catch (Exception e) {
            result = new ImportResult(
                task.correlationId(),
                task.eventId(),
                task.userId(),
                ImportResult.Status.FAILED,
                List.of(),
                List.of(),
                List.of("Unexpected error: " + e.getMessage()),
                Instant.now()
            );

        }

        if (result != null) {
            try {
                outboundQueuePort.publish(result);
            } catch (RuntimeException e) {
                stateRepository.releaseClaim(task.eventId());
                throw e;
            }
            stateRepository.markProcessed(task.eventId(), result);
        }

    }


   /* private ParsedDocument parseDocument(ImportTask task) throws DocumentParserException {
        return switch (task.type()) {

            case URL -> documentParser.parseFromUrl(task.content());

            case FILE -> documentParser.parse(getDecoder().decode(task.content()), task.metadata());

            case TEXT -> new ParsedDocument(task.content(), task.metadata());//не нужно парсить, текст уже контент
        };
    }*/
}
