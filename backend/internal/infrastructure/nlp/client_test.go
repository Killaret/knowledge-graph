package nlp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNLPClient_ExtractKeywords_Success тестирует успешное извлечение ключевых слов
func TestNLPClient_ExtractKeywords_Success(t *testing.T) {
	// Создаём тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод и путь
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/extract_keywords" {
			t.Errorf("expected /extract_keywords, got %s", r.URL.Path)
		}

		// Проверяем Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		// Проверяем тело запроса
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Возвращаем тестовый ответ
		response := map[string]interface{}{
			"keywords": []Keyword{
				{Keyword: "machine", Weight: 0.8},
				{Keyword: "learning", Weight: 0.7},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Создаём клиент
	client := NewNLPClient(server.URL, nil, 5*time.Minute)

	// Вызываем метод
	ctx := context.Background()
	keywords, err := client.ExtractKeywords(ctx, "machine learning", 5)

	// Проверяем результат
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(keywords))
	}
	if keywords[0].Keyword != "machine" {
		t.Errorf("expected first keyword 'machine', got %s", keywords[0].Keyword)
	}
	if keywords[0].Weight != 0.8 {
		t.Errorf("expected weight 0.8, got %f", keywords[0].Weight)
	}
}

// TestNLPClient_ExtractKeywords_HTTPError тестирует обработку HTTP ошибок
func TestNLPClient_ExtractKeywords_HTTPError(t *testing.T) {
	// Создаём тестовый сервер, возвращающий ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewNLPClient(server.URL, nil, 5*time.Minute)
	ctx := context.Background()

	_, err := client.ExtractKeywords(ctx, "test", 5)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Проверяем что ошибка содержит информацию о статусе 500
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain '500', got %v", err)
	}
}

// TestNLPClient_ExtractKeywords_InvalidJSON тестирует обработку невалидного JSON
func TestNLPClient_ExtractKeywords_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewNLPClient(server.URL, nil, 5*time.Minute)
	ctx := context.Background()

	_, err := client.ExtractKeywords(ctx, "test", 5)

	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestNLPClient_Embed_Success тестирует успешное получение эмбеддинга
func TestNLPClient_Embed_Success(t *testing.T) {
	expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embed" {
			t.Errorf("expected /embed, got %s", r.URL.Path)
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		response := map[string]interface{}{
			"embedding": expectedEmbedding,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNLPClient(server.URL, nil, 5*time.Minute)
	ctx := context.Background()

	embedding, err := client.Embed(ctx, "test text")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(embedding) != len(expectedEmbedding) {
		t.Errorf("expected embedding length %d, got %d", len(expectedEmbedding), len(embedding))
	}
	for i, v := range expectedEmbedding {
		if embedding[i] != v {
			t.Errorf("expected embedding[%d] = %f, got %f", i, v, embedding[i])
		}
	}
}

// TestNLPClient_Embed_CacheHit тестирует получение эмбеддинга из кэша
func TestNLPClient_Embed_CacheHit(t *testing.T) {
	expectedEmbedding := []float32{0.5, 0.6, 0.7}
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"embedding": expectedEmbedding,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Создаём мок Redis клиента (nil - кэширование отключено для этого теста)
	client := NewNLPClient(server.URL, nil, 5*time.Minute)
	ctx := context.Background()

	// Первый вызов
	_, err := client.Embed(ctx, "test text")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Второй вызов с тем же текстом (без кэша Redis должен сделать новый запрос)
	_, err = client.Embed(ctx, "test text")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	// Проверяем что сервер был вызван 2 раза (без кэша)
	if callCount != 2 {
		t.Errorf("expected 2 calls without cache, got %d", callCount)
	}
}

// TestNLPClient_Embed_EmptyText тестирует эмбеддинг пустого текста
func TestNLPClient_Embed_EmptyText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"embedding": []float32{0.0, 0.0, 0.0},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNLPClient(server.URL, nil, 5*time.Minute)
	ctx := context.Background()

	_, err := client.Embed(ctx, "")

	if err != nil {
		t.Fatalf("unexpected error for empty text: %v", err)
	}
}

// TestNLPClient_ExtractKeywords_EmptyResponse тестирует пустой ответ
func TestNLPClient_ExtractKeywords_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"keywords": []Keyword{},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNLPClient(server.URL, nil, 5*time.Minute)
	ctx := context.Background()

	keywords, err := client.ExtractKeywords(ctx, "test", 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keywords) != 0 {
		t.Errorf("expected 0 keywords, got %d", len(keywords))
	}
}

// TestNLPClient_NewNLPClient тестирует создание клиента
func TestNLPClient_NewNLPClient(t *testing.T) {
	client := NewNLPClient("http://localhost:8080", nil, 5*time.Minute)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL http://localhost:8080, got %s", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", client.httpClient.Timeout)
	}
}

// TestKeywordStruct тестирует структуру Keyword
func TestKeywordStruct(t *testing.T) {
	kw := Keyword{
		Keyword: "test",
		Weight:  0.9,
	}

	if kw.Keyword != "test" {
		t.Errorf("expected keyword 'test', got %s", kw.Keyword)
	}
	if kw.Weight != 0.9 {
		t.Errorf("expected weight 0.9, got %f", kw.Weight)
	}
}
