package com.alximac.knowledgegraph.texthandler.config;

import com.alximac.knowledgegraph.texthandler.application.ImportDocumentHandler;
import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.application.port.OutboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.model.*;
import com.alximac.knowledgegraph.texthandler.domain.service.ChunkingStrategy;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;
import com.alximac.knowledgegraph.texthandler.domain.service.LinkDetector;
import com.alximac.knowledgegraph.texthandler.infrastructure.db.RedisImportStateRepository;
import com.alximac.knowledgegraph.texthandler.infrastructure.parser.ParserFactory;
import com.alximac.knowledgegraph.texthandler.infrastructure.parser.TextDocumentParser;
import com.alximac.knowledgegraph.texthandler.infrastructure.parser.TikaDocumentParser;
import com.alximac.knowledgegraph.texthandler.infrastructure.queue.AsynqInboundAdapter;
import com.alximac.knowledgegraph.texthandler.infrastructure.queue.AsynqOutboundAdapter;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.fasterxml.jackson.module.paramnames.ParameterNamesModule;
import io.lettuce.core.RedisClient;

import java.util.List;
import java.util.Map;
import java.util.UUID;

public class AppConfig {
    private final long ttlSeconds = 604800;

    private ObjectMapper buildObjectMapper () {
        return JsonMapper.builder()
                .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
                .propertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE)
                .addModule(new ParameterNamesModule())
                .addModule(new JavaTimeModule())   // <-- добавляем поддержку java.time
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS) // <-- ISO-8601 вместо чисел
                .build();
    }

    private RedisClient createRedisClient () {
        String uri = System.getenv().getOrDefault("REDIS_URI", "redis://localhost:6379/0");//change redis://localhost:6379/0 later
        return RedisClient.create(uri);
    }

    public void start() {



        RedisClient redisClient = createRedisClient();
        ObjectMapper objectMapper = buildObjectMapper();

        RedisImportStateRepository repository = new RedisImportStateRepository(
                redisClient,
                "import:event:",
                ttlSeconds, objectMapper);

        AsynqInboundAdapter asynqInboundAdapter = new AsynqInboundAdapter(
                redisClient,
                "asynq:import:document:pending",
                objectMapper);

        DocumentParser tikaParser = new TikaDocumentParser();
        DocumentParser textParser = new TextDocumentParser();
        ParserFactory parserFactory = new ParserFactory(tikaParser, textParser);



        ChunkingStrategy chunk = new ChunkingStrategy() {
            @Override
            public List<DocumentChunk> chunk(String text, ImportOptions options) {
                if (text == null || text.isBlank()) return List.of();
                return List.of(new DocumentChunk(text, 0, Map.of(), null));
            }
        };

        LinkDetector linkDetector = chunks -> List.of();

        NoteCreatorPort noteCreatorPort = new NoteCreatorPort() {
            @Override
            public String createNote(DocumentChunk chunk) throws RemoteServiceException {
                return UUID.randomUUID().toString();
            }

            @Override
            public void createLink(Link link) throws RemoteServiceException {

            }
        };

        OutboundQueuePort outboundQueuePort = new AsynqOutboundAdapter(
                redisClient,
                "asynq:import:responses:pending",
                objectMapper);

        ImportDocumentHandler handler = new ImportDocumentHandler(chunk, parserFactory, linkDetector, noteCreatorPort,
                outboundQueuePort, repository);

        Thread worker = new Thread(() -> asynqInboundAdapter.subscribe(handler::handle));
        worker.setDaemon(false);
        worker.start();

        try {
            Thread.currentThread().join();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }

    }
}

