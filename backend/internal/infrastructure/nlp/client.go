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

// Keyword представляет одно ключевое слово с весом
type Keyword struct {
	Keyword string  `json:"keyword"`
	Weight  float64 `json:"weight"`
}

// NLPClient — клиент для вызова Python-микросервиса
type NLPClient struct {
	httpClient *http.Client
	baseURL    string
	redis      *redis.Client
	cacheTTL   time.Duration
}

// NewNLPClient создаёт новый клиент
// redisClient может быть nil, тогда кэширование отключено
func NewNLPClient(baseURL string, redisClient *redis.Client, cacheTTL time.Duration) *NLPClient {
	return &NLPClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		redis:      redisClient,
		cacheTTL:   cacheTTL,
	}
}

// ExtractKeywords вызывает /extract_keywords и возвращает список ключевых слов
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

// Embed вызывает /embed и возвращает вектор ([]float32)
// Результат кэшируется в Redis (если redisClient передан)
func (c *NLPClient) Embed(ctx context.Context, text string) ([]float32, error) {
	// Проверяем кэш
	if c.redis != nil {
		hash := sha256.Sum256([]byte(text))
		key := "embed:" + hex.EncodeToString(hash[:])
		cached, err := c.redis.Get(ctx, key).Bytes()
		if err == nil {
			// десериализуем из JSON
			var embedding []float32
			if err := json.Unmarshal(cached, &embedding); err == nil {
				return embedding, nil
			}
			// если ошибка декодирования, просто продолжаем
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

	// Сохраняем в кэш
	if c.redis != nil && len(result.Embedding) > 0 {
		hash := sha256.Sum256([]byte(text))
		key := "embed:" + hex.EncodeToString(hash[:])
		data, _ := json.Marshal(result.Embedding)
		_ = c.redis.Set(ctx, key, data, c.cacheTTL).Err() // ошибку игнорируем
	}

	return result.Embedding, nil
}
