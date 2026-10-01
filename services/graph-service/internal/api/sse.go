package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"knowledge-graph-graph-service/internal/config"

	"github.com/redis/go-redis/v9"
)

// GraphChangeEvent is the payload an SSE client receives — a bare "the graph
// changed" signal. The client then asks /graph/delta for what changed, so the
// stream carries no graph data and cannot leak note content.
type GraphChangeEvent struct {
	EventID string `json:"event_id,omitempty"`
	Event   string `json:"event"`
}

// eventsBroker subscribes to the shared graph events Redis channel and fans
// each event out to the SSE connections owned by that event's user. Scope "*"
// (SKIP_AUTH dev/test mode) receives every event.
type eventsBroker struct {
	redis   *redis.Client
	channel string

	mu      sync.Mutex
	clients map[chan GraphChangeEvent]string // send channel -> scope
}

// NewEventsBroker creates the fan-out broker for the SSE endpoint.
func NewEventsBroker(redisClient *redis.Client, channel string) *eventsBroker {
	return &eventsBroker{
		redis:   redisClient,
		channel: channel,
		clients: make(map[chan GraphChangeEvent]string),
	}
}

// redisEvent mirrors the envelope published by backend/internal/infrastructure/events.
type redisEvent struct {
	EventID string          `json:"event_id"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

// Start runs the Redis subscription until ctx is cancelled. The subscription
// is independent from the cache-invalidation subscriber so either can restart
// without disturbing the other.
func (b *eventsBroker) Start(ctx context.Context) {
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			b.run(ctx)
			// Brief pause before resubscribing after a Redis drop.
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

func (b *eventsBroker) run(ctx context.Context) {
	pubsub := b.redis.Subscribe(ctx, b.channel)
	defer func() {
		if err := pubsub.Close(); err != nil {
			log.Printf("[SSE] failed to close pubsub: %v", err)
		}
	}()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			b.dispatch(msg.Payload)
		}
	}
}

func (b *eventsBroker) dispatch(raw string) {
	var env redisEvent
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		log.Printf("[SSE] dropping undecodable event: %v", err)
		return
	}
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		log.Printf("[SSE] dropping event with undecodable payload: %v", err)
		return
	}

	out := GraphChangeEvent{EventID: env.EventID, Event: env.Event}

	b.mu.Lock()
	defer b.mu.Unlock()
	for ch, scope := range b.clients {
		if scope != "*" && scope != payload.UserID {
			continue
		}
		select {
		case ch <- out:
		default:
			// Slow consumer: the client refetches via the delta endpoint
			// anyway, so dropping is safe and keeps the broker unblocked.
		}
	}
}

func (b *eventsBroker) subscribe(scope string) (chan GraphChangeEvent, func()) {
	ch := make(chan GraphChangeEvent, 16)
	b.mu.Lock()
	b.clients[ch] = scope
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
	}
}

// subscriberCount reports how many SSE connections are registered — used by
// tests to wait until the handler is listening.
func (b *eventsBroker) subscriberCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients)
}

// SSEHandler serves GET /api/v1/graph/events. Registered without AuthMiddleware
// because EventSource cannot send the Authorization header — the JWT arrives as
// the `access_token` query parameter instead, validated with the same secret.
func (b *eventsBroker) SSEHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := "*"
		if !cfg.SkipAuth {
			userID, ok := authenticateRequest(r, cfg)
			if !ok || userID == "" {
				if token := r.URL.Query().Get("access_token"); token != "" {
					userID, ok = validateJWT(token, cfg.JWTSecret)
				}
			}
			if !ok || userID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			scope = userID
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, `{"error":"streaming unsupported"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// Ask nginx not to buffer the stream.
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		events, unsubscribe := b.subscribe(scope)
		defer unsubscribe()

		heartbeat := time.NewTicker(25 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case ev := <-events:
				payload, err := json.Marshal(ev)
				if err != nil {
					continue
				}
				if ev.EventID != "" {
					fmt.Fprintf(w, "id: %s\n", ev.EventID)
				}
				fmt.Fprintf(w, "event: graph\ndata: %s\n\n", payload)
				flusher.Flush()
			case <-heartbeat.C:
				if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}
