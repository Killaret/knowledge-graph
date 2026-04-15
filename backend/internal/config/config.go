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
