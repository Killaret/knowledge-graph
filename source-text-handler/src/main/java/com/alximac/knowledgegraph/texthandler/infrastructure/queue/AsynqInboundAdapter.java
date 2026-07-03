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

public class AsynqInboundAdapter implements InboundQueuePort {

    private final RedisClient redisClient;
    private final String queueKey;
    private final ObjectMapper objectMapper;
    private final AtomicBoolean running = new AtomicBoolean(true);
    private StatefulRedisConnection<String, String> connection;

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
                    System.err.println("Redis error, retrying: " + e.getMessage());
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

    private void processMessage(String envelopeJson, Consumer<ImportTask> handler) {
        try {
            AsynqTaskEnvelope envelope = objectMapper.readValue(envelopeJson, AsynqTaskEnvelope.class);
            String payloadJson = envelope.payload();
            ImportTask task = objectMapper.readValue(payloadJson, ImportTask.class);
            handler.accept(task);
        } catch (Exception e) {
            System.err.println("Error while processing message: " + e.getMessage());
        }
    }

    public void shutdown() {
        running.set(false);
        if (connection != null) {
            connection.close();
        }
    }
}
