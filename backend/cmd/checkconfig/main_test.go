package main

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"knowledge-graph/internal/config"
)

func TestCheckZeroValues(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		issues int
	}{
		{
			name:   "valid values",
			input:  `{"backend": {"database": {"retry_max_attempts": 3, "retry_delay_seconds": 5}}}`,
			issues: 0,
		},
		{
			name:   "negative retry attempts",
			input:  `{"backend": {"database": {"retry_max_attempts": -1, "retry_delay_seconds": 5}}}`,
			issues: 1,
		},
		{
			name:   "negative delay",
			input:  `{"backend": {"database": {"retry_max_attempts": 3, "retry_delay_seconds": -5}}}`,
			issues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonCfg config.JSONConfig
			if err := json.Unmarshal([]byte(tt.input), &jsonCfg); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			var issues []string
			checkZeroValues(reflect.ValueOf(&jsonCfg).Elem(), "", &issues)

			if len(issues) != tt.issues {
				t.Errorf("expected %d issues, got %d: %v", tt.issues, len(issues), issues)
			}
		})
	}
}

func TestValidateCriticalFields(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*config.JSONConfig)
		issues int
	}{
		{
			name:   "valid config",
			issues: 0,
		},
		{
			name: "invalid recommendation depth",
			modify: func(c *config.JSONConfig) {
				c.Backend.Recommendation.Depth = 0
			},
			issues: 1,
		},
		{
			name: "alpha out of range",
			modify: func(c *config.JSONConfig) {
				c.Backend.Recommendation.Alpha = 1.5
			},
			issues: 1,
		},
		{
			name: "pagination max less than default",
			modify: func(c *config.JSONConfig) {
				c.Backend.Pagination.DefaultLimit = 20
				c.Backend.Pagination.MaxLimit = 10
			},
			issues: 1,
		},
		{
			name: "invalid bfs aggregation",
			modify: func(c *config.JSONConfig) {
				c.Backend.Recommendation.BFSAggregation = "invalid"
			},
			issues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validJSONConfig()
			if tt.modify != nil {
				tt.modify(&cfg)
			}

			issues := validateCriticalFields(&cfg)
			if len(issues) != tt.issues {
				t.Errorf("expected %d issues, got %d: %v", tt.issues, len(issues), issues)
			}
		})
	}
}

func TestValidateLoadedConfig(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*config.Config)
		issues int
	}{
		{
			name:   "valid config",
			issues: 0,
		},
		{
			name: "zero recommendation depth",
			modify: func(c *config.Config) {
				c.RecommendationDepth = 0
			},
			issues: 1,
		},
		{
			name: "pagination max less than default",
			modify: func(c *config.Config) {
				c.PaginationDefaultLimit = 20
				c.PaginationMaxLimit = 10
			},
			issues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validLoadedConfig()
			if tt.modify != nil {
				tt.modify(&cfg)
			}

			issues := validateLoadedConfig(&cfg)
			if len(issues) != tt.issues {
				t.Errorf("expected %d issues, got %d: %v", tt.issues, len(issues), issues)
			}
		})
	}
}

func validJSONConfig() config.JSONConfig {
	c := config.JSONConfig{}
	c.Backend.Recommendation.Depth = 3
	c.Backend.Recommendation.TopN = 5
	c.Backend.Recommendation.CacheTTLSeconds = 3600
	c.Backend.Recommendation.Alpha = 0.5
	c.Backend.Recommendation.Beta = 0.5
	c.Backend.Recommendation.Gamma = 0.5
	c.Backend.Recommendation.BatchRateLimit = 10
	c.Backend.Recommendation.TaskDelaySeconds = 0
	c.Backend.Recommendation.FallbackTTLSeconds = 3600
	c.Backend.Recommendation.BFSAggregation = "max"
	c.Backend.Database.RetryMaxAttempts = 3
	c.Backend.Database.RetryDelaySeconds = 5
	c.Backend.Graph.LoadDepth = 3
	c.Backend.Graph.DefaultLimit = 20
	c.Backend.Graph.MaxLimit = 100
	c.Backend.Graph.LinkDefaultLimit = 20
	c.Backend.Graph.LinkMaxLimit = 100
	c.Backend.Pagination.DefaultLimit = 20
	c.Backend.Pagination.MaxLimit = 100
	c.Backend.Server.RateLimit.Enabled = false
	c.Backend.Asynq.Concurrency = 10
	c.Backend.Asynq.QueueDefault = 1
	c.Backend.Asynq.QueueMaxLen = 1000
	c.Backend.Embedding.SimilarityLimit = 5
	return c
}

func validLoadedConfig() config.Config {
	c := config.Config{}
	c.RecommendationDepth = 3
	c.RecommendationCacheTTL = time.Hour
	c.GraphLoadDepth = 3
	c.PaginationDefaultLimit = 20
	c.PaginationMaxLimit = 100
	return c
}
