package nlp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// Keyword represents a single keyword with a weight
type Keyword struct {
	Keyword string  `json:"keyword"`
	Weight  float64 `json:"weight"`
}

// NLPClient is the client for calling the Python microservice
type NLPClient struct {
	httpClient *http.Client
	baseURL    string
	redis      *redis.Client
	cacheTTL   time.Duration
}

// NewNLPClient creates a new client
// redisClient may be nil, in which case caching is disabled
func NewNLPClient(baseURL string, redisClient *redis.Client, cacheTTL time.Duration) *NLPClient {
	return &NLPClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		redis:      redisClient,
		cacheTTL:   cacheTTL,
	}
}

// ExtractKeywords calls /extract_keywords and returns the list of keywords
func (c *NLPClient) ExtractKeywords(ctx context.Context, text string, topN int) ([]Keyword, error) {
	reqBody := map[string]interface{}{
		"text":  text,
		"top_n": topN,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/extract_keywords", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nlp service returned %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Keywords []Keyword `json:"keywords"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Keywords, nil
}

// Embed calls /embed and returns the vector ([]float32)
// The result is cached in Redis (if redisClient is provided)
func (c *NLPClient) Embed(ctx context.Context, text string) ([]float32, error) {
	// Check the cache
	if c.redis != nil {
		hash := sha256.Sum256([]byte(text))
		key := "embed:" + hex.EncodeToString(hash[:])
		cached, err := c.redis.Get(ctx, key).Bytes()
		if err == nil {
			// deserialize from JSON
			var embedding []float32
			if err := json.Unmarshal(cached, &embedding); err == nil {
				return embedding, nil
			}
			// on a decoding error, just continue
		}
	}

	reqBody := map[string]interface{}{"text": text}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nlp service returned %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Store in the cache
	if c.redis != nil && len(result.Embedding) > 0 {
		hash := sha256.Sum256([]byte(text))
		key := "embed:" + hex.EncodeToString(hash[:])
		data, _ := json.Marshal(result.Embedding)
		_ = c.redis.Set(ctx, key, data, c.cacheTTL).Err() // ignore the error
	}

	return result.Embedding, nil
}

// HealthCheck checks the availability of the NLP service
func (c *NLPClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nlp health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nlp health check returned status %d", resp.StatusCode)
	}

	return nil
}
