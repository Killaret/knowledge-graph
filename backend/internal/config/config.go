package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

// JSONConfig represents the structure of knowledge-graph.config.json
type JSONConfig struct {
	Backend struct {
		Server struct {
			RateLimit struct {
				Enabled       bool           `json:"enabled"`
				Requests      int            `json:"requests"`
				WindowSeconds int            `json:"window_seconds"`
				Endpoints     map[string]int `json:"endpoints"`
				FallbackPorts []string       `json:"fallback_ports"`
			} `json:"rate_limit"`
		} `json:"server"`
		Database struct {
			RetryMaxAttempts      int  `json:"retry_max_attempts"`
			RetryDelaySeconds     int  `json:"retry_delay_seconds"`
			MigrationsFailOnError bool `json:"migrations_fail_on_error"`
		} `json:"database"`
		Search struct {
			FulltextLanguages []string           `json:"fulltext_languages"`
			RankingWeights    map[string]float64 `json:"ranking_weights"`
			FallbackToILike   bool               `json:"fallback_to_ilike"`
		} `json:"search"`
		Recommendation struct {
			Depth                   int     `json:"depth"`
			Decay                   float64 `json:"decay"`
			TopN                    int     `json:"top_n"`
			Alpha                   float64 `json:"alpha"`
			Beta                    float64 `json:"beta"`
			Gamma                   float64 `json:"gamma"`
			CacheTTLSeconds         int     `json:"cache_ttl_seconds"`
			TaskDelaySeconds        int     `json:"task_delay_seconds"`
			BatchRateLimit          int     `json:"batch_rate_limit"`
			FallbackEnabled         bool    `json:"fallback_enabled"`
			FallbackTTLSeconds      int     `json:"fallback_ttl_seconds"`
			FallbackSemanticEnabled bool    `json:"fallback_semantic_enabled"`
			KeywordEnabled          bool    `json:"keyword_enabled"`
			BFSAggregation          string  `json:"bfs_aggregation"`
			BFSNormalize            bool    `json:"bfs_normalize"`
			KeywordSimilarityMethod string  `json:"keyword_similarity_method"`
			KeywordTverskyAlpha     float64 `json:"keyword_tversky_alpha"`
			KeywordTverskyBeta      float64 `json:"keyword_tversky_beta"`
		} `json:"recommendation"`
		Pagination struct {
			DefaultLimit int `json:"default_limit"`
			MaxLimit     int `json:"max_limit"`
		} `json:"pagination"`
		Graph struct {
			LoadDepth        int `json:"load_depth"`
			MaxNodes         int `json:"max_nodes"`
			DefaultLimit     int `json:"default_limit"`
			MaxLimit         int `json:"max_limit"`
			LinkDefaultLimit int `json:"link_default_limit"`
			LinkMaxLimit     int `json:"link_max_limit"`
		} `json:"graph"`
		Embedding struct {
			SimilarityLimit int `json:"similarity_limit"`
		} `json:"embedding"`
		Asynq struct {
			Concurrency  int `json:"concurrency"`
			QueueDefault int `json:"queue_default"`
			QueueMaxLen  int `json:"queue_max_len"`
		} `json:"asynq"`
	} `json:"backend"`
}

type Config struct {
	// Сервер
	ServerPort                   string
	ServerRateLimitEnabled       bool
	ServerRateLimitRequests      int
	ServerRateLimitWindowSeconds int
	ServerRateLimitEndpoints     map[string]int
	ServerFallbackPorts          []string

	// База данных
	DatabaseURL               string
	DatabaseRetryMaxAttempts  int
	DatabaseRetryDelaySeconds int
	MigrationsFailOnError     bool

	// Redis
	RedisURL string

	// NLP
	NLPServiceURL string

	// Search
	SearchFulltextLanguages []string
	SearchRankingWeights    map[string]float64
	SearchFallbackToILike   bool

	// Рекомендации (граф)
	RecommendationAlpha      float64
	RecommendationBeta       float64
	RecommendationDepth      int
	RecommendationDecay      float64
	RecommendationCacheTTL   time.Duration
	EmbeddingSimilarityLimit int

	// Загрузка графа (визуализация)
	GraphLoadDepth int // Глубина загрузки графа для визуализации

	// Pagination
	PaginationDefaultLimit int
	PaginationMaxLimit     int

	// Graph API limits
	GraphDefaultLimit     int
	GraphMaxLimit         int
	GraphLinkDefaultLimit int
	GraphLinkMaxLimit     int

	// Новые параметры для улучшения алгоритма рекомендаций
	RecommendationGamma                   float64       // Коэффициент для дополнительного компонента
	BFSAggregation                        string        // Метод агрегации BFS: "max", "sum", "avg"
	BFSNormalize                          bool          // Нормализация весов в BFS
	RecommendationTopN                    int           // Количество рекомендаций для сохранения
	RecommendationTaskDelaySeconds        int           // Задержка перед выполнением задачи (dedup)
	RecommendationBatchRateLimit          int           // Rate limit для batch обработки
	RecommendationFallbackEnabled         bool          // Включить fallback на Redis
	RecommendationFallbackTTL             time.Duration // TTL для fallback-кэша
	RecommendationFallbackSemanticEnabled bool          // Включить fallback на семантических соседей
	RecommendationKeywordEnabled          bool          // Включить keyword-компонент (gamma)
	RecommendationKeywordSimilarityMethod string        // Метод сходства ключевых слов: jaccard, overlap, tversky, weighted_jaccard, cosine
	RecommendationKeywordTverskyAlpha   float64       // Alpha параметр для Tversky index
	RecommendationKeywordTverskyBeta    float64       // Beta параметр для Tversky index
	AsynqConcurrency                      int           // Уровень параллелизма Asynq
	AsynqQueueDefault                     int           // Приоритет дефолтной очереди Asynq
	AsynqQueueMaxLen                      int           // Максимальная длина очереди
}

// loadJSONConfig загружает конфигурацию из knowledge-graph.config.json
// Возвращает nil если файл не существует или не может быть прочитан
func loadJSONConfig() *JSONConfig {
	// Определяем путь к корню проекта (где находится knowledge-graph.config.json)
	// Пробуем несколько вариантов поиска файла
	possiblePaths := []string{
		"knowledge-graph.config.json", // текущая директория
	}

	// Добавляем путь относительно расположения этого файла (backend/internal/config)
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// От корня проекта: backend/internal/config -> ../../../knowledge-graph.config.json
		configDir := filepath.Dir(filename)
		projectRoot := filepath.Join(configDir, "..", "..", "..")
		possiblePaths = append(possiblePaths, filepath.Join(projectRoot, "knowledge-graph.config.json"))
	}

	var data []byte
	var err error

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			log.Printf("[Config] Loading JSON config from: %s", path)
			break
		}
	}

	if err != nil {
		// Файл не найден - это нормально, используем только env vars
		log.Printf("[Config] knowledge-graph.config.json not found, using env vars only")
		return nil
	}

	var jsonCfg JSONConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		log.Printf("[Config] Failed to parse knowledge-graph.config.json: %v, using env vars only", err)
		return nil
	}

	log.Printf("[Config] JSON config loaded successfully")
	return &jsonCfg
}

// Load загружает конфигурацию из переменных окружения.
// Если переменная не задана, используется значение по умолчанию из JSON-конфига (если он есть),
// иначе используется встроенное значение по умолчанию.
// Для обязательных переменных (DatabaseURL) используется mustGetEnv.
func Load() (*Config, error) {
	// Загружаем JSON конфиг как источник дефолтных значений
	jsonCfg := loadJSONConfig()

	cfg := &Config{
		// Server
		ServerPort:                   getEnv("SERVER_PORT", "8080"),
		ServerRateLimitEnabled:       getBoolEnv("SERVER_RATE_LIMIT_ENABLED", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Server.RateLimit.Enabled }, true)),
		ServerRateLimitRequests:      getIntEnv("SERVER_RATE_LIMIT_REQUESTS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Server.RateLimit.Requests }, 100)),
		ServerRateLimitWindowSeconds: getIntEnv("SERVER_RATE_LIMIT_WINDOW_SECONDS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Server.RateLimit.WindowSeconds }, 60)),
		ServerRateLimitEndpoints:     nil, // Loaded separately below
		ServerFallbackPorts:          nil, // Loaded separately below

		// Database
		DatabaseURL:               "", // Will be set below after error check
		DatabaseRetryMaxAttempts:  getIntEnv("DATABASE_RETRY_MAX_ATTEMPTS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Database.RetryMaxAttempts }, 3)),
		DatabaseRetryDelaySeconds: getIntEnv("DATABASE_RETRY_DELAY_SECONDS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Database.RetryDelaySeconds }, 5)),
		MigrationsFailOnError:     getBoolEnv("MIGRATIONS_FAIL_ON_ERROR", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Database.MigrationsFailOnError }, false)),

		// Redis & NLP
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		NLPServiceURL: getEnv("NLP_SERVICE_URL", "http://localhost:5000"),

		// Search
		SearchFulltextLanguages: getJSONStringSliceOrDefault(jsonCfg, func(j *JSONConfig) []string { return j.Backend.Search.FulltextLanguages }, []string{"russian", "simple"}),
		SearchRankingWeights:    nil, // Loaded separately below
		SearchFallbackToILike:   getBoolEnv("SEARCH_FALLBACK_TO_ILIKE", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Search.FallbackToILike }, true)),

		// Recommendation settings (env var > JSON config > built-in default)
		RecommendationAlpha:      getFloatEnv("RECOMMENDATION_ALPHA", getJSONFloatOrDefault(jsonCfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.Alpha }, 0.5)),
		RecommendationBeta:       getFloatEnv("RECOMMENDATION_BETA", getJSONFloatOrDefault(jsonCfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.Beta }, 0.5)),
		RecommendationDepth:      getIntEnv("RECOMMENDATION_DEPTH", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Recommendation.Depth }, 3)),
		RecommendationDecay:      getFloatEnv("RECOMMENDATION_DECAY", getJSONFloatOrDefault(jsonCfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.Decay }, 0.5)),
		RecommendationCacheTTL:   time.Duration(getIntEnv("RECOMMENDATION_CACHE_TTL_SECONDS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Recommendation.CacheTTLSeconds }, 300))) * time.Second,
		EmbeddingSimilarityLimit: getIntEnv("EMBEDDING_SIMILARITY_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Embedding.SimilarityLimit }, 30)),

		// Загрузка графа
		GraphLoadDepth: getIntEnv("GRAPH_LOAD_DEPTH", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Graph.LoadDepth }, 2)),

		// Pagination
		PaginationDefaultLimit: getIntEnv("PAGINATION_DEFAULT_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Pagination.DefaultLimit }, 20)),
		PaginationMaxLimit:     getIntEnv("PAGINATION_MAX_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Pagination.MaxLimit }, 100)),

		// Graph API limits
		GraphDefaultLimit:     getIntEnv("GRAPH_DEFAULT_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Graph.DefaultLimit }, 100)),
		GraphMaxLimit:         getIntEnv("GRAPH_MAX_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Graph.MaxLimit }, 1000)),
		GraphLinkDefaultLimit: getIntEnv("GRAPH_LINK_DEFAULT_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Graph.LinkDefaultLimit }, 500)),
		GraphLinkMaxLimit:     getIntEnv("GRAPH_LINK_MAX_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Graph.LinkMaxLimit }, 5000)),

		// Новые параметры
		RecommendationGamma:                   getFloatEnv("RECOMMENDATION_GAMMA", getJSONFloatOrDefault(jsonCfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.Gamma }, 0.2)),
		BFSAggregation:                        getEnv("BFS_AGGREGATION", getJSONStringOrDefault(jsonCfg, func(j *JSONConfig) string { return j.Backend.Recommendation.BFSAggregation }, "max")),
		BFSNormalize:                          getBoolEnv("BFS_NORMALIZE", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Recommendation.BFSNormalize }, true)),
		RecommendationTopN:                    getIntEnv("RECOMMENDATION_TOP_N", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Recommendation.TopN }, 20)),
		RecommendationTaskDelaySeconds:        getIntEnv("RECOMMENDATION_TASK_DELAY_SECONDS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Recommendation.TaskDelaySeconds }, 5)),
		RecommendationBatchRateLimit:          getIntEnv("RECOMMENDATION_BATCH_RATE_LIMIT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Recommendation.BatchRateLimit }, 10)),
		RecommendationFallbackEnabled:         getBoolEnv("RECOMMENDATION_FALLBACK_ENABLED", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Recommendation.FallbackEnabled }, true)),
		RecommendationFallbackTTL:             time.Duration(getIntEnv("RECOMMENDATION_FALLBACK_TTL_SECONDS", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Recommendation.FallbackTTLSeconds }, 3600))) * time.Second,
		RecommendationFallbackSemanticEnabled: getBoolEnv("RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Recommendation.FallbackSemanticEnabled }, true)),
		RecommendationKeywordEnabled:          getBoolEnv("RECOMMENDATION_KEYWORD_ENABLED", getJSONBoolOrDefault(jsonCfg, func(j *JSONConfig) bool { return j.Backend.Recommendation.KeywordEnabled }, true)),
		RecommendationKeywordSimilarityMethod: getEnv("RECOMMENDATION_KEYWORD_SIMILARITY_METHOD", getJSONStringOrDefault(jsonCfg, func(j *JSONConfig) string { return j.Backend.Recommendation.KeywordSimilarityMethod }, "jaccard")),
		RecommendationKeywordTverskyAlpha:     getFloatEnv("RECOMMENDATION_KEYWORD_TVERSKY_ALPHA", getJSONFloatOrDefault(jsonCfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.KeywordTverskyAlpha }, 0.5)),
		RecommendationKeywordTverskyBeta:      getFloatEnv("RECOMMENDATION_KEYWORD_TVERSKY_BETA", getJSONFloatOrDefault(jsonCfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.KeywordTverskyBeta }, 0.5)),
		AsynqConcurrency:                      getIntEnv("ASYNQ_CONCURRENCY", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Asynq.Concurrency }, 10)),
		AsynqQueueDefault:                     getIntEnv("ASYNQ_QUEUE_DEFAULT", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Asynq.QueueDefault }, 1)),
		AsynqQueueMaxLen:                      getIntEnv("ASYNQ_QUEUE_MAX_LEN", getJSONIntOrDefault(jsonCfg, func(j *JSONConfig) int { return j.Backend.Asynq.QueueMaxLen }, 10000)),
	}

	// Load complex types from JSON (no env var override for these)
	if jsonCfg != nil {
		cfg.ServerRateLimitEndpoints = getJSONIntMapOrDefault(jsonCfg, func(j *JSONConfig) map[string]int { return j.Backend.Server.RateLimit.Endpoints }, map[string]int{
			"notes_create": 30,
			"links_create": 50,
			"notes_update": 20,
		})
		cfg.ServerFallbackPorts = getJSONStringSliceOrDefault(jsonCfg, func(j *JSONConfig) []string { return j.Backend.Server.RateLimit.FallbackPorts }, []string{"8081", "8082"})
		cfg.SearchRankingWeights = getJSONFloatMapOrDefault(jsonCfg, func(j *JSONConfig) map[string]float64 { return j.Backend.Search.RankingWeights }, map[string]float64{
			"russian": 1.0,
			"simple":  1.0,
		})
	} else {
		// Set defaults when no JSON config
		cfg.ServerRateLimitEndpoints = map[string]int{
			"notes_create": 30,
			"links_create": 50,
			"notes_update": 20,
		}
		cfg.ServerFallbackPorts = []string{"8081", "8082"}
		cfg.SearchRankingWeights = map[string]float64{
			"russian": 1.0,
			"simple":  1.0,
		}
	}

	// Load required environment variable
	dbURL, err := mustGetEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	cfg.DatabaseURL = dbURL

	return cfg, nil
}

// Helper functions for JSON config with fallbacks
func getJSONIntOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) int, defaultValue int) int {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

func getJSONFloatOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) float64, defaultValue float64) float64 {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

func getJSONStringOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) string, defaultValue string) string {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

func getJSONBoolOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) bool, defaultValue bool) bool {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

func getJSONStringSliceOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) []string, defaultValue []string) []string {
	if jsonCfg == nil {
		return defaultValue
	}
	result := getter(jsonCfg)
	if result == nil {
		return defaultValue
	}
	return result
}

func getJSONIntMapOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) map[string]int, defaultValue map[string]int) map[string]int {
	if jsonCfg == nil {
		return defaultValue
	}
	result := getter(jsonCfg)
	if result == nil {
		return defaultValue
	}
	return result
}

func getJSONFloatMapOrDefault(jsonCfg *JSONConfig, getter func(*JSONConfig) map[string]float64, defaultValue map[string]float64) map[string]float64 {
	if jsonCfg == nil {
		return defaultValue
	}
	result := getter(jsonCfg)
	if result == nil {
		return defaultValue
	}
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}

func getIntEnv(key string, defaultValue int) int {
	if str := os.Getenv(key); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if str := os.Getenv(key); str != "" {
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return val
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if str := os.Getenv(key); str != "" {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return defaultValue
}
