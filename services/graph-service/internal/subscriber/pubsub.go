package subscriber

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
)

type RedisSubscriber struct {
	redisClient *redis.Client
	postgres     db.PostgresClient
	cache        *cache.RedisCache
	channel      string
	limit        int
}

func NewRedisSubscriber(redisClient *redis.Client, postgres db.PostgresClient, cache *cache.RedisCache, channel string, limit int) *RedisSubscriber {
	return &RedisSubscriber{redisClient: redisClient, postgres: postgres, cache: cache, channel: channel, limit: limit}
}

func (s *RedisSubscriber) Start(ctx context.Context) error {
	pubsub := s.redisClient.Subscribe(ctx, s.channel)
	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	ch := pubsub.Channel()
	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				log.Printf("[GraphService] Received invalidation event: %s", msg.Payload)
				if err := s.cache.InvalidateAll(ctx); err != nil {
					log.Printf("[GraphService] Failed to invalidate cache: %v", err)
				}
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			}
		}
	}()

	return nil
}
