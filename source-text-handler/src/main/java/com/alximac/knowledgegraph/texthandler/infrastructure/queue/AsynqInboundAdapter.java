package com.alximac.knowledgegraph.texthandler.infrastructure.queue;

import com.alximac.knowledgegraph.texthandler.application.port.InboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportTask;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.lettuce.core.RedisClient;
import io.lettuce.core.RedisException;
import io.lettuce.core.api.StatefulRedisConnection;
import io.lettuce.core.api.sync.RedisCommands;

import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.Consumer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class AsynqInboundAdapter implements InboundQueuePort {

    private final RedisClient redisClient;
    private final String queueKey;
    private final ObjectMapper objectMapper;
    private final AtomicBoolean running = new AtomicBoolean(true);
    private StatefulRedisConnection<String, String> connection;
    private static final Logger log = LoggerFactory.getLogger(AsynqInboundAdapter.class);

    public AsynqInboundAdapter(RedisClient redisClient, String queueKey, ObjectMapper objectMapper) {
        this.redisClient = redisClient;
        this.queueKey = queueKey;
        this.objectMapper = objectMapper;
    }

    @Override
    public void subscribe(Consumer<ImportTask> handler) {
        connection = redisClient.connect();
        try {
            RedisCommands<String, String> sync = connection.sync();
            while (running.get() && !Thread.currentThread().isInterrupted()) {
                try {
                    var result = sync.blpop(5, queueKey);
                    if (result == null) {
                        continue;
                    }
                    processMessage(result.getValue(), handler);
                } catch (RedisException e) {
                    if (!running.get() || Thread.currentThread().isInterrupted()) {
                        break;
                    }
                    log.error("Redis error, retrying", e);
                    try {
                        Thread.sleep(1000);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            }
        } finally {
            if (connection != null) {
                connection.close();
            }
        }
    }

    private void processMessage(String json, Consumer<ImportTask> handler) {
        try {
            ImportTask task = objectMapper.readValue(json, ImportTask.class);
            handler.accept(task);
        } catch (Exception e) {
            //момент с утечкой jwt - jackson кладет первые 500 символов
            log.warn("Failed to parse incoming JSON task: {}. Skipped.", e.getClass().getSimpleName());
        }
    }

    public void shutdown() {
        running.set(false);
        if (connection != null) {
            connection.close();
        }
    }
}
