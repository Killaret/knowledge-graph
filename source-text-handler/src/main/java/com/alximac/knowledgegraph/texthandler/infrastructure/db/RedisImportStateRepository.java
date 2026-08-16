package com.alximac.knowledgegraph.texthandler.infrastructure.db;

import com.alximac.knowledgegraph.texthandler.domain.model.ImportResult;
import com.alximac.knowledgegraph.texthandler.domain.repository.ImportStateRepository;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.lettuce.core.RedisClient;
import io.lettuce.core.SetArgs;
import io.lettuce.core.api.StatefulRedisConnection;
import io.lettuce.core.api.sync.RedisCommands;

public class RedisImportStateRepository implements ImportStateRepository {
    private final RedisClient redisClient;
    private final String keyPrefix;
    private final long ttlSecond;         // для markProcessed
    private final long claimTtlSeconds;   // для tryClaim
    private final ObjectMapper mapper;


    public RedisImportStateRepository(RedisClient redisClient, String keyPrefix,
                                      long ttlSecond, long claimTtlSeconds, ObjectMapper mapper) {
        this.redisClient = redisClient;
        this.keyPrefix = keyPrefix;
        this.ttlSecond = ttlSecond;
        this.claimTtlSeconds = claimTtlSeconds;
        this.mapper = mapper;
    }

    @Override
    public boolean tryClaim(String eventId) {
        try (StatefulRedisConnection<String, String> connection = redisClient.connect()) {
            RedisCommands<String, String> sync = connection.sync();
            String response = sync.set(keyPrefix + eventId, "processing",
                    SetArgs.Builder.nx().ex(claimTtlSeconds));
            return "OK".equals(response);
        }
    }

    @Override
    public void markProcessed(String eventId, ImportResult result) {
        try (StatefulRedisConnection<String, String> connection = redisClient.connect()) {
            RedisCommands<String, String> sync = connection.sync();
            String jsonResult = mapper.writeValueAsString(result);
            sync.set(keyPrefix + eventId, jsonResult, SetArgs.Builder.ex(ttlSecond));
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
    }


}
