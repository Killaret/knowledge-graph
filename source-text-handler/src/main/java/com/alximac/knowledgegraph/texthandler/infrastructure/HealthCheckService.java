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

        // Регистрируем контексты
        httpServer.createContext("/health/liveness", exchange -> {
            try {
                sendResponse(exchange, 200, "{\"status\":\"UP\"}");
            } catch (IOException e) {
                e.printStackTrace();
            }
        });

        httpServer.createContext("/health/readiness", exchange -> {
            try {
                boolean redisOk = checkRedis(redisClient);
                int code = redisOk ? 200 : 503;
                String body = redisOk ? "{\"status\":\"UP\"}" : "{\"status\":\"DOWN\",\"reason\":\"Redis unavailable\"}";
                sendResponse(exchange, code, body);
            } catch (IOException e) {
                e.printStackTrace();
            }
        });

        httpServer.setExecutor(Executors.newSingleThreadExecutor());
        System.out.println("HealthCheckService created on port " + port);
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
        byte[] bytes = body.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(code, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    public void start() {
        httpServer.start();
        System.out.println("HealthCheckService started and listening");
    }

    public void stop() {
        httpServer.stop(0);
    }
}