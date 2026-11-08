package com.alximac.knowledgegraph.texthandler.infrastructure.queue;

import com.alximac.knowledgegraph.texthandler.application.port.OutboundQueuePort;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportResult;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.lettuce.core.RedisClient;
import io.lettuce.core.api.StatefulRedisConnection;
import io.lettuce.core.api.sync.RedisCommands;


public class AsynqOutboundAdapter implements OutboundQueuePort {
    private final RedisClient redisClient;
    private final String queueKey;
    private final ObjectMapper objectMapper;

    public AsynqOutboundAdapter(RedisClient redisClient, String queueKey, ObjectMapper objectMapper) {
        this.redisClient = redisClient;
        this.queueKey = queueKey;
        this.objectMapper = objectMapper;
    }

    @Override
    public void publish(ImportResult result) {

        try (StatefulRedisConnection<String, String> connection = redisClient.connect()) {
            RedisCommands<String, String> sync = connection.sync();

            String json = objectMapper.writeValueAsString(result);//сериализуем result в json строку
            sync.rpush(queueKey, json);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("ImportResult serialization error ", e);
        }


    }
}
