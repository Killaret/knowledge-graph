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

	"github.com/cenkalti/backoff/v4"
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
	maxRetries uint64
	retryDelay time.Duration
}

// NewNLPClient создаёт новый клиент
// redisClient может быть nil, тогда кэширование отключено
func NewNLPClient(baseURL string, redisClient *redis.Client, cacheTTL time.Duration) *NLPClient {
	return &NLPClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		redis:      redisClient,
		cacheTTL:   cacheTTL,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

// requestBuilder создаёт новый HTTP запрос с заданным телом
type requestBuilder func() (*http.Request, error)

// doWithRetry выполняет HTTP запрос с retry логикой и exponential backoff
func (c *NLPClient) doWithRetry(ctx context.Context, buildReq requestBuilder) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	// Настраиваем exponential backoff
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = c.retryDelay
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 15 * time.Second

	op := func() error {
		// Создаём новый запрос для каждой попытки
		req, err := buildReq()
		if err != nil {
			return backoff.Permanent(err)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			return fmt.Errorf("http request failed: %w", err)
		}

		// Retry on 5xx errors and 429 (rate limit)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("nlp service returned %d", resp.StatusCode)
			return lastErr
		}

		return nil
	}

	notify := func(err error, duration time.Duration) {
		fmt.Printf("[NLP] Retry after %v: %v\n", duration, err)
	}

	// Выполняем с retry
	if err := backoff.RetryNotify(op, backoff.WithMaxRetries(b, c.maxRetries), notify); err != nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return resp, nil
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

	buildReq := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/extract_keywords", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}

	resp, err := c.doWithRetry(ctx, buildReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed after retries: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
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

	buildReq := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}

	resp, err := c.doWithRetry(ctx, buildReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed after retries: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
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

// HealthCheck проверяет доступность NLP сервиса
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
