//go:build integration
// +build integration

package queue

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	infevents "knowledge-graph/internal/infrastructure/events"
	"knowledge-graph/internal/infrastructure/nlp"
	"knowledge-graph/internal/testutil"
)

// identicalVec builds a 384-dim vector; identical vectors give cosine 1.0.
func identicalVec(v float32) []float32 {
	vec := make([]float32, 384)
	for i := range vec {
		vec[i] = v
	}
	return vec
}

// LINKS-1 integration: a compute-embedding task stores the embedding, then
// creates at most two gamma links and publishes one LinkCreated per link.
func TestWorker_ComputeEmbeddingCreatesGammaLinks(t *testing.T) {
	database, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	database.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err := database.AutoMigrate(&postgres.UserModel{}, &postgres.NoteModel{}, &postgres.NoteEmbeddingModel{}, &postgres.LinkModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	ctx := context.Background()
	noteRepo := postgres.NewNoteRepository(database, nil)
	embeddingRepo := postgres.NewEmbeddingRepository(database, "test-model")
	linkRepo := postgres.NewLinkRepository(database)

	saveNote := func(title string) *note.Note {
		tv, _ := note.NewTitle(title)
		cv, _ := note.NewContent("body of " + title)
		md, _ := note.NewMetadata(nil)
		n := note.NewNote(tv, cv, note.MustType("star"), md)
		require.NoError(t, noteRepo.Save(ctx, n))
		return n
	}

	source := saveNote("source")
	targets := []*note.Note{saveNote("t1"), saveNote("t2"), saveNote("t3")}

	// Targets already have embeddings identical to what the NLP stub returns.
	vec := pgvector.NewVector(identicalVec(0.1))
	for _, tn := range targets {
		require.NoError(t, embeddingRepo.Upsert(ctx, tn.ID(), vec))
	}

	nlpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/embed", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"embedding": identicalVec(0.1)})
	}))
	defer nlpServer.Close()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	publisher := infevents.NewPublisher(rdb, "graph:events")

	sub := rdb.Subscribe(ctx, "graph:events")
	defer sub.Close()
	for {
		msg, err := sub.ReceiveTimeout(ctx, time.Second)
		if err != nil {
			t.Fatal("subscription not confirmed")
		}
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	nlpClient := nlp.NewNLPClient(nlpServer.URL, nil, time.Hour)
	gammaGen := recommendation.NewGammaLinkGenerator(embeddingRepo, linkRepo, 2, 0.6)
	w := NewWorker(noteRepo, nil, embeddingRepo, nlpClient, nil, nil, gammaGen, publisher, nil, 0)

	payload, err := json.Marshal(ComputeEmbeddingTaskPayload{NoteID: source.ID().String()})
	require.NoError(t, err)
	task := asynq.NewTask(TypeComputeEmbedding, payload)

	require.NoError(t, w.HandleComputeEmbedding(ctx, task))

	// At most two gamma links, all from the source, all source_type='gamma'.
	links, err := linkRepo.FindBySourceType(ctx, "gamma")
	require.NoError(t, err)
	require.LessOrEqual(t, len(links), 2)
	require.NotEmpty(t, links)
	for _, l := range links {
		assert.Equal(t, source.ID(), l.SourceNoteID())
		assert.Equal(t, "gamma", l.SourceType().String())
	}

	// One LinkCreated event per created link.
	for i := 0; i < len(links); i++ {
		msg, err := sub.ReceiveTimeout(ctx, 2*time.Second)
		require.NoError(t, err, "expected LinkCreated event %d", i)
		pubMsg, ok := msg.(*redis.Message)
		require.True(t, ok)

		var ev infevents.Event
		require.NoError(t, json.Unmarshal([]byte(pubMsg.Payload), &ev))
		assert.Equal(t, "LinkCreated", ev.Event)

		var lp infevents.LinkEventPayload
		require.NoError(t, json.Unmarshal(ev.Payload, &lp))
		assert.Equal(t, source.ID().String(), lp.SourceNoteID)
	}
}

// Deleting a target note cascades its gamma links — FK ON DELETE CASCADE on
// note_links must not leave orphan rows.
func TestWorker_GammaLinkRemovedWithTarget(t *testing.T) {
	database, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	database.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err := database.AutoMigrate(&postgres.UserModel{}, &postgres.NoteModel{}, &postgres.NoteEmbeddingModel{}, &postgres.LinkModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	ctx := context.Background()
	noteRepo := postgres.NewNoteRepository(database, nil)
	linkRepo := postgres.NewLinkRepository(database)

	saveNote := func(title string) *note.Note {
		tv, _ := note.NewTitle(title)
		cv, _ := note.NewContent("body")
		md, _ := note.NewMetadata(nil)
		n := note.NewNote(tv, cv, note.MustType("star"), md)
		require.NoError(t, noteRepo.Save(ctx, n))
		return n
	}

	source := saveNote("src")
	target := saveNote("tgt")

	l, err := linkRepo.FindBySourceType(ctx, "gamma")
	require.NoError(t, err)
	require.Empty(t, l)

	// Insert a gamma link directly.
	gamma := newGammaLink(t, source.ID(), target.ID())
	require.NoError(t, linkRepo.Save(ctx, gamma))

	require.NoError(t, noteRepo.Delete(ctx, target.ID()))

	links, err := linkRepo.FindBySourceType(ctx, "gamma")
	require.NoError(t, err)
	assert.Empty(t, links, "gamma link must be removed when the target note is deleted")
}
