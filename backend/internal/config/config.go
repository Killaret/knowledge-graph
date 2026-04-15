package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Сервер
	ServerPort string

	// База данных
	DatabaseURL string

	// Redis
	RedisURL string

	// NLP
	NLPServiceURL string

	// Рекомендации (граф)
	RecommendationAlpha      float64
	RecommendationBeta       float64
	RecommendationDepth      int
	RecommendationDecay      float64
	RecommendationCacheTTL   time.Duration
	EmbeddingSimilarityLimit int

	// Загрузка графа (визуализация)
	GraphLoadDepth int // Глубина загрузки графа для визуализации

	// Новые параметры для улучшения алгоритма рекомендаций
	RecommendationGamma float64 // Коэффициент для дополнительного компонента
	BFSAggregation      string  // Метод агрегации BFS: "max", "sum", "avg"
	BFSNormalize        bool    // Нормализация весов в BFS
	AsynqConcurrency    int     // Уровень параллелизма Asynq
	AsynqQueueDefault   int     // Приоритет дефолтной очереди Asynq
}

// Load загружает конфигурацию из переменных окружения.
// Если переменная не задана, используется значение по умолчанию.
// Для обязательных переменных (DatabaseURL) используется mustGetEnv.
func Load() *Config {
	return &Config{
		ServerPort:               getEnv("SERVER_PORT", "8080"),
		DatabaseURL:              mustGetEnv("DATABASE_URL"),
		RedisURL:                 getEnv("REDIS_URL", "localhost:6379"),
		NLPServiceURL:            getEnv("NLP_SERVICE_URL", "http://localhost:5000"),
		RecommendationAlpha:      getFloatEnv("RECOMMENDATION_ALPHA", 0.5),
		RecommendationBeta:       getFloatEnv("RECOMMENDATION_BETA", 0.5),
		RecommendationDepth:      getIntEnv("RECOMMENDATION_DEPTH", 3),
		RecommendationDecay:      getFloatEnv("RECOMMENDATION_DECAY", 0.5),
		RecommendationCacheTTL:   time.Duration(getIntEnv("RECOMMENDATION_CACHE_TTL_SECONDS", 300)) * time.Second,
		EmbeddingSimilarityLimit: getIntEnv("EMBEDDING_SIMILARITY_LIMIT", 30),

		// Загрузка графа
		GraphLoadDepth: getIntEnv("GRAPH_LOAD_DEPTH", 2),

		// Новые параметры
		RecommendationGamma: getFloatEnv("RECOMMENDATION_GAMMA", 0.2),
		BFSAggregation:      getEnv("BFS_AGGREGATION", "max"),
		BFSNormalize:        getBoolEnv("BFS_NORMALIZE", true),
		AsynqConcurrency:    getIntEnv("ASYNQ_CONCURRENCY", 10),
		AsynqQueueDefault:   getIntEnv("ASYNQ_QUEUE_DEFAULT", 1),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return value
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
