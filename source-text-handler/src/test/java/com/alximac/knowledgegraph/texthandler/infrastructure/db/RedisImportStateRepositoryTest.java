package com.alximac.knowledgegraph.texthandler.infrastructure.db;

import com.alximac.knowledgegraph.texthandler.domain.model.ImportResult;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportResult.Status;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import io.lettuce.core.RedisClient;
import io.lettuce.core.api.StatefulRedisConnection;
import org.junit.jupiter.api.*;
/*import org.testcontainers.containers.GenericContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.testcontainers.utility.DockerImageName;*/
import java.time.Instant;
import java.util.List;
import static org.assertj.core.api.Assertions.*;


class RedisImportStateRepositoryTest {

    private static RedisClient redisClient;
    private static ObjectMapper objectMapper;
    private RedisImportStateRepository repository;

    private static final String REDIS_URI = "redis://localhost:6379/0";
    private static final String KEY_PREFIX = "test:event:";

    @BeforeAll
    static void setUp() {
        redisClient = RedisClient.create(REDIS_URI);

        try (var conn = redisClient.connect()) {
            conn.sync().ping();
        } catch (Exception e) {
            throw new IllegalStateException(
                    "Redis is not available at localhost:6379." +
                            " Start Redis manually (e.g. docker run -p 6379:6379 redis:7.4-alpine)", e);
        }

        objectMapper = JsonMapper.builder()
                .addModule(new com.fasterxml.jackson.datatype.jsr310.JavaTimeModule())
                .disable(com.fasterxml.jackson.databind.SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)
                .build();
    }

    @AfterAll
    static void tearDown() {
        if (redisClient != null) {
            redisClient.shutdown();
        }
    }

    @BeforeEach
    void initAndClean() {
        repository = new RedisImportStateRepository(redisClient, KEY_PREFIX, 2, 2, objectMapper);
        try (StatefulRedisConnection<String, String> conn = redisClient.connect()) {
            conn.sync().keys(KEY_PREFIX + "*").forEach(key -> conn.sync().del(key));
        }
    }

    @Test
    @DisplayName("tryClaim must return true for new event")
    void tryClaimMustReturnTrueForNewEvent() {
        assertThat(repository.tryClaim("evt-new")).isTrue();
    }

    @Test
    @DisplayName("tryClaim must return false for already claimed event")
    void tryClaimMustReturnFalseForAlreadyClaimed() {
        repository.tryClaim("evt-dup");
        assertThat(repository.tryClaim("evt-dup")).isFalse();
    }

    @Test
    @DisplayName("markProcessed must store result")
    void markProcessedMustStoreResult() throws JsonProcessingException {
        String eventId = "evt-store";
        repository.tryClaim(eventId);
        ImportResult result = new ImportResult(
                "corr-1", Status.COMPLETED, List.of("n1"), List.of(), List.of(), Instant.now()
        );
        repository.markProcessed(eventId, result);
        try (StatefulRedisConnection<String, String> conn = redisClient.connect()) {
            String json = conn.sync().get(KEY_PREFIX + eventId);
            assertThat(json).contains("\"status\":\"COMPLETED\"");
            ImportResult deserialized = objectMapper.readValue(json, ImportResult.class);
            assertThat(deserialized.correlationId()).isEqualTo("corr-1");
            assertThat(deserialized.status()).isEqualTo(Status.COMPLETED);
            assertThat(deserialized.noteIds()).containsExactly("n1");
        }

    }

    @Test
    @DisplayName("Claim must expire after TTL")
    void claimMustExpireAfterTTL() throws InterruptedException {
        String eventId = "evt-expire";
        assertThat(repository.tryClaim(eventId)).isTrue();
        Thread.sleep(2100);
        assertThat(repository.tryClaim(eventId)).isTrue();
    }

    @Test
    @DisplayName("markProcessed must succeed even if claim key expired")
    void markProcessedWithoutActiveClaim() throws InterruptedException {
        String eventId = "evt-noclaim";
        repository.tryClaim(eventId);
        Thread.sleep(2100);

        ImportResult result = new ImportResult(
                "corr-1",
                ImportResult.Status.COMPLETED,
                List.of("n1"),
                List.of(),
                List.of(),
                Instant.now()
        );

        assertThatCode(() -> repository.markProcessed(eventId, result))
                .doesNotThrowAnyException();

        try (var conn = redisClient.connect()) {
            String json = conn.sync().get(KEY_PREFIX + eventId);
            assertThat(json).contains("COMPLETED");
        }
    }


}




    /*private static final int REDIS_PORT = 6379;

    @Container
    static GenericContainer<?> redis = new GenericContainer<>(DockerImageName.parse("redis:7.4-alpine"))
            .withExposedPorts(REDIS_PORT);

    private static RedisClient redisClient;
    private static ObjectMapper objectMapper;
    private RedisImportStateRepository repository;

    @BeforeAll
    static void setUp() {
        String redisUri = "redis://" + redis.getHost() + ":" + redis.getMappedPort(REDIS_PORT) + "/0";
        redisClient = RedisClient.create(redisUri);
        objectMapper = JsonMapper.builder().build();
    }

    @AfterAll
    static void tearDown() {
        if (redisClient != null) {
            redisClient.shutdown();
        }
    }

    @BeforeEach
    void initRepository() {
        repository = new RedisImportStateRepository(redisClient, "test:event:", 2, 2, objectMapper);
        try (StatefulRedisConnection<String, String> conn = redisClient.connect()) {
            conn.sync().keys("test:event:*").forEach(key -> conn.sync().del(key));
        }
    }

    @Test
    @DisplayName("tryClaim must return true for new event")
    void tryClaimMustReturnTrueForNewEvent() {
        boolean claimed = repository.tryClaim("evt-new");
        assertThat(claimed).isTrue();
    }

    @Test
    @DisplayName("tryClaim must return false for already claimed event")
    void tryClaimMustReturnFalseForAlreadyClaimed() {
        repository.tryClaim("evt-dup"); // первый раз true
        boolean claimedAgain = repository.tryClaim("evt-dup");
        assertThat(claimedAgain).isFalse();
    }

    @Test
    @DisplayName("markProcessed must store result and overwrite claim")
    void markProcessedMustStoreResult() {
        String eventId = "evt-store";
        repository.tryClaim(eventId);

        ImportResult result = new ImportResult(
                "corr-1", Status.COMPLETED, List.of("n1"), List.of(), List.of(), Instant.now()
        );
        repository.markProcessed(eventId, result);

        // Проверка, что ключ теперь содержит сериализованный результат, а не "processing"
        try (StatefulRedisConnection<String, String> conn = redisClient.connect()) {
            String json = conn.sync().get("test:event:" + eventId);
            assertThat(json).isNotNull();
            assertThat(json).contains("\"status\":\"COMPLETED\"");
        }
    }

    @Test
    @DisplayName("Claim must expire after TTL and allow re-claim")
    void claimMustExpireAfterTTL() throws InterruptedException {
        String eventId = "evt-expire";
        assertThat(repository.tryClaim(eventId)).isTrue();

        // Ждём, чтобы ключ истек
        Thread.sleep(2100);

        // Теперь ключ должен быть удалён, и мы захватываем снова
        assertThat(repository.tryClaim(eventId)).isTrue();
    }*/
