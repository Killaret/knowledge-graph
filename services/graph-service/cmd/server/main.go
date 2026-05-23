package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"knowledge-graph-graph-service/internal/api"
	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/subscriber"
)

func envOrDefault(key, def string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return val
}

func mustIntEnv(key string, def int) int {
	raw := envOrDefault(key, "")
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return val
}

func main() {
	grpcPort := envOrDefault("GRPC_PORT", "9090")
	httpPort := envOrDefault("HTTP_PORT", "9091")
	postgresURL := envOrDefault("POSTGRES_URL", "postgresql://postgres:postgres@postgres:5432/knowledge_base?sslmode=disable")
	redisURL := envOrDefault("REDIS_URL", "redis:6379")
	redisChannel := envOrDefault("EVENT_CHANNEL", "graph:events")
	fullLimit := mustIntEnv("GRAPH_FULL_LIMIT", 1000)

	log.Printf("[GraphService] Starting with POSTGRES=%s REDIS=%s GRPC=%s HTTP=%s EVENT_CHANNEL=%s", postgresURL, redisURL, grpcPort, httpPort, redisChannel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pgPool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: redisURL})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
		}
	}()

	pgClient := db.NewPostgresClient(pgPool)
	cacheClient := cache.NewRedisCache(redisClient)
	service := api.NewGraphServer(pgClient, cacheClient, fullLimit)

	// Start gRPC server
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to start gRPC listener: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterGraphServiceServer(grpcServer, service)

	// Start HTTP fallback server
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	api.RegisterHTTPHandlers(httpMux, service)

	subscriberSvc := subscriber.NewRedisSubscriber(redisClient, pgClient, cacheClient, redisChannel, fullLimit)
	if err := subscriberSvc.Start(ctx); err != nil {
		log.Fatalf("failed to start subscriber: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Printf("[GraphService] gRPC listening on %s", grpcPort)
		if err := grpcServer.Serve(grpcLis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Printf("[GraphService] HTTP fallback listening on %s", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), httpMux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[GraphService] Shutdown signal received")
	grpcServer.GracefulStop()
	wg.Wait()
}
