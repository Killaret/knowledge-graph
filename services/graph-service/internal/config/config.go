package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ── JSON Config Structure (matches knowledge-graph.config.json) ────

type JSONConfig struct {
	GraphService struct {
		GRPCPort     string `json:"grpc_port"`
		HTTPPort     string `json:"http_port"`
		FullLimit    int    `json:"full_limit"`
		DefaultDepth int    `json:"default_depth"`
		EventChannel string `json:"event_channel"`
		Cache        struct {
			NoteLayoutTTLSeconds int `json:"note_layout_ttl_seconds"`
			FullLayoutTTLSeconds int `json:"full_layout_ttl_seconds"`
			DeltaTTLSeconds      int `json:"delta_ttl_seconds"`
		} `json:"cache"`
		Layout struct {
			Radius2D        float64 `json:"2d_radius"`
			Radius3D        float64 `json:"3d_radius"`
			ZStep3D         float64 `json:"3d_z_step"`
			DefaultNodeSize float64 `json:"default_node_size"`
		} `json:"layout"`
		StreamChunkSize                      int `json:"stream_chunk_size"`
		EventTrackingTTLHours                int `json:"event_tracking_ttl_hours"`
		UnprocessedEventCheckIntervalMinutes int `json:"unprocessed_event_check_interval_minutes"`
	} `json:"graph_service"`
}

// ── Runtime Config ─────────────────────────────────────────────────

type Config struct {
	GRPCPort     string
	HTTPPort     string
	PostgresURL  string
	RedisURL     string
	EventChannel string
	FullLimit    int
	DefaultDepth int

	// Cache TTLs
	NoteLayoutTTL time.Duration
	FullLayoutTTL time.Duration
	DeltaTTL      time.Duration

	// Layout engine
	Layout2DRadius  float64
	Layout3DRadius  float64
	Layout3DZStep   float64
	DefaultNodeSize float64
	NodeTypeNote    string

	// Streaming
	StreamChunkSize int

	// Event processing
	EventTrackingTTL               time.Duration
	UnprocessedEventCheckInterval  time.Duration
	UnprocessedEventRetryThreshold time.Duration

	// Hash
	LayoutHashLength int

	// Auth
	JWTSecret         string
	InternalAuthToken string
	SkipAuth          bool

	// NLP
	NLPModelName string
}

// ── Load ───────────────────────────────────────────────────────────

func Load() (*Config, error) {
	jsonCfg := loadJSONConfig()

	cfg := &Config{
		GRPCPort:     getEnv("GRPC_PORT", getJSONString(jsonCfg, func(j *JSONConfig) string { return j.GraphService.GRPCPort }, "9090")),
		HTTPPort:     getEnv("HTTP_PORT", getJSONString(jsonCfg, func(j *JSONConfig) string { return j.GraphService.HTTPPort }, "9091")),
		PostgresURL:  getEnv("POSTGRES_URL", "postgresql://postgres:postgres@postgres:5432/knowledge_base?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "redis:6379"),
		EventChannel: getEnv("EVENT_CHANNEL", getJSONString(jsonCfg, func(j *JSONConfig) string { return j.GraphService.EventChannel }, "graph:events")),
		FullLimit:    getIntEnv("GRAPH_FULL_LIMIT", getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.FullLimit }, 1000)),
		DefaultDepth: getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.DefaultDepth }, 2),

		// Cache TTLs (env overrides JSON)
		NoteLayoutTTL: time.Duration(getIntEnv("CACHE_NOTE_TTL_SECONDS", getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.Cache.NoteLayoutTTLSeconds }, 300))) * time.Second,
		FullLayoutTTL: time.Duration(getIntEnv("CACHE_FULL_TTL_SECONDS", getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.Cache.FullLayoutTTLSeconds }, 300))) * time.Second,
		DeltaTTL:      time.Duration(getIntEnv("CACHE_DELTA_TTL_SECONDS", getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.Cache.DeltaTTLSeconds }, 60))) * time.Second,

		// Layout engine constants
		Layout2DRadius:  getJSONFloat(jsonCfg, func(j *JSONConfig) float64 { return j.GraphService.Layout.Radius2D }, 100.0),
		Layout3DRadius:  getJSONFloat(jsonCfg, func(j *JSONConfig) float64 { return j.GraphService.Layout.Radius3D }, 120.0),
		Layout3DZStep:   getJSONFloat(jsonCfg, func(j *JSONConfig) float64 { return j.GraphService.Layout.ZStep3D }, 5.0),
		DefaultNodeSize: getJSONFloat(jsonCfg, func(j *JSONConfig) float64 { return j.GraphService.Layout.DefaultNodeSize }, 1.0),
		NodeTypeNote:    "note",

		// Streaming
		StreamChunkSize: getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.StreamChunkSize }, 100),

		// Event processing
		EventTrackingTTL:               time.Duration(getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.EventTrackingTTLHours }, 24)) * time.Hour,
		UnprocessedEventCheckInterval:  time.Duration(getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.UnprocessedEventCheckIntervalMinutes }, 5)) * time.Minute,
		UnprocessedEventRetryThreshold: 5 * time.Minute,

		// Hash
		LayoutHashLength: 32,

		// Auth (env only; secrets are never read from JSON)
		JWTSecret:         getEnv("JWT_SECRET", ""),
		InternalAuthToken: getEnv("GRAPH_SERVICE_INTERNAL_TOKEN", ""),
		SkipAuth:          getBoolEnv("SKIP_AUTH", false),

		// NLP (env only)
		NLPModelName: getEnv("NLP_MODEL_NAME", "all-MiniLM-L6-v2"),
	}

	return cfg, nil
}

// ── JSON Loading ───────────────────────────────────────────────────

func loadJSONConfig() *JSONConfig {
	possiblePaths := []string{
		"knowledge-graph.config.json",
	}

	// Try relative to this file: services/graph-service/internal/config -> ../../../../knowledge-graph.config.json
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		configDir := filepath.Dir(filename)
		projectRoot := filepath.Join(configDir, "..", "..", "..", "..")
		possiblePaths = append(possiblePaths, filepath.Join(projectRoot, "knowledge-graph.config.json"))
	}

	var data []byte
	var err error
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil
	}

	var fullCfg JSONConfig
	if err := json.Unmarshal(data, &fullCfg); err != nil {
		return nil
	}
	return &fullCfg
}

// ── Helpers ────────────────────────────────────────────────────────

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if s := os.Getenv(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if s := os.Getenv(key); s != "" {
		s = strings.ToLower(s)
		if s == "true" || s == "1" || s == "yes" || s == "on" {
			return true
		}
		if s == "false" || s == "0" || s == "no" || s == "off" {
			return false
		}
	}
	return defaultValue
}

func getJSONString(jsonCfg *JSONConfig, getter func(*JSONConfig) string, defaultValue string) string {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

func getJSONInt(jsonCfg *JSONConfig, getter func(*JSONConfig) int, defaultValue int) int {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

func getJSONFloat(jsonCfg *JSONConfig, getter func(*JSONConfig) float64, defaultValue float64) float64 {
	if jsonCfg == nil {
		return defaultValue
	}
	return getter(jsonCfg)
}

// CacheKey builds a Redis key under the graph-service namespace.
func CacheKey(parts ...string) string {
	return fmt.Sprintf("graph-service:%s", strings.Join(parts, ":"))
}
