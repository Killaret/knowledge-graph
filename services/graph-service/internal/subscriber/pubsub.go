package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/config"
	"knowledge-graph-graph-service/internal/db"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Event represents the event structure from the monolith
type Event struct {
	EventID string          `json:"event_id"`
	Event   string          `json:"event"` // NoteCreated, NoteUpdated, NoteDeleted, LinkCreated, LinkUpdated, LinkDeleted
	Payload json.RawMessage `json:"payload"`
}

// NoteEventPayload represents the payload for note events
type NoteEventPayload struct {
	NoteID string `json:"note_id"`
	UserID string `json:"user_id"`
}

// LinkEventPayload represents the payload for link events
type LinkEventPayload struct {
	SourceNoteID string `json:"source_note_id"`
	TargetNoteID string `json:"target_note_id"`
	UserID       string `json:"user_id"`
}

// EventTracking stores event acknowledgment data in Redis
type EventTracking struct {
	TimestampReceived  time.Time `json:"timestamp_received"`
	TimestampProcessed time.Time `json:"timestamp_processed"`
	IsAcknowledged     bool      `json:"is_acknowledged"`
	ErrorMessage       string    `json:"error_message,omitempty"`
}

type RedisSubscriber struct {
	redisClient      *redis.Client
	postgres         db.PostgresClient
	cache            *cache.RedisCache
	channel          string
	limit            int
	eventTrackingTTL time.Duration
	checkInterval    time.Duration
	retryThreshold   time.Duration
	// refreshRequests carries closure-view refresh requests; capacity 1 so a
	// burst of link/note events coalesces into a single pending refresh.
	refreshRequests chan struct{}
	// closureDebounce is the quiet window waited before refreshing, so a
	// batch of events triggers one REFRESH instead of one per event.
	closureDebounce time.Duration
}

func NewRedisSubscriber(redisClient *redis.Client, postgres db.PostgresClient, cache *cache.RedisCache, channel string, limit int) *RedisSubscriber {
	return &RedisSubscriber{
		redisClient:      redisClient,
		postgres:         postgres,
		cache:            cache,
		channel:          channel,
		limit:            limit,
		eventTrackingTTL: 24 * time.Hour,
		checkInterval:    5 * time.Minute,
		retryThreshold:   5 * time.Minute,
		refreshRequests:  make(chan struct{}, 1),
		closureDebounce:  2 * time.Second,
	}
}

// NewRedisSubscriberWithConfig creates a subscriber with config-driven TTLs and intervals.
func NewRedisSubscriberWithConfig(redisClient *redis.Client, postgres db.PostgresClient, cache *cache.RedisCache, cfg *config.Config) *RedisSubscriber {
	return &RedisSubscriber{
		redisClient:      redisClient,
		postgres:         postgres,
		cache:            cache,
		channel:          cfg.EventChannel,
		limit:            cfg.FullLimit,
		eventTrackingTTL: cfg.EventTrackingTTL,
		checkInterval:    cfg.UnprocessedEventCheckInterval,
		retryThreshold:   cfg.UnprocessedEventRetryThreshold,
		refreshRequests:  make(chan struct{}, 1),
		closureDebounce:  2 * time.Second,
	}
}

func (s *RedisSubscriber) Start(ctx context.Context) error {
	pubsub := s.redisClient.Subscribe(ctx, s.channel)
	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to channel: %w", err)
	}

	ch := pubsub.Channel()

	// Start worker for processing unprocessed events periodically
	go s.processUnprocessedEvents(ctx)

	// Start the debounced closure-view refresher: link/note events request a
	// refresh instead of running REFRESH synchronously per event.
	go s.closureRefreshLoop(ctx)

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				s.handleEvent(ctx, msg.Payload)
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			}
		}
	}()

	return nil
}

func (s *RedisSubscriber) handleEvent(ctx context.Context, payload string) {
	var event Event
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("[GraphService] Failed to parse event payload: %v", err)
		return
	}

	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}

	log.Printf("[GraphService] Received event: %s (ID: %s)", event.Event, event.EventID)

	// Record event receipt in Redis
	eventKey := fmt.Sprintf("event:%s", event.EventID)
	tracking := EventTracking{
		TimestampReceived: time.Now(),
		IsAcknowledged:    false,
	}

	trackingData, _ := json.Marshal(tracking)
	if err := s.redisClient.HSet(ctx, eventKey, "data", trackingData).Err(); err != nil {
		log.Printf("[GraphService] Failed to record event receipt: %v", err)
	}

	// Set TTL for event tracking
	s.redisClient.Expire(ctx, eventKey, s.eventTrackingTTL)

	// Process the event
	if err := s.processEvent(ctx, event); err != nil {
		log.Printf("[GraphService] Failed to process event %s: %v", event.EventID, err)
		tracking.ErrorMessage = err.Error()
		trackingData, _ = json.Marshal(tracking)
		s.redisClient.HSet(ctx, eventKey, "data", trackingData)
		return
	}

	// Mark event as acknowledged
	tracking.IsAcknowledged = true
	tracking.TimestampProcessed = time.Now()
	trackingData, _ = json.Marshal(tracking)
	if err := s.redisClient.HSet(ctx, eventKey, "data", trackingData).Err(); err != nil {
		log.Printf("[GraphService] Failed to mark event as acknowledged: %v", err)
	}

	log.Printf("[GraphService] Event %s acknowledged successfully", event.EventID)
}

func (s *RedisSubscriber) processEvent(ctx context.Context, event Event) error {
	switch event.Event {
	case "NoteCreated", "NoteUpdated", "NoteDeleted":
		return s.handleNoteEvent(ctx, event)
	case "LinkCreated", "LinkUpdated", "LinkDeleted":
		return s.handleLinkEvent(ctx, event)
	default:
		log.Printf("[GraphService] Unknown event type: %s", event.Event)
		return nil
	}
}

func (s *RedisSubscriber) handleNoteEvent(ctx context.Context, event Event) error {
	var payload NoteEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse note event payload: %w", err)
	}

	userID := payload.UserID
	if userID == "" {
		userID = "public"
	}

	log.Printf("[GraphService] Handling note event %s for user %s, note %s", event.Event, userID, payload.NoteID)

	// Invalidate note-specific cache for all users/depths
	noteCacheKey := config.CacheKey("note", "*", payload.NoteID, "*")
	s.invalidatePattern(ctx, noteCacheKey)

	// Invalidate user's full layout cache
	fullLayoutKey := config.CacheKey("full", userID)
	s.redisClient.Del(ctx, fullLayoutKey)

	// Invalidate user's delta cache
	deltaPattern := config.CacheKey("delta", userID, "*")
	s.invalidatePattern(ctx, deltaPattern)

	// Publish/unpublish affects the public graph, so invalidate public caches too.
	publicFullLayoutKey := config.CacheKey("full", "public")
	s.redisClient.Del(ctx, publicFullLayoutKey)

	publicDeltaPattern := config.CacheKey("delta", "public", "*")
	s.invalidatePattern(ctx, publicDeltaPattern)

	// Note deletion may cascade to links, so refresh the closure view.
	if event.Event == "NoteDeleted" {
		s.requestClosureRefresh()
	}

	log.Printf("[GraphService] Cache invalidated for note event %s", event.Event)
	return nil
}

func (s *RedisSubscriber) handleLinkEvent(ctx context.Context, event Event) error {
	var payload LinkEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse link event payload: %w", err)
	}

	userID := payload.UserID
	if userID == "" {
		userID = "public"
	}

	log.Printf("[GraphService] Handling link event %s for user %s", event.Event, userID)

	// Invalidate user's full layout cache
	fullLayoutKey := config.CacheKey("full", userID)
	s.redisClient.Del(ctx, fullLayoutKey)

	// Invalidate user's delta cache
	deltaPattern := config.CacheKey("delta", userID, "*")
	s.invalidatePattern(ctx, deltaPattern)

	// Invalidate note-specific caches for both source and target
	if payload.SourceNoteID != "" {
		noteCachePattern1 := config.CacheKey("note", "*", payload.SourceNoteID, "*")
		s.invalidatePattern(ctx, noteCachePattern1)
	}
	if payload.TargetNoteID != "" {
		noteCachePattern2 := config.CacheKey("note", "*", payload.TargetNoteID, "*")
		s.invalidatePattern(ctx, noteCachePattern2)
	}

	// Links change the transitive closure, so refresh the materialized view —
	// asynchronously: REFRESH on a dense graph takes seconds to minutes and
	// must not block the subscriber on every LinkCreated.
	s.requestClosureRefresh()

	log.Printf("[GraphService] Cache invalidated for link event %s", event.Event)
	return nil
}

// requestClosureRefresh schedules a closure-view refresh without blocking the
// caller. If a request is already pending the extra signal is dropped — one
// pending refresh covers all changes seen so far.
func (s *RedisSubscriber) requestClosureRefresh() {
	select {
	case s.refreshRequests <- struct{}{}:
	default:
	}
}

// closureRefreshLoop waits for refresh requests, debounces bursts into a
// single REFRESH, and runs it off the event-handling path.
func (s *RedisSubscriber) closureRefreshLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.refreshRequests:
			timer := time.NewTimer(s.closureDebounce)
			for quiet := false; !quiet; {
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-s.refreshRequests:
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(s.closureDebounce)
				case <-timer.C:
					quiet = true
				}
			}
			if err := s.postgres.RefreshClosureView(ctx); err != nil {
				log.Printf("[GraphService] Failed to refresh closure view: %v", err)
			}
		}
	}
}

func (s *RedisSubscriber) invalidatePattern(ctx context.Context, pattern string) {
	iter := s.redisClient.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := s.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("[GraphService] Failed to delete key %s: %v", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		log.Printf("[GraphService] Error scanning pattern %s: %v", pattern, err)
	}
}

// processUnprocessedEvents periodically checks for events that were not acknowledged
func (s *RedisSubscriber) processUnprocessedEvents(ctx context.Context) {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.scanAndRetryEvents(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *RedisSubscriber) scanAndRetryEvents(ctx context.Context) {
	log.Printf("[GraphService] Scanning for unprocessed events...")

	pattern := "event:*"
	iter := s.redisClient.Scan(ctx, 0, pattern, 100).Iterator()

	retryCount := 0
	for iter.Next(ctx) {
		eventKey := iter.Val()

		// Get event tracking data
		data, err := s.redisClient.HGetAll(ctx, eventKey).Result()
		if err != nil {
			log.Printf("[GraphService] Failed to get event tracking for %s: %v", eventKey, err)
			continue
		}

		if len(data) == 0 {
			continue
		}

		var tracking EventTracking
		trackingData := []byte(data["data"])

		if len(trackingData) > 0 {
			if err := json.Unmarshal(trackingData, &tracking); err == nil {
				// Check if event is unacknowledged and older than retry threshold
				if !tracking.IsAcknowledged && time.Since(tracking.TimestampReceived) > s.retryThreshold {
					log.Printf("[GraphService] Retrying unacknowledged event: %s", eventKey)
					// In a real implementation, we would need to replay the event
					// For now, we just log it
					retryCount++
				}
			}
		}
	}

	if retryCount > 0 {
		log.Printf("[GraphService] Found %d unacknowledged events", retryCount)
	}

	if err := iter.Err(); err != nil {
		log.Printf("[GraphService] Error scanning for unprocessed events: %v", err)
	}
}
