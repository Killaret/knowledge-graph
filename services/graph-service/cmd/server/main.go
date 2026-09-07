package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"knowledge-graph-graph-service/internal/api"
	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/config"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/subscriber"
	graphservice "knowledge-graph-graph-service/proto"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("[GraphService] Starting with POSTGRES=%s REDIS=%s GRPC=%s HTTP=%s EVENT_CHANNEL=%s",
		cfg.PostgresURL, cfg.RedisURL, cfg.GRPCPort, cfg.HTTPPort, cfg.EventChannel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Configure pgxpool with connection pool settings for graph-service
	pgConfig, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to parse postgres config: %v", err)
	}

	// Connection pool settings for graph-service (lighter than main backend)
	pgConfig.MaxConns = 10                       // Maximum number of connections
	pgConfig.MinConns = 2                        // Minimum number of connections
	pgConfig.MaxConnLifetime = 5 * time.Minute   // Maximum lifetime of a connection
	pgConfig.MaxConnIdleTime = 1 * time.Minute   // Maximum idle time of a connection
	pgConfig.HealthCheckPeriod = 1 * time.Minute // Health check interval

	pgPool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()

	log.Printf("[GraphService] PostgreSQL pool configured: MaxConns=10, MinConns=2, MaxLifetime=5m, MaxIdleTime=1m")

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		PoolSize:     10,              // Maximum number of connections
		MinIdleConns: 3,               // Minimum number of idle connections
		PoolTimeout:  4 * time.Second, // Timeout for getting connection from pool
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
		}
	}()

	pgClient := db.NewPostgresClient(pgPool, cfg.NLPModelName)
	cacheClient := cache.NewRedisCacheWithConfig(redisClient, cfg)
	service := api.NewGraphService(pgClient, cacheClient, cfg.FullLimit, cfg.DefaultDepth, cfg.StreamChunkSize)

	// Start gRPC server
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to start gRPC listener: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(api.GRPCAuthUnaryInterceptor(cfg)),
		grpc.StreamInterceptor(api.GRPCAuthStreamInterceptor(cfg)),
	)
	graphservice.RegisterGraphServiceServer(grpcServer, service)

	// Start HTTP fallback server
	httpMux := http.NewServeMux()
	api.RegisterHTTPHandlers(httpMux, service, cfg)

	subscriberSvc := subscriber.NewRedisSubscriberWithConfig(redisClient, pgClient, cacheClient, cfg)
	if err := subscriberSvc.Start(ctx); err != nil {
		log.Fatalf("failed to start subscriber: %v", err)
	}

	// Start periodic pool statistics logging
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Log PostgreSQL pool statistics
				pgStats := pgPool.Stat()
				log.Printf("[GraphService] PostgreSQL Pool: TotalConns=%d, IdleConns=%d, MaxConns=%d, AcquireCount=%d, AcquireDuration=%v",
					pgStats.TotalConns(),
					pgStats.IdleConns(),
					pgStats.MaxConns(),
					pgStats.AcquireCount(),
					pgStats.AcquireDuration())

				// Log Redis pool statistics
				redisStats := redisClient.PoolStats()
				log.Printf("[GraphService] Redis Pool: Hits=%d, Misses=%d, Timeouts=%d, TotalConns=%d, IdleConns=%d, StaleConns=%d",
					redisStats.Hits,
					redisStats.Misses,
					redisStats.Timeouts,
					redisStats.TotalConns,
					redisStats.IdleConns,
					redisStats.StaleConns)
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Printf("[GraphService] gRPC listening on %s", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Printf("[GraphService] HTTP fallback listening on %s", cfg.HTTPPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), httpMux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[GraphService] Shutdown signal received")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop gRPC server gracefully
	log.Println("[GraphService] Stopping gRPC server...")
	grpcServer.GracefulStop()

	// Stop HTTP server gracefully
	log.Println("[GraphService] Stopping HTTP server...")
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: httpMux,
	}
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[GraphService] Error stopping HTTP server: %v", err)
	}

	// Wait for all servers to stop
	wg.Wait()

	// Close connections
	log.Println("[GraphService] Closing database connection...")
	pgPool.Close()

	log.Println("[GraphService] Closing redis connection...")
	if err := redisClient.Close(); err != nil {
		log.Printf("[GraphService] Error closing redis: %v", err)
	}

	log.Println("[GraphService] Graceful shutdown completed")
}
