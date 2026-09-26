//go:build !integration
// +build !integration

package notehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appquality "knowledge-graph/internal/application/quality"
	"knowledge-graph/internal/config"
)

type fakeQualityReader struct{ entry *appquality.LogEntry }

func (f *fakeQualityReader) LatestQuality(context.Context, uuid.UUID) (*appquality.LogEntry, error) {
	return f.entry, nil
}

func (f *fakeQualityReader) ListQuality(_ context.Context, _ uuid.UUID, _ int) ([]appquality.LogEntry, error) {
	if f.entry == nil {
		return nil, nil
	}
	return []appquality.LogEntry{*f.entry}, nil
}

type fakeQualityEnq struct{ calls []string }

func (f *fakeQualityEnq) EnqueueAssessQuality(_ context.Context, noteID string, trigger string) error {
	f.calls = append(f.calls, noteID+":"+trigger)
	return nil
}

func newQualityHandler(t *testing.T) *Handler {
	cfg := &config.Config{}
	return New(nil, nil, nil, nil, 0, nil, nil, nil, cfg, nil, nil, nil)
}

func qualityRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/notes/:id/quality", h.GetQuality)
	r.POST("/notes/:id/quality/assess", h.AssessQuality)
	return r
}

func TestGetQuality_Disabled(t *testing.T) {
	h := newQualityHandler(t) // SetQuality never called
	r := qualityRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes/"+uuid.NewString()+"/quality", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled": false}`, rec.Body.String())
}

func TestGetQuality_NoAssessmentYet(t *testing.T) {
	h := newQualityHandler(t)
	h.SetQuality(true, &fakeQualityReader{}, &fakeQualityEnq{})
	r := qualityRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes/"+uuid.NewString()+"/quality", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, true, body["enabled"])
	assert.Nil(t, body["quality"])
}

func TestGetQuality_ReturnsRecord(t *testing.T) {
	entry := &appquality.LogEntry{
		NoteID:     uuid.New(),
		SourceHash: "abc123",
		Trigger:    appquality.TriggerAuto,
		Record: appquality.Record{
			Signals: appquality.Signals{
				Kind:              appquality.KindStub,
				TruncatedByImport: true,
				HasSourceLink:     true,
			},
			Gates:           []string{appquality.GateStub, appquality.GateTruncated},
			Verdict:         appquality.VerdictEnrich,
			Reasons:         []string{appquality.GateStub, appquality.GateTruncated},
			Attempt:         1,
			ComputedAt:      time.Now().UTC(),
			PipelineVersion: appquality.PipelineVersion,
		},
	}
	h := newQualityHandler(t)
	h.SetQuality(true, &fakeQualityReader{entry: entry}, &fakeQualityEnq{})
	r := qualityRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes/"+entry.NoteID.String()+"/quality", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Enabled bool `json:"enabled"`
		Quality struct {
			Verdict    string   `json:"verdict"`
			Reasons    []string `json:"reasons"`
			Attempt    int      `json:"attempt"`
			CanRefetch bool     `json:"can_refetch"`
			Signals    struct {
				Kind              string `json:"kind"`
				TruncatedByImport bool   `json:"truncated_by_import"`
			} `json:"signals"`
		} `json:"quality"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Enabled)
	assert.Equal(t, "enrich", body.Quality.Verdict)
	assert.Equal(t, []string{"stub", "truncated"}, body.Quality.Reasons)
	assert.Equal(t, 1, body.Quality.Attempt)
	// truncated + source link => the re-fetch action is offered
	assert.True(t, body.Quality.CanRefetch)
	assert.Equal(t, "stub", body.Quality.Signals.Kind)
	assert.True(t, body.Quality.Signals.TruncatedByImport)
}

func TestGetQuality_BadID(t *testing.T) {
	h := newQualityHandler(t)
	h.SetQuality(true, &fakeQualityReader{}, nil)
	r := qualityRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes/not-a-uuid/quality", nil))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssessQuality_EnqueuesManual(t *testing.T) {
	enq := &fakeQualityEnq{}
	h := newQualityHandler(t)
	h.SetQuality(true, &fakeQualityReader{}, enq)
	r := qualityRouter(h)

	noteID := uuid.NewString()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+noteID+"/quality/assess", nil))
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Len(t, enq.calls, 1)
	assert.Equal(t, noteID+":manual", enq.calls[0])
}

func TestAssessQuality_Disabled(t *testing.T) {
	enq := &fakeQualityEnq{}
	h := newQualityHandler(t) // enabled=false — nothing is set
	r := qualityRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+uuid.NewString()+"/quality/assess", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled": false}`, rec.Body.String())
	assert.Empty(t, enq.calls)
}
