package com.alximac.knowledgegraph.texthandler.config;

import com.alximac.knowledgegraph.texthandler.application.ImportDocumentHandler;
import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.application.port.OutboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.service.ChunkingStrategy;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.LinkDetector;
import com.alximac.knowledgegraph.texthandler.infrastructure.HealthCheckService;
import com.alximac.knowledgegraph.texthandler.infrastructure.chunking.HybridChunker;
import com.alximac.knowledgegraph.texthandler.infrastructure.db.RedisImportStateRepository;
import com.alximac.knowledgegraph.texthandler.infrastructure.http.NoteCreatorHttpClient;
import com.alximac.knowledgegraph.texthandler.infrastructure.http.ResilientNoteCreator;
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
import io.github.resilience4j.circuitbreaker.CircuitBreaker;
import io.github.resilience4j.circuitbreaker.CircuitBreakerConfig;
import io.github.resilience4j.retry.Retry;
import io.github.resilience4j.retry.RetryConfig;
import io.lettuce.core.RedisClient;

import java.io.IOException;
import java.net.http.HttpClient;
import java.time.Duration;
import java.util.List;

public class AppConfig {
    private static final long TIMEOUT_DURATION = 5;

    private ObjectMapper buildObjectMapper() {
        return JsonMapper.builder()
                .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
                .propertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE)
                .addModule(new ParameterNamesModule())
                .addModule(new JavaTimeModule())   // <-- добавляем поддержку java.time
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS) // <-- ISO-8601 вместо чисел
                .build();
    }

    private RedisClient createRedisClient() {
        String uri = System.getenv().getOrDefault("REDIS_URI", "redis://localhost:6379/0");//change redis://localhost:6379/0 later
        return RedisClient.create(uri);
    }


    public void start() throws IOException {


        RedisClient redisClient = createRedisClient();
        ObjectMapper objectMapper = buildObjectMapper();
        HttpClient httpClient = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(TIMEOUT_DURATION)).build();
        NoteCreatorHttpClient noteCreatorHttpClient =
                new NoteCreatorHttpClient(httpClient, objectMapper, "http://backend:8080");
        HealthCheckService healthService = new HealthCheckService(8081, redisClient);
        healthService.start();

        RetryConfig retryConfig = RetryConfig.custom()
                .maxAttempts(3)
                .waitDuration(Duration.ofSeconds(2))
                .retryExceptions(RemoteServiceException.class)
                .build();
        Retry retry = Retry.of("noteRetry", retryConfig);


        CircuitBreakerConfig breakerConfig = CircuitBreakerConfig.custom()
                .failureRateThreshold(50)
                .waitDurationInOpenState(Duration.ofSeconds(30))
                .slidingWindowSize(10)
                .build();
        CircuitBreaker circuitBreaker = CircuitBreaker.of("noteCB", breakerConfig);

        NoteCreatorPort noteCreator = new ResilientNoteCreator(noteCreatorHttpClient, retry, circuitBreaker);


        long ttlSeconds = 604800;
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


        ChunkingStrategy chunk = new HybridChunker();

        LinkDetector linkDetector = chunks -> List.of();


        OutboundQueuePort outboundQueuePort = new AsynqOutboundAdapter(
                redisClient,
                "asynq:import:responses:pending",
                objectMapper);

        ImportDocumentHandler handler = new ImportDocumentHandler(chunk, parserFactory, linkDetector, noteCreator,
                outboundQueuePort, repository);


        Thread worker = new Thread(() -> asynqInboundAdapter.subscribe(handler::handle));
        worker.setName("async-worker");
        worker.setDaemon(false);
        worker.start();

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            System.out.println("Graceful shutdown initiated...");
            asynqInboundAdapter.shutdown();   //  флаг + закрытие сокета

            // 30 секунд на завершение текущей задачи и выход из BLPOP
            try {
                worker.join(30_000);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }

            // Если поток всё ещё жив — что-то зависло, принудительно прерываем
            if (worker.isAlive()) {
                System.err.println("Worker stuck, interrupting...");
                worker.interrupt();
                try {
                    worker.join(5_000);
                } catch (InterruptedException ex) {
                    Thread.currentThread().interrupt();
                }
            }

            if (healthService != null) {
                healthService.stop();
            }
            System.out.println("Shutdown complete.");
        }));
        // Главный поток ждёт завершения воркера, а не висит бесконечно
        try {
            worker.join();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }

    }
}

