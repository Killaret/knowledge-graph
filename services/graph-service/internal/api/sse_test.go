package api

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"knowledge-graph-graph-service/internal/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBroker(t *testing.T) (*eventsBroker, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	return NewEventsBroker(redisClient, "graph:events"), mr
}

func newTestBrokerServer(t *testing.T, cfg *config.Config) (*httptest.Server, *miniredis.Miniredis, *eventsBroker) {
	t.Helper()
	broker, mr := newTestBroker(t)
	ctx, cancel := context.WithCancel(context.Background())
	broker.Start(ctx)
	t.Cleanup(cancel)
	srv := httptest.NewServer(broker.SSEHandler(cfg))
	t.Cleanup(srv.Close)
	return srv, mr, broker
}

// waitForClient blocks until the handler registered an SSE connection.
func waitForClient(t *testing.T, b *eventsBroker) {
	t.Helper()
	require.Eventually(t, func() bool {
		return b.subscriberCount() > 0
	}, 3*time.Second, 10*time.Millisecond, "SSE client did not register")
}

func sseConfig() *config.Config {
	return &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
}

// sseReader drains an SSE body once and queues every `data:` line — a plain
// bufio.Reader polled from several goroutines loses lines to a leftover
// reader that has already timed out.
type sseReader struct {
	lines chan string
}

func newSSEReader(body *bufio.Reader) *sseReader {
	r := &sseReader{lines: make(chan string, 16)}
	go func() {
		for {
			line, err := body.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data:") {
				r.lines <- line
			}
		}
	}()
	return r
}

// next waits for the next `data:` line or times out.
func (r *sseReader) next(t *testing.T, timeout time.Duration) (string, bool) {
	t.Helper()
	select {
	case line := <-r.lines:
		return line, true
	case <-time.After(timeout):
		return "", false
	}
}

func publishEvent(t *testing.T, mr *miniredis.Miniredis, userID string) {
	t.Helper()
	// The broker subscribes in a goroutine — wait until Redis sees it, or the
	// published message lands before anyone is listening.
	require.Eventually(t, func() bool {
		return mr.PubSubNumSub("graph:events")["graph:events"] > 0
	}, 3*time.Second, 10*time.Millisecond, "broker did not subscribe to graph:events")
	payload := fmt.Sprintf(
		`{"event_id":"ev-1","event":"NoteCreated","payload":{"note_id":"n1","user_id":"%s"}}`,
		userID,
	)
	mr.Publish("graph:events", payload)
}

func TestSSEHandlerRejectsUnauthenticated(t *testing.T) {
	broker, _ := newTestBroker(t)
	handler := broker.SSEHandler(sseConfig())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/events", nil)
	handler(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSSEHandlerDeliversEventToOwnerOnly(t *testing.T) {
	srv, mr, broker := newTestBrokerServer(t, sseConfig())

	token := generateTestToken("user-1", "test-secret", "access")
	resp, err := http.Get(srv.URL + "?access_token=" + token)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "no", resp.Header.Get("X-Accel-Buffering"))

	reader := newSSEReader(bufio.NewReader(resp.Body))
	waitForClient(t, broker)

	// An event for another user must never reach this connection.
	publishEvent(t, mr, "user-2")
	_, got := reader.next(t, 300*time.Millisecond)
	assert.False(t, got, "event for user-2 leaked to user-1 connection")

	publishEvent(t, mr, "user-1")
	line, got := reader.next(t, 3*time.Second)
	require.True(t, got, "event for user-1 was not delivered")
	assert.Contains(t, line, `"event":"NoteCreated"`)
	assert.Contains(t, line, `"event_id":"ev-1"`)
	// The payload stays a bare signal — no note ids, no user ids.
	assert.NotContains(t, line, "note_id")
	assert.NotContains(t, line, "user_id")
}

func TestSSEHandlerAcceptsAuthorizationHeader(t *testing.T) {
	srv, mr, broker := newTestBrokerServer(t, sseConfig())

	token := generateTestToken("user-7", "test-secret", "access")
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	waitForClient(t, broker)

	publishEvent(t, mr, "user-7")
	line, got := newSSEReader(bufio.NewReader(resp.Body)).next(t, 3*time.Second)
	require.True(t, got)
	assert.Contains(t, line, "NoteCreated")
}

func TestSSEHandlerSkipAuthReceivesAll(t *testing.T) {
	cfg := sseConfig()
	cfg.SkipAuth = true
	srv, mr, broker := newTestBrokerServer(t, cfg)

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	waitForClient(t, broker)

	publishEvent(t, mr, "anyone")
	line, got := newSSEReader(bufio.NewReader(resp.Body)).next(t, 3*time.Second)
	require.True(t, got)
	assert.Contains(t, line, "NoteCreated")
}

func TestSSEBrokerDropsMalformedEvents(t *testing.T) {
	srv, mr, broker := newTestBrokerServer(t, sseConfig())

	token := generateTestToken("user-1", "test-secret", "access")
	resp, err := http.Get(srv.URL + "?access_token=" + token)
	require.NoError(t, err)
	defer resp.Body.Close()

	reader := newSSEReader(bufio.NewReader(resp.Body))
	waitForClient(t, broker)
	mr.Publish("graph:events", "not-json")
	mr.Publish("graph:events", `{"event_id":"x","event":"NoteCreated","payload":"broken"}`)
	_, got := reader.next(t, 300*time.Millisecond)
	assert.False(t, got, "malformed events must be dropped")

	// A good event still arrives afterwards — the subscription survived.
	publishEvent(t, mr, "user-1")
	_, got = reader.next(t, 3*time.Second)
	assert.True(t, got)
}
