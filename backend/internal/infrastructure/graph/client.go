package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/graph"

	"github.com/google/uuid"
)

// Client is an HTTP client for the graph-service analytics API.
type Client struct {
	baseURL      string
	internalToken string
	httpClient   *http.Client
}

// NewClient creates a graph-service client from configuration.
// If cfg.GraphServiceURL is empty, the client is disabled and all calls fail fast.
func NewClient(cfg *config.Config) *Client {
	if cfg == nil {
		return &Client{}
	}
	timeout := cfg.GraphServiceTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		baseURL:       cfg.GraphServiceURL,
		internalToken: cfg.GraphServiceInternalToken,
		httpClient:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) enabled() bool {
	return c.baseURL != "" && c.httpClient != nil
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values) (*http.Response, error) {
	if !c.enabled() {
		return nil, fmt.Errorf("graph service client disabled")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.internalToken != "" {
		req.Header.Set("X-Internal-Auth", c.internalToken)
	}
	if userID, ok := graph.UserIDFromContext(ctx); ok && userID != "" {
		req.Header.Set("X-User-Id", userID)
	}

	return c.httpClient.Do(req)
}

func (c *Client) GetRecommendations(ctx context.Context, noteID uuid.UUID, limit int) ([]graph.SuggestionResult, error) {
	q := url.Values{}
	q.Set("note_id", noteID.String())
	q.Set("limit", strconv.Itoa(limit))

	resp, err := c.request(ctx, http.MethodGet, "/api/v1/graph/recommendations", q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph service returned status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Recommendations []struct {
			NoteID        string  `json:"note_id"`
			Title         string  `json:"title"`
			Score         float64 `json:"score"`
			GraphScore    float64 `json:"graph_score"`
			SemanticScore float64 `json:"semantic_score"`
		} `json:"recommendations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	results := make([]graph.SuggestionResult, 0, len(payload.Recommendations))
	for _, r := range payload.Recommendations {
		id, err := uuid.Parse(r.NoteID)
		if err != nil {
			continue
		}
		results = append(results, graph.SuggestionResult{
			NodeID:        id,
			Title:         r.Title,
			Score:         r.Score,
			GraphScore:    r.GraphScore,
			SemanticScore: r.SemanticScore,
		})
	}

	return results, nil
}

func (c *Client) GetNeighbors(ctx context.Context, noteID uuid.UUID, depth int) (map[uuid.UUID]float64, error) {
	q := url.Values{}
	q.Set("depth", strconv.Itoa(depth))

	resp, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/api/v1/graph/note/%s/neighbors", noteID.String()), q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph service returned status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Nodes []struct {
			ID     string  `json:"id"`
			Weight float64 `json:"weight"`
		} `json:"nodes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	weights := make(map[uuid.UUID]float64, len(payload.Nodes))
	for _, n := range payload.Nodes {
		id, err := uuid.Parse(n.ID)
		if err != nil {
			continue
		}
		weights[id] = n.Weight
	}

	return weights, nil
}

func (c *Client) GetPath(ctx context.Context, from, to uuid.UUID) ([]uuid.UUID, error) {
	q := url.Values{}
	q.Set("from", from.String())
	q.Set("to", to.String())

	resp, err := c.request(ctx, http.MethodGet, "/api/v1/graph/path", q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph service returned status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		NoteIDs []string `json:"note_ids"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	path := make([]uuid.UUID, 0, len(payload.NoteIDs))
	for _, s := range payload.NoteIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		path = append(path, id)
	}

	return path, nil
}
