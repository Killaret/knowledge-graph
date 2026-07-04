package com.alximac.knowledgegraph.texthandler.infrastructure.queue;

import com.alximac.knowledgegraph.texthandler.application.port.InboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportTask;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.lettuce.core.RedisClient;
import io.lettuce.core.api.StatefulRedisConnection;
import io.lettuce.core.api.sync.RedisCommands;

import java.util.function.Consumer;

public class AsynqInboundAdapter implements InboundQueuePort {

    private final RedisClient redisClient;
    private final String queueKey;
    private final ObjectMapper objectMapper;

    public AsynqInboundAdapter(RedisClient redisClient, String queueKey, ObjectMapper objectMapper) {
        this.redisClient = redisClient;
        this.queueKey = queueKey;
        this.objectMapper = objectMapper;
    }


    @Override
    public void subscribe(Consumer<ImportTask> handler) {
        try (StatefulRedisConnection<String, String> connection = redisClient.connect()) {
            RedisCommands<String, String> sync = connection.sync();
            while (!Thread.currentThread().isInterrupted()) {
                var result = sync.blpop(5, queueKey);
                if (result == null) continue; // timeout, continue
                try {


                    String envelopeJson = result.getValue(); // read envelope
                    AsynqTaskEnvelope envelope = objectMapper.readValue(envelopeJson, AsynqTaskEnvelope.class); // structure by AsynqTaskEnvelope record class
                    String payloadJson = envelope.payload(); // convert to string

                    ImportTask task = objectMapper.readValue(payloadJson, ImportTask.class); // structure by ImportTask record
                    handler.accept(task);
                } catch (Exception e) {
                    System.err.println("Error while processing message " + e.getMessage());
                }
            }
        } catch (Exception e) {
            throw new RuntimeException("Error in inbound adapter loop", e);
        }
    }

}

