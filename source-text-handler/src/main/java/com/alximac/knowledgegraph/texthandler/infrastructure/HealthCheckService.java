package com.alximac.knowledgegraph.texthandler.infrastructure;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import io.lettuce.core.RedisClient;
import io.lettuce.core.api.StatefulRedisConnection;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;

public class HealthCheckService {
    private final HttpServer httpServer;

    public HealthCheckService(int port, RedisClient redisClient) throws IOException {

        httpServer = HttpServer.create(new InetSocketAddress(port), 0);


        httpServer.createContext("/health/liveliness", exchange -> {
            sendResponse(exchange, 200, "{\"status\":\"UP\"}");
        });

        // проверяет Redis
        httpServer.createContext("/health/readiness", exchange -> {
            boolean redisOk = checkRedis(redisClient);
            int code = redisOk ? 200 : 503;
            String body = redisOk ? "{\"status\":\"UP\"}" : "{\"status\":\"DOWN\",\"reason\":\"Redis unavailable\"}";
            sendResponse(exchange, code, body);
        });

        httpServer.setExecutor(Executors.newSingleThreadExecutor());
    }

    private boolean checkRedis(RedisClient redisClient) {
        try (StatefulRedisConnection<String, String> conn = redisClient.connect()) {
            String pong = conn.sync().ping();
            return "PONG".equals(pong);
        } catch (Exception e) {
            return false;
        }
    }

    private void sendResponse(HttpExchange exchange, int code, String body) throws IOException {
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(code, body.getBytes(StandardCharsets.UTF_8).length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(body.getBytes(StandardCharsets.UTF_8));
        }
    }

    public void start() {
        httpServer.start();
    }

    public void stop() {
        httpServer.stop(0);
    }
}
